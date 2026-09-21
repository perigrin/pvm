// ABOUTME: Runs one corpus case: perl adjudicates, then our lexer and parser answer.
// ABOUTME: A case that perl itself refuses is a corpus bug, reported as such.
//
// The format this runs against is docs/plans/2026-09-21-graded-conformance-corpus.md,
// and the questions that design left open are settled here rather than in
// prose: a repeated section is an error, expected output is byte-exact,
// and a refusal is a SKIP carrying its citation so the suite stays
// pristine while still naming what does not work.
package conformance

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lexer"
	"tamarou.com/pvm/internal/parse"
)

// Run checks one corpus case.
//
// The order is deliberate and is the design's: PERL ADJUDICATES FIRST. A
// file whose `--- expect parses` disagrees with perl is a bug in the
// corpus, not a finding about the parser, and saying so before reporting
// on our parser keeps a wrong expectation from being read as a refusal.
func Run(t *testing.T, f *File) {
	t.Helper()

	v := verdict(t, f)
	switch v.kind {
	case corpusBug:
		t.Fatalf("%s", strings.Join(v.msgs, "\n\t"))
	case knownRefusal:
		t.Skip(v.reason)
	default:
		for _, m := range v.msgs {
			t.Error(m)
		}
	}
}

// verdictKind is what running one corpus file concluded.
type verdictKind int

const (
	// passed: our lexer and parser handled the file, and it claimed no
	// refusal. The only kind that reports nothing.
	passed verdictKind = iota

	// corpusBug: perl disagrees with the file's own expectation, so the
	// file is wrong and nothing was measured about our parser.
	corpusBug

	// knownRefusal: our parser does not handle this yet, and the file
	// says so. Skipped rather than failed, because a corpus whose point
	// is to name what does not work cannot also be all-green.
	knownRefusal

	// staleMarker: the file claims a refusal it no longer has. Reported
	// as an error, because this is how a corpus silently stops measuring.
	staleMarker

	// refused: our parser does not handle this and the file does NOT say
	// so. An ordinary failure.
	refused
)

func (k verdictKind) String() string {
	switch k {
	case passed:
		return "passed"
	case corpusBug:
		return "corpusBug"
	case knownRefusal:
		return "knownRefusal"
	case staleMarker:
		return "staleMarker"
	case refused:
		return "refused"
	}
	return "unknown"
}

// verdictResult is what verdict concluded and what it would say.
//
// Separated from the reporting so the decision can be TESTED. Run's own
// calls are Fatalf and Skipf, which end the calling goroutine, so a test
// that called Run could observe the first verdict and never the rest.
type verdictResult struct {
	kind verdictKind
	msgs []string
	// reason is the skip text for a knownRefusal, which names the
	// refusal and what our parser actually said.
	reason string
}

