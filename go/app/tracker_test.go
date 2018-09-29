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
		dependentURLs   []string
		independentURLs []string
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
			"dependentURLs": tc.dependentURLs,
		})

		tracker := defaultTracker(tc.urls)
		for _, dependentURL := range tc.dependentURLs {
			tracker.AddDependentURL(dependentURL)
		}

		got := tracker.DependentURLs()
		exp := tc.dependentURLs
		sort.Strings(got)
		sort.Strings(exp)

		if !cmp.Equal(got, exp) {
			t.Error(context.DiffString("dependentURLs", got, exp, cmp.Diff(got, exp)))
		}
		got, err := tracker.IndependentURLs()
		if err != nil {
			t.Error(context.String(err))
		}
		exp = tc.independentURLs
		sort.Strings(got)
		sort.Strings(exp)
		if !cmp.Equal(got, exp) {
			t.Error(context.DiffString("independentURLs", got, exp, cmp.Diff(got, exp)))
		}
	}
}
