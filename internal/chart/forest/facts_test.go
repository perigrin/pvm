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
		// a call, under a prototype or none: f's \@ slot takes @a by reference, fitted slot by slot (protoRefs); g's
		// prototype resolve cannot see, so it is Unresolved; h, declared with a $ slot, takes none
		{"call.pl", `{"ok":true,"prototypes":{"f":"\\@","h":"$"},"call_sites":[
			{"line":3,"end_line":3,"name":"f","took_reference":false},
			{"kind":"reference","line":3,"end_line":3,"took_reference":true},
			{"line":4,"end_line":4,"name":"g","took_reference":false,"unresolved":true},
			{"line":6,"end_line":6,"name":"h","took_reference":false}]}`},
		// a call to a word nothing declares, where a module resolve cannot see may give it a \ prototype: perl
		// takes @array by reference here, so the call is Unresolved, not "took no reference"
		{"undeclared.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"line":3,"end_line":3,"name":"is_shared","took_reference":false,"unresolved":true}]}`},
		// a \ the source wrote, on its statement's first and last lines; code interpolated in a string or a pattern
		// is one token the forest does not read, so its statement hedges every marker (parse.InterpolatesCode),
		// while the pattern's own match stays decided
		{"reference.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"reference","line":2,"end_line":2,"took_reference":true},
			{"kind":"reference","line":3,"end_line":4,"took_reference":true},
			{"kind":"match","line":6,"end_line":6},
			{"kind":"reference","line":5,"end_line":5,"took_reference":false,"unresolved":true},
			{"kind":"hash","line":5,"end_line":5,"unresolved":true},
			{"kind":"match","line":5,"end_line":5,"unresolved":true},
			{"kind":"readline","line":5,"end_line":5,"unresolved":true},
			{"kind":"anonhash","line":5,"end_line":5,"unresolved":true},
			{"kind":"reference","line":6,"end_line":6,"took_reference":false,"unresolved":true},
			{"kind":"hash","line":6,"end_line":6,"unresolved":true},
			{"kind":"match","line":6,"end_line":6,"unresolved":true},
			{"kind":"readline","line":6,"end_line":6,"unresolved":true},
			{"kind":"anonhash","line":6,"end_line":6,"unresolved":true}]}`},
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
		// six ambiguous statements make the block 4^6 readings, above 2^10, so it collapses; a collapsed statement
		// hedges only the sites its readings disagree on (perigrin, 2026-11-16, "Narrow it"): each Foo::bar and its
		// { a => 1 } stay Unresolved, and the \@a every reading holds is decided, as is the one after the block
		{"collapsed.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"line":2,"end_line":6,"name":"Foo::bar","took_reference":false,"unresolved":true},
			{"kind":"anonhash","line":2,"end_line":6,"unresolved":true},
			{"line":2,"end_line":6,"name":"Foo::bar","took_reference":false,"unresolved":true},
			{"kind":"anonhash","line":2,"end_line":6,"unresolved":true},
			{"line":2,"end_line":6,"name":"Foo::bar","took_reference":false,"unresolved":true},
			{"kind":"anonhash","line":2,"end_line":6,"unresolved":true},
			{"line":2,"end_line":6,"name":"Foo::bar","took_reference":false,"unresolved":true},
			{"kind":"anonhash","line":2,"end_line":6,"unresolved":true},
			{"line":2,"end_line":6,"name":"Foo::bar","took_reference":false,"unresolved":true},
			{"kind":"anonhash","line":2,"end_line":6,"unresolved":true},
			{"line":2,"end_line":6,"name":"Foo::bar","took_reference":false,"unresolved":true},
			{"kind":"anonhash","line":2,"end_line":6,"unresolved":true},
			{"kind":"reference","line":2,"end_line":6,"took_reference":true},
			{"kind":"reference","line":7,"end_line":7,"took_reference":true}]}`},
		// under use utf8 a handle's name is runes, not bytes: perl reads <FÉ> as readline, not a glob
		{"utf8readline.pl", `{"ok":true,"prototypes":{},"call_sites":[{"kind":"readline","line":3,"end_line":3}]}`},
		// without use utf8 in force, perl reads <FÉ> as a glob: none at all, one a block ended, one no utf8 turned off
		{"noutf8readline.pl", `{"ok":true,"prototypes":{},"call_sites":[]}`},
		// known limit (facts.go utf8At): a utf8 a BEGIN block imports is not seen, so <FÉ> reports nothing where
		// perl reads it as readline on line 2
		{"beginutf8readline.pl", `{"ok":true,"prototypes":{},"call_sites":[]}`},
		// an empty import list calls neither import nor unimport: use utf8 () leaves <FÉ> a glob, and no utf8 ()
		// leaves use utf8 in force, so perl reads its <FÉ> as readline
		{"emptyutf8readline.pl", `{"ok":true,"prototypes":{},"call_sites":[]}`},
		{"noemptyutf8readline.pl", `{"ok":true,"prototypes":{},"call_sites":[{"kind":"readline","line":3,"end_line":3}]}`},
		// known limit (facts.go utf8At): a utf8 use if imports is not seen, so <FÉ> reports nothing where perl
		// reads it as readline on line 2
		{"ifutf8readline.pl", `{"ok":true,"prototypes":{},"call_sites":[]}`},
		// perl reads <<>> as readline, though parse.IsReadlineTerm calls it a glob
		{"doublediamondreadline.pl", `{"ok":true,"prototypes":{},"call_sites":[{"kind":"readline","line":1,"end_line":1}]}`},
		// a name may open with a digit: perl reads <1> and <$1> as readline
		{"digitreadline.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"readline","line":1,"end_line":1},{"kind":"readline","line":2,"end_line":2}]}`},
		// known limit (facts.go): the file's own r is read before the later import of R's r, so its call is decided
		// on the file's declaration; perl applies the import's \@, reports prototypes {"r":"\\@"} and a \ on line 4
		{"own_then_import.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"line":4,"end_line":4,"name":"r","took_reference":false}]}`},
		// code interpolated in a heredoc body hedges every marker over its statement and the body's lines: perl
		// reports the \ on line 3
		{"heredoc.pl", `{"ok":true,"prototypes":{},"call_sites":[
			{"kind":"reference","line":2,"end_line":4,"took_reference":false,"unresolved":true},
			{"kind":"hash","line":2,"end_line":4,"unresolved":true},
			{"kind":"match","line":2,"end_line":4,"unresolved":true},
			{"kind":"readline","line":2,"end_line":4,"unresolved":true},
			{"kind":"anonhash","line":2,"end_line":4,"unresolved":true}]}`},
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
