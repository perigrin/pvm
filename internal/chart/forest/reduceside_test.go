// ABOUTME: Tests the judged recognizer's reduce side through the forest: its constructors, Build's refusal records,
// ABOUTME: and the reduce-side differential (judged recognizer, Build, Constrain) against V1 over toys and seeds.

package forest

import (
	"maps"
	"os"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/prectab"
	"tamarou.com/pvm/internal/chart/recognize"
)

// reduceSideCandidate swaps the differential's candidates for the reduce side's for the rest of t.
func reduceSideCandidate(t *testing.T) {
	t.Helper()
	src0, toy0 := CandidateForest, CandidateToyForest
	t.Cleanup(func() { CandidateForest, CandidateToyForest = src0, toy0 })
	CandidateForest, CandidateToyForest = ReduceSideForest, ReduceSideToyForest
}

// feasibleToys is the toy fixtures prectab.Feasible accepts, by name, which the judged recognizer judges, and the
// names of those it refuses, which the judged recognizer refuses (TestJudgedRecognizerRefusesInfeasibleGrammar) and
// Constrain alone judges over the no-precedence recognizer.
func feasibleToys(t *testing.T) (feasible map[string]toy, refused []string) {
	t.Helper()
	feasible = map[string]toy{}
	for _, name := range toyNames {
		ty := loadToy(t, name)
		if _, err := prectab.Feasible(ty.prec); err != nil {
			refused = append(refused, name)
			continue
		}
		feasible[name] = ty
	}
	return feasible, refused
}

// chartSets is each node's items in c, the recognizer's set content, comparable across recognizers.
func chartSets(c *recognize.Chart) [][]recognize.Item {
	out := make([][]recognize.Item, len(c.Sets))
	for i, s := range c.Sets {
		out[i] = s.Items()
	}
	return out
}

// sampleLattices is the perly.y sample, lexed, as linear lattices.
func sampleLattices(t *testing.T, tab *grammar.Tables) []*lex.Lattice[lex.None] {
	t.Helper()
	var out []*lex.Lattice[lex.None]
	for _, p := range samplePaths(t) {
		src, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		s, _ := lex.Lex(src, tab.Tokens)
		out = append(out, lex.Linear(s))
	}
	return out
}

