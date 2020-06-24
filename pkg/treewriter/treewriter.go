/*
package treewriter implements a utility to output trees
(or similar datastructures) beautifully to a specified
writer.

The output looks similar to the `tree` command:

	├─ gitlab
	│   ├─ gitlab.go
	│   └─ projects.go
	├─ log
	│   └─ log.go
	└─ treewriter
	    ├─ treewriter.go
	    └─ treewriter_test.go

*/
package treewriter

const (
	listItem = "│  "
	midItem  = "├─ "
	lastItem = "└─ "
	noItem   = "   "
)

// Writer contains methods to write a beautiful tree.
type Writer interface {
	// Element prints an element on the current level.
	Element(e string) string

	// Sub creates a new writer which prints subitems
	// for the last printed element. It needs to know
	// the number of sub-elements to write in order to
	// output the correct characters.
	//
	// The write-behavior after the number of elements
	// are written is unspecified and may produce garbage.
	Sub(elements int) Writer
}

// rootWriter is used to output the topmost element, the root.
// It exists because it is easier to just create a separate
// writer instead of special-casing the subWriter.
type rootWriter struct{}

// New returns a new Writer, ready to use.
func New() Writer {
	return &rootWriter{}
}

func (*rootWriter) Element(e string) string {
	return e
}

func (*rootWriter) Sub(elements int) Writer {
	return &subWriter{
		expectedElements: elements,
		nextElement:      1,
	}
}

type subWriter struct {
	expectedElements, nextElement int

	// treePrefix records the current "prefix" that will
	// be printed in a tree. For example:
	//
	//   ├─ gitlab
	//   │   ├─ gitlab.go
	//   │   └─ projects.go
	//
	// For the first line, no prefix is needed.
	// For the second and third line, the tree prefix will
	// be "│   ".
	treePrefix string
}

func (sw *subWriter) isLast() bool {
	return sw.nextElement >= sw.expectedElements
}

func (sw *subWriter) Element(e string) string {
	defer func() {
		sw.nextElement++
	}()

	if sw.isLast() {
		return sw.treePrefix + lastItem + e
	}
	return sw.treePrefix + midItem + e
}

func (sw *subWriter) Sub(elements int) Writer {
	prefix := sw.treePrefix

	// if this is true, then we already printed the last
	// element of the top-level tree. Thus, te subtree gets
	// a noItem.
	if sw.nextElement > sw.expectedElements {
		prefix += noItem
	} else {
		prefix += listItem
	}

	return &subWriter{
		expectedElements: elements,
		nextElement:      1,
		treePrefix:       prefix,
	}
}
