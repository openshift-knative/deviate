package config

import (
	"github.com/openshift-knative/deviate/pkg/files"
	"github.com/openshift-knative/hack/pkg/dockerfilegen"
)

// newDefaults creates a new default configuration.
func newDefaults(project Project) Config {
	const (
		releaseTemplate = "release-{{ .Major }}.{{ .Minor }}"
		releaseSearch   = `^release-(\d+)\.(\d+)$`
	)
	return Config{
		DeleteFromUpstream: files.Filters{
			Include: []string{
				".github/workflows/knative-*.y?ml",
			},
		},
		CopyFromMidstream: files.Filters{
			Include: []string{"**"},
		},
		Branches: Branches{
			Main:          "main",
			ReleaseNext:   "release-next",
			CheckPrPrefix: "ci/",
			ReleaseTemplates: ReleaseTemplates{
				Upstream:   releaseTemplate,
				Downstream: releaseTemplate,
			},
			Searches: Searches{
				UpstreamReleases:   releaseSearch,
				DownstreamReleases: releaseSearch,
			},
		},
		Tags: Tags{
			RefSpec: "v*",
		},
		ResyncReleases: ResyncReleases{
			NumberOf: 6, //nolint:mnd
		},
		Messages: Messages{
			TriggerCI: ":robot: Synchronize branch `%s` to " +
				"`upstream/%s`",
			TriggerCIBody: "This automated PR is to make sure the " +
				"forked project's `%s` branch (forked upstream's `%s` branch) passes" +
				" a CI.",
			ApplyForkFiles:  ":open_file_folder: Apply fork specific files",
			ImagesGenerated: ":vhs: Images generated",
		},
		SyncLabels: []string{"kind/sync-fork-to-upstream"},
		DockerfileGen: DockerfileGen{
			Params: dockerfilegen.DefaultParams(project.Path),
		},
	}
}
