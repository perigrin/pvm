// ABOUTME: Runs the graded conformance corpus: every case AllCases reads
// ABOUTME: is checked for parses/parsent, perl output, and declared token facts.
package conformance

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// corpusDir is the corpus root, relative to this package.
const corpusDir = "../../conformance"

// pin builds a File.ExpectOutput for a test constructing one by hand,
// since a string literal has no address.
func pin(s string) *string { return &s }

// TestParseFileSections reads all five sections off one whole file.
//
// THE INPUT USED TO BE A CORPUS FILE ON DISK, `01_literals/06_leading_
// decimal.t`, and is now the bytes that file held. The corpus moved to
// mdtest topics, which `ParseFile` does not read and never will -- so
// pointing this at a corpus path again would mean reading a topic with
// the wrong parser. What is asserted is unchanged, down to the token
// facts, because the claim was always about `ParseFile` rather than
// about that file: a whole file with every section in it parses into
// exactly these fields.
//
// The one thing that cannot survive is the coupling. This no longer
// fails when a corpus file changes shape, because no corpus file is in
// this format any more. `TestCorpus` is what reads the corpus now.
func TestParseFileSections(t *testing.T) {
	// `--- expect parses` before `--- expect output` is deliberate: it is
	// a bodiless section followed by another marker, which is the shape
	// TestParseFileMarkerNeedsBlankLineBefore's third case exists for.
	const raw = "#!perl\n# TIER 01 literals\n\n" +
		"--- source\nmy $x = .5;\nprint \"$x\\n\";\n\n" +
		"--- expect parses\n\n" +
		"--- expect output\n0.5\n\n" +
		"--- expect tokens\n" +
		"one numeric literal whose text is \".5\"\n" +
		"no operator whose text is \".\"\n"

	f, err := ParseFile(raw)
	if err != nil {
		t.Fatalf("parsing sections: %v", err)
	}

	if got, want := strings.TrimSpace(f.Source), "my $x = .5;\nprint \"$x\\n\";"; got != want {
		t.Errorf("Source = %q, want %q", got, want)
	}
	if !f.ExpectParses {
		t.Error("ExpectParses = false, want true")
	}
	if f.ExpectOutput == nil {
		t.Fatal("ExpectOutput = nil, want a pin: the file has an `--- expect output` section")
	}
	if got, want := *f.ExpectOutput, "0.5\n"; got != want {
		t.Errorf("*ExpectOutput = %q, want %q", got, want)
	}

	// Pinned-but-empty and unpinned are DIFFERENT states, and a nil
	// pointer is what tells them apart: a plain string made both read as
	// "", so the runner skipped the check on a file whose whole claim was
	// that it prints nothing. No sentinel string can do this, because a
	// corpus file may legitimately pin any string at all.
	empty, err := ParseFile("--- source\nmy $x = 1;\n\n--- expect output\n\n--- expect parses\n")
	if err != nil {
		t.Fatalf("parsing a file pinning empty output: %v", err)
	}
	if empty.ExpectOutput == nil || *empty.ExpectOutput != "" {
		t.Errorf("ExpectOutput = %v, want a pin of the empty string", empty.ExpectOutput)
	}
	absent, err := ParseFile("--- source\nmy $x = 1;\n\n--- expect parses\n")
	if err != nil {
		t.Fatalf("parsing a file with no expect-output section: %v", err)
	}
	if absent.ExpectOutput != nil {
		t.Errorf("ExpectOutput = %q, want nil for a file with no such section", *absent.ExpectOutput)
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
// The happy path above is a whole well-formed file; these are the cases
// no corpus file should ever be, and each must fail LOUDLY rather than
// parse into something half-formed. A corpus that silently accepts a
// file asserting nothing is a corpus that measures nothing.
func TestParseFileErrors(t *testing.T) {
	// The smallest valid file, and the base every case below mutates.
	const valid = "--- source\nmy $x = 1;\n\n--- expect parses\n\n"

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
		raw:  "--- source\nmy $x = 1;\n\n--- source\nmy $y = 2;\n\n--- expect parses\n",
		want: `duplicate section "source"`,
	}, {
		name: "duplicate expect output",
		raw:  valid + "--- expect output\n1\n\n--- expect output\n2\n",
		want: `duplicate section "expect output"`,
	}, {
		name: "neither parses nor parsent",
		raw:  "--- source\nmy $x = 1;\n",
		want: "asserts neither parses nor parsent",
	}, {
		// The marker is present, so "no --- source section" would send
		// the author looking for a line that is already there.
		name: "source section present but empty",
		raw:  "--- source\n\n--- expect parses\n",
		want: "--- source section is empty",
	}, {
		// Whitespace-only is the same mistake with an invisible cause.
		// perl compiles it as an empty program, so without this it is a
		// file that passes while measuring nothing.
		name: "source section only whitespace",
		raw:  "--- source\n \n\n--- expect parses\n",
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
		raw:  "--- \nmy $x = 1;\n\n--- source\nmy $y = 2;\n\n--- expect parses\n",
	}, {
		name: "between sections",
		raw:  "--- source\nmy $x = 1;\n\n--- \n--- expect parses\n",
	}, {
		name: "at end of file",
		raw:  "--- source\nmy $x = 1;\n\n--- expect parses\n--- \n",
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

// TestParseFileMarkerNeedsBlankLineBefore covers a marker inside a heredoc.
//
// A heredoc body, POD block or `__DATA__` section is opaque to the section
// split, so a line inside one beginning `--- ` was read as a real marker.
// The two spellings behaved differently and only one was safe: `--- not a
// marker` was rejected as an unknown section, while `--- expect output` was
// accepted SILENTLY -- the source truncated at that line, leaving an
// unterminated heredoc opener, and the rest of the program became the
// pinned output. A valid Perl program was mis-read as a differently-shaped
// corpus file with nothing said.
//
// The fix is the blank line: a marker must be the only content on its line
// AND be preceded by a blank line or start of file. Measured across all 134
// corpus files at dc1bea2c, every one of the 472 markers already satisfies
// this, so it costs the existing corpus nothing. It is a weak check by
// construction -- a heredoc body CAN contain a blank line followed by
// `--- expect output` -- but it turns the common accident loud, which is
// what was missing. See conformance/13_opaque/README.md.
func TestParseFileMarkerNeedsBlankLineBefore(t *testing.T) {
	// The dangerous case: the marker names a REAL section, so nothing
	// downstream objects. Without the rule this parses cleanly and pins
	// "END\nprint $t;" as the expected output of a truncated program.
	heredoc := "--- source\nmy $t = <<'END';\n--- expect output\nEND\nprint $t;\n\n--- expect parses\n"

	if _, err := ParseFile(heredoc); err == nil {
		t.Fatal("ParseFile accepted a `--- expect output` line inside a heredoc body, " +
			"want an error: the source truncates there and the rest is read as a pin")
	} else if want := "blank line"; !strings.Contains(err.Error(), want) {
		t.Errorf("ParseFile error = %q, want it to contain %q", err, want)
	}

	// The legitimate shapes must keep parsing. A marker at the very start
	// of a file has no preceding line at all, which is start-of-file
	// rather than a violation -- though no real corpus file is shaped that
	// way, every ParseFile test in this file is.
	for _, tc := range []struct{ name, raw string }{{
		name: "marker at start of input",
		raw:  "--- source\nmy $x = 1;\n\n--- expect parses\n",
	}, {
		name: "marker after a comment block",
		raw:  "#!perl\n# TIER 01_literals\n\n--- source\nmy $x = 1;\n\n--- expect parses\n",
	}, {
		// A marker DIRECTLY after another, with no blank line between
		// them. A section with an empty body is ordinary -- `--- expect
		// parses` carries nothing at all -- so the preceding marker is
		// what makes this line openable, not a blank line. Written with
		// the two markers genuinely adjacent: with a blank line between
		// them it would pass under a rule that forbade this, which is
		// what the first draft of this case did.
		name: "consecutive markers, the earlier one bodiless",
		raw:  "--- source\nmy $x = 1;\n\n--- expect parses\n--- expect output\n1\n\n--- expect tokens\n",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ParseFile(tc.raw); err != nil {
				t.Errorf("ParseFile(%q) = %v, want it to parse", tc.raw, err)
			}
		})
	}
}

// TestParseFileBodyRunsToTheNextMarker pins the blank separator as content.
//
// A section body runs to the next `--- ` marker INCLUDING the blank
// separator line before it. For most sections that line is trailing
// whitespace in a Perl program and nobody notices; inside a `__DATA__`
// section it is DATA, so a two-line data section prints THREE lines.
//
// THIS WAS PINNED AGAINST `13_opaque/08_data_section.t`, WHICH WAS THE
// ONE PLACE IN THE CORPUS THE PROPERTY WAS VISIBLE -- and that is the
// coupling the topic format dissolved. An mdtest case's ```perl block is
// TRIMMED and its ```output block is a fence rather than a run-to-marker
// body, so no separator line can reach the data section: the topic case
// pins `one\ntwo\n`, two lines, where the `.t` file pinned three. The
// property is not wrong, it simply stopped having a witness in the
// corpus, so the bytes move here.
//
// That matters more than it sounds, not less: the property now has NO
// reader outside this test, so if `ParseFile` terminated a body at the
// blank line instead of the marker, this is the only thing that would
// say so. It still earns its place for the reason it always did -- that
// change would satisfy every other test in this file.
func TestParseFileBodyRunsToTheNextMarker(t *testing.T) {
	// The bytes `13_opaque/08_data_section.t` carried, minus its header
	// prose. The blank line before `--- expect output` is the whole
	// point and is why this is not written as a tidier fixture.
	const raw = "--- source\nmy @lines = <DATA>;\nprint @lines;\n__DATA__\none\ntwo\n\n" +
		"--- expect parses\n\n" +
		"--- expect output\none\ntwo\n\n\n" +
		"--- expect tokens\n" +
		"one readline operator whose text is \"<DATA>\"\n" +
		"one data section whose text is \"__DATA__\\none\\ntwo\\n\\n\"\n"

	f, err := ParseFile(raw)
	if err != nil {
		t.Fatalf("parsing the data-section file: %v", err)
	}

	// The trailing "\n\n" is the point: one newline ends `two`, the other
	// is the format's own separator line, which perl reads as a third
	// (empty) line of DATA.
	if want := "__DATA__\none\ntwo\n\n"; !strings.HasSuffix(f.Source, want) {
		t.Errorf("Source = %q, want it to end with %q -- the blank separator "+
			"line is part of the data section", f.Source, want)
	}
	if f.ExpectOutput == nil {
		t.Fatal("ExpectOutput = nil, want the three-line pin")
	}
	if got, want := *f.ExpectOutput, "one\ntwo\n\n"; got != want {
		t.Errorf("ExpectOutput = %q, want %q -- three lines for a two-line "+
			"data section, which is what perl prints", got, want)
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
	raw := "--- source\r\nmy $x = .5;\r\n\r\n--- expect output\r\n0.5\r\n\r\n--- expect parses\r\n"

	f, err := ParseFile(raw)
	if err != nil {
		t.Fatalf("ParseFile on a CRLF file: %v", err)
	}
	if f.ExpectOutput == nil {
		t.Fatal("ExpectOutput = nil, want a pin")
	}
	if got, want := *f.ExpectOutput, "0.5\n"; got != want {
		t.Errorf("ExpectOutput = %q, want %q", got, want)
	}
	// The source body runs to the next marker and so includes the blank
	// separator line before it -- the same property 13_opaque's
	// 08_data_section.t pins, where that blank line is DATA. Asserted with
	// it present rather than trimmed away, because a CR hiding in the
	// separator is exactly the byte this test exists to catch.
	if got, want := f.Source, "my $x = .5;\n\n"; got != want {
		t.Errorf("Source = %q, want %q", got, want)
	}
	if strings.Contains(*f.ExpectOutput+f.Source, "\r") {
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
	const head = "--- source\nmy $x = 1;\n\n--- expect parses\n--- expect output\n"

	withSeparator, err := ParseFile(head + "0.5\n\n--- expect tokens\n")
	if err != nil {
		t.Fatalf("with a blank separator: %v", err)
	}
	if withSeparator.ExpectOutput == nil {
		t.Fatal("with a blank separator: ExpectOutput = nil, want a pin")
	}
	if got, want := *withSeparator.ExpectOutput, "0.5\n"; got != want {
		t.Errorf("with separator: ExpectOutput = %q, want %q -- this is what perl prints", got, want)
	}

	withoutSeparator, err := ParseFile(head + "0.5\n")
	if err != nil {
		t.Fatalf("without a blank separator: %v", err)
	}
	if withoutSeparator.ExpectOutput == nil {
		t.Fatal("without a blank separator: ExpectOutput = nil, want a pin")
	}
	if got, want := *withoutSeparator.ExpectOutput, "0.5"; got != want {
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
	raw := "--- source\nmy $x = 1;\n\n--- expect parses\n--- expect parsent\n"

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
		// The one rule an author cannot infer from a well-formed example,
		// because every example already satisfies it. Its absence from
		// the README is how the next `--- expect output` inside a heredoc
		// gets written.
		{"a marker needs a blank line before it", "blank line before it"},
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
	cases, err := AllCases(corpusDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		t.Run(c.Key, func(t *testing.T) {
			Run(t, c.File)
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
			ExpectOutput: pin("b\n"),
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
	//
	// Indirect object syntax rather than the lexical sub this used to carry,
	// which replaced `try`/`catch`, `defer { ... }`, `undef @a;` and before
	// that `my $x = .5;`, for the reason `TestRefusalCodeMismatchFails`
	// records: a fixture whose premise is a refusal has to be replaced every
	// time that refusal is fixed. The decimal stopped refusing under
	// 01a0c13f-97f5, the unary `undef` under 01a0dd43-bc9e, `defer` under
	// 01a0d087-28dd when `WORD BLOCK` became readable, `try` under
	// 01a0de43-ff83 when parseTry landed, and `my sub e ($);` under
	// 01a0ea2e-9771 -- each caught by this assertion.
	//
	// `new Foo "a"` is valid perl -- `perl -c` says syntax OK on 5.42.0,
	// outside the 5.36 bundle that disables indirect calls. 01a0ddc5-b591
	// read indirect notation on a class this parse KNOWS; `Foo` here is not
	// one, and deciding it needs the symbol table the parser has only in
	// part, which is why it is the longest-lived candidate available. 01a0ebb0-6329
	// owns what is left.
	//
	// The issue cited is the one that still records this one.
	refusing := &File{
		Source:       "my $x = new Foo \"a\";\n",
		ExpectParses: true,
		Refuses:      "01a0ebb0-6329-7a13-b398-a5120ec9bbe8",
	}
	v := verdict(t, refusing)
	if v.kind != knownRefusal {
		t.Fatalf("verdict = %v, want knownRefusal", v.kind)
	}
	if !strings.Contains(v.reason, refusing.Refuses) {
		t.Errorf("skip reason = %q, want it to name %q", v.reason, refusing.Refuses)
	}

	// Every skip the real corpus produces must carry a refusal too. This
	// is the observation rather than the assertion: it reads the cases on
	// disk instead of trusting the rule above to have been followed.
	cases, err := AllCases(corpusDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		if v := verdict(t, c.File); v.kind == knownRefusal && c.Refuses == "" {
			t.Errorf("%s skips with no refusal recorded", c.Key)
		}
	}
}

// TestRefusalCitationMustResolve checks that every issue a corpus file
// cites still exists.
//
// A citation can outlive its issue -- which nearly happened during a
// chain prune that destroyed one ref. A refusal pointing at nothing is a
// skip with no explanation behind it: the file still claims a known gap,
// but the record of what the gap IS is gone.
func TestRefusalCitationMustResolve(t *testing.T) {
	// The resolver, checked against ids whose existence is known, so this
	// proves the lookup and not the corpus. The fabricated id is a real
	// one with its last character bumped: same shape, same length, so a
	// resolver that merely pattern-matches would pass it.
	t.Run("resolver", func(t *testing.T) {
		requireZhi(t)

		// A live id, ASKED FOR rather than named. Naming one couples this
		// fixture to tracker state that moves underneath it: the first
		// spelling was `01a0c13f-97f5`, which went done; the second was
		// `01a0d087-28dd`, which went done the moment the last corpus
		// refusal was implemented -- and this subtest then asserted that a
		// CLOSED issue resolves, which is the exact thing the sibling
		// subtest below exists to forbid.
		//
		// A fixture that has to be remembered will be wrong. Deriving one
		// costs a second `git-zhi` call and cannot go stale.
		real, err := anyLiveIssue()
		if err != nil {
			t.Skipf("no live issue to check the resolver against: %v", err)
		}
		if err := citationResolves(real); err != nil {
			t.Errorf("the live id %s did not resolve: %v", real, err)
		}

		const fake = "01a0c13f-97f5-7f98-b32d-07245ec6ddff"
		err = citationResolves(fake)
		if err == nil {
			t.Fatalf("fabricated id %s resolved", fake)
		}
		if !strings.Contains(err.Error(), fake) {
			t.Errorf("error = %q, want it to name the id %q", err, fake)
		}
	})

	// selfRecorded is a complete citation, not an id, and looking it up
	// would fail every corpus file that uses it. Checked without
	// requireZhi because it must not reach git-zhi at all.
	t.Run("this file is not looked up", func(t *testing.T) {
		if err := citationResolves(selfRecorded); err != nil {
			t.Errorf("%q was looked up: %v", selfRecorded, err)
		}
	})

	// A CLOSED issue is not a live record, and existence was never the
	// question. Three corpus cases were found citing done issues from a
	// finished milestone -- `01a0c730`, `01a0c432` and `01a0c35f`, all
	// corpus-CONSTRUCTION issues that never owned a parser gap. Each
	// resolved, so this test passed them while the thing it exists to
	// prevent had already happened: a live refusal whose record says the
	// work is finished.
	//
	// The id below is `01a0c730`, measured done in milestone
	// m3-conformance-corpus. If it is ever reopened this test will fail
	// and want a different done id, which is the correct failure: a
	// fixture pinned to real tracker state.
	t.Run("a done issue does not resolve", func(t *testing.T) {
		requireZhi(t)

		const done = "01a0c730-b241-7765-aaa2-5260d050dce9"
		err := citationResolves(done)
		if err == nil {
			t.Fatalf("done issue %s resolved; a closed record is not a live one", done)
		}
		if !strings.Contains(err.Error(), done) {
			t.Errorf("error = %q, want it to name the id %q", err, done)
		}
	})

	// The same claim in the topic format's spelling. Sixteen cases carry
	// it, and every one was a `.t` file that said `refuses as of this
	// file` -- so a resolver that knew only the old phrase would have
	// reported all sixteen as citations pointing at nothing, which is a
	// claim about the corpus that is not true.
	t.Run("unfiled is not looked up", func(t *testing.T) {
		if err := citationResolves(unfiledRefusal); err != nil {
			t.Errorf("%q was looked up: %v", unfiledRefusal, err)
		}
	})

	// The corpus itself: every id it cites, resolved.
	t.Run("corpus", func(t *testing.T) {
		requireZhi(t)

		cases, err := AllCases(corpusDir)
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range cases {
			if err := citationResolves(c.Refuses); err != nil {
				t.Errorf("%s: %v", c.Key, err)
			}
		}
	})
}

// requireZhi skips when the tracker is not installed.
//
// The check needs git-zhi; a machine without it cannot answer the
// question, and failing there would make the suite depend on a tool that
// is not a build dependency. So it SKIPS, loudly and by name, rather than
// passing vacuously -- a check that silently reports success when it
// could not run is worse than no check, because it converts "unknown" to
// "fine".
func requireZhi(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath(zhiBinary); err != nil {
		t.Skipf("%s is not installed, so refusal citations cannot be resolved", zhiBinary)
	}
}

// anyLiveIssue returns the id of some open issue, for a fixture that needs one
// and must not name one.
//
// Naming a live id couples the fixture to state that moves: two spellings have
// already gone `done` underneath this test, the second on the very commit that
// implemented the last corpus refusal. `git-zhi issue list` omits closed
// issues, so its first entry is live by construction.
//
// Deliberately NOT filtered to a milestone or a title. The claim being tested
// is "the resolver accepts a live id", and any live id demonstrates it; a
// narrower pick would be another thing to keep in step.
func anyLiveIssue() (string, error) {
	out, err := exec.Command(zhiBinary, "issue", "list", "--format", "json").Output()
	if err != nil {
		return "", fmt.Errorf("listing issues: %w", err)
	}
	var issues []struct {
		ID    string `json:"id"`
		State string `json:"state"`
	}
	if err := json.Unmarshal(out, &issues); err != nil {
		return "", fmt.Errorf("reading %s output: %w", zhiBinary, err)
	}
	for _, i := range issues {
		if !closedStates[i.State] && i.ID != "" {
			return i.ID, nil
		}
	}
	return "", fmt.Errorf("%s lists no open issue", zhiBinary)
}

// TestEmptyExpectedOutputIsChecked pins that a pinned EMPTY output is an
// assertion rather than an absence.
//
// A file whose whole point is that a construct prints NOTHING has to be
// able to say so. While ExpectOutput was a plain string, such a file was
// indistinguishable in the struct from one with no `--- expect output`
// section at all, so the check was skipped and perl could print anything.
func TestEmptyExpectedOutputIsChecked(t *testing.T) {
	// The blank line is the separator the one-newline strip removes, so
	// the pinned output is the empty string rather than "\n".
	raw := "--- source\nprint \"a\\n\";\n\n--- expect output\n\n--- expect parses\n"

	f, err := ParseFile(raw)
	if err != nil {
		t.Fatalf("parsing a file that pins empty output: %v", err)
	}
	if f.ExpectOutput == nil {
		t.Fatal("ExpectOutput = nil for a file WITH an `--- expect output` section")
	}
	if got := *f.ExpectOutput; got != "" {
		t.Fatalf("*ExpectOutput = %q, want the empty string", got)
	}

	// The source prints "a\n" against a pin of nothing, so the file is
	// wrong and the runner must say so rather than skipping the check.
	v := verdict(t, f)
	if v.kind != corpusBug {
		t.Fatalf("verdict = %v, want corpusBug: the file pins empty output and perl prints %q",
			v.kind, "a\n")
	}
	if len(v.msgs) != 1 || !strings.Contains(v.msgs[0], "pinned output") {
		t.Errorf("messages = %q, want exactly one naming the pinned output", v.msgs)
	}
}

// TestAbsentExpectedOutputSkipsCheck pins the other half of the same
// distinction: no section means no claim, which is the normal parsent
// case, and must stay a skip rather than becoming an assertion that perl
// prints nothing.
func TestAbsentExpectedOutputSkipsCheck(t *testing.T) {
	raw := "--- source\nprint \"a\\n\";\n\n--- expect parses\n"

	f, err := ParseFile(raw)
	if err != nil {
		t.Fatalf("parsing a file with no expect-output section: %v", err)
	}
	if f.ExpectOutput != nil {
		t.Fatalf("ExpectOutput = %q for a file with NO `--- expect output` section, want nil",
			*f.ExpectOutput)
	}

	v := verdict(t, f)
	if v.kind == corpusBug {
		t.Fatalf("verdict = corpusBug (%q), want the output check skipped entirely", v.msgs)
	}
}

// TestRefusalCodeMismatchFails: a file that names the code it waits on
// must FAIL when the parser refuses for a different reason.
//
// Skipping would be wrong, and the distinction is the issue's whole
// point. A skip says "this gap is known and recorded"; a file whose
// refusal has changed cause is no longer measuring what its header
// documents, and the record behind the skip now describes something
// else. That is the same failure mode as a stale marker -- a corpus that
// quietly stops measuring -- so it gets the same treatment.
//
// A file that names NO code keeps skipping. Seventeen files in the tree
// predate codes and must keep working; a code is an additional promise,
// not a new requirement.
func TestRefusalCodeMismatchFails(t *testing.T) {
	// `new Foo "a"`, indirect object syntax on a class this parse does not
	// know: a method call perl decides from its symbol table. Valid perl --
	// `perl -c` says syntax OK on 5.42.0 -- so it reaches the parser rather
	// than stopping at perl's adjudication. 01a0ebb0-6329 owns it; 01a0ddc5-b591 read
	// the half a static parser can.
	//
	// It is the SIXTH fixture here, and each replacement is the fixture
	// guarding itself: a test whose premise is that something refuses has
	// to notice when it stops. `my $x = .5;` stopped refusing when
	// 01a0c13f-97f5 taught the lexer that a `.` before a digit starts a
	// number in term position; `undef @a;` stopped when 01a0dd43-bc9e added
	// `undef` to `parse.namedUnary`; `defer { ... }` stopped when 01a0d087-28dd
	// made `WORD BLOCK` a statement; `try { 1 } catch ($e) { 2 }` stopped
	// when 01a0de43-ff83 added parseTry; `my sub e ($);` stopped when
	// 01a0ea2e-9771 read lexical subs. The `t.Fatalf` below reported all
	// five.
	//
	// The code is READ from the parse rather than written here: which
	// site declines it is the parser's business and may change, and a
	// literal in this test would then assert the old answer.
	const src = "my $x = new Foo \"a\";\n"

	const wrong = parse.RefusalCode("missing_operand")

	actual := refusalCodes(parse.Parse([]byte(src)))
	if len(actual) == 0 || actual[0] == "" {
		t.Fatalf("%q no longer refuses with a code; replace this fixture", src)
	}
	if hasCode(actual, wrong) {
		t.Fatalf("%q now refuses with %s, which this test uses as the "+
			"WRONG code; pick another", src, wrong)
	}

	t.Run("mismatch fails", func(t *testing.T) {
		f := &File{
			Source:       src,
			ExpectParses: true,
			Refuses:      "01a0d087-28dd-711f-a0d5-54cb8515910c",
			RefusalCode:  wrong,
		}
		v := verdict(t, f)
		if v.kind != staleMarker {
			t.Fatalf("verdict = %v, want staleMarker: a changed cause is a "+
				"marker that no longer describes the file", v.kind)
		}
		if len(v.msgs) != 1 {
			t.Fatalf("reported %d messages, want 1:\n\t%s",
				len(v.msgs), strings.Join(v.msgs, "\n\t"))
		}
		// Both codes must appear: the one the file claimed and the one it
		// actually got. A message naming only one leaves the reader to
		// guess which half moved.
		for _, want := range []string{string(wrong), string(actual[0])} {
			if !strings.Contains(v.msgs[0], want) {
				t.Errorf("message = %q, want it to name %q", v.msgs[0], want)
			}
		}
	})

	t.Run("match still skips", func(t *testing.T) {
		f := &File{
			Source:       src,
			ExpectParses: true,
			Refuses:      "01a0d087-28dd-711f-a0d5-54cb8515910c",
			RefusalCode:  actual[0],
		}
		if v := verdict(t, f); v.kind != knownRefusal {
			t.Errorf("verdict = %v, want knownRefusal: the file names the "+
				"code it actually has", v.kind)
		}
	})

	t.Run("no code still skips", func(t *testing.T) {
		f := &File{
			Source:       src,
			ExpectParses: true,
			Refuses:      "01a0d087-28dd-711f-a0d5-54cb8515910c",
		}
		if v := verdict(t, f); v.kind != knownRefusal {
			t.Errorf("verdict = %v, want knownRefusal: the seventeen files "+
				"predating codes must keep working", v.kind)
		}
	})
}

// TestRefusalCodeParsesFromHeader: the code round-trips through
// ParseFile.
//
// The code attaches to the STATUS line that already carries the refusal,
// rather than to a new section. A refusal is one fact -- what does not
// work, why it is known, and now which site declines -- and splitting it
// across two places would let the halves disagree.
func TestRefusalCodeParsesFromHeader(t *testing.T) {
	const body = "\n\n--- source\nmy $x = 1;\n\n--- expect parses\n"

	t.Run("named", func(t *testing.T) {
		f, err := ParseFile("#!perl\n# STATUS refuses as of 0ce515cb. " +
			"Issue 01a0c13f-97f5-7f98-b32d-07245ec6ddfe. " +
			"Refusal unimplemented_statement." + body)
		if err != nil {
			t.Fatal(err)
		}
		if f.RefusalCode != "unimplemented_statement" {
			t.Errorf("RefusalCode = %q, want %q",
				f.RefusalCode, "unimplemented_statement")
		}
		if f.Refuses != "01a0c13f-97f5-7f98-b32d-07245ec6ddfe" {
			t.Errorf("Refuses = %q; the code must not eat the issue id",
				f.Refuses)
		}
	})

	t.Run("absent", func(t *testing.T) {
		f, err := ParseFile("#!perl\n# STATUS refuses as of this file." + body)
		if err != nil {
			t.Fatal(err)
		}
		if f.RefusalCode != "" {
			t.Errorf("RefusalCode = %q, want empty for a file naming none",
				f.RefusalCode)
		}
		if f.Refuses != selfRecorded {
			t.Errorf("Refuses = %q, want %q", f.Refuses, selfRecorded)
		}
	})

	t.Run("not refusing", func(t *testing.T) {
		f, err := ParseFile("#!perl\n# an ordinary header." + body)
		if err != nil {
			t.Fatal(err)
		}
		if f.RefusalCode != "" {
			t.Errorf("RefusalCode = %q, want empty for a passing file",
				f.RefusalCode)
		}
	})
}
