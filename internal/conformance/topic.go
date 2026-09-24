// ABOUTME: Reads Chalk's mdtest corpus format -- `##` per case, fenced blocks by language tag.
// ABOUTME: Each block is one IMPLEMENTATION's answer about a shared Perl program; each reads its own and ignores the rest.
package conformance

import (
	"fmt"
	"regexp"
	"strings"

	"tamarou.com/pvm/internal/parse"
)

// Case is one `##` section of an mdtest topic file.
//
// The fields are deliberately the same claims `File` carries, because the
// format change is a CONTAINER change: what a case asserts about Perl is
// unchanged, only how many of them live in a file and how the prose around
// them is written.
type Case struct {
	// Title is the `##` heading, which names the case the way a file name
	// used to.
	Title string

	// Source is the ```perl block.
	Source string

	// ExpectParses and ExpectParsent come from the ```behavior block's
	// `parses:` line; ExpectOutput comes from a separate ```output block.
	//
	// ExpectOutput is the GROUND TRUTH rather than a claim of ours -- it
	// is what perl prints, which no implementation reading this corpus
	// gets to define and all of them are checked against. A pointer for
	// `File`'s reason: pinning empty output and pinning nothing are
	// different claims.
	ExpectOutput  *string
	ExpectParses  bool
	ExpectParsent bool

	// TokenFacts are the ```tokens block's lines, in GLOSSARY.md's
	// vocabulary exactly as the `--- expect tokens` section carried them.
	TokenFacts []string

	// Refuses and RefusalCode come from a `refuses:` / `refusal:` pair in
	// the behavior block, replacing the `# STATUS refuses` header line.
	Refuses     string
	RefusalCode string

	// Line is the heading's 1-based line, so a failure can be reported at
	// a position a reader can open.
	Line int
}

// reTopicTier reads the tier a topic declares, from the `**Tier 04
// operators.**` line its own prose carries.
//
// DECLARED, not derived. A tier cannot be computed from a topic's
// contents -- the optimiser erases the very construct a case is about,
// which is why `lintOps` checks a declaration rather than making one.
// What changed with topics is only WHERE the declaration lives: in the
// document a reader reads, rather than in a Go map keyed on file name
// that a reader has to be told about separately.
var reTopicTier = regexp.MustCompile(`(?m)^\*\*Tier (\d\d) (\w+)`)

// topicTier returns the corpus tier directory a topic declares.
func topicTier(raw string) (string, error) {
	m := reTopicTier.FindStringSubmatch(raw)
	if m == nil {
		return "", fmt.Errorf("topic declares no tier; " +
			"every topic needs a `**Tier NN name.**` line")
	}
	return m[1] + "_" + m[2], nil
}

var (
	reHeading = regexp.MustCompile(`^##\s+(.+?)\s*$`)
	reFence   = regexp.MustCompile("^```(\\w*)\\s*$")
	reKeyVal  = regexp.MustCompile(`^\s*([\w-]+)\s*:\s*(.*)$`)
)

// ParseTopic reads an mdtest topic file into its cases.
//
// A CASE IS ONE PERL PROGRAM AND SEVERAL IMPLEMENTATIONS' ANSWERS ABOUT
// IT. The ```perl block is the program; the rest is who answers what:
//
//	perl       the program -- shared, and the only block everyone reads
//	output     what PERL prints -- the ground truth, which no
//	           implementation defines and all of them are checked against
//	behavior   whether perl compiles it, plus a recorded refusal
//	tokens     the LEXICAL answer, which is this parser's
//	ir         the GRAPH answer, which is B::SoN's
//
// THESE ARE NOT STAGES OF ONE SYSTEM. Chalk and B::SoN are separate
// implementations in other languages, and both reach a program THROUGH
// PERL -- B::SoN reads perl's optree, Chalk builds on the IR that
// produces. This parser reads the bytes, which is what makes a case it
// answers correctly evidence that Perl can be parsed without perl.
//
// Not that we have no IR. PSC lowers to the same SoN vocabulary for
// type inference, because that IR was built for it. The difference is
// what the IR is FOR: for B::SoN the graph IS the product, and for us
// it is a tool downstream of a parse that already happened. So the
// ```ir block is not permanently foreign here -- it is unfilled because
// our IR answers a different question, and it is where PSC would speak
// if it had something to say in that vocabulary.
//
// That independence is why an unrecognised block is skipped rather than
// rejected. An ```ir block written by B::SoN is its answer about a
// program we share, reached by a route with none of our code in it, and
// not ours to check or to be blocked by. A reader that refused unknown
// blocks would force every implementation to implement every other
// one's answer, making the corpus a single system again and destroying
// exactly the independence that makes it evidence.
func ParseTopic(raw string) ([]*Case, error) {
	var (
		cases   []*Case
		cur     *Case
		inFence bool
		lang    string
		body    []string
	)

	for i, line := range strings.Split(raw, "\n") {
		if !inFence {
			if m := reHeading.FindStringSubmatch(line); m != nil {
				cur = &Case{Title: m[1], Line: i + 1}
				cases = append(cases, cur)
				continue
			}
			if m := reFence.FindStringSubmatch(line); m != nil {
				lang, body, inFence = m[1], nil, true
				continue
			}
			continue
		}

		if strings.TrimSpace(line) == "```" {
			inFence = false
			if cur != nil {
				if err := cur.addBlock(lang, strings.Join(body, "\n")); err != nil {
					return nil, fmt.Errorf("%s: %w", cur.Title, err)
				}
			}
			lang, body = "", nil
			continue
		}
		body = append(body, line)
	}

	if inFence {
		return nil, fmt.Errorf("unclosed fence")
	}
	return cases, nil
}

