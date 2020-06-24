package gitlab

import (
	"testing"

	gl "github.com/xanzy/go-gitlab"
)

func TestConstructProjectTreeCase(t *testing.T) {
	rootGroup := &gl.Group{
		Name:     "GROUP",
		FullPath: "group",
	}

	subProjects := []*gl.Project{
		{
			PathWithNamespace: "group/myproject",
			Name:              "myproject",
			Namespace: &gl.ProjectNamespace{
				FullPath: "group",
			},
		},
		{
			PathWithNamespace: "group/build/bazel",
			Name:              "bAZEL",
			Namespace: &gl.ProjectNamespace{
				FullPath: "group/build",
			},
		},
	}

	root := newGroup(rootGroup)
	addSubProjects(root, subProjects)

	if root.name != "GROUP" {
		t.Errorf("root node name not correct. expected=%q, got=%q", "GROUP", root.name)
	}
	if root.namespace != "" {
		t.Errorf("root namespace not correct. expected=%q, got=%q", "", root.namespace)
	}

	if len(root.nodes()) != 2 {
		t.Fatalf("root number of subnodes not correct. expected=%v, got=%v", 2, len(root.nodes()))
	}

	// we could also get it by name to test that function, but this ensures that the
	// nodes are ordered alphabetically, which is fine too.
	build := root.nodes()[0].(*Group)
	if build.Name() != "build" {
		t.Errorf("build name not correct. expected=%q, got=%q", "build", build.Name())
	}

	if build.FullPath() != "group/build" {
		t.Errorf("build fullpath not correct. expected=%q, got=%q", "group/build", build.FullPath())
	}

	if build.Namespace() != "group" {
		t.Errorf("build namespace not correct. expected=%q, got=%q", "group", build.Namespace())
	}
	if len(build.nodes()) != 1 {
		t.Fatalf("build number of subnodes not correct. expected=%v, got=%v", 1, len(build.nodes()))
	}
	if n := build.getNode("bAZEL"); n == nil {
		t.Fatalf("project name not found. expected=%q, got=%v", "bAZEL", nil)
	}
}

func TestConstructProjectTree(t *testing.T) {
	rootGroup := &gl.Group{
		Name:     "mygroup",
		FullPath: "test/mygroup",
	}

	subProjects := []*gl.Project{
		{
			PathWithNamespace: "test/mygroup/myproject",
			Name:              "myproject",
			Namespace: &gl.ProjectNamespace{
				FullPath: "test/mygroup",
			},
		},
		{
			PathWithNamespace: "test/mygroup/build/bazel",
			Name:              "bazel",
			Namespace: &gl.ProjectNamespace{
				FullPath: "test/mygroup/build",
			},
		},
		{
			PathWithNamespace: "test/mygroup/build/buck",
			Name:              "buck",
			Namespace: &gl.ProjectNamespace{
				FullPath: "test/mygroup/build",
			},
		},
		{
			PathWithNamespace: "test/mygroup/build/shared/tools",
			Name:              "tools",
			Namespace: &gl.ProjectNamespace{
				FullPath: "test/mygroup/build/shared",
			},
		},
	}

	root := newGroup(rootGroup)
	addSubProjects(root, subProjects)

	if root.name != "mygroup" {
		t.Errorf("root node name not correct. expected=%q, got=%q", "mygroup", root.name)
	}
	if root.namespace != "test" {
		t.Errorf("root namespace not correct. expected=%q, got=%q", "test", root.namespace)
	}

	if len(root.nodes()) != 2 {
		t.Fatalf("root number of subnodes not correct. expected=%v, got=%v", 2, len(root.nodes()))
	}

	// we expect the nodes to be ordered alphabetically.
	build := root.nodes()[0].(*Group)
	if build.Name() != "build" {
		t.Errorf("build name not correct. expected=%q, got=%q", "build", build.Name())
	}

	if build.FullPath() != "test/mygroup/build" {
		t.Errorf("build fullpath not correct. expected=%q, got=%q", "test/mygroup/build", build.FullPath())
	}

	if build.Namespace() != "test/mygroup" {
		t.Errorf("build namespace not correct. expected=%q, got=%q", "test/mygroup", build.Namespace())
	}
	if len(build.nodes()) != 3 {
		t.Fatalf("build number of subnodes not correct. expected=%v, got=%v", 2, len(build.nodes()))
	}

	shared := build.nodes()[2].(*Group)
	if shared.Name() != "shared" {
		t.Errorf("group name not correct. expected=%q, got=%q", "shared", shared.Name())
	}
	if len(shared.nodes()) != 1 {
		t.Fatalf("build/shared number of subnodes not correct. expected=%v, got=%v", 1, len(shared.nodes()))
	}
	if shared.nodes()[0].Name() != "tools" {
		t.Errorf("project name not correct. expected=%q, got=%q", "tools", shared.nodes()[0].Name())
	}

	project := root.nodes()[1].(*Project)
	if project.Name() != "myproject" {
		t.Errorf("project name not correct. expected=%q, got=%q", "myproject", project.Name())
	}

	if project.FullPath() != "test/mygroup/myproject" {
		t.Errorf("project fullpath not correct. expected=%q, got=%q", "test/mygroup/myproject", project.FullPath())
	}

	if project.Namespace() != "test/mygroup" {
		t.Errorf("project namespace not correct. expected=%q, got=%q", "test/mygroup", project.Namespace())
	}
}

func TestNamespace(t *testing.T) {
	n := Namespace("test/my/group/subgroup/project")

	if len(n.elements()) != 5 {
		t.Errorf("namespace does not have correct number of elements. expected=%v, got=%v", 5, len(n.elements()))
	}

	subNS := n.relative("test/my/group")
	if subNS != Namespace("subgroup/project") {
		t.Errorf("subnamespace not correct. expected=%q, got=%q", Namespace("subgroup/project"), subNS)
	}

	if len(subNS.elements()) != 2 {
		t.Errorf("subnamespace does not have correct number of elements. expected=%v, got=%v", 2, len(n.elements()))
	}
	if subNS.elements()[0] != "subgroup" || subNS.elements()[1] != "project" {
		t.Errorf("namespace does not have correct elements. got=%v", subNS.elements())
	}

	n = Namespace("test/my/group")

	if n.withProject("mycoolproject") != "test/my/group/mycoolproject" {
		t.Errorf("namespace with project not correct. expected=%q, got=%q", "test/my/group/mycoolproject", n.withProject("mycoolproject"))
	}

	if n.Join("sub", "group") != "test/my/group/sub/group" {
		t.Errorf("namespace joined not correct. expected=%q, got=%q", "test/my/group/sub/group", n.Join("sub,", "group"))
	}
}
