package gitlab

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/tommyknows/gitlab-cli/pkg/log"
	"github.com/tommyknows/gitlab-cli/pkg/treewriter"
	gl "github.com/xanzy/go-gitlab"
	"golang.org/x/sync/errgroup"
)

// ProjectNode is a node in a tree.
type ProjectNode interface {
	// The Name of a node. They may be stylised (e.g. capital letters)
	Name() string

	// Namespace returns the namespace of the node.
	Namespace() Namespace

	// FullPath returns the full path (namespace+name) of a node.
	FullPath() Namespace

	// Depth returns the depth of the node, starting at 0 for root-level nodes
	Depth() int

	// seal
	projectNode()
}

// noder extends the projectNode and describes an entity
// that can hold subnodes. This applies to (as currently
// known) users and groups.
type noder interface {
	ProjectNode
	nodes() []ProjectNode
	addNodes(...ProjectNode)
	getNode(name string) ProjectNode
	numNodes(includeArchived bool) int
}

type Visitor func(p ProjectNode) error
type ContextVisitor func(ctx context.Context, p ProjectNode) error

// Walk a Project (tree) depth-first. Stops walking on error.
func Walk(root ProjectNode, walkFunc Visitor) error {
	if err := walkFunc(root); err != nil {
		return errors.Wrapf(err, "could not walk node")
	}

	node, ok := root.(noder)
	if !ok {
		return nil
	}

	for _, n := range node.nodes() {
		if err := Walk(n, walkFunc); err != nil {
			return errors.Wrapf(err, "could not walk project")
		}
	}
	return nil
}

// WalkConcurrent walks a Project (tree) depth-first, starting a goroutine
// for every child's node AFTER the node has been visited.
func WalkConcurrent(ctx context.Context, root ProjectNode, walkFunc ContextVisitor) error {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	if err := walkFunc(ctx, root); err != nil {
		return errors.Wrapf(err, "could not walk node")
	}

	node, ok := root.(noder)
	if !ok {
		return nil
	}

	g, ctx := errgroup.WithContext(ctx)
	for _, n := range node.nodes() {
		n := n // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			return WalkConcurrent(ctx, n, walkFunc)
		})
	}

	return g.Wait()
}

func sortNodes(s []ProjectNode) {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Name() < s[j].Name()
	})
}

type PrintOptions struct {
	PrintArchived,
	PrintDescription bool
	Depth int
}

// PrintProject pretty-prints a projectNode with the supplied settings.
func PrintProject(g ProjectNode, opts PrintOptions) string {
	const (
		archived = " (archived)"
		project  = "project"
		group    = "group"
	)

	var (
		header   = "name\ttype\tgroup\n----\t----\t-----\n"
		printFmt = "%s\t%s\t%s\n"

		b         = new(strings.Builder)
		tabWrite  = tabwriter.NewWriter(b, 4, 5, 3, ' ', tabwriter.StripEscape)
		treeWrite = treewriter.New()
	)

	if opts.PrintDescription {
		header = "name\ttype\tgroup\tdescription\n----\t----\t-----\t-----------\n"
		printFmt = "%s\t%s\t%s\t%s\n"
	}

	fmt.Fprint(tabWrite, header)

	writers := map[Namespace]treewriter.Writer{
		g.Namespace(): treeWrite,
	}

	printProject := func(name, typ string, ns Namespace, description string) {
		if opts.PrintDescription {
			fmt.Fprintf(tabWrite, printFmt, name, typ, ns, description)
		} else {
			fmt.Fprintf(tabWrite, printFmt, name, typ, ns)
		}
	}

	printGroup := func(name, typ string, ns Namespace) {
		if opts.PrintDescription {
			fmt.Fprintf(tabWrite, printFmt, name, typ, ns, "")
		} else {
			fmt.Fprintf(tabWrite, printFmt, name, typ, ns)
		}
	}

	depthReached := func(d int) bool {
		if opts.Depth == 0 {
			return false
		}
		return d-g.Depth() > opts.Depth
	}

	if opts.PrintArchived {
		log.Debugf("printing archived repositories too!")
	}

	err := Walk(g, func(p ProjectNode) error {
		if depthReached(p.Depth()) {
			return nil
		}

		switch n := p.(type) {
		case *Project:
			tw, ok := writers[n.Namespace()]
			if !ok {
				return errors.Errorf("no writer for project %v (namespace %v)", n.FullPath(), n.Namespace())
			}

			typ := project

			if n.gp.Archived {
				if !opts.PrintArchived {
					return nil
				}
				typ += archived
			}

			printProject(tw.Element(n.Name()), typ, n.Namespace(), n.gp.Description)

		case noder:
			tw, ok := writers[n.Namespace()]
			if !ok {
				return errors.Errorf("no writer for group %v", n.FullPath())
			}

			printGroup(tw.Element(n.Name()), group, n.Namespace())

			writers[n.FullPath()] = tw.Sub(n.numNodes(opts.PrintArchived))
		}

		return nil
	})
	if err != nil {
		// this really should not happen.
		panic(err)
	}

	if err := tabWrite.Flush(); err != nil {
		panic(err)
	}

	return b.String()
}

