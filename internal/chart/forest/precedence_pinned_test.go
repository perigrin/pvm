// ABOUTME: Pins branches of the state-free judgement no corpus reaches, and the reference's one known divergence.
// ABOUTME: Each toy is judged by Constrain and ConstrainReference; the comparison harness is checked for aliasing.

package forest

import (
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/recognize"
)

// pinnedToy is a toy grammar's tables and precedence; statements are one rule each, rulePrec their %prec tokens.
func pinnedToy(t *testing.T, levels []grammar.Level, rulePrec []string, statements ...string) (*grammar.Tables, *Precedence) {
	t.Helper()
	tab := toyTables(t, statements...)
	p, err := NewPrecedence(tab, levels, rulePrec, nil)
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
	tab, p := pinnedToy(t, []grammar.Level{{N: 1, Assoc: "left", Tokens: []string{"A"}}, {N: 2, Assoc: "left", Tokens: []string{"HIGH"}}},
		[]string{"", "", "", "", "", "", "", "HIGH", "", ""},
		"G ::= Y X", "Y ::= C", "X ::= W", "X ::= E A K", "W ::= P Q", "P ::= A", "P ::= B", "E ::= ", "Q ::= Q1", "K ::= K1")
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

// TestPrecedenceLeftParentItemDotZero: a: X %prec LOW before T is reduced after TWO and forbidden after ONE, where
// s: ONE X T shifts T, so the pair needs its parent item, and e: . a T, at dot 0 of a two-symbol rule, is the one
// that forbids it. Of ONE X T's two derivations only s: ONE X T survives, as in bison.
func TestPrecedenceLeftParentItemDotZero(t *testing.T) {
	tab, p := pinnedToy(t, []grammar.Level{{N: 1, Assoc: "left", Tokens: []string{"LOW"}}, {N: 2, Assoc: "left", Tokens: []string{"T"}}},
		[]string{"", "", "", "", "", "LOW"}, "s ::= ONE e", "s ::= ONE X T", "s ::= TWO f", "e ::= a T", "f ::= a T", "a ::= X")
	k := slices.Index(p.reduceAt, 2)
	pi := p.base[3] // e: . a T
	if k < 0 || !p.parentForbid[int(p.mixAt[k])*len(p.topAt)+int(pi)] {
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
	if !p.shiftForbidden(p.base[0], p.base[0]+2, int32(symbol(tab, "T"))) {
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
	productionConstrain = func(f *Forest, p *Precedence) bool {
		ok := f.Constrain(p)
		f.Nodes[f.Root].Alt = f.Alts[f.Nodes[f.Root].Alt].Next
		return ok
	}
	if d, _, _ := judgeBoth(f, ty.prec); d == "" {
		t.Error("judgeBoth reports no difference for a judgement that over-prunes the root")
	}
}
