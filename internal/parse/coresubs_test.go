// ABOUTME: The subs perl defines itself at startup -- utf8::encode, Internals::SvREADONLY,
// ABOUTME: PerlIO::get_layers and the rest of universal.c's table -- are known without a declaration.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestInterpreterDefinedSubs holds perl's own XSUBs, which exist before any
// file is compiled: universal.c registers them from `these_details[]` at
// interpreter start. A parenless call to one resolves on 5.42.0 with nothing
// loaded:
//
//	$ perl -e 'my $s = "\x{100}"; utf8::encode $s; print length $s'
//	2
//
// The parser declared nothing for them, so a parenless call with an argument
// refused. 12 perl.git t/ files first refuse at one, `utf8::encode $x` and
// `Internals::SvREADONLY $x, 1` among them.
//
// The prototypes are universal.c's: NULL is no prototype, a list operator;
// `\[$%@];$` makes Internals::SvREADONLY take one reference and one optional
// scalar.
func TestInterpreterDefinedSubs(t *testing.T) {
	cases := []struct {
		src   string
		canon string
	}{
		{`utf8::encode $x;`, `utf8::encode($x);`},
		{`utf8::upgrade $x if $u;`, `utf8::upgrade($x) if $u;`},
		{`Internals::SvREADONLY $x, 1;`, `Internals::SvREADONLY($x , 1);`},
		{`my @l = PerlIO::get_layers $fh;`, `my @l = PerlIO::get_layers($fh);`},
		{`my $p = re::regexp_pattern $qr;`, `my $p = re::regexp_pattern($qr);`},
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
	}

	// NEGATIVE: an arbitrary qualified name is still unknown. Only the table
	// perl itself registers is pre-declared.
	src := []byte("Foo::bar $x, 1;")
	if got := countUnknown(parse.Parse(src)); got == 0 {
		t.Errorf("Parse(%q) has no Unknown: Foo::bar is not an interpreter sub", src)
	}
}
