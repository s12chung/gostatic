package router

import (
	"fmt"
	"mime"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"

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
	".js":   "application/javascript",
	".png":  "image/png",
	".svg":  "image/svg+xml",
	".txt":  "text/plain; charset=utf-8",
	".xml":  "text/xml; charset=utf-8",
}

var AllGetTypesWithResponse = []struct {
	pattern  string
	mimeType string
	response string
}{
	{RootURL, "text/html; charset=utf-8", `<p>the root of it all</p>`},
	{"/page", "text/html; charset=utf-8", `<html>some page</html>`},
	{"/another_page", "text/html; charset=utf-8", `<html>another_page</html>`},
	{"/something.atom", "application/xml; charset=utf-8", `<?xml version="1.0" encoding="UTF-8"?>`},
	{"/robots.txt", "text/plain; charset=utf-8", "User-agent: *\nDisallow: /"},
	{"/files/haha.txt", "text/plain; charset=utf-8", "some test text"},
	{"/files/waa", "text/html; charset=utf-8", `<html>waa</html>`},
	{"/files/hmm.html", "text/html; charset=utf-8", `<html>hmm</html>`},
	{"/files/more/morez", "text/html; charset=utf-8", `<html>morez</html>`},
	{"/files/more/deep.txt", "text/plain; charset=utf-8", "deep"},
}

func SetupAllGetTypesWithResponse(router Router) {
	for _, allGetTypeWithResponse := range AllGetTypesWithResponse {
		response := allGetTypeWithResponse.response
		handler := func(ctx Context) error {
			ctx.Respond([]byte(response))
			return nil
		}

		pattern := allGetTypeWithResponse.pattern
		switch pattern {
		case RootURL:
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

	router.GetRootHTML(func(ctx Context) error {
		return nil
	})
	for _, htmlRoute := range allGetType.htmlRoutes {
		router.GetHTML(htmlRoute, func(ctx Context) error {
			return nil
		})
	}
	for index, route := range allGetType.otherRoutes {
		router.Get(route, func(ctx Context) error {
			ctx.SetContentType(allGetType.mimeTypes[index])
			return nil
		})
	}
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

func eachRouterSetup(t *testing.T, callback func(setup RouterSetup)) {
	setups := map[string]RouterSetup{
		"Generate": NewGenerateRouterSetup(),
		"Web":      NewWebRouterSetup(),
	}
	for name, setup := range setups {
		t.Log(name)
		callback(setup)
	}
}

func TestRouter_Around(t *testing.T) {
	eachRouterSetup(t, func(setup RouterSetup) {
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
			context := test.NewContext(t).SetFields(test.ContextFields{
				"index":       testCaseIndex,
				"handlersLen": len(tc.handlers),
			})

			router, _, _ := setup.DefaultRouter()
			router.GetRootHTML(func(ctx Context) error {
				testPreviousContext(ctx)

				got = append(got, "call")
				return nil
			})

			for _, handler := range tc.handlers {
				router.Around(handler)
			}

			setup.RunServer(router, func() {
				_, err := setup.Requester(router).Get(RootURL)
				context.AssertError(err, "Requester.Get")
				context.AssertArray("arounds", got, tc.expected)
			})
		}
	})
}

func TestRouter_GetInvalidRoute(t *testing.T) {
	eachRouterSetup(t, func(setup RouterSetup) {
		router, _, _ := setup.DefaultRouter()
		setup.RunServer(router, func() {
			response, err := setup.Requester(router).Get("/does_not_exist")
			if err == nil {
				t.Error("expecting error")
			}
			if response != nil {
				t.Error("expecting no response")
			}
		})
	})
}

func TestRouter_GetRootHTML(t *testing.T) {
	testRouteSetup(t, RootURL, "text/html; charset=utf-8", func(router Router, url string, handler ContextHandler) {
		if url == RootURL {
			router.GetRootHTML(handler)
		} else {
			router.GetHTML(url, handler)
		}
	})
}

func TestRouter_GetHTML(t *testing.T) {
	testRouteSetup(t, "/blah", "text/html; charset=utf-8", func(router Router, url string, handler ContextHandler) {
		router.GetHTML(url, handler)
	})
}

func TestRouter_Get(t *testing.T) {
	testRouteSetup(t, "/blah.atom", "application/xml; charset=utf-8", func(router Router, url string, handler ContextHandler) {
		router.Get(url, handler)
	})
}

func TestRouter_GetWithContentTypeSet(t *testing.T) {
	testRouteSetup(t, "/something.fakeext", "text/plain; charset=utf-8", func(router Router, url string, handler ContextHandler) {
		router.Get(url, func(ctx Context) error {
			ctx.SetContentType("text/plain; charset=utf-8")
			return handler(ctx)
		})
	})
}

func testRouteSetup(t *testing.T, url, contentType string, setRoute func(router Router, url string, handler ContextHandler)) {
	eachRouterSetup(t, func(setup RouterSetup) {
		testRouterContext(t, setup, url, contentType, setRoute)
		testRouterErrors(t, setup, url, setRoute)
		testRouterSlash(t, setup, url, setRoute)
	})
}

func testRouterContext(t *testing.T, setup RouterSetup, url, contentType string, setRoute func(router Router, url string, handler ContextHandler)) {
	called := false
	router, log, _ := setup.DefaultRouter()

	expResponse := "The Response"
	setRoute(router, url, func(ctx Context) error {
		called = true
		test.AssertLabel(t, "ctx.Log()", ctx.Log(), log)
		test.AssertLabel(t, "ctx.URL()", ctx.URL(), url)
		test.AssertLabel(t, "ctx.ContentType()", ctx.ContentType(), contentType)
		ctx.Respond([]byte(expResponse))
		return nil
	})
	setup.RunServer(router, func() {
		response, err := setup.Requester(router).Get(url)
		test.AssertError(t, err, "setup.Requester")
		test.AssertLabel(t, "response", string(response.Body), expResponse)
		test.AssertLabel(t, "called", called, true)
	})
}

func testRouterErrors(t *testing.T, setup RouterSetup, url string, setRoute func(router Router, url string, handler ContextHandler)) {
	expError := "test error"

	router, _, _ := setup.DefaultRouter()
	setRoute(router, url, func(ctx Context) error {
		return fmt.Errorf(expError)
	})

	setup.RunServer(router, func() {
		_, err := setup.Requester(router).Get(url)
		test.AssertLabel(t, "Handler error", err.Error(), expError)

		_, err = setup.Requester(router).Get("/multipart/url")
		if err == nil {
			t.Error("Multipart URLs are not giving errors")
		}
	})
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Did not panic for duplicate route setup.")
			}
		}()
		setRoute(router, url, func(ctx Context) error {
			return nil
		})
	}()
}

