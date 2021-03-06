package main

import (
	"os"

	"github.com/s12chung/gostatic/blueprint/go/content"

	"github.com/s12chung/gostatic/go/app"
	"github.com/s12chung/gostatic/go/cli"
	"github.com/s12chung/gostatic/go/lib/utils"
)

func main() {
	log := app.DefaultLog()

	settings := app.DefaultSettings()
	contentSettings := content.DefaultSettings()
	settings.Content = contentSettings
	utils.SettingsFromFile("./settings.json", settings, log)

	theContent := content.NewContent(settings.GeneratedPath, contentSettings, log)
	err := cli.Run(app.NewApp(theContent, settings, log))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
