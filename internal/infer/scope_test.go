// ABOUTME: Tests for Analyze's lexical scoping: sub bodies, bare blocks and foreach loops each scope their own `my`.
// ABOUTME: Each silent case is one perl 5.42 -w is silent on, and each control one perl warns about.

package infer_test

import (
	"testing"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// A `my` in a sub body declares a variable of that body, and narrows it,
// not a same-named variable outside the sub. perl 5.42 -w prints 6 and 2
// silently for the first two, and warns "Use of uninitialized value $x in
// addition" on the control, where the sub's own $x is undef.
func TestSubBodyMyDoesNotLeakOutward(t *testing.T) {
	wantSilent(t, "my $x = 5;\nsub f { my $x = undef; return defined $x ? 1 : 0 }\nprint $x + 1, \"\\n\";\n")
	wantSilent(t, "my @a = (1, 2);\nsub g { my @a = (\"x\"); return scalar @a }\nprint $a[0] + 1, \"\\n\";\n")
	wantReported(t, "my $x = 5;\nsub f { my $x = undef; return $x + 1 }\nf();\n", infer.CodeCoercionMismatch)
}

// A `my` in a named sub's body narrows to its initialiser, as one at file
// level does. perl 5.42 -w warns "Use of uninitialized value $x in
// addition" on the first, and is silent on the second.
func TestSubBodyDeclNarrows(t *testing.T) {
	wantReported(t, "sub f { my $x = undef; return $x + 1 }\nf();\n", infer.CodeCoercionMismatch)
	wantSilent(t, "sub f { my $x = 1; return $x + 1 }\nf();\n")

	src := "sub f { my $x = 1; return $x }\n"
	root := parse.Parse([]byte(src))
	f := lower.Lower(root)
	_, _, st := infer.Analyze(f, lower.NewPositions(root, f), nil, infer.Options{})
	if sym, ok := st.LookupAll("$x"); !ok || sym.Type != types.Int {
		t.Errorf("%q: $x is %v (found %v), want Int", src, sym.Type, ok)
	}
}

// A bare block is a scope: its `my` is not the variable outside it. perl
// 5.42 -w prints 6 silently for the first, and warns "Use of
// uninitialized value $x in addition" on the control.
func TestBareBlockIsAScope(t *testing.T) {
	wantSilent(t, "my $x = 5;\n{ my $x = undef; }\nprint $x + 1, \"\\n\";\n")
	wantReported(t, "{ my $x = undef; print $x + 1, \"\\n\"; }\n", infer.CodeCoercionMismatch)
}

// `for my $i ($n)` runs its body once with $i aliased to $n: $i is the
// loop's own variable, and what the body writes to it is written to $n.
// perl 5.42 -w prints 4 silently for the first, and warns "Use of
// uninitialized value $n in addition" on the control, whose body set $n
// undef through $j.
func TestForMyDefinesItsVariable(t *testing.T) {
	wantSilent(t, "my $n = 3;\nmy $i = undef;\nfor my $i ($n) { print $i + 1, \"\\n\"; }\n")
	wantReported(t, "my $n = 3;\nforeach my $j ($n) { $j = undef; }\nprint $n + 1, \"\\n\";\n", infer.CodeCoercionMismatch)
}

// A `for my $i` loop scopes $i: the outer $i is not the loop's, and the
// body's writes to $i do not reach it. perl 5.42 -w prints 6 silently for
// the first two, and warns "Use of uninitialized value $i in addition" on
// the control, which reads the loop's own $i after setting it undef.
func TestForMyScopesItsVariable(t *testing.T) {
	wantSilent(t, "my $i = 5;\nmy $n = 3;\nfor my $i ($n) { $i = undef; }\nprint $i + 1, \"\\n\";\n")
	wantSilent(t, "my $i = 5;\nmy @l = (1, 2);\nfor my $i (@l) { $i = undef; }\nprint $i + 1, \"\\n\";\n")
	wantReported(t, "my $n = 3;\nfor my $i ($n) { $i = undef; print $i + 1, \"\\n\"; }\n", infer.CodeCoercionMismatch)
}
