// ABOUTME: FuzzRecognize, the recognizer's fuzz target, and the oracle that holds it to the reference recognizer.
// ABOUTME: Acceptance and Expected agree with reference_test.go's under every Options combination, on stmtGrammar and fuzzed small grammars. Run by `make fuzz-chart-recognize`.

package recognize

import (
	"fmt"
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

// withDeadEnd is l with a path that goes nowhere: edge i's token again, from
// node i to a new node placed after node i+1, so node i+1 and the new node
// are inside a branch. The last node is unchanged, so acceptance is too.
// l needs at least two edges, and i at most len(l.Edges)-2.
func withDeadEnd(l *lex.Lattice[lex.None], i int) *lex.Lattice[lex.None] {
	at := i + 2
	shift := func(n int) int {
		if n >= at {
			return n + 1
		}
		return n
	}
	out := &lex.Lattice[lex.None]{Nodes: slices.Insert(slices.Clone(l.Nodes), at, l.Nodes[i+1])}
	for _, e := range l.Edges {
		out.Edges = append(out.Edges, lex.Edge{From: shift(e.From), To: shift(e.To), Token: e.Token})
	}
	out.Edges = append(out.Edges, lex.Edge{From: i, To: at, Token: l.Edges[i].Token})
	return out
}

// stmtTerms are stmtGrammar's terminals; a fuzz input byte b is stmtTerms[b % 9].
var stmtTerms = []string{"My", "Var", "Assign", "Num", "Semi", "Print", "Add", "LBrace", "RBrace"}

// fuzzLattice is src, a byte to a terminal's singleton set, as a linear
// lattice over tab, so every input the fuzzer produces is a valid terminal
// string.
func fuzzLattice(tab *grammar.Tables, src string) *lex.Lattice[lex.None] {
	terms := make([]string, len(src))
	for i := range src {
		terms[i] = stmtTerms[int(src[i])%len(stmtTerms)]
	}
	return tokens(tab, terms...)
}

// enc is the fuzz input for the terminal names in words.
func enc(words string) string {
	var b []byte
	for _, w := range strings.Fields(words) {
		i := slices.Index(stmtTerms, w)
		if i < 0 {
			panic("enc: no terminal " + w)
		}
		b = append(b, byte(i))
	}
	return string(b)
}

// TestFuzzInputsAreTerminalStrings: whatever bytes the fuzzer produces, each
// token of the lattice names a terminal of stmtGrammar, so no mutation is
// refused at its first token for want of a terminal.
func TestFuzzInputsAreTerminalStrings(t *testing.T) {
	tab := toyStmt(t)
	term := map[string]bool{}
	for _, s := range tab.Symbols {
		term[s.Name] = s.Terminal
	}
	for _, src := range append([]string{"Print Num SQmi Semi", "0A", "\xff\x00 z"}, fuzzSeeds...) {
		l := fuzzLattice(tab, src)
		if len(l.Edges) != len(src) {
			t.Errorf("%q: %d tokens, want one per byte", src, len(l.Edges))
		}
		for _, e := range l.Edges {
			if set := tab.TerminalSets[e.Token.TerminalSet]; len(set) != 1 || !term[tab.Symbols[set[0]].Name] {
				t.Errorf("%q: token set %v is not one terminal", src, set)
			}
		}
	}
}

// fuzzSeeds is FuzzRecognize's seed corpus, some accepted and some refused.
var fuzzSeeds = []string{
	"", enc(decl), enc("Print Num"), enc("Num Num"), enc("Var Var"), enc("My Var Assign"),
	enc(decl + " Print Var Add Num Semi"),
	enc("LBrace Print Num Semi LBrace " + decl + " RBrace RBrace"),
	enc("LBrace LBrace Print Num Semi RBrace"),
}

// fuzzGrammar is gram, bytes, as a small grammar's statements: N0 is the
// start symbol, and a byte b is a symbol, one of N0 to N2 or T0 to T2 by b%6,
// so the start symbol may sit on a right-hand side. A rule is three reads: a
// left-hand side (b%3), a length (b%4), then that many symbols. A symbol with
// no rule is a terminal, as toyTables has it.
func fuzzGrammar(gram string) []string {
	var out []string
	for i := 0; i+1 < len(gram); {
		rule := fmt.Sprintf("N%d ::=", gram[i]%3)
		n := int(gram[i+1] % 4)
		i += 2
		for ; n > 0 && i < len(gram); n-- {
			rule += fmt.Sprintf(" %c%d", "NNNTTT"[gram[i]%6], gram[i]%3)
			i++
		}
		out = append(out, rule)
	}
	return out
}

// checkGrammar holds the recognizer to the reference on a fuzzed grammar:
// src picks the input's first 24 tokens among the grammar's terminals. A grammar the
// generator refuses is skipped.
func checkGrammar(t *testing.T, src, gram string) {
	tab, err := buildToy(fuzzGrammar(gram))
	if err != nil {
		t.Skip(err)
	}
	var terms []string
	for _, s := range tab.Symbols {
		if s.Terminal {
			terms = append(terms, s.Name)
		}
	}
	var words []string
	for i := 0; i < min(len(src), 24) && len(terms) > 0; i++ { // the reference is cubic
		words = append(words, terms[int(src[i])%len(terms)])
	}
	if d := oracleDisagreement(tab, builds(tab), tokens(tab, words...)); d != "" {
		t.Fatalf("grammar %q, input %q: %s", fuzzGrammar(gram), words, d)
	}
}

// FuzzRecognize takes an input and a grammar; with no grammar, the input is
// bytes of stmtGrammar's terminals, and otherwise gram is a small grammar.
func FuzzRecognize(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s, "")
	}
	// start on a right-hand side, N0 ::= T0 N0 | T0 and N1 ::= N0 T0.
	// The start symbol on a right-hand side, where a Leo chain once ran up
	// through a start item from node 0: S ::= V B T, T ::= A | V, V ::= A | S
	// on A B A (N0 is S, N1 is V, N2 is T, T0 is A, T1 is B).
	f.Add("\x00\x01\x00", "\x00\x03\x01\x04\x02\x02\x01\x03\x02\x01\x01\x01\x01\x03\x01\x01\x00")
	tab := toyStmt(f)
	// warm is shared across inputs, so its cache serves earlier inputs'
	// transitions to later ones.
	var mu sync.Mutex
	warm := New(tab, Options{})
	f.Fuzz(func(t *testing.T, src, gram string) {
		if gram != "" {
			checkGrammar(t, src, gram)
			return
		}
		lattices := []*lex.Lattice[lex.None]{fuzzLattice(tab, src)}
		if n := len(lattices[0].Edges); n >= 2 {
			lattices = append(lattices, withDeadEnd(lattices[0], len(src)%(n-1)))
		}
		if d := oracleDisagreement(tab, builds(tab), lattices[0]); d != "" {
			t.Fatalf("%q: %s", src, d)
		}
		refSets := refRun(tab, tokenIDs(lattices[0]))
		want := refAccepts(tab, refSets)
		mu.Lock()
		defer mu.Unlock()
		for i, l := range lattices {
			cold := mustRecognize(t, New(tab, Options{NoCache: true}), l)
			if cold.Accepted() != want {
				t.Fatalf("%q lattice %d: accepted %v cold, want %v", src, i, cold.Accepted(), want)
			}
			for range 2 {
				c := mustRecognize(t, warm, l)
				if c.Accepted() != want {
					t.Fatalf("%q lattice %d: accepted %v with the cache, %v without", src, i, c.Accepted(), want)
				}
				// Expected is memoized per recognizer across inputs, so the warm
				// chart's answers are held to the reference (the dead-end
				// lattice has other nodes: to a cold chart's).
				for pos := -1; pos <= len(c.Sets); pos++ {
					got, err := c.Expected(pos)
					if pos < 0 || pos >= len(c.Sets) {
						continue // out of range: an error, and no panic
					}
					var wantNames []string
					if i == 0 {
						wantNames = refExpected(tab, refSets[pos])
					} else {
						wantNames, _ = cold.Expected(pos)
					}
					if err != nil || !slices.Equal(got, wantNames) {
						t.Fatalf("%q lattice %d node %d: Expected %v (err %v) with the cache, %v without", src, i, pos, got, err, wantNames)
					}
				}
			}
		}
	})
}

