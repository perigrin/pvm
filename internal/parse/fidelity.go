// ABOUTME: Faithful compares canon's emission against the SOURCE, not against itself.
// ABOUTME: One difference is forgiven: the parens a parenless call must acquire.

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
// differences, each one a spelling the tree cannot record because it does
// not change meaning:
//
//  1. A call written without parens acquires them. `print "x"` and
//     `print("x")` parse to the SAME TREE, correctly, so a tree-faithful
//     emitter must print one form for both. It was 519 of 658 disagreeing
//     T1 files.
//  2. A block's last statement acquires its semicolon. Measured:
//     `sub f { 1 }` and `sub g { 1; }` both return 1.
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
// -- a call's argument list, a grouping binding power demands, and two
// paths that also write a pair from a node's Paren flag, left for the
// Paren-once issue (01a1664d): emitWraps's pair for a Paren node with a
// TrailingComma, `(A,)/0`, and the `->(` argument pair, `$f->($a+1)`.
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
	ruleCallParens    = "call-parens"
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

		// The one forgiven difference: the emission opened a paren the
		// source did not, immediately after a call's name.
		if got[g] == "(" && g > 0 && isName(got[g-1]) {
			skip, ok := matchParen(got, g)
			if !ok {
				return false, forgiven, fmt.Sprintf("token %d: unbalanced %q in the emission", g, got[g])
			}
			// The source must contain the same arguments, unparenthesised.
			// A block ARGUMENT inside those parens may have acquired its own
			// trailing `;`, which is difference 2 one level down: `print {$fh}
			// "x"` is emitted `print({$fh;} "x")` and `map { $_ } @a` as
			// `map({$_;} @a)`. Forgiving it only in the outer loop failed both,
			// and the `;` is no more meaningful inside a paren than outside one.
			if n, inner, ok := matchForgivingSemicolons(want, w, got, g+1, skip); ok {
				forgiven = append(forgiven, Forgiveness{CanonIndex: g, SrcIndex: w, Rule: ruleCallParens})
				forgiven = append(forgiven, inner...)
				w += n
				g = skip + 1
				continue
			}
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

// matchForgivingSemicolons matches inner against the front of want, forgiving a
// `;` the emission wrote immediately before a `}` where the source went straight
// to the brace.
//
// The same rewrite Faithful's main loop forgives, applied inside a forgiven
// paren: `print {$fh} "x"` is emitted as `print({$fh;} "x")` and `map { $_ } @a`
// as `map({$_;} @a)`, so the arguments agree except for that one semicolon. A
// `;` is no more meaningful inside a paren than outside one, and forgiving it
// only outside failed every block argument in the corpus.
//
// It matches got[from:to] against want from w on, and returns how many of
// want's tokens were consumed and each difference it forgave, indexed into
// want and got.
func matchForgivingSemicolons(want []string, w int, got []string, from, to int) (int, []Forgiveness, bool) {
	start := w
	var forgiven []Forgiveness
	for i := from; i < to; i++ {
		if w < len(want) && want[w] == got[i] {
			w++
			continue
		}
		if got[i] == ";" && i+1 < to && got[i+1] == "}" &&
			w < len(want) && want[w] == "}" {
			forgiven = append(forgiven, Forgiveness{CanonIndex: i, SrcIndex: w, Rule: ruleLastSemicolon})
			continue
		}
		// A call's added parens inside these ones: the forgiven difference
		// one level down, forgiven the same way. `print create Foo sub {...}`
		// is emitted `print(create Foo(sub {...}))`.
		if got[i] == "(" && i > from && isName(got[i-1]) {
			if skip, ok := matchParen(got[:to], i); ok {
				if n, inner, ok := matchForgivingSemicolons(want, w, got, i+1, skip); ok {
					forgiven = append(forgiven, Forgiveness{CanonIndex: i, SrcIndex: w, Rule: ruleCallParens})
					forgiven = append(forgiven, inner...)
					w += n
					i = skip
					continue
				}
			}
		}
		return 0, nil, false
	}
	return w - start, forgiven, true
}

// isName reports whether a token could be a call's name -- the only place
// the forgiven paren may appear.
//
// A leading letter or underscore covers every named call, and a file test is
// the one call whose name begins with punctuation: `-e` is ONE operator token
// whose text includes the minus, and canon emits it with the same parens it
// gives `length`. Without this clause `-e $f` emitted as `-e($f)` failed the
// forgiveness that `length $x` emitted as `length($x)` receives, which is a
// difference in the spelling of the name and not in the tree.
func isName(tok string) bool {
	if tok == "" {
		return false
	}
	if lexer.IsFileTest(tok) {
		return true
	}
	c := tok[0]
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
