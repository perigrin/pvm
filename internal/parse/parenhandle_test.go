// ABOUTME: A filehandle inside print's parens is a handle when no comma follows it, as
// ABOUTME: outside them: `print($fh "x")` is canon's own spelling and must re-parse.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestParenthesisedFilehandle holds perl's reading of a handle in a
// parenthesised print. Measured on 5.42.0 with -MO=Deparse:
//
//	print($fh "x");      print $fh 'x';       a handle
//	print(STDERR "y");   print STDERR 'y';    a handle
//	print($a, "z");      print $a, 'z';       a list: the comma says so
//
// The parenthesised path read only the `{$fh}` block form, on the belief that
// inside parens a scalar or bareword in the first position is an ordinary
// argument. perl says otherwise, and canon writes a handle as `print($fh "x")`
// -- so its own emission refused on re-parse.
func TestParenthesisedFilehandle(t *testing.T) {
	cases := []struct {
		src   string
		canon string
	}{
		{`print($fh "x");`, `print($fh "x");`},
		{`print(STDERR "y");`, `print(STDERR "y");`},
		{`print($a, "z");`, `print($a , "z");`},
		{`print({$fh} "x");`, `print({$fh;} "x");`},
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
