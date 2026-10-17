// ABOUTME: Tests for Analyze's flow rules: guard narrowing inside logical operators, and what a join or a sub reads.
// ABOUTME: Each silent case is one perl 5.42 -w is silent on, and each control one perl warns about or that is truly wrong.

package infer_test

import (
	"testing"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
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

// A compound guard with one side that is not a guard keeps its operator:
// `defined $x && $y` failing says nothing about $x, since $y may be what
// failed, and `!defined $x || $y` holding says nothing about $x either.
// perl 5.42 -w, run with the argument 5, is silent on each and prints 6.
func TestPartialCompoundGuardKeepsOperator(t *testing.T) {
	for _, src := range []string{
		"my $x = $ARGV[0];\nmy $y = 0;\nif (defined $x && $y) { exit 0 }\nprint $x + 1, \"\\n\";\n",
		"my $x = $ARGV[0];\nmy $y = 0;\n(defined $x && $y) || print $x + 1, \"\\n\";\n",
		"my $x = $ARGV[0];\nmy $y = 1;\nif (!defined $x || $y) { print $x + 1, \"\\n\" }\n",
		"my $x = $ARGV[0];\nmy $y = 0;\nif (!defined $x || $y) { exit 0 }\nprint $x + 1, \"\\n\";\n",
	} {
		wantSilent(t, src)
	}

	// The known side still narrows where the operator lets it: perl -w,
	// run with no argument, warns "Use of uninitialized value $x in
	// addition (+)" on each.
	for _, src := range []string{
		"my $x = $ARGV[0];\nmy $y = 1;\nif (!defined $x && $y) { print $x + 1, \"\\n\" }\n",
		"my $x = $ARGV[0];\nmy $y = 0;\nif (defined $x || $y) { } else { print $x + 1, \"\\n\" }\n",
	} {
		wantReported(t, src, infer.CodeCoercionMismatch)
	}
}

// An `or` whose other side is not a guard narrows nothing in its
// then-branch, which runs whenever the other side is true; its
// else-branch runs only where both sides are false, so the guard's
// negation holds there. perl 5.42 -w is silent on the first two, run
// with the argument "a" and then 5.
func TestOrWithNonGuardSideDoesNotNarrowThen(t *testing.T) {
	wantSilent(t, "my $ch = $ARGV[0];\nif (!defined $ch or $ch !~ /\\d/) { print \"a\" . $ch, \"\\n\" }\n")
	wantSilent(t, "my $ch = $ARGV[0];\nif (!defined $ch or $ch !~ /\\d/) { } else { print $ch + 1, \"\\n\" }\n")
	// perl -w, run with no argument, warns "Use of uninitialized value
	// $ch in concatenation (.) or string".
	wantReported(t, "my $ch = $ARGV[0];\nif (!defined $ch) { print \"a\" . $ch, \"\\n\" }\n", infer.CodeCoercionMismatch)
}

// A negated compound negates every level of it, by De Morgan:
// `!((A && B) || C)` is `(!A || !B) && !C`, which narrows nothing from A
// or B. perl 5.42 -w, given 1 and 3, is silent on the first and prints 4.
func TestNegateGuardIsDeep(t *testing.T) {
	wantSilent(t, "my $a1 = $ARGV[0];\nmy $b1 = $ARGV[1];\nmy $c1 = undef;\nif (!((!defined $a1 && !defined $b1) || defined $c1)) { print $b1 + 1, \"\\n\" }\n")
	// `!((B || A) || C)` is `!B && !A && !C`, so $b1 is undef in the
	// branch: perl -w, given no argument, warns "Use of uninitialized
	// value $b1 in concatenation (.) or string".
	wantReported(t, "my $a1 = undef;\nmy $b1 = $ARGV[0];\nmy $c1 = undef;\nif (!((defined $b1 || defined $a1) || defined $c1)) { print \"a\" . $b1, \"\\n\" }\n", infer.CodeCoercionMismatch)
}

// A guard tests a scalar, so a variable typed Any or Unknown narrows from
// Scalar: what `!ref $v` leaves of it is a Scalar less its references,
// which a Num or Str slot takes, not a union of Array, Hash, IO and Void.
// perl 5.42 -w is silent on each and prints 2, a1 and a1.
func TestNegatedRefGuardOnAny(t *testing.T) {
	for _, src := range []string{
		"my $v = eval \"1\";\nunless (ref $v) { print $v + 1, \"\\n\" }\n",
		"my $v = eval \"1\";\nif (defined $v && !ref $v) { print \"a\" . $v, \"\\n\" }\n",
		"my $v = eval \"1\";\nif (ref $v) { } else { print \"a\" . $v, \"\\n\" }\n",
	} {
		wantSilent(t, src)
	}
	// perl -w warns "Use of uninitialized value $v in concatenation (.)
	// or string": a failed defined still narrows Any to Undef.
	wantReported(t, "my $v = eval \"undef\";\nunless (defined $v) { print \"a\" . $v, \"\\n\" }\n", infer.CodeCoercionMismatch)
}

// A call is an exit when the signature it resolves to returns None, as
// CORE.tpm declares die and exit, and `CORE::name` resolves to the
// builtin name, as a guard and as a checked call. perl 5.42 -w, given no
// argument, is silent on the first two and prints 4.
func TestExitsAndGuardsByCoreAndCarp(t *testing.T) {
	for _, src := range []string{
		"my $y = @ARGV ? undef : 3;\nunless (CORE::defined $y) { die \"y\" }\nprint $y + 1, \"\\n\";\n",
		"my $y = @ARGV ? undef : 3;\nunless (defined $y) { CORE::die \"y\" }\nprint $y + 1, \"\\n\";\n",
	} {
		wantSilent(t, src)
	}
	// perl -w, given an argument, warns "Use of uninitialized value $y in
	// concatenation (.) or string" on both.
	for _, src := range []string{
		"my $y = @ARGV ? undef : 3;\nunless (CORE::defined $y) { print \"a\" . $y, \"\\n\" }\n",
		"my $y = @ARGV ? undef : 3;\nif (CORE::defined $y) { CORE::exit 0 }\nprint \"a\" . $y, \"\\n\";\n",
	} {
		wantReported(t, src, infer.CodeCoercionMismatch)
	}
	// perl refuses it at compile time: "Type of arg 1 to push must be
	// array (not private hash)".
	wantReported(t, "my %h;\nCORE::push %h, 1;\n", infer.CodeTypeMismatch)

	t.Run("croak", func(t *testing.T) {
		// croak and confess exit as Carp.tpm declares them, returning
		// None, and the parser hands each call the signature it resolved:
		// perl 5.42 -w, given no argument, is silent on both and prints
		// 4. carp returns, so perl, given an argument, warns "Use of
		// uninitialized value $x in addition (+)". A parse that resolves
		// modules reads Carp.tpm; one that finds nothing else on disk is
		// enough.
		diags := func(src string) []infer.Diagnostic {
			root := parse.ParseWithLoader([]byte(src), func(string) ([]byte, bool) { return nil, false })
			f := lower.Lower(root)
			_, d, _ := infer.Analyze(f, lower.NewPositions(root, f), nil, infer.Options{})
			return d
		}
		for _, exit := range []string{"croak", "confess"} {
			src := "use Carp;\nmy $x = @ARGV ? undef : 3;\nunless (defined $x) { " + exit + " \"x\" }\nprint $x + 1, \"\\n\";\n"
			for _, d := range diags(src) {
				t.Errorf("%q: diagnostic %q, want none", src, d.Message)
			}
		}
		src := "use Carp;\nmy $x = @ARGV ? undef : 3;\nunless (defined $x) { carp \"x\" }\nprint $x + 1, \"\\n\";\n"
		if d := diags(src); len(d) != 1 || d[0].Code != infer.CodeCoercionMismatch {
			t.Errorf("%q: diagnostics %v, want one %s", src, d, infer.CodeCoercionMismatch)
		}
	})
}

// A statement-modifier exit, and an `or die`, narrow what follows them as
// the block form does: past `die unless defined $x`, $x is defined. perl
// 5.42 -w, given an argument, is silent on each and prints 4.
func TestModifierGuardNarrowsContinuation(t *testing.T) {
	for _, src := range []string{
		"my $x = @ARGV ? 3 : undef;\ndie \"x\" unless defined $x;\nprint $x + 1, \"\\n\";\n",
		"my $x = @ARGV ? 3 : undef;\ndefined $x or die \"x\";\nprint $x + 1, \"\\n\";\n",
		"my $x = @ARGV ? 3 : undef;\n!defined $x and die \"x\";\nprint $x + 1, \"\\n\";\n",
		"sub f { my $x = $_[0] ? 3 : undef; return 0 unless defined $x; return $x + 1 }\nprint f(1), \"\\n\";\n",
	} {
		wantSilent(t, src)
	}
	// perl -w, given no argument, warns "Use of uninitialized value $x"
	// on each: past an exit the guard failed, and a body that does not
	// exit narrows nothing after it.
	for _, src := range []string{
		"my $x = @ARGV ? 3 : undef;\nCORE::exit 0 if defined $x;\nprint \"a\" . $x, \"\\n\";\n",
		"my $x = @ARGV ? 3 : undef;\ndefined $x and die \"x\";\nprint \"a\" . $x, \"\\n\";\n",
		"my $x = @ARGV ? 3 : undef;\nprint \"n\\n\" unless defined $x;\nprint $x + 1, \"\\n\";\n",
	} {
		wantReported(t, src, infer.CodeCoercionMismatch)
	}
}

// resolvedDiags is the diagnostics src draws on the lowered path from a
// parse that resolves modules, finding none on disk: a `use Carp` reads
// Carp.tpm.
func resolvedDiags(src string) []infer.Diagnostic {
	root := parse.ParseWithLoader([]byte(src), func(string) ([]byte, bool) { return nil, false })
	f := lower.Lower(root)
	_, d, _ := infer.Analyze(f, lower.NewPositions(root, f), nil, infer.Options{})
	return d
}

// A qualified croak or confess ends its path as the bare one does, whether
// Carp is imported or only loaded: perl 5.42 -w, given no argument, is
// silent on each and prints 4.
func TestQualifiedCroakConfessAreExits(t *testing.T) {
	for _, use := range []string{"use Carp;", "use Carp ();"} {
		for _, exit := range []string{"Carp::croak", "Carp::confess"} {
			src := use + "\nmy $x = @ARGV ? undef : 3;\nunless (defined $x) { " + exit + "(\"x\") }\nprint $x + 1, \"\\n\";\n"
			for _, d := range resolvedDiags(src) {
				t.Errorf("%q: diagnostic %q, want none", src, d.Message)
			}
		}
	}
}

// Foo::confess is Foo's own sub, not Carp's, and returns: perl 5.42 -w,
// given an argument, warns "Use of uninitialized value $x in addition (+)".
func TestQualifiedCallOtherPackageConfessNotExit(t *testing.T) {
	src := "use Carp;\npackage Foo { sub confess { 1 } }\nmy $x = @ARGV ? undef : 3;\nunless (defined $x) { Foo::confess(\"x\") }\nprint $x + 1, \"\\n\";\n"
	if d := resolvedDiags(src); len(d) != 1 || d[0].Code != infer.CodeCoercionMismatch {
		t.Errorf("%q: diagnostics %v, want one %s", src, d, infer.CodeCoercionMismatch)
	}
}

// \&Carp::confess takes a reference and calls nothing: perl 5.42 -w, given
// an argument, warns "Useless use of single ref constructor in void
// context" and "Use of uninitialized value $x in addition (+)".
func TestQualifiedCodeRefIsNotExit(t *testing.T) {
	src := "use Carp;\nmy $x = @ARGV ? undef : 3;\nunless (defined $x) { \\&Carp::confess }\nprint $x + 1, \"\\n\";\n"
	if d := resolvedDiags(src); len(d) != 1 || d[0].Code != infer.CodeCoercionMismatch {
		t.Errorf("%q: diagnostics %v, want one %s", src, d, infer.CodeCoercionMismatch)
	}
}
