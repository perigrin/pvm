// ABOUTME: Tests the forest against a brute-force enumerator over toy grammars' rules, and over perly.y on a corpus sample.
// ABOUTME: Leo on and off give one forest, a slipper ; is a zero-width leaf, refused charts have none, enumeration is bounded.

package forest

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
)

// toyTables is the tables of a toy grammar, one string per statement:
// "LHS ::= alt | alt", symbols space-separated, an empty alternative empty,
// `X*` a sequence. A symbol no statement defines is a terminal. Its terminal
// sets are the empty set at id 0, then one singleton set per terminal.
func toyTables(t *testing.T, statements ...string) *grammar.Tables {
	t.Helper()
	g := &grammar.Grammar{}
	defined := map[string]bool{}
	for _, st := range statements {
		lhs, rhs, _ := strings.Cut(st, "::=")
		lhs = strings.TrimSpace(lhs)
		defined[lhs] = true
		for _, alt := range strings.Split(rhs, "|") {
			r := grammar.Rule{LHS: lhs}
			for _, w := range strings.Fields(alt) {
				if base, ok := strings.CutSuffix(w, "*"); ok {
					r.Seq, w = grammar.Star, base
				}
				r.RHS = append(r.RHS, grammar.Symbol{Kind: grammar.Name, Text: w})
			}
			g.G1 = append(g.G1, r)
		}
	}
	seen := map[string]bool{}
	for _, r := range g.G1 {
		for _, sym := range r.RHS {
			if !defined[sym.Text] && !seen[sym.Text] {
				seen[sym.Text] = true
				g.G0 = append(g.G0, grammar.Rule{LHS: sym.Text})
			}
		}
	}
	tab, err := grammar.BuildTables(g)
	if err != nil {
		t.Fatal(err)
	}
	tab.TerminalSets = [][]int{{}}
	for id, s := range tab.Symbols {
		if s.Terminal {
			tab.TerminalSets = append(tab.TerminalSets, []int{id})
		}
	}
	return tab
}

// toyLattice is words as a linear lattice of singleton terminal sets; a word
// that is no terminal gets the empty set.
func toyLattice(tab *grammar.Tables, words []string) *lex.Lattice[lex.None] {
	l := &lex.Lattice[lex.None]{Nodes: make([]lex.Node[lex.None], len(words)+1)}
	for i, w := range words {
		id := max(0, slices.IndexFunc(tab.TerminalSets, func(s []int) bool { return len(s) == 1 && tab.Symbols[s[0]].Name == w }))
		l.Edges = append(l.Edges, lex.Edge{From: i, To: i + 1, Token: lex.Token{TerminalSet: id}})
	}
	return l
}

// symbol is the id of the symbol named name.
func symbol(tab *grammar.Tables, name string) int {
	return slices.IndexFunc(tab.Symbols, func(s grammar.TableSymbol) bool { return s.Name == name })
}

// brute counts the derivations of sym over words[i:j] from the grammar's
// rules alone, no chart code: every tree whose leaves are those words, in
// which no (symbol, span) sits inside itself.
func brute(tab *grammar.Tables, words []string, sym, i, j int, path map[[3]int]bool) int {
	if tab.Symbols[sym].Terminal {
		if j == i+1 && words[i] == tab.Symbols[sym].Name {
			return 1
		}
		return 0
	}
	k := [3]int{sym, i, j}
	if path[k] {
		return 0
	}
	path[k] = true
	defer delete(path, k)
	n := 0
	for _, r := range tab.Rules {
		if r.LHS == sym {
			n += bruteSeq(tab, words, r.RHS, i, j, path)
		}
	}
	return n
}

// bruteSeq counts the ways rhs derives words[i:j].
func bruteSeq(tab *grammar.Tables, words []string, rhs []int, i, j int, path map[[3]int]bool) int {
	if len(rhs) == 0 {
		if i == j {
			return 1
		}
		return 0
	}
	n := 0
	for k := i; k <= j; k++ {
		if a := brute(tab, words, rhs[0], i, k, path); a > 0 {
			n += a * bruteSeq(tab, words, rhs[1:], k, j, path)
		}
	}
	return n
}

