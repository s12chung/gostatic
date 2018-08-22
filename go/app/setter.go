package app

import "github.com/s12chung/gostatic/go/lib/router"

type Setter interface {
	SetRoutes(r router.Router, tracker *Tracker)
	WildcardUrls() ([]string, error)
	AssetsUrl() string
	GeneratedAssetsPath() string
}
