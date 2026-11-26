// ABOUTME: Pins branches of the state-free judgement no corpus reaches, the reference's one known divergence and the nullable-shift limit.
// ABOUTME: Each toy is judged by Constrain and ConstrainReference; the comparison harness is checked for aliasing.

package forest

import (
	"slices"
	"sort"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/prectab"
	"tamarou.com/pvm/internal/chart/recognize"
)

// pinnedToy is a toy grammar's tables and precedence; statements are one rule each, rulePrec their %prec tokens.
func pinnedToy(t *testing.T, levels []grammar.Level, rulePrec []string, statements ...string) (*grammar.Tables, *prectab.Precedence) {
	t.Helper()
	tab := toyTables(t, statements...)
	p, err := prectab.NewPrecedence(tab, levels, rulePrec, nil)
	if err != nil {
		t.Fatal(err)
	}
	return tab, p
}

// TestPrecedenceReferenceKnownDivergence pins the reference's known defect (perigrin's ruling, 2026-11-24): its
// shifts forbids PLUS where E's empty reduction would win a conflict, in a state that does not shift PLUS, so
// there is no conflict. bison 3.8.2 accepts NUM PLUS NUM: its state 1 (S: NUM . A) has no shift on PLUS and
// reduces E by default. Constrain agrees with bison; the reference refuses. A reference that called Constrain
// would accept, and turns this red.
func TestPrecedenceReferenceKnownDivergence(t *testing.T) {
	const bison = true // bison 3.8.2 on the toy: ACCEPT
	tab, p := pinnedToy(t, []grammar.Level{{N: 1, Assoc: "left", Tokens: []string{"PLUS"}}, {N: 2, Assoc: "left", Tokens: []string{"HIGH"}}},
		[]string{"", "", "HIGH", ""}, "S ::= NUM A", "A ::= E PLUS NUM", "E ::= ", "Z ::= HIGH")
	f, _, _ := toyForest(t, tab, strings.Fields("NUM PLUS NUM"), recognize.Options{})
	if f == nil {
		t.Fatal("the chart refuses NUM PLUS NUM")
	}
	if got := f.Copy().Constrain(p); got != bison {
		t.Errorf("Constrain: %v, bison %v", got, bison)
	}
	if got := f.Copy().ConstrainReference(p); got {
		t.Error("ConstrainReference accepts NUM PLUS NUM; its known defect refuses it. Fixed or bypassed? Update the README and this test")
	}
}

// TestPrecedenceMarkFirstFilter: the token after Y admits A and B, and shifting A after Y is forbidden (E's empty
// reduction wins on A there), so of P's two alternatives only P: B survives. mark's first-terminal filter is what
// drops P: A, through the unit rule X: W (seq's m == 1 call) and through W: P Q's left child.
func TestPrecedenceMarkFirstFilter(t *testing.T) {
	tab, p := markFirstFilterToy(t)
	l := toyLattice(tab, []string{"C", "A", "Q1"})
	tab.TerminalSets = append(tab.TerminalSets, []int{symbol(tab, "A"), symbol(tab, "B")})
	l.Edges[1].Token.TerminalSet = len(tab.TerminalSets) - 1
	c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), l)
	if err != nil || !c.Accepted() {
		t.Fatalf("the chart refuses C {A B} Q1: %v", err)
	}
	f := Build(tab, c, l)
	pn := slices.IndexFunc(f.Nodes, func(nd Node) bool { return nd.Sym == int32(symbol(tab, "P")) })
	if pn < 0 || len(keptAlts(f, pn)) != 2 {
		t.Fatal("no P node with two alternatives")
	}
	d, ok, _ := judgeBoth(f, p)
	if d != "" || !ok {
		t.Fatalf("reference verdict %v, difference %q; want accepted, none", ok, d)
	}
	prod := f.Copy()
	prod.Constrain(p)
	if kept := keptAlts(prod, pn); len(kept) != 1 || f.Nodes[f.Alts[kept[0]].Right].Sym != int32(symbol(tab, "B")) {
		t.Errorf("P keeps %v; want only P: B", kept)
	}
}

