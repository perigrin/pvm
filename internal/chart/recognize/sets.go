// ABOUTME: Interned Earley sets: a set is its (state, distance) items, interned by one hash of its states and one map lookup.
// ABOUTME: Equal content at two positions is one *Set; sets of one core share its postdot index.

package recognize

import (
	"math/bits"
	"slices"
)

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
	dh    uint64   // hash of the distances, set by intern
}

// setCore is a set core's postdot index (YAEP's core_symb_vect, Marpa's
// postdot items): per symbol, the items awaiting it, as (index into the
// set's items, target state). Items are sorted by state, so one index
// holds for every set of the core.
type setCore struct {
	id    int
	first *Set     // the core's first interned set, which stands for its states
	tab   []*Set   // the core's interned sets, open-addressed by distance hash
	n     int      // sets in tab
	next  *setCore // the next core under the same hash
	off   []int32  // by symbol, into ent; built with the core (intern)
	ent   []coreEnt
}

type coreEnt struct{ idx, target int32 }

// intern is r.sets.intern, which indexes a new core: the index is built
// when the core is, so a finished chart's reads never write it.
func (r *Recognizer) intern(items []Item) (*Set, bool) {
	s, existed := r.sets.intern(items)
	if s.core.off == nil {
		r.index(s)
	}
	return s, existed
}

// index builds s's core's postdot index.
func (r *Recognizer) index(s *Set) {
	c := s.core
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

// awaiting is s's items awaiting sym, in item order; s must have a core.
func (r *Recognizer) awaiting(s *Set, sym int) []coreEnt {
	c := s.core
	return c.ent[c.off[sym]:c.off[sym+1]]
}

// Items is the set's items; a nil set has none.
func (s *Set) Items() []Item {
	if s == nil {
		return nil
	}
	return s.items
}

// interner is YAEP's: a set is its core (the states) and its distances. One
// hash of the states keys a table of cores, each holding its sets in a small
// open-addressed table keyed by a hash of their distances. Hashes only narrow
// the search: items decide identity, so a collision never merges two sets.
type interner struct {
	cores         map[uint64]*setCore
	nsets, ncores int
	// hashMask is all ones; a test clears it to force every hash to collide.
	hashMask uint64
	// interns counts intern calls; lookups and inserts count reads and
	// writes of cores.
	interns, lookups, inserts int
}

func newInterner() interner {
	return interner{cores: map[uint64]*setCore{}, hashMask: ^uint64(0)}
}

// hashSeed is nonzero so that leading zero values change a hash.
const hashSeed = 0x9e3779b97f4a7c15

// mix folds x into h.
func mix(h, x uint64) uint64 {
	return bits.RotateLeft64(h^x, 29) * 0x9e3779b97f4a7c15
}

// slot is where dh starts probing in a table of size len(tab).
func slot(dh uint64, size int) int { return int((dh ^ dh>>32) & uint64(size-1)) }

// intern is the one set with items' content, and whether it existed. A new
// set copies items, so the caller may reuse them.
func (in *interner) intern(items []Item) (*Set, bool) {
	ch, dh := uint64(hashSeed), uint64(hashSeed)
	for _, it := range items {
		ch, dh = mix(ch, uint64(it.State)), mix(dh, uint64(it.Dist))
	}
	ch, dh = ch&in.hashMask, dh&in.hashMask
	in.interns++
	in.lookups++
	head := in.cores[ch]
	co := head
	for co != nil && !sameStates(co.first.items, items) {
		co = co.next
	}
	if co == nil {
		co = &setCore{id: in.ncores, next: head, tab: make([]*Set, 4)}
		in.ncores++
		in.inserts++
		in.cores[ch] = co
	}
	i := slot(dh, len(co.tab))
	for ; co.tab[i] != nil; i = (i + 1) & (len(co.tab) - 1) {
		if s := co.tab[i]; s.dh == dh && slices.Equal(s.items, items) {
			return s, true
		}
	}
	s := &Set{id: in.nsets, items: slices.Clone(items), core: co, dh: dh}
	in.nsets++
	if co.first == nil {
		co.first = s
	}
	co.n++
	if 2*co.n > len(co.tab) { // keep the table at most half full
		old := co.tab
		co.tab = make([]*Set, 2*len(old))
		for _, o := range old {
			if o != nil {
				co.place(o)
			}
		}
		co.place(s)
	} else {
		co.tab[i] = s
	}
	return s, false
}

// place stores s in the first free slot from its distance hash.
func (co *setCore) place(s *Set) {
	i := slot(s.dh, len(co.tab))
	for co.tab[i] != nil {
		i = (i + 1) & (len(co.tab) - 1)
	}
	co.tab[i] = s
}

func sameStates(a, b []Item) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].State != b[i].State {
			return false
		}
	}
	return true
}
