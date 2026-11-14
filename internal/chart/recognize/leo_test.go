// ABOUTME: Tests Leo memos: the test that fails if Leo does nothing, set identity, left recursion and short inputs.
// ABOUTME: Also that Leo on and off agree in acceptance and in the number of distinct interned sets.

package recognize

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

// rightRec is S ::= A S | A, every prefix of a run of As a complete S: the
// right recursion that is quadratic at lookahead 0 without Leo.
var rightRec = []string{"S ::= A S | A"}

// tokens is a linear toy lattice of one-byte tokens, one terminal each.
func tokens(tab *grammar.Tables, terms ...string) *lex.Lattice[lex.None] {
	offsets := []int{0}
	var edges []lex.Edge
	for i, term := range terms {
		edges = append(edges, toyEdge(tab, i, i+1, term))
		offsets = append(offsets, i+1)
	}
	return toyLinear(offsets, edges...)
}

// items is the number of items over every set of c.
func items(c *Chart) int {
	n := 0
	for _, s := range c.Sets {
		n += len(s.Items())
	}
	return n
}

// TestLeoLinearAtLookaheadZero is the test that fails if Leo does nothing
// (§10), on S ::= A S | A at lookahead 0.
func TestLeoLinearAtLookaheadZero(t *testing.T) {
	tab := toyTables(t, rightRec...)
	checkLinear(t, tab, func(n int) *lex.Lattice[lex.None] { return tokens(tab, slices.Repeat([]string{"A"}, n)...) }, 200)
}

// TestLeoStatementSeqAtLookaheadZero: the toy statement grammar's
// right-recursive, nullable Seq, the shape of Guacamole's StatementSeq that
// interleaved mode will recognize at lookahead 0 (§5, "One coupled choice"),
// is linear with Leo too.
func TestLeoStatementSeqAtLookaheadZero(t *testing.T) {
	tab := toyStmt(t)
	checkLinear(t, tab, func(n int) *lex.Lattice[lex.None] {
		return tokens(tab, slices.Repeat([]string{"Print", "Num", "Semi"}, n)...)
	}, 100)
}

// checkLinear holds the item count, at lookahead 0, of input(n) and
// input(2n): with Leo, doubling the input at most doubles it, give or take
// a constant; without Leo it more than triples, which holds the check to
// measuring something.
func checkLinear(t *testing.T, tab *grammar.Tables, input func(int) *lex.Lattice[lex.None], n int) {
	t.Helper()
	count := func(n int, noLeo bool) int {
		c := mustRecognize(t, New(tab, Options{NoLookahead: true, NoLeo: noLeo}), input(n))
		if !c.Accepted() {
			t.Fatalf("input of %d refused", n)
		}
		return items(c)
	}
	if off1, off2 := count(n, true), count(2*n, true); off2 < 3*off1 {
		t.Fatalf("without Leo, %d items at %d and %d at %d: not quadratic, so the input tests nothing", off1, n, off2, 2*n)
	}
	if on1, on2 := count(n, false), count(2*n, false); on2 > 2*on1+8 {
		t.Errorf("with Leo, %d items at %d and %d at %d: not linear", on1, n, on2, 2*n)
	}
}

// TestLeoTopOutsideSetIdentity: with lookahead, the sets after each A but
// the last hold the same items, so they intern to one set, while their memos
// resolve to tops at different distances. The top is in the side table, not
// in the set.
func TestLeoTopOutsideSetIdentity(t *testing.T) {
	tab := toyTables(t, rightRec...)
	c := mustRecognize(t, New(tab, Options{}), tokens(tab, slices.Repeat([]string{"A"}, 5)...))
	if !c.Accepted() {
		t.Fatal("refused")
	}
	a, b := c.memoAt(2, tab.Start), c.memoAt(3, tab.Start)
	if a == nil || b == nil {
		t.Fatal("no memo at nodes 2 and 3")
	}
	if 2-a.top[1] == 3-b.top[1] {
		t.Fatalf("tops at distances %d and %d, want them to differ", 2-a.top[1], 3-b.top[1])
	}
	if c.Sets[2] != c.Sets[3] || !c.Interned[3] {
		t.Errorf("nodes 2 and 3 differ only in their tops but do not share one interned set")
	}
}

