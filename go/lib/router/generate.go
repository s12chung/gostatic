package router

import (
	"net/http"
	"strconv"

	"fmt"
	"github.com/sirupsen/logrus"
	"mime"
	"path"
)

// RunFileServer hosts the files of targetDir into given port with the log
func RunFileServer(targetDir string, port int, log logrus.FieldLogger) error {
	log.Infof("Serving files from '%v' at http://localhost:%v/", targetDir, port)
	handler := http.FileServer(http.Dir(targetDir))
	return http.ListenAndServe(":"+strconv.Itoa(port), handler)
}

type generateRoute struct {
	ContentType string
	handler     ContextHandler
}

func newGenerateRoute(contentType string, handler ContextHandler) *generateRoute {
	return &generateRoute{contentType, handler}
}

// GenerateRouter generates static files, note that ContentType is respected by the router's Response struct,
// by default via calling mime.TypeByExtension (Go std lib) on the route pattern
// or setting it via Context. However, generated files DO NOT have a ContentType,
// as ContentType is a http thing and will be set when files are uploaded to S3.
//
// See the Router interface.
type GenerateRouter struct {
	log     logrus.FieldLogger
	routes  map[string]*generateRoute
	folders map[string]bool

	arounds []AroundHandler
}

// NewGenerateRouter returns a new instance of GenerateRouter
func NewGenerateRouter(log logrus.FieldLogger) *GenerateRouter {
	return &GenerateRouter{
		log,
		make(map[string]*generateRoute),
		make(map[string]bool),
		nil,
	}
}

// Around is a callback/handler that is called around all routes
func (router *GenerateRouter) Around(handler AroundHandler) {
	router.arounds = append(router.arounds, handler)
}

// GetRootHTML defines a HTML handler for the root URL `/`
func (router *GenerateRouter) GetRootHTML(handler ContextHandler) {
	router.checkAndSetHTMLRoutes(RootURL, handler)
}

// GetHTML defines a HTML handler given a URL (shorthand for Get with Content-Type set for .html files)
func (router *GenerateRouter) GetHTML(url string, handler ContextHandler) {
	router.checkAndSetHTMLRoutes(url, handler)
}

// Get define a handler for any file type given a URL
func (router *GenerateRouter) Get(url string, handler ContextHandler) {
	router.checkAndSetRoutes(url, mime.TypeByExtension(path.Ext(url)), handler)
}

func (router *GenerateRouter) checkAndSetHTMLRoutes(url string, handler ContextHandler) {
	router.checkAndSetRoutes(url, mime.TypeByExtension(".html"), handler)
}

func (router *GenerateRouter) hasRoute(url string) bool {
	_, has := router.routes[url]
	return has
}

func (router *GenerateRouter) checkAndSetRoutes(url, contentType string, handler ContextHandler) {
	if router.hasRoute(url) {
		panicDuplicateRoute(url)
	}
	checkAndSetFolders(url, router.folders, router.hasRoute)
	router.routes[url] = newGenerateRoute(contentType, handler)
}

func (router *GenerateRouter) get(url string) (*Response, error) {
	route := router.routes[url]
	if route == nil {
		return nil, fmt.Errorf("url not found: %v", url)
	}

	ctx := NewContext(router.log)
	ctx.url = url
	ctx.contentType = route.ContentType

	err := callArounds(router.arounds, route.handler, ctx)
	if err != nil {
		return nil, err
	}
	return NewResponse(ctx.response, ctx.contentType), nil
}

// URLs returns a list the URLs defined on the router
func (router *GenerateRouter) URLs() []string {
	staticRoutes := make([]string, len(router.routes))
	i := 0
	for k := range router.routes {
		staticRoutes[i] = k
		i++
	}
	return staticRoutes
}

// Requester returns a requester for the given router, to make requests and return the response
func (router *GenerateRouter) Requester() Requester {
	return &GenerateRequester{
		router,
	}
}

// GenerateRequester makes requests on the GenerateRouter
type GenerateRequester struct {
	router *GenerateRouter
}

// Get gets the response of the route's handler given the url
func (requester *GenerateRequester) Get(url string) (*Response, error) {
	if url[:1] != "/" {
		url = "/" + url
	}
	return requester.router.get(url)
}
