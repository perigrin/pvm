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
