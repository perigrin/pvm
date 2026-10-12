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
