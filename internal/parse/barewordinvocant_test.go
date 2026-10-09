// ABOUTME: A bareword before `->` is a class-name string unless a sub of that name is declared.
// ABOUTME: `Foo::Bar->new` invokes new on 'Foo::Bar'; `sub Foo {} Foo->new` calls Foo first.

package parse_test

import (
	"os/exec"
	"testing"

	"tamarou.com/pvm/internal/conformance"
	"tamarou.com/pvm/internal/parse"
)

// TestBarewordInvocant: perl reads a bareword before `->` as a string
// invocant when no sub of that name is declared at that point, and as a call
// to the sub when one is. Measured on 5.42.0 with -MO=Deparse:
//
//	Foo::Bar->new; Foo->new(1);         'Foo::Bar'->new; 'Foo'->new(1);
//	Foo->$method; Foo->SUPER::new       'Foo'->$method; 'Foo'->SUPER::new;
//	STDOUT->flush                       'STDOUT'->flush;
//	CORE::GLOBAL::die->new              'CORE::GLOBAL::die'->new;
//	Foo->new; sub Foo {1}               'Foo'->new;   (declared after)
//	sub Foo {1} Foo->new                Foo()->new;
//	sub Foo; Foo->new                   Foo()->new;
//	sub Foo::Bar {1} Foo::Bar->new      Foo::Bar()->new;
//	__PACKAGE__->new                    __PACKAGE__->new;
//	shift->x                            shift(@ARGV)->x;
//
// The tree said `Foo::Bar()->new` -- call the sub Foo::Bar and invoke new on
// its result -- which is a different program wherever no such sub exists.
func TestBarewordInvocant(t *testing.T) {
	perl, perlErr := conformance.PerlPath()
	deparse := func(src string) string {
		out, err := exec.Command(perl, "-MO=Deparse", "-e", src).CombinedOutput()
		if err != nil {
			t.Errorf("perl refuses %q: %v\n%s", src, err, out)
		}
		return string(out)
	}

	for _, tc := range []struct {
		src, word string
		call      bool
	}{
		// A string invocant: no sub of that name is declared.
		{`Foo::Bar->new;`, "Foo::Bar", false},
		{`Foo->new(1);`, "Foo", false},
		{`Foo ->new;`, "Foo", false},
		{`my $method; Foo->$method;`, "Foo", false},
		{`Foo->SUPER::new;`, "Foo", false},
		{`Foo::Bar->new->baz;`, "Foo::Bar", false},
		{`print Foo->x;`, "Foo", false},
		{`foo->new;`, "foo", false},
		{`CORE::GLOBAL::die->new;`, "CORE::GLOBAL::die", false},
		{`STDOUT->flush;`, "STDOUT", false},
		{`STDERR->print(1);`, "STDERR", false},
		{`Foo->new; sub Foo {1}`, "Foo", false},
		{`package Foo { sub new {} } Foo->new;`, "Foo", false},

		// A call: a sub of that name is declared before the arrow.
		{`sub Foo {1} Foo->new;`, "Foo", true},
		{`sub Foo; Foo->new;`, "Foo", true},
		{`sub Foo::Bar {1} Foo::Bar->new;`, "Foo::Bar", true},
		{`package Foo { sub new {} } sub Foo {1} Foo->new;`, "Foo", true},
		{`sub Foo {1} Foo->$m;`, "Foo", true},
		{`Foo()->new;`, "Foo", true},
		{`shift->x;`, "shift", true},

		// `__PACKAGE__` is the niladic builtin naming the current package,
		// never the string `__PACKAGE__`.
		{`__PACKAGE__->new;`, "__PACKAGE__", true},
	} {
		root := parse.Parse([]byte(tc.src))
		if n := countUnknown(root); n != 0 {
			t.Errorf("%s: Unknown=%d\n%s", tc.src, n, dumpTree(root, 2))
			continue
		}
		if got := callsWord(root, tc.word); got != tc.call {
			t.Errorf("%s: %s is a call = %v, perl reads call = %v\n%s",
				tc.src, tc.word, got, tc.call, dumpTree(root, 2))
		}
		if ok, why := canonMatches([]byte(tc.src)); !ok {
			t.Errorf("%s: canon is not a fixpoint: %s", tc.src, why)
		}
		if perlErr != nil {
			continue
		}
		canon := parse.Canon(root, []byte(tc.src))
		if a, b := deparse(tc.src), deparse(canon); a != b {
			t.Errorf("perl reads %q and its canon %q differently:\n%s\n%s", tc.src, canon, a, b)
		}
	}

	if perlErr != nil {
		t.Logf("no perl 5.42 to compare readings with: %v", perlErr)
	}
}

// callsWord reports whether any Call in n is named word.
func callsWord(n *parse.Node, word string) bool {
	if n.Kind == parse.Call && n.Text == word {
		return true
	}
	for _, c := range n.Children {
		if callsWord(c, word) {
			return true
		}
	}
	return false
}
