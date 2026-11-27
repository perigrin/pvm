// ABOUTME: The context a predicted item carries (V2 design §4.2) after §4.3's collapse, computed statically per predictor.
// ABOUTME: Contexts reads ruleTops and descendants from the feasibility checks, so it rests on Feasible's counters being 0.

package prectab

import "slices"

// Contexts is the context each predictor gives the item it predicts: the canonical spine top, or -1 for none. Every
// item reached from a predicted item at its origin inherits that item's context, so this table is every context an
// item can carry.
type Contexts struct {
	Start  int32   // the start state's item at node 0
	Kernel []int32 // by state: the context of the predicted item a kernel state with an Epsilon adds; -1 for the rest
}

// PredictionContexts is p's Contexts. A predictor's rules all take its one non-none canonical top (§4.2: `do`,
// UNIOP and `my \` begin some rules under none, and taking the real top is exact); Feasible's context counter b is 0
// on a grammar the recognizer judges, so there is at most one, and where there are more the least is taken. That top
// collapses to none (§4.3) where it forbids no shift any core item makes in a state reachable at its origin, the
// predicted state's descendants: every item carrying it lies there, so it changes no verdict any of them reaches, and
// the derivations under it are exactly those under none, so it selects no parent a none could not.
func PredictionContexts(p *Precedence) Contexts {
	tab := p.Tab
	c := Contexts{Kernel: slices.Repeat([]int32{-1}, len(tab.States))}
	for _, k := range predictors(tab) {
		ctx := rawTop(p, k)
		if ctx >= 0 && !forbidsAt(p, ctx, k.epsilon) {
			ctx = -1
		}
		if k.kernel < 0 {
			c.Start = ctx
		} else {
			c.Kernel[k.kernel] = ctx
		}
	}
	return c
}

// rawTop is k's context before §4.3's collapse: its rules' one non-none canonical top (§4.2), the least where there
// are more, or -1.
func rawTop(p *Precedence, k predictor) int32 {
	raw := int32(-1)
	for _, ts := range ruleTops(p, k) {
		for t := range ts {
			if t >= 0 && (raw < 0 || t < raw) {
				raw = t
			}
		}
	}
	return raw
}

// forbidsAt reports whether top forbids a shift some core item makes in a state reachable from predicted state e at
// its origin.
func forbidsAt(p *Precedence, top int32, e int) bool {
	for _, s := range descendants(p.Tab, e) {
		for _, i := range p.Tab.States[s].Items {
			for _, f := range itemShifts(p.Tab, i) {
				if p.ShiftForbidden(top, p.Item(i), int32(f)) {
					return true
				}
			}
		}
	}
	return false
}
