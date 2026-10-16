// ABOUTME: A decimal `use VERSION` pads its fraction to three digits, as perl does, before the signatures test.
// ABOUTME: `use 5.04` is 5.040 and turns signatures on; read as 5.4 the signature lexed as a prototype.

package lexer

import "testing"

// TestLexerVersionPadding: perl reads a decimal version's fraction in groups
// of three digits, padding a short one with zeros. Measured on 5.42.0 with
// -MO=Deparse, `use 5.04; sub f ($x) { $x }` enables `:5.40` and keeps the
// signature, as `use 5.036` enables `:5.36`; `use 5.034` has no signatures,
// and its `($x)` is a prototype. The parser already padded (perlVersion), so
// the lexer's prototype and the parser's bundle disagreed and the signature
// was lost.
func TestLexerVersionPadding(t *testing.T) {
	for src, proto := range map[string]bool{
		"use 5.04; sub f ($x) { $x }":  false,
		"use 5.036; sub f ($x) { $x }": false,
		"use 5.034; sub f ($x) { $x }": true,
	} {
		got := false
		for _, tok := range Tokenize([]byte(src)) {
			got = got || tok.Kind == Prototype
		}
		if got != proto {
			t.Errorf("%q: lexed a prototype %v, want %v", src, got, proto)
		}
	}
}
