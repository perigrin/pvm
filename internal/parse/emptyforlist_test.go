// ABOUTME: A block-form for/foreach over an empty () is refused, as perl refuses it.
// ABOUTME: The C-style head, a nested empty list and the modifier form still parse.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestEmptyForeachListIsRefused: perl 5.42 calls `for () {}` a syntax error
// ("near "() ""), with or without a loop variable, while `for (())`,
// `for (;;)` and `1 for ();` compile. Accepting the empty head
// let `for()()` canonicalise into something that re-read differently.
func TestEmptyForeachListIsRefused(t *testing.T) {
	for _, src := range []string{
		"for(){}",
		"for () {}",
		"foreach(){}",
		"for()()",
		"for my $x () {}",
		"foreach $x () {}",
	} {
		if !containsKind(parse.Parse([]byte(src)), parse.Unknown) {
			t.Errorf("%q: perl refuses an empty foreach list; want an Unknown", src)
		}
	}
	for _, src := range []string{
		"for (()) {}",
		"for ((), ()) {}",
		"for(;;){}",
		"for (1) {}",
		"1 for ();",
	} {
		if root := parse.Parse([]byte(src)); containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; want no Unknown:\n%s", src, structure(root))
		}
	}
}
