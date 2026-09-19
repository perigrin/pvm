// ABOUTME: Canon emits Perl from a tree, parenthesised the way the TREE groups.
// ABOUTME: The parens it writes are what makes the comparison against source mean anything.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestCanonEmitsTreeParens: the emission's grouping comes from the tree.
//
// An emitter that prints each leaf's original text in order is SourceText
// with extra steps -- it round-trips by construction and proves nothing. The
// parens below are the whole point: Canon must write one wherever a child
// binds looser than its parent, whether or not the source had it, and must
// NOT write one where the tree's own shape already says the grouping.
func TestCanonEmitsTreeParens(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		// Precedence alone: * binds tighter, so the tree needs no paren
		// and Canon must not invent one.
		{"1 + 2 * 3;", "1 + 2 * 3;"},

		// The source's paren changed the grouping, so the tree holds it and
		// Canon must write it back.
		{"(1 + 2) * 3;", "(1 + 2) * 3;"},

		// A paren the source had but the tree did not need. Canon prints
		// the TREE, so the redundant paren is gone.
		{"1 + (2 * 3);", "1 + 2 * 3;"},

		// Left associative: the right operand at equal power needs a paren,
		// the left one does not. `1 - 2 - 3` is (1-2)-3.
		{"1 - 2 - 3;", "1 - 2 - 3;"},
		{"1 - (2 - 3);", "1 - (2 - 3);"},

		// Right associative: the mirror image. `2 ** 3 ** 2` is 2**(3**2),
		// so the right operand needs no paren and the left one does.
		{"2 ** 3 ** 2;", "2 ** 3 ** 2;"},
		{"(2 ** 3) ** 2;", "(2 ** 3) ** 2;"},
	} {
		got := parse.Canon(parse.Parse([]byte(tc.src)), []byte(tc.src))
		if strings.TrimSpace(got) != tc.want {
			t.Errorf("Canon(parse(%q)):\n  got  %q\n  want %q", tc.src, strings.TrimSpace(got), tc.want)
		}
	}
}

// TestCanonCatchesMisgrouping: a tree that groups differently from its source
// re-emits differently, and round-trip does not notice.
//
// The tree is built by hand rather than parsed, because the parser gets this
// case right -- and a check whose failure mode cannot be produced is not a
// check. `print (1+2)*3` is the case chapter 7 names: perl reads it as
// `print(1+2) * 3`, and a parser that read it as `print((1+2)*3)` has the
// same leaves in the same order.
//
// That is why this test asserts BOTH halves. SourceText agrees with the
// wrong tree; Canon does not. The two checks are disjoint, not weaker and
// stronger versions of one thing.
func TestCanonCatchesMisgrouping(t *testing.T) {
	src := []byte("print (1+2)*3;")

	// The wrong tree: print swallowing the whole product.
	//
	//	statement
	//	  call print          span: the whole expression
	//	    binary *
	//	      binary +
	//	        term 1
	//	        term 2
	//	      term 3
	mul := &parse.Node{Kind: parse.Binary, Text: "*", Start: 6, End: 13, Children: []*parse.Node{
		{Kind: parse.Binary, Text: "+", Start: 7, End: 10, Children: []*parse.Node{
			{Kind: parse.Term, Text: "1", Start: 7, End: 8},
			{Kind: parse.Term, Text: "2", Start: 9, End: 10},
		}},
		{Kind: parse.Term, Text: "3", Start: 12, End: 13},
	}}
	wrong := &parse.Node{Kind: parse.SourceFile, Start: 0, End: len(src), Children: []*parse.Node{
		{Kind: parse.Statement, Start: 0, End: len(src), Children: []*parse.Node{
			{Kind: parse.Call, Text: "print", Resolved: true, Start: 0, End: 13, Children: []*parse.Node{mul}},
		}},
	}}

	// Round-trip is blind to it: same leaves, same order, every byte back.
	if got := wrong.SourceText(src); got != string(src) {
		t.Fatalf("the wrong tree must still round-trip, or it is not the case "+
			"this test is about:\n  got  %q\n  want %q", got, string(src))
	}

	// Canon is not blind to it, because it writes its OWN grouping.
	got := strings.TrimSpace(parse.Canon(wrong, src))
	right := strings.TrimSpace(parse.Canon(parse.Parse(src), src))
	if got == right {
		t.Errorf("Canon does not distinguish the mis-grouped tree from the correct one; "+
			"both emitted %q", got)
	}
}
