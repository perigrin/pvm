// ABOUTME: Hexadecimal, octal and binary floating-point literals parse as one
// ABOUTME: term: perl 5.44.0 compiles `0_.4p1` (1.0) and `0xfp-1` (7.5).

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestNonDecimalFloatLiteralsParse: every source is a Guacamole v0.008 case
// byte for byte, internal/chart/testdata/guacamole/parses/
// Statements-Expressions-Value-Literal-023.pl to -054.pl, which perl 5.44.0
// compiles. Each statement is the literal alone, so its tree is one term.
func TestNonDecimalFloatLiteralsParse(t *testing.T) {
	for _, lit := range []string{
		"07p1", "07p-2", "01.4p1", "01.4_p1", "01_.4p1", "0_.4p1",
		"0xfP1", "0xeP2_", "0xfp-1", "0x_ap+2", "0xap+2_",
		"0x1.8p1", "0x1.8_p1", "0x1_.8p1", "0x_.8p1",
		"0b1.1p1", "0b1.1_p1", "0b1_.1p1", "0b_.1p1",
	} {
		src := "use feature ':5.36';\n" + lit
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) || countUnknown(root) != 0 {
			t.Errorf("%q: perl accepts this; got %s", lit, shape(root))
			continue
		}
		stmts := root.Children
		if got := shape(skipWrappers(stmts[len(stmts)-1])); got != lit {
			t.Errorf("%q: want the literal as one term, got %s", lit, got)
		}
	}
}
