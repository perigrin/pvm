// ABOUTME: Tests for lowering calls (§4.14.2): Call from five call sites and MethodCall from four arrow shapes.
// ABOUTME: Each flag is read off the parse node -- Handle, Resolved, Indirect, the span -- and each perl claim is 5.42's.

package lower_test

import (
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// lowered is the type of checkShapes' lx: it lowers any node of the tree.
type lowered = func(*parse.Node) lower.Node

// kindOf is the kind of node a Node-typed field holds, nil for none.
func kindOf(f reflect.Value) reflect.Type {
	if f.IsNil() {
		return nil
	}
	return f.Elem().Type()
}

// The filehandle slot of print, printf and say is not argument 1: the
// parser marks it with the Handle flag, and the lowering reads the flag
// rather than counting children. perl 5.42's -MO=Deparse keeps the slot
// apart, `print $fh 'x'`, and reads `print $fh, "x"` as printing $fh and
// "x" to STDOUT. A block slot of one expression is that expression.
func TestLowerCallHandleFromFlag(t *testing.T) {
	type L = []lower.Node
	checkShapes(t, []shapeCase{
		{`print $fh "x";`, func(e *parse.Node, lx lowered) lower.Node {
			return &lower.Call{Name: "print", Handle: lx(e.Children[0]), Args: L{lx(e.Children[1])}, Resolved: true}
		}},
		{`print STDERR "x";`, func(e *parse.Node, lx lowered) lower.Node {
			return &lower.Call{Name: "print", Handle: lx(e.Children[0]), Args: L{lx(e.Children[1])}, Resolved: true}
		}},
		{`say $fh "x";`, func(e *parse.Node, lx lowered) lower.Node {
			return &lower.Call{Name: "say", Handle: lx(e.Children[0]), Args: L{lx(e.Children[1])}, Resolved: true}
		}},
		{`print $fh, "x";`, func(e *parse.Node, lx lowered) lower.Node {
			comma := e.Children[0]
			return &lower.Call{Name: "print", Args: L{lx(comma.Children[0]), lx(comma.Children[1])}, Resolved: true}
		}},
		{`print {$fh} "x";`, func(e *parse.Node, lx lowered) lower.Node {
			fh := e.Children[0].Children[0].Children[0]
			return &lower.Call{Name: "print", Handle: lx(fh), Args: L{lx(e.Children[1])}, Resolved: true}
		}},
		// A block of more than one statement is the lowered block.
		{`print {$fh; $g} "x";`, func(e *parse.Node, lx lowered) lower.Node {
			return &lower.Call{Name: "print", Handle: lx(e.Children[0]), Args: L{lx(e.Children[1])}, Resolved: true}
		}},
	})

	// The handle is the variable, not a Hole standing for the block.
	root := parse.Parse([]byte(`print {$fh} "x";`))
	got := lower.LowerExpr(root, expression(t, root))
	if !sameKind(got, &lower.Call{}) || kindOf(field(got, "Handle")) != reflect.TypeOf(&lower.Var{}) {
		t.Errorf(`print {$fh} "x": lowered to %#v, want a Call whose Handle is the Var $fh`, got)
	}
}

// Parenthesized is read off the call's span (§4.8.1): `f(1,2)` has its
// arguments in parentheses and `join ",", @a` does not. `map $_+1, @a`,
// the spec's own unparenthesised example, lowers to MapGrepSort, which has
// no such field.
//
// §4.8.3: `&foo` bypasses prototypes, and with no parentheses passes the
// caller's @_. perl 5.42: with `sub foo { scalar @_ }`, `sub g { &foo }`
// and `sub h { &foo() }`, g(1,2,3) is 3 and h(1,2,3) is 0.
func TestLowerCallFlags(t *testing.T) {
	type L = []lower.Node
	checkShapes(t, []shapeCase{
		{"f(1,2);", func(e *parse.Node, lx lowered) lower.Node {
			comma := e.Children[0]
			return &lower.Call{Name: "f", Args: L{lx(comma.Children[0]), lx(comma.Children[1])}, Parenthesized: true}
		}},
		{"f (1);", func(e *parse.Node, lx lowered) lower.Node {
			return &lower.Call{Name: "f", Args: L{lx(e.Children[0])}, Parenthesized: true}
		}},
		{"f();", func(e *parse.Node, lx lowered) lower.Node {
			return &lower.Call{Name: "f", Parenthesized: true}
		}},
		{"f;", func(e *parse.Node, lx lowered) lower.Node {
			return &lower.Call{Name: "f"}
		}},
		{`join ",", @a;`, func(e *parse.Node, lx lowered) lower.Node {
			comma := e.Children[0]
			return &lower.Call{Name: "join", Args: L{lx(comma.Children[0]), lx(comma.Children[1])}, Resolved: true}
		}},
		{"&foo;", func(e *parse.Node, lx lowered) lower.Node {
			return &lower.Call{Name: "foo", Ampersand: true, BareAmpersand: true}
		}},
		{"&foo();", func(e *parse.Node, lx lowered) lower.Node {
			return &lower.Call{Name: "foo", Parenthesized: true, Ampersand: true}
		}},
		{"&foo(1, 2);", func(e *parse.Node, lx lowered) lower.Node {
			comma := e.Children[1]
			return &lower.Call{Name: "foo", Args: L{lx(comma.Children[0]), lx(comma.Children[1])}, Parenthesized: true, Ampersand: true}
		}},
	})

	if got := lowerOne(t, "map $_+1, @a;"); !sameKind(got, &lower.MapGrepSort{}) {
		t.Errorf("map $_+1, @a: lowered to %#v, want a MapGrepSort", got)
	}
}

// `&foo` names the sub without calling it where it is the operand of `\`,
// defined, exists, undef or goto: perl 5.42's -MO=Concise compiles each of
// `\&foo`, `\(&foo)`, `defined &foo`, `exists &foo`, `undef &foo` and
// `goto &foo` with no entersub, where `&foo;` and `&foo + 1` have one.
// There it lowers to the Var &foo, as a named sub's own Var is.
func TestLowerCodeNameIsNotACall(t *testing.T) {
	foo := &lower.Var{Sigil: "&", Name: "foo"}
	cases := []struct {
		src  string
		want lower.Node
	}{
		{`\&foo;`, &lower.Ref{Operand: foo}},
		{`\(&foo);`, &lower.Ref{Operand: withParen(&lower.Var{Sigil: "&", Name: "foo"})}},
		{"defined &foo;", &lower.Call{Name: "defined", Args: []lower.Node{foo}, Resolved: true}},
		{"defined(&foo);", &lower.Call{Name: "defined", Args: []lower.Node{foo}, Parenthesized: true, Resolved: true}},
		{"exists &foo;", &lower.Call{Name: "exists", Args: []lower.Node{foo}, Resolved: true}},
		{"undef &foo;", &lower.Call{Name: "undef", Args: []lower.Node{foo}, Resolved: true}},
		{"goto &foo;", &lower.LoopEx{Op: "goto", Label: foo}},
		{"&foo + 1;", &lower.Binary{Op: "+", L: &lower.Call{Name: "foo", Ampersand: true, BareAmpersand: true}, R: &lower.NumLit{Text: "1", IsInt: true}}},
	}
	for _, c := range cases {
		sameShape(t, c.src, zeroIDs(lowerOne(t, c.src)), c.want)
	}
}

// A call through a code reference has Fn, the reference, and no Name
// (§4.8.3). `$code->(...)` is a `binary "->"` over the parenthesised
// arguments; `&{$code}(...)` and `&$code(...)` are an `index "("` over the
// ampersand, which bypasses prototypes as `&foo(...)` does. `&$code` is one
// term whose inner expression the parser never built, so its Fn is a Hole
// carrying the term's id.
func TestLowerCallThroughReference(t *testing.T) {
	type L = []lower.Node
	checkShapes(t, []shapeCase{
		{"$code->(1);", func(e *parse.Node, lx lowered) lower.Node {
			return &lower.Call{Fn: lx(e.Children[0]), Args: L{lx(e.Children[1])}, Parenthesized: true}
		}},
		{"$code->(1, 2);", func(e *parse.Node, lx lowered) lower.Node {
			list := e.Children[1]
			return &lower.Call{Fn: lx(e.Children[0]), Args: L{lx(list.Children[0]), lx(list.Children[1])}, Parenthesized: true}
		}},
		{"$code->();", func(e *parse.Node, lx lowered) lower.Node {
			return &lower.Call{Fn: lx(e.Children[0]), Parenthesized: true}
		}},
		{"&{$code}(1);", func(e *parse.Node, lx lowered) lower.Node {
			return &lower.Call{Fn: lx(e.Children[0].Children[1]), Args: L{lx(e.Children[1])}, Parenthesized: true, Ampersand: true}
		}},
		{"&$code(1);", func(e *parse.Node, lx lowered) lower.Node {
			fn := &lower.Hole{}
			fn.Meta().ID = lx(e.Children[0]).Meta().ID
			return &lower.Call{Fn: fn, Args: L{lx(e.Children[1])}, Parenthesized: true, Ampersand: true}
		}},
	})
}

// Four method-call shapes lower to one MethodCall. The method is a `call`
// when it has parentheses and a word when it has none, so HasParens is the
// right side's kind and span; a dynamic method, `$obj->$name`, is Dynamic
// in place of Name. Indirect object syntax keeps the parser's Indirect flag:
// with `package Foo` seen, perl 5.42's -MO=Deparse gives `new Foo(1)` as
// `'Foo'->new(1)`.
func TestLowerMethodCall(t *testing.T) {
	type L = []lower.Node
	checkShapes(t, []shapeCase{
		{"$obj->method(1);", func(e *parse.Node, lx lowered) lower.Node {
			m := e.Children[1]
			return &lower.MethodCall{Invocant: lx(e.Children[0]), Name: "method", Args: L{lx(m.Children[0])}, HasParens: true}
		}},
		{"$obj->method(1, 2);", func(e *parse.Node, lx lowered) lower.Node {
			comma := e.Children[1].Children[0]
			return &lower.MethodCall{Invocant: lx(e.Children[0]), Name: "method", Args: L{lx(comma.Children[0]), lx(comma.Children[1])}, HasParens: true}
		}},
		{"$obj->method;", func(e *parse.Node, lx lowered) lower.Node {
			return &lower.MethodCall{Invocant: lx(e.Children[0]), Name: "method"}
		}},
		{"Foo->new;", func(e *parse.Node, lx lowered) lower.Node {
			return &lower.MethodCall{Invocant: lx(e.Children[0]), Name: "new"}
		}},
		{"Foo->new();", func(e *parse.Node, lx lowered) lower.Node {
			return &lower.MethodCall{Invocant: lx(e.Children[0]), Name: "new", HasParens: true}
		}},
		// The arguments flatten as a list does, the word left of => quoted.
		{"Foo->new(a => 1);", func(e *parse.Node, lx lowered) lower.Node {
			fat := e.Children[1].Children[0]
			a := &lower.Bareword{Name: "a", Quoted: true}
			a.Meta().ID = lx(fat.Children[0]).Meta().ID
			return &lower.MethodCall{Invocant: lx(e.Children[0]), Name: "new", Args: L{a, lx(fat.Children[1])}, HasParens: true}
		}},
		{"$obj->$name(1);", func(e *parse.Node, lx lowered) lower.Node {
			idx := e.Children[1]
			return &lower.MethodCall{Invocant: lx(e.Children[0]), Dynamic: lx(idx.Children[0]), Args: L{lx(idx.Children[1])}, HasParens: true}
		}},
		{"$obj->$name;", func(e *parse.Node, lx lowered) lower.Node {
			return &lower.MethodCall{Invocant: lx(e.Children[0]), Dynamic: lx(e.Children[1])}
		}},
	})

	src := "package Foo; sub new {1} package main; new Foo(1);"
	root := parse.Parse([]byte(src))
	last := root.Children[len(root.Children)-1]
	if last.Kind != parse.Statement || len(last.Children) != 1 || !last.Children[0].Indirect {
		t.Fatalf("%q: the last statement is not one indirect call", src)
	}
	e := last.Children[0]
	lx := func(n *parse.Node) lower.Node { return lower.LowerExpr(root, n) }
	want := &lower.MethodCall{Invocant: lx(e.Children[0]), Name: "new", Args: []lower.Node{lx(e.Children[1])}, HasParens: true, Indirect: true}
	want.Meta().ID = lower.AssignIDs(root)[e]
	if got := lx(e); !reflect.DeepEqual(got, want) {
		t.Errorf("%q: lowered to %#v, want %#v", src, got, want)
	}
}

// perl 5.42 rejects both fragments (`perl -c`: syntax error near "(;"), so
// the parser cannot grow to accept them. The argument it refuses is an
// Unknown, and the call or method call around it lowers with its name and
// invocant intact and a Hole for that argument.
func TestLowerCallOverBrokenInput(t *testing.T) {
	cases := []struct {
		src  string
		want lower.Node
	}{
		{"foo(;", &lower.Call{Name: "foo", Args: []lower.Node{&lower.Hole{}}}},
		{"Foo->new(;", &lower.MethodCall{Invocant: &lower.Bareword{Name: "Foo"}, Name: "new", Args: []lower.Node{&lower.Hole{}}}},
	}
	for _, c := range cases {
		sameShape(t, c.src, zeroIDs(lowerOne(t, c.src)), c.want)
	}
}

// A builtin that takes no operand, called with empty parentheses, is the
// same call as without them: perl 5.42 -MO=Concise compiles `time()` to
// time and `wantarray()` to wantarray, and `return wantarray() ? "l" : "s"`
// gives "l" in list context. The parser gives each an `index "("` over the
// childless call. The control: `time(1)`, which perl refuses ("syntax
// error near "(1""), is not that call.
func TestNiladicBuiltinWithParens(t *testing.T) {
	cases := []struct {
		src  string
		want lower.Node
	}{
		{"time();", &lower.Call{Name: "time", Resolved: true, Parenthesized: true}},
		{"wantarray();", &lower.Wantarray{}},
		{"wantarray ();", &lower.Wantarray{}},
		{"time;", &lower.Call{Name: "time", Resolved: true}},
	}
	for _, c := range cases {
		sameShape(t, c.src, zeroIDs(lowerOne(t, c.src)), c.want)
	}
	if got := zeroIDs(lowerOne(t, "time(1);")); reflect.DeepEqual(got, &lower.Call{Name: "time", Resolved: true, Parenthesized: true}) {
		t.Errorf("time(1): lowered to the niladic call time(), which perl refuses")
	}
}
