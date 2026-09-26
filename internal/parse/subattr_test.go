// ABOUTME: A sub attribute sits between the name and the body and belongs to the sub.
// ABOUTME: `sub f :lvalue { 1 }` is one declaration, not a declaration and a leftover block.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestSubAttribute: an attribute list between a sub's name and its body is
// part of the declaration.
//
// Measured at `b2084901`, before this test existed, every attributed spelling
// finished the sub AT the name and left the attribute and the whole body as
// trailing tokens:
//
//	sub f :lvalue { 1 }              Unknown=1  canon "sub f;:lvalue { 1 }"
//	my $s = sub :lvalue { 1 };       Unknown=1  canon "my $s = sub();:lvalue { 1 }"
//	my $s = sub :prototype($) { 1 }; Unknown=1  canon "my $s = sub();:prototype($) { 1 };"
//
// One attribute therefore costs the rest of the statement. parseAttributes
// already existed for `class` and `field`; neither sub path called it.
//
// Every spelling below is valid perl 5.42.0:
//
//	perl -e 'sub f :lvalue { 1 } print f(),"\n"'                    -> 1
//	perl -e 'my $s = sub :lvalue { 1 }; print $s->(),"\n"'          -> 1
//	perl -e 'my $s = sub :prototype($) { 1 }; print $s->(1),"\n"'   -> 1
func TestSubAttribute(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		// A named sub. The attribute is a child of the declaration and the
		// block is the last child, so no semicolon is written after it.
		{"sub f :lvalue { 1 }", "sub f :lvalue {1;}"},

		// Two attributes are a LIST, not a nested anything.
		{"sub f :lvalue :method { 1 }", "sub f :lvalue :method {1;}"},

		// An anonymous sub takes them too. This is the reason the anon path
		// needs its own call: `sub :lvalue {...}` has no name for the
		// declaration branch to find, so it reaches parseTerm.
		{"my $s = sub :lvalue { 1 };", "my $s = sub :lvalue {1;};"},

		// An attribute with a parenthesised argument. `$)` is a real perl
		// punctuation variable, so the argument must be scanned opaquely or
		// the closing paren is swallowed into a Variable token and the sub
		// body goes with it.
		{"my $s = sub :prototype($) { 1 };", "my $s = sub :prototype($) {1;};"},
		{"sub f :prototype($$) { 1 }", "sub f :prototype($$) {1;}"},

		// A prototype AND an attribute, in perl's order: prototype first.
		{"sub f () :lvalue { 1 }", "sub f () :lvalue {1;}"},

		// No attribute: unchanged, and the reason the ratchet must not move
		// for these.
		{"my $s = sub { 1 };", "my $s = sub {1;};"},
		{"sub x () { 8 }", "sub x () {8;}"},

		// THE TERNARY IS THE HAZARD THIS CHANGE CREATED. `inSubAttrs` keeps
		// a block expected across an attribute's colon, and a ternary's
		// colon sits in exactly the same place after an anonymous sub. What
		// separates them is the `}`: a closing bracket clears the carry, so
		// the colon after `sub { 1 }` is a ternary's and the one after a
		// bare `sub` is an attribute's.
		{"my $c = $x ? sub { 1 } : sub { 2 };", "my $c = $x ? sub {1;} : sub {2;};"},

		// A CLASS attribute takes an argument, and canon DROPPED it: the
		// node's End was extended past `(Shape)` and its Text was not, so
		// `class Point :isa(Shape)` re-emitted as `class Point :isa` and
		// lost its parent class. Unknown was 0 throughout, which is why the
		// parse ratchet never saw it.
		{"class Point :isa(Shape) { }", "class Point :isa(Shape) {}"},
	} {
		n := parse.Parse([]byte(tc.src))
		if got := countUnknown(n); got != 0 {
			t.Errorf("Parse(%q): Unknown=%d, want 0", tc.src, got)
		}
		if got := strings.TrimSpace(parse.Canon(n, []byte(tc.src))); got != tc.want {
			t.Errorf("Canon(parse(%q)):\n  got  %q\n  want %q", tc.src, got, tc.want)
		}
		if ok, why := parse.Faithful(n, []byte(tc.src)); !ok {
			t.Errorf("Faithful(%q): %s", tc.src, why)
		}
	}
}