// TestLeoLeftRecursiveNoMemo: a left recursion has no Leo item, so no memo
// is made and the item count is the same with Leo on and off.
func TestLeoLeftRecursiveNoMemo(t *testing.T) {
	tab := toyTables(t, "L ::= L X | Y", "X ::= C")
	l := tokens(tab, append([]string{"Y"}, slices.Repeat([]string{"C"}, 20)...)...)
	for _, la := range []bool{false, true} {
		on := mustRecognize(t, New(tab, Options{NoLookahead: la}), l)
		off := mustRecognize(t, New(tab, Options{NoLookahead: la, NoLeo: true}), l)
		if !on.Accepted() || !off.Accepted() {
			t.Fatalf("refused: %v with Leo, %v without", on.Accepted(), off.Accepted())
		}
		for pos := range on.Sets {
			for sym := range tab.Symbols {
				if on.memoAt(pos, sym) != nil {
					t.Errorf("a memo at node %d for %s", pos, tab.Symbols[sym].Name)
				}
			}
		}
		if items(on) != items(off) {
			t.Errorf("%d items with Leo, %d without", items(on), items(off))
		}
	}
}

// TestLeoShortInputs: inputs of length 0 and 1 for S ::= A S | A are judged
// the same with Leo on and off.
func TestLeoShortInputs(t *testing.T) {
	tab := toyTables(t, rightRec...)
	for _, l := range []*lex.Lattice[lex.None]{tokens(tab), tokens(tab, "A")} {
		for _, la := range []bool{false, true} {
			on := mustRecognize(t, New(tab, Options{NoLookahead: la}), l).Accepted()
			off := mustRecognize(t, New(tab, Options{NoLookahead: la, NoLeo: true}), l).Accepted()
			if on != off || on != (len(l.Edges) == 1) {
				t.Errorf("%d tokens, lookahead off %v: accepted %v with Leo, %v without", len(l.Edges), la, on, off)
			}
		}
	}
}

// leoAgree recognizes l with Leo on and off, both with and without
// lookahead, each on a fresh recognizer, and reports any difference in
// acceptance or in the number of distinct interned sets.
func leoAgree(t testing.TB, tab *grammar.Tables, l *lex.Lattice[lex.None]) string {
	for _, la := range []bool{false, true} {
		on, off := New(tab, Options{NoLookahead: la}), New(tab, Options{NoLookahead: la, NoLeo: true})
		a, b := mustRecognize(t, on, l).Accepted(), mustRecognize(t, off, l).Accepted()
		if a != b {
			return fmt.Sprintf("lookahead off %v: accepted %v with Leo, %v without", la, a, b)
		}
		if x, y := len(on.sets.sets), len(off.sets.sets); x != y {
			return fmt.Sprintf("lookahead off %v: %d distinct sets with Leo, %d without", la, x, y)
		}
	}
	return ""
}

// TestLeoKeepsAcceptance: Leo on and off accept and reject the same inputs,
// and intern the same number of distinct sets, on toy right recursions and
// on the toy statement grammar over small cases and the fuzz seeds.
func TestLeoKeepsAcceptance(t *testing.T) {
	// The third grammar's chain runs up through S ::= V B T from node 0, a
	// start-rule item that acceptance needs to see (#549's gate, F1).
	for _, g := range [][]string{
		rightRec,
		{"S ::= A T | A", "T ::= S O", "O ::= C*"},
		{"S ::= V B T", "T ::= A | V", "V ::= A | S"},
	} {
		tab := toyTables(t, g...)
		for _, in := range [][]string{nil, {"A"}, {"A", "A", "A"}, {"A", "C", "A"}, {"A", "A", "C", "C"}, {"C"}, {"A", "A", "A", "A", "A", "A", "A", "C"}, {"A", "B", "A"}, {"A", "B", "A", "B", "A"}} {
			if d := leoAgree(t, tab, tokens(tab, in...)); d != "" {
				t.Errorf("%q on %v: %s", g, in, d)
			}
		}
	}
	stmt := toyStmt(t)
	for _, l := range []*lex.Lattice[lex.None]{
		tokens(stmt, strings.Fields("Print Print Print Num Semi")...), tokens(stmt, strings.Fields(decl+" "+decl)...),
		tokens(stmt, strings.Fields("LBrace Print Num Semi LBrace Print Num Semi RBrace RBrace")...),
		tokens(stmt, strings.Fields("Print Num Semi Semi")...), tokens(stmt, strings.Fields("Print Print Num Semi")...), tokens(stmt, strings.Fields(decl+" My Var")...),
	} {
		if d := leoAgree(t, stmt, l); d != "" {
			t.Errorf("%v: %s", l.Edges, d)
		}
	}
	for _, src := range fuzzSeeds {
		if d := leoAgree(t, stmt, fuzzLattice(stmt, src)); d != "" {
			t.Errorf("%q: %s", src, d)
		}
	}
}

