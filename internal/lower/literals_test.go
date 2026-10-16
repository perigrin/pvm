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

// An unresolved word means four things (§4.14.1 rule 3), told apart by
// where it stands. perl 5.42: with %h = ("Foo::Bar" => 1, k => 2), $h{k} is
// 2 and exists $h{Foo::Bar} is true; join ",", (a => 1) is "a,1".
func TestLowerBarewordByPosition(t *testing.T) {
	// child picks the word out of the fragment's expression.
	child := func(i int) func(*parse.Node) *parse.Node {
		return func(n *parse.Node) *parse.Node { return n.Children[i] }
	}
	cases := []struct {
		src   string
		word  func(*parse.Node) *parse.Node
		lower func(root, n *parse.Node) lower.Node
		want  lower.Node
	}{
		// The subscript of an `index "{"`: a hash key, quoted.
		{"$h{k};", child(1), lower.LowerHashKey, &lower.Bareword{Name: "k", Quoted: true}},
		{"$h{Foo::Bar};", child(1), lower.LowerHashKey, &lower.Bareword{Name: "Foo::Bar", Quoted: true}},
		// A key that is not a bare word lowers as itself.
		{"$h{'k'};", child(1), lower.LowerHashKey, &lower.StrLit{Value: "k"}},
		// A `call` with Fat set: the left of a fat comma, quoted, in a
		// list or an anonymous constructor alike.
		{"(a => 1);", child(0), lower.LowerExpr, &lower.Bareword{Name: "a", Quoted: true}},
		{"[a => 1];", child(0), lower.LowerExpr, &lower.Bareword{Name: "a", Quoted: true}},
		// The left operand of `->`: a class name.
		{"Foo->new;", child(0), lower.LowerExpr, &lower.Bareword{Name: "Foo"}},
		{"Foo::Bar->new;", child(0), lower.LowerExpr, &lower.Bareword{Name: "Foo::Bar"}},
		// The operand of a file test: the stat buffer.
		{"-r _;", child(0), lower.LowerExpr, &lower.Bareword{Name: "_"}},
	}
	for _, c := range cases {
		root := parse.Parse([]byte(c.src))
		sameShape(t, c.src, c.lower(root, c.word(expression(t, root))), c.want)
	}

	// Without the fat comma the word is a call to `a`, which this slice
	// does not quote.
	root := parse.Parse([]byte("(a, 1);"))
	got := lower.LowerExpr(root, expression(t, root).Children[0])
	got.Meta().ID = lower.ID{}
	if reflect.DeepEqual(got, &lower.Bareword{Name: "a", Quoted: true}) {
		t.Errorf("(a, 1): the call lowered to a quoted Bareword")
	}
}

