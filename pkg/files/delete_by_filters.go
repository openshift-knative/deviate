package files

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift-knative/deviate/pkg/errors"
)

// ErrCantDeleteFiles when cannot delete files.
var ErrCantDeleteFiles = errors.New("cannot delete files")

// DeleteFiles will delete all matching files starting at given root directory.
func (f Filters) DeleteFiles(root string) error {
	matcher := f.Matcher()
	err := filepath.WalkDir(root, func(pth string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if de.IsDir() {
			// continue
			return nil
		}
		pth = filepath.ToSlash(pth)
		relPath := strings.TrimPrefix(pth, root+"/")
		if matcher.Matches(relPath) {
			return os.Remove(pth)
		}
		return nil
	})
	return errors.Wrap(err, ErrCantDeleteFiles)
}
