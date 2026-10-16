// ABOUTME: Tests for lowering variables and subscripts: Var, Deref, Index and Key, with Arrow and deref-ness.
// ABOUTME: Each shape is the parse measured on the fragment, and each claim about what it reads is perl 5.42's.

package lower_test

import (
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// lowerOne parses a one-expression program and lowers its expression.
func lowerOne(t *testing.T, src string) lower.Node {
	t.Helper()
	root := parse.Parse([]byte(src))
	return lower.LowerExpr(root, expression(t, root))
}

// zeroIDs clears every id under n, so a nested tree compares by shape.
func zeroIDs(n lower.Node) lower.Node {
	n.Meta().ID = lower.ID{}
	for _, c := range lower.Children(n) {
		zeroIDs(c)
	}
	return n
}

func TestLowerVarAndSubscript(t *testing.T) {
	v := func(sigil, name string) *lower.Var { return &lower.Var{Sigil: sigil, Name: name} }
	cases := []struct {
		src  string
		want lower.Node
	}{
		// The sigil is the leading bytes, $# included, and the name is
		// the rest (§4.14.1 rule 7).
		{"$x;", v("$", "x")},
		{"@a;", v("@", "a")},
		{"%h;", v("%", "h")},
		{"*foo;", v("*", "foo")},
		{"$#a;", v("$#", "a")},
		// perl: with $main::x = "pkg" and $Foo::x = "foo", $::x is pkg
		// and $Foo::x is foo. The name keeps its spelling.
		{"$Foo::x;", v("$", "Foo::x")},
		{"$::x;", v("$", "::x")},
		// $_ and @_ differ in the sigil alone.
		{"$_;", v("$", "_")},
		{"@_;", v("@", "_")},
		// perl -MO=Deparse prints ${x} and ${ x } as $x, and $#{x} as $#x.
		{"${x};", v("$", "x")},
		{"${ x };", v("$", "x")},
		{"$#{x};", v("$#", "x")},
		{"${^WARNING_BITS};", v("$", "^WARNING_BITS")},
		{"$$;", v("$", "$")},

		// perl: with %h = (k => "hash") and $h = {k => "ref"}, $h{k} is
		// hash and $h->{k} is ref. The two Keys differ in Arrow alone.
		{"$h{k};", &lower.Key{Base: v("$", "h"), Key: &lower.Bareword{Name: "k", Quoted: true}}},
		{"$h->{k};", &lower.Key{Base: v("$", "h"), Key: &lower.Bareword{Name: "k", Quoted: true}, Arrow: true}},
		{"$a[0];", &lower.Index{Base: v("$", "a"), Idx: &lower.NumLit{Text: "0", IsInt: true}}},
		{"$x->[$i];", &lower.Index{Base: v("$", "x"), Idx: v("$", "i"), Arrow: true}},
		// Chains nest, and only a link spelled with -> carries Arrow.
		{"$h->{a}[0]{b};", &lower.Key{
			Base: &lower.Index{
				Base: &lower.Key{Base: v("$", "h"), Key: &lower.Bareword{Name: "a", Quoted: true}, Arrow: true},
				Idx:  &lower.NumLit{Text: "0", IsInt: true},
			},
			Key: &lower.Bareword{Name: "b", Quoted: true},
		}},
	}
	for _, c := range cases {
		got := lowerOne(t, c.src)
		sameShape(t, c.src, zeroIDs(got), c.want)
	}

	// The ids are AssignIDs', down the chain.
	root := parse.Parse([]byte("$h->{a}[0];"))
	n := expression(t, root)
	ids := lower.AssignIDs(root)
	got := lower.LowerExpr(root, n)
	if len(lower.Children(got)) != 2 {
		t.Fatalf("$h->{a}[0]: lowered to %#v, want an Index over a Key", got)
	}
	base := lower.Children(got)[0]
	if got.Meta().ID != ids[n] || base.Meta().ID != ids[n.Children[0]] {
		t.Errorf("ids %v, %v are not the index nodes' %v, %v", got.Meta().ID, base.Meta().ID, ids[n], ids[n.Children[0]])
	}
}

// A slice is an index over an @ or % base, or over a list: its own kind,
// a Slice, not an Index or a Key. &{$x}, a code call, is a Deref
// (TestLowerAmpBlockDeref), and *foo{CODE}, a glob slot, is neither, and
// this lowering leaves it a Hole.
func TestLowerSubscriptDeclinesSlices(t *testing.T) {
	for _, src := range []string{"@a[0];", "@h{a};", "%a[0];", "%h{a};", "(1,2)[0];", "($x)[0];", "@$x[0];", "@{$x}{a};"} {
		if got := lowerOne(t, src); reflect.TypeOf(got) != reflect.TypeOf(&lower.Slice{}) {
			t.Errorf("%q: lowered to %#v, want a Slice", src, got)
		}
	}
	for _, src := range []string{"*foo{CODE};"} {
		got := lowerOne(t, src)
		sameShape(t, src, got, &lower.Hole{})
	}
}

// Deref-ness is whether a subscript reads through a reference: Arrow, or a
// base that is itself a dereference (§4.14.1 rule 2). perl 5.42, with
// $x = [7], %h = (a => [7]) and $h = {k => 7}: $$x[0], ${$x}[0] and
// $h->{k} are 7 through a reference, $h{k} reads %h, and $h{a}[0] is 7
// because the arrow between two subscripts is implied.
func TestLowerDerefness(t *testing.T) {
	for src, want := range map[string]bool{
		"$h->{k};":    true,
		"$$x[0];":     true,
		"${$x}[0];":   true,
		"$h{a}[0];":   true,
		"$x->[0][1];": true,
		"$h{k};":      false,
		"$a[0];":      false,
	} {
		if got := lower.IsDeref(lowerOne(t, src)); got != want {
			t.Errorf("%q: IsDeref is %v, want %v", src, got, want)
		}
	}
	// Anything not a subscript is not one.
	if lower.IsDeref(lowerOne(t, "$x;")) {
		t.Errorf("$x: IsDeref is true")
	}
}

// Prefix, block and postfix spellings of one dereference lower to one Deref
// (§4.4.6); Postfix records the ->@* spelling. perl 5.42, with $x =
// [1,2,3]: $x->$#*, $#{$x} and $#$x are all 2.
func TestLowerDeref(t *testing.T) {
	x := &lower.Var{Sigil: "$", Name: "x"}
	cases := []struct {
		src  string
		want lower.Node
	}{
		{"${$x};", &lower.Deref{Sigil: "$", Expr: x}},
		{"$$x;", &lower.Deref{Sigil: "$", Expr: x}},
		{"@$x;", &lower.Deref{Sigil: "@", Expr: x}},
		{"@{$x};", &lower.Deref{Sigil: "@", Expr: x}},
		{"%$x;", &lower.Deref{Sigil: "%", Expr: x}},
		// The glob deref's one child is its Expr (§4.14.2). perl: with
		// $g = \*foo and $foo = "g", ${*$g} and ${*{"foo"}} are both g.
		{"*$x;", &lower.Deref{Sigil: "*", Expr: x}},
		{`*{"foo"};`, &lower.Deref{Sigil: "*", Expr: &lower.StrLit{Value: "foo", Interpolated: true}}},

		{"$x->@*;", &lower.Deref{Sigil: "@", Expr: x, Postfix: true}},
		{"$x->%*;", &lower.Deref{Sigil: "%", Expr: x, Postfix: true}},
		{"$x->$#*;", &lower.Deref{Sigil: "$#", Expr: x, Postfix: true}},
		{"$x->$*;", &lower.Deref{Sigil: "$", Expr: x, Postfix: true}},
		{"$x->&*;", &lower.Deref{Sigil: "&", Expr: x, Postfix: true}},
		{"$x->**;", &lower.Deref{Sigil: "*", Expr: x, Postfix: true}},

		// $#{$x}, $#$x and &$x are each one childless term: the inner
		// expression was never parsed, so it is a Hole, not re-lexed.
		{"$#{$x};", &lower.Deref{Sigil: "$#", Expr: &lower.Hole{}}},
		{"$#$x;", &lower.Deref{Sigil: "$#", Expr: &lower.Hole{}}},
		{"&$x;", &lower.Deref{Sigil: "&", Expr: &lower.Hole{}}},
	}
	for _, c := range cases {
		got := lowerOne(t, c.src)
		sameShape(t, c.src, zeroIDs(got), c.want)
	}

	// Any other -> is not a dereference this lowering claims.
	if got := lowerOne(t, "$x->foo;"); reflect.TypeOf(got) == reflect.TypeOf(&lower.Deref{}) {
		t.Errorf("$x->foo: lowered to a Deref")
	}
}

// perl 5.42 rejects both fragments (`perl -c`: syntax error). The parser
// keeps the subscript as an Unknown inside an intact index, so the lowering
// keeps the Index or Key, its base and its Arrow, and holds a Hole where the
// subscript would be.
func TestLowerSubscriptOverBrokenInput(t *testing.T) {
	cases := []struct {
		src  string
		want lower.Node
	}{
		{"$x->[;", &lower.Index{Base: &lower.Var{Sigil: "$", Name: "x"}, Idx: &lower.Hole{}, Arrow: true}},
		{"$h{;", &lower.Key{Base: &lower.Var{Sigil: "$", Name: "h"}, Key: &lower.Hole{}}},
	}
	for _, c := range cases {
		root := parse.Parse([]byte(c.src))
		var n *parse.Node
		for _, s := range root.Children {
			if s.Kind == parse.Statement && len(s.Children) == 1 && s.Children[0].Kind == parse.Index {
				n = s.Children[0]
			}
		}
		if n == nil {
			t.Fatalf("%q: no statement holds an index", c.src)
		}
		sameShape(t, c.src, zeroIDs(lower.LowerExpr(root, n)), c.want)
	}
}

// `&{$c}` and `&$c` are one operation, a call through the code reference
// that passes the caller's @_: perl 5.42, with `sub f { print scalar @_ }`
// and `sub g { my $c = \&f; &{$c}; &$c; &{$c}() }`, g(1,2) prints 2, 2, 0,
// and -MO=Concise compiles both bare forms to entersub. So both lower to
// Deref{&}, the braced one with its operand built. The control:
// `&{$c}(1)` stays a Call with its arguments.
func TestLowerAmpBlockDeref(t *testing.T) {
	type L = []lower.Node
	checkShapes(t, []shapeCase{
		{"&{$c};", func(e *parse.Node, lx lowered) lower.Node {
			return &lower.Deref{Sigil: "&", Expr: lx(e.Children[1])}
		}},
		{"&{$c}(1);", func(e *parse.Node, lx lowered) lower.Node {
			return &lower.Call{Fn: lx(e.Children[0].Children[1]), Args: L{lx(e.Children[1])}, Parenthesized: true, Ampersand: true}
		}},
	})
	got := zeroIDs(lowerOne(t, "&$c;"))
	if d, ok := got.(*lower.Deref); !ok || d.Sigil != "&" {
		t.Errorf("&$c: lowered to %#v, want a Deref{&}", got)
	}
}

// Two small shapes. A `&` followed by one punctuation character calls the
// sub of that name: perl 5.42 -MO=Concise compiles `my $x = &=;` to an
// entersub of *=, and with subs named +, * and % installed, `&+, &*, &%`
// calls all three. A trailing `::` ends a class name and is not part of
// it: -MO=Deparse gives `Foo::->new` as `'Foo'->new` and `my $s = Foo::`
// as `my $s = 'Foo'`. The controls: `&$c` stays a Deref, and `::Foo`
// keeps its leading separator, as Deparse's `'::Foo'->new` does.
func TestLowerMinorShapes(t *testing.T) {
	for src, want := range map[string]lower.Node{
		"&=;":         &lower.Call{Name: "=", Ampersand: true, BareAmpersand: true},
		"&+;":         &lower.Call{Name: "+", Ampersand: true, BareAmpersand: true},
		"&%;":         &lower.Call{Name: "%", Ampersand: true, BareAmpersand: true},
		"Foo::->new;": &lower.MethodCall{Invocant: &lower.Bareword{Name: "Foo"}, Name: "new"},
		"Foo::Bar::;": &lower.Bareword{Name: "Foo::Bar"},
		"::Foo->new;": &lower.MethodCall{Invocant: &lower.Bareword{Name: "::Foo"}, Name: "new"},
		"&$c;":        &lower.Deref{Sigil: "&", Expr: &lower.Hole{}},
	} {
		if got := zeroIDs(lowerOne(t, src)); !reflect.DeepEqual(got, want) {
			t.Errorf("%q: lowered to %#v, want %#v", src, got, want)
		}
	}
}