// checkTree reports how t is not a derivation of its grammar over l: a rule
// whose LHS is not the node's symbol, children that are not the rule's RHS
// or do not cover the span in order, a leaf no token or inserted ; admits,
// or a (symbol, span) inside itself.
func checkTree(tab *grammar.Tables, l *lex.Lattice[lex.None], inserted []int, t *Tree, path map[[3]int]bool) error {
	k := [3]int{t.Sym, t.From, t.To}
	if path[k] {
		return fmt.Errorf("%s over %d-%d sits inside itself", tab.Symbols[t.Sym].Name, t.From, t.To)
	}
	if tab.Symbols[t.Sym].Terminal {
		if t.Rule != -1 || t.Kids != nil {
			return fmt.Errorf("leaf %s has a rule or children", tab.Symbols[t.Sym].Name)
		}
		if t.From == t.To {
			if tab.Symbols[t.Sym].Name != "PERLY_SEMICOLON" || !slices.Contains(inserted, t.From) {
				return fmt.Errorf("zero-width leaf %s at %d, where no ; was inserted", tab.Symbols[t.Sym].Name, t.From)
			}
			return nil
		}
		if t.To != t.From+1 || !slices.Contains(tab.TerminalSets[l.Edges[t.From].Token.TerminalSet], t.Sym) {
			return fmt.Errorf("leaf %s over %d-%d matches no token", tab.Symbols[t.Sym].Name, t.From, t.To)
		}
		return nil
	}
	if t.Rule < 0 {
		return fmt.Errorf("%s over %d-%d has no derivation", tab.Symbols[t.Sym].Name, t.From, t.To)
	}
	r := tab.Rules[t.Rule]
	if r.LHS != t.Sym || len(r.RHS) != len(t.Kids) {
		return fmt.Errorf("%s over %d-%d uses rule %d with %d children", tab.Symbols[t.Sym].Name, t.From, t.To, t.Rule, len(t.Kids))
	}
	path[k] = true
	defer delete(path, k)
	at := t.From
	for i, kid := range t.Kids {
		if kid.Sym != r.RHS[i] || kid.From != at {
			return fmt.Errorf("rule %d's child %d is %s from %d, want %s from %d", t.Rule, i, tab.Symbols[kid.Sym].Name, kid.From, tab.Symbols[r.RHS[i]].Name, at)
		}
		if err := checkTree(tab, l, inserted, kid, path); err != nil {
			return err
		}
		at = kid.To
	}
	if at != t.To {
		return fmt.Errorf("rule %d's children end at %d, not %d", t.Rule, at, t.To)
	}
	return nil
}

// selfLoop is a node reachable in f with an Alt whose child is the node itself, or nil.
func selfLoop(f *Forest) *Node {
	seen := map[*Node]bool{}
	work := []*Node{f.Root}
	for len(work) > 0 {
		n := work[len(work)-1]
		work = work[:len(work)-1]
		if n == nil || seen[n] {
			continue
		}
		seen[n] = true
		for _, a := range n.Alts {
			if a.Left == n || a.Right == n {
				return n
			}
			work = append(work, a.Left, a.Right)
		}
	}
	return nil
}

// treeText is t as text, to tell two derivations apart.
func treeText(t *Tree) string {
	var b strings.Builder
	fmt.Fprintf(&b, "(%d %d %d-%d", t.Sym, t.Rule, t.From, t.To)
	for _, k := range t.Kids {
		b.WriteString(" " + treeText(k))
	}
	return b.String() + ")"
}

// leaves is t's leaves, in order.
func leaves(t *Tree) []*Tree {
	if t.Kids == nil && t.Rule == -1 {
		return []*Tree{t}
	}
	var out []*Tree
	for _, k := range t.Kids {
		out = append(out, leaves(k)...)
	}
	return out
}

