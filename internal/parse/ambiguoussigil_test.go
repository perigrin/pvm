// ABOUTME: A `&` or `*` read as an operator after an undeclared bareword, before a quote-op name, refuses.
// ABOUTME: The lexer shielded the name as a sub's or glob's; perl reads it as the operator it begins.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestAmbiguousSigilBeforeQuoteOpRefuses is the parser fuzzer's `A &y` and
// `A*y`. After a bareword perl has not seen declared, it expects an
// operator: `&` is bitwise-and, `*` multiplies, and `y` begins a
// transliteration. Measured on 5.42.0 with -MO=Deparse: `A &y;`, `A*y`,
// `A *y;` and `A &s(1);` are each refused ("Transliteration ... not
// terminated" or a syntax error at EOF), while `sub A {} A &y;` and
// `sub A {} A*y;` call A with `&y` and `*y`.
//
// The lexer has no symbol table, so it read the `&` and `*` as sigils and
// the name after them as a word; the parser, which knows A is undeclared,
// reads the operator. Neither guess is perl's token stream, so the parser
// reports the disagreement rather than building `A() & y()`, whose canon
// re-lexed as a transliteration.
func TestAmbiguousSigilBeforeQuoteOpRefuses(t *testing.T) {
	for _, src := range []string{"A &y", "A &y;", "A*y", "A *y;", "A &s(1);"} {
		found := collectUnknown(parse.Parse([]byte(src)))
		if len(found) == 0 || found[0].Refusal != parse.AmbiguousSigil {
			t.Errorf("%q: want an ambiguous_sigil refusal, got %v", src, refusalCodes(found))
		}
	}
	// What perl compiles still parses: a declared A takes `&y` and `*y` as
	// its argument, an operand that is no quote-op name is the operator's,
	// and `s =>` is the string "s" whatever is before it.
	for _, src := range []string{
		"sub A {} A &y;", "sub A {} A*y;", "A & $x;", "A * 2;", "my $x; $x*s => 1;",
	} {
		if found := collectUnknown(parse.Parse([]byte(src))); len(found) != 0 {
			t.Errorf("%q: perl compiles this, but it refuses: %v", src, refusalCodes(found))
		}
	}
}
