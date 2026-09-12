// ABOUTME: Adapts our tree-sitter parser onto the portable SubjectFacts contract.
// ABOUTME: All grammar-specific knowledge lives here, so the comparison core stays parser-agnostic.

package parseoracle

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/parser"
)

// TreeSitterSubject derives SubjectFacts from our own parser.
//
// This is the only place in the harness that knows a tree-sitter node kind.
// The comparison core (compare_facts.go) sees facts alone, which is what lets
// a subject written in Rust, Java or Perl be measured on equal terms. When the
// hand-written Go parser of the specification exists, it gets an adapter of
// its own beside this one and nothing else changes.
//
// It is an in-process shortcut rather than a subprocess: our parser is already
// linked in, and spawning a copy of ourselves to ask it a question we can ask
// directly would cost the sweep a process per file for nothing. The contract
// is the same either way — that is the point of defining it as data.
func TreeSitterSubject(tree *parser.Tree) SubjectFacts {
	root := treeRoot(tree)
	if root == nil {
		return SubjectFacts{
			OK:             false,
			Declined:       true,
			DeclinedReason: "our parser produced no tree",
		}
	}

	if root.HasError() {
		return SubjectFacts{
			OK:             false,
			Declined:       true,
			DeclinedReason: "our parse contains an error node",
		}
	}

	// A missing error node is not the same as a parse. Two different failures
	// reach this point, and the wording has to cover both without claiming
	// either. The grammar accepts `my $x = ;` -- which perl rejects -- by
	// dropping the right-hand side and reporting no error; it accepts
	// `@{ \@a }` -- which perl compiles -- by keeping every byte and building
	// a `varname` that holds a reference instead of a name. One lost source,
	// the other lost structure. Comparing against either measures nothing, so
	// we decline, which keeps the limitation in the coverage number instead of
	// inflating the fidelity one.
	//
	// Note this is the tree-sitter-specific FORM of a general question. The
	// contract asks "is this tree a parse of this source"; a rule contradicting
	// itself is how this parser answers no. A hand-written parser has no hidden
	// rules to leak and will answer it some other way.
	if kinds := tree.DegenerateKinds(); len(kinds) > 0 {
		return SubjectFacts{
			OK:       false,
			Declined: true,
			DeclinedReason: fmt.Sprintf(
				"our parse is not a parse of this source, and no error node says so "+
					"(rule%s contradicting itself: %s)",
				plural(len(kinds)), strings.Join(kinds, ", ")),
		}
	}

	return SubjectFacts{
		OK:             true,
		CallSites:      treeSitterCallSites(root),
		KnowsCallSites: true,
	}
}

// treeSitterCallSites translates the tree walk into per-call conclusions.
//
// The old Compare counted three things across the whole tree: explicit refgen
// nodes, hedged calls, and committed calls. The contract asks the same
// questions per call site, so an explicit `\` is reported as a call site that
// took a reference. Totals are what the comparison uses, and they are
// preserved exactly.
func treeSitterCallSites(root *parser.Node) []SubjectCallSite {
	var sites []SubjectCallSite

	walk(root, func(n *parser.Node) {
		switch {
		case isRefgen(n):
			// The source wrote `\`, so a reference here is accounted for.
			sites = append(sites, SubjectCallSite{TookReference: true})
		case isHedgedCall(n):
			// The grammar emitted a node kind that names its own uncertainty.
			sites = append(sites, SubjectCallSite{Unresolved: true})
		case isCommittedCall(n):
			sites = append(sites, SubjectCallSite{})
		}
	})

	return sites
}
