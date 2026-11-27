// ABOUTME: Tests the judgements the context makes: a completion pairs only with the parents that predicted it under
// ABOUTME: its context, and a scan V1's shift verdict forbids under an item's context is refused (V2 design §4.1, §4.3).

package recognize

import (
	"slices"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

// coreItem is the tables' core item of the rule lhs ::= rhs (symbol names) with its dot at dot; it fails tb where
// there is none.
func coreItem(tb testing.TB, tab *grammar.Tables, dot int, lhs string, rhs ...string) int {
	tb.Helper()
	i := slices.IndexFunc(tab.Items, func(it grammar.CoreItem) bool {
		rule := tab.Rules[it.Rule]
		if it.Dot != dot || tab.Symbols[rule.LHS].Name != lhs || len(rule.RHS) != len(rhs) {
			return false
		}
		for k, s := range rule.RHS {
			if tab.Symbols[s].Name != rhs[k] {
				return false
			}
		}
		return true
	})
	if i < 0 {
		tb.Fatalf("no core item %s ::= %v with its dot at %d", lhs, rhs, dot)
	}
	return i
}

// itemCtx is the context (Item.Ctx) whose spine top is core item i.
func itemCtx(r *Recognizer, i int) int { return itemContext(r.judge.p.Item(i)) }

// holding is the items of s whose state holds core item i.
func holding(tab *grammar.Tables, s []Item, i int) (out []Item) {
	for _, it := range s {
		if slices.Contains(tab.States[it.State].Items, i) {
			out = append(out, it)
		}
	}
	return out
}

// TestCompletionPairsUnderContext: in `foo -$x * $y;` token 1 is {ADDOP, PERLY_MINUS}, so at node 2 binary minus
// (term ADDOP . term, from node 0) and unary minus (PERLY_MINUS . term, from node 1) both await a term, and their
// predicted items there carry each one's real top. Completing `$x * $y`, from node 2 under binary minus's context,
// at node 5 advances binary minus and never unary minus: the completion pairs only with the parents that predicted
// it under its context (F6-R4, F7-A2). The completion is closed on its own, so the scan of * under unary minus,
// which the shift side judges, plays no part.
func TestCompletionPairsUnderContext(t *testing.T) {
	tab := committedTables(t)
	r := mustJudged(t, tab, nil, Options{NoLeo: true})
	c := perlyChart(t, r, "foo -$x * $y;")
	bin := coreItem(t, tab, 2, "termbinop", "term", "ADDOP", "term")
	un := coreItem(t, tab, 1, "termunop", "PERLY_MINUS", "term")
	bctx, uctx := itemCtx(r, bin), itemCtx(r, un)
	if bctx == 0 || uctx == 0 || bctx == uctx {
		t.Fatalf("binary minus's context %d, unary minus's %d: want two real tops", bctx, uctx)
	}
	bp, up := holding(tab, c.Sets[2].Items(), bin), holding(tab, c.Sets[2].Items(), un)
	if len(bp) == 0 || len(up) == 0 {
		t.Fatalf("node 2 %v: binary minus's parents %v, unary minus's %v; want both", c.Sets[2].Items(), bp, up)
	}
	term := symbolID(tab, "term")
	mul := coreItem(t, tab, 3, "termbinop", "term", "MULOP", "term")
	var seed []Item // $x * $y, complete at node 5 from node 2 under binary minus
	for _, it := range holding(tab, c.Sets[5].Items(), mul) {
		if it.Dist == 3 && it.Ctx == bctx {
			seed = append(seed, it)
		}
	}
	if len(seed) == 0 {
		t.Fatalf("node 5 %v holds no $x * $y from node 2 under binary minus's context", c.Sets[5].Items())
	}
	b := &builder{r: r, c: c, pos: 5}
	for _, it := range seed {
		b.add(it.State, 5-it.Dist, it.Ctx)
	}
	b.complete(nil)
	advanced := func(parents []Item) (n int) {
		for _, p := range parents {
			k, ok := transition(tab, p.State, term)
			if ok && slices.Contains(b.items, [3]int{k, 2 - p.Dist, p.Ctx}) {
				n++
			}
		}
		return n
	}
	if n := advanced(bp); n != len(bp) {
		t.Errorf("$x * $y under binary minus advances %d of binary minus's %d parents %v; want every one", n, len(bp), bp)
	}
	if n := advanced(up); n != 0 {
		t.Errorf("$x * $y under binary minus advances %d of unary minus's parents %v; want none", n, up)
	}
}

// TestShiftSideRefusesForbiddenScan: in `foo -$x * $y;` the term after the minus carries binary minus's context and
// unary minus's. Unary minus binds tighter than *, so V1 forbids the shift of * after `term . MULOP term` under its
// top: that scan is refused, while under binary minus's context, and under none in the no-precedence chart, it is
// kept. With the pairing rule, the chart never completes `$x * $y` into unary minus.
func TestShiftSideRefusesForbiddenScan(t *testing.T) {
	tab := committedTables(t)
	r := mustJudged(t, tab, nil, Options{})
	const src = "foo -$x * $y;"
	c := perlyChart(t, r, src)
	plain := perlyChart(t, NewNoPrecedence(tab, Options{}), src)
	if !c.Accepted() || !plain.Accepted() {
		t.Fatalf("%q: judged accepts %v, no-precedence %v", src, c.Accepted(), plain.Accepted())
	}
	bctx := itemCtx(r, coreItem(t, tab, 2, "termbinop", "term", "ADDOP", "term"))
	uctx := itemCtx(r, coreItem(t, tab, 1, "termunop", "PERLY_MINUS", "term"))
	mul := symbolID(tab, "MULOP")
	before := coreItem(t, tab, 1, "termbinop", "term", "MULOP", "term")
	after := coreItem(t, tab, 2, "termbinop", "term", "MULOP", "term")
	ctxs := func(s []Item, i, dist int) (out []int) {
		for _, it := range holding(tab, s, i) {
			if it.Dist == dist {
				out = append(out, it.Ctx)
			}
		}
		return out
	}
	// At node 3 $x awaits * under both contexts; the state's verdict on * is refused under unary minus only.
	if got := ctxs(c.Sets[3].Items(), before, 1); !slices.Contains(got, uctx) || !slices.Contains(got, bctx) {
		t.Fatalf("node 3 %v: $x awaits * under contexts %v; want unary minus's %d and binary minus's %d", c.Sets[3].Items(), got, uctx, bctx)
	}
	for _, it := range holding(tab, c.Sets[3].Items(), before) {
		if it.Dist != 1 {
			continue
		}
		if got, want := r.scanRefused(it.State, it.Ctx, mul), it.Ctx == uctx; got != want {
			t.Errorf("state %d under context %d: the scan of * refused %v, want %v", it.State, it.Ctx, got, want)
		}
		if r.scanRefused(it.State, 0, mul) {
			t.Errorf("state %d under none: the scan of * refused", it.State)
		}
	}
	if got := ctxs(c.Sets[4].Items(), after, 2); !slices.Equal(got, []int{bctx}) {
		t.Errorf("node 4 %v: $x * under contexts %v; want binary minus's %d only, unary minus's %d refused", c.Sets[4].Items(), got, bctx, uctx)
	}
	if got := ctxs(plain.Sets[4].Items(), after, 2); !slices.Equal(got, []int{0}) {
		t.Errorf("no-precedence node 4 %v: $x * under contexts %v; want none's, kept", plain.Sets[4].Items(), got)
	}
	term := symbolID(tab, "term")
	for _, p := range holding(tab, c.Sets[2].Items(), coreItem(t, tab, 1, "termunop", "PERLY_MINUS", "term")) {
		k, _ := transition(tab, p.State, term)
		if slices.Contains(c.Sets[5].Items(), Item{State: k, Dist: 5 - (2 - p.Dist), Ctx: p.Ctx}) {
			t.Errorf("node 5 %v: unary minus %v completed over $x * $y", c.Sets[5].Items(), p)
		}
	}
}

// TestPredictorTakesNonNoneTop: `do`, UNIOP and `my \` begin some rules under none and some under their own item, a
// real top (§4.2). Each gives the items it predicts that top, and under it the scan of an operator V1 forbids inside
// the operand is refused, where none, which taking none would give, keeps it: the operand's term awaits the operator
// under the top at the node before it, and the chart holds no item that scanned it from the operand's start, while
// the no-precedence chart does.
func TestPredictorTakesNonNoneTop(t *testing.T) {
	tab := committedTables(t)
	for _, tc := range []struct {
		name, src string
		top       []string // the predictor's item: dot, lhs, rhs
		dot       int
		pred, op  int // the node the predictor predicts at, and the operator's token
		opName    string
	}{
		{"do", "do $f || 1;", []string{"termdo", "KW_DO", "term"}, 1, 1, 2, "OROR"},
		{"UNIOP", "ref $x || 1;", []string{"term", "UNIOP", "term"}, 1, 1, 2, "OROR"},
		{"my-backslash", `my \$x + 1;`, []string{"myattrterm", "KW_MY", "REFGEN", "term"}, 2, 2, 3, "ADDOP"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := mustJudged(t, tab, nil, Options{})
			c := perlyChart(t, r, tc.src)
			plain := perlyChart(t, NewNoPrecedence(tab, Options{}), tc.src)
			top := coreItem(t, tab, tc.dot, tc.top[0], tc.top[1:]...)
			ctx := itemCtx(r, top)
			kernels := holding(tab, c.Sets[tc.pred].Items(), top)
			if len(kernels) == 0 {
				t.Fatalf("node %d %v holds no item of %v", tc.pred, c.Sets[tc.pred].Items(), tc.top)
			}
			for _, k := range kernels {
				if got := r.predictedContext(k.State); got != ctx || got == 0 {
					t.Errorf("state %d predicts under context %d; want its top %d, not none", k.State, got, ctx)
				}
			}
			predicted := 0
			for _, it := range c.Sets[tc.pred].Items() {
				if it.Dist == 0 {
					predicted++
					if it.Ctx != ctx {
						t.Errorf("node %d: predicted item %v under context %d; want the top %d", tc.pred, it, it.Ctx, ctx)
					}
				}
			}
			if predicted == 0 {
				t.Fatalf("node %d %v predicts nothing", tc.pred, c.Sets[tc.pred].Items())
			}
			op := symbolID(tab, tc.opName)
			before := coreItem(t, tab, 1, "termbinop", "term", tc.opName, "term")
			after := coreItem(t, tab, 2, "termbinop", "term", tc.opName, "term")
			waiting := 0
			for _, it := range holding(tab, c.Sets[tc.op].Items(), before) {
				if it.Dist != tc.op-tc.pred {
					continue
				}
				waiting++
				if it.Ctx != ctx || !r.scanRefused(it.State, it.Ctx, op) || r.scanRefused(it.State, 0, op) {
					t.Errorf("node %d: %v awaits %s under context %d, refused %v, refused under none %v; want the top %d, refused, kept under none",
						tc.op, it, tc.opName, it.Ctx, r.scanRefused(it.State, it.Ctx, op), r.scanRefused(it.State, 0, op), ctx)
				}
			}
			if waiting == 0 {
				t.Fatalf("node %d %v: the operand awaits no %s", tc.op, c.Sets[tc.op].Items(), tc.opName)
			}
			scanned := func(s []Item) bool {
				return slices.ContainsFunc(holding(tab, s, after), func(it Item) bool { return it.Dist == tc.op+1-tc.pred })
			}
			if scanned(c.Sets[tc.op+1].Items()) {
				t.Errorf("node %d %v: the operand scanned %s under the top", tc.op+1, c.Sets[tc.op+1].Items(), tc.opName)
			}
			if !scanned(plain.Sets[tc.op+1].Items()) {
				t.Errorf("no-precedence node %d %v: the operand did not scan %s under none", tc.op+1, plain.Sets[tc.op+1].Items(), tc.opName)
			}
		})
	}
}

