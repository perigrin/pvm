// ABOUTME: `$#{ EXPR }` hides an expression the parser never built, so it must
// ABOUTME: hedge what it holds; `$#$r` and `$#{$r}` hold nothing and stay quiet.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// TestLastIndexBlockHedges: the lexer emits `$#{ \@a }` as one Term, so the
// reference inside it is invisible to a walk over nodes. Silence there is
// the claim that nothing is inside, which scores WRONG wherever perl found
// something. Measured on 5.42.0 with -MO=Concise:
//
//	my @a; my $n = $#{ \@a };        srefgen, rv2av, av2arylen
//	my $r; my $n = $#{$r->{a}};      rv2hv, rv2av, av2arylen
//	my $n = $#{ {a=>[1]}->{a} };     anonhash, anonlist, rv2hv, ...
//
// op/goto.t:668 and op/gmagic.t:227 are this shape.
func TestLastIndexBlockHedges(t *testing.T) {
	for _, c := range []struct {
		src  string
		kind string
	}{
		{`my @a; my $n = $#{ \@a };`, parseoracle.SiteKindReference},
		{`my $r; my $n = $#{$r->{a}};`, parseoracle.SiteKindHash},
		{`my $n = $#{$$r{a}};`, parseoracle.SiteKindHash},
		{`my $n = $#{ {a=>[1]}->{a} };`, parseoracle.SiteKindAnonhash},
	} {
		if !has(hedgedKinds(c.src), c.kind) && !has(decidedKinds(c.src), c.kind) {
			t.Errorf("%q: perl reports a %s site inside the block; the subject "+
				"neither decided nor hedged one", c.src, c.kind)
		}
	}
}

// TestLastIndexOfPlainDerefDoesNotHedge is the negative half. A `$#` of a
// plain reference holds nothing but the variable, and perl reports none of
// the five markers for it -- measured: rv2av and av2arylen only.
func TestLastIndexOfPlainDerefDoesNotHedge(t *testing.T) {
	for _, src := range []string{
		`my $r; my $n = $#$r;`,
		`my $r; my $n = $#{$r};`,
		`my @a; my $n = $#a;`,
		`my @a; my $n = $#{a};`,
	} {
		if got := hedgedKinds(src); len(got) != 0 {
			t.Errorf("%q: nothing is hidden here, want no hedge, got %v", src, got)
		}
	}
}
