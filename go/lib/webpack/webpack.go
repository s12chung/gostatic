package webpack

import (
	"fmt"
	"html/template"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

var imgRegex = regexp.MustCompile(`<img[^>]+(src="([^"]*)")`)

type Webpack struct {
	generatedPath string
	settings      *Settings
	manifest      *Manifest
	responsive    *Responsive
	log           logrus.FieldLogger
}

func NewWebpack(generatedPath string, settings *Settings, log logrus.FieldLogger) *Webpack {
	return &Webpack{
		generatedPath,
		settings,
		NewManifest(generatedPath, settings.AssetsPath, log),
		NewResponsive(generatedPath, settings.AssetsPath, log),
		log,
	}
}

func (w *Webpack) AssetsPath() string {
	return w.settings.AssetsPath
}

func (w *Webpack) AssetsUrl() string {
	return fmt.Sprintf("/%v/", w.AssetsPath())
}

func (w *Webpack) GeneratedAssetsPath() string {
	return filepath.Join(w.generatedPath, w.AssetsPath())
}

func (w *Webpack) ManifestUrl(key string) string {
	return w.manifest.ManifestUrl(key)
}

func (w *Webpack) GetResponsiveImage(originalSrc string) *ResponsiveImage {
	manifestImage := func() *ResponsiveImage {
		return &ResponsiveImage{Src: w.ManifestUrl(originalSrc)}
	}

	if !HasResponsive(originalSrc) {
		return manifestImage()
	}
	responsiveImage := w.responsive.GetResponsiveImage(originalSrc)
	if responsiveImage == nil {
		return manifestImage()
	}
	return responsiveImage
}

func (w *Webpack) ReplaceResponsiveAttrs(srcPrefix, html string) string {
	return imgRegex.ReplaceAllStringFunc(html, func(imgTag string) string {
		matches := imgRegex.FindStringSubmatch(imgTag)
		responsiveImage := w.GetResponsiveImage(path.Join(srcPrefix, matches[2]))
		return strings.Replace(imgTag, matches[1], responsiveImage.HtmlAttrs(), 1)
	})
}

func (w *Webpack) ResponsiveHtmlAttrs(originalSrc string) template.HTMLAttr {
	responsiveImage := w.GetResponsiveImage(originalSrc)
	return template.HTMLAttr(responsiveImage.HtmlAttrs())
}

func (w *Webpack) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"webpackUrl":             w.ManifestUrl,
		"responsiveAttrs":        w.ResponsiveHtmlAttrs,
		"replaceResponsiveAttrs": w.ReplaceResponsiveAttrs,
	}
}
