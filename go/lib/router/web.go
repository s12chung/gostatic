package router

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type webHandler func(w http.ResponseWriter, r *http.Request) error

// See the Context interface.
type WebContext struct {
	log         logrus.FieldLogger
	contentType string

	url string

	responseWriter http.ResponseWriter
	request        *http.Request
}

func NewWebContext(log logrus.FieldLogger) *WebContext {
	return &WebContext{log: log}
}

func (ctx *WebContext) Log() logrus.FieldLogger {
	return ctx.log
}

func (ctx *WebContext) SetLog(log logrus.FieldLogger) {
	ctx.log = log
}

func (ctx *WebContext) ContentType() string {
	return ctx.contentType
}

func (ctx *WebContext) SetContentType(contentType string) {
	ctx.contentType = contentType
}

func (ctx *WebContext) Url() string {
	return ctx.url
}

func (ctx *WebContext) Respond(bytes []byte) error {
	_, err := ctx.responseWriter.Write(bytes)
	return err
}

// The router to host a web application server. It's simplified such that all errors
// are give http.StatusBadRequest and print out the error.
//
// Content-Type is respected by default via calling mime.TypeByExtension (Go std lib) on the route pattern
// or setting it via Context.
//
// See the Router interface.
type WebRouter struct {
	serveMux *http.ServeMux
	log      logrus.FieldLogger

	arounds []AroundHandler
	routes  map[string]bool

	rootHandler     http.HandlerFunc
	wildcardHandler http.HandlerFunc

	port int
}

func NewWebRouter(port int, log logrus.FieldLogger) *WebRouter {
	defaultHandler := func(w http.ResponseWriter, r *http.Request) {
		s := fmt.Sprintf("%v is not being handled", r.URL)
		log.Errorf(s)
		http.Error(w, s, http.StatusBadRequest)
		return
	}

	router := &WebRouter{
		http.NewServeMux(),
		log,
		nil,
		make(map[string]bool),
		defaultHandler,
		defaultHandler,
		port,
	}
	router.handleWildcard()
	return router
}

func (router *WebRouter) Around(handler AroundHandler) {
	router.arounds = append(router.arounds, handler)
}

func (router *WebRouter) GetRootHTML(handler ContextHandler) {
	router.checkAndSetRoutes(RootUrlPattern)
	router.rootHandler = router.getRequestHandler(router.htmlHandler(handler))
}

func (router *WebRouter) GetHTML(pattern string, handler ContextHandler) {
	router.checkAndSetRoutes(pattern)
	router.get(pattern, router.htmlHandler(handler))
}

func (router *WebRouter) Get(pattern string, handler ContextHandler) {
	router.checkAndSetRoutes(pattern)
	router.get(pattern, router.handler(mime.TypeByExtension(path.Ext(pattern)), handler))
}

func (router *WebRouter) checkAndSetRoutes(pattern string) error {
	_, has := router.routes[pattern]
	if has {
		panicDuplicateRoute(pattern)
	}
	router.routes[pattern] = true
	return nil
}

func (router *WebRouter) Urls() []string {
	staticRoutes := make([]string, len(router.routes))
	i := 0
	for k := range router.routes {
		staticRoutes[i] = k
		i++
	}
	return staticRoutes
}

func (router *WebRouter) Requester() Requester {
	return newWebRequester(router.port)
}

func (router *WebRouter) htmlHandler(handler ContextHandler) webHandler {
	return router.handler(mime.TypeByExtension(".html"), handler)
}

func (router *WebRouter) handler(contentType string, handler ContextHandler) webHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := NewWebContext(router.log)
		ctx.contentType = contentType
		ctx.url = r.URL.String()
		ctx.responseWriter = w
		ctx.request = r

		err := callArounds(router.arounds, handler, ctx)
		w.Header().Set("Content-Type", ctx.contentType)

		return err
	}
}

func (router *WebRouter) getRequestHandler(handler webHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			err := handler(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		}
	}
}

func (router *WebRouter) handleWildcard() {
	router.serveMux.HandleFunc(RootUrlPattern, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/" {
			router.rootHandler(w, r)
		} else {
			router.wildcardHandler(w, r)
		}
	})
}

// FileServe sets the router to redirect requests with a pattern to a file directory.
// Content-Type is respected via calling mime.TypeByExtension (Go std lib).
func (router *WebRouter) FileServe(pattern, dirPath string) {
	router.get(pattern, func(w http.ResponseWriter, r *http.Request) error {
		regex := regexp.MustCompile(strings.Replace(`^/`+pattern+`/`, "//", "/", -1))
		assetFilePath := path.Join(dirPath, regex.ReplaceAllString(r.URL.String(), ""))

		file, err := os.Open(assetFilePath)
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(assetFilePath)))
		_, err = io.Copy(w, file)
		return err
	})
}

func (router *WebRouter) get(pattern string, handler webHandler) {
	if pattern == RootUrlPattern {
		router.log.Errorf("Can not use pattern that touches root, use GetRootHTML or GetWildcardHTML instead")
		return
	}

	router.serveMux.HandleFunc(pattern, router.getRequestHandler(handler))
}

func (router *WebRouter) Run() error {
	router.log.Infof("Running server at http://localhost:%v/", router.port)
	server := &http.Server{Addr: ":" + strconv.Itoa(router.port), Handler: router.serveMux}
	return server.ListenAndServe()
}

// Object to make requests on the router
type WebRequester struct {
	hostname string
	port     int
}

func newWebRequester(port int) *WebRequester {
	return &WebRequester{
		"localhost",
		port,
	}
}

func (requester *WebRequester) Get(url string) (*Response, error) {
	response, err := http.Get(fmt.Sprintf("http://%v:%v%v", requester.hostname, requester.port, url))
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf(strings.TrimSpace(string(body)))
	}
	return NewResponse(body, response.Header.Get("Content-Type")), nil
}
