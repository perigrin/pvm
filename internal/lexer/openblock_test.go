// ABOUTME: A block's opening brace leaves a statement boundary, so a brace right after it is a block.
// ABOUTME: perl's yyl_leftcurly sets PL_expect = XSTATE after a block brace; this holds the same.

package lexer

import "testing"

// TestOpenBlockLeavesStatementState holds the state a block's `{` leaves.
//
// after(Operator, ...) returned XTerm for every operator, a block's `{`
// included, so the first token inside a block was lexed in term position. A
// second `{` there was classified from XTerm -- an anonymous hash -- and its
// `}` reported a closed subscript. perl sets PL_expect = XSTATE after a block
// brace in each of yyl_leftcurly's block cases (toke.c:6688-6697).
//
// The second brace of each source below opens a block, measured by perl
// running the inner statements as statements.
func TestOpenBlockLeavesStatementState(t *testing.T) {
	for _, src := range []string{
		"{ { $s = 1; $x = 2; } }",
		"sub r { { $s = 1; } }",
		"if (1) { { $s = 1; } }",
	} {
		opens, found := opensBlockAt(src, 1)
		if !found {
			t.Fatalf("%q: no second brace", src)
		}
		if !opens {
			t.Errorf("%q: the second `{` is a block's first statement and must open a block", src)
		}
	}

	// A SUBSCRIPT brace is not a block brace and must not leave XState: the
	// `{k}` in `$h{k}{j}` continues a chain.
	if opens, _ := opensBlockAt("$h{k}{j} = 1;", 1); opens {
		t.Errorf("the second brace of `$h{k}{j}` is a subscript, not a block")
	}
}
