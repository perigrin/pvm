// ABOUTME: Leo (1991) memos: a right recursion's completions resolved to their top item, once per position and symbol.
// ABOUTME: Memos live in a per-position side table outside set identity; every set a memo read is a cache dependency.

package recognize

// memo is a Leo memo at node at for a symbol B: the Earley set there holds
// one item awaiting B, a Leo item A ::= α . B from node origin. Completing B
// back to at would add A ::= α B . from origin, which completes A back to
// origin, and so on up the recursion. top is where that chain ends, as
// (state, absolute origin), and next is the memo at origin for A, by
// reference. A memo is not part of any set, so two positions with equal
// sets share one *Set whatever their memos (spec §5).
type memo struct {
	at, origin int
	top        [2]int
	next       *memo
}

// leoItems is, per (state, symbol), the LHS of the state's Leo item for that
// symbol (grammar's static eligibility).
func leoItems(r *Recognizer) map[[2]int]int {
	m := map[[2]int]int{}
	for s, st := range r.tab.States {
		for _, i := range st.Leo {
			it := r.tab.Items[i]
			rule := r.tab.Rules[it.Rule]
			m[[2]int{s, rule.RHS[it.Dot]}] = rule.LHS
		}
	}
	return m
}

// leo is the top item completing sym back to origin resolves to, or false
// where no memo applies. With lookahead la, a chain whose top cannot be
// followed by la is not taken: FOLLOW only shrinks up a chain, so the plain
// completion stops at the same link it would without Leo.
func (b *builder) leo(origin, sym int, la bitset) ([2]int, bool) {
	if b.r.opts.NoLeo {
		return [2]int{}, false
	}
	m := b.c.memoAt(origin, sym)
	if m == nil {
		return [2]int{}, false
	}
	if la != nil && !b.r.follow[b.r.lhs[m.top[0]][0]].intersects(la) {
		return [2]int{}, false
	}
	// The chain read the set at every link and at its end; a cache hit
	// needs them all to be the same (2eec's F1).
	// shortcut: walks the chain, O(chain) per use; intern memos when lookahead 0 runs with the cache.
	if b.interned {
		for l := m; l != nil; l = l.next {
			b.read(l.at, b.c.Sets[l.at])
			if l.next == nil {
				b.read(l.origin, b.c.Sets[l.origin])
			}
		}
	}
	return m.top, true
}
