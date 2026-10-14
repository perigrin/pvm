// ABOUTME: Tests for AnalyzeLowered's flow rules: guard narrowing inside logical operators, and what a join or a sub reads.
// ABOUTME: Each silent case is one perl 5.42 -w is silent on, and each control one perl warns about or that is truly wrong.

package infer_test

import (
	"testing"

	"tamarou.com/pvm/internal/infer"
)

// wantSilent fails for each diagnostic src draws on the lowered path.
func wantSilent(t *testing.T, src string) {
	t.Helper()
	for _, d := range loweredDiags(src) {
		t.Errorf("%q: diagnostic %q, want none", src, d.Message)
	}
}

// wantReported fails unless src draws exactly one diagnostic of code.
func wantReported(t *testing.T, src string, code string) {
	t.Helper()
	if diags := loweredDiags(src); len(diags) != 1 || diags[0].Code != code {
		t.Errorf("%q: diagnostics %v, want one %s", src, diags, code)
	}
}

// perl short-circuits `&&` and `and`, so their right operand runs only
// where the left is true, and `||` and `or` theirs only where it is false.
// perl 5.42 -w is silent on each case below: the right operand never sees
// $ch undef, and JSON::PP writes every one of them.
func TestDefinedAndNarrowsRightOperand(t *testing.T) {
	for _, src := range []string{
		"my $ch = undef;\nif (defined $ch and $ch eq '}') { print \"close\\n\" }\n",
		"my $ch = undef;\nif (defined $ch && $ch eq '}') { print \"close\\n\" }\n",
		"my $ch = undef;\nmy $at = 1;\n$at-- if defined $ch and $ch ne '';\n",
		"my $ch = undef;\nif (!defined $ch or $ch ne ')') { print \"x\\n\" }\n",
		"my $ch = undef;\nif (!defined $ch || $ch ne ')') { print \"x\\n\" }\n",
		// A broad Scalar narrowed by defined, and a Hole's Any, may
		// still hold a Str.
		"my $ch;\nif (defined $ch) { if (defined $ch and $ch eq '}') { print \"close\\n\" } }\n",
		"my $ch = eval 'undef';\nif (defined $ch and $ch eq '}') { print \"close\\n\" }\n",
	} {
		wantSilent(t, src)
	}

	// perl -w warns "Use of uninitialized value $ch in string eq" on
	// each: the right operand runs with $ch undef.
	for _, src := range []string{
		"my $ch = undef;\nif ($ch eq '}') { print \"c\\n\" }\n",
		"my $ch = undef;\nif (defined $ch or $ch eq '}') { print \"c\\n\" }\n",
		"my $ch = undef;\nif (!defined $ch and $ch eq '}') { print \"c\\n\" }\n",
	} {
		wantReported(t, src, infer.CodeCoercionMismatch)
	}
}

// What a logical operator's right operand assigns reaches the code after
// it, joined with the path that skipped the right operand, on which the
// left operand's guard held as it was tested. perl 5.42 -w is silent on
// `defined $x or $x = 'a'` followed by `$x eq 'a'`, and warns "Use of
// uninitialized value $x in string eq" when `and` assigns only where $x
// was already defined.
func TestLogicalRightAssignmentJoinsSkippedPath(t *testing.T) {
	wantSilent(t, "my $x = undef;\ndefined $x or $x = 'a';\nprint \"y\\n\" if $x eq 'a';\n")
	wantSilent(t, "my $x = undef;\n!defined $x and $x = 'a';\nprint \"y\\n\" if $x eq 'a';\n")
	wantSilent(t, "my $c = 0;\nmy $x = $c ? 'a' : undef;\ndefined $x or $x = 'b';\nprint \"y\\n\" if $x eq 'b';\n")
	wantReported(t, "my $x = undef;\ndefined $x and $x = 'b';\nprint \"y\\n\" if $x eq 'a';\n", infer.CodeCoercionMismatch)
}

// A sub reads a file-scope variable's value when it is called, which any
// assignment to it may have set: in another sub, or at file scope after
// the sub. perl 5.42 -w is silent on each case below (Sys::Syslog's
// $current_proto is the first), and prints "n".
func TestFileScopeUndefReadInSubIsNotUndef(t *testing.T) {
	for _, src := range []string{
		"my $proto = undef;\nsub log_it { if ($proto eq 'native') { print \"n\\n\" } }\nsub connect_log { $proto = 'native' }\nconnect_log(); log_it();\n",
		"my $proto = undef;\nsub log_it { if ($proto eq 'native') { print \"n\\n\" } }\n$proto = 'native';\nlog_it();\n",
		"my $proto = undef;\nsub connect_log { $proto = 'native' }\nsub log_it { if ($proto eq 'native') { print \"n\\n\" } } connect_log(); log_it();\n",
		"my $proto = undef;\nsub connect_log { $proto = 'native' }\nsub log_it { connect_log(); print \"n\\n\" if $proto eq 'native' } log_it();\n",
	} {
		wantSilent(t, src)
	}

	// Nothing assigns $proto but its declaration, so the sub reads undef:
	// perl -w warns "Use of uninitialized value $proto in string eq".
	wantReported(t, "my $proto = undef;\nsub log_it { if ($proto eq 'native') { print \"n\\n\" } }\nlog_it();\n", infer.CodeCoercionMismatch)
}
