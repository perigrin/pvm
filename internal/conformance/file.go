// ABOUTME: Parses a corpus .t file into its `--- name` sections.
// ABOUTME: Everything before the first section is commentary, not source.
package conformance

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"tamarou.com/pvm/internal/parse"
)

// File is one corpus case: the source to run and what is expected of it.
type File struct {
	// Source is the Perl the case runs. Never empty in a valid file.
	Source string

	// ExpectParses and ExpectParsent are the must-parse / must-not-parse
	// bit, from `--- expect parses` or `--- expect parsent`. Exactly one
	// is set: a file asserting neither says nothing.
	ExpectParses  bool
	ExpectParsent bool

	// ExpectOutput is what perl prints, verbatim including the trailing
	// newline. NIL when the file has no `--- expect output` section,
	// which is the normal case for a parsent file.
	//
	// A pointer rather than a string, because PINNING EMPTY OUTPUT AND
	// PINNING NOTHING ARE DIFFERENT CLAIMS. As a string both read as "",
	// so a file whose whole point is that a construct prints nothing had
	// its check skipped and perl could print anything unnoticed. A
	// sentinel is not available: the pin is byte-exact against perl's
	// stdout, so any string a sentinel could be is a string a corpus file
	// may legitimately pin. A paired bool would say the same thing in two
	// fields that can disagree; the pointer cannot be set and absent at
	// once, and it is the one field here whose absence is meaningful --
	// the expectation bits next to it are false-when-absent by nature.
	ExpectOutput *string

	// TokenFacts are declared lexical claims, one per line, in the
	// vocabulary of conformance/GLOSSARY.md rather than of our lexer.
	TokenFacts []string

	// Refuses is the issue id from a `# STATUS refuses` line, empty when
	// the file is expected to pass.
	//
	// A known refusal runs as a Go subtest failure would be too loud: the
	// suite must stay pristine, and a corpus whose point is to name what
	// does not work yet cannot also be all-green. So a refusing file is
	// SKIPPED with its issue id, and the ratchet counts it. What must
	// never happen silently is the reverse -- a file marked refusing that
	// now passes, which Run reports as an error so the marker gets removed.
	Refuses string

	// RefusalCode is the parse.RefusalCode a `Refusal <code>.` clause on
	// the STATUS line names, empty when the file names none.
	//
	// Empty is not a lesser refusal, it is an older one: seventeen files
	// predate codes and cite only an issue or themselves. A code is an
	// ADDITIONAL promise -- "this file waits on THIS site declining" --
	// and a file that makes it gets checked against it, while a file that
	// does not keeps skipping as before.
	RefusalCode parse.RefusalCode
}

