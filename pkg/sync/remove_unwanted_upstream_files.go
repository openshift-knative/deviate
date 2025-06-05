package sync

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/openshift-knative/deviate/pkg/errors"
)

func (o Operation) removeUnwantedUpstreamFiles() error {
	o.Println("- Remove unwanted upstream files")

	return errors.Wrap(filepath.WalkDir(o.State.Project.Path, func(pth string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if o.Config.DeleteFromUpstream.Matches(pth) {
			fp := path.Join(o.State.Project.Path, pth)
			return os.RemoveAll(fp)
		}
		return nil
	}), ErrSyncFailed)
}
