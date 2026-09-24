// ABOUTME: The mdtest spike: one topic file makes the same claims six .t files did, checked the same way.
// ABOUTME: Proves the port is a CONTAINER change -- every existing check runs against a case unmodified.
package conformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestMdtestTopicsMakeTheSameClaims runs every topic file through the
// same three checks `TestCorpus` runs over a `.t` file.
//
// THE SPIKE'S QUESTION is whether the existing machinery survives the
// container change. It does: `askPerl` validates the pinned output,
// `parse.Parse` answers the parses bit, and `checkTokenFact` reads the
// token facts unmodified -- because a `Case` carries the same four claims
// a `File` does.
//
// What the container BUYS, measured over tier 04: 34 files become 6
// topics, the pragma pair that could not share a file now shares one,
// and the prose is a document rather than a 40-line comment header
// repeating the tier boilerplate 34 times.
func TestMdtestTopicsMakeTheSameClaims(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(corpusDir, "mdtest", "*.md"))
	if err != nil || len(paths) == 0 {
		t.Skipf("no topic files: %v", err)
	}

	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	total := 0
	for _, path := range paths {
		// FORMAT.md documents the format; it is not a topic. Skipped by
		// NAME rather than by "has no tier line", so a real topic that
		// forgets its tier still fails loudly.
		if filepath.Base(path) == "FORMAT.md" {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		cases, err := ParseTopic(string(raw))
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}

		// The tier is DECLARED by the topic, in the same `**Tier NN
		// name.**` line a reader sees. Read from the file rather than
		// held in a Go map keyed on file name, which is the machinery
		// this migration removes.
		tier, err := topicTier(string(raw))
		if err != nil {
			t.Fatalf("%s: %v", filepath.Base(path), err)
		}
		if len(cases) == 0 {
			t.Errorf("%s holds no cases", filepath.Base(path))
		}

		for _, c := range cases {
			total++
			t.Run(filepath.Base(path)+"/"+c.Title, func(t *testing.T) {
				if c.Source == "" {
					t.Fatalf("case has no source block")
				}

				// Collected rather than reported, because a case that
				// RECORDS a refusal is expected to fail some of these
				// and skips instead -- the shape `run.go` uses over a
				// .t file, and the reason the suite can stay pristine
				// while the corpus names what does not work yet.
				ours := &recorder{}

				// Claim 1: perl agrees with the parses bit and the pin.
				compiles, output := askPerl(t, c.Source)
				switch {
				case c.ExpectParses && !compiles:
					ours.Errorf("case says parses: yes, perl -c refuses it")
				case c.ExpectParsent && compiles:
					ours.Errorf("case says parses: no, perl -c accepts it")
				}
				if c.ExpectOutput != nil && output != *c.ExpectOutput {
					ours.Errorf("pinned %q, perl prints %q", *c.ExpectOutput, output)
				}

				// Claim 2: our parser reaches the same verdict.
				codes := refusalCodes(parse.Parse([]byte(c.Source)))
				if c.ExpectParses && len(codes) > 0 {
					ours.Errorf("parser refuses: %d Unknown node(s), %s",
						len(codes), joinCodes(codes))
				}

				// Claim 3: the token facts, read by the SAME checker the
				// `--- expect tokens` section uses.
				for _, fact := range c.TokenFacts {
					checkTokenFact(ours, fact, []byte(c.Source))
				}

				// Claim 4: the op budget. A case may use only ops its
				// tier or an earlier one introduces, which is what keeps
				// the corpus graded. A `parses: no` case emits no ops --
				// perl builds no optree for a program it will not
				// compile -- so there is nothing to lint.
				if !c.ExpectParsent {
					if err := lintOps(t, c.Source, tier, tiers); err != nil {
						ours.Errorf("%s", err)
					}
				}

				switch {
				case c.Refuses != "" && len(ours.msgs) == 0:
					t.Errorf("case records a refusal (%s) but now PASSES.\n"+
						"\tDrop the `refuses:` line -- a stale record hides "+
						"a regression.", c.Refuses)
				case c.RefusalCode != "" && !hasCode(codes, parse.RefusalCode(c.RefusalCode)) &&
					len(codes) > 0:
					t.Errorf("case names refusal %s, parser refuses with %s.\n"+
						"\tThe CAUSE changed; re-measure it.",
						c.RefusalCode, joinCodes(codes))
				case c.Refuses != "":
					t.Skipf("refuses (%s):\n\t%s", c.Refuses,
						strings.Join(ours.msgs, "\n\t"))
				default:
					for _, m := range ours.msgs {
						t.Errorf("%s", m)
					}
				}
			})
		}
	}
	t.Logf("%d cases across %d topic files", total, len(paths))
}

// TestMdtestIgnoresBlocksItDoesNotOwn is the property that makes the
// corpus shareable.
//
// A topic file holds one Perl program and several IMPLEMENTATIONS'
// answers about it. Chalk and B::SoN are separate implementations in
// other languages, and both go through perl -- B::SoN reads perl's
// optree, Chalk builds on the IR that produces. This parser reads the
// bytes, which is what makes a case it answers correctly evidence that
// Perl can be parsed without perl.
//
// So an ```ir block is another implementation's answer about a program
// we share, reached with none of our code in it. A reader that REJECTED
// it would force every implementation to implement every other one's
// answer, making the corpus a single system again and destroying the
// independence that makes it evidence.
//
// So an unknown tag is skipped, and the case around it still parses.
func TestMdtestIgnoresBlocksItDoesNotOwn(t *testing.T) {
	const topic = "## A case another implementation also answers\n" +
		"\n```perl\nmy $x = 1;\n```\n" +
		"\n```behavior\nparses: yes\n```\n" +
		"\n```ir\n%c1 = Constant(1) :Int\nL: GREEN\n```\n" +
		"\n```tokens\none numeric literal whose text is \"1\"\n```\n"

	cases, err := ParseTopic(topic)
	if err != nil {
		t.Fatalf("a block we do not own must not break the parse: %v", err)
	}
	if len(cases) != 1 {
		t.Fatalf("got %d cases, want 1", len(cases))
	}
	c := cases[0]
	if c.Source != "my $x = 1;\n" {
		t.Errorf("source = %q", c.Source)
	}
	if len(c.TokenFacts) != 1 {
		t.Errorf("token facts = %v, want the one we own", c.TokenFacts)
	}
	if !c.ExpectParses {
		t.Errorf("parses bit lost")
	}
}

// TestMdtestSeparatesEmptyOutputFromAbsentOutput keeps `File`'s
// distinction.
//
// Pinning EMPTY output and pinning NOTHING are different claims, which
// `File.ExpectOutput` is a pointer to express. A `key: value` block cannot
// spell an empty value unambiguously, so output is a FENCED block: an
// empty one pins empty, an absent one pins nothing, and this test is why.
func TestMdtestSeparatesEmptyOutputFromAbsentOutput(t *testing.T) {
	withEmpty, err := ParseTopic("## x\n```output\n```\n")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if withEmpty[0].ExpectOutput == nil {
		t.Errorf("an empty output block must PIN the empty string, not leave it absent")
	} else if *withEmpty[0].ExpectOutput != "" {
		t.Errorf("empty output block pinned %q", *withEmpty[0].ExpectOutput)
	}

	absent, err := ParseTopic("## x\n```behavior\nparses: yes\n```\n")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if absent[0].ExpectOutput != nil {
		t.Errorf("a case with no output block must pin nothing, got %q",
			*absent[0].ExpectOutput)
	}
}
