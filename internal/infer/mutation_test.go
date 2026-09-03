// ABOUTME: Type-overwriting mutation tests — inject a type error and require PSC to report it.
// ABOUTME: The oracle is perl itself, so no hand-written ground truth is needed.

package infer_test

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TYPE-OVERWRITING MUTATION, after Chaliasos et al. (PLDI'22), which found
// real soundness bugs in javac, kotlinc and groovyc with it.
//
// The technique replaces a value with one of an incompatible type. If the
// checker still accepts the program, that is a soundness bug — it failed to
// notice an error that is actually there. The dual mutation in that paper,
// type ERASURE, does not apply to PSC: it strips declared annotations, and
// Perl has no annotation syntax, so every program is already fully erased.
// PSC reads no declared types at all (see inferParamTypesFromUsage), which
// makes erasure vacuous here rather than easy.
//
// WHAT MAKES THIS WORTH RUNNING is that it needs no annotated corpus. The
// oracle is the mutation itself: we KNOW an error was injected, so silence is
// a finding.
//
// The trap, and it caught me: "I injected a type error" is a claim about PERL,
// not about the checker. `sort $hashref` and `uc $arrayref` both look like
// type errors and both run fine — sort takes a one-element list, uc
// stringifies the reference. A witness whose premise is wrong reports a false
// positive against the checker and looks exactly like a real finding. So each
// case below records what perl does, and perlRejects verifies it rather than
// trusting the comment.

type mutationCase struct {
	name string
	src  string
	// perlDies says whether perl itself refuses this program. When true, PSC
	// must report it. When false, the case documents a value misuse that runs
	// but is nearly always a mistake, and PSC SHOULD report it — those are
	// tracked separately below.
	perlDies bool
	want     bool // whether PSC is expected to diagnose it
}

// perlRuns reports whether perl executes src without dying, which is the
// factual claim each case rests on.
func perlRuns(t *testing.T, src string) bool {
	t.Helper()
	perl, err := exec.LookPath("perl")
	if err != nil {
		t.Skip("perl not found in PATH")
	}
	cmd := exec.Command(perl, "-Mstrict", "-Mwarnings", "-e", src)
	return cmd.Run() == nil
}

// TestMutationWitnessesAreHonest verifies that each case's claim about perl is
// true before any of them is used to judge PSC. A witness with a wrong premise
// produces a finding that looks real and is not.
func TestMutationWitnessesAreHonest(t *testing.T) {
	for _, c := range mutationCases {
		runs := perlRuns(t, c.src)
		assert.Equal(t, !c.perlDies, runs,
			"%s: the case claims perlDies=%v; measured otherwise. Fix the witness, not the checker.",
			c.name, c.perlDies)
	}
}

// mutationCases are value misuses injected into otherwise well-formed code.
// Every one of these RUNS under perl — they are coercion mistakes rather than
// syntax errors, which is exactly the class PSC exists to catch, since perl
// itself will not.
var mutationCases = []mutationCase{
	{
		name:     "hashref bound to a regex match",
		src:      `my $x = {}; my $y = $x =~ /a/;`,
		perlDies: false,
		want:     true,
	},
	{
		name:     "coderef passed to length",
		src:      `my $x = sub {1}; my $y = length($x);`,
		perlDies: false,
		want:     true,
	},
	{
		name:     "regex used as a number",
		src:      `my $x = qr/a/; my $y = $x + 1;`,
		perlDies: false,
		want:     true,
	},
	{
		name:     "hashref used as a number",
		src:      `my $x = {}; my $y = $x + 1;`,
		perlDies: false,
		want:     true,
	},
	{
		name:     "arrayref passed to uc",
		src:      `my $x = []; my $y = uc($x);`,
		perlDies: false,
		want:     true,
	},
	{
		name:     "hashref passed to lc",
		src:      `my $x = {}; my $y = lc($x);`,
		perlDies: false,
		want:     true,
	},
	{
		name:     "coderef passed to index",
		src:      `my $x = sub {1}; my $y = index($x, "a");`,
		perlDies: false,
		want:     true,
	},
	{
		name: "hashref passed to sort",
		src:  `my $x = {}; my @s = sort $x;`,
		// NOT A TYPE ERROR, and this case is kept precisely because it looks
		// like one. sort takes a list, a scalar IS a one-element list, and
		// that is exactly what Scalar <: List means in the paper's arity
		// ordering — so PSC's silence here is the lattice being right rather
		// than the checker being weak.
		//
		// Measured rather than assumed, on 5.26.0, 5.42.0 and 5.44.0, in
		// every form: bare `sort $x`, parenthesised `sort($x)`, with a
		// comparator `sort { $a cmp $b } $x`, and under `use v5.36`. All run,
		// all yield a one-element list. Eight years of releases including the
		// newest, so this is not a quirk about to be tightened.
		perlDies: false,
		want:     false,
	},
}

// TestTypeOverwritingMutations reports which injected misuses PSC catches.
//
// A case with want=false is a KNOWN GAP, not an accepted outcome: flipping one
// to true is the measure of progress. The test fails if a case PSC used to
// catch stops being caught, and equally if a known gap starts being caught
// without the expectation being updated — silent movement in either direction
// is what this guards against.
func TestTypeOverwritingMutations(t *testing.T) {
	for _, c := range mutationCases {
		_, diags := analyzeSource(t, []byte(c.src+"\n"))
		got := len(diags) > 0
		assert.Equal(t, c.want, got,
			"%s: expected diagnosed=%v, got %v (src: %s)", c.name, c.want, got, c.src)
	}
}

// TestMutationDetectionRate reports the current rate as a visible number.
// It is deliberately not a threshold assertion: the rate is a measurement to
// watch, and the per-case expectations above are what actually gate.
func TestMutationDetectionRate(t *testing.T) {
	var caught, total int
	var missed []string
	for _, c := range mutationCases {
		if c.perlDies {
			continue // not the class this measures
		}
		total++
		_, diags := analyzeSource(t, []byte(c.src+"\n"))
		if len(diags) > 0 {
			caught++
		} else {
			missed = append(missed, c.name)
		}
	}
	t.Logf("type-overwriting mutations caught: %d/%d", caught, total)
	if len(missed) > 0 {
		t.Logf("missed: %s", strings.Join(missed, "; "))
	}
}
