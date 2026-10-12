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

// Each special form the parser gives as a call, a term, a unary or a
// loop_control lowers to its own §4.14 kind. Read in perl 5.42:
//   - -MO=Deparse gives `<>` as `readline ARGV`, `<STDIN>` as `readline
//     STDIN` and `<*.c>` as `glob '*.c'`, so the brackets' body decides
//     readline from glob (§4.15.2);
//   - with @a three elements, \(@a) makes 3 references and \@a one, so Ref
//     keeps the operand's Paren bit;
//   - `{a=>1};` at statement start deparses as `+{'a', 1};`, a hash.
func TestLowerSpecialRenames(t *testing.T) {
	type L = []lower.Node
	checkShapes(t, []shapeCase{
		{"do { 1 };", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.DoBlock{Body: lx(e.Children[0])}
		}},
		{"do 'file.pl';", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.DoFile{Path: lx(e.Children[0])}
		}},
		{"wantarray;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Wantarray{}
		}},
		{"-e $f;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.FileTest{Op: 'e', Operand: lx(e.Children[0])}
		}},
		// No operand means $_.
		{"-e;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.FileTest{Op: 'e'}
		}},
		// Stacked tests nest.
		{"-e -f $x;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.FileTest{Op: 'e', Operand: lx(e.Children[0])}
		}},
		{`\@a;`, func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Ref{Operand: lx(e.Children[0])}
		}},
		{`\($a,$b);`, func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Ref{Operand: lx(e.Children[0])}
		}},
		{"<STDIN>;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Readline{}
		}},
		{"<$fh>;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Readline{}
		}},
		{"<>;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Readline{Magic: true}
		}},
		{"<<>>;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Readline{Magic: true}
		}},
		{"<*.c>;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Glob{}
		}},
		{"<$x/*>;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Glob{}
		}},
		{`glob("*.c");`, func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Glob{Pattern: lx(e.Children[0])}
		}},
		{"[1,2];", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.ArrayLiteral{Elems: L{lx(e.Children[0]), lx(e.Children[1])}}
		}},
		{"[];", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.ArrayLiteral{}
		}},
		// At statement start.
		{"{a=>1};", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.HashLiteral{Elems: L{lx(e.Children[0]), lx(e.Children[1])}}
		}},
		{"next;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.LoopEx{Op: "next"}
		}},
		{"goto &foo;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.LoopEx{Op: "goto", Label: lx(e.Children[0])}
		}},
		// return is a loop_control too; its value stands in Label.
		{"return (1,2);", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.LoopEx{Op: "return", Label: lx(e.Children[0])}
		}},
	})

	// In expression position: the anon_hash under unary plus.
	root := parse.Parse([]byte("+{a=>1};"))
	h := expression(t, root).Children[0]
	got := lower.LowerExpr(root, h)
	if reflect.TypeOf(got) != reflect.TypeOf(&lower.HashLiteral{}) || got.Meta().ID != lower.AssignIDs(root)[h] {
		t.Errorf("+{a=>1}: the anon_hash lowered to %#v, want a HashLiteral with its id", got)
	}

	// A label is a Bareword carrying the label node's id.
	root = parse.Parse([]byte("last FOO;"))
	e := expression(t, root)
	label := &lower.Bareword{Name: "FOO"}
	label.Meta().ID = lower.AssignIDs(root)[e.Children[0]]
	want := &lower.LoopEx{Op: "last", Label: label}
	want.Meta().ID = lower.AssignIDs(root)[e]
	if got := lower.LowerExpr(root, e); !reflect.DeepEqual(got, want) {
		t.Errorf("last FOO: lowered to %#v, want %#v", got, want)
	}
}

