package webpack

import (
	"testing"

	"github.com/s12chung/gostatic/go/test"
)

func TestDefaultSettings(t *testing.T) {
	test.EnvSetting(t, "ASSETS_PATH", "assets", func() string {
		return DefaultSettings().AssetsPath
	})
}
