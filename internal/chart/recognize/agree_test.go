// ABOUTME: Holds the recognizer equal to the test-only reference on toy grammars and a deterministic sample of the corpus.
// ABOUTME: Acceptance, Inserted and Expected at every node; a refused file must be refused at the same node by both.

package recognize

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

// agreeToys are the grammars of TestReferenceRecognizer: left and right
// recursion, nullable chains, ambiguity and a dead end, then statements
// ending in a ; that sit in braces, where the slipper applies.
var agreeToys = []struct {
	name  string
	rules []string
	words []string
}{
	{"left recursion", []string{"S ::= S A | A"}, []string{"A", "B"}},
	{"right recursion", []string{"S ::= A S | A"}, []string{"A", "B"}},
	{"nullable chain", []string{"S ::= X Y Z", "X ::= Y", "Y ::= Z |", "Z ::= A |"}, []string{"A", "B"}},
	{"ambiguity", []string{"S ::= S S | A"}, []string{"A", "B"}},
	{"dead end", []string{"S ::= A B | A C D | E"}, []string{"A", "B", "C", "E"}},
	{"statements in braces", slipperGrammar, []string{"X", ";", "{", "}", "HASH"}},
	{"statements that are a ;", slipperEmptyGrammar, []string{"X", ";", "{", "}"}},
}

// TestReferenceRecognizer: the recognizer, under every Options combination,
// accepts what the reference accepts, inserts a ; at the nodes it does, and
// expects what it expects at every node, over every input of up to five
// tokens of each toy grammar and over the corpus sample.
func TestReferenceRecognizer(t *testing.T) {
	for _, toy := range agreeToys {
		tab := toyTables(t, toy.rules...)
		warm := map[Options]*Recognizer{}
		build := func(o Options) *Recognizer {
			if warm[o] == nil {
				warm[o] = New(tab, o)
			}
			return warm[o]
		}
		inputs := [][]string{nil}
		for n := 0; n < len(inputs); n++ {
			if len(inputs[n]) < 5 {
				for _, w := range toy.words {
					inputs = append(inputs, append(slices.Clone(inputs[n]), slip.Replace(w)))
				}
			}
		}
		for _, in := range inputs {
			if d := oracleDisagreement(tab, build, tokens(tab, in...)); d != "" {
				t.Fatalf("%s %q: %s", toy.name, in, d)
			}
		}
	}
	tab, files := corpusSample(t)
	for _, f := range files {
		if d := oracleDisagreement(tab, builds(tab), f.lattice(len(f.l.Edges))); d != "" {
			t.Errorf("%s: %s", f.name, d)
		}
	}
}

// sampleFile is one file of the corpus sample, lexed against the committed
// grammar, and the token count it is cut at, mid-statement: after the first
// =, else after the first token that is not a ; (use, package, ...); -1 for a
// file with none, which has no mid-statement point.
type sampleFile struct {
	name string
	l    *lex.Lattice[lex.None]
	cut  int
}

// maxTokens bounds what the reference reads of a file; it is cubic.
const maxTokens = 300

// lattice is the first n tokens of f, as a lattice of their own.
func (f sampleFile) lattice(n int) *lex.Lattice[lex.None] {
	n = min(n, maxTokens, len(f.l.Edges))
	return &lex.Lattice[lex.None]{Nodes: f.l.Nodes[:n+1], Edges: f.l.Edges[:n]}
}

// corpusSample is the committed grammar's tables and the sample: std5k.pl,
// every port snippet, and every 24th Guacamole case in name order.
func corpusSample(t *testing.T) (*grammar.Tables, []sampleFile) {
	t.Helper()
	g, ref, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	tab, err := grammar.Generate(g, ref)
	if err != nil {
		t.Fatal(err)
	}
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
	var files []sampleFile
	for _, p := range paths {
		src, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		s, _ := lex.Lex(src, tab.Tokens)
		f := sampleFile{name: p, l: lex.Linear(s), cut: -1}
		for i, tok := range s.Tokens {
			r := s.Raw[tok.Index]
			if string(src[r.Start:r.End]) == "=" {
				f.cut = i + 1
				break
			}
			if f.cut < 0 && string(src[r.Start:r.End]) != ";" {
				f.cut = i + 1 // else after the first token that is not a ;
			}
		}
		files = append(files, f)
	}
	return tab, files
}

// deadNode is the first node of c with no Earley set, or -1.
func deadNode(c *Chart) int {
	return slices.IndexFunc(c.Sets, func(s *Set) bool { return s == nil })
}

// TestReferenceAgreesOnTruncatedFiles: each sample file cut mid-statement
// (see sampleFile), is refused by the recognizer and the reference at the same
// node, with the same Expected there.
func TestReferenceAgreesOnTruncatedFiles(t *testing.T) {
	tab, files := corpusSample(t)
	cut, skipped := 0, 0
	for _, f := range files {
		if f.cut < 0 {
			skipped++
			t.Logf("%s: no token to cut after but ;", f.name)
			continue
		}
		cut++
		l := f.lattice(f.cut)
		c := mustRecognize(t, New(tab, Options{}), l)
		sets, _, accepted := refRun(tab, tokenIDs(l))
		wantDead := slices.IndexFunc(sets, func(s []refItem) bool { return len(s) == 0 })
		if c.Accepted() || accepted {
			t.Errorf("%s: cut after token %d, the recognizer accepts %v and the reference %v", f.name, f.cut, c.Accepted(), accepted)
		}
		if got := deadNode(c); got != wantDead {
			t.Errorf("%s: the recognizer refuses at node %d, the reference at %d", f.name, got, wantDead)
		}
		if d := oracleDisagreement(tab, builds(tab), l); d != "" {
			t.Errorf("%s: %s", f.name, d)
		}
	}
	if skipped*10 > len(files) { // a few files hold nothing to cut; most must be cut
		t.Errorf("only %d of %d files were cut", cut, len(files))
	}
}

// TestReferenceRecognizerCatchesDroppedTransition: the committed tables with
// the start state's transition on a sample file's first token removed
// disagree with the reference reading the intact rules. The file is the first
// of the sample whose first token the grammar takes: std5k.pl once the port
// reaches it, and until then (its first token, package, has no terminal
// yet) the next file the grammar takes.
func TestReferenceRecognizerCatchesDroppedTransition(t *testing.T) {
	tab, files := corpusSample(t)
	for _, f := range files {
		l := f.lattice(maxTokens)
		if len(l.Edges) == 0 || mustRecognize(t, New(tab, Options{}), l).Sets[1] == nil {
			continue
		}
		if d := oracleDisagreement(tab, builds(tab), l); d != "" {
			t.Fatalf("%s: the intact tables disagree: %s", f.name, d)
		}
		first := tab.TerminalSets[l.Edges[0].Token.TerminalSet]
		skewed := copyTables(tab)
		st := &skewed.States[0]
		i := slices.IndexFunc(st.Transitions, func(tr grammar.Transition) bool { return slices.Contains(first, tr.Symbol) })
		if i < 0 {
			t.Fatalf("%s: the start state has no transition on its first token, set %v", f.name, first)
		}
		st.Transitions = slices.Delete(st.Transitions, i, i+1)
		if d := oracleDisagreement(tab, builds(skewed), l); !strings.Contains(d, "the reference") {
			t.Errorf("%s: the tables without the transition were reported as %q", f.name, d)
		}
		return
	}
	t.Fatal("the grammar takes the first token of no file of the sample")
}
