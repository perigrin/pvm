// ABOUTME: Holds the oracle to perl's own verdict on files whose BEGIN blocks print, warn-die, exit or die.
// ABOUTME: A file perl compiles is answered; one perl rejects is declined; one that exits early is never exact.

package parseoracle_test

import (
	"context"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
	"tamarou.com/pvm/internal/perltest"
)

// TestOracleAnswersFileThatPerlCompiles is op/magic.t reduced. perl -c says
// "syntax OK", yet the oracle declined the file. The cause is the op-list
// run: parse_facts.pl takes "did it compile" from `-MO=Concise,-exec`, and
// O.pm runs B::Concise in a CHECK block, after the file's BEGIN has
// installed `$SIG{__WARN__} = sub { die ... }`. B::Concise is `use v5.40`,
// so its own sequence() warns "Deep recursion" on a file whose statement
// chain runs through more than 100 LOGOPs -- here, 101 statement-modifier
// ifs -- and the file's handler turns that warning into "CHECK failed--call
// queue aborted". Concise exits 255, and its exit status became ok:0.
// The warning is Concise's, not the file's; perl never raises it.
func TestOracleAnswersFileThatPerlCompiles(t *testing.T) {
	requirePOSIXOracle(t)
	p := perltest.PinnedPerl(t)
	src := "BEGIN { $| = 1; print \"1..1\\nok 1\\n\"; $SIG{__WARN__} = sub { die \"Dying on warning: \", @_ } }\n" + // 1
		"my %h = (a => 1);\n" + // 2
		strings.Repeat("print \"ok\\n\" if $h{a};\n", 101) // 3..103

	facts, err := parseoracle.Ask(context.Background(), []byte(src), parseoracle.Options{Perl: p.Path})
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if !facts.OK {
		t.Fatalf("perl compiles this file, but the oracle reported ok:0; stderr: %s", facts.Stderr)
	}
	if facts.OpCount == 0 || !facts.Walked {
		t.Fatalf("the oracle answered without facts: op_count %d, walked %t", facts.OpCount, facts.Walked)
	}
	if v := parseoracle.CompareFacts(facts, answeringSubject()); v.Bucket == parseoracle.BucketNoAnswer {
		t.Errorf("verdict = %s (%s), want an answer", v.Bucket, v.Detail)
	}
}

// TestOracleStillDeclinesFileDyingInBEGIN: perl rejects a file whose BEGIN
// dies (perl -c: "BEGIN failed--compilation aborted", exit 255), so the
// oracle must decline it too.
func TestOracleStillDeclinesFileDyingInBEGIN(t *testing.T) {
	requirePOSIXOracle(t)
	p := perltest.PinnedPerl(t)
	src := "BEGIN { $SIG{__WARN__} = sub { die @_ } }\n" +
		"BEGIN { die \"no\\n\" }\n" +
		"my $h = { a => 1 };\n"

	facts, err := parseoracle.Ask(context.Background(), []byte(src), parseoracle.Options{Perl: p.Path})
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if facts.OK {
		t.Fatalf("perl rejects this file, but the oracle reported ok:1")
	}
	if !strings.Contains(facts.Stderr, "BEGIN failed") {
		t.Errorf("stderr = %q, want perl's BEGIN failure", facts.Stderr)
	}
	if v := parseoracle.CompareFacts(facts, answeringSubject()); v.Bucket != parseoracle.BucketNoAnswer {
		t.Errorf("verdict = %s, want no-answer", v.Bucket)
	}
}

// TestOracleExitInBEGINIsNotExact: a file that exits 0 from BEGIN, as
// skip_all does, compiles "successfully" having built no optree. Nothing
// was parsed, so nothing can agree with it.
func TestOracleExitInBEGINIsNotExact(t *testing.T) {
	requirePOSIXOracle(t)
	p := perltest.PinnedPerl(t)
	src := "BEGIN { $SIG{__WARN__} = sub { die @_ } }\n" +
		"BEGIN { print \"1..0 # skip\\n\"; exit 0 }\n" +
		"my $h = { a => 1 };\n"

	facts, err := parseoracle.Ask(context.Background(), []byte(src), parseoracle.Options{Perl: p.Path})
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if facts.OpCount != 0 || len(facts.Ops) != 0 {
		t.Fatalf("op_count %d, ops %d: want no optree from a file that exits in BEGIN", facts.OpCount, len(facts.Ops))
	}
	if v := parseoracle.CompareFacts(facts, answeringSubject()); v.Bucket == parseoracle.BucketExact {
		t.Errorf("verdict = exact, want anything else (%s)", v.Detail)
	}
}

// TestOracleStderrFromBEGINIsNotAFailure: a file that writes to STDERR
// while compiling, and compiles, is answered. Output is not an error.
func TestOracleStderrFromBEGINIsNotAFailure(t *testing.T) {
	requirePOSIXOracle(t)
	p := perltest.PinnedPerl(t)
	src := "BEGIN { print STDERR \"# syntax error at line 1, near nothing\\n\"; print \"ok 1\\n\" }\n" +
		"my $h = { a => 1 };\n"

	facts, err := parseoracle.Ask(context.Background(), []byte(src), parseoracle.Options{Perl: p.Path})
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if !facts.OK || !facts.Walked {
		t.Fatalf("ok %t, walked %t: want an answered file; stderr: %s", facts.OK, facts.Walked, facts.Stderr)
	}
	if k := parseoracle.Classify(facts.Stderr); k != parseoracle.NoFailure {
		t.Errorf("Classify = %s, want NoFailure", k)
	}
	if got := facts.Sites[parseoracle.MarkerAnonhash]; len(got) != 1 || got[0] != 2 {
		t.Errorf("Sites[anonhash] = %v, want [2]", got)
	}
}

// answeringSubject is a subject that parsed the file and reports no call
// sites, so any verdict short of no-answer is the oracle's doing.
func answeringSubject() parseoracle.SubjectFacts {
	return parseoracle.SubjectFacts{OK: true, KnowsCallSites: true}
}
