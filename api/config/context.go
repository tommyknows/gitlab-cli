package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
	"github.com/tommyknows/gitlab-cli/pkg/gitlab"
	"github.com/tommyknows/gitlab-cli/pkg/log"
	gogitlab "github.com/xanzy/go-gitlab"
)

type ErrInvalidContext struct {
	name string
}

func (e ErrInvalidContext) Error() string {
	return fmt.Sprintf("invalid context %q does not exist", e.name)
}

// GetGetCurrentContext returns the right Context depending on
// the configuration. It considers the (filesystem-)local git
// repository and tries to find an instance config that relates
// to the git remote. If that exist, it constructs a new, temporary
// context that points to that instance with the correct group set.
func (c *Config) GetCurrentContext() (*Context, error) {
	if c.PreferConfigContext || c.useConfigContext {
		log.Debugf("preferring config context")
		return c.getCurrentConfigContext()
	}

	repo, err := git.PlainOpen(localPath)
	if repo != nil && err != git.ErrRepositoryNotExists {
		log.Debugf("currently in git repo, creating git context")
		return c.newGitRepoContext(repo)
	}

	log.Debugf("not in git repo, using default context")
	return c.getCurrentConfigContext()
}

func (c *Context) Instance() *InstanceConfig {
	return c.instanceConfig
}

// GitlabClient creates a Gitlab Client from the given context
func (c *Context) GitlabClient() (*gitlab.Client, error) {
	isUser := false
	space := c.Group
	if space == "" && c.User != "" {
		isUser = true
		space = c.User
	}

	cl, err := gogitlab.NewClient(
		c.Instance().Authentication.Token,
		gogitlab.WithBaseURL(c.Instance().apiURL()),
	)
	if err != nil {
		return nil, err
	}

	return gitlab.New(cl, space, isUser), nil
}

// WithGroup returns a copy of the context, with the new group set
func (c *Context) WithGroup(group string) *Context {
	return &Context{
		Group:          group,
		User:           c.User,
		InstanceName:   c.InstanceName,
		instanceConfig: c.instanceConfig,
	}
}

const (
	localPath = "."
	origin    = "origin"
)

// newGitRepoContext creates a new context out of the local git repository.
// It checks all "origin" remote URLs and checks if there's an URL where
// the host matches a known host from the context, and then creates a temporary
// Context that is pointing to that instance and the right group.
func (c *Config) newGitRepoContext(repo *git.Repository) (*Context, error) {
	remote, err := repo.Remote(origin)
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate context from git repository")
	}

	gitRemoteURL := remote.Config().URLs[0]

	repoURL, err := parseGitURL(gitRemoteURL)
	if err != nil {
		log.Infof("instance %s is not a valid URL: %v", remote.String(), err)
	}

	log.Debugf("found repo remote: %v", repoURL)

	for instName, instCfg := range c.Instances {
		if repoURL.Host == instCfg.url.Host {
			log.Debugf("repo URL matches Instance URL %q, creating context with Group %v", repoURL.Host, repoURL.Path)
			return &Context{
				InstanceName:   instName,
				Group:          repoURL.Path,
				instanceConfig: instCfg,
			}, nil
		}
	}

	return nil, errors.Errorf("no instance match found for git repository %q", repoURL)
}

const (
	// these names are not really correct, but I don't know
	// what they're actually called...TODO
	gitSSHImplicit = "git@"
	gitSSHExplicit = "ssh://"

	gitHTTPS = "https://"

	gitSuffix = ".git"
)

// parseGitURL parses a git remote and returns a URL to it, if it is valid.
func parseGitURL(gitURL string) (*url.URL, error) {
	gitURL = strings.TrimPrefix(gitURL, gitSSHExplicit)
	gitURL = strings.TrimSuffix(gitURL, gitSuffix)
	switch {
	case strings.HasPrefix(gitURL, gitSSHImplicit):
		gitURL = strings.TrimPrefix(gitURL, gitSSHImplicit)
		gitURL = strings.Replace(gitURL, ":", "/", 1)
		gitURL = "https://" + gitURL

	case strings.HasPrefix(gitURL, gitHTTPS):
		// nothing to do

	default:
		return nil, errors.Errorf("unknown git remote URL specification: %q", gitURL)
	}

	return url.Parse(gitURL)
}

func (c *Config) getCurrentConfigContext() (*Context, error) {
	ctx, ok := c.Contexts[c.CurrentContext]
	if !ok {
		return nil, ErrInvalidContext{c.CurrentContext}
	}
	ctx.instanceConfig, ok = c.Instances[ctx.InstanceName]
	if !ok {
		// TODO: custom error?
		return nil, ErrInvalidContext{c.CurrentContext}
	}

	log.Debugf("Using context %q", c.CurrentContext)
	return ctx, nil
}
