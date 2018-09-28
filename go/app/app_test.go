package app_test

import (
	"io/ioutil"
	"path"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/gostatic/go/app"
	"github.com/s12chung/gostatic/go/lib/router"
	"github.com/s12chung/gostatic/go/lib/utils"
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

func TestApp_Generate(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	setter := mocks.NewMockSetter(controller)
	setter.EXPECT().SetRoutes(gomock.Any(), gomock.Any()).Do(func(r router.Router, tracker *app.Tracker) {
		handler := func(ctx router.Context) error {
			return ctx.Respond([]byte(ctx.URL()))
		}
		r.GetRootHTML(handler)
		r.GetHTML("/dep", handler)
		r.GetHTML("/non_dep", handler)
		tracker.AddDependentUrl("/dep")
	})

	generatedPath, clean := test.SandboxDir(t, "generated")
	defer clean()

	a, _, _ := defaultApp(setter, generatedPath)
	a.Generate()

	filenames := []string{"index.html", "dep", "non_dep"}
	generatedFiles, err := utils.FilePaths("", generatedPath)
	if err != nil {
		t.Error(err)
	}

	got := len(generatedFiles)
	test.AssertLabel(t, "filename len", got, len(filenames))

	for _, filename := range filenames {
		context := test.NewContext().SetFields(test.ContextFields{
			"filename": filename,
		})

		bytes, err := ioutil.ReadFile(path.Join(generatedPath, filename))
		if err != nil {
			t.Error(context.String(err))
		}
		got := strings.TrimSpace(string(bytes))
		exp := "/" + filename
		if filename == "index.html" {
			exp = "/"
		}
		test.AssertLabel(t, "File", got, exp)
	}
}
