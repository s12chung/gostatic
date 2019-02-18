package html

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/gostatic/go/test"
	"github.com/s12chung/gostatic/go/test/factory"
	"github.com/s12chung/gostatic/go/test/testfile"
)

var updateFixturesPtr = testfile.UpdateFixtureFlag()

func defaultRenderer() (*Renderer, *logTest.Hook) {
	settings := DefaultSettings()
	settings.TemplatePath = testfile.FixturePath
	log, hook := logTest.NewNullLogger()
	return NewRenderer(settings, []Plugin{}, log), hook
}

type layoutData struct {
	Title       string
	ContentData interface{}
}

func TestRenderer_RenderWithLayout(t *testing.T) {
	renderer, hook := defaultRenderer()

	testCases := []struct {
		layoutName string
		name       string
		layoutData layoutData
	}{
		{"layout_with_title", "title", layoutData{"", nil}},
		{"layout_with_title", "title", layoutData{"The Given", nil}},
		{"layout_with_title", "title", layoutData{"something", nil}},
		{"", "no_template_content", layoutData{}},
		{"layout", "helpers", layoutData{"", map[string]interface{}{"HTML": `<span>html_data</span>`, "Date": factory.Time(1)}}},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext(t).SetFields(test.ContextFields{
			"index":      testCaseIndex,
			"layoutName": tc.layoutName,
			"name":       tc.name,
			"layoutData": tc.layoutData,
		})

		rendered, err := renderer.RenderWithLayout(tc.layoutName, tc.name, tc.layoutData)
		if err != nil {
			test.PrintLogEntries(t, hook)
			t.Error(context.String(err))
		}

		got := strings.TrimSpace(string(rendered))

		fixtureName := tc.name + ".html"
		if tc.name == "title" {
			fixtureName = tc.name + strconv.Itoa(testCaseIndex) + ".html"
		}
		if *updateFixturesPtr {
			testfile.WriteFixture(t, fixtureName, []byte(got))
			continue
		}

		exp := strings.TrimSpace(string(testfile.ReadFixture(t, fixtureName)))
		if got != exp {
			t.Error(context.DiffString("Result", got, exp, cmp.Diff(got, exp)))
		}
	}
}

func TestRenderer_Render_Settings(t *testing.T) {
	testCases := []struct {
		layoutName  string
		templateExt string
	}{
		{"layout", ".gohtml"},
		{"", ".gohtml"},
		{"layout", ".tmpl"},
		{"", ".tmpl"},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext(t).SetFields(test.ContextFields{
			"index":       testCaseIndex,
			"layoutName":  tc.layoutName,
			"templateExt": tc.templateExt,
		})

		renderer, hook := defaultRenderer()
		renderer.settings.LayoutName = tc.layoutName
		renderer.settings.TemplateExt = tc.templateExt
		rendered, err := renderer.Render("settings", nil)
		if err != nil {
			test.PrintLogEntries(t, hook)
			t.Error(context.String(err))
		}

		got := strings.TrimSpace(string(rendered))

		fixtureName := fmt.Sprintf("settings%v.html", testCaseIndex)
		if *updateFixturesPtr {
			testfile.WriteFixture(t, fixtureName, []byte(got))
			continue
		}

		exp := strings.TrimSpace(string(testfile.ReadFixture(t, fixtureName)))
		if got != exp {
			t.Error(context.DiffString("Result", got, exp, cmp.Diff(got, exp)))
		}
	}
}

type intPlugin struct{}

func (p *intPlugin) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"pInt": func() int {
			return 999
		},
	}
}

type stringPlugin struct{}

func (p *stringPlugin) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"pString": func() string {
			return "Strings are forever"
		},
	}
}

func TestRenderer_Render_Plugins(t *testing.T) {
	testCases := []struct {
		plugins []Plugin
	}{
		{[]Plugin{}},
		{[]Plugin{&stringPlugin{}, &intPlugin{}}},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext(t).SetFields(test.ContextFields{
			"index":   testCaseIndex,
			"plugins": tc.plugins,
		})

		renderer, hook := defaultRenderer()
		renderer.plugins = tc.plugins
		rendered, err := renderer.Render("plugins", nil)
		if err != nil {
			if len(tc.plugins) != 0 {
				test.PrintLogEntries(t, hook)
				t.Error(context.String(err))
			}
			continue
		}

		got := strings.TrimSpace(string(rendered))
		fixtureName := "plugins.html"
		if *updateFixturesPtr {
			testfile.WriteFixture(t, fixtureName, []byte(got))
			continue
		}

		exp := strings.TrimSpace(string(testfile.ReadFixture(t, fixtureName)))
		if got != exp {
			t.Error(context.DiffString("Result", got, exp, cmp.Diff(got, exp)))
		}
	}
}
