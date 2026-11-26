// ABOUTME: Tests the judged recognizer's reduce side: its two precedence sources, the verdict per completion and
// ABOUTME: parent state, the vacuous case, refusal records beside the sets, and Leo crossing a refused completion.

package recognize

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/prectab"
)

// mustJudged is NewJudged(tab, p, opts), failing tb on its error.
func mustJudged(tb testing.TB, tab *grammar.Tables, p *prectab.Precedence, opts Options) *Recognizer {
	tb.Helper()
	r, err := NewJudged(tab, p, opts)
	if err != nil {
		tb.Fatal(err)
	}
	return r
}

// precToy is a toy's tables and the Precedence of levels over them, rulePrec naming the first rules' %prec tokens
// ("" for none, the rest none) and refused the rules bison's states hold that the tables lack.
func precToy(tb testing.TB, levels []grammar.Level, rulePrec []string, refused []grammar.Rule, statements ...string) (*grammar.Tables, *prectab.Precedence) {
	tb.Helper()
	tab := toyTables(tb, statements...)
	p, err := prectab.NewPrecedence(tab, levels, append(rulePrec, make([]string, len(tab.Rules)-len(rulePrec))...), refused)
	if err != nil {
		tb.Fatal(err)
	}
	return tab, p
}

// opToy is S ::= E | Q X, Q ::= NUM, E ::= E PLUS NUM %prec PLUS | E PLUS NUM TIMES NUM | E TIMES NUM %prec TIMES |
// NUM, PLUS below TIMES, both left: the state after E PLUS NUM shifts TIMES, which wins, so E PLUS NUM completes
// before PLUS and is refused before TIMES. X follows Q, never E.
func opToy(tb testing.TB) (*grammar.Tables, *prectab.Precedence) {
	return precToy(tb, []grammar.Level{{N: 1, Assoc: "left", Tokens: []string{"PLUS"}}, {N: 2, Assoc: "left", Tokens: []string{"TIMES"}}},
		[]string{"", "", "", "PLUS", "TIMES", "TIMES"}, nil,
		"S ::= E | Q X", "Q ::= NUM", "E ::= E PLUS NUM | E PLUS NUM TIMES NUM | E TIMES NUM | NUM")
}

// symbolID is the id of the symbol named name, -1 for none.
func symbolID(tab *grammar.Tables, name string) int {
	return slices.IndexFunc(tab.Symbols, func(s grammar.TableSymbol) bool { return s.Name == name })
}

// opLattice is NUM PLUS NUM, a token whose set is next's terminals, then NUM: node 3 is where E PLUS NUM completes
// from node 0, before that token. The set is appended to tab's.
func opLattice(tab *grammar.Tables, next ...string) *lex.Lattice[lex.None] {
	l := tokens(tab, "NUM", "PLUS", "NUM", "NUM", "NUM")
	var set []int
	for _, n := range next {
		set = append(set, symbolID(tab, n))
	}
	slices.Sort(set)
	l.Edges[3].Token.TerminalSet = len(tab.TerminalSets)
	tab.TerminalSets = append(tab.TerminalSets, set)
	return l
}

// opNode3 is the judged chart of opToy's opLattice(next) at node 3: its items, and whether a completion was refused
// there; and the no-precedence chart's items there.
func opNode3(t *testing.T, opts Options, next ...string) (judged []Item, refused bool, plain []Item) {
	t.Helper()
	tab, p := opToy(t)
	l := opLattice(tab, next...)
	c := mustRecognize(t, mustJudged(t, tab, p, opts), l)
	u := mustRecognize(t, NewNoPrecedence(tab, opts), l)
	return c.Sets[3].Items(), c.RefusedAt(3), u.Sets[3].Items()
}