func testRouterSlash(t *testing.T, setup RouterSetup, url string, setRoute func(router Router, url string, handler ContextHandler)) {
	trimmedURL := strings.TrimLeft(url, "/")

	testCases := []struct {
		routeURL   string
		requestURL string
	}{
		{url, url},
		{trimmedURL, url},
		{url, trimmedURL},
		{trimmedURL, trimmedURL},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext(t).SetFields(test.ContextFields{
			"index":      testCaseIndex,
			"routeURL":   tc.routeURL,
			"requestURL": tc.requestURL,
		})

		router, _, _ := setup.DefaultRouter()
		setRoute(router, tc.routeURL, func(ctx Context) error {
			return nil
		})

		setup.RunServer(router, func() {
			_, err := setup.Requester(router).Get(tc.requestURL)
			context.AssertError(err, "Requester.Get")
		})
	}
}

func TestRouter_URLs(t *testing.T) {
	eachRouterSetup(t, func(setup RouterSetup) {
		for testCaseIndex, allGetType := range AllGetTypesVaried {
			context := test.NewContext(t).SetFields(test.ContextFields{
				"index":       testCaseIndex,
				"htmlRoutes":  allGetType.htmlRoutes,
				"otherRoutes": allGetType.otherRoutes,
			})

			router, _, _ := setup.DefaultRouter()
			SetupAllGetTypeVaried(router, allGetType)

			got := router.URLs()
			exp := append(allGetType.htmlRoutes, allGetType.otherRoutes...)
			exp = append(exp, RootURL)
			if allGetType.htmlRoutes == nil {
				exp = []string{}
			}

			sort.Strings(got)
			sort.Strings(exp)

			context.AssertArray("result", got, exp)
		}
	})
}

func TestRequester_Get(t *testing.T) {
	eachRouterSetup(t, func(setup RouterSetup) {
		router, _, _ := setup.DefaultRouter()
		SetupAllGetTypesWithResponse(router)

		setup.RunServer(router, func() {
			requester := setup.Requester(router)
			for getIndex, allGetTypeWithResponse := range AllGetTypesWithResponse {
				pattern := allGetTypeWithResponse.pattern
				context := test.NewContext(t).SetFields(test.ContextFields{
					"index":    getIndex,
					"pattern":  pattern,
					"mimeType": allGetTypeWithResponse.mimeType,
					"response": allGetTypeWithResponse.response,
				})

				response, err := requester.Get(pattern)
				context.AssertError(err, "requester.Get")
				context.Assert("Response.Body", string(response.Body), allGetTypeWithResponse.response)
				context.Assert("Response.ContentType", response.MimeType, allGetTypeWithResponse.mimeType)

				if pattern != RootURL {
					_, err := requester.Get(pattern[1:])
					if err != nil {
						t.Error(context.String("Can't handle requester.Get without / prefix"))
					}
				}
			}
		})
	})
}

func TestRequester_GetBadFolder(t *testing.T) {
	eachRouterSetup(t, func(setup RouterSetup) {
		testCases := []struct {
			urls  []string
			panic bool
		}{
			{[]string{"/blah"}, false},
			{[]string{"/blah", "/haha.atom"}, false},
			{[]string{"/blah", "/blah/haha.atom"}, true},
			{[]string{"/blah/haha.atom", "/blah"}, true},
			{[]string{"/blah", "/blah/haha"}, true},
			{[]string{"/blah/haha", "/blah"}, true},
			{[]string{"/blah/he", "/blah/haha.atom"}, false},
			{[]string{"/blah/he", "/blah/haha.atom", "/blah/wa.txt"}, false},
			{[]string{"/blah/he", "/blah/he/ni.atom"}, true},
			{[]string{"/blah/he/ni.atom", "/blah/he"}, true},
			{[]string{"/blah/he", "/blah/he/ni"}, true},
			{[]string{"/blah/he/ni", "/blah/he"}, true},
		}

		handler := func(ctx Context) error {
			ctx.Respond([]byte(ctx.URL()))
			return nil
		}

		for testCaseIndex, tc := range testCases {
			context := test.NewContext(t).SetFields(test.ContextFields{
				"index": testCaseIndex,
				"urls":  tc.urls,
			})

			router, _, _ := setup.DefaultRouter()
			router.GetRootHTML(handler)

			func() {
				if tc.panic {
					defer func() {
						if r := recover(); r == nil {
							t.Error(context.String("did not panic"))
						}
					}()
				}

				for _, url := range tc.urls {
					if path.Ext(url) == "" {
						router.GetHTML(url, handler)
					} else {
						router.Get(url, handler)
					}
				}
			}()
		}
	})
}
