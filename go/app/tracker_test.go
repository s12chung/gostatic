package app

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/s12chung/gostatic/go/test"
)

func defaultTracker(urls []string) *Tracker {
	return NewTracker(func() []string {
		return urls
	})
}

func TestTracker_Urls(t *testing.T) {
	testCases := []struct {
		urls            []string
		dependentUrls   []string
		independentUrls []string
	}{
		{[]string{}, []string{}, []string{}},
		{[]string{"a", "b"}, []string{"a", "b"}, []string{}},
		{[]string{"a", "b", "c", "d"}, []string{"a", "b"}, []string{"c", "d"}},
		{[]string{"a", "b", "c", "d"}, []string{}, []string{"a", "b", "c", "d"}},
		{[]string{"a", "b"}, []string{}, []string{"a", "b"}},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":         testCaseIndex,
			"urls":          tc.urls,
			"dependentUrls": tc.dependentUrls,
		})

		tracker := defaultTracker(tc.urls)
		for _, dependentUrl := range tc.dependentUrls {
			tracker.AddDependentUrl(dependentUrl)
		}

		got := tracker.DependentUrls()
		exp := tc.dependentUrls
		sort.Strings(got)
		sort.Strings(exp)

		if !cmp.Equal(got, exp) {
			t.Error(context.DiffString("dependentUrls", got, exp, cmp.Diff(got, exp)))
		}
		got, err := tracker.IndependentUrls()
		if err != nil {
			t.Error(context.String(err))
		}
		exp = tc.independentUrls
		sort.Strings(got)
		sort.Strings(exp)
		if !cmp.Equal(got, exp) {
			t.Error(context.DiffString("independentUrls", got, exp, cmp.Diff(got, exp)))
		}
	}
}
