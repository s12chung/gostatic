/*
Package app glues your routes together to generate files concurrently or host them in a server.
*/
package app

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/s12chung/gostatic/go/lib/pool"
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
	routeSetter Setter
	settings    *Settings
	log         logrus.FieldLogger
}

// NewApp returns a new instance of App
func NewApp(routeSetter Setter, settings *Settings, log logrus.FieldLogger) *App {
	return &App{
		routeSetter,
		settings,
		log,
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
	r.FileServe(app.routeSetter.AssetsURL(), app.routeSetter.GeneratedAssetsPath())
	app.setRoutes(r)

	return r.Run()
}

// ServerPort returns the port of the web application server
func (app *App) ServerPort() int {
	return app.settings.ServerPort
}

// Generate generates the static web pages concurrently.
//
// For speed and concurrency reasons (like file/map read/writing), this is done in two stages:
// First, the Tracker.IndependentURLs routes are generated. After, the Tracker.DependentURLs.
//
// Use Tracker.AddDependentURL to generate the route's file during the second stage.
func (app *App) Generate() error {
	start := time.Now()
	defer func() {
		app.log.Infof("Build generated in %v.", time.Since(start))
	}()

	err := utils.MkdirAll(app.settings.GeneratedPath)
	if err != nil {
		return err
	}

	r := router.NewGenerateRouter(app.log)
	routeTracker := app.setRoutes(r)
	err = app.requestRoutes(r.Requester(), routeTracker)
	if err != nil {
		return err
	}
	return nil
}

func (app *App) setRoutes(r router.Router) *Tracker {
	r.Around(func(ctx *router.Context, handler router.ContextHandler) error {
		ctx.SetLog(ctx.Log().WithFields(logrus.Fields{
			"type": "routes",
			"URL":  ctx.URL(),
		}))

		var err error

		ctx.Log().Infof("Running route")
		start := time.Now()
		defer func() {
			ending := fmt.Sprintf(" for route")

			log := ctx.Log().WithField("time", time.Since(start))
			if err != nil {
				log.Errorf("Error"+ending+" - %v", err)
			} else {
				log.Infof("Success" + ending)
			}
		}()

		err = handler(ctx)
		return err
	})

	routeTracker := NewTracker(r.Urls)
	app.routeSetter.SetRoutes(r, routeTracker)
	return routeTracker
}

func (app *App) requestRoutes(requester router.Requester, tracker *Tracker) error {
	var urlBatches [][]string

	independentUrls, err := tracker.IndependentURLs()
	if err != nil {
		return err
	}

	urlBatches = append(urlBatches, independentUrls)
	urlBatches = append(urlBatches, tracker.DependentURLs())

	for _, urlBatch := range urlBatches {
		app.runTasks(app.urlsToTasks(requester, urlBatch))
	}
	return nil
}

func (app *App) urlsToTasks(requester router.Requester, urls []string) []*pool.Task {
	tasks := make([]*pool.Task, len(urls))
	for i, url := range urls {
		tasks[i] = app.getURLTask(requester, url)
	}
	return tasks
}

func (app *App) getURLTask(requester router.Requester, url string) *pool.Task {
	log := app.log.WithFields(logrus.Fields{
		"type": "task",
		"url":  url,
	})

	return pool.NewTask(log, func() error {
		response, err := requester.Get(url)
		if err != nil {
			return err
		}

		filename := url
		if url == router.RootURL {
			filename = "index.html"
		}

		generatedFilePath := path.Join(app.settings.GeneratedPath, filename)

		generatedDir := path.Dir(generatedFilePath)
		_, err = os.Stat(generatedDir)
		if os.IsNotExist(err) {
			err = utils.MkdirAll(generatedDir)
			if err != nil {
				return err
			}
		}

		log.Infof("Writing response into %v", generatedFilePath)
		return utils.WriteFile(generatedFilePath, response.Body)
	})
}

func (app *App) runTasks(tasks []*pool.Task) {
	p := pool.NewPool(tasks, app.settings.Concurrency)
	p.Run()
	p.EachError(func(task *pool.Task) {
		task.Log.Errorf("Error for task - %v", task.Error)
	})
}
