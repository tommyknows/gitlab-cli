package gitlab

import (
	"context"
	"os"

	"github.com/go-git/go-git/plumbing/transport"
	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
	"github.com/tommyknows/gitlab-cli/pkg/log"
)

// Clone the ProjectNodes into the current directory, mirroring the upstream structure.
// The rootGroup is needed in order to strip away the leading path from the full Project
// Paths. If skipRoot is true and the root namespace matches the node's full path, the
// node will be skipped.
func Clone(root Namespace, skipRoot bool, auth transport.AuthMethod) (ContextVisitor, error) {

	return func(ctx context.Context, n ProjectNode) error {
		switch n := n.(type) {
		case *Project:
			log.Debugf("cloning %s to ./%s", n.Name(), Namespace(n.FullPath()).relative(root))

			_, err := git.PlainCloneContext(ctx, n.FullPath().relative(root).String(), false, &git.CloneOptions{
				URL:  n.gp.HTTPURLToRepo,
				Auth: auth,
				//URL:  n.gp.SSHURLToRepo, // TODO: maybe make HTTP possible too?
				//Auth: sshAuth,
			})
			if err != nil {
				return errors.Wrapf(err, "could not clone project %v", n.Name())
			}

			log.Debugf("cloned %v to ./%v", n.Name(), Namespace(n.FullPath()).relative(root))

		case *Group:
			if skipRoot && n.FullPath() == root {
				break
			}

			log.Debugf("creating folder %v for group %v\n", n.FullPath().relative(root), n.Name())
			if err := os.MkdirAll(n.FullPath().relative(root).String(), 0700); err != nil {
				return errors.Wrapf(err, "could not create folder for group %v", n.Name())
			}
			log.Debugf("created folder %v for group %v", n.FullPath().relative(root), n.Name())
		}
		return nil
	}, nil
}
