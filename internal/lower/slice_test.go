// ABOUTME: Tests for lowering slices to Slice{Kind}: the five kinds, told apart by the base's sigil and the bracket.
// ABOUTME: Each fragment's parse is measured, and each claim about what a slice gives is perl 5.42's.

package lower_test

import (
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// sliceKind returns n's Kind if n is a Slice. It reads the field through
// reflect, so the test asserts no lowered type.
func sliceKind(n lower.Node) (lower.SliceKind, bool) {
	v := reflect.ValueOf(n)
	if v.Type() != reflect.TypeOf(&lower.Slice{}) {
		return 0, false
	}
	return lower.SliceKind(v.Elem().FieldByName("Kind").Uint()), true
}

// The five kinds give four result types (§4.4.7). perl 5.42, with @a =
// (10,20,30) and %h = (x=>1, y=>2): @a[0,2] is 10,30 (elements), @h{qw(x
// y)} is 1,2 (values), %a[0,1] is 0,10,1,20 (index/element pairs), %h{"x"}
// is x,1 (key/value pairs), and (1,2)[0] is 1.
func TestLowerSliceKind(t *testing.T) {
	for _, c := range []struct {
		src  string
		want lower.SliceKind
	}{
		{"@a[0,1];", lower.SliceArray},
		{"@h{'a','b'};", lower.SliceHash},
		{"%a[0,1];", lower.SliceKVArray},
		{"%h{'a'};", lower.SliceKVHash},
		{"(1,2)[0];", lower.SliceList},
	} {
		got := lowerOne(t, c.src)
		if k, ok := sliceKind(got); !ok || k != c.want {
			t.Errorf("%q: lowered to %#v, want a Slice of kind %d", c.src, got, c.want)
		}
	}
}

// The sigil on the base separates a slice from an element, not the
// bracket. perl 5.42, with @a = (10,20,30): @a[0] is the one-element list
// (10), $a[0] the scalar 10.
func TestLowerSliceIsNotSubscript(t *testing.T) {
	a := &lower.Var{Sigil: "@", Name: "a"}
	zero := &lower.NumLit{Text: "0", IsInt: true}
	sameShape(t, "@a[0];", zeroIDs(lowerOne(t, "@a[0];")), &lower.Slice{Base: a, Idx: zero, Kind: lower.SliceArray})
	sameShape(t, "$a[0];", zeroIDs(lowerOne(t, "$a[0];")), &lower.Index{Base: &lower.Var{Sigil: "$", Name: "a"}, Idx: zero})
}

// perl 5.42 rejects both fragments (`perl -c`: syntax error, near "[;" and
// near "{;"). The parser keeps an Unknown subscript inside an intact index,
// so the Slice keeps its base and kind and holds a Hole for Idx.
func TestLowerSliceOverBrokenInput(t *testing.T) {
	cases := []struct {
		src  string
		want lower.Node
	}{
		{"@a[;", &lower.Slice{Base: &lower.Var{Sigil: "@", Name: "a"}, Idx: &lower.Hole{}, Kind: lower.SliceArray}},
		{"%h{;", &lower.Slice{Base: &lower.Var{Sigil: "%", Name: "h"}, Idx: &lower.Hole{}, Kind: lower.SliceKVHash}},
	}
	for _, c := range cases {
		root := parse.Parse([]byte(c.src))
		if len(root.Children) == 0 || len(root.Children[0].Children) != 1 || root.Children[0].Children[0].Kind != parse.Index {
			t.Fatalf("%q: the first statement holds no index", c.src)
		}
		sameShape(t, c.src, zeroIDs(lower.LowerExpr(root, root.Children[0].Children[0])), c.want)
	}
}

// A slice's base and subscript lower as the same spelling's element would.
// perl -MO=Deparse 5.42 prints @h{a} as @h{'a'}, so a lone word in braces
// is quoted; qw(a b c)[1] as ('a', 'b', 'c')[1], a list slice; and
// $x->@[0], $x->%{a} as @$x[0], %$x{'a'}, so a postfix slice's base is the
// postfix Deref of what the arrow follows. With $x = {a=>1}, $x->@{a} is 1
// and $x->%{a} is a,1.
func TestLowerSliceShape(t *testing.T) {
	zero := &lower.NumLit{Text: "0", IsInt: true}
	paren := func(n lower.Node) lower.Node { n.Meta().Paren = true; return n }
	a := &lower.Bareword{Name: "a", Quoted: true}
	deref := func(sigil string, postfix bool) *lower.Deref {
		return &lower.Deref{Sigil: sigil, Expr: &lower.Var{Sigil: "$", Name: "x"}, Postfix: postfix}
	}
	cases := []struct {
		src  string
		want lower.Node
	}{
		{"@h{a};", &lower.Slice{Base: &lower.Var{Sigil: "@", Name: "h"}, Idx: a, Kind: lower.SliceHash}},
		{"%h{a};", &lower.Slice{Base: &lower.Var{Sigil: "%", Name: "h"}, Idx: a, Kind: lower.SliceKVHash}},
		{"@$x[0];", &lower.Slice{Base: deref("@", false), Idx: zero, Kind: lower.SliceArray}},
		{"%$x{a};", &lower.Slice{Base: deref("%", false), Idx: a, Kind: lower.SliceKVHash}},
		{"($x)[0];", &lower.Slice{Base: &lower.Var{Sigil: "$", Name: "x"}, Idx: zero, Kind: lower.SliceList}},
		{"(($x))[0];", &lower.Slice{Base: paren(&lower.Var{Sigil: "$", Name: "x"}), Idx: zero, Kind: lower.SliceList}},
		{"qw(a b c)[1];", &lower.Slice{Base: &lower.QwList{Words: []string{"a", "b", "c"}}, Idx: &lower.NumLit{Text: "1", IsInt: true}, Kind: lower.SliceList}},
		{"$x->@[0];", &lower.Slice{Base: deref("@", true), Idx: zero, Kind: lower.SliceArray}},
		{"$x->@{a};", &lower.Slice{Base: deref("@", true), Idx: a, Kind: lower.SliceHash}},
		{"$x->%[0];", &lower.Slice{Base: deref("%", true), Idx: zero, Kind: lower.SliceKVArray}},
		{"$x->%{a};", &lower.Slice{Base: deref("%", true), Idx: a, Kind: lower.SliceKVHash}},
		{"$x->@[;", &lower.Slice{Base: deref("@", true), Idx: &lower.Hole{}, Kind: lower.SliceArray}},
	}
	for _, c := range cases {
		root := parse.Parse([]byte(c.src))
		sameShape(t, c.src, zeroIDs(lower.LowerExpr(root, root.Children[0].Children[0])), c.want)
	}

	// A postfix slice is the arrow's id, and its Deref the bare sigil's.
	root := parse.Parse([]byte("$x->@[0];"))
	n := root.Children[0].Children[0]
	ids := lower.AssignIDs(root)
	got := lower.LowerExpr(root, n)
	base := lower.Children(got)[0]
	if got.Meta().ID != ids[n] || base.Meta().ID != ids[n.Children[1].Children[0]] {
		t.Errorf("ids %v, %v are not the arrow's and the sigil's %v, %v", got.Meta().ID, base.Meta().ID, ids[n], ids[n.Children[1].Children[0]])
	}
}
