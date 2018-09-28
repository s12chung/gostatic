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

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/gostatic/go/lib/utils"
	"github.com/s12chung/gostatic/go/test"
)

func defaultWebRouter() (*WebRouter, logrus.FieldLogger, *logTest.Hook) {
	log, hook := logTest.NewNullLogger()
	return NewWebRouter(8080, log), log, hook
}

func webRouterTester() *RouterTester {
	return NewRouterTester(NewWebRouterSetup())
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

func TestWebRouter_Around(t *testing.T) {
	webRouterTester().TestRouter_Around(t)
}

func TestWebRouter_GetInvalidRoute(t *testing.T) {
	generateRouterTester().TestRouter_GetInvalidRoute(t)
}

func TestWebRouter_GetRootHTML(t *testing.T) {
	webRouterTester().TestRouter_GetRootHTML(t)
}

func TestWebRouter_GetHTML(t *testing.T) {
	webRouterTester().TestRouter_GetHTML(t)
}

func TestWebRouter_Get(t *testing.T) {
	webRouterTester().TestRouter_Get(t)
}

func TestWebRouterRouter_GetWithContentTypeSet(t *testing.T) {
	generateRouterTester().TestRouter_GetWithContentTypeSet(t)
}

func TestWebRouter_StaticUrls(t *testing.T) {
	webRouterTester().TestRouter_StaticUrls(t)
}

func TestWebRouter_FileServe(t *testing.T) {
	router, _, _ := defaultWebRouter()
	SetupAllGetTypesWithResponse(router)
	router.FileServe(fmt.Sprintf("/%v/", utils.CleanFilePath(test.FixturePath)), test.FixturePath)

	setup := NewWebRouterSetup()
	setup.RunServer(router, func() {
		requester := setup.Requester(router)

		for getIndex, allGetTypeWithResponse := range AllGetTypesWithResponse {
			context := test.NewContext().SetFields(test.ContextFields{
				"index":    getIndex,
				"pattern":  allGetTypeWithResponse.pattern,
				"mimeType": allGetTypeWithResponse.mimeType,
				"response": allGetTypeWithResponse.response,
			})

			reqUrl := allGetTypeWithResponse.pattern
			_, err := requester.Get(reqUrl)
			if err != nil {
				t.Error(context.String(err))
			}
		}

		filePaths, err := utils.FilePaths("", test.FixturePath)
		if err != nil {
			t.Fatal(err)
		}
		if len(contentTypes) != len(filePaths) {
			t.Error("Mime types does not match number of test files")
		}

		failedMimeExts := testAllFiles(t, filePaths, requester)
		for ext, got := range failedMimeExts {
			context := test.NewContext().SetFields(test.ContextFields{
				"ext": ext,
			})
			t.Error(context.GotExpString("mimeType", got, contentTypes[ext]))
		}
	})
}

func testAllFiles(t *testing.T, filePaths []string, requester Requester) map[string]string {
	failedMimeTypes := make(map[string]string)
	for _, filePath := range filePaths {
		context := test.NewContext().SetFields(test.ContextFields{
			"filePath": filePath,
		})
		response, err := requester.Get("/" + strings.Join([]string{utils.CleanFilePath(test.FixturePath), path.Base(filePath)}, "/"))
		if err != nil {
			t.Error(context.String(err))
		}

		ext := path.Ext(filePath)
		if response.MimeType != contentTypes[ext] {
			failedMimeTypes[ext] = response.MimeType
		}

		expBody, err := ioutil.ReadFile(path.Join(test.FixturePath, path.Base(filePath)))
		if err != nil {
			t.Error(context.String(err))
		}
		if !cmp.Equal(response.Body, expBody) {
			t.Error(context.GotExpString("Response.Body", response.Body, expBody))
		}
	}
	return failedMimeTypes
}