// TestReduceSideRefusesForbiddenCompletion: a completion forbidden before every terminal of the lookahead that can
// follow its LHS is refused, recorded, and advances no parent, on opToy and on perly.y's 1 + 2 * 3;, while one
// allowed is kept; and each of perly.y's six parent-dependent pairs is judged by the parent state it completes into,
// refused under some and kept under others, as V1's parentForbid has it, also through the indirob toy's sentences.
func TestReduceSideRefusesForbiddenCompletion(t *testing.T) {
	t.Run("toy", func(t *testing.T) {
		judged, refused, plain := opNode3(t, Options{}, "TIMES")
		if !refused || len(judged) >= len(plain) {
			t.Errorf("before TIMES: refused %v, %d items judged and %d not; want a refusal and fewer items", refused, len(judged), len(plain))
		}
		judged, refused, plain = opNode3(t, Options{}, "PLUS")
		if refused || !slices.Equal(judged, plain) {
			t.Errorf("before PLUS: refused %v, items %v judged and %v not; want none refused, the same items", refused, judged, plain)
		}
	})
	t.Run("perly", func(t *testing.T) {
		tab := committedTables(t)
		s, _ := lex.Lex([]byte("1 + 2 * 3;"), tab.Tokens)
		l := lex.Linear(s)
		c := mustRecognize(t, mustJudged(t, tab, nil, Options{}), l)
		u := mustRecognize(t, NewNoPrecedence(tab, Options{}), l)
		if !c.Accepted() || !c.RefusedAt(3) || len(c.Sets[3].Items()) >= len(u.Sets[3].Items()) {
			t.Errorf("1 + 2 * 3;: accepted %v, a refusal at node 3 (before *) %v, %d items there judged and %d not; want accepted, refused, fewer",
				c.Accepted(), c.RefusedAt(3), len(c.Sets[3].Items()), len(u.Sets[3].Items()))
		}
	})
	t.Run("parent-pairs", func(t *testing.T) {
		tab := committedTables(t)
		r := mustJudged(t, tab, nil, Options{})
		j, p := r.judge, r.judge.p
		pairs, both := 0, 0
		for k, v := range p.ReduceAt {
			if v != 2 {
				continue
			}
			pairs++
			rule, term := k/p.NT, k%p.NT
			lhs := tab.Rules[rule].LHS
			la := r.g.bitsOf([]int{term})
			follows := r.g.follow[lhs].intersects(la)
			kept, refused := 0, 0
			for s := range tab.States {
				pi := j.parentItem(s, lhs)
				if pi < 0 {
					continue
				}
				got, _ := j.verdict(rule, r.g.follow[lhs], la, pi)
				// Where the terminal cannot follow the LHS the completion is refused vacuously, whatever the parent.
				if want := !follows || !p.ReducesFree(rule, term, pi); got != want {
					t.Errorf("%s before %s under state %d: refused %v, want %v", tab.Symbols[lhs].Name, tab.Symbols[term].Name, s, got, want)
				}
				if got {
					refused++
				} else {
					kept++
				}
			}
			t.Logf("rule %d (%s) before %s, in FOLLOW %v: refused under %d parent states, kept under %d", rule, tab.Symbols[lhs].Name, tab.Symbols[term].Name, follows, refused, kept)
			if follows {
				both++
				if refused == 0 || kept == 0 {
					t.Errorf("rule %d before %s: refused under %d parent states, kept under %d; want some of each", rule, tab.Symbols[term].Name, refused, kept)
				}
			}
		}
		if both == 0 {
			t.Error("no parent-dependent pair's terminal can follow its LHS, so no parent state decides one")
		}
		if pairs != 6 {
			t.Errorf("%d parent-dependent pairs on perly.y; want 6", pairs)
		}
	})
	t.Run("indirob-toy", func(t *testing.T) {
		tab, p := fixtureToy(t, "indirob")
		indirob := symbolID(tab, "indirob")
		r := mustJudged(t, tab, p, Options{})
		for _, tc := range []struct {
			words   string
			refused bool
		}{{"LSTOP VAR LB NUM RB", true}, {"SNAIL VAR LB NUM RB", false}} {
			c := mustRecognize(t, r, tokens(tab, strings.Fields(tc.words)...))
			got := false
			for _, it := range c.Sets[1].Items() {
				got = got || c.Refused(2, indirob, 1, it.State)
			}
			if got != tc.refused || !c.Accepted() {
				t.Errorf("%s: indirob from node 1 refused before LB %v, accepted %v; want refused %v, accepted", tc.words, got, c.Accepted(), tc.refused)
			}
		}
	})
}

