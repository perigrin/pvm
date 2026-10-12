// ABOUTME: Tests for lowering statements and the file: IfStmt, the loops, UseStmt, PhaserStmt, Block and lower.Lower.
// ABOUTME: Each want is built from the CST of its own fragment, so ids and the expressions other slices lower compare exactly.

package lower_test

import (
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

// kit builds a want over one parsed fragment: x lowers a node as an
// expression, at stamps a want with a node's id, and block spells the Block
// a CST block of expression statements lowers to.
type kit struct {
	root *parse.Node
	ids  map[*parse.Node]lower.ID
}

func (k kit) x(n *parse.Node) lower.Node { return lower.LowerExpr(k.root, n) }

func (k kit) at(n *parse.Node, out lower.Node) lower.Node {
	out.Meta().ID = k.ids[n]
	return out
}

func (k kit) block(b *parse.Node) lower.Node {
	var stmts []lower.Node
	for _, s := range b.Children {
		if s.Kind == parse.Statement {
			stmts = append(stmts, k.at(s, &lower.ExprStmt{X: k.x(s.Children[0])}))
		}
	}
	return k.at(b, &lower.Block{Stmts: stmts})
}

// label is the Bareword a statement label lowers to, as LoopEx's target does.
func (k kit) label(n *parse.Node) lower.Node { return k.at(n, &lower.Bareword{Name: n.Text}) }

// stmtCase is a program whose first statement s lowers to want(s, k).
type stmtCase struct {
	src  string
	want func(s *parse.Node, k kit) lower.Node
}

func checkStmts(t *testing.T, cases []stmtCase) {
	t.Helper()
	for _, c := range cases {
		root := parse.Parse([]byte(c.src))
		if len(root.Children) == 0 || root.Children[0].Kind != parse.Statement {
			t.Fatalf("%q: no leading statement", c.src)
		}
		s := root.Children[0]
		want := c.want(s, kit{root: root, ids: lower.AssignIDs(root)})
		if got := lower.LowerStmt(root, s); !reflect.DeepEqual(got, want) {
			t.Errorf("%q: lowered to %#v, want %#v", c.src, got, want)
		}
	}
}

// Every conditional form lowers to one IfStmt shape (§4.14.5). The modifier
// form's children are body then condition, the reverse of the block form's,
// and the Modifier flag says which (parse.go). `unless` is a branch swap,
// never a synthesised Not: perl 5.42's -MO=Deparse gives `unless ($c) {f()}
// else {g()}` as `if ($c) {g()} else {f()}`. An elsif is an IfStmt in the
// Else slot; under `unless` the chain lands in Then, because perl runs
// `unless ($c) {A} elsif ($d) {B} else {C}` as A for a false $c, and B or C
// by $d for a true one (measured over the four cases: A A C B).
func TestLowerConditionals(t *testing.T) {
	checkStmts(t, []stmtCase{
		{"if ($x) { 1 }", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			return k.at(c, &lower.IfStmt{Cond: k.x(c.Children[0]), Then: k.block(c.Children[1])})
		}},
		{"$y = 1 if $x;", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			return k.at(c, &lower.IfStmt{Cond: k.x(c.Children[1]), Then: k.x(c.Children[0])})
		}},
		{"$y = 1 unless $x;", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			return k.at(c, &lower.IfStmt{Cond: k.x(c.Children[1]), Else: k.x(c.Children[0])})
		}},
		{"unless ($x) { 1 }", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			return k.at(c, &lower.IfStmt{Cond: k.x(c.Children[0]), Else: k.block(c.Children[1])})
		}},
		{"unless ($x) { 1 } else { 2 }", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			return k.at(c, &lower.IfStmt{Cond: k.x(c.Children[0]), Then: k.block(c.Children[2].Children[0]), Else: k.block(c.Children[1])})
		}},
		{"if ($a) { 1 } elsif ($b) { 2 } else { 3 }", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			elsif := c.Children[2]
			return k.at(c, &lower.IfStmt{
				Cond: k.x(c.Children[0]),
				Then: k.block(c.Children[1]),
				Else: k.at(elsif, &lower.IfStmt{
					Cond: k.x(elsif.Children[0]),
					Then: k.block(elsif.Children[1]),
					Else: k.block(c.Children[3].Children[0]),
				}),
			})
		}},
		{"unless ($c) { 1 } elsif ($d) { 2 } else { 3 }", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			elsif := c.Children[2]
			return k.at(c, &lower.IfStmt{
				Cond: k.x(c.Children[0]),
				Then: k.at(elsif, &lower.IfStmt{
					Cond: k.x(elsif.Children[0]),
					Then: k.block(elsif.Children[1]),
					Else: k.block(c.Children[3].Children[0]),
				}),
				Else: k.block(c.Children[1]),
			})
		}},
	})

	// given, when, try and defer share the conditional kind and have no
	// IfStmt reading: each is a Hole until a consumer needs it.
	for _, src := range []string{
		"given ($x) { 1 }",
		"given ($x) { when (1) { 2 } }",
		"try { 1 } catch ($e) { 2 }",
		"use feature 'defer'; defer { 1 }",
	} {
		root := parse.Parse([]byte(src))
		s := root.Children[len(root.Children)-1]
		if got := lower.LowerStmt(root, s); reflect.TypeOf(got) != reflect.TypeOf(&lower.Hole{}) {
			t.Errorf("%q: lowered to %T, want *lower.Hole", src, got)
		}
	}
}

