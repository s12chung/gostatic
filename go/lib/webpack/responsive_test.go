package webpack

import (
	"github.com/google/go-cmp/cmp"
	"github.com/s12chung/gostatic/go/test"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
)

func defaultResponsive() (*Responsive, *logTest.Hook) {
	log, hook := logTest.NewNullLogger()
	return NewResponsive(generatedPath, DefaultSettings().AssetsPath, log), hook
}

func TestHasResponsive(t *testing.T) {
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
		context := test.NewContext().SetFields(test.ContextFields{
			"index":       testCaseIndex,
			"originalSrc": tc.originalSrc,
		})
		got := HasResponsive(tc.originalSrc)
		if got != tc.exp {
			t.Error(context.GotExpString("result", got, tc.exp))
		}
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
		context := test.NewContext().SetFields(test.ContextFields{
			"index":       testCaseIndex,
			"originalSrc": tc.originalSrc,
		})

		responsive, _ := defaultResponsive()
		got := responsive.GetResponsiveImage(tc.originalSrc)

		if !cmp.Equal(got, tc.exp) {
			t.Error(context.GotExpString("result", got, tc.exp))
		}
	}
}
