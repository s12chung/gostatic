package app

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/gostatic/go/lib/router"
	"github.com/s12chung/gostatic/go/test"
	"github.com/s12chung/gostatic/go/test/mocks"
)

//go:generate mockgen -destination=../test/mocks/router_router.go -package=mocks github.com/s12chung/gostatic/go/lib/router Router
//go:generate mockgen -destination=../test/mocks/router_context.go -package=mocks github.com/s12chung/gostatic/go/lib/router Context

func TestSetDefaultRouterAroundHandlers(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	testCases := []struct {
		err error
	}{
		{nil},
		{fmt.Errorf("err")},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext(t).SetFields(test.ContextFields{
			"index": testCaseIndex,
			"err":   tc.err,
		})

		log, hook := logTest.NewNullLogger()
		url := "some_url"
		message := "waka"

		var ctxLog logrus.FieldLogger = log
		ctx := mocks.NewMockContext(controller)
		ctx.EXPECT().Log().AnyTimes().DoAndReturn(func() logrus.FieldLogger {
			return ctxLog
		})
		ctx.EXPECT().SetLog(gomock.Any()).DoAndReturn(func(log logrus.FieldLogger) {
			ctxLog = log
		})
		ctx.EXPECT().URL().Return(url)

		var err error
		r := mocks.NewMockRouter(controller)
		r.EXPECT().Around(gomock.Any()).DoAndReturn(func(handler router.AroundHandler) {
			err = handler(ctx, func(ctx router.Context) error {
				ctx.Log().Info(message)
				return tc.err
			})
		})

		SetDefaultRouterAroundHandlers(r)

		context.Assert("logEntries", len(hook.AllEntries()), 3)
		if tc.err == nil {
			context.Assert("SafeLogEntries", test.SafeLogEntries(hook), true)
		} else {
			if err == nil {
				t.Error(context.String("request did not return err"))
			}

			exp := []logrus.Level{logrus.InfoLevel, logrus.InfoLevel, logrus.ErrorLevel}
			if !cmp.Equal(test.LogEntryLevels(hook), exp) {
				t.Error(context.AssertString("Log.Entry.Levels", test.LogEntryLevels(hook), exp))
			}
		}

		entryTestCases := []struct {
			dataLength int
		}{
			{2},
			{2},
			{3},
		}

		for i, entryTc := range entryTestCases {
			entry := hook.AllEntries()[i]
			context.Assert(fmt.Sprintf("Log.Entry[%v].Data", i), len(entry.Data), entryTc.dataLength)
			context.Assert(fmt.Sprintf("Log.Entry[%v].Data.type", i), entry.Data["type"], LogRouteType)
			context.Assert(fmt.Sprintf("Log.Entry[%v].Data.URL", i), entry.Data["URL"], url)
		}

		context.Assert("Log.Entry[1].Message", hook.AllEntries()[1].Message, message)

		_, exists := hook.AllEntries()[2].Data["duration"]
		context.Assert("Log.Entry[2].Data.duration.exists", exists, true)
	}
}
