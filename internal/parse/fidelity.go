// ABOUTME: Faithful compares canon's emission against the SOURCE, not against itself.
// ABOUTME: Two differences are forgiven: a block's last semicolon, and parens nested inside a node's own pair.

package parse

import (
	"fmt"
	"slices"

	"tamarou.com/pvm/internal/lexer"
)

// Faithful reports whether canon(n) says the same thing as src.
//
// The stability sweep compares canon's output against itself, which cannot
// see a tree that misread its source -- the source is gone after the first
// emission, so a consistently wrong tree is still a fixpoint. Measured:
//
//	WRONG tree: once="print((1 + 2) * 3);"
//	            twice="print((1 + 2) * 3);"   fixpoint
//
// This puts the source back on one side, and forgives exactly two
// differences, spellings the tree cannot record because they do not change
// meaning (perigrin's ruling, 2026-10-23):
//
//   - last-semicolon: a block's last statement acquires its semicolon.
//     Measured: `sub f { 1 }` and `sub g { 1; }` both return 1. That covers
//     a block argument too: `print {$fh} "x"` is emitted `print {$fh;} "x"`.
//     Only where the TREE reads the braces as a Block (CanonToken.BlockSemi):
//     `$h{a;}` and `{a => 1;}` are syntax errors, never a spelling of a hash.
//   - extra-nesting: a source pair nested directly inside the pair canon
//     wrote for a node's Paren flag, around the same node. perl keeps that
//     paren as one boolean bit, OPf_PARENS: measured on perl 5.44.0,
//     `ok((($y eq "x")))` and `ok(($y eq "x"))` both build `seq sKPM`. Only
//     down to one pair: `(($y)) x 3` against `$y x 3` is a difference.
//
// A call's parens are not forgiven. Canon writes them only where the source
// did (the tree's ArgParen) or binding power needs them, so a pair the
// source lacks is a grouping the tree imposed.
//
// Nothing else is forgiven. A paren that changes grouping is the defect this
// exists to catch, and the quotient must not reach it -- see
// TestCanonQuotientIsNarrow, which asserts that each rewrite is accepted and
// that a grouping difference of the same shape is not.
func Faithful(n *Node, src []byte) (bool, string) {
	ok, _, why := faithfulTokens(significantTokens(src), canonTokens(n, src))
	return ok, why
}

// canonTokens is canon(n)'s significant tokens, each `(` and `)` written for
// a node's Paren flag marked ParenNode.
func canonTokens(n *Node, src []byte) []CanonToken {
	var b canonBuf
	emit(&b, n, src, 0)
	text := []byte(b.String())
	var out []CanonToken
	for _, tok := range significant(text) {
		out = append(out, CanonToken{
			Text:      string(text[tok.Start:tok.End]),
			ParenNode: slices.Contains(b.parenNode, tok.Start),
			BlockSemi: slices.Contains(b.blockSemi, tok.Start),
		})
	}
	return out
}

// CanonToken is one significant token of canon's emission.
//
// ParenNode marks a `(` or `)` canon wrote for a node's Paren flag: the one
// pair emit writes for every parenthesised node, `(A,)/0` and `(not $x)`
// among them. Every other paren is unmarked -- a call's argument list, a
// grouping binding power demands, and the three pairs the syntax requires,
// where perl keeps no OPf_PARENS bit: the `->(` argument pair, `$f->($a+1)`,
// the `not (` operand's, `not ($a == 1)`, and a list slice's, `(1,2,3)[1]`
// (Node.SyntaxParen).
//
// BlockSemi marks the `;` canon wrote as the terminator of a Block's last
// statement: the one `;` last-semicolon may forgive. A `;` inside braces the
// tree reads as anything but a Block -- an anonymous hash, a subscript --
// is unmarked.
type CanonToken struct {
	Text      string
	ParenNode bool
	BlockSemi bool
}

// Forgiveness is one difference faithfulTokens forgave: the canon token and
// the source token where it was found, and the rule that forgave it.
type Forgiveness struct {
	CanonIndex int
	SrcIndex   int
	Rule       string
}

// The rules faithfulTokens forgives a difference by. See Faithful;
// TestFaithfulForgivenessIsClosed holds the set closed.
const (
	ruleLastSemicolon = "last-semicolon"
	ruleExtraNesting  = "extra-nesting"
)