// verdict decides one corpus case without reporting it.
//
// The order is the design's and is not negotiable: PERL ADJUDICATES
// FIRST, and a disagreement returns immediately with the parser never
// consulted. Reporting on our parser first would let a wrong expectation
// read as a refusal, and a refusal is a claim about US.
func verdict(t *testing.T, f *File) verdictResult {
	t.Helper()

	src := []byte(f.Source)
	compiles, output := askPerl(t, f.Source)

	// 1. The corpus agrees with perl, or the corpus is wrong. Each of
	//    these returns rather than collecting, so exactly one CORPUS BUG
	//    is reported and the parser is never reached.
	switch {
	case f.ExpectParses && !compiles:
		return verdictResult{kind: corpusBug, msgs: []string{
			"CORPUS BUG: file says `expect parses`, perl -c refuses it"}}
	case f.ExpectParsent && compiles:
		return verdictResult{kind: corpusBug, msgs: []string{
			"CORPUS BUG: file says `expect parsent`, perl -c accepts it"}}
	}
	// A PIN is checked; an ABSENT section is not. The test is presence,
	// never emptiness -- `!= ""` skipped exactly the file whose claim is
	// that the construct prints nothing.
	if f.ExpectOutput != nil && output != *f.ExpectOutput {
		return verdictResult{kind: corpusBug, msgs: []string{fmt.Sprintf(
			"CORPUS BUG: pinned output %q, perl prints %q", *f.ExpectOutput, output)}}
	}

	// 2. Our lexer and parser, collected rather than reported, because a
	//    file marked refusing must not fail the suite and a file that has
	//    started passing must not stay marked.
	ours := &recorder{}

	var codes []parse.RefusalCode
	if f.ExpectParses {
		codes = refusalCodes(parse.Parse(src))
		if len(codes) > 0 {
			ours.Errorf("parser refuses: %d Unknown node(s), %s",
				len(codes), joinCodes(codes))
		}
	}
	for _, fact := range f.TokenFacts {
		checkTokenFact(ours, fact, src)
	}

	switch {
	case f.Refuses != "" && len(ours.msgs) == 0:
		return verdictResult{kind: staleMarker, msgs: []string{fmt.Sprintf(
			"file is marked `STATUS refuses` (%s) but now PASSES.\n"+
				"\tRemove the STATUS line -- a stale marker hides a regression.",
			f.Refuses)}}

	// A file that NAMES the site it waits on, refusing at a different
	// one. The same failure as a stale marker and reported as one: the
	// file still skips on a claim its own header no longer describes, so
	// it has stopped measuring what it documents. Skipping would hide
	// exactly the event the code was added to expose -- a refusal that
	// changed cause while staying a refusal.
	//
	// Only for a file that made the claim. One with no code has promised
	// nothing about WHICH site declines, and there is nothing to be wrong
	// about.
	case f.RefusalCode != "" && !hasCode(codes, f.RefusalCode):
		return verdictResult{kind: staleMarker, msgs: []string{fmt.Sprintf(
			"file names refusal %s, but the parser refuses with %s.\n"+
				"\tThe CAUSE changed: this file no longer measures what its "+
				"header documents.\n"+
				"\tRe-measure it and update the `Refusal` clause, or the "+
				"STATUS line if the gap closed.",
			f.RefusalCode, joinCodes(codes))}}

	case f.Refuses != "":
		return verdictResult{
			kind:   knownRefusal,
			msgs:   ours.msgs,
			reason: fmt.Sprintf("refuses (%s):\n\t%s", f.Refuses, strings.Join(ours.msgs, "\n\t")),
		}
	case len(ours.msgs) > 0:
		return verdictResult{kind: refused, msgs: ours.msgs}
	}
	return verdictResult{kind: passed}
}

// reporter is the part of testing.TB that a check needs, so a check can be
// run for its verdict rather than for its side effect on the suite.
type reporter interface {
	Errorf(format string, args ...any)
}

// recorder collects what a check would have reported.
type recorder struct{ msgs []string }

func (r *recorder) Errorf(format string, args ...any) {
	r.msgs = append(r.msgs, fmt.Sprintf(format, args...))
}

// askPerl runs the source under perl, returning whether it compiles and
// what it prints. Compilation is checked separately from running because
// croak cases fail at runtime having parsed perfectly well.
func askPerl(t *testing.T, source string) (compiles bool, output string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "case.pl")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}

	perl, err := perlPath()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if err := exec.Command(perl, "-c", path).Run(); err != nil {
		return false, ""
	}
	out, err := exec.Command(perl, path).Output()
	if err != nil {
		// Compiles but dies. Output is whatever reached stdout first.
		return true, string(out)
	}
	return true, string(out)
}

// refusalCodes collects the code of every Unknown in a tree, in tree
// order.
//
// Replaces a plain count. The count is still here -- it is the length --
// and the codes are what a corpus file can name, so a refusal is
// reportable as WHICH site declined rather than only as how many did.
func refusalCodes(n *parse.Node) []parse.RefusalCode {
	if n == nil {
		return nil
	}
	var out []parse.RefusalCode
	if n.Kind == parse.Unknown {
		out = append(out, n.Refusal)
	}
	for _, c := range n.Children {
		out = append(out, refusalCodes(c)...)
	}
	return out
}

// hasCode reports whether a file's named refusal is among those the
// parser actually produced.
//
// Membership rather than equality, because a file may refuse in more than
// one place and naming one of them is a true claim about it. A file whose
// named code is nowhere in the list is describing a refusal it does not
// have.
func hasCode(codes []parse.RefusalCode, want parse.RefusalCode) bool {
	return slices.Contains(codes, want)
}

// joinCodes renders refusal codes for a message, each named once and in
// tree order, so a file refusing four times at one site does not print
// that site four times.
func joinCodes(codes []parse.RefusalCode) string {
	if len(codes) == 0 {
		return "no refusal"
	}
	var seen []string
	for _, c := range codes {
		name := string(c)
		if name == "" {
			name = "(no code)"
		}
		if !slices.Contains(seen, name) {
			seen = append(seen, name)
		}
	}
	return strings.Join(seen, ", ")
}

// significant returns the tokens that carry meaning, dropping whitespace.
func significant(src []byte) []lexer.Token {
	var out []lexer.Token
	for _, tk := range lexer.Tokenize(src) {
		if tk.Kind != lexer.Whitespace {
			out = append(out, tk)
		}
	}
	return out
}
