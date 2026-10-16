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

// An attribute is head, not initialiser: `my $x :shared` gives an
// `attribute ":shared"` second child, kept whole as the parser spells it.
// `my \$x = $y` is a ref alias: its variable is the `ref`'s operand.
func TestLowerDeclAttrsAndRefAlias(t *testing.T) {
	cases := []struct {
		src      string
		attrs    []string
		init     bool
		refAlias bool
	}{
		{"my $x :shared;", []string{":shared"}, false, false},
		{"field $x :param :reader = 1;", []string{":param", ":reader"}, true, false},
		{`my \$x = $y;`, nil, true, true},
		{"my $x = 1;", nil, true, false},
		// A typed lexical's class is head too, not the variable.
		{"my Foo $f;", nil, false, false},
	}
	for _, c := range cases {
		d := lowerDecl(t, c.src)
		if !reflect.DeepEqual(d.Attrs, c.attrs) {
			t.Errorf("%q: Attrs = %q, want %q", c.src, d.Attrs, c.attrs)
		}
		if len(d.Vars) != 1 || (d.Init != nil) != c.init || d.RefAlias != c.refAlias {
			t.Errorf("%q: %d Vars, Init %v, RefAlias %v; want 1 Var, Init %v, RefAlias %v",
				c.src, len(d.Vars), d.Init != nil, d.RefAlias, c.init, c.refAlias)
		}
	}
}

// lowerLast lowers the one expression of src's last statement.
func lowerLast(t *testing.T, src string) lower.Node {
	t.Helper()
	root := parse.Parse([]byte(src))
	last := root.Children[len(root.Children)-1]
	if last.Kind != parse.Statement || len(last.Children) != 1 {
		t.Fatalf("%q: last statement does not hold one expression", src)
	}
	return lower.LowerExpr(root, last.Children[0])
}

// clearIDs zeroes the id of n and of everything a walk from it reaches, so a
// shape compares without its ids.
func clearIDs(n lower.Node) {
	n.Meta().ID = lower.ID{}
	for _, c := range lower.Children(n) {
		clearIDs(c)
	}
}

// field reads one named field of a lowered node through reflect, since a
// type assertion in a test trips go-critic run without type information.
func field(n lower.Node, name string) reflect.Value {
	return reflect.ValueOf(n).Elem().FieldByName(name)
}

// A named sub is perl's compile-time `*f = sub {...}`: a Decl of kind sub
// whose one Var is &f and whose Init is the AnonSub; `sub f;` declares
// without defining, so its Init is nil. method follows sub. package and
// class name a Bareword, with a version or block as Init and class
// attributes in Attrs. A sub without a name is an AnonSub in either
// position.
//
// Every body, and a package or class block, is the Block its statements
// lower to: `{ 1 }` is one ExprStmt, `{ }` none.
func TestLowerSubDeclaration(t *testing.T) {
	one := func() *lower.Block {
		return &lower.Block{Stmts: []lower.Node{&lower.ExprStmt{X: &lower.NumLit{Text: "1", IsInt: true}}}}
	}
	body := func(attrs ...string) *lower.AnonSub {
		return &lower.AnonSub{Body: one(), Attrs: attrs}
	}
	amp := func(name string) []lower.Node { return []lower.Node{&lower.Var{Sigil: "&", Name: name}} }
	word := func(name string) []lower.Node { return []lower.Node{&lower.Bareword{Name: name}} }
	cases := []struct {
		src  string
		want lower.Node
	}{
		{"sub f { 1 }", &lower.Decl{Kind: lower.DeclSub, Vars: amp("f"), Init: body()}},
		{"sub f;", &lower.Decl{Kind: lower.DeclSub, Vars: amp("f")}},
		{"sub f($);", &lower.Decl{Kind: lower.DeclSub, Vars: amp("f")}},
		{"sub { 1 };", body()},
		{"sub f :prototype($) { 1 }", &lower.Decl{Kind: lower.DeclSub, Vars: amp("f"), Init: body(":prototype($)")}},
		{"use v5.38; class Foo; method m { 1 }", &lower.Decl{Kind: lower.DeclMethod, Vars: amp("m"), Init: body()}},
		{"package Foo;", &lower.Decl{Kind: lower.DeclPackage, Vars: word("Foo")}},
		{"package Foo 1.2;", &lower.Decl{Kind: lower.DeclPackage, Vars: word("Foo"), Init: &lower.NumLit{Text: "1.2"}}},
		{"package Foo { 1 }", &lower.Decl{Kind: lower.DeclPackage, Vars: word("Foo"), Init: one()}},
		{"use v5.38; class Foo :isa(Bar) { }", &lower.Decl{Kind: lower.DeclClass, Vars: word("Foo"), Attrs: []string{":isa(Bar)"}, Init: &lower.Block{}}},
	}
	for _, c := range cases {
		got := lowerLast(t, c.src)
		clearIDs(got)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%q: lowered to %#v, want %#v", c.src, got, c.want)
		}
	}

	// A signature's parameters are the sub's Sig: a `list paren` gives one
	// each, a lone `term paren` gives one. Without the signatures feature
	// the parens are a prototype, which is not a signature. The parameters'
	// own shapes are the variables slice's, so only their count is checked.
	sigs := []struct {
		src string
		n   int
	}{
		{"use v5.36; sub f ($x, $y) { 1 }", 2},
		{"use v5.36; sub f ($x, $y = 5, @r) { 1 }", 3},
		{"sub f ($x, $y) { 1 }", 0},
		{"use v5.36; sub ($x) { 1 };", 1},
		{"use v5.36; my $f = sub ($x) { 1 };", 1},
	}
	for _, c := range sigs {
		sub := lowerLast(t, c.src)
		if field(sub, "Kind").IsValid() {
			sub = field(sub, "Init").Interface().(lower.Node)
		}
		if got := field(sub, "Sig").Len(); got != c.n {
			t.Errorf("%q: %d signature parameters, want %d", c.src, got, c.n)
		}
		if b := field(sub, "Body"); b.IsNil() || reflect.TypeOf(b.Interface()) != reflect.TypeOf(&lower.Block{}) {
			t.Errorf("%q: Body is %v, want a Block", c.src, b)
		}
	}

	// In expression position the sub is the declaration's Init.
	init := field(lowerLast(t, "my $f = sub { 1 };"), "Init").Interface().(lower.Node)
	clearIDs(init)
	if !reflect.DeepEqual(init, lower.Node(body())) {
		t.Errorf("my $f = sub { 1 }: Init is %#v, want %#v", init, body())
	}
}

