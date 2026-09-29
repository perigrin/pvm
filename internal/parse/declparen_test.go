// ABOUTME: A declaration with an initialiser binds like an assignment, so canon keeps
// ABOUTME: the parens a tighter operator needs: `(my $s = "abc") =~ /x/` is not `my $s = "abc" =~ /x/`.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestParenthesisedDeclarationCanon holds the grouping canon dropped at
// Unknown=0. bindingPower gave every Declaration an atom's power, so a
// declaration with an initialiser was never parenthesised, and
//
//	() = (my $str = "abc") =~ /(...)/;
//
// re-emitted as `() = my $str = "abc" =~ /(...)/;`, which binds the match to
// "abc" before the assignment -- a different program. perl's own
// re/pat_advanced.t line 23 is this statement. Measured on 5.42.0, Deparse
// keeps the parens:
//
//	$ perl -MO=Deparse -e '() = (my $str = "abc") =~ /(...)/;'
//	() = (my $str = 'abc') =~ /(...)/;
func TestParenthesisedDeclarationCanon(t *testing.T) {
	cases := []struct {
		src   string
		canon string
	}{
		{`() = (my $str = "abc") =~ /(...)/;`, `() = (my $str = "abc") =~ /(...)/;`},
		{`my $y = (my ($a, $b) = @_) + 1;`, `my $y = (my ($a, $b) = @_) + 1;`},
		// A declaration standing alone as a statement is not wrapped.
		{`my $x = 1;`, `my $x = 1;`},
	}
	for _, c := range cases {
		src := []byte(c.src)
		n := parse.Parse(src)
		if got := countUnknown(n); got != 0 {
			t.Errorf("Parse(%q) has %d Unknown, want 0", c.src, got)
		}
		got := strings.TrimSpace(parse.Canon(n, src))
		if got != c.canon {
			t.Errorf("Canon(%q):\n  got  %s\n  want %s", c.src, got, c.canon)
		}
		again := strings.TrimSpace(parse.Canon(parse.Parse([]byte(got)), []byte(got)))
		if again != got {
			t.Errorf("canon is not a fixpoint:\n  once  %s\n  twice %s", got, again)
		}
	}
}
