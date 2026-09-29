// ABOUTME: A comma directly before a fat comma is an empty slot and is dropped, as
// ABOUTME: two commas in a row are: `tie $@, => 'main', 1` ties with two arguments.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestCommaBeforeFatComma: perl drops a separator with no element after it,
// and a fat comma is a separator. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'tie $@, => "main", 1; my @a = (1, => 2);'
//	tie $@, 'main', 1;
//	my(@a) = (1, 2);
//	$ perl -MO=Deparse -e 'sub foo {7} my @x = (foo, => 1);'
//	my(@x) = (foo(), 1);
//
// The fat comma does not reach back past the dropped comma to quote a word.
// perl.git t/op/tie_fetch_count.t:331-335.
func TestCommaBeforeFatComma(t *testing.T) {
	for _, src := range []string{
		`tie $@, => 'main', 1;`,
		`my @a = (1, => 2);`,
		`sub foo {7} my @x = (foo, => 1);`,
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
	root := parse.Parse([]byte(`sub foo {7} my @x = (foo, => 1);`))
	if call := findCall(root, "foo"); call == nil || call.Fat {
		t.Errorf("`foo, => 1`: foo is a call, not a quoted word; got %s", shape(root))
	}
}
