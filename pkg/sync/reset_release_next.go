package sync

import (
	"github.com/openshift-knative/deviate/pkg/config/git"
	"github.com/openshift-knative/deviate/pkg/errors"
)

func (o Operation) resetReleaseNext() error {
	o.Printf("Reset %s branch to upstream/%s.\n",
		o.ReleaseNext, o.Main)
	remote := git.Remote{
		Name: "upstream",
		URL:  o.Upstream,
	}
	if err := o.Repository.Checkout(remote, o.Config.Branches.Main).
		As(o.ReleaseNext); err != nil {
		return errors.Wrap(err, ErrSyncFailed)
	}

	return nil
}
