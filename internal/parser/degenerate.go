// ABOUTME: Detects trees where error recovery dropped source without emitting an error node.
// ABOUTME: Makes a known grammar limitation visible instead of silently scoring as a clean parse.

package parser

import (
	"sort"
	"strings"
)

// IsDegenerate reports whether this tree is structurally degenerate: the
// grammar accepted the source and produced no error node, but the tree it
// built is not a parse of that source.
//
// This exists because the grammar has a documented gap. `perl -c` rejects all
// of these, and the grammar reports HasError() == false for every one:
//
//	my $x = ;    my @a = ;    my %h = ;    $x = ;
//	1 +;         my $y = 1 +;              my ($a) = ;
//
// The right-hand side is dropped with no diagnostic. `my $y = 1 +;` is the
// clearest case: it comes back as TWO sibling statements with the `+` gone
// entirely, so a consumer reading that tree sees two unrelated terms where the
// source wrote one expression.
//
// The signal is not a guess about Perl semantics -- reimplementing perl's
// expression grammar to second-guess the parser would be writing the parser
// twice. It is a statement about the TREE: a node kind beginning with `_` is a
// hidden tree-sitter rule (`_term` here, tree-sitter-perl's expression
// supertype). Hidden rules are inlined into their parents in a successful
// parse and are never supposed to surface as a node. When one does, recovery
// has bailed out mid-rule and kept the fragment it had. That makes the
// underscore an artifact of the failure rather than a property of the source,
// which is why it separates these cases from valid Perl that merely looks
// similar: `return ;` compiles, and parses to a real `return_expression`.
//
// ponytail: kind-prefix check, not a grammar reimplementation. If a future
// grammar version starts naming hidden rules without recovering, or fixes the
// gap so real error nodes appear, TestDroppedRHSFamilyHasNoErrorNode fails and
// points here.
func (t *Tree) IsDegenerate() bool {
	return len(t.DegenerateKinds()) > 0
}

// DegenerateKinds returns the sorted, deduplicated hidden rule names that
// surfaced in this tree, empty for a clean parse. Callers report these so a
// fidelity run can say WHICH internal rule leaked rather than only that
// something was wrong.
func (t *Tree) DegenerateKinds() []string {
	root := t.RootNode()
	if root == nil {
		return nil
	}

	seen := map[string]struct{}{}
	var walk func(*Node)
	walk = func(n *Node) {
		if n == nil {
			return
		}
		// Anonymous nodes are punctuation and carry no rule name worth
		// reporting; only a named node leaking a hidden rule is a signal.
		if k := n.Kind(); n.IsNamed() && strings.HasPrefix(k, "_") {
			seen[k] = struct{}{}
		}
		for i := 0; i < n.NamedChildCount(); i++ {
			walk(n.NamedChild(i))
		}
	}
	walk(root)

	if len(seen) == 0 {
		return nil
	}
	kinds := make([]string, 0, len(seen))
	for k := range seen {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	return kinds
}
