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
			// A block ARGUMENT inside those parens may have acquired its own
			// trailing `;`, which is difference 2 one level down: `print {$fh}
			// "x"` is emitted `print({$fh;} "x")` and `map { $_ } @a` as
			// `map({$_;} @a)`. Forgiving it only in the outer loop failed both,
			// and the `;` is no more meaningful inside a paren than outside one.
			inner := got[g+1 : skip]
			if n, ok := matchForgivingSemicolons(want[w:], inner); ok {
				w += n
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
// Returns how many of want's tokens were consumed.
func matchForgivingSemicolons(want, inner []string) (int, bool) {
	w := 0
	for i := 0; i < len(inner); i++ {
		if w < len(want) && want[w] == inner[i] {
			w++
			continue
		}
		if inner[i] == ";" && i+1 < len(inner) && inner[i+1] == "}" &&
			w < len(want) && want[w] == "}" {
			continue
		}
		return 0, false
	}
	return w, true
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