// TestWarmExpectedMatchesReference: one recognizer, its Expected memo warmed
// by a fixed sequence of inputs, answers each as the reference does. A memo
// entry served at a node whose origin sets differ from the ones it read is
// the 2eec class, which only a later input can expose.
func TestWarmExpectedMatchesReference(t *testing.T) {
	tab := toyStmt(t)
	warm := New(tab, Options{})
	rng := rand.New(rand.NewSource(1))
	inputs := slices.Clone(fuzzSeeds)
	for range 400 {
		b := make([]byte, rng.Intn(12))
		for i := range b {
			b[i] = byte(rng.Intn(len(stmtTerms)))
		}
		inputs = append(inputs, string(b))
	}
	for _, src := range inputs {
		l := fuzzLattice(tab, src)
		c := mustRecognize(t, warm, l)
		for pos, items := range refRun(tab, tokenIDs(l)) {
			if got, err := c.Expected(pos); err != nil || !slices.Equal(got, refExpected(tab, items)) {
				t.Fatalf("%q node %d: Expected %v (err %v) from the warm recognizer, %v from the reference", src, pos, got, err, refExpected(tab, items))
			}
		}
	}
}

// TestFuzzSeedsAcceptAndRefuse: the seeds hold an accepted and a refused
// input, and the dead-end lattice keeps each verdict, so FuzzRecognize cannot
// pass by never reaching a refusal.
func TestFuzzSeedsAcceptAndRefuse(t *testing.T) {
	tab := toyStmt(t)
	var accepted, refused int
	for _, src := range fuzzSeeds {
		l := fuzzLattice(tab, src)
		want := mustRecognize(t, New(tab, Options{}), l).Accepted()
		if want {
			accepted++
		} else {
			refused++
		}
		if len(l.Edges) < 2 {
			continue
		}
		d := withDeadEnd(l, 0)
		if len(d.Nodes) != len(l.Nodes)+1 {
			t.Errorf("%q: the dead end added %d nodes, want 1", src, len(d.Nodes)-len(l.Nodes))
		}
		if got := mustRecognize(t, New(tab, Options{}), d).Accepted(); got != want {
			t.Errorf("%q: accepted %v with a dead end, %v without", src, got, want)
		}
	}
	if accepted == 0 || refused == 0 {
		t.Errorf("%d accepted and %d refused seeds, want both", accepted, refused)
	}
}

