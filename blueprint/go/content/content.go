/*
Package content contains the high level structure of the content of your site--your routes.
*/
package content

import (
	"mime"

	"github.com/s12chung/gostatic-packages/robots"
	"github.com/sirupsen/logrus"

	"github.com/s12chung/gostatic/go/app"
	"github.com/s12chung/gostatic/go/lib/html"
	"github.com/s12chung/gostatic/go/lib/router"
	"github.com/s12chung/gostatic/go/lib/webpack"
)

// Content represents contains the logic/routing for the content of your site
type Content struct {
	Settings *Settings
	Log      logrus.FieldLogger

	HTMLRenderer *html.Renderer
	Webpack      *webpack.Webpack
}

// NewContent returns Content with default config
func NewContent(generatedPath string, settings *Settings, log logrus.FieldLogger) *Content {
	w := webpack.NewWebpack(generatedPath, settings.Webpack, log)
	htmlRenderer := html.NewRenderer(settings.HTML, []html.Plugin{w}, log)
	return &Content{settings, log, htmlRenderer, w}
}

// AssetsURL is the URL path prefix of all your assets.
// Used when App.Host()-ing to generate the routes real-time, so the server can redirect this prefix to your assets
func (content *Content) AssetsURL() string {
	return content.Webpack.AssetsURL()
}

// GeneratedAssetsPath is the local file path of the generated assets.
// Used when App.Host()-ing to generate the routes real-time.
func (content *Content) GeneratedAssetsPath() string {
	return content.Webpack.GeneratedAssetsPath()
}

func (content *Content) renderHTML(ctx router.Context, name string, layoutD interface{}) error {
	bytes, err := content.HTMLRenderer.Render(name, layoutD)
	if err != nil {
		return err
	}
	ctx.Respond(bytes)
	return nil
}

// SetRoutes is where you set the routes
func (content *Content) SetRoutes(r router.Router, tracker *app.Tracker) error {
	r.GetRootHTML(content.getRoot)
	r.GetHTML("/404.html", content.get404)
	r.Get("/robots.txt", content.getRobots)
	return nil
}

func (content *Content) getRoot(ctx router.Context) error {
	return content.renderHTML(ctx, "root", layoutData{"", "Hello World!"})
}

func (content *Content) get404(ctx router.Context) error {
	return content.renderHTML(ctx, "404", layoutData{"404", nil})
}

func (content *Content) getRobots(ctx router.Context) error {
	userAgents := []*robots.UserAgent{
		robots.NewUserAgent(robots.EverythingUserAgent, []string{"/"}),
	}
	ctx.SetContentType(mime.TypeByExtension(".txt"))
	ctx.Respond([]byte(robots.ToFileString(userAgents)))
	return nil
}
