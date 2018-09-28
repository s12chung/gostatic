package router

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/gostatic/go/test"
)

func TestMain(m *testing.M) {
	err := setExtraMimeTypes()
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(m.Run())
}

var extraMimeTypes = map[string]bool{
	".atom": true,
	".ico":  true,
	".txt":  true,
}

var contentTypes = map[string]string{
	".atom": "application/xml; charset=utf-8",
	".css":  "text/css; charset=utf-8",
	".gif":  "image/gif",
	".html": "text/html; charset=utf-8",
	".ico":  "image/x-icon",
	".jpg":  "image/jpeg",
	".js":   "application/x-javascript",
	".png":  "image/png",
	".svg":  "image/svg+xml",
	".txt":  "text/plain; charset=utf-8",
	".xml":  "text/xml; charset=utf-8",
}

func setExtraMimeTypes() error {
	for ext := range extraMimeTypes {
		err := mime.AddExtensionType(ext, contentTypes[ext])
		if err != nil {
			return err
		}
	}
	return nil
}

type RouterSetup interface {
	DefaultRouter() (Router, logrus.FieldLogger, *logTest.Hook)
	RunServer(router Router, callback func())
	Requester(router Router) Requester
}

type RouterTester struct {
	setup RouterSetup
}

func NewRouterTester(setup RouterSetup) *RouterTester {
	return &RouterTester{setup}
}

func (tester *RouterTester) TestRouter_Around(t *testing.T) {
	var got []string
	var previousContext Context

	testPreviousContext := func(ctx Context) {
		if previousContext == nil {
			previousContext = ctx
		} else {
			test.AssertLabel(t, "ctx", ctx, previousContext)
		}
	}

	h := func(before, after string) AroundHandler {
		return func(ctx Context, handler ContextHandler) error {
			testPreviousContext(ctx)

			if before != "" {
				got = append(got, before)
			}
			err := handler(ctx)
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
		previousContext = nil
		context := test.NewContext().SetFields(test.ContextFields{
			"index":       testCaseIndex,
			"handlersLen": len(tc.handlers),
		})

		router, _, _ := tester.setup.DefaultRouter()
		router.GetRootHTML(func(ctx Context) error {
			testPreviousContext(ctx)

			got = append(got, "call")
			return nil
		})

		for _, handler := range tc.handlers {
			router.Around(handler)
		}

		tester.setup.RunServer(router, func() {
			_, err := tester.setup.Requester(router).Get(RootURLPattern)
			if err != nil {
				t.Error(context.String(err))
			}
			if !cmp.Equal(got, tc.expected) {
				t.Error(context.GotExpString("state", got, tc.expected))
			}
		})
	}
}

var AllGetTypesWithResponse = []struct {
	pattern  string
	mimeType string
	response string
}{
	{RootURLPattern, "text/html; charset=utf-8", `<p>the root of it all</p>`},
	{"/page", "text/html; charset=utf-8", `<html>some page</html>`},
	{"/another_page", "text/html; charset=utf-8", `<html>another_page</html>`},
	{"/something.atom", "application/xml; charset=utf-8", `<?xml version="1.0" encoding="UTF-8"?>`},
	{"/robots.txt", "text/plain; charset=utf-8", "User-agent: *\nDisallow: /"},
}

func SetupAllGetTypesWithResponse(router Router) {
	for _, allGetTypeWithResponse := range AllGetTypesWithResponse {
		response := allGetTypeWithResponse.response
		handler := func(ctx Context) error {
			return ctx.Respond([]byte(response))
		}

		pattern := allGetTypeWithResponse.pattern
		switch pattern {
		case RootURLPattern:
			router.GetRootHTML(handler)
		default:
			if filepath.Ext(pattern) == "" {
				router.GetHTML(pattern, handler)
			} else {
				router.Get(pattern, handler)
			}
		}

	}
}

type AllGetType struct {
	htmlRoutes  []string
	otherRoutes []string
	mimeTypes   []string
}

var AllGetTypesVaried = []AllGetType{
	{nil, nil, nil},
	{[]string{}, []string{}, []string{}},
	{[]string{"/some"}, []string{"/something.atom"}, []string{"application/xml"}},
	{[]string{"/some", "/ha", "/works"}, []string{"/something.atom", "/robots.txt"}, []string{"application/xml", "text/plain"}},
}

func SetupAllGetTypeVaried(router Router, allGetType AllGetType) {
	if allGetType.htmlRoutes == nil {
		return
	}

	handler := func(ctx Context) error {
		return nil
	}

	router.GetRootHTML(handler)

	for _, htmlRoute := range allGetType.htmlRoutes {
		router.GetHTML(htmlRoute, handler)
	}
	for _, route := range allGetType.otherRoutes {
		router.Get(route, handler)
	}
}

func (tester *RouterTester) TestRequester_Get(t *testing.T) {
	router, _, _ := tester.setup.DefaultRouter()
	SetupAllGetTypesWithResponse(router)

	tester.setup.RunServer(router, func() {
		requeseter := tester.setup.Requester(router)
		for getIndex, allGetTypeWithResponse := range AllGetTypesWithResponse {
			context := test.NewContext().SetFields(test.ContextFields{
				"index":    getIndex,
				"pattern":  allGetTypeWithResponse.pattern,
				"mimeType": allGetTypeWithResponse.mimeType,
				"response": allGetTypeWithResponse.response,
			})

			response, err := requeseter.Get(allGetTypeWithResponse.pattern)
			if err != nil {
				t.Errorf(context.String(err))
			}

			got := string(response.Body)
			exp := allGetTypeWithResponse.response
			if got != exp {
				t.Error(context.GotExpString("Response.Body", got, exp))
			}

			got = response.MimeType
			exp = allGetTypeWithResponse.mimeType
			if got != exp {
				t.Error(context.GotExpString("Response.ContentType", got, exp))
			}
		}
	})
}

func (tester *RouterTester) NewGetTester(requestURL, contentType string, testFunc func(router Router, handler ContextHandler)) *GetTester {
	if testFunc == nil {
		testFunc = func(router Router, handler ContextHandler) {}
	}
	return &GetTester{
		tester.setup,
		requestURL,
		contentType,
		testFunc,
	}
}

type GetTester struct {
	setup       RouterSetup
	requestURL  string
	contentType string
	testFunc    func(router Router, handler ContextHandler)
}

func (getTester *GetTester) Test_Get(t *testing.T) {
	getTester.testRouterContext(t)
	getTester.testRouterErrors(t)
}

func (getTester *GetTester) testRouterContext(t *testing.T) {
	called := false
	expResponse := "The Response"
	router, log, _ := getTester.setup.DefaultRouter()
	getTester.testFunc(router, func(ctx Context) error {
		called = true
		test.AssertLabel(t, "ctx.Log()", ctx.Log(), log)
		test.AssertLabel(t, "ctx.URL()", ctx.URL(), getTester.requestURL)
		test.AssertLabel(t, "ctx.ContentType()", ctx.ContentType(), getTester.contentType)
		return ctx.Respond([]byte(expResponse))
	})
	getTester.setup.RunServer(router, func() {
		response, err := getTester.setup.Requester(router).Get(getTester.requestURL)
		if err != nil {
			t.Error(err)
		}
		test.AssertLabel(t, "response", string(response.Body), expResponse)
		test.AssertLabel(t, "called", called, true)
	})
}

func (getTester *GetTester) testRouterErrors(t *testing.T) {
	expError := "test error"
	router, _, _ := getTester.setup.DefaultRouter()
	getTester.testFunc(router, func(ctx Context) error {
		return fmt.Errorf(expError)
	})

	getTester.setup.RunServer(router, func() {
		_, err := getTester.setup.Requester(router).Get(getTester.requestURL)
		test.AssertLabel(t, "Handler error", err.Error(), expError)

		_, err = getTester.setup.Requester(router).Get("/multipart/url")
		if err == nil {
			t.Error("Multipart Urls are not giving errors")
		}
	})
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Did not panic for duplicate route setup.")
			}
		}()
		getTester.testFunc(router, func(ctx Context) error {
			return nil
		})
	}()
}

