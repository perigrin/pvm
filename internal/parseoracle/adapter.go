// ABOUTME: Adapts our tree-sitter parser onto the portable SubjectFacts contract.
// ABOUTME: All grammar-specific knowledge lives here, so the comparison core stays parser-agnostic.

package parseoracle

import (
	"bytes"
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
		CallSites:      treeSitterCallSites(root, tree.Source()),
		KnowsCallSites: true,
	}
}

// treeSitterCallSites translates the tree walk into per-call conclusions.
//
// Each site carries the line of the STATEMENT it sits in, not its own line,
// because that is the line perl attributes the matching op to: a nextstate
// names the first line of its statement, and every op until the next
// nextstate belongs to it. A `\@a` on the second line of a three-line call
// is perl's srefgen on the call's first line. Reporting the backslash's own
// line would put the two sides one line apart and turn agreement into a
// deficit.
//
// A statement is any `*_statement` node, plus `elsif`: perl gives an elsif
// condition a nextstate of its own (perly.y wraps it in newSTATEOP), so its
// sites belong to the elsif line rather than to the enclosing if.
func treeSitterCallSites(root *parser.Node, src []byte) []SubjectCallSite {
	var sites []SubjectCallSite

	// Pre-order visits statements in source order, so one running count of
	// newlines up to each statement's start yields its line in a single pass.
	pos, line := 0, 1
	var visit func(n *parser.Node, stmt int)
	visit = func(n *parser.Node, stmt int) {
		if isStatement(n) {
			start := int(n.StartByte())
			line += bytes.Count(src[pos:start], []byte{'\n'})
			pos = start
			stmt = line
		}
		switch {
		case isRefgen(n):
			// The source wrote `\`, so a reference here is accounted for.
			sites = append(sites, SubjectCallSite{Line: stmt, TookReference: true})
		case isHedgedCall(n):
			// The grammar emitted a node kind that names its own uncertainty.
			sites = append(sites, SubjectCallSite{Line: stmt, Unresolved: true})
		case isCommittedCall(n):
			sites = append(sites, SubjectCallSite{Line: stmt})
		}
		for i := 0; i < n.NamedChildCount(); i++ {
			visit(n.NamedChild(i), stmt)
		}
	}
	visit(root, 0)

	return sites
}

func isStatement(n *parser.Node) bool {
	kind := n.Kind()
	return strings.HasSuffix(kind, "_statement") || kind == "elsif"
}
