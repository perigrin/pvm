// ABOUTME: Expected(pos): the terminals the Earley set at a lattice node expects next, by name.
// ABOUTME: The union over every state in the set; an expectation class projects it onto the conflicting pairs (§5).

package recognize

import (
	"fmt"
	"slices"
)

// Expected is the terminals some state in the set at node pos expects, by
// name, sorted. A node with no set expects nothing; a pos outside the
// lattice is an error. With lookahead on, a completion the next token
// cannot follow was not made, so its expectations are absent.
func (c *Chart) Expected(pos int) ([]string, error) {
	if pos < 0 || pos >= len(c.Sets) {
		return nil, fmt.Errorf("recognize: node %d is outside the lattice's %d nodes", pos, len(c.Sets))
	}
	var ids []int
	for _, it := range c.Sets[pos].Items() {
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
