// ABOUTME: Tests the verdict keys over a vfile population with stubbed chart, recursive-descent and perl verdicts.
// ABOUTME: Membership rules, Q11's lexer-list exclusion, error on a missing label or a recognizer fault, finite ratios.

package main

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// vfile is one stubbed file: what each side says about it.
type vfile struct {
	rd, chart bool
	perl      *perlVerdict // nil: no label
	res       chartResult  // the chart's counts; Accepted is overridden by chart
	err       error
}

func accept() *perlVerdict         { return &perlVerdict{} }
func refuse(c string) *perlVerdict { return &perlVerdict{Refused: true, Category: c} }
func vfiles(m map[string]vfile) []File {
	var fs []File
	for p := range m {
		fs = append(fs, File{Path: p})
	}
	slices.SortFunc(fs, func(a, b File) int { return strings.Compare(a.Path, b.Path) })
	return fs
}

func vinputs(m map[string]vfile, excluded ...string) verdictInputs {
	ex := map[string]bool{}
	for _, p := range excluded {
		ex[p] = true
	}
	return verdictInputs{
		Chart: func(p string) (chartResult, error) {
			r := m[p].res
			r.Accepted = m[p].chart
			r.Recognized = r.Recognized || r.Accepted // the pipeline accepts only what recognition does
			return r, m[p].err
		},
		RD:       func(p string) (bool, error) { return m[p].rd, nil },
		Perl:     func(p string) *perlVerdict { return m[p].perl },
		Excluded: ex,
	}
}

func run(t *testing.T, m map[string]vfile, excluded ...string) (map[string]string, map[string][]string) {
	t.Helper()
	measured := map[string]string{}
	refused, err := measureVerdicts(vfiles(m), vinputs(m, excluded...), measured)
	if err != nil {
		t.Fatal(err)
	}
	return measured, refused
}

func TestVerdictKeys(t *testing.T) {
	m := map[string]vfile{
		"ok.pl":         {rd: true, chart: true, perl: accept()},
		"refused1.pl":   {rd: true, chart: false, perl: accept()},
		"refused2.pl":   {rd: true, chart: false, perl: accept()},
		"listed.pl":     {rd: true, chart: false, perl: accept()}, // on a lexer list: left out
		"rdonly.pl":     {rd: false, chart: false, perl: accept()},
		"perlrefuse.pl": {rd: true, chart: false, perl: refuse("grammar")}, // not in the population
		"both.pl":       {rd: false, chart: true, perl: refuse("grammar")},
		"bothlisted.pl": {rd: false, chart: true, perl: refuse("grammar")}, // lexer list: left out
		"rdaccepts.pl":  {rd: true, chart: true, perl: refuse("grammar")},  // RD accepts: not "both refused"
	}
	got, refused := run(t, m, "listed.pl", "bothlisted.pl")
	// population: perl and RD accept, less the lexer-listed: ok, refused1, refused2.
	for k, want := range map[string]string{
		"rd-accepted-chart-refused":   "2",
		"chart-accepted-both-refused": "1",
		// perl refuses and the chart accepts: both, bothlisted, rdaccepts (lists do not apply).
		"chart-accepted-perl-refused": "3",
		"superset-recognized":         "0.333",
	} {
		if got[k] != want {
			t.Errorf("%s = %q, want %q", k, got[k], want)
		}
	}
	if g := strings.Join(refused["rd-accepted-chart-refused"], " "); g != "refused1.pl refused2.pl" {
		t.Errorf("refused list = %q", g)
	}
	if g := strings.Join(refused["chart-accepted-both-refused"], " "); g != "both.pl" {
		t.Errorf("both-refused list = %q", g)
	}
}

func TestListRises(t *testing.T) {
	ms := []Measure{{"rd-accepted-chart-refused", "2"}, {"chart-accepted-both-refused", "0"}}
	lists := map[string][]string{"rd-accepted-chart-refused": {"a.pl", "b.pl"}, "chart-accepted-both-refused": nil}
	base := Format([]Measure{{"rd-accepted-chart-refused", "1"}, {"chart-accepted-both-refused", "0"}})
	var w strings.Builder
	listRises(&w, base, ms, lists)
	if want := "chart-ratchet: rd-accepted-chart-refused: a.pl\nchart-ratchet: rd-accepted-chart-refused: b.pl\n"; w.String() != want {
		t.Errorf("a rise listed %q, want %q", w.String(), want)
	}
	w.Reset()
	listRises(&w, Format([]Measure{{"rd-accepted-chart-refused", "2"}}), ms, lists)
	listRises(&w, Format([]Measure{{"rd-accepted-chart-refused", "na"}}), ms, lists)
	if w.Len() != 0 {
		t.Errorf("no rise (equal baseline, or na) still listed %q", w.String())
	}
}

