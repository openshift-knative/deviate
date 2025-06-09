package files_test

import (
	"os"
	"path"
	"slices"
	"testing"

	gitv5 "github.com/go-git/go-git/v5"
	"github.com/openshift-knative/deviate/pkg/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteByFilters(t *testing.T) {
	root := t.TempDir()
	paths := []string{
		"a.txt",
		"a/b.txt",
		"a/b/c.md",
		"a/b/c.txt",
		"a/b/d.md",
		"a/b/d.txt",
		"a/b/e/f.md",
		"a/b/e/f.txt",
		"a/c.txt",
		"c.txt",
	}
	for _, file := range paths {
		fp := path.Join(root, file)
		dir := path.Dir(fp)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(fp, []byte("test"), 0o600))
	}

	filters := files.Filters{
		Include: []string{"a/b/**"},
		Exclude: []string{"**.txt"},
	}
	require.NoError(t, filters.DeleteFiles(root))

	repo, err := gitv5.PlainInit(root, false)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	st, serr := wt.Status()
	require.NoError(t, serr)
	got := make([]string, 0, len(st))
	for f := range st {
		got = append(got, f)
	}
	slices.Sort(got)
	want := []string{
		"a.txt",
		"a/b.txt",
		"a/b/c.txt",
		"a/b/d.txt",
		"a/b/e/f.txt",
		"a/c.txt",
		"c.txt",
	}
	assert.Equal(t, want, got)
}
