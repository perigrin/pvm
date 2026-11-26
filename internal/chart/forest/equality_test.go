// ABOUTME: Tests the forest-equality differential: perturbed forests reported, renumbered and unreachable changes not.
// ABOUTME: Runs the candidate against V1's pipeline over the toys, seeds 4 and 44, and a union-shortcut fixture.

package forest

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
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

	tab := committedTables(t)
	sides := func(t *testing.T, src string) (cand, v1 *Forest) {
		t.Helper()
		cand, _ = CandidateForest(tab, []byte(src), Scope{})
		v1, _ = V1Forest(tab, []byte(src), Scope{})
		return cand, v1
	}

	t.Run("empty-input", func(t *testing.T) {
		cand, v1 := sides(t, "")
		if v1 == nil {
			t.Fatal("V1 refuses the empty input")
		}
		quiet(t, cand, v1)
	})

	t.Run("both-refuse", func(t *testing.T) {
		for _, src := range []string{"1 +;", "1 .. 2 .. 3;"} { // refused by the chart, and by Constrain
			cand, v1 := sides(t, src)
			if cand != nil || v1 != nil {
				t.Fatalf("%q: candidate forest %v, V1 forest %v; want both refused", src, cand != nil, v1 != nil)
			}
			quiet(t, cand, v1)
		}
	})
}

// precedenceRefusals are three inputs V1 refuses only by precedence: recognition accepts each and Parse refuses it
// as precedence (measured at e3432c0f).
var precedenceRefusals = []struct{ name, src string }{
	{"indirob", "print $fh [];"},
	{"nceqop-chain", "$a <=> $b <=> $c;"},
	{"dotdot-chain", "1 .. 2 .. 3;"},
}

// TestForestEqualityV1SideRecognizesPrecedenceRefusals: the V1 side's chart accepts each input V1 refuses only by
// precedence, and its own Constrain refuses it, so a V1 side whose recognizer judged precedence would refuse at
// the chart, as grammar, and fail here.
func TestForestEqualityV1SideRecognizesPrecedenceRefusals(t *testing.T) {
	tab := committedTables(t)
	for _, in := range precedenceRefusals {
		t.Run(in.name, func(t *testing.T) {
			c, err := recognize.Recognize(V1Recognizer(tab), lex.Linear(lexOnly(tab, in.src)))
			if err != nil {
				t.Fatal(err)
			}
			if !c.Accepted() {
				t.Errorf("%q: the V1 recognizer's chart refuses it", in.src)
			}
			if f, ref := V1Forest(tab, []byte(in.src), Scope{}); f != nil || ref.Reason != "precedence" {
				t.Errorf("%q: V1Forest = (forest %v, %+v), want no forest and a precedence refusal", in.src, f != nil, ref)
			}
		})
	}
}

// lexOnly is src's token stream over tab.
func lexOnly(tab *grammar.Tables, src string) *lex.Stream {
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	return s
}

