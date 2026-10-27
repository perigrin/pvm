// ABOUTME: FuzzParser's pre-existing canon-fixpoint finds: `no &0`, `A*%x*(%A` and `;A&~ ~0`.
// ABOUTME: Each is read as perl 5.44.0 reads it, or refused where perl -c rejects it.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// clean parses src and fails unless it has no refusal and no error, then
// returns its canon.
func clean(t *testing.T, src string) string {
	t.Helper()
	root := parse.Parse([]byte(src))
	if found := collectUnknown(root); len(found) != 0 || root.HasError {
		t.Errorf("%q: perl compiles this, but it refuses: %v", src, refusalCodes(found))
	}
	return parse.Canon(root, []byte(src))
}

// TestStarPercentAfterWordParses: after an undeclared word, `*%x` is a
// multiplication by `%x`, not the glob `*%` followed by the repetition `x`.
// perl -c on 5.44.0 accepts `A*%x`, `A*@x`, `A*&x` and `A*%$h`, and rejects
// `A*%x*(%A` at EOF; the glob reading made the latter's canon re-read
// differently. A declared A takes the glob, `sub A {} A*%x3` being
// `A(*% x 3)`, which the parser declines rather than mis-read.
func TestStarPercentAfterWordParses(t *testing.T) {
	for src, want := range map[string]string{
		"A*%x":  "A * %x",
		"A*@x":  "A * @x",
		"A*&x":  "A * &x",
		"A*%$h": "A * %$h",
	} {
		if got := clean(t, src); got != want {
			t.Errorf("%q: canon %q, want %q", src, got, want)
		}
	}
	if found := collectUnknown(parse.Parse([]byte("sub A {} A *%x3"))); len(found) == 0 || found[0].Refusal != parse.AmbiguousSigil {
		t.Errorf("`sub A {} A *%%x3`: want an ambiguous_sigil refusal, got %v", refusalCodes(found))
	}
	if ok, why := canonMatches([]byte("A*%x*(%A")); !ok {
		t.Errorf("`A*%%x*(%%A`: canon is not a fixpoint: %s", why)
	}
}

// TestUseWithoutModuleRefuses: `use` and `no` take a module name or a
// version. perl -c on 5.44.0 rejects `no &0`, `no & ! no & 00`, `use
// 'strict'`, `use $x`, `use +strict` and `use;`; read on, `no &0` took `&`
// as an import list whose canon `&;` re-read differently. `use => 1` is the
// string "use", which perl compiles.
func TestUseWithoutModuleRefuses(t *testing.T) {
	for _, src := range []string{"no &0", "no & ! no & 00", "use 'strict';", "use $x;", "use +strict;", "use;", "no"} {
		if found := collectUnknown(parse.Parse([]byte(src))); len(found) == 0 {
			t.Errorf("%q: perl rejects this, but it parses", src)
		}
	}
	for _, src := range []string{"use => 1;", "no => 1;", "no 6;", "use v5.10;", "use 5.010;", "use strict;"} {
		clean(t, src)
	}
}

// TestAmpBeforeTildeAfterWord: `;A&~ ~0` is `'A' & ~~0` to perl 5.44.0, and
// so is its canon `;A & ~~0`. The lexer took the `&` for a sub's sigil and
// the `~~` after it for smartmatch, so the canon refused.
func TestAmpBeforeTildeAfterWord(t *testing.T) {
	for _, src := range []string{";A&~ ~0", ";A & ~~0", "A & ~~0"} {
		clean(t, src)
		if ok, why := canonMatches([]byte(src)); !ok {
			t.Errorf("%q: canon is not a fixpoint: %s", src, why)
		}
	}
}
