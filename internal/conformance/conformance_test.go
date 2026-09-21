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
		// The marker is present, so "no --- source section" would send
		// the author looking for a line that is already there.
		name: "source section present but empty",
		raw:  "--- source\n--- expect parses\n",
		want: "--- source section is empty",
	}, {
		// Whitespace-only is the same mistake with an invisible cause.
		// perl compiles it as an empty program, so without this it is a
		// file that passes while measuring nothing.
		name: "source section only whitespace",
		raw:  "--- source\n\n--- expect parses\n",
		want: "--- source section is empty",
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

// TestParseFileNamelessMarker covers a marker that opens no section.
//
// `--- ` with nothing after it matched the section prefix, trimmed to an
// empty name, and then hit the same guard that skips the leading comment
// block -- so its body was discarded with no error, and every line after
// it was treated as commentary. A file losing SOURCE TEXT silently is
// worse than the contradiction this format already rejects: that lost an
// expectation bit, this loses the program.
//
// It is a plausible typo for `--- source`, and `--- ` is also a legal
// prefix inside Perl source (a heredoc body line, or a printed string), so
// this is reachable by accident in both directions.
func TestParseFileNamelessMarker(t *testing.T) {
	for _, tc := range []struct{ name, raw string }{{
		name: "before a real section",
		raw:  "--- \nmy $x = 1;\n--- source\nmy $y = 2;\n--- expect parses\n",
	}, {
		name: "between sections",
		raw:  "--- source\nmy $x = 1;\n--- \n--- expect parses\n",
	}, {
		name: "at end of file",
		raw:  "--- source\nmy $x = 1;\n--- expect parses\n--- \n",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseFile(tc.raw)
			if err == nil {
				t.Fatalf("ParseFile(%q) = nil error, want a nameless-marker error", tc.raw)
			}
			if want := "no name"; !strings.Contains(err.Error(), want) {
				t.Errorf("ParseFile error = %q, want it to contain %q", err, want)
			}
		})
	}
}

// TestParseFileCRLF keeps a line ending out of the byte-exact field.
//
// Markers match under CRLF because the name is trimmed, so such a file
// parses "successfully" -- but SplitAfter keeps the \r on every line, so
// ExpectOutput becomes "0.5\r\n\r" where perl prints "0.5\n". Run then
// reports CORPUS BUG and blames the author for a line ending.
//
// The corpus's whole claim is byte-exactness against perl, which makes a
// silent CR in that field the worst place for one. Normalised in the
// parser rather than in .gitattributes, because the parser cannot be
// bypassed by a contributor's local git config.
func TestParseFileCRLF(t *testing.T) {
	// The output section is followed by another, as in a real file: the
	// one-newline strip means a section at EOF keeps no trailing newline,
	// which is a property of position rather than of line endings and
	// would otherwise be conflated with the CR here.
	raw := "--- source\r\nmy $x = .5;\r\n--- expect output\r\n0.5\r\n\r\n--- expect parses\r\n"

	f, err := ParseFile(raw)
	if err != nil {
		t.Fatalf("ParseFile on a CRLF file: %v", err)
	}
	if got, want := f.ExpectOutput, "0.5\n"; got != want {
		t.Errorf("ExpectOutput = %q, want %q", got, want)
	}
	if got, want := f.Source, "my $x = .5;\n"; got != want {
		t.Errorf("Source = %q, want %q", got, want)
	}
	if strings.Contains(f.ExpectOutput+f.Source, "\r") {
		t.Error("a carriage return survived into a byte-exact field")
	}
}

