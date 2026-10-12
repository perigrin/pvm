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
