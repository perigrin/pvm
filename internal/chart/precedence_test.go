// ABOUTME: Recognition ignores precedence (spec §8); the forest's constraint applies it, refereed by bison on every
// ABOUTME: kept generated sentence (testdata/precedence/), and refuses nothing perl accepts in the Guacamole corpus or std5k.pl.

package chart

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"tamarou.com/pvm/internal/chart/forest"
	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/prectab"
	"tamarou.com/pvm/internal/chart/recognize"
)

// perlyPrecedence is the committed grammar's precedence over its tables.
var perlyPrecedence = sync.OnceValues(func() (*prectab.Precedence, error) {
	g, _, err := grammar.Committed()
	if err != nil {
		return nil, err
	}
	tab, err := committed()
	if err != nil {
		return nil, err
	}
	return newPrecedence(g, tab)
})

// TestPrecedenceRefusedRulesLeftRecursive pins NewPrecedence's shortcut: each rule refused by precedence begins with
// its own left-hand side, which derives no empty string, so leaving it out of the tables' FIRST sets and nullability
// changes neither (a nullable A in A: A X would add FIRST(X) to FIRST(A)).
func TestPrecedenceRefusedRulesLeftRecursive(t *testing.T) {
	g, _, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	nullable := map[string]bool{}
	for _, s := range tab.Symbols {
		nullable[s.Name] = s.Nullable
	}
	if len(g.Refused) != 4 {
		t.Errorf("%d rules refused by precedence, want perly.y's four comparison chains", len(g.Refused))
	}
	for _, r := range g.Refused {
		if r.RHS[0].Text != r.LHS {
			t.Errorf("%s %s: does not begin with %s", r.Cite, r.LHS, r.LHS)
		}
		if nullable[r.LHS] {
			t.Errorf("%s %s: derives the empty string", r.Cite, r.LHS)
		}
	}
}

// constrainSource is the verdict on src of chart/lex, recognition, Build's forest before resolve and the constraint.
func constrainSource(t *testing.T, src []byte) bool {
	t.Helper()
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	p, err := perlyPrecedence()
	if err != nil {
		t.Fatal(err)
	}
	f := builtForest(tab, src)
	return f != nil && f.Constrain(p)
}