// ParseFile splits a corpus file into its sections.
//
// The format is a leading comment block, then `--- <name>` sections. The
// comment block carries the measured perl invocations and the refusal
// status; it is documentation and is not returned.
func ParseFile(raw string) (*File, error) {
	f := &File{}

	// Normalise line endings before anything looks at the bytes. Markers
	// would match under CRLF anyway -- the name is trimmed -- so such a
	// file parses "successfully" while carrying a \r into ExpectOutput,
	// which is compared against perl's stdout byte for byte. The result is
	// a CORPUS BUG report blaming the author for a line ending. Done here
	// rather than in .gitattributes because a contributor's local git
	// config cannot bypass it.
	raw = strings.ReplaceAll(raw, "\r\n", "\n")

	// Split on section markers at the start of a line. The first chunk is
	// the comment block, which has no marker and is discarded.
	var (
		name    string
		started bool // a marker has been seen; we are past the comment block
		body    strings.Builder
		comment strings.Builder
		seen    = map[string]bool{}
	)

	// started and name are tracked separately on purpose. Testing only
	// `name == ""` conflates "still in the comment block" with "the last
	// marker had no name", and the second case would then be skipped by
	// the guard meant for the first -- discarding that section's body,
	// which for a `--- ` typo above the source means losing the program
	// with no error at all.
	flush := func() error {
		if !started {
			return nil
		}
		if name == "" {
			return fmt.Errorf("section marker with no name")
		}
		if seen[name] {
			return fmt.Errorf("duplicate section %q", name)
		}
		seen[name] = true
		return f.setSection(name, body.String())
	}

	// openable says the line just read leaves the next one free to open a
	// section: it was blank, or a marker of its own, or there was no
	// previous line because we are at the start of the file.
	//
	// A heredoc body, POD block or data section is OPAQUE to this split, so
	// a line inside one beginning `--- ` was read as a real marker. The two
	// spellings failed differently and only one failed safely: `--- not a
	// marker` was rejected as an unknown section, while `--- expect output`
	// was accepted SILENTLY -- the source truncated at that line, leaving an
	// unterminated heredoc opener, and the rest of the program became the
	// pinned output.
	//
	// Requiring a blank line before a marker is what separates the two.
	// Measured across all 134 corpus files, every one of their 472 markers
	// already has one, so the rule costs the existing corpus nothing and it
	// is stated in conformance/README.md for new files. A marker directly
	// after another marker is allowed because a section with an empty body
	// is ordinary -- `--- expect parses` carries nothing at all.
	//
	// This is a WEAK check by construction: a heredoc body may contain a
	// blank line followed by `--- expect output` and still slip through.
	// Closing the hole properly would mean lexing the source section to find
	// its heredoc openers, which is the parser this corpus exists to test --
	// a corpus file must be readable without it, or a parser bug becomes a
	// corpus that cannot be read. The rule turns the reachable accident
	// loud, which is the part that was missing.
	openable := true
	for _, line := range strings.SplitAfter(raw, "\n") {
		if marker, ok := sectionName(line); ok {
			if !openable {
				return nil, fmt.Errorf("section marker %q needs a blank line before it, "+
					"or it is a line inside a heredoc, POD block or data section",
					strings.TrimRight(line, "\n"))
			}
			if err := flush(); err != nil {
				return nil, err
			}
			name, started = marker, true
			body = strings.Builder{}
			openable = true
			continue
		}
		openable = strings.TrimSpace(line) == ""
		if !started {
			// Still in the comment block; collected whole, because the
			// STATUS marker and the issue id it refers to need not share
			// a line. The rest of the block is the measured-perl record
			// and is read by people, not here.
			comment.WriteString(line)
			continue
		}
		body.WriteString(line)
	}
	if err := flush(); err != nil {
		return nil, err
	}

	f.Refuses = refusalIssue(comment.String())
	f.RefusalCode = refusalCode(comment.String())

	// Absent and empty are different mistakes and get different messages:
	// told "no --- source section" when the marker is right there, an
	// author goes looking for a line that already exists. Whitespace-only
	// counts as empty because perl compiles it as an empty program, so it
	// would otherwise be a file that passes while measuring nothing.
	if !seen["source"] {
		return nil, fmt.Errorf("no --- source section")
	}
	if strings.TrimSpace(f.Source) == "" {
		return nil, fmt.Errorf("--- source section is empty")
	}
	if !f.ExpectParses && !f.ExpectParsent {
		return nil, fmt.Errorf("file asserts neither parses nor parsent")
	}
	if f.ExpectParses && f.ExpectParsent {
		// Resolving this by field order would make the contradiction
		// invisible: Run consults ExpectParses first, so a file claiming
		// both would quietly measure only the must-parse half.
		return nil, fmt.Errorf("file asserts both parses and parsent")
	}
	return f, nil
}

// reRefusal matches the one machine-read line of the comment block:
//
//	# STATUS refuses as of 38c95d23. Issue 01a0c13f-97f5.
//	# STATUS refuses as of 38c95d23, and this was NOT a known gap.
//
// An issue id is optional because a refusal can be FOUND BY WRITING THE
// FILE, which is the corpus doing its job. Such a refusal cites the file
// itself; see selfRecorded.
var reRefusal = regexp.MustCompile(`(?m)^#\s*STATUS refuses\b`)

// reIssue finds a crochet issue id anywhere in the comment block.
//
// The full 36-character form is tried FIRST, because alternation is
// ordered and the short form is a prefix of it -- matching the short one
// first would capture 13 characters of a full id and stop.
//
// The short form remains accepted so files already in the tree keep
// parsing, but it is not unique: these are UUIDv7, whose leading 8
// characters are a millisecond timestamp, so a batch-created chain of
// issues collides there by construction. Measured across the 27 ids in
// docs/plans/2026-09-21-deferred-chain-m1-m2.md, the 8-character prefix
// gives 10 distinct values and the 13-character prefix 23. Uniqueness
// rests on 16 bits, which is why new citations spell the id whole.
var reIssue = regexp.MustCompile(`\bIssue ([0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}|[0-9a-f]{8}-[0-9a-f]{4})\b`)

// refusalIssue returns the issue id a comment block's STATUS line refers
// to, "this file" when it refuses without citing one, or "" when the file
// is expected to pass.
func refusalIssue(comment string) string {
	if !reRefusal.MatchString(comment) {
		return ""
	}
	if m := reIssue.FindStringSubmatch(comment); m != nil {
		return m[1]
	}
	return selfRecorded
}

