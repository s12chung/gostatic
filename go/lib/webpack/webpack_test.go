package webpack

import (
	"fmt"
	"path"
	"strings"
	"testing"

	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/gostatic/go/test"
	"github.com/s12chung/gostatic/go/test/testfile"
)

var generatedPath = path.Join(testfile.FixturePath, "generated")

var jpgResponsiveImage = &ResponsiveImage{
	"assets/content/images/test-37a65f446db3e9da33606b7eb48721bb-325.jpg",
	"assets/content/images/test-37a65f446db3e9da33606b7eb48721bb-325.jpg 325w, assets/content/images/test-c9d1dad468456287c20a476ade8a4d3f-750.jpg 750w, assets/content/images/test-be268849aa760a62798817c27db7c430-1500.jpg 1500w, assets/content/images/test-38e5ee006bf91e6af6d508bce2a9da4c-3000.jpg 3000w, assets/content/images/test-84800b3286f76133d1592c9e68fa10be-4000.jpg 4000w",
}
var pngResponsiveImage = &ResponsiveImage{
	"assets/content/images/test-afe607afeab81578d972f0ce9a92bdf4-325.png",
	"assets/content/images/test-afe607afeab81578d972f0ce9a92bdf4-325.png 325w, assets/content/images/test-d31be3db558b4fe54b2c098abdd96306-750.png 750w, assets/content/images/test-e4b7c37523ea30081ad02f6191b299f6-1440.png 1440w",
}

func defaultWebpack() (*Webpack, *logTest.Hook) {
	log, hook := logTest.NewNullLogger()
	settings := DefaultSettings()
	return NewWebpack(generatedPath, settings, log), hook
}

func TestWebpack_AssetsUrl(t *testing.T) {
	webpack, _ := defaultWebpack()
	got := webpack.AssetsURL()
	test.AssertLabel(t, "Result", got, "/assets/")
}

func TestWebpack_GeneratedAssetsPath(t *testing.T) {
	webpack, _ := defaultWebpack()
	got := webpack.GeneratedAssetsPath()
	test.AssertLabel(t, "Result", got, path.Join(generatedPath, webpack.settings.AssetsPath))
}

func TestWebpack_ManifestUrl(t *testing.T) {
	webpack, hook := defaultWebpack()
	got := webpack.ManifestURL("vendor.css")

	test.PrintLogEntries(t, hook)
	test.AssertLabel(t, "Result", got, path.Join(webpack.settings.AssetsPath, "vendor-32267303b2484ed8b3aa.css"))
}

func TestWebpack_GetResponsiveImage(t *testing.T) {
	webpack, hook := defaultWebpack()

	testCases := []struct {
		originalSrc string
		exp         *ResponsiveImage
		unsafeLog   bool
	}{
		{"content/images/test.jpg", jpgResponsiveImage, false},
		{"content/images/test.png", pngResponsiveImage, false},
		{"test.gif", &ResponsiveImage{Src: "assets/test.gif"}, false},
		{"http://testy.com/test.png", &ResponsiveImage{Src: "http://testy.com/test.png"}, false},
		{"http://testy.com/something.bad", &ResponsiveImage{Src: "http://testy.com/something.bad"}, false},
		{"does_not_exist.png", &ResponsiveImage{Src: "assets/does_not_exist.png"}, true},
		{"content/images/test_again.png", &ResponsiveImage{Src: "assets/content/images/test_again-1440.png"}, true},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext(t).SetFields(test.ContextFields{
			"index":       testCaseIndex,
			"originalSrc": tc.originalSrc,
			"unsafeLog":   tc.unsafeLog,
		})

		context.AssertArray("result", webpack.GetResponsiveImage(tc.originalSrc), tc.exp)
		if test.SafeLogEntries(hook) == tc.unsafeLog {
			test.PrintLogEntries(t, hook)
			t.Error(context.AssertString("test.SafeLogEntries(hook)", test.SafeLogEntries(hook), tc.unsafeLog))
		}
	}
}

func TestWebpack_ReplaceResponsiveAttrs(t *testing.T) {
	testCases := []struct {
		imageFilename      string
		srcPrefix          string
		input              string
		skipEmptyNamespace bool
		expected           string
	}{
		{"test.jpg", "content/images", `<img src="SRC"/>`, false, fmt.Sprintf(`<img %v/>`, jpgResponsiveImage.HTMLAttrs())},
		{"test.jpg", "content/images", `<img src="SRC" class="haha"/>`, false, fmt.Sprintf(`<img %v class="haha"/>`, jpgResponsiveImage.HTMLAttrs())},
		{"test.jpg", "content/images", `<img alt="blah" src="SRC"/>`, false, fmt.Sprintf(`<img alt="blah" %v/>`, jpgResponsiveImage.HTMLAttrs())},
		{"test.jpg", "content/images", `<img alt="blah" src="SRC" class="haha"/>`, false, fmt.Sprintf(`<img alt="blah" %v class="haha"/>`, jpgResponsiveImage.HTMLAttrs())},
		{"test_again.png", "content/images", `<img src="SRC"/>`, false, `<img src="assets/content/images/test_again-1440.png"/>`},
		{"test.gif", "", `<img src="SRC"/>`, false, `<img src="assets/test.gif"/>`},
		{"test.jpg", "doesnt_exist", `<img src="SRC"/>`, false, `<img src="assets/doesnt_exist/test.jpg"/>`},
		{"http://testy.com/test.png", "content/images", `<img src="SRC"/>`, true, `<img src="http://testy.com/test.png"/>`},
		{"http://testy.com/something.bad", "content/images", `<img src="SRC"/>`, true, `<img src="http://testy.com/something.bad"/>`},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext(t).SetFields(test.ContextFields{
			"index":         testCaseIndex,
			"imageFilename": tc.imageFilename,
			"srcPrefix":     tc.srcPrefix,
			"input":         tc.input,
		})

		webpack, _ := defaultWebpack()

		srcPrefixCases := []struct {
			srcPrefix string
			input     string
		}{
			{tc.srcPrefix, strings.Replace(tc.input, "SRC", tc.imageFilename, -1)},
			{"", strings.Replace(tc.input, "SRC", path.Join(tc.srcPrefix, tc.imageFilename), -1)},
		}

		for i, stc := range srcPrefixCases {
			if i == 1 && tc.skipEmptyNamespace {
				continue
			}
			context.Assert(
				fmt.Sprintf("result with srcPrefix: %v", stc.srcPrefix),
				webpack.ReplaceResponsiveAttrs(stc.srcPrefix, stc.input),
				tc.expected,
			)
		}
	}
}

func TestWebpack_ResponsiveHtmlAttrs(t *testing.T) {
	webpack, _ := defaultWebpack()
	got := string(webpack.ResponsiveHTMLAttrs("content/images/test.jpg"))
	exp := jpgResponsiveImage.HTMLAttrs()
	if got != exp {
		t.Error(test.AssertLabelString("result", got, exp))
	}
}
