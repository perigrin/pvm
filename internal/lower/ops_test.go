// ABOUTME: Tests for lowering operators: Binary, Unary, Repeat, Assign, Range, Ternary, CmpChain and List.
// ABOUTME: Each operator is read off parse.Node fields; each claim about perl's reading is measured on 5.42.

package lower_test

import (
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// lowerOne parses a one-statement program and lowers its expression.
func lowerOne(t *testing.T, src string) lower.Node {
	t.Helper()
	root := parse.Parse([]byte(src))
	return lower.LowerExpr(root, expression(t, root))
}

// clearIDs zeroes the id of n and of every node under it, so two lowerings
// compare by shape alone.
func clearIDs(n lower.Node) lower.Node {
	if n == nil {
		return nil
	}
	n.Meta().ID = lower.ID{}
	for _, c := range lower.Children(n) {
		clearIDs(c)
	}
	return n
}

// withParen sets n's Paren bit, as source parentheses around it would.
func withParen(n lower.Node) lower.Node {
	n.Meta().Paren = true
	return n
}

// sameKind reports whether got is the same kind of node as want.
func sameKind(got, want lower.Node) bool {
	return reflect.TypeOf(got) == reflect.TypeOf(want)
}

// field reads the named field of a lowered node. Operands of variables are
// another slice's lowering, so these tests read the field under test rather
// than comparing whole trees that hold them.
func field(n lower.Node, name string) any {
	return reflect.ValueOf(n).Elem().FieldByName(name).Interface()
}

// opOf names the operator a lowered node was built from: the Op field where
// the kind has one, and otherwise the operator the kind itself stands for.
func opOf(n lower.Node) string {
	switch {
	case sameKind(n, &lower.Binary{}), sameKind(n, &lower.Assign{}):
		return field(n, "Op").(string)
	case sameKind(n, &lower.Repeat{}):
		return "x"
	case sameKind(n, &lower.Range{}):
		if field(n, "Exclusive").(bool) {
			return "..."
		}
		return ".."
	}
	return ""
}

// The operator is the node's Text, never a rescan of its bytes.
// findOperatorText scanned a node's children for text matching an operator
// table, so an operand spelled like an operator could answer for it.
func TestLowerBinaryOpFromField(t *testing.T) {
	cases := []struct {
		src  string
		want lower.Node
		op   string
	}{
		{"$a + $b;", &lower.Binary{}, "+"},
		{"$a - $b;", &lower.Binary{}, "-"},
		{"$a . $b;", &lower.Binary{}, "."},
		{"$a x 3;", &lower.Repeat{}, "x"},
		// A lone comparison is a cmp_chain of one link in the CST and a
		// plain binary op in perl: `$q = $a < $b` compiles to `lt`, where
		// `$a < $b < $c` compiles to cmpchain_and.
		{"$a == $b;", &lower.Binary{}, "=="},
		{"$a eq $b;", &lower.Binary{}, "eq"},
		{"$a < $b;", &lower.Binary{}, "<"},
		{"$a lt $b;", &lower.Binary{}, "lt"},
		{"$a // $b;", &lower.Binary{}, "//"},
		{"$a .. $b;", &lower.Range{}, ".."},
		{"$a ... $b;", &lower.Range{}, "..."},
		{"$a =~ $b;", &lower.Binary{}, "=~"},
		{"$a !~ $b;", &lower.Binary{}, "!~"},
		{"$a = $b;", &lower.Assign{}, "="},
		{"$a += 1;", &lower.Assign{}, "+="},
		{"$a <=> $b;", &lower.Binary{}, "<=>"},
		{"$a && $b;", &lower.Binary{}, "&&"},
		{"$a and $b;", &lower.Binary{}, "and"},
		// Operands spelled like operators.
		{"'..' . '.';", &lower.Binary{}, "."},
		{"'+' - '+';", &lower.Binary{}, "-"},
		{"'x' x 2;", &lower.Repeat{}, "x"},
		{"'eq' eq '==';", &lower.Binary{}, "eq"},
	}
	for _, c := range cases {
		got := lowerOne(t, c.src)
		if !sameKind(got, c.want) || opOf(got) != c.op {
			t.Errorf("%q: lowered to %T op %q, want %T op %q", c.src, got, opOf(got), c.want, c.op)
		}
	}

	// Both operands of a Binary are lowered.
	sameShape(t, "'..' . '.';", clearIDs(lowerOne(t, "'..' . '.';")),
		&lower.Binary{Op: ".", L: &lower.StrLit{Value: ".."}, R: &lower.StrLit{Value: "."}})

	// No rescan: Op follows Text even where the bytes say otherwise.
	root := parse.Parse([]byte("1 - 2;"))
	n := expression(t, root)
	n.Text = "+"
	if got := opOf(lower.LowerExpr(root, n)); got != "+" {
		t.Errorf("Op is %q after Text became \"+\"; the lowering read the source", got)
	}
}

// §4.10: a parenthesised left operand, or a qw list, repeats a list; any
// other repeats a string. perl 5.42: ("a") x 3, qw(a b) x 2 and (1,2) x 2
// give 3, 4 and 4 elements; "a" x 3 gives 1, and so does (1,2)[0] x 2,
// whose parens close before the subscript.
func TestLowerListRepeatFromParenFlag(t *testing.T) {
	cases := []struct {
		src  string
		want bool
	}{
		{`("a") x 3;`, true},
		{`"a" x 3;`, false},
		{`'a' x 3;`, false},
		{"qw(a b) x 2;", true},
		{"(qw(a b)) x 2;", true},
		{"(1,2) x 2;", true},
		{"(1,2)[0] x 2;", false},
	}
	for _, c := range cases {
		got := lowerOne(t, c.src)
		if !sameKind(got, &lower.Repeat{}) || field(got, "ListRepeat") != c.want {
			t.Errorf("%q: lowered to %#v, want a Repeat with ListRepeat %v", c.src, got, c.want)
		}
	}
	sameShape(t, `("a") x 3;`, clearIDs(lowerOne(t, `("a") x 3;`)), &lower.Repeat{
		Left:       withParen(&lower.StrLit{Value: "a", Interpolated: true}),
		Count:      &lower.NumLit{Text: "3", IsInt: true},
		ListRepeat: true,
	})
}

// §4.14.2, Range row: `1..5` is `binary ".."` over two terms, so a literal
// endpoint is an operand like any other. perl: `my @r = (1..5); $r[0]+1` is
// 2, and `...` differs from `..` only as a flip-flop.
func TestLowerRangeEndpoints(t *testing.T) {
	cases := []struct {
		src       string
		exclusive bool
	}{
		{"$a..$b;", false},
		{"1..5;", false},
		{"$a...$b;", true},
	}
	for _, c := range cases {
		got := lowerOne(t, c.src)
		if !sameKind(got, &lower.Range{}) {
			t.Errorf("%q: lowered to %T, want a Range", c.src, got)
			continue
		}
		if field(got, "Lo") == nil || field(got, "Hi") == nil {
			t.Errorf("%q: lowered to %#v, want both endpoints", c.src, got)
		}
		if field(got, "Exclusive") != c.exclusive {
			t.Errorf("%q: Exclusive is %v, want %v", c.src, field(got, "Exclusive"), c.exclusive)
		}
	}
	sameShape(t, "1..5;", clearIDs(lowerOne(t, "1..5;")), &lower.Range{
		Lo: &lower.NumLit{Text: "1", IsInt: true},
		Hi: &lower.NumLit{Text: "5", IsInt: true},
	})
}
