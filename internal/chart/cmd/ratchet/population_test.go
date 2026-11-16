// ABOUTME: Tests the superset population: membership, lexer-list exclusion, missing labels and corpora, RD's verdict.
// ABOUTME: Also the disagreement keys and lex-overhead-aggregate's file set, over stubbed verdicts and a fake clock.

package main

import (
	"math"
	"os"
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
	if _, _, err := corpus(); err == nil || !strings.Contains(err.Error(), filepath.FromSlash("/nonexistent/t1")) {
		t.Errorf("an absent T1 corpus: err = %v, want one naming %s", err, filepath.FromSlash("/nonexistent/t1"))
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
		"rdonly2.pl":  {rd: true, chart: true, perl: refuse("scope")},   // a second one, so the two sides count differently
		"both.pl":     {rd: true, chart: true, perl: accept()},
		"neither.pl":  {rd: false, chart: true, perl: refuse("grammar")}, // both refuse: counted nowhere
	}
	got, _ := run(t, m)
	if got["superset-population"] != "1" || got["rd-accepted-perl-refused"] != "2" {
		t.Errorf("population %q, rd-accepted-perl-refused %q; want 1 and 2", got["superset-population"], got["rd-accepted-perl-refused"])
	}
	pop, err := supersetPopulation(vfiles(m), vinputs(m))
	if err != nil || !slices.Equal(paths(pop), []string{"both.pl"}) {
		t.Errorf("population %v, %v; want only both.pl", paths(pop), err)
	}
}

// TestLexOverheadFileSet: lex-overhead-aggregate times exactly the aggregate of the
// population. The fake chart and reference sides record every path they run.
func TestLexOverheadFileSet(t *testing.T) {
	var now time.Duration
	std5k, listed := filepath.Join("d", "std5k.pl"), "listed.pl"
	// outside.pl is huge, so a threshold taken over all files (not the
	// population) would keep heavy.pl and mid.pl, which exceed 10% of the population.
	cost := map[string]time.Duration{"a.pl": 10, "b.pl": 10, "heavy.pl": 1000, "mid.pl": 150, std5k: 10, "edit5k.pl": 10, "outside.pl": 10000, listed: 10}
	rd := func() Runner { return func(p string) { now += cost[p] } }
	seen := map[string]bool{}
	side := func(factor time.Duration) Side {
		return func() Runner { return func(p string) { seen[p] = true; now += cost[p] * factor / 10 } }
	}
	p := Protocol{Chart: side(11), Ref: side(10), Now: func() time.Duration { return now }, GC: func() {}, Stderr: &strings.Builder{}}
	m := map[string]vfile{
		"a.pl": {rd: true, perl: accept()}, "b.pl": {rd: true, perl: accept()}, "heavy.pl": {rd: true, perl: accept()}, "mid.pl": {rd: true, perl: accept()},
		std5k: {rd: true, perl: accept()}, "edit5k.pl": {rd: true, perl: accept()},
		"outside.pl": {rd: true, perl: refuse("grammar")}, // not in the population
		listed:       {rd: true, perl: accept()},          // lexer-listed: not in the population
	}
	kept, err := aggregateFiles(p, rd, vfiles(m), vinputs(m, listed))
	if err != nil {
		t.Fatal(err)
	}
	got := p.Ratio("lex-overhead-aggregate", kept, false)
	if math.Abs(got-1.1) > 1e-9 {
		t.Errorf("lex-overhead-aggregate %v, want 1.1", got)
	}
	var timed []string
	for f := range seen {
		timed = append(timed, f)
	}
	slices.Sort(timed)
	if want := []string{"a.pl", "b.pl"}; !slices.Equal(timed, want) {
		t.Errorf("lex-overhead-aggregate times %v, want %v (outside, listed, std5k, edit5k, the >10%% files left out)", timed, want)
	}
}

