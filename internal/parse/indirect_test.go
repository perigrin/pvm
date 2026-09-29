// ABOUTME: Indirect object notation on a class: `new Foo ARGS` is `'Foo'->new(ARGS)`, by
// ABOUTME: toke.c's intuit_method rule, off under the 5.36 bundle.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestIndirectObjectNotation holds toke.c's S_intuit_method for a word that is
// not a keyword. Measured on 5.42.0 with -MO=Deparse:
//
//	my $x = new Foo "a";                 my $x = 'Foo'->new('a');
//	my $x = new Foo(1, 2);               my $x = 'Foo'->new(1, 2);
//	new Foo;                             'Foo'->new;
//	my $o = new Foo::Bar;                my $o = 'Foo::Bar'->new;
//	$foo = doit $object "FOO";           $foo = $object->doit('FOO');
//	package Foo; sub new {1}
//	  package main; my $o = new Foo;     my $o = 'Foo'->new;       a known PACKAGE
//	sub new {1} my $o = new Foo;         my $o = new('Foo');       a known SUB
//	use v5.36; my $o = new Foo;          syntax error: indirect is off
//
// The class must be one this parse knows -- declared, or loaded by `use` or
// `require` -- for the reason parseIndirect records.
//
// The tree carries Indirect; canon keeps the spelling and adds the argument
// parens, which re-parses to the same call and stays faithful to the source. The `WORD $var` half -- `doit $object
// "FOO"` is `$object->doit('FOO')` -- is not read without a loader; see
// parseIndirect for the measurement behind that, and TestIndirectOnScalar
// for the half read with one.
func TestIndirectObjectNotation(t *testing.T) {
	cases := []struct {
		src   string
		canon string
	}{
		{`package Foo; package main; my $x = new Foo "a";`, `package Foo;package main;my $x = new Foo("a");`},
		{`package Foo; package main; my $x = new Foo(1, 2);`, `package Foo;package main;my $x = new Foo(1 , 2);`},
		{`use Foo; new Foo;`, `use Foo; new Foo();`},
		{`use Foo::Bar; my $o = new Foo::Bar;`, `use Foo::Bar; my $o = new Foo::Bar();`},
		{`package Foo; sub new {1} package main; my $o = new Foo;`,
			`package Foo;sub new {1;} package main;my $o = new Foo();`},
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

	// Not indirect: a class this parse does not know (see parseIndirect for
	// why that stays a refusal), a known sub with an unknown bareword, a fat
	// comma, a filehandle, and anything under the 5.36 bundle.
	for _, src := range []string{
		`my $x = new Foo "a";`,
		`sub new {1} my $o = new Foo;`,
		`my %h = (new => 1);`,
		`print STDERR "x";`,
		`use v5.36; my $o = new Foo;`,
	} {
		got := strings.TrimSpace(parse.Canon(parse.Parse([]byte(src)), []byte(src)))
		for _, node := range allNodes(parse.Parse([]byte(src))) {
			if node.Indirect {
				t.Errorf("Parse(%q) built an indirect call where perl does not (canon %s)", src, got)
			}
		}
	}
}
