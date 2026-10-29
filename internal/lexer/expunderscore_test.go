// ABOUTME: Underscores may stand on either side of a decimal exponent's sign:
// ABOUTME: `1e_+_1` is the one number 10, and `1e_+_A` ends the number before e.

package lexer

import (
	"strconv"
	"strings"
	"testing"
)

// TestExponentUnderscores: measured on 5.44.0 with
// `perl -MO=Deparse -e 'my $x = LIT;'`. The first two are Guacamole v0.008's
// internal/chart/testdata/guacamole/parses/Statements-Expressions-Value-Literal-020.pl
// and -021.pl.
func TestExponentUnderscores(t *testing.T) {
	for _, tc := range []struct {
		lit  string
		want float64
	}{
		{"1e_+_1", 10},
		{"1e_+__1", 10},
		{"1e+_1", 10},
		{"1e__+1", 10},
		{"1e_-_1", 0.1},
		{"1e_-1", 0.1},
		{"1e_1", 10},
		{"1e__1", 10},
		{"1E_+_1", 10},
		{"1.5e_+_2", 150},
		{"1e_+_1_0", 1e10},
		{"5e-1", 0.5},
	} {
		src := "use feature ':5.36';\n" + tc.lit
		toks := significant(Tokenize([]byte(src)))
		last := toks[len(toks)-1]
		if got := src[last.Start:last.End]; last.Kind != Number || got != tc.lit {
			t.Errorf("%q: want one Number %q, got %v %q (%d tokens)", tc.lit, tc.lit, last.Kind, got, len(toks))
			continue
		}
		v, err := strconv.ParseFloat(strings.ReplaceAll(tc.lit, "_", ""), 64)
		if err != nil || v != tc.want {
			t.Errorf("%q: value %v (%v), perl gives %v", tc.lit, v, err, tc.want)
		}
	}
	// An `e` with no digit after its sign starts no exponent, so the number
	// ends before it. perl 5.44.0: "Bareword found where operator expected
	// (Missing operator before "e_"?)" for each.
	for _, lit := range []string{"1e_", "1e_+", "1e_+_", "1e_+_A", "1e+_", "0e+A"} {
		toks := significant(Tokenize([]byte("my $x = " + lit + ";")))
		for _, tok := range toks {
			if tok.Kind == Number && tok.End-tok.Start != 1 {
				t.Errorf("%q: the number must end before the e, got Number of %d bytes", lit, tok.End-tok.Start)
			}
		}
	}
}
