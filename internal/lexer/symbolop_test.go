// ABOUTME: A non-ASCII symbol is one token: an operator a library declares, `⊕`, and its
// ABOUTME: name after `sub` in a `.pmt`, rather than one Error token per byte.

package lexer

import "testing"

// TestNonASCIISymbolIsOneToken: an operator plugin may register a
// non-ASCII symbol, so `2 ⊕ 3` lexes `⊕` as one Operator, under `use utf8`
// or not, and a `.pmt`'s `sub ⊕` names the sub by it, as `sub +` does. A
// non-ASCII letter under `use utf8` is still an identifier's.
func TestNonASCIISymbolIsOneToken(t *testing.T) {
	for _, src := range []string{"2 ⊕ 3", "use utf8; 2 ⊕ 3"} {
		toks := significant(Tokenize([]byte(src)))
		op := toks[len(toks)-2]
		if op.Kind != Operator || src[op.Start:op.End] != "⊕" {
			t.Errorf("%q: got %v %q, want one Operator ⊕", src, op.Kind, src[op.Start:op.End])
		}
	}
	src := "sub ⊕ :infix :equiv(&) (Int $x, Int $y) Int;"
	toks := significant(TokenizeTyped([]byte(src)))
	if name := toks[1]; name.Kind != Word || src[name.Start:name.End] != "⊕" {
		t.Errorf("%q: got %v %q, want the Word ⊕", src, name.Kind, src[name.Start:name.End])
	}
	src = "use utf8; my $é = 1;"
	for _, tok := range significant(Tokenize([]byte(src))) {
		if tok.Kind == Operator && src[tok.Start:tok.End] == "é" {
			t.Errorf("%q: é lexed as an operator", src)
		}
	}
}
