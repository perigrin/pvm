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

// TestMethodBodyIsOneStatement: `method m { 1 }` is one statement holding a
// declaration whose block is the body, as a named sub's is. Measured on
// 5.42.0:
//
//	$ perl -MO=Deparse -e 'use feature "class"; class P { method m { 1 } }'
//	    method m {
//	        1;
//	    }
func TestMethodBodyIsOneStatement(t *testing.T) {
	src := `use feature 'class'; class P { method m { 1 } }`
	root := parse.Parse([]byte(src))
	if containsKind(root, parse.Unknown) {
		t.Fatalf("%q: perl accepts this; got %s", src, shape(root))
	}
	var body *parse.Node
	for _, d := range findAll(root, parse.Declaration) {
		if d.Text == "class" {
			body = find(d, parse.Block)
		}
	}
	if body == nil {
		t.Fatalf("%q: no class body:\n%s", src, dumpTree(root, 0))
	}
	want := `block
  statement
    declaration method
      term m
      block
        statement
          term 1
`
	if got := structure(body); got != want {
		t.Errorf("%q: want the class body to hold one method statement:\ngot:\n%swant:\n%s",
			src, got, want)
	}
}

// TestSubKnownOnlyInItsPackage: a sub declared in a package is a known sub
// there and not elsewhere. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'package Pack; sub method { "m" }
//	                       package main; my $x = method Pack ("a");'
//	my $x = 'Pack'->method('a');
//
//	$ perl -MO=Deparse -e 'sub method { "m" } my $x = method Pack ("a");'
//	my $x = main::method(Pack('a'));
//
// That pair is issue 01a0ee05's, but it does not tell a package-blind sub
// table from a scoped one: `package Pack;` makes Pack a known package, and
// perl reads `method Pack (...)` as the indirect call for that alone --
// measured, `package Pack; package main; sub method { "m" } method Pack ("a")`
// is `'Pack'->method('a')` too.
//
// The cases that do tell them apart declare the sub in a package that is
// NOT the invocant, so only the sub table decides. perl reads each from main
// as an indirect call:
//
//	$ perl -MO=Deparse -e 'package Foo; sub method { "m" }
//	                       package main; my $x = method Pack ("a");'
//	my $x = 'Pack'->method('a');
//	$ perl -MO=Deparse -e 'package Foo; sub method { "m" }
//	                       package main; my $x = method $obj;'
//	my $x = $obj->method;
//	$ perl -MO=Deparse -e 'package Foo; sub f { "m" } package main; f Bar;'
//	'Bar'->f;
//
// A package-blind table made each the direct call `method(Pack("a"))`,
// `method($obj)` and `f(Bar())` at Unknown=0. The parser refuses them now
// rather than reading the indirect call, so the assertion is the claim
// itself: the tree never calls Foo's sub from main.
func TestSubKnownOnlyInItsPackage(t *testing.T) {
	for _, tc := range []struct {
		src      string
		indirect bool
	}{
		{`package Pack; sub method { "m" } package main; my $x = method Pack ("a");`, true},
		{`sub method { "m" } my $x = method Pack ("a");`, false},
	} {
		root := parse.Parse([]byte(tc.src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", tc.src, shape(root))
			continue
		}
		call := findCall(root, "method")
		if call == nil || call.Indirect != tc.indirect {
			t.Errorf("%q: want a call to method with Indirect=%v; got %s",
				tc.src, tc.indirect, shape(root))
		}
	}

	for _, tc := range []struct{ src, name string }{
		{`package Foo; sub method { "m" } package main; my $x = method Pack ("a");`, "method"},
		{`package Foo; sub method { "m" } package main; my $x = method $obj;`, "method"},
		{`package Foo; sub f { "m" } package main; f Bar;`, "f"},
	} {
		root := parse.Parse([]byte(tc.src))
		if call := findCall(root, tc.name); call != nil && !call.Indirect {
			t.Errorf("%q: %s is Foo's, not main's, and perl calls it as a "+
				"method; got the direct call %s", tc.src, tc.name, shape(root))
		}
	}
}
