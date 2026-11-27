// ABOUTME: Tests the judged recognizer's forest with no Constrain: Build walking by context, and the differential
// ABOUTME: (judged recognizer then Build) against V1 over the toys Feasible passes and seeds 4 and 44.

package forest

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/prectab"
	"tamarou.com/pvm/internal/chart/recognize"
)

// deadContextToy is S ::= X END1 | Q X END2, X ::= X T X | C T A | X U | C U | P X | A, with T and U below P, which
// binds tighter than both, the infix T and the postfix U. After a token that is both P and Q, X is predicted under
// P's real top, which forbids the shift of T after X . T X and of U after X . U, and under none, Q's, which forbids
// nothing; C T A and C U shift their T and U under either.
func deadContextToy(t *testing.T) (*grammar.Tables, *prectab.Precedence) {
	return pinnedToy(t, []grammar.Level{{N: 1, Assoc: "left", Tokens: []string{"T", "U"}}, {N: 2, Assoc: "right", Tokens: []string{"P"}}},
		make([]string, 8), "S ::= X END1 | Q X END2", "X ::= X T X | C T A | X U | C U | P X | A")
}

// deadContextLattice is words with {P Q} first and {A C} second: the two sets are appended to tab's.
func deadContextLattice(tab *grammar.Tables, words string) *lex.Lattice[lex.None] {
	l := toyLattice(tab, strings.Fields(words))
	for i, set := range map[int][]string{0: {"P", "Q"}, 1: {"A", "C"}} {
		ids := []int{symbol(tab, set[0]), symbol(tab, set[1])}
		slices.Sort(ids)
		l.Edges[i].Token.TerminalSet = len(tab.TerminalSets)
		tab.TerminalSets = append(tab.TerminalSets, ids)
	}
	return l
}

// TestBuildDropsAlternativeOfDeadContext: on deadContextToy's {P Q} {A C} T A END1, X over 1 to 4 completes under
// Q's context, none, as X T X and as C T A, and under P's real top as C T A only, the shift of T after X . T X
// refused. Q's parent chain dies at END1, so the node is reached only under P's top, and V1 keeps C T A there alone.
// Build walks by context: it attaches C T A and never X T X, the alternative completed under the context whose
// parent chain died, and the forest equals V1's. On {P Q} {A C} U END1 the same holds of X over 1 to 3, X U and C U,
// where X . U stands under P's top too and only the shift of U is refused.
func TestBuildDropsAlternativeOfDeadContext(t *testing.T) {
	for _, tc := range []struct {
		words      string
		to         int
		dead, live []string // the rules' right-hand sides
	}{
		{"P A T A END1", 4, []string{"X", "T", "X"}, []string{"C", "T", "A"}},
		{"P A U END1", 3, []string{"X", "U"}, []string{"C", "U"}},
	} {
		t.Run(strings.Join(tc.dead, "-"), func(t *testing.T) {
			tab, p := deadContextToy(t)
			if _, err := prectab.Feasible(p); err != nil {
				t.Fatalf("the toy is not feasible: %v", err)
			}
			l := deadContextLattice(tab, tc.words)
			r, err := recognize.NewJudged(tab, p, recognize.Options{})
			if err != nil {
				t.Fatal(err)
			}
			c, err := recognize.Recognize(r, l)
			if err != nil || !c.Accepted() {
				t.Fatalf("the judged chart refuses %s (%v)", tc.words, err)
			}
			rule := func(rhs []string) int {
				return slices.IndexFunc(tab.Rules, func(r grammar.BNFRule) bool {
					return slices.EqualFunc(r.RHS, rhs, func(x int, n string) bool { return tab.Symbols[x].Name == n })
				})
			}
			dead, live := rule(tc.dead), rule(tc.live)
			// The chart completes X over 1 to to by the dead rule under one context only, and by the live one under two.
			ctxs := func(rule int) (out []int) {
				for _, it := range c.Sets[tc.to].Items() {
					if it.Dist == tc.to-1 && slices.Contains(tab.States[it.State].Reduces, rule) {
						out = append(out, it.Ctx)
					}
				}
				return out
			}
			t.Logf("X over 1 to %d: %v under contexts %v, %v under %v", tc.to, tc.dead, ctxs(dead), tc.live, ctxs(live))
			if len(ctxs(dead)) != 1 || len(ctxs(live)) != 2 || !slices.Contains(ctxs(live), ctxs(dead)[0]) {
				t.Fatalf("X over 1 to %d: %v under contexts %v, %v under %v; want the first under one, the second under it and another",
					tc.to, tc.dead, ctxs(dead), tc.live, ctxs(live))
			}
			v1, err := V1ToyForest(tab, l, p)
			if err != nil || v1 == nil {
				t.Fatalf("V1 refuses %s (%v)", tc.words, err)
			}
			f := Build(tab, c, l)
			x := NodeKey{Sym: int32(symbol(tab, "X")), Rule: -1, Dot: 0, From: 1, To: int32(tc.to)}
			rules := func(f *Forest) (out []int32) {
				alts, _ := reachableAlts(f)
				for _, a := range alts[x] {
					out = append(out, a.Rule)
				}
				return out
			}
			if got, want := rules(v1), []int32{int32(live)}; !slices.Equal(got, want) {
				t.Fatalf("V1 keeps rules %v on X over 1 to %d; want %v's %v alone", got, tc.to, tc.live, want)
			}
			if got := rules(f); !slices.Equal(got, []int32{int32(live)}) {
				t.Errorf("Build attaches rules %v to X over 1 to %d; want %v's %d alone, never %v's %d", got, tc.to, tc.live, live, tc.dead, dead)
			}
			if d := ForestDiff(f, v1); d != "" {
				t.Errorf("the judged forest differs from V1's: %s", d)
			}
		})
	}
}

