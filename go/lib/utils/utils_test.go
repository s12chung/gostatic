package utils

import (
	"path"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/s12chung/gostatic/go/test"
)

func TestCleanFilePath(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"/go/src/github.com/s12chung/gostatic", "go/src/github.com/s12chung/gostatic"},
		{"/go/src/github.com/s12chung/gostatic/", "go/src/github.com/s12chung/gostatic"},
		{"go/src/github.com/s12chung/gostatic", "go/src/github.com/s12chung/gostatic"},
		{"./go/src/github.com/s12chung/gostatic", "go/src/github.com/s12chung/gostatic"},
		{"./../go/src/github.com/s12chung/gostatic", "../go/src/github.com/s12chung/gostatic"},
		{"", ""},
		{"./", ""},
		{".", ""},
	}

	for _, tc := range testCases {
		got := CleanFilePath(tc.input)
		test.AssertInput(t, tc.input, got, tc.expected)
	}
}

func TestToSimpleQuery(t *testing.T) {
	testCases := []struct {
		input    map[string]string
		expected string
	}{
		{map[string]string{"a": "1", "b": "2", "c": "3"}, "a=1&b=2&c=3"},
		{map[string]string{"a": "1"}, "a=1"},
		{map[string]string{}, ""},
	}

	for _, tc := range testCases {
		got := ToSimpleQuery(tc.input)
		test.AssertInput(t, tc.input, got, tc.expected)
	}
}

func TestSliceList(t *testing.T) {
	testCases := []struct {
		input    []string
		expected string
	}{
		{[]string{"Johnny", "Eugene", "Kate", "Katherine"}, "Johnny, Eugene, Kate & Katherine"},
		{[]string{"Mike", "Cedric"}, "Mike & Cedric"},
		{[]string{"Steve"}, "Steve"},
		{[]string{}, ""},
	}

	for _, tc := range testCases {
		got := SliceList(tc.input)
		test.AssertInput(t, tc.input, got, tc.expected)
	}
}

func TestFilePaths(t *testing.T) {
	testCases := []struct {
		ext      string
		dirPaths []string
		expected map[string][]string
		error    bool
	}{
		{"", []string{""}, map[string][]string{"": {"a.md", "b.md"}}, false},
		{".md", []string{""}, map[string][]string{"": {"a.md", "b.md"}}, false},
		{".md", []string{"dir1"}, map[string][]string{"dir1": {"1.md"}}, false},
		{".md", []string{"dir1", "dir2"}, map[string][]string{"dir1": {"1.md"}}, false},
		{".md", []string{"dir1", "dir2", "dir3"}, map[string][]string{"dir1": {"1.md"}}, false},
		{".md", []string{"", "dir1", "dir2", "dir3"}, map[string][]string{"": {"a.md", "b.md"}, "dir1": {"1.md"}}, false},
		{".txt", []string{""}, map[string][]string{}, false},
		{".txt", []string{"dir1"}, map[string][]string{"dir1": {"1.txt", "2.txt"}}, false},
		{".txt", []string{"dir1", "dir2"}, map[string][]string{"dir1": {"1.txt", "2.txt"}, "dir2": {"a.txt"}}, false},
		{".txt", []string{"dir1", "dir2", "dir3"}, map[string][]string{"dir1": {"1.txt", "2.txt"}, "dir2": {"a.txt"}}, false},
		{".txt", []string{"", "dir1", "dir2", "dir3"}, map[string][]string{"dir1": {"1.txt", "2.txt"}, "dir2": {"a.txt"}}, false},
		{".txt", []string{"does not exist"}, nil, true},
		{".md", []string{"", "does not exist"}, nil, true},
	}

	for testCaseIndex, tc := range testCases {
		context := test.NewContext().SetFields(test.ContextFields{
			"index":    testCaseIndex,
			"ext":      tc.ext,
			"dirPaths": tc.dirPaths,
		})

		dirPaths := make([]string, len(tc.dirPaths))
		for i, d := range tc.dirPaths {
			dirPaths[i] = path.Join(test.FixturePath, d)
		}

		got, err := FilePaths(tc.ext, dirPaths...)
		if tc.error && err != nil {
			continue
		}
		if err != nil {
			t.Error(context.String(err))
			continue
		}

		var exp []string
		for relativePath, files := range tc.expected {
			for _, file := range files {
				exp = append(exp, path.Join(test.FixturePath, relativePath, file))
			}
		}

		sort.Strings(exp)
		sort.Strings(got)

		if !cmp.Equal(got, exp) {
			t.Error(context.DiffString("Result", got, exp, cmp.Diff(got, exp)))
		}
	}
}
