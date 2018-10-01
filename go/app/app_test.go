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

//go:generate mockgen -destination=../test/mocks/app_setter.go -package=mocks github.com/s12chung/gostatic/go/app Setter

func defaultApp(setter app.Setter, generatedPath string) (*app.App, logrus.FieldLogger, *logTest.Hook) {
	settings := app.DefaultSettings()
	settings.GeneratedPath = generatedPath
	log, hook := logTest.NewNullLogger()
	return app.NewApp(setter, settings, log), log, hook
}

func setupGenerate(t *testing.T, setRoutes func(r router.Router, tracker *app.Tracker), callback func(generatedPath string)) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	setter := mocks.NewMockSetter(controller)
	setter.EXPECT().SetRoutes(gomock.Any(), gomock.Any()).Do(setRoutes)

	generatedPath, clean := test.SandboxDir(t, "generated")
	defer clean()

	a, _, _ := defaultApp(setter, generatedPath)
	err := a.Generate()
	if err != nil {
		t.Error(err)
	}

	callback(generatedPath)
}

func TestApp_Generate(t *testing.T) {
	setupGenerate(t,
		func(r router.Router, tracker *app.Tracker) {
			handler := func(ctx router.Context) error {
				ctx.Respond([]byte(ctx.URL()))
				return nil
			}

			r.GetRootHTML(handler)
			r.GetHTML("/dep", handler)
			r.GetHTML("/non_dep", handler)
			r.GetHTML("/fold/me", handler)
			r.GetHTML("/fold/me_again", handler)
			r.GetHTML("/fold/go.json", handler)
			r.GetHTML("/fold/deeper/in.txt", handler)
			r.GetHTML("/fold/deeper/out", handler)
			r.GetHTML("/fold/deeper/with", handler)
			tracker.AddDependentURL("/dep")
		},
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

			got := len(generatedFiles)
			test.AssertLabel(t, "filename len", got, 9)

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
	testCases := []struct {
		urlDeps []bool
	}{
		{[]bool{true}},
		{[]bool{false}},
		{[]bool{true, false, true, false, false}},
		{[]bool{false, true, false, true, true}},
		{[]bool{false, true, true, false}},
		{[]bool{true, false, false, true}},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":   testCaseIndex,
			"urlDeps": tc.urlDeps,
		})

		var requestOrder []string
		setupGenerate(t,
			func(r router.Router, tracker *app.Tracker) {
				handler := func(ctx router.Context) error {
					requestOrder = append(requestOrder, ctx.URL())
					ctx.Respond([]byte(ctx.URL()))
					return nil
				}

				r.GetRootHTML(handler)

				for i, dep := range tc.urlDeps {
					url := fmt.Sprintf("/non_dep%v", i)
					if dep {
						url = fmt.Sprintf("/dep%v", i)
						tracker.AddDependentURL(url)
					}
					r.GetHTML(url, handler)
				}
			},
			func(generatedPath string) {
				reachedDep := false
				for _, url := range requestOrder {
					if strings.Contains(url, "/non_dep") {
						if reachedDep {
							t.Error(context.Stringf("Contains a non-dep after dep"))
						}
					} else {
						if url != router.RootURL {
							reachedDep = true
						}
					}
				}
			},
		)
	}
}
