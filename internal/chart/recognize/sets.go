// ABOUTME: Interned Earley sets: a set is its (state, distance) items, interned as a (core, distances) pair.
// ABOUTME: Cores and distance vectors are each deduplicated, so equal content at two positions is one *Set.

package recognize

import "encoding/binary"

// Item is an Earley item: an LR(0) state and the distance in lattice nodes
// back to its origin. Distance 0 is a prediction at this node.
type Item struct {
	State, Dist int
}

// Set is an Earley set's content, sorted by state then distance. An
// interned set has an id; one built inside a branch has id -1.
type Set struct {
	id    int
	items []Item
	core  *setCore // nil for a set built inside a branch
}

// setCore is a set core's postdot index (YAEP's core_symb_vect, Marpa's
// postdot items): per symbol, the items awaiting it, as (index into the
// set's items, target state). Items are sorted by state, so one index
// holds for every set of the core.
type setCore struct {
	id  int
	off []int32 // by symbol, into ent; nil until first used
	ent []coreEnt
}

type coreEnt struct{ idx, target int32 }

// awaiting is s's items awaiting sym, in item order; s must have a core.
// The core's index is built on first use, by the recognizer that interned
// it, which owns it.
func (r *Recognizer) awaiting(s *Set, sym int) []coreEnt {
	c := s.core
	if c.off == nil {
		states := r.tab.States
		c.off = make([]int32, len(r.tab.Symbols)+2)
		n := 0
		for _, it := range s.items {
			for _, tr := range states[it.State].Transitions {
				c.off[tr.Symbol+2]++
				n++
			}
		}
		for i := 2; i < len(c.off); i++ {
			c.off[i] += c.off[i-1]
		}
		c.ent = make([]coreEnt, n)
		for i, it := range s.items {
			for _, tr := range states[it.State].Transitions {
				c.ent[c.off[tr.Symbol+1]] = coreEnt{int32(i), int32(tr.Target)}
				c.off[tr.Symbol+1]++
			}
		}
	}
	return c.ent[c.off[sym]:c.off[sym+1]]
}

// Items is the set's items; a nil set has none.
func (s *Set) Items() []Item {
	if s == nil {
		return nil
	}
	return s.items
}

// interner is YAEP's: the states of a set are its core, its distances a
// vector, each deduplicated, and a set is interned as the pair of their ids.
type interner struct {
	cores map[string]*setCore
	dists map[string]int
	sets  map[[2]int]*Set
}

func newInterner() interner {
	return interner{cores: map[string]*setCore{}, dists: map[string]int{}, sets: map[[2]int]*Set{}}
}

// intern is the one set with items' content, and whether it existed.
func (in *interner) intern(items []Item) (*Set, bool) {
	var core, dist []byte
	for _, it := range items {
		core = binary.AppendUvarint(core, uint64(it.State))
		dist = binary.AppendUvarint(dist, uint64(it.Dist))
	}
	co, ok := in.cores[string(core)]
	if !ok {
		co = &setCore{id: len(in.cores)}
		in.cores[string(core)] = co
	}
	k := [2]int{co.id, idOf(in.dists, string(dist))}
	if s, ok := in.sets[k]; ok {
		return s, true
	}
	s := &Set{id: len(in.sets), items: items, core: co}
	in.sets[k] = s
	return s, false
}

func idOf(m map[string]int, k string) int {
	if id, ok := m[k]; ok {
		return id
	}
	m[k] = len(m)
	return m[k]
}
