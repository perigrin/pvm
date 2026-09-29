// ABOUTME: The braces after `&` hold a block as `${...}`'s do: `&{; $c }(1)`
// ABOUTME: and `\&{; do { ... } }` read their statements.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestAmpDerefBlock: measured on 5.42.0,
//
//	$ perl -MO=Deparse -e 'prototype \&{; do { f: sub ($) {} } }; &{; $c }(1);'
//	prototype \&{do {
//	    f: sub ($) {
//	    }
//	};};
//	&{$c;}(1);
//
// perl.git t/op/goto.t:674.
func TestAmpDerefBlock(t *testing.T) {
	for _, src := range []string{
		`prototype \&{; do { f: sub ($) {} } };`,
		`&{; $c }(1);`,
		`&{; 1 };`,
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
