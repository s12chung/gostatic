/*
Package html maps your filesystem to `html/template` and handles templates.
*/
package html

import (
	"bytes"
	"fmt"
	"html/template"
	"path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/s12chung/gostatic/go/lib/utils"
)

// Renderer holds settings and config to Render with
type Renderer struct {
	settings *Settings
	plugins  []Plugin
	log      logrus.FieldLogger
}

// NewRenderer returns a new instance of Renderer
func NewRenderer(settings *Settings, plugins []Plugin, log logrus.FieldLogger) *Renderer {
	return &Renderer{
		settings,
		plugins,
		log,
	}
}

// Plugin for Renderer to add template functions with
type Plugin interface {
	TemplateFuncs() template.FuncMap
}

func (renderer *Renderer) partialPaths() ([]string, error) {
	filePaths, err := utils.FilePaths(renderer.settings.TemplateExt, renderer.settings.TemplatePath)
	if err != nil {
		return nil, err
	}

	var partialPaths []string
	for _, filePath := range filePaths {
		if strings.HasPrefix(filepath.Base(filePath), "_") {
			partialPaths = append(partialPaths, filePath)
		}
	}
	return partialPaths, nil
}

func (renderer *Renderer) templateFuncs() template.FuncMap {
	defaults := defaultTemplateFuncs()
	mergeFuncMap(defaults, template.FuncMap{
		"title": func(t string) string {
			if t == "" {
				return renderer.settings.WebsiteTitle
			}
			return fmt.Sprintf("%v - %v", t, renderer.settings.WebsiteTitle)
		},
	})
	for _, plugin := range renderer.plugins {
		mergeFuncMap(defaults, plugin.TemplateFuncs())
	}
	return defaults
}

func mergeFuncMap(dest, src template.FuncMap) {
	for k, v := range src {
		dest[k] = v
	}
}

// RenderWithLayout renders the given templateName with the given layoutName and data
// It finds the templates within the given html.Settings.TemplatePath.
//
// Default template functions are provided in addition to the plugin template functions.
// See https://github.com/s12chung/gostatic/blob/master/go/lib/html/helpers.go for a list
// of default helper functions.
func (renderer *Renderer) RenderWithLayout(layoutName, templateName string, layoutData interface{}) ([]byte, error) {
	partialPaths, err := renderer.partialPaths()
	if err != nil {
		return nil, err
	}

	rootTemplateFilename := templateName + renderer.settings.TemplateExt
	templatePaths := append(partialPaths, path.Join(renderer.settings.TemplatePath, rootTemplateFilename))
	if layoutName != "" {
		rootTemplateFilename = layoutName + renderer.settings.TemplateExt
		templatePaths = append(templatePaths, path.Join(renderer.settings.TemplatePath, rootTemplateFilename))
	}

	gohtml, err := template.New("self").Funcs(renderer.templateFuncs()).ParseFiles(templatePaths...)
	if err != nil {
		return nil, err
	}

	buffer := &bytes.Buffer{}
	err = gohtml.ExecuteTemplate(buffer, rootTemplateFilename, layoutData)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Render calls RenderWithLayout with the default layoutName from html.Settings.LayoutName
func (renderer *Renderer) Render(name string, layoutData interface{}) ([]byte, error) {
	return renderer.RenderWithLayout(renderer.settings.LayoutName, name, layoutData)
}
