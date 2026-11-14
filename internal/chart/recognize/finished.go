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
// though its set holds the chain's top.
func (c *Chart) LeoLink(at, sym int) (rule, origin int, more, ok bool) {
	if c.r.opts.NoLeo || at < 0 || at >= len(c.Sets) {
		return 0, 0, false, false
	}
	n, state := 0, 0
	for _, it := range c.Sets[at].Items() {
		if _, ok := transition(c.r.tab, it.State, sym); ok {
			n, state, origin = n+1, it.State, at-it.Dist
		}
	}
	lhs, ok := c.r.g.leo[[2]int{state, sym}]
	if n != 1 || !ok {
		return 0, 0, false, false
	}
	for _, i := range c.r.tab.States[state].Leo {
		if it := c.r.tab.Items[i]; c.r.tab.Rules[it.Rule].RHS[it.Dot] == sym {
			rule = it.Rule
		}
	}
	return rule, origin, origin < at && (lhs != c.r.tab.Start || origin != 0), true
}

// Slipped is the items the ; inserted at node pos advances, closed under
// completion as recognition closed them before the token whose terminal set
// is next, or at the end of input where next is -1; nil where no item at
// pos awaits a ;. They are beside the set at pos, never in it (slipper.go).
// Where recognition built the node after pos, or pos is the last, they are
// the closure it kept, Leo on; the caller reads them and writes none. Where
// the transition cache served that node, no closure was made, so they are
// recomputed with Leo off, link by link, since making a memo here would
// write the chart; those are the caller's, a copy. The two give the forest
// the same facts once Leo chains are expanded (TestSlipperClosuresAgree).
// shortcut: a cache-served node recomputes its closure on every call; cache
// the closure with the transition (R6) if that recomputation shows.
func (c *Chart) Slipped(pos, next int) []Item {
	if c.r.g.semi < 0 || pos < 0 || pos >= len(c.Sets) {
		return nil
	}
	if s, ok := c.slips[[2]int{pos, next}]; ok {
		return s
	}
	return c.slippedLeoOff(pos, next)
}

// slippedLeoOff is Slipped's recomputation, Leo off, whether or not the
// chart kept a closure at pos.
func (c *Chart) slippedLeoOff(pos, next int) []Item {
	la := c.r.g.eofBits
	if next >= 0 {
		la = c.r.cls.bits[next]
	}
	// The copy shares the intern table and cache, but semiAt with Leo off
	// writes neither: its builder reads sets and keeps its items to itself,
	// in buffers of its own, so concurrent calls share no memory they write.
	r := *c.r
	r.opts.NoLeo = true
	r.itemsBuf, r.depsBuf, r.relBuf = nil, nil, nil
	nc := *c
	nc.r = &r
	v := r.semiAt(&nc, pos, la)
	if v == nil {
		return nil
	}
	return v.relative()
}
