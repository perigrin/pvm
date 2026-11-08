// ABOUTME: The transition cache: (set, token, lookahead) to the next set, valid while the origin sets it read match.
// ABOUTME: Completion reads the sets at item origins, so an entry records them and a hit checks them, along a trie.

package recognize

import "slices"

// cacheKey is YAEP's (set, token, lookahead), with the edge's span in nodes,
// which turns the source set's distances into the new node's.
type cacheKey struct {
	set, span, tok, la int
}

// dep is an origin set a transition's completion read: the set's id, at
// dist nodes back from the new node.
type dep struct {
	dist, set int
}

// cacheNode is a node of a trie over a key's entries' dependencies, in
// the order they were read. The key alone does not fix a result:
// completion reads the sets at the origins of complete items, and those
// are not part of the source set's content. So an entry holds for a node
// only where those sets are the ones it read. Which set a transition reads
// next follows from the key and the sets already read, so a trie node
// normally has one distance: a lookup reads one set per dependency and
// follows its id (cacheEdge), instead of checking every entry (a key's
// entries grew by one per flat statement).
type cacheNode struct {
	seq    int      // the order of the entry ending here among the stores, 0 for none
	result *Set     // the transition ending here
	names  []string // or the Expected ending here (expected.go)
	dists  []int    // the distances of the next dependencies
	// first is the first edge stored from here, and where it leads, checked
	// before the trie map: most nodes have one edge, and a map access costs
	// more than the scan it replaces on a short list (std5k.pl).
	first   cacheEdge
	firstTo *cacheNode
}

// cacheEdge is a trie edge: from a node, the set with this id at dist nodes back.
type cacheEdge struct {
	from     *cacheNode
	dist, id int
}

// lookup is the cached set for key at node pos, or nil.
func (r *Recognizer) lookup(c *Chart, pos int, key cacheKey) *Set {
	if n := r.hit(c, pos, key); n != nil {
		return n.result
	}
	return nil
}

// hit is the entry for key at node pos, or nil: of the entries whose
// dependencies hold, the one stored first.
func (r *Recognizer) hit(c *Chart, pos int, key cacheKey) *cacheNode {
	root := r.cache[key]
	if root == nil {
		return nil
	}
	var hit *cacheNode
	scanned := 0
	r.find(c, pos, root, &hit, &scanned)
	r.longestScan = max(r.longestScan, scanned)
	return hit
}

// find walks every edge from n whose set holds at pos, keeping in hit the
// entry stored first, and counts the distances it examined in scanned.
func (r *Recognizer) find(c *Chart, pos int, n *cacheNode, hit **cacheNode, scanned *int) {
	if n.seq != 0 && (*hit == nil || n.seq < (*hit).seq) {
		*hit = n
	}
	for _, d := range n.dists {
		*scanned++
		at := pos - d
		if at < 0 || c.Sets[at] == nil {
			continue
		}
		if child := r.edge(cacheEdge{n, d, c.Sets[at].id}); child != nil {
			r.find(c, pos, child, hit, scanned)
		}
	}
}

// edge is where e leads, or nil.
func (r *Recognizer) edge(e cacheEdge) *cacheNode {
	if e.from.firstTo != nil && e.from.first == e {
		return e.from.firstTo
	}
	return r.trie[e]
}

// store caches result for key, read from deps.
func (r *Recognizer) store(key cacheKey, deps []dep, result *Set) {
	if n := r.slot(key, deps); n != nil {
		n.result = result
	}
}

// slot is the trie node where key's entry read from deps ends, for the
// caller to fill, or nil where an earlier entry has the same dependencies:
// that one is found first, so a second is never reached.
func (r *Recognizer) slot(key cacheKey, deps []dep) *cacheNode {
	n := r.cache[key]
	if n == nil {
		n = &cacheNode{}
		r.cache[key] = n
	}
	for _, d := range deps {
		if d.dist == key.span && d.set == key.set {
			continue // the source set, which the key already fixes
		}
		if !slices.Contains(n.dists, d.dist) {
			n.dists = append(n.dists, d.dist)
		}
		e := cacheEdge{n, d.dist, d.set}
		child := r.edge(e)
		if child == nil {
			child = &cacheNode{}
			if n.firstTo == nil {
				n.first, n.firstTo = e, child
			} else {
				r.trie[e] = child
			}
		}
		n = child
	}
	if n.seq != 0 {
		return nil
	}
	r.stored++
	n.seq = r.stored
	return n
}