// TestParseFileExpectOutputNewline pins what the one-newline strip is for.
//
// It is the BLANK SEPARATOR LINE that carries the output's own trailing
// newline, not the position of the section. A corpus file writes
//
//	--- expect output
//	0.5
//	<blank>
//	--- expect tokens
//
// so the body is "0.5\n\n", the strip removes the separator, and what is
// left is "0.5\n" -- exactly perl's four bytes. Written without the blank
// line, the body is "0.5\n" and the strip leaves "0.5", which no `print
// "$x\n"` ever produces.
//
// This is the rule a corpus author most easily gets wrong, and getting it
// wrong produces a CORPUS BUG report rather than a parse error, so it is
// pinned here and stated in the README.
func TestParseFileExpectOutputNewline(t *testing.T) {
	const head = "--- source\nmy $x = 1;\n--- expect parses\n--- expect output\n"

	withSeparator, err := ParseFile(head + "0.5\n\n--- expect tokens\n")
	if err != nil {
		t.Fatalf("with a blank separator: %v", err)
	}
	if got, want := withSeparator.ExpectOutput, "0.5\n"; got != want {
		t.Errorf("with separator: ExpectOutput = %q, want %q -- this is what perl prints", got, want)
	}

	withoutSeparator, err := ParseFile(head + "0.5\n")
	if err != nil {
		t.Fatalf("without a blank separator: %v", err)
	}
	if got, want := withoutSeparator.ExpectOutput, "0.5"; got != want {
		t.Errorf("without separator: ExpectOutput = %q, want %q", got, want)
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

	// The two checks above only catch drift in ONE direction. Dropping a
	// name from sectionNames() while leaving its case in setSection is
	// never iterated, so nothing above notices -- and because
	// TestReadmeDocumentsEverySection iterates the same list, the omission
	// would silently shrink the README's checked surface too.
	//
	// The count is asserted rather than derived: there is no way to
	// enumerate a switch's cases at run time, so this is the line that has
	// to be updated deliberately when a section is added or removed.
	if got, want := len(sectionNames()), 5; got != want {
		t.Errorf("sectionNames() has %d entries, want %d -- update this "+
			"count and setSection together, or the README test quietly "+
			"stops checking the missing one", got, want)
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

// TestRefusalIssueFullID pins that a whole 36-character id survives.
//
// The id is a UUIDv7, whose first 8 characters are a millisecond
// timestamp: a batch-created chain of issues collides there by
// construction. Measured against the 27 ids in
// docs/plans/2026-09-21-deferred-chain-m1-m2.md, the 8-character prefix
// yields 10 distinct values and the 13-character prefix 23 -- so the
// short form works today only because 16 bits of randomness have not yet
// repeated, and this corpus is designed to grow.
//
// A citation that cannot be looked up unambiguously cannot be VERIFIED,
// which is what the runner needs of it.
func TestRefusalIssueFullID(t *testing.T) {
	const full = "01a0c447-8bd2-7f1e-9a3c-5e4d6b8c9a01"

	f, err := ParseFile("#!perl\n# STATUS refuses as of 0ce515cb. Issue " +
		full + ".\n\n--- source\nmy $x = 1;\n\n--- expect parses\n")
	if err != nil {
		t.Fatal(err)
	}
	if f.Refuses != full {
		t.Errorf("Refuses = %q, want the whole id %q", f.Refuses, full)
	}
}

// TestRefusalIssueShortIDStillWorks keeps the files already in the tree
// parsing.
//
// The short form is what conformance/README.md documented and what every
// existing citation used, so widening the pattern must not narrow it.
func TestRefusalIssueShortIDStillWorks(t *testing.T) {
	const short = "01a0c13f-97f5"

	f, err := ParseFile("#!perl\n# STATUS refuses as of 38c95d23. Issue " +
		short + ".\n\n--- source\nmy $x = 1;\n\n--- expect parses\n")
	if err != nil {
		t.Fatal(err)
	}
	if f.Refuses != short {
		t.Errorf("Refuses = %q, want %q", f.Refuses, short)
	}
}

// TestPerlAdjudicatesFirst pins the order the whole corpus rests on.
//
// A file whose expectation perl does not honour is not measuring the
// parser -- it is wrong. Reporting on our parser first would let a wrong
// expectation read as a refusal, and a refusal is a claim about US. So
// perl is consulted first and a disagreement is reported as CORPUS BUG
// with the parser never consulted at all.
//
// The three disagreements are checked separately because they are three
// different mistakes: claiming a file parses when perl refuses it,
// claiming it does not when perl accepts it, and pinning output perl does
// not print.
func TestPerlAdjudicatesFirst(t *testing.T) {
	for _, tc := range []struct {
		name string
		file *File
		want string
	}{{
		name: "expect parses, but perl refuses",
		file: &File{Source: "my $x = ;\n", ExpectParses: true},
		want: "perl -c refuses it",
	}, {
		name: "expect parsent, but perl accepts",
		file: &File{Source: "my $x = 1;\n", ExpectParsent: true},
		want: "perl -c accepts it",
	}, {
		name: "pinned output perl does not print",
		file: &File{
			Source:       "print \"a\\n\";\n",
			ExpectParses: true,
			ExpectOutput: "b\n",
		},
		want: "pinned output",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			v := verdict(t, tc.file)

			if v.kind != corpusBug {
				t.Fatalf("verdict = %v, want corpusBug", v.kind)
			}
			if len(v.msgs) != 1 {
				// More than one would mean the file's wrongness had been
				// mixed with a claim about our parser, which is the
				// conflation this ordering exists to prevent.
				t.Fatalf("reported %d messages, want exactly 1:\n\t%s",
					len(v.msgs), strings.Join(v.msgs, "\n\t"))
			}
			if !strings.Contains(v.msgs[0], "CORPUS BUG") {
				t.Errorf("message = %q, want it to say CORPUS BUG", v.msgs[0])
			}
			if !strings.Contains(v.msgs[0], tc.want) {
				t.Errorf("message = %q, want it to contain %q", v.msgs[0], tc.want)
			}
		})
	}
}

// TestStaleRefusalMarkerFails is the half of the skip discipline that
// rots.
//
// A file marked `STATUS refuses` that now PASSES must fail loudly. The
// opposite -- letting it keep skipping -- is how a corpus silently stops
// measuring: the construct works, nobody is told, and the marker sits
// there forever claiming a gap that closed.
func TestStaleRefusalMarkerFails(t *testing.T) {
	f := &File{
		Source:       "my $x = 1;\n",
		ExpectParses: true,
		Refuses:      "01a0c13f-97f5-7f98-b32d-07245ec6ddfe",
	}

	v := verdict(t, f)

	if v.kind != staleMarker {
		t.Fatalf("verdict = %v, want staleMarker", v.kind)
	}
	if len(v.msgs) != 1 {
		t.Fatalf("reported %d messages, want 1:\n\t%s",
			len(v.msgs), strings.Join(v.msgs, "\n\t"))
	}
	for _, want := range []string{"PASSES", f.Refuses} {
		if !strings.Contains(v.msgs[0], want) {
			t.Errorf("message = %q, want it to contain %q", v.msgs[0], want)
		}
	}
}

// TestCorpusSkipsAreDocumented makes "skips limited to documented
// refusals" observed rather than asserted.
//
// A skip with no refusal behind it is a test that does not run and does
// not say why, which in a summary line is indistinguishable from a test
// that passed. The only legitimate skip in this suite is a file whose
// refusal is recorded in its own header.
func TestCorpusSkipsAreDocumented(t *testing.T) {
	passing := &File{Source: "my $x = 1;\n", ExpectParses: true}
	if v := verdict(t, passing); v.kind == knownRefusal {
		t.Errorf("a file with no STATUS refuses would skip, saying %q",
			strings.Join(v.msgs, "; "))
	}

	// The converse, without which this would pass against a runner that
	// never skips at all: a genuinely refusing file DOES skip, naming its
	// refusal.
	refusing := &File{
		Source:       "my $x = .5;\n",
		ExpectParses: true,
		Refuses:      "01a0c13f-97f5-7f98-b32d-07245ec6ddfe",
	}
	v := verdict(t, refusing)
	if v.kind != knownRefusal {
		t.Fatalf("verdict = %v, want knownRefusal", v.kind)
	}
	if !strings.Contains(v.reason, refusing.Refuses) {
		t.Errorf("skip reason = %q, want it to name %q", v.reason, refusing.Refuses)
	}

	// Every skip the real corpus produces must carry a refusal too. This
	// is the observation rather than the assertion: it reads the files on
	// disk instead of trusting the rule above to have been followed.
	paths, err := filepath.Glob(filepath.Join(corpusDir, "*", "*.t"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		f, err := ParseFile(string(raw))
		if err != nil {
			t.Errorf("%s: %v", path, err)
			continue
		}
		if v := verdict(t, f); v.kind == knownRefusal && f.Refuses == "" {
			t.Errorf("%s skips with no refusal recorded", path)
		}
	}
}