func (tester *RouterTester) TestRouter_GetInvalidRoute(t *testing.T) {
	router, _, _ := tester.setup.DefaultRouter()
	tester.setup.RunServer(router, func() {
		response, err := tester.setup.Requester(router).Get("/does_not_exist")
		if err == nil {
			t.Error("expecting error")
		}
		if response != nil {
			t.Error("expecting no response")
		}
	})
}

func (tester *RouterTester) TestRouter_GetRootHTML(t *testing.T) {
	tester.NewGetTester(RootURLPattern, "text/html; charset=utf-8", func(router Router, handler ContextHandler) {
		router.GetRootHTML(handler)
	}).Test_Get(t)
}

func (tester *RouterTester) TestRouter_GetHTML(t *testing.T) {
	tester.NewGetTester("/blah", "text/html; charset=utf-8", func(router Router, handler ContextHandler) {
		router.GetHTML("/blah", handler)
	}).Test_Get(t)
}

func (tester *RouterTester) TestRouter_Get(t *testing.T) {
	tester.NewGetTester("/blah.atom", "application/xml; charset=utf-8", func(router Router, handler ContextHandler) {
		router.Get("/blah.atom", handler)
	}).Test_Get(t)
}

func (tester *RouterTester) TestRouter_GetWithContentTypeSet(t *testing.T) {
	tester.NewGetTester("/something.fakeext", "text/plain; charset=utf-8", func(router Router, handler ContextHandler) {
		router.Get("/something.fakeext", func(ctx Context) error {
			ctx.SetContentType("text/plain; charset=utf-8")
			return handler(ctx)
		})
	}).Test_Get(t)
}

func (tester *RouterTester) TestRouter_Urls(t *testing.T) {
	for testCaseIndex, allGetType := range AllGetTypesVaried {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":       testCaseIndex,
			"htmlRoutes":  allGetType.htmlRoutes,
			"otherRoutes": allGetType.otherRoutes,
		})

		router, _, _ := tester.setup.DefaultRouter()
		SetupAllGetTypeVaried(router, allGetType)

		got := router.Urls()
		exp := append(allGetType.htmlRoutes, allGetType.otherRoutes...)
		exp = append(exp, RootURLPattern)
		if allGetType.htmlRoutes == nil {
			exp = []string{}
		}

		sort.Strings(got)
		sort.Strings(exp)

		if !cmp.Equal(got, exp) {
			t.Error(context.GotExpString("Result", got, exp))
		}
	}
}
