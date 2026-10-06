// ABOUTME: A decimal literal ends at a letter that is not its exponent: `5x3` is
// ABOUTME: 5 x 3 and `1if $b` is 1 if $b, while radix literals keep their letters.

package lexer

import "testing"

// TestDecimalEndsAtLetter: toke.c's scan_num takes a decimal's `e` only when
// a sign, a digit or `_` follows it (toke.c:13142-13144), and no other
// letter. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e '0-5x-l{0}; $a=5x3; print 1if $b; $c = 1eq 1;'
//	0 - 5 x (-l {0});
//	$a = '555';
//	print 1 if $b;
//	$c = !0;
//
// Octal floats keep their letters: `01.1p0` is 1.125, perl.git
// t/op/hexfp.t:273.
//
// perl.git t/base/lex.t:516.
func TestDecimalEndsAtLetter(t *testing.T) {
	for src, want := range map[string]string{
		"$n=5x-l{0};":    "5",
		"$a=5x3;":        "5",
		"print 1if $b;":  "1",
		"$c = 1eq 1;":    "1",
		"$d = 1.5e10;":   "1.5e10",
		"$d = 5e-1;":     "5e-1",
		"$d = 1e_5;":     "1e_5",
		"$e = 0b101;":    "0b101",
		"$f = 0o17;":     "0o17",
		"$g = 0x1f;":     "0x1f",
		"$h = 1_000;":    "1_000",
		"$i = 0x1.8p1;":  "0x1.8p1",
		"$j = 01.1p0;":   "01.1p0",
		"$k = 00.400p1;": "00.400p1",
	} {
		tok, ok := firstOfKind(src, Number)
		if !ok {
			t.Errorf("%q: no Number", src)
			continue
		}
		if got := src[tok.Start:tok.End]; got != want {
			t.Errorf("%q: first Number %q, want %q", src, got, want)
		}
	}
	// A v-string does not keep a dot with no digit after it: `v10.v257` is
	// `v10 . v257`, measured to deparse as "\n\x{101}". perl.git t/op/lc.t:198.
	src := "$a = v10.v257;"
	for _, tok := range Tokenize([]byte(src)) {
		if text := src[tok.Start:tok.End]; tok.Kind == Quote && text[len(text)-1] == '.' {
			t.Errorf("%q: v-string %q keeps the concatenation's dot", src, text)
		}
	}
}
