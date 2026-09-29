// ABOUTME: `*]` is the glob named `]`, not a star and a closing bracket that would
// ABOUTME: pop whatever bracket is open.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestGlobCloseBracket: after a glob-sigil star toke.c's scan_ident takes
// one punctuation byte as the name, `]` among them. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'my $g = *]; my @a = (1, *], 2);'
//	my $g = *];
//	my(@a) = (1, *], 2);
//
// perl.git t/op/tie_fetch_count.t:189, `tie my $var4 => 'main', *];`. As a
// closer the `]` popped the enclosing bracket and the rest of the file
// shifted.
func TestGlobCloseBracket(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"my $g = *];", "my $g = *];"},
		{"my @a = (1, *], 2);", "my @a = (1 , *] , 2);"},
		{"my @b = ([1], 2);", "my @b = ([1], 2);"},
	} {
		root := parse.Parse([]byte(tc.src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", tc.src, shape(root))
			continue
		}
		got := parse.Canon(root, []byte(tc.src))
		if again := parse.Canon(parse.Parse([]byte(got)), []byte(got)); again != got {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", tc.src, got, again)
		}
	}
}