// TestForestEqualityV1SideRunsOldPipeline: V1Forest enters its own LibraryOperators and Constrain once per input the
// chart accepts, and V1ToyForest its own Constrain once per toy sentence the chart accepts, counted through the V1
// side's stage hooks, so a V1 side wired to production's output, which never passes those hooks, fails.
func TestForestEqualityV1SideRunsOldPipeline(t *testing.T) {
	libs, cons := 0, 0
	lib0, con0 := v1LibraryOperators, v1Constrain
	t.Cleanup(func() { v1LibraryOperators, v1Constrain = lib0, con0 })
	v1LibraryOperators = func(tab *grammar.Tables, f *Forest, s *lex.Stream, src []byte, scope Scope) int {
		libs++
		return lib0(tab, f, s, src, scope)
	}
	v1Constrain = func(f *Forest, p *Precedence) bool {
		cons++
		return con0(f, p)
	}

	tab := committedTables(t)
	srcs := []string{"1;", "$x = 1 + 2 * 3;", "f(1, 2);", "print $fh [];", "1 .. 2 .. 3;"}
	accepted := 0
	for _, src := range srcs {
		if f, _ := V1Forest(tab, []byte(src), Scope{}); f != nil {
			accepted++
		}
	}
	if libs != len(srcs) || cons != len(srcs) {
		t.Errorf("over %d inputs the chart accepts: LibraryOperators entered %d times, Constrain %d; want %d each", len(srcs), libs, cons, len(srcs))
	}
	if accepted == 0 || accepted == len(srcs) {
		t.Errorf("V1Forest accepted %d of %d; want some accepted and some refused by Constrain", accepted, len(srcs))
	}

	ty := loadToy(t, "ladder")
	cons = 0
	for _, s := range ty.sentences[:20] {
		if _, err := V1ToyForest(ty.tab, toyLattice(ty.tab, s.words), ty.prec); err != nil {
			t.Fatal(err)
		}
	}
	want := 0
	for _, s := range ty.sentences[:20] {
		if f, _, _ := toyForest(t, ty.tab, s.words, recognize.Options{}); f != nil {
			want++
		}
	}
	if want == 0 || cons != want {
		t.Errorf("V1ToyForest entered Constrain %d times over %d ladder sentences the chart accepts; want one each", cons, want)
	}
}

// toyDiffs compares the candidate's forest of each of ty's sentences with V1's under ty's precedence: how many it
// compared, how many V1 accepts, and each difference.
func toyDiffs(t *testing.T, ty toy) (compared, accepted int, diffs []string) {
	t.Helper()
	for _, s := range ty.sentences {
		l := toyLattice(ty.tab, s.words)
		cand, err := CandidateToyForest(ty.tab, l, ty.prec)
		if err != nil {
			t.Fatalf("%v: candidate: %v", s.words, err)
		}
		v1, err := V1ToyForest(ty.tab, l, ty.prec)
		if err != nil {
			t.Fatalf("%v: V1: %v", s.words, err)
		}
		compared++
		if v1 != nil {
			accepted++
		}
		if d := ForestDiff(cand, v1); d != "" {
			diffs = append(diffs, strings.Join(s.words, " ")+": "+d)
		}
	}
	return compared, accepted, diffs
}

// sentenceDiffs compares the candidate's forest of each source with V1's under the empty scope: how many it
// compared, how many V1 accepts, and each difference.
func sentenceDiffs(tab *grammar.Tables, srcs []string) (compared, accepted int, diffs []string) {
	for _, src := range srcs {
		cand, _ := CandidateForest(tab, []byte(src), Scope{})
		v1, _ := V1Forest(tab, []byte(src), Scope{})
		compared++
		if v1 != nil {
			accepted++
		}
		if d := ForestDiff(cand, v1); d != "" {
			diffs = append(diffs, fmt.Sprintf("%q: %s", src, d))
		}
	}
	return compared, accepted, diffs
}

// reportDiffs fails t with the first ten of diffs and their count over compared.
func reportDiffs(t *testing.T, what string, compared int, diffs []string) {
	t.Helper()
	for _, d := range diffs[:min(10, len(diffs))] {
		t.Errorf("%s %s", what, d)
	}
	if len(diffs) > 0 {
		t.Errorf("%s: %d of %d forests differ from V1's", what, len(diffs), compared)
	}
}

// TestForestEqualityToys: on every toy fixture sentence, the candidate's toy forest under the toy's own precedence
// equals V1's. It logs how many it compared of how many it loaded, and requires each toy to have sentences V1
// accepts and sentences it refuses, so the comparison is not of refusals alone.
func TestForestEqualityToys(t *testing.T) {
	loaded, compared := 0, 0
	for _, name := range toyNames {
		ty := loadToy(t, name)
		n, accepted, diffs := toyDiffs(t, ty)
		loaded += len(ty.sentences)
		compared += n
		if accepted == 0 || accepted == n {
			t.Errorf("%s: V1 accepts %d of %d sentences; want some of each verdict", name, accepted, n)
		}
		reportDiffs(t, name, n, diffs)
	}
	t.Logf("toys: compared %d of %d sentences", compared, loaded)
	if compared == 0 || compared != loaded {
		t.Errorf("compared %d of %d toy sentences; want every one, and some", compared, loaded)
	}
}

