// ABOUTME: A glob's name is any variable name: `*1`, `*^R` and `*-` as well as `*foo`,
// ABOUTME: so a digit, caret or punctuation name directly after `*` is part of the glob.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestGlobWithNonIdentifierName holds perl's reading of a glob whose name is
// not an identifier. Measured on 5.42.0, each deparses as written:
//
//	local *a = *1;        *^R = *foo;        *X = *-;
//
// The glob term read only a Word, a Variable or `{` after the star, so `*1`
// became a star and a stray number -- op/method.t, re/pat.t and re/subst.t
// each first refused there.
func TestGlobWithNonIdentifierName(t *testing.T) {
	cases := []struct {
		src   string
		canon string
	}{
		{"local *a = *1;", "local *a = *1;"},
		{"local *1 = sub { 123 };", "local *1 = sub {123;};"},
		{"*^R = *foo;", "*^R = *foo;"},
		{"*X = *-;", "*X = *-;"},
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
