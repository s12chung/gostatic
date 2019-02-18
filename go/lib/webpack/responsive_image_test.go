package webpack

import (
	"path"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/s12chung/gostatic/go/lib/utils"
	"github.com/s12chung/gostatic/go/test"
)

func copyResponsiveImage(v *ResponsiveImage) *ResponsiveImage {
	cp := *v
	return &cp
}

func TestResponsiveImage_ChangeSrcPrefix(t *testing.T) {
	placeholder := "PREFIX/"

	testCases := []struct {
		img     *ResponsiveImage
		exp     *ResponsiveImage
		safeLog bool
	}{
		{
			&ResponsiveImage{"", ""},
			&ResponsiveImage{"", ""},
			true,
		},
		{
			&ResponsiveImage{"blah.png", ""},
			&ResponsiveImage{placeholder + "blah.png", ""},
			true,
		},
		{
			&ResponsiveImage{"blah.png", "blah-125.png 125w"},
			&ResponsiveImage{placeholder + "blah.png", placeholder + "blah-125.png 125w"},
			true,
		},
		{
			&ResponsiveImage{"blah.png", "blah-125.png 125w, blah-125.png 250w, blah-125.png 125w, blah-500.png 500w"},
			&ResponsiveImage{placeholder + "blah.png", placeholder + "blah-125.png 125w, " + placeholder + "blah-125.png 250w, " + placeholder + "blah-125.png 125w, " + placeholder + "blah-500.png 500w"},
			true,
		},
		{
			&ResponsiveImage{"content/images/blah.png", ""},
			&ResponsiveImage{placeholder + "blah.png", ""},
			true,
		},
		{
			&ResponsiveImage{"content/images/blah.png", "content/images/blah-125.png 125w"},
			&ResponsiveImage{placeholder + "blah.png", placeholder + "blah-125.png 125w"},
			true,
		},
		{
			&ResponsiveImage{"content/images/blah.png", "content/images/blah-125.png 125w, content/images/blah-125.png 250w, content/images/blah-125.png 125w, content/images/blah-500.png 500w"},
			&ResponsiveImage{placeholder + "blah.png", placeholder + "blah-125.png 125w, " + placeholder + "blah-125.png 250w, " + placeholder + "blah-125.png 125w, " + placeholder + "blah-500.png 500w"},
			true,
		},
		{
			&ResponsiveImage{"content/images/blah.png", "content/images/blah-125.png 125w,"},
			&ResponsiveImage{placeholder + "blah.png", placeholder + "blah-125.png 125w"},
			false,
		}, {
			&ResponsiveImage{"content/images/blah.png", ",content/images/blah-125.png 125w,,"},
			&ResponsiveImage{placeholder + "blah.png", placeholder + "blah-125.png 125w"},
			false,
		},
		{
			&ResponsiveImage{"content/images/blah.png", "content/images/blah-125.png 125w, content/images/blah-125.png 250w, content/images/blah-125.png 125w, content/images/blah-500.png 500w,"},
			&ResponsiveImage{placeholder + "blah.png", placeholder + "blah-125.png 125w, " + placeholder + "blah-125.png 250w, " + placeholder + "blah-125.png 125w, " + placeholder + "blah-500.png 500w"},
			false,
		},
		{
			&ResponsiveImage{"content/images/blah.png", "content/images/blah-125.png 125w, , content/images/blah-125.png 250w, , content/images/blah-125.png 125w, content/images/blah-500.png 500w,"},
			&ResponsiveImage{placeholder + "blah.png", placeholder + "blah-125.png 125w, " + placeholder + "blah-125.png 250w, " + placeholder + "blah-125.png 125w, " + placeholder + "blah-500.png 500w"},
			false,
		},
	}

	for testCaseIndex, tc := range testCases {
		prefixes := []string{
			"",
			"/",
			"testy",
			"/testy",
			"testy/",
			"/testy/",
			"long/long",
			"long/long/way/",
		}

		for _, prefix := range prefixes {
			context := test.NewContext(t).SetFields(test.ContextFields{
				"index":  testCaseIndex,
				"img":    tc.img,
				"prefix": prefix,
			})

			log, hook := logTest.NewNullLogger()
			got := copyResponsiveImage(tc.img)
			got.PrependSrcPath(prefix, log)

			exp := copyResponsiveImage(tc.exp)

			fullPrefix := path.Join(utils.CleanFilePath(prefix), path.Dir(tc.img.Src)) + "/"
			if fullPrefix == "./" {
				fullPrefix = ""
			}
			exp.Src = strings.Replace(exp.Src, placeholder, fullPrefix, 1)
			exp.SrcSet = strings.Replace(exp.SrcSet, placeholder, fullPrefix, -1)

			if !cmp.Equal(got, exp) {
				t.Error(context.AssertString("Result", got, exp))
			}
			context.Assert("test.SafeLogEntries(hook)", test.SafeLogEntries(hook), tc.safeLog)
		}
	}
}

func TestResponsiveImage_HtmlAttrs(t *testing.T) {
	testCases := []struct {
		img *ResponsiveImage
		exp string
	}{
		{
			&ResponsiveImage{"", ""},
			"",
		},
		{
			&ResponsiveImage{"blah.png", ""},
			`src="blah.png"`,
		},

		{
			&ResponsiveImage{"blah.png", "blah-125.png 125w"},
			`src="blah.png" srcset="blah-125.png 125w"`,
		},
		{
			&ResponsiveImage{"", "blah-125.png 125w"},
			`srcset="blah-125.png 125w"`,
		},
		{
			&ResponsiveImage{"blah.png", "blah-125.png 125w, blah-125.png 250w, blah-125.png 125w, blah-500.png 500w"},
			`src="blah.png" srcset="blah-125.png 125w, blah-125.png 250w, blah-125.png 125w, blah-500.png 500w"`,
		},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext(t).SetFields(test.ContextFields{
			"index": testCaseIndex,
			"img":   tc.img,
		})
		context.Assert("Result", tc.img.HTMLAttrs(), tc.exp)
	}
}
