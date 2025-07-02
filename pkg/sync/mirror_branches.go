package sync

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"text/template"

	"github.com/openshift-knative/deviate/pkg/config/git"
	"github.com/openshift-knative/deviate/pkg/errors"
	"github.com/openshift-knative/deviate/pkg/log/color"
)

func (o Operation) mirrorReleases() error {
	o.Println("Check if there's an upstream release we need " +
		"to mirror downstream")

	missing, err := o.findMissingDownstreamReleases()
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		o.Printf("Found missing releases: %s\n", color.Blue(fmt.Sprintf("%+q", missing)))
		for _, rel := range missing {
			err = o.mirrorRelease(rel)
			if err != nil {
				return err
			}
		}
	} else {
		o.Println("No missing releases found")
	}

	return o.resyncReleases(missing)
}

type release interface {
	String() string
	Name(tpl string) (string, error)
	less(o release) bool
	Tag() string
}

type stdRelease struct {
	Major, Minor int
}

func (r stdRelease) String() string {
	return strconv.Itoa(r.Major) + "." + strconv.Itoa(r.Minor)
}

func (r stdRelease) Name(tpl string) (string, error) {
	eng, err := template.New("release").Parse(tpl)
	if err != nil {
		return "", errors.Wrap(err, ErrSyncFailed)
	}
	var buff bytes.Buffer
	err = eng.Execute(&buff, r)
	if err != nil {
		return "", errors.Wrap(err, ErrSyncFailed)
	}
	return buff.String(), nil
}

func (r stdRelease) Tag() string {
	return "knative-v" + r.String()
}

func (r stdRelease) less(o release) bool {
	if so, ok := o.(stdRelease); ok {
		return r.Major < so.Major || (r.Major == so.Major && r.Minor < so.Minor)
	}
	return false
}

func (o Operation) findMissingDownstreamReleases() ([]release, error) {
	var upstreamReleases, downstreamReleases []release
	var err error
	downstreamReleases, err = o.listReleases(false)
	if err != nil {
		return nil, errors.Wrap(err, ErrSyncFailed)
	}
	upstreamReleases, err = o.listReleases(true)
	if err != nil {
		return nil, errors.Wrap(err, ErrSyncFailed)
	}

	missing := make([]release, 0, len(upstreamReleases))
	for _, candidate := range upstreamReleases {
		found := false
		for _, downstreamRelease := range downstreamReleases {
			if candidate == downstreamRelease {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, candidate)
		}
	}

	return missing, nil
}

func (o Operation) listReleases(upstream bool) ([]release, error) {
	url := o.Downstream
	re := regexp.MustCompile(o.DownstreamReleases)
	if upstream {
		url = o.Upstream
		re = regexp.MustCompile(o.UpstreamReleases)
	}

	refs, err := o.ListRemote(git.Remote{Name: "origin", URL: url})
	if err != nil {
		return nil, errors.Wrap(err, ErrSyncFailed)
	}

	releases := make([]release, 0)

	for _, ref := range refs {
		name := ref.Name()
		if name.IsBranch() {
			branch := name.Short()
			if matches := re.FindStringSubmatch(branch); matches != nil {
				version := stdRelease{atoi(matches[1]), atoi(matches[2])}
				releases = append(releases, version)
			}
		}
	}

	sort.Slice(releases, func(i, j int) bool {
		return releases[i].less(releases[j])
	})
	return releases, nil
}

func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}
