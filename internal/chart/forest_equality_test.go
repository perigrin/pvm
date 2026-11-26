// ABOUTME: The forest-equality differential over every waypoint2 row and every referee sentence, with V1's side from forest/equality.go.
// ABOUTME: The referee's V1 path is Build, narrow and Constrain with no LibraryOperators; its comparison drops zero-width slipper ; leaves.

package chart

import (
	"fmt"
	"slices"
	"testing"

	"tamarou.com/pvm/internal/chart/forest"
	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/recognize"
)

// waypoint2Sources is every source the waypoint2 tables hold, scope rows after their preamble, with the verdicts
// dropped: the differential compares forests, not perl's verdicts.
func waypoint2Sources() []string {
	var srcs []string
	for _, c := range waypoint2Refusals {
		srcs = append(srcs, c.src)
	}
	for _, c := range waypoint2ScopeRefusals {
		srcs = append(srcs, c.pre+c.src)
	}
	return srcs
}

// TestForestEqualityWaypoint2: on every waypoint2 row, refused rows included, the candidate's forest under the
// empty scope equals V1's section 5 pipeline. It logs how many rows it compared of how many it loaded and how many
// V1 refused, and requires every row compared and some refused, so a skipped row or an empty table cannot pass.
func TestForestEqualityWaypoint2(t *testing.T) {
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	srcs := waypoint2Sources()
	compared, refused := 0, 0
	var diffs []string
	for _, src := range srcs {
		cand, _ := forest.CandidateForest(tab, []byte(src), forest.Scope{})
		v1, _ := forest.V1Forest(tab, []byte(src), forest.Scope{})
		compared++
		if v1 == nil {
			refused++
		}
		if d := forest.ForestDiff(cand, v1); d != "" {
			diffs = append(diffs, fmt.Sprintf("%q: %s", src, d))
		}
	}
	for _, d := range diffs[:min(10, len(diffs))] {
		t.Error(d)
	}
	t.Logf("waypoint2: compared %d of %d rows, refused: %d", compared, len(srcs), refused)
	if len(diffs) > 0 {
		t.Errorf("%d of %d forests differ from V1's", len(diffs), compared)
	}
	if compared == 0 || compared != len(srcs) || refused == 0 {
		t.Errorf("compared %d of %d rows, V1 refused %d; want every row compared, and some refused", compared, len(srcs), refused)
	}
}

// refereeV1Forest is V1's forest on the referee's path: V1Recognizer's chart, Build, keep narrows the forest,
// then Constrain under p, nil where any refuses. It has no LibraryOperators: the referee judges bison's token
// sequence, which scope does not decide (F4-2). Its recognizer is V1's, never recognize.New or the referee's own.
func refereeV1Forest(tab *grammar.Tables, p *forest.Precedence, src string, keep func(*forest.Forest) bool) *forest.Forest {
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	l := lex.Linear(s)
	c, err := recognize.Recognize(forest.V1Recognizer(tab), l)
	if err != nil || !c.Accepted() {
		return nil
	}
	f := forest.Build(tab, c, l)
	if !keep(f) || !f.Constrain(p) {
		return nil
	}
	return f
}

// refereeCandidateForest is the forest the referee judges: narrowed to the sentence's terminals, then constrained.
// A later issue may point it at the recognizer that judges precedence itself.
var refereeCandidateForest = func(r *referee, s precedenceSentence) *forest.Forest {
	f, _ := r.narrowed(s)
	if f == nil || !f.Constrain(r.p) {
		return nil
	}
	return f
}

// dropSlipperLeaves is f with the zero-width leaves of terminal semi removed, and their places in every
// alternative's children. A leaf of another terminal, and a semi leaf one token wide, stay.
func dropSlipperLeaves(f *forest.Forest, semi int32) *forest.Forest {
	if f == nil {
		return nil
	}
	slipper := func(n int32) bool {
		nd := f.Nodes[n]
		return n != 0 && nd.Alt == 0 && nd.Sym == semi && nd.From == nd.To
	}
	g := *f
	g.Alts = slices.Clone(f.Alts)
	for i := range g.Alts {
		if slipper(g.Alts[i].Left) {
			g.Alts[i].Left = 0
		}
		if slipper(g.Alts[i].Right) {
			g.Alts[i].Right = 0
		}
	}
	return &g
}

// refereeDiff is where the candidate's forest differs from V1's, with zero-width slipper leaves ignored on both.
func refereeDiff(cand, v1 *forest.Forest, semi int32) string {
	return forest.ForestDiff(dropSlipperLeaves(cand, semi), dropSlipperLeaves(v1, semi))
}

// refereeDiffs compares the referee's candidate forest of each kept sentence with V1's path: how many it
// compared, how many V1 refused, and each difference.
func refereeDiffs(r *referee, sentences []precedenceSentence) (compared, refused int, diffs []string) {
	semi := r.terminal["PERLY_SEMICOLON"]
	for _, s := range sentences {
		if s.rendering == "-" {
			continue
		}
		want, skip := r.target(s)
		v1 := refereeV1Forest(r.tab, r.p, s.rendering, func(f *forest.Forest) bool { return narrow(f, want, skip) })
		compared++
		if v1 == nil {
			refused++
		}
		if d := refereeDiff(refereeCandidateForest(r, s), v1, semi); d != "" {
			diffs = append(diffs, fmt.Sprintf("%s %q: %s", s.id, s.rendering, d))
		}
	}
	return compared, refused, diffs
}

