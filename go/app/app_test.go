package app

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/gostatic/go/lib/router"
	"github.com/s12chung/gostatic/go/test"
	"github.com/s12chung/gostatic/go/test/mocks"
	"github.com/s12chung/gostatic/go/test/testfile"
)

//go:generate mockgen -destination=../test/mocks/app_setter.go -package=mocks github.com/s12chung/gostatic/go/app Setter

func defaultApp(setter Setter, generatedPath string) (*App, logrus.FieldLogger, *logTest.Hook) {
	settings := DefaultSettings()
	settings.GeneratedPath = generatedPath
	log, hook := logTest.NewNullLogger()
	return NewApp(setter, settings, log), log, hook
}

func runGenerate(t *testing.T, setter Setter, callback func(generatedPath string)) {
	generatedPath, clean := testfile.SandboxDir(t, "generated")
	defer clean()

	app, _, _ := defaultApp(setter, generatedPath)
	test.AssertError(t, app.Generate(), "app.Generate()")
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
			test.AssertError(t, err, "filepath.Walk()")
			test.AssertLabel(t, "filename len", len(generatedFiles), 7)

			for index, generatedFile := range generatedFiles {
				context := test.NewContext(t).SetFields(test.ContextFields{
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
				context.Assert("File Contents", got, exp)
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

		context := test.NewContext(t).SetFields(test.ContextFields{
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

func TestApp_Around(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	var got []string

	h := func(before, after string) AroundHandler {
		return func(handler func() error) error {
			if before != "" {
				got = append(got, before)
			}
			err := handler()
			if after != "" {
				got = append(got, after)
			}
			return err
		}
	}

	testCases := []struct {
		handlers []AroundHandler
		expected []string
	}{
		{[]AroundHandler{}, []string{"call"}},
		{[]AroundHandler{h("b1", "")}, []string{"b1", "call"}},
		{[]AroundHandler{h("b1", ""), h("b2", "")}, []string{"b1", "b2", "call"}},
		{[]AroundHandler{h("", "a1")}, []string{"call", "a1"}},
		{[]AroundHandler{h("", "a1"), h("", "a2")}, []string{"call", "a2", "a1"}},
		{[]AroundHandler{h("ar1", "ar2")}, []string{"ar1", "call", "ar2"}},
		{[]AroundHandler{h("ar1", "ar2"), h("arr1", "arr2")}, []string{"ar1", "arr1", "call", "arr2", "ar2"}},
		{[]AroundHandler{h("ar1", "ar2"), h("", "a1"), h("b1", ""), h("arr1", "arr2")}, []string{"ar1", "b1", "arr1", "call", "arr2", "a1", "ar2"}},
	}

	for testCaseIndex, tc := range testCases {
		got = nil
		context := test.NewContext(t).SetFields(test.ContextFields{
			"index":       testCaseIndex,
			"handlersLen": len(tc.handlers),
		})

		setter := mocks.NewMockSetter(controller)
		setter.EXPECT().SetRoutes(gomock.Any()).DoAndReturn(func(r router.Router) error {
			got = append(got, "call")
			return nil
		})
		setter.EXPECT().URLBatches(gomock.Any())

		generatedPath, clean := testfile.SandboxDir(t, "generated")
		app, _, _ := defaultApp(setter, generatedPath)
		for _, handler := range tc.handlers {
			app.Around(handler)
		}

		context.AssertError(app.Generate(), "app.Generate")
		if !cmp.Equal(got, tc.expected) {
			t.Error(context.AssertString("state", got, tc.expected))
		}

		clean()
	}
}
