package sync

import (
	"github.com/openshift-knative/hack/pkg/dockerfilegen"
)

func (o Operation) generateImages() error {
	o.Println("- Generating images")

	p := dockerfilegen.DefaultParams(o.Path)
	p.ScanImports = true

	return dockerfilegen.GenerateDockerfiles(p)
}
