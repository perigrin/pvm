// ABOUTME: A declaration that ends in its own block has finished its statement, so a
// ABOUTME: `for`, `if` or `while` on the next line starts a new one rather than modifying it.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestBlockDeclarationTakesNoModifier holds perl's reading of a sub, package
// block or lexical sub followed by a control-flow statement. Measured on
// 5.42.0, Deparse gives two siblings each time:
//
//	$ perl -MO=Deparse -e 'sub t { 1 } if ($x) { 2 }'
//	sub t { 1; }
//	if ($x) { '???'; }
//
// applyModifier stops at a `;` and at a Call ending in a block, but not at a
// declaration ending in one, so `sub test { ... }` followed by a loop read the
// loop keyword as a statement MODIFIER on the sub. comp/utf.t has exactly that
// shape, and its nested loops fell apart behind it.
func TestBlockDeclarationTakesNoModifier(t *testing.T) {
	for _, src := range []string{
		"sub t { 1 }\nfor my $b (0, 1) { f(); }\n",
		"sub t { 1 }\nif ($x) { 2 }\n",
		"sub t { 1 }\nwhile ($x) { 2 }\n",
		"package P { 1 }\nfor (1) { 2 }\n",
		"my sub t { 1 }\nforeach my $x (1) { 2 }\n",
	} {
		n := parse.Parse([]byte(src))
		if got := countUnknown(n); got != 0 {
			t.Errorf("Parse(%q) has %d Unknown, want 0", src, got)
		}
		if len(statements(n)) != 2 {
			t.Errorf("Parse(%q): %d statements, want 2 -- the loop is its own", src, len(statements(n)))
		}
		got := strings.TrimSpace(parse.Canon(n, []byte(src)))
		again := strings.TrimSpace(parse.Canon(parse.Parse([]byte(got)), []byte(got)))
		if again != got {
			t.Errorf("canon is not a fixpoint:\n  once  %s\n  twice %s", got, again)
		}
	}

	// REGRESSION GUARD: an expression ending in a block still takes one.
	src := []byte("$x = sub { 1 } if $y;")
	if got := strings.TrimSpace(parse.Canon(parse.Parse(src), src)); got != "$x = sub {1;} if $y;" {
		t.Errorf("Canon(%q) = %s, want the modifier kept", src, got)
	}
}

// statements is the root's children that are not trivia.
func statements(root *parse.Node) []*parse.Node {
	var out []*parse.Node
	for _, c := range root.Children {
		if c.Kind != parse.Trivia {
			out = append(out, c)
		}
	}
	return out
}