func TestVerdictKeysNeedEveryLabel(t *testing.T) {
	m := map[string]vfile{"a.pl": {rd: true, chart: true, perl: accept()}, "b.pl": {rd: true}}
	if _, err := measureVerdicts(vfiles(m), vinputs(m), map[string]string{}); err == nil || !strings.Contains(err.Error(), "b.pl") {
		t.Errorf("an unlabelled file: err = %v, want one naming b.pl", err)
	}
}

// TestRatchetVerdictIsPipeline: the chart's verdict is the whole pipeline's
// (forest.Parse), so a comparison chain (precedence) and a method outside a
// class under the class feature (scope), which recognition accepts and perl
// -c refuses, are refused, while the same method word with the feature off,
// which perl reads as a call, is accepted; a call to a sub a module beside the
// file exports is accepted only under the file's scope facts (parse.Imports
// over the verdict RD's tree). And chart-accepted-both-refused counts a file
// perl refuses for precedence or scope as it counts a grammar one, never a
// semantic one.
func TestRatchetVerdictIsPipeline(t *testing.T) {
	dir := t.TempDir()
	srcs := map[string][]byte{}
	for name, text := range map[string]string{
		"prec.pl":    "$a <=> $b <=> $c;",
		"scope.pl":   "use feature 'class'; no warnings; method foo {}",
		"word.pl":    "method foo;",
		"imports.pl": "use Mine;\nmine 1, 2;\n",
		"Mine.pm":    "package Mine;\nuse Exporter 'import';\nour @EXPORT = qw(mine);\nsub mine :prototype($$) { 1 }\n1;\n",
	} {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		srcs[name] = []byte(text)
	}
	dirs := func(string) []string { return []string{dir} }
	tab := committedTables(t)
	in := chartInputs(tab, srcs, nil, nil, dirs)
	for p, accepted := range map[string]bool{"prec.pl": false, "scope.pl": false, "word.pl": true, "imports.pl": true} {
		if r, err := in.Chart(p); err != nil || !r.Recognized || r.Accepted != accepted {
			t.Errorf("%s: %+v, %v; want recognized, and accepted=%v by the pipeline", p, r, err, accepted)
		}
	}
	if r, _ := chartInputs(tab, srcs, nil, nil, ownDir).Chart("imports.pl"); r.Accepted {
		t.Error("imports.pl is accepted without its module, so it proves nothing about the scope facts")
	}
	m := map[string]vfile{
		"prec.pl":  {chart: true, perl: refuse("precedence")},
		"scope.pl": {chart: true, perl: refuse("scope")},
		"gram.pl":  {chart: true, perl: refuse("grammar")},
		"sem.pl":   {chart: true, perl: refuse("semantic")},
	}
	got, refused := run(t, m)
	if g := strings.Join(refused["chart-accepted-both-refused"], " "); got["chart-accepted-both-refused"] != "3" || g != "gram.pl prec.pl scope.pl" {
		t.Errorf("chart-accepted-both-refused %q over %q; want 3: gram.pl prec.pl scope.pl", got["chart-accepted-both-refused"], g)
	}
}

// TestRatchetCountsPipelineOverRefusal: a file recognition accepts and the
// pipeline refuses is the chart's refusal, so in the population it counts in
// rd-accepted-chart-refused and an over-refusing constraint cannot hide.
func TestRatchetCountsPipelineOverRefusal(t *testing.T) {
	m := map[string]vfile{"over.pl": {rd: true, perl: accept(), res: chartResult{Recognized: true}}}
	got, refused := run(t, m)
	if got["rd-accepted-chart-refused"] != "1" || len(refused["rd-accepted-chart-refused"]) != 1 || got["superset-recognized"] != "0.000" {
		t.Errorf("rd-accepted-chart-refused %q (%v), superset-recognized %q; want 1 (over.pl) and 0.000",
			got["rd-accepted-chart-refused"], refused["rd-accepted-chart-refused"], got["superset-recognized"])
	}
}

