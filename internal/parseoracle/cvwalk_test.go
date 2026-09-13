// ABOUTME: Pins the oracle's population: every CV perl compiled for the file, attributed to the file's own statements.
// ABOUTME: A main-program-only count silently omits every prototype-forced reference inside a sub body.

package parseoracle_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// TestOracleCountsReferencesInEveryCV is the defect: `-MO=Concise,-exec` dumps
// PL_main_root only. Named subs, anonymous subs and BEGIN blocks are separate
// CVs, so a prototype-forced reference inside any of them never reached the
// count. Measured with perl 5.42.0, this source has FOUR prototype-driven
// srefgen ops and the main-only oracle reported one.
func TestOracleCountsReferencesInEveryCV(t *testing.T) {
	src := "sub f(\\@){}\n" + // 1
		"sub g { my @a; f(@a) }\n" + // 2  named sub
		"my $c = sub { my @b; f(@b) };\n" + // 3  anonymous sub
		"BEGIN { my @d; f(@d) }\n" + // 4  BEGIN block
		"my @m; f(@m);\n" // 5  main program

	facts, err := parseoracle.Ask(context.Background(), []byte(src), parseoracle.Options{})
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if !facts.OK {
		t.Fatalf("probe must compile, stderr: %s", facts.Stderr)
	}
	if facts.Srefgen != 4 {
		t.Errorf("Srefgen = %d, want 4: one per CV (named sub, anon sub, BEGIN, main)", facts.Srefgen)
	}
	if want := []int{2, 3, 4, 5}; !reflect.DeepEqual(facts.RefLines, want) {
		t.Errorf("RefLines = %v, want %v (the statement line of each reference)", facts.RefLines, want)
	}
	if !facts.Walked {
		t.Error("Walked must report that the optree walk ran")
	}
}

// TestOracleAttributesReferencesToThisFileOnly: widening the population to
// every CV pulls in every sub that t/test.pl defines, because perl's tests
// `require` it into main. Those subs are not the file under measurement, and
// a count that included them would blame every corpus file for test.pl.
func TestOracleAttributesReferencesToThisFileOnly(t *testing.T) {
	dir := t.TempDir()
	helper := "sub f(\\@){}\nsub g { my @a; f(@a) }\n1;\n"
	if err := os.WriteFile(filepath.Join(dir, "helper.pl"), []byte(helper), 0o644); err != nil {
		t.Fatal(err)
	}
	src := "BEGIN { require './helper.pl' }\n" + // 1
		"my @m; f(@m);\n" // 2
	if err := os.WriteFile(filepath.Join(dir, "probe.pl"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	facts, err := parseoracle.AskFile(context.Background(), "probe.pl", parseoracle.Options{Dir: dir})
	if err != nil {
		t.Fatalf("AskFile: %v", err)
	}
	if !facts.OK {
		t.Fatalf("probe must compile, stderr: %s", facts.Stderr)
	}
	if facts.Srefgen != 1 {
		t.Errorf("Srefgen = %d, want 1: the reference inside helper.pl's sub g is not this file's", facts.Srefgen)
	}
	if want := []int{2}; !reflect.DeepEqual(facts.RefLines, want) {
		t.Errorf("RefLines = %v, want %v", facts.RefLines, want)
	}
}

// TestOraclePopulationIsBackslashesAndPrototypes states what a reference site
// IS, so the two sides count the same thing. Each row was measured against
// perl 5.42.0 with -MO=Concise before being written down.
func TestOraclePopulationIsBackslashesAndPrototypes(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want []int
	}{
		// `goto &NAME` is a `goto` whose operand perly.y parsed as an
		// entersub term; op.c's newLOOPEX then wraps it in a REFGEN. That
		// srefgen is goto's rewrite, not a reference taken at a call site,
		// and the source wrote no backslash for it.
		{"goto &sub is not a call-site reference", "sub h{}\nsub s1 { goto &h }\n", nil},
		// `\1` and `\"x"` are folded to a const holding the reference: the
		// backslash was real, the srefgen op is gone. Not a site.
		{"folded constant refs are not sites", "my $a = \\1;\nmy $b = \\\"x\";\n", nil},
		// `\-1` is not folded, and neither is a reference to a variable.
		{"unfolded refs are sites", "my $x;\nmy $a = \\-1;\nmy $b = \\$x;\n", []int{2, 3}},
		// A list-form `\( ... )` is ONE refgen op for the whole list, which
		// is also one backslash in the source. Counting it once keeps the
		// two sides in the same units.
		{"list-form refgen is one site", "my (@a, @b);\nmy @r = \\(@a, @b);\n", []int{2}},
		// A multi-line statement is attributed to the line it starts on,
		// because that is the line perl's nextstate records.
		{"multi-line statement attributes to its first line",
			"sub f(\\@){}\nmy @a;\nf(\n  @a\n);\n", []int{3}},
		// The condition of an elsif has a nextstate of its own.
		{"elsif condition attributes to the elsif line",
			"sub f(\\@){}\nmy @a;\nif (0) { }\nelsif (f(@a)) { }\n", []int{4}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			facts, err := parseoracle.Ask(context.Background(), []byte(c.src), parseoracle.Options{})
			if err != nil {
				t.Fatalf("Ask: %v", err)
			}
			if !facts.OK {
				t.Fatalf("probe must compile, stderr: %s", facts.Stderr)
			}
			if len(facts.RefLines) == 0 && len(c.want) == 0 {
				return
			}
			if !reflect.DeepEqual(facts.RefLines, c.want) {
				t.Errorf("RefLines = %v, want %v", facts.RefLines, c.want)
			}
		})
	}
}
