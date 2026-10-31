// ABOUTME: Tests the recognizer's input contract: tokens as terminal-set ids, multi-terminal sets, lattice validation.
// ABOUTME: Also that the per-token class lookup is an integer key that allocates nothing.

package recognize

import (
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

// setLattice is a linear lattice of one token per set id given.
func setLattice(sets ...int) *lex.Lattice[lex.None] {
	l := &lex.Lattice[lex.None]{Nodes: []lex.Node[lex.None]{{}}}
	for i, s := range sets {
		l.Nodes = append(l.Nodes, lex.Node[lex.None]{Offset: i + 1})
		l.Edges = append(l.Edges, lex.Edge{From: i, To: i + 1, Token: lex.Token{TerminalSet: s}})
	}
	return l
}

// addSet adds a set of the named terminals to tab and is its id.
func addSet(tab *grammar.Tables, names ...string) int {
	var ids []int
	for _, n := range names {
		ids = append(ids, tab.TerminalSets[termSet(tab, n)][0])
	}
	slices.Sort(ids)
	tab.TerminalSets = append(tab.TerminalSets, ids)
	return len(tab.TerminalSets) - 1
}

// TestRecognizeTakesSetIDs: tokens given only as set ids are recognized as
// their terminal names were, and a token whose set holds two terminals is
// accepted exactly where one of its terminals in its place would be.
func TestRecognizeTakesSetIDs(t *testing.T) {
	tab := toyStmt(t)
	varNum := addSet(tab, "Var", "Num")
	r := New(tab, Options{})
	// ids is src's set ids, with _ the Var-or-Num token.
	ids := func(src string) []int {
		var out []int
		for _, w := range strings.Fields(src) {
			if w == "_" {
				out = append(out, varNum)
			} else {
				out = append(out, termSet(tab, w))
			}
		}
		return out
	}
	for _, src := range []string{"", "Num Num", "Print Num Semi", decl + " Print Var Semi", "My Var Assign"} {
		if got, want := mustRecognize(t, r, setLattice(ids(src)...)).Accepted(), run(t, New(toyStmt(t), Options{}), src).Accepted(); got != want {
			t.Errorf("%q as set ids: accepted %v, by name %v", src, got, want)
		}
	}
	for src, want := range map[string]bool{"Print _ Semi": true, "My _ Assign Num Semi": true, "My Var Assign _ Add _ Semi": true, "Print _ _ Semi": false, "Print Var Add _ Semi": true, "_ _": false} {
		if got := mustRecognize(t, r, setLattice(ids(src)...)).Accepted(); got != want {
			t.Errorf("%q with _ as {Var, Num}: accepted %v, want %v", src, got, want)
		}
	}
}

// TestRecognizeValidatesLattice: a backward edge, an edge to a node outside
// the lattice and an out-of-range set id each return an error, not a panic.
func TestRecognizeValidatesLattice(t *testing.T) {
	tab := toyStmt(t)
	r := New(tab, Options{})
	num := termSet(tab, "Num")
	for name, l := range map[string]*lex.Lattice[lex.None]{
		"backward edge":       {Nodes: make([]lex.Node[lex.None], 3), Edges: []lex.Edge{{From: 2, To: 1, Token: lex.Token{TerminalSet: num}}}},
		"node past the end":   {Nodes: make([]lex.Node[lex.None], 2), Edges: []lex.Edge{{From: 0, To: 2, Token: lex.Token{TerminalSet: num}}}},
		"negative node":       {Nodes: make([]lex.Node[lex.None], 2), Edges: []lex.Edge{{From: -1, To: 1, Token: lex.Token{TerminalSet: num}}}},
		"set id out of range": setLattice(len(tab.TerminalSets) + 5),
		"negative set id":     setLattice(-1),
	} {
		if c, err := Recognize(r, l); err == nil {
			t.Errorf("%s: no error (chart %v)", name, c)
		}
	}
}

// TestRecognizeValidatesLatticeBoundaries: a set id equal to the table's
// length and a zero-length edge are errors; the last valid set id and an
// empty lattice are not, and the grammar decides the empty lattice.
func TestRecognizeValidatesLatticeBoundaries(t *testing.T) {
	tab := toyStmt(t)
	r := New(tab, Options{})
	if _, err := Recognize(r, setLattice(len(tab.TerminalSets))); err == nil {
		t.Error("a set id equal to the table's length: no error")
	}
	zero := &lex.Lattice[lex.None]{Nodes: make([]lex.Node[lex.None], 2), Edges: []lex.Edge{{From: 1, To: 1, Token: lex.Token{TerminalSet: 0}}}}
	if _, err := Recognize(r, zero); err == nil {
		t.Error("a zero-length edge: no error")
	}
	if _, err := Recognize(r, setLattice(len(tab.TerminalSets)-1)); err != nil {
		t.Errorf("the last valid set id: %v", err)
	}
	for name, tc := range map[string]struct {
		tab  *grammar.Tables
		want bool
	}{"nullable Program": {tab, true}, "S ::= A S | A": {toyTables(t, rightRec...), false}} {
		for _, l := range []*lex.Lattice[lex.None]{{}, setLattice()} {
			c, err := Recognize(New(tc.tab, Options{}), l)
			if err != nil {
				t.Errorf("%s, %d nodes: the empty lattice was reported invalid: %v", name, len(l.Nodes), err)
				continue
			}
			if want := tc.want && len(l.Nodes) > 0; c.Accepted() != want {
				t.Errorf("%s, %d nodes: accepted %v, want %v", name, len(l.Nodes), c.Accepted(), want)
			}
		}
	}
}

// TestRecognizeSetIDsRefusesWhenNoneExpected: after `Print` a term is
// expected, so a {Semi, RBrace} token there is refused and Expected at that
// node names the terms; after `Print Num` exactly one of them, Semi, is
// expected, and the token is accepted.
func TestRecognizeSetIDsRefusesWhenNoneExpected(t *testing.T) {
	tab := toyStmt(t)
	closer := addSet(tab, "Semi", "RBrace")
	r := New(tab, Options{})
	c := mustRecognize(t, r, setLattice(termSet(tab, "Print"), closer))
	if c.Accepted() || c.Sets[2] != nil {
		t.Errorf("Print {Semi, RBrace}: accepted %v, set after it %v; want refused there", c.Accepted(), c.Sets[2])
	}
	exp, err := c.Expected(1)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(exp, "Var") || !slices.Contains(exp, "Num") || slices.Contains(exp, "Semi") || slices.Contains(exp, "RBrace") {
		t.Errorf("Expected after Print = %v, want Var and Num, neither Semi nor RBrace", exp)
	}
	if !mustRecognize(t, r, setLattice(termSet(tab, "Print"), termSet(tab, "Num"), closer)).Accepted() {
		t.Error("Print Num {Semi, RBrace} refused, want accepted through Semi")
	}
}

// TestClassKeysAllocationFree: the lookahead class of a node, from the set
// ids leaving it or the end of input, is an integer key and its prebuilt
// bitset, found without allocating.
func TestClassKeysAllocationFree(t *testing.T) {
	tab := toyStmt(t)
	r := New(tab, Options{})
	semi := termSet(tab, "Semi")
	out := []int{semi}
	key, eofKey := -2, -2
	var bits, eofBits bitset
	if n := testing.AllocsPerRun(100, func() {
		key, bits = r.lookahead(out)
		eofKey, eofBits = r.lookahead(nil)
	}); n != 0 {
		t.Errorf("the class lookup allocated %v times per run, want 0", n)
	}
	if key != semi || !slices.Equal(bits, r.bits[semi]) {
		t.Errorf("lookahead({Semi}) = %d, want the set id %d and its bitset", key, semi)
	}
	if eofKey != len(tab.TerminalSets) || eofBits == nil || eofBits[tab.EOF/64]&(1<<(tab.EOF%64)) == 0 {
		t.Errorf("lookahead at the end = %d, want the end-of-input class %d holding EOF", eofKey, len(tab.TerminalSets))
	}
}

// TestRecognizeSetAddedAfterNew: New keeps its own copy of the set table,
// so sets added to the tables afterwards are outside that recognizer's
// table: an error, never a panic, and its verdicts on known sets unchanged.
func TestRecognizeSetAddedAfterNew(t *testing.T) {
	tab := toyStmt(t)
	r := New(tab, Options{})
	prt, num, semi := termSet(tab, "Print"), termSet(tab, "Num"), termSet(tab, "Semi")
	added := []int{addSet(tab, "Semi"), addSet(tab, "Semi", "RBrace")}
	for _, s := range added {
		if _, err := Recognize(r, setLattice(prt, num, s)); err == nil {
			t.Errorf("set %d, added after New: no error", s)
		}
	}
	if !mustRecognize(t, r, setLattice(prt, num, semi)).Accepted() {
		t.Error("Print Num Semi refused after sets were added to the tables")
	}
}
