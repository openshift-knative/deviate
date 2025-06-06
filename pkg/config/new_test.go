package config_test

import (
	_ "embed"
	"os"
	"path"
	"testing"

	"github.com/openshift-knative/deviate/pkg/config"
	"github.com/openshift-knative/deviate/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed example_config.yaml
var configYaml string

func TestNew(t *testing.T) {
	tmp := t.TempDir()
	configPath := path.Join(tmp, ".deviate.yaml")
	if err := os.WriteFile(configPath, []byte(configYaml), 0o600); err != nil {
		require.NoError(t, err)
	}
	project := config.Project{
		Path:       tmp,
		ConfigPath: configPath,
	}
	cfg, err := config.New(project, log.TestingLogger{T: t}, noopInformer{})
	require.NoError(t, err)
	assert.True(t, cfg.DockerfileGen.Skip)
	assert.Equal(t, []string{"eventing"}, cfg.DockerfileGen.ImagesFromRepositories)
}

type noopInformer struct{}

func (n noopInformer) Remote(name string) (string, error) {
	return name, nil
}
