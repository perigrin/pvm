// ABOUTME: Quote-like operators: arbitrary delimiters, bracket nesting, and closer-before-escape.
// ABOUTME: Every expectation here was run against perl 5.42 before it was written down.

package lexer

import "testing"

// quoteSpan returns the span of the single Quote token in src, or fails.
// Tests assert on the SPAN rather than on extracted content: delimitation is
// the property under test, and a lexer that got the boundaries wrong while
// producing the right inner text would still be broken.
func quoteSpan(t *testing.T, src string) (int, int) {
	t.Helper()
	var found []Token
	for _, tok := range Tokenize([]byte(src)) {
		if tok.Kind == Quote {
			found = append(found, tok)
		}
	}
	if len(found) != 1 {
		t.Fatalf("%q: got %d Quote tokens, want exactly 1 (all tokens: %v)",
			src, len(found), Tokenize([]byte(src)))
	}
	return found[0].Start, found[0].End
}

// TestQuoteArbitraryDelimiter: perl takes the next non-whitespace character
// as the delimiter, with no allow-list. toke.c Perl_scan_str (12272): "after
// skipping whitespace, the next character is the delimiter".
//
// The word-character case is the one that matters, because spec §2.10.3 says
// word characters may NOT delimit and the spec is wrong. Measured:
//
//	$ perl -e 'print q xfoox'
//	foo
//
// A spec correction is owed; the measurement is recorded here so the next
// reader does not re-derive it.
func TestQuoteArbitraryDelimiter(t *testing.T) {
	for _, tc := range []struct{ name, src string }{
		{"slash", `q/foo/`},
		{"tilde", `q~foo~`},
		{"bang", `q!foo!`},
		{"comma", `q,foo,`},
		{"hash glued", `q#foo#`},
		{"word after space", `q xfoox`},
		{"digit after space", `q 9foo9`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start, end := quoteSpan(t, tc.src)
			if start != 0 || end != len(tc.src) {
				t.Errorf("%q: span [%d,%d), want [0,%d) — the whole operator",
					tc.src, start, end, len(tc.src))
			}
		})
	}
}

// TestQuoteNesting: bracketing delimiters count depth, everything else stops
// at the first closer. toke.c:12412 — `if (PL_multi_open == PL_multi_close)`
// turns nesting off.
//
//	$ perl -e 'print q{a{b}c}'
//	a{b}c
func TestQuoteNesting(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		wantEnd   int
	}{
		{"brace nests", `q{a{b}c}`, 8},
		{"paren nests", `q(a(b)c)`, 8},
		{"bracket nests", `q[a[b]c]`, 8},
		{"angle nests", `q<a<b>c>`, 8},
		// A non-bracketing delimiter does not nest: the first closer wins,
		// so the operator ends early and the rest is other tokens.
		{"slash does not nest", `q/a/b/`, 4},
		{"tilde does not nest", `q~a~b~`, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start, end := quoteSpan(t, tc.src)
			if start != 0 || end != tc.wantEnd {
				t.Errorf("%q: span [%d,%d), want [0,%d)", tc.src, start, end, tc.wantEnd)
			}
		})
	}
}

// TestQuoteThreePart: s, tr and y take a second delimiter pair only when the
// first pair is bracketing. `s{a}{b}` has two pairs; `s/a/b/` reuses one
// delimiter for three positions.
func TestQuoteThreePart(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		wantEnd   int
	}{
		{"s slash", `s/a/b/`, 6},
		{"s brace", `s{a}{b}`, 7},
		{"tr slash", `tr/a/b/`, 7},
		{"y slash", `y/a/b/`, 6},
		{"tr brace", `tr{a}{b}`, 8},
		// Modifiers belong to the operator's span.
		{"s with modifiers", `s/a/b/gi`, 8},
		{"s brace with modifiers", `s{a}{b}ge`, 9},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start, end := quoteSpan(t, tc.src)
			if start != 0 || end != tc.wantEnd {
				t.Errorf("%q: span [%d,%d), want [0,%d)", tc.src, start, end, tc.wantEnd)
			}
		})
	}
}

// TestQuoteBackslashDelimiter is the case three implementations got wrong.
//
// When the delimiter IS a backslash, escapes are impossible: every `\` is a
// delimiter. perl guards it at toke.c:12445 with `close_delim_code != '\\'`.
// Measured:
//
//	$ perl -e '$_="a"; s\a\b\; print'
//	b
//
// Check the closer BEFORE the escape and this works. Check the escape first
// and the scan eats its own closer, runs to EOF, and derails every line after
// it. PerlOnJava checks the closer first and is right; gotreesitter and
// perl-lsp check the escape first and are both wrong — and gotreesitter is
// the grammar this project currently depends on, which is how the bug was
// found.
func TestQuoteBackslashDelimiter(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		wantEnd   int
	}{
		{"q backslash", `q\a\`, 4},
		{"s backslash three parts", `s\a\b\`, 6},
		{"s backslash with modifier", `s\a\b\r`, 7},
		{"tr backslash", `tr\a\b\`, 7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start, end := quoteSpan(t, tc.src)
			if start != 0 || end != tc.wantEnd {
				t.Errorf("%q: span [%d,%d), want [0,%d) — the closer must be "+
					"checked before the escape", tc.src, start, end, tc.wantEnd)
			}
		})
	}

	// The discriminating case: a backslash-delimited body containing what
	// would otherwise be an escape. `q\a\nb\` is NOT "a", newline, "b" — the
	// second backslash closes the string.
	start, end := quoteSpan(t, `q\a\nb\`)
	if start != 0 || end != 4 {
		t.Errorf(`q\a\nb\: span [%d,%d), want [0,4) — the second backslash closes it`, start, end)
	}
}

// TestQuoteCommentAsymmetry pins two rules that look like edge cases and are
// in the corpus.
//
// Spec §2.10.3: `q#a#` is a string delimited by `#`, but `q #a#` is a
// COMMENT. Measured, and the real rule is more interesting than "the rest of
// the line is a comment": perl's skipspace runs BEFORE delimiter selection,
// so the comment is skipped and the delimiter is the next non-whitespace
// character — which may be on a later line.
//
//	$ perl -e 'print q #comment
//	xfoox'
//	foo
//
// The delimiter there is `x`, two lines down. Spec §2.10.5: a comment may
// also sit between the two parts of a bracketing three-part operator.
//
//	$ perl -e '$_="aXc"; s{X} # comment
//	{Y}; print'
//	aYc
func TestQuoteCommentAsymmetry(t *testing.T) {
	// Glued: # is the delimiter.
	start, end := quoteSpan(t, `q#a#`)
	if start != 0 || end != 4 {
		t.Errorf(`q#a#: span [%d,%d), want [0,4) — glued # delimits`, start, end)
	}

	// Spaced: # opens a comment, and the delimiter is the next non-whitespace
	// character after it, here on the following line.
	const spaced = "q #comment\nxfoox"
	start, end = quoteSpan(t, spaced)
	if start != 0 || end != len(spaced) {
		t.Errorf("%q: span [%d,%d), want [0,%d) — the comment is skipped and "+
			"x is the delimiter", spaced, start, end, len(spaced))
	}

	// A comment between the two parts of a bracketing three-part operator.
	const between = "s{X} # comment\n{Y}"
	start, end = quoteSpan(t, between)
	if start != 0 || end != len(between) {
		t.Errorf("%q: span [%d,%d), want [0,%d) — the second pair follows the comment",
			between, start, end, len(between))
	}
}
