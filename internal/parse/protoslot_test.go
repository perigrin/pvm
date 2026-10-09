// ABOUTME: A call that fills a \, * or + prototype slot reports the reference perl takes for it.
// ABOUTME: Each count is perl 5.42's srefgen ops for the same source, measured with -MO=Concise.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// TestPrototypeSlotReportsReference: perl compiles an argument in a `\X`
// slot as a reference to it, a glob in a `*` slot as a reference to the
// glob, and an array or hash in a `+` slot as a reference to it. Each is a
// srefgen with no `\` in the source. Measured on 5.42.0:
//
//	$ perl -MO=Concise -e 'sub f(\@){1} my @a; f(@a);'
//	-> srefgen
//
// The prototype applies only where perl sees it: not to `&f(...)`, not to
// a method, and not to a call above the declaration.
func TestPrototypeSlotReportsReference(t *testing.T) {
	for _, c := range []struct {
		src  string
		want int
	}{
		{`sub f(\@){1} my @a; f(@a);`, 1},
		{`sub f(\@){1} my @a; f @a;`, 1},
		{`sub f(\%){1} my %h; f(%h);`, 1},
		{`sub f(\$){1} my $s; f($s);`, 1},
		{`sub f(\&){1} sub g{} f(&g);`, 1},
		{`sub f(\[$@%]){1} my @a; f(@a);`, 1},
		{`sub f(\[$@%]){1} my %h; f(%h);`, 1},
		{`sub f(\@\@){1} my (@a,@b); f(@a,@b);`, 2},
		{`sub f($\@){1} my @b; f(1,@b);`, 1},
		{`sub f(\@@){1} my @a; f(@a, 1, 2);`, 1},
		{`sub f(\@){1} my $r; f(@$r);`, 1},

		// A `*` slot takes a reference to a GLOB, and to nothing else.
		{`sub f(*){1} f(*FOO);`, 1},
		{`sub f(*&){1} f *FOO, sub {1};`, 1},
		{`sub f(*){1} my $x; f(*$x);`, 1},
		{`sub f(*){1} f(FOO);`, 0},
		{`sub f(*){1} my $fh; f($fh);`, 0},
		{`sub f(*){1} f(\*FOO);`, 1}, // the `\` in the source, once

		// A `+` slot takes a reference to an array or a hash.
		{`sub f(+){1} my @a; f(@a);`, 1},
		{`sub f(+){1} my %h; f(%h);`, 1},
		{`sub f(+){1} my $x; f($x);`, 0},

		// A lexical sub's prototype counts as a package sub's does.
		{`my sub f(\@){1} my @a; f(@a);`, 1},
		{`use feature "state"; state sub f(\@){1} my @a; f(@a);`, 1},

		// Where perl applies no prototype, it takes no reference.
		{`sub f{1} my @a; f(@a);`, 0},
		{`sub f(;\@){1} f();`, 0},
		{`sub f(\@){1} my @a; &f(@a);`, 0},
		{`sub f(\@){1} my @a; main->f(@a);`, 0},
		{`my @a; f(@a); sub f(\@){1}`, 0},
	} {
		got := 0
		for _, k := range decidedKinds(c.src) {
			if k == parseoracle.SiteKindReference {
				got++
			}
		}
		if got != c.want {
			t.Errorf("%q: %d decided reference sites, want %d (perl's srefgen count)", c.src, got, c.want)
		}
	}
}
