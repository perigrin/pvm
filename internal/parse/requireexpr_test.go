// ABOUTME: `require MODULE` is a named unary in an expression as well as a statement:
// ABOUTME: `... or require mro, diag "x"` and `(require Foo, 2)`.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestRequireInExpression: the parser read `require BAREWORD` only as a
// statement, where anything after the name was an import list, so a
// require inside an expression refused. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'sub diag {} ok($x) or require mro, diag "x";
//	      my @a = (require Foo, 2);'
//	require mro, diag('x') unless ok($x);
//	my(@a) = (require Foo, 2);
//
// perl.git t/op/universal.t:311.
func TestRequireInExpression(t *testing.T) {
	for _, src := range []string{
		`sub diag {} ok($x) or require mro, diag "x";`,
		`my @a = (require Foo, 2);`,
		`require Foo;`,
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
