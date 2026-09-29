// ABOUTME: `return` inside an expression is the same list operator it is at a
// ABOUTME: statement's start: `$a and return 1` returns 1.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestReturnInsideAnExpression: toke.c lexes return as OLDLOP(OP_RETURN), a
// list operator wherever it stands, and only the statement-start path read
// it as one. In an expression it went to the generic call path and refused
// on its first argument. Measured on 5.42.0:
//
//	$ perl -MO=Deparse,-p -e 'sub f { $a and return 1, 2; $a || return 3;
//	      my @r = (2, 3, return @a ? 1 : ()); }'
//	($a and (return 1, 2));
//	($a or (return 3));
//	(my(@r) = (2, 3, (return (@a ? 1 : ()))));
//
// perl.git t/ spells `COND and return EXPR` often: op/pack.t:46 has
// `$err =~ $e and return 1;`, op/stat.t:694 `... and return 1;`, and
// op/sprintf2.t:645 `(...) && return $1;`.
func TestReturnInsideAnExpression(t *testing.T) {
	for _, src := range []string{
		"$a and return 1;",
		"$a or return 1, 2;",
		"$a && return 3;",
		"$err =~ $e and return 1;",
		"my @r = (2, 3, return @a ? 1 : ());",
		"$a and return;",
		// A bare return takes nothing before a ternary's `:` or a closer:
		// `($c ? (return) : uc($_))`, `f((return))`.
		"my @r = map { $_ eq 'm' ? return : uc($_) } @a;",
		"f(return);",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		if firstOfKind(root, parse.LoopControl) == nil {
			t.Errorf("%q: the return is a call, not a return; got %s", src, shape(root))
		}
		canon := parse.Canon(root, []byte(src))
		again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon))
		if canon != again {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", src, canon, again)
		}
	}
}
