// ABOUTME: A package or class version is part of the declaration's head: canon
// ABOUTME: writes it after the name, and a block after it is the body.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestPackageVersionIsHead: `package NAME VERSION` and `class NAME VERSION`
// were read with the version as a plain child, so canon joined it with the
// initialiser's ` = ` -- `package Foo = 1.0;`, which is not Perl -- and the
// lexer closed the declaration's head at the number, so a block after it
// was not the body. Measured on 5.42.0, all of them compile (Deparse keeps
// the blocks and does not print a version):
//
//	$ perl -MO=Deparse -e 'package Foo 1.0; package Bar 1.0 { 1 } package main;
//	      use feature "class"; no warnings; class T6A 1.23 {}
//	      class T6B 1.23 :isa(T6A) {}'
//	-e syntax OK
//
// perl.git t/class/inherit.t:55 and 110.
func TestPackageVersionIsHead(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"package Foo 1.0;", "package Foo 1.0;"},
		{"package Bar 1.0 { 1 }", "package Bar 1.0 {1;}"},
		{"use feature \"class\"; class T6A 1.23 {}", "use feature \"class\"; class T6A 1.23 {}"},
		{"use feature \"class\"; class T6B 1.23 :isa(T6A) {}", "use feature \"class\"; class T6B 1.23 :isa(T6A) {}"},
	} {
		root := parse.Parse([]byte(tc.src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", tc.src, shape(root))
			continue
		}
		if got := strings.TrimSpace(parse.Canon(root, []byte(tc.src))); got != tc.want {
			t.Errorf("%q: canon %q, want %q", tc.src, got, tc.want)
		}
	}
}
