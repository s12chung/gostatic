/*
Package testfile is a set of helpers for tests related to files
*/
package testfile

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/s12chung/gostatic/go/test"
)

// FixturePath is the path of the fixtures
const FixturePath = "./testdata"

func cleanFilePath(filePath string) string {
	filePath = strings.TrimLeft(filePath, ".")
	return strings.Trim(filePath, "/")
}

// SandboxDir sets up a TempDir for a sandbox
func SandboxDir(t *testing.T, originalPath string) (string, func()) {
	dir, err := ioutil.TempDir("", "sandbox")
	test.AssertError(t, err, "ioutil.TempDir")
	return filepath.Join(dir, cleanFilePath(originalPath)), func() {
		err := os.RemoveAll(dir)
		test.AssertError(t, err, "os.RemoveAll")
	}
}

// UpdateFixtureFlag checks for the -update flag for go test to update the fixtures
func UpdateFixtureFlag() *bool {
	return flag.Bool("update", false, "Update fixtures")
}

// ReadFixture reads the fixture given the filename
func ReadFixture(t *testing.T, filename string) []byte {
	bytes, err := ioutil.ReadFile(filepath.Join(FixturePath, filename))
	test.AssertError(t, err, "ioutil.ReadFile")
	return bytes
}

// WriteFixture writes the fixture given the filename
func WriteFixture(t *testing.T, filename string, data []byte) {
	err := ioutil.WriteFile(filepath.Join(FixturePath, filename), data, 0755)
	test.AssertError(t, err, "ioutil.WriteFile")
}