// addBlock files one fenced block into the case, ignoring tags it does not
// know.
func (c *Case) addBlock(lang, content string) error {
	switch lang {
	case "perl":
		// A CASE IS ONE COMPILATION UNIT. Perl's compilation unit is
		// the FILE: `my` scope, BEGIN ordering, `use strict`'s lexical
		// effect, `__DATA__` and constant folding are all per unit, so
		// two programs in one case would fold across a boundary a
		// reader sees as separating them.
		//
		//	my $x = "abc";        folds into the pad -- no VarDecl
		//	my $x = "abc" . $0;   runtime -- the VarDecl survives
		//
		// Those are two fixtures a reader expects to be independent.
		// Merged into one unit they are not, and what you read is not
		// what perl compiled.
		//
		// Ty's mdtest merges consecutive unnamed blocks into one file;
		// this format deliberately does not. Rejecting the second block
		// rather than overwriting is the difference between the rule
		// being ENFORCED and being remembered -- a silent overwrite
		// drops a program and makes no diagnostic, which is how every
		// other defect in this corpus started.
		if c.Source != "" {
			return fmt.Errorf("a second ```perl block; a case is ONE " +
				"compilation unit, because perl's is the file -- split " +
				"it into two `##` cases")
		}
		c.Source = strings.TrimSpace(content) + "\n"
	case "behavior":
		return c.parseBehavior(content)
	case "output":
		// A FENCED block rather than a behavior key, because output is
		// byte-exact stdout and is routinely several lines. A
		// `key: value` line cannot carry a newline, and Chalk never hit
		// this because it pins a single-line `return:` -- our corpus
		// pins what `print` wrote, which is a different claim.
		//
		// The block's content is the pin verbatim, plus the trailing
		// newline `print` leaves. An EMPTY block pins empty output,
		// which is a real claim and distinct from no block at all.
		s := content
		if s != "" && !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		c.ExpectOutput = &s
	case "tokens":
		for _, l := range strings.Split(content, "\n") {
			if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "#") {
				c.TokenFacts = append(c.TokenFacts, l)
			}
		}
	}
	// Every other tag -- `ir`, and whatever another implementation adds
	// -- is its answer about a program we share, and not this reader's to
	// validate.
	return nil
}

// parseBehavior reads the `key: value` lines of a behavior block.
func (c *Case) parseBehavior(content string) error {
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		m := reKeyVal.FindStringSubmatch(line)
		if m == nil {
			return fmt.Errorf("behavior line is not `key: value`: %q", line)
		}
		key, val := m[1], strings.TrimRight(m[2], " \t")

		switch key {
		case "parses":
			switch val {
			case "yes":
				c.ExpectParses = true
			case "no":
				c.ExpectParsent = true
			default:
				return fmt.Errorf("parses must be yes or no, got %q", val)
			}
		case "refuses":
			c.Refuses = val
		case "refusal":
			c.RefusalCode = val
		default:
			return fmt.Errorf("unknown behavior key %q", key)
		}
	}
	return nil
}

// asFile presents a case as a `File`, which is what the tier tests
// already take.
//
// The two carry the same claims -- a case IS a corpus file that stopped
// needing its own path -- so this is a field copy rather than a
// conversion, and it is what let fifteen tier test files migrate to
// topics without being touched.
func (c *Case) asFile() *File {
	return &File{
		Source:        c.Source,
		ExpectOutput:  c.ExpectOutput,
		ExpectParses:  c.ExpectParses,
		ExpectParsent: c.ExpectParsent,
		TokenFacts:    c.TokenFacts,
		Refuses:       c.Refuses,
		RefusalCode:   parse.RefusalCode(c.RefusalCode),
	}
}