// A match binds its target with =~ or !~; a bare match reads $_, which a nil
// Target means. Op is the quote operator, `m` for a bare /.../. perl 5.42:
// with $x = "ab", `$x =~ /a/` is true and `$x !~ /a/` false; with $_ = "a",
// `/a/` is true.
func TestLowerMatchTargetAndNegated(t *testing.T) {
	checkShapes(t, []shapeCase{
		{"$x =~ /a/;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Match{Op: "m", Target: lx(e.Children[0])}
		}},
		{"$x !~ /a/;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Match{Op: "m", Target: lx(e.Children[0]), Negated: true}
		}},
		{"$x =~ s/a/b/g;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Match{Op: "s", Target: lx(e.Children[0])}
		}},
		{"$x =~ tr/a/b/;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Match{Op: "tr", Target: lx(e.Children[0])}
		}},
		{"/a/;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Match{Op: "m"}
		}},
		{"m{x};", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Match{Op: "m"}
		}},
		{"y/a/b/;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Match{Op: "y"}
		}},
		{"qr/x/;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Match{Op: "qr"}
		}},
	})
}

// hasHole reports whether a walk from n reaches a Hole.
func hasHole(n lower.Node) bool {
	if reflect.TypeOf(n) == reflect.TypeOf(&lower.Hole{}) {
		return true
	}
	for _, c := range lower.Children(n) {
		if hasHole(c) {
			return true
		}
	}
	return false
}

// A special form perl rejects still lowers to its kind, with a Hole where
// the parse failed. perl 5.42 rejects all three: `map {;` and `eval {;` with
// "Missing right curly or square bracket", `$x =~ ;` with a syntax error.
func TestLowerSpecialFormsOverBrokenInput(t *testing.T) {
	hole := reflect.TypeOf(&lower.Hole{})

	// The brace that never closes leaves an Unknown in the block slot.
	checkShapes(t, []shapeCase{
		{"map {;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.MapGrepSort{Op: "map", Block: lx(e.Children[0])}
		}},
		// No pattern to carry, so not a Match.
		{"$x =~ ;", func(e *parse.Node, lx func(*parse.Node) lower.Node) lower.Node {
			return &lower.Binary{Op: "=~", L: lx(e.Children[0]), R: lx(e.Children[1])}
		}},
	})
	for _, c := range []struct {
		src  string
		slot func(*parse.Node) *parse.Node
	}{
		{"map {;", func(e *parse.Node) *parse.Node { return e.Children[0] }},
		{"$x =~ ;", func(e *parse.Node) *parse.Node { return e.Children[1] }},
	} {
		root := parse.Parse([]byte(c.src))
		if got := lower.LowerExpr(root, c.slot(expression(t, root))); reflect.TypeOf(got) != hole {
			t.Errorf("%q: the failed slot lowered to %T, want *lower.Hole", c.src, got)
		}
	}

	// The block parses and holds the Unknown; the body reaches a Hole.
	root := parse.Parse([]byte("eval {;"))
	got := lower.LowerExpr(root, expression(t, root))
	if reflect.TypeOf(got) != reflect.TypeOf(&lower.EvalBlock{}) {
		t.Fatalf("eval {;: lowered to %T, want *lower.EvalBlock", got)
	}
	if kids := lower.Children(got); len(kids) != 1 || !hasHole(kids[0]) {
		t.Errorf("eval {;: the body does not hold a Hole")
	}
}

// §4.5.4: => quotes the word to its left in a map, grep or sort list as it
// does in any other list. perl 5.42's -MO=Deparse gives `map { $_ } a => 1`
// as `map {$_;} 'a', 1`.
func TestLowerMapGrepSortListQuotesFatWord(t *testing.T) {
	word := &lower.Bareword{Name: "a", Quoted: true}
	for _, src := range []string{"map { $_ } a => 1;", "grep { $_ } a => 1;", "sort a => 1;"} {
		got := lowerOne(t, src)
		if !sameKind(got, &lower.MapGrepSort{}) || field(got, "List").Len() != 2 {
			t.Errorf("%q: lowered to %#v, want a MapGrepSort listing two elements", src, got)
			continue
		}
		if first := field(zeroIDs(got), "List").Index(0).Interface(); !reflect.DeepEqual(first, word) {
			t.Errorf("%q: the list's first element is %#v, want %#v", src, first, word)
		}
	}
}
