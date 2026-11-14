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
	id   int
	sets []*Set   // interned sets of this core, in creation order
	next *setCore // the next core under the same hash
	off  []int32  // by symbol, into ent; nil until first used
	ent  []coreEnt
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

// interner is YAEP's: a set is its core (the states) and its distances. One
// hash of the states keys a table of cores, each holding its sets, which
// carry a hash of their distances. Hashes only narrow the search: items
// decide identity, so a collision never merges two sets.
type interner struct {
	cores         map[uint64]*setCore
	nsets, ncores int
	// hashMask is all ones; a test clears it to force every hash to collide.
	hashMask uint64
	// interns and lookups count intern calls and the map reads they made.
	interns, lookups int
}

func newInterner() interner {
	return interner{cores: map[uint64]*setCore{}, hashMask: ^uint64(0)}
}

// mix folds x into h.
func mix(h, x uint64) uint64 {
	return bits.RotateLeft64(h^x, 29) * 0x9e3779b97f4a7c15
}

// intern is the one set with items' content, and whether it existed. A new
// set copies items, so the caller may reuse them.
func (in *interner) intern(items []Item) (*Set, bool) {
	var ch, dh uint64
	for _, it := range items {
		ch, dh = mix(ch, uint64(it.State)), mix(dh, uint64(it.Dist))
	}
	ch, dh = ch&in.hashMask, dh&in.hashMask
	in.interns++
	in.lookups++
	head := in.cores[ch]
	co := head
	for co != nil && !sameStates(co.sets[0].items, items) {
		co = co.next
	}
	if co == nil {
		co = &setCore{id: in.ncores, next: head}
		in.ncores++
		in.cores[ch] = co
	}
	// shortcut: a linear scan of the core's sets by distance hash, 453 long at most over corpusSample; index them if a profile shows it.
	for _, s := range co.sets {
		if s.dh == dh && slices.Equal(s.items, items) {
			return s, true
		}
	}
	s := &Set{id: in.nsets, items: slices.Clone(items), core: co, dh: dh}
	in.nsets++
	co.sets = append(co.sets, s)
	return s, false
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
