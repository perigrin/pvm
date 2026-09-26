// ABOUTME: The three word operators below assignment — `and`, `or`, `xor` — parsed both parenthesised and bare.
// ABOUTME: The paren is not a style choice: it changes which program perl compiles, so both shapes are asserted.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestParenthesisedWordInfixParses: `($a and $b)` is an ordinary `&&`.
//
// Measured on perl 5.42.0 -- the paren regroups, so the whole expression is
// the initialiser:
//
//	$ perl -MO=Deparse -e 'my $a=1; my $b=0; my $y = ($a and $b);'
//	my $y = $a && $b;
//
// These refused with `not_a_term` because parseParenList parsed each element
// at the COMMA's binding power, which is above levels 4 and 5, so the word
// operator was never reached as an infix. `cmp` and `x` bind tighter than
// the comma and were clean for that reason alone.
func TestParenthesisedWordInfixParses(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want string
	}{
		{"my $y = ($a and $b);", "(my $y (and $a $b))"},
		{"my $y = ($a or $b);", "(my $y (or $a $b))"},
		{"my $y = ($a xor $b);", "(my $y (xor $a $b))"},
	} {
		root := parseOneExpr(t, tc.src)
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: refused, want a parse", tc.src)
			continue
		}
		assertShape(t, root, tc.want)
	}

	// Not every parenthesised occurrence is an initialiser. Nothing about
	// this shape needs an assignment to its left, and the measurement that
	// filed the issue used this form to prove it.
	root := parseOneExpr(t, "print(($a and 1), \"\\n\");")
	if containsKind(root, parse.Unknown) {
		t.Errorf("print(($a and 1), ...): refused, want a parse")
	}
}

// TestBareWordInfixStillRegroups: without the paren, the assignment wins.
//
// This is the other half of the same cliff, and it is the reason the paren
// matters. `and` binds LOOSER than `=`, so the assignment takes only `$a`
// and the `and` sits above it:
//
//	$ perl -MO=Deparse -e 'my $a=1; my $b=0; my $y = $a and $b;'
//	$b if my $y = $a;
//
// So the bare and parenthesised forms are different programs, and a fix that
// made one parse by flattening it into the other would be a regression
// dressed as a pass. The declarator's initialiser must therefore stop at
// level 4/5 rather than swallow it.
func TestBareWordInfixStillRegroups(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want string
	}{
		{"my $y = $a and $b;", "(and (my $y $a) $b)"},
		{"my $y = $a or $b;", "(or (my $y $a) $b)"},
		{"my $y = $a xor $b;", "(xor (my $y $a) $b)"},
	} {
		root := parseOneExpr(t, tc.src)
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: refused, want a parse", tc.src)
			continue
		}
		assertShape(t, root, tc.want)
	}

	// The COMMA is not one of the three and must still be reached from an
	// initialiser. Stopping the initialiser at the assignment's own right
	// power excluded it too and made this refuse, which is legal perl:
	//
	//	$ perl -MO=Deparse -e 'my $x = 1, my $y = 2;'
	//	my $x = 1, my $y = 2;
	//
	// So the floor is bpBelowComma, and this is the case that says why.
	root := parseOneExpr(t, "my $x = 1, my $y = 2;")
	if containsKind(root, parse.Unknown) {
		t.Errorf("my $x = 1, my $y = 2: refused, want a parse")
	}
}
