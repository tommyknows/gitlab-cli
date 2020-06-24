package gitlab

import (
	"fmt"
)

// Clone the ProjectNodes into the current director, mirroring the upstream structure.
// The rootGroup is needed in order to strip away the leading path from the full Project
// Paths
func Clone(root Namespace) Visitor {
	return func(n ProjectNode) error {
		switch n := n.(type) {
		case *Project:
			fmt.Println("cloning", n.Name(), "to ./"+Namespace(n.FullPath()).relative(root))

		case *Group:
			fmt.Println("creating folder", n.namespace.relative(root))
		}
		return nil
	}
}
