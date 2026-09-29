// ABOUTME: `DESTROY { ... }` and `AUTOLOAD { ... }` at statement start declare the sub,
// ABOUTME: as `BEGIN { ... }` does -- perl reads them as `sub DESTROY { ... }`.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestSpecialSubWithoutSub holds toke.c's rule for the words that declare a
// sub with no `sub` before them:
//
//	case KEY_AUTOLOAD: case KEY_DESTROY: case KEY_BEGIN: ... case KEY_END:
//	    if (PL_expect == XSTATE)
//	        return yyl_sub(aTHX_ PL_bufptr, key);
//
// The phasers were read; AUTOLOAD and DESTROY were not, so their block read as
// an anonymous hash after a call. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'package Thingie; DESTROY { 1 } AUTOLOAD { 2 }'
//	package Thingie;
//	sub DESTROY { 1; }
//	sub AUTOLOAD { 2; }
func TestSpecialSubWithoutSub(t *testing.T) {
	cases := []struct {
		src   string
		canon string
	}{
		{"DESTROY { $messages .= 'destroyed ' }\nf();", "sub DESTROY {$messages .= 'destroyed ';} f();"},
		{"AUTOLOAD { return 2 }", "sub AUTOLOAD {return 2;}"},
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

	// Not at statement start, the word is a word: a method named DESTROY.
	src := []byte("$obj->DESTROY;")
	if got := countUnknown(parse.Parse(src)); got != 0 {
		t.Errorf("Parse(%q) has %d Unknown, want 0", src, got)
	}
}
