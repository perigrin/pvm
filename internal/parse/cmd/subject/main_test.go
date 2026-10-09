// ABOUTME: Runs the internal/parse subject as the oracle does, a built binary over a file on disk.
// ABOUTME: Every marker is answered from the tree, and the prototypes it reports are the ones the tree carries.

package main

import (
	"context"
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