// slipShiftToy is S ::= X, X ::= X ; | X } X | X T X | P X | A, with ; postfix, } and T infix, P prefix at level 2,
// T below it, and the levels given to ; and }: the slipper's inserted ; before }, and the } scanned after it, are
// both shifts V1 judges under P's top, which T keeps a real top. With } below P, a } after X . } X under P's top is
// forbidden; with ; below P, a ; after X . ; is.
func slipShiftToy(t *testing.T, semi, brace int) (*grammar.Tables, *Recognizer) {
	t.Helper()
	levels := []grammar.Level{{N: 1, Assoc: "left", Tokens: []string{"T"}}, {N: 2, Assoc: "right", Tokens: []string{"P"}}}
	for _, l := range []grammar.Level{{N: semi, Assoc: "left", Tokens: []string{"PERLY_SEMICOLON"}}, {N: brace, Assoc: "left", Tokens: []string{"PERLY_BRACE_CLOSE"}}} {
		if k := slices.IndexFunc(levels, func(x grammar.Level) bool { return x.N == l.N }); k >= 0 {
			levels[k].Tokens = append(levels[k].Tokens, l.Tokens...)
		} else {
			levels = append(levels, l)
		}
	}
	slices.SortFunc(levels, func(a, b grammar.Level) int { return a.N - b.N })
	tab, p := precToy(t, levels, nil, nil, "S ::= X", "X ::= X PERLY_SEMICOLON | X PERLY_BRACE_CLOSE X | X T X | P X | A")
	return tab, mustJudged(t, tab, p, Options{})
}

