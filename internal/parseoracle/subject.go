// ABOUTME: The subprocess contract by which any Perl implementation can be measured.
// ABOUTME: Subjects report facts about their parse as JSON; no syntax tree crosses this boundary.

package parseoracle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// SubjectFacts is what an implementation reports about its own parse.
//
// The contract is deliberately FACTS rather than a tree. An earlier design
// passed *parser.Tree directly, which restricted the harness to one parser; a
// Go interface would have restricted it to Go. The implementations worth
// measuring are not all Go — perl-lsp is Rust, PerlOnJava is JVM, Chalk is
// Perl, perl itself is C — so the boundary has to be a subprocess, and the
// payload has to be something every one of them can produce.
//
// Emitting a tree would not do. Measured, our own `psc parse --format sexpr`
// answers in tree-sitter's vocabulary (`ambiguous_function_call_expression`,
// `subroutine_declaration_statement`). Another implementation names the same
// constructs differently, so comparing trees across implementations first
// requires agreeing a universal Perl AST — a research project, not a harness.
// These questions, by contrast, are ones any parser can answer about itself.
type SubjectFacts struct {
	// OK is whether the subject believes the file is valid Perl.
	OK bool `json:"ok"`

	// Prototypes maps sub name to prototype string, for prototypes the
	// subject resolved.
	//
	// Neither this nor CallSites carries omitempty, and that is the contract
	// rather than an oversight: an empty map or list is an answer ("I looked
	// and found none") and nil is not ("I did not look"). Marshalling nil
	// emits JSON null, which decodes as absent; see decodeSubjectFacts.
	Prototypes map[string]string `json:"prototypes"`

	// CallSites reports what the subject decided at each call, and at each
	// reference the source took outside a call; see SubjectCallSite.Kind.
	CallSites []SubjectCallSite `json:"call_sites"`

	// Declined is the portable form of "I gave up here". It replaces the
	// tree-sitter-specific IsDegenerate: a hidden grammar rule surfacing as a
	// node is one implementation's way of saying it dropped source it could
	// not handle, and every parser can answer that question in its own terms.
	Declined bool `json:"declined,omitempty"`

	// DeclinedReason says where and why, for triage.
	DeclinedReason string `json:"declined_reason,omitempty"`

	// KnowsCallSites distinguishes "no call sites" from "cannot answer".
	// A subject that omits the field must not be scored as agreeing; see
	// CompareFacts. Set during decoding, never by the subject.
	KnowsCallSites bool `json:"-"`

	// KnowsPrototypes is the same distinction for prototypes.
	KnowsPrototypes bool `json:"-"`
}

// SiteKindReference marks a site that is not a call: the source took a
// reference itself, as in `my $r = \@a;`. Perl reports that with the same
// srefgen it uses for a prototype-driven reference, so a subject has to be
// able to account for it -- and a faithful subject that reports only calls
// would otherwise be scored WRONG on every file containing a backslash.
const SiteKindReference = "reference"

// SubjectCallSite is one call, and what the subject concluded about it.
type SubjectCallSite struct {
	// Kind is empty for a call. SiteKindReference is the one other kind:
	// a reference the source wrote, reported with TookReference set.
	Kind string `json:"kind,omitempty"`

	// Line is the first line of the STATEMENT the site is in, and EndLine
	// its last; a site is matched to perl's by statement, because that is
	// the finest attribution perl offers and perl may record any line of
	// the statement for it. EndLine may be omitted, meaning Line.
	Line    int    `json:"line"`
	EndLine int    `json:"end_line,omitempty"`
	Name    string `json:"name"`

	// TookReference is the question the oracle answers with srefgen: did this
	// call pass a reference rather than a flattened list? A prototype can make
	// that true without the source saying so.
	TookReference bool `json:"took_reference"`

	// Unresolved is the subject saying it could not settle this call -- it
	// knows a prototype might apply and has not committed either way. This is
	// what separates `wider` from `WRONG`: declining to answer is honest,
	// committing to the wrong answer is not.
	//
	// Our own grammar expresses this by emitting a distinct node kind for an
	// unqualified f(...) call; another implementation will express it its own
	// way. The contract asks for the conclusion, not the representation.
	Unresolved bool `json:"unresolved,omitempty"`
}

// Subject is an implementation under measurement, invoked as a subprocess.
//
// Command is argv; the file path is appended. Anything executable qualifies —
// the tests use a shell script, which is the proof that the boundary holds.
type Subject struct {
	Command []string
	Dir     string
	Timeout time.Duration
}

