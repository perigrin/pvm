// ABOUTME: Tests for the call rules of Analyze: which calls are checked as a CORE builtin, and how.
// ABOUTME: Each case is a program perl 5.42 -w runs silently, paired with a control perl rejects or warns about.

package infer_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
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

// An array in a builtin's non-slurpy `$` slot is in scalar context, so it
// arrives as its count: perl 5.42 -w is silent on `int(rand(@m))`, a
// random index of @m, and on `lc(@a)`, which is "2" for two elements.
// The builtins perl refuses or warns on an aggregate in are still
// reported: `length(@a)` warns "length() used on @a (did you mean
// "scalar(@a)"?)", `defined(@a)` dies "Can't use 'defined(@array)'", and
// `exists(@a)` dies "exists argument is not a HASH or ARRAY element or a
// subroutine".
func TestArrayInScalarSlotIsCount(t *testing.T) {
	for _, src := range []string{
		"my @m = (1, 2, 3);\nmy $i = int(rand(@m));\n",
		"my @a = (1, 2);\nmy $c = lc(@a);\n",
	} {
		wantSilent(t, src)
	}
	for _, src := range []string{
		"my @a = (1, 2);\nmy $l = length(@a);\n",
		"my @a = (1, 2);\nmy $d = defined(@a);\n",
		"my @a = (1, 2);\nmy $e = exists(@a);\n",
	} {
		wantCode(t, src, infer.CodeTypeMismatch)
	}
}

// srand and sleep take their `$` argument in scalar context, so an array
// or a hash there is its count: perl 5.42 -w, `my @e; srand(@e)` is "0
// but true", `srand(@a)` with two elements 2, and `my %h; srand(%h)` "0
// but true", each silently, and `sleep(@e)` sleeps 0 seconds. A hash in
// length's Str slot is still reported: perl warns "length() used on %h
// (did you mean "scalar(keys %h)"?)".
func TestArrayInScalarSlotIsItsCount(t *testing.T) {
	for _, src := range []string{
		"my @e;\nsrand(@e);\n",
		"my @a = (1, 2);\nsrand(@a);\n",
		"my %h;\nsrand(%h);\n",
		"my @e;\nsleep(@e);\n",
	} {
		wantSilent(t, src)
	}
	wantCode(t, "my %h = (a => 1);\nmy $l = length(%h);\n", infer.CodeTypeMismatch)
}

// succeeded judges a builtin's result without the Undef it returns only
// when it fails: perl 5.42 -w is silent on `fileno(STDOUT) . ">"`. A
// builtin that returns undef in normal use keeps it: with STDIN at end of
// file and %h empty, perl warns "Use of uninitialized value in
// concatenation (.) or string" on `getc(STDIN) . "x"` and `each(%h) .
// "y"`, and so on a sub returning getc's result. A builtin `use subs`
// overrides is a user sub, and its result is its own: perl warns the
// same on `use subs "fileno"; sub fileno { $_[0] ? "a" : undef }
// fileno(0) . ">"`.
func TestSucceededKeepsNormalUndef(t *testing.T) {
	wantSilent(t, "print fileno(STDOUT) . \">\";\n")
	for _, src := range []string{
		"print getc(STDIN) . \"x\";\n",
		"my %h;\nprint each(%h) . \"y\";\n",
		"sub g { return getc(STDIN) }\nprint g() . \"x\";\n",
		"use subs \"fileno\";\nsub fileno { $_[0] ? \"a\" : undef }\nprint fileno(0) . \">\";\n",
	} {
		wantCode(t, src, infer.CodeCoercionMismatch)
	}
}

