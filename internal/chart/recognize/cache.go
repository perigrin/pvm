// ABOUTME: The transition cache: (set, token, lookahead) to the next set, valid while the origin sets it read match.
// ABOUTME: Completion reads the sets at item origins, so an entry records them and a hit checks them.

package recognize

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

// cacheEntry is one result for a key. The key alone does not fix it:
// completion reads the sets at the origins of complete items, and those
// are not part of the source set's content. So an entry holds for a node
// only where those sets are the ones it read.
type cacheEntry struct {
	deps   []dep
	result *Set
}

// lookup is the cached set for key at node pos, or nil.
func (r *Recognizer) lookup(c *Chart, pos int, key cacheKey) *Set {
	for _, e := range r.cache[key] {
		if depsHold(c, pos, e.deps) {
			return e.result
		}
	}
	return nil
}

func depsHold(c *Chart, pos int, deps []dep) bool {
	for _, d := range deps {
		at := pos - d.dist
		if at < 0 || c.Sets[at] == nil || c.Sets[at].id != d.set {
			return false
		}
	}
	return true
}

// store caches result for key, read from deps.
func (r *Recognizer) store(key cacheKey, deps []dep, result *Set) {
	r.cache[key] = append(r.cache[key], cacheEntry{deps: deps, result: result})
}
