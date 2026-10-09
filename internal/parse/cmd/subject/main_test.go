// ABOUTME: Runs the internal/parse subject as the oracle does, a built binary over a file on disk.
// ABOUTME: Every marker is answered from the tree, and the prototypes it reports are the ones the tree carries.

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// TestSubjectAnswersAllFiveMarkers runs the subject over a file that makes
// each of the oracle's five parse decisions once, and holds it to answering
// every one -- decided, not hedged -- and to answering both optional
// questions honestly.
//
// The hash is a dereference, `%$r`, because that is the form whose op is
// decidable from text; a named `%h` hedges by design (hashhedge_test.go).
//
// KnowsCallSites and KnowsPrototypes are set by the DECODER from what the
// subject printed, so they are asserted through parseoracle.Subject.Parse,
// the path every measurement takes, rather than by reading main's struct.
//
// The prototypes are the ones the tree carries: a sub this file declares,
// and one a module it uses exports, found the way perl's own test suite finds
// it -- in ../lib from a t/ working directory. An empty map would claim "I
// looked and there are none" about a file declaring two.
func TestSubjectAnswersAllFiveMarkers(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "subject")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("building the subject: %v\n%s", err, out)
	}

	write := func(rel, src string) {
		t.Helper()
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("lib/Shape.pm", "package Shape;\n"+
		"use Exporter 'import';\n"+
		"our @EXPORT = ('one');\n"+
		"sub one ($) { $_[0] }\n"+
		"1;\n")
	write("t/probe.t", "use Shape;\n"+
		"sub twice ($$) { 1 }\n"+
		"my %h = (a => 1);\n"+
		"my $r = \\%h;\n"+
		"my $n = keys %$r;\n"+
		"if ($n =~ /1/) { print one $n }\n"+
		"my $line = <STDIN>;\n"+
		"my $anon = { a => 1 };\n")

	s := parseoracle.Subject{Command: []string{bin}, Dir: filepath.Join(root, "t")}
	facts, err := s.Parse(context.Background(), "probe.t")
	if err != nil {
		t.Fatalf("the subject failed: %v", err)
	}

	if !facts.OK {
		t.Errorf("the subject rejected a valid file")
	}
	if !facts.KnowsCallSites {
		t.Errorf("KnowsCallSites is false: the subject did not answer the call-site question")
	}

	decided := map[string]bool{}
	for _, site := range facts.CallSites {
		if site.Unresolved {
			continue
		}
		if site.Kind == parseoracle.SiteKindReference && !site.TookReference {
			continue
		}
		decided[site.Kind] = true
	}
	for _, kind := range []string{
		parseoracle.SiteKindReference,
		parseoracle.SiteKindHash,
		parseoracle.SiteKindMatch,
		parseoracle.SiteKindReadline,
		parseoracle.SiteKindAnonhash,
	} {
		if !decided[kind] {
			t.Errorf("no decided %s site; the subject reported %+v", kind, facts.CallSites)
		}
	}

	if !facts.KnowsPrototypes {
		t.Fatalf("KnowsPrototypes is false, but the tree carries the file's declarations")
	}
	// perl's spelling, which is prototype()'s: no parens.
	for name, want := range map[string]string{"twice": "$$", "one": "$"} {
		if got, ok := facts.Prototypes[name]; !ok || got != want {
			t.Errorf("Prototypes[%q] = %q (present %v), want %q; all: %v",
				name, got, ok, want, facts.Prototypes)
		}
	}
}

// TestSubjectHonoursLineDirective holds the subject to perl's line numbers,
// which a `#line N` directive renumbers. The oracle reports each site at the
// line its nextstate recorded, so a subject reporting physical lines never
// meets perl past a directive (lib/deprecate.t's `#line 1001`).
//
// Every expectation below is perl 5.42's own, measured with -MO=Concise and
// with the oracle's parse_facts.pl over this exact file: srefgen at
// 2,100,102,104,108,109,203,400. perl's lexer (toke.c S_incline) honours a
// `#line` comment that starts its line, and every line of POD; it ignores
// one indented, one with no number, one inside a string and one after code
// on the same line. A directive naming another file moves the statements
// after it out of this file, and the oracle's CvFILE filter drops them, so
// the subject reports nothing there either.
func TestSubjectHonoursLineDirective(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "subject")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("building the subject: %v\n%s", err, out)
	}
	src := "my @a;\n" +
		"my $x = \\@a;\n" + // 2
		"#line 100\n" +
		"my $y = \\@a;\n" + // 100
		"  #line 500\n" + // indented: not a directive
		"my $z = \\@a;\n" + // 102
		"#line\n" + // no number: not a directive
		"my $w = \\@a;\n" + // 104
		"my $s = \"\n" +
		"#line 900\n" + // inside a string: not a directive
		"\";\n" +
		"my $m = \\@a; #line 700\n" + // 108; after code: not a directive
		"my $v = \\@a;\n" + // 109
		"\n=pod\n\n" +
		"#line 200\n" + // inside POD: a directive
		"\n=cut\n\n" +
		"my $u = \\@a;\n" + // 203
		"#line 300 \"elsewhere.pl\"\n" +
		"my $t = \\@a;\n" + // another file's line 300
		"#line 400 \"probe.t\"\n" +
		"my $q = \\@a;\n" // 400, this file again
	if err := os.WriteFile(filepath.Join(root, "probe.t"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	s := parseoracle.Subject{Command: []string{bin}, Dir: root}
	facts, err := s.Parse(context.Background(), "probe.t")
	if err != nil {
		t.Fatalf("the subject failed: %v", err)
	}

	var got []int
	for _, site := range facts.CallSites {
		if site.Kind != parseoracle.SiteKindReference || site.Unresolved {
			continue
		}
		if site.EndLine != 0 && site.EndLine != site.Line {
			t.Errorf("a one-line statement spans %d-%d", site.Line, site.EndLine)
		}
		got = append(got, site.Line)
	}
	want := []int{2, 100, 102, 104, 108, 109, 203, 400}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("reference sites at %v, want perl's %v", got, want)
	}
}
