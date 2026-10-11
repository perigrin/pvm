// ABOUTME: Tests for lowering the special forms of §4.14.2: MapGrepSort, the eval boundary, Match, and the renames.
// ABOUTME: Each want is built from the lowering of the CST children, so a child another slice lowers is compared exactly.

package lower_test

import (
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// shapeCase is one fragment whose expression e lowers to want(e, lx), where
// lx lowers any node of the same tree. Building the want from lx keeps it
// exact, ids included, whichever slice lowers the children.
type shapeCase struct {
	src  string
	want func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node
}

func checkShapes(t *testing.T, cases []shapeCase) {
	t.Helper()
	for _, c := range cases {
		root := parse.Parse([]byte(c.src))
		e := expression(t, root)
		lx := func(n *parse.Node) lower.Node { return lower.LowerExpr(root, n) }
		want := c.want(e, lx)
		want.Meta().ID = lower.AssignIDs(root)[e]
		if got := lx(e); !reflect.DeepEqual(got, want) {
			t.Errorf("%q: lowered to %#v, want %#v", c.src, got, want)
		}
	}
}

// map, grep and sort are one kind told apart by Op, a field, so nothing
// downstream matches on source text to separate map from grep. The lowering
// never sees the source: LowerExpr takes the tree alone.
//
// perl 5.42 reads each shape this way: `map $_+1, 1, 2` is 2,3 (the first
// comma operand is the expression, the rest the list); `sort byname qw(a c
// b)` with byname descending is c,b,a (the comparator is not an element);
// -MO=Deparse gives `map {$_;} @a, @b` and `sort byname @a`.
func TestLowerMapGrepSortOp(t *testing.T) {
	type L = []lower.Node
	checkShapes(t, []shapeCase{
		{"map { $_ } @a;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.MapGrepSort{Op: "map", Block: lx(e.Children[0]), List: L{lx(e.Children[1])}}
		}},
		{"grep { $_ } @a;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.MapGrepSort{Op: "grep", Block: lx(e.Children[0]), List: L{lx(e.Children[1])}}
		}},
		{"sort { $a <=> $b } @a;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.MapGrepSort{Op: "sort", Block: lx(e.Children[0]), List: L{lx(e.Children[1])}}
		}},
		// The list after a block is one comma tree, flattened.
		{"map { $_ } @a, @b;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			comma := e.Children[1]
			return &lower.MapGrepSort{Op: "map", Block: lx(e.Children[0]), List: L{lx(comma.Children[0]), lx(comma.Children[1])}}
		}},
		// The expression form: the first comma operand is the expression.
		{"map $_+1, @a, @b;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			outer := e.Children[0]
			inner := outer.Children[0]
			return &lower.MapGrepSort{Op: "map", FirstExpr: lx(inner.Children[0]), List: L{lx(inner.Children[1]), lx(outer.Children[1])}}
		}},
		{"map($_, @a);", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			comma := e.Children[0]
			return &lower.MapGrepSort{Op: "map", FirstExpr: lx(comma.Children[0]), List: L{lx(comma.Children[1])}}
		}},
		{"grep /x/, @a;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			comma := e.Children[0]
			return &lower.MapGrepSort{Op: "grep", FirstExpr: lx(comma.Children[0]), List: L{lx(comma.Children[1])}}
		}},
		// sort's expression form has no expression: every operand is listed.
		{"sort @a, @b;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			comma := e.Children[0]
			return &lower.MapGrepSort{Op: "sort", List: L{lx(comma.Children[0]), lx(comma.Children[1])}}
		}},
		// A comparator slot is FirstExpr, not an element.
		{"sort byname @a;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.MapGrepSort{Op: "sort", FirstExpr: lx(e.Children[0]), List: L{lx(e.Children[1])}}
		}},
	})

	// The comparator word lowers as a Bareword.
	root := parse.Parse([]byte("sort byname @a;"))
	if reflect.TypeOf(lower.LowerExpr(root, expression(t, root).Children[0])) != reflect.TypeOf(&lower.Bareword{}) {
		t.Errorf("sort byname @a: the comparator is not a Bareword")
	}
}

// eval BLOCK is code the analysis can read; eval EXPR is a string compiled
// at run time, the soundness boundary whose result is Any. They are two
// kinds, told apart by whether the child is a block. A missing operand
// means $_: perl 5.42's -MO=Deparse gives `eval;` as `eval $_`.
func TestLowerEvalBoundary(t *testing.T) {
	checkShapes(t, []shapeCase{
		{"eval { 1 };", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.EvalBlock{Body: lx(e.Children[0])}
		}},
		{`eval "1";`, func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.EvalStr{Code: lx(e.Children[0])}
		}},
		{"eval $s;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.EvalStr{Code: lx(e.Children[0])}
		}},
		{"eval;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.EvalStr{}
		}},
	})
}
