// ABOUTME: Static Leo eligibility per (state, symbol): the state's one item awaiting B, A ::= α . B, B last.
// ABOUTME: The recognizer memoizes Leo items only for these (spec §5); the generator writes them to leo.txt.

package grammar

// leo sets each state's Leo items: for a nonterminal B, the state's one
// item awaiting B, A ::= α . B, where B is the rule's last symbol and
// derives A at its right edge. Completing such an item is the step a right
// recursion repeats, which Leo's memo skips. Other items in the state do not
// matter: S ::= A . S is eligible beside S ::= A . (Leo 1991). This is
// necessary, not sufficient: the recognizer memoizes only where the whole
// Earley set holds one item awaiting B.
//
// B must be last, not followed by nullables. The memo skips the item
// A ::= α B . β, and one whose β can still read a token would lose the
// parses that read it. Trailing nullables count only in the right edge.
func (t *Tables) leo() {
	right := t.rightEdges()
	postdot := func(i int) (int, bool) {
		b, ok := t.after(i)
		return b, ok && !t.Symbols[b].Terminal
	}
	for s := range t.States {
		st := &t.States[s]
		awaiting := map[int]int{} // nonterminal to the items awaiting it
		for _, i := range st.Items {
			if b, ok := postdot(i); ok {
				awaiting[b]++
			}
		}
		for _, i := range st.Items {
			b, ok := postdot(i)
			if !ok || awaiting[b] != 1 {
				continue
			}
			it := t.Items[i]
			rule := t.Rules[it.Rule]
			if it.Dot == len(rule.RHS)-1 && right[b][rule.LHS] {
				st.Leo = append(st.Leo, i)
			}
		}
	}
}

// rightEdges is, for each nonterminal, the nonterminals that can be the last
// non-empty symbol of something it derives: each rule's last symbol, and the
// one before it while those after are nullable, transitively.
func (t *Tables) rightEdges() map[int]map[int]bool {
	last := map[int][]int{}
	for _, r := range t.Rules {
		for i := len(r.RHS) - 1; i >= 0; i-- {
			x := r.RHS[i]
			if t.Symbols[x].Terminal {
				break
			}
			last[r.LHS] = append(last[r.LHS], x)
			if !t.Symbols[x].Nullable {
				break
			}
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