// TestShiftRefusalThroughSlipper: on P A } A the slipper inserts a ; before the }, where X . ; awaits it from node 1,
// under P's top, and from node 0, under none. Where ; is below P, V1 forbids its shift under P's top: the slipper's
// closure at node 2 holds no item that scanned it from node 1, while the no-precedence closure does. Where } is below
// P and ; above, the ; is kept and completes X from node 1, so X . } X awaits the } under P's top from node 1 only
// through the inserted ;, the direct scan having refused it: V1 forbids that shift, and the chart holds no item
// that scanned the } from node 1, while the no-precedence chart does. Where both are above P, the } is kept through
// the inserted ;, and the item it advances carries P's top.
func TestShiftRefusalThroughSlipper(t *testing.T) {
	t.Run("semicolon", func(t *testing.T) {
		tab, r := slipShiftToy(t, 1, 3)
		l := tokens(tab, "P", "A", "PERLY_BRACE_CLOSE", "A")
		next := l.Edges[2].Token.TerminalSet
		c := mustRecognize(t, r, l)
		plain := mustRecognize(t, NewNoPrecedence(tab, Options{}), l)
		if !slices.Contains(c.Inserted(), 2) || !slices.Contains(plain.Inserted(), 2) {
			t.Fatalf("no ; inserted at node 2: judged %v, no-precedence %v", c.Inserted(), plain.Inserted())
		}
		after := coreItem(t, tab, 2, "X", "X", "PERLY_SEMICOLON")
		from1 := func(items []Item) []Item {
			var out []Item
			for _, it := range holding(tab, items, after) {
				if it.Dist == 1 {
					out = append(out, it)
				}
			}
			return out
		}
		if got := from1(c.Slipped(2, next)); len(got) != 0 {
			t.Errorf("the slipper's closure at node 2 %v holds X ; from node 1, %v, under P's top", c.Slipped(2, next), got)
		}
		if got := from1(plain.Slipped(2, next)); len(got) == 0 {
			t.Errorf("the no-precedence closure at node 2 %v lacks X ; from node 1: the fixture never inserts it there", plain.Slipped(2, next))
		}
	})
	t.Run("brace", func(t *testing.T) {
		tab, r := slipShiftToy(t, 3, 1)
		l := tokens(tab, "P", "A", "PERLY_BRACE_CLOSE", "A")
		next := l.Edges[2].Token.TerminalSet
		c := mustRecognize(t, r, l)
		plain := mustRecognize(t, NewNoPrecedence(tab, Options{}), l)
		before := coreItem(t, tab, 1, "X", "X", "PERLY_BRACE_CLOSE", "X")
		after := coreItem(t, tab, 2, "X", "X", "PERLY_BRACE_CLOSE", "X")
		ctx := itemCtx(r, coreItem(t, tab, 1, "X", "P", "X"))
		// Through the inserted ;, X ; from node 1 awaits the } under P's top; in the set itself it never does.
		if got := holding(tab, c.Slipped(2, next), before); !slices.ContainsFunc(got, func(it Item) bool { return it.Dist == 1 && it.Ctx == ctx }) {
			t.Fatalf("the slipper's closure at node 2 %v: X . } X awaits under %v; want from node 1 under P's top %d", c.Slipped(2, next), got, ctx)
		}
		from1 := func(c *Chart) bool {
			return slices.ContainsFunc(holding(tab, c.Sets[3].Items(), after), func(it Item) bool { return it.Dist == 2 })
		}
		if from1(c) {
			t.Errorf("node 3 %v: X } scanned from node 1 under P's top", c.Sets[3].Items())
		}
		if !from1(plain) {
			t.Errorf("no-precedence node 3 %v: X } never scanned from node 1", plain.Sets[3].Items())
		}
	})
	t.Run("carried", func(t *testing.T) {
		// With ; and } both above P, the } after the inserted ; is kept from node 1, and the item it advances keeps
		// P's top, as the direct scan's does: node 3 holds X } from node 1 under P's top and under no other context.
		tab, r := slipShiftToy(t, 3, 3)
		l := tokens(tab, "P", "A", "PERLY_BRACE_CLOSE", "A")
		c := mustRecognize(t, r, l)
		ctx := itemCtx(r, coreItem(t, tab, 1, "X", "P", "X"))
		var got []int
		for _, it := range holding(tab, c.Sets[3].Items(), coreItem(t, tab, 2, "X", "X", "PERLY_BRACE_CLOSE", "X")) {
			if it.Dist == 2 {
				got = append(got, it.Ctx)
			}
		}
		if ctx == 0 || !slices.Equal(got, []int{ctx}) || !slices.Contains(c.Inserted(), 2) {
			t.Errorf("node 3 %v: X } from node 1 under contexts %v, a ; inserted at node 2 %v; want P's top %d alone, and a ;",
				c.Sets[3].Items(), got, slices.Contains(c.Inserted(), 2), ctx)
		}
	})
}

