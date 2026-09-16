// ABOUTME: Declarations: my/our/local/state, sub in both body forms, and package in both forms.
// ABOUTME: Prototype recognition is the lexer's; this asserts the declaration holds what it found.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestVariableDeclarations: the four declarators, with single and list
// targets.
//
// `my` is a named unary in perly.y (level 19, §4.6) but it is written here as
// a statement form, because a declaration's target list is not an ordinary
// expression: `my ($a, $b) = @_` declares two variables and assigns to both,
// and the parens are not a call.
func TestVariableDeclarations(t *testing.T) {
	for _, src := range []string{
		"my $x;",
		"my $x = 1;",
		"my ($a, $b) = @_;",
		"my @list = (1, 2);",
		"my %h = (a => 1);",
		"our $VERSION = '1.0';",
		"our ($x, $y);",
		"local $_ = $line;",
		"local ($a, $b) = (1, 2);",
		"state $count = 0;",
	} {
		root := parse.Parse([]byte(src))
		d := firstOfKind(root, parse.Declaration)
		if d == nil {
			t.Errorf("%q must parse as a Declaration: %v", src, kinds(root))
			continue
		}
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q parsed with an Unknown: %v", src, kinds(root))
		}
	}
}

// TestSubDeclarationForms: a body or a bare `;`.
//
// §5.5.2 gives the grammar as `( Block | ";" )`, so the bodiless form is just
// the `;` alternative. §0.13 rank 4 counts 7 corpus files using it.
func TestSubDeclarationForms(t *testing.T) {
	for _, tc := range []struct {
		src      string
		hasBlock bool
	}{
		{"sub f { 1 }", true},
		{"sub f;", false},
		{"sub f ($$) { 1 }", true},
		{"sub f ($$);", false},
		{"sub f () { 1 }", true},
	} {
		root := parse.Parse([]byte(tc.src))
		d := firstOfKind(root, parse.Declaration)
		if d == nil {
			t.Errorf("%q must parse as a Declaration: %v", tc.src, kinds(root))
			continue
		}
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q parsed with an Unknown: %v", tc.src, kinds(root))
			continue
		}
		hasBlock := firstOfKind(d, parse.Block) != nil
		if hasBlock != tc.hasBlock {
			t.Errorf("%q: hasBlock = %v, want %v", tc.src, hasBlock, tc.hasBlock)
		}
	}
}

// TestPackageForms: the statement form and the block form.
//
// `package NAME;` scopes to the end of the enclosing block; `package NAME {}`
// scopes to its own braces. Both are in the T2 corpus.
func TestPackageForms(t *testing.T) {
	for _, tc := range []struct {
		src      string
		hasBlock bool
	}{
		{"package Foo;", false},
		{"package Foo::Bar;", false},
		{"package Foo { 1 }", true},
		{"package Foo 1.0;", false},
	} {
		root := parse.Parse([]byte(tc.src))
		d := firstOfKind(root, parse.Declaration)
		if d == nil {
			t.Errorf("%q must parse as a Declaration: %v", tc.src, kinds(root))
			continue
		}
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q parsed with an Unknown: %v", tc.src, kinds(root))
			continue
		}
		hasBlock := firstOfKind(d, parse.Block) != nil
		if hasBlock != tc.hasBlock {
			t.Errorf("%q: hasBlock = %v, want %v", tc.src, hasBlock, tc.hasBlock)
		}
	}
}

// TestDeclarationCarriesItsPrototype: the lexer found it; the declaration
// holds it.
//
// Resolution -- what the prototype does to a CALL -- is M4's. This is only
// that the recognition survives into the tree, so M4 has something to read.
func TestDeclarationCarriesItsPrototype(t *testing.T) {
	src := []byte("sub f ($$) { 1 }")
	root := parse.Parse(src)

	d := firstOfKind(root, parse.Declaration)
	if d == nil {
		t.Fatalf("no Declaration: %v", kinds(root))
	}
	p := firstOfKind(d, parse.PrototypeNode)
	if p == nil {
		t.Fatalf("the declaration must carry its prototype: %v", kinds(d))
	}
	if got := string(src[p.Start:p.End]); got != "($$)" {
		t.Errorf("prototype spans %q, want %q", got, "($$)")
	}
}