// A builtin a module's default @EXPORT overrides is the module's sub, not
// the builtin, as the parser resolves it: perl 5.42 -w, with Foo2.pm
// exporting `sleep` by default, `use Foo2; sleep("abc")` prints "slept
// abc" silently. Without the import the builtin is called: `use Foo2 ();
// sleep("abc")` warns "Argument "abc" isn't numeric in sleep", and so does
// `sleep("abc")` when Foo2 cannot be found. A lexical sub overrides the
// builtin, `my sub sleep { "s" } sleep("abc")` being "s", and a package
// sub the file declares does not: `sub sleep { "s" } sleep("abc")` warns
// "Ambiguous call resolved as CORE::sleep()" and "isn't numeric".
func TestDefaultExportOverridesBuiltin(t *testing.T) {
	foo2 := "package Foo2;\nuse Exporter 'import';\nour @EXPORT = ('sleep');\nsub sleep { return \"slept $_[0]\" }\n1;\n"
	load := func(name string) ([]byte, bool) {
		if name == "Foo2" {
			return []byte(foo2), true
		}
		return nil, false
	}
	diags := func(src string, load parse.Loader) []infer.Diagnostic {
		root := parse.ParseWithLoader([]byte(src), load)
		f := lower.Lower(root)
		_, d, _ := infer.Analyze(f, lower.NewPositions(root, f), nil, infer.Options{})
		return d
	}

	src := "use Foo2;\nmy $x = sleep(\"abc\");\n"
	for _, s := range []string{src, "my sub sleep { \"s\" }\nmy $x = sleep(\"abc\");\n"} {
		for _, d := range diags(s, load) {
			t.Errorf("%q: diagnostic %q, want none", s, d.Message)
		}
	}
	for _, c := range []struct {
		src  string
		load parse.Loader
	}{
		{"use Foo2 ();\nmy $x = sleep(\"abc\");\n", load},
		{src, nil},
		{"sub sleep { \"s\" }\nmy $x = sleep(\"abc\");\n", nil},
	} {
		if d := diags(c.src, c.load); len(d) != 1 || d[0].Code != infer.CodeTypeMismatch {
			t.Errorf("%q: diagnostics %v, want one type mismatch", c.src, d)
		}
	}
}

// TestArityCheckReachable names the input that reaches the arity check: a
// sub of a builtin's name that perl does not let override it. perl 5.42
// resolves `getgrgid()` to CORE::getgrgid and refuses it, "Not enough
// arguments for getgrgid"; the parser reads the call as the user sub's and
// does not refuse it, so the arity check is what reports it. A refused
// call such as `push()` never reaches it: psc reports the refusal.
func TestArityCheckReachable(t *testing.T) {
	for _, src := range []string{
		"sub getgrgid { \"u\" }\nprint getgrgid(), \"\\n\";\n",
		"sub rename { \"fake\" }\nprint rename(), \"\\n\";\n",
	} {
		var refused func(n *parse.Node) bool
		refused = func(n *parse.Node) bool {
			return n.Kind == parse.Unknown || slices.ContainsFunc(n.Children, refused)
		}
		if refused(parse.Parse([]byte(src))) {
			t.Errorf("%q: the parser refuses it", src)
		}
		wantCode(t, src, infer.CodeArityMismatch)
	}
}

// TestSliceArgumentFlattens: I7, declaredCall must decide whether an
// argument flattens from its own syntactic form, not its inferred
// type -- a slice is untyped (vars.go's slice, "the slice itself is not
// typed") and a call's type is whatever it returns in scalar context,
// neither ever carrying the Array or Hash bit the old, type-based test
// required. POSIX's strftime and mktime take fixed, typed Int
// parameters with no prototype of their own (measured, `prototype(\&
// POSIX::strftime)` is undef), so perl parses every argument position
// as an ordinary list operator's: ambient list context, in which a
// slice, a map, or a call flattens into however many values it holds.
// perl 5.42 -w runs every one of these silently.
func TestSliceArgumentFlattens(t *testing.T) {
	for _, src := range []string{
		`use POSIX qw(strftime); my $s = strftime("%Y", (localtime)[0..5]);`,
		`use POSIX qw(strftime); my @lt = localtime; my $s = strftime("%Y", @lt[0..5]);`,
		`use POSIX qw(strftime); my @lt = localtime; my $r = [@lt]; my $s = strftime("%Y", @$r[0..5]);`,
		`use POSIX qw(strftime); my @lt = localtime; my $s = strftime("%Y", map { $_ } @lt);`,
		`use POSIX qw(mktime); my @lt = localtime; my $t = mktime(@lt[0..5]);`,
		`use POSIX qw(fmod); sub two { return (7, 3) } my $f = fmod(two());`,
		`use POSIX qw(fmod); my $f = fmod(eval { (7, 3) });`,
		`use POSIX qw(fmod); my $f = fmod(do { (7, 3) });`,
	} {
		wantSilentResolved(t, src)
	}
}