// committedReferee is the committed grammar's referee over testdata/precedence.
func committedReferee(t *testing.T) *referee {
	t.Helper()
	g, _, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	r, err := newReferee(g, tab, precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// TestPrecedenceReferee requires bison's verdict and grouping on every kept generated sentence.
func TestPrecedenceReferee(t *testing.T) {
	r := committedReferee(t)
	sentences, err := readSentences(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	kept, bad := 0, 0
	for _, s := range sentences {
		if s.rendering == "-" {
			continue
		}
		kept++
		if d := r.disagreement(s); d != "" {
			if bad++; bad <= 20 {
				t.Error(d)
			}
		}
	}
	covered, total, err := checkPrecedenceFixtures(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	if kept == 0 || bad > 0 {
		t.Fatalf("%d of %d kept sentences differ from bison", bad, kept)
	}
	t.Logf("%d kept sentences agree with bison; they cover %d of %d bison resolutions", kept, covered, total)
}

// TestCountPrecedenceCovered: CountPrecedence's Covered is the resolutions a kept sentence covers, which on the
// committed fixture is fewer than bison's total, so a count of the total cannot pass.
func TestCountPrecedenceCovered(t *testing.T) {
	g, _, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	c, err := CountPrecedence(g, tab, precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	covered, total, err := checkPrecedenceFixtures(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	if covered >= total {
		t.Fatalf("the fixture covers %d of %d; want some uncovered, or this proves nothing", covered, total)
	}
	if c.Covered != covered {
		t.Errorf("CountPrecedence covers %d, the fixture's coverage is %d of %d", c.Covered, covered, total)
	}
}

// TestPrecedenceRefereeDetectsMismatch requires the comparison to report a sentence with its grouping swapped or its
// verdict flipped, so TestPrecedenceReferee cannot pass by comparing nothing.
func TestPrecedenceRefereeDetectsMismatch(t *testing.T) {
	r := committedReferee(t)
	sentences, err := readSentences(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	swapped, accepted, refused := false, false, false
	for _, s := range sentences {
		if s.rendering == "-" || r.disagreement(s) != "" {
			continue
		}
		if g := strings.Split(s.grouping, ","); s.verdict == "accept" && len(g) > 1 && g[0] != g[1] && !swapped {
			g[0], g[1] = g[1], g[0]
			m := s
			m.grouping = strings.Join(g, ",")
			swapped = true
			if !strings.Contains(r.disagreement(m), "the chart groups") {
				t.Errorf("%s with its grouping swapped (%s) is not reported", s.id, m.grouping)
			}
		}
		m := s
		switch {
		case s.verdict == "accept" && !accepted:
			accepted, m.verdict = true, "refuse@1"
			if !strings.Contains(r.disagreement(m), "the chart accepts") {
				t.Errorf("%s with bison's verdict flipped to refuse is not reported", s.id)
			}
		case s.verdict != "accept" && !refused:
			refused, m.verdict = true, "accept"
			if !strings.Contains(r.disagreement(m), "the chart refuses") {
				t.Errorf("%s with bison's verdict flipped to accept is not reported", s.id)
			}
		}
	}
	if !swapped || !accepted || !refused {
		t.Fatalf("no agreeing sentence to mutate: grouping swapped %v, accepted %v, refused %v", swapped, accepted, refused)
	}
}

// TestPrecedenceKeepsCorpusAccepted requires the constraint to refuse no in-tree file (the Guacamole snippets,
// std5k.pl and edit5k.pl) that perl's label accepts and recognition accepts.
func TestPrecedenceKeepsCorpusAccepted(t *testing.T) {
	labels, err := ReadLabels(filepath.Join(labelsDir, "in-tree.txt"))
	if err != nil {
		t.Fatal(err)
	}
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	p, err := perlyPrecedence()
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, l := range labels.Labels {
		if l.Refused() {
			continue
		}
		src, err := os.ReadFile(filepath.Join("..", "..", l.Path))
		if err != nil {
			t.Fatal(err)
		}
		f := builtForest(tab, src)
		if f == nil {
			continue
		}
		n++
		if !f.Constrain(p) {
			t.Errorf("%s: perl and recognition accept it, the precedence constraint refuses it", l.Path)
		}
	}
	if n == 0 {
		t.Fatal("no file perl and recognition accept")
	}
	t.Logf("the constraint keeps all %d files perl and recognition accept", n)
}

// TestPrecedenceRefereeMatchesReference requires the referee's judged forest to give ConstrainReference's verdict and
// kept alternatives over the no-precedence narrowed forest (referee.narrowed) on every kept sentence, zero-width
// slipper ; leaves dropped from both. Where the reference refuses, the judged chart dies and has no forest, so only
// the verdicts are compared (design §5).
func TestPrecedenceRefereeMatchesReference(t *testing.T) {
	r := committedReferee(t)
	sentences, err := readSentences(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	semi := r.terminal["PERLY_SEMICOLON"]
	compared, verdictOnly := 0, 0
	for _, s := range sentences {
		if s.rendering == "-" {
			continue
		}
		ref, _ := r.narrowed(s)
		if ref == nil {
			t.Errorf("%s %q: the no-precedence narrowed chart refuses, so the reference has no forest", s.id, s.rendering)
			continue
		}
		compared++
		judged, _, _ := r.judged(s)
		if !ref.ConstrainReference(r.p) {
			verdictOnly++
			if judged != nil {
				t.Errorf("%s %q: ConstrainReference refuses, the judged chart accepts", s.id, s.rendering)
			}
			continue
		}
		if d := refereeDiff(judged, ref, semi); d != "" {
			t.Errorf("%s %q: the judged forest differs from ConstrainReference's: %s", s.id, s.rendering, d)
		}
	}
	t.Logf("referee: compared %d of %d", compared, keptSentences(sentences))
	t.Logf("%s verdict only: %d", t.Name(), verdictOnly)
	if compared == 0 || verdictOnly == 0 {
		t.Errorf("compared %d, %d by verdict only; want some of each", compared, verdictOnly)
	}
}

// TestPrecedenceRefereeJudgesRecognizer: on every kept sentence the referee's verdict and grouping match bison's, and
// judging them all enters Constrain no time, counted by Constrain's own hook, so the verdicts are the judged
// recognizer's over the narrowed lattice and not the constraint's over a built forest.
func TestPrecedenceRefereeJudgesRecognizer(t *testing.T) {
	r := committedReferee(t)
	sentences, err := readSentences(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	compared, refused, bad, n0 := 0, 0, 0, forest.ConstrainEntries()
	for _, s := range sentences {
		if s.rendering == "-" {
			continue
		}
		compared++
		if s.verdict != "accept" {
			refused++
		}
		if d := r.disagreement(s); d != "" {
			if bad++; bad <= 10 {
				t.Error(d)
			}
		}
	}
	entries := forest.ConstrainEntries() - n0
	t.Logf("referee: compared %d of %d, bison refuses %d, disagree: %d, constrain entries: %d", compared, keptSentences(sentences), refused, bad, entries)
	if compared == 0 || refused == 0 || bad > 0 || entries != 0 {
		t.Errorf("compared %d, bison refused %d, %d disagree, Constrain entered %d times; want some of each compared, none disagreeing, no entry", compared, refused, bad, entries)
	}
}

// TestPrecedenceRefereeAllSentencesAgree: precedence-disagree, as the ratchet reads it from CountPrecedence, is 0 over
// every kept sentence, and it logs how many it compared of how many were loaded.
func TestPrecedenceRefereeAllSentencesAgree(t *testing.T) {
	g, _, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	sentences, err := readSentences(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	c, err := CountPrecedence(g, tab, precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	kept := keptSentences(sentences)
	for _, d := range c.Disagree[:min(10, len(c.Disagree))] {
		t.Error(d)
	}
	t.Logf("referee: compared %d of %d", c.Compared, kept)
	t.Logf("disagree: %d", len(c.Disagree))
	if kept == 0 || c.Compared != kept || len(c.Disagree) != 0 {
		t.Errorf("compared %d of %d kept sentences, %d disagree; want every one compared, none disagreeing", c.Compared, kept, len(c.Disagree))
	}
}

// TestRefereeSingletonSetIdsOutsideCommittedTables: narrowing every kept sentence's lattice to its own terminals
// gives the 19 singleton sets the committed tables lack ids past recognize.NarrowedBase, and the tables still hold
// their 1880 sets, each as it was.
func TestRefereeSingletonSetIdsOutsideCommittedTables(t *testing.T) {
	r := committedReferee(t)
	sentences, err := readSentences(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	before := make([][]int, len(r.tab.TerminalSets))
	for i, set := range r.tab.TerminalSets {
		before[i] = slices.Clone(set)
	}
	outside := map[int]bool{}
	for _, s := range sentences {
		if s.rendering == "-" {
			continue
		}
		l, skip := r.lattice(s)
		for _, e := range l.Edges[skip:] {
			if e.Token.TerminalSet >= recognize.NarrowedBase {
				outside[e.Token.TerminalSet] = true
			}
		}
	}
	t.Logf("singleton sets outside the committed tables: %d", len(outside))
	if len(outside) != 19 {
		t.Errorf("%d singleton sets outside the committed tables, want 19", len(outside))
	}
	if len(r.tab.TerminalSets) != 1880 || !slices.EqualFunc(r.tab.TerminalSets, before, slices.Equal) {
		t.Errorf("the committed tables hold %d sets after narrowing, or one changed; want the 1880 as they were", len(r.tab.TerminalSets))
	}
}

// TestPrecedenceRefereeKeepsSlipperLeaves (F8-P1): lattice narrowing keeps the slipper's zero-width ; leaves, which the
// forest-side narrow dropped. Over every kept sentence it counts the charts the slipper inserted a ; into and the
// forests that keep such a leaf, and requires the verdict and grouping with the leaves dropped to be the referee's,
// and no sentence ambiguous or cut.
func TestPrecedenceRefereeKeepsSlipperLeaves(t *testing.T) {
	r := committedReferee(t)
	sentences, err := readSentences(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	semi := r.terminal["PERLY_SEMICOLON"]
	kept, inserted, leaves, moved, ambiguous, cut := 0, 0, 0, 0, 0, 0
	for _, s := range sentences {
		if s.rendering == "-" {
			continue
		}
		kept++
		f, c, skip := r.judged(s)
		if len(c.Inserted()) > 0 {
			inserted++
		}
		if f != nil && slices.ContainsFunc(f.Nodes[1:], func(n forest.Node) bool { return n.Alt == 0 && n.Sym == semi && n.From == n.To }) {
			leaves++
		}
		v, g := r.verdict(f, skip)
		if dv, dg := r.verdict(dropSlipperLeaves(f, semi), skip); dv != v || dg != g {
			if moved++; moved <= 10 {
				t.Errorf("%s %q: %s %s with the slipper's leaves, %s %s without", s.id, s.rendering, v, g, dv, dg)
			}
		}
		switch {
		case strings.HasPrefix(g, "ambiguous"):
			ambiguous++
		case strings.HasPrefix(g, "more than"):
			cut++
		}
	}
	t.Logf("slipper ; insertions: %d of %d; forests keeping a zero-width ; leaf: %d; moved: %d; ambiguous: %d; cut: %d", inserted, kept, leaves, moved, ambiguous, cut)
	if inserted == 0 || leaves == 0 || moved > 0 || ambiguous > 0 || cut > 0 {
		t.Errorf("want some insertions and kept leaves, and nothing moved, ambiguous or cut")
	}
}
