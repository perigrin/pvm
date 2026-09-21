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
