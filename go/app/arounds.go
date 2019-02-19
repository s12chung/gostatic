package app

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/s12chung/gostatic/go/lib/router"
)

// LogRouteType is the value set for "type" in the logs given by SetDefaultRouterArounds
const LogRouteType = "routes"

// SetDefaultRouterAroundHandlers adds the default around handlers on the Router
func SetDefaultRouterAroundHandlers(r router.Router) {
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

// ContextHandler is the handler for Generate routes
type ContextHandler func() error

// AroundHandler is the handler for Generate callbacks
type AroundHandler func(handler func() error) error

func callArounds(arounds []AroundHandler, handler ContextHandler) error {
	if len(arounds) == 0 {
		return handler()
	}

	aroundToNext := make([]ContextHandler, len(arounds))
	for index := range arounds {
		reverseIndex := len(arounds) - 1 - index
		around := arounds[reverseIndex]
		if index == 0 {
			aroundToNext[reverseIndex] = func() error {
				return around(handler)
			}
		} else {
			aroundToNext[reverseIndex] = func() error {
				return around(aroundToNext[reverseIndex+1])
			}
		}
	}
	return aroundToNext[0]()
}
