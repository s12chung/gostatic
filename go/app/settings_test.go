package app

import (
	"testing"

	"github.com/s12chung/gostatic/go/test"
)

func TestDefaultSettings(t *testing.T) {
	test.TestEnvSetting(t, "GENERATED_PATH", "./generated", func() string {
		return DefaultSettings().GeneratedPath
	})
}
