// ABOUTME: The builtin:: functions are the interpreter's own, and `use builtin LIST`
// ABOUTME: imports them with builtin.c's prototypes, so `blessed $obj` is a call.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestBuiltinImportsHaveShapes: builtin.c's boot_core_builtin registers every
// builtin:: function at interpreter start, with a prototype chosen by its
// call checker -- "" for the constants, "$" for the one-argument functions,
// "@" for indexed. `use builtin LIST` imports them by name. Unknown, a word
// before `$obj` is an indirect method call; known, it is a named unary.
// Measured on 5.42.0:
//
//	$ perl -MO=Deparse,-p -e 'use builtin qw(blessed reftype);
//	      is(blessed $obj, "T", "x"); my $w = builtin::reftype $r, 1;
//	      use builtin "true"; my @a = (true, 1);'
//	is(scalar(builtin::blessed($obj)), 'T', 'x');
//	((my($w) = builtin::reftype($r)), '???');
//	(my(@a) = ((!0), 1));
//
// class/construct.t:29 has `is(blessed $obj, "Testcase1", ...)` after
// `use builtin qw( blessed reftype );`.
func TestBuiltinImportsHaveShapes(t *testing.T) {
	for _, tc := range []struct {
		src, name string
		args      int
	}{
		{`use builtin qw( blessed reftype ); is(blessed $obj, "T", "x");`, "blessed", 1},
		{`my $w = builtin::reftype $r, 1;`, "builtin::reftype", 1},
		{`use builtin 'true'; my @a = (true, 1);`, "true", 0},
		// A list operator: it takes everything, so only resolution is asserted.
		{`use builtin qw(indexed); my @p = indexed @a, 1;`, "indexed", -1},
	} {
		root := parse.Parse([]byte(tc.src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", tc.src, shape(root))
			continue
		}
		call := findCall(root, tc.name)
		if call == nil {
			t.Errorf("%q: no call to %s; got %s", tc.src, tc.name, shape(root))
			continue
		}
		if !call.Resolved {
			t.Errorf("%q: %s is the interpreter's and must resolve", tc.src, tc.name)
		}
		if tc.args >= 0 && len(call.Children) != tc.args {
			t.Errorf("%q: %s takes %d argument(s), got %d; tree %s",
				tc.src, tc.name, tc.args, len(call.Children), shape(root))
		}
	}
}
