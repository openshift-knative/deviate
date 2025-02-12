package sync

import (
	"github.com/openshift-knative/deviate/pkg/config/git"
	"github.com/openshift-knative/deviate/pkg/errors"
)

func (o Operation) resetReleaseNext() error {
	o.Printf("Reset %s branch to upstream/%s.\n",
		o.Config.Branches.ReleaseNext, o.Config.Branches.Main)
	remote := git.Remote{
		Name: "upstream",
		URL:  o.Config.Upstream,
	}
	if err := o.Repository.Checkout(remote, o.Config.Branches.Main).
		As(o.Config.Branches.ReleaseNext); err != nil {
		return errors.Wrap(err, ErrSyncFailed)
	}

	return nil
}
