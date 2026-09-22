// ABOUTME: A leading-decimal literal parses as a term, and the concatenation spelling beside it still parses.
// ABOUTME: Pins the PARSE-level win at the shape that earned it, so the ratchet number is not the only record.
package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestLeadingDecimalParses is the parse half of issue 01a0c13f. The lexer
// half lives in `internal/lexer/scanrows_test.go`.
//
// Each source here produced at least one Unknown before the fix and none
// after. They are kept as a test rather than left to the ratchet because
// a ratchet records a COUNT: it would go on passing if the count stayed
// at zero for an unrelated reason, and it names no construct.
//
// The third case is the one that moved the corpus most. `code_too_large.t`
// went from 932 Unknowns to 0, and the shape driving it is `('.$v.")` --
// a `(` puts the lexer in term position, so the `.` that follows was read
// as an operator with no left operand and the statement fell to Unknown.
// The file repeats that shape across 10,001 lines, which is why one
// lexical rule moved a four-figure number.
func TestLeadingDecimalParses(t *testing.T) {
	for _, src := range []string{
		`my $x = .5;`,
		`my $x = int($f * 100 + .5);`,
		`print 'ok 1 - (' . $v . ")\n";`,
		`$v = (" " / "1"); print 'ok 1 - (' . $v . ")\n";`,
		`my @a = (1, .5);`,
		`my $x = .5 + .25;`,
	} {
		if n := len(collectUnknownNodes(parse.Parse([]byte(src)))); n != 0 {
			t.Errorf("%q has %d Unknown(s), want 0", src, n)
		}
	}
}

// TestLeadingDecimalConcatenationParses is the negative, at the parse
// level: the spellings that must stay CONCATENATION still parse, and a
// version string still parses.
//
// `print $a .5` is deliberately absent. Measured 5.42.0, perl reads it as
// `print {$a} 0.5` -- a filehandle print -- so it is neither a parse
// failure nor concatenation, and asserting either reading here would
// record something false. It is documented in the lexer test instead.
func TestLeadingDecimalConcatenationParses(t *testing.T) {
	for _, src := range []string{
		`my $r = $a . 5;`,
		`my $r = $a .5;`,
		`my $r = "x" . .5;`,
		`my @r = (1 .. 5);`,
		`use v5.36;`,
		`require(v5.5.630);`,
	} {
		if n := len(collectUnknownNodes(parse.Parse([]byte(src)))); n != 0 {
			t.Errorf("%q has %d Unknown(s), want 0", src, n)
		}
	}
}
