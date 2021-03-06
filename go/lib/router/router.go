/*
Package router is a router for static websites. Provides a GenerateRouter to generate files and a WebRouter,
a simplified web router, which have the same interface.
*/
package router

import (
	"fmt"
	"path"

	"github.com/sirupsen/logrus"
)

// RootURL is the URL of the Root of the router
const RootURL = "/"

// ContextHandler is the handler for Router routes
type ContextHandler func(ctx Context) error

// AroundHandler is the handler for Router callbacks
type AroundHandler func(ctx Context, handler ContextHandler) error

// Context is an interface of the context provided for every route,
// it's an interface to make testing easier
type Context interface {
	// Log returns the log of the context
	Log() logrus.FieldLogger
	// SetLog sets the log of the Context, so that you can set the context of the log
	SetLog(log logrus.FieldLogger)

	// ContentType returns the Content-Type to be sent to the response
	ContentType() string
	// SetContentType sets the Content-Type of the response
	SetContentType(contentType string)

	// URL returns the URL of the request
	URL() string

	// Respond sets the response data of the request
	Respond(bytes []byte)
}

// context provided for every route
type context struct {
	log         logrus.FieldLogger
	contentType string

	url      string
	response []byte
}

// newContext returns a new instance of Context
func newContext(log logrus.FieldLogger) *context {
	return &context{log: log}
}

// Log returns the log of the context
func (ctx *context) Log() logrus.FieldLogger {
	return ctx.log
}

// SetLog sets the log of the Context, so that you can set the context of the log
func (ctx *context) SetLog(log logrus.FieldLogger) {
	ctx.log = log
}

// ContentType returns the Content-Type to be sent to the response
func (ctx *context) ContentType() string {
	return ctx.contentType
}

// SetContentType sets the Content-Type of the response
func (ctx *context) SetContentType(contentType string) {
	ctx.contentType = contentType
}

// URL returns the URL of the request
func (ctx *context) URL() string {
	return ctx.url
}

// Respond sets the response data of the request
func (ctx *context) Respond(bytes []byte) {
	ctx.response = bytes
}

// Router is the interface for all routers.
type Router interface {
	// Around is a callback/handler that is called around all routes
	Around(handler AroundHandler)

	// GetRootHTML defines a HTML handler for the root URL `/`
	GetRootHTML(handler ContextHandler)
	// GetHTML defines a HTML handler given a url (shorthand for Get with Content-Type set for .html files)
	GetHTML(url string, handler ContextHandler)
	// Get define a handler for any file type given a url
	Get(url string, handler ContextHandler)

	// URLs returns a list the URLs defined on the router
	URLs() []string
	// Requester returns a requester for the given router, to make requests and return the response
	Requester() Requester
}

// Response given by all routers
type Response struct {
	Body     []byte
	MimeType string
}

// NewResponse returns a new instance of Response
func NewResponse(body []byte, mimeType string) *Response {
	return &Response{body, mimeType}
}

// Requester is an abstraction for making router requests
type Requester interface {
	// Calls the route and returns the response given the url
	Get(url string) (*Response, error)
}

func handleURLSlash(url string) string {
	if len(url) == 0 || url[:1] != "/" {
		url = "/" + url
	}
	return url
}

func panicDuplicateRoute(url string) {
	panic(fmt.Sprintf("%v is a duplicate route", url))
}

func checkAndSetFolders(url string, folders map[string]bool, hasRoute func(url string) bool) {
	_, has := folders[url]
	if has {
		panic(fmt.Sprintf("%v is a route, which is a folder of another route", url))
	}

	dir := path.Dir(url)
	if dir == RootURL {
		return
	}

	if hasRoute(dir) {
		panic(fmt.Sprintf("%v the folder of this URL is another URL", url))
	}
	folders[dir] = true
}

func callArounds(arounds []AroundHandler, handler ContextHandler, ctx Context) error {
	if len(arounds) == 0 {
		return handler(ctx)
	}

	aroundToNext := make([]ContextHandler, len(arounds))
	for index := range arounds {
		reverseIndex := len(arounds) - 1 - index
		around := arounds[reverseIndex]
		if index == 0 {
			aroundToNext[reverseIndex] = func(ctx Context) error {
				return around(ctx, handler)
			}
		} else {
			aroundToNext[reverseIndex] = func(ctx Context) error {
				return around(ctx, aroundToNext[reverseIndex+1])
			}
		}
	}
	return aroundToNext[0](ctx)
}
