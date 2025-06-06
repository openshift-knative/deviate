package files_test

import (
	"testing"

	"github.com/openshift-knative/deviate/pkg/files"
	"github.com/stretchr/testify/assert"
)

func TestFilters_Match(t *testing.T) {
	filelist := []string{
		"a/b/c.txt",
		"a/d.txt",
		"a/b/d.md",
		"a/b/d.txt",
		"a.txt",
		"b.md",
	}
	tcs := []testFiltersMatchCases{{
		name: "all md files",
		Filters: files.Filters{
			Include: []string{"**/*.md", "*.md"},
		},
		want: []string{
			"a/b/d.md",
			"b.md",
		},
	}, {
		name: "root level md files",
		Filters: files.Filters{
			Include: []string{"*.md"},
		},
		want: []string{
			"b.md",
		},
	}, {
		"just one txt",
		files.Filters{
			Include: []string{"**.txt"},
			Exclude: []string{"a/b**", "a.txt"},
		},
		[]string{
			"a/d.txt",
		},
	}}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			matcher := tc.Matcher()
			got := make([]string, 0, len(tc.want))
			for _, f := range filelist {
				if matcher.Matches(f) {
					got = append(got, f)
				}
			}

			assert.Equal(t, tc.want, got)
		})
	}
}

type testFiltersMatchCases struct {
	name string
	files.Filters
	want []string
}
