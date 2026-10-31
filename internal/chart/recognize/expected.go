// ABOUTME: Expected(pos): the terminals the Earley set at a lattice node expects next, by name.
// ABOUTME: The union of the expected terminals of every state in the set, as recognition without lookahead builds it.

package recognize

import (
	"fmt"
	"slices"
)

// Expected is the terminals some item in the Earley set at node pos expects,
// by name, sorted: the union over every state in the set (§5). Lookahead
// prunes completions the next token cannot follow, so with it on the set is
// first closed again under every completion. The closure is not stored and
// changes neither the interned set nor the cache. A pruned completion is
// never needed by a later set, so this is the set recognition without
// lookahead builds. A node with no set expects nothing; a pos outside the
// lattice is an error.
func (c *Chart) Expected(pos int) ([]string, error) {
	if pos < 0 || pos >= len(c.Sets) {
		return nil, fmt.Errorf("recognize: node %d is outside the lattice's %d nodes", pos, len(c.Sets))
	}
	items := c.Sets[pos].Items()
	if items != nil && !c.r.opts.NoLookahead {
		b := builder{r: c.r, c: c, pos: pos, seen: map[[2]int]bool{}}
		for _, it := range items {
			b.add(it.State, pos-it.Dist)
		}
		b.complete(nil)
		items = b.relative()
	}
	var ids []int
	for _, it := range items {
		ids = append(ids, c.r.tab.States[it.State].Expected...)
	}
	slices.Sort(ids)
	ids = slices.Compact(ids)
	names := make([]string, len(ids))
	for i, id := range ids {
		names[i] = c.r.tab.Symbols[id].Name
	}
	slices.Sort(names)
	return names, nil
}
