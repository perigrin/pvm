// ABOUTME: Tests the judgements the context makes: a completion pairs only with the parents that predicted it under
// ABOUTME: its context, and a scan V1's shift verdict forbids under an item's context is refused (V2 design §4.1, §4.3).

package recognize

import (
	"slices"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
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
