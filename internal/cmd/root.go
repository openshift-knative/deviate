package cmd

import (
	"os"

	"github.com/openshift-knative/deviate/pkg/cli"
	"github.com/openshift-knative/deviate/pkg/metadata"
	"github.com/spf13/cobra"
	"github.com/wavesoftware/go-commandline"
)

// Options hold the overrides for regular execution of main cobra.Command.
var Options = make([]commandline.Option, 0, 1) //nolint:gochecknoglobals

// App is the main application.
type App struct{}

func (a App) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:     metadata.Name,
		Short:   metadata.Description,
		Version: metadata.Version,
	}
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.PersistentPreRun = func(cmd *cobra.Command, _ []string) {
		cmd.SilenceUsage = true
	}
	opts := &cli.Options{}
	subs := []subcommand{
		sync{opts},
	}
	addFlags(cmd, opts)
	for _, sub := range subs {
		cmd.AddCommand(sub.command())
	}

	return cmd
}

type subcommand interface {
	command() *cobra.Command
}