// TestShiftRefusalThroughExpected: on S ::= Y END | Q Y T, Y ::= P X, X ::= X T X | A, with T below P, and P A END,
// X . T X awaits T at node 2 from node 1 under P's top. Q Y T puts T in the lookahead of Y ::= P X, which bison
// reduces before T, so V1 forbids that shift, and nothing else at node 2 can take a T, since this Y is followed by
// END. Expected omits T, which the no-precedence chart's Expected offers, and still offers END; the same holds with
// lookahead off and with no cache.
func TestShiftRefusalThroughExpected(t *testing.T) {
	tab, p := precToy(t, []grammar.Level{{N: 1, Assoc: "left", Tokens: []string{"T"}}, {N: 2, Assoc: "right", Tokens: []string{"P"}}}, nil, nil,
		"S ::= Y END | Q Y T", "Y ::= P X", "X ::= X T X | A")
	l := tokens(tab, "P", "A", "END")
	for _, opts := range []Options{{}, {NoLookahead: true}, {NoCache: true}} {
		r := mustJudged(t, tab, p, opts)
		c := mustRecognize(t, r, l)
		ctx := itemCtx(r, coreItem(t, tab, 1, "Y", "P", "X"))
		if !slices.ContainsFunc(holding(tab, c.Sets[2].Items(), coreItem(t, tab, 1, "X", "X", "T", "X")), func(it Item) bool {
			return it.Dist == 1 && it.Ctx == ctx && ctx != 0
		}) {
			t.Fatalf("%+v: node 2 %v: X . T X does not await T from node 1 under P's top %d", opts, c.Sets[2].Items(), ctx)
		}
		got, err := c.Expected(2)
		if err != nil {
			t.Fatal(err)
		}
		plain, err := mustRecognize(t, NewNoPrecedence(tab, opts), l).Expected(2)
		if err != nil {
			t.Fatal(err)
		}
		if slices.Contains(got, "T") || !slices.Contains(got, "END") {
			t.Errorf("%+v: Expected at node 2 is %v; want END and never T, whose shift P's top forbids", opts, got)
		}
		if !slices.Contains(plain, "T") {
			t.Errorf("%+v: the no-precedence Expected at node 2 is %v; want T, or the fixture forbids nothing", opts, plain)
		}
	}
}

