// ABOUTME: A bare deref sigil takes only the variable after it: `$$r[0]` is element
// ABOUTME: 0 of @$r, the subscript applying to the dereference.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestBareDerefTakesTheVariable: bpDeref was 300, below the subscripts at
// 320, so the sigil's operand parse swallowed the subscript and `$$r[0]` was
// a deref of `$r[0]`. Measured on 5.42.0:
//
//	$ perl -e 'my $r=[5,6]; print $$r[1], " ", join(",", @$r[0,1]), "\n"'
//	6 5,6
//	$ perl -MO=Deparse -e 'my $y = $$r[0]; my @x = @$r[1,2]; my $z = $$h{k};'
//	my $y = $r->[0];
//	my(@x) = @$r[1, 2];
//	my $z = $h->{'k'};
func TestBareDerefTakesTheVariable(t *testing.T) {
	for _, src := range []string{
		"my $y = $$r[0];",
		"my @x = @$r[1,2];",
		"my $z = $$h{k};",
	} {
		root := parse.Parse([]byte(src))
		idx := firstOfKind(root, parse.Index)
		if idx == nil || len(idx.Children) == 0 || idx.Children[0].Kind != parse.Unary {
			t.Errorf("%q: want a subscript applied to the dereference; got %s", src, shape(root))
		}
	}
}
