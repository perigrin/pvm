// ABOUTME: Tests for a sub's return type and user guards: which returns are the sub's own, and what each says.
// ABOUTME: Each silent case is one perl 5.42 -w is silent on, and each control one perl warns about.

package infer_test

import (
	"testing"

	"tamarou.com/pvm/internal/infer"
)

// A sub one of whose returns, or whose fall-through value, is of no known
// type has no known return type: the arms that are known do not say what
// it returns. perl 5.42 -w prints 5 silently for the first, and warns
// "Use of uninitialized value $u in addition" on the control, whose only
// return is undef.
func TestReturnTypeKeepsUnknownArms(t *testing.T) {
	wantSilent(t, "sub k { my $x = shift; return undef unless $x; eval { 4 } }\nmy $u = k(3);\nprint $u + 1, \"\\n\";\n")
	wantReported(t, "sub k { my $x = shift; return undef }\nmy $u = k(3);\nprint $u + 1, \"\\n\";\n", infer.CodeCoercionMismatch)
}

// A return in a sub nested in another, anonymous or named, returns from
// the nested sub, so it is no arm of the outer sub's return type, nor of
// what a guard or a builtin-result judgement reads of it. perl 5.42 -w
// prints 2 silently for the first two, and warns "Use of uninitialized
// value in addition" on the control, whose own return is undef.
func TestNestedSubReturnDoesNotLeak(t *testing.T) {
	wantSilent(t, "sub outer { my $cb = sub { return undef }; return 1 }\nmy $r = outer() + 1;\nprint \"$r\\n\";\n")
	wantSilent(t, "sub outer { sub inner { return undef } return 1 }\nmy $r = outer() + 1;\nprint \"$r\\n\";\n")
	wantSilent(t, "sub fd { sub one { return 1 } return fileno $_[0] }\nprint fd(\\*STDOUT) . \">\\n\";\n")
	wantReported(t, "sub outer { my $cb = sub { return 1 }; return undef }\nmy $r = outer() + 1;\nprint \"$r\\n\";\n", infer.CodeCoercionMismatch)
}

// A user guard narrows only its parameter, and only through a returned
// expression that is every value the sub hands back. chk tests its own
// $t, not the caller's, and isnull returns 1 whenever $strict is false,
// so neither narrows the caller's Undef|Int to Undef: perl 5.42 -w prints
// 6 for each with @ARGV empty, and warns "Use of uninitialized value" in
// the addition with an argument, which is the Undef|Int psc reports. The
// guard on a parameter still narrows: perl warns "Use of uninitialized
// value $w in addition" on the control with @ARGV empty, and psc reports
// the Undef.
func TestUserGuardOnlyNarrowsParameters(t *testing.T) {
	wantReportedWith(t, "sub chk { my $v = shift; my $t = $v; return !defined $t }\nmy $t = @ARGV ? undef : 5;\nmy $w;\nif (chk($w)) { my $q = $t + 1; print \"$q\\n\" }\n", infer.CodeCoercionMismatch, "Undef|Int")
	wantReportedWith(t, "our $strict = 0;\nsub isnull { my $v = shift; return !defined $v if $strict; 1 }\nmy $y = @ARGV ? undef : 5;\nif (isnull($y)) { my $q = $y + 1; print \"$q\\n\" }\n", infer.CodeCoercionMismatch, "Undef|Int")
	wantReportedWith(t, "sub chk { my $v = shift; return !defined $v }\nmy $w = @ARGV ? 5 : undef;\nif (chk($w)) { my $q = $w + 1; print \"$q\\n\" }\n", infer.CodeCoercionMismatch, "got Undef")
}
