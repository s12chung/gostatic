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

type Renderer struct {
	settings *Settings
	plugins  []Plugin
	log      logrus.FieldLogger
}

func NewRenderer(settings *Settings, plugins []Plugin, log logrus.FieldLogger) *Renderer {
	return &Renderer{
		settings,
		plugins,
		log,
	}
}

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

func (renderer *Renderer) RenderWithLayout(layoutName, name string, layoutData interface{}) ([]byte, error) {
	partialPaths, err := renderer.partialPaths()
	if err != nil {
		return nil, err
	}

	rootTemplateFilename := name + renderer.settings.TemplateExt
	templatePaths := append(partialPaths, path.Join(renderer.settings.TemplatePath, rootTemplateFilename))
	if layoutName != "" {
		rootTemplateFilename = layoutName + renderer.settings.TemplateExt
		templatePaths = append(templatePaths, path.Join(renderer.settings.TemplatePath, rootTemplateFilename))
	}

	tmpl, err := template.New("self").
		Funcs(renderer.templateFuncs()).
		ParseFiles(templatePaths...)
	if err != nil {
		return nil, err
	}

	buffer := &bytes.Buffer{}
	err = tmpl.ExecuteTemplate(buffer, rootTemplateFilename, layoutData)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (renderer *Renderer) Render(name string, layoutData interface{}) ([]byte, error) {
	return renderer.RenderWithLayout(renderer.settings.LayoutName, name, layoutData)
}