// leoChainGrammar is a grammar whose memo chains read sets completion
// does not, so a warm cache serves A A B A's transitions to B A B A only
// if it keeps those dependencies.
var leoChainGrammar = []string{"S ::= U U T | N V", "T ::= A", "U ::= V | T U", "V ::= V N | B | V S S", "N ::= C*"}

// TestLeoChainReadsAreCacheDeps: a memo chain reads sets beyond the origin
// sets completion reads, and each is a transition-cache dependency. A
// recognizer warmed on A A B A caches transitions whose chains read sets B
// A B A does not have there, so without those dependencies it serves B A B A
// a wrong set (#549's gate, F2).
func TestLeoChainReadsAreCacheDeps(t *testing.T) {
	tab := toyTables(t, leoChainGrammar...)
	warm := New(tab, Options{})
	mustRecognize(t, warm, tokens(tab, "A", "A", "B", "A"))
	l := tokens(tab, "B", "A", "B", "A")
	c, cold := mustRecognize(t, warm, l), mustRecognize(t, New(tab, Options{NoCache: true}), l)
	for pos := range c.Sets {
		if !slices.Equal(c.Sets[pos].Items(), cold.Sets[pos].Items()) {
			t.Errorf("node %d: items %v with the warm cache, %v without", pos, c.Sets[pos].Items(), cold.Sets[pos].Items())
		}
	}
}

// leoMemoEntriesBound is the most stored Leo memo entries per node the
// corpus sample may leave under default options: the review prototype's
// count, rounded up, recorded with the base tree's as chart-recovery's
// "R3 bound" bullet.
const leoMemoEntriesBound = 0.036

// memoStore is the number of entries c's memo store holds, and how many of
// them are memos rather than a stored "none".
func memoStore(c *Chart) (entries, real int) {
	for _, byNode := range c.leo {
		for _, m := range byNode {
			entries++
			if m != nil {
				real++
			}
		}
	}
	return entries, real
}

// storedMemo is the entry c's memo store holds for (at, sym), if any.
func storedMemo(c *Chart, at, sym int) (*memo, bool) {
	m, ok := c.leo[at][sym]
	return m, ok
}

// TestLeoMemoEntriesBounded: memoAt stores only real memos, so the entries
// per node over corpusSample stay within leoMemoEntriesBound. A store that
// keeps every computation, "none" included, holds about one per
// computation.
func TestLeoMemoEntriesBounded(t *testing.T) {
	tab, files := corpusSample(t)
	r := New(tab, Options{})
	nodes, entries := 0, 0
	for _, f := range files {
		c := mustRecognize(t, r, f.l)
		n, _ := memoStore(c)
		nodes, entries = nodes+len(c.Sets), entries+n
	}
	per := float64(entries) / float64(nodes)
	t.Logf("%d stored Leo memo entries over %d nodes: %.4f per node", entries, nodes, per)
	if per > leoMemoEntriesBound {
		t.Errorf("%.4f stored Leo memo entries per node, above the bound %v", per, leoMemoEntriesBound)
	}
}

// TestLeoOnByDefault: Options{} recognizing a right recursion makes a memo.
func TestLeoOnByDefault(t *testing.T) {
	tab := toyTables(t, rightRec...)
	c := mustRecognize(t, New(tab, Options{}), tokens(tab, slices.Repeat([]string{"A"}, 5)...))
	if !c.Accepted() {
		t.Fatal("refused")
	}
	if _, real := memoStore(c); real == 0 {
		t.Error("Options{} made no Leo memo on S ::= A S | A")
	}
}

