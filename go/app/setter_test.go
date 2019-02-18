package app

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/s12chung/gostatic/go/lib/router"
	"github.com/s12chung/gostatic/go/test"
	"github.com/sirupsen/logrus"
	"testing"

	logTest "github.com/sirupsen/logrus/hooks/test"
)

func setupRouter(setRoutes func(r router.Router)) (*logTest.Hook, *router.Response, error) {
	log, hook := logTest.NewNullLogger()

	r := router.NewGenerateRouter(log)
	setRoutes(r)

	response, err := r.Requester().Get(router.RootURL)
	return hook, response, err
}

func TestSetDefaultAroundHandlers(t *testing.T) {
	testCases := []struct {
		content string
	}{
		{"Hi"},
		{""},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext(t).SetFields(test.ContextFields{
			"index":   testCaseIndex,
			"content": tc.content,
		})

		hook, response, err := setupRouter(func(r router.Router) {
			r.GetRootHTML(func(ctx router.Context) error {
				if tc.content == "" {
					return fmt.Errorf("err")
				}
				ctx.Respond([]byte(tc.content))
				return nil
			})
			SetDefaultAroundHandlers(r)
		})

		context.Assert("logEntries", len(hook.AllEntries()), 2)
		if tc.content == "" {
			if err == nil {
				t.Error(context.String("request did not return err"))
			}

			exp := []logrus.Level{logrus.InfoLevel, logrus.ErrorLevel}
			if !cmp.Equal(test.LogEntryLevels(hook), exp) {
				t.Error(context.AssertString("Log.Entry.Levels", test.LogEntryLevels(hook), exp))
			}

		} else {
			context.Assert("response", string(response.Body), tc.content)
			context.Assert("SafeLogEntries", test.SafeLogEntries(hook), true)
		}

		entryTestCases := []struct {
			dataLength int
		}{
			{2},
			{3},
		}

		for i, entryTc := range entryTestCases {
			entry := hook.AllEntries()[i]
			context.Assert(fmt.Sprintf("Log.Entry[%v].Data", i), len(entry.Data), entryTc.dataLength)
			context.Assert(fmt.Sprintf("Log.Entry[%v].Data.type", i), entry.Data["type"], LogRouteType)
			context.Assert(fmt.Sprintf("Log.Entry[%v].Data.URL", i), entry.Data["URL"], router.RootURL)
		}

		_, exists := hook.AllEntries()[1].Data["duration"]
		context.Assert("Log.Entry[1].Data.duration.exists", exists, true)
	}
}
