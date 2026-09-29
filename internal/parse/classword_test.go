// ABOUTME: A bareword ending in `::` is a class-name string, never a call:
// ABOUTME: `bless [], Foo::` and `Foo::->new` name the package Foo.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestClassNameBareword: toke.c:8072 strips the `::` from `Foo::` and makes
// the rest a constant without looking for a sub, so there is no callee to
// resolve. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'f(Foo::); my $x = Foo::;'
//	f('Foo');
//	my $x = 'Foo';
//
// perl.git t/op/gv.t:984, `bless [], _121242::`.
func TestClassNameBareword(t *testing.T) {
	for _, src := range []string{
		`f(Foo::);`,
		`my $x = Foo::;`,
		`Foo::->new;`,
		`*$$foo = bless [], _121242::;`,
	} {
		root := parse.Parse([]byte(src))
		var walk func(*parse.Node)
		walk = func(n *parse.Node) {
			if n.Kind == parse.Call && strings.HasSuffix(n.Text, "::") {
				t.Errorf("%q: %q is a call; perl reads a string", src, n.Text)
			}
			for _, c := range n.Children {
				walk(c)
			}
		}
		walk(root)
		canon := parse.Canon(root, []byte(src))
		if strings.Contains(canon, "::()") {
			t.Errorf("%q: canon %q calls the class name", src, canon)
		}
		if again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon)); again != canon {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", src, canon, again)
		}
	}
}
