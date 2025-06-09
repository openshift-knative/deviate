package git

import (
	"github.com/openshift-knative/deviate/pkg/files"
)

type Checkout interface {
	As(branch string) error
	OntoWorkspace(filters files.Filters) error
}
