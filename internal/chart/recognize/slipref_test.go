// ABOUTME: The Leo-off recomputation of a slipper closure, the reference the slipper tests hold kept closures to.
// ABOUTME: Test-only: Slipped returns the closure recognition kept and recomputes nothing.

package recognize

// slippedLeoOff is the slipper's closure at pos recomputed with Leo off, the
// reference TestSlipperClosuresAgree and TestSlippedCacheServedNodesKeepClosure
// hold the kept closures to. It writes no memo.
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
	r.itemsBuf, r.depsBuf = nil, nil
	nc := *c
	nc.r = &r
	v := r.semiAt(&nc, pos, la)
	if v == nil {
		return nil
	}
	r.relBuf = make([]Item, 0, len(v.items)) // the result, in one allocation
	return v.relative()
}
