// ABOUTME: Expected(pos): the terminals the Earley set at a lattice node expects next, by name.
// ABOUTME: The union of the expected terminals of every state in the set without lookahead, less what precedence forbids.

package recognize

import (
	"fmt"
	"slices"
)

// Expected is the terminals some item in the Earley set at node pos expects,
// by name, sorted: the union over every state in the set (§5), less each
// terminal the shift side refuses an item to scan under its context. Lookahead
// prunes completions the next token cannot follow, so with it on the set is
// first closed again under every completion. The closure is not stored and
// changes neither the interned set nor the cache. A pruned completion is
// never needed by a later set, so this is the set recognition without
// lookahead builds. On a judged chart it is less each terminal precedence
// forbids there (allowed), so it is empty where precedence forbids them all.
// A node with no set expects nothing; a pos outside the lattice is an error.
func (c *Chart) Expected(pos int) ([]string, error) {
	if pos < 0 || pos >= len(c.Sets) {
		return nil, fmt.Errorf("recognize: node %d is outside the lattice's %d nodes", pos, len(c.Sets))
	}
	set := c.Sets[pos]
	// The answer depends on the set and on the origin sets its closure
	// read, so it is memoized in the transition cache's trie, under a key
	// no transition has (span 0). A node with no set, or a branch set
	// (id -1), is not memoized.
	memo := set != nil && set.id >= 0 && !c.r.opts.NoCache
	var key cacheKey
	if memo {
		key = cacheKey{set: set.id}
		if n := c.r.hit(c, pos, key); n != nil {
			return slices.Clone(n.names), nil
		}
	}
	c.r.expectedComputed++
	items := set.Items()
	judged := c.r.judge != nil && items != nil
	b := builder{r: c.r, c: c, pos: pos, interned: memo}
	if items != nil && (judged || !c.r.opts.NoLookahead) {
		for _, it := range items {
			b.add(it.State, pos-it.Dist, it.Ctx)
		}
		b.complete(nil)
		items = b.relative()
	}
	var ids []int
	for _, it := range items {
		for _, t := range c.r.tab.States[it.State].Expected {
			if !c.r.scanRefused(it.State, it.Ctx, t) {
				ids = append(ids, t)
			}
		}
	}
	slices.Sort(ids)
	ids = slices.Compact(ids)
	if judged {
		ids = b.allowed(set, ids)
	}
	deps := b.deps
	memo = memo && b.interned
	names := make([]string, len(ids))
	for i, id := range ids {
		names[i] = c.r.tab.Symbols[id].Name
	}
	slices.Sort(names)
	if memo {
		if n := c.r.slot(key, deps); n != nil {
			n.names = names
		}
		return slices.Clone(names), nil
	}
	return names, nil
}

// allowed is the candidates ids that precedence allows after set, the set at b's node (V2 design §4.4): each is
// judged as the lookahead, by closing the set's items no completion made again with it, the reduce side judging every
// completion before it, and is kept where an item of that closure scans it unrefused. The sets the closures read are
// b's dependencies.
func (b *builder) allowed(set *Set, ids []int) []int {
	r := b.r
	var seeds []Item
	for _, it := range set.Items() {
		if !r.judge.completed[it.State] {
			seeds = append(seeds, it)
		}
	}
	kept := ids[:0]
	for _, t := range ids {
		la := newBitset(r.g.eof + 1)
		la.set(t)
		b.items, b.seen, b.refused = b.items[:0], nil, nil
		b.la, b.laKey = la, lookaheadKey(-1, la)
		for _, it := range seeds {
			b.add(it.State, b.pos-it.Dist, it.Ctx)
		}
		b.complete(la)
		if slices.ContainsFunc(b.items, func(it [3]int) bool {
			return slices.Contains(r.tab.States[it[0]].Expected, t) && !r.scanRefused(it[0], it[2], t)
		}) {
			kept = append(kept, t)
		}
	}
	return kept
}
