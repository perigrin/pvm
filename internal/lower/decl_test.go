// ABOUTME: Tests for lowering declarations: Decl kinds, the list form, attributes, ref aliases and sub declarations.
// ABOUTME: Each shape is the parse measured on the fragment, and each context claim is perl 5.42's reading.

package lower_test

import (
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// firstDecl returns the first declaration of src in preorder, and its root.
func firstDecl(t *testing.T, src string) (root, decl *parse.Node) {
	t.Helper()
	root = parse.Parse([]byte(src))
	var find func(n *parse.Node) *parse.Node
	find = func(n *parse.Node) *parse.Node {
		if n.Kind == parse.Declaration {
			return n
		}
		for _, c := range n.Children {
			if d := find(c); d != nil {
				return d
			}
		}
		return nil
	}
	if decl = find(root); decl == nil {
		t.Fatalf("%q: no declaration", src)
	}
	return root, decl
}

// lowerDecl lowers the first declaration of src and returns it as a Decl.
func lowerDecl(t *testing.T, src string) *lower.Decl {
	t.Helper()
	root, n := firstDecl(t, src)
	got := lower.LowerExpr(root, n)
	// Set through reflect: go-critic, run without type information, reads
	// a type assertion here as one between identical types.
	var d *lower.Decl
	v := reflect.ValueOf(got)
	if v.Type() != reflect.TypeOf(d) {
		t.Fatalf("%q: lowered to %T, want *lower.Decl", src, got)
	}
	reflect.ValueOf(&d).Elem().Set(v)
	return d
}

// The list form of a declaration is one child, a `binary "="`; the scalar
// form is two children, the variable then its initialiser. perl: with
// sub f {(1,2,3)}, my ($x) = f() gives 1 and my $y = f() gives 3. The
// parenthesised left side carries Paren as well, so the two signals agree.
//
// A ref alias is the one-child shape too, but its context is the paren's:
// perl with declared_refs reports scalar context for my \$x = f() and list
// for my \($x) = f() and my (\$x) = f().
func TestLowerDeclListForm(t *testing.T) {
	cases := []struct {
		src        string
		listAssign bool
	}{
		{"my $x = f();", false},
		{"my ($x) = f();", true},
		{"my ($x,$y) = f();", true},
		{`my \$x = f();`, false},
		{`my \($x) = f();`, true},
		{`my (\$x) = f();`, true},
	}
	for _, c := range cases {
		d := lowerDecl(t, c.src)
		if d.ListAssign != c.listAssign {
			t.Errorf("%q: ListAssign = %v, want %v", c.src, d.ListAssign, c.listAssign)
		}
		if len(d.Vars) != 1 || d.Init == nil {
			t.Fatalf("%q: want one Vars entry and an Init, got %d and %v", c.src, len(d.Vars), d.Init)
		}
		// The Paren bit on the lowered left side agrees with ListAssign
		// wherever the declaration is not a ref alias.
		if !d.RefAlias && d.Vars[0].Meta().Paren != c.listAssign {
			t.Errorf("%q: Vars[0].Paren = %v, want %v", c.src, d.Vars[0].Meta().Paren, c.listAssign)
		}
	}
}

// Every declarator of the T1 census lowers to its own Decl.Kind: the
// lowering carries which declarator was written, and typing it is
// inference's job. Each case lowers the declaration of its last statement.
func TestLowerDeclKinds(t *testing.T) {
	cases := []struct {
		src  string
		kind lower.DeclKind
	}{
		{"my $x;", lower.DeclMy},
		{"sub f;", lower.DeclSub},
		{"package Foo;", lower.DeclPackage},
		{"local $x;", lower.DeclLocal},
		{"our $x;", lower.DeclOur},
		{"use v5.36; state $x;", lower.DeclState},
		{"use v5.38; class Foo; field $x;", lower.DeclField},
		{"use v5.38; class Foo;", lower.DeclClass},
		{"use v5.38; class Foo; method m { 1 }", lower.DeclMethod},
	}
	for _, c := range cases {
		root := parse.Parse([]byte(c.src))
		last := root.Children[len(root.Children)-1]
		if last.Kind != parse.Statement || len(last.Children) != 1 || last.Children[0].Kind != parse.Declaration {
			t.Fatalf("%q: last statement is not one declaration", c.src)
		}
		got := lower.LowerExpr(root, last.Children[0])
		v := reflect.ValueOf(got).Elem()
		if v.Type() != reflect.TypeOf(lower.Decl{}) || v.FieldByName("Kind").Interface() != any(c.kind) {
			t.Errorf("%q: lowered to %#v, want a Decl of kind %v", c.src, got, c.kind)
		}
	}
}
