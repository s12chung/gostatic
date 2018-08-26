package webpack

import (
	"os"
)

type Settings struct {
	AssetsPath string `json:"assets_path,omitempty"`
}

func DefaultSettings() *Settings {
	assetsPath := os.Getenv("ASSETS_PATH")
	if assetsPath == "" {
		assetsPath = "assets"
	}
	return &Settings{
		assetsPath,
	}
}
