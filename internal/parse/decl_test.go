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

		// Attributes are not a `field` thing -- `my` and `our` take them
		// too, which is why they live on the shared declarator path rather
		// than in a class-syntax branch. Verified:
		//
		//	perl -e 'use threads::shared; my $x :shared = 1;'   ok
		//	perl -MO=Deparse -e 'our $y :shared;'               our $y;
		"my $x :shared;",
		"my $x :shared = 1;",
		"our $y :shared;",
		"field $f :param;",
		"field $f :param = 1;",
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

// TestFormatDeclaration: `format NAME = <body> .` is a declaration whose
// body is one opaque token.
//
// The lexer already delimits the body -- `scanFormatBody` emits one
// FormatBody token running from the `=` to a line holding a lone `.` -- so
// what is asserted here is above it: that the declaration ENDS at that
// token rather than looking for a `;` it does not have.
//
// `format` with no name is the STDOUT default, which perl accepts.
// Measured on perl 5.42.0:
//
//	$ perl -e 'format =
//	> a line
//	> .
//	> write;'
//	a line
func TestFormatDeclaration(t *testing.T) {
	for _, tc := range []struct {
		src  string
		name string
	}{
		{"format STDOUT =\na fixed report line\n.\n", "STDOUT"},
		{"format =\na fixed report line\n.\n", ""},
		{"format REPORT =\n@<<<<<<< @>>>\n$name,   $qty\n.\n", "REPORT"},
	} {
		root := parse.Parse([]byte(tc.src))
		d := firstOfKind(root, parse.Declaration)
		if d == nil {
			t.Errorf("%q must parse as a Declaration: %v", tc.src, kinds(root))
			continue
		}
		if d.Text != "format" {
			t.Errorf("%q: declarator is %q, want %q", tc.src, d.Text, "format")
		}
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q parsed with an Unknown: %v", tc.src, kinds(root))
			continue
		}
		var name string
		if len(d.Children) > 0 {
			name = d.Children[0].Text
		}
		if name != tc.name {
			t.Errorf("%q: name is %q, want %q", tc.src, name, tc.name)
		}
	}
}

// TestFormatDoesNotSwallowTheNextStatement: the declaration stops at its
// body's closing `.`, so the `write` that uses it is a statement of its own.
//
// This is what the corpus measured while `format` was unimplemented, stated
// directly rather than through an Unknown count: the keyword fell to
// `skipToStatementEnd`, which looks for a `;` a format declaration does not
// have and takes the next statement's instead.
func TestFormatDoesNotSwallowTheNextStatement(t *testing.T) {
	src := []byte("format STDOUT =\na fixed report line\n.\nwrite;\nprint \"x\";\n")
	root := parse.Parse(src)

	if containsKind(root, parse.Unknown) {
		t.Fatalf("parsed with an Unknown: %v", kinds(root))
	}
	d := firstOfKind(root, parse.Declaration)
	if d == nil {
		t.Fatalf("no Declaration: %v", kinds(root))
	}
	want := "format STDOUT =\na fixed report line\n.\n"
	if got := string(src[d.Start:d.End]); got != want {
		t.Errorf("the declaration spans %q, want %q", got, want)
	}
}
