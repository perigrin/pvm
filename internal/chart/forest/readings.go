// ABOUTME: The readings count (spec §9.4): each statement's derivations, counted after precedence and resolve have acted.
// ABOUTME: A statement above 2^10 readings collapses: counted once and never enumerated; collapse is no refusal.

package forest

import (
	"fmt"
	"math"
	"math/bits"

	"tamarou.com/pvm/internal/chart/grammar"
)

// CollapseAbove is the most readings a statement keeps; one with more collapses (spec §9.4).
const CollapseAbove = 1 << 10

// Readings is what Parse counts over its forest's statements, perly.y's fullstmt at any depth, each reachable from
// the root and wider than the slipper's zero-width ;. A statement's readings are the derivations left under its
// node, a product over its packed nodes, summed over their alternatives, saturating at math.MaxUint64.
type Readings struct {
	Statements int    // the statements, a nested one counted once as its own
	Collapsed  int    // the statements with more than CollapseAbove readings
	Total      uint64 // each statement's readings summed, a collapsed one counted once, saturating
}

// countReadings is f's Readings with stmt the statement symbol. It enumerates no derivation: each node's count is
// memoized, so the work is linear in the forest.
// shortcut: a derivation never holds a node inside itself, so a node met again on the current path counts 0, and a
// count memoized below such a cycle omits that path's node wherever it is reached; exact for an acyclic forest, as
// perly.y's are on every corpus the ratchet reads; count per path if a cyclic forest's count is ever read.
func (f *Forest) countReadings(stmt int) Readings {
	var r Readings
	count := f.counter()
	seen := make([]bool, len(f.Nodes))
	seen[f.Root] = true
	for work := []int32{f.Root}; len(work) > 0; {
		n := work[len(work)-1]
		work = work[:len(work)-1]
		nd := f.Nodes[n]
		if int(nd.Sym) == stmt && nd.To > nd.From {
			r.Statements++
			c := count(n)
			if c > CollapseAbove {
				r.Collapsed++
				c = 1
			}
			r.Total = addSat(r.Total, c)
		}
		for x := nd.Alt; x != 0; x = f.Alts[x].Next {
			for _, k := range [2]int32{f.Alts[x].Left, f.Alts[x].Right} {
				if k != 0 && !seen[k] {
					seen[k] = true
					work = append(work, k)
				}
			}
		}
	}
	return r
}

// counter is a count of each node's derivations, memoized across calls.
func (f *Forest) counter() func(n int32) uint64 {
	memo := make([]uint64, len(f.Nodes))
	state := make([]uint8, len(f.Nodes)) // 0 not reached, 1 on the path, 2 counted
	var count func(n int32) uint64
	count = func(n int32) uint64 {
		nd := f.Nodes[n]
		switch {
		case nd.Alt == 0:
			return 1
		case state[n] == 1:
			return 0
		case state[n] == 2:
			return memo[n]
		}
		state[n] = 1
		var sum uint64
		for x := nd.Alt; x != 0; x = f.Alts[x].Next {
			a, c := f.Alts[x], uint64(1)
			if a.Left != 0 {
				c = count(a.Left)
			}
			if a.Right != 0 && c != 0 {
				c = mulSat(c, count(a.Right))
			}
			sum = addSat(sum, c)
		}
		state[n], memo[n] = 2, sum
		return sum
	}
	return count
}

func addSat(a, b uint64) uint64 {
	if s, carry := bits.Add64(a, b, 0); carry == 0 {
		return s
	}
	return math.MaxUint64
}

func mulSat(a, b uint64) uint64 {
	if hi, lo := bits.Mul64(a, b); hi == 0 {
		return lo
	}
	return math.MaxUint64
}

// parsePrecedence is the committed grammar's precedence over a tables value, built once per value; err where the
// tables are not the committed grammar's, whose rule ids the %prec list is indexed by.
type parsePrecedence struct {
	p   *Precedence
	err error
}

func buildParsePrecedence(tab *grammar.Tables) parsePrecedence {
	g, _, err := grammar.Committed()
	if err != nil {
		return parsePrecedence{err: err}
	}
	if len(tab.Rules) != len(g.G1) {
		return parsePrecedence{err: fmt.Errorf("forest: %d table rules for perly.y's %d: not the committed grammar's tables", len(tab.Rules), len(g.G1))}
	}
	rulePrec := make([]string, len(g.G1))
	for r, rule := range g.G1 {
		if tab.Symbols[tab.Rules[r].LHS].Name != rule.LHS {
			return parsePrecedence{err: fmt.Errorf("forest: table rule %d is %s's, perly.y's is %s's", r, tab.Symbols[tab.Rules[r].LHS].Name, rule.LHS)}
		}
		rulePrec[r] = rule.Prec
	}
	p, err := NewPrecedence(tab, g.Levels, rulePrec, g.Refused)
	return parsePrecedence{p, err}
}
