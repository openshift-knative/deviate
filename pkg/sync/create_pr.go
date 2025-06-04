package sync

import (
	"encoding/json"
	"fmt"

	"github.com/openshift-knative/deviate/pkg/errors"
	"github.com/openshift-knative/deviate/pkg/git"
	"github.com/openshift-knative/deviate/pkg/github"
	"github.com/openshift-knative/deviate/pkg/log/color"
)

func (o Operation) createSyncReleaseNextPR() error {
	branches := o.Config.Branches
	return o.createPR(
		o.triggerCIMessage(),
		o.triggerCIBody(),
		branches.ReleaseNext,
		branches.SyncCi+branches.ReleaseNext,
	)
}

func (o Operation) createPR(title, body, base, head string) error {
	o.Println("Create a sync PR for:", color.Blue(base))
	pr := createPR{Operation: o, title: title, body: body, base: base, head: head}
	url, err := pr.active()
	if err != nil {
		if errors.Is(err, errPrNotFound) {
			return pr.open()
		}
		return err
	}

	o.Printf("Thr PR for %s is already active: %s\n",
		color.Blue(base), color.Yellow(*url))
	return nil
}

type createPR struct {
	Operation
	title string
	body  string
	base  string
	head  string
}

var errPrNotFound = errors.New("PR not found")

func (c createPR) active() (*string, error) {
	repo, err := c.repository()
	if err != nil {
		return nil, errors.Wrap(err, ErrSyncFailed)
	}
	args := []string{
		"pr", "list",
		"--repo", repo,
		"--state", "open",
		"--author", "@me",
		"--search", c.title,
		"--json", "url",
	}
	for _, label := range c.Config.SyncLabels {
		args = append(args, "--label", label)
	}
	cl := github.NewClient(args...)
	cl.DisableColor = true
	cl.ProjectDir = c.Project.Path
	buff, err := cl.Execute(c.Context)
	if err != nil {
		return nil, errors.Wrap(err, ErrSyncFailed)
	}
	un := make([]map[string]interface{}, 0)
	err = json.Unmarshal(buff, &un)
	if err != nil {
		return nil, errors.Wrap(err, ErrSyncFailed)
	}

	if len(un) > 0 {
		u := fmt.Sprintf("%s", un[0]["url"])
		return &u, nil
	}
	return nil, errPrNotFound
}

func (c createPR) open() error {
	repo, err := c.repository()
	if err != nil {
		return errors.Wrap(err, ErrSyncFailed)
	}
	args := []string{
		"pr", "create",
		"--repo", repo,
		"--body", c.body,
		"--title", c.title,
		"--base", c.base,
		"--head", c.head,
	}
	for _, label := range c.Config.SyncLabels {
		args = append(args, "--label", label)
	}
	cl := github.NewClient(args...)
	cl.ProjectDir = c.Project.Path
	_, err = cl.Execute(c.Context)

	return errors.Wrap(err, ErrSyncFailed)
}

func (c createPR) repository() (string, error) {
	addr, err := git.ParseAddress(c.Config.Downstream)
	if err != nil {
		return "", errors.Wrap(err, ErrSyncFailed)
	}
	return addr.Path, nil
}
