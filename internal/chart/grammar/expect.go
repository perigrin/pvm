// ABOUTME: Per-state expected-terminal sets: the terminals after a dot in a state's items.
// ABOUTME: Filled in as the tables are built; the recognizer's lookahead pruning reads them.

package grammar

// expected sets each state's Expected: the terminals after a dot in its
// items, in symbol order, which are its transitions on terminals.
func (t *Tables) expected() {
	for s := range t.States {
		st := &t.States[s]
		for _, tr := range st.Transitions {
			if t.Symbols[tr.Symbol].Terminal {
				st.Expected = append(st.Expected, tr.Symbol)
			}
		}
	}
}
