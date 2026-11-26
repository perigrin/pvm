// ABOUTME: Terminal-set ids for the sets a narrowed lattice needs that the committed tables lack, from NarrowedBase up.
// ABOUTME: Kept beside one tables value, shared by its recognizers, never reassigned, never written into tab.TerminalSets.

package recognize

import (
	"encoding/binary"
	"slices"
	"sync"
	"sync/atomic"

	"tamarou.com/pvm/internal/chart/grammar"
)

// NarrowedBase is the first id SetID gives a set the tables lack: far above any table's ids, so a set appended to
// tab.TerminalSets later never takes one, and an id never indexes tab.TerminalSets.
const NarrowedBase = 1 << 30

// narrowedSets is the sets SetID named that the tables lack, beside one tables value: their ids by content, and
// their classes, published as extend publishes the committed sets', index id-NarrowedBase. An id, once given, names
// its set for the tables' life, so no recognizer's intern table or transition cache, which key on set ids, can serve
// one set's transition for another's.
type narrowedSets struct {
	mu        sync.Mutex
	committed map[string]int // the tables' own sets by content, the first id of each, read at the first SetID
	ids       map[string]int
	cls       atomic.Pointer[classes]
}

// SetID is the terminal-set id of terms, a sorted set of tab's terminals: the tables' own id where they have the
// set, else one from NarrowedBase, the same for every call over tab. Recognize reads either; tab is not written. It
// is safe from any goroutine.
func SetID(tab *grammar.Tables, terms []int) int {
	g := grammar.Derived(tab, buildGrammar)
	n := &g.narrowed
	key := setKey(terms)
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.committed == nil {
		n.committed, n.ids = map[string]int{}, map[string]int{}
		for id, set := range tab.TerminalSets {
			if _, ok := n.committed[setKey(set)]; !ok {
				n.committed[setKey(set)] = id
			}
		}
	}
	if id, ok := n.committed[key]; ok {
		return id
	}
	if id, ok := n.ids[key]; ok {
		return id
	}
	old := n.cls.Load()
	// Clipped, so append copies rather than write past what old's readers see.
	c := &classes{terms: slices.Clip(old.terms), bits: slices.Clip(old.bits), closes: slices.Clip(old.closes)}
	g.add(c, slices.Clone(terms))
	id := NarrowedBase + len(old.terms)
	n.ids[key] = id
	n.cls.Store(c)
	return id
}

// setKey is terms as a map key.
func setKey(terms []int) string {
	b := make([]byte, 0, 2*len(terms))
	for _, t := range terms {
		b = binary.AppendUvarint(b, uint64(t))
	}
	return string(b)
}

// class is the classes holding set id and its index there: the committed sets', or SetID's from NarrowedBase.
func (r *Recognizer) class(id int) (*classes, int) {
	if id < NarrowedBase {
		return r.cls, id
	}
	return r.ncls, id - NarrowedBase
}

// Terminals is the terminals of set id, an id of c's lattice: the tables' set, or SetID's.
func (c *Chart) Terminals(id int) []int {
	cls, i := c.r.class(id)
	return cls.terms[i]
}

// Entries is how many sets r has interned and transitions it has cached, which only grow: equal counts before and
// after a call are a call that added none.
func (r *Recognizer) Entries() (sets, transitions int) { return r.sets.nsets, r.stored }
