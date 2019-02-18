/*
Package settings is a set of helpers for gostatic settings tests
*/
package settings

import (
	"os"
	"testing"

	"github.com/s12chung/gostatic/go/test"
)

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
		context := test.NewContext(t).SetFields(test.ContextFields{
			"index": testCaseIndex,
			"env":   tc.env,
		})
		context.AssertError(os.Setenv(envKey, tc.env), "os.Setenv")

		err := callDefaultSettings()
		context.Assert("Result", err, tc.exp)
	}
	test.AssertError(t, os.Setenv(envKey, ""), "os.Setenv")
}