// reRefusalCode finds the refusal code a comment block names:
//
//	# STATUS refuses as of 9102578c. Issue 01a0c35e-....  Refusal missing_operand.
//
// Attached to the STATUS line that already carries the refusal, rather
// than given a section of its own. A refusal is ONE fact -- what does not
// work, why it is known, and which site declines -- and splitting it
// across two places in a file would let the halves disagree, which is the
// drift this corpus keeps paying for elsewhere.
//
// The clause is optional, and its absence is not a defect: seventeen
// files predate codes, and a format change that invalidated them would be
// paid for by a bulk edit of files nobody was otherwise touching.
//
// lower_snake_case is the whole of the grammar. A code is an identifier
// rather than prose precisely so this pattern can be strict.
var reRefusalCode = regexp.MustCompile(`\bRefusal ([a-z][a-z0-9_]*)\b`)

// refusalCode returns the code a comment block's STATUS line names, or ""
// when it names none or the file does not refuse at all.
func refusalCode(comment string) parse.RefusalCode {
	if !reRefusal.MatchString(comment) {
		return ""
	}
	if m := reRefusalCode.FindStringSubmatch(comment); m != nil {
		return parse.RefusalCode(m[1])
	}
	return ""
}

// selfRecorded is the citation for a refusal the corpus found itself.
//
// A construct discovered by writing its file has nowhere earlier to have
// been filed, and the file already holds the source, the measured perl
// behaviour and the tokens we produce instead. An issue would be a second
// copy of that, free to go stale. So this is a complete citation rather
// than a placeholder for one.
const selfRecorded = "this file"

// sectionNames lists every section a corpus file may contain, in the
// order the README presents them.
//
// setSection's switch is the parser's copy of this list. They are checked
// against each other rather than merged into one table, because the switch
// assigns to differently-typed fields and a table would need an interface
// or a closure per entry to say the same thing.
func sectionNames() []string {
	return []string{
		"source",
		"expect parses",
		"expect parsent",
		"expect output",
		"expect tokens",
	}
}

// sectionName reports whether a line opens a section, and which.
func sectionName(line string) (string, bool) {
	rest, ok := strings.CutPrefix(line, "--- ")
	if !ok {
		return "", false
	}
	return strings.TrimSpace(rest), true
}

func (f *File) setSection(name, body string) error {
	switch name {
	case "source":
		f.Source = body
	case "expect parses":
		f.ExpectParses = true
	case "expect parsent":
		f.ExpectParsent = true
	case "expect output":
		// Expected output must match perl's bytes exactly, so the blank
		// line separating this section from the next is not part of it.
		// One trailing newline is stripped; everything else is content,
		// including a deliberate trailing blank line written as two.
		//
		// Assigned through a local so the pointer is non-nil even when
		// the body strips to "": the section being PRESENT is the claim,
		// and its emptiness is the claim's content.
		pinned := strings.TrimSuffix(body, "\n")
		f.ExpectOutput = &pinned
	case "expect tokens":
		for line := range strings.SplitSeq(body, "\n") {
			if line = strings.TrimSpace(line); line != "" {
				f.TokenFacts = append(f.TokenFacts, line)
			}
		}
	default:
		return fmt.Errorf("unknown section %q", name)
	}
	return nil
}

// zhiBinary is the issue tracker a refusal citation names an issue in.
const zhiBinary = "git-zhi"

// citationResolves reports whether a refusal citation still stands.
//
// Three citations, three answers. "" is a file that does not refuse.
// selfRecorded is COMPLETE as it stands -- the file itself is the record
// -- so looking it up would fail the fifteen corpus files that use it.
// Anything else is an issue id, and an id that no longer resolves is a
// skip with nothing behind it: the file still claims a known gap while
// the record of what the gap IS has been destroyed, which nearly happened
// during a chain prune.
func citationResolves(citation string) error {
	if citation == "" || citation == selfRecorded {
		return nil
	}
	return issueResolves(citation)
}

var (
	issueOnce sync.Mutex
	issueSeen = map[string]error{}
)

// issueResolves asks the tracker whether one id exists, once per process.
//
// `git-zhi issue show` exits 0 for an id it finds and 1 for one it does
// not, including for an issue in any state -- a done issue resolves even
// though the default `issue list` omits it. The answer is memoised
// because it cannot change mid-run and each call costs ~0.4s, which the
// corpus would otherwise pay per citing file.
func issueResolves(id string) error {
	issueOnce.Lock()
	defer issueOnce.Unlock()

	if err, ok := issueSeen[id]; ok {
		return err
	}
	err := exec.Command(zhiBinary, "issue", "show", id).Run()
	if err != nil {
		err = fmt.Errorf("cited issue %s does not resolve: %s cannot find it", id, zhiBinary)
	}
	issueSeen[id] = err
	return err
}
