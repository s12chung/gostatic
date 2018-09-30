/*
Package webpack lets Go see into the generated asset paths, `Manifest.json`, and `images/responsive` folder of JSON files from Webpack.

Webpack struct implements github.com/s12chung/gostatic/go/lib/router/html.Plugin
*/
package webpack

import (
	"fmt"
	"html/template"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

var imgRegex = regexp.MustCompile(`<img[^>]+(src="([^"]*)")`)

// Webpack represents of a webpack generated setup
type Webpack struct {
	generatedPath string
	settings      *Settings
	manifest      *Manifest
	responsive    *Responsive
	log           logrus.FieldLogger
}

// NewWebpack returns a new instance of Webpack
func NewWebpack(generatedPath string, settings *Settings, log logrus.FieldLogger) *Webpack {
	return &Webpack{
		generatedPath,
		settings,
		NewManifest(generatedPath, settings.AssetsPath, log),
		NewResponsive(generatedPath, settings.AssetsPath, log),
		log,
	}
}

// AssetsURL returns the URL path prefix of all your assets
func (w *Webpack) AssetsURL() string {
	return fmt.Sprintf("/%v/", w.settings.AssetsPath)
}

// GeneratedAssetsPath returns the file path of the generated assets
func (w *Webpack) GeneratedAssetsPath() string {
	return filepath.Join(w.generatedPath, w.settings.AssetsPath)
}

// ManifestURL calls Manifest.ManifestURL
func (w *Webpack) ManifestURL(key string) string {
	return w.manifest.ManifestURL(key)
}

func (w *Webpack) manifestImage(originalSrc string) *ResponsiveImage {
	return &ResponsiveImage{Src: w.ManifestURL(originalSrc)}
}

// GetResponsiveImage returns the struct representation of a *ResponsiveImage given a originalSrc.
// originalSrc should give Webpack a filepath to the generated images folder.
//
// If the src points to a non-responsive image, will return a *ResponsiveImage
// with src set as:
// - the result of using originalSrc as a key for webpack.ManifestURL
// - the given originalSrc
func (w *Webpack) GetResponsiveImage(originalSrc string) *ResponsiveImage {
	u, err := url.Parse(originalSrc)
	if err == nil && u.Hostname() != "" {
		return &ResponsiveImage{Src: originalSrc}
	}

	if !HasResponsiveExt(originalSrc) {
		return w.manifestImage(originalSrc)
	}
	responsiveImage := w.responsive.GetResponsiveImage(originalSrc)
	if responsiveImage == nil {
		return w.manifestImage(originalSrc)
	}
	return responsiveImage
}

// ReplaceResponsiveAttrs replaces the img.src HTML attrs, to responsive img.src and img.srcset attrs
// within the HTML string. It takes the existing img.src values as the originalSrc to
// call webpack.GetResponsiveImage.
//
// You may add a srcPrefix to the img.src, so webpack.GetResponsiveImage can work.
func (w *Webpack) ReplaceResponsiveAttrs(srcPrefix, html string) string {
	return imgRegex.ReplaceAllStringFunc(html, func(imgTag string) string {
		matches := imgRegex.FindStringSubmatch(imgTag)
		originalSrc := matches[2]

		u, err := url.Parse(originalSrc)
		if err == nil && u.Hostname() != "" {
			return imgTag
		}

		responsiveImage := w.GetResponsiveImage(path.Join(srcPrefix, originalSrc))
		return strings.Replace(imgTag, matches[1], responsiveImage.HTMLAttrs(), 1)
	})
}

// ResponsiveHTMLAttrs calls GetResponsiveImage and returns the html attr (img.src and img.srcset) representation of the *ResponsiveImage
func (w *Webpack) ResponsiveHTMLAttrs(originalSrc string) template.HTMLAttr {
	responsiveImage := w.GetResponsiveImage(originalSrc)
	return template.HTMLAttr(responsiveImage.HTMLAttrs()) // #nosec G203
}

// TemplateFuncs implements github.com/s12chung/gostatic/go/lib/router/html.Plugin
func (w *Webpack) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"webpackURL":             w.ManifestURL,
		"responsiveAttrs":        w.ResponsiveHTMLAttrs,
		"replaceResponsiveAttrs": w.ReplaceResponsiveAttrs,
	}
}
