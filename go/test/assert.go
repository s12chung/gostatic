package test

import (
	"fmt"
	"testing"
)

// AssertError checks if there's an error and reports it
func AssertError(t *testing.T, err error, label string) {
	if err != nil {
		t.Error(AssertErrorString(err, label))
	}
}

// AssertErrorString returns a string format for AssertError
func AssertErrorString(err error, label string) string {
	return fmt.Sprintf("error - %v - %v", label, err)
}

// AssertLabel does a simple assertion
func AssertLabel(t *testing.T, label string, got, exp interface{}) {
	if got != exp {
		t.Error(AssertLabelString(label, got, exp))
	}
}

// AssertLabelString returns a string format for assertions
func AssertLabelString(label string, got, exp interface{}) string {
	return fmt.Sprintf("%v - got: %v, exp: %v", label, got, exp)
}

// DiffString returns a string format for diffs
func DiffString(label string, got, exp, diff interface{}) string {
	return fmt.Sprintf("%v, diff: %v", AssertLabelString(label, got, exp), diff)
}
