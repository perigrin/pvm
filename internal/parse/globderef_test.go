// ABOUTME: A glob dereferences what the scalar sigils do: a deref'd scalar
// ABOUTME: `*$$foo`, and braces that hold statements `*{;undef}`.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestGlobDeref: `*` took a name, a variable or one braced expression, so
// the two deref forms `$` and `@` already had refused after it. Measured on
// 5.42.0:
//
//	$ perl -MO=Deparse -e '*$$foo = bless [], _121242::;'
//	*$$foo = bless([], '_121242');
//	$ perl -MO=Deparse -e '*{;undef} = 3'
//	*{undef;} = 3;
//
// perl.git t/op/gv.t:984 and :1020.
func TestGlobDeref(t *testing.T) {
	for _, src := range []string{
		`*$$foo = bless [], _121242::;`,
		`eval { *{;undef} = 3 };`,
		`*$$$foo = 1;`,
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
