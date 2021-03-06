package router

import (
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/gostatic/go/lib/utils"
	"github.com/s12chung/gostatic/go/test"
	"github.com/s12chung/gostatic/go/test/testfile"
)

func defaultWebRouter() (*WebRouter, logrus.FieldLogger, *logTest.Hook) {
	log, hook := logTest.NewNullLogger()
	return NewWebRouter(8080, log), log, hook
}

type WebRouterSetup struct {
	server *httptest.Server
}

func NewWebRouterSetup() *WebRouterSetup {
	return &WebRouterSetup{}
}

func (setup *WebRouterSetup) DefaultRouter() (Router, logrus.FieldLogger, *logTest.Hook) {
	return defaultWebRouter()
}

func (setup *WebRouterSetup) RunServer(router Router, callback func()) {
	r, ok := router.(*WebRouter)
	if !ok {
		panic("Not a *WebRouter being passed")
	}
	setup.server = httptest.NewServer(r.serveMux)
	callback()
	setup.server.Close()
}

func (setup *WebRouterSetup) Requester(router Router) Requester {
	if setup.server == nil {
		panic("Server not running, please run within RunServer callback")
	}

	urlObject, err := url.Parse(setup.server.URL)
	if err != nil {
		panic(err)
	}
	port, err := strconv.ParseInt(urlObject.Port(), 10, 32)
	if err != nil {
		panic(err)
	}

	requester := newWebRequester(int(port))
	requester.hostname = urlObject.Hostname()
	return requester
}

func TestWebRouter_FileServe(t *testing.T) {
	router, _, _ := defaultWebRouter()
	router.FileServe(fmt.Sprintf("/%v/", utils.CleanFilePath(testfile.FixturePath)), testfile.FixturePath)

	setup := NewWebRouterSetup()
	setup.RunServer(router, func() {
		filePaths, err := utils.FilePaths("", testfile.FixturePath)
		if err != nil {
			t.Fatal(err)
		}
		if len(contentTypes) != len(filePaths) {
			t.Error("Mime types does not match number of test files")
		}

		requester := setup.Requester(router)
		for index, filePath := range filePaths {
			context := test.NewContext(t).SetFields(test.ContextFields{
				"index":    index,
				"filePath": filePath,
			})
			response, err := requester.Get("/" + strings.Join([]string{utils.CleanFilePath(testfile.FixturePath), path.Base(filePath)}, "/"))
			context.AssertError(err, "requester.Get")
			context.Assert("mimeType", response.MimeType, contentTypes[path.Ext(filePath)])

			expBody, err := ioutil.ReadFile(path.Join(testfile.FixturePath, path.Base(filePath)))
			context.AssertError(err, "ioutil.ReadFile")
			context.AssertArray("Response.Body", response.Body, expBody)
		}
	})
}

func TestWebRouter_FileServe_PathChecks(t *testing.T) {
	router, _, _ := defaultWebRouter()
	router.FileServe(fmt.Sprintf("/%v/", utils.CleanFilePath(testfile.FixturePath)), testfile.FixturePath)

	setup := NewWebRouterSetup()
	setup.RunServer(router, func() {
		requester := setup.Requester(router)

		testCases := []struct {
			url      string
			hasError bool
		}{
			{"/test.atom", false},
			{"/dir/inner.atom", false},
			{"/../test.atom", true},
			{"/../dir/inner.atom", true},
			{"/dir/../inner.atom", true},
		}

		for index, tc := range testCases {
			context := test.NewContext(t).SetFields(test.ContextFields{
				"index": index,
				"url":   tc.url,
			})

			_, err := requester.Get("/" + strings.Join([]string{utils.CleanFilePath(testfile.FixturePath), tc.url}, "/"))
			if err != nil {
				if tc.hasError {
					continue
				}
				context.AssertError(err, "requester.Get")
			}
		}
	})
}
