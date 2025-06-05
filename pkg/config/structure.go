package config

import (
	"github.com/openshift-knative/deviate/pkg/files"
	"github.com/openshift-knative/hack/pkg/dockerfilegen"
)

// Config for a deviate to operate.
type Config struct {
	Upstream           string               `json:"upstream"           valid:"required"`
	Downstream         string               `json:"downstream"         valid:"required"`
	DryRun             bool                 `json:"dryRun"`
	CopyFromMidstream  files.Filters        `json:"copyFromMidstream"  valid:"required"`
	DeleteFromUpstream files.Filters        `json:"deleteFromUpstream" valid:"required"`
	SyncLabels         []string             `json:"syncLabels"         valid:"required"`
	DockerfileGen      dockerfilegen.Params `json:"dockerfileGen"`
	ResyncReleases     `json:"resyncReleases"`
	Branches           `json:"branches"`
	Tags               `json:"tags"`
	Messages           `json:"messages"`
}

// ResyncReleases holds configuration for resyncing past releases.
type ResyncReleases struct {
	Enabled  bool `json:"enabled"`
	NumberOf int  `json:"numberOf"`
}

// Tags holds configuration for tags.
type Tags struct {
	Synchronize bool   `json:"synchronize"`
	RefSpec     string `json:"refSpec"     valid:"required"`
}

// Messages holds messages that are used to commit changes and create PRs.
type Messages struct {
	TriggerCI       string `json:"triggerCi"       valid:"required"`
	TriggerCIBody   string `json:"triggerCiBody"   valid:"required"`
	ApplyForkFiles  string `json:"applyForkFiles"  valid:"required"`
	ImagesGenerated string `json:"imagesGenerated" valid:"required"`
}

// Branches holds configuration for branches.
type Branches struct {
	Main             string `json:"main"             valid:"required"`
	ReleaseNext      string `json:"releaseNext"      valid:"required"`
	SyncCi           string `json:"syncCi"`
	ReleaseTemplates `json:"releaseTemplates"`
	Searches         `json:"searches"`
}

// ReleaseTemplates contains templates for release names.
type ReleaseTemplates struct {
	Upstream   string `json:"upstream"   valid:"required"`
	Downstream string `json:"downstream" valid:"required"`
}

// Searches contains regular expressions used to search for branches.
type Searches struct {
	UpstreamReleases   string `json:"upstreamReleases"   valid:"required"`
	DownstreamReleases string `json:"downstreamReleases" valid:"required"`
}
