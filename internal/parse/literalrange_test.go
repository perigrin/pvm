// ABOUTME: A literal range is the `..` operator over two numeric endpoints, not one token.
// ABOUTME: Holds the parser half of 01a0afc0; type inference over it is a separate question.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestLiteralRangeHasEndpoints: `(1..5)` is `..` over the numbers 1 and 5,
// the same shape `($a..$b)` takes. perl 5.42 folds it to a constant array of
// 1 through 5 (`-MO=Concise`), so the endpoints are the two numbers, never a
// string `1..5` or a float `1.` followed by `.5`.
func TestLiteralRangeHasEndpoints(t *testing.T) {
	for _, src := range []string{"my @r = (1..5);", "my @r = (1 .. 5);"} {
		root := parse.Parse([]byte(src))
		var rng *parse.Node
		for _, b := range findAll(root, parse.Binary) {
			if b.Text == ".." {
				rng = b
			}
		}
		if rng == nil {
			t.Errorf("%q: no `..` operator:\n%s", src, structure(root))
			continue
		}
		want := "binary ..\n  term 1\n  term 5\n"
		if got := structure(rng); got != want {
			t.Errorf("%q: want `..` over the two numbers:\ngot:\n%swant:\n%s", src, got, want)
		}
	}
}