// defaultSubjectTimeout bounds one file. The corpus contains files that wedge
// a parser; a sweep must not wedge with them.
const defaultSubjectTimeout = 60 * time.Second

// Parse runs the subject against one file and decodes what it reports.
func (s Subject) Parse(ctx context.Context, path string) (SubjectFacts, error) {
	if len(s.Command) == 0 {
		return SubjectFacts{}, fmt.Errorf("subject: no command configured")
	}

	timeout := s.Timeout
	if timeout <= 0 {
		timeout = defaultSubjectTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	args := append(append([]string{}, s.Command[1:]...), path)
	cmd := exec.CommandContext(ctx, s.Command[0], args...)
	cmd.Dir = s.Dir
	// A killed subject may leave a grandchild holding the pipe; without this
	// the read blocks past cancellation. The oracle learned the same lesson.
	cmd.WaitDelay = 5 * time.Second
	killGroup(cmd)

	var out cappedBuffer
	var stderr strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// A timeout is named as one. "signal: killed" reads as a crash and
		// sends triage the wrong way; the oracle makes the same distinction.
		if ctxErr := ctx.Err(); ctxErr != nil {
			return SubjectFacts{}, fmt.Errorf("subject %s on %s: %w", s.Command[0], path, ctxErr)
		}
		// The subject's own diagnosis is the useful part of a failure: "exit
		// status 2" says nothing, "cannot read op/sub.t" says which directory
		// it ran from.
		return SubjectFacts{}, fmt.Errorf("subject %s on %s: %w (%s)", s.Command[0], path, err,
			strings.TrimSpace(stderr.String()))
	}
	if out.overflow {
		return SubjectFacts{}, fmt.Errorf("subject %s on %s: output exceeds %d bytes",
			s.Command[0], path, maxSubjectOutput)
	}

	return decodeSubjectFacts(out.buf.Bytes())
}

// maxSubjectOutput bounds what one subject may print. The facts for a file are
// kilobytes; a subject that streams megabytes is broken, and one buffer per
// worker of unbounded size is how a broken subject takes the machine down.
const maxSubjectOutput = 16 << 20

// cappedBuffer keeps the first maxSubjectOutput bytes and remembers that more
// arrived. It accepts the surplus rather than refusing it, because a Write
// error stops the copy and leaves the child blocked on a full pipe until its
// timeout -- which would report a size problem as a hang.
type cappedBuffer struct {
	buf      bytes.Buffer
	overflow bool
}

func (c *cappedBuffer) Write(p []byte) (int, error) {
	if room := maxSubjectOutput - c.buf.Len(); len(p) > room {
		c.overflow = true
		c.buf.Write(p[:room])
		return len(p), nil
	}
	return c.buf.Write(p)
}

// decodeSubjectFacts parses the JSON and records which optional questions the
// subject actually answered. Presence is decided from the raw object rather
// than from zero values, because an empty call_sites list is a real answer
// ("no calls here") and a missing one is not an answer at all.
//
// A JSON null counts as missing. A Go subject without omitempty, a JSON::PP
// undef and a Jackson null field all spell "not computed" that way, and
// scoring it as "computed, found nothing" would hand out verdicts the subject
// never gave.
func decodeSubjectFacts(out []byte) (SubjectFacts, error) {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(out, &probe); err != nil {
		return SubjectFacts{}, fmt.Errorf("subject output is not a JSON object: %w", err)
	}
	// `null` and `{}` are what a subject prints when it crashed before
	// forming an opinion. Neither is a verdict, so neither reaches one: the
	// file becomes a runner error rather than "the subject rejected it".
	if probe == nil {
		return SubjectFacts{}, fmt.Errorf("subject output is null, not a JSON object")
	}
	if !answered(probe, "ok") {
		return SubjectFacts{}, fmt.Errorf("subject output has no \"ok\" field, so it never said whether the file is Perl")
	}

	var facts SubjectFacts
	if err := json.Unmarshal(out, &facts); err != nil {
		return SubjectFacts{}, fmt.Errorf("decode subject facts: %w", err)
	}

	facts.KnowsCallSites = answered(probe, "call_sites")
	facts.KnowsPrototypes = answered(probe, "prototypes")
	return facts, nil
}

// answered reports whether the subject gave field a value: present, and not null.
func answered(probe map[string]json.RawMessage, field string) bool {
	raw, present := probe[field]
	return present && string(raw) != "null"
}
