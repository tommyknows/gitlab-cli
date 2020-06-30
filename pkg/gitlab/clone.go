package gitlab

import (
	"context"
	"os"

	"github.com/go-git/go-git/plumbing/transport"
	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
	"github.com/tommyknows/gitlab-cli/pkg/log"
	"github.com/xanzy/go-gitlab"
)

// Clone the ProjectNodes into the current directory, mirroring the upstream structure.
// The rootGroup is needed in order to strip away the leading path from the full Project
// Paths. If skipRoot is true and the root namespace matches the node's full path, the
// node will be skipped.
func Clone(root Namespace, skipRoot bool, auth transport.AuthMethod) (ContextVisitor, error) {
	return func(ctx context.Context, n ProjectNode) error {
		switch n := n.(type) {
		case *Project:
			path := Namespace(n.FullPath()).relative(root)

			repo, err := git.PlainOpen(path.String())
			if repo != nil && err != git.ErrRepositoryNotExists {
				log.Debugf("Pulling %s in %s", n.Name(), path)
				if err := pull(repo, n.gp, auth); err != nil {
					return errors.Wrapf(err, "could not pull existing repo at %v", path.String())
				}
				return nil
			}

			log.Debugf("cloning %s to ./%s", n.Name(), path)

			_, err = git.PlainCloneContext(ctx, n.FullPath().relative(root).String(), false, &git.CloneOptions{
				URL:  n.gp.HTTPURLToRepo,
				Auth: auth,
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

func pull(repo *git.Repository, proj *gitlab.Project, auth transport.AuthMethod) error {
	w, err := repo.Worktree()
	if err != nil {
		return errors.Wrapf(err, "could not get worktree from git repository")
	}

	remote, err := determineRemote(repo, proj)
	if err != nil {
		return errors.Wrapf(err, "could not get remote")
	}

	err = w.Pull(&git.PullOptions{
		RemoteName: remote,
		Auth:       auth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}
	return nil
}

func determineRemote(repo *git.Repository, proj *gitlab.Project) (string, error) {
	remotes, err := repo.Remotes()
	if err != nil {
		return "", errors.Wrapf(err, "could not get remotes")
	}

	for _, remote := range remotes {
		for _, url := range remote.Config().URLs {
			if url == proj.SSHURLToRepo || url == proj.HTTPURLToRepo {
				return remote.Config().Name, nil
			}
		}
	}

	// TODO: configure and pull it?
	return "", errors.New("could not determine remote, maybe not configured")
}
