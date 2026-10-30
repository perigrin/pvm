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
	cores, dists map[string]int
	sets         map[[2]int]*Set
}

func newInterner() interner {
	return interner{cores: map[string]int{}, dists: map[string]int{}, sets: map[[2]int]*Set{}}
}

// intern is the one set with items' content, and whether it existed.
func (in *interner) intern(items []Item) (*Set, bool) {
	var core, dist []byte
	for _, it := range items {
		core = binary.AppendUvarint(core, uint64(it.State))
		dist = binary.AppendUvarint(dist, uint64(it.Dist))
	}
	k := [2]int{idOf(in.cores, string(core)), idOf(in.dists, string(dist))}
	if s, ok := in.sets[k]; ok {
		return s, true
	}
	s := &Set{id: len(in.sets), items: items}
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