// TestSumResultInArithmeticIsSilent: I8, List::Util::sum's declared
// return, Num|Inf|NaN|Object joined with the empty candidate's Undef,
// must not report in ordinary arithmetic: an addend overloading `+` or
// `*` comes back unchanged (List/Util.tpm's own comment), so the
// Object bit is as legitimate a number there as a plain Num, and the
// Undef is the same possible-undef a declared call's own catch-all
// candidate always carries. perl 5.42 -w runs every one of these
// silently.
func TestSumResultInArithmeticIsSilent(t *testing.T) {
	for _, src := range []string{
		`use List::Util qw(sum); my @n = (1, 2, 3); my $s = sum(@n) + 1;`,
		`use List::Util qw(sum); my @n = (1, 2, 3); print abs(sum(@n)), "\n";`,
		`use List::Util qw(sum); my @n = (1, 2, 3); print sqrt(sum(@n)), "\n";`,
		`use List::Util qw(sum); use POSIX qw(floor); my @n = (1, 2, 3); my $f = floor(sum(@n));`,
		`use List::Util qw(sum); my @n = (1, 2, 3); my $avg = sum(@n) / @n;`,
	} {
		wantSilentResolved(t, src)
	}
}

// wantArity fails t unless src, analyzed with no options, gives exactly
// one diagnostic, a call arity mismatch.
func wantArity(t *testing.T, src string) {
	t.Helper()
	if diags := diagsWithOptions(src, infer.Options{}); len(diags) != 1 || diags[0].Code != infer.CodeArityMismatch {
		t.Errorf("%q: diagnostics %v, want one %s", src, diags, infer.CodeArityMismatch)
	}
}

// TestScalarCallArgumentKeepsArity: a call flattens into a declared
// call's arguments only when what it returns in list context is a list.
// length, scalar, lc, join and abs return one value however they are
// called, so fmod is still called with one argument: perl 5.42 -w dies
// on each, "Usage: POSIX::fmod(x, y)".
//
// List::Util's first is declared to return Any, which carries the Array
// and Hash bits but is one value: perl 5.42 -w dies the same way on
// `fmod(first { 1 } @n)`.
func TestScalarCallArgumentKeepsArity(t *testing.T) {
	for _, arg := range []string{`length("abc")`, `scalar(@n)`, `lc("ab")`, `join(",", @n)`, `abs(2)`, `first { 1 } @n`} {
		wantArity(t, "use POSIX qw(fmod); use List::Util qw(first); my @n = (1, 2, 3); my $f = fmod("+arg+");")
	}
}

// TestListFormsFlatten: every argument form perl evaluates to a list in
// an ordinary argument list flattens, so its count is not one argument:
// a call whose list-context return is a list (localtime, reverse, a
// method call, a sub of this file, a coderef, `&$c`), a ternary either of whose
// arms flattens, a list repetition, a range, qw, a parenthesised list, a
// list assignment, a readline and a glob, and a logical or's right
// operand. Measured on 5.42.0, `sub cnt { scalar @_ }` counts 2 for
// `cnt($o->two)`, `cnt(1 ? @a : @a)`, `cnt(0 ? 1 : @a)`, `cnt((7,3))`,
// `cnt(my @q = (7,3))`, `cnt(my ($x, $y) = (7,3))` and `cnt(0 || @a)`,
// 6 for `cnt((0) x 6)`, 3 for `cnt(7..9)` and `cnt(qw(a b c))`; and perl
// 5.42 -w runs each program below silently. A readline yields a value
// per line, so perl's own result there depends on its input: given
// "7\n3\n" it runs `fmod(<STDIN>)` silently. A brace glob with no
// wildcard yields each of its words, `<{7,3}>` the two.
func TestListFormsFlatten(t *testing.T) {
	const pre = "use POSIX qw(fmod strftime); package O { sub new { bless {}, shift } sub two { (7, 3) } } package main; my $o = O->new; my @a = (7, 3); my @lt = localtime; sub two { (7, 3) } my $c = \\&two; "
	for _, call := range []string{
		`strftime("%Y", localtime)`,
		`fmod(reverse @a)`,
		`fmod($o->two)`,
		`fmod(O->two)`,
		`fmod(two())`,
		`fmod($c->())`,
		`fmod(&$c)`,
		`strftime("%Y", @ARGV ? @lt : @lt)`,
		`fmod(@ARGV ? 1 : @a)`,
		`fmod(@ARGV ? (7, 3) : (8, 3))`,
		`strftime("%Y", (0) x 6)`,
		`fmod(7..8)`,
		`strftime("%Y", qw(0 0 0 1 0 100))`,
		`fmod(qw(7 3))`,
		`fmod((7, 3))`,
		`fmod(my @q = (7, 3))`,
		`fmod(my ($x, $y) = (7, 3))`,
		`fmod(0 || @a)`,
		`fmod(<STDIN>)`,
		`fmod(<{7,3}>)`,
		`fmod(glob("{7,3}"))`,
		`fmod(eval "(7, 3)")`,
		`fmod(do "./two.pl")`,
		`fmod(@a = (7, 3))`,
		`fmod(@ARGV ? @a : 1)`,
	} {
		wantSilentResolved(t, pre+"my $r = "+call+";")
	}
}

