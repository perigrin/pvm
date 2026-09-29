// ABOUTME: `my TYPE VAR` declares a typed lexical: the class between the declarator
// ABOUTME: and the variable belongs to the declaration, not to a call.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestTypedLexicalDeclaration: toke.c's KEY_my branch takes a package name
// after `my`, `our` or `state` as the variable's class
// (`PL_in_my_stash = find_in_my_stash(...)`). The parser read it as a call
// and split the statement -- `my Foo();$f;`. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'package Foo::Bar; package Foo; package main;
//	      use feature "state"; my Foo $f; our Foo $g; state Foo $h;
//	      my Foo::Bar ($i, $j) = (1, 2); my Foo $k = 3;'
//	my Foo $f;
//	our Foo $g;
//	state Foo $h;
//	(my Foo::Bar ($i, $j)) = (1, 2);
//	my Foo $k = 3;
//
// perl.git t/op/multideref.t:240.
func TestTypedLexicalDeclaration(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"my Foo $f;", "my Foo $f;"},
		{"our Foo $g;", "our Foo $g;"},
		{"state Foo $h;", "state Foo $h;"},
		{"my Foo $k = 3;", "my Foo $k = 3;"},
		{"my Foo::Bar ($i, $j) = (1, 2);", "my Foo::Bar ($i, $j) = (1, 2);"},
	} {
		root := parse.Parse([]byte(tc.src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", tc.src, shape(root))
			continue
		}
		if firstOfKind(root, parse.Call) != nil {
			t.Errorf("%q: the class is not a call; got %s", tc.src, shape(root))
		}
		if got := strings.TrimSpace(parse.Canon(root, []byte(tc.src))); got != tc.want {
			t.Errorf("%q: canon %q, want %q", tc.src, got, tc.want)
		}
	}
}
