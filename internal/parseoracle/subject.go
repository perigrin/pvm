// ABOUTME: The subprocess contract by which any Perl implementation can be measured.
// ABOUTME: Subjects report facts about their parse as JSON; no syntax tree crosses this boundary.

package parseoracle

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
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
	Prototypes map[string]string `json:"prototypes,omitempty"`

	// CallSites reports what the subject decided at each call.
	CallSites []SubjectCallSite `json:"call_sites,omitempty"`

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

// SubjectCallSite is one call, and what the subject concluded about it.
type SubjectCallSite struct {
	Line int    `json:"line"`
	Name string `json:"name"`

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

	out, err := cmd.Output()
	if err != nil {
		return SubjectFacts{}, fmt.Errorf("subject %s: %w", s.Command[0], err)
	}

	return decodeSubjectFacts(out)
}

// decodeSubjectFacts parses the JSON and records which optional questions the
// subject actually answered. Presence is decided from the raw object rather
// than from zero values, because an empty call_sites list is a real answer
// ("no calls here") and a missing one is not an answer at all.
func decodeSubjectFacts(out []byte) (SubjectFacts, error) {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(out, &probe); err != nil {
		return SubjectFacts{}, fmt.Errorf("subject output is not a JSON object: %w", err)
	}

	var facts SubjectFacts
	if err := json.Unmarshal(out, &facts); err != nil {
		return SubjectFacts{}, fmt.Errorf("decode subject facts: %w", err)
	}

	_, facts.KnowsCallSites = probe["call_sites"]
	_, facts.KnowsPrototypes = probe["prototypes"]
	return facts, nil
}
