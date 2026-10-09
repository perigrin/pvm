// ABOUTME: A `my` or `our` in term position with no variable after it refuses as not_a_term.
// ABOUTME: Read as a call to a sub named my, its canon `my()` re-read as an empty declaration.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestDeclaratorWithoutTargetRefuses is the parser fuzzer's `!my=0`. `my`
// and `our` are always keywords, and in term position they take a variable,
// a `(` list, or a class name before one. Measured on 5.42.0, `!my=0` is
// "syntax error ... near "my="", and `!my`, `-my`, `f(my)`, `f(my, 1)`,
// `[my]`, `print my;` and `$x = our;` are each syntax errors. Read as a call
// to a sub named `my`, canon wrote `!my() = 0`, which re-read as a
// declaration of the empty list.
//
// The same word is still a bareword where perl reads one: a hash key, and
// the left of a fat comma.
func TestDeclaratorWithoutTargetRefuses(t *testing.T) {
	for _, src := range []string{
		"!my=0", "!my", "-my", "f(my);", "f(my, 1);", "my $x = [my];", "print my;", "$x = our;",
	} {
		found := collectUnknown(parse.Parse([]byte(src)))
		if len(found) == 0 || found[0].Refusal != parse.NotATerm {
			t.Errorf("%q: want a not_a_term refusal, got %v", src, refusalCodes(found))
		}
	}
	for _, src := range []string{
		"my %h; $h{my};", "my %h = (my => 1);", "f(my $x);", "f(my ($a, $b));", "our $x; !our $x;",
	} {
		if found := collectUnknown(parse.Parse([]byte(src))); len(found) != 0 {
			t.Errorf("%q: perl compiles this, but it refuses: %v", src, refusalCodes(found))
		}
	}
}
