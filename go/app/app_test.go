package app_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/gostatic/go/app"
	"github.com/s12chung/gostatic/go/lib/router"
	"github.com/s12chung/gostatic/go/test"
	"github.com/s12chung/gostatic/go/test/mocks"
)

func defaultApp(setter app.Setter, generatedPath string) (*app.App, logrus.FieldLogger, *logTest.Hook) {
	settings := app.DefaultSettings()
	settings.GeneratedPath = generatedPath
	log, hook := logTest.NewNullLogger()
	return app.NewApp(setter, settings, log), log, hook
}

func runGenerate(t *testing.T, setter app.Setter, callback func(generatedPath string)) {
	generatedPath, clean := test.SandboxDir(t, "generated")
	defer clean()

	a, _, _ := defaultApp(setter, generatedPath)
	if err := a.Generate(); err != nil {
		t.Error(err)
	}

	callback(generatedPath)
}

func TestApp_Generate(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	setter := mocks.NewMockSetter(controller)
	setter.EXPECT().SetRoutes(gomock.Any()).Do(func(r router.Router) error {
		handler := func(ctx router.Context) error {
			ctx.Respond([]byte(ctx.URL()))
			return nil
		}

		r.GetRootHTML(handler)
		r.GetHTML("/fold/me", handler)
		r.GetHTML("/fold/me_again", handler)
		r.GetHTML("/fold/go.json", handler)
		r.GetHTML("/fold/deeper/in.txt", handler)
		r.GetHTML("/fold/deeper/out", handler)
		r.GetHTML("/fold/deeper/with", handler)

		return nil
	})
	setter.EXPECT().URLBatches(gomock.Any()).DoAndReturn(func(r router.Router) ([][]string, error) {
		return [][]string{r.URLs()}, nil
	})

	runGenerate(t,
		setter,
		func(generatedPath string) {
			var generatedFiles []string
			err := filepath.Walk(generatedPath, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				generatedFiles = append(generatedFiles, path)
				return nil
			})
			if err != nil {
				t.Error(err)
			}

			test.AssertLabel(t, "filename len", len(generatedFiles), 7)

			for index, generatedFile := range generatedFiles {
				context := test.NewContext().SetFields(test.ContextFields{
					"index":         index,
					"generatedFile": generatedFile,
				})

				bytes, err := ioutil.ReadFile(generatedFile)
				if err != nil {
					t.Error(context.String(err))
				}
				got := strings.TrimSpace(string(bytes))

				exp := strings.TrimPrefix(generatedFile, generatedPath)
				if exp == "/index.html" {
					exp = "/"
				}
				if got != exp {
					t.Error(context.GotExpString("File Contents", got, exp))
				}
			}
		},
	)
}

func TestApp_Generate_Order(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	testCases := []struct {
		isSecond []bool
	}{
		{[]bool{true}},
		{[]bool{false}},
		{[]bool{true, false, true, false, false}},
		{[]bool{false, true, false, true, true}},
		{[]bool{false, true, true, false}},
		{[]bool{true, false, false, true}},
	}

	isSecondURL := func(url string) bool {
		return strings.Contains(url, "/second")
	}

	for testCaseIndex, tc := range testCases {
		var requestOrder []string
		setter := mocks.NewMockSetter(controller)
		setter.EXPECT().SetRoutes(gomock.Any()).DoAndReturn(func(r router.Router) error {
			handler := func(ctx router.Context) error {
				requestOrder = append(requestOrder, ctx.URL())
				ctx.Respond([]byte(ctx.URL()))
				return nil
			}

			r.GetRootHTML(handler)

			for i, isSecond := range tc.isSecond {
				url := fmt.Sprintf("/first-%v", i)
				if isSecond {
					url = fmt.Sprintf("/second-%v", i)
				}
				r.GetHTML(url, handler)
			}
			return nil
		})
		setter.EXPECT().URLBatches(gomock.Any()).DoAndReturn(func(r router.Router) ([][]string, error) {
			var first []string
			var second []string

			for _, url := range r.URLs() {
				if isSecondURL(url) {
					second = append(second, url)
				} else {
					first = append(first, url)
				}
			}
			return [][]string{first, second}, nil
		})

		context := test.NewContext().SetFields(test.ContextFields{
			"index":    testCaseIndex,
			"isSecond": tc.isSecond,
		})

		runGenerate(t,
			setter,
			func(generatedPath string) {
				test.AssertLabel(t, "requestOrder len", len(requestOrder), len(tc.isSecond)+1)

				reachedSecond := false
				for _, url := range requestOrder {
					if isSecondURL(url) {
						reachedSecond = true
					} else {
						if reachedSecond {
							t.Error(context.Stringf("A /second route went before other route"))
						}
					}
				}
			},
		)
	}
}
