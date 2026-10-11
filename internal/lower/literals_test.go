// ABOUTME: Tests for lowering literals and names: NumLit, StrLit, QwList and Bareword from their CST terms.
// ABOUTME: Each shape is the parse measured on the fragment, and each value claim is perl 5.42's reading.

package lower_test

import (
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// expression returns the one expression of a one-statement program.
func expression(t *testing.T, root *parse.Node) *parse.Node {
	t.Helper()
	if len(root.Children) != 1 || root.Children[0].Kind != parse.Statement || len(root.Children[0].Children) != 1 {
		t.Fatalf("want one statement holding one expression, got %d children", len(root.Children))
	}
	return root.Children[0].Children[0]
}

// sameShape compares a lowered node against want, ignoring ids: the test
// for ids is that they come from AssignIDs, checked separately.
func sameShape(t *testing.T, src string, got, want lower.Node) {
	t.Helper()
	if got == nil {
		t.Fatalf("%q: lowered to nil", src)
	}
	got.Meta().ID = lower.ID{}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%q: lowered to %#v, want %#v", src, got, want)
	}
}

func TestLowerLiterals(t *testing.T) {
	paren := func(n lower.Node) lower.Node { n.Meta().Paren = true; return n }
	cases := []struct {
		src  string
		want lower.Node
	}{
		// IsInt comes from Text. perl: 0x1f is 31, 1_000 is 1000.
		{"42;", &lower.NumLit{Text: "42", IsInt: true}},
		{"0x1f;", &lower.NumLit{Text: "0x1f", IsInt: true}},
		{"1_000;", &lower.NumLit{Text: "1_000", IsInt: true}},
		{"1.5;", &lower.NumLit{Text: "1.5"}},
		{"1e3;", &lower.NumLit{Text: "1e3"}},
		{"(42);", paren(&lower.NumLit{Text: "42", IsInt: true})},

		// The delimiter decides interpolation. A non-interpolating body
		// has its escapes resolved; perl: q(a\)b) is a)b, q(a\\b) is a\b.
		{"'ab';", &lower.StrLit{Value: "ab"}},
		{`'a\'b';`, &lower.StrLit{Value: "a'b"}},
		{`q(a\)b);`, &lower.StrLit{Value: "a)b"}},
		{`q(a\\b);`, &lower.StrLit{Value: `a\b`}},
		{"q{a{b}c};", &lower.StrLit{Value: "a{b}c"}},
		{`"ab";`, &lower.StrLit{Value: "ab", Interpolated: true}},
		{"qq{a};", &lower.StrLit{Value: "a", Interpolated: true}},
		// Interp is PARSER: the lexer defers interpolation, so the text is
		// kept whole rather than refused.
		{`"a$b c";`, &lower.StrLit{Value: "a$b c", Interpolated: true}},

		// A backtick string is a command. qx'' does not interpolate; perl:
		// qx'echo $HOME' runs the shell's $HOME, not perl's.
		{"`ls`;", &lower.StrLit{Value: "ls", Interpolated: true, Backtick: true}},
		{"qx/ls/;", &lower.StrLit{Value: "ls", Interpolated: true, Backtick: true}},
		{"qx{ls -l};", &lower.StrLit{Value: "ls -l", Interpolated: true, Backtick: true}},
		{"qx'echo $HOME';", &lower.StrLit{Value: "echo $HOME", Backtick: true}},

		// The lowering splits the qw body on whitespace. perl: (qw(a b)) x 2
		// and ("a","b") x 2 both give a b a b.
		{"qw(a b);", &lower.QwList{Words: []string{"a", "b"}}},
		{"qw(a b c);", &lower.QwList{Words: []string{"a", "b", "c"}}},
		{"qw{ a\n b };", &lower.QwList{Words: []string{"a", "b"}}},
		{"qw//;", &lower.QwList{Words: []string{}}},
	}
	for _, c := range cases {
		root := parse.Parse([]byte(c.src))
		n := expression(t, root)
		got := lower.LowerExpr(root, n)
		if got != nil && got.Meta().ID != lower.AssignIDs(root)[n] {
			t.Errorf("%q: id %v is not the term's id", c.src, got.Meta().ID)
		}
		sameShape(t, c.src, got, c.want)
	}
}

// Input perl itself rejects is Unknown by construction, whatever the parser
// learns later. perl 5.42: both fragments fail `perl -c` with "Can't find
// string terminator". A Hole carries the Unknown's id, which is what keys
// its bytes in the position map, and has no children even when the Unknown
// does.
func TestLowerUnknownToHole(t *testing.T) {
	for _, src := range []string{"qw(a b", `"abc`} {
		root := parse.Parse([]byte(src))
		if len(root.Children) != 1 || root.Children[0].Kind != parse.Unknown {
			t.Fatalf("%q: want one Unknown, got %d children", src, len(root.Children))
		}
		u := root.Children[0]
		if u.Start != 0 || u.End != len(src) {
			t.Fatalf("%q: Unknown spans %d-%d, want the whole file", src, u.Start, u.End)
		}
		got := lower.LowerExpr(root, u)
		if reflect.TypeOf(got) != reflect.TypeOf(&lower.Hole{}) {
			t.Fatalf("%q: lowered to %T, want *lower.Hole", src, got)
		}
		if got.Meta().ID != lower.AssignIDs(root)[u] {
			t.Errorf("%q: Hole id %v is not the Unknown's", src, got.Meta().ID)
		}
	}

	// `if ($x) {` holds its block inside the Unknown; the Hole does not
	// lower it.
	root := parse.Parse([]byte("if ($x) {"))
	var u *parse.Node
	var find func(*parse.Node)
	find = func(n *parse.Node) {
		if n.Kind == parse.Unknown {
			u = n
		}
		for _, c := range n.Children {
			find(c)
		}
	}
	find(root)
	if u == nil || len(u.Children) == 0 {
		t.Fatalf("want an Unknown with children in `if ($x) {`")
	}
	if kids := lower.Children(lower.LowerExpr(root, u)); len(kids) != 0 {
		t.Errorf("Hole has %d children, want none", len(kids))
	}
}
