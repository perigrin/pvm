// ABOUTME: Pins the trees behind every §4.14.2 status that moved off the 3c997a6b
// ABOUTME: baseline, so the spec's status table and the parser cannot drift apart.
package parse_test

import (
	"fmt"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestStatusTableAtHead measures the fragments chapter 4 §4.14.2 and §4.14.3
// cite for each row whose status changed when the table was re-measured,
// and asserts the tree the table now quotes.
//
// It pins reality rather than asking for it: each want is what the parser
// produced when the table was written. A row that still refuses is not
// pinned here (perigrin, 2026-10-11), because that would turn this test red
// whenever another milestone's parser work closed the row. When a pinned
// tree changes, the §4.14.2 row that quotes it changes in the same commit.
func TestStatusTableAtHead(t *testing.T) {
	cases := []struct {
		row      string
		src      string
		unknowns int
		tree     string
	}{
		{"StrLit.Backtick", "qx/ls/;", 0,
			`[statement [term "qx/ls/"]]`},
		{"Var", "$::x;", 0,
			`[statement [term "$::x"]]`},
		{"Deref", "${$x};", 0,
			`[statement [unary "$" [term "$x"]]]`},
		{"Deref", "$$x;", 0,
			`[statement [unary "$" [term "$x"]]]`},
		{"Deref", "@$x;", 0,
			`[statement [unary "@" [term "$x"]]]`},
		{"Deref", "%$x;", 0,
			`[statement [unary "%" [term "$x"]]]`},
		{"Deref", `*{"foo"};`, 0,
			`[statement [term "*" [term "\"foo\""]]]`},
		{"Deref.Postfix", "$x->$#*;", 0,
			`[statement [binary "->" [term "$x"] [term "$#*"]]]`},
		{"Deref.Postfix", "$x->@[0,1];", 0,
			`[statement [binary "->" [term "$x"] [index "[" [term "@"] [binary "," [term "0"] [term "1"]]]]]`},
		{"Index.Arrow", "$h->{k};", 0,
			`[statement [index "{" arrow [term "$h"] [term "k"]]]`},
		{"Call.Fn", "&$code(1);", 0,
			`[statement [index "(" [term "&$code"] [term "1"]]]`},
		{"Repeat.ListRepeat", `("a") x 3;`, 0,
			`[statement [binary "x" [term "\"a\"" paren] [term "3"]]]`},
		{"Range", "1..2;", 0,
			`[statement [binary ".." [term "1"] [term "2"]]]`},
		{"List.Fat", "(a => 1, b => 2);", 0,
			`[statement [list paren [call "a" fat] [term "1"] [call "b" fat] [term "2"]]]`},
		{"Paren", "($x);", 0,
			`[statement [term "$x" paren]]`},
		{"Decl.Attrs", "my $x :shared;", 0,
			`[statement [declaration "my" [term "$x"] [attribute ":shared"]]]`},
		{"sub attributes", "sub f :prototype($) {1}", 0,
			`[statement [declaration "sub" [term "f"] [attribute ":prototype($)"] [block [statement [term "1"]]]]]`},
		{"signatures", "use v5.36; sub f ($x, $y) { 1 }", 0,
			`[statement [use "use" [term "v5.36"]]] [statement [declaration "sub" [term "f"] [list paren [term "$x"] [term "$y"]] [block [statement [term "1"]]]]]`},
		{"Call.Handle", `print $fh "x";`, 0,
			`[statement [call "print" resolved [term "$fh" handle] [term "\"x\""]]]`},
		{"Call.Handle", `print {$fh} "x";`, 0,
			`[statement [call "print" resolved [block handle [statement [term "$fh"]]] [term "\"x\""]]]`},
		{"Call args", "f(1,2);", 0,
			`[statement [call "f" [binary "," [term "1"] [term "2"]]]]`},
		{"MethodCall.Indirect", "package Foo; sub new {1} package main; new Foo(1);", 0,
			`[statement [declaration "package" [term "Foo"]]] [statement [declaration "sub" [term "new"] [block [statement [term "1"]]]]] [statement [declaration "package" [term "main"]]] [statement [call "new" resolved indirect [term "Foo"] [term "1"]]]`},
		{"MapGrepSort", "map { $_ } @a;", 0,
			`[statement [call "map" resolved [block [statement [term "$_"]]] [term "@a"]]]`},
		{"MapGrepSort", "grep { $_ } @a;", 0,
			`[statement [call "grep" resolved [block [statement [term "$_"]]] [term "@a"]]]`},
		{"MapGrepSort", "sort { $a <=> $b } @a;", 0,
			`[statement [call "sort" resolved [block [statement [binary "<=>" [term "$a"] [term "$b"]]]] [term "@a"]]]`},
		{"MapGrepSort comparator", "sort byname @a;", 0,
			`[statement [call "sort" resolved [term "byname" comparator] [term "@a"]]]`},
		{"DoFile", "do 'file.pl';", 0,
			`[statement [call "do" resolved [term "'file.pl'"]]]`},
		{"AnonSub", "sub { 1 };", 0,
			`[statement [declaration "sub" [block [statement [term "1"]]]]]`},
		{"AnonSub.Sig", "use v5.36; my $f = sub ($x) { 1 };", 0,
			`[statement [use "use" [term "v5.36"]]] [statement [declaration "my" [term "$f"] [declaration "sub" [term "$x" paren] [block [statement [term "1"]]]]]]`},
		{"AnonHash", "{a=>1};", 0,
			`[statement [anon_hash [call "a" fat] [term "1"]]]`},
		{"Readline.Magic", "<<>>;", 0,
			`[statement [term "<<>>"]]`},
		{"Glob", "<*.c>;", 0,
			`[statement [term "<*.c>"]]`},
		{"LoopEx", "goto &foo;", 0,
			`[statement [loop_control "goto" [term "&foo"]]]`},
	}
	for _, c := range cases {
		root := parse.Parse([]byte(c.src))
		if got := countUnknown(root); got != c.unknowns {
			t.Errorf("%s: %q has %d Unknown, the table says %d\n%s",
				c.row, c.src, got, c.unknowns, statusTree(root))
		}
		if got := statusTree(root); got != c.tree {
			t.Errorf("%s: %q\n got  %s\n want %s", c.row, c.src, got, c.tree)
		}
	}
}

// statusTree renders the statements under root the way §4.14.2 quotes a
// tree: kind, then Text, then each flag that is set, then the children.
// Trivia is dropped, since no row of the table quotes it.
func statusTree(root *parse.Node) string {
	var b strings.Builder
	var walk func(*parse.Node)
	walk = func(n *parse.Node) {
		b.WriteString("[" + n.Kind.String())
		if n.Text != "" {
			fmt.Fprintf(&b, " %q", n.Text)
		}
		for _, f := range []struct {
			set  bool
			name string
		}{
			{n.Resolved, "resolved"}, {n.Arrow, "arrow"}, {n.Paren, "paren"},
			{n.Fat, "fat"}, {n.Handle, "handle"}, {n.Comparator, "comparator"},
			{n.Indirect, "indirect"},
		} {
			if f.set {
				b.WriteString(" " + f.name)
			}
		}
		for _, c := range n.Children {
			if c.Kind != parse.Trivia {
				b.WriteString(" ")
				walk(c)
			}
		}
		b.WriteString("]")
	}
	var parts []string
	for _, c := range root.Children {
		if c.Kind == parse.Trivia {
			continue
		}
		b.Reset()
		walk(c)
		parts = append(parts, b.String())
	}
	return strings.Join(parts, " ")
}
