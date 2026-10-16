// ABOUTME: Tests for Analyze's branch joins: each body of a conditional starts from the types before it, and the code after joins their ends.
// ABOUTME: Each silent case is one perl 5.42 -w is silent on, and each control one perl warns about.

package infer_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/infer"
)

// wantReportedWith fails unless src draws exactly one diagnostic of code
// whose message holds text.
func wantReportedWith(t *testing.T, src, code, text string) {
	t.Helper()
	if diags := loweredDiags(src); len(diags) != 1 || diags[0].Code != code || !strings.Contains(diags[0].Message, text) {
		t.Errorf("%q: diagnostics %v, want one %s holding %q", src, diags, code, text)
	}
}

// What one body of an if writes is not there in another body, and the
// code after the statement joins what each body left. perl 5.42 -w prints
// 2 silently for the first; on the controls, after `$n = undef` ran or
// did not, $n + 1 is Undef|Int, and perl warns "Use of uninitialized
// value $n in addition" whenever @ARGV is not empty.
func TestIfBranchWriteDoesNotLeakIntoElse(t *testing.T) {
	wantSilent(t, "my $n = 1;\nif (@ARGV) { $n = undef } else { print $n + 1, \"\\n\" }\n")
	wantSilent(t, "my $n = 1;\nif (@ARGV) { $n = undef } elsif ($0) { print $n + 1, \"\\n\" }\n")
	wantReportedWith(t, "my $n = 1;\nif (@ARGV) { $n = undef } else { $n = 2 }\nprint $n + 1, \"\\n\";\n", infer.CodeCoercionMismatch, "Undef|Int")
	wantReportedWith(t, "my $y = 3;\nmy $c = 0;\nif ($c) { 1 } elsif ($c > 1) { $y = undef }\nmy $z = $y + 1;\n", infer.CodeCoercionMismatch, "Undef|Int")
}

// An elsif's write to the variable its own guard narrows reaches the join.
// perl 5.42 -w warns "Use of uninitialized value $y in addition" on it,
// with @ARGV empty.
func TestElsifWriteJoins(t *testing.T) {
	src := "my $y = @ARGV ? 3 : {};\nmy $c = 0;\nif ($c) { print \"c\\n\" } elsif (ref $y) { $y = undef }\nmy $z = $y + 1;\n"
	wantReportedWith(t, src, infer.CodeCoercionMismatch, "Undef")
}

// An else its guard does not narrow still joins with the type it leaves:
// after `if ($x isa Foo) {...} else {...}` on an undef $x, $x is undef.
// perl 5.42 -w warns "Use of uninitialized value $x in addition" on the
// first, and prints 6 silently for the second.
func TestElseBranchJoinsUnnarrowed(t *testing.T) {
	wantReported(t, "use v5.36;\nmy $x = undef;\nif ($x isa Foo) { 1 } else { 2 }\nprint $x + 1, \"\\n\";\n", infer.CodeCoercionMismatch)
	wantSilent(t, "my $x = 5;\nif (!defined $x) { print \"u\\n\" } else { print \"d\\n\" }\nprint $x + 1, \"\\n\";\n")
}
