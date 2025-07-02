package github

import (
	"context"
	"os"

	"github.com/cli/cli/v2/pkg/cmd/factory"
	ghroot "github.com/cli/cli/v2/pkg/cmd/root"
	"github.com/openshift-knative/deviate/pkg/errors"
	"github.com/openshift-knative/deviate/pkg/files"
	"github.com/openshift-knative/deviate/pkg/metadata"
)

// ErrClientFailed when client operations has failed.
var ErrClientFailed = errors.New("client failed")

// NewClient creates new client.
func NewClient(args ...string) Client {
	return Client{Args: args}
}

// Client a client for Github CLI.
type Client struct {
	Args         []string
	DisableColor bool
	ProjectDir   string
}

// Execute a Github client CLI command.
func (c Client) Execute(ctx context.Context) (bytes []byte, err error) {
	buildVersion := metadata.Version
	cmdFactory := factory.New(buildVersion)
	cmd, gerr := ghroot.NewCmdRoot(cmdFactory, buildVersion, "-")
	if gerr != nil {
		err = errors.Join(err, errors.Wrap(gerr, ErrClientFailed))
		return bytes, err
	}
	cmd.SetArgs(c.Args)
	tmpf, terr := os.CreateTemp("", "gh-")
	if terr != nil {
		err = errors.Join(err, errors.Wrap(terr, ErrClientFailed))
		return bytes, err
	}
	defer func() {
		rmerr := os.Remove(tmpf.Name())
		err = errors.Join(err, errors.Wrap(rmerr, ErrClientFailed))
	}()
	cmdFactory.IOStreams.Out = tmpf
	cmdFactory.IOStreams.ErrOut = os.Stderr
	if c.DisableColor {
		cmdFactory.IOStreams.SetColorEnabled(false)
	}
	runner := func() error {
		return errors.Wrap(cmd.ExecuteContext(ctx), ErrClientFailed)
	}
	if c.ProjectDir != "" {
		previousRunner := runner
		runner = func() error {
			return errors.Wrap(files.WithinDirectory(c.ProjectDir, previousRunner),
				ErrClientFailed)
		}
	}
	err = errors.Join(err, runner())
	bytes, ferr := os.ReadFile(tmpf.Name())
	if ferr != nil {
		err = errors.Join(err, errors.Wrap(ferr, ErrClientFailed))
	}
	return bytes, err
}