// dump is everything reachable from f's root, one line per node with its
// alternatives, sorted, so two forests compare as text.
func dump(f *Forest) string {
	if f == nil {
		return "no forest"
	}
	key := func(n *Node) string {
		if n == nil {
			return "-"
		}
		return fmt.Sprintf("%d/%d.%d@%d-%d", n.Sym, n.Rule, n.Dot, n.From, n.To)
	}
	seen := map[*Node]bool{}
	var lines []string
	work := []*Node{f.Root}
	for len(work) > 0 {
		n := work[len(work)-1]
		work = work[:len(work)-1]
		if n == nil || seen[n] {
			continue
		}
		seen[n] = true
		var alts []string
		for _, a := range n.Alts {
			alts = append(alts, fmt.Sprintf("%d:%s,%s", a.Rule, key(a.Left), key(a.Right)))
			work = append(work, a.Left, a.Right)
		}
		sort.Strings(alts)
		lines = append(lines, key(n)+" "+strings.Join(alts, " "))
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

// optionSets are the recognizer's options a forest must not depend on.
var optionSets = []recognize.Options{{}, {NoLeo: true}, {NoLookahead: true}, {NoLeo: true, NoLookahead: true}, {NoCache: true}}

// toyForest recognizes words under opts and builds the forest; nil where the chart refuses.
func toyForest(t *testing.T, tab *grammar.Tables, words []string, opts recognize.Options) (*Forest, *recognize.Chart, *lex.Lattice[lex.None]) {
	t.Helper()
	l := toyLattice(tab, words)
	c, err := recognize.Recognize(recognize.New(tab, opts), l)
	if err != nil {
		t.Fatal(err)
	}
	if !c.Accepted() {
		return nil, c, l
	}
	return build(tab, c, l), c, l
}

// inputs is every string of up to n of words.
func inputs(words []string, n int) [][]string {
	out := [][]string{nil}
	for prev := out; n > 0; n-- {
		var next [][]string
		for _, p := range prev {
			for _, w := range words {
				next = append(next, append(slices.Clone(p), w))
			}
		}
		out, prev = append(out, next...), next
	}
	return out
}

// forestToys are left and right recursion, nullable chains, ambiguity, a
// cycle, and right recursion Leo memoizes beside ambiguity.
var forestToys = []struct {
	name  string
	rules []string
	words []string
	n     int
}{
	{"left recursion", []string{"S ::= S A | A"}, []string{"A", "B"}, 5},
	{"right recursion", []string{"S ::= A S | A"}, []string{"A", "B"}, 6},
	{"nullable chain", []string{"S ::= X Y Z", "X ::= Y", "Y ::= Z |", "Z ::= A |"}, []string{"A", "B"}, 4},
	{"ambiguity", []string{"S ::= S S | A"}, []string{"A", "B"}, 6},
	{"cycle", []string{"S ::= S | T | A", "T ::= S B | S"}, []string{"A", "B"}, 4},
	{"nullable right recursion", []string{"S ::= A S | B S |"}, []string{"A", "B"}, 5},
	{"Leo beside ambiguity", []string{"S ::= U U T | N V", "T ::= A", "U ::= V | T U", "V ::= V N | B | V S S", "N ::= C*"}, []string{"A", "B", "C"}, 5},
}

// TestForestFromChart: the forest built from a recognized chart agrees with
// a brute-force enumerator over the grammar's rules on toy grammars, under
// every option, and is the same forest with Leo on and off; on perly.y a
// slipper ; is a zero-width leaf, and over the corpus sample each file has
// a forest exactly where the chart accepts it, one of whose derivations has
// the file's tokens as its leaves, in order.
func TestForestFromChart(t *testing.T) {
	t.Run("toys", func(t *testing.T) {
		chains := 0
		for _, toy := range forestToys {
			tab := toyTables(t, toy.rules...)
			for _, words := range inputs(toy.words, toy.n) {
				want := brute(tab, words, tab.Start, 0, len(words), map[[3]int]bool{})
				var first string
				for i, opts := range optionSets {
					f, c, l := toyForest(t, tab, words, opts)
					if (f != nil) != (want > 0) {
						t.Fatalf("%s %q %+v: forest %v, %d derivations by brute force", toy.name, words, opts, f != nil, want)
					}
					if i == 0 {
						first = dump(f)
					} else if got := dump(f); got != first {
						t.Errorf("%s %q: the forest under %+v differs from the one under %+v:\n%s\nwant\n%s", toy.name, words, opts, optionSets[0], got, first)
					}
					if f == nil {
						continue
					}
					trees, cut := f.Derivations(f.Root, want+1)
					if cut {
						t.Errorf("%s %q %+v: enumeration cut short below its limit", toy.name, words, opts)
					}
					distinct := map[string]bool{}
					for _, tr := range trees {
						if err := checkTree(tab, l, c.Inserted(), tr, map[[3]int]bool{}); err != nil {
							t.Fatalf("%s %q %+v: %v", toy.name, words, opts, err)
						}
						distinct[treeText(tr)] = true
					}
					if len(trees) != want || len(distinct) != want {
						t.Errorf("%s %q %+v: %d derivations, %d distinct; brute force counts %d", toy.name, words, opts, len(trees), len(distinct), want)
					}
					if opts == (recognize.Options{}) {
						for pos := range c.Sets {
							for sym := range tab.Symbols {
								if _, _, _, ok := c.LeoLink(pos, sym); ok {
									chains++
								}
							}
						}
					}
				}
			}
		}
		if chains == 0 {
			t.Error("no toy input made a Leo memo, so the forest's Leo expansion went untested")
		}
	})

	t.Run("slipper leaf", func(t *testing.T) {
		tab := committedTables(t)
		semi := symbol(tab, "PERLY_SEMICOLON")
		for _, src := range []string{"{ 1 }", "use strict", "if ($x) { print 1 }\n"} {
			f, c, l := perlForest(t, tab, src, recognize.Options{})
			if f == nil {
				t.Fatalf("%q: no forest", src)
			}
			trees, _ := f.Derivations(f.Root, 1)
			if len(trees) != 1 {
				t.Fatalf("%q: %d derivations, want 1", src, len(trees))
			}
			if err := checkTree(tab, l, c.Inserted(), trees[0], map[[3]int]bool{}); err != nil {
				t.Fatalf("%q: %v", src, err)
			}
			if !slices.ContainsFunc(leaves(trees[0]), func(lf *Tree) bool { return lf.Sym == semi && lf.From == lf.To }) {
				t.Errorf("%q: the derivation has no zero-width ; leaf; inserted at %v", src, c.Inserted())
			}
		}
		// The inserted ; is no extra null statement beside a written one.
		if f, _ := Parse(tab, []byte("{ 1; }")); f == nil {
			t.Error("{ 1; }: no forest")
		} else if trees, cut := f.Derivations(f.Root, 10); len(trees) != 1 || cut {
			t.Errorf("{ 1; }: %d derivations, cut %v; want exactly one", len(trees), cut)
		}
	})

	// A cache hit at node j runs no completion, so the side table holds no
	// memo for a chain whose top j's set holds; the forest must not read it.
	t.Run("Leo off on cached sets", func(t *testing.T) {
		tab := committedTables(t)
		var src strings.Builder
		for _, name := range []string{"new", "make", "build", "clone"} {
			fmt.Fprintf(&src, "sub %s { my ($class, %%args) = @_; my $self = {%%args}; return bless($self, $class); }\n", name)
		}
		on, c, _ := perlForest(t, tab, src.String(), recognize.Options{})
		off, _, _ := perlForest(t, tab, src.String(), recognize.Options{NoLeo: true})
		if on == nil || dump(on) != dump(off) {
			t.Errorf("subs returning bless(...): the forest with Leo on (%v) differs from the one with it off", on != nil)
		}
		if !slices.Contains(c.Cached, true) {
			t.Error("no node was served by the transition cache, so the case is not exercised")
		}
	})

	t.Run("corpus sample", func(t *testing.T) {
		tab := committedTables(t)
		accepted := 0
		for _, p := range samplePaths(t) {
			src, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			f, c, l := perlForest(t, tab, string(src), recognize.Options{})
			if (f != nil) != c.Accepted() {
				t.Fatalf("%s: forest %v where the chart's verdict is %v", p, f != nil, c.Accepted())
			}
			if g, _ := Parse(tab, src); (g != nil) != c.Accepted() {
				t.Fatalf("%s: Parse gives forest %v where the chart's verdict is %v", p, g != nil, c.Accepted())
			}
			if f == nil {
				continue
			}
			accepted++
			checkYield(t, p, tab, f, c, l)
			if n := selfLoop(f); n != nil {
				t.Errorf("%s: %s over %d-%d has an Alt whose child is itself", p, tab.Symbols[n.Sym].Name, n.From, n.To)
			}
			if p == samplePaths(t)[0] {
				continue // std5k.pl: its Leo-off forest is checked by the toys' and smaller files' agreement
			}
			off, _, _ := perlForest(t, tab, string(src), recognize.Options{NoLeo: true})
			if dump(off) != dump(f) {
				t.Errorf("%s: the forest with Leo off differs from the one with it on", p)
			}
		}
		if accepted == 0 {
			t.Error("the chart accepted no file of the sample")
		}
	})
}

// checkYield fails t unless one derivation of f is a derivation of the
// grammar over l whose non-zero-width leaves are every token, in order.
func checkYield(t *testing.T, name string, tab *grammar.Tables, f *Forest, c *recognize.Chart, l *lex.Lattice[lex.None]) {
	t.Helper()
	trees, _ := f.Derivations(f.Root, 1)
	if len(trees) != 1 {
		t.Fatalf("%s: %d derivations, want 1", name, len(trees))
	}
	if err := checkTree(tab, l, c.Inserted(), trees[0], map[[3]int]bool{}); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	var at []int
	for _, lf := range leaves(trees[0]) {
		if lf.From != lf.To {
			at = append(at, lf.From)
		}
	}
	if len(at) != len(l.Edges) || slices.ContainsFunc(at, func(i int) bool { return at[i] != i }) {
		t.Errorf("%s: the derivation's leaves start at %v, want each of the %d tokens in order", name, at, len(l.Edges))
	}
}

// perlForest lexes src, recognizes it under opts, and builds the forest; nil where the chart refuses.
func perlForest(t *testing.T, tab *grammar.Tables, src string, opts recognize.Options) (*Forest, *recognize.Chart, *lex.Lattice[lex.None]) {
	t.Helper()
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	l := lex.Linear(s)
	c, err := recognize.Recognize(recognize.New(tab, opts), l)
	if err != nil {
		t.Fatal(err)
	}
	if !c.Accepted() {
		return nil, c, l
	}
	return build(tab, c, l), c, l
}

// TestForestRefusedChartHasNoForest: a file the chart refuses gets a
// refusal naming the byte of the token it is refused at, or the end of the
// file, and no forest, never a partial one.
func TestForestRefusedChartHasNoForest(t *testing.T) {
	tab := committedTables(t)
	for _, want := range []struct {
		src string
		Refusal
	}{
		{"my $x = ;\n", Refusal{8, "grammar"}},
		{"print 1;\n$h{1;};\n", Refusal{13, "grammar"}},
		{"if ($x) {\n", Refusal{10, "grammar"}},
		{"1 +", Refusal{3, "grammar"}},
		{"print <<EOT;\nno terminator\n", Refusal{6, "lex-unmapped"}},
	} {
		f, ref := Parse(tab, []byte(want.src))
		if f != nil {
			t.Errorf("%q: a forest of %v for a refused file", want.src, f.Root)
		}
		if ref != want.Refusal {
			t.Errorf("%q: refusal %+v, want %+v", want.src, ref, want.Refusal)
		}
	}
	f, ref := Parse(tab, []byte("my $x = 1;\n"))
	if f == nil || ref != (Refusal{}) {
		t.Errorf("my $x = 1;: forest %v, refusal %+v; want a forest and no refusal", f != nil, ref)
	}
}

// TestForestEnumerationBounded: enumerating a highly ambiguous, cyclic and
// nullable toy's derivations stops at the limit, however many there are;
// and on a diamond of unit-rule cycles, whose dead ends grow exponentially
// with its depth, it stops at its step budget and says it was cut short.
func TestForestEnumerationBounded(t *testing.T) {
	tab := toyTables(t, "S ::= S S | S | A |")
	f, _, _ := toyForest(t, tab, slices.Repeat([]string{"A"}, 40), recognize.Options{})
	if f == nil {
		t.Fatal("no forest")
	}
	for _, limit := range []int{1, 7, 500} {
		if got, cut := f.Derivations(f.Root, limit); len(got) != limit || !cut {
			t.Errorf("Derivations(%d) gave %d, cut %v", limit, len(got), cut)
		}
	}

	const k = 28
	rules := []string{"A1 ::= A2 | B2 | X"}
	for i := 2; i < k; i++ {
		rules = append(rules, fmt.Sprintf("A%d ::= A%d | B%d", i, i+1, i+1), fmt.Sprintf("B%d ::= A%d | B%d", i, i+1, i+1))
	}
	rules = append(rules, fmt.Sprintf("A%d ::= A1", k), fmt.Sprintf("B%d ::= A1", k))
	f, _, _ = toyForest(t, toyTables(t, rules...), []string{"X"}, recognize.Options{})
	if f == nil {
		t.Fatal("diamond: no forest")
	}
	start := time.Now()
	got, cut := f.Derivations(f.Root, 10)
	if d := time.Since(start); d > 2*time.Second || len(got) > 1 || len(got) == 0 && !cut {
		t.Errorf("diamond of depth %d: %d derivations, cut %v, in %v; want at most its one, or a cut, within 2s", k, len(got), cut, d)
	}
}

// TestForestLinearLatticesOnly: build refuses a lattice that branches, which
// it would misread (shortcut in build), and builds the linear one beside it.
func TestForestLinearLatticesOnly(t *testing.T) {
	tab := toyTables(t, "S ::= X C", "X ::= A | B B")
	set := func(name string) lex.Token { return toyLattice(tab, []string{name}).Edges[0].Token }
	branched := &lex.Lattice[lex.None]{Nodes: make([]lex.Node[lex.None], 4), Edges: []lex.Edge{
		{From: 0, To: 2, Token: set("A")}, {From: 0, To: 1, Token: set("B")}, {From: 1, To: 2, Token: set("B")}, {From: 2, To: 3, Token: set("C")},
	}}
	c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), branched)
	if err != nil || !c.Accepted() {
		t.Fatalf("the branched lattice: %v, accepted %v", err, err == nil && c.Accepted())
	}
	if f := build(tab, c, branched); f != nil {
		t.Error("build made a forest of a branched lattice")
	}
	if f, _, _ := toyForest(t, tab, []string{"B", "B", "C"}, recognize.Options{}); f == nil {
		t.Error("build refused the linear lattice B B C")
	}
}

