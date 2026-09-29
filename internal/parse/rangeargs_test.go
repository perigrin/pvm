// ABOUTME: A range operator after a named unary or sub name is infix, so the word
// ABOUTME: takes no argument: `undef..2` is `(undef) .. 2`.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestRangeEndsTheArgumentList: `..` and `...` cannot start a term, so a word
// before one takes nothing, as it takes nothing before `.` or `==`. Measured
// on 5.42.0:
//
//	$ perl -MO=Deparse,-p -e 'sub g{1} sub f { my @x = (undef..2);
//	      my @y = (shift .. 2); my @z = (lc .. 2); my @w = (g ... 2) }'
//	(my(@x) = ((undef) .. 2));
//	(my(@y) = ((shift()) .. 2));
//	(my(@z) = (lc($_) .. 2));
//	(my(@w) = (g() ... 2));
//
// perl.git t/op/range.t has `is(join(":",undef..2), '0:1:2');`, which read
// `..` as undef's argument and refused twice.
func TestRangeEndsTheArgumentList(t *testing.T) {
	for _, src := range []string{
		"my @x = (undef..2);",
		"my @y = (shift .. 2);",
		"my @z = (lc .. 2);",
		"sub g {1} my @w = (g ... 2);",
		`is(join(":",undef..2), 1);`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		canon := parse.Canon(root, []byte(src))
		again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon))
		if canon != again {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", src, canon, again)
		}
	}
}
