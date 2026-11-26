// ABOUTME: The set-1626 observer (design §3, §5): no input's token carrying {HASHBRACK, PERLY_BRACE_OPEN} follows a completed indirob ::= scalar.
// ABOUTME: Over the in-tree corpus and testdata/set1626-inputs.txt here; perl.git's half is in precedence_perlgit_test.go.

package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
)

// set1626Inputs is the committed constructed inputs, each from a "### <name>" line to the next.
var set1626Inputs = filepath.Join("internal", "chart", "cmd", "ratchet", "testdata", "set1626-inputs.txt")

// pairObserver counts, over the inputs it observes, the tokens carrying terminal set set, the token positions whose
// chart set holds rule completed, and of those the ones whose token carries set: a completion before that set.
type pairObserver struct {
	tab                                     *grammar.Tables
	r                                       *recognize.Recognizer
	rule, set                               int
	inputs, carrying, completions, reaching int
}

// newPairObserver names the rule by its text ("LHS ::= RHS") and the terminal set by its members, so a renumbered
// table cannot be observed by coincidence.
func newPairObserver(t *testing.T, tab *grammar.Tables, rule string, set ...string) *pairObserver {
	t.Helper()
	o := &pairObserver{tab: tab, r: recognize.New(tab, recognize.Options{}), rule: -1, set: -1}
	for r := range tab.Rules {
		if ruleText(tab, r) == rule {
			o.rule = r
		}
	}
	for id, members := range tab.TerminalSets {
		if setText(tab, members) == strings.Join(set, " ") {
			o.set = id
		}
	}
	if o.rule < 0 || o.set < 0 {
		t.Fatalf("rule %q is %d, terminal set {%s} is %d", rule, o.rule, strings.Join(set, " "), o.set)
	}
	return o
}

// observe recognizes src and adds what it holds.
func (o *pairObserver) observe(src []byte) (completions, reaching int, err error) {
	s, _ := lex.Lex(src, o.tab.Tokens)
	c, err := recognize.Recognize(o.r, lex.Linear(s))
	if err != nil {
		return 0, 0, err
	}
	o.inputs++
	completed := func(it recognize.Item) bool { return slices.Contains(o.tab.States[it.State].Reduces, o.rule) }
	for k, tk := range s.Tokens {
		if tk.TerminalSet == o.set {
			o.carrying++
		}
		if k < len(c.Sets) && slices.ContainsFunc(c.Sets[k].Items(), completed) {
			completions++
			if tk.TerminalSet == o.set {
				reaching++
			}
		}
	}
	o.completions += completions
	o.reaching += reaching
	return completions, reaching, nil
}

// TestSet1626NeverReachedCorpus: over every in-tree corpus file and every committed constructed input, no token
// carrying set 1626 follows a completed indirob ::= scalar. Reaching it is a design §7 stop, not a number to update.
func TestSet1626NeverReachedCorpus(t *testing.T) {
	t.Chdir("../../../..")
	files, srcs, err := corpus()
	if err != nil {
		t.Fatal(err)
	}
	o := newPairObserver(t, committedTables(t), "indirob ::= scalar", "HASHBRACK", "PERLY_BRACE_OPEN")
	for _, f := range files {
		if _, _, err := o.observe(srcs[f.Path]); err != nil {
			t.Fatalf("%s: %v", f.Path, err)
		}
	}
	nfiles := o.inputs
	inputs := constructedInputs(t)
	for _, in := range inputs {
		c, _, err := o.observe(in.src)
		if err != nil {
			t.Fatalf("%s: %v", in.name, err)
		}
		if c == 0 { // an input that never completes the rule observes nothing
			t.Errorf("constructed input %q: indirob ::= scalar never completes", in.name)
		}
	}
	t.Logf("files: %d constructed inputs: %d tokens carrying set 1626: %d indirob ::= scalar completions: %d reaching set 1626: %d",
		nfiles, len(inputs), o.carrying, o.completions, o.reaching)
	if nfiles == 0 || len(inputs) < 14 || o.carrying == 0 || o.completions == 0 {
		t.Errorf("observed nothing: %d files, %d constructed inputs, %d tokens carrying set 1626, %d completions", nfiles,
			len(inputs), o.carrying, o.completions)
	}
	if o.reaching != 0 {
		t.Errorf("set 1626 reached %d times after indirob ::= scalar: a design §7 stop", o.reaching)
	}
}

// TestSet1626ObserverCountsAnOccurringPair: the observer counts a pair the corpus does reach, indirob ::= scalar
// before set 1626's neighbour {PERLY_BRACE_OPEN} ($$ref{key}, *$glob{IO}), over the corpus and in a constructed input,
// and counts nothing for a set no token carries after the rule, so its 0 for set 1626 is a count.
func TestSet1626ObserverCountsAnOccurringPair(t *testing.T) {
	t.Chdir("../../../..")
	files, srcs, err := corpus()
	if err != nil {
		t.Fatal(err)
	}
	o := newPairObserver(t, committedTables(t), "indirob ::= scalar", "PERLY_BRACE_OPEN")
	for _, f := range files {
		if _, _, err := o.observe(srcs[f.Path]); err != nil {
			t.Fatalf("%s: %v", f.Path, err)
		}
	}
	t.Logf("corpus: files: %d completions: %d reaching {PERLY_BRACE_OPEN}: %d", o.inputs, o.completions, o.reaching)
	if o.reaching == 0 {
		t.Error("the corpus's indirob ::= scalar before { counted 0")
	}
	for _, c := range []struct {
		src  string
		want bool
	}{{"my $r = {}; my $v = $$r{key};", true}, {"my $g = \\*STDOUT; my $io = *$g{IO};", true}, {"my $r = {}; my $v = $$r;", false}} {
		_, n, err := o.observe([]byte(c.src))
		t.Logf("%q: reaching {PERLY_BRACE_OPEN}: %d", c.src, n)
		if err != nil || (n > 0) != c.want {
			t.Errorf("%q: reaching %d, %v; want reached %t", c.src, n, err, c.want)
		}
	}
}

// constructedInput is one input of set1626Inputs.
type constructedInput struct {
	name string
	src  []byte
}

func constructedInputs(t *testing.T) []constructedInput {
	t.Helper()
	b, err := os.ReadFile(set1626Inputs)
	if err != nil {
		t.Fatal(err)
	}
	var ins []constructedInput
	for _, line := range strings.SplitAfter(string(b), "\n") {
		if name, ok := strings.CutPrefix(line, "### "); ok {
			ins = append(ins, constructedInput{name: strings.TrimSpace(name)})
		} else if len(ins) > 0 {
			ins[len(ins)-1].src = append(ins[len(ins)-1].src, line...)
		}
	}
	return ins
}

// ruleText is rule r as "LHS ::= RHS".
func ruleText(tab *grammar.Tables, r int) string {
	parts := []string{tab.Symbols[tab.Rules[r].LHS].Name, "::="}
	for _, x := range tab.Rules[r].RHS {
		parts = append(parts, tab.Symbols[x].Name)
	}
	return strings.Join(parts, " ")
}

// setText is a terminal set's member names, in id order.
func setText(tab *grammar.Tables, set []int) string {
	names := make([]string, len(set))
	for i, x := range set {
		names[i] = tab.Symbols[x].Name
	}
	return strings.Join(names, " ")
}
