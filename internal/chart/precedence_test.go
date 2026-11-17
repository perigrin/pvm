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
)

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

// perlyPrecedence is the committed grammar's precedence over its tables.
var perlyPrecedence = sync.OnceValues(func() (*forest.Precedence, error) {
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
