package sync

import (
	"github.com/openshift-knative/deviate/pkg/errors"
)

func (o Operation) removeUnwantedUpstreamFiles() error {
	o.Println("- Remove unwanted upstream files")

	return errors.Wrap(
		o.Config.DeleteFromUpstream.DeleteFiles(o.State.Project.Path),
		ErrSyncFailed)
}
