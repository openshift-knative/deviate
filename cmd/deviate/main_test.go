package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/openshift-knative/deviate/internal/cmd"
	"github.com/openshift-knative/deviate/pkg/metadata"
	"github.com/spf13/cobra"
	"github.com/wavesoftware/go-commandline"
	"gotest.tools/v3/assert"
)

func TestMainFunc(t *testing.T) {
	var o bytes.Buffer
	var retcode *int
	withOptions(func() {
		main()
	},
		commandline.WithCommand(func(cmd *cobra.Command) {
			cmd.SetArgs([]string{"--version"})
			cmd.SetOut(&o)
		}),
		commandline.WithExit(func(code int) {
			retcode = &code
		}),
	)

	assert.Equal(t, retcode, (*int)(nil))
	assert.Equal(t, o.String(), fmt.Sprintf("%s version %s\n",
		metadata.Name, metadata.Version))
}

func withOptions(fn func(), newOpts ...commandline.Option) {
	old := cmd.Options
	cmd.Options = newOpts
	defer func() {
		cmd.Options = old
	}()
	fn()
}