// TestForestEqualityReduceSide: the reduce side's forest, the judged recognizer then Build then Constrain, equals
// V1's on every sentence of every toy Feasible accepts, under the toy's own precedence, and on every seed-4 and
// seed-44 sentence. Each logs how many it compared of how many it loaded. The candidates take their recognizer
// from reduceSideRecognizer, which judges, so neither is V1 under another name.
func TestForestEqualityReduceSide(t *testing.T) {
	t.Run("candidate-is-judged", func(t *testing.T) {
		calls := 0
		rec0 := reduceSideRecognizer
		t.Cleanup(func() { reduceSideRecognizer = rec0 })
		reduceSideRecognizer = func(tab *grammar.Tables, p *prectab.Precedence) (*recognize.Recognizer, error) {
			calls++
			return rec0(tab, p)
		}
		tab := committedTables(t)
		ReduceSideForest(tab, []byte("1 + 2 * 3;"), Scope{})
		ty := loadToy(t, "indirob")
		ReduceSideToyForest(ty.tab, toyLattice(ty.tab, strings.Fields("VAR")), ty.prec)
		if calls != 2 {
			t.Errorf("the reduce-side candidates took %d recognizers from reduceSideRecognizer; want one each, 2", calls)
		}
		s, _ := lex.Lex([]byte("1 + 2 * 3;"), tab.Tokens)
		l := lex.Linear(s)
		r, err := rec0(tab, nil)
		if err != nil {
			t.Fatal(err)
		}
		judged, _ := recognize.Recognize(r, l)
		plain, _ := recognize.Recognize(recognize.NewNoPrecedence(tab, recognize.Options{}), l)
		if reflect.DeepEqual(chartSets(judged), chartSets(plain)) {
			t.Error("reduceSideRecognizer's sets on 1 + 2 * 3; equal the no-precedence recognizer's: it judges nothing")
		}
	})
	reduceSideCandidate(t)
	toys, refused := feasibleToys(t)
	t.Logf("toys Feasible refuses, judged by Constrain alone: %v", refused)
	if !slices.Equal(refused, []string{"empty", "emptyred", "ladder", "nonassoc", "prefix"}) {
		t.Errorf("toys Feasible refuses: %v; want [empty emptyred ladder nonassoc prefix]", refused)
	}
	loaded, compared := 0, 0
	for _, name := range slices.Sorted(maps.Keys(toys)) {
		n, accepted, diffs := toyDiffs(t, toys[name])
		loaded += len(toys[name].sentences)
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
	tab := committedTables(t)
	for _, seed := range []struct {
		name string
		srcs []string
	}{{"4", operatorSentences()}, {"44", chainSentences()}} {
		compared, accepted, diffs := sentenceDiffs(tab, seed.srcs)
		t.Logf("seed %s: compared %d of %d sentences, V1 accepts %d", seed.name, compared, len(seed.srcs), accepted)
		if compared == 0 || compared != len(seed.srcs) || accepted == 0 || accepted == compared {
			t.Errorf("seed %s: compared %d of %d, V1 accepts %d; want every one, and some of each verdict", seed.name, compared, len(seed.srcs), accepted)
		}
		reportDiffs(t, "seed "+seed.name, compared, diffs)
	}
}

// TestJudgedRecognizerIsExplicit: on the perly.y sample the judged constructor's sets differ from the no-precedence
// constructor's in at least one set, and the differential's V1 side takes its recognizer from the no-precedence
// constructor, never recognize.New, which the switch-over makes the judged one.
func TestJudgedRecognizerIsExplicit(t *testing.T) {
	tab := committedTables(t)
	judged, err := recognize.NewJudged(tab, nil, recognize.Options{})
	if err != nil {
		t.Fatal(err)
	}
	plain := recognize.NewNoPrecedence(tab, recognize.Options{})
	differ := 0
	for _, l := range sampleLattices(t, tab) {
		a, errA := recognize.Recognize(judged, l)
		b, errB := recognize.Recognize(plain, l)
		if errA != nil || errB != nil {
			t.Fatalf("recognize: %v, %v", errA, errB)
		}
		x, y := chartSets(a), chartSets(b)
		for i := range x {
			if !slices.Equal(x[i], y[i]) {
				differ++
			}
		}
	}
	t.Logf("the judged constructor's sets differ from the no-precedence constructor's at %d nodes of the sample", differ)
	if differ == 0 {
		t.Error("the judged and no-precedence constructors' sets are equal over the sample")
	}
	name := runtime.FuncForPC(reflect.ValueOf(v1NewRecognizer).Pointer()).Name()
	t.Logf("the V1 side's recognizer comes from %s", name)
	if !strings.HasSuffix(name, "/recognize.NewNoPrecedence") {
		t.Errorf("the V1 side's recognizer comes from %s; want recognize.NewNoPrecedence", name)
	}
}

// infeasibleToys are the toys prectab.Feasible refuses that the issue names, one per failing counter shape.
var infeasibleToys = []struct {
	name string
	load func(*testing.T) (*grammar.Tables, *prectab.Precedence)
}{
	{"emptyred", toyOf("emptyred")},
	{"nullableShift", nullableShiftToy},
	{"MarkFirstFilter", markFirstFilterToy},
	{"empty", toyOf("empty")},
	{"scancollide", scancollideToy},
}

// toyInputs is every string of up to five of tab's terminals, as lattices.
func toyInputs(tab *grammar.Tables) []*lex.Lattice[lex.None] {
	var terms []string
	for _, s := range tab.Symbols {
		if s.Terminal && s.Name != "EOF" {
			terms = append(terms, s.Name)
		}
	}
	words := [][]string{nil}
	for n := 0; n < len(words); n++ {
		if len(words[n]) < 5 {
			for _, w := range terms {
				words = append(words, append(slices.Clone(words[n]), w))
			}
		}
	}
	var out []*lex.Lattice[lex.None]
	for _, w := range words {
		out = append(out, toyLattice(tab, w))
	}
	return out
}

// TestJudgedRecognizerRefusesInfeasibleGrammar: on each toy Feasible refuses, the judged constructor given the
// toy's own precedence returns Feasible's refusal, as text, never the not-the-committed-grammar error, and Recognize
// returns it for every input; the no-precedence constructor still recognizes the toy, accepting some input.
func TestJudgedRecognizerRefusesInfeasibleGrammar(t *testing.T) {
	for _, ty := range infeasibleToys {
		t.Run(ty.name, func(t *testing.T) {
			tab, p := ty.load(t)
			_, want := prectab.Feasible(p)
			if want == nil {
				t.Fatalf("%s: Feasible accepts it", ty.name)
			}
			r, err := recognize.NewJudged(tab, p, recognize.Options{})
			if err == nil || err.Error() != want.Error() {
				t.Fatalf("%s: NewJudged's error %v; want Feasible's, %v", ty.name, err, want)
			}
			t.Logf("%s: %v", ty.name, err)
			accepted := 0
			for _, l := range toyInputs(tab) {
				if _, err := recognize.Recognize(r, l); err == nil || err.Error() != want.Error() {
					t.Fatalf("%s: Recognize's error %v; want Feasible's", ty.name, err)
				}
				c, err := recognize.Recognize(recognize.NewNoPrecedence(tab, recognize.Options{}), l)
				if err != nil {
					t.Fatalf("%s: the no-precedence recognizer: %v", ty.name, err)
				}
				if c.Accepted() {
					accepted++
				}
			}
			if accepted == 0 {
				t.Errorf("%s: the no-precedence recognizer accepts no input of up to five terminals", ty.name)
			}
		})
	}
}

// pairToy is a toy where X: b, completed before t, is allowed for one parent and refused for another in one set:
// after a, P: a . X t (whose state after b shifts nothing) and F: Y . X t, through Y: a (whose state after b also
// holds Z: b . t, which wins). F: Y X . t stands at that node all the same, through Y empty and X: a b, so Build
// meets the refused pair while splitting it.
func pairToy(t *testing.T) (*grammar.Tables, *prectab.Precedence) {
	return pinnedToy(t, []grammar.Level{{N: 1, Assoc: "left", Tokens: []string{"b"}}, {N: 2, Assoc: "left", Tokens: []string{"t"}}},
		make([]string, 10), "S ::= P | F", "P ::= a X t", "F ::= Y X t | Y Z", "Y ::= a |", "X ::= b | a b", "Z ::= b t")
}

// TestBuildNeverPairsRefusedParent: on pairToy's a b t, the judged recognizer records X: b from node 1 refused for
// F's parent state and not for P's, and Build, splitting F: Y X . t over 0 to 2, pairs X over 1 to 2 with P's item
// and never with F's, while it keeps F's pair with X over 0 to 2. The no-precedence chart, which refuses nothing,
// gives Build that pair, so the fixture reaches it.
func TestBuildNeverPairsRefusedParent(t *testing.T) {
	tab, p := pairToy(t)
	l := toyLattice(tab, strings.Fields("a b t"))
	r, err := recognize.NewJudged(tab, p, recognize.Options{})
	if err != nil {
		t.Fatal(err)
	}
	c, err := recognize.Recognize(r, l)
	if err != nil || !c.Accepted() {
		t.Fatalf("judged: %v, accepted %v", err, c != nil && c.Accepted())
	}
	X, F, P := symbol(tab, "X"), symbol(tab, "F"), symbol(tab, "P")
	parents := map[int]bool{} // each parent state at node 1 awaiting X, and whether the judgement refused it
	for _, it := range c.Sets[1].Items() {
		if slices.ContainsFunc(tab.States[it.State].Items, func(i int) bool {
			ci := tab.Items[i]
			rhs := tab.Rules[ci.Rule].RHS
			return ci.Dot < len(rhs) && rhs[ci.Dot] == X
		}) {
			parents[it.State] = c.Refused(2, X, 1, it.State)
		}
	}
	refusedN := 0
	for _, refused := range parents {
		if refused {
			refusedN++
		}
	}
	if len(parents) != 2 || refusedN != 1 {
		t.Fatalf("parent states at node 1 awaiting X and whether X from 1 was refused for each: %v; want two, one refused", parents)
	}
	pairs := func(f *Forest) (withP, withF, fWide bool) {
		for n, nd := range f.Nodes[1:] {
			for a := nd.Alt; a != 0; a = f.Alts[a].Next {
				alt := f.Alts[a]
				if alt.Right == 0 {
					continue
				}
				rt := f.Nodes[alt.Right]
				lhs := tab.Rules[alt.Rule].LHS
				if int(rt.Sym) != X || rt.To != 2 {
					continue
				}
				switch {
				case lhs == P && rt.From == 1:
					withP = true
				case lhs == F && rt.From == 1:
					withF = true
					t.Logf("node %d pairs X over 1 to 2 with F's item", n+1)
				case lhs == F && rt.From == 0:
					fWide = true
				}
			}
		}
		return withP, withF, fWide
	}
	if withP, withF, fWide := pairs(Build(tab, c, l)); !withP || withF || !fWide {
		t.Errorf("judged chart's forest: X over 1 to 2 paired with P %v, with F %v; F with X over 0 to 2 %v; want true, false, true", withP, withF, fWide)
	}
	plain, _ := recognize.Recognize(recognize.NewNoPrecedence(tab, recognize.Options{}), l)
	if _, withF, _ := pairs(Build(tab, plain, l)); !withF {
		t.Error("the no-precedence chart's forest never pairs X over 1 to 2 with F, so the fixture tests nothing")
	}
}

// TestReduceSideNoLookaheadSameAcceptance: with the reduce side on, NoLookahead changes no acceptance over the
// sentences of every toy Feasible accepts, under its own precedence, and over the seed-4 and seed-44 sentences.
func TestReduceSideNoLookaheadSameAcceptance(t *testing.T) {
	check := func(what string, tab *grammar.Tables, p *prectab.Precedence, lattices []*lex.Lattice[lex.None]) {
		t.Helper()
		on, errOn := recognize.NewJudged(tab, p, recognize.Options{})
		off, errOff := recognize.NewJudged(tab, p, recognize.Options{NoLookahead: true})
		if errOn != nil || errOff != nil {
			t.Fatalf("%s: %v, %v", what, errOn, errOff)
		}
		accepted, differ := 0, 0
		for _, l := range lattices {
			a, errA := recognize.Recognize(on, l)
			b, errB := recognize.Recognize(off, l)
			if errA != nil || errB != nil {
				t.Fatalf("%s: %v, %v", what, errA, errB)
			}
			if a.Accepted() {
				accepted++
			}
			if a.Accepted() != b.Accepted() {
				differ++
			}
		}
		t.Logf("%s: %d inputs, %d accepted, %d whose acceptance NoLookahead changes", what, len(lattices), accepted, differ)
		if accepted == 0 || accepted == len(lattices) || differ != 0 {
			t.Errorf("%s: %d of %d accepted, %d changed by NoLookahead; want some of each verdict, none changed", what, accepted, len(lattices), differ)
		}
	}
	toys, _ := feasibleToys(t)
	for _, name := range slices.Sorted(maps.Keys(toys)) {
		ty := toys[name]
		var ls []*lex.Lattice[lex.None]
		for _, s := range ty.sentences {
			ls = append(ls, toyLattice(ty.tab, s.words))
		}
		check(name, ty.tab, ty.prec, ls)
	}
	tab := committedTables(t)
	for _, seed := range []struct {
		name string
		srcs []string
	}{{"seed 4", operatorSentences()}, {"seed 44", chainSentences()}} {
		var ls []*lex.Lattice[lex.None]
		for _, src := range seed.srcs {
			s, _ := lex.Lex([]byte(src), tab.Tokens)
			ls = append(ls, lex.Linear(s))
		}
		check(seed.name, tab, nil, ls)
	}
}
