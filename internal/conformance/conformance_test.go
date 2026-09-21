// ABOUTME: Runs the graded conformance corpus: parses each .t file's sections
// ABOUTME: and checks parses/parsent, perl output, and declared token facts.
package conformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// corpusDir is the corpus root, relative to this package.
const corpusDir = "../../conformance"

func TestParseFileSections(t *testing.T) {
	path := filepath.Join(corpusDir, "01_literals", "03_leading_decimal.t")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}

	f, err := ParseFile(string(raw))
	if err != nil {
		t.Fatalf("parsing sections: %v", err)
	}

	if got, want := strings.TrimSpace(f.Source), "my $x = .5;\nprint \"$x\\n\";"; got != want {
		t.Errorf("Source = %q, want %q", got, want)
	}
	if !f.ExpectParses {
		t.Error("ExpectParses = false, want true")
	}
	if got, want := f.ExpectOutput, "0.5\n"; got != want {
		t.Errorf("ExpectOutput = %q, want %q", got, want)
	}
	if len(f.TokenFacts) != 2 {
		t.Fatalf("TokenFacts = %d facts, want 2: %q", len(f.TokenFacts), f.TokenFacts)
	}
	if got, want := f.TokenFacts[0], `one numeric literal whose text is ".5"`; got != want {
		t.Errorf("TokenFacts[0] = %q, want %q", got, want)
	}
}

// TestParseFileErrors covers what a malformed corpus file does.
//
// The happy path above reads a real file; these are the cases no corpus
// file should ever be, and each must fail LOUDLY rather than parse into
// something half-formed. A corpus that silently accepts a file asserting
// nothing is a corpus that measures nothing.
func TestParseFileErrors(t *testing.T) {
	// The smallest valid file, and the base every case below mutates.
	const valid = "--- source\nmy $x = 1;\n--- expect parses\n"

	if _, err := ParseFile(valid); err != nil {
		t.Fatalf("the base case must be valid, or every case below tests the wrong thing: %v", err)
	}

	for _, tc := range []struct {
		name string
		raw  string
		want string
	}{{
		name: "no source section",
		raw:  "--- expect parses\n",
		want: "no --- source section",
	}, {
		name: "duplicate source",
		raw:  "--- source\nmy $x = 1;\n--- source\nmy $y = 2;\n--- expect parses\n",
		want: `duplicate section "source"`,
	}, {
		name: "duplicate expect output",
		raw:  valid + "--- expect output\n1\n--- expect output\n2\n",
		want: `duplicate section "expect output"`,
	}, {
		name: "neither parses nor parsent",
		raw:  "--- source\nmy $x = 1;\n",
		want: "asserts neither parses nor parsent",
	}, {
		name: "unknown section",
		raw:  valid + "--- expect vibes\ngood\n",
		want: `unknown section "expect vibes"`,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseFile(tc.raw)
			if err == nil {
				t.Fatalf("ParseFile(%q) = nil error, want %q", tc.raw, tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("ParseFile error = %q, want it to contain %q", err, tc.want)
			}
		})
	}
}

// TestParseFileBothExpectations pins the one case the spike left open.
//
// A file claiming both `--- expect parses` and `--- expect parsent` is
// asserting a contradiction, and the spike accepted it: both fields were
// set and Run consulted ExpectParses first, so parsent was silently
// ignored. Contradictions must be rejected, not resolved by field order.
func TestParseFileBothExpectations(t *testing.T) {
	raw := "--- source\nmy $x = 1;\n--- expect parses\n--- expect parsent\n"

	_, err := ParseFile(raw)
	if err == nil {
		t.Fatal("ParseFile accepted a file asserting both parses and parsent, want an error")
	}
	if want := "both"; !strings.Contains(err.Error(), want) {
		t.Errorf("ParseFile error = %q, want it to contain %q", err, want)
	}
}

// TestReadmeDocumentsEverySection keeps the format's documentation honest.
//
// The README is the corpus's public interface: a parser author reading it
// should not have to read setSection to learn what a file may contain. So
// every section name the code accepts must appear there, and the two rules
// that are not obvious from an example -- any order, no repeats -- must be
// stated rather than implied.
//
// This is a coupling test on purpose. Adding a section to setSection and
// not to the README fails here, which is the only moment anyone would
// notice.
func TestReadmeDocumentsEverySection(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(corpusDir, "README.md"))
	if err != nil {
		t.Fatalf("reading the corpus README: %v", err)
	}
	readme := string(raw)

	for _, section := range sectionNames() {
		if !strings.Contains(readme, "--- "+section) {
			t.Errorf("README does not document the %q section", section)
		}
	}

	// Stated as rules, not left to be inferred from the example.
	for _, rule := range []struct{ name, want string }{
		{"sections may appear in any order", "any order"},
		{"a repeated section is an error", "repeated section"},
		{"exactly one of parses/parsent", "Exactly one"},
	} {
		if !strings.Contains(readme, rule.want) {
			t.Errorf("README does not state that %s (looked for %q)", rule.name, rule.want)
		}
	}
}

// TestSectionNamesMatchParser checks the two lists that must agree.
//
// sectionNames() is what the README is tested against; setSection's switch
// is what a file is actually parsed by. If they drift, the README can
// document a section the parser rejects, or miss one it accepts -- and the
// README test would still pass. This is the check that makes
// sectionNames() a record of the parser rather than a second opinion.
func TestSectionNamesMatchParser(t *testing.T) {
	for _, section := range sectionNames() {
		// A file carrying only this section either parses or fails on a
		// missing-section rule -- never on an unknown-section rule.
		_, err := ParseFile("--- " + section + "\n")
		if err != nil && strings.Contains(err.Error(), "unknown section") {
			t.Errorf("sectionNames() lists %q but the parser rejects it as unknown", section)
		}
	}

	// And the converse: a name not in the list must be rejected.
	if _, err := ParseFile("--- expect vibes\n"); err == nil ||
		!strings.Contains(err.Error(), "unknown section") {
		t.Errorf("parser accepted a section absent from sectionNames(); err = %v", err)
	}
}

// TestCorpus runs every case in the corpus: perl adjudicates the source,
// then our lexer and parser are checked against what perl said.
func TestCorpus(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(corpusDir, "*_*", "*.t"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no corpus files found")
	}

	for _, path := range files {
		name := filepath.ToSlash(strings.TrimPrefix(path, corpusDir+string(filepath.Separator)))
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			f, err := ParseFile(string(raw))
			if err != nil {
				t.Fatalf("bad corpus file: %v", err)
			}
			Run(t, f)
		})
	}
}