// TestScalarCallArgumentStillChecked: an argument that is a call
// returning one value leaves the call's arity known, so every argument
// is still checked against the candidate it selects. perl 5.42 -w warns
// "Use of uninitialized value in subroutine entry" on `fmod(undef,
// abs(2))` and `fmod(max(@n), undef)`, and `Argument "abc" isn't
// numeric` on `fmod(sum(@n), "abc")`.
func TestScalarCallArgumentStillChecked(t *testing.T) {
	for _, call := range []string{`fmod(undef, abs(2))`, `fmod(max(@n), undef)`, `fmod(sum(@n), "abc")`} {
		src := "use POSIX qw(fmod); use List::Util qw(sum max); my @n = (1, 2, 3); my $f = " + call + ";"
		diags := diagsWithOptions(src, infer.Options{})
		if len(diags) != 1 || diags[0].Code != infer.CodeCoercionMismatch {
			t.Errorf("%q: diagnostics %v, want one %s", src, diags, infer.CodeCoercionMismatch)
		}
	}
}

// TestArgumentsBeforeAMethodCallAreChecked: a method call may return a
// list, so it flattens, but every argument before it still stands at its
// own position, and is checked against the parameter there. perl 5.42 -w
// warns "Use of uninitialized value in subroutine entry" on `fmod(undef,
// $o->one)` and `Argument "abc" isn't numeric` on `fmod("abc", $o->one)`,
// and is silent on `fmod(7, $o->one)`.
func TestArgumentsBeforeAMethodCallAreChecked(t *testing.T) {
	const pre = "use POSIX qw(fmod); package O { sub new { bless {}, shift } sub one { 3 } } package main; my $o = O->new; "
	for _, call := range []string{`fmod(undef, $o->one)`, `fmod("abc", $o->one)`, `fmod("abc", O->one)`} {
		src := pre + "my $f = " + call + ";"
		diags := diagsWithOptions(src, infer.Options{})
		if len(diags) != 1 || diags[0].Code != infer.CodeCoercionMismatch {
			t.Errorf("%q: diagnostics %v, want one %s", src, diags, infer.CodeCoercionMismatch)
		}
	}
	for _, call := range []string{`fmod(7, $o->one)`, `fmod($o->one, "abc")`} {
		src := pre + "my $f = " + call + ";"
		assert.Empty(t, diagsWithOptions(src, infer.Options{}), src)
	}
}

