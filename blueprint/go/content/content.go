package content

import (
	"github.com/sirupsen/logrus"

	"github.com/s12chung/gostatic/go/app"
	"github.com/s12chung/gostatic/go/lib/html"
	"github.com/s12chung/gostatic/go/lib/router"
	"github.com/s12chung/gostatic/go/lib/webpack"
)

type Content struct {
	Settings *Settings
	Log      logrus.FieldLogger

	HtmlRenderer *html.Renderer
	Webpack      *webpack.Webpack
}

func NewContent(generatedPath string, settings *Settings, log logrus.FieldLogger) *Content {
	w := webpack.NewWebpack(generatedPath, settings.Webpack, log)
	htmlRenderer := html.NewRenderer(settings.Html, []html.Plugin{w}, log)
	return &Content{settings, log, htmlRenderer, w}
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

func (content *Content) RenderHtml(ctx router.Context, name, defaultTitle string, data interface{}) error {
	bytes, err := content.HtmlRenderer.Render(name, defaultTitle, data)
	if err != nil {
		return err
	}
	return ctx.Respond(bytes)
}

func (content *Content) SetRoutes(r router.Router, tracker *app.Tracker) {
	r.GetRootHTML(content.getRoot)
	r.GetHTML("/404.html", content.get404)
	r.GetHTML("/robots.txt", content.getRobots)
}

func (content *Content) getRoot(ctx router.Context) error {
	return content.RenderHtml(ctx, "root", "", "Hello World!")
}

func (content *Content) get404(ctx router.Context) error {
	return content.RenderHtml(ctx, "404", "404", nil)
}

func (content *Content) getRobots(ctx router.Context) error {
	//userAgents := []*robots.UserAgent {
	//	robots.NewUserAgent(robots.EverythingUserAgent, []string { "/" }),
	//}
	//return ctx.Respond([]byte(robots.ToFileString(userAgents)))
	return ctx.Respond([]byte{})
}