func addSubProjects(rootNode noder, subProjects []*gl.Project) {
	for _, subProj := range subProjects {
		namespaces := Namespace(subProj.PathWithNamespace).relative(rootNode.FullPath()).elements()
		group := rootNode

		for i, ns := range namespaces {
			// if we're at the last element, we have a project that we add to the nodes.
			if i == len(namespaces)-1 {
				group.addNodes(newProject(subProj))
				break
			}

			switch n := group.getNode(ns).(type) {
			case noder:
				group = n

			case *Project:
				panic("node found but is project, not group")

			default:
				// no match found, create node
				subGroup := &Group{
					name:      ns,
					namespace: rootNode.FullPath().Join(namespaces[:i]...),
				}
				// the name of the group may not be the exact name from the API. However,
				// the projects do not contain the actual name, but rather the "pathName".
				// Because we alreday have the pathName, we can construct the fullPath
				// from it together with the namespace.
				subGroup.fullPath = subGroup.namespace.withProject(subGroup.name)

				group.addNodes(subGroup)
				group = subGroup
			}
		}
	}

	// we never return an error, so we don't expect one from walk.
	_ = Walk(rootNode, func(p ProjectNode) error {
		n, ok := p.(noder)
		if !ok {
			return nil
		}

		sortNodes(n.nodes())
		return nil
	})
}

type User struct {
	fullname string
	username string

	projects []ProjectNode
}

func (u *User) projectNode()              {}
func (u *User) Namespace() Namespace      { return Namespace("") }
func (u *User) Name() string              { return u.username }
func (u *User) FullPath() Namespace       { return Namespace(u.username) }
func (u *User) Depth() int                { return 1 }
func (u *User) addNodes(n ...ProjectNode) { u.projects = append(u.projects, n...) }
func (u *User) nodes() []ProjectNode      { return u.projects }
func (u *User) getNode(name string) ProjectNode {
	for _, node := range u.projects {
		if node.Name() == name {
			return node
		}
	}
	return nil
}
func (u *User) numNodes(includeArchived bool) int {
	if includeArchived {
		return len(u.projects)
	}

	var archived int
	for _, n := range u.projects {
		if p, ok := n.(*Project); ok {
			if p.gp.Archived {
				archived++
			}
		}
	}

	return len(u.projects) - archived
}
func newUser(u *gl.User) *User {
	return &User{
		fullname: u.Name,
		username: u.Username,
	}
}

type Group struct {
	name      string
	namespace Namespace
	// fullPath may differ from <namespace>/<name> due
	// to the name containing uppercase letters.
	fullPath string

	subNodes []ProjectNode
}

func newGroup(g *gl.Group) *Group {
	return &Group{
		name:      g.Name,
		namespace: extractNamespace(g.FullPath),
		fullPath:  g.FullPath,
	}
}

func (g *Group) projectNode()              {}
func (g *Group) Namespace() Namespace      { return g.namespace }
func (g *Group) Name() string              { return g.name }
func (g *Group) FullPath() Namespace       { return Namespace(g.fullPath) }
func (g *Group) Depth() int                { return len(g.namespace.elements()) }
func (g *Group) addNodes(n ...ProjectNode) { g.subNodes = append(g.subNodes, n...) }
func (g *Group) nodes() []ProjectNode      { return g.subNodes }

func (g *Group) getNode(name string) ProjectNode {
	for _, node := range g.nodes() {
		if node.Name() == name {
			return node
		}
	}
	return nil
}

func (g *Group) numNodes(includeArchived bool) int {
	if includeArchived {
		return len(g.subNodes)
	}

	var archived int
	for _, n := range g.subNodes {
		if p, ok := n.(*Project); ok {
			if p.gp.Archived {
				archived++
			}
		}
	}

	return len(g.subNodes) - archived
}

type Project struct {
	gp *gl.Project
}

func newProject(pr *gl.Project) *Project {
	p := Project{pr}
	return &p
}

func (p *Project) projectNode()         {}
func (p *Project) Namespace() Namespace { return Namespace(p.gp.Namespace.FullPath) }
func (p *Project) Name() string         { return p.gp.Name }
func (p *Project) FullPath() Namespace  { return Namespace(p.gp.PathWithNamespace) }
func (p *Project) Depth() int           { return len(p.Namespace().elements()) }
