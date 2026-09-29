// ABOUTME: An empty slot inside `[...]` or `{...}` is dropped as it is inside
// ABOUTME: `(...)`: `["x",, "y"]` has two elements.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestBracketEmptySlot: parseParenList drops a separator with no element
// after it, and parseBracketed did not. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'my $a = ["x",, "y"]; my $h = {a => 1,, b => 2};
//	      my $c = [1, => 2];'
//	my $a = ['x', 'y'];
//	my $h = {'a', 1, 'b', 2};
//	my $c = [1, 2];
//
// perl.git t/porting/bench.t:274-282, `"--bisect=Ir",,`.
func TestBracketEmptySlot(t *testing.T) {
	for _, src := range []string{
		`my $a = ["x",, "y"];`,
		`my $h = {a => 1,, b => 2};`,
		`my $c = [1, => 2];`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		canon := parse.Canon(root, []byte(src))
		if again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon)); again != canon {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", src, canon, again)
		}
	}
}
