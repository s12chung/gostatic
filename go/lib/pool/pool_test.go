package pool

import (
	"fmt"
	"testing"

	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/gostatic/go/test"
)

func TestPool(t *testing.T) {
	testCases := []struct {
		tasksWithSuccess []bool
	}{
		{nil},
		{[]bool{}},
		{[]bool{true}},
		{[]bool{true, true}},
		{[]bool{true, true, true}},
		{[]bool{true, false, true}},
		{[]bool{false}},
		{[]bool{false, false, false}},
		{[]bool{false, true, true}},
	}

	log, _ := logTest.NewNullLogger()
	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":            testCaseIndex,
			"tasksWithSuccess": tc.tasksWithSuccess,
		})

		var errorTasks []*Task
		runCount := 0
		tasks := tasksWithSuccessToTasks(tc.tasksWithSuccess, func(err error) *Task {
			task := NewTask(log, func() error {
				runCount++
				return err
			})

			if err != nil {
				errorTasks = append(errorTasks, task)
			}
			return task
		})
		p := NewPool(tasks, 10)
		p.Run()

		if runCount != len(tc.tasksWithSuccess) {
			t.Error(context.GotExpString("runCount", runCount, len(tc.tasksWithSuccess)))
		}

		errorCount := 0
		p.EachError(func(task *Task) {
			errorCount++
			if task.Error == nil {
				t.Error("EachError found task without error")
			}
			if !isErrorTask(errorTasks, task) {
				t.Error("EachError Error not found in errorTasks")
			}
		})
		if errorCount != len(errorTasks) {
			t.Error("errorCount does not match number of errorTasks")
		}
	}
}

func tasksWithSuccessToTasks(tasksWithSuccess []bool, makeTask func(err error) *Task) []*Task {
	tasks := make([]*Task, len(tasksWithSuccess))
	if tasksWithSuccess == nil {
		tasks = nil
	} else {
		for i, success := range tasksWithSuccess {
			ret := fmt.Errorf("error")
			if success {
				ret = nil
			}
			tasks[i] = makeTask(ret)
		}
	}
	return tasks
}

func isErrorTask(errorTasks []*Task, task *Task) bool {
	for _, errorTask := range errorTasks {
		if errorTask == task {
			return true
		}
	}
	return false
}