// TestCwdChdirRefArgumentNoError: Cwd::chdir takes a string or a handle,
// and any other reference is stringified on its way to CORE::chdir:
// perl 5.42 -w is silent on `Cwd::chdir({})`, which returns 0. Its
// declaration says so with a handle candidate beside a string one, so a
// reference selects the string candidate and is a lossy coercion,
// reported only under WarnLossyCoercion, not a default error; a handle
// is no coercion at all. POSIX::chdir, Time::HiRes's stat and lstat and
// File::stat's lstat stringify a stray reference the same way, measured
// silent under perl 5.42 -w.
func TestCwdChdirRefArgumentNoError(t *testing.T) {
	for _, ref := range []string{
		"use Cwd (); my $h = {}; my $q = Cwd::chdir($h);",
		"use POSIX (); my $q = POSIX::chdir({});",
		"use Time::HiRes (); my @s = Time::HiRes::stat({});",
		"use Time::HiRes (); my $s = Time::HiRes::lstat([]);",
		"use File::stat (); my $s = File::stat::lstat({});",
	} {
		assert.Empty(t, diagsWithOptions(ref, infer.Options{}), ref)
		diags := diagsWithOptions(ref, infer.Options{WarnLossyCoercion: true})
		if assert.Len(t, diags, 1, ref) {
			assert.Equal(t, infer.CodeLossyCoercion, diags[0].Code, ref)
		}
	}
	const handle = "use Cwd (); opendir(my $dh, '.') or die; my $q = Cwd::chdir($dh); my $r = Cwd::chdir(\\*STDIN);"
	assert.Empty(t, diagsWithOptions(handle, infer.Options{WarnLossyCoercion: true}))
}

// TestGlobDerefParamStillReportsRef: whether a parameter stringifies a
// stray reference belongs to each declaration, not to its type. File::stat's
// stat and Symbol's qualify_to_ref dereference a reference as a glob, so
// perl 5.42 -w dies "Not a GLOB reference" on each call below, and each
// stays a coercion mismatch.
func TestGlobDerefParamStillReportsRef(t *testing.T) {
	for _, src := range []string{
		"use File::stat (); my $s = File::stat::stat({});",
		"use File::stat (); my $s = File::stat::stat([]);",
		"use Symbol (); my $s = Symbol::qualify_to_ref({});",
		"use Symbol (); my $s = Symbol::qualify_to_ref([]);",
	} {
		diags := diagsWithOptions(src, infer.Options{})
		if len(diags) != 1 || diags[0].Code != infer.CodeCoercionMismatch || diags[0].Severity != infer.Error {
			t.Errorf("%q: diagnostics %v, want one %s error", src, diags, infer.CodeCoercionMismatch)
		}
	}
}

// TestUndefUnionCoercesToItsMember: a parameter typed X|Undef takes what
// perl coerces to X, as X itself does. Socket's sockaddr_in takes `Int|Undef
// $port`, and perl 5.42 -w is silent on a string port, literal or held in a
// variable; Time::HiRes::utime's `Num|Undef` times take "1.5" the same way.
// A literal that is no number is still reported: perl warns `Argument "abc"
// isn't numeric` on `sockaddr_in("abc", ...)`.
func TestUndefUnionCoercesToItsMember(t *testing.T) {
	for _, src := range []string{
		`use Socket qw(sockaddr_in inet_aton); my $x = sockaddr_in("80", inet_aton("127.0.0.1"));`,
		`use Socket qw(sockaddr_in inet_aton); my $a = inet_aton("127.0.0.1"); my $p = "80"; my $x = sockaddr_in($p, $a);`,
		`use Time::HiRes (); my $n = Time::HiRes::utime("1.5", "1.5", "f.txt");`,
	} {
		assert.Empty(t, diagsWithOptions(src, infer.Options{}), src)
	}
	const abc = `use Socket qw(sockaddr_in inet_aton); my $x = sockaddr_in("abc", inet_aton("127.0.0.1"));`
	if diags := diagsWithOptions(abc, infer.Options{}); assert.Len(t, diags, 1, abc) {
		assert.Equal(t, infer.CodeCoercionMismatch, diags[0].Code)
	}
}

// TestFileparseTakesItsPath: File::Basename::fileparse's path is a
// required Str of its own, ahead of the suffixes. perl 5.42 -w is silent
// on `fileparse({})`, which stringifies the reference, so it is a lossy
// coercion, reported only under WarnLossyCoercion; and perl dies
// "fileparse(): need a valid pathname" on `fileparse()`.
func TestFileparseTakesItsPath(t *testing.T) {
	const ref = "use File::Basename qw(fileparse); my @p = fileparse({});"
	assert.Empty(t, diagsWithOptions(ref, infer.Options{}), ref)
	if diags := diagsWithOptions(ref, infer.Options{WarnLossyCoercion: true}); assert.Len(t, diags, 1, ref) {
		assert.Equal(t, infer.CodeLossyCoercion, diags[0].Code, ref)
	}
	wantArity(t, "use File::Basename qw(fileparse); my @p = fileparse();")
}
