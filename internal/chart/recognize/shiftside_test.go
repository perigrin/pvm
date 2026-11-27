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
