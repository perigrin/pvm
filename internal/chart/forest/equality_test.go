// ABOUTME: Tests the forest-equality differential: perturbed forests reported, renumbered and unreachable changes not.
// ABOUTME: Runs the candidate against V1's pipeline over the toys, seeds 4 and 44, and a union-shortcut fixture.

package forest

import (
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/recognize"
)

// perturbationForest is an ambiguous toy forest with rule-position nodes: S ::= S S S | S S | A over four A's.
func perturbationForest(t *testing.T) *Forest {
	t.Helper()
	tab := toyTables(t, "S ::= S S S | S S | A")
	f, _, _ := toyForest(t, tab, []string{"A", "A", "A", "A"}, recognize.Options{})
	if f == nil {
		t.Fatal("the chart refuses A A A A")
	}
	return f
}

// reachableNodes is the indexes of f's nodes reachable from its root.
func reachableNodes(f *Forest) []int32 {
	seen := make([]bool, len(f.Nodes))
	var out []int32
	work := []int32{f.Root}
	for len(work) > 0 {
		n := work[len(work)-1]
		work = work[:len(work)-1]
		if seen[n] {
			continue
		}
		seen[n] = true
		out = append(out, n)
		for a := f.Nodes[n].Alt; a != 0; a = f.Alts[a].Next {
			for _, c := range []int32{f.Alts[a].Left, f.Alts[a].Right} {
				if c != 0 {
					work = append(work, c)
				}
			}
		}
	}
	return out
}

// dropAlt unlinks alternative a from node n's list.
func dropAlt(f *Forest, n, a int32) {
	p := &f.Nodes[n].Alt
	for *p != a {
		p = &f.Alts[*p].Next
	}
	*p = f.Alts[a].Next
}

// findDrop is a reachable node of f and one of its alternatives whose removal leaves the reachable node count
// unchanged (same), or lowers it (!same).
func findDrop(t *testing.T, f *Forest, same bool) (int32, int32) {
	t.Helper()
	before := len(reachableNodes(f))
	for _, n := range reachableNodes(f) {
		alts := keptAlts(f, int(n))
		if len(alts) < 2 {
			continue
		}
		for _, a := range alts {
			g := f.Copy()
			dropAlt(g, n, a)
			if (len(reachableNodes(g)) == before) == same {
				return n, a
			}
		}
	}
	t.Fatalf("no reachable node with an alternative whose removal keeps the node count same=%v", same)
	return 0, 0
}

// renumbered is f with its nodes and alternatives stored in reverse order, every index remapped.
func renumbered(f *Forest) *Forest {
	g := &Forest{Nodes: make([]Node, len(f.Nodes)), Alts: make([]Alt, len(f.Alts))}
	node := func(n int32) int32 { return map0(n, len(f.Nodes)) }
	alt := func(a int32) int32 { return map0(a, len(f.Alts)) }
	g.Root = node(f.Root)
	for n, nd := range f.Nodes[1:] {
		nd.Alt = alt(nd.Alt)
		g.Nodes[node(int32(n+1))] = nd
	}
	for a, x := range f.Alts[1:] {
		x.Left, x.Right, x.Next = node(x.Left), node(x.Right), alt(x.Next)
		g.Alts[alt(int32(a+1))] = x
	}
	return g
}

// map0 is index i of n reversed, 0 (none) kept: 1 becomes n-1.
func map0(i int32, n int) int32 {
	if i == 0 {
		return 0
	}
	return int32(n) - i
}

