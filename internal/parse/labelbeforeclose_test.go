// ABOUTME: A label may end a block with no statement after it: `sub f { ...; _: }`
// ABOUTME: labels an empty statement, and the `}` still closes the block.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestLabelBeforeClose: measured on 5.42.0,
//
//	$ perl -MO=Deparse -e 'sub f { $x = 1; _: } sub g { L: }'
//	sub f {
//	    $x = 1;
//	    _: ;
//
// perl.git t/op/attrs.t:306 and 324, `sub MODIFY_CODE_ATTRIBUTES { ...; _: }`.
func TestLabelBeforeClose(t *testing.T) {
	for _, src := range []string{
		`sub f { $x = 1; _: }`,
		`sub g { L: }`,
		`L:`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		if got := root.SourceText([]byte(src)); got != src {
			t.Errorf("%q: round-trip gave %q", src, got)
		}
		canon := parse.Canon(root, []byte(src))
		if again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon)); again != canon {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", src, canon, again)
		}
	}
}
