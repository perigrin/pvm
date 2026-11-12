// ABOUTME: The shared packed forest, built after recognition from a chart's completed items, Leo memos and inserted ;s.
// ABOUTME: Nodes are (symbol or rule position, span); nothing is added to an Earley item, and the recognizer stays Boolean.

// Package forest builds the chart's forest after recognition (spec §9.4).
package forest

import (
	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
)

// Forest is a file's shared packed forest.
type Forest struct {
	Root *Node
}

// Node is a symbol, or a rule position, over a span of lattice nodes.
type Node struct {
	Sym, Rule, Dot int
	From, To       int
	Alts           []Alt
}

// Alt is one packed derivation of a node.
type Alt struct {
	Rule        int
	Left, Right *Node
}

// Tree is one derivation.
type Tree struct {
	Sym, Rule, From, To int
	Kids                []*Tree
}

// Refusal is why a file has no forest; its zero value is no refusal.
type Refusal struct {
	Byte   int
	Reason string
}

// Parse recognizes src and builds its forest.
func Parse(tab *grammar.Tables, src []byte) (*Forest, Refusal) {
	return nil, Refusal{}
}

func build(tab *grammar.Tables, c *recognize.Chart, l *lex.Lattice[lex.None]) *Forest {
	return nil
}

// Derivations is at most limit of n's derivations.
func (n *Node) Derivations(limit int) []*Tree {
	return nil
}