// TestLeoMemoOnlyWhenUnique: recognition stores a memo for (node, symbol)
// where exactly one item awaits the symbol and it is a Leo item, and stores
// nothing where two items await it or the one is not a Leo item. In each
// case completion asks memoAt at node 1 for S.
func TestLeoMemoOnlyWhenUnique(t *testing.T) {
	for _, tc := range []struct {
		name    string
		grammar []string
		in      []string
		memo    bool
	}{
		{"one Leo item", rightRec, []string{"A", "A", "A"}, true},
		// After A: S ::= A . S and, through T ::= A, S ::= T . S.
		{"two items", []string{"S ::= A S | T S | A", "T ::= A"}, []string{"A", "A"}, false},
		// After C: only S ::= C . S D, which is no Leo item, since D follows S.
		{"not a Leo item", []string{"S ::= C S D | A"}, []string{"C", "A", "D"}, false},
	} {
		tab := toyTables(t, tc.grammar...)
		c := mustRecognize(t, New(tab, Options{}), tokens(tab, tc.in...))
		if !c.Accepted() {
			t.Fatalf("%s: refused", tc.name)
		}
		m, ok := storedMemo(c, 1, tab.Start)
		if ok != tc.memo || (ok && m == nil) {
			t.Errorf("%s: node 1 stores %v (%v) for S, want a memo %v", tc.name, m, ok, tc.memo)
		}
		if entries, real := memoStore(c); entries != real {
			t.Errorf("%s: %d entries stored, only %d of them memos", tc.name, entries, real)
		}
	}
}

// rightRecLattice is n As for S ::= A S | A; with branch, also one A from
// node n/2 to n/2+2, so node n/2+1 holds a branch set, which has no postdot
// index and takes memoAt's whole-set fallback.
func rightRecLattice(tab *grammar.Tables, n int, branch bool) *lex.Lattice[lex.None] {
	l := tokens(tab, slices.Repeat([]string{"A"}, n)...)
	if branch {
		l.Edges = append(l.Edges, toyEdge(tab, n/2, n/2+2, "A"))
		slices.SortStableFunc(l.Edges, func(a, b lex.Edge) int { return a.From - b.From })
	}
	return l
}

// TestLeoDeepRightRecursionAgrees: on a right recursion 2,000 deep, plain
// and through a branch set, Leo on and off agree on acceptance and on
// Expected at every node, with and without lookahead.
func TestLeoDeepRightRecursionAgrees(t *testing.T) {
	tab := toyTables(t, rightRec...)
	for _, branch := range []bool{false, true} {
		l := rightRecLattice(tab, 2000, branch)
		for _, la := range []bool{false, true} {
			on := mustRecognize(t, New(tab, Options{NoLookahead: la}), l)
			off := mustRecognize(t, New(tab, Options{NoLookahead: la, NoLeo: true}), l)
			if on.Accepted() != off.Accepted() || !on.Accepted() {
				t.Fatalf("branch %v, lookahead off %v: accepted %v with Leo, %v without", branch, la, on.Accepted(), off.Accepted())
			}
			for pos := range l.Nodes {
				a, errA := on.Expected(pos)
				b, errB := off.Expected(pos)
				if errA != nil || errB != nil || !slices.Equal(a, b) {
					t.Fatalf("branch %v, lookahead off %v, node %d: Expected %v (%v) with Leo, %v (%v) without", branch, la, pos, a, errA, b, errB)
				}
			}
		}
	}
}

// TestNoLeoStoresNoMemo: with NoLeo, recognition through a branch set
// stores no memo and makes no whole-set scan in memoAt; with Leo on, the
// same input makes one, so the count is live.
func TestNoLeoStoresNoMemo(t *testing.T) {
	tab := toyTables(t, rightRec...)
	l := rightRecLattice(tab, 50, true)
	for _, noLeo := range []bool{true, false} {
		r := New(tab, Options{NoLeo: noLeo})
		c := mustRecognize(t, r, l)
		entries, _ := memoStore(c)
		if !c.Accepted() {
			t.Fatalf("NoLeo %v: refused", noLeo)
		}
		if noLeo && (entries != 0 || r.memoScans != 0) {
			t.Errorf("NoLeo: %d memo entries stored, %d whole-set scans", entries, r.memoScans)
		}
		if !noLeo && (entries == 0 || r.memoScans == 0) {
			t.Errorf("Leo on: %d memo entries stored, %d whole-set scans: the branch set took no fallback", entries, r.memoScans)
		}
	}
}
