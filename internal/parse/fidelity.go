// ABOUTME: Faithful compares canon's emission against the SOURCE, not against itself.
// ABOUTME: One difference is forgiven: the parens a parenless call must acquire.

package parse

import (
	"fmt"

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
	want := significantTokens(src)
	got := significantTokens([]byte(Canon(n, src)))

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
			g++
			continue
		}

		// The one forgiven difference: the emission opened a paren the
		// source did not, immediately after a call's name.
		if got[g] == "(" && g > 0 && isName(got[g-1]) {
			skip, ok := matchParen(got, g)
			if !ok {
				return false, fmt.Sprintf("token %d: unbalanced %q in the emission", g, got[g])
			}
			// The source must contain the same arguments, unparenthesised.
			inner := got[g+1 : skip]
			if w+len(inner) <= len(want) && equalTokens(want[w:w+len(inner)], inner) {
				w += len(inner)
				g = skip + 1
				continue
			}
		}
		return false, fmt.Sprintf("token %d: source has %q, canon has %q", w, want[w], got[g])
	}
	if w != len(want) || g != len(got) {
		return false, fmt.Sprintf("source has %d significant tokens, canon has %d "+
			"(matched %d and %d)", len(want), len(got), w, g)
	}
	return true, ""
}

// significantTokens is the text of every token that carries meaning.
//
// Whitespace, comments and POD are dropped: canon drops trivia on purpose,
// so keeping them would fail every file for a reason that says nothing about
// the tree.
func significantTokens(src []byte) []string {
	var out []string
	for _, tok := range lexer.Tokenize(src) {
		switch tok.Kind {
		case lexer.Whitespace, lexer.Comment, lexer.Pod, lexer.DataSection:
			continue
		}
		out = append(out, string(src[tok.Start:tok.End]))
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

// equalTokens reports whether two token runs are identical.
func equalTokens(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// isName reports whether a token could be a call's name -- the only place
// the forgiven paren may appear.
func isName(tok string) bool {
	if tok == "" {
		return false
	}
	c := tok[0]
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
