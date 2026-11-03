// ABOUTME: Tests the superset population: membership, lexer-list exclusion, missing labels and corpora, RD's verdict.
// ABOUTME: Also the disagreement keys and lex-overhead's file set, over stubbed verdicts and a fake clock.

package main

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/parse"
)

func paths(fs []File) []string {
	var out []string
	for _, f := range fs {
		out = append(out, f.Path)
	}
	return out
}

func TestSupersetPopulation(t *testing.T) {
	m := map[string]vfile{
		"both.pl":     {rd: true, perl: accept()},
		"rdonly.pl":   {rd: true, perl: refuse("grammar")},
		"perlonly.pl": {rd: false, perl: accept()},
		"neither.pl":  {rd: false, perl: refuse("grammar")},
		"listed.pl":   {rd: true, perl: accept()},
		"alsolist.pl": {rd: true, perl: accept()},
	}
	pop, err := supersetPopulation(vfiles(m), vinputs(m, "listed.pl", "alsolist.pl"))
	if err != nil || !slices.Equal(paths(pop), []string{"both.pl"}) {
		t.Errorf("population %v, %v; want only both.pl", paths(pop), err)
	}

	m["nolabel.pl"] = vfile{rd: true}
	if _, err := supersetPopulation(vfiles(m), vinputs(m)); err == nil || !strings.Contains(err.Error(), "nolabel.pl") {
		t.Errorf("an unlabelled file: err = %v, want one naming nolabel.pl", err)
	}

	// An absent corpus directory is an error naming it, never a smaller corpus.
	t.Chdir("../../../..")
	t.Setenv("T1_CORPUS", "/nonexistent/t1")
	if _, _, err := corpus(); err == nil || !strings.Contains(err.Error(), "/nonexistent/t1") {
		t.Errorf("an absent T1 corpus: err = %v, want one naming /nonexistent/t1", err)
	}
}

func TestRDAcceptNeedsNoRefusedNode(t *testing.T) {
	leaf := func(k parse.Kind) *parse.Node { return &parse.Node{Kind: k} }
	deep := &parse.Node{Kind: parse.Trivia, Children: []*parse.Node{{Kind: parse.Block, Children: []*parse.Node{leaf(parse.Unknown)}}}}
	clean := &parse.Node{Kind: parse.Trivia, Children: []*parse.Node{{Kind: parse.Block, Children: []*parse.Node{leaf(parse.Block)}}}}
	for _, c := range []struct {
		name string
		tree *parse.Node
		want bool
	}{{"one deep refused leaf", deep, false}, {"no refused node", clean, true}} {
		got, err := rdAccepts("x.pl", func() *parse.Node { return c.tree })
		if err != nil || got != c.want {
			t.Errorf("%s: accepted %v, %v; want %v", c.name, got, err, c.want)
		}
	}
}

func TestRDPanicIsAnError(t *testing.T) {
	_, err := rdAccepts("boom.pl", func() *parse.Node { panic("bang") })
	if err == nil || !strings.Contains(err.Error(), "boom.pl") {
		t.Fatalf("a panic: err = %v, want one naming boom.pl", err)
	}
	if ok, err := rdAccepts("fine.pl", func() *parse.Node { return parse.Parse([]byte("$x = 1;")) }); err != nil || !ok {
		t.Errorf("a file RD parses: accepted %v, %v; want true, nil", ok, err)
	}
	m := map[string]vfile{"a.pl": {perl: accept()}}
	in := vinputs(m)
	in.RD = func(p string) (bool, error) { return rdAccepts(p, func() *parse.Node { panic("bang") }) }
	if _, err := supersetPopulation(vfiles(m), in); err == nil || !strings.Contains(err.Error(), "a.pl") {
		t.Errorf("the population on a panic: err = %v, want one naming a.pl", err)
	}
	if _, err := measureVerdicts(vfiles(m), in, map[string]string{}); err == nil || !strings.Contains(err.Error(), "a.pl") {
		t.Errorf("the verdict keys on a panic: err = %v, want one naming a.pl", err)
	}
}

func TestPopulationDisagreementsReported(t *testing.T) {
	m := map[string]vfile{
		"perlonly.pl": {rd: false, chart: true, perl: accept()},         // perl accepts, RD refuses
		"rdonly.pl":   {rd: true, chart: true, perl: refuse("grammar")}, // RD accepts, perl refuses
		"both.pl":     {rd: true, chart: true, perl: accept()},
	}
	got, _ := run(t, m)
	if got["superset-population"] != "1" || got["rd-accepted-perl-refused"] != "1" {
		t.Errorf("population %q, rd-accepted-perl-refused %q; want 1 and 1", got["superset-population"], got["rd-accepted-perl-refused"])
	}
	pop, err := supersetPopulation(vfiles(m), vinputs(m))
	if err != nil || !slices.Equal(paths(pop), []string{"both.pl"}) {
		t.Errorf("population %v, %v; want only both.pl", paths(pop), err)
	}
}

func TestLexOverheadFileSet(t *testing.T) {
	var now time.Duration
	std5k, listed := filepath.Join("d", "std5k.pl"), "listed.pl"
	cost := map[string]time.Duration{"a.pl": 10, "b.pl": 10, "heavy.pl": 1000, std5k: 10, "edit5k.pl": 10, "outside.pl": 10, listed: 10}
	rd := func() Runner { return func(p string) { now += cost[p] } }
	p := Protocol{Now: func() time.Duration { return now }, GC: func() {}, Stderr: &strings.Builder{}}
	m := map[string]vfile{
		"a.pl": {rd: true, perl: accept()}, "b.pl": {rd: true, perl: accept()}, "heavy.pl": {rd: true, perl: accept()},
		std5k: {rd: true, perl: accept()}, "edit5k.pl": {rd: true, perl: accept()},
		"outside.pl": {rd: true, perl: refuse("grammar")}, // not in the population
		listed:       {rd: true, perl: accept()},          // lexer-listed: not in the population
	}
	got, err := lexOverheadFiles(p, rd, vfiles(m), vinputs(m, listed))
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"a.pl", "b.pl"}; !slices.Equal(got, want) {
		t.Errorf("lex-overhead times %v, want %v (outside.pl, listed.pl, std5k.pl, edit5k.pl and the 10%% file left out)", got, want)
	}
}
