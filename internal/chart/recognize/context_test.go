// ABOUTME: Tests the context items carry in the judged recognizer: §4.3's collapse interning equal sets as one, the
// ABOUTME: transition cache keyed through the source set's contexts, and a Leo memo's top carrying its item's context.

package recognize

import (
	"slices"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

// perlyChart is the judged chart of src under r, over perly.y's tables.
func perlyChart(t *testing.T, r *Recognizer, src string) *Chart {
	t.Helper()
	s, _ := lex.Lex([]byte(src), r.tab.Tokens)
	return mustRecognize(t, r, lex.Linear(s))
}

// TestContextCollapseInternsEqualSets: the predictors after `if (`, `while (`, `foo(` and `-(` are four kernel states,
// and the last stands under unary minus's real top. None of their tops forbids a shift, so each predicts under none,
// and the sets after the 1 that follows, whose items all began after the (, are one *Set of none-context items.
func TestContextCollapseInternsEqualSets(t *testing.T) {
	tab := committedTables(t)
	r := mustJudged(t, tab, nil, Options{})
	var at2 [][]Item
	var at3 []*Set
	for _, src := range []string{"if (1 + 2) {}", "while (1 + 2) {}", "foo(1 + 2);", "-(1 + 2);"} {
		c := perlyChart(t, r, src)
		if !c.Accepted() {
			t.Fatalf("%q is refused", src)
		}
		at2, at3 = append(at2, c.Sets[2].Items()), append(at3, c.Sets[3])
	}
	for i := range at2 {
		for j := range i {
			if slices.Equal(at2[i], at2[j]) {
				t.Errorf("inputs %d and %d reach one set before the 1: %v", j, i, at2[i])
			}
		}
	}
	if !slices.ContainsFunc(at2[3], func(it Item) bool { return it.Ctx != 0 }) {
		t.Errorf("after -( no item carries unary minus's top: %v", at2[3])
	}
	for i, s := range at3 {
		if s != at3[0] {
			t.Errorf("input %d's set after the 1 is %v, input 0's %v: want one *Set", i, s.Items(), at3[0].Items())
		}
	}
	if slices.ContainsFunc(at3[0].Items(), func(it Item) bool { return it.Ctx != 0 }) {
		t.Errorf("the set after the 1 carries a context: %v", at3[0].Items())
	}
}

// TestTransitionCacheKeysOnContext: after `1 and $y->[0]` and `1 or $y->[0]` the set's states and distances are one,
// under the tops of `and` and of `or`, which judge a shift of `or` differently. The two contexts intern as two sets,
// and the transition over the + from each is its own: the second input's is not served by the entry the first
// stored, and carries its own context, while recognizing the first input again is served.
func TestTransitionCacheKeysOnContext(t *testing.T) {
	tab := committedTables(t)
	r := mustJudged(t, tab, nil, Options{})
	and, or := "1 and $y->[0] + 1;", "1 or $y->[0] + 1;"
	const node = 7 // after ]: the + is the next token
	a := perlyChart(t, r, and)
	o := perlyChart(t, r, or)
	strip := func(items []Item) (out []Item) {
		for _, it := range items {
			out = append(out, Item{State: it.State, Dist: it.Dist})
		}
		return out
	}
	ctx := func(s *Set) int { return s.Items()[0].Ctx }
	if !slices.Equal(strip(a.Sets[node].Items()), strip(o.Sets[node].Items())) {
		t.Fatalf("the sets after ] differ in states or distances: %v and %v", a.Sets[node].Items(), o.Sets[node].Items())
	}
	ca, co := ctx(a.Sets[node]), ctx(o.Sets[node])
	if p := r.judge.p; ca == 0 || co == 0 || ca == co ||
		!p.ShiftForbidden(int32(ca-1), p.Item(slices.IndexFunc(tab.Items, func(it grammar.CoreItem) bool {
			rule := tab.Rules[it.Rule]
			return it.Dot == 1 && len(rule.RHS) == 3 && tab.Symbols[rule.RHS[1]].Name == "OROP"
		})), int32(symbolID(tab, "OROP"))) {
		t.Fatalf("contexts %d and %d: want two real tops, and the first (and's) forbidding a shift of or", ca, co)
	}
	if a.Sets[node] == o.Sets[node] {
		t.Fatalf("the two contexts intern as one set: %v", a.Sets[node].Items())
	}
	if o.Cached[node+1] || o.Sets[node+1] == a.Sets[node+1] {
		t.Errorf("after the + of %q: cached %v, the set %q left there %v", or, o.Cached[node+1], and, o.Sets[node+1] == a.Sets[node+1])
	}
	if !slices.ContainsFunc(o.Sets[node+1].Items(), func(it Item) bool { return it.Ctx == co }) ||
		slices.ContainsFunc(o.Sets[node+1].Items(), func(it Item) bool { return it.Ctx == ca }) {
		t.Errorf("after the + of %q the items are %v: want or's context %d and never and's %d", or, o.Sets[node+1].Items(), co, ca)
	}
	if again := perlyChart(t, r, and); !again.Cached[node+1] || again.Sets[node+1] != a.Sets[node+1] {
		t.Errorf("%q again: the transition over the + cached %v; want it served", and, again.Cached[node+1])
	}
	cold := perlyChart(t, mustJudged(t, tab, nil, Options{NoCache: true}), or)
	for i, s := range o.Sets {
		if !slices.Equal(s.Items(), cold.Sets[i].Items()) {
			t.Errorf("%q at node %d: warm %v, no cache %v", or, i, s.Items(), cold.Sets[i].Items())
		}
	}
}

// TestLeoMemoCarriesContext: in `- 1 ** 2;` the item unary minus predicts at node 1, under its real top, is the one
// item there awaiting termbinop (** binds tighter than unary minus, so the shift side keeps its scan), so a Leo memo is made at node 1 for termbinop, and its top carries that context. Where
// the chain is taken the top item is added with it, so every item of every set is one completing link by link
// (NoLeo) adds too.
func TestLeoMemoCarriesContext(t *testing.T) {
	tab := committedTables(t)
	const src = "- 1 ** 2;"
	r := mustJudged(t, tab, nil, Options{})
	c := perlyChart(t, r, src)
	m := c.leo[[2]int{1, symbolID(tab, "termbinop")}]
	if !c.Accepted() || r.leoSteps == 0 || m == nil || m.top[2] == 0 {
		t.Fatalf("%q: accepted %v, Leo chains taken %d, the memo at node 1 for termbinop %+v: want accepted, a chain, "+
			"and a memo whose top carries a context", src, c.Accepted(), r.leoSteps, m)
	}
	top := Item{State: m.top[0], Dist: 4 - m.top[1], Ctx: m.top[2]}
	if !slices.Contains(c.Sets[4].Items(), top) {
		t.Errorf("the set after 2 %v lacks the memo's top %v", c.Sets[4].Items(), top)
	}
	plain := perlyChart(t, mustJudged(t, tab, nil, Options{NoLeo: true}), src)
	for i, s := range c.Sets {
		for _, it := range s.Items() {
			if !slices.Contains(plain.Sets[i].Items(), it) {
				t.Errorf("with Leo node %d holds %v, which link by link never adds: %v", i, it, plain.Sets[i].Items())
			}
		}
	}
}
