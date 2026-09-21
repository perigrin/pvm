// ABOUTME: Runs one corpus case: perl adjudicates, then our lexer and parser answer.
// ABOUTME: A case that perl itself refuses is a corpus bug, reported as such.
package conformance

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	src := []byte(f.Source)
	compiles, output := askPerl(t, f.Source)

	// 1. The corpus agrees with perl, or the corpus is wrong.
	switch {
	case f.ExpectParses && !compiles:
		t.Fatalf("CORPUS BUG: file says `expect parses`, perl -c refuses it")
	case f.ExpectParsent && compiles:
		t.Fatalf("CORPUS BUG: file says `expect parsent`, perl -c accepts it")
	}
	if f.ExpectOutput != "" && output != f.ExpectOutput {
		t.Fatalf("CORPUS BUG: pinned output %q, perl prints %q", f.ExpectOutput, output)
	}

	// 2. Our lexer and parser, collected rather than reported, because a
	//    file marked refusing must not fail the suite and a file that has
	//    started passing must not stay marked.
	ours := &recorder{}

	if f.ExpectParses {
		if n := unknowns(parse.Parse(src)); n > 0 {
			ours.Errorf("parser refuses: %d Unknown node(s)", n)
		}
	}
	for _, fact := range f.TokenFacts {
		checkTokenFact(ours, fact, src)
	}

	switch {
	case f.Refuses != "" && len(ours.msgs) == 0:
		t.Errorf("file is marked `STATUS refuses` (%s) but now PASSES.\n"+
			"\tRemove the STATUS line -- a stale marker hides a regression.", f.Refuses)
	case f.Refuses != "":
		t.Skipf("refuses (%s):\n\t%s", f.Refuses, strings.Join(ours.msgs, "\n\t"))
	default:
		for _, m := range ours.msgs {
			t.Error(m)
		}
	}
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

	if err := exec.Command("perl", "-c", path).Run(); err != nil {
		return false, ""
	}
	out, err := exec.Command("perl", path).Output()
	if err != nil {
		// Compiles but dies. Output is whatever reached stdout first.
		return true, string(out)
	}
	return true, string(out)
}

func unknowns(n *parse.Node) int {
	if n == nil {
		return 0
	}
	count := 0
	if n.Kind == parse.Unknown {
		count++
	}
	for _, c := range n.Children {
		count += unknowns(c)
	}
	return count
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
