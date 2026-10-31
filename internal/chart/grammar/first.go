// ABOUTME: FIRST and FOLLOW of every symbol, computed once by the generator and stored on the tables.
// ABOUTME: The recognizer's lookahead pruning reads FOLLOW; the end of input is the id Tables.EOF.

package grammar

// firstFollow sets Tables.First and Tables.Follow, each symbol's sorted
// terminal ids, with Tables.EOF in a FOLLOW set where the end of input can
// follow. A terminal's FIRST is itself.
func (t *Tables) firstFollow() {
	n := len(t.Symbols)
	t.EOF = n
	first, follow := newSets(n), newSets(n)
	for s, sym := range t.Symbols {
		if sym.Terminal {
			first[s][s] = true
		}
	}
	nullable := func(s int) bool { return t.Symbols[s].Nullable }
	for changed := true; changed; {
		changed = false
		for _, r := range t.Rules {
			for _, x := range r.RHS {
				changed = first[r.LHS].add(first[x]) || changed
				if !nullable(x) {
					break
				}
			}
		}
	}
	follow[t.Start][t.EOF] = true
	for changed := true; changed; {
		changed = false
		for _, r := range t.Rules {
			for i, x := range r.RHS {
				if t.Symbols[x].Terminal {
					continue
				}
				restNullable := true
				for _, y := range r.RHS[i+1:] {
					changed = follow[x].add(first[y]) || changed
					if !nullable(y) {
						restNullable = false
						break
					}
				}
				if restNullable {
					changed = follow[x].add(follow[r.LHS]) || changed
				}
			}
		}
	}
	t.First, t.Follow = make([][]int, n), make([][]int, n)
	for s := range n {
		t.First[s], t.Follow[s] = first[s].ids(), follow[s].ids()
	}
}

// set is a set of symbol ids, and the end of input as id len(Symbols).
type set []bool

func newSets(n int) []set {
	sets := make([]set, n)
	for i := range sets {
		sets[i] = make(set, n+1)
	}
	return sets
}

// add puts o in s and reports whether s grew.
func (s set) add(o set) bool {
	grew := false
	for i, in := range o {
		if in && !s[i] {
			s[i], grew = true, true
		}
	}
	return grew
}

func (s set) ids() []int {
	var ids []int
	for i, in := range s {
		if in {
			ids = append(ids, i)
		}
	}
	return ids
}
