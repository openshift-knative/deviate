package sync

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/openshift-knative/deviate/pkg/config/git"
	"github.com/openshift-knative/deviate/pkg/errors"
	"github.com/openshift-knative/deviate/pkg/log/color"
)

func (o Operation) triggerCI() error {
	return triggerCI{o}.run()
}

func (o Operation) triggerCIMessage() string {
	return fmt.Sprintf(
		o.Config.Messages.TriggerCI,
		o.Config.Branches.ReleaseNext,
		o.Config.Branches.Main)
}

type triggerCI struct {
	Operation
}

func (c triggerCI) run() error {
	c.Println("Trigger CI")

	if c.Config.Branches.SkipCheckPr {
		c.Println(color.Yellow("Skipping CI Check PRs trigger"))
		return nil
	}

	return runSteps([]step{
		c.checkout,
		c.addChange,
		c.commitChanges(c.triggerCIMessage()),
		c.pushBranch(c.Config.Branches.CheckPrPrefix + c.Config.Branches.ReleaseNext),
	})
}

func (c triggerCI) checkout() error {
	remote := git.Remote{
		Name: "downstream",
		URL:  c.Config.Downstream,
	}
	err := c.Repository.Checkout(remote, c.Config.Branches.ReleaseNext).
		As(c.Config.Branches.CheckPrPrefix + c.Config.Branches.ReleaseNext)
	return errors.Wrap(err, ErrSyncFailed)
}

func (c triggerCI) addChange() error {
	filePath := path.Join(c.Project.Path, "ci")
	content := time.Now().Format(time.RFC3339)
	const fileReadableToOwnerPerm = 0o600
	err := os.WriteFile(filePath, []byte(content), fileReadableToOwnerPerm)
	return errors.Wrap(err, ErrSyncFailed)
}