// fixtureToy loads the committed toy fixture name (testdata/precedence/toys) as tables and their Precedence.
func fixtureToy(t *testing.T, name string) (*grammar.Tables, *prectab.Precedence) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "testdata", "precedence", "toys", name+".txt"))
	if err != nil {
		t.Fatal(err)
	}
	var levels []grammar.Level
	var statements, rulePrec []string
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		switch {
		case len(f) > 1 && f[0] == "level":
			levels = append(levels, grammar.Level{N: len(levels) + 1, Assoc: f[1], Tokens: f[2:]})
		case len(f) > 1 && f[0] == "rule":
			lhs, rhs, _ := strings.Cut(strings.Join(f[2:], " "), ":")
			rhs, prec, _ := strings.Cut(rhs, "%prec")
			statements = append(statements, lhs+" ::= "+rhs)
			rulePrec = append(rulePrec, strings.TrimSpace(prec))
		}
	}
	return precToy(t, levels, rulePrec, nil, statements...)
}

// TestReduceSideMixedVerdictSetKept: E PLUS NUM before {PLUS, TIMES}, forbidden before TIMES and allowed before PLUS,
// both in FOLLOW(E), is kept: nothing refused, and the set the no-precedence recognizer builds.
func TestReduceSideMixedVerdictSetKept(t *testing.T) {
	judged, refused, plain := opNode3(t, Options{}, "PLUS", "TIMES")
	if refused || !slices.Equal(judged, plain) {
		t.Errorf("before {PLUS, TIMES}: refused %v, items %v judged and %v not; want none refused, the same items", refused, judged, plain)
	}
	if _, refused, _ := opNode3(t, Options{}, "TIMES"); !refused {
		t.Error("before TIMES alone the completion is not refused, so the mixed set's keeping proves nothing")
	}
}

// TestReduceSideIgnoresTerminalsOutsideFollow: E PLUS NUM before {TIMES, X}, X outside FOLLOW(E), is refused as
// before TIMES alone: a terminal that cannot follow E does not save it.
func TestReduceSideIgnoresTerminalsOutsideFollow(t *testing.T) {
	judged, refused, _ := opNode3(t, Options{}, "TIMES", "X")
	alone, _, _ := opNode3(t, Options{}, "TIMES")
	if !refused || !slices.Equal(judged, alone) {
		t.Errorf("before {TIMES, X}: refused %v, items %v; want refused, as before TIMES alone, %v", refused, judged, alone)
	}
}

// TestReduceSideNoFollowingTerminal pins the vacuous case: before X, which cannot follow E, no terminal of the set
// can follow the completion's LHS, and the rule holds vacuously, so the completion is refused, with lookahead and
// with NoLookahead alike, and recorded nowhere (no parent advances over E there from any origin, so Build has
// nothing to pair). Without precedence, NoLookahead completes it, so the case is live.
func TestReduceSideNoFollowingTerminal(t *testing.T) {
	with, recWith, _ := opNode3(t, Options{}, "X")
	without, recWithout, plain := opNode3(t, Options{NoLookahead: true}, "X")
	if !slices.Equal(with, without) || recWith || recWithout {
		t.Errorf("before X: items %v with lookahead, %v without, records %v and %v; want the same items, no records", with, without, recWith, recWithout)
	}
	if len(plain) <= len(without) {
		t.Errorf("before X under NoLookahead: %d items without precedence, %d judged; want the completion refused", len(plain), len(without))
	}
}

