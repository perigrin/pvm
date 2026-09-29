// ABOUTME: `system { "ls" } "ls", "-l"` and `exec $shell "-sh"`: the program slot takes a block,
// ABOUTME: scalar or bareword with no comma after it, the same slot print has for a handle.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestIndirectProgramSlot holds perl's reading of system's and exec's first
// slot. Measured on 5.42.0 with -MO=Deparse:
//
//	my $rc = system { "ls" } "ls", "-l";    system({'ls';} 'ls', '-l')
//	exec { $p } @args;                      exec {$p;} @args;
//	system $shell "-sh";                    system $shell '-sh';
//
// The parser gave only print, printf and say a handle slot, so the block
// became system's sole argument and the list a statement of its own, at
// Unknown=0 -- `system({"ls"});"ls" , "-l";`.
func TestIndirectProgramSlot(t *testing.T) {
	cases := []struct {
		src   string
		canon string
	}{
		{`my $rc = system { "ls" } "ls", "-l";`, `my $rc = system({"ls";} "ls" , "-l");`},
		{`exec { $p } @args;`, `exec({$p;} @args);`},
		{`system $shell "-sh";`, `system($shell "-sh");`},
		{`system "ls", "-l";`, `system("ls" , "-l");`},
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
