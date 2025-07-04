package sync

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/openshift-knative/deviate/pkg/config/git"
	"github.com/openshift-knative/deviate/pkg/errors"
	"github.com/openshift-knative/deviate/pkg/log/color"
	"github.com/openshift-knative/deviate/pkg/state"
)

func (o Operation) mirrorRelease(rel release) error {
	return runSteps([]step{
		o.createNewRelease(rel),
		o.addForkFiles(rel),
		o.applyPatches,
		o.switchToMain,
		o.pushRelease(rel),
	})
}

func (o Operation) createNewRelease(rel release) step {
	o.Printf("- Creating new release: %s\n", color.Blue(rel.String()))
	upstream := git.Remote{Name: "upstream", URL: o.Upstream}
	cnr := createNewRelease{State: o.State, rel: rel, remote: upstream}
	return cnr.step
}

func (o Operation) pushRelease(rel release) step {
	return func() error {
		o.Printf("- Publishing release: %s\n", color.Blue(rel.String()))
		branch, err := rel.Name(o.ReleaseTemplates.Downstream)
		if err != nil {
			return errors.Wrap(err, ErrSyncFailed)
		}
		pr := push{State: o.State, branch: branch}
		return runSteps(pr.steps())
	}
}

type createNewRelease struct {
	state.State
	rel    release
	remote git.Remote
}

func (r createNewRelease) step() error {
	upstreamBranch, err := r.rel.Name(r.ReleaseTemplates.Upstream)
	if err != nil {
		return errors.Wrap(err, ErrSyncFailed)
	}
	downstreamBranch, err := r.rel.Name(r.ReleaseTemplates.Downstream)
	if err != nil {
		return errors.Wrap(err, ErrSyncFailed)
	}
	return runSteps([]step{
		r.fetch,
		r.checkoutAsNewRelease(upstreamBranch, downstreamBranch),
	})
}

func (r createNewRelease) fetch() error {
	return errors.Wrap(r.Fetch(r.remote), ErrSyncFailed)
}

func (r createNewRelease) checkoutAsNewRelease(upstreamBranch, downstreamBranch string) step {
	return func() error {
		return errors.Wrap(
			r.Repository.Checkout(r.remote, upstreamBranch).As(downstreamBranch),
			ErrSyncFailed)
	}
}

type push struct {
	state.State
	branch     string
	skipDelete bool
}

func (p push) steps() []step {
	st := []step{p.push}
	if !p.skipDelete {
		st = append(st, p.delete)
	}
	return st
}

func (p push) push() error {
	refName := plumbing.NewBranchReferenceName(p.branch)
	return publish(p.State, "release push", refName)
}

func (p push) delete() error {
	return errors.Wrap(p.DeleteBranch(p.branch), ErrSyncFailed)
}

func publish(state state.State, title string, refName plumbing.ReferenceName) error {
	if state.DryRun {
		state.Println(color.Yellow(fmt.Sprintf(
			"- Skipping %s, because of dry run", title)))
		return nil
	}
	remote := git.Remote{
		Name: "downstream",
		URL:  state.Downstream,
	}
	return errors.Wrap(state.Push(remote, refName), ErrSyncFailed)
}
