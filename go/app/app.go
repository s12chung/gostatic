/*
Package app glues your routes together to generate files concurrently or host them in a server.
*/
package app

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/s12chung/gostatic/go/lib/router"
	"github.com/s12chung/gostatic/go/lib/utils"
)

// DefaultLog returns the default log used for the App
func DefaultLog() logrus.FieldLogger {
	return &logrus.Logger{
		Out: os.Stderr,
		Formatter: &logrus.TextFormatter{
			ForceColors: true,
		},
		Hooks: make(logrus.LevelHooks),
		Level: logrus.InfoLevel,
	}
}

// App is a wrapper around the router, to provide the functionality of the cli.App interface.
// Provides defaults to give the app structure and connects things together to generate
// route responses concurrently.
//
// See cli.App interface for core functions.
type App struct {
	Setter
	settings *Settings
	log      logrus.FieldLogger
	arounds  []AroundHandler
}

// NewApp returns a new instance of App
func NewApp(setter Setter, settings *Settings, log logrus.FieldLogger) *App {
	return &App{
		setter,
		settings,
		log,
		nil,
	}
}

// RunFileServer runs the server to host the generated files of the static web page
func (app *App) RunFileServer() error {
	return router.RunFileServer(app.settings.GeneratedPath, app.settings.FileServerPort, app.log)
}

// FileServerPort returns the port of the file server
func (app *App) FileServerPort() int {
	return app.settings.FileServerPort
}

// GeneratedPath returns the path of the generates files of the static web page
func (app *App) GeneratedPath() string {
	return app.settings.GeneratedPath
}

// Host runs a web application server that computes the route responses in real time
func (app *App) Host() error {
	r := router.NewWebRouter(app.settings.ServerPort, app.log)
	r.FileServe(app.AssetsURL(), app.GeneratedAssetsPath())

	if err := app.SetRoutes(r); err != nil {
		return err
	}
	return r.Run()
}

// ServerPort returns the port of the web application server
func (app *App) ServerPort() int {
	return app.settings.ServerPort
}

// Generate generates the static web pages concurrently.
//
// Generated concurrently in the batches, in the order given by Setter.URLBatches()
func (app *App) Generate() error {
	return callArounds(app.arounds, func() error {
		if err := utils.MkdirAll(app.settings.GeneratedPath); err != nil {
			return err
		}

		r := router.NewGenerateRouter(app.log)
		if err := app.SetRoutes(r); err != nil {
			return err
		}
		return app.requestRoutes(r)
	})
}

// Around is a callback/handler that is called around Generate
func (app *App) Around(handler func(handler func() error) error) {
	app.arounds = append(app.arounds, handler)
}

// Log returns the log of the App
func (app *App) Log() logrus.FieldLogger {
	return app.log
}

func (app *App) requestRoutes(r router.Router) error {
	urlBatches, err := app.URLBatches(r)
	if err != nil {
		return err
	}

	generator := newGenerator(app.settings.GeneratedPath, r.Requester(), app.settings.GeneratorSettings, app.log)
	for _, urlBatch := range urlBatches {
		generator.generate(urlBatch)
	}
	return nil
}