// keptSentences is how many of sentences carry a rendering.
func keptSentences(sentences []precedenceSentence) (kept int) {
	for _, s := range sentences {
		if s.rendering != "-" {
			kept++
		}
	}
	return kept
}

// TestForestEqualityRefereeSentences: on every kept referee sentence the referee's candidate forest equals V1's on
// the referee's own path. It logs how many it compared of how many it loaded and how many V1 refused, and requires
// every one compared and some refused.
func TestForestEqualityRefereeSentences(t *testing.T) {
	r := committedReferee(t)
	sentences, err := readSentences(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	loaded := keptSentences(sentences)
	compared, refused, diffs := refereeDiffs(r, sentences)
	for _, d := range diffs[:min(10, len(diffs))] {
		t.Error(d)
	}
	t.Logf("referee: compared %d of %d sentences, refused: %d", compared, loaded, refused)
	if len(diffs) > 0 {
		t.Errorf("%d of %d forests differ from V1's", len(diffs), compared)
	}
	if compared == 0 || compared != loaded || refused == 0 {
		t.Errorf("compared %d of %d sentences, V1 refused %d; want every one compared, and some refused", compared, loaded, refused)
	}
}

// TestForestEqualityRefereeReportsADifferingCandidate: with a candidate that refuses everything, the referee loop
// reports every sentence V1 accepts, so it does not compare V1 with itself.
func TestForestEqualityRefereeReportsADifferingCandidate(t *testing.T) {
	cand0 := refereeCandidateForest
	t.Cleanup(func() { refereeCandidateForest = cand0 })
	refereeCandidateForest = func(*referee, precedenceSentence) *forest.Forest { return nil }
	sentences, err := readSentences(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	compared, refused, diffs := refereeDiffs(committedReferee(t), sentences)
	if accepted := compared - refused; accepted == 0 || len(diffs) != accepted {
		t.Errorf("V1 accepts %d of %d, %d reported; want each accepted sentence reported", accepted, compared, len(diffs))
	}
}

// TestForestEqualityRefereePathOmitsLibraryOperators: on a sentence where LibraryOperators drops an alternative,
// the referee's V1 forest keeps it, so it differs from section 5's pipeline's and equals Build then Constrain.
func TestForestEqualityRefereePathOmitsLibraryOperators(t *testing.T) {
	const src = "use My::Ops; 1 plus 2;"
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	p, err := forest.ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	all := func(*forest.Forest) bool { return true }
	ref := refereeV1Forest(tab, p, src, all)
	s5, refusal := forest.V1Forest(tab, []byte(src), libopsScope(t, src))
	if ref == nil || s5 == nil {
		t.Fatalf("the referee path gives %v, section 5's pipeline %v (%+v); want both to accept", ref != nil, s5 != nil, refusal)
	}
	if forest.ForestDiff(ref, s5) == "" {
		t.Errorf("the referee's V1 forest equals section 5's pipeline's; LibraryOperators dropped nothing on %q", src)
	}
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	l := lex.Linear(s)
	c, _ := recognize.Recognize(forest.V1Recognizer(tab), l)
	plain := forest.Build(tab, c, l)
	if !plain.Constrain(p) {
		t.Fatal("Build then Constrain refuses the sentence")
	}
	if d := forest.ForestDiff(ref, plain); d != "" {
		t.Errorf("the referee's V1 forest differs from Build then Constrain: %s", d)
	}
}

// TestForestEqualityRefereeIgnoresOnlyZeroWidthSlipperLeaves pins the referee comparison's one allowance: a
// zero-width ; leaf is dropped from both sides, and a ; one token wide or a zero-width leaf of another terminal is not.
func TestForestEqualityRefereeIgnoresOnlyZeroWidthSlipperLeaves(t *testing.T) {
	r := committedReferee(t)
	semi, word := r.terminal["PERLY_SEMICOLON"], r.terminal["BAREWORD"]
	// pair is the forest of a root over a one-token leaf and then a leaf of sym spanning from to.
	pair := func(sym, from, to int32) *forest.Forest {
		f := &forest.Forest{Root: 1, Nodes: make([]forest.Node, 3, 4), Alts: make([]forest.Alt, 2)}
		f.Nodes[1] = forest.Node{Sym: 0, Rule: -1, From: 0, To: 1, Alt: 1}
		f.Nodes[2] = forest.Node{Sym: word, Rule: -1, From: 0, To: 1}
		right := int32(0)
		if sym >= 0 {
			f.Nodes = append(f.Nodes, forest.Node{Sym: sym, Rule: -1, From: from, To: to})
			right = 3
		}
		f.Alts[1] = forest.Alt{Rule: 0, Left: 2, Right: right}
		return f
	}
	for _, c := range []struct {
		name      string
		with      *forest.Forest
		wantEqual bool
	}{
		{"zero-width-semicolon-equal", pair(semi, 1, 1), true},
		{"one-byte-semicolon-differs", pair(semi, 1, 2), false},
		{"zero-width-other-terminal-differs", pair(word, 1, 1), false},
	} {
		t.Run(c.name, func(t *testing.T) {
			without := pair(-1, 0, 0)
			for _, sides := range [][2]*forest.Forest{{c.with, without}, {without, c.with}} {
				if d := refereeDiff(sides[0], sides[1], semi); (d == "") != c.wantEqual {
					t.Errorf("equal=%v, want %v (%q)", d == "", c.wantEqual, d)
				}
			}
			if d := forest.ForestDiff(c.with, without); d == "" {
				t.Errorf("the plain ForestDiff finds the pair equal, so the pair does not differ before the filter")
			}
		})
	}
}
