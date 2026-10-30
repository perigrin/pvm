//go:build !chart_noleo

// ABOUTME: The Leo memo step: a memo where the whole Earley set holds one item awaiting a symbol, and it is a Leo item.
// ABOUTME: Its pair, leo_memo_noleo.go under the chart_noleo tag, makes no memo, so the Leo test can be shown to fail.

package recognize

// memoAt is the Leo memo at node at for sym, or nil, made once and kept in
// the chart's side table. A memo needs the whole set to hold exactly one
// item awaiting sym, across its states, and that item to be a Leo item
// (Leo 1991); static eligibility alone is not enough.
func (c *Chart) memoAt(at, sym int) *memo {
	if m, ok := c.leo[at][sym]; ok {
		return m
	}
	var m *memo
	n, state, origin, k := 0, 0, 0, 0
	for _, it := range c.Sets[at].Items() {
		if to, ok := c.r.goTo(it.State, sym); ok {
			n++
			state, origin, k = it.State, at-it.Dist, to
		}
	}
	if lhs, ok := c.r.leo[[2]int{state, sym}]; ok && n == 1 {
		m = &memo{at: at, origin: origin, top: [2]int{k, origin}}
		// A predicted item's origin is this node; its chain stops here. So
		// does one at a start-rule item from node 0, which acceptance reads.
		if origin < at && (lhs != c.r.tab.Start || origin != 0) {
			if next := c.memoAt(origin, lhs); next != nil {
				m.top, m.next = next.top, next
			}
		}
	}
	if c.leo[at] == nil {
		c.leo[at] = map[int]*memo{}
	}
	c.leo[at][sym] = m
	return m
}
