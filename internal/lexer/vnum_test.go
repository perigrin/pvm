// ABOUTME: Two scanNumber rows: the sign of an exponent, and the second dot that makes a v-string.
// ABOUTME: Both produced plausible wrong tokens rather than an error -- a split literal and a number where perl has a string.

package lexer_test

import (
	"testing"

	"tamarou.com/pvm/internal/lexer"
)

// TestSignedExponentIsOneToken: the sign of an EXPONENT is part of the
// literal, where a sign in front of a number never is.
//
// `scanNumber`'s loop took `isWordByte(c) || c == '.'`, and `-` is neither,
// so the literal ended at the `e`. Measured on perl 5.42.0, the split is
// wrong rather than merely different:
//
//	$ perl -e 'print 5e-1'
//	0.5
//	$ perl -e 'print 5e'
//	Bareword found where operator expected (Missing operator before "e"?)
//
// `5e` is not a number, so `Number("5e")` is a token perl would reject.
func TestSignedExponentIsOneToken(t *testing.T) {
	for _, c := range []struct {
		src  string
		want string
	}{
		{"5e-1", "Number(5e-1)"},
		{"5e+1", "Number(5e+1)"},
		{"5E-1", "Number(5E-1)"},
		{"1.5e-3", "Number(1.5e-3)"},
		{".5e-3", "Number(.5e-3)"},
		{"1_000e-2", "Number(1_000e-2)"},
	} {
		got := significant(c.src)
		if len(got) != 1 || got[0] != c.want {
			t.Errorf("%q lexes as %v, want [%s]", c.src, got, c.want)
		}
	}
}

// TestSignAfterNumberIsStillAnOperator is the negative half, and the reason
// the rule is ADJACENCY TO `e` rather than "a sign inside a number".
//
// Measured on perl 5.42.0, each of these is arithmetic and not a literal:
//
//	$ perl -e 'print 5-1'      4
//	$ perl -e 'print 0x1e-1'   29
//
// The second is the one a naive rule breaks: `e` is a hex DIGIT, so the
// `-` after `0x1e` is subtraction. A radix prefix rules the exponent out.
func TestSignAfterNumberIsStillAnOperator(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		{"5-1", []string{"Number(5)", "Operator(-)", "Number(1)"}},
		{"5+1", []string{"Number(5)", "Operator(+)", "Number(1)"}},
		{"0x1e-1", []string{"Number(0x1e)", "Operator(-)", "Number(1)"}},
		{"0X1E-1", []string{"Number(0X1E)", "Operator(-)", "Number(1)"}},
		{"1.5-3", []string{"Number(1.5)", "Operator(-)", "Number(3)"}},
		// A second sign is not part of the literal either: `5e-1-1` is
		// the number then subtraction.
		{"5e-1-1", []string{"Number(5e-1)", "Operator(-)", "Number(1)"}},
	} {
		if !streamIs(c.src, c.want...) {
			t.Errorf("%q lexes as %v, want %v",
				c.src, significant(c.src), c.want)
		}
	}
}

// TestVStringIsOneStringToken: TWO DOTS make a string, not a number.
//
// Measured on perl 5.42.0, both spellings give the same three-character
// string and neither is the number it looks like:
//
//	$ perl -e 'my $v = 65.66.67; print $v'      ABC
//	$ perl -e 'my $v = v65.66.67; print $v'     ABC
//	$ perl -e 'my $x = 5.42.0; print $x+0'      0
//
// Bare `65.66.67` lexed as one `Number` -- the right boundary in the wrong
// CATEGORY -- and `v65.66.67` as `Word(v65) Operator(.) Number(66.67)`,
// because `v65` is a legal identifier the word scanner claims first.
//
// The kind is Quote, which is what this lexer calls a string: perl holds
// a string of ordinals here, and `internal/conformance/categories.go`
// reads a Quote with no operator name as the glossary's "string literal".
func TestVStringIsOneStringToken(t *testing.T) {
	for _, src := range []string{
		"65.66.67", "v65.66.67", "5.42.0", "1.2.3.4", "v1.2.3",
	} {
		got := significant(src)
		want := "Quote(" + src + ")"
		if len(got) != 1 || got[0] != want {
			t.Errorf("%q lexes as %v, want [%s]", src, got, want)
		}
	}
}

// TestVStringIsNotANumericLiteral states the category claim directly, since
// the boundary being right is the half that was already true.
func TestVStringIsNotANumericLiteral(t *testing.T) {
	for _, src := range []string{"65.66.67", "v65.66.67"} {
		for _, tok := range lexer.Tokenize([]byte(src)) {
			if tok.Kind == lexer.Number {
				t.Errorf("%q emits Number(%s); perl makes it a string",
					src, src[tok.Start:tok.End])
			}
		}
	}
}

// TestVStringVersionStillEnablesSignatures is the consequence the v-string
// rule nearly broke silently, and the reason `use v5.42.0` is in this file
// at all.
//
// `use v5.42.0;` has TWO dots, so it now arrives as ONE token where the
// signatures machinery expected Word("v5") Operator(".") Number("42.0").
// With nothing reading the version out of a v-string, the feature bundle
// stops turning on -- and then `sub g ($a, $b)` is a PROTOTYPE rather than
// a signature, which changes what the rest of the file MEANS. Measured
// 5.42.0:
//
//	$ perl -e 'use v5.42.0; sub g ($a, $b) {1} print prototype(\&g) // "undef"'
//	undef
//
// No prototype, so perl read a signature. `use v5.36` is the one-dot
// spelling of the same claim and still lexes split, so both are checked
// here: the two spellings must agree about the feature.
func TestVStringVersionStillEnablesSignatures(t *testing.T) {
	for _, src := range []string{
		"use v5.42.0;\nsub g ($a, $b) { 1 }\n",
		"use v5.36;\nsub g ($a, $b) { 1 }\n",
		"use v5.36.0;\nsub g ($a, $b) { 1 }\n",
	} {
		for _, tok := range lexer.Tokenize([]byte(src)) {
			if tok.Kind == lexer.Prototype {
				t.Errorf("%q lexes %q as a Prototype; the version enables "+
					"signatures, so it is a signature",
					src, src[tok.Start:tok.End])
			}
		}
	}

	// The negative half: a version BELOW the bundle leaves signatures off,
	// so the same parens stay a prototype. `v5.5.630` is the three-part
	// spelling `require` uses.
	for _, src := range []string{
		"use v5.10.0;\nsub g ($) { 1 }\n",
		"use v5.5.630;\nsub g ($) { 1 }\n",
	} {
		found := false
		for _, tok := range lexer.Tokenize([]byte(src)) {
			if tok.Kind == lexer.Prototype {
				found = true
			}
		}
		if !found {
			t.Errorf("%q lexes no Prototype; that version does not enable "+
				"signatures, so `($)` is one", src)
		}
	}
}

// TestOneDotIsStillANumber is the negative half of the dot count. A single
// dot is a float and stays one:
//
//	$ perl -e 'my $n = 65.66; print $n'    65.66
//	$ perl -e 'my $x = 5.36; print $x+0'   5.36
func TestOneDotIsStillANumber(t *testing.T) {
	for _, src := range []string{"65.66", "1.5", ".5", "1.", "1.5e-3"} {
		got := significant(src)
		want := "Number(" + src + ")"
		if len(got) != 1 || got[0] != want {
			t.Errorf("%q lexes as %v, want [%s]", src, got, want)
		}
	}
}
