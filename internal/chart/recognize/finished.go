// ABOUTME: Read-only accessors over a finished chart, for the forest: the completion a Leo memo skips, and an inserted ;'s items.
// ABOUTME: They read the chart's sets as recognition left them and write nothing: no item, set, memo or intern key.

package recognize

// LeoLink is the first completion a Leo memo at node at for sym skips:
// completing sym from at to a later node also completes rule from origin
// there, though no set holds that item (Leo 1991). ok is false where no memo
// applies, or with Leo off. Where more is true the memo's chain goes on with
// LeoLink(origin, the rule's LHS); where it is false the chain stops at this
// link, at a predicted item or a start rule from node 0, as memoAt stops. It
// is the link memoAt makes, read from the sets and not stored, because a
// node the transition cache served ran no completion and so made no memo,
// though its set holds the chain's top. An interned set is read through its
// postdot index, as memoAt reads it; a branch set has no core and is scanned.
func (c *Chart) LeoLink(at, sym int) (rule, origin int, more, ok bool) {
	if c.r.opts.NoLeo || at < 0 || at >= len(c.Sets) {
		return 0, 0, false, false
	}
	if set := c.Sets[at]; set != nil && set.core != nil && !c.r.g.leoSym[sym] {
		return 0, 0, false, false
	}
	it, ok := c.awaitingOne(at, sym)
	if !ok {
		return 0, 0, false, false
	}
	lhs, ok := c.r.g.leo[[2]int{it.State, sym}]
	if !ok {
		return 0, 0, false, false
	}
	for _, i := range c.r.tab.States[it.State].Leo {
		if x := c.r.tab.Items[i]; c.r.tab.Rules[x.Rule].RHS[x.Dot] == sym {
			rule = x.Rule
		}
	}
	origin = at - it.Dist
	more = origin < at && (lhs != c.r.tab.Start || origin != 0)
	if p, ok := c.awaitingOne(origin, lhs); more && ok && c.r.predicts(p.State, p.Ctx) != it.Ctx {
		more = false // the next link does not pair with this one's completion, as memoAt stops (§4.3)
	}
	return rule, origin, more, true
}

// LeoLinkContext is the context of the completion LeoLink(at, sym) reports: the context of the one item at at
// awaiting sym, which its rule's item from origin would carry. It is 0 where no item or several await sym.
func (c *Chart) LeoLinkContext(at, sym int) int {
	it, _ := c.awaitingOne(at, sym)
	return it.Ctx
}

// Predicts is the context an item of state under context ctx predicts its children under (§4.3's pairing): Build
// walks by it (forest.Build).
func (c *Chart) Predicts(state, ctx int) int { return c.r.predicts(state, ctx) }

// awaitingOne is the one item of the set at node at awaiting sym, and false where it holds none or more.
func (c *Chart) awaitingOne(at, sym int) (Item, bool) {
	var one Item
	n := 0
	if set := c.Sets[at]; set != nil && set.core != nil {
		en := c.r.awaiting(set, sym)
		if n = len(en); n == 1 {
			one = set.items[en[0].idx]
		}
	} else {
		for _, it := range c.Sets[at].Items() {
			if _, ok := transition(c.r.tab, it.State, sym); ok {
				n, one = n+1, it
			}
		}
	}
	return one, n == 1
}

// Slipped is the items the ; inserted at node pos advances, closed under
// completion as recognition closed them before the token whose terminal set
// is next, or at the end of input where next is -1; nil where no item at
// pos awaits a ;. They are beside the set at pos, never in it (slipper.go).
// They are the closure recognition kept, Leo on: built at the node after
// pos, or at pos where it is the last, or, where the transition cache served
// that node, the closure its storing node kept, cached with the transition
// (R6). Every node the slipper ran at keeps one, so nothing is recomputed.
// The result is shared and read-only: the caller writes none of it.
func (c *Chart) Slipped(pos, next int) []Item {
	if c.r.g.semi < 0 || pos < 0 || pos >= len(c.Sets) {
		return nil
	}
	return c.slips[[2]int{pos, next}]
}
