package content

import (
	"github.com/s12chung/gostatic/go/lib/webpack"
	"github.com/s12chung/gostatic/go/lib/html"
)

type Settings struct {
	Html      *html.Settings      `json:"html,omitempty"`
	Webpack   *webpack.Settings   `json:"webpack,omitempty"`
}

func DefaultSettings() *Settings {
	return &Settings{
		html.DefaultSettings(),
		webpack.DefaultSettings(),
	}
}