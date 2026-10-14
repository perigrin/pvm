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

// An array or a hash given to a list operator's slurpy slot flattens into
// its elements, and each element is what the slot checks: perl 5.42 -w
// runs `my @parts = ("a", "b"); print @parts; print "x", @parts; warn
// @parts` and `system(@cmd)`, `exec @cmd` and `chomp(@$arr)` silently. An
// array whose elements are undef still is reported: `my @a = (undef,
// undef); print @a` warns "Use of uninitialized value $a[0] in print". A
// `$` slot does not flatten: `length(@a)` warns "length() used on @a".
func TestArrayArgumentToListOpFlattens(t *testing.T) {
	for _, src := range []string{
		"my @parts = (\"a\", \"b\");\nprint @parts; print \"x\", @parts, \"\\n\";\nwarn @parts;\n",
		"sub run { my @cmd = @_; system(@cmd) == 0 or die \"failed\"; exec @cmd if 0 }\nrun(\"true\");\n",
		"my $arr = [\"a\\n\"];\nchomp(@$arr);\n",
		"my %h = (a => 1);\nprint %h, \"\\n\";\n",
		"my @a;\nprint STDERR @a;\n",
		"sub f { print @{$_[0]} }\n",
		"print(;",
		"warn(;",
	} {
		wantSilent(t, src)
	}

	wantCode(t, "my @a = (undef, undef);\nprint @a, \"\\n\";\n", infer.CodeTypeMismatch)
	wantCode(t, "my @a = (1, 2);\nprint length(@a), \"\\n\";\n", infer.CodeTypeMismatch)
}

// A builtin that returns undef only when it fails is judged by the value
// it returns when it succeeds: perl 5.42 -w runs `fileno(STDOUT) . ">"`
// (prints "1>"), `vec($rin, fileno(STDIN), 1) = 1`, `fileno STDERR !=
// fileno STDOUT` and `1 while read($fh, $buf, 4096) > 0` silently. A value
// that is known to be undef still is reported: `undef() + 1` warns "Use of
// uninitialized value in addition (+)". So is a user sub's maybe-undef
// result: `sub f { return if $_[0]; "s" } f(1) . "x"` warns in
// concatenation.
func TestFilenoAndReadResultsInStrAndNumSlots(t *testing.T) {
	for _, src := range []string{
		"my $fd = fileno(STDOUT) . \">\";\nmy $rin = \"\"; vec($rin, fileno(STDIN), 1) = 1;\nprint \"$fd\\n\";\n",
		"print \"differ\\n\" if defined fileno STDERR && fileno STDERR != fileno STDOUT;\n",
		"open my $fh, \"<\", $0 or die; my $buf;\n1 while read($fh, $buf, 4096) > 0;\n",
	} {
		wantSilent(t, src)
	}

	wantCode(t, "my $x = undef() + 1;\n", infer.CodeCoercionMismatch)
	wantCode(t, "sub f { return if $_[0]; \"s\" }\nmy $x = f(1) . \"x\";\n", infer.CodeCoercionMismatch)
}
