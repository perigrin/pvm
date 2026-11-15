// ABOUTME: Tests SubjectFacts, the chart as a parseoracle subject, on committed samples decoded by parseoracle itself.
// ABOUTME: Each sample states its facts; the pinned perl -c accepts every sample (testdata/subject).

package forest

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// TestOracleSubject computes each sample's facts, marshals them, and has parseoracle decode them as it decodes any
// subject's, through a subject command that prints them; the decoded facts must be the ones the sample states.
func TestOracleSubject(t *testing.T) {
	tab := committedTables(t)
	samples := []struct{ file, want string }{
		// a call, under a prototype or none; f's may take @a by reference, which is Unresolved, not fitted (facts.go)
		{"call.pl", `{"ok":true,"prototypes":{"f":"\\@","h":"$"},"call_sites":[
			{"line":3,"end_line":3,"name":"f","took_reference":false,"unresolved":true},
			{"line":4,"end_line":4,"name":"g","took_reference":false},
			{"line":6,"end_line":6,"name":"h","took_reference":false}]}`},
		// a \ the source wrote, on its statement's first and last lines; code interpolated in a string is one token
		// the forest does not read, so its \ reports nothing (README "The chart as a parseoracle subject")
		{"reference.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"reference","line":2,"end_line":2,"took_reference":true},
			{"kind":"reference","line":3,"end_line":4,"took_reference":true}]}`},
		// a hash read: %h, $h{a}, $r->{b}, its {c}, %$r and @h{'a'}, one site each
		{"hash.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"hash","line":1,"end_line":1},
			{"kind":"reference","line":2,"end_line":2,"took_reference":true},
			{"kind":"hash","line":2,"end_line":2},
			{"kind":"hash","line":3,"end_line":3},{"kind":"hash","line":3,"end_line":3},{"kind":"hash","line":3,"end_line":3},
			{"kind":"hash","line":3,"end_line":3},{"kind":"hash","line":3,"end_line":3}]}`},
		// a match: =~ against a pattern or an expression, and a bare pattern; not s///, not division
		{"match.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"match","line":2,"end_line":2},{"kind":"match","line":3,"end_line":3},{"kind":"match","line":5,"end_line":5}]}`},
		// a readline: <STDIN> and readline(...); <*.c> is a glob
		{"readline.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"readline","line":1,"end_line":1},{"kind":"readline","line":3,"end_line":3}]}`},
		// an anonymous hash, braced and after +
		{"anonhash.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"anonhash","line":1,"end_line":1},{"kind":"anonhash","line":2,"end_line":2}]}`},
		// Foo::bar's shape is unknown, so its { a => 1 } is an anonymous hash in some readings and a block in
		// others: Unresolved, as the call is; the \@a every reading holds is decided
		{"ambiguous.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"line":2,"end_line":2,"name":"Foo::bar","took_reference":false,"unresolved":true},
			{"kind":"anonhash","line":2,"end_line":2,"unresolved":true},
			{"kind":"reference","line":3,"end_line":3,"took_reference":true}]}`},
		// #line renumbers what follows it, and a name other than the file's moves it out of the file's sites
		{"line.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"reference","line":2,"end_line":2,"took_reference":true},
			{"kind":"reference","line":10,"end_line":10,"took_reference":true}]}`},
	}
	dir := t.TempDir()
	t.Setenv(subjectEchoEnv, "1")
	subject := parseoracle.Subject{Command: []string{os.Args[0], "-test.run=^TestSubjectEcho$"}}
	for _, s := range samples {
		src, err := os.ReadFile(filepath.Join("testdata", "subject", s.file))
		if err != nil {
			t.Fatal(err)
		}
		out, err := json.Marshal(SubjectFacts(tab, src, s.file, scopeOf(string(src))))
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, s.file+".json")
		if err := os.WriteFile(path, out, 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := subject.Parse(context.Background(), path)
		if err != nil {
			t.Fatalf("%s: parseoracle refuses the facts: %v\n%s", s.file, err, out)
		}
		var want parseoracle.SubjectFacts
		if err := json.Unmarshal([]byte(s.want), &want); err != nil {
			t.Fatalf("%s: the stated facts: %v", s.file, err)
		}
		want.KnowsCallSites, want.KnowsPrototypes = true, true
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: facts\n%s\nwant\n%s", s.file, out, s.want)
		}
	}
}

// subjectEchoEnv makes TestSubjectEcho the subject TestOracleSubject runs: this test binary, printing a file.
const subjectEchoEnv = "CHART_SUBJECT_ECHO"

// TestSubjectEcho is no test of its own: run as TestOracleSubject's subject command, it prints the file named last
// on its command line, the facts TestOracleSubject wrote, and exits before the testing package prints anything.
func TestSubjectEcho(t *testing.T) {
	if os.Getenv(subjectEchoEnv) == "" {
		t.Skip("TestOracleSubject's subject command")
	}
	b, err := os.ReadFile(os.Args[len(os.Args)-1])
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
	os.Stdout.Write(b)
	os.Exit(0)
}

// TestSubjectFactsRefusedAndEmpty: a file the chart refuses answers ok false, an empty one ok true, each with both
// lists empty and never null, which parseoracle would read as a question not asked.
func TestSubjectFactsRefusedAndEmpty(t *testing.T) {
	tab := committedTables(t)
	for _, c := range []struct{ src, want string }{
		{"1 +;", `{"ok":false,"prototypes":{},"call_sites":[]}`}, // perl -c: syntax error
		{"", `{"ok":true,"prototypes":{},"call_sites":[]}`},
	} {
		out, err := json.Marshal(SubjectFacts(tab, []byte(c.src), "f.pl", Scope{}))
		if err != nil {
			t.Fatal(err)
		}
		if string(out) != c.want {
			t.Errorf("%q: %s, want %s", c.src, out, c.want)
		}
	}
}
