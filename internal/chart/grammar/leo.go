// ABOUTME: Per-state Leo eligibility: a state whose single item is a right-recursive completion candidate.
// ABOUTME: The recognizer memoizes Leo items only for these states (spec §5); the generator writes them to leo.txt.

package grammar

// leo marks each state whose one item has its dot before the last symbol
// of its rule, that symbol a nonterminal that derives the rule's own LHS at
// its right edge. Completing such an item is the step a right recursion
// repeats, which Leo's memo skips.
func (t *Tables) leo() {
	right := t.rightEdges()
	for s := range t.States {
		st := &t.States[s]
		if len(st.Items) != 1 {
			continue
		}
		it := t.Items[st.Items[0]]
		rhs := t.Rules[it.Rule].RHS
		if it.Dot != len(rhs)-1 || t.Symbols[rhs[it.Dot]].Terminal {
			continue
		}
		st.Leo = right[rhs[it.Dot]][t.Rules[it.Rule].LHS]
	}
}

// rightEdges is, for each nonterminal, the nonterminals that can be the last
// symbol of something it derives: its rules' last symbols, transitively.
func (t *Tables) rightEdges() map[int]map[int]bool {
	last := map[int][]int{}
	for _, r := range t.Rules {
		if n := len(r.RHS); n > 0 && !t.Symbols[r.RHS[n-1]].Terminal {
			last[r.LHS] = append(last[r.LHS], r.RHS[n-1])
		}
	}
	edges := map[int]map[int]bool{}
	for nt := range last {
		seen := map[int]bool{}
		work := append([]int(nil), last[nt]...)
		for len(work) > 0 {
			x := work[len(work)-1]
			work = work[:len(work)-1]
			if seen[x] {
				continue
			}
			seen[x] = true
			work = append(work, last[x]...)
		}
		edges[nt] = seen
	}
	return edges
}
