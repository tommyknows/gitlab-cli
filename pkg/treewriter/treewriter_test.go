package treewriter

import (
	"fmt"
	"strings"
	"testing"
)

type testTree struct {
	subtrees []testTree
	name     string
}

func TestTreeWriting(t *testing.T) {
	tTree := &testTree{
		name: "root",
		subtrees: []testTree{
			{
				name: "sub1",
				subtrees: []testTree{
					{name: "subsub11"},
					{name: "subsub12"},
				},
			},
			{
				name: "sub2",
				subtrees: []testTree{
					{name: "subsub21"},
					{
						name: "subsub22",
						subtrees: []testTree{
							{name: "subsubsub221"},
							{
								name:     "subsubsub222",
								subtrees: []testTree{{name: "subsubsubsub2221"}},
							},
						},
					},
					{name: "subsub23"},
				},
			},
			{
				name: "sub3",
				subtrees: []testTree{
					{name: "subsub31"},
					{name: "subsub32"},
				},
			},
		},
	}

	var b strings.Builder
	w := New()
	recursiveTreePrint(tTree, w, &b)

	expected := `root
├─ sub1
│  ├─ subsub11
│  └─ subsub12
├─ sub2
│  ├─ subsub21
│  ├─ subsub22
│  │  ├─ subsubsub221
│  │  └─ subsubsub222
│  │     └─ subsubsubsub2221
│  └─ subsub23
└─ sub3
   ├─ subsub31
   └─ subsub32
`

	if b.String() != expected {
		t.Errorf("tree is not as expected. got=\n%s, expected=\n%s", b.String(), expected)
	}
}

func recursiveTreePrint(t *testTree, w Writer, b *strings.Builder) {
	fmt.Fprintf(b, "%s\n", w.Element(t.name))

	wsub := w.Sub(len(t.subtrees))

	for _, st := range t.subtrees {
		recursiveTreePrint(&st, wsub, b)
	}
}
