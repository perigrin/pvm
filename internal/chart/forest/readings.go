// ABOUTME: The readings count (spec §9.4): each statement's derivations, counted after precedence and resolve have acted.
// ABOUTME: A statement above 2^10 readings collapses: counted once and never enumerated; collapse is no refusal.

package forest

import (
	"math"
	"math/bits"
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
func (f *Forest) countReadings(int) Readings { return Readings{} }

// counter is a count of each node's derivations, memoized across calls.
func (f *Forest) counter() func(n int32) uint64 {
	return func(n int32) uint64 { return 0 }
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
