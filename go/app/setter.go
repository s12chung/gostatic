package app

import "github.com/s12chung/gostatic/go/lib/router"

// An interface for App to talk to your code and set routes.
type Setter interface {
	// SetRoutes is where you set the routes
	SetRoutes(r router.Router, tracker *Tracker)
	// AssetsURL is the URL path prefix of all your assets, so the server can redirect this prefix to your assets
	AssetsURL() string
	// GeneratedAssetsPath is the local file path of the generated assets
	GeneratedAssetsPath() string
}
