package webpack

import (
	"os"
)

// Settings is the settings of this package
type Settings struct {
	AssetsPath string `json:"assets_path,omitempty"`
}

// DefaultSettings returns the default settings of this package
func DefaultSettings() *Settings {
	assetsPath := os.Getenv("ASSETS_PATH")
	if assetsPath == "" {
		assetsPath = "assets"
	}
	return &Settings{
		assetsPath,
	}
}
