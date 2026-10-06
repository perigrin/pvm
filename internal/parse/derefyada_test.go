// ABOUTME: A deref's braces may hold the yada statement: `${...}++` dies
// ABOUTME: Unimplemented, as `sub f { ... }` does.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestDerefYada: `...` is a statement and only that (perly.y:759,
// bare_statement_yadayada), so braces that begin with it hold statements.
// Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'eval { ${...}++ };'
//	eval {
//	    do {
//	        ++${die 'Unimplemented';}
//	    }
//	};
//
// perl.git t/base/lex.t:504.
func TestDerefYada(t *testing.T) {
	for _, src := range []string{
		`eval { ${...}++ };`,
		`@{...};`,
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
