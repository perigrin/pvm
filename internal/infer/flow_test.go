// ABOUTME: Tests for Analyze's flow rules: guard narrowing inside logical operators, and what a join or a sub reads.
// ABOUTME: Each silent case is one perl 5.42 -w is silent on, and each control one perl warns about or that is truly wrong.

package infer_test

import (
	"testing"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/types"
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

// A value PSC does not type, a line read from a handle, still replaces
// what the variable held, so it no longer holds the Undef it was narrowed
// to. perl 5.42 -w is silent on each case below, given "42" on STDIN:
// Unicode::UCD reads its $UNICODEVERSION so. Assigning a typed undef is
// still reported, as perl warns "Use of uninitialized value $V in pattern
// match".
func TestUntypedAssignmentForgetsUndef(t *testing.T) {
	wantSilent(t, "my $V = undef;\n$V = <STDIN>;\nprint \"y\\n\" if $V =~ /^\\d/;\n")
	wantSilent(t, "my $V;\nsub v { unless (defined $V) { chomp($V = <STDIN>); die \"bad\" unless $V =~ /^\\d/; } return $V }\nprint v(), \"\\n\";\n")
	wantReported(t, "my $V = 1;\n$V = undef;\nprint \"y\\n\" if $V =~ /^\\d/;\n", infer.CodeCoercionMismatch)
}

// A package variable may be assigned by any code, through its full name
// or a symbolic reference, as Math::BigInt's downgrade() sets
// ${"${class}::downgrade"}, so a sub reads it at its sigil type. perl 5.42
// -w is silent on each case below.
func TestBlessClassFromAssignedPackageVar(t *testing.T) {
	for _, src := range []string{
		"our $downgrade = undef;\nsub f { my $self = bless {}, \"main\"; bless $self, $downgrade if defined $downgrade; $self }\nf();\n",
		"package Foo;\nour $downgrade = undef;\nsub f { my $self = bless {}, \"Foo\"; bless $self, $downgrade; $self }\npackage main;\n$Foo::downgrade = \"Bar\";\nprint ref(Foo::f()), \"\\n\";\n",
		"package Foo;\nour $downgrade = undef;\nsub downgrade { no strict 'refs'; my $c = shift; ${\"${c}::downgrade\"} = shift if @_; ${\"${c}::downgrade\"} }\nsub _dng { my $self = shift; my $class = ref($self); my $downgrade = $class->downgrade(); return $self unless $downgrade; bless $self, $downgrade; return $self }\npackage main;\nFoo->downgrade(\"Bar\");\nprint ref(Foo::_dng(bless {}, \"Foo\")), \"\\n\";\n",
	} {
		wantSilent(t, src)
	}

	// A lexical nothing else assigns is undef in the sub: perl -w warns
	// "Use of uninitialized value $d in bless".
	wantReported(t, "my $d = undef;\nsub f { my $self = bless {}, \"main\"; bless $self, $d; $self }\nf();\n", infer.CodeTypeMismatch)
}

// What a scalar variable holds after a branch join is a scalar value: a
// Hole's Any, or eval STRING's, narrowed by a guard and assigned to it,
// is never Void, an array or a hash there. perl 5.42 -w is silent on each
// case below (Math::Complex sets its $Inf so), and prints Inf.
func TestBranchJoinNeverHoldsVoid(t *testing.T) {
	for _, src := range []string{
		"our ($Inf);\nmy $i = eval \"9**9**9\";\nif (defined $i && $i > 1e308) { $Inf = $i }\n$Inf = 1.79e308 unless defined $Inf;\n",
		"my $Inf;\nmy $i = eval \"9**9**9\";\nif (defined $i && $i > 1e308) { $Inf = $i }\n$Inf = 1.79e308 unless defined $Inf;\n",
	} {
		wantSilent(t, src)
	}
	// Where Any failed a defined test it is undef: perl -w warns "Use of
	// uninitialized value $i in string eq".
	wantReported(t, "my $i = eval \"undef\";\nif (!defined $i) { print \"y\\n\" if $i eq 'x' }\n", infer.CodeCoercionMismatch)

	src := "my $Inf;\nmy $i = eval \"9**9**9\";\nif (defined $i) { $Inf = $i }\nprint \"$Inf\\n\";\n"
	got, _ := loweredSymbolType(t, src, "$Inf")
	if got != types.Any && got&types.Void != 0 {
		t.Errorf("%q: $Inf is %v, want Any or a union without Void", src, got)
	}
}
