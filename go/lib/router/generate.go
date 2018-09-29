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
	log    logrus.FieldLogger
	routes map[string]*generateRoute

	arounds []AroundHandler
}

func NewGenerateRouter(log logrus.FieldLogger) *GenerateRouter {
	return &GenerateRouter{
		log,
		make(map[string]*generateRoute),
		nil,
	}
}

func (router *GenerateRouter) Around(handler AroundHandler) {
	router.arounds = append(router.arounds, handler)
}

func (router *GenerateRouter) GetRootHTML(handler ContextHandler) {
	router.checkAndSetHTMLRoutes(RootURLPattern, handler)
}

func (router *GenerateRouter) GetHTML(pattern string, handler ContextHandler) {
	router.checkAndSetHTMLRoutes(pattern, handler)
}

func (router *GenerateRouter) Get(pattern string, handler ContextHandler) {
	router.checkAndSetRoutes(pattern, mime.TypeByExtension(path.Ext(pattern)), handler)
}

func (router *GenerateRouter) checkAndSetHTMLRoutes(pattern string, handler ContextHandler) {
	router.checkAndSetRoutes(pattern, mime.TypeByExtension(".html"), handler)
}

func (router *GenerateRouter) checkAndSetRoutes(pattern, contentType string, handler ContextHandler) {
	_, has := router.routes[pattern]
	if has {
		panicDuplicateRoute(pattern)
	}
	router.routes[pattern] = newGenerateRoute(contentType, handler)
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

func (router *GenerateRouter) Urls() []string {
	staticRoutes := make([]string, len(router.routes))
	i := 0
	for k := range router.routes {
		staticRoutes[i] = k
		i++
	}
	return staticRoutes
}

func (router *GenerateRouter) Requester() Requester {
	return &GenerateRequester{
		router,
	}
}

// GenerateRequester makes requests on the GenerateRouter
type GenerateRequester struct {
	router *GenerateRouter
}

func (requester *GenerateRequester) Get(url string) (*Response, error) {
	if url[:1] != "/" {
		url = "/" + url
	}
	return requester.router.get(url)
}
