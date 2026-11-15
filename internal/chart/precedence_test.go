// ABOUTME: Recognition ignores precedence (spec §8); the forest's constraint applies it, refereed by bison on every
// ABOUTME: kept generated sentence (testdata/precedence/), and refuses nothing perl accepts in the Guacamole corpus or std5k.pl.

package chart

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"tamarou.com/pvm/internal/chart/forest"
	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
)

// builtForest is src's forest as forest.Build makes it from the chart, before resolve: the precedence constraint and
// its bison referee read the token sequence's grammar, which scope does not decide.
// A recognition error gives no forest, as forest.Parse's refusal did.
func builtForest(tab *grammar.Tables, src []byte) *forest.Forest {
	s, _ := lex.Lex(src, tab.Tokens)
	l := lex.Linear(s)
	c, err := recognize.Recognize(recognize.New(tab, recognize.Options{}), l)
	if err != nil {
		return nil
	}
	return forest.Build(tab, c, l)
}

func TestPrecedenceIgnoredByRecognition(t *testing.T) {
	g, _, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	for _, tok := range []string{"NCEQOP", "DOTDOT", "PERLY_BRACKET_OPEN", "ADDOP", "MULOP"} {
		if !slices.Contains(g.Prec, tok) {
			t.Errorf("the committed grammar declares no precedence for %s", tok)
		}
	}
	if !slices.ContainsFunc(g.G1, func(r grammar.Rule) bool { return r.Prec != "" }) {
		t.Error("no committed rule carries a precedence token (Rule.Prec)")
	}
	// The first three perl refuses by precedence alone: <=> chained by termbinop's `termeqop %prec PREC_LOW`
	// (perly.y:1438) losing to NCEQOP, which shifts into the error rule perly.y:1484 (166 is %left);
	// 1 .. 2 .. 3 by DOTDOT's %nonassoc; and print $fh [] by indirob's PREC_LOW.
	for _, src := range []string{"$a <=> $b <=> $c;", "1 .. 2 .. 3;", "print $fh [];", "$a + $b * $c;"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at byte %d (expected %v); recognition must ignore precedence", src, v.at, v.expected)
		}
	}
}

