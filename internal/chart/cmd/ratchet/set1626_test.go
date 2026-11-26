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

// pairCount is what a pairObserver saw: the tokens carrying its terminal set, the token positions whose chart set
// holds its rule completed, and of those the ones whose token carries the set, a completion right before it. gap is the
// fewest tokens from a completion to a later token carrying the set, -1 where none follows one.
type pairCount struct{ carrying, completions, reaching, gap int }

// pairObserver adds up, over the inputs it observes, the pairCount of rule and terminal set set.
type pairObserver struct {
	tab       *grammar.Tables
	r         *recognize.Recognizer
	rule, set int
	inputs    int
	pairCount
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

// observe recognizes src, adds what it holds and returns it.
func (o *pairObserver) observe(src []byte) (pairCount, error) {
	n := pairCount{gap: -1}
	s, _ := lex.Lex(src, o.tab.Tokens)
	c, err := recognize.Recognize(o.r, lex.Linear(s))
	if err != nil {
		return n, err
	}
	o.inputs++
	completed := func(it recognize.Item) bool { return slices.Contains(o.tab.States[it.State].Reduces, o.rule) }
	last := -1 // the latest position holding a completion
	for k, tk := range s.Tokens {
		if k < len(c.Sets) && slices.ContainsFunc(c.Sets[k].Items(), completed) {
			n.completions++
			last = k
		}
		if tk.TerminalSet != o.set {
			continue
		}
		n.carrying++
		if last == k {
			n.reaching++
		}
		if last >= 0 && (n.gap < 0 || k-last < n.gap) {
			n.gap = k - last
		}
	}
	o.carrying += n.carrying
	o.completions += n.completions
	o.reaching += n.reaching
	return n, nil
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
		if _, err := o.observe(srcs[f.Path]); err != nil {
			t.Fatalf("%s: %v", f.Path, err)
		}
	}
	nfiles := o.inputs
	inputs := constructedInputs(t)
	for _, in := range inputs {
		n, err := o.observe(in.src)
		if err != nil {
			t.Fatalf("%s: %v", in.name, err)
		}
		t.Logf("constructed input %q: set-1626 tokens: %d completions: %d nearest set-1626 token after a completion: %d",
			in.name, n.carrying, n.completions, n.gap)
		// An input probes set 1626 where it carries the set right after a completion of the rule, or as near as the
		// lexer allows: one or two tokens on (the fixture's header says why).
		if n.gap < 0 || n.gap > 2 {
			t.Errorf("constructed input %q: %d completions of indirob ::= scalar, %d set-1626 tokens, nearest %d after one; want 0 to 2",
				in.name, n.completions, n.carrying, n.gap)
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
		if _, err := o.observe(srcs[f.Path]); err != nil {
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
		n, err := o.observe([]byte(c.src))
		t.Logf("%q: reaching {PERLY_BRACE_OPEN}: %d", c.src, n.reaching)
		if err != nil || (n.reaching > 0) != c.want {
			t.Errorf("%q: reaching %d, %v; want reached %t", c.src, n.reaching, err, c.want)
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