// TestRecognitionKeysUnchangedByPipeline: leo-disagree, items-per-token,
// set-reuse and lex-unmapped stay recognition's. On a file recognition
// accepts and the pipeline refuses, each reads the same as under a verdict
// that takes recognition's acceptance.
func TestRecognitionKeysUnchangedByPipeline(t *testing.T) {
	tab := committedTables(t)
	srcs := map[string][]byte{"prec.pl": []byte("$a <=> $b <=> $c;")}
	files := []File{{"prec.pl", "x"}}
	piped := chartInputs(tab, srcs, nil, func(string) *perlVerdict { return refuse("precedence") }, ownDir)
	if r, err := piped.Chart("prec.pl"); err != nil || r.Accepted || !r.Recognized {
		t.Fatalf("the fixture: %+v, %v; want recognized and refused by the pipeline, or it proves nothing", r, err)
	}
	recognized := piped
	recognized.Chart = func(p string) (chartResult, error) {
		r, err := piped.Chart(p)
		r.Accepted = r.Recognized
		return r, err
	}
	keys := func(in verdictInputs) map[string]string {
		m := map[string]string{}
		if _, err := measureVerdicts(files, in, m); err != nil {
			t.Fatal(err)
		}
		if err := measureLeo(tab, files, srcs, m); err != nil {
			t.Fatal(err)
		}
		m["lex-unmapped"] = strconv.Itoa(len(lexUnmapped(tab.Tokens, files, srcs)))
		return m
	}
	got, want := keys(piped), keys(recognized)
	for _, k := range []string{"leo-disagree", "items-per-token", "set-reuse", "lex-unmapped"} {
		if got[k] == "" || got[k] != want[k] {
			t.Errorf("%s = %q under the pipeline's verdict, %q under recognition's; want the same number", k, got[k], want[k])
		}
	}
}

func TestVerdictKeysRecognizerErrorIsAnError(t *testing.T) {
	boom := errors.New("invalid lattice")
	m := map[string]vfile{"ok.pl": {rd: true, chart: true, perl: accept()}, "bad.pl": {rd: true, perl: accept(), err: boom}}
	measured := map[string]string{}
	_, err := measureVerdicts(vfiles(m), vinputs(m), measured)
	if !errors.Is(err, boom) || !strings.Contains(err.Error(), "bad.pl") {
		t.Fatalf("err = %v, want the recognizer's error naming bad.pl", err)
	}
	// A refused file is no error and is counted.
	m["bad.pl"] = vfile{rd: true, perl: accept()}
	got, _ := run(t, m)
	if got["rd-accepted-chart-refused"] != "1" {
		t.Errorf("a refused file counted %q, want 1", got["rd-accepted-chart-refused"])
	}
}

func TestVerdictRatiosFinite(t *testing.T) {
	empty := map[string]vfile{}
	got, _ := run(t, empty)
	for _, k := range []string{"superset-recognized", "items-per-token", "terminals-per-token", "set-reuse"} {
		if v, ok := got[k]; ok && v != na {
			t.Errorf("empty corpus: %s = %q, want na", k, v)
		}
	}
	// An empty population with labels: the fraction is na, the count is 0.
	got, _ = run(t, map[string]vfile{"p.pl": {perl: refuse("grammar")}})
	if _, ok := got["superset-recognized"]; ok || got["rd-accepted-chart-refused"] != "0" {
		t.Errorf("empty population: %v", got)
	}
	// An empty file adds nothing; a non-empty one gives finite numbers.
	m := map[string]vfile{
		"empty.pl": {perl: accept(), res: chartResult{}},
		"full.pl":  {chart: true, perl: accept(), res: chartResult{Tokens: 4, Terminals: 6, Items: 10, Hits: 1, Positions: 4}},
	}
	got, _ = run(t, m)
	for k, want := range map[string]string{"items-per-token": "2.500", "terminals-per-token": "1.500", "set-reuse": "0.250"} {
		if got[k] != want {
			t.Errorf("%s = %q, want %q", k, got[k], want)
		}
	}
}

func TestVerdictKeysUnlistedUnmappedCounts(t *testing.T) {
	// A file the chart's table cannot map is the chart's refusal unless it is listed.
	m := map[string]vfile{
		"unmapped.pl": {rd: true, chart: false, perl: accept()},
		"defect.pl":   {rd: true, chart: false, perl: accept()},
	}
	got, _ := run(t, m, "defect.pl")
	if got["rd-accepted-chart-refused"] != "1" {
		t.Errorf("rd-accepted-chart-refused = %q, want 1 (the unlisted unmapped file)", got["rd-accepted-chart-refused"])
	}
	if got, err := listedPaths("# a comment\nT1/a.t 12 01a18fa6-aaaa\n\nb.pl 3 01a18fa6-bbbb\n"), error(nil); err != nil || !got["T1/a.t"] || !got["b.pl"] || len(got) != 2 {
		t.Errorf("listedPaths = %v", got)
	}
}