// TestBuildReadsNoRefusedRule: in `_ - $h{1} - $fh;` the term over `_ - $h{1}` before the second - completes as
// termbinop and as UNIOPSUB term, `_(- $h{1})`, whose reduction V1 forbids before ADDOP, the one terminal of that
// token that can follow a term. The judged recognizer refuses that completion into every parent and records it, its
// item still standing in the set beside termbinop's, which completes the same term from the same node; Build reads
// no completion of the refused rule, so the term keeps termbinop alone, as V1 does; and so on a toy chart where no
// item carries a context.
func TestBuildReadsNoRefusedRule(t *testing.T) {
	tab := committedTables(t)
	const src = "_ - $h{1} - $fh;"
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	l := lex.Linear(s)
	r, err := recognize.NewJudged(tab, nil, recognize.Options{})
	if err != nil {
		t.Fatal(err)
	}
	c, err := recognize.Recognize(r, l)
	if err != nil || !c.Accepted() {
		t.Fatalf("%q: the judged chart refuses it (%v)", src, err)
	}
	uniop := slices.IndexFunc(tab.Rules, func(r grammar.BNFRule) bool {
		return len(r.RHS) == 2 && tab.Symbols[r.RHS[0]].Name == "UNIOPSUB" && tab.Symbols[r.LHS].Name == "term"
	})
	term := symbol(tab, "term")
	if !slices.ContainsFunc(c.Sets[6].Items(), func(it recognize.Item) bool {
		return it.Dist == 6 && slices.Contains(tab.States[it.State].Reduces, uniop)
	}) || !c.RuleRefused(6, uniop, term, 0) {
		t.Fatalf("node 6 %v: want an item completing UNIOPSUB term from node 0, its completion refused into every parent (%v)",
			c.Sets[6].Items(), c.RuleRefused(6, uniop, term, 0))
	}
	key := NodeKey{Sym: int32(term), Rule: -1, Dot: 0, From: 0, To: 6}
	alts, _ := reachableAlts(Build(tab, c, l))
	if len(alts[key]) == 0 || slices.ContainsFunc(alts[key], func(a altKey) bool { return int(a.Rule) == uniop }) {
		t.Errorf("Build keeps %+v on the term over 0 to 6; want some, and never UNIOPSUB term's rule %d", alts[key], uniop)
	}
	v1, _ := V1Forest(tab, []byte(src), Scope{})
	if d := ForestDiff(Build(tab, c, l), v1); d != "" {
		t.Errorf("%q: the judged forest differs from V1's: %s", src, d)
	}

	// The same where no item carries a context: on S ::= E, E ::= E PLUS NUM | E PLUS NUM TIMES NUM | E TIMES NUM |
	// M PLUS NUM %prec HIGH | NUM, PLUS below TIMES below HIGH, and {NUM M} PLUS NUM TIMES NUM, E over 0 to 3
	// completes as E PLUS NUM, refused before TIMES, and as M PLUS NUM, kept, and V1 keeps the second alone.
	ttab, tp := pinnedToy(t, []grammar.Level{{N: 1, Assoc: "left", Tokens: []string{"PLUS"}}, {N: 2, Assoc: "left", Tokens: []string{"TIMES"}},
		{N: 3, Assoc: "left", Tokens: []string{"HIGH"}}}, []string{"", "PLUS", "TIMES", "TIMES", "HIGH", ""},
		"S ::= E", "E ::= E PLUS NUM | E PLUS NUM TIMES NUM | E TIMES NUM | M PLUS NUM | NUM")
	tl := toyLattice(ttab, strings.Fields("NUM PLUS NUM TIMES NUM"))
	tl.Edges[0].Token.TerminalSet = len(ttab.TerminalSets)
	ttab.TerminalSets = append(ttab.TerminalSets, slices.Sorted(slices.Values([]int{symbol(ttab, "NUM"), symbol(ttab, "M")})))
	tr, err := recognize.NewJudged(ttab, tp, recognize.Options{})
	if err != nil {
		t.Fatal(err)
	}
	tc, err := recognize.Recognize(tr, tl)
	if err != nil || !tc.Accepted() {
		t.Fatalf("the toy's judged chart refuses {NUM M} PLUS NUM TIMES NUM (%v)", err)
	}
	for i, s := range tc.Sets {
		if slices.ContainsFunc(s.Items(), func(it recognize.Item) bool { return it.Ctx != 0 }) {
			t.Fatalf("the toy's node %d %v carries a context", i, s.Items())
		}
	}
	e := symbol(ttab, "E")
	if !tc.RuleRefused(3, 1, e, 0) {
		t.Fatalf("the toy's E PLUS NUM from node 0 is not refused into every parent at node 3")
	}
	tv1, err := V1ToyForest(ttab, tl, tp)
	if err != nil || tv1 == nil {
		t.Fatalf("V1 refuses the toy's {NUM M} PLUS NUM TIMES NUM (%v)", err)
	}
	ek := NodeKey{Sym: int32(e), Rule: -1, Dot: 0, From: 0, To: 3}
	talts, _ := reachableAlts(Build(ttab, tc, tl))
	if len(talts[ek]) != 1 || talts[ek][0].Rule != 4 {
		t.Errorf("Build keeps %+v on the toy's E over 0 to 3; want M PLUS NUM's rule 4 alone", talts[ek])
	}
	if d := ForestDiff(Build(ttab, tc, tl), tv1); d != "" {
		t.Errorf("the toy's judged forest differs from V1's: %s", d)
	}
}

