// ABOUTME: A foreach loop variable may be a dereferenced scalar: `for ${*$f} (...)`
// ABOUTME: and `for $$r (...)` iterate through what the reference names.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestForeachDerefLoopVariable: the loop-variable slot took a plain
// Variable or a declarator, and a deref sigil fell through to the head,
// where `${*$f} (5,11,33)` is no parenthesised list. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'no strict "refs"; for ${*$f} (5,11,33) { $i++ }
//	      for $$r (1,2) { 1 }'
//	foreach ${*$f;} (5, 11, 33) {
//	foreach $$r (1, 2) {
//
// perl.git t/op/for.t:767.
func TestForeachDerefLoopVariable(t *testing.T) {
	for _, src := range []string{
		"for ${*$f} (5,11,33) { $i++ }",
		"for $$r (1,2) { 1 }",
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
		again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon))
		if canon != again {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", src, canon, again)
		}
	}
}
