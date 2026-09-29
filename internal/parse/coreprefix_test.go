// ABOUTME: `CORE::X` is the keyword X, so `CORE::state`, `CORE::say` and `CORE::length`
// ABOUTME: parse as their keywords while canon keeps the prefix the source wrote.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestCorePrefixedKeyword holds perl's reading of the `CORE::` prefix, which
// names the builtin whatever else is in scope. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'use feature "state","say"; CORE::state $x = 42;
//	      CORE::say "ok"; my $n = CORE::length $s; CORE::my $y;'
//	state $x = 42;
//	say 'ok';
//	my $n = length $s;
//	my $y;
//
// The parser matched the whole word `CORE::state` against its keyword tables
// and found nothing, so each refused -- or, for `CORE::length $s`, became a call
// with no argument followed by a stray `$s`, at Unknown=0.
//
// The prefix stays in the tree and in canon: `CORE::say` works without the
// `say` feature where bare `say` does not, so dropping it changes the program.
func TestCorePrefixedKeyword(t *testing.T) {
	cases := []struct {
		src   string
		canon string
	}{
		{`CORE::state $x = 42;`, `CORE::state $x = 42;`},
		{`CORE::my $y;`, `CORE::my $y;`},
		{`CORE::say "ok";`, `CORE::say("ok");`},
		{`my $n = CORE::length $s;`, `my $n = CORE::length($s);`},
		{`CORE::require bleah;`, `CORE::require bleah;`},
		// A loop's declarator is looked up in its own table, which needs
		// the prefix off as well -- op/for.t line 727.
		{`for CORE::my $v (@o) { 1 }`, `for CORE::my $v (@o) {1;}`},
	}
	for _, c := range cases {
		src := []byte(c.src)
		n := parse.Parse(src)
		if got := countUnknown(n); got != 0 {
			t.Errorf("Parse(%q) has %d Unknown, want 0", c.src, got)
		}
		got := strings.TrimSpace(parse.Canon(n, src))
		if got != c.canon {
			t.Errorf("Canon(%q):\n  got  %s\n  want %s", c.src, got, c.canon)
		}
		again := strings.TrimSpace(parse.Canon(parse.Parse([]byte(got)), []byte(got)))
		if again != got {
			t.Errorf("canon is not a fixpoint:\n  once  %s\n  twice %s", got, again)
		}
	}
}
