// ABOUTME: Build's walk by context (C6): a node keeps only the alternatives the chart completed under a context it is
// ABOUTME: reached under, so neither one of a dead context nor one of a rule refused into every parent is attached.

package forest

import "slices"

// leoFact is a completion a Leo memo skipped at node at: links[link]'s rule from its origin.
type leoFact struct{ at, link int }

// byContext walks the forest from the root by context (the V2 design's §4.3): the root is reached under the start
// item's context, and each kept alternative's children under the contexts that predicted them, its left one under
// its own and its right one under the one the item before it predicts (recognize.Chart.Predicts). A node reached
// keeps the alternatives the chart derived under one of the contexts it is reached under, in Build's order, and drops
// the rest: one completed only under a context whose parent chain died, and one of a rule the recognizer refused into
// every parent. A chart with no context and no refusal derives every alternative Build made under none, so it keeps
// them all.
func (b *builder) byContext() {
	if !b.walk || b.Root == 0 {
		return
	}
	type reach struct{ n, ctx int32 }
	kept := make([]bool, len(b.Alts))
	reached := make([]bool, len(b.Nodes))
	first := make([]int32, len(b.Nodes)) // the context each reached node was first reached under
	var more map[reach]bool              // the other (node, context) pairs, which are few
	work := []reach{{b.Root, int32(b.c.Sets[0].Items()[0].Ctx)}}
	var x reach
	var a int32
	keep := func(right int) {
		if !kept[a] {
			kept[a] = true
			if l := b.Alts[a].Left; l != 0 {
				work = append(work, reach{l, x.ctx})
			}
		}
		if r := b.Alts[a].Right; r != 0 {
			work = append(work, reach{r, int32(right)})
		}
	}
	for len(work) > 0 {
		x = work[len(work)-1]
		work = work[:len(work)-1]
		switch {
		case !reached[x.n]:
			reached[x.n], first[x.n] = true, x.ctx
		case first[x.n] == x.ctx:
			continue
		default:
			if more == nil {
				more = map[reach]bool{}
			}
			if more[x] {
				continue
			}
			more[x] = true
		}
		for a = b.Nodes[x.n].Alt; a != 0; a = b.Alts[a].Next {
			b.under(x.n, a, int(x.ctx), keep)
		}
	}
	for n := range b.Nodes {
		if !reached[n] {
			continue
		}
		for p := &b.Nodes[n].Alt; *p != 0; {
			if kept[*p] {
				p = &b.Alts[*p].Next
			} else {
				*p = b.Alts[*p].Next
			}
		}
	}
}

// under calls keep with each context alternative a of node n, reached under context ctx, reaches its right child
// under, where the chart derived a under ctx: a symbol node's rule completed from its start under ctx; the item
// before the right child standing at the split under ctx; and that child, a nonterminal, completed there under the
// context the item predicts, or the left corner's, ctx itself. A rule position's own item was found under ctx by
// the parent that reached it.
func (b *builder) under(n int32, a int32, ctx int, keep func(right int)) {
	nd, alt := b.Nodes[n], b.Alts[a]
	r, m := int(alt.Rule), int(nd.Dot)
	from, to := int(nd.From), int(nd.To)
	if nd.Sym >= 0 {
		m = len(b.tab.Rules[r].RHS)
		if !b.completedUnder(int(nd.Sym), r, from, to, ctx) {
			return
		}
	}
	if m == 0 {
		keep(ctx)
		return
	}
	k := int(b.Nodes[alt.Right].From)
	x := b.tab.Rules[r].RHS[m-1]
	terminal := b.tab.Symbols[x].Terminal
	if m == 1 {
		if terminal || b.completedUnder(x, -1, k, to, ctx) {
			keep(ctx)
		}
		return
	}
	core := b.g.core[r] + m - 1
	for _, p := range b.placedAt(from, k) {
		if int(p.ctx) != ctx {
			continue
		}
		if _, ok := slices.BinarySearch(b.tab.States[p.state].Items, core); !ok {
			continue
		}
		if terminal {
			keep(ctx)
			return
		}
		if right := b.c.Predicts(int(p.state), ctx); b.completedUnder(x, -1, k, to, right) {
			keep(right)
		}
	}
}

// completedUnder reports whether the chart completed sym from node from at node to under context ctx, by rule, or by
// any rule where rule is -1: an item of its set or the slipper's closure there, or a completion a Leo memo skipped.
func (b *builder) completedUnder(sym, rule, from, to, ctx int) bool {
	for _, p := range b.placedAt(from, to) {
		if int(p.ctx) != ctx {
			continue
		}
		for _, x := range b.tab.States[p.state].Reduces {
			if (x == rule || rule < 0 && b.tab.Rules[x].LHS == sym) && !b.c.RuleRefused(to, x, sym, from) {
				return true
			}
		}
	}
	i, _ := slices.BinarySearchFunc(b.leoAt, to, func(f leoFact, k int) int { return f.at - k })
	for ; i < len(b.leoAt) && b.leoAt[i].at == to; i++ {
		l := b.links[b.leoAt[i].link]
		if l.origin == from && l.ctx == ctx && (l.rule == rule || rule < 0 && b.tab.Rules[l.rule].LHS == sym) {
			return true
		}
	}
	return false
}

// placedAt is the items at node at from node from.
func (b *builder) placedAt(from, at int) []placed {
	run := b.items[b.from[from]:b.from[from+1]]
	lo, _ := slices.BinarySearchFunc(run, at, func(p placed, k int) int { return int(p.at) - k })
	hi, _ := slices.BinarySearchFunc(run, at+1, func(p placed, k int) int { return int(p.at) - k })
	return run[lo:hi]
}
