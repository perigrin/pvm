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
