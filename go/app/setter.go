package app

import (
	"time"

	"github.com/sirupsen/logrus"

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

// LogRouteType is the value set for "type" in the logs given by SetDefaultAroundHandlers
const LogRouteType = "routes"

// SetDefaultAroundHandlers adds the default around handlers on the route
func SetDefaultAroundHandlers(r router.Router) {
	r.Around(func(ctx router.Context, handler router.ContextHandler) error {
		ctx.SetLog(ctx.Log().WithFields(logrus.Fields{
			"type": LogRouteType,
			"URL":  ctx.URL(),
		}))
		ctx.Log().Infof("Running route")

		var err error
		start := time.Now()
		defer func() {
			log := ctx.Log().WithField("duration", time.Since(start))

			ending := " for route"
			if err != nil {
				log.Errorf("Error"+ending+" - %v", err)
				return
			}

			log.Infof("Success" + ending)
		}()

		err = handler(ctx)
		return err
	})
}
