package sync

import (
	"bytes"
	"os"
	"path"
	"text/template"
	"time"

	"github.com/openshift-knative/deviate/pkg/config/git"
	"github.com/openshift-knative/deviate/pkg/errors"
)

func (o Operation) triggerCI() error {
	return triggerCI{o}.run()
}

// triggerCIMessageData holds the data for the TriggerCI message template.
type triggerCIMessageData struct {
	ReleaseBranch string
	MainBranch    string
}

func (o Operation) triggerCIMessage() string {
	tmplData := triggerCIMessageData{
		ReleaseBranch: o.Config.Branches.ReleaseNext,
		MainBranch:    o.Config.Branches.Main,
	}

	t, err := template.New("triggerCIMessage").Parse(o.Config.Messages.TriggerCI)
	if err != nil {
		o.Printf("Error parsing TriggerCI message template: %v. Using raw template string.", err)
		return o.Config.Messages.TriggerCI // Return raw template on parsing error
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, tmplData); err != nil {
		o.Printf("Error executing TriggerCI message template: %v. Using raw template string.", err)
		return o.Config.Messages.TriggerCI // Return raw template on execution error
	}

	return buf.String()
}

func (o Operation) triggerCIBody() string {
	tmplData := triggerCIMessageData{
		ReleaseBranch: o.Config.Branches.ReleaseNext,
		MainBranch:    o.Config.Branches.Main,
	}

	t, err := template.New("triggerCIBody").Parse(o.Config.Messages.TriggerCIBody)
	if err != nil {
		o.Printf("Error parsing TriggerCIBody message template: %v. Using raw template string.", err)
		return o.Config.Messages.TriggerCIBody // Return raw template on parsing error
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, tmplData); err != nil {
		o.Printf("Error executing TriggerCIBody message template: %v. Using raw template string.", err)
		return o.Config.Messages.TriggerCIBody // Return raw template on execution error
	}

	return buf.String()
}

type triggerCI struct {
	Operation
}

func (c triggerCI) run() error {
	c.Println("Trigger CI")
	return runSteps([]step{
		c.checkout,
		c.addChange,
		c.commitChanges(c.triggerCIMessage()),
		c.pushBranch(c.Config.Branches.SynchCI + c.Config.Branches.ReleaseNext),
	})
}

func (c triggerCI) checkout() error {
	remote := git.Remote{
		Name: "downstream",
		URL:  c.Config.Downstream,
	}
	err := c.Repository.Checkout(remote, c.Config.Branches.ReleaseNext).
		As(c.Config.Branches.SynchCI + c.Config.Branches.ReleaseNext)
	return errors.Wrap(err, ErrSyncFailed)
}

func (c triggerCI) addChange() error {
	filePath := path.Join(c.Project.Path, "ci")
	content := time.Now().Format(time.RFC3339)
	const fileReadableToOwnerPerm = 0o600
	err := os.WriteFile(filePath, []byte(content), fileReadableToOwnerPerm)
	return errors.Wrap(err, ErrSyncFailed)
}
