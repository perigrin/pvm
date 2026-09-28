// ABOUTME: `$::{n}` names the main stash, so the sigil's name is `::` even when
// ABOUTME: a subscript follows rather than an identifier.
package lexer

import "testing"

// TestMainStashSubscript covers the empty package name before a subscript.
//
// `leadingPackageSeparator` already admitted `$::x` -- it requires a word byte
// after the two colons. `$::{n}` has a `{` there, so the guard declined and
// `scanVarName` took `$:` (a real punctuation variable, the output field
// separator) and left the second colon stranded as an Operator:
//
//	$::{n}   ->  Variable($:) Operator(:) Operator({) Word(n) CloseBracket(})
//
// Perl reads it as the main stash, verified on 5.42.0:
//
//	perl -MO=Deparse -e 'our $n = 5; my $r = $::{n};'
//	  ->  my $r = $main::{'n'};
//
// `$main::{n}` and `$Pkg::{n}` were already correct, so only the EMPTY package
// name was affected -- the same shape as `${^TAINT}` being right while `$^O` was
// wrong, and as `$h{m}` being right in an interpolated string while the bare
// form was broken. One spelling of a rule repaired, its sibling left.
func TestMainStashSubscript(t *testing.T) {
	// The sigil's name must be `$::`, with the subscript its own tokens.
	for _, src := range []string{
		"$::{n}",
		"$::{'n'}",
		"%::",
		"@::",
		// A `;` ends the name as surely as a subscript does. A first fix
		// listed the openers `{ [ :` and left `keys %::;` refusing.
		"%::;",
		"$::{n} = 1",
	} {
		toks := significant(Tokenize([]byte(src)))
		if len(toks) == 0 {
			t.Fatalf("Tokenize(%q) produced no significant tokens", src)
		}
		first := string([]byte(src)[toks[0].Start:toks[0].End])
		want := src[:1] + "::"
		if first != want {
			t.Errorf("Tokenize(%q): first token is %q, want %q", src, first, want)
		}
		if toks[0].Kind != Variable {
			t.Errorf("Tokenize(%q): first token is %v, want Variable", src, toks[0].Kind)
		}
	}

	// REGRESSION GUARD, passes today: `$:` alone is a real punctuation
	// variable and must stay one token. perl's `$:` is the set of characters
	// a format may break a line on, so this is not a hypothetical.
	for _, src := range []string{"$:", "$: = 1"} {
		toks := significant(Tokenize([]byte(src)))
		if len(toks) == 0 {
			t.Fatalf("Tokenize(%q) produced no significant tokens", src)
		}
		if got := string([]byte(src)[toks[0].Start:toks[0].End]); got != "$:" {
			t.Errorf("Tokenize(%q): first token is %q, want %q", src, got, "$:")
		}
	}

	// REGRESSION GUARD, passes today: a NAMED package before the subscript is
	// unaffected, and so is the identifier spelling the old guard allowed.
	for _, tc := range []struct{ src, want string }{
		{"$main::{n}", "$main::"},
		{"$Pkg::{n}", "$Pkg::"},
		{"$::x", "$::x"},
		{"%main::", "%main::"},
	} {
		toks := significant(Tokenize([]byte(tc.src)))
		if len(toks) == 0 {
			t.Fatalf("Tokenize(%q) produced no significant tokens", tc.src)
		}
		if got := string([]byte(tc.src)[toks[0].Start:toks[0].End]); got != tc.want {
			t.Errorf("Tokenize(%q): first token is %q, want %q", tc.src, got, tc.want)
		}
	}
}

// significant drops trivia so a test can assert on the tokens that carry
// meaning.
func significant(toks []Token) []Token {
	out := make([]Token, 0, len(toks))
	for _, t := range toks {
		if !IsTrivia(t.Kind) {
			out = append(out, t)
		}
	}
	return out
}