// markFirstFilterToy is TestPrecedenceMarkFirstFilter's toy.
func markFirstFilterToy(t *testing.T) (*grammar.Tables, *prectab.Precedence) {
	return pinnedToy(t, []grammar.Level{{N: 1, Assoc: "left", Tokens: []string{"A"}}, {N: 2, Assoc: "left", Tokens: []string{"HIGH"}}},
		[]string{"", "", "", "", "", "", "", "HIGH", "", ""},
		"G ::= Y X", "Y ::= C", "X ::= W", "X ::= E A K", "W ::= P Q", "P ::= A", "P ::= B", "E ::= ", "Q ::= Q1", "K ::= K1")
}

// scancollideToy is minute row F6-C2's toy: the predicted state after P, under the spine top s: P . a, scans X for
// c: . X Q, a dot-0 item V1 never checks, and for a: n . X R, where d: n %prec HIGH wins over X and the shift is
// forbidden. One scan would have to be both refused and allowed.
func scancollideToy(t *testing.T) (*grammar.Tables, *prectab.Precedence) {
	return pinnedToy(t, []grammar.Level{{N: 1, Assoc: "left", Tokens: []string{"X"}}, {N: 2, Assoc: "left", Tokens: []string{"HIGH"}}},
		[]string{"", "", "", "", "", "HIGH", "", ""},
		"s ::= P c", "s ::= P a", "s ::= P d X W", "c ::= X Q", "a ::= n X R", "d ::= n", "n ::= ", "n ::= M")
}

// TestPrecedenceLeftParentItemDotZero: a: X %prec LOW before T is reduced after TWO and forbidden after ONE, where
// s: ONE X T shifts T, so the pair needs its parent item, and e: . a T, at dot 0 of a two-symbol rule, is the one
// that forbids it. Of ONE X T's two derivations only s: ONE X T survives, as in bison.
func TestPrecedenceLeftParentItemDotZero(t *testing.T) {
	tab, p := pinnedToy(t, []grammar.Level{{N: 1, Assoc: "left", Tokens: []string{"LOW"}}, {N: 2, Assoc: "left", Tokens: []string{"T"}}},
		[]string{"", "", "", "", "", "LOW"}, "s ::= ONE e", "s ::= ONE X T", "s ::= TWO f", "e ::= a T", "f ::= a T", "a ::= X")
	k := slices.Index(p.ReduceAt, 2)
	pi := p.Base[3] // e: . a T
	if k < 0 || !p.ParentForbid[int(p.MixAt[k])*len(p.TopAt)+int(pi)] {
		t.Fatal("a before T is not forbidden under the parent item e: . a T")
	}
	f, _, _ := toyForest(t, tab, strings.Fields("ONE X T"), recognize.Options{})
	if f == nil {
		t.Fatal("the chart refuses ONE X T")
	}
	if d, ok, _ := judgeBoth(f, p); d != "" || !ok {
		t.Fatalf("reference verdict %v, difference %q; want accepted, none", ok, d)
	}
	prod := f.Copy()
	prod.Constrain(p)
	if kept := keptAlts(prod, int(prod.Root)); len(kept) != 1 || prod.Alts[kept[0]].Rule != 1 {
		t.Errorf("the root keeps %v; want only s: ONE X T", kept)
	}
}

// TestPrecedenceShiftForbiddenPastDotTwo: after P Q, x: Q %prec HIGH reduces with T next, so the shift of T at
// s: P Q . T, rule position 2 under the start item, is forbidden. The generated tables hold that key, and of
// P Q T's two derivations only s: P x T survives, as in bison.
func TestPrecedenceShiftForbiddenPastDotTwo(t *testing.T) {
	tab, p := pinnedToy(t, []grammar.Level{{N: 1, Assoc: "left", Tokens: []string{"T"}}, {N: 2, Assoc: "left", Tokens: []string{"HIGH"}}},
		[]string{"", "", "HIGH"}, "s ::= P Q T", "s ::= P x T", "x ::= Q")
	if !p.ShiftForbidden(p.Base[0], p.Base[0]+2, int32(symbol(tab, "T"))) {
		t.Error("the tables do not forbid shifting T at s: P Q . T under s: . P Q T")
	}
	f, _, _ := toyForest(t, tab, strings.Fields("P Q T"), recognize.Options{})
	if f == nil {
		t.Fatal("the chart refuses P Q T")
	}
	if d, ok, _ := judgeBoth(f, p); d != "" || !ok {
		t.Fatalf("reference verdict %v, difference %q; want accepted, none", ok, d)
	}
	prod := f.Copy()
	prod.Constrain(p)
	if kept := keptAlts(prod, int(prod.Root)); len(kept) != 1 || prod.Alts[kept[0]].Rule != 1 {
		t.Errorf("the root keeps %v; want only s: P x T", kept)
	}
}