// TestJudgedBranchSetsAgree: with an edge over the whole of `foo -$x * $y;` added, every node between is inside a
// branch, so its set is built by searching every item rather than through the postdot index; the shift side and
// the pairing rule judge there too, so each of those sets equals the linear lattice's, where * is refused under
// unary minus and $x * $y never completes into it.
func TestJudgedBranchSetsAgree(t *testing.T) {
	tab := committedTables(t)
	s, _ := lex.Lex([]byte("foo -$x * $y;"), tab.Tokens)
	l := lex.Linear(s)
	linear := mustRecognize(t, mustJudged(t, tab, nil, Options{}), l)
	n := len(l.Nodes)
	l.Edges = append(l.Edges, lex.Edge{From: 0, To: n - 1, Token: l.Edges[len(l.Edges)-1].Token})
	branched := mustRecognize(t, mustJudged(t, tab, nil, Options{}), l)
	for i := 1; i < n-1; i++ {
		if branched.Sets[i] == nil || branched.Sets[i].id != -1 {
			t.Fatalf("node %d is not a branch set", i)
		}
		if !slices.Equal(branched.Sets[i].Items(), linear.Sets[i].Items()) {
			t.Errorf("node %d: inside the branch %v, linear %v", i, branched.Sets[i].Items(), linear.Sets[i].Items())
		}
	}
}