// Items and set reuse are the cost on recognized input: a file the chart refuses adds nothing.
func TestVerdictCostOverAcceptedFilesOnly(t *testing.T) {
	m := map[string]vfile{
		"yes.pl": {chart: true, perl: accept(), res: chartResult{Positions: 4, Items: 10, Hits: 1}},
		"no.pl":  {perl: accept(), res: chartResult{Positions: 100, Items: 900, Hits: 90}},
	}
	got, _ := run(t, m)
	if got["items-per-token"] != "2.500" || got["set-reuse"] != "0.250" {
		t.Errorf("items-per-token %q, set-reuse %q; want 2.500 and 0.250", got["items-per-token"], got["set-reuse"])
	}
}

// committedTables is the committed grammar's tables, generated once for the package's tests, none of which writes them.
func committedTables(t *testing.T) *grammar.Tables {
	t.Helper()
	committed.Do(func() {
		g, ref, err := grammar.Committed()
		if err == nil {
			committed.tab, err = grammar.Generate(g, ref)
		}
		committed.err = err
	})
	if committed.err != nil {
		t.Fatal(committed.err)
	}
	return committed.tab
}

var committed struct {
	sync.Once
	tab *grammar.Tables
	err error
}

// The real chart side over the committed tables: "1;" is accepted; "1 \x01;"
// has a token the table cannot map, so it is refused, and that token neither
// counts as a token nor holds a terminal, so it does not lower the width.
func TestChartInputsUnmappedIsRefused(t *testing.T) {
	srcs := map[string][]byte{"ok.pl": []byte("1;"), "rep.pl": []byte("1; 1; 1; 1;"), "bad.pl": []byte("1 \x01;")}
	in := chartInputs(committedTables(t), srcs, nil, nil, ownDir)
	ok, err := in.Chart("ok.pl")
	if err != nil || !ok.Accepted {
		t.Fatalf("1; = %+v, %v; want accepted", ok, err)
	}
	// The exact counts feed baselined keys: node 0 is neither a position nor
	// an item, and hits and positions are not swapped.
	if ok.Positions != 2 || ok.Items != 11 || ok.Hits != 0 {
		t.Errorf("1; = %+v; want 2 positions, 11 items, 0 hits", ok)
	}
	rep, err := in.Chart("rep.pl")
	if err != nil || rep.Positions != 8 || rep.Items != 44 || rep.Hits != 3 {
		t.Errorf("1; 1; 1; 1; = %+v, %v; want 8 positions, 44 items, 3 hits", rep, err)
	}
	bad, err := in.Chart("bad.pl")
	if err != nil || bad.Accepted {
		t.Fatalf("1 \\x01; = %+v, %v; want refused", bad, err)
	}
	if bad.Tokens != 2 || bad.Terminals != 2 {
		t.Errorf("1 \\x01; counts %d tokens, %d terminals; want the 2 mapped ones, 1 terminal each", bad.Tokens, bad.Terminals)
	}
}

func TestLexerListedReadsBothLists(t *testing.T) {
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
	got, err := lexerListed()
	if err != nil || len(got) != 2 || !got["a.pl"] || !got["b.pl"] {
		t.Errorf("lexerListed = %v, %v; want a.pl from the known list and b.pl from the defects list", got, err)
	}
}

// TestLabelVerdicts: the committed label files give perl's verdict by the
// ratchet's corpus path (T1 files as T1/<name>.t, perl.git's as
// perlgit/<path under t/>); an unlabelled path is nil,
// and a missing label file is an error, never an empty set of verdicts.
func TestLabelVerdicts(t *testing.T) {
	dir := t.TempDir()
	if _, err := labelVerdicts(dir); err == nil {
		t.Fatal("no label files: want an error")
	}
	if err := os.WriteFile(filepath.Join(dir, "in-tree.txt"), []byte("# perl: p\na.pl\taccept\nb.pl\trefuse\tsemantic\treviewed\tCan't modify\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "t1.txt"), []byte("x.t\trefuse\tgrammar\treviewed\tsyntax error\treferee-noprec refuse@3; referee-prec refuse@3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "perlgit.txt"), []byte("op/y.t\taccept\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	perl, err := labelVerdicts(dir)
	if err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]*perlVerdict{
		"a.pl": {}, "b.pl": {Refused: true, Category: "semantic"}, "T1/x.t": {Refused: true, Category: "grammar"},
		"perlgit/op/y.t": {}, "x.t": nil, "c.pl": nil, "op/y.t": nil,
	} {
		if got := perl(path); (got == nil) != (want == nil) || got != nil && *got != *want {
			t.Errorf("%s: %+v, want %+v", path, got, want)
		}
	}
}
