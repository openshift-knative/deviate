package git

import (
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/mitchellh/go-homedir"
	"github.com/openshift-knative/deviate/pkg/config/git"
	"github.com/openshift-knative/deviate/pkg/errors"
	"github.com/openshift-knative/deviate/pkg/url"
	sshagent "github.com/xanzy/ssh-agent"
)

func authentication(remote git.Remote) (ssh.AuthMethod, error) { //nolint:ireturn
	if url.IsHTTP(remote.URL) {
		// The auth isn't required for getting HTTP remote from GH
		return nil, nil //nolint:nilnil
	}
	if sshagent.Available() {
		user := ""
		if addr, err := ParseAddress(remote.URL); err == nil {
			user = addr.User
		}
		auth, err := ssh.NewSSHAgentAuth(user)
		if err != nil {
			return nil, errors.Wrap(err, ErrRemoteOperationFailed)
		}
		return auth, nil
	}
	idRsa, err := homedir.Expand("~/.ssh/id_rsa")
	if err != nil {
		return nil, errors.Wrap(err, ErrRemoteOperationFailed)
	}
	auth, err := ssh.NewPublicKeysFromFile("git", idRsa, "")
	return auth, errors.Wrap(err, ErrRemoteOperationFailed)
}
