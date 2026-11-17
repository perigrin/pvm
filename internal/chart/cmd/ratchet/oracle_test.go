// ABOUTME: Tests the oracle grading measure (oracle.go) on committed samples whose perl facts each row states.
// ABOUTME: Every sample's facts are the pinned perl's (testdata/oracle; PERL5LIB=testdata/oracle/lib for import.pl).

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/parseoracle"
)

var oracleSamples = filepath.Join("testdata", "oracle")

// samplePerl is the pinned perl's facts for each sample, as oracle-facts.sh records them.
var samplePerl = map[string]parseoracle.Facts{
	"exact.pl":      {OK: true, Walked: true, OpCount: 8, Sites: map[parseoracle.Marker][]int{"srefgen": {2}}},
	"wider.pl":      {OK: true, Walked: true, OpCount: 10, Sites: map[parseoracle.Marker][]int{"srefgen": {3}}, Prototypes: map[string]string{"f": `\@`}},
	"collapsed.pl":  {OK: true, Walked: true, OpCount: 51, Sites: map[parseoracle.Marker][]int{"srefgen": {5, 7}}},
	"import.pl":     {OK: true, Walked: true, OpCount: 10, Sites: map[parseoracle.Marker][]int{"srefgen": {3}}, Prototypes: map[string]string{"r": `\@`}},
	"smartmatch.pl": {OK: true, Walked: true, OpCount: 65, Sites: map[parseoracle.Marker][]int{"srefgen": {3, 4, 5, 5, 6, 8, 9}}},
}

// sampleCorpus writes a facts file holding the named samples' records and returns it as a corpus over testdata/oracle.
func sampleCorpus(t *testing.T, roots []string, names ...string) oracleCorpus {
	t.Helper()
	var lines []string
	for _, n := range names {
		l, err := chart.FormatOracleRecord(chart.OracleRecord{Key: n, Facts: samplePerl[n]})
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, l)
	}
	return oracleCorpus{name: "sample", facts: writeFacts(t, strings.Join(lines, "\n")+"\n"),
		path: func(k string) string { return filepath.Join(oracleSamples, k) }, roots: roots}
}