// TestForestEqualityGeneratedSentences: decision §2's seed-4 and seed-44 generated sentences, the candidate's
// forest against V1's on each. Each seed logs how many it compared of how many it generated, and requires
// sentences V1 accepts and sentences it refuses.
func TestForestEqualityGeneratedSentences(t *testing.T) {
	tab := committedTables(t)
	for _, seed := range []struct {
		name string
		srcs []string
	}{{"4", operatorSentences()}, {"44", chainSentences()}} {
		t.Logf("seed %s: %d sentences", seed.name, len(seed.srcs))
		compared, accepted, diffs := sentenceDiffs(tab, seed.srcs)
		t.Logf("seed %s: compared %d of %d sentences, V1 accepts %d", seed.name, compared, len(seed.srcs), accepted)
		if compared == 0 || compared != len(seed.srcs) {
			t.Errorf("seed %s: compared %d of %d; want every one, and some", seed.name, compared, len(seed.srcs))
		}
		if accepted == 0 || accepted == compared {
			t.Errorf("seed %s: V1 accepts %d of %d; want some of each verdict", seed.name, accepted, compared)
		}
		reportDiffs(t, "seed "+seed.name, compared, diffs)
	}
}

// TestForestEqualityHarnessesUseCandidate: with the candidate swapped for one that refuses everything, the toy and
// sentence loops report every input V1 accepts, so neither compares V1 with itself.
func TestForestEqualityHarnessesUseCandidate(t *testing.T) {
	src0, toy0 := CandidateForest, CandidateToyForest
	t.Cleanup(func() { CandidateForest, CandidateToyForest = src0, toy0 })
	CandidateForest = func(*grammar.Tables, []byte, Scope) (*Forest, Refusal) { return nil, Refusal{Reason: "test"} }
	CandidateToyForest = func(*grammar.Tables, *lex.Lattice[lex.None], *Precedence) (*Forest, error) { return nil, nil }

	if _, accepted, diffs := sentenceDiffs(committedTables(t), []string{"1;", "$x = 1 + 2;", "1 +;"}); accepted != 2 || len(diffs) != accepted {
		t.Errorf("sentences: V1 accepts %d, %d reported; want 2 accepted and each reported", accepted, len(diffs))
	}
	if _, accepted, diffs := toyDiffs(t, loadToy(t, "ladder")); accepted == 0 || len(diffs) != accepted {
		t.Errorf("toys: V1 accepts %d, %d reported; want each accepted sentence reported", accepted, len(diffs))
	}
}

// contextAltSets is, for each node of f, the alternatives each of its contexts keeps on its own under V1's
// judgement by p, in the order of the node's list: a context is eval's (the terminal after the node, its spine top,
// its parent item), and what it keeps is mark's choice over every first terminal it was marked with. Constrain
// keeps the union of a node's sets (its shortcut:); this reads them apart. f is not changed.
func contextAltSets(f *Forest, p *Precedence) map[int32][][]int32 {
	j := newJudgement(f, p)
	if len(j.eval(f.Root, p.tab.EOF, rootTop, -1)) == 0 {
		return nil
	}
	j.mark(f.Root, p.tab.EOF, rootTop, -1, -1)
	out := map[int32][][]int32{}
	for n := range f.Nodes {
		type ctx struct{ t, top, pi int32 }
		var order []ctx
		sets := map[ctx][]int32{}
		for e := j.markAt[n]; e != 0; e = j.marks[e].next {
			m := j.marks[e]
			k := ctx{m.t, m.top, m.pi}
			if _, ok := sets[k]; !ok {
				order = append(order, k)
				sets[k] = []int32{}
			}
			keep := func(x int32) bool { return m.first < 0 || x == m.first }
			for a := f.Nodes[n].Alt; a != 0; a = f.Alts[a].Next { // mark's loop, without marking
				r, mm, at := j.alt(int32(n), a, m.top)
				if f.Nodes[n].Sym >= 0 && !p.reducesFree(r, int(m.t), m.pi) {
					continue
				}
				if !slices.Contains(sets[k], a) && slices.ContainsFunc(j.seq(a, r, mm, int(m.t), at, nil), keep) {
					sets[k] = append(sets[k], a)
				}
			}
		}
		for _, k := range order {
			out[int32(n)] = append(out[int32(n)], sets[k])
		}
	}
	return out
}

