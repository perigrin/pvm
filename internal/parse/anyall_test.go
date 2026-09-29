// ABOUTME: `any` and `all` take a block as grep does when their features are on,
// ABOUTME: in the parenthesised spelling too.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestAnyAllTakeABlock: `use feature qw(keyword_any keyword_all)` makes both
// builtins that take BLOCK LIST. Measured on 5.42.0:
//
//	$ perl -e 'use feature "keyword_any"; no warnings;
//	      print +(any( { $_ > 10 } 1 .. 20)) ? "y\n" : "n\n";'
//	y
//	$ perl -e 'use v5.42; print +(any( { $_ > 10 } 1 .. 20)) ? "y\n" : "n\n";'
//	syntax error at -e line 1, near "} 1"
//
// perl.git t/op/any_all.t:19-20.
func TestAnyAllTakeABlock(t *testing.T) {
	for _, src := range []string{
		`use feature qw(keyword_any keyword_all); ok( (any( { $_ > 10 } 1 .. 20) ), "x");`,
		`use feature qw(keyword_any keyword_all); ok( (all( { $_ < 10 } 1 .. 9) ), "x");`,
		`use feature qw(keyword_any keyword_all); ok( !(all { $_ < 10 } 1 .. 20), "x");`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		if firstOfKind(root, parse.Block) == nil {
			t.Errorf("%q: the braces are the block; got %s", src, shape(root))
		}
		canon := parse.Canon(root, []byte(src))
		if again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon)); again != canon {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", src, canon, again)
		}
	}
}
