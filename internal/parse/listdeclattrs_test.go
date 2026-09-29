// ABOUTME: A list declaration takes attributes between the list and its initialiser:
// ABOUTME: `my ($c, @g, %b) : teapots = qw[a b c];`.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestListDeclarationAttributes: the list target was parsed as one
// expression, where `: teapots` read as a ternary's colon with nothing to
// match. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'package n; sub MODIFY_SCALAR_ATTRIBUTES{}
//	      sub MODIFY_ARRAY_ATTRIBUTES{} sub MODIFY_HASH_ATTRIBUTES{}
//	      my ($c, @g, %b) : teapots = qw[a b c];'
//	my($c, @g, %b) :teapots = ('a', 'b', 'c');
//
// perl.git t/op/attrs.t:348 and t/uni/attrs.t:190.
func TestListDeclarationAttributes(t *testing.T) {
	for _, src := range []string{
		`my ($c, @g, %b) : teapots = qw[a b c];`,
		`my ($d, $e) :shared;`,
		`my ($f, $g) = @_;`,
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
