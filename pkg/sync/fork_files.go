package sync

import (
	"github.com/openshift-knative/deviate/pkg/config/git"
	"github.com/openshift-knative/deviate/pkg/errors"
)

func (o Operation) addForkFiles() error {
	return runSteps([]step{
		o.removeGithubWorkflows,
		o.unpackForkOntoWorkspace,
		o.commitChanges(o.Config.Messages.ApplyForkFiles),
		o.generateImages,
		o.commitChanges(o.Config.Messages.ImagesGenerated),
	})
}

func (o Operation) unpackForkOntoWorkspace() error {
	o.Println("- Add fork's files")
	upstream := git.Remote{Name: "upstream", URL: o.Config.Upstream}
	err := o.Repository.Checkout(upstream, o.Config.Branches.Main).
		OntoWorkspace()
	return errors.Wrap(err, ErrSyncFailed)
}
