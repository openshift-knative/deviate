package git_test

import (
	"slices"
	"testing"

	gitv5 "github.com/go-git/go-git/v5"
	gitv5config "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/openshift-knative/deviate/pkg/config"
	configgit "github.com/openshift-knative/deviate/pkg/config/git"
	"github.com/openshift-knative/deviate/pkg/files"
	"github.com/openshift-knative/deviate/pkg/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckout_OntoWorkspace(t *testing.T) {
	projectPath := t.TempDir()
	remote := configgit.Remote{
		Name: "origin",
		URL:  "https://github.com/cardil/ghet",
	}
	mainBranch := "main"
	gr, err := gitv5.PlainClone(projectPath, false, &gitv5.CloneOptions{
		URL:           remote.URL,
		RemoteName:    "origin",
		Depth:         1,
		SingleBranch:  true,
		ReferenceName: plumbing.NewBranchReferenceName(mainBranch),
	})
	require.NoError(t, err)
	wt, gerr := gr.Worktree()
	require.NoError(t, gerr)
	initCommit := "9ab17a360b240c506cf98dfa83997563aa6d9a28"
	require.NoError(t, gr.Fetch(&gitv5.FetchOptions{
		RemoteName: "origin",
		RefSpecs: []gitv5config.RefSpec{
			gitv5config.RefSpec(initCommit + ":refs/heads/target"),
		},
		Depth: 1,
	}))
	require.NoError(t, wt.Checkout(&gitv5.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("target"),
	}))
	repo := &git.Repository{
		Context: t.Context(),
		Project: config.Project{
			Path: projectPath,
		},
		Repository: gr,
	}
	filters := files.Filters{
		Include: []string{
			"**.go",
		},
		Exclude: []string{
			"**internal**",
			"**pkg**",
			"**Magefile**",
			"**_test.go",
		},
	}
	err = repo.Checkout(remote, mainBranch).OntoWorkspace(filters)
	require.NoError(t, err)

	st, serr := wt.Status()
	require.NoError(t, serr)
	got := make([]string, 0, len(st))
	for f := range st {
		got = append(got, f)
	}
	slices.Sort(got)
	assert.Equal(t, []string{"build/mage.go", "cmd/ght/main.go"}, got)
}
