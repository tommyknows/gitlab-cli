/*
	Package gitlab serves as an abstraction over gitlab-go,
	providing easier access to things we need
*/
package gitlab

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/tommyknows/gitlab-cli/pkg/log"
	gl "github.com/xanzy/go-gitlab"
)

func New(c *gl.Client, namespace string) *Client {
	return &Client{c, namespace}
}

type Client struct {
	c         *gl.Client
	namespace string
}

func normalize(ns string) string {
	return strings.ToLower(strings.Trim(ns, "/"))
}

func extractNamespace(fullPath string) Namespace {
	fullPath = normalize(fullPath)
	lastSlash := strings.LastIndex(fullPath, "/")
	if lastSlash == -1 {
		return ""
	}

	return Namespace(fullPath[:lastSlash])
}

// GetProjects gets the project of the set namespace, returning the root of a Project-tree.
func (c *Client) GetProjects(ctx context.Context, includeArchived bool) (root ProjectNode, err error) {
	ns, resp, err := c.c.Namespaces.GetNamespace(url.QueryEscape(c.namespace), gl.WithContext(ctx))
	if err == nil {
		switch ns.Kind {
		case "group":
			return c.getGroup(ctx, c.namespace, includeArchived)
		case "user":
			return c.getUser(ctx, c.namespace, includeArchived)
		default:
			return nil, errors.New("unknown kind: " + ns.Kind)
		}
	}

	// some connection error - no internet connection for example.
	if resp == nil && err != nil {
		return nil, err
	}

	// it could be that the namespace really not exists.
	// it could also be that it is a project. In this case,
	// (and we don't really know), we try getting the project.
	if resp != nil && resp.StatusCode != http.StatusNotFound {
		return nil, err
	}

	tr := true
	p, resp, err := c.c.Projects.GetProject(c.namespace, &gl.GetProjectOptions{
		Statistics: &tr,
	}, gl.WithContext(ctx))
	if err == nil {
		return newProject(p), nil
	}

	if resp.StatusCode != http.StatusNotFound {
		return nil, err
	}
	// the error message is a lengthy string that includes the
	// URL of the request. We don't need it, as we know exactly
	// what it is, so we just create a custom error
	return nil, errors.New("no such namespace or project")
}

// getUser and getGroup have an extreme amount of duplicated code. Yet, I cannot find a simple
// solution to unify them without adding a ton of abstraction.
func (c *Client) getUser(ctx context.Context, user string, includeArchived bool) (root ProjectNode, err error) {
	getProjects := func(archived bool) ([]*gl.Project, error) {
		var projects []*gl.Project

		opts := &gl.ListProjectsOptions{
			Archived: &archived,
			ListOptions: gl.ListOptions{
				Page:    1,
				PerPage: 100, // this is the max value from gitlab
			},
		}
		for {
			select {
			case <-ctx.Done():
				return nil, context.Canceled
			default:
			}

			p, resp, err := c.c.Projects.ListUserProjects(user, opts, gl.WithContext(ctx))
			if err != nil {
				return nil, err
			}

			projects = append(projects, p...)

			// Exit the loop when we've seen all pages.
			if resp.CurrentPage >= resp.TotalPages {
				log.Debugf("got all results from the API")
				break
			}

			// update page number to fetch next page
			log.Debugf("getting next page from the API")
			opts.Page = resp.NextPage
		}

		return projects, nil
	}

	u, _, err := c.c.Users.ListUsers(&gl.ListUsersOptions{Username: &user})
	if err != nil {
		return nil, err
	}

	if len(u) == 0 {
		// TODO: Better message, maybe a custom error?
		return nil, errors.New("did not find user. this should not happen!")
	}

	if len(u) > 1 {
		log.Infof("found more than one user with the username %v, using first one: %v", user, u[0].Name)
	}

	usr := newUser(u[0])

	projects, err := getProjects(false)
	if err != nil {
		return nil, err
	}
	addSubProjects(usr, projects)

	if includeArchived {
		projects, err := getProjects(true)
		if err != nil {
			return nil, err
		}

		addSubProjects(usr, projects)
	}

	return usr, nil
}

func (c *Client) getGroup(ctx context.Context, group string, includeArchived bool) (root ProjectNode, err error) {
	getProjects := func(archived bool) ([]*gl.Project, error) {
		var projects []*gl.Project
		tr := true
		opts := &gl.ListGroupProjectsOptions{
			IncludeSubgroups: &tr,
			Archived:         &archived,
			ListOptions: gl.ListOptions{
				Page:    1,
				PerPage: 100, // this is the max value from gitlab
			},
		}
		for {
			select {
			case <-ctx.Done():
				return nil, context.Canceled
			default:
			}

			p, resp, err := c.c.Groups.ListGroupProjects(group, opts, gl.WithContext(ctx))
			if err != nil {
				return nil, err
			}

			projects = append(projects, p...)

			// Exit the loop when we've seen all pages.
			if resp.CurrentPage >= resp.TotalPages {
				log.Debugf("got all results from the API")
				break
			}

			// update page number to fetch next page
			log.Debugf("getting next page from the API")
			opts.Page = resp.NextPage
		}

		return projects, nil
	}

	rootGroup, _, err := c.c.Groups.GetGroup(group)
	if err != nil {
		return nil, err
	}

	g := newGroup(rootGroup)

	projects, err := getProjects(false)
	if err != nil {
		return nil, err
	}
	addSubProjects(g, projects)

	if includeArchived {
		projects, err := getProjects(true)
		if err != nil {
			return nil, err
		}

		addSubProjects(g, projects)
	}

	return g, nil
}

type Namespace string

func (n Namespace) String() string                 { return string(n) }
func (n Namespace) withProject(proj string) string { return path.Join(string(n), proj) }

func (n Namespace) relative(root Namespace) Namespace {
	return Namespace(normalize(strings.TrimPrefix(string(n), string(root))))
}

func (n Namespace) Join(subPaths ...string) Namespace {
	return Namespace(path.Join(string(n),
		path.Join(subPaths...),
	))
}

func (n Namespace) elements() []string {
	if n == "" {
		return nil
	}
	return strings.Split(string(n), "/")
}
