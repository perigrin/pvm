// ABOUTME: `map ({...} LIST)` with a space before the paren takes a block, as the
// ABOUTME: unspaced `map({...} LIST)` does.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestListOpSpacedParenBlock: the lexer carries a block-taking word's brace
// lookahead into its `(`, and the whitespace token between them reset the
// carry, so `map ({"$_\n"} @x)` read its braces as an anonymous hash.
// Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'my @e = map ({"$_\n"} @x); my @g = grep ({ $_ } @y);
//	      my @s = sort ({ $a <=> $b } @z);'
//	my(@e) = map({"$_\n";} @x);
//	my(@g) = grep({$_;} @y);
//	my(@s) = (sort {$a <=> $b} @z);
//
// perl.git t/comp/retainedlines.t:55.
func TestListOpSpacedParenBlock(t *testing.T) {
	for _, src := range []string{
		`my @e = map ({"$_\n"} @x);`,
		`my @g = grep ({ $_ } @y);`,
		`my @s = sort ({ $a <=> $b } @z);`,
		`my @e = (undef, map ({"$_\n"} split "\n", $prog), "\n");`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		if firstOfKind(root, parse.Block) == nil {
			t.Errorf("%q: the braces are the block, not a hash; got %s", src, shape(root))
		}
	}
}
