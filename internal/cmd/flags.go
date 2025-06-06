package cmd

import (
	"github.com/openshift-knative/deviate/pkg/cli"
	"github.com/openshift-knative/deviate/pkg/metadata"
	"github.com/spf13/cobra"
)

func addFlags(root *cobra.Command, opts *cli.Options) {
	fl := root.PersistentFlags()
	fl.StringVar(&opts.ConfigPath, "config", ".deviate.yaml",
		metadata.Name+" configuration file")
}
