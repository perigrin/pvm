// ABOUTME: Tests for the call rules of AnalyzeLowered: which calls are checked as a CORE builtin, and how.
// ABOUTME: Each case is a program perl 5.42 -w runs silently, paired with a control perl rejects or warns about.

package infer_test

import (
	"testing"

	"tamarou.com/pvm/internal/infer"
)

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

// defined's operand is in scalar context: perl 5.42 -w runs
// `defined($self{lastline} = readline($fh))`, `defined(my $x =
// readline($fh))` and `1 while defined caller ++$lvl` silently, each
// a line or a package name, or undef. An aggregate there is perl's
// compile error, "Can't use 'defined(@array)'", and is still reported.
func TestDefinedOfAssignmentIsScalar(t *testing.T) {
	for _, src := range []string{
		"open my $fh, \"<\", $0 or die;\nmy %self;\nprint \"l\\n\" if defined($self{lastline} = readline($fh));\n",
		"my $lvl = 0;\nsub depth { 1 while defined caller ++$lvl; $lvl }\nopen my $fh, \"<\", $0; print \"l\\n\" if defined(my $x = readline($fh)); print depth(), \"\\n\";\n",
		"my $p = defined caller;\n",
		"defined(;",
	} {
		wantSilent(t, src)
	}

	wantCode(t, "my @a = (1, 2);\nprint defined(@a);\n", infer.CodeTypeMismatch)
	wantCode(t, "my %h;\nprint defined(%h);\n", infer.CodeTypeMismatch)
}

// 4-arg select's timeout may be undef, which blocks until a handle is
// ready: perl 5.42 -w compiles and runs `select($rout = $rin, undef, undef,
// undef) if 0` silently. A string timeout is still reported: `select(undef,
// undef, undef, "abc")` warns 'Argument "abc" isn't numeric in select
// system call'.
func TestSelectAcceptsUndefTimeout(t *testing.T) {
	for _, src := range []string{
		"my ($rin, $rout) = (\"\", \"\");\nselect($rout = $rin, undef, undef, undef) if 0;\n",
		"select(undef, undef, undef, 0.25);\n",
		"my $t;\nselect(undef, undef, undef, $t);\n",
	} {
		wantSilent(t, src)
	}

	wantCode(t, "select(undef, undef, undef, \"abc\");\n", infer.CodeTypeMismatch)
}

// A builtin a `use` list names is overridden by the sub it imports or
// predeclares: perl 5.42 -w runs `use subs "rename"; sub rename { "fake" }
// print rename()` and prints "fake", and compiles `use POSIX
// qw(localtime); localtime(1, 2)`. A sub of the same name without the
// `use` does not override: `sub rename { "fake" } rename()` is "Not enough
// arguments for rename". Nor can a strong keyword be overridden: under
// `use subs "defined"`, `defined(@a)` is still "Can't use
// 'defined(@array)'".
func TestUseSubsOverridesBuiltinCheck(t *testing.T) {
	for _, src := range []string{
		"use subs \"rename\";\nsub rename { \"fake\" }\nprint rename(), \"\\n\";\n",
		"use subs qw(rename);\nmy $r = rename();\n",
		"use POSIX qw(localtime);\nmy @t = localtime(1, 2);\n",
	} {
		wantSilent(t, src)
	}

	wantCode(t, "sub rename { \"fake\" }\nprint rename(), \"\\n\";\n", infer.CodeArityMismatch)
	wantCode(t, "use subs \"defined\";\nsub defined { 1 }\nmy @a;\nprint defined(@a);\n", infer.CodeTypeMismatch)
}

// getsockopt and prototype return undef only when they fail, so a Str
// slot judges what they return on success: perl 5.42 -w runs `unpack 'i',
// getsockopt($s, SOL_SOCKET, SO_SNDBUF)` (prints "ok" for a positive
// buffer) and `defined prototype($b) && prototype($b) =~ /^;?\*/`
// silently. A known undef in a Str slot, `(undef) . "x"`, warns "Use of
// uninitialized value in concatenation".
func TestUndefUnionBuiltinResultInStrSlot(t *testing.T) {
	for _, src := range []string{
		"use Socket;\nsocket(my $s, PF_INET, SOCK_STREAM, 0) or die;\nmy $sndbuf = unpack 'i', getsockopt($s, SOL_SOCKET, SO_SNDBUF);\nprint $sndbuf > 0 ? \"ok\\n\" : \"no\\n\";\n",
		"my $builtinname = \"CORE::push\";\nprint \"p\\n\" if defined prototype($builtinname) && prototype($builtinname) =~ /^;?\\*/;\n",
	} {
		wantSilent(t, src)
	}

	wantCode(t, "my $y = (undef) . \"x\";\n", infer.CodeCoercionMismatch)
}

// `scalar localtime()` is the date string: perl 5.42 -w runs `"on: " .
// scalar localtime() . "\n"` and `"g: " . scalar gmtime(0)` silently,
// printing "on: Wed Oct 14 ..." and "g: Thu Jan  1 00:00:00 1970".
// Without `scalar`, in list context, localtime is nine numbers.
func TestScalarLocaltimeIsStr(t *testing.T) {
	for _, src := range []string{
		"my $stamp = \"on: \" . scalar localtime() . \"\\n\";\nprint $stamp;\n",
		"my $g = \"g: \" . scalar gmtime(0);\n",
	} {
		wantSilent(t, src)
	}

	wantCode(t, "my $y = (undef) . \"x\";\n", infer.CodeCoercionMismatch)
}
