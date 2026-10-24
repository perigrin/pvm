// ABOUTME: `print $fh -1`: a scalar after a handle-taking list operator, then whitespace, is the
// ABOUTME: handle when what follows starts a term -- toke.c's yyl_dollar rule, held as canon.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestScalarHandleBeforeATermStart holds perl's reading of a scalar in a list
// operator's first slot. toke.c's yyl_dollar peeks past the whitespace after
// the scalar and, when the next characters can only start a term, lexes them
// as one; the grammar then has no reading but the handle. Measured on 5.42.0
// with -MO=Deparse:
//
//	print $x &y;      print $x &y;        a call to y
//	print $x *y;      print $x *y;        a glob
//	print $x -1;      print $x -1;        a negative number
//	print $x +1;      print $x 1;
//	print $x -$y;     print $x -$y;
//	print $x --$y;    print $x --$y;
//	print $x -e;      print $x -e $_;     a filetest
//	print $x <y>;     print $x readline y;
//	print($x -1);     print $x -1;        parenthesised too
//	sort $x -1;       sort $x -1;         sort's comparator slot
//
// and, with no lexer state involved, a prefix operator no infix spells:
//
//	print $x !1;      print $x !1;
//	print $x ~1;      print $x 18446744073709551614;
//	print $x!1;       print $x !1;
//
// The neighbours that stay operators:
//
//	print $x & $y;    print $x & $y;      space after the `&`
//	print $x &$y;     print $x & $y;      `&` before a non-identifier
//	print $x &&1;     print $x && 1;
//	print $x - 1;     print $x - 1;       space after the `-`
//	print $x-1;       print $x - 1;       no space after the scalar
//	print $x -=1;     print $x -= 1;
//	print $x ->m;     print $x->m;
//	print $x ?1:2;    print $x ? 1 : 2;   `?` is a ternary either way
//	print $x !~ 1;    print !($x =~ /1/);
//	print STDOUT $x -1;   print STDOUT $x - 1;   not the first slot
//	print $$x -1;     print $$x - 1;      not a plain scalar
//	print $x[0] -1;   print $x[0] - 1;
//
// The fuzzer's `print$0 &y` read `&` as a bitwise and, and canon's `y()`
// re-lexed as a transliteration.
func TestScalarHandleBeforeATermStart(t *testing.T) {
	cases := []struct {
		src   string
		canon string
	}{
		{`print$0 &y;`, `print $0 &y;`},
		{`print $x &y;`, `print $x &y;`},
		{`print $x &y::z;`, `print $x &y::z;`},
		{`print $x *y;`, `print $x *y;`},
		{`print $x -1;`, `print $x -1;`},
		{`print $x +1;`, `print $x +1;`},
		{`print $x -$y;`, `print $x -$y;`},
		{`print $x --$y;`, `print $x --$y;`},
		{`print $x <y>;`, `print $x <y>;`},
		{`print($x -1);`, `print($x -1);`},
		{`printf $x -1;`, `printf $x -1;`},
		{`say $x -1;`, `say $x -1;`},
		{`system $x -1;`, `system $x -1;`},
		{`exec $x -1;`, `exec $x -1;`},
		{`sort $x -1;`, `sort $x -1;`},
		{`print $x !1;`, `print $x !1;`},
		{`print $x ~1;`, `print $x ~1;`},
		{`print $x!1;`, `print $x !1;`},

		{`print $x & $y;`, `print $x & $y;`},
		{`print $x &$y;`, `print $x & $y;`},
		{`print $x &&1;`, `print $x && 1;`},
		{`print $x - 1;`, `print $x - 1;`},
		{`print $x-1;`, `print $x - 1;`},
		{`print $x -=1;`, `print $x -= 1;`},
		{`print $x ?1:2;`, `print $x ? 1 : 2;`},
		{`print $x !~ 1;`, `print $x !~ 1;`},
		{`print STDOUT $x -1;`, `print STDOUT $x - 1;`},
		{`print $$x -1;`, `print $$x - 1;`},
		{`print $x[0] -1;`, `print $x[0] - 1;`},
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
