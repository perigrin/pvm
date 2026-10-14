// ABOUTME: Tests for the call rules of AnalyzeLowered: which calls are checked as a CORE builtin, and how.
// ABOUTME: Each case is a program perl 5.42 -w runs silently, paired with a control perl rejects or warns about.

package infer_test

import (
	"testing"

	"tamarou.com/pvm/internal/infer"
)

// wantSilent fails t for every diagnostic the lowered path gives src.
func wantSilent(t *testing.T, src string) {
	t.Helper()
	for _, d := range loweredDiags(src) {
		t.Errorf("%q: diagnostic %q, want none", src, d.Message)
	}
}

// wantCode fails t unless the lowered path gives src a diagnostic of code.
func wantCode(t *testing.T, src, code string) {
	t.Helper()
	diags := loweredDiags(src)
	for _, d := range diags {
		if d.Code == code {
			return
		}
	}
	t.Errorf("%q: diagnostics %v, want a %s", src, diags, code)
}

// `&name` never calls CORE: perl 5.42 -w runs `sub getgrgid { "user:@_" }
// sub lookup { &getgrgid } print lookup(7)` and prints "user:7", the sub
// handed the caller's @_. Without the ampersand perl takes the builtin and
// dies at compile time, "Not enough arguments for getgrgid".
func TestAmpersandCallIsTheUserSub(t *testing.T) {
	for _, src := range []string{
		"sub getgrgid { \"user:@_\" }\nsub lookup { &getgrgid }\nprint lookup(7), \"\\n\";\n",
		"sub getgrnam { \"g:@_\" } sub getgrgid { \"i:@_\" }\nsub getgr :prototype($) { ($_[0] =~ /^\\d+/) ? &getgrgid : &getgrnam }\nprint getgr(\"root\"), \"\\n\";\n",
		"sub gethostbyaddr { \"h\" }\nmy $h = &gethostbyaddr(\"a\");\n",
	} {
		wantSilent(t, src)
	}

	wantCode(t, "sub getgrgid { \"u\" }\nprint getgrgid(), \"\\n\";\n", infer.CodeArityMismatch)
}

// die throws any defined scalar: perl 5.42 -w runs `eval { die E->new };
// print ref($@)` and prints "E", and `die {code => 1}` leaves a HASH in $@.
// An undefined one is what perl warns about, "Use of uninitialized value
// $x in die", so it is still reported.
func TestDieTakesAnObject(t *testing.T) {
	for _, src := range []string{
		"package E; sub new { bless {}, shift }\npackage main;\neval { die E->new }; print ref($@), \"\\n\";\n",
		"eval { die { code => 1 } };\n",
		"eval { die \"a\", \"b\" };\n",
		"eval { die 1 < 2 };\n",
		"sub f { die $_[0] }\n",
		"die(;",
	} {
		wantSilent(t, src)
	}

	wantCode(t, "eval { die undef };\n", infer.CodeTypeMismatch)
}
