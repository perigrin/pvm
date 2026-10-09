// ABOUTME: A brace before `->`, and `+{}`, are anonymous hashes; perl compiles one for each.
// ABOUTME: The same brace without the arrow, and a block at statement start, compile none.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// TestBraceBeforeArrowIsAHashref holds two answers to "is this `{` a block
// or a hash", each counted as decided anonhash sites against what perl
// compiles. Measured on perl 5.42.0 with -MO=Concise,-exec (an anon sub's
// body is not printed by plain Concise, so `+{}` is measured at top level
// and with -MO=Concise,f in a named sub):
//
//	my @res = grep {a=>$_}->{a}, ("x");   anonhash
//	my $x = +{};                          emptyavhv/ANONHASH
//	sub f { +{} }                         emptyavhv/ANONHASH
//	my @r = grep { $_ } ("x");            none
//	{ 1 }                                 none: a bare block
//
// op/grep.t:105 is the `->` case. The arrow after the `}` is what makes the
// brace a hash. Without it, `grep {a=>$_} ("x")` is a syntax error -- perl
// guesses a hash there too and then finds no comma -- so the negatives are a
// brace perl keeps a block: grep's `{ $_ }` and a bare `{ 1 }`.
func TestBraceBeforeArrowIsAHashref(t *testing.T) {
	for _, c := range []struct {
		src  string
		want int
	}{
		{`my @res = grep {a=>$_}->{a}, ("x");`, 1},
		{`my $x = +{};`, 1},
		{`sub f { +{} }`, 1},
		{`my @r = grep { $_ } ("x");`, 0},
		{`{ 1 }`, 0},
	} {
		n := 0
		for _, k := range decidedKinds(c.src) {
			if k == parseoracle.SiteKindAnonhash {
				n++
			}
		}
		if n != c.want {
			t.Errorf("%q: want %d decided anonhash sites, as perl compiles, got %d",
				c.src, c.want, n)
		}
	}
}
