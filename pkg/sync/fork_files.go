package sync

import (
	"github.com/openshift-knative/deviate/pkg/config/git"
	"github.com/openshift-knative/deviate/pkg/errors"
)

func (o Operation) addForkFiles(rel release) step {
	return multiStep([]step{
		o.removeGithubWorkflows,
		o.unpackForkOntoWorkspace,
		o.commitChanges(o.Config.Messages.ApplyForkFiles),
		o.generateImages(rel),
		o.commitChanges(o.Config.Messages.ImagesGenerated),
	}).runSteps
}

func (o Operation) unpackForkOntoWorkspace() error {
	o.Println("- Add fork's files")
	upstream := git.Remote{Name: "upstream", URL: o.Config.Upstream}
	err := o.Repository.Checkout(upstream, o.Config.Branches.Main).
		OntoWorkspace()
	return errors.Wrap(err, ErrSyncFailed)
}
