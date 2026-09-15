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
// An explicit backslash is reported as a reference site --
// SiteKindReference, the contract's name for a reference taken outside a
// call -- so a subject need never invent a call site to report one.
//
// Each site carries the span of the STATEMENT it sits in, not its own line,
// because a statement is the unit perl attributes the matching op to: every
// op until the next nextstate belongs to that nextstate. Which line of the
// statement the nextstate records is not fixed, and that is why the whole
// span is reported rather than the first line. Measured on perl 5.42.0: a
// plain multi-line call, a hash literal and an if/elsif/for/while condition
// take the FIRST line; a statement whose term carries a block -- op/qr.t's
// sub { ... } followed by a deref arrow -- takes the line the statement
// ENDS on, because the block resets the parser's copline and newSTATEOP
// then uses the lexer's current line. A reference anywhere inside a
// statement is perl's srefgen somewhere inside the same statement, and the
// span is what the two sides share.
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
			site.Kind = SiteKindReference
			site.TookReference = true
			sites = append(sites, site)
		case isHedgedCall(n):
			// The grammar emitted a node kind that names its own uncertainty.
			site.Unresolved = true
			sites = append(sites, site)
		case isCommittedCall(n):
			sites = append(sites, site)
		case isHashAccess(n):
			site.Kind = SiteKindHash
			sites = append(sites, site)
		case isMatch(n):
			site.Kind = SiteKindMatch
			sites = append(sites, site)
		case isReadline(n):
			site.Kind = SiteKindReadline
			sites = append(sites, site)
		case isAnonHash(n):
			site.Kind = SiteKindAnonhash
			sites = append(sites, site)
		}
		for i := 0; i < n.NamedChildCount(); i++ {
			visit(n.NamedChild(i), stmt)
		}
	}
	visit(root, SubjectCallSite{})

	return sites
}

// The four predicates below are the tree-sitter side of spec §7.5.4's
// marker table. Each names the node kinds our grammar emits for a construct
// perl compiles to the marker op, and every kind was read off `psc parse
// --format sexpr` on the sample lines in markers_test.go rather than
// assumed. They may over-report relative to perl -- a lexical `%h` is padhv,
// `$h{a}` folds to multideref, a match under `if (0)` is discarded -- and
// that is the safe direction: the comparison consults only perl's sites, so
// a subject site with no perl op behind it accounts for nothing and costs
// nothing. What they must never do is under-report a construct perl does
// emit the op for, because that is scored WRONG.

// isHashAccess: `%h`, `%$r`, `%{...}`, `->%*`, `$h{a}`, `$$r{a}`, `$r->{a}`,
// a hash slice `@h{...}`/`@$r{...}`/`->@{...}` and a key/value slice
// `%h{...}`. slice_expression and keyval_expression serve array slices too,
// and the subscript's opening token says which container was sliced.
func isHashAccess(n *parser.Node) bool {
	switch n.Kind() {
	case "hash", "hash_deref_expression", "hash_element_expression":
		return true
	case "slice_expression", "keyval_expression":
		return hasToken(n, "{")
	}
	return false
}

// isMatch: a regex literal `/.../` or `m//`, or a `=~`/`!~` binding to
// anything that is not s/// or tr/// (measured: a scalar, a string and a
// qr// on the right all compile to regcomp + match). A binding whose right
// side is itself a match_regexp is not counted twice: the literal is the
// site, the binding merely names its target.
func isMatch(n *parser.Node) bool {
	switch n.Kind() {
	case "match_regexp":
		return true
	case "binary_expression":
		if !hasToken(n, "=~") && !hasToken(n, "!~") {
			return false
		}
		if rhs := n.NamedChild(n.NamedChildCount() - 1); rhs != nil {
			switch rhs.Kind() {
			case "match_regexp", "substitution_regexp", "transliteration_expression":
				return false
			}
		}
		return true
	}
	return false
}

// isReadline: `<FH>`, `<$fh>`, `<>`, `<<>>`, and the readline builtin with
// or without parentheses. `<*.c>` is a fileglob_expression and is not one.
func isReadline(n *parser.Node) bool {
	switch n.Kind() {
	case "readline_expression":
		return true
	case "func1op_call_expression":
		return hasToken(n, "readline")
	}
	return false
}

// isAnonHash: `{ ... }` the grammar read as a hash constructor rather than
// a block, including the empty and the `+{ }` forms.
func isAnonHash(n *parser.Node) bool {
	return n.Kind() == "anonymous_hash_expression"
}

// hasToken reports whether one of n's anonymous children is the literal
// token. Operators, sigils and builtin names are anonymous in this grammar,
// so `$a =~ $b` and `$a % $b` are both binary_expression and only the token
// tells them apart.
func hasToken(n *parser.Node, token string) bool {
	for i := 0; i < n.ChildCount(); i++ {
		if n.Child(i).Kind() == token {
			return true
		}
	}
	return false
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