// TestForestRightRecursionLinear: the forest's work (facts recorded and
// splits tried) and size grow linearly in a right recursion Leo keeps linear
// for the recognizer, if-elsif chains and runs of !, and in the left
// recursion of a statement sequence.
func TestForestRightRecursionLinear(t *testing.T) {
	tab := committedTables(t)
	for name, src := range map[string]func(n int) string{
		"elsif":      func(n int) string { return "if ($x) { 1 }" + strings.Repeat(" elsif ($x) { 1 }", n) + "\n" },
		"!":          func(n int) string { return strings.Repeat("!", n) + "$x;\n" },
		"statements": func(n int) string { return strings.Repeat("$x = 1;\n", n) }, // left recursion
	} {
		var steps, size [2]int
		for i, n := range []int{400, 800} {
			f, _, _ := perlForest(t, tab, src(n), recognize.Options{})
			if f == nil {
				t.Fatalf("%s × %d: no forest", name, n)
			}
			steps[i], size[i] = f.steps, f.size
		}
		if steps[1] > 5*steps[0]/2 || size[1] > 5*size[0]/2 {
			t.Errorf("%s: doubling the chain took the forest's work from %d to %d and its size from %d to %d, want at most 2.5×", name, steps[0], steps[1], size[0], size[1])
		}
	}
}

