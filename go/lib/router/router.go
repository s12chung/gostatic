/*
	A router for static websites. Provides a GenerateRouter to generate files and a WebRouter,
	a simplified web router, which provide have the same interface.

	Please note that only 1 level of routes are supported: home.com/, home.com/about, home.com/*
	all work, but home.com/projects/about and home.com/projects/* will not work.
*/
package router

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

const WildcardUrlPattern = "*"
const RootUrlPattern = "/"

var IsRootUrlPart = func(urlParts []string) bool { return len(urlParts) == 0 }

type ContextHandler func(ctx Context) error
type AroundHandler func(ctx Context, handler ContextHandler) error

// The Context object provided for every route
type Context interface {
	// Sets the response of the route
	Respond(bytes []byte) error

	// Returns the log of the route
	Log() logrus.FieldLogger
	// Sets the log of the route, can be used with Router.Around to track state
	SetLog(log logrus.FieldLogger)

	// The content type of the response, only respected WebRouter (ContentType is an HTTP thing, which will be set when uploading to S3)
	ContentType() string
	SetContentType(contentType string)

	// Url split cleanly by `/`
	UrlParts() []string
	// The url of request
	Url() string
}

// Interface for all routers.
type Router interface {
	// A callback/handler that is called around all routes
	Around(handler AroundHandler)

	// Define a HTML handler for any undefined route
	GetWildcardHTML(handler ContextHandler)

	// Define a HTML handler for the root
	GetRootHTML(handler ContextHandler)
	// Define a HTML handler given a pattern
	GetHTML(pattern string, handler ContextHandler)
	// Define a handler for any file given a pattern
	Get(pattern string, handler ContextHandler)

	// A list the urls defined set on the router
	StaticUrls() []string
	// Returns a requester, an abstraction for making router requests
	Requester() Requester
}

// Responsive give by all routers
type Response struct {
	Body     []byte
	MimeType string
}

func NewResponse(body []byte, mimeType string) *Response {
	return &Response{body, mimeType}
}

// An abstraction for making router requests
type Requester interface {
	// Calls the route and returns the response given the url
	Get(url string) (*Response, error)
}

func panicDuplicateRoute(route string) {
	panic(fmt.Sprintf("%v is a duplicate route", route))
}

func urlParts(url string) ([]string, error) {
	var parts []string
	for _, part := range strings.Split(url, "/") {
		if part != "" {
			parts = append(parts, part)
		}
	}
	if len(parts) > 1 {
		return nil, fmt.Errorf("currently can't handle more than 1 UrlPart - %v", url)
	}
	return parts, nil
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
