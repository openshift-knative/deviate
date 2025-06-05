package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift-knative/deviate/pkg/cli"
	"github.com/openshift-knative/deviate/pkg/config"
	"github.com/spf13/cobra"
)

type sync struct {
	*cli.Options
}

func (s sync) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "sync",
		Short:     "Synchronize to the upstream releases",
		ValidArgs: []string{"REPOSITORY"},
		Args:      cobra.OnlyValidArgs,
		RunE:      s.run,
	}
	return cmd
}

func (s sync) run(cmd *cobra.Command, args []string) error {
	return cli.Sync(cmd, s.project(args)) //nolint:wrapcheck
}

// findGitRoot searches upwards from startPath for a directory containing a .git subdirectory.
func findGitRoot(startPath string) (string, error) {
	currentPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %s: %w", startPath, err)
	}

	for {
		gitPath := filepath.Join(currentPath, ".git")
		stat, err := os.Stat(gitPath)
		if err == nil && stat.IsDir() {
			return currentPath, nil
		}

		if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("error checking for .git directory at %s: %w", gitPath, err)
		}
		if err == nil && !stat.IsDir() {
			return "", fmt.Errorf(".git found at %s but it is not a directory", currentPath)
		}

		// .git not found here, try parent
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached the root of the filesystem
			return "", fmt.Errorf("'.git' directory not found in %s or any of its parent directories", startPath)
		}
		currentPath = parentPath
	}
}

func (s sync) project(args []string) func() config.Project {
	return func() config.Project {
		configDir := filepath.Dir(s.ConfigPath)
		projectRoot, err := findGitRoot(configDir)
		if err != nil {
			panic(fmt.Errorf("failed to determine project root: %w", err))
		}

		project := config.Project{
			ConfigPath: s.ConfigPath,
			Path:       projectRoot,
		}
		if len(args) > 0 {
			project.Path = args[0] // Override if path is provided as a command-line argument
		}
		return project
	}
}
