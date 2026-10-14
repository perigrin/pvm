// ABOUTME: A declaration with a parenthesised list ends at its closing paren, so a
// ABOUTME: comma after it separates the enclosing list rather than joining the declaration.
package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestParenDeclarationEndsAtItsParen: `my`, `our`, `state` and `local` take
// only their parenthesised list; a comma after the `)` belongs to the list
// the declaration sits in. Read at the lowest power, the list swallowed it:
// `read $fh, my ($buf), 3` was `my (($buf), 3)` and read lost its count.
// Measured on 5.42.0 with -MO=Deparse:
//
//	read $fh, my ($buf), 3;    read $fh, my $buf, 3;
//	f($a, my ($b), $c);        f($a, my $b, $c);
//	my ($b), $c;               my $b, $c;
//	my ($b, $c), $d;           my($b, $c), $d;
//	our (@x), 1;               our @x, '???';
//	local ($y), 2;             local $y, '???';
//	state ($s), 1;             state $s, '???';
//	opendir my ($dh), ".";     opendir my $dh, '.';
//	sysread STDIN, my ($x), 1; sysread STDIN, my $x, 1;
//	my ($a), my $b = (1, 2);   my $a, my $b = (...);
//	my $b, $c;                 my $b, $c;
//	my ($x) = @a or die;       die unless my($x) = @a;
//
// An assignment still binds inside: `my ($a, $b) = @_, 1` keeps `= @_` in
// the declaration and leaves `, 1` to the list above it.
func TestParenDeclarationEndsAtItsParen(t *testing.T) {
	cases := []struct {
		src  string
		decl string // shape of the first declaration
	}{
		{`read $fh, my ($buf), 3;`, `(my $buf)`},
		{`f($a, my ($b), $c);`, `(my $b)`},
		{`my ($b), $c;`, `(my $b)`},
		{`my ($b, $c), $d;`, `(my ( $b $c))`},
		{`our (@x), 1;`, `(our @x)`},
		{`local ($y), 2;`, `(local $y)`},
		{"use feature 'state'; state ($s), 1;", `(state $s)`},
		{`opendir my ($dh), ".";`, `(my $dh)`},
		{`sysread STDIN, my ($x), 1;`, `(my $x)`},
		{`my ($a), my $b = (1, 2);`, `(my $a)`},
		{`my $b, $c;`, `(my $b)`},
		{`my ($a, $b) = @_, 1;`, `(my (= ( $a $b) @_))`},
		{`my ($x) = @a or die;`, `(my (= $x @a))`},
	}
	for _, c := range cases {
		root := parse.Parse([]byte(c.src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", c.src, shape(root))
			continue
		}
		d := firstOfKind(root, parse.Declaration)
		if d == nil {
			t.Errorf("%q: no declaration in %s", c.src, shape(root))
			continue
		}
		if got := shape(d); got != c.decl {
			t.Errorf("%q: declaration\n  got  %s\n  want %s", c.src, got, c.decl)
		}
	}
}