// Every loop form lowers to WhileStmt, CForStmt or ForeachStmt, and a bare
// block to BlockStmt. A plain `until` is a WhileStmt with Negated: two
// spellings of one loop. `do {} until` is not: perl 5.42 leaves 0 after
// `$m++ until 1` and 1 after `do { $n++ } until 1`, because the do-form tests
// after its body, and PostCond records it. `for` and `foreach` are one loop
// (toke.c routes both to yyl_foreach). A loop variable is lowered as written,
// `my $i` a declaration and `$i` a Var; none means $_. A label is a Bareword
// with the label node's id, as LoopEx's target is, and labels stack.
func TestLowerLoops(t *testing.T) {
	type L = []lower.Node
	checkStmts(t, []stmtCase{
		{"while ($x) { 1 }", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			return k.at(c, &lower.WhileStmt{Cond: k.x(c.Children[0]), Body: k.block(c.Children[1])})
		}},
		{"until ($x) { 1 }", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			return k.at(c, &lower.WhileStmt{Cond: k.x(c.Children[0]), Body: k.block(c.Children[1]), Negated: true})
		}},
		{"while ($x) { 1 } continue { 2 }", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			return k.at(c, &lower.WhileStmt{Cond: k.x(c.Children[0]), Body: k.block(c.Children[1]), Continue: k.block(c.Children[2].Children[0])})
		}},
		{"$m++ until $x;", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			return k.at(c, &lower.WhileStmt{Cond: k.x(c.Children[1]), Body: k.x(c.Children[0]), Negated: true})
		}},
		{"do { 1 } until $x;", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			return k.at(c, &lower.WhileStmt{Cond: k.x(c.Children[1]), Body: k.x(c.Children[0]), Negated: true, PostCond: true})
		}},
		{"do { 1 } while $x;", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			return k.at(c, &lower.WhileStmt{Cond: k.x(c.Children[1]), Body: k.x(c.Children[0]), PostCond: true})
		}},
		{"foreach my $i (@a) { 1 }", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			return k.at(c, &lower.ForeachStmt{Var: k.x(c.Children[0]), List: k.x(c.Children[1]), Body: k.block(c.Children[2])})
		}},
		{"for $i (@a) { 1 }", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			return k.at(c, &lower.ForeachStmt{Var: k.x(c.Children[0]), List: k.x(c.Children[1]), Body: k.block(c.Children[2])})
		}},
		{"for (@a) { 1 } continue { 2 }", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			return k.at(c, &lower.ForeachStmt{List: k.x(c.Children[0]), Body: k.block(c.Children[1]), Continue: k.block(c.Children[2].Children[0])})
		}},
		{"print 1 for @a;", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			return k.at(c, &lower.ForeachStmt{List: k.x(c.Children[1]), Body: k.x(c.Children[0])})
		}},
		{"for (my $i = 0; $i < 3; $i++) { 1 }", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			return k.at(c, &lower.CForStmt{Init: k.x(c.Children[0]), Cond: k.x(c.Children[1]), Post: k.x(c.Children[2]), Body: k.block(c.Children[3])})
		}},
		// An empty slot is nil: ForSlots says which slots the clauses fill.
		{"for (; $i < 3;) { 1 }", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[0]
			return k.at(c, &lower.CForStmt{Cond: k.x(c.Children[0]), Body: k.block(c.Children[1])})
		}},
		{"LABEL: for (;;) { last LABEL }", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[1]
			return k.at(c, &lower.CForStmt{Labels: L{k.label(s.Children[0])}, Body: k.block(c.Children[0])})
		}},
		{"A: B: while ($x) { 1 }", func(s *parse.Node, k kit) lower.Node {
			c := s.Children[2]
			return k.at(c, &lower.WhileStmt{Labels: L{k.label(s.Children[0]), k.label(s.Children[1])}, Cond: k.x(c.Children[0]), Body: k.block(c.Children[1])})
		}},
		// A bare block is a loop that runs once (last works in it).
		{"L: { 1 }", func(s *parse.Node, k kit) lower.Node {
			return k.at(s, &lower.BlockStmt{Labels: L{k.label(s.Children[0])}, Body: k.block(s.Children[1])})
		}},
		// The parser nests a bare block with a continue in a statement of
		// its own; the label stays on the outer one.
		{"L: { 1 } continue { 2 }", func(s *parse.Node, k kit) lower.Node {
			in := s.Children[1]
			return k.at(in, &lower.BlockStmt{Labels: L{k.label(s.Children[0])}, Body: k.block(in.Children[0]), Continue: k.block(in.Children[1].Children[0])})
		}},
	})
}
