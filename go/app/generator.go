package app

import (
	"os"
	"path"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/s12chung/gostatic/go/lib/pool"
	"github.com/s12chung/gostatic/go/lib/router"
	"github.com/s12chung/gostatic/go/lib/utils"
)

type generator struct {
	generatedPath string
	requester     router.Requester
	settings      *GeneratorSettings
	log           logrus.FieldLogger

	dirs      map[string]bool
	dirsMutex *sync.RWMutex
}

func newGenerator(generatedPath string, requester router.Requester, settings *GeneratorSettings, log logrus.FieldLogger) *generator {
	return &generator{
		generatedPath,
		requester,
		settings,
		log,
		map[string]bool{},
		&sync.RWMutex{},
	}
}

func (gen *generator) generate(urls []string) {
	tasks := gen.urlsToTasks(urls)
	gen.runTasks(tasks)
}

func (gen *generator) urlsToTasks(urls []string) []*pool.Task {
	tasks := make([]*pool.Task, len(urls))
	for i, url := range urls {
		tasks[i] = gen.getURLTask(url)
	}
	return tasks
}

func (gen *generator) getURLTask(url string) *pool.Task {
	log := gen.log.WithFields(logrus.Fields{
		"type": "task",
		"url":  url,
	})

	return pool.NewTask(log, func() error {
		response, err := gen.requester.Get(url)
		if err != nil {
			return err
		}

		filename := url
		if url == router.RootURL {
			filename = "index.html"
		}

		generatedFilePath := path.Join(gen.generatedPath, filename)

		generatedDir := path.Dir(generatedFilePath)

		gen.dirsMutex.RLock()
		_, has := gen.dirs[generatedDir]
		gen.dirsMutex.RUnlock()

		if !has {
			_, err = os.Stat(generatedDir)
			if os.IsNotExist(err) {
				err = utils.MkdirAll(generatedDir)
				if err != nil {
					return err
				}
			}
			gen.dirsMutex.Lock()
			gen.dirs[generatedDir] = true
			gen.dirsMutex.Unlock()
		}

		log.Infof("Writing response into %v", generatedFilePath)
		return utils.WriteFile(generatedFilePath, response.Body)
	})
}

func (gen *generator) runTasks(tasks []*pool.Task) {
	p := pool.NewPool(tasks, gen.settings.Concurrency)
	p.Run()
	p.EachError(func(task *pool.Task) {
		task.Log.Errorf("Error for task - %v", task.Error)
	})
}
