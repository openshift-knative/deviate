package config

import (
	"github.com/openshift-knative/deviate/pkg/config/git"
	"github.com/openshift-knative/deviate/pkg/log"
)

// New creates a new default configuration.
func New(
	project Project,
	log log.Logger,
	informer git.RemoteURLInformer,
) (Config, error) {
	c := newDefaults(project)
	err := c.load(project, log, informer)
	if err != nil {
		return Config{}, err
	}
	err = c.overrides()
	if err != nil {
		return Config{}, err
	}
	err = c.validate()
	if err != nil {
		return Config{}, err
	}
	return c, nil
}
