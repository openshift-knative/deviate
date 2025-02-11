package cmd_test

import (
	"testing"

	"github.com/openshift-knative/deviate/internal/cmd"
	"gotest.tools/v3/assert"
)

func TestRoot(t *testing.T) {
	c := new(cmd.App).Command()

	assert.Equal(t, len(c.Commands()), 1)
	assert.Equal(t, c.Name(), "deviate")
	assert.Equal(t, c.Commands()[0].Name(), "sync")
}
