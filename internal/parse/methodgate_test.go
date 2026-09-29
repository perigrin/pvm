// ABOUTME: `method` declares only under the class feature; elsewhere it is an
// ABOUTME: ordinary name, and `method Pack (...)` is an indirect method call.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestMethodKeywordNeedsClassFeature: the parser read `method` as the class
// feature's declarator everywhere, so op/method.t's `method Pack ("a")` --
// a sub named method called through indirect object syntax -- became a
// method declaration named Pack. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'sub Pack::method { "m" } my $x = method Pack ("a");'
//	my $x = 'Pack'->method('a');
//
// The qualified declaration is what makes Pack a package the indirect
// reading can name, as op/method.t:54 declares it.
//
// perl.git t/op/method.t:59-82.
func TestMethodKeywordNeedsClassFeature(t *testing.T) {
	src := `sub Pack::method { "m" } my $x = method Pack ("a");`
	root := parse.Parse([]byte(src))
	if containsKind(root, parse.Unknown) {
		t.Fatalf("%q: perl accepts this; got %s", src, shape(root))
	}
	if call := findCall(root, "method"); call == nil || !call.Indirect {
		t.Errorf("%q: want the indirect call Pack->method; got %s", src, shape(root))
	}

	// Under the feature it declares, as class/*.t relies on, and so it does
	// once a `class` has been declared: Object::Pad provides the same two
	// keywords without the feature.
	for _, decl := range []string{
		`use feature 'class'; class P { method m { 1 } }`,
		`class Q { method m { 1 } }`,
	} {
		root = parse.Parse([]byte(decl))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", decl, shape(root))
		}
	}
}
