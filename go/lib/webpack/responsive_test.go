package webpack

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/gostatic/go/test"
)

func defaultResponsive() (*Responsive, *logTest.Hook) {
	log, hook := logTest.NewNullLogger()
	return NewResponsive(generatedPath, DefaultSettings().AssetsPath, log), hook
}

func TestHasResponsiveExt(t *testing.T) {
	testCases := []struct {
		originalSrc string
		exp         bool
	}{
		{"test.jpg", true},
		{"test.png", true},
		{"test.gif", false},
		{"test.svg", false},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext(t).SetFields(test.ContextFields{
			"index":       testCaseIndex,
			"originalSrc": tc.originalSrc,
		})
		context.Assert("result", HasResponsiveExt(tc.originalSrc), tc.exp)
	}
}

func TestResponsive_GetResponsiveImage(t *testing.T) {
	testCases := []struct {
		originalSrc string
		exp         *ResponsiveImage
	}{
		{"content/images/test.jpg", jpgResponsiveImage},
		{"content/images/test.png", pngResponsiveImage},
		{"test.gif", nil},
		{"content/images/test.gif", nil},
		{"does_not_exist/test.png", nil},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext(t).SetFields(test.ContextFields{
			"index":       testCaseIndex,
			"originalSrc": tc.originalSrc,
		})

		responsive, _ := defaultResponsive()
		got := responsive.GetResponsiveImage(tc.originalSrc)

		if !cmp.Equal(got, tc.exp) {
			t.Error(context.AssertString("result", got, tc.exp))
		}
	}
}
