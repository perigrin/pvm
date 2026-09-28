// ABOUTME: A block that is the first statement of another block is a block, not an anonymous hash.
// ABOUTME: The lexer leaves a statement boundary after a block's `{`, as perl's yyl_leftcurly does.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestNestedBareBlock holds a block opened as the first statement of another.
//
// The lexer's expect machine returned XTerm after EVERY operator, including a
// `{` that had just opened a block, so the first token inside a block stood in
// term position and a `{` there was classified as an anonymous hash. The
// inner block then closed after its own first statement:
//
//	sub r { { $s = 1; $x = 2; } }   ->   sub r {{$s = 1};$x = 2;} }
//
// perl's yyl_leftcurly sets PL_expect = XSTATE after a block brace
// (toke.c:6688-6697), so the first thing inside is at a statement boundary --
// where a `{` is a block unless intuit_curly's lookahead says hash.
// re/pat_advanced.t opens `{ { ... } }` dozens of times.
func TestNestedBareBlock(t *testing.T) {
	cases := []struct {
		name  string
		src   string
		canon string
	}{
		{"in a sub", "sub r {\n {\n  $s = 1;\n  $x = 2;\n }\n}\n", "sub r {{$s = 1;$x = 2;} }"},
		{"in an if", "if (1) {\n {\n  $s = 1;\n  $x = 2;\n }\n}\n", "if (1) {{$s = 1;$x = 2;} }"},
		{"in a bare block", "{\n {\n  $s = 1;\n  $x = 2;\n }\n}\n", "{{$s = 1;$x = 2;} }"},
		{"in an anonymous sub", "my $c = sub {\n {\n  $s = 1;\n  $x = 2;\n }\n};\n", "my $c = sub {{$s = 1;$x = 2;} };"},
		// 01a0ddac-bd24's shape: a bare block as a class body's first member,
		// with the method after it lost.
		{"in a class body", "class C { { 1 } method m { 1 } }\n", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := []byte(c.src)
			n := parse.Parse(src)
			if got := countUnknown(n); got != 0 {
				t.Errorf("Parse(%q) has %d Unknown, want 0", c.src, got)
			}
			got := strings.TrimSpace(parse.Canon(n, src))
			if c.canon != "" && got != c.canon {
				t.Errorf("Canon(%q):\n  got  %s\n  want %s", c.src, got, c.canon)
			}
			again := strings.TrimSpace(parse.Canon(parse.Parse([]byte(got)), []byte(got)))
			if again != got {
				t.Errorf("canon is not a fixpoint:\n  once  %s\n  twice %s", got, again)
			}
		})
	}

	// NEGATIVE: an anonymous hash as a block's first statement is still a
	// hash. perl runs intuit_curly at a statement-start brace, and a bareword
	// then `=>` says hash -- measured, `sub f { {a=>1} }` deparses the inner
	// braces as `+{'a', 1}`.
	src := []byte("sub f { {a => 1} }\n")
	var sawHash bool
	var walk func(*parse.Node)
	walk = func(n *parse.Node) {
		if n.Kind == parse.AnonHash {
			sawHash = true
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(parse.Parse(src))
	if !sawHash {
		t.Errorf("Parse(%q): the inner braces are a hash per intuit_curly, and no AnonHash was built", src)
	}
}

// TestBareBlockInClassBody holds issue 01a0ddac-bd24's shape and the
// neighbours its scope named. A class body is a block, so a bare block as its
// first member was the same misreading as any block's first statement, and
// the method after it was lost. Each source runs on 5.42.0 and the method
// named last is reached:
//
//	$ perl -e 'use feature "class"; no warnings;
//	      class C { { 1 } method m { 2 } } print C->new->m'
//	2
func TestBareBlockInClassBody(t *testing.T) {
	for _, src := range []string{
		"class C { { 1 } method m { 2 } }\n",
		"class C { field $x = 3; { 1 } method m { $x } }\n",
		"class C { method m { 4 } { 1 } }\n",
		"class C { { 1 } { 2 } method m { 5 } }\n",
	} {
		n := parse.Parse([]byte("use feature 'class';\n" + src))
		if got := countUnknown(n); got != 0 {
			t.Errorf("Parse(%q) has %d Unknown, want 0", src, got)
		}
	}
}