// TestRefusalRecordsOutsideSetIdentity: on one recognizer, node 3 of NUM PLUS NUM before TIMES and before X holds the
// same items, E PLUS NUM's completion refused before both, by precedence before TIMES, which records it, and
// vacuously before X, which does not: the two nodes share one interned *Set while their records differ.
func TestRefusalRecordsOutsideSetIdentity(t *testing.T) {
	tab, p := opToy(t)
	r := mustJudged(t, tab, p, Options{})
	a := mustRecognize(t, r, opLattice(tab, "TIMES"))
	b := mustRecognize(t, r, opLattice(tab, "X"))
	if a.RefusedAt(3) == b.RefusedAt(3) {
		t.Fatalf("records at node 3: %v before TIMES, %v before X; want them to differ", a.RefusedAt(3), b.RefusedAt(3))
	}
	if a.Sets[3] != b.Sets[3] || !b.Interned[3] {
		t.Errorf("node 3 before TIMES and before X differ only in their records but do not share one interned set")
	}
}

// TestRefusalRecordsServedByCache: a cache hit serves the records its storing node kept, so a warm recognizer's
// chart refuses where a cold one does.
func TestRefusalRecordsServedByCache(t *testing.T) {
	tab, p := opToy(t)
	l := opLattice(tab, "TIMES")
	r := mustJudged(t, tab, p, Options{})
	mustRecognize(t, r, l)
	warm := mustRecognize(t, r, l)
	if !warm.Cached[3] || !warm.RefusedAt(3) {
		t.Errorf("node 3 on a warm recognizer: cached %v, records %v; want served from the cache with its records", warm.Cached[3], warm.RefusedAt(3))
	}
}

// powToy is S ::= E | E T NUM, E ::= NUM POW E | NUM, POW right below T left, with E ::= E T NUM in bison's states
// but not the tables (a refused rule): the state after NUM POW E shifts T there, so E ::= NUM POW E is refused
// before T, while the chart's set after POW holds one item awaiting E, a Leo item. A Leo chain up NUM POW NUM POW
// NUM before T crosses that refused completion.
func powToy(tb testing.TB) (*grammar.Tables, *prectab.Precedence) {
	name := func(s string) grammar.Symbol { return grammar.Symbol{Kind: grammar.Name, Text: s} }
	return precToy(tb, []grammar.Level{{N: 1, Assoc: "right", Tokens: []string{"POW"}}, {N: 2, Assoc: "left", Tokens: []string{"T"}}},
		nil, []grammar.Rule{{LHS: "E", RHS: []grammar.Symbol{name("E"), name("T"), name("NUM")}}},
		"S ::= E | E T NUM", "E ::= NUM POW E | NUM")
}

// TestLeoCrossesRefusedCompletion: on powToy, the judged recognizer with Leo and without agree on acceptance,
// Expected and the refusal records at every node, with lookahead and without, over inputs whose Leo chains are
// taken and one whose chain crosses a refused completion. It logs both counts and requires both non-zero.
func TestLeoCrossesRefusedCompletion(t *testing.T) {
	tab, p := powToy(t)
	steps, crossed := 0, 0
	for _, in := range []string{"NUM POW NUM POW NUM T NUM", "NUM POW NUM POW NUM", "NUM POW NUM POW NUM POW NUM", "NUM T NUM", "NUM POW NUM T NUM"} {
		l := tokens(tab, strings.Fields(in)...)
		for _, noLA := range []bool{false, true} {
			on := mustJudged(t, tab, p, Options{NoLookahead: noLA})
			off := mustJudged(t, tab, p, Options{NoLookahead: noLA, NoLeo: true})
			a, b := mustRecognize(t, on, l), mustRecognize(t, off, l)
			steps, crossed = steps+on.leoSteps, crossed+on.leoRefused
			if a.Accepted() != b.Accepted() {
				t.Errorf("%s, lookahead off %v: accepted %v with Leo, %v without", in, noLA, a.Accepted(), b.Accepted())
			}
			for pos := range a.Sets {
				x, _ := a.Expected(pos)
				y, _ := b.Expected(pos)
				if !slices.Equal(x, y) || !slices.Equal(recordsAt(a, pos), recordsAt(b, pos)) {
					t.Errorf("%s, lookahead off %v, node %d: Expected %v and records %v with Leo, %v and %v without", in, noLA, pos, x, recordsAt(a, pos), y, recordsAt(b, pos))
				}
			}
		}
	}
	t.Logf("leo steps: %d", steps)
	t.Logf("refused completions crossed: %d", crossed)
	if steps == 0 || crossed == 0 {
		t.Errorf("leo steps %d, refused completions crossed %d; want both non-zero", steps, crossed)
	}
}