// Input perl rejects (perl -c on 5.42) still lowers: what parsed keeps its
// shape and the Unknown becomes a Hole. `my $x = (;` keeps its variable and
// has a Hole Init. `sub {;` is an AnonSub whose block, which the parser
// closes, holds an Unknown: its Body is a Block holding the one Hole.
// `sub f {` keeps its name, and its unclosed body is an Unknown, so a Hole.
func TestLowerDeclOverBrokenInput(t *testing.T) {
	hole := reflect.TypeOf(&lower.Hole{})

	d := lowerDecl(t, "my $x = (;")
	if d.Kind != lower.DeclMy || len(d.Vars) != 1 || d.Vars[0] == nil || reflect.TypeOf(d.Init) != hole {
		t.Errorf("my $x = (;: lowered to %#v, want a my Decl with one Var and a Hole Init", d)
	}

	s := lowerLast(t, "sub {;")
	clearIDs(s)
	wantSub := &lower.AnonSub{Body: &lower.Block{Stmts: []lower.Node{&lower.Hole{}}}}
	if !reflect.DeepEqual(s, lower.Node(wantSub)) {
		t.Errorf("sub {;: lowered to %#v, want %#v", s, wantSub)
	}

	f := lowerLast(t, "sub f {")
	clearIDs(f)
	want := &lower.Decl{Kind: lower.DeclSub, Vars: []lower.Node{&lower.Var{Sigil: "&", Name: "f"}},
		Init: &lower.AnonSub{Body: &lower.Hole{}}}
	if !reflect.DeepEqual(f, lower.Node(want)) {
		t.Errorf("sub f {: lowered to %#v, want %#v", f, want)
	}
}

// A name is what the lexer scans as one: a leading `::`, an apostrophe
// separator, and, under `use utf8`, any XID letter. perl 5.42 runs
// `sub ::foo { 1 } sub foo'bar { 2 } sub café { 3 }` and prints 123 from
// main::foo(), foo::bar() and café(); `$h{café}` is the key café, and
// `::Foo->new` and `Foo->new->café` call into package Foo. So each named
// sub lowers to its Decl, and the id layer opens an anchor for exactly
// the subs the lowering names. The control: `sub { 1 }` and `sub ($x)
// { 1 }` stay anonymous and open none.
func TestLowerUnicodeAndPackageNames(t *testing.T) {
	for src, name := range map[string]string{
		"sub ::foo { 1 }":          "::foo",
		"sub foo'bar { 2 }":        "foo'bar",
		"use utf8; sub café { 3 }": "café",
		"sub f { 1 }":              "f",
	} {
		root := parse.Parse([]byte(src))
		last := root.Children[len(root.Children)-1].Children[0]
		decl, ok := lowerLast(t, src).(*lower.Decl)
		if !ok || len(decl.Vars) != 1 || !reflect.DeepEqual(zeroIDs(decl.Vars[0]), &lower.Var{Sigil: "&", Name: name}) {
			t.Errorf("%q: lowered to %#v, want the Decl of &%s", src, decl, name)
		}
		if got := lower.AssignIDs(root)[last.Children[1]].Anchor; got != "sub "+name+"#0" {
			t.Errorf("%q: the body is anchored %q, want %q", src, got, "sub "+name+"#0")
		}
	}
	for _, src := range []string{"sub { 1 };", "use v5.36; sub ($x) { 1 };"} {
		if _, ok := lowerLast(t, src).(*lower.AnonSub); !ok {
			t.Errorf("%q: not an AnonSub", src)
		}
		root := parse.Parse([]byte(src))
		for _, n := range everyNode(root) {
			if id := lower.AssignIDs(root)[n]; id.Anchor != "" {
				t.Errorf("%q: %s %q is anchored %q", src, n.Kind, n.Text, id.Anchor)
			}
		}
	}

	for src, want := range map[string]lower.Node{
		"use utf8; $h{café};": &lower.Key{Base: &lower.Var{Sigil: "$", Name: "h"}, Key: &lower.Bareword{Name: "café", Quoted: true}},
		"::Foo->new;":         &lower.MethodCall{Invocant: &lower.Bareword{Name: "::Foo"}, Name: "new"},
		"use utf8; $o->café;": &lower.MethodCall{Invocant: &lower.Var{Sigil: "$", Name: "o"}, Name: "café"},
		"&::foo;":             &lower.Call{Name: "::foo", Ampersand: true, BareAmpersand: true},
	} {
		if got := zeroIDs(lowerLast(t, src)); !reflect.DeepEqual(got, want) {
			t.Errorf("%q: lowered to %#v, want %#v", src, got, want)
		}
	}
}
