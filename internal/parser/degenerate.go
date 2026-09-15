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
// This reports on the TREE, not on perl's verdict. It is not a syntax check:
// a true result means "this tree is not a faithful parse", which is strictly
// weaker than "perl would reject this". Measured over 3306 cleanly-parsed
// files from a perl5 checkout, 2 flag true and both compile -- consecutive
// `sub NAME;` forward declarations, where the grammar also drops the `sub`
// keyword. So the tree really is degraded in those files too; only the
// inference to "malformed source" would be wrong. See
// TestForwardDeclarationsAlsoLeak and docs/specs/perl-parser/00-findings.md
// §0.4.1. Callers should treat a true result as "do not trust this tree",
// never as a diagnostic to show a user.
//
// ponytail: kind-prefix check, not a grammar reimplementation. If a future
// grammar version starts naming hidden rules without recovering, or fixes the
// gap so real error nodes appear, TestDroppedRHSFamilyHasNoErrorNode fails and
// points here.
func (t *Tree) IsDegenerate() bool {
	return len(t.DegenerateKinds()) > 0
}

// DegenerateKinds returns the sorted, deduplicated names of the rules that
// contradicted themselves in this tree, empty for a clean parse. Callers
// report these so a fidelity run can say WHAT it saw rather than only that
// something was wrong.
//
// Two signals feed it, and they are the same kind of claim about the tree
// rather than about Perl. A hidden rule surfacing (`_term`) is a rule that is
// never supposed to be a node; a `varname` whose text is not a name is a node
// whose contents deny its own rule. See isCollapsedVarname for the second.
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
		if t.isCollapsedVarname(n) {
			seen[n.Kind()] = struct{}{}
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

// isCollapsedVarname reports whether this node is a `varname` that holds an
// explicit reference rather than a name.
//
// The grammar mis-parses a deref block whose entire content is one `\` against
// a sigilled variable. `@{ \@a }` comes back as an `array` whose braces are
// anonymous tokens and whose `varname` child spans the literal text `\@a`; the
// `block` node that a correct parse builds is absent, and with it the refgen
// inside. Perl compiles that source and its optree holds one srefgen, so a
// consumer reading this tree sees an array named `\@a` where the source took a
// reference. Neither HasError nor the hidden-rule check fires.
//
// The test is deliberately the narrowest thing that separates it, because the
// grammar gets every neighbouring shape right and flagging those would cost a
// sweep nearly every dereference in a corpus. `@{ $r }`, `@$r`, `$r->@*` and
// `$$r[0]` all build the block. So do `@{ \@a, }`, `@{ +\@a }` and even
// `@{ \ @a }` -- the collapse needs the backslash tight against the sigil and
// nothing else in the braces. A varname with a child is therefore already
// correct.
//
// The sigil after the backslash is load-bearing, not decoration. `$\` is the
// output record separator, and the grammar parses it correctly into a varname
// whose text is a lone backslash. A prefix test alone flags that, which is a
// false positive on valid Perl -- measured, as two corpus files moving bucket
// for no reason (t/op/tiehandle.t, t/uni/lex_utf8.t). Requiring `\` followed
// by a sigil separates the punctuation variable from the collapse.
//
// This says nothing about whether perl accepts the source: every case it flags
// compiles. It is the same weaker claim IsDegenerate is documented to make,
// "this tree is not a faithful parse", which is why the honest verdict for one
// is no-answer rather than a comparison nobody can trust.
//
// ponytail: a text test on one node kind, not a grammar reimplementation. The
// real fix belongs in gotreesitter's compiled grammar table, which is outside
// this repo. If it lands, TestDerefOfExplicitRefCollapses fails and points here.
func (t *Tree) isCollapsedVarname(n *Node) bool {
	if n == nil || n.Kind() != "varname" || n.NamedChildCount() > 0 {
		return false
	}
	text := n.Text(t.source)
	if !strings.HasPrefix(text, `\`) || len(text) < 2 {
		return false
	}
	return strings.ContainsRune(`$@%&*`, rune(text[1]))
}
