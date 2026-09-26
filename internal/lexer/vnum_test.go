// ABOUTME: Two scanNumber rows: the sign of an exponent, and the second dot that makes a v-string.
// ABOUTME: Both produced plausible wrong tokens rather than an error -- a split literal and a number where perl has a string.

package lexer_test

import (
	"strings"
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
// spelling of the same claim and, since `scanVString` stopped requiring a
// second dot, arrives as one Quote too -- so both are checked here because
// the two spellings must agree about the feature, not because they lex
// differently.
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

// TestOneDotVString: a `v` prefix makes a v-string of ONE dot too, which is
// the asymmetry the dot count alone cannot express.
//
// Measured perl 5.42.0 -- the same one dot, two categories, and the `v` is
// the only difference between them:
//
//	$ perl -e 'my $v = v5.36; print length($v)'   2
//	$ perl -e 'print 5.36'                        5.36
//
// Two characters, so perl holds a string of ordinals; `5.36` is the float it
// looks like. We lexed the first as `Word(v5) Operator(.) Number(36)` -- a
// bareword concatenated with a number, an expression the parser accepts
// without complaint, so the token assertions are the only place the
// disagreement shows.
//
// `v5.36` is also the spelling every `use` line in this repository's corpus
// writes, which is why `TestVStringVersionStillEnablesSignatures` covers the
// feature-gating consequence: the version now reaches `noteSignatures` as one
// Quote rather than as a split it reassembles.
func TestOneDotVString(t *testing.T) {
	for _, src := range []string{"v5.36", "v5.10", "v65.66", "v1.0"} {
		got := significant(src)
		want := "Quote(" + src + ")"
		if len(got) != 1 || got[0] != want {
			t.Errorf("%q lexes as %v, want [%s]", src, got, want)
		}
	}

	// CAPITAL V IS NOT A V-STRING, and an earlier revision of this test
	// asserted that it was. Measured on 5.42.0:
	//
	//	$ perl -e 'my $v = V5.36; print "[$v]"'
	//	[V536]
	//	$ perl -Mstrict -e 'my $v = V5.36; print $v'
	//	Bareword "V5" not allowed while "strict subs" in use
	//	$ perl -e 'use V5.36; print "ok"'
	//	Can't locate V5.pm in @INC
	//
	// So `V5.36` is the bareword `V5` concatenated with `.36`, and
	// `use V5.36` LOADS A MODULE. Reading it as a version made
	// `noteSignatures` turn signatures on for a module load -- the same
	// silent change of meaning this whole family of tests exists to catch,
	// introduced by the test that was supposed to prevent it.
	for _, src := range []string{"V5.36", "V5.36.0"} {
		got := significant(src)
		if len(got) == 1 && strings.HasPrefix(got[0], "Quote(") {
			t.Errorf("%q lexes as %v; capital V is a bareword, not a v-string",
				src, got)
		}
	}

	// The negative half, and the whole subtlety: ONE dot with NO `v` stays a
	// number. `5.36` is a float in perl, so a rule that counted dots without
	// looking for the prefix would break arithmetic.
	for _, src := range []string{"5.36", "65.66", "1.0"} {
		got := significant(src)
		want := "Number(" + src + ")"
		if len(got) != 1 || got[0] != want {
			t.Errorf("%q lexes as %v, want [%s]; one dot and no `v` is a "+
				"float", src, got, want)
		}
	}

	// `v5` alone is a BAREWORD, not a v-string: perl needs a dot. Measured:
	// `perl -e 'my $v = v5; print length($v)'` dies with "Bareword found
	// where operator expected"... it is `v5` the identifier, so the lexer
	// must not claim a `v` run with no dot in it.
	if !streamIs("v5", "Word(v5)") {
		t.Errorf("%q lexes as %v, want [Word(v5)]", "v5", significant("v5"))
	}
}

// TestOneDotVStringInUseRoutes covers the three routes a version reaches the
// lexer by, because `use`, `require` and the three-part spelling are
// different paths through `noteSignatures` and `internal/parse/use.go`.
func TestOneDotVStringInUseRoutes(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		{`use v5.36;`, []string{
			"Word(use)", "Quote(v5.36)", "Semicolon(;)",
		}},
		{`require v5.36;`, []string{
			"Word(require)", "Quote(v5.36)", "Semicolon(;)",
		}},
		{`use v5.36.0;`, []string{
			"Word(use)", "Quote(v5.36.0)", "Semicolon(;)",
		}},
	} {
		if !streamIs(c.src, c.want...) {
			t.Errorf("%q lexes as %v, want %v",
				c.src, significant(c.src), c.want)
		}
	}
}
