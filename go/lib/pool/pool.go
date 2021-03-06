/*
Package pool runs your tasks concurrently.
*/
package pool

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// Task that runs and stores any errors.
type Task struct {
	run   func() error
	Log   logrus.FieldLogger
	Error error
}

// NewTask returns a new instance of Task
func NewTask(log logrus.FieldLogger, run func() error) *Task {
	return &Task{run, log, nil}
}

// Run runs function of the task
func (task *Task) Run(waitGroup *sync.WaitGroup) {
	task.Error = task.run()
	waitGroup.Done()
}

// Pool represents a set of tasks.
type Pool struct {
	Tasks []*Task

	concurrency int
	tasksChan   chan *Task
	waitGroup   sync.WaitGroup
}

// NewPool returns a new instance of pool
func NewPool(tasks []*Task, concurrency int) *Pool {
	return &Pool{
		Tasks:       tasks,
		concurrency: concurrency,
		tasksChan:   make(chan *Task),
	}
}

// EachError loops through all the Pool's Tasks' errors (after they run)
func (pool *Pool) EachError(callback func(*Task)) {
	for _, task := range pool.Tasks {
		if task.Error != nil {
			callback(task)
		}
	}
}

// Run runs all the Tasks of the pool with the given concurrency
func (pool *Pool) Run() {
	for i := 0; i < pool.concurrency; i++ {
		go pool.work()
	}

	pool.waitGroup.Add(len(pool.Tasks))
	for _, task := range pool.Tasks {
		pool.tasksChan <- task
	}
	close(pool.tasksChan)

	pool.waitGroup.Wait()
}

func (pool *Pool) work() {
	for task := range pool.tasksChan {
		task.Run(&pool.waitGroup)
	}
}