// perlyPrecedence is the committed grammar's precedence over its tables: its levels, each rule's %prec, and the
// rules refused by precedence, which the tables leave out.
var perlyPrecedence = sync.OnceValues(func() (*forest.Precedence, error) {
	g, _, err := grammar.Committed()
	if err != nil {
		return nil, err
	}
	tab, err := committed()
	if err != nil {
		return nil, err
	}
	if len(tab.Rules) != len(g.G1) {
		return nil, fmt.Errorf("%d table rules for %d grammar rules: a sequence rule moved the rule ids", len(tab.Rules), len(g.G1))
	}
	rulePrec := make([]string, len(g.G1))
	for r, rule := range g.G1 {
		rulePrec[r] = rule.Prec
	}
	return forest.NewPrecedence(tab, g.Levels, rulePrec, g.Refused)
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

// constrainSource is the whole pipeline's verdict on src: chart/lex, recognition, the forest and the constraint.
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

// referee compares the chart with bison's committed verdict and grouping on the generated sentences.
type referee struct {
	tab      *grammar.Tables
	p        *forest.Precedence
	lexemes  map[string]lexeme
	terminal map[string]int32
	line     []int  // by rule: the perly.y alternative it cites
	carries  []bool // by rule: a %prec, or a last terminal with a level, as the referee's grouping trace reads it
}

func newReferee(t *testing.T) *referee {
	t.Helper()
	g, _, err := grammar.Committed()
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
	lexemes, err := readLexemes(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	r := &referee{tab: tab, p: p, lexemes: lexemes, terminal: map[string]int32{}}
	for id, s := range tab.Symbols {
		if s.Terminal {
			r.terminal[s.Name] = int32(id)
		}
	}
	for i, rule := range g.G1 {
		ref, _, _ := strings.Cut(strings.TrimPrefix(rule.Cite, "perly.y:"), ",")
		n, err := strconv.Atoi(ref)
		if err != nil {
			t.Fatalf("rule %d cites %q", i, rule.Cite)
		}
		last := ""
		for _, s := range slices.Backward(tab.Rules[i].RHS) {
			if tab.Symbols[s].Terminal {
				last = tab.Symbols[s].Name
				break
			}
		}
		r.line = append(r.line, n)
		r.carries = append(r.carries, rule.Prec != "" || slices.Contains(g.Prec, last))
	}
	return r
}

// judge is the chart's verdict ("accept" or "refuse") and grouping on s's rendering: forest.Parse, its derivations
// narrowed to s's own terminals (a token admits a set of terminals, and bison read one of them), then the constraint.
// The grouping is the referee's: each reduction of a rule carrying a precedence, line:from-to over terminal indexes
// (GRAMPROG is 0), children first, comma-separated, - for none; "ambiguous" where the derivations left disagree on it.
func (r *referee) judge(s precedenceSentence) (verdict, grouping string) {
	preamble, _, _, _ := renderTerms(r.lexemes, s.terms)
	skip := 0
	if preamble != "" {
		pre, _ := lex.Lex([]byte(preamble), r.tab.Tokens)
		skip = len(pre.Tokens)
	}
	var want []int32
	for _, name := range s.terms[1:] { // after GRAMPROG
		want = append(want, r.terminal[name])
	}
	f := builtForest(r.tab, []byte(s.rendering))
	if f == nil || !narrow(f, want, skip) || !f.Constrain(r.p) {
		return "refuse", "-"
	}
	trees, cut := f.Derivations(f.Root, 16)
	if cut {
		return "accept", "more than 16 derivations left, not all compared"
	}
	groupings := map[string]bool{}
	for _, tr := range trees {
		var g []string
		var walk func(*forest.Tree)
		walk = func(tr *forest.Tree) {
			for _, k := range tr.Kids {
				walk(k)
			}
			if tr.Rule >= 0 && r.carries[tr.Rule] && int(tr.From) >= skip { // bison reads no preamble
				g = append(g, fmt.Sprintf("%d:%d-%d", r.line[tr.Rule], tr.From-skip+1, tr.To-skip+1))
			}
		}
		walk(tr)
		if len(g) == 0 {
			g = []string{"-"}
		}
		groupings[strings.Join(g, ",")] = true
	}
	if len(groupings) != 1 {
		return "accept", fmt.Sprintf("ambiguous (%d groupings) %v", len(groupings), groupings)
	}
	for g := range groupings {
		grouping = g
	}
	return "accept", grouping
}

// narrow keeps the derivations of f whose leaves past the first skip tokens are want's terminals, one token each, and
// reports whether the root keeps one. A slipper ; is zero-width, and bison reads no ; that is not written.
func narrow(f *forest.Forest, want []int32, skip int) bool {
	alive := make([]bool, len(f.Nodes))
	for n, nd := range f.Nodes {
		i := int(nd.From) - skip
		alive[n] = n > 0 && nd.Alt == 0 && (i < 0 || nd.To == nd.From+1 && i < len(want) && nd.Sym == want[i])
	}
	ok := func(a forest.Alt) bool { return (a.Left == 0 || alive[a.Left]) && (a.Right == 0 || alive[a.Right]) }
	for changed := true; changed; { // to a fixed point: a unit rule's cycle keeps a node only through another way
		changed = false
		for n := range f.Nodes {
			for a := f.Nodes[n].Alt; a != 0 && !alive[n]; a = f.Alts[a].Next {
				if ok(f.Alts[a]) {
					alive[n], changed = true, true
				}
			}
		}
	}
	for n := range f.Nodes {
		if !alive[n] {
			continue // unreachable from a kept alternative
		}
		for a := &f.Nodes[n].Alt; *a != 0; {
			if ok(f.Alts[*a]) {
				a = &f.Alts[*a].Next
			} else {
				*a = f.Alts[*a].Next
			}
		}
	}
	return alive[f.Root]
}

// disagreement is how the chart differs from bison's verdict and grouping on s, or "" where it agrees. Two
// refusals are not compared by grouping, since bison's stops at its first error.
func (r *referee) disagreement(s precedenceSentence) string {
	v, g := r.judge(s)
	switch {
	case v == "refuse" && s.verdict == "accept":
		return fmt.Sprintf("%s %q: the chart refuses, bison accepts %s", s.id, s.rendering, s.grouping)
	case v == "accept" && s.verdict != "accept":
		return fmt.Sprintf("%s %q: the chart accepts %s, bison %s", s.id, s.rendering, g, s.verdict)
	case v == "accept" && g != s.grouping:
		return fmt.Sprintf("%s %q: the chart groups %s, bison %s", s.id, s.rendering, g, s.grouping)
	}
	return ""
}

// TestPrecedenceReferee requires bison's verdict and grouping on every kept generated sentence.
func TestPrecedenceReferee(t *testing.T) {
	r := newReferee(t)
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

// TestPrecedenceRefereeDetectsMismatch requires the comparison to report a sentence with its grouping swapped or its
// verdict flipped, so TestPrecedenceReferee cannot pass by comparing nothing.
func TestPrecedenceRefereeDetectsMismatch(t *testing.T) {
	r := newReferee(t)
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
