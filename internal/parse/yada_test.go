// ABOUTME: `...` is the yada-yada STATEMENT -- `sub f { ... }` -- and nothing else;
// ABOUTME: perl rejects it as a term and with a modifier, so the parser reads it only whole.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestYadaYadaStatement holds perl's ellipsis statement. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'sub f { ... }'
//	sub f { die 'Unimplemented'; }
//	$ perl -c -e 'my $x = ...;'        syntax error near "= ..."
//	$ perl -c -e '... if 0;'           syntax error near "... if"
//
// A statement form only, and the parser read `...` nowhere: it is the range
// operator's spelling in the infix table, so at a statement start it refused.
func TestYadaYadaStatement(t *testing.T) {
	cases := []struct {
		src   string
		canon string
	}{
		{"sub f { ... }", "sub f {...;}"},
		{"{ ...; }", "{...;}"},
		{"eval { $i++; ...; $i += 10; 123 };", "eval {$i++;...;$i += 10;123;};"},
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

	// NEGATIVE: perl rejects `...` as a term, so it must still refuse there.
	src := []byte("my $x = ...;")
	if got := countUnknown(parse.Parse(src)); got == 0 {
		t.Errorf("Parse(%q) has no Unknown: `...` is a statement, not a term", src)
	}
}
