// ABOUTME: Pins what "committed" means after the grammar fork made every parenthesised call function_call_expression.
// ABOUTME: A call whose prototype the source never disambiguated is a hedge, not a commitment.

package parseoracle

import (
	"testing"

	"tamarou.com/pvm/internal/parser"
)

// TestPrototypedBuiltinIsNotACommitment is op/splice.t, reduced. Before the
// grammar fork `Internals::SvREADONLY(@a, 1)` parsed as
// ambiguous_function_call_expression and the file scored wider; after it,
// every parenthesised call is function_call_expression, and reading that as a
// commitment turns a prototype we cannot know about into a WRONG verdict.
//
// We did not get this parse wrong. We cannot know that SvREADONLY has a
// `\[$@%]` prototype without having seen its definition, which is the whole
// reason the wider bucket exists. A plain `f(...)` must therefore stay a
// hedge; only a form the SOURCE disambiguated -- `&f(@a)`, a method call --
// is a commitment.
func TestPrototypedBuiltinIsNotACommitment(t *testing.T) {
	tree, err := parser.New().Parse([]byte("my @a = 10..11;\nInternals::SvREADONLY(@a, 1);\n"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	var hedged, committed int
	walk(treeRoot(tree), func(n *parser.Node) {
		switch {
		case isHedgedCall(n):
			hedged++
		case isCommittedCall(n):
			committed++
		}
	})

	if committed != 0 {
		t.Errorf("committed: got %d, want 0 -- a bare f(...) commits to nothing "+
			"the source disambiguated", committed)
	}
	if hedged != 1 {
		t.Errorf("hedged: got %d, want 1 -- an unresolvable prototype is what "+
			"wider exists to report", hedged)
	}
}

// TestAmpersandCallIsACommitment guards the other direction: WRONG must stay
// reachable. `&f(@a)` bypasses the prototype -- perl documents it as doing so
// -- so the source HAS settled the parse, and a disagreement there is real.
func TestAmpersandCallIsACommitment(t *testing.T) {
	tree, err := parser.New().Parse([]byte("sub f {}\n&f(@a);\n"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	var committed int
	walk(treeRoot(tree), func(n *parser.Node) {
		if isCommittedCall(n) {
			committed++
		}
	})
	if committed != 1 {
		t.Errorf("committed: got %d, want 1 -- &f(...) is the settled form", committed)
	}
}