// judgedCandidate swaps the differential's candidates for the judged recognizer's, with no Constrain, for the rest
// of t.
func judgedCandidate(t *testing.T) {
	t.Helper()
	src0, toy0 := CandidateForest, CandidateToyForest
	t.Cleanup(func() { CandidateForest, CandidateToyForest = src0, toy0 })
	CandidateForest, CandidateToyForest = JudgedForest, JudgedToyForest
}

// TestForestEqualityJudgedRecognizer: the judged recognizer's forest with no Constrain, the judged recognizer then
// Build, equals V1's on every sentence of every toy Feasible accepts, under the toy's own precedence, and on every
// seed-4 and seed-44 sentence, each logging how many it compared of how many it loaded. The toys Feasible refuses
// are excepted: the judged constructor refuses each with Feasible's error, so they are judged by Constrain over the
// no-precedence recognizer, V1's own path. Each candidate recognizes with the recognizer it takes from
// reduceSideRecognizer, perly.y's precedence derived for a source and the toy's own for a toy.
func TestForestEqualityJudgedRecognizer(t *testing.T) {
	t.Run("candidate-is-judged", func(t *testing.T) {
		er := loadToy(t, "emptyred")
		marker, feasErr := recognize.NewJudged(er.tab, er.prec, recognize.Options{})
		if feasErr == nil {
			t.Fatal("emptyred's judged recognizer has no error to mark it")
		}
		var got []*prectab.Precedence
		rec0 := reduceSideRecognizer
		t.Cleanup(func() { reduceSideRecognizer = rec0 })
		reduceSideRecognizer = func(_ *grammar.Tables, p *prectab.Precedence) (*recognize.Recognizer, error) {
			got = append(got, p)
			return marker, nil
		}
		if f, ref := JudgedForest(committedTables(t), []byte("1 + 2 * 3;"), Scope{}); f != nil || ref.Reason != feasErr.Error() {
			t.Errorf("JudgedForest with the marker: forest %v, refusal %q; want none, the marker's %q: it recognizes with another recognizer", f != nil, ref.Reason, feasErr)
		}
		ty := loadToy(t, "indirob")
		if f, err := JudgedToyForest(ty.tab, toyLattice(ty.tab, strings.Fields("VAR")), ty.prec); f != nil || err != feasErr {
			t.Errorf("JudgedToyForest with the marker: forest %v, error %v; want none, the marker's: it recognizes with another recognizer", f != nil, err)
		}
		if len(got) != 2 || got[0] != nil || got[1] != ty.prec {
			t.Errorf("the judged candidates handed reduceSideRecognizer %v; want nil, then the toy's precedence %p", got, ty.prec)
		}
	})
	judgedCandidate(t)
	toys, refused := feasibleToys(t)
	t.Logf("toys Feasible refuses, judged by Constrain over the no-precedence recognizer: %v", refused)
	if !slices.Equal(refused, []string{"empty", "emptyred", "ladder", "nonassoc", "prefix"}) {
		t.Errorf("toys Feasible refuses: %v; want [empty emptyred ladder nonassoc prefix]", refused)
	}
	for _, name := range refused {
		ty := loadToy(t, name)
		_, feasErr := prectab.Feasible(ty.prec)
		if _, err := recognize.NewJudged(ty.tab, ty.prec, recognize.Options{}); err == nil || err.Error() != feasErr.Error() {
			t.Errorf("%s: the judged constructor's error %v; want Feasible's %v", name, err, feasErr)
		}
		accepted := 0
		for _, s := range ty.sentences {
			if f, err := V1ToyForest(ty.tab, toyLattice(ty.tab, s.words), ty.prec); err == nil && f != nil {
				accepted++
			}
		}
		if accepted == 0 {
			t.Errorf("%s: Constrain over the no-precedence recognizer accepts none of %d sentences", name, len(ty.sentences))
		}
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

// TestForestEqualityJudgedCandidateNeverConstrains: counted by Constrain's own hook, the judged candidates enter
// Constrain no time over inputs whose charts accept, toys and sources, and V1's side over the same inputs enters it
// once per input its chart accepts, so the judged differential never compares V1 with itself.
func TestForestEqualityJudgedCandidateNeverConstrains(t *testing.T) {
	tab := committedTables(t)
	srcs := []string{"1 + 2 * 3;", "foo -$x * $y;", "do $f || 1;", "$a <=> $b <=> $c;", "print $fh [];", unionShortcutInput}
	ty := loadToy(t, "mixed")
	sentences := ty.sentences[:40]
	accepted, n0 := 0, constrainEntries.Load()
	for _, src := range srcs {
		if f, _ := JudgedForest(tab, []byte(src), Scope{}); f != nil {
			accepted++
		}
	}
	for _, s := range sentences {
		f, err := JudgedToyForest(ty.tab, toyLattice(ty.tab, s.words), ty.prec)
		if err != nil {
			t.Fatal(err)
		}
		if f != nil {
			accepted++
		}
	}
	cand := constrainEntries.Load() - n0
	charts, n1 := 0, constrainEntries.Load()
	for _, src := range srcs {
		s, _ := lex.Lex([]byte(src), tab.Tokens)
		if c, err := recognize.Recognize(V1Recognizer(tab), lex.Linear(s)); err == nil && c.Accepted() {
			charts++
		}
		V1Forest(tab, []byte(src), Scope{})
	}
	for _, s := range sentences {
		l := toyLattice(ty.tab, s.words)
		if c, err := recognize.Recognize(V1Recognizer(ty.tab), l); err == nil && c.Accepted() {
			charts++
		}
		if _, err := V1ToyForest(ty.tab, l, ty.prec); err != nil {
			t.Fatal(err)
		}
	}
	v1 := constrainEntries.Load() - n1
	t.Logf("over %d inputs: the judged candidates accepted %d and entered Constrain %d times; V1's charts accepted %d and V1 entered Constrain %d times",
		len(srcs)+len(sentences), accepted, cand, charts, v1)
	if accepted == 0 || cand != 0 {
		t.Errorf("the judged candidates accepted %d inputs and entered Constrain %d times; want some accepted, and no entry", accepted, cand)
	}
	if charts == 0 || v1 != int64(charts) {
		t.Errorf("V1's side entered Constrain %d times over %d inputs its charts accept; want one each", v1, charts)
	}
}