func writeFacts(t *testing.T, text string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "facts.txt")
	if err := os.WriteFile(p, []byte("# perl: the pinned perl\n"+text), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// fakeSubject answers every file with the facts its row gives, by sample name.
func fakeSubject(name string, by map[string]parseoracle.SubjectFacts) oracleSubject {
	return oracleSubject{name, func(path string, _ []byte, _ *parse.Node) parseoracle.SubjectFacts {
		return by[filepath.Base(path)]
	}}
}

func totals(e, w, x, n int) parseoracle.Totals {
	var t parseoracle.Totals
	t[parseoracle.BucketExact], t[parseoracle.BucketWider], t[parseoracle.BucketWrong], t[parseoracle.BucketNoAnswer] = e, w, x, n
	return t
}

// TestOracleGrading: per site, the chart grades an EXACT site, a WIDER one from a call whose reading hangs on a \
// prototype slot it does not fit, and a collapsed statement's site every reading holds EXACT, as is the one after it;
// recursive descent is graded the same way; a deliberately wrong fake subject grades WRONG, its site listed on
// stderr, and a fake agreeing with perl everywhere grades no WRONG.
func TestOracleGrading(t *testing.T) {
	c := sampleCorpus(t, nil, "exact.pl", "wider.pl", "collapsed.pl")
	ref := func(line int) parseoracle.SubjectCallSite {
		return parseoracle.SubjectCallSite{Kind: parseoracle.SiteKindReference, Line: line, EndLine: line, TookReference: true}
	}
	agree := fakeSubject("agree", map[string]parseoracle.SubjectFacts{
		"exact.pl":     {OK: true, CallSites: []parseoracle.SubjectCallSite{ref(2)}},
		"wider.pl":     {OK: true, CallSites: []parseoracle.SubjectCallSite{ref(3)}},
		"collapsed.pl": {OK: true, CallSites: []parseoracle.SubjectCallSite{ref(5), ref(7)}},
	})
	// a call on line 2 that took no reference: a committed answer to the reference question, and perl took one
	wrong := fakeSubject("wrong", map[string]parseoracle.SubjectFacts{
		"exact.pl":     {OK: true, CallSites: []parseoracle.SubjectCallSite{{Name: "f", Line: 2, EndLine: 2}}},
		"wider.pl":     {OK: true, CallSites: []parseoracle.SubjectCallSite{ref(3)}},
		"collapsed.pl": {OK: true, CallSites: []parseoracle.SubjectCallSite{ref(5), ref(7)}},
	})
	var stderr bytes.Buffer
	got, err := gradeOracle([]oracleCorpus{c}, append(defaultSubjects(committedTables(t)), agree, wrong), &stderr)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range []struct {
		subject string
		want    parseoracle.Totals
	}{
		{"chart", totals(3, 1, 0, 0)}, // a collapsed statement hedges only the sites its readings disagree on
		{"rd", totals(4, 0, 0, 0)},    // it fits f's \@ slot, and nothing collapses for it
		{"agree", totals(4, 0, 0, 0)},
		{"wrong", totals(3, 0, 1, 0)},
	} {
		if g := got.Sites[w.subject]["sample"]; g != w.want {
			t.Errorf("%s: %v (exact, wider, wrong, no-answer by Bucket), want %v", w.subject, g, w.want)
		}
	}
	if line := "oracle-grade: wrong WRONG " + filepath.Join(oracleSamples, "exact.pl") + ":2 srefgen\n"; !strings.Contains(stderr.String(), line) {
		t.Errorf("stderr does not list the WRONG site %q:\n%s", line, stderr.String())
	}
	if strings.Contains(stderr.String(), "oracle-grade: chart WRONG") {
		t.Errorf("stderr lists a chart WRONG site:\n%s", stderr.String())
	}
}

// TestOracleGradingSmartmatch: perl takes a reference to an array or hash operand of ~~ and to one that is given's
// or when's condition (op.c's ref_array_or_hash), not to a slice or an anonymous array; the chart grades every
// such site EXACT (perl.git op/smartmatch.t:168-171 and op/switch.t:1101 were WRONG).
func TestOracleGradingSmartmatch(t *testing.T) {
	got, err := gradeOracle([]oracleCorpus{sampleCorpus(t, nil, "smartmatch.pl")}, defaultSubjects(committedTables(t)), &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if g := got.Sites["chart"]["sample"]; g != totals(7, 0, 0, 0) {
		t.Errorf("chart: %v (exact, wider, wrong, no-answer by Bucket), want all 7 exact", g)
	}
}

// TestOracleGradingWithoutFactsFails: no committed facts is an error, never WRONG 0; a malformed record is an
// error; a record whose file cannot be read is named on stderr and counted, its sites no-answer.
func TestOracleGradingWithoutFactsFails(t *testing.T) {
	subjects := defaultSubjects(committedTables(t))
	if _, err := gradeOracle(nil, subjects, &bytes.Buffer{}); err == nil {
		t.Error("grading no corpus: no error")
	}
	empty := oracleCorpus{name: "empty", facts: writeFacts(t, ""), path: func(k string) string { return k }}
	if _, err := gradeOracle([]oracleCorpus{empty}, subjects, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "no committed facts") {
		t.Errorf("grading a facts file with no record: %v, want an error naming no committed facts", err)
	}
	bad := oracleCorpus{name: "bad", facts: writeFacts(t, "exact.pl\tfacts\t{\"ok\":tru\n"), path: empty.path}
	if _, err := gradeOracle([]oracleCorpus{bad}, subjects, &bytes.Buffer{}); err == nil {
		t.Error("grading a malformed record: no error")
	}

	c := sampleCorpus(t, nil, "exact.pl")
	gone, err := chart.FormatOracleRecord(chart.OracleRecord{Key: "gone.pl", Facts: samplePerl["collapsed.pl"]})
	if err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(c.facts)
	if err != nil {
		t.Fatal(err)
	}
	c.facts = writeFacts(t, string(text)+gone+"\n")
	var stderr bytes.Buffer
	got, err := gradeOracle([]oracleCorpus{c}, subjects, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if got.Unreadable != 1 || !strings.Contains(stderr.String(), "gone.pl") {
		t.Errorf("unreadable %d, want 1, named on stderr:\n%s", got.Unreadable, stderr.String())
	}
	if g := got.Sites["chart"]["sample"]; g != totals(1, 0, 0, 2) {
		t.Errorf("chart: %v, want exact.pl's site exact and gone.pl's two no-answer", g)
	}
}

// TestOracleGradingNoAnswerIsNotExact: a subject that reports no call sites, or none of the marker's kind, has not
// answered: perl's sites grade no-answer, never EXACT.
func TestOracleGradingNoAnswerIsNotExact(t *testing.T) {
	c := sampleCorpus(t, nil, "exact.pl", "collapsed.pl")
	silent := fakeSubject("silent", map[string]parseoracle.SubjectFacts{"exact.pl": {OK: true}, "collapsed.pl": {OK: true}})
	hashOnly := []parseoracle.SubjectCallSite{{Kind: parseoracle.SiteKindHash, Line: 2, EndLine: 2}}
	other := fakeSubject("other", map[string]parseoracle.SubjectFacts{
		"exact.pl": {OK: true, CallSites: hashOnly}, "collapsed.pl": {OK: true, CallSites: hashOnly}})
	got, err := gradeOracle([]oracleCorpus{c}, []oracleSubject{silent, other}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"silent", "other"} {
		if g := got.Sites[s]["sample"]; g != totals(0, 0, 0, 3) {
			t.Errorf("%s: %v, want every site no-answer", s, g)
		}
	}
}

// TestOracleGradingRDResolvesImports: import.pl's r(@a) takes @a by reference only under R's \@ prototype, from
// lib/R.pm; recursive descent grades it EXACT with lib searched, and not without.
func TestOracleGradingRDResolvesImports(t *testing.T) {
	subjects := defaultSubjects(committedTables(t))
	for _, c := range []struct {
		roots []string
		exact int
	}{{[]string{filepath.Join(oracleSamples, "lib")}, 1}, {nil, 0}} {
		got, err := gradeOracle([]oracleCorpus{sampleCorpus(t, c.roots, "import.pl")}, subjects, &bytes.Buffer{})
		if err != nil {
			t.Fatal(err)
		}
		if g := got.Sites["rd"]["sample"]; g[parseoracle.BucketExact] != c.exact {
			t.Errorf("roots %v: recursive descent %v, want %d exact", c.roots, g, c.exact)
		}
	}
}
