package cli

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/gostatic/go/test"
	"github.com/s12chung/gostatic/go/test/mocks"
)

//go:generate mockgen -destination=../test/mocks/cli_app.go -package=mocks github.com/s12chung/gostatic/go/cli App

func TestRunDefault(t *testing.T) {
	testCases := []struct {
		args         []string
		functionName string
	}{
		{nil, "Generate"},
		{[]string{}, "Generate"},
		{[]string{"-file-server"}, "RunFileServer"},
		{[]string{"-server"}, "Host"},
		{[]string{"-file-server", "-server"}, "RunFileServer"},
		{[]string{"-blah"}, ""},
		{[]string{"-file-server", "-blah"}, ""},
	}
	for testCaseIndex, tc := range testCases {
		context := test.NewContext(t).SetFields(test.ContextFields{
			"index":        testCaseIndex,
			"args":         tc.args,
			"functionName": tc.functionName,
		})

		controller := gomock.NewController(t)
		app := mocks.NewMockApp(controller)
		expect := app.EXPECT()

		expect.GeneratedPath().Return("the_generated")
		expect.FileServerPort().Return(999)
		expect.ServerPort().Return(100)

		map[string]func() *gomock.Call{
			"":              func() *gomock.Call { return nil },
			"Generate":      expect.Generate,
			"RunFileServer": expect.RunFileServer,
			"Host":          expect.Host,
		}[tc.functionName]()

		context.AssertError(Run("random name", app, tc.args), "Run")

		t.Log(context.FieldsString())
		controller.Finish()
	}
}

func TestSetDefaultAppARoundHandlers(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	log, hook := logTest.NewNullLogger()

	app := mocks.NewMockApp(controller)
	app.EXPECT().Log().Return(log)
	app.EXPECT().Around(gomock.Any()).DoAndReturn(func(handler func(handler func() error) error) {
		err := handler(func() error {
			return nil
		})
		test.AssertError(t, err, "handler")
	})

	SetDefaultAppARoundHandlers(app)

	test.AssertArray(t, "test.LogEntryLevels", test.LogEntryLevels(hook), []logrus.Level{logrus.InfoLevel})
}