// A heredoc is a string at its opener's position, and its quote decides
// interpolation as a string's delimiter does. perl 5.42, with $x = "X":
// <<"A", <<B, <<~E and << "F" interpolate ("a X"), <<'C' and <<\D do not
// ("c $x"), and <<`G` runs its body as a command. The body is a separate
// node after the statement, so Value is not carried and Heredoc says so.
// The controls: `<<>>` is a readline and `1 << 3` a shift.
func TestHeredocLowers(t *testing.T) {
	cases := []struct {
		src  string
		want lower.Node
	}{
		{"<<\"A\";\na $x\nA\n", &lower.StrLit{Interpolated: true, Heredoc: true}},
		{"<<B;\nb $x\nB\n", &lower.StrLit{Interpolated: true, Heredoc: true}},
		{"<<'C';\nc $x\nC\n", &lower.StrLit{Heredoc: true}},
		{"<<\\D;\nd $x\nD\n", &lower.StrLit{Heredoc: true}},
		{"<<~E;\n  e $x\n  E\n", &lower.StrLit{Interpolated: true, Heredoc: true}},
		{"<< \"F\";\nf $x\nF\n", &lower.StrLit{Interpolated: true, Heredoc: true}},
		{"<<`G`;\necho g\nG\n", &lower.StrLit{Interpolated: true, Backtick: true, Heredoc: true}},
		{"<<~'H';\n  h\n  H\n", &lower.StrLit{Heredoc: true}},
	}
	for _, c := range cases {
		root := parse.Parse([]byte(c.src))
		stmt := root.Children[0]
		if stmt.Kind != parse.Statement || len(stmt.Children) != 2 || !stmt.Children[1].HeredocBody {
			t.Fatalf("%q: want one statement holding the opener and its body", c.src)
		}
		n := stmt.Children[0]
		got := lower.LowerExpr(root, n)
		if got.Meta().ID != lower.AssignIDs(root)[n] {
			t.Errorf("%q: id %v is not the opener's id", c.src, got.Meta().ID)
		}
		sameShape(t, c.src, got, c.want)
	}

	// Each opener of `f(<<A, <<B)` is its own string.
	src := "f(<<A, <<'B');\na\nA\nb\nB\n"
	want := &lower.Call{Name: "f", Parenthesized: true, Args: []lower.Node{
		&lower.StrLit{Interpolated: true, Heredoc: true}, &lower.StrLit{Heredoc: true},
	}}
	root := parse.Parse([]byte(src))
	sameShape(t, src, zeroIDs(lower.LowerExpr(root, root.Children[0].Children[0])), want)

	for src, want := range map[string]lower.Node{
		"<<>>;":   &lower.Readline{Magic: true},
		"1 << 3;": &lower.Binary{Op: "<<", L: &lower.NumLit{Text: "1", IsInt: true}, R: &lower.NumLit{Text: "3", IsInt: true}},
	} {
		sameShape(t, src, zeroIDs(lowerOne(t, src)), want)
	}
}

// A number may start with its point, and a v-string is a string. perl 5.42:
// `.5 + 1` is 1.5 and `.5e1` is 5; `v5.36.0`, `5.36.0` and `v1.2` are
// strings of the characters numbered by their parts (`5.36.0 eq v5.36.0`,
// `length(1.2.3)` is 3, `ref \v5.36.0` is VSTRING); and `my $v = v65` is
// "A" when no sub v65 is declared before it. The controls: with `sub v65`
// declared first, `v65` calls it (perl prints "sub"), `v65 => 1` quotes
// the word v65, and 1.5 stays a number.
func TestLowerLeadingDotAndVStrings(t *testing.T) {
	cases := []struct {
		src  string
		want lower.Node
	}{
		{".5;", &lower.NumLit{Text: ".5"}},
		{".5e1;", &lower.NumLit{Text: ".5e1"}},
		{"v5.36.0;", &lower.StrLit{Value: "\x05\x24\x00"}},
		{"v1.2;", &lower.StrLit{Value: "\x01\x02"}},
		{"5.36.0;", &lower.StrLit{Value: "\x05\x24\x00"}},
		{"v65;", &lower.StrLit{Value: "A"}},
		{"1.5;", &lower.NumLit{Text: "1.5"}},
	}
	for _, c := range cases {
		sameShape(t, c.src, zeroIDs(lowerOne(t, c.src)), c.want)
	}

	src := "sub v65 {\"sub\"} v65;"
	if got := zeroIDs(lowerLast(t, src)); !reflect.DeepEqual(got, &lower.Call{Name: "v65", Resolved: true}) {
		t.Errorf("%q: the call lowered to %#v, want the Call v65", src, got)
	}
	src = "my %h = (v65 => 1);"
	root := parse.Parse([]byte(src))
	list := expression(t, root).Children[1]
	if got := zeroIDs(lower.LowerExpr(root, list.Children[0])); !reflect.DeepEqual(got, &lower.Bareword{Name: "v65", Quoted: true}) {
		t.Errorf("%q: the key lowered to %#v, want the quoted word v65", src, got)
	}
}
