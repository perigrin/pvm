// ABOUTME: The token lattice the chart recognizes over: nodes at token boundaries, each with a lexer state, joined by tokens.
// ABOUTME: Up front it is linear; the rejoin rule says where two branched paths meet again, and refuses past a bound.

package lex

import (
	"fmt"

	"tamarou.com/pvm/internal/lexer"
)

// State is the lexer state a lattice node carries, as the chart compares it
// (spec §5, "Input model"): Equal when lexing forward from equal states at
// one offset gives the same tokens, as lexer.State.Equal and
// TokenCache.resyncsAt use it, but with the production expect field left
// out, which Node holds apart.
type State[S any] interface {
	Equal(S) bool
}

// Node is a lattice position: a token boundary's byte offset, the production
// lexer's expect state there, and the rest of the lexer state.
type Node[S State[S]] struct {
	Offset int
	// Expect is not compared: interleaved mode takes its expectation from the
	// Earley set, so two paths perl's grammar merges can differ here.
	Expect lexer.Expect
	State  S
}

// Edge is a token from one node to another, by index into Nodes.
type Edge struct {
	From, To int
	Token    Token
}

// Lattice is the recognizer's input: nodes, and the tokens between them.
type Lattice[S State[S]] struct {
	Nodes []Node[S]
	Edges []Edge
}

// None is the state of an up-front lattice's nodes, which carry no lexer
// state: up-front lexing never branches, so nothing is compared, and the
// production lexer's state is unexported.
type None struct{}

// Equal is always true; nothing is carried to differ.
func (None) Equal(None) bool { return true }

// Linear is an up-front stream as a lattice: one node per token boundary,
// at byte 0 and at each token's end, and one edge per token.
func Linear(s *Stream) *Lattice[None] {
	l := &Lattice[None]{Nodes: []Node[None]{{}}}
	for i, t := range s.Tokens {
		l.Nodes = append(l.Nodes, Node[None]{Offset: t.End})
		l.Edges = append(l.Edges, Edge{From: i, To: i + 1, Token: t})
	}
	return l
}

// Rejoin is where two paths forked from one node meet again: the indexes
// into a and b of the first nodes whose offset and state are both equal,
// the expect field aside. a[0] and b[0] are the fork, so an index counts
// tokens from it. A branch not rejoined within bound tokens on each path is
// refused.
func Rejoin[S State[S]](a, b []Node[S], bound int) (int, int, error) {
	for i := 1; i < len(a) && i <= bound; i++ {
		for j := 1; j < len(b) && j <= bound; j++ {
			if a[i].Offset == b[j].Offset && a[i].State.Equal(b[j].State) {
				return i, j, nil
			}
		}
	}
	return 0, 0, fmt.Errorf("byte %d: the branch does not rejoin within %d tokens", a[0].Offset, bound)
}
