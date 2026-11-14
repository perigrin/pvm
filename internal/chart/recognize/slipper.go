// ABOUTME: The narrow ruby slipper: a zero-width ; scanned where toke.c inserts one, before } and at end of input.
// ABOUTME: Its items close beside a node's set, never in it, so Sets and Expected stay what the source may hold.

package recognize

import (
	"slices"

	"tamarou.com/pvm/internal/chart/lex"
)

// semiAt is the items a zero-width ; advances at node pos, closed under
// completion with lookahead la, as a builder at pos; nil where nothing at pos
// awaits a ; (spec §5.2). toke.c inserts that ; before the } that closes a
// scope (yyl_rightcurly) and at the end of the file (yyl_fake_eof).
func (r *Recognizer) semiAt(c *Chart, pos int, la bitset) *builder {
	v := &builder{r: r, c: c, pos: pos, interned: true}
	for _, it := range c.Sets[pos].Items() {
		if k, ok := r.goTo(it.State, r.g.semi); ok {
			v.add(k, pos-it.Dist)
		}
	}
	if len(v.items) == 0 {
		return nil
	}
	if r.opts.NoLookahead {
		la = nil
	}
	v.complete(la)
	return v
}

// slip scans e's }, beside its scan from e.From's set, from the items an
// inserted ; advances there; their origin sets are this node's dependencies.
func (b *builder) slip(e lex.Edge) {
	v := b.r.semiAt(b.c, e.From, b.r.cls.bits[e.Token.TerminalSet])
	if v == nil {
		return
	}
	b.c.keep(e.From, e.Token.TerminalSet, v)
	b.interned = b.interned && v.interned
	for _, d := range v.deps {
		b.read(e.From-d.dist, b.c.Sets[e.From-d.dist])
	}
	for _, it := range v.items {
		// shortcut: an item predicted after the ; is not carried, since its
		// origin would name the set before it; exact while no rule's } is
		// preceded only by nullable symbols (the nullable closure keeps such
		// items predicted), as holds for perly.y. Give the ; a node of its own if one is.
		if b.r.tab.States[it[0]].Predicted {
			continue
		}
		for _, t := range b.r.cls.terms[e.Token.TerminalSet] {
			if k, ok := b.r.goTo(it[0], t); ok {
				b.add(k, it[1])
			}
		}
	}
}

// slipped records the nodes where a ; was inserted, those before a } or at
// the end of input whose set awaits one, and whether the one at the end of
// input completes the start. out is the set ids leaving each node.
func (c *Chart) slipped(out [][]int) {
	r, last := c.r, len(c.Sets)-1
	if r.g.semi < 0 {
		return
	}
	for pos, s := range c.Sets {
		if pos != last && !slices.ContainsFunc(out[pos], func(id int) bool { return r.cls.closes[id] }) {
			continue
		}
		if slices.ContainsFunc(s.Items(), func(it Item) bool { _, ok := r.goTo(it.State, r.g.semi); return ok }) {
			c.inserted = append(c.inserted, pos)
		}
	}
	if v := r.semiAt(c, last, r.g.eofBits); v != nil {
		c.keep(last, -1, v)
		c.slipAccepted = slices.ContainsFunc(v.items, func(it [2]int) bool {
			return it[1] == 0 && slices.Contains(r.g.lhs[it[0]], r.tab.Start)
		})
	}
}

// keep keeps v, the slipper's closure at pos before the terminal set next
// (-1 at the end of input), for Slipped: a copy, since v's items are
// relative in the recognizer's buffer.
func (c *Chart) keep(pos, next int, v *builder) {
	if c.slips == nil {
		c.slips = map[[2]int][]Item{}
	}
	c.slips[[2]int{pos, next}] = slices.Clone(v.relative())
}

// Inserted is the nodes where the slipper scanned a ;, ascending.
func (c *Chart) Inserted() []int { return c.inserted }
