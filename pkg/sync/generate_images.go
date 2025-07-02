package sync

import (
	"os"
	"path"
	"strings"

	"github.com/openshift-knative/deviate/pkg/errors"
	"github.com/openshift-knative/hack/pkg/dockerfilegen"
	"sigs.k8s.io/yaml"
)

func (o Operation) generateImages(rel release) step {
	return func() error {
		params := o.DockerfileGen
		if params.Skip {
			o.Println("- Skipping image generation")
			return nil
		}
		o.Println("- Generating images")
		var closer func()
		var err error
		params.ProjectFilePath, closer, err = tempProjectFile(o, rel)
		defer closer()
		if err != nil {
			return err
		}
		return dockerfilegen.GenerateDockerfiles(params.Params)
	}
}

func tempProjectFile(o Operation, rel release) (string, func(), error) {
	closer := func() {}
	f, err := os.CreateTemp("", "project-*.yaml")
	if err != nil {
		return "", closer, errors.Wrap(err, ErrSyncFailed)
	}
	if err = f.Close(); err != nil {
		return "", closer, errors.Wrap(err, ErrSyncFailed)
	}
	closer = func() {
		_ = os.Remove(f.Name())
	}
	data := map[string]any{
		"project": map[string]any{
			"tag":         rel.Tag(),
			"imagePrefix": upstreamToName(o.Upstream),
		},
	}
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return "", closer, errors.Wrap(err, ErrSyncFailed)
	}
	const allowRead = 0o644
	if err = os.WriteFile(f.Name(), bytes, allowRead); err != nil {
		return "", closer, errors.Wrap(err, ErrSyncFailed)
	}
	o.Println("- Project file written:", f.Name())
	return f.Name(), closer, nil
}

func upstreamToName(upstream string) string {
	return strings.TrimSuffix(path.Base(upstream), ".git")
}
