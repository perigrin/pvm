// ABOUTME: Tests for the position map: offsets resolve to lowered nodes, erased bytes map forward, order is source order.
// ABOUTME: Each case lowers a whole program with lower.Lower and queries the map built over that program's CST.

package lower_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

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

// An offset no statement covers resolves to no node, and never panics: past
// the end, before the start, in whitespace after the last statement, and
// anywhere in an empty file. The File's bytes run from its first
// statement's start to its last one's end, so trivia between statements is
// the File's and trivia outside them is nobody's.
func TestPositionMapOutOfRange(t *testing.T) {
	for _, c := range []struct {
		src string
		off int
	}{{"1;", 2}, {"1;", 100}, {"1;", -1}, {"1;\n\n", 3}, {"", 0}} {
		_, _, p := positions(c.src)
		if got := p.At(c.off); got != nil {
			t.Errorf("%q: offset %d resolves to %#v, want no node", c.src, c.off, got)
		}
	}
	_, _, p := positions("")
	if got := p.LastBefore(1, func(lower.Node) bool { return true }); got != nil {
		t.Errorf("an empty file has %#v before offset 1, want no node", got)
	}
}

// Span is the map's other direction: a lowered node's id to its bytes. A
// node with no parse node of its own, a named sub's AnonSub or the handle of
// `<STDIN>`, has its parent's span; the File's span is its statements', without the trivia
// around them; an id the tree never gave has none.
func TestPositionMapSpan(t *testing.T) {
	src := "# a\nsub f { 1 }\n# b\n2;\n\n"
	_, f, p := positions(src)
	decl := lower.Children(f.Stmts[0])[0]
	kids := lower.Children(decl)
	sub := kids[len(kids)-1]
	for _, c := range []struct {
		name       string
		id         lower.ID
		start, end int
	}{
		{"the Decl", decl.Meta().ID, 4, 15},
		{"its AnonSub", sub.Meta().ID, 4, 15},
		{"the File", f.Meta().ID, 4, 22},
	} {
		if s, e, ok := p.Span(c.id); !ok || s != c.start || e != c.end {
			t.Errorf("%q: %s spans [%d,%d) ok=%v, want [%d,%d)", src, c.name, s, e, ok, c.start, c.end)
		}
	}
	src2 := "<STDIN>;"
	_, f2, p2 := positions(src2)
	handle := lower.Children(lower.Children(f2.Stmts[0])[0])[0]
	if s, e, ok := p2.Span(handle.Meta().ID); !ok || s != 0 || e != 7 {
		t.Errorf("%q: the Readline's handle spans [%d,%d) ok=%v, want the term's [0,7)", src2, s, e, ok)
	}
	if _, _, ok := p.Span(lower.ID{Anchor: "sub g#0"}); ok {
		t.Errorf("%q: an id the tree never gave has a span", src)
	}
}

// A Match keeps its pattern inside its one quote token (§4.14.2, PARSER),
// and the map hands that token back by the Match's id, bound or not, so $1
// can read the groups of the match before it. Anything but a Match has no
// quote token.
func TestPositionMapQuote(t *testing.T) {
	src := `$s =~ /(\d+)/; s{a}{b}g; ($t !~ m!x!); qr/y/; tr/a/b/; $x; $u =~ $re;`
	_, f, p := positions(src)
	want := []string{`/(\d+)/`, `s{a}{b}g`, `m!x!`, `qr/y/`, `tr/a/b/`, ``, ``}
	for i, stmt := range f.Stmts {
		n := lower.Children(stmt)[0]
		if got := p.Quote(n.Meta().ID); got != want[i] {
			t.Errorf("%q statement %d: Quote is %q, want %q", src, i, got, want[i])
		}
	}
}

// $1 asks LastBefore for the match before it, once per capture, so a
// linear scan over the file per question is quadratic: 4,000 lines of
// `if ($s =~ /(a)/) { print $1 }` took 6.7 s through psc check. Each
// question here costs a search and the few nodes back to its match, so
// 4,000 of them visit a bounded number of nodes and finish well within a
// second. The control: each answer is the match on its own line.
func TestCaptureResolutionIsNotQuadratic(t *testing.T) {
	const lines = 4000
	line := "if ($s =~ /(a)/) { print $1 }\n"
	src := strings.Repeat(line, lines)
	_, _, pos := positions(src)
	visits := 0
	keep := func(n lower.Node) bool { visits++; return isMatch(n) }
	start := time.Now()
	for i := 0; i < lines; i++ {
		off := i*len(line) + strings.Index(line, "$1")
		m := pos.LastBefore(off, keep)
		if m == nil {
			t.Fatalf("line %d: no match before $1", i)
		}
		if s, _, _ := pos.Span(m.Meta().ID); s/len(line) != i {
			t.Fatalf("line %d: the match before $1 is on line %d", i, s/len(line))
		}
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("%d captures resolved in %v, want under 1s", lines, elapsed)
	}
	if visits > 20*lines {
		t.Errorf("%d captures visited %d nodes, want at most %d", lines, visits, 20*lines)
	}
}
