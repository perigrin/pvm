// ABOUTME: `*@` is the glob named `@`, not a star and an array: toke.c's yyl_star
// ABOUTME: reads the name with scan_ident wherever an operator is not expected.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestGlobAtSign: in term position toke.c's yyl_star scans the name after
// the star with scan_ident, which takes one punctuation byte. Measured on
// 5.42.0:
//
//	$ perl -MO=Deparse -e 'sub f { local *@; my $g = \*@; my $n = $a * @b; }'
//	    local *@;
//	    my $g = \*@;
//	    my $n = $a * @b;
//
// perl.git t/op/local.t:824. After a term the star is multiplication and
// `@b` is still an array.
func TestGlobAtSign(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"local *@;", "local *@;"},
		{"my $g = \\*@;", "my $g = \\*@;"},
		{"my $n = $a * @b;", "my $n = $a * @b;"},
	} {
		root := parse.Parse([]byte(tc.src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", tc.src, shape(root))
			continue
		}
		if got := parse.Canon(root, []byte(tc.src)); got != tc.want {
			t.Errorf("%q: canon %q, want %q", tc.src, got, tc.want)
		}
	}
}