// recordsAt is c's refusal records at node pos.
func recordsAt(c *Chart, pos int) []refusal {
	if c.refused == nil {
		return nil
	}
	return c.refused[pos]
}

// TestJudgedOverLevellessTablesRecognizesAsBefore: tables that are not perly.y's declare no levels, so the derived
// form judges nothing by that fact: over package recognize's level-less toy grammars and FuzzRecognize's seeds'
// grammars, under every Options combination, NewJudged with no precedence gives no error and the no-precedence
// recognizer's sets, Expected and acceptance at every node.
func TestJudgedOverLevellessTablesRecognizesAsBefore(t *testing.T) {
	type input struct {
		tab *grammar.Tables
		l   *lex.Lattice[lex.None]
	}
	same := func(t *testing.T, ins []input) {
		t.Helper()
		for _, in := range ins {
			for _, o := range []Options{{}, {NoLookahead: true}, {NoLeo: true}, {NoCache: true}, {NoLookahead: true, NoLeo: true}} {
				r, err := NewJudged(in.tab, nil, o)
				if err != nil {
					t.Fatalf("NewJudged: %v", err)
				}
				a, b := mustRecognize(t, r, in.l), mustRecognize(t, NewNoPrecedence(in.tab, o), in.l)
				if a.Accepted() != b.Accepted() {
					t.Fatalf("%v under %+v: accepted %v judged, %v not", in.l.Edges, o, a.Accepted(), b.Accepted())
				}
				for pos := range a.Sets {
					x, _ := a.Expected(pos)
					y, _ := b.Expected(pos)
					if !slices.Equal(a.Sets[pos].Items(), b.Sets[pos].Items()) || !slices.Equal(x, y) {
						t.Fatalf("%v under %+v, node %d: items %v and Expected %v judged, %v and %v not", in.l.Edges, o, pos, a.Sets[pos].Items(), x, b.Sets[pos].Items(), y)
					}
				}
			}
		}
		t.Logf("%d inputs", len(ins))
	}
	t.Run("recognize-toys", func(t *testing.T) {
		var ins []input
		for _, toy := range agreeToys {
			tab := toyTables(t, toy.rules...)
			words := [][]string{nil}
			for n := 0; n < len(words); n++ {
				if len(words[n]) < 4 {
					for _, w := range toy.words {
						words = append(words, append(slices.Clone(words[n]), slip.Replace(w)))
					}
				}
			}
			for _, w := range words {
				ins = append(ins, input{tab, tokens(tab, w...)})
			}
		}
		for _, g := range [][]string{rightRec, leoChainGrammar} {
			tab := toyTables(t, g...)
			for _, w := range []string{"A A A A A", "A A B A", "B A B A", "A C A"} {
				ins = append(ins, input{tab, tokens(tab, strings.Fields(w)...)})
			}
		}
		same(t, ins)
	})
	t.Run("fuzz-seeds", func(t *testing.T) {
		stmt := toyStmt(t)
		var ins []input
		for _, src := range fuzzSeeds {
			ins = append(ins, input{stmt, fuzzLattice(stmt, src)})
		}
		// FuzzRecognize's seeded grammar, on its seeded input A B A.
		g, err := buildToy(fuzzGrammar("\x00\x03\x01\x04\x02\x02\x01\x03\x02\x01\x01\x01\x01\x03\x01\x01\x00"))
		if err != nil {
			t.Fatal(err)
		}
		ins = append(ins, input{g, tokens(g, "T0", "T1", "T0")})
		same(t, ins)
	})
}

