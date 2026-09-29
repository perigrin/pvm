// ABOUTME: An anonymous sub takes a prototype or a signature as a named one does,
// ABOUTME: and its body is a block after it, not a subscript on a call.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestAnonymousSubHead: the lexer scanned a prototype only after a sub's
// NAME and parseAnonSub read no signature, so an anonymous sub's parens
// stayed code -- `sub (&) {...}` read as a call to `sub` with a subscript,
// and `sub () { $f++ if 0; 1 }` split at the modifier. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'my $f2 = sub () { $f++ if 0; 1 };
//	      my $h = sub (&) { my $c = shift; };
//	      use feature "signatures"; my $s = sub ($x, $) { $x };'
//	my $f2 = sub () { 1 }
//	my $h = sub (&) {
//	my $s = sub ($x, $) {
//
// perl.git t/op/hash.t:79, t/op/current_sub.t:83 and 85.
func TestAnonymousSubHead(t *testing.T) {
	for _, src := range []string{
		"my $f2 = sub () { $f++ if 0; 1 };",
		"my $h = sub (&) { my $c = shift; };",
		"my $g = sub ($$) { 1 };",
		"use feature 'signatures'; my $s = sub ($x, $) { $x };",
		"use feature 'signatures'; my $t = sub :lvalue ($x) { $x };",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		sub := firstAnonSub(root)
		if sub == nil || len(sub.Children) < 2 || sub.Children[len(sub.Children)-1].Kind != parse.Block {
			t.Errorf("%q: want an anonymous sub with a head and a block body; got %s", src, shape(root))
		}
		canon := parse.Canon(root, []byte(src))
		again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon))
		if canon != again {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", src, canon, again)
		}
	}
	// A method named sub is still a method call, and a sub NAMED method --
	// an ordinary name outside the class feature -- is still a call.
	for _, src := range []string{
		"$o->sub(1);",
		"sub method { 1 } method('foo', 'bar');",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) || firstOfKind(root, parse.PrototypeNode) != nil {
			t.Errorf("%q: the parens are arguments, not a prototype; got %s", src, shape(root))
		}
	}
}

// firstAnonSub returns the first `sub` Declaration.
func firstAnonSub(n *parse.Node) *parse.Node {
	if n.Kind == parse.Declaration && n.Text == "sub" {
		return n
	}
	for _, c := range n.Children {
		if got := firstAnonSub(c); got != nil {
			return got
		}
	}
	return nil
}