// TestPrecedenceHarnessesDoNotAlias: judgeBoth reports a production judgement that drops one alternative more than
// Constrain, which it could not if its two copies were one forest.
func TestPrecedenceHarnessesDoNotAlias(t *testing.T) {
	ty := loadToy(t, "ladder")
	f := toyForestOf(t, ty, "NUM PLUS NUM STAR NUM")
	old := productionConstrain
	t.Cleanup(func() { productionConstrain = old })
	productionConstrain = func(f *Forest, p *prectab.Precedence) bool {
		ok := f.Constrain(p)
		f.Nodes[f.Root].Alt = f.Alts[f.Nodes[f.Root].Alt].Next
		return ok
	}
	if d, _, _ := judgeBoth(f, ty.prec); d == "" {
		t.Error("judgeBoth reports no difference for a judgement that over-prunes the root")
	}
}

// fBelowHigh is the nullable-shift toys' levels: %left F below %left HIGH.
var fBelowHigh = []grammar.Level{{N: 1, Assoc: "left", Tokens: []string{"F"}}, {N: 2, Assoc: "left", Tokens: []string{"HIGH"}}}

// nullableShiftToy is REVIEWFIX3's toy (perigrin's ruling on REVIEWFIX2's gate G1): after TA TC X the state shifts
// F through K2: TC X . F TH and reduces E: %prec HIGH with F next, and the reduction wins, so the shift key
// (V: TC . P, P: X . Y, F) is recorded as forbidden; but under P: X . Y, F is reached only after E derives the
// empty string, and bison shifts it later, from the state after E.
func nullableShiftToy(t *testing.T) (*grammar.Tables, *prectab.Precedence) {
	return pinnedToy(t, fBelowHigh,
		[]string{"", "", "", "", "", "", "HIGH", ""},
		"S ::= TA V | TA K2", "V ::= TC P", "K2 ::= TC X F TH", "P ::= X Y", "Y ::= E F", "E ::= ", "X ::= TX")
}

// nullableShiftKeys is p's shift keys of the known limit's shape, as "f after item under top". The shape, in the
// generated automaton: a shift key (spine-top item, rule-position item A: α . Bβ, terminal f) whose bit is set in
// shiftForbid, so every state holding the key shifts f and forbids it there (a state that does not shift f counts
// as allowed), where f is in FIRST(B) only after a nullable symbol: some derivation B ⇒* N γ, through first
// symbols of any number of rules, has every symbol of N derive ε, N not empty, and f in FIRST(γ). That f is shifted
// in a later state, after the empty reductions, not in the one the key was judged in. A terminal at the dot has no
// such f. rules are the automaton's, the tables' then the refused ones'; FIRST and nullability are the tables', as
// generation's are.
func nullableShiftKeys(p *prectab.Precedence, rules []grammar.BNFRule) []string {
	tab := p.Tab
	after := make([][]bool, len(tab.Symbols)) // after[x][f]: f begins some derivation of x after a nullable prefix
	for x := range after {
		after[x] = make([]bool, p.NT)
	}
	for changed := true; changed; {
		changed = false
		for _, r := range tab.Rules {
			for i, y := range r.RHS {
				for f := range p.NT {
					// past position 0 every symbol before y derived ε; at 0, whatever y itself reaches that way
					if !after[r.LHS][f] && (i > 0 && slices.Contains(tab.First[y], f) || i == 0 && after[y][f]) {
						after[r.LHS][f], changed = true, true
					}
				}
				if !tab.Symbols[y].Nullable {
					break
				}
			}
		}
	}
	text := func(r, dot int) string {
		w := []string{tab.Symbols[rules[r].LHS].Name + ":"}
		for d, s := range rules[r].RHS {
			if d == dot {
				w = append(w, ".")
			}
			w = append(w, tab.Symbols[s].Name)
		}
		return strings.Join(w, " ")
	}
	at := func(i int) (r, dot int) { // item i's rule and dot: base is ascending, a rule's items consecutive
		r = sort.Search(len(p.Base), func(r int) bool { return int(p.Base[r]) > i }) - 1
		return r, i - int(p.Base[r])
	}
	ni := len(p.TopAt)
	var out []string
	for top := range ni {
		if p.TopAt[top] < 0 {
			continue
		}
		for i := range ni {
			set := p.ShiftAt[int(p.TopAt[top])*ni+i]
			if set < 0 {
				continue
			}
			r, j := at(i)
			for f := range p.NT {
				if p.ShiftForbid[set][f/64]&(1<<(f%64)) != 0 && after[rules[r].RHS[j]][f] {
					tr, tj := at(top)
					out = append(out, tab.Symbols[f].Name+" after "+text(r, j)+" under "+text(tr, tj))
				}
			}
		}
	}
	return out
}

