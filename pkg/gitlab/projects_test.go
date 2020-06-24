package gitlab

import (
	"testing"

	gl "github.com/xanzy/go-gitlab"
)

func TestProject(t *testing.T) {
	gp := &gl.Project{
		Name: "MyCoolProject",
		Namespace: &gl.ProjectNamespace{
			FullPath: "root/sub",
		},
		PathWithNamespace: "root/sub/mycoolproject",
	}

	p := &Project{gp}

	if p.Name() != "MyCoolProject" {
		t.Errorf("name of project is wrong. expected=%q, got=%q", "MyCoolProject", p.Name())
	}

	if p.Namespace() != "root/sub" {
		t.Errorf("namespace is wrong. expected=%q, got=%q", "root/sub", p.Namespace())
	}

	if p.FullPath() != "root/sub/mycoolproject" {
		t.Errorf("full path is wrong. expected=%q, got=%q", "root/sub/mycoolproject", p.FullPath())
	}

	if p.Depth() != 2 {
		t.Errorf("depth is wrong. expected=%v, got=%v", 2, p.Depth())
	}
}

func TestGroup(t *testing.T) {
	gp := &gl.Group{
		Name:     "MyCoolGroup",
		FullPath: "root/sub/mycoolgroup",
	}

	g := newGroup(gp)

	if g.Name() != "MyCoolGroup" {
		t.Errorf("name of project is wrong. expected=%q, got=%q", "MyCoolGroup", g.Name())
	}

	if g.Namespace() != "root/sub" {
		t.Errorf("namespace is wrong. expected=%q, got=%q", "root/sub", g.Namespace())
	}

	if g.FullPath() != "root/sub/mycoolgroup" {
		t.Errorf("full path is wrong. expected=%q, got=%q", "root/sub/mycoolproject", g.FullPath())
	}

	if g.Depth() != 2 {
		t.Errorf("depth is wrong. expected=%v, got=%v", 2, g.Depth())
	}

	gp = &gl.Group{
		Name:     "rootGroup",
		FullPath: "rootGroup",
	}
	g = newGroup(gp)

	if g.Depth() != 0 {
		t.Errorf("depth is wrong. expected=%v, got=%v", 0, g.Depth())
	}
}

func TestPrintProjectTree(t *testing.T) {
	rootGroup := &gl.Group{
		Name:     "mygroup",
		FullPath: "test/mygroup",
	}

	subProjects := []*gl.Project{
		{
			PathWithNamespace: "test/mygroup/myproject",
			Description:       "myproject is a cool project",
			Name:              "myproject",
			Namespace: &gl.ProjectNamespace{
				FullPath: "test/mygroup",
			},
		},
		{
			PathWithNamespace: "test/mygroup/build/bazel",
			Name:              "bazel",
			Description:       "On pizza, bazel tastes great",
			Namespace: &gl.ProjectNamespace{
				FullPath: "test/mygroup/build",
			},
		},
		{
			PathWithNamespace: "test/mygroup/build/buck",
			Name:              "buck",
			Archived:          true,
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

	// the whitespace is important because of padding
	expected := `name           type                 group                description
----           ----                 -----                -----------
mygroup        group                test                 
├─ build       group                test/mygroup         
│  ├─ bazel    project              test/mygroup/build   On pizza, bazel tastes great
│  ├─ buck     project (archived)   test/mygroup/build   
│  └─ shared   group                test/mygroup/build   
└─ myproject   project              test/mygroup         myproject is a cool project
`

	node := newGroup(rootGroup)
	addSubProjects(node, subProjects)

	actual := PrintProject(node, PrintOptions{
		PrintArchived:    true,
		PrintDescription: true,
		Depth:            2,
	})

	if actual != expected {
		t.Errorf("PrintProject output differs. expected=\n%qgot=\n%q", expected, actual)
	}
}
