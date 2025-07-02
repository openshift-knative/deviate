package sync

import (
	gitv5 "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/openshift-knative/deviate/pkg/config/git"
	"github.com/openshift-knative/deviate/pkg/errors"
	"github.com/openshift-knative/deviate/pkg/log/color"
	"github.com/openshift-knative/deviate/pkg/state"
)

// ErrSyncFailed when the sync failed.
var ErrSyncFailed = errors.New("sync failed")

// Operation performs sync - the upstream synchronization.
type Operation struct {
	state.State
}

func (o Operation) Run() error {
	if err := runSteps([]step{
		o.mirrorReleases,
		o.syncTags,
		o.syncReleaseNext,
		o.triggerCI,
		o.createSyncReleaseNextPR,
	}); err != nil {
		return err
	}
	return o.switchToMain()
}

func (o Operation) switchToMain() error {
	downstream := git.Remote{Name: "downstream", URL: o.Downstream}
	err := o.Fetch(downstream)
	if err != nil {
		return errors.Wrap(err, ErrSyncFailed)
	}
	return errors.Wrap(
		o.Repository.Checkout(downstream, o.Config.Main).As(o.Main),
		ErrSyncFailed,
	)
}

func (o Operation) commitChanges(message string, onCommit ...step) step {
	return func() error {
		o.Println("- Committing changes:", message)
		commit, err := o.CommitChanges(message)
		if err != nil {
			if errors.Is(err, gitv5.NoErrAlreadyUpToDate) {
				o.Println("-- No changes to commit")
				return nil
			}
			return errors.Wrap(err, ErrSyncFailed)
		}
		stats, err := commit.StatsContext(o.Context)
		if err == nil {
			o.Printf("-- Statistics:\n%s\n", stats)
		}
		err = errors.Wrap(err, ErrSyncFailed)
		for _, st := range onCommit {
			if serr := st(); serr != nil {
				err = errors.Join(err, serr)
				break
			}
		}
		return err
	}
}

func (o Operation) syncTags() error {
	refName := plumbing.NewTagReferenceName(o.RefSpec)
	o.Println("- Syncing tags:", color.Blue(refName))
	return publish(o.State, "tag synchronization", refName)
}