// TestForestEqualityDetectsPerturbation: ForestDiff reports each kind of difference through the comparison that
// owns it, and reports nothing for equal forests, a renumbered forest or a change no root derivation reaches.
func TestForestEqualityDetectsPerturbation(t *testing.T) {
	base := perturbationForest(t)
	reports := func(t *testing.T, cand, v1 *Forest, want string) {
		t.Helper()
		if d := ForestDiff(cand, v1); !strings.Contains(d, want) {
			t.Errorf("ForestDiff = %q, want a report containing %q", d, want)
		}
	}
	quiet := func(t *testing.T, cand, v1 *Forest) {
		t.Helper()
		if d := ForestDiff(cand, v1); d != "" {
			t.Errorf("ForestDiff = %q, want no difference", d)
		}
	}

	t.Run("equal", func(t *testing.T) { quiet(t, base.Copy(), base.Copy()) })

	t.Run("drop-alternative", func(t *testing.T) {
		g := base.Copy()
		n, a := findDrop(t, g, true)
		dropAlt(g, n, a)
		reports(t, g, base, "keeps")
		reports(t, base, g, "keeps")
	})

	t.Run("reorder-alternatives", func(t *testing.T) {
		g := base.Copy()
		n := slices.IndexFunc(reachableNodes(g), func(n int32) bool { return len(keptAlts(g, int(n))) >= 2 })
		if n < 0 {
			t.Fatal("no reachable node with two alternatives")
		}
		m := reachableNodes(g)[n]
		alts := keptAlts(g, int(m))
		g.Nodes[m].Alt = 0
		for _, a := range alts { // add puts each first, so the list ends reversed
			g.Nodes[m].Alt, g.Alts[a].Next = a, g.Nodes[m].Alt
		}
		if slices.Equal(keptAlts(g, int(m)), alts) {
			t.Fatal("the alternatives were not reordered")
		}
		reports(t, g, base, "keeps")
	})

	t.Run("add-reachable-node", func(t *testing.T) {
		g := base.Copy()
		leaf := g.Nodes[g.Alts[g.Nodes[g.Root].Alt].Right]
		leaf.From, leaf.To, leaf.Alt = 100, 101, 0 // a span no node of base has
		x := g.newNode(leaf)
		g.add(g.Root, Alt{Rule: g.Alts[g.Nodes[g.Root].Alt].Rule, Right: x})
		reports(t, g, base, "reachable in the candidate only")
	})

	t.Run("remove-reachable-node", func(t *testing.T) {
		g := base.Copy()
		n, a := findDrop(t, g, false)
		dropAlt(g, n, a)
		reports(t, g, base, "reachable in V1 only")
	})

	t.Run("flip-accept", func(t *testing.T) {
		reports(t, nil, base, "the candidate refuses")
		reports(t, base, nil, "V1 refuses")
	})

	t.Run("rule-position", func(t *testing.T) {
		at := slices.IndexFunc(reachableNodes(base), func(n int32) bool { return base.Nodes[n].Sym < 0 })
		if at < 0 {
			t.Fatal("no reachable rule-position node")
		}
		n := reachableNodes(base)[at]
		other := base.Nodes[n].Rule + 1 // any other rule
		for name, change := range map[string]func(*Node){
			"rule": func(nd *Node) { nd.Rule = other },
			"dot":  func(nd *Node) { nd.Dot-- },
		} {
			g := base.Copy()
			change(&g.Nodes[n])
			if g.Nodes[n].Sym != base.Nodes[n].Sym || g.Nodes[n].From != base.Nodes[n].From || g.Nodes[n].To != base.Nodes[n].To {
				t.Fatalf("%s: the symbol or span moved", name)
			}
			if d := ForestDiff(g, base); !strings.Contains(d, "reachable") {
				t.Errorf("%s changed on node %d: ForestDiff = %q, want the node reported", name, n, d)
			}
		}
	})

	t.Run("renumbered-nodes-equal", func(t *testing.T) {
		g := renumbered(base)
		if g.Root == base.Root {
			t.Fatalf("the root kept index %d", g.Root)
		}
		quiet(t, g, base)
		quiet(t, base, g)
	})

	t.Run("unreachable-node-ignored", func(t *testing.T) {
		g := base.Copy()
		x := g.newNode(Node{Sym: g.Nodes[g.Root].Sym, Rule: -1, From: 0, To: g.Nodes[g.Root].To + 1})
		g.add(x, Alt{Rule: g.Alts[g.Nodes[g.Root].Alt].Rule, Left: g.Root, Right: g.Root})
		if slices.Contains(reachableNodes(g), x) {
			t.Fatal("the added node is reachable")
		}
		quiet(t, g, base)
	})
}
