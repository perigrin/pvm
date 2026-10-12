// ABOUTME: Tests for the position map: offsets resolve to lowered nodes, erased bytes map forward, order is source order.
// ABOUTME: Each case lowers a whole program with lower.Lower and queries the map built over that program's CST.

package lower_test

import (
	"reflect"
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

// isMatch is the predicate nearestPatternBefore (infer.go:1726) asks with.
func isMatch(n lower.Node) bool { return reflect.TypeOf(n) == reflect.TypeOf(&lower.Match{}) }

// $1 reads the last match before it in source order (§4.14's preamble;
// infer.go:1726). Between two matches, the table answers with the first;
// after both, with the second; before either, with nothing.
func TestPositionMapPreservesSourceOrder(t *testing.T) {
	src := "$a =~ /a/;\n$1;\n$b =~ /b/;\n$1;"
	_, f, p := positions(src)
	first, second := lower.Children(f.Stmts[0])[0], lower.Children(f.Stmts[2])[0]
	for _, c := range []struct {
		off  int
		want lower.Node
	}{{0, nil}, {11, first}, {15, first}, {26, second}} {
		if got := p.LastBefore(c.off, isMatch); got != c.want {
			t.Errorf("%q: the last match before offset %d is %#v, want %#v", src, c.off, got, c.want)
		}
	}
}

// An Unknown lowers to an opaque Hole (§4.14.2, Hole row), and every byte
// it covers resolves to that Hole, the bytes of its own children included,
// so the LSP answers "no information" there rather than a type. Perl
// rejects both programs.
func TestPositionMapInsideHole(t *testing.T) {
	src := "$x = ;"
	_, f, p := positions(src)
	hole := lower.Children(lower.Children(f.Stmts[0])[0])[1] // the Assign's R
	if got := p.At(5); got != hole || reflect.TypeOf(got) != reflect.TypeOf(&lower.Hole{}) {
		t.Errorf("%q: offset 5 resolves to %#v, want the Hole %#v", src, got, hole)
	}

	src = "if ($x) {"
	_, f, p = positions(src)
	hole = lower.Children(f.Stmts[0])[1] // the IfStmt's Then: the Unknown holding a block
	if got := p.At(8); got != hole || reflect.TypeOf(got) != reflect.TypeOf(&lower.Hole{}) {
		t.Errorf("%q: offset 8, in the Unknown's block, resolves to %#v, want the Hole %#v", src, got, hole)
	}
}

// parseNode is the syntax pointer no lowered type may hold.
var parseNode = reflect.TypeOf((*parse.Node)(nil))

// holdsSyntax reports whether a value of type t can hold a *parse.Node in a
// field, an element or a key. A Node-typed field is an interface over the
// lowered kinds, each of which is checked in its own right.
func holdsSyntax(t reflect.Type, seen map[reflect.Type]bool) bool {
	if t == parseNode {
		return true
	}
	if seen[t] {
		return false
	}
	seen[t] = true
	switch t.Kind() {
	case reflect.Map:
		return holdsSyntax(t.Key(), seen) || holdsSyntax(t.Elem(), seen)
	case reflect.Pointer, reflect.Slice, reflect.Array:
		return holdsSyntax(t.Elem(), seen)
	case reflect.Struct:
		for i := range t.NumField() {
			if holdsSyntax(t.Field(i).Type, seen) {
				return true
			}
		}
	}
	return false
}

// The lowered tree carries no syntax pointers (§6.1.6 item 3): the position
// map is the one connection back to the CST, so the tree is a value of its
// own. Every kind TestZeroValueNodeIsUsable lists is checked, and a control
// shows the check finds a pointer behind a slice.
func TestLoweredTreeHasNoSyntaxPointers(t *testing.T) {
	if !holdsSyntax(reflect.TypeOf(struct{ N []*parse.Node }{}), map[reflect.Type]bool{}) {
		t.Fatal("the check misses a []*parse.Node field")
	}
	for _, n := range allKinds() {
		if holdsSyntax(reflect.TypeOf(n), map[reflect.Type]bool{}) {
			t.Errorf("%T can hold a *parse.Node", n)
		}
	}
}
