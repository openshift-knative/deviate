package sync

import (
	"os"
	"path"
	"strings"

	"github.com/openshift-knative/deviate/pkg/errors"
	pkgfiles "github.com/openshift-knative/deviate/pkg/files"
	"github.com/openshift-knative/deviate/pkg/log/color"
	"github.com/openshift-knative/deviate/pkg/sh"
)

func (o Operation) applyPatches() error {
	o.Println("- Apply patches if present")
	patchesDir := path.Join(o.Path, "openshift", "patches")
	files, err := os.ReadDir(patchesDir)
	if err != nil {
		o.Println("-- No patches found")
		return nil //nolint:nilerr
	}
	o.Printf("-- Found %d patche(s)\n", len(files))
	for _, file := range files {
		if !file.Type().IsRegular() || !strings.HasSuffix(file.Name(), ".patch") {
			continue
		}
		filePath := path.Join(patchesDir, file.Name())
		o.Printf("-- Applying %s\n", color.Blue(filePath))

		// TODO: Consider rewriting this to Go native code instead shell invocation.
		err = pkgfiles.WithinDirectory(o.Path, func() error {
			return errors.Wrap(sh.Run("git", "apply", filePath),
				ErrSyncFailed)
		})
		if err != nil {
			return errors.Wrap(err, ErrSyncFailed)
		}
	}

	return runSteps([]step{
		o.commitChanges(":fire: Apply carried patches"),
	})
}
