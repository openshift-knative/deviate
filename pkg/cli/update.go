package cli

import (
	"errors"

	"github.com/openshift-knative/deviate/pkg/config"
	pkgerrors "github.com/openshift-knative/deviate/pkg/errors"
	"github.com/openshift-knative/deviate/pkg/git"
	"github.com/openshift-knative/deviate/pkg/log"
	"github.com/openshift-knative/deviate/pkg/log/color"
	"github.com/openshift-knative/deviate/pkg/state"
	"github.com/openshift-knative/deviate/pkg/sync"
)

// ErrConfigurationIsInvalid when configuration is invalid.
var ErrConfigurationIsInvalid = errors.New("configuration is invalid")

// Sync will perform synchronization to upstream branches.
func Sync(logger log.Logger, projectFactory func() config.Project) error {
	color.SetupMode()
	st := state.New(log.LabeledLogger{
		Label: color.Green("[deviate:sync]"),
		Logger: log.TimedLogger{
			Logger: logger,
		},
	})
	defer st.Close()
	project, err := git.NewProject(projectFactory(), st)
	if err != nil {
		return pkgerrors.Wrap(err, ErrConfigurationIsInvalid)
	}
	cfg, err := config.New(project.Project, st, project.Repository())
	if err != nil {
		return pkgerrors.Wrap(err, ErrConfigurationIsInvalid)
	}
	st.Project = &project.Project
	st.Repository = project.Repository()
	st.Config = &cfg
	op := sync.Operation{State: st}
	return pkgerrors.Wrap(op.Run(), sync.ErrSyncFailed)
}