// TestRealInputsExcludeTheLexerLists: the real verdict inputs carry both
// lexer lists and perl's labels.
func TestRealInputsExcludeTheLexerLists(t *testing.T) {
	dir := t.TempDir()
	known, defects := filepath.Join(dir, "known"), filepath.Join(dir, "defects")
	for f, text := range map[string]string{known: "a.pl 1 01a18fa6-aaaa\n", defects: "b.pl 2 01a18fa6-bbbb\n"} {
		if err := os.WriteFile(f, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	oldK, oldD := knownFile, defectsFile
	knownFile, defectsFile = known, defects
	defer func() { knownFile, defectsFile = oldK, oldD }()
	t.Chdir("../../../..")
	in, err := realInputs(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !in.Excluded["a.pl"] || !in.Excluded["b.pl"] {
		t.Errorf("the real inputs exclude %v, want a.pl and b.pl from the two lists", in.Excluded)
	}
	if in.Perl("internal/chart/testdata/guacamole/parses/Statements-Block-001.pl") == nil {
		t.Error("the real inputs hold no perl label for a Guacamole file")
	}
}

const resolveFixture = "use Test::More;\nplan skip_all => 'x';\n"

func TestRDVerdictsResolveUse(t *testing.T) {
	srcs := map[string][]byte{"f.pl": []byte(resolveFixture)}
	if bare, _ := rdAccepts("f.pl", func() *parse.Node { return parse.Parse(srcs["f.pl"]) }); bare {
		t.Fatal("bare Parse accepts the fixture, so it proves nothing about resolving")
	}
	if ok, err := chartInputs(nil, srcs, nil, nil, ownDir).RD("f.pl"); err != nil || !ok {
		t.Errorf("the verdict side's RD: accepted %v, %v; want true, nil", ok, err)
	}
}

func TestSpeedRDUnresolved(t *testing.T) {
	srcs := map[string][]byte{"f.pl": []byte(resolveFixture)}
	var tree *parse.Node
	old := speedParse
	speedParse = func(src []byte) *parse.Node { tree = old(src); return tree }
	defer func() { speedParse = old }()
	rdRunner(srcs)()("f.pl")
	if tree == nil || !refusedAnywhere(tree) {
		t.Error("the speed side's RD did not parse the fixture bare (a bare parse refuses it)")
	}
}

// The verdict RD is parse.ParseFileFrom's on the real path: a module beside
// the file is followed, and a file with none is judged the same.
func TestRDVerdictsMatchParseFile(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"Mine.pm":  "package Mine;\nuse Exporter 'import';\nour @EXPORT = qw(mine);\nsub mine :prototype($$) { 1 }\n1;\n",
		"uses.pl":  "use Mine;\nmine 1, 2;\n",
		"none.pl":  resolveFixture,
		"alone.pl": "use Missing::Mod;\n$x = 1;\n",
	}
	srcs := map[string][]byte{}
	for name, text := range files {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		srcs[p] = []byte(text)
	}
	in := chartInputs(nil, srcs, nil, nil, ownDir)
	for p := range srcs {
		if strings.HasSuffix(p, ".pm") {
			continue
		}
		tree, err := parse.ParseFileFrom(p)
		if err != nil {
			t.Fatal(err)
		}
		got, err := in.RD(p)
		if err != nil || got != !refusedAnywhere(tree) {
			t.Errorf("%s: verdict RD %v, %v; ParseFileFrom accepts %v", filepath.Base(p), got, err, !refusedAnywhere(tree))
		}
	}
	// The module beside the file is what accepts it: with no search directory it is refused.
	p := filepath.Join(dir, "uses.pl")
	if bare := parse.ParseWithLoader(srcs[p], parse.DirLoader()); !refusedAnywhere(bare) {
		t.Error("uses.pl is accepted without its sibling module, so the sample proves nothing about the search directory")
	}
	if ok, _ := in.RD(p); !ok {
		t.Error("the verdict RD refuses uses.pl, which its sibling module declares")
	}
}

// measureLex times the bare parse: it takes its recursive descent from
// rdRunner, not from the verdict side. The first thing it times is the
// population's RD, so a panic there stands in for the rest of the run.
func TestMeasureLexTimesBareRD(t *testing.T) {
	t.Chdir("../../../..")
	called := false
	old := speedParse
	speedParse = func([]byte) *parse.Node { called = true; panic("stop") }
	defer func() { speedParse = old }()
	in := verdictInputs{
		RD:   func(string) (bool, error) { return true, nil },
		Perl: func(string) *perlVerdict { return &perlVerdict{} },
	}
	srcs := map[string][]byte{"a.pl": []byte("1;")}
	func() {
		defer func() { recover() }()
		measureLex(committedTables(t), []File{{"a.pl", "x"}}, in, srcs, map[string]string{})
	}()
	if !called {
		t.Error("measureLex did not time parse.Parse through rdRunner")
	}
}
