// ABOUTME: Tests waypoint 2's keys (KEYS, perigrin's 2026-11-10 "Yes, in one issue"): printed once, in order, as numbers.
// ABOUTME: Also their compare rules, and that a missing precedence fixture or oracle facts file is an error, never a 0.

package main

import (
	"bytes"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/parseoracle"
)

// waypoint2Keys are KEYS's ten keys, in the contract's order after the others.
var waypoint2Keys = []string{
	"oracle-wrong", "oracle-exact-chart", "oracle-exact-rd", "precedence-disagree", "precedence-covered",
	"precedence-uncovered-unlisted", "readings-collapsed", "collapse-over-limit", "superset-population-perlgit", "chart-beats-rd",
}

var precedenceFixture = filepath.Join("..", "..", "testdata", "precedence")

// waypoint2Measured measures every waypoint-2 key over stubbed verdicts, the oracle samples and the committed
// precedence fixture.
func waypoint2Measured(t *testing.T) map[string]string {
	t.Helper()
	measured := map[string]string{}
	files, in := collapseCorpus()
	if _, err := measureVerdicts(files, in, measured); err != nil {
		t.Fatal(err)
	}
	c := sampleCorpus(t, nil, "exact.pl", "collapsed.pl")
	if err := measureOracle([]oracleCorpus{c}, oracleKeySubjects(), measured, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	g, _, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	if err := measurePrecedence(g, committedTables(t), precedenceFixture, measured); err != nil {
		t.Fatal(err)
	}
	return measured
}

// TestWaypoint2KeysPrinted: the ten keys close the fixed order and each prints once, as a number.
func TestWaypoint2KeysPrinted(t *testing.T) {
	if len(keys) < len(waypoint2Keys) || !slices.Equal(keys[len(keys)-len(waypoint2Keys):], waypoint2Keys) {
		t.Fatalf("keys end %q, want waypoint 2's %q", keys, waypoint2Keys)
	}
	ms := report(waypoint2Measured(t))
	for _, k := range waypoint2Keys {
		n := 0
		for _, m := range ms {
			if m.Key != k {
				continue
			}
			n++
			if _, err := strconv.ParseFloat(m.Value, 64); err != nil {
				t.Errorf("%s printed %q, want a number", k, m.Value)
			}
		}
		if n != 1 {
			t.Errorf("%s printed %d times, want once", k, n)
		}
	}
}

// collapseCorpus is four corpora's files with their readings: big crosses the limit (6 of 100 statements, more
// than 1% and at least 5), few does not (4 of 10, fewer than 5), thin does not (5 of 1000, not more than 1%), and
// perlgit holds a population file, a file perl and the chart accept that recursive descent refuses, and a collapsed
// file the chart refuses, whose readings count nowhere.
func collapseCorpus() ([]File, verdictInputs) {
	type f struct {
		corpus        string
		rd, chart     bool
		stmts, collap int
	}
	m := map[string]f{
		"big.pl":     {"big", true, true, 100, 6},
		"few.pl":     {"few", true, true, 10, 4},
		"thin.pl":    {"thin", true, true, 1000, 5},
		"pop.t":      {"perlgit", true, true, 3, 0},
		"beats.t":    {"perlgit", false, true, 2, 0},
		"refused.t":  {"perlgit", true, false, 50, 40},
		"beats-2.pl": {"few", false, true, 0, 0},
	}
	var files []File
	for _, p := range slices.Sorted(maps.Keys(m)) {
		files = append(files, File{p, m[p].corpus})
	}
	return files, verdictInputs{
		Chart: func(p string) (chartResult, error) {
			r := chartResult{Accepted: m[p].chart, Recognized: true}
			if r.Accepted {
				r.Statements, r.Collapsed = m[p].stmts, m[p].collap
			}
			return r, nil
		},
		RD:       func(p string) (bool, error) { return m[p].rd, nil },
		Perl:     func(string) *perlVerdict { return accept() },
		Excluded: map[string]bool{},
	}
}

// TestWaypoint2VerdictKeys: readings-collapsed sums the accepted files' collapsed statements over every corpus,
// collapse-over-limit counts the corpora over spec §10's limit, superset-population-perlgit the perl.git files in the
// population, and chart-beats-rd the files perl and the chart accept that recursive descent refuses.
func TestWaypoint2VerdictKeys(t *testing.T) {
	measured := map[string]string{}
	files, in := collapseCorpus()
	if _, err := measureVerdicts(files, in, measured); err != nil {
		t.Fatal(err)
	}
	for k, want := range map[string]string{
		"readings-collapsed": "15", "collapse-over-limit": "1", "superset-population-perlgit": "2", "chart-beats-rd": "2",
	} {
		if measured[k] != want {
			t.Errorf("%s = %q, want %s", k, measured[k], want)
		}
	}
}

// oracleKeySubjects are a chart that grades exact.pl's site WRONG (a call on line 2 that took no reference, where
// perl took one) and collapsed.pl's EXACT, and a recursive descent that agrees with perl everywhere.
func oracleKeySubjects() []oracleSubject {
	ref := func(line int) parseoracle.SubjectCallSite {
		return parseoracle.SubjectCallSite{Kind: parseoracle.SiteKindReference, Line: line, EndLine: line, TookReference: true}
	}
	collapsed := parseoracle.SubjectFacts{OK: true, CallSites: []parseoracle.SubjectCallSite{ref(5), ref(7)}}
	return []oracleSubject{
		fakeSubject("chart", map[string]parseoracle.SubjectFacts{
			"exact.pl": {OK: true, CallSites: []parseoracle.SubjectCallSite{{Name: "f", Line: 2, EndLine: 2}}}, "collapsed.pl": collapsed}),
		fakeSubject("rd", map[string]parseoracle.SubjectFacts{
			"exact.pl": {OK: true, CallSites: []parseoracle.SubjectCallSite{ref(2)}}, "collapsed.pl": collapsed}),
	}
}

// TestWaypoint2OracleKeys: the oracle keys sum ORACLEGRADE's sites over every corpus: the chart's WRONG and EXACT,
// and recursive descent's EXACT.
func TestWaypoint2OracleKeys(t *testing.T) {
	c1, c2 := sampleCorpus(t, nil, "exact.pl"), sampleCorpus(t, nil, "exact.pl", "collapsed.pl")
	c2.name = "other"
	subjects := oracleKeySubjects()
	g, err := gradeOracle([]oracleCorpus{c1, c2}, subjects, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	var wrong, chartExact, rdExact int
	for _, c := range []string{"sample", "other"} {
		wrong += g.Sites["chart"][c][parseoracle.BucketWrong]
		chartExact += g.Sites["chart"][c][parseoracle.BucketExact]
		rdExact += g.Sites["rd"][c][parseoracle.BucketExact]
	}
	if wrong == 0 || rdExact <= chartExact {
		t.Fatalf("the fixture grades chart WRONG %d, exact %d, RD exact %d; want a WRONG site and RD ahead, or it proves nothing", wrong, chartExact, rdExact)
	}
	measured := map[string]string{}
	if err := measureOracle([]oracleCorpus{c1, c2}, subjects, measured, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	for k, want := range map[string]int{"oracle-wrong": wrong, "oracle-exact-chart": chartExact, "oracle-exact-rd": rdExact} {
		if measured[k] != strconv.Itoa(want) {
			t.Errorf("%s = %q, want %d", k, measured[k], want)
		}
	}
}

// TestWaypoint2PrecedenceKeys: over the committed fixture, the counts are the referee's. A kept refused sentence
// marked accepted by bison is one disagreement.
func TestWaypoint2PrecedenceKeys(t *testing.T) {
	measured := waypoint2Measured(t)
	if measured["precedence-covered"] == "0" || measured["precedence-covered"] == "" {
		t.Errorf("precedence-covered = %q, want the kept sentences' coverage", measured["precedence-covered"])
	}
	dir := t.TempDir()
	for _, name := range []string{"bison-conflicts.txt", "resolutions.txt", "sentences.txt", "uncovered-reviewed.txt", "rendering.txt"} {
		b, err := os.ReadFile(filepath.Join(precedenceFixture, name))
		if err != nil {
			t.Fatal(err)
		}
		if name == "sentences.txt" {
			b = flipFirstKeptRefusal(t, b)
		}
		if err := os.WriteFile(filepath.Join(dir, name), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	flipped := map[string]string{}
	g, _, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	if err := measurePrecedence(g, committedTables(t), dir, flipped); err != nil {
		t.Fatal(err)
	}
	want, _ := strconv.Atoi(measured["precedence-disagree"])
	if flipped["precedence-disagree"] != strconv.Itoa(want+1) {
		t.Errorf("one bison verdict flipped: precedence-disagree %q, want %d", flipped["precedence-disagree"], want+1)
	}
}

// flipFirstKeptRefusal marks the first kept sentence bison refuses as accepted.
func flipFirstKeptRefusal(t *testing.T, b []byte) []byte {
	t.Helper()
	lines := bytes.Split(b, []byte("\n"))
	for i, l := range lines {
		f := bytes.Split(l, []byte("\t"))
		if len(f) == 5 && !bytes.Equal(f[4], []byte("-")) && !bytes.Equal(f[2], []byte("accept")) && l[0] != '#' {
			f[2] = []byte("accept")
			lines[i] = bytes.Join(f, []byte("\t"))
			return bytes.Join(lines, []byte("\n"))
		}
	}
	t.Fatal("no kept sentence bison refuses")
	return nil
}

// TestWaypoint2KeysCompareRules: the gated counts may not rise, precedence-covered may not fall, a new key printing
// na regresses, the reported keys move freely, and the timing keys pass moving either way.
func TestWaypoint2KeysCompareRules(t *testing.T) {
	for _, tc := range []struct {
		key, base, measure string
		regress            bool
	}{
		{"oracle-wrong", "13", "14", true},
		{"oracle-wrong", "13", "12", false},
		{"precedence-disagree", "0", "1", true},
		{"precedence-uncovered-unlisted", "0", "1", true},
		{"collapse-over-limit", "0", "1", true},
		{"precedence-covered", "1304", "1303", true},
		{"precedence-covered", "1304", "1305", false},
		{"oracle-exact-chart", "100", "90", false},
		{"oracle-exact-rd", "100", "110", false},
		{"readings-collapsed", "245", "300", false},
		{"superset-population-perlgit", "515", "400", false},
		{"chart-beats-rd", "3", "9", false},
		{"lex-overhead-aggregate", "1.2", "0.9", false},
		{"speed-ratio-aggregate-perfile", "5.2", "9.9", false},
		{"speed-ratio-aggregate-perfile", "5.2", "1.1", false},
		{"speed-stage-forest-aggregate-longlived", "0.7", "2.0", false},
		{"alloc-per-token-forest", "0.06", "0.01", false},
		{"gc-cpu-share-chart", "0.18", "0.9", false},
	} {
		if got := compareOne(tc.key, tc.base, tc.measure, ""); (len(got) > 0) != tc.regress {
			t.Errorf("%s %s -> %s: %q, want regression %v", tc.key, tc.base, tc.measure, got, tc.regress)
		}
	}
	for _, k := range waypoint2Keys {
		if got := compareOne(k, "0", na, ""); len(got) != 1 {
			t.Errorf("%s 0 -> na: %q, want one regression", k, got)
		}
	}
	if slices.Contains(keys, "speed-ratio-std5k-cold") {
		t.Error("the retired speed-ratio-std5k-cold is still a key (perigrin, 2026-11-14)")
	}
}

// TestWaypoint2KeysMissingInputIsError: a precedence fixture or oracle facts file that is missing or unreadable is an
// error, which main turns into exit 2, and measures none of the keys it feeds.
func TestWaypoint2KeysMissingInputIsError(t *testing.T) {
	g, _, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	unreadable := t.TempDir()
	if err := os.Mkdir(filepath.Join(unreadable, "sentences.txt"), 0o755); err != nil { // a directory reads as an error
		t.Fatal(err)
	}
	for _, name := range []string{"bison-conflicts.txt", "resolutions.txt", "uncovered-reviewed.txt", "rendering.txt"} {
		b, err := os.ReadFile(filepath.Join(precedenceFixture, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(unreadable, name), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for name, dir := range map[string]string{"missing": t.TempDir(), "unreadable": unreadable} {
		measured := map[string]string{}
		if err := measurePrecedence(g, committedTables(t), dir, measured); err == nil || len(measured) != 0 {
			t.Errorf("a %s precedence fixture: err %v, measured %v; want an error and no key", name, err, measured)
		}
	}
	subjects := defaultSubjects(committedTables(t))
	for name, facts := range map[string]string{"missing": filepath.Join(t.TempDir(), "gone.txt"), "unreadable": t.TempDir()} {
		c := sampleCorpus(t, nil, "exact.pl")
		c.facts = facts
		measured := map[string]string{}
		if err := measureOracle([]oracleCorpus{c}, subjects, measured, &bytes.Buffer{}); err == nil || len(measured) != 0 {
			t.Errorf("a %s oracle facts file: err %v, measured %v; want an error and no key", name, err, measured)
		}
	}
}
