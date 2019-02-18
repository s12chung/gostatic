package app

import (
	"path"
	"testing"

	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/gostatic/go/test"
)

func TestDefaultSettings(t *testing.T) {
	test.EnvSetting(t, "GENERATED_PATH", "./generated", func() string {
		return DefaultSettings().GeneratedPath
	})
}

type basicSetting struct {
	GeneratedPath string `json:"generated_path,omitempty"`
	Concurrency   int    `json:"concurrency,omitempty"`
	IsEmpty       string `json:"is_empty,omitempty"`
}

type embeddedSetting struct {
	TopLevel string      `json:"top_level,omitempty"`
	Basic    interface{} `json:"basic,omitempty"`
}

func TestSettingsFromFile(t *testing.T) {
	testCases := []struct {
		path           string
		settings       interface{}
		structName     string
		safeLogEntries bool
	}{
		{"does not exist", nil, "", false},
		{"does not exist", &basicSetting{}, "basicSetting", false},
		{"a.md", &basicSetting{}, "basicSetting", false},
		{"basic_setting.json", &basicSetting{}, "basicSetting", true},
		{"embedded_setting.json", &embeddedSetting{Basic: &basicSetting{}}, "embeddedSetting", true},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext(t).SetFields(test.ContextFields{
			"index":      testCaseIndex,
			"path":       tc.path,
			"structName": tc.structName,
		})
		log, hook := logTest.NewNullLogger()

		SettingsFromFile(path.Join(test.FixturePath, tc.path), tc.settings, log)

		got := test.SafeLogEntries(hook)
		if got != tc.safeLogEntries {
			test.PrintLogEntries(t, hook)
			t.Error(context.AssertString("SafeLogEntries", got, tc.safeLogEntries))
		}
		if len(hook.AllEntries()) > 0 {
			continue
		}
		testSettingStruct(t, context, tc.settings, tc.structName)
	}
}

func testSettingStruct(t *testing.T, context *test.Context, setting interface{}, structName string) {
	switch structName {
	case "basicSetting":
		s, ok := setting.(*basicSetting)
		if !ok {
			t.Error(context.String("Failed to cast: " + structName))
		}
		test.AssertLabel(t, context.String("basicSetting.GeneratedPath"), s.GeneratedPath, "some_path")
		test.AssertLabel(t, context.String("basicSetting.Concurrency"), s.Concurrency, 22)
		test.AssertLabel(t, context.String("basicSetting.IsEmpty"), s.IsEmpty, "")
	case "embeddedSetting":
		s, ok := setting.(*embeddedSetting)
		if !ok {
			t.Error(context.String("Failed to cast: " + structName))
		}
		test.AssertLabel(t, context.String("basicSetting.TopLevel"), s.TopLevel, "the top of the world")
		testSettingStruct(t, context, s.Basic, "basicSetting")
	}
}
