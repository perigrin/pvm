// ABOUTME: Faithful compares canon's emission against the SOURCE, not against itself.
// ABOUTME: One difference is forgiven: the semicolon canon writes after a block's last statement.

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
// This puts the source back on one side, and forgives exactly one
// difference, a spelling the tree cannot record because it does not change
// meaning: a block's last statement acquires its semicolon. Measured:
// `sub f { 1 }` and `sub g { 1; }` both return 1. That covers a block
// argument too: `print {$fh} "x"` is emitted `print {$fh;} "x"`.
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
		})
	}
	return out
}

// CanonToken is one significant token of canon's emission.
//
// ParenNode marks a `(` or `)` canon wrote on one of two paths: a
// parenthesised List's own pair, and the outermost pair of a parenthesised
// Term, which is written as its source text. Every other paren is unmarked
// -- a call's argument list, a grouping binding power demands, and three
// paths that also write a pair from a node's Paren flag, left for the
// Paren-once issue (01a1664d): emitWraps's pair for a Paren node with a
// TrailingComma, `(A,)/0`, the `->(` argument pair, `$f->($a+1)`, and
// emitNot's pair for a parenthesised `not` or its parenthesised operand,
// `(not $x)` and `not ($a == 1)`.
type CanonToken struct {
	Text      string
	ParenNode bool
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

	w, g := 0, 0
	for w < len(want) && g < len(got) {
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
		// its closing `}`, where the source went straight to the brace.
		if got[g] == ";" && g+1 < len(got) && got[g+1] == "}" && want[w] == "}" {
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

// droppedKinds are the token kinds significant leaves out.
var droppedKinds = []lexer.Kind{lexer.Whitespace, lexer.Comment, lexer.Pod, lexer.DataSection}

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
