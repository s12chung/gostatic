package content

import (
	"github.com/s12chung/gostatic/go/lib/html"
	"github.com/s12chung/gostatic/go/lib/webpack"
)

// Settings contains the settings of your site, each field is for each package used, you can add additional fields for more features,
// such as packages in github.com/s12chung/gostatic-packages.
//
// The settings are read from a JSON file in main.go.
type Settings struct {
	HTML    *html.Settings    `json:"html,omitempty"`
	Webpack *webpack.Settings `json:"webpack,omitempty"`
}

// DefaultSettings is the default settings of your App, when JSON data is not given
func DefaultSettings() *Settings {
	return &Settings{
		html.DefaultSettings(),
		webpack.DefaultSettings(),
	}
}
