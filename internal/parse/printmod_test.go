// ABOUTME: `print $x if $c` prints $x -- a statement modifier after the scalar does not make
// ABOUTME: it a filehandle, because a modifier word does not start a term.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestPrintScalarBeforeModifier holds perl's reading of a scalar followed by a
// statement modifier in print's first slot. Measured on 5.42.0 with
// -MO=Deparse, `print $warn if length $warn;` is kept as written: $warn is
// printed, not used as a handle.
//
// The handle slot asks startsTerm of the token after the scalar, and
// startsTerm answered yes for every word that is not an infix operator --
// `if`, `unless`, `while`, `until`, `for` and `foreach` included -- so the
// scalar became a handle and the modifier was left with nothing to attach to.
func TestPrintScalarBeforeModifier(t *testing.T) {
	for _, c := range []struct{ src, canon string }{
		{`print $warn if length $warn;`, `print($warn) if length($warn);`},
		{`print $w unless $q;`, `print($w) unless $q;`},
		{`print $w foreach @a;`, `print($w) foreach @a;`},
	} {
		src := []byte(c.src)
		n := parse.Parse(src)
		if got := countUnknown(n); got != 0 {
			t.Errorf("Parse(%q) has %d Unknown, want 0", c.src, got)
		}
		if got := strings.TrimSpace(parse.Canon(n, src)); got != c.canon {
			t.Errorf("Canon(%q):\n  got  %s\n  want %s", c.src, got, c.canon)
		}
	}
	// REGRESSION GUARD: a handle followed by a real term is still a handle.
	src := []byte(`print $fh "x" if $c;`)
	if got := strings.TrimSpace(parse.Canon(parse.Parse(src), src)); got != `print($fh "x") if $c;` {
		t.Errorf("Canon(%q) = %s, want the handle kept", src, got)
	}
}
