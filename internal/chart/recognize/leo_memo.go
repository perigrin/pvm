// ABOUTME: The Leo memo step: a memo where the whole Earley set holds one item awaiting a symbol, and it is a Leo item.
// ABOUTME: Options.NoLeo, checked in leo.go before this is called, is the only switch that turns it off.

package recognize

// memoAt is the Leo memo at node at for sym, or nil, made once and kept in
// the chart's memo store. A memo needs the whole set to hold exactly one
// item awaiting sym, across its states, and that item to be a Leo item
// (Leo 1991); static eligibility alone is not enough. The test reads the
// postdot index, so it is constant time and runs before the store is
// read, and only real memos are stored.
func (c *Chart) memoAt(at, sym int) *memo {
	if !c.r.g.leoSym[sym] {
		return nil // no state has a Leo item awaiting sym
	}
	n, state, origin, k := 0, 0, 0, 0
	if set := c.Sets[at]; set != nil && set.core != nil {
		en := c.r.awaiting(set, sym)
		if n = len(en); n == 1 {
			it := set.items[en[0].idx]
			state, origin, k = it.State, at-it.Dist, int(en[0].target)
		}
	} else {
		// shortcut: as in scan, a branch set searches every item, on every
		// call, since a "none" is not stored.
		c.r.memoScans++
		for _, it := range set.Items() {
			if to, ok := c.r.goTo(it.State, sym); ok {
				n++
				state, origin, k = it.State, at-it.Dist, to
			}
		}
	}
	lhs, ok := c.r.g.leo[[2]int{state, sym}]
	if n != 1 || !ok {
		return nil
	}
	key := [2]int{at, sym}
	if m := c.leo[key]; m != nil {
		return m
	}
	m := &memo{at: at, origin: origin, top: [2]int{k, origin}}
	// A predicted item's origin is this node; its chain stops here. So
	// does one at a start-rule item from node 0, which acceptance reads.
	// The origin != 0 test is the one position-dependent step here: it is
	// cache-safe only because node 0's set is {(0, 0)} and every other
	// node's set has a kernel item with origin < pos, so no other node
	// shares node 0's set id. A mode seeding node 0 differently, or
	// starting mid-file, must revisit it.
	if origin < at && (lhs != c.r.tab.Start || origin != 0) {
		if next := c.memoAt(origin, lhs); next != nil {
			m.top, m.next = next.top, next
		}
	}
	if c.leo == nil {
		c.leo = map[[2]int]*memo{}
	}
	c.leo[key] = m
	return m
}
