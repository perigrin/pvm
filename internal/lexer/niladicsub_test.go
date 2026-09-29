// ABOUTME: A niladic builtin's name after `sub` is a declared name, so the brace
// ABOUTME: after it opens the sub's body rather than following a value.

package lexer

import "testing"

// TestNiladicSubNameOpensABlock: `sub time { 1 }` declares a sub named time.
// Measured on 5.42.0, `perl -MO=Deparse -e 'sub time { 1 } sub wait { 2 }'`
// keeps both bodies. The niladic table answered first and left an operator
// expected, so the `{` was never the body and the source split into a
// forward declaration and a bare block at Unknown=0.
func TestNiladicSubNameOpensABlock(t *testing.T) {
	for _, src := range []string{"sub time { 1 }", "sub wait { 2 }", "sub getppid { 3 }"} {
		opens, found := lastBraceOpensBlock(src)
		if !found || !opens {
			t.Errorf("%q: the `{` after a sub name opens its body", src)
		}
	}
}
