package cmd

import (
	"os"
	"path"

	"github.com/openshift-knative/deviate/pkg/cli"
	"github.com/openshift-knative/deviate/pkg/metadata"
	"github.com/spf13/cobra"
)

func addFlags(root *cobra.Command, opts *cli.Options) {
	fl := root.PersistentFlags()
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Default to .deviate.yaml or .github/deviate.yaml in the root
	config := path.Join(wd, ".deviate.yaml")
	if _, err := os.Stat(config); os.IsNotExist(err) {
		githubConfig := path.Join(wd, ".github", "deviate.yaml")
		if _, errStatGithub := os.Stat(githubConfig); errStatGithub == nil {
			config = githubConfig
		}
	}
	fl.StringVar(&opts.ConfigPath, "config", config,
		metadata.Name+" configuration file")
}
