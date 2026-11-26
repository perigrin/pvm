// ABOUTME: The forest-equality differential: a candidate's forest against V1's pipeline, accept or refuse and reachable nodes.
// ABOUTME: Nodes are matched by symbol, rule position and span; the V1 side, the candidate and ForestDiff are exported for reuse.

package forest

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
)

// NodeKey is what a node derives, by which ForestDiff matches nodes across forests: its symbol (-1 for a rule
// position), its rule position (Rule and Dot; Rule -1 for a symbol) and its span.
type NodeKey struct{ Sym, Rule, Dot, From, To int32 }

// noNode is the key of an absent child, index 0, which no node's key equals.
var noNode = NodeKey{-1, -1, -1, -1, -1}

// altKey is one kept alternative by what it derives: its rule and its children's keys.
type altKey struct {
	Rule        int32
	Left, Right NodeKey
}

// ForestDiff is where cand's forest differs from v1's, nil for a refusal: the verdict, then a node reachable from
// the root in one only, then the first reachable node, by key, whose kept alternatives differ in content or order;
// "" where they are equal. A node unreachable from the root is never compared.
func ForestDiff(cand, v1 *Forest) string {
	switch {
	case cand == nil && v1 == nil:
		return ""
	case cand == nil:
		return "the candidate refuses, V1 accepts"
	case v1 == nil:
		return "the candidate accepts, V1 refuses"
	}
	c, v := reachableAlts(cand), reachableAlts(v1)
	keys := slices.SortedFunc(maps.Keys(c), cmpNodeKey)
	for _, k := range slices.SortedFunc(maps.Keys(v), cmpNodeKey) {
		if _, ok := c[k]; !ok {
			return fmt.Sprintf("node %+v is reachable in V1 only", k)
		}
	}
	for _, k := range keys {
		if _, ok := v[k]; !ok {
			return fmt.Sprintf("node %+v is reachable in the candidate only", k)
		}
	}
	for _, k := range keys {
		if !slices.Equal(c[k], v[k]) {
			return fmt.Sprintf("node %+v: the candidate keeps %+v, V1 %+v", k, c[k], v[k])
		}
	}
	return ""
}

// reachableAlts is each node reachable from f's root, by key, with its kept alternatives in order.
func reachableAlts(f *Forest) map[NodeKey][]altKey {
	key := func(n int32) NodeKey {
		if n == 0 {
			return noNode
		}
		nd := f.Nodes[n]
		return NodeKey{nd.Sym, nd.Rule, nd.Dot, nd.From, nd.To}
	}
	out := map[NodeKey][]altKey{}
	seen := make([]bool, len(f.Nodes))
	work := []int32{f.Root}
	for len(work) > 0 {
		n := work[len(work)-1]
		work = work[:len(work)-1]
		if seen[n] {
			continue
		}
		seen[n] = true
		alts := []altKey{}
		for a := f.Nodes[n].Alt; a != 0; a = f.Alts[a].Next {
			x := f.Alts[a]
			alts = append(alts, altKey{x.Rule, key(x.Left), key(x.Right)})
			for _, ch := range []int32{x.Left, x.Right} {
				if ch != 0 {
					work = append(work, ch)
				}
			}
		}
		out[key(n)] = alts
	}
	return out
}

func cmpNodeKey(x, y NodeKey) int {
	return cmp.Or(cmp.Compare(x.From, y.From), cmp.Compare(x.To, y.To), cmp.Compare(x.Sym, y.Sym), cmp.Compare(x.Rule, y.Rule), cmp.Compare(x.Dot, y.Dot))
}
