// ABOUTME: Tests Expected's per-set memo and the Leo on/off comparison that the ratchet's leo-disagree counts.
// ABOUTME: Skews are made by editing one chart's sets, so each position the comparison looks at is shown to matter.

package recognize

import (
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/lex"
)

// TestExpectedMemoizedPerSet: a second Expected at a position does no
// recomputation, counted rather than timed, and answers the same.
func TestExpectedMemoizedPerSet(t *testing.T) {
	r := New(toyStmt(t), Options{})
	c := run(t, r, decl+" Print Var Semi")
	for pos := range c.Sets {
		first, err := c.Expected(pos)
		if err != nil {
			t.Fatal(err)
		}
		before := r.expectedComputed
		again, _ := c.Expected(pos)
		if r.expectedComputed != before || !slices.Equal(first, again) {
			t.Errorf("node %d: second Expected computed %d more, %v then %v", pos, r.expectedComputed-before, first, again)
		}
	}
	// A new chart over the same input meets the same sets and origins.
	c2 := run(t, r, decl+" Print Var Semi")
	before := r.expectedComputed
	for pos := range c2.Sets {
		_, _ = c2.Expected(pos)
	}
	if r.expectedComputed != before {
		t.Errorf("a second chart over the same input computed Expected %d more times", r.expectedComputed-before)
	}
}

// TestExpectedMemoDistinguishesOrigins: `p a b` and `q a b` end in one
// interned set, but completing E reads the set where E began, so the first
// expects x and the second y. The memo answers each its own, and a repeat
// of either is served from it.
func TestExpectedMemoDistinguishesOrigins(t *testing.T) {
	tab := toyTables(t, "S ::= P E X | Q E Y", "E ::= A B")
	r := New(tab, Options{})
	p, q := run(t, r, "P A B"), run(t, r, "Q A B")
	if p.Sets[3] != q.Sets[3] {
		t.Fatal("the premise failed: node 3's set differs between the inputs")
	}
	for _, tc := range []struct {
		c    *Chart
		want []string
	}{{p, []string{"X"}}, {q, []string{"Y"}}, {p, []string{"X"}}, {q, []string{"Y"}}} {
		if got, _ := tc.c.Expected(3); !slices.Equal(got, tc.want) {
			t.Errorf("Expected(3) = %v, want %v", got, tc.want)
		}
	}
	if r.expectedComputed != 2 {
		t.Errorf("computed %d times, want 2: once per origin", r.expectedComputed)
	}
}

// onOff is a lattice's charts with Leo on and off.
func onOff(t *testing.T, src string, rules ...string) (on, off *Chart) {
	t.Helper()
	tab := toyTables(t, rules...)
	l := tokens(tab, strings.Fields(src)...)
	return mustRecognize(t, New(tab, Options{}), l), mustRecognize(t, New(tab, Options{NoLeo: true}), l)
}

// TestLeoDisagreeComparesExpected: Leo on and off agree on a right
// recursion, in acceptance and in Expected; Expected skewed on one side is
// a disagreement.
func TestLeoDisagreeComparesExpected(t *testing.T) {
	tab := toyTables(t, rightRec...)
	l := tokens(tab, "A", "A", "A")
	if d, err := LeoDisagree(tab, l); d || err != nil {
		t.Fatalf("LeoDisagree = %v, %v; want false, nil", d, err)
	}
	on, off := onOff(t, "A A A", rightRec...)
	if chartsDisagree(on, off) {
		t.Fatal("unskewed charts disagree")
	}
	off.Sets[1] = nil
	if !chartsDisagree(on, off) {
		t.Error("Expected skewed at node 1 went unreported")
	}
}

// TestLeoDisagreeEveryPosition: a skew at the first position and one at
// the last are each reported, and so is acceptance differing alone.
func TestLeoDisagreeEveryPosition(t *testing.T) {
	for _, pos := range []int{0, 3} {
		on, off := onOff(t, "A A A", rightRec...)
		off.Sets[pos] = nil
		if !chartsDisagree(on, off) {
			t.Errorf("a skew at node %d went unreported", pos)
		}
	}
	on, off := onOff(t, "A", "S ::= A | A B")
	last := len(off.Sets) - 1
	// The accepting item moved off its origin: still expecting B, no longer accepting.
	moved := slices.Clone(off.Sets[last].Items())
	for i := range moved {
		moved[i].Dist = 0
	}
	off.Sets[last] = &Set{id: -1, items: moved}
	if on.Accepted() == off.Accepted() {
		t.Fatal("the premise failed: acceptance does not differ")
	}
	if a, _ := on.Expected(last); !slices.Equal(a, expectedOf(t, off, last)) {
		t.Fatal("the premise failed: Expected differs, so acceptance is not alone")
	}
	if !chartsDisagree(on, off) {
		t.Error("acceptance differing alone went unreported")
	}
}

func expectedOf(t *testing.T, c *Chart, pos int) []string {
	t.Helper()
	e, err := c.Expected(pos)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// TestLeoDisagreeInvalidLatticeErrors: a lattice that cannot be run is an
// error and not "no disagreement"; a valid one is neither.
func TestLeoDisagreeInvalidLatticeErrors(t *testing.T) {
	tab := toyTables(t, rightRec...)
	bad := &lex.Lattice[lex.None]{Nodes: make([]lex.Node[lex.None], 3), Edges: []lex.Edge{{From: 2, To: 1}}}
	if d, err := LeoDisagree(tab, bad); err == nil || d {
		t.Errorf("invalid lattice: %v, %v; want false and an error", d, err)
	}
	if d, err := LeoDisagree(tab, tokens(tab, "A", "A")); err != nil || d {
		t.Errorf("valid lattice: %v, %v; want false, nil", d, err)
	}
}