// TestPrecedenceNullableShiftLimitAbsent counts the known limit's keys (nullableShiftKeys): none on perly.y, whose
// automaton holds its four refused rules, and at least one on its toy, so the count is shown to see the shape.
func TestPrecedenceNullableShiftLimitAbsent(t *testing.T) {
	tab := committedTables(t)
	p, err := prectab.ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	g, _, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	rules := slices.Clone(tab.Rules)
	for _, r := range g.Refused {
		rule := grammar.BNFRule{LHS: symbol(tab, r.LHS)}
		for _, s := range r.RHS {
			rule.RHS = append(rule.RHS, symbol(tab, s.Text))
		}
		rules = append(rules, rule)
	}
	if len(rules) != len(p.Base) {
		t.Fatalf("%d rules for the automaton's %d", len(rules), len(p.Base))
	}
	if got := nullableShiftKeys(p, rules); len(got) != 0 {
		t.Errorf("perly.y: %d keys of the limit's shape, want 0: %q", len(got), got)
	}
	toyTab, toy := nullableShiftToy(t)
	got := nullableShiftKeys(toy, toyTab.Rules)
	if !slices.Contains(got, "F after P: X . Y under V: TC . P") {
		t.Errorf("toy: %q; want F after P: X . Y under V: TC . P among them", got)
	}
	// A chain: F is reached under P: X . Z through Z: Y, Y: E F and E: E1 E2, two empty rules deep.
	chainTab, chain := pinnedToy(t, fBelowHigh,
		[]string{"", "", "", "", "", "", "", "", "HIGH", "", ""},
		"S ::= TA V | TA K2", "V ::= TC P", "K2 ::= TC X F TH", "P ::= X Z", "Z ::= Y", "Y ::= E F", "E ::= E1 E2", "E1 ::= ", "E2 ::= ", "X ::= TX")
	if got := nullableShiftKeys(chain, chainTab.Rules); !slices.Contains(got, "F after P: X . Z under V: TC . P") {
		t.Errorf("chain toy: %q; want F after P: X . Z under V: TC . P among them", got)
	}
	t.Logf("perly.y keys=0 toy keys=%d", len(got))
}

// TestPrecedenceNullableShiftKnownLimit pins the known limit (perigrin's ruling, 2026-11-24, "Check perly.y, then
// pin"): on nullableShiftToy, bison 3.8.2 accepts TA TC TX F. Its state 9 (K2: TC X . F TH, P: X . Y, Y: . E F,
// E: %empty . [F]) resolves the conflict between E and F as reduce (F < HIGH), then shifts F from state 11
// (Y: E . F). Constrain and ConstrainReference both refuse it, because the shift key (V: TC . P, P: X . Y, F) is
// forbidden. perly.y holds no key of the shape (TestPrecedenceNullableShiftLimitAbsent). A fix that accepts it
// turns this red: update the README and this test.
func TestPrecedenceNullableShiftKnownLimit(t *testing.T) {
	const bison = true // bison 3.8.2 on the toy: ACCEPT
	tab, p := nullableShiftToy(t)
	f, _, _ := toyForest(t, tab, strings.Fields("TA TC TX F"), recognize.Options{})
	if f == nil {
		t.Fatal("the chart refuses TA TC TX F")
	}
	for name, judge := range map[string]func(*Forest, *prectab.Precedence) bool{"Constrain": (*Forest).Constrain, "ConstrainReference": (*Forest).ConstrainReference} {
		if got := judge(f.Copy(), p); got == bison {
			t.Errorf("%s accepts TA TC TX F, as bison does; the known limit is fixed or bypassed: update the README and this test", name)
		}
	}
}