// TestForestDegenerateFiles: an empty file, one of comments only and one
// that is only ; each have a forest, a derivation of which has their tokens
// as its leaves, a ; the slipper inserted counted where one is used.
func TestForestDegenerateFiles(t *testing.T) {
	tab := committedTables(t)
	for _, src := range []string{"", "# only a comment\n# and another\n", ";", ";;\n", "{ }"} {
		f, c, l := perlForest(t, tab, src, recognize.Options{})
		if f == nil {
			t.Fatalf("%q: no forest", src)
		}
		if g, ref := Parse(tab, []byte(src)); g == nil {
			t.Fatalf("%q: Parse refuses it: %+v", src, ref)
		}
		checkYield(t, fmt.Sprintf("%q", src), tab, f, c, l)
	}
}

// committedTables is the committed grammar's tables, generated once for the package's tests, none of which writes them.
func committedTables(t testing.TB) *grammar.Tables {
	t.Helper()
	committed.Do(func() {
		g, ref, err := grammar.Committed()
		if err == nil {
			committed.tab, err = grammar.Generate(g, ref)
		}
		committed.err = err
	})
	if committed.err != nil {
		t.Fatal(committed.err)
	}
	return committed.tab
}

var committed struct {
	sync.Once
	tab *grammar.Tables
	err error
}

// samplePaths is the corpus sample: std5k.pl, every port snippet, and every
// 24th Guacamole case in name order.
func samplePaths(t *testing.T) []string {
	t.Helper()
	chart := filepath.Join("..", "testdata")
	paths := []string{filepath.Join(chart, "std5k.pl")}
	port, _ := filepath.Glob(filepath.Join(chart, "port", "*", "*.pl"))
	cases, _ := filepath.Glob(filepath.Join(chart, "guacamole", "parses", "*.pl"))
	paths = append(paths, port...)
	for i := 0; i < len(cases); i += 24 {
		paths = append(paths, cases[i])
	}
	if len(port) == 0 || len(cases) == 0 {
		t.Fatalf("the sample found %d port snippets and %d Guacamole cases", len(port), len(cases))
	}
	return paths
}
