/*
Package test is a set of helper functions of tests
*/
package test

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// FixturePath is the path of the fixtures
const FixturePath = "./testdata"

// AssertInput calls AssertLabel, but sets the context based on input
func AssertInput(t *testing.T, input, got, exp interface{}) {
	context := NewContext().SetFields(ContextFields{
		"input": input,
	})
	if got != exp {
		t.Error(context.GotExpString("Result", got, exp))
	}
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

// EnvSetting is a standard test for environment variables's interaction with DefaultSettings()
func EnvSetting(t *testing.T, envKey, defaultValue string, callDefaultSettings func() string) {
	testCases := []struct {
		env string
		exp string
	}{
		{"", defaultValue},
		{"test", "test"},
	}

	for testCaseIndex, tc := range testCases {
		context := NewContext().SetFields(ContextFields{
			"index": testCaseIndex,
			"env":   tc.env,
		})

		err := os.Setenv(envKey, tc.env)
		if err != nil {
			t.Error(err)
		}
		got := callDefaultSettings()
		if got != tc.exp {
			t.Error(context.GotExpString("Result", got, tc.exp))
		}
	}
	err := os.Setenv(envKey, "")
	if err != nil {
		t.Error(err)
	}
}

// RandSeed sets the rand.Seed
func RandSeed() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// TimeDiff finds the time difference between the callack
func TimeDiff(callback func()) time.Duration {
	start := time.Now()
	callback()
	return time.Now().Sub(start)
}

// Time returns a standard time to test with
func Time(i int) time.Time {
	return time.Date(2018, 1, i, i, i, i, i, time.UTC)
}

func cleanFilePath(filePath string) string {
	filePath = strings.TrimLeft(filePath, ".")
	return strings.Trim(filePath, "/")
}

// SandboxDir sets up a TempDir for a sandbox
func SandboxDir(t *testing.T, originalPath string) (string, func()) {
	dir, err := ioutil.TempDir("", "sandbox")
	if err != nil {
		t.Error(err)
	}
	return filepath.Join(dir, cleanFilePath(originalPath)), func() {
		err := os.RemoveAll(dir)
		if err != nil {
			t.Error(err)
		}
	}
}

// UpdateFixtureFlag checks for the -update flag for go test to update the fixtures
func UpdateFixtureFlag() *bool {
	return flag.Bool("update", false, "Update fixtures")
}

// ReadFixture reads the fixture given the filename
func ReadFixture(t *testing.T, filename string) []byte {
	bytes, err := ioutil.ReadFile(filepath.Join(FixturePath, filename)) // #nosec G304
	if err != nil {
		t.Error(err)
	}
	return bytes
}

// WriteFixture writes the fixture given the filename
func WriteFixture(t *testing.T, filename string, data []byte) {
	err := ioutil.WriteFile(filepath.Join(FixturePath, filename), data, 0755)
	if err != nil {
		t.Error(err)
	}
}
