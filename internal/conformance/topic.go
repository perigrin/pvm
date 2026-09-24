// ABOUTME: Reads Chalk's mdtest corpus format -- `##` per case, fenced blocks by language tag.
// ABOUTME: A block tag this reader does not know is IGNORED, which is what lets three projects share one file.
package conformance

import (
	"fmt"
	"regexp"
	"strings"
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
	// ExpectOutput is a pointer for `File`'s reason: pinning empty output
	// and pinning nothing are different claims.
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

var (
	reHeading = regexp.MustCompile(`^##\s+(.+?)\s*$`)
	reFence   = regexp.MustCompile("^```(\\w*)\\s*$")
	reKeyVal  = regexp.MustCompile(`^\s*([\w-]+)\s*:\s*(.*)$`)
)

// ParseTopic reads an mdtest topic file into its cases.
//
// THE UNKNOWN-TAG RULE IS THE POINT. A fenced block whose language this
// reader does not recognise is skipped rather than rejected, which is what
// lets one corpus file serve several projects: Chalk fills an ```ir block
// with typed-graph node signatures, B::SoN could fill one with optree
// shape, and this reader ignores both while they ignore ```tokens.
//
// A corpus that refused unknown blocks would force every consumer to
// implement every other consumer's claims, which is the coupling the whole
// format is meant to avoid.
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
	// Every other tag -- `ir`, and whatever a later consumer adds -- is
	// another project's claim and is not this reader's to validate.
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
