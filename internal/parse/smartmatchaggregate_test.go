// ABOUTME: Smartmatch takes an implicit reference to an array or hash operand, so
// ABOUTME: `1 ~~ @l` and `when (@l)` report the srefgen perl compiles them to.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// TestSmartmatchAggregateReportsReference: perl compiles each array or hash
// operand of `~~`, and the condition of `when`, through ref_array_or_hash,
// so `1 ~~ @l` is `1 ~~ \@l` (Deparse agrees). Measured on 5.42.0 with
// -MO=Concise, each one srefgen per aggregate operand:
//
//	my @l; 1 ~~ @l;  @l ~~ 1;           1 srefgen
//	my %h; 1 ~~ %h;  %h ~~ 1;           1 srefgen
//	my @l; my @m; @l ~~ @m;             2 srefgen
//	1 ~~ @_;  1 ~~ @$r;  1 ~~ @{$r};    1 srefgen
//	1 ~~ %$r;  1 ~~ $r->@*;  $r->%*     1 srefgen
//	when (@l) { ... }; when (%h) { ... } 1 srefgen
//	print 1 when @l;                    1 srefgen
//
// op/switch.t:1101,1120 are `when (@list)` and `when (%list)`.
func TestSmartmatchAggregateReportsReference(t *testing.T) {
	for _, c := range []struct {
		src  string
		want int
	}{
		{`my @l; my $x = 1 ~~ @l;`, 1},
		{`my @l; my $x = @l ~~ 1;`, 1},
		{`my %h; my $x = 1 ~~ %h;`, 1},
		{`my %h; my $x = %h ~~ 1;`, 1},
		{`my @l; my @m; my $x = @l ~~ @m;`, 2},
		{`my $x = 1 ~~ @_;`, 1},
		{`my $r; my $x = 1 ~~ @$r;`, 1},
		{`my $r; my $x = 1 ~~ @{$r};`, 1},
		{`my $r; my $x = 1 ~~ %$r;`, 1},
		{`my $r; my $x = 1 ~~ $r->@*;`, 1},
		{`my $r; my $x = 1 ~~ $r->%*;`, 1},
		{`my @l; for (1) { when (@l) { 1 } }`, 1},
		{`my %h; for (1) { when (%h) { 1 } }`, 1},
		{`my @l; for (1) { print 1 when @l; }`, 1},
	} {
		if n := decidedReferences(c.src); n != c.want {
			t.Errorf("%q: want %d decided reference site(s), perl's srefgen; got %d",
				c.src, c.want, n)
		}
	}
}

// TestSmartmatchScalarTakesNoReference is the negative half. A scalar, a
// slice, an anonymous array and a boolean-looking `when` condition reach
// smartmatch without ref_array_or_hash taking anything. Measured: no
// srefgen in any of these.
func TestSmartmatchScalarTakesNoReference(t *testing.T) {
	for _, src := range []string{
		`my $x; my $y = 1 ~~ $x;`,
		`my @l; my $x = 1 ~~ @l[0,1];`,
		`my @l; for (1) { when ([1]) { 1 } }`,
		`my @l; for (1) { when (!@l) { 1 } }`,
		`my @l; for (1) { when (@l > 1) { 1 } }`,
		`my @l; my $n = @l + 1;`,
	} {
		if n := decidedReferences(src); n != 0 {
			t.Errorf("%q: perl takes no reference here, got %d reference site(s)", src, n)
		}
	}
}

func decidedReferences(src string) int {
	n := 0
	for _, k := range decidedKinds(src) {
		if k == parseoracle.SiteKindReference {
			n++
		}
	}
	return n
}
