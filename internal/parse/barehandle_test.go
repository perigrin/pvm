// ABOUTME: A bareword after print is a filehandle when a term follows it and it is
// ABOUTME: neither a builtin nor a declared sub -- whatever its case, as perl decides.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestBarewordFilehandleCase holds perl's rule for a bareword in print's
// handle slot. Measured on 5.42.0 with -MO=Deparse:
//
//	print foo "x";                 print foo 'x';      a handle, lower case
//	print foo 1;                   print foo 1;        a handle
//	sub foo {1} print foo "x";     print foo('x');     a declared sub: a call
//	print length "x";              print 1;            a builtin
//
// The parser took only ALL-CAPS words as handles, on the belief that every
// corpus file follows that convention. io/print.t (`print foo "ok 6\n"`) and
// op/while.t (`print tmp "..."`) do not.
func TestBarewordFilehandleCase(t *testing.T) {
	cases := []struct {
		src   string
		canon string
	}{
		{`print foo "x";`, `print(foo "x");`},
		{`print tmp "y";`, `print(tmp "y");`},
		{`print STDERR "z";`, `print(STDERR "z");`},
		{`sub foo {1} print foo "x";`, `sub foo {1;} print(foo("x"));`},
		{`print length "x";`, `print(length("x"));`},
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
