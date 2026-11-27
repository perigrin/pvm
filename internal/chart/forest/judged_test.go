// ABOUTME: Tests the judged recognizer's forest with no Constrain: Build walking by context, and reading no completion
// ABOUTME: of a rule the recognizer refused into every parent.

package forest

import (
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/prectab"
	"tamarou.com/pvm/internal/chart/recognize"
)

// deadContextToy is S ::= X END1 | Q X END2, X ::= X T X | C T A | P X | A, with T below P: P binds tighter than T. After a token that is both P and Q, X is predicted under P's real top, which forbids the shift of T after
// X . T X, and under none, Q's, which forbids nothing; C T A shifts its T under either.
func deadContextToy(t *testing.T) (*grammar.Tables, *prectab.Precedence) {
	return pinnedToy(t, []grammar.Level{{N: 1, Assoc: "left", Tokens: []string{"T"}}, {N: 2, Assoc: "right", Tokens: []string{"P"}}},
		make([]string, 6), "S ::= X END1 | Q X END2", "X ::= X T X | C T A | P X | A")
}

// deadContextLattice is {P Q} {A C} T A END1: the two sets are appended to tab's.
func deadContextLattice(tab *grammar.Tables) *lex.Lattice[lex.None] {
	l := toyLattice(tab, strings.Fields("P A T A END1"))
	for i, set := range map[int][]string{0: {"P", "Q"}, 1: {"A", "C"}} {
		ids := []int{symbol(tab, set[0]), symbol(tab, set[1])}
		slices.Sort(ids)
		l.Edges[i].Token.TerminalSet = len(tab.TerminalSets)
		tab.TerminalSets = append(tab.TerminalSets, ids)
	}
	return l
}

// TestBuildDropsAlternativeOfDeadContext: on deadContextToy's {P Q} {A C} T A END1, X over 1 to 4 completes under
// Q's context, none, as X T X and as C T A, and under P's real top as C T A only, the shift of T after X . T X
// refused. Q's parent chain dies at END1, so the node is reached only under P's top, and V1 keeps C T A there alone.
// Build walks by context: it attaches C T A and never X T X, the alternative completed under the context whose
// parent chain died, and the forest equals V1's.
func TestBuildDropsAlternativeOfDeadContext(t *testing.T) {
	tab, p := deadContextToy(t)
	if _, err := prectab.Feasible(p); err != nil {
		t.Fatalf("the toy is not feasible: %v", err)
	}
	l := deadContextLattice(tab)
	r, err := recognize.NewJudged(tab, p, recognize.Options{})
	if err != nil {
		t.Fatal(err)
	}
	c, err := recognize.Recognize(r, l)
	if err != nil || !c.Accepted() {
		t.Fatalf("the judged chart refuses {P Q} {A C} T A END1 (%v)", err)
	}
	rule := func(first string) int {
		return slices.IndexFunc(tab.Rules, func(r grammar.BNFRule) bool { return len(r.RHS) == 3 && tab.Symbols[r.RHS[0]].Name == first })
	}
	xtx, cta := rule("X"), rule("C")
	// The chart completes X over 1 to 4 as X T X under one context only, and as C T A under two.
	ctxs := func(rule int) (out []int) {
		for _, it := range c.Sets[4].Items() {
			if it.Dist == 3 && slices.Contains(tab.States[it.State].Reduces, rule) {
				out = append(out, it.Ctx)
			}
		}
		return out
	}
	t.Logf("X over 1 to 4: X T X under contexts %v, C T A under %v", ctxs(xtx), ctxs(cta))
	if len(ctxs(xtx)) != 1 || len(ctxs(cta)) != 2 || !slices.Contains(ctxs(cta), ctxs(xtx)[0]) {
		t.Fatalf("X over 1 to 4: X T X under contexts %v, C T A under %v; want X T X under one, C T A under it and another",
			ctxs(xtx), ctxs(cta))
	}
	v1, err := V1ToyForest(tab, l, p)
	if err != nil || v1 == nil {
		t.Fatalf("V1 refuses {P Q} {A C} T A END1 (%v)", err)
	}
	f := Build(tab, c, l)
	x := NodeKey{Sym: int32(symbol(tab, "X")), Rule: -1, Dot: 0, From: 1, To: 4}
	rules := func(f *Forest) (out []int32) {
		alts, _ := reachableAlts(f)
		for _, a := range alts[x] {
			out = append(out, a.Rule)
		}
		return out
	}
	if got, want := rules(v1), []int32{int32(cta)}; !slices.Equal(got, want) {
		t.Fatalf("V1 keeps rules %v on X over 1 to 4; want C T A's %v alone", got, want)
	}
	if got := rules(f); !slices.Equal(got, []int32{int32(cta)}) {
		t.Errorf("Build attaches rules %v to X over 1 to 4; want C T A's %d alone, never X T X's %d", got, cta, xtx)
	}
	if d := ForestDiff(f, v1); d != "" {
		t.Errorf("the judged forest differs from V1's: %s", d)
	}
}

// TestBuildReadsNoRefusedRule: in `_ - $h{1} - $fh;` the term over `_ - $h{1}` before the second - completes as
// termbinop and as UNIOPSUB term, `_(- $h{1})`, whose reduction V1 forbids before ADDOP, the one terminal of that
// token that can follow a term. The judged recognizer refuses that completion into every parent and records it, its
// item still standing in the set beside termbinop's, which completes the same term from the same node; Build reads
// no completion of the refused rule, so the term keeps termbinop alone, as V1 does.
func TestBuildReadsNoRefusedRule(t *testing.T) {
	tab := committedTables(t)
	const src = "_ - $h{1} - $fh;"
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	l := lex.Linear(s)
	r, err := recognize.NewJudged(tab, nil, recognize.Options{})
	if err != nil {
		t.Fatal(err)
	}
	c, err := recognize.Recognize(r, l)
	if err != nil || !c.Accepted() {
		t.Fatalf("%q: the judged chart refuses it (%v)", src, err)
	}
	uniop := slices.IndexFunc(tab.Rules, func(r grammar.BNFRule) bool {
		return len(r.RHS) == 2 && tab.Symbols[r.RHS[0]].Name == "UNIOPSUB" && tab.Symbols[r.LHS].Name == "term"
	})
	term := symbol(tab, "term")
	if !slices.ContainsFunc(c.Sets[6].Items(), func(it recognize.Item) bool {
		return it.Dist == 6 && slices.Contains(tab.States[it.State].Reduces, uniop)
	}) || !c.RuleRefused(6, uniop, term, 0) {
		t.Fatalf("node 6 %v: want an item completing UNIOPSUB term from node 0, its completion refused into every parent (%v)",
			c.Sets[6].Items(), c.RuleRefused(6, uniop, term, 0))
	}
	key := NodeKey{Sym: int32(term), Rule: -1, Dot: 0, From: 0, To: 6}
	alts, _ := reachableAlts(Build(tab, c, l))
	if len(alts[key]) == 0 || slices.ContainsFunc(alts[key], func(a altKey) bool { return int(a.Rule) == uniop }) {
		t.Errorf("Build keeps %+v on the term over 0 to 6; want some, and never UNIOPSUB term's rule %d", alts[key], uniop)
	}
	v1, _ := V1Forest(tab, []byte(src), Scope{})
	if d := ForestDiff(Build(tab, c, l), v1); d != "" {
		t.Errorf("%q: the judged forest differs from V1's: %s", src, d)
	}
}
