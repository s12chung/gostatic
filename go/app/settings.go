package app

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
)

// Settings represents the settings of App
type Settings struct {
	GeneratedPath     string             `json:"generated_path,omitempty"`
	ServerPort        int                `json:"server_port,omitempty"`
	FileServerPort    int                `json:"file_server_port,omitempty"`
	GeneratorSettings *GeneratorSettings `json:"generator_settings,omitempty"`

	Content interface{} `json:"content,omitempty"`
}

// GeneratorSettings represents the settings for Generating files
type GeneratorSettings struct {
	Concurrency int `json:"concurrency,omitempty"`
}

// DefaultSettings returns the default settings of the App
func DefaultSettings() *Settings {
	generatedPath := os.Getenv("GENERATED_PATH")
	if generatedPath == "" {
		generatedPath = "./generated"
	}
	return &Settings{
		generatedPath,
		8080,
		3000,
		&GeneratorSettings{
			10,
		},
		nil,
	}
}

// SettingsFromFile loads settings from the given file path into the given Settings
func SettingsFromFile(path string, settings interface{}, log logrus.FieldLogger) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		log.Warnf("%v not found, using defaults...", path)
		return
	}

	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Warnf("error reading %v, using defaults...", path)
		return
	}

	err = json.Unmarshal(file, settings)
	if err != nil {
		log.Warnf("error reading %v, using defaults...", path)
		return
	}
}