// optionCombos is every combination of Options' three switches.
// shortcut: the three switches are hard-coded; a new Options field needs a
// bit here, and the count raised to match.
var optionCombos = func() (out []Options) {
	for m := range 8 {
		out = append(out, Options{NoCache: m&1 != 0, NoLookahead: m&2 != 0, NoLeo: m&4 != 0})
	}
	return out
}()

// tokenIDs is the terminal-set ids of a linear lattice's tokens, in order.
func tokenIDs(l *lex.Lattice[lex.None]) []int {
	ids := make([]int, len(l.Edges))
	for i, e := range l.Edges {
		ids[i] = e.Token.TerminalSet
	}
	return ids
}

// oracleDisagreement is "" where the recognizers build makes, under every
// Options combination, accept l and expect at each node what the reference
// does over rules; otherwise it says what differs. l is a linear lattice, one
// edge per node but the first, so edge i is token i.
func oracleDisagreement(rules *grammar.Tables, build func(Options) *Recognizer, l *lex.Lattice[lex.None]) string {
	sets := refRun(rules, tokenIDs(l))
	want := refAccepts(rules, sets)
	for _, opts := range optionCombos {
		c, err := Recognize(build(opts), l)
		if err != nil {
			return fmt.Sprintf("%+v: %v", opts, err)
		}
		if got := c.Accepted(); got != want {
			return fmt.Sprintf("%+v: accepted %v, the reference %v", opts, got, want)
		}
		for pos, items := range sets {
			if got, err := c.Expected(pos); err != nil || !slices.Equal(got, refExpected(rules, items)) {
				return fmt.Sprintf("%+v node %d: Expected %v (err %v), the reference %v", opts, pos, got, err, refExpected(rules, items))
			}
		}
	}
	return ""
}

// builds is the recognizers over tab, one fresh per call.
func builds(tab *grammar.Tables) func(Options) *Recognizer {
	return func(o Options) *Recognizer { return New(tab, o) }
}

// copyTables is tab with its states copied one level down, so a test can
// alter a state without altering tab.
func copyTables(tab *grammar.Tables) *grammar.Tables {
	c := *tab
	c.States = slices.Clone(tab.States)
	for i := range c.States {
		c.States[i].Transitions = slices.Clone(c.States[i].Transitions)
		c.States[i].Expected = slices.Clone(c.States[i].Expected)
	}
	return &c
}

// without is tab with the first transition on the terminal sym removed from
// the first state that has one.
func without(t *testing.T, tab *grammar.Tables, sym string) *grammar.Tables {
	c := copyTables(tab)
	for s, st := range c.States {
		for i, tr := range st.Transitions {
			if tab.Symbols[tr.Symbol].Name == sym {
				c.States[s].Transitions = slices.Delete(st.Transitions, i, i+1)
				return c
			}
		}
	}
	t.Fatalf("no state has a transition on %s", sym)
	return nil
}

