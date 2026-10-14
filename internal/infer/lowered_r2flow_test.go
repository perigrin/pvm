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

// Where `ref $x` is false $x is any scalar but a reference: perl 5.42 -w
// runs the else branch for "Foo", undef, 3, 1.5, *STDOUT and !!0 alike.
// So an untyped scalar narrowed there is still one of unknown content, as
// it is before the guard, and passes a Str slot. perl is silent on every
// case below, the first B::Concise's and B::Showlex's. The control is an
// undef that reaches the else branch, which perl warns on.
func TestNegatedRefGuardNarrowsToNonRef(t *testing.T) {
	for _, src := range []string{
		"my @args = (\"Foo\", \"x::y\");\nfor my $objname (@args) {\n    if (ref $objname) { } else { $objname = \"main::\" . $objname unless $objname =~ /::/; }\n}\n",
		"my @args = (\"Foo\", \\\"x\");\nfor my $objname (@args) {\n    if (ref $objname) { $objname = \"main::\" . $objname unless $objname =~ /::/; }\n}\n",
		"for my $x (\"a\") { if (!ref $x) { print \"main::\" . $x } }\n",
		"for my $x (\"a\") { unless (ref $x) { print \"main::\" . $x } }\n",
		"for my $x (\"a\") { if (defined $x and !ref $x) { print \"main::\" . $x } }\n",
		"sub f { my ($x) = @_; if (ref $x) { } else { print \"main::\" . $x } }\nf('a');\n",
	} {
		wantSilent(t, src)
	}
	wantReported(t, "my $x = undef;\nif (ref $x) { } else { print \"main::\" . $x }\n", "coercion-mismatch")
}
