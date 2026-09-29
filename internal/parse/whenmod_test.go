// ABOUTME: `when` is a statement modifier inside a given block: `$ok = 1 when undef;`
// ABOUTME: is a conditional statement, not trailing tokens.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestWhenModifier: perly.y admits `expr KW_WHEN condition` under the switch
// feature. Measured on 5.42.0:
//
//	$ perl -e 'use feature "switch"; no warnings; my $ok;
//	      given (undef) { $ok = 1 when undef; } print "ok=$ok\n"'
//	ok=1
//
// perl.git t/op/switch.t:810.
func TestWhenModifier(t *testing.T) {
	for _, src := range []string{
		`use feature "switch"; given ($x) { $ok = 1 when undef; }`,
		`use feature "switch"; given ($x) { print "a" when [1,2]; }`,
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