// TestJudgedJudgesCopyOfCommittedTables: the derived form decides perly.y by content, not by pointer: a copy of the
// committed tables and tables regenerated from grammar.Committed() are each judged with perly.y's precedence, their
// sets on the perly.y sample equal to the committed value's judged sets and different from the no-precedence
// recognizer's at some node.
func TestJudgedJudgesCopyOfCommittedTables(t *testing.T) {
	tab, files := corpusSample(t)
	sets := func(r *Recognizer) [][]Item {
		var out [][]Item
		for _, f := range files {
			c := mustRecognize(t, r, f.l)
			for _, s := range c.Sets {
				out = append(out, slices.Clone(s.Items()))
			}
		}
		return out
	}
	want := sets(mustJudged(t, tab, nil, Options{}))
	plain := sets(NewNoPrecedence(tab, Options{}))
	cp := *tab
	g, ref, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	regen, err := grammar.Generate(g, ref)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		tab  *grammar.Tables
	}{{"copy", &cp}, {"regenerated", regen}} {
		t.Run(tc.name, func(t *testing.T) {
			got := sets(mustJudged(t, tc.tab, nil, Options{}))
			differ := 0
			for i := range got {
				if !slices.Equal(got[i], plain[i]) {
					differ++
				}
				if !slices.Equal(got[i], want[i]) {
					t.Fatalf("%s: set %d differs from the committed value's judged set", tc.name, i)
				}
			}
			t.Logf("%s: %d sets, %d differing from the no-precedence recognizer's", tc.name, len(got), differ)
			if differ == 0 {
				t.Errorf("%s: judged as the no-precedence recognizer: nothing judged", tc.name)
			}
		})
	}
	// Rule for rule includes the RHS: perly.y's rule count and LHSes with one RHS symbol renamed are not perly.y's.
	t.Run("rhs-differs", func(t *testing.T) {
		other := cp
		other.Rules = slices.Clone(tab.Rules)
		r := slices.IndexFunc(other.Rules, func(r grammar.BNFRule) bool { return len(r.RHS) > 0 })
		other.Rules[r].RHS = append([]int{other.Rules[r].LHS}, other.Rules[r].RHS[1:]...)
		if d := deriveJudgement(&other); d.j != nil || d.err != nil {
			t.Errorf("rule %d's RHS changed: judged %v, error %v; want neither", r, d.j != nil, d.err)
		}
	})
}

// TestJudgedRefusesFailedDerivationOnPerlyTables: with the derivation of perly.y's precedence forced to fail through
// its grammar.Derived hook (prectab.Parsed), as TestPrecedenceGenerationErrorRefusesEveryFile forces it, the derived
// form returns that error from NewJudged and from Recognize for every input of the perly.y sample, and never judges
// nothing; the no-precedence recognizer over the same value still accepts the sample's files.
func TestJudgedRefusesFailedDerivationOnPerlyTables(t *testing.T) {
	_, files := corpusSample(t)
	g, ref, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	tab, err := grammar.Generate(g, ref) // its own value, whose precedence the forced error stands as
	if err != nil {
		t.Fatal(err)
	}
	forced := errors.New("forced: perly.y's precedence failed to derive")
	grammar.Derived(tab, func(*grammar.Tables) prectab.Parsed { return prectab.Parsed{Err: forced} })
	r, err := NewJudged(tab, nil, Options{})
	if err != forced {
		t.Fatalf("NewJudged's error %v; want the forced one", err)
	}
	accepted := 0
	for _, f := range files {
		if _, err := Recognize(r, f.l); err != forced {
			t.Fatalf("%s: Recognize's error %v; want the forced one", f.name, err)
		}
		if mustRecognize(t, NewNoPrecedence(tab, Options{}), f.l).Accepted() {
			accepted++
		}
	}
	if accepted == 0 {
		t.Error("the no-precedence recognizer accepts no file of the sample over the forced value")
	}
}
