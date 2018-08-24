package content

import (

	"github.com/sirupsen/logrus"

	"github.com/s12chung/gostatic/go/app"
	"github.com/s12chung/gostatic/go/lib/router"
	"github.com/s12chung/gostatic/go/lib/webpack"
	"github.com/s12chung/gostatic/go/lib/html"
)

type Content struct {
	Settings *Settings
	Log      logrus.FieldLogger

	HtmlRenderer      *html.Renderer
	Webpack  *webpack.Webpack
}

func NewContent(generatedPath string, settings *Settings, log logrus.FieldLogger) *Content {
	w := webpack.NewWebpack(generatedPath, settings.Webpack, log)
	htmlRenderer := html.NewRenderer(settings.Html, []html.Plugin{ w }, log)
	return &Content{ settings, log, htmlRenderer, w }
}

func (content *Content) RenderHtml(ctx router.Context, name, defaultTitle string, data interface{}) error {
	bytes, err := content.HtmlRenderer.Render(name, defaultTitle, data)
	if err != nil {
		return err
	}
	return ctx.Respond(bytes)
}

func (content *Content) SetRoutes(r router.Router, tracker *app.Tracker) {
	r.GetRootHTML(func(ctx router.Context) error {
		return content.RenderHtml(ctx, "root", "", "Hello World!")
	})
	r.GetHTML("/404.html", func(ctx router.Context) error {
		return content.RenderHtml(ctx, "404", "404", nil)
	})
}

func (content *Content) WildcardUrls() ([]string, error) {
	return []string{}, nil
}

func (content *Content) AssetsUrl() string {
	return content.Webpack.AssetsUrl()
}

func (content *Content) GeneratedAssetsPath() string {
	return content.Webpack.GeneratedAssetsPath()
}
