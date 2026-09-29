// ABOUTME: A comma with no element after it is a separator with nothing to separate,
// ABOUTME: before another comma, a terminator, a modifier or a low-precedence word.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestEmptyCommaSlot holds perl's reading of a comma with no element after
// it, beyond the closer case 652f4792 covered. Measured on 5.42.0 with
// `perl -MO=Deparse`:
//
//	sub f {} f(1, , 2);                 f 1, 2;
//	sub skip {} skip "x", 2, if $m;     skip 'x', 2 if $m;
//	close $fh, or die;                  die unless close $fh;
//	open my $fh, , "<", $f;             open my $fh, '<', $f;
//	my $aa, $bb, $cc;                   my $aa, $bb, $cc;
func TestEmptyCommaSlot(t *testing.T) {
	cases := []struct {
		src   string
		canon string
	}{
		{"sub f {} f(1, , 2);", "sub f {} f(1 , 2);"},
		{"sub f {} f 1, , 2;", "sub f {} f(1 , 2);"},
		{`sub skip {} skip "x", 2, if $m;`, `sub skip {} skip("x" , 2) if $m;`},
		{"close $fh, or die;", "close($fh) or die();"},
		{`open my $fh, , "<", $f;`, `open(my $fh , "<" , $f);`},
		{"my $aa, $bb, $cc;", "my $aa , $bb , $cc;"},
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
