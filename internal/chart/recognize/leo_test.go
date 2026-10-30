// ABOUTME: Tests Leo memos: the test that fails if Leo does nothing, set identity, left recursion and short inputs.
// ABOUTME: Also that Leo on and off agree in acceptance and in the number of distinct interned sets.

package recognize

import (
	"fmt"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

// rightRec is S ::= A S | A, every prefix of a run of As a complete S: the
// right recursion that is quadratic at lookahead 0 without Leo.
const rightRec = "S ::= A S | A\nA ~ 'a'\n"

// tokens is a linear toy lattice of one-byte tokens, one terminal each.
func tokens(terms ...string) *lex.Lattice[lex.None] {
	offsets := []int{0}
	var edges []lex.Edge
	for i, term := range terms {
		edges = append(edges, toyEdge(i, i+1, term, i, i+1))
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
	tab := toyTables(t, rightRec)
	checkLinear(t, tab, func(n int) *lex.Lattice[lex.None] { return tokens(slicesRepeat("A", n)...) }, 200)
}

// TestLeoStatementSeqAtLookaheadZero: Guacamole's right-recursive
// StatementSeq, which interleaved mode will recognize at lookahead 0 (§5,
// "One coupled choice"), is linear with Leo too.
func TestLeoStatementSeqAtLookaheadZero(t *testing.T) {
	_, x := guacamole(t)
	checkLinear(t, guacTab, func(n int) *lex.Lattice[lex.None] {
		s, _ := x.Lex([]byte(strings.Repeat("1;\n", n)))
		return lex.Linear(s)
	}, 100)
}

// checkLinear holds the item count, at lookahead 0, of input(n) and
// input(2n): with Leo, doubling the input at most doubles it, give or take
// a constant; without Leo it more than triples, which holds the check to
// measuring something.
func checkLinear(t *testing.T, tab *grammar.Tables, input func(int) *lex.Lattice[lex.None], n int) {
	t.Helper()
	count := func(n int, noLeo bool) int {
		c := Recognize(New(tab, Options{NoLookahead: true, NoLeo: noLeo}), input(n))
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

func slicesRepeat(s string, n int) []string {
	return strings.Fields(strings.Repeat(s+" ", n))
}

// TestLeoTopOutsideSetIdentity: with lookahead, the sets after each A but
// the last hold the same items, so they intern to one set, while their memos
// resolve to tops at different distances. The top is in the side table, not
// in the set.
func TestLeoTopOutsideSetIdentity(t *testing.T) {
	tab := toyTables(t, rightRec)
	c := Recognize(New(tab, Options{}), tokens(slicesRepeat("A", 5)...))
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
	tab := toyTables(t, "L ::= L X | Y\nX ::= C\nC ~ 'c'\nY ~ 'y'\n")
	l := tokens(append([]string{"Y"}, slicesRepeat("C", 20)...)...)
	for _, la := range []bool{false, true} {
		on := Recognize(New(tab, Options{NoLookahead: la}), l)
		off := Recognize(New(tab, Options{NoLookahead: la, NoLeo: true}), l)
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
	tab := toyTables(t, rightRec)
	for _, l := range []*lex.Lattice[lex.None]{tokens(), tokens("A")} {
		for _, la := range []bool{false, true} {
			on := Recognize(New(tab, Options{NoLookahead: la}), l).Accepted()
			off := Recognize(New(tab, Options{NoLookahead: la, NoLeo: true}), l).Accepted()
			if on != off || on != (len(l.Edges) == 1) {
				t.Errorf("%d tokens, lookahead off %v: accepted %v with Leo, %v without", len(l.Edges), la, on, off)
			}
		}
	}
}

// leoAgree recognizes l with Leo on and off, both with and without
// lookahead, each on a fresh recognizer, and reports any difference in
// acceptance or in the number of distinct interned sets.
func leoAgree(tab *grammar.Tables, l *lex.Lattice[lex.None]) string {
	for _, la := range []bool{false, true} {
		on, off := New(tab, Options{NoLookahead: la}), New(tab, Options{NoLookahead: la, NoLeo: true})
		a, b := Recognize(on, l).Accepted(), Recognize(off, l).Accepted()
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
// on Guacamole over small cases and the fuzz seeds.
func TestLeoKeepsAcceptance(t *testing.T) {
	for _, g := range []string{rightRec, "S ::= A T | A\nT ::= S O\nO ::= C*\nA ~ 'a'\nC ~ 'c'\n"} {
		tab := toyTables(t, g)
		for _, in := range [][]string{nil, {"A"}, {"A", "A", "A"}, {"A", "C", "A"}, {"A", "A", "C", "C"}, {"C"}, {"A", "A", "A", "A", "A", "A", "A", "C"}} {
			if d := leoAgree(tab, tokens(in...)); d != "" {
				t.Errorf("%q on %v: %s", g, in, d)
			}
		}
	}
	_, x := guacamole(t)
	srcs := append([]string{
		"print print print 1;", "1; 2; 3; 4; 5;", "{ 1; { 2; { 3; } } }", "sub f { 1 } sub g { 2 } 3;",
		"my $x = 1; my $y = 2;", "return return 1;", "my $x = $y = $z = 1;", "1;;",
	}, fuzzSeeds...)
	for _, src := range srcs {
		s, _ := x.Lex([]byte(src))
		if d := leoAgree(guacTab, lex.Linear(s)); d != "" {
			t.Errorf("%q: %s", src, d)
		}
	}
}
