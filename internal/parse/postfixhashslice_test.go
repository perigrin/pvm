// ABOUTME: A postfix hash slice, `->@{...}` or `->%{...}`, reads a hash through a
// ABOUTME: reference, so it reports the rv2hv perl emits; an array slice does not.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// TestPostfixHashSliceReportsRv2hv: `$h->@{1,2}` is `@$h{1,2}` spelled
// postfix, and perl emits rv2hv for both. Measured on 5.42.0 with
// -MO=Concise:
//
//	my $h; my @x = $h->@{1,2};          hslice, rv2hv
//	my $h; my @x = $h->%{1,3};          kvhslice, rv2hv
//	my @x = {1..4}->%{1,3};             anonhash, kvhslice, rv2hv
//	my $h; my @x = $h->{a}->@{1,2};     multideref, hslice, rv2hv
//
// op/postfixderef.t:148,316 are this shape.
func TestPostfixHashSliceReportsRv2hv(t *testing.T) {
	for _, src := range []string{
		`my $h; my @x = $h->@{1,2};`,
		`my $h; my @x = $h->%{1,3};`,
		`my @x = {1..4}->%{1,3};`,
		`my $h; my @x = $h->{a}->@{1,2};`,
		`my $h; my @x = $h->@{f(1)};`,
	} {
		n := 0
		for _, k := range decidedKinds(src) {
			if k == parseoracle.SiteKindHash {
				n++
			}
		}
		if n != 1 {
			t.Errorf("%q: want one decided hash site, perl's rv2hv; got %d", src, n)
		}
	}
}

// TestPostfixArraySliceReportsNoHash is the negative half. A BRACKET slice
// reads an array, rv2av, and `->@*` dereferences one whole: measured, none
// of the three emits rv2hv.
func TestPostfixArraySliceReportsNoHash(t *testing.T) {
	for _, src := range []string{
		`my $a; my @x = $a->@[1,2];`,
		`my $a; my @x = $a->%[1,2];`,
		`my $a; my @x = $a->@*;`,
	} {
		if got := decidedKinds(src); has(got, parseoracle.SiteKindHash) {
			t.Errorf("%q: perl emits no rv2hv here, got %v", src, got)
		}
		if got := hedgedKinds(src); has(got, parseoracle.SiteKindHash) {
			t.Errorf("%q: nothing here is a hash, yet it hedges %v", src, got)
		}
	}
}
