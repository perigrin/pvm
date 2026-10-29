// ABOUTME: Hexadecimal, octal and binary floating-point literals with a `p`
// ABOUTME: exponent are one Number, signed exponent included, valued as perl does.

package lexer

import (
	"math"
	"strings"
	"testing"
)

// TestRadixFloatLiterals: values measured on 5.44.0 with
// `perl -we 'print LIT'` (the underscored ones also warn "Misplaced _").
// Every literal from 07p1 to 0b_.1p1 is a Guacamole v0.008 case, from
// internal/chart/testdata/guacamole/parses/Statements-Expressions-Value-Literal-023.pl
// to -054.pl.
func TestRadixFloatLiterals(t *testing.T) {
	for _, tc := range []struct {
		lit  string
		want float64
	}{
		{"07p1", 14}, {"07p-2", 1.75},
		{"01.4p1", 3}, {"01.4_p1", 3}, {"01_.4p1", 3}, {"0_.4p1", 1},
		{"0xfP1", 30}, {"0xeP2_", 56}, {"0xfp-1", 7.5},
		{"0x_ap+2", 40}, {"0xap+2_", 40},
		{"0x1.8p1", 3}, {"0x1.8_p1", 3}, {"0x1_.8p1", 3}, {"0x_.8p1", 1},
		{"0b1.1p1", 3}, {"0b1.1_p1", 3}, {"0b1_.1p1", 3}, {"0b_.1p1", 1},
		{"0x1p-2", 0.25}, {"0x.8p0", 0.5}, {"0o7p1", 14}, {"0o1.4p1", 3},
		{"0x1P+1", 2}, {"0x1p1_0", 1024},
	} {
		src := "use feature ':5.36';\n" + tc.lit
		toks := significant(Tokenize([]byte(src)))
		last := toks[len(toks)-1]
		if got := src[last.Start:last.End]; last.Kind != Number || got != tc.lit {
			t.Errorf("%q: want one Number %q, got %v %q as the last of %d tokens",
				tc.lit, tc.lit, last.Kind, got, len(toks))
			continue
		}
		if v, ok := radixFloatValue(tc.lit); !ok || v != tc.want {
			t.Errorf("%q: value %v (ok %v), perl gives %v", tc.lit, v, ok, tc.want)
		}
	}
	// In a hex literal `e` is a digit, so a sign after it still subtracts:
	// perl 5.44.0 prints 29 for `0x1e-1`.
	toks := significant(Tokenize([]byte("print 0x1e-1;")))
	if len(toks) != 5 {
		t.Errorf("`0x1e-1` is 0x1e minus 1, got %d tokens", len(toks))
	}
}

// radixFloatValue reads a radix floating-point literal the way perl does:
// digits of the radix with an optional fraction, then a binary exponent.
// It exists to hold the token's text to perl's measured value.
func radixFloatValue(lit string) (float64, bool) {
	s := strings.ReplaceAll(lit, "_", "")
	base := 8.0
	switch {
	case strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X"):
		base, s = 16, s[2:]
	case strings.HasPrefix(s, "0b") || strings.HasPrefix(s, "0B"):
		base, s = 2, s[2:]
	case strings.HasPrefix(s, "0o") || strings.HasPrefix(s, "0O"):
		s = s[2:]
	}
	p := strings.IndexAny(s, "pP")
	if p < 0 {
		return 0, false
	}
	mant, exp := s[:p], s[p+1:]
	v, scale, frac := 0.0, 1.0, false
	for _, c := range mant {
		if c == '.' {
			frac = true
			continue
		}
		d := strings.IndexRune("0123456789abcdef", c|0x20)
		if d < 0 || float64(d) >= base {
			return 0, false
		}
		v = v*base + float64(d)
		if frac {
			scale *= base
		}
	}
	sign := 1.0
	if exp != "" && (exp[0] == '+' || exp[0] == '-') {
		if exp[0] == '-' {
			sign = -1
		}
		exp = exp[1:]
	}
	e := 0.0
	for _, c := range exp {
		if c < '0' || c > '9' {
			return 0, false
		}
		e = e*10 + float64(c-'0')
	}
	return v / scale * math.Pow(2, sign*e), exp != ""
}
