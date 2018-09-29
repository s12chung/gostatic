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

// WebRouter is the router to host a web application server. It's simplified such that all errors
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

	rootHandler http.HandlerFunc
	port        int
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
		port,
	}
	router.serveMux.HandleFunc(RootURLPattern, func(w http.ResponseWriter, r *http.Request) {
		router.rootHandler(w, r)
	})
	return router
}

func (router *WebRouter) Around(handler AroundHandler) {
	router.arounds = append(router.arounds, handler)
}

func (router *WebRouter) GetRootHTML(handler ContextHandler) {
	router.checkAndSetRoutes(RootURLPattern)
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

func (router *WebRouter) checkAndSetRoutes(pattern string) {
	_, has := router.routes[pattern]
	if has {
		panicDuplicateRoute(pattern)
	}
	router.routes[pattern] = true
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
		ctx := NewContext(router.log)
		ctx.contentType = contentType
		ctx.url = r.URL.String()

		err := callArounds(router.arounds, handler, ctx)
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", ctx.contentType)
		_, err = w.Write(ctx.response)
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

var dangerPathRegex = regexp.MustCompile(`\/\.\.\/`)

// FileServe sets the router to redirect requests with a pattern to a file directory.
// Content-Type is respected via calling mime.TypeByExtension (Go std lib).
func (router *WebRouter) FileServe(pattern, dirPath string) {
	router.get(pattern, func(w http.ResponseWriter, r *http.Request) error {
		url := r.URL.String()
		if dangerPathRegex.MatchString(url) {
			return fmt.Errorf("url has dangerous characters for security reasons (%v)", url)
		}

		regex := regexp.MustCompile(strings.Replace(`^/`+pattern+`/`, "//", "/", -1))
		assetFilePath := path.Join(dirPath, regex.ReplaceAllString(url, ""))

		file, err := os.Open(filepath.Clean(assetFilePath))
		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(assetFilePath)))
		_, err = io.Copy(w, file)
		return err
	})
}

func (router *WebRouter) get(pattern string, handler webHandler) {
	if pattern == RootURLPattern {
		router.log.Errorf("Can not use pattern that touches root, use GetRootHTML instead")
		return
	}

	router.serveMux.HandleFunc(pattern, router.getRequestHandler(handler))
}

func (router *WebRouter) Run() error {
	router.log.Infof("Running server at http://localhost:%v/", router.port)
	server := &http.Server{Addr: ":" + strconv.Itoa(router.port), Handler: router.serveMux}
	return server.ListenAndServe()
}

// WebRequester makes requests on the WebRouter
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
	if url[:1] != "/" {
		url = "/" + url
	}

	response, err := http.Get(fmt.Sprintf("http://%v:%v%v", requester.hostname, requester.port, url))
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf(strings.TrimSpace(string(body)))
	}
	return NewResponse(body, response.Header.Get("Content-Type")), nil
}