// shortcutNode is the first node of f at which two contexts keep two different non-empty sets of alternatives
// under p, with every context's set there and how many distinct non-empty sets they hold; -1 where there is none.
func shortcutNode(f *Forest, p *Precedence) (n int32, sets [][]int32, distinct int) {
	all := contextAltSets(f, p)
	for _, n := range slices.Sorted(maps.Keys(all)) {
		var seen [][]int32
		for _, s := range all[n] {
			if len(s) > 0 && !slices.ContainsFunc(seen, func(d []int32) bool { return sameAlts(d, s) }) {
				seen = append(seen, s)
			}
		}
		if len(seen) >= 2 {
			return n, all[n], len(seen)
		}
	}
	return -1, nil, 0
}

// unionDiff is where judged, Constrain's judgement of a copy of built, keeps on some node other than the union of
// that node's contexts' sets in built under p, in the node's order; "" where it keeps the union on every node.
func unionDiff(built, judged *Forest, p *Precedence) string {
	all := contextAltSets(built, p)
	for n := range built.Nodes {
		var union []int32
		for _, a := range keptAlts(built, n) {
			if slices.ContainsFunc(all[int32(n)], func(s []int32) bool { return slices.Contains(s, a) }) {
				union = append(union, a)
			}
		}
		if !slices.Equal(keptAlts(judged, n), union) {
			return fmt.Sprintf("node %d: Constrain keeps %v, the union of its contexts' sets is %v", n, keptAlts(judged, n), union)
		}
	}
	return ""
}

// sameAlts is whether a and b hold the same alternatives, in any order.
func sameAlts(a, b []int32) bool {
	return len(a) == len(b) && !slices.ContainsFunc(a, func(x int32) bool { return !slices.Contains(b, x) })
}

// unionShortcutInput is a constructed input on which V1 keeps a node's alternatives under two contexts with two
// different non-empty sets, the union Constrain's shortcut: takes (C5): the term over `isa my $m | 1` keeps one
// alternative in one context and two in the other. No corpus file exercises it.
const unionShortcutInput = "STDOUT isa my $m | 1;"

