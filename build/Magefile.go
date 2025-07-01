//go:build mage
// +build mage

package main

import (
	"github.com/openshift-knative/deviate/pkg/metadata"

	// mage:import
	"knative.dev/toolbox/magetasks"
	"knative.dev/toolbox/magetasks/config"
	"knative.dev/toolbox/magetasks/pkg/artifact"
	"knative.dev/toolbox/magetasks/pkg/artifact/platform"
	"knative.dev/toolbox/magetasks/pkg/checks"
	"knative.dev/toolbox/magetasks/pkg/git"
)

// Default target is set to binary.
//
//goland:noinspection GoUnusedGlobalVariable
var Default = magetasks.Build //nolint:deadcode,gochecknoglobals

func init() { //nolint:gochecknoinits
	bin := artifact.Binary{
		Metadata:  config.Metadata{Name: "deviate"},
		Platforms: []artifact.Platform{{OS: platform.Linux, Architecture: platform.AMD64}},
	}

	magetasks.Configure(config.Config{
		Version: &config.Version{
			Path:     metadata.VersionPath(),
			Resolver: git.NewVersionResolver(),
		},
		Artifacts: []config.Artifact{bin},
		Checks:    []config.Task{checks.GolangCiLint(withVersion("v2.0.2"))},
	})
}

func withVersion(v string) checks.GolangCiLintParam {
	return func(opts *checks.GolangCiLintOptions) {
		opts.Version = v
	}
}
