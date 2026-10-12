// ABOUTME: Tests for lowering operators: Binary, Unary, Repeat, Assign, Range, Ternary, CmpChain and List.
// ABOUTME: Each operator is read off parse.Node fields; each claim about perl's reading is measured on 5.42.

package lower_test

import (
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// withParen sets n's Paren bit, as source parentheses around it would.
func withParen(n lower.Node) lower.Node {
	n.Meta().Paren = true
	return n
}

// sameKind reports whether got is the same kind of node as want.
func sameKind(got, want lower.Node) bool {
	return reflect.TypeOf(got) == reflect.TypeOf(want)
}

// fieldOf reads the named field of a lowered node. Operands of variables are
// another slice's lowering, so these tests read the field under test rather
// than comparing whole trees that hold them.
func fieldOf(n lower.Node, name string) any {
	return reflect.ValueOf(n).Elem().FieldByName(name).Interface()
}

// opOf names the operator a lowered node was built from: the Op field where
// the kind has one, and otherwise the operator the kind itself stands for.
func opOf(n lower.Node) string {
	switch {
	case sameKind(n, &lower.Binary{}), sameKind(n, &lower.Assign{}):
		return fieldOf(n, "Op").(string)
	case sameKind(n, &lower.Repeat{}):
		return "x"
	case sameKind(n, &lower.Range{}):
		if fieldOf(n, "Exclusive").(bool) {
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
	sameShape(t, "'..' . '.';", zeroIDs(lowerOne(t, "'..' . '.';")),
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
		if !sameKind(got, &lower.Repeat{}) || fieldOf(got, "ListRepeat") != c.want {
			t.Errorf("%q: lowered to %#v, want a Repeat with ListRepeat %v", c.src, got, c.want)
		}
	}
	sameShape(t, `("a") x 3;`, zeroIDs(lowerOne(t, `("a") x 3;`)), &lower.Repeat{
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
		if fieldOf(got, "Lo") == nil || fieldOf(got, "Hi") == nil {
			t.Errorf("%q: lowered to %#v, want both endpoints", c.src, got)
		}
		if fieldOf(got, "Exclusive") != c.exclusive {
			t.Errorf("%q: Exclusive is %v, want %v", c.src, fieldOf(got, "Exclusive"), c.exclusive)
		}
	}
	sameShape(t, "1..5;", zeroIDs(lowerOne(t, "1..5;")), &lower.Range{
		Lo: &lower.NumLit{Text: "1", IsInt: true},
		Hi: &lower.NumLit{Text: "5", IsInt: true},
	})
}

// lowerArgs lowers the argument chain of a one-call program: the call's
// only child, a `binary ","` tree rather than a `list`.
func lowerArgs(t *testing.T, src string) lower.Node {
	t.Helper()
	root := parse.Parse([]byte(src))
	call := expression(t, root)
	if call.Kind != parse.Call || len(call.Children) != 1 {
		t.Fatalf("%q: want a call with one argument chain", src)
	}
	return lower.LowerExpr(root, call.Children[0])
}

// §4.5.4: => quotes the word to its left. perl 5.42: (a => 1, b => 2) is
// ("a", 1, "b", 2). Outside a `list` node the commas are a left-nested
// binary chain, and there => is an operator rather than a Fat flag.
func TestLowerListFat(t *testing.T) {
	word := func(s string) lower.Node { return &lower.Bareword{Name: s, Quoted: true} }
	num := func(s string) lower.Node { return &lower.NumLit{Text: s, IsInt: true} }
	sameShape(t, "(a => 1, b => 2);", zeroIDs(lowerOne(t, "(a => 1, b => 2);")), withParen(&lower.List{
		Elems: []lower.Node{word("a"), num("1"), word("b"), num("2")},
		Fat:   []bool{true, false, true, false},
	}))

	got := lowerOne(t, "(a, 1, b, 2);")
	if !sameKind(got, &lower.List{}) || !reflect.DeepEqual(fieldOf(got, "Fat"), []bool{false, false, false, false}) {
		t.Errorf("(a, 1, b, 2): lowered to %#v, want a List of four with no Fat", got)
	}

	// The chain lowers to the List its parenthesised form does, but for
	// the parens themselves.
	for _, c := range []struct{ chain, paren string }{
		{"f(1, 2, 3);", "(1, 2, 3);"},
		{"f(a => 1, b => 2);", "(a => 1, b => 2);"},
	} {
		got := zeroIDs(lowerArgs(t, c.chain))
		want := zeroIDs(lowerOne(t, c.paren))
		want.Meta().Paren = false
		if !sameKind(got, &lower.List{}) || !reflect.DeepEqual(got, want) {
			t.Errorf("%q: lowered to %#v, want %#v as %q does", c.chain, got, want, c.paren)
		}
	}
}

// §4.14.2, Assign row: list assignment is the LHS's syntactic shape. perl
// 5.42 compiles each true case to aassign and each false one to sassign;
// `($a) = 1` is a list assignment for its parens alone, and `$#a` is a
// scalar for all its `#`.
func TestLowerAssignListness(t *testing.T) {
	cases := []struct {
		src  string
		want bool
	}{
		{"($a,$b) = (1,2);", true},
		{"@a = (1,2);", true},
		{"%h = (1,2);", true},
		{"($a) = 1;", true},
		{"@$r = (1,2);", true},
		{"@h{1,2} = (1,2);", true},
		{"$r->@* = (1,2);", true},
		{"$r->@[0,1] = (1,2);", true},
		{"$a = 1;", false},
		{"$a = (1,2);", false},
		{"$#a = 1;", false},
		{"$h{a} = (1,2);", false},
		{"$r->[0] = 1;", false},
	}
	for _, c := range cases {
		got := lowerOne(t, c.src)
		if !sameKind(got, &lower.Assign{}) || fieldOf(got, "ListAssign") != c.want {
			t.Errorf("%q: lowered to %#v, want an Assign with ListAssign %v", c.src, got, c.want)
		}
	}
}

// Each input is one perl 5.42 rejects with `perl -c`, so the parser cannot
// grow to accept it: the operand it refuses is an Unknown, and the operator
// around it lowers intact with a Hole in that operand.
func TestLowerOperatorsOverBrokenInput(t *testing.T) {
	assign := lowerOne(t, "$x = ;")
	if !sameKind(assign, &lower.Assign{}) || !sameKind(fieldOf(assign, "R").(lower.Node), &lower.Hole{}) {
		t.Errorf("$x = ;: lowered to %#v, want an Assign whose R is a Hole", assign)
	}

	assign = lowerOne(t, "$x = 1 +;")
	plus, _ := fieldOf(assign, "R").(lower.Node)
	if !sameKind(assign, &lower.Assign{}) || !sameKind(plus, &lower.Binary{}) ||
		fieldOf(plus, "Op") != "+" || !sameKind(fieldOf(plus, "R").(lower.Node), &lower.Hole{}) {
		t.Errorf("$x = 1 +;: lowered to %#v, want an Assign over a Binary \"+\" whose R is a Hole", assign)
	}

	repeat := lowerOne(t, "1 x;")
	if !sameKind(repeat, &lower.Repeat{}) || !sameKind(fieldOf(repeat, "Count").(lower.Node), &lower.Hole{}) {
		t.Errorf("1 x;: lowered to %#v, want a Repeat whose Count is a Hole", repeat)
	}
}

// §4.14.2: Unary's Op is the node's Text and Postfix is the CST's Postfix
// kind; a Ternary keeps its three operands; a chain of two or more links is
// one CmpChain with flat operands, as perl 5.42 compiles `$a < $b < $c` to
// one cmpchain_and.
func TestLowerUnaryTernaryAndChain(t *testing.T) {
	num := func(s string) lower.Node { return &lower.NumLit{Text: s, IsInt: true} }
	cases := []struct {
		src  string
		want lower.Node
	}{
		{"!1;", &lower.Unary{Op: "!", Operand: num("1")}},
		{"not 1;", &lower.Unary{Op: "not", Operand: num("1")}},
		{"-1;", &lower.Unary{Op: "neg", Operand: num("1")}},
		{"+1;", &lower.Unary{Op: "pos", Operand: num("1")}},
		{"~1;", &lower.Unary{Op: "~", Operand: num("1")}},
		{"1 ? 2 : 3;", &lower.Ternary{Cond: num("1"), Then: num("2"), Else: num("3")}},
		{"1 < 2 <= 3;", &lower.CmpChain{
			Operands: []lower.Node{num("1"), num("2"), num("3")},
			Ops:      []string{"<", "<="},
			Class:    lower.ChRelOp,
		}},
		{"1 == 2 != 3;", &lower.CmpChain{
			Operands: []lower.Node{num("1"), num("2"), num("3")},
			Ops:      []string{"==", "!="},
			Class:    lower.ChEqOp,
		}},
	}
	for _, c := range cases {
		sameShape(t, c.src, zeroIDs(lowerOne(t, c.src)), c.want)
	}

	// ++ and -- need a variable, which another slice lowers: read the
	// fields under test.
	for _, c := range []struct {
		src     string
		op      string
		postfix bool
	}{
		{"++$a;", "++", false},
		{"--$a;", "--", false},
		{"$a++;", "++", true},
		{"$a--;", "--", true},
	} {
		got := lowerOne(t, c.src)
		if !sameKind(got, &lower.Unary{}) || fieldOf(got, "Op") != c.op || fieldOf(got, "Postfix") != c.postfix {
			t.Errorf("%q: lowered to %#v, want a Unary %q with Postfix %v", c.src, got, c.op, c.postfix)
		}
	}
}

// An arrow is a method call or a dereference, never an operator with an
// Op: until an arm claims it, it is a Hole rather than a Binary "->".
func TestLowerArrowIsNotABinary(t *testing.T) {
	if got := lowerOne(t, "Foo->new;"); sameKind(got, &lower.Binary{}) {
		t.Errorf("Foo->new: lowered to %#v, want anything but a Binary", got)
	}
}
