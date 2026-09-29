// ABOUTME: `ok`, `main::ok` and `::ok` name one sub, so a parenless call resolves whichever
// ABOUTME: spelling declared it; a sub qualified into another package keeps its package.
package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestSubNameSpellings holds issue 01a0df59's measured table. The in-file sub
// table was keyed by the literal spelling, so a parenless call resolved only
// when it spelled the callee exactly as the declaration did. perl strips a
// leading `::` and then a leading `main::`; valid on 5.42.0 in every case:
//
//	$ perl -e 'sub main::ok {print "A\n"} ::ok(1);'
//	A
//
// 15 perl.git t/ files first refuse at a `::is ...` or `main::ok ...` call to
// test.pl's subs, which the require declares under their bare names.
func TestSubNameSpellings(t *testing.T) {
	for _, src := range []string{
		"sub ok {1} ok 1, 2;",
		"sub main::ok {1} main::ok 1, 2;",
		"sub Foo::bar {1} Foo::bar 1, 2;",
		"sub ok {1} main::ok 1, 2;",
		"sub ok {1} ::ok 1, 2;",
		"sub main::ok {1} ok 1, 2;",
		"sub main::ok {1} ::ok 1, 2;",
	} {
		if got := countUnknown(parse.Parse([]byte(src))); got != 0 {
			t.Errorf("Parse(%q) has %d Unknown, want 0", src, got)
		}
	}

	// NEGATIVE: a sub qualified into ANOTHER package is not the bare name.
	// `sub Foo::bar {1} bar 1, 2;` calls an undeclared main::bar, so the
	// parenless list stays unresolved and refuses, as it does today.
	src := "sub Foo::bar {1} bar 1, 2;"
	if got := countUnknown(parse.Parse([]byte(src))); got == 0 {
		t.Errorf("Parse(%q) has no Unknown: Foo::bar must not collapse to bar", src)
	}
}
