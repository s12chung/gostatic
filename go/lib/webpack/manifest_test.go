package webpack

import (
	"testing"

	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/gostatic/go/test"
)

func defaultManifest() (*Manifest, *logTest.Hook) {
	log, hook := logTest.NewNullLogger()
	return NewManifest(generatedPath, "assets", log), hook
}

func TestManifest_ManifestUrl(t *testing.T) {
	testCases := []struct {
		assetsFolder string
		key          string
		exp          string
		safeLog      bool
	}{
		{"assets", "test.gif", "assets/test.gif", true},
		{"assets", "vendor.css", "assets/vendor-32267303b2484ed8b3aa.css", true},
		{"assets", "content/images/test.png", "assets/content/images/test-1440.png", true},
		{"assets", "does_not_exist.gif", "assets/does_not_exist.gif", false},
		{"does_not_exist", "test.gif", "does_not_exist/test.gif", false},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext(t).SetFields(test.ContextFields{
			"index":        testCaseIndex,
			"assetsFolder": tc.assetsFolder,
			"key":          tc.key,
		})

		manifest, hook := defaultManifest()
		manifest.assetsFolder = tc.assetsFolder
		context.Assert("Result", manifest.ManifestURL(tc.key), tc.exp)
		context.Assert("test.SafeLogEntries(hook)", test.SafeLogEntries(hook), tc.safeLog)
	}
}
