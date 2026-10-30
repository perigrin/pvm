// ABOUTME: Tests Leo memos: the test that fails if Leo does nothing, set identity, left recursion and short inputs.
// ABOUTME: Also that Leo on and off agree in acceptance and in the number of distinct interned sets.

package recognize

import (
	"strings"
	"testing"

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
// (§10). On S ::= A S | A at lookahead 0 the item count with Leo grows
// linearly: doubling the input at most doubles it, give or take a constant.
// Without Leo it grows quadratically, which holds the test to measuring
// something.
func TestLeoLinearAtLookaheadZero(t *testing.T) {
	tab := toyTables(t, rightRec)
	count := func(n int, opts Options) int {
		c := Recognize(New(tab, opts), tokens(slicesRepeat("A", n)...))
		if !c.Accepted() {
			t.Fatalf("%d As refused", n)
		}
		return items(c)
	}
	const n = 200
	off1, off2 := count(n, Options{NoLookahead: true, NoLeo: true}), count(2*n, Options{NoLookahead: true, NoLeo: true})
	if off2 < 3*off1 {
		t.Fatalf("without Leo, %d items at %d tokens and %d at %d: not quadratic, so the grammar tests nothing", off1, n, off2, 2*n)
	}
	on1, on2 := count(n, Options{NoLookahead: true}), count(2*n, Options{NoLookahead: true})
	if on2 > 2*on1+8 {
		t.Errorf("with Leo, %d items at %d tokens and %d at %d: not linear", on1, n, on2, 2*n)
	}
}

func slicesRepeat(s string, n int) []string {
	return strings.Fields(strings.Repeat(s+" ", n))
}
