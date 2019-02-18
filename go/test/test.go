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
		context := NewContext(t).SetFields(ContextFields{
			"index": testCaseIndex,
			"env":   tc.env,
		})
		context.AssertError(os.Setenv(envKey, tc.env), "os.Setenv")
		context.Assert("Result", callDefaultSettings(), tc.exp)
	}
	AssertError(t, os.Setenv(envKey, ""), "os.Setenv")
}

// RandSeed sets the rand.Seed
func RandSeed() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// TimeDiff finds the time difference between the callack
func TimeDiff(callback func()) time.Duration {
	start := time.Now()
	callback()
	return time.Since(start)
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
	AssertError(t, err, "ioutil.TempDir")
	return filepath.Join(dir, cleanFilePath(originalPath)), func() {
		err := os.RemoveAll(dir)
		AssertError(t, err, "os.RemoveAll")
	}
}

// UpdateFixtureFlag checks for the -update flag for go test to update the fixtures
func UpdateFixtureFlag() *bool {
	return flag.Bool("update", false, "Update fixtures")
}

// ReadFixture reads the fixture given the filename
func ReadFixture(t *testing.T, filename string) []byte {
	bytes, err := ioutil.ReadFile(filepath.Join(FixturePath, filename))
	AssertError(t, err, "ioutil.ReadFile")
	return bytes
}

// WriteFixture writes the fixture given the filename
func WriteFixture(t *testing.T, filename string, data []byte) {
	err := ioutil.WriteFile(filepath.Join(FixturePath, filename), data, 0755)
	AssertError(t, err, "ioutil.WriteFile")
}
