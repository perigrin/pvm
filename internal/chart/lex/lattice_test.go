// ABOUTME: Tests the token lattice: the up-front lattice is linear, and the rejoin rule holds on a toy lattice.
// ABOUTME: Two paths rejoin only where offset and state agree, the expect field aside, and within the bound.

package lex

import (
	"testing"

	"tamarou.com/pvm/internal/lexer"
)

// toy is a toy lexer state: a pending heredoc queue, the kind of state the
// rejoin rule exists for (`foo <<EOF;` as shift or heredoc).
type toy struct{ pending string }

func (s toy) Equal(o toy) bool { return s == o }

// path is a toy path from a fork at byte 0: one node per offset, all in
// state st, with the fork node first.
func path(st toy, offsets ...int) []Node[toy] {
	nodes := []Node[toy]{{Offset: 0}}
	for _, off := range offsets {
		nodes = append(nodes, Node[toy]{Offset: off, State: st})
	}
	return nodes
}

func TestLinearLattice(t *testing.T) {
	x := guacamole(t)
	src := []byte("my $x = 1;\n")
	s, err := x.Lex(src)
	if err != nil {
		t.Fatal(err)
	}
	l := Linear(s)
	if len(l.Nodes) != len(s.Tokens)+1 || len(l.Edges) != len(s.Tokens) {
		t.Fatalf("%d nodes and %d edges for %d tokens, want one more node than edges", len(l.Nodes), len(l.Edges), len(s.Tokens))
	}
	for i, e := range l.Edges {
		if e.From != i || e.To != i+1 || l.Nodes[e.To].Offset != e.Token.End {
			t.Errorf("edge %d is %d->%d ending at %d, want %d->%d at the token's end %d",
				i, e.From, e.To, l.Nodes[e.To].Offset, i, i+1, e.Token.End)
		}
	}
}

// TestRejoinNeedsOffsetAndState: paths meeting at one offset in different
// states do not rejoin, nor do equal states at different offsets; offset and
// state both equal do.
func TestRejoinNeedsOffsetAndState(t *testing.T) {
	shift, heredoc := toy{}, toy{pending: "EOF"}
	if _, _, err := Rejoin(path(shift, 3, 6, 10), path(heredoc, 6, 10), 8); err == nil {
		t.Error("paths at the same offsets in different states rejoined")
	}
	if _, _, err := Rejoin(path(shift, 3, 7, 11), path(shift, 6, 10), 8); err == nil {
		t.Error("paths in equal states at different offsets rejoined")
	}
	i, j, err := Rejoin(path(shift, 3, 6, 10), path(shift, 6, 10), 8)
	if err != nil || i != 2 || j != 1 {
		t.Errorf("Rejoin = %d, %d, %v; want 2, 1 at offset 6", i, j, err)
	}
}

// TestRejoinIgnoresExpect: states that differ only in the production expect
// field rejoin. The pair of the check above: the comparison is not so strict
// that nothing rejoins.
func TestRejoinIgnoresExpect(t *testing.T) {
	a, b := path(toy{}, 4, 9), path(toy{}, 9)
	a[2].Expect, b[1].Expect = lexer.XOperator, lexer.XTerm
	i, j, err := Rejoin(a, b, 8)
	if err != nil || i != 2 || j != 1 {
		t.Errorf("Rejoin = %d, %d, %v; want 2, 1: expect is not compared", i, j, err)
	}
}

// TestRejoinBound: a branch rejoining at exactly the bound is accepted, and
// one rejoining a token later is refused.
func TestRejoinBound(t *testing.T) {
	st := toy{}
	if _, _, err := Rejoin(path(st, 1, 2, 3), path(st, 3), 3); err != nil {
		t.Errorf("a branch rejoining at the bound, 3 tokens, was refused: %v", err)
	}
	if _, _, err := Rejoin(path(st, 1, 2, 3, 4), path(st, 4), 3); err == nil {
		t.Error("a branch rejoining after 4 tokens was accepted under a bound of 3")
	}
}
