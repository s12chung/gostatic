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

func TestTracker_AddDependentURL(t *testing.T) {
	testCases := []struct {
		url string
		exp string
	}{
		{"/", "/"},
		{"/blah", "/blah"},
		{"blah", "/blah"},
		{"/fold/me", "/fold/me"},
		{"fold/me", "/fold/me"},
	}

	allTracker := defaultTracker(nil)
	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index": testCaseIndex,
			"url":   tc.url,
		})

		currentTracker := defaultTracker(nil)
		allTracker.AddDependentURL(tc.url)
		currentTracker.AddDependentURL(tc.url)

		var got string
		for k := range currentTracker.dependentURLs {
			if got != "" {
				t.Error(context.String("currentTracker.dependentURLs has > 1 key"))
			}
			got = k
			if got != tc.exp {
				t.Error(context.GotExpString("Result", got, tc.exp))
			}
		}
	}

	exp := []string{"/", "/blah", "/fold/me"}
	var got []string
	for k := range allTracker.dependentURLs {
		got = append(got, k)
	}
	sort.Strings(exp)
	sort.Strings(got)

	if !cmp.Equal(got, exp) {
		t.Error(test.AssertLabelString("allTracker.dependentURLs keys", got, exp))
	}
}

func TestTracker_URLs(t *testing.T) {
	testCases := []struct {
		urls            []string
		dependentURLs   []string
		independentURLs []string
	}{
		{[]string{}, []string{}, []string{}},
		{[]string{"/a", "/b"}, []string{"/a", "/b"}, []string{}},
		{[]string{"/a", "/b", "/c", "/d"}, []string{"/a", "/b"}, []string{"/c", "/d"}},
		{[]string{"/a", "/b", "/c", "/d"}, []string{}, []string{"/a", "/b", "/c", "/d"}},
		{[]string{"/a", "/b"}, []string{}, []string{"/a", "/b"}},
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
