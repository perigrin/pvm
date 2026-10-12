// ABOUTME: Tests for the position map: offsets resolve to lowered nodes, erased bytes map forward, order is source order.
// ABOUTME: Each case lowers a whole program with lower.Lower and queries the map built over that program's CST.

package lower_test

import (
	"testing"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// positions lowers src and builds its position map.
func positions(src string) (*parse.Node, *lower.File, *lower.Positions) {
	root := parse.Parse([]byte(src))
	f := lower.Lower(root)
	return root, f, lower.NewPositions(root, f)
}

// The paren of `($x)` is erased into a bit on the Var, and its bytes still
// resolve to that Var, as lower.rs:1630 maps the paren forward. A comment
// between two statements lowers to nothing and resolves to the File that
// encloses it rather than to nothing.
func TestPositionMapErasedNodeMapsForward(t *testing.T) {
	_, f, p := positions("($x);")
	v := lower.Children(f.Stmts[0])[0]
	for _, off := range []int{0, 1, 3} {
		if got := p.At(off); got != v {
			t.Errorf("($x); offset %d resolves to %#v, want the Var %#v", off, got, v)
		}
	}

	src := "1;\n# hi\n2;"
	_, f, p = positions(src)
	if got := p.At(4); got != lower.Node(f) {
		t.Errorf("%q: offset 4, in the comment, resolves to %#v, want the File", src, got)
	}
}
