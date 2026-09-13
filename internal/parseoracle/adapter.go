// ABOUTME: Adapts our tree-sitter parser onto the portable SubjectFacts contract.
// ABOUTME: All grammar-specific knowledge lives here, so the comparison core stays parser-agnostic.

package parseoracle

import (
	"fmt"
	"sort"
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
// Each site carries the span of the STATEMENT it sits in, not its own line,
// because a statement is the unit perl attributes the matching op to: every
// op until the next nextstate belongs to that nextstate. Which line of the
// statement the nextstate records is not fixed, and that is why the whole
// span is reported rather than the first line. Measured on perl 5.42.0: a
// plain multi-line call, a hash literal and an if/elsif/for/while condition
// take the FIRST line; a statement whose term carries a block -- op/qr.t's
// `sub { ... }\n ->((\my%hash)->{key})` -- takes the line the statement
// ENDS on, because the block resets the parser's copline and newSTATEOP
// then uses the lexer's current line. A `\@a` anywhere inside a statement
// is perl's srefgen somewhere inside the same statement, and the span is
// what the two sides share.
//
// A statement is any `*_statement` node, plus `elsif`: perl gives an elsif
// condition a nextstate of its own (perly.y wraps it in newSTATEOP), so its
// sites belong to the elsif rather than to the enclosing if.
func treeSitterCallSites(root *parser.Node, src []byte) []SubjectCallSite {
	var sites []SubjectCallSite
	lines := newLineIndex(src)

	var visit func(n *parser.Node, stmt SubjectCallSite)
	visit = func(n *parser.Node, stmt SubjectCallSite) {
		if isStatement(n) {
			stmt = SubjectCallSite{
				Line:    lines.of(int(n.StartByte())),
				EndLine: lines.of(int(n.EndByte()) - 1),
			}
		}
		site := stmt
		switch {
		case isRefgen(n):
			// The source wrote `\`, so a reference here is accounted for.
			site.TookReference = true
			sites = append(sites, site)
		case isHedgedCall(n):
			// The grammar emitted a node kind that names its own uncertainty.
			site.Unresolved = true
			sites = append(sites, site)
		case isCommittedCall(n):
			sites = append(sites, site)
		}
		for i := 0; i < n.NamedChildCount(); i++ {
			visit(n.NamedChild(i), stmt)
		}
	}
	visit(root, SubjectCallSite{})

	return sites
}

// lineIndex turns a byte offset into a 1-based line. Statement ends are not
// visited in source order -- an outer statement ends after its inner ones
// begin -- so a running count will not do; the newline offsets are indexed
// once and searched.
type lineIndex []int

func newLineIndex(src []byte) lineIndex {
	var nl lineIndex
	for i, b := range src {
		if b == '\n' {
			nl = append(nl, i)
		}
	}
	return nl
}

func (nl lineIndex) of(offset int) int {
	return sort.SearchInts(nl, offset) + 1
}

func isStatement(n *parser.Node) bool {
	kind := n.Kind()
	return strings.HasSuffix(kind, "_statement") || kind == "elsif"
}