// disagreements is oracleDisagreement over each of the fuzz seeds.
func disagreements(tab *grammar.Tables, build func(Options) *Recognizer) (out []string) {
	for _, src := range fuzzSeeds {
		if d := oracleDisagreement(tab, build, fuzzLattice(tab, src)); d != "" {
			out = append(out, d)
		}
	}
	return out
}

// TestOracleCatchesSkew: the oracle is green on intact tables and reports a
// recognizer whose tables lack one transition, against the intact rules.
func TestOracleCatchesSkew(t *testing.T) {
	tab := toyStmt(t)
	if d := disagreements(tab, builds(tab)); len(d) > 0 {
		t.Fatalf("intact tables disagree with the reference: %v", d)
	}
	if d := disagreements(tab, builds(without(t, tab, "Semi"))); len(d) == 0 {
		t.Error("the oracle did not report tables without a Semi transition")
	}
}

// withExpected is tab with Expected changed in the first state it fits: name
// dropped from the first state that expects it, or, with add, added to the
// first state that expects something but not name. Transitions are left, so
// acceptance does not change.
func withExpected(t *testing.T, tab *grammar.Tables, name string, add bool) *grammar.Tables {
	sym := slices.IndexFunc(tab.Symbols, func(s grammar.TableSymbol) bool { return s.Name == name })
	c := copyTables(tab)
	for s, st := range c.States {
		i := slices.Index(st.Expected, sym)
		switch {
		case !add && i >= 0:
			c.States[s].Expected = slices.Delete(st.Expected, i, i+1)
			return c
		case add && i < 0 && len(st.Expected) > 0:
			c.States[s].Expected = slices.Concat(st.Expected, []int{sym})
			return c
		}
	}
	t.Fatalf("no state fits (add %v, %s)", add, name)
	return nil
}

// TestOracleCatchesExpectedSkew: tables altered so acceptance is unchanged
// but one state's expected terminals differ, fewer or more, are reported as an
// Expected disagreement, never an acceptance one, and at a node before the
// last, so an oracle that reads only the last node or only one direction
// fails.
func TestOracleCatchesExpectedSkew(t *testing.T) {
	tab := toyStmt(t)
	for name, skewed := range map[string]*grammar.Tables{
		"Num dropped": withExpected(t, tab, "Num", false),
		"Semi added":  withExpected(t, tab, "Semi", true),
	} {
		early := false
		for _, src := range fuzzSeeds {
			l := fuzzLattice(tab, src)
			d := oracleDisagreement(tab, builds(skewed), l)
			if d == "" {
				continue
			}
			_, at, _ := strings.Cut(d, "node ")
			var pos int
			if !strings.Contains(d, "Expected") {
				t.Errorf("%s: a skew in Expected alone was reported as: %s", name, d)
			} else if _, err := fmt.Sscanf(at, "%d", &pos); err == nil && pos < len(l.Edges) {
				early = true
			}
		}
		if !early {
			t.Errorf("%s: no seed reported it at a node before the last", name)
		}
	}
}

// TestOracleCatchesSkewUnderOneOption: a recognizer wrong under one Options
// combination and right under the other seven is reported, whichever the
// combination and whether the skew is a transition or Expected, fewer or more.
func TestOracleCatchesSkewUnderOneOption(t *testing.T) {
	if n := reflect.TypeOf(Options{}).NumField(); n != 3 {
		t.Fatalf("Options has %d switches and optionCombos covers 3: extend optionCombos", n)
	}
	tab := toyStmt(t)
	for name, bad := range map[string]*grammar.Tables{
		"transition":  without(t, tab, "Semi"),
		"Num dropped": withExpected(t, tab, "Num", false),
		"Semi added":  withExpected(t, tab, "Semi", true),
	} {
		for _, only := range optionCombos {
			build := func(o Options) *Recognizer {
				if o == only {
					return New(bad, o)
				}
				return New(tab, o)
			}
			d := disagreements(tab, build)
			if len(d) == 0 {
				t.Errorf("%s: a skew under %+v alone was not reported", name, only)
			}
			for _, msg := range d {
				if !strings.Contains(msg, fmt.Sprintf("%+v", only)) {
					t.Errorf("%s: a skew under %+v was reported as: %s", name, only, msg)
				}
			}
		}
	}
}
