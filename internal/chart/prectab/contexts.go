// ABOUTME: The context a predicted item carries (V2 design §4.2) computed statically per predictor.
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
// on a grammar the recognizer judges, so there is at most one, and where there are more the least is taken. §4.3's
// collapse of a top that forbids no reachable shift adds nothing at this granularity: a canonical top forbids shifts
// only in rules predicted from it, which lie in its own predictor's region, so it never fires.
func PredictionContexts(p *Precedence) Contexts {
	tab := p.Tab
	c := Contexts{Kernel: slices.Repeat([]int32{-1}, len(tab.States))}
	for _, k := range predictors(tab) {
		ctx := rawTop(p, k)
		if k.kernel < 0 {
			c.Start = ctx
		} else {
			c.Kernel[k.kernel] = ctx
		}
	}
	return c
}

// rawTop is k's context: its rules' one non-none canonical top (§4.2), the least where there
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
