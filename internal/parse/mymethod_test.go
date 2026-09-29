// ABOUTME: `my method NAME {...}` declares a lexical method under class syntax, as
// ABOUTME: `my sub NAME {...}` declares a lexical sub.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestLexicalMethod: measured on 5.42.0,
//
//	$ perl -MO=Deparse -e 'use feature "class"; no warnings;
//	      class X { field $f = 1; my method priv { return $f } method pub { 1 } }'
//	    my method priv {
//	        return $f;
//	    }
//
// perl.git t/class/method.t:104, 120 and 153.
func TestLexicalMethod(t *testing.T) {
	for _, src := range []string{
		`use feature 'class'; class X { field $f = 1; my method priv { return $f } method pub { priv($self) } }`,
		`use feature 'class'; class Y { my method priv ( $y ) { return $y } }`,
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
