// ABOUTME: Tests for AnalyzeLowered's guard narrowing on the round-2 cutover false positives: defined and ref guards.
// ABOUTME: Each silent case is one perl 5.42 -w is silent on, and each control one perl warns about.

package infer_test

import "testing"

// A branch a guard proves unreachable contributes nothing to the join
// after the conditional: $x keeps the type it had, a named one. perl 5.42
// -w is silent on every case below, the first two JSON::PP's; a sub's
// parameter is the broad case. The control reads $ch where the defined
// guard failed, which perl warns on.
func TestDefinedGuardLeavesANamedType(t *testing.T) {
	for _, src := range []string{
		"my $ch = substr(\"1.e\", 0, 1);\nif (defined $ch and $ch eq '.') { if (!defined $ch or $ch !~ /\\d/) { } }\nif (defined $ch and ($ch eq 'e' or $ch eq 'E')) { print \"exp\\n\" }\n",
		"my $x = \"a\";\nif (!defined $x) { print 1 }\nprint $x eq 'a';\n",
		"my $x = \"a\";\nif (!defined $x) { print 1 } elsif ($x) { print 2 }\nprint $x eq 'a';\n",
		"sub f { my ($x) = @_; if (!defined $x) { print 1 } print $x eq 'a' }\nf('a');\n",
		"my $ch;\nsub next_chr { $ch = substr(\"e\", 0, 1) } next_chr();\nif (defined $ch and ($ch eq 'e' or $ch eq 'E')) { print \"exp\\n\" }\n",
	} {
		wantSilent(t, src)
	}
	wantReported(t, "my $ch = undef;\nif (!defined $ch) { print $ch eq 'e' }\n", "coercion-mismatch")
}