// TestForestEqualityUnionShortcut: on unionShortcutInput, V1's judgement keeps one node's alternatives under two
// contexts with two different non-empty sets, which it logs; V1Forest keeps on every node the union of its
// contexts' sets, so the measurement reads Constrain's own choice; and the candidate's forest equals V1's.
func TestForestEqualityUnionShortcut(t *testing.T) {
	tab := committedTables(t)
	prec, err := ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	src := []byte(unionShortcutInput)
	s, _ := lex.Lex(src, tab.Tokens)
	l := lex.Linear(s)
	c, err := recognize.Recognize(V1Recognizer(tab), l)
	if err != nil || !c.Accepted() {
		t.Fatalf("%q: the V1 chart refuses it (%v)", src, err)
	}
	built := Build(tab, c, l)
	if LibraryOperators(tab, built, s, src, Scope{}) >= 0 {
		t.Fatalf("%q: LibraryOperators refuses it", src)
	}
	n, sets, distinct := shortcutNode(built, prec)
	if n < 0 {
		t.Fatalf("%q: no node keeps different non-empty alternative sets under two contexts", src)
	}
	nd := built.Nodes[n]
	name := "a rule position"
	if nd.Sym >= 0 {
		name = tab.Symbols[nd.Sym].Name
	}
	t.Logf("%q: node %d (%s, %d-%d), alternatives %v: contexts: %d, distinct non-empty alternative sets: %d, sets %v",
		src, n, name, nd.From, nd.To, keptAlts(built, int(n)), len(sets), distinct, sets)
	if len(sets) != 2 || distinct != 2 {
		t.Errorf("contexts %d, distinct non-empty alternative sets %d; want 2 and 2", len(sets), distinct)
	}

	v1, ref := V1Forest(tab, src, Scope{})
	if v1 == nil {
		t.Fatalf("%q: V1 refuses it: %+v", src, ref)
	}
	if g := v1.Nodes[n]; g.Sym != nd.Sym || g.Rule != nd.Rule || g.Dot != nd.Dot || g.From != nd.From || g.To != nd.To {
		t.Fatalf("V1Forest's node %d is %+v, the built forest's %+v", n, g, nd)
	}
	if d := unionDiff(built, v1, prec); d != "" {
		t.Errorf("%q: %s", src, d)
	}

	// The measurement is held to Constrain beyond this input: every toy sentence and the 83 handcrafted operator
	// sentences V1 accepts, `_ - $h{1} - $fh;`, where its reduce check decides, and TestPrecedenceMarkFirstFilter's
	// toy, where mark's first-terminal filter does.
	checked := 0
	ftab, fp := markFirstFilterToy(t)
	fl := toyLattice(ftab, []string{"C", "A", "Q1"})
	ftab.TerminalSets = append(ftab.TerminalSets, []int{symbol(ftab, "A"), symbol(ftab, "B")})
	fl.Edges[1].Token.TerminalSet = len(ftab.TerminalSets) - 1
	fv1, err := V1ToyForest(ftab, fl, fp)
	fc, cerr := recognize.Recognize(V1Recognizer(ftab), fl)
	if err != nil || cerr != nil || fv1 == nil {
		t.Fatalf("the mark-first-filter toy: V1 refuses C {A B} Q1 (%v, %v)", err, cerr)
	}
	if d := unionDiff(Build(ftab, fc, fl), fv1, fp); d != "" {
		t.Errorf("the mark-first-filter toy: %s", d)
	}
	checked++
	for _, name := range toyNames {
		ty := loadToy(t, name)
		for _, s := range ty.sentences {
			l := toyLattice(ty.tab, s.words)
			v1, err := V1ToyForest(ty.tab, l, ty.prec)
			if err != nil || v1 == nil {
				continue
			}
			b, _, _ := toyForest(t, ty.tab, s.words, recognize.Options{})
			if d := unionDiff(b, v1, ty.prec); d != "" {
				t.Errorf("%s %s: %s", name, strings.Join(s.words, " "), d)
			}
			checked++
		}
	}
	for _, src := range append(slices.Clone(operatorCrafted), "_ - $h{1} - $fh;") {
		v1, _ := V1Forest(tab, []byte(src), Scope{})
		if v1 == nil {
			continue
		}
		s, _ := lex.Lex([]byte(src), tab.Tokens)
		l := lex.Linear(s)
		c, _ := recognize.Recognize(V1Recognizer(tab), l)
		b := Build(tab, c, l)
		LibraryOperators(tab, b, s, []byte(src), Scope{})
		if d := unionDiff(b, v1, prec); d != "" {
			t.Errorf("%q: %s", src, d)
		}
		checked++
	}
	t.Logf("the per-context measurement matches Constrain's kept set on every node of %d accepted inputs", checked)

	cand, _ := CandidateForest(tab, src, Scope{})
	if d := ForestDiff(cand, v1); d != "" {
		t.Errorf("%q: %s", src, d)
	}
}
