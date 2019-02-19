package app

import (
	"github.com/s12chung/gostatic/go/lib/router"
)

// Setter is an interface for App to talk to your code and set routes.
type Setter interface {
	// SetRoutes is where you set the routes
	SetRoutes(r router.Router) error

	// URLBatches returns an array of URL batches (arrays). When generating the static web pages,
	// each URL of each batch is generated concurrently, in the order of the URL batches.
	URLBatches(r router.Router) ([][]string, error)

	// AssetsURL is the URL path prefix of all your assets, so the server can redirect this prefix to your assets
	AssetsURL() string
	// GeneratedAssetsPath is the local file path of the generated assets
	GeneratedAssetsPath() string
}
