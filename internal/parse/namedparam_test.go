// ABOUTME: A signature's named parameter, blead's `:$name` with an optional
// ABOUTME: default, parses; a `:` before a scalar anywhere else still refuses.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestNamedParameter: blead's perly.y, `sigscalarelem: optcolon PERLY_DOLLAR
// sigvar [ASSIGNOP optsigscalardefault]` -- a `:` before a scalar parameter
// makes it named, and it may default with `=`, `//=` or `||=` as a positional
// one does (perlsub, "Named parameters", since 5.43.5). No released perl
// has it -- 5.42.0 rejects `sub f (:$x) {}` -- so accepting it changes no
// parse of a program 5.42 accepts. perl.git t/op/signatures.t:927-1005.
func TestNamedParameter(t *testing.T) {
	for _, src := range []string{
		`use feature 'signatures'; sub f (:$alpha, :$beta) { $alpha }`,
		`use feature 'signatures'; sub f (:$alpha = "A", :$beta //= "B", :$g ||= 1) { }`,
		`use feature 'signatures'; sub f ($a, $b, :$x, :$y) { }`,
		`use feature 'signatures'; sub f (:$x, :$y, @rest) { }`,
		`use feature 'signatures'; my $f = sub (:$x) { $x };`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: blead accepts this; got %s", src, shape(root))
			continue
		}
		canon := parse.Canon(root, []byte(src))
		if again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon)); again != canon {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", src, canon, again)
		}
	}
	if root := parse.Parse([]byte(`my $x; my @a = (:$x);`)); !containsKind(root, parse.Unknown) {
		t.Errorf("`(:$x)` outside a signature has no reading; got %s", shape(root))
	}
}