// faithfulTokens is Faithful over token sequences: the source's significant
// tokens against canon's. It reports each difference it forgave, with the
// rule that forgave it, on a failing verdict as well: the ones it forgave
// before the difference it could not.
func faithfulTokens(want []string, canon []CanonToken) (bool, []Forgiveness, string) {
	got := make([]string, len(canon))
	for i, c := range canon {
		got[i] = c.Text
	}
	var forgiven []Forgiveness
	// closes holds the source index of each `)` closing a pair forgiven as
	// extra nesting, which the source has and canon does not.
	closes := map[int]bool{}

	w, g := 0, 0
	for w < len(want) && g < len(got) {
		if closes[w] {
			forgiven = append(forgiven, Forgiveness{CanonIndex: g, SrcIndex: w, Rule: ruleExtraNesting})
			w++
			continue
		}
		// A source pair nested directly inside the pair canon wrote for a
		// node's Paren flag: the source's `(` follows the `(` matched to
		// canon's marked one, and closes right before it. Measured on perl
		// 5.44.0, `ok((($y eq "x")))` and `ok(($y eq "x"))` are the same
		// `seq sKPM`. A call's pair is unmarked, so `print((1))+2` is not
		// forgiven against `print(1)+2`: `const sP` against `const s`.
		//
		// Asked before the tokens are compared, so a `(` canon writes next
		// for an operand is not taken for it: `((($a+$b)*2))` against
		// `(($a + $b) * 2)`. Not where canon nests its own pair the same
		// way, which is then the source's pair. canon[g-1] is the `(` matched
		// to want[w-1], or to the pair want[w-1] was itself forgiven inside:
		// nestedPair says want[w-1] is a `(`.
		if g > 0 && canon[g-1].ParenNode && nestedPair(want, w) &&
			!(got[g] == "(" && nestedPair(got, g)) {
			inner, _ := matchParen(want, w)
			forgiven = append(forgiven, Forgiveness{CanonIndex: g, SrcIndex: w, Rule: ruleExtraNesting})
			closes[inner] = true
			w++
			continue
		}
		if want[w] == got[g] {
			w++
			g++
			continue
		}
		// A block's LAST statement may omit its semicolon in the source, and
		// canon always writes one. Measured to be no difference at all:
		//
		//	$ perl -e 'sub f { 1 } sub g { 1; } print f(), g(), "\n";'
		//	11
		//
		// Narrow on purpose: only a `;` the emission has, immediately before
		// its closing `}`, where the source went straight to the brace -- and
		// only one canon marked as a Block's, never a `;` inside braces the
		// tree reads as a hash or a subscript: perl 5.44.0 rejects `$h{a;}`
		// and `{a => 1;}` alike.
		if canon[g].BlockSemi && got[g] == ";" && g+1 < len(got) && got[g+1] == "}" && want[w] == "}" {
			forgiven = append(forgiven, Forgiveness{CanonIndex: g, SrcIndex: w, Rule: ruleLastSemicolon})
			g++
			continue
		}

		return false, forgiven, fmt.Sprintf("token %d: source has %q, canon has %q", w, want[w], got[g])
	}
	if w != len(want) || g != len(got) {
		return false, forgiven, fmt.Sprintf("source has %d significant tokens, canon has %d "+
			"(matched %d and %d)", len(want), len(got), w, g)
	}
	return true, forgiven, ""
}

// significantTokens is the text of every token that carries meaning.
//
// Whitespace, comments and POD are dropped: canon drops trivia on purpose,
// so keeping them would fail every file for a reason that says nothing about
// the tree.
func significantTokens(src []byte) []string {
	var out []string
	for _, tok := range significant(src) {
		out = append(out, string(src[tok.Start:tok.End]))
	}
	return out
}

// droppedKinds are the token kinds significant leaves out. A data section,
// `__END__` or `__DATA__` and everything after it, is not one of them:
// perigrin's ruling (2026-10-25) is that canon writes it back verbatim and
// Faithful compares it exactly, as one token. Its bytes are what DATA
// yields, so a canon that drops or alters them runs differently.
var droppedKinds = []lexer.Kind{lexer.Whitespace, lexer.Comment, lexer.Pod}

// significant is src's tokens with droppedKinds left out.
func significant(src []byte) []lexer.Token {
	var out []lexer.Token
	for _, tok := range lexer.Tokenize(src) {
		if !slices.Contains(droppedKinds, tok.Kind) {
			out = append(out, tok)
		}
	}
	return out
}

// nestedPair reports whether toks[i] opens a pair nested directly inside
// the pair opened at i-1: it closes right before that pair does. That
// toks[i-1] is a `(` follows: from any other token, matchParen finds no `)`
// right after the one closing toks[i].
func nestedPair(toks []string, i int) bool {
	if i == 0 || toks[i] != "(" {
		return false
	}
	inner, okInner := matchParen(toks, i)
	outer, okOuter := matchParen(toks, i-1)
	return okInner && okOuter && inner+1 == outer
}

// matchParen returns the index of the `)` closing the `(` at open.
func matchParen(toks []string, open int) (int, bool) {
	depth := 0
	for i := open; i < len(toks); i++ {
		switch toks[i] {
		case "(":
			depth++
		case ")":
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}
	return 0, false
}
