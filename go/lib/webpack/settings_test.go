package webpack

import (
	"testing"

	"github.com/s12chung/gostatic/go/test/settings"
)

func TestDefaultSettings(t *testing.T) {
	settings.EnvSetting(t, "ASSETS_PATH", "assets", func() string {
		return DefaultSettings().AssetsPath
	})
}
