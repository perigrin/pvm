// ABOUTME: Tests that a call whose arity perl refuses at compile time is refused, and every call perl compiles is not.
// ABOUTME: Each case was measured with perl 5.42.0's `perl -c`; the arity comes from the declarations, never a builtin's name.
package parse

import "testing"

// wantArityRefusal wants each source refused, once, with the arity code.
func wantArityRefusal(t *testing.T, srcs ...string) {
	t.Helper()
	for _, src := range srcs {
		if got := refusedAs(src); len(got) != 1 || got[0] != "call_arity" {
			t.Errorf("%q: refusals %v, want one call_arity", src, got)
		}
	}
}

// TestBuiltinArityPerlRefusesIsRefused: RFC 0001 "Refusing what perl
// refuses", a call whose arity no candidate accepts. Measured on 5.42.0
// with `perl -c`: `select(1, 2)` dies "Not enough arguments for select
// system call" and `select(1, 2, 3, 4, 5)` "Too many arguments";
// `localtime(1, 2)` and `gmtime(1, 2)` "Too many arguments"; `each()`,
// `grep()`, `map()` and `push()` "Not enough arguments".
func TestBuiltinArityPerlRefusesIsRefused(t *testing.T) {
	wantArityRefusal(t,
		"select(1, 2);\n", "select(1, 2, 3);\n", "select(1, 2, 3, 4, 5);\n",
		"localtime(1, 2);\n", "gmtime(1, 2);\n",
		"each();\n", "grep();\n", "map();\n", "push();\n",
	)
}

// TestArityRefusalKeepsWhatPerlAccepts: a call perl compiles is never
// refused. Measured on 5.42.0 with `perl -c`, each of these compiles: the
// issue's `chomp()`, `localtime()`, `select()`, `push(@a)`, `grep { 1 }
// ()`, `sort(())` and `my @e; sort @e`; `not()` and `CORE::not()`, whose
// own parse takes no operand though CORE.pmt's `(Scalar $x)` requires one;
// and one call to each builtin whose own parse is looser than its
// declaration, found by calling every CORE.pmt builtin with none to six
// arguments, plain and as `CORE::NAME`, and keeping what perl compiles.
// `CORE::dump($x)` compiles, its operand a label. Without its feature, `fc`,
// `evalbytes`, `any`, `all`, `catch`, `method`, `isa`, `break`,
// `__CLASS__` and `__SUB__` name a user's sub.
func TestArityRefusalKeepsWhatPerlAccepts(t *testing.T) {
	var srcs []string
	for _, call := range []string{
		"chomp()", "localtime()", "select()", "push(@a)", "grep { 1 } ()",
		"sort(())", "my @e; sort @e", "not()", "CORE::not()",
		"system()", "system", "exec()", "exec", "do()",
		"eval($x, $x)", "scalar($x, $x)", "prototype($x, $x)", "write($x, $x)",
		"stat($x, $x)", "lstat($x, $x)", "readline($x, $x)", "tell($x, $x)",
		"telldir($x, $x)", "close($x, $x)", "closedir($x, $x)", "eof($x, $x)",
		"fileno($x, $x)", "getc($x, $x)", "getpeername($x, $x)",
		"getsockname($x, $x)", "readdir($x, $x)", "rewinddir($x, $x)",
		"fc($x, $x)", "evalbytes($x, $x)", "any()", "all()", "catch($x)",
		"method($x)", "isa($x)", "break($x)", "__CLASS__($x)", "__SUB__($x)",
		"CORE::dump($x)",
	} {
		srcs = append(srcs, "my ($x, @a); "+call+";\n")
	}
	wantNoRefusal(t, srcs...)
}

// TestArityRefusalLeavesAutoquotedWordsAlone: a builtin's name autoquoted
// by a `=>` or alone in a hash subscript is a string, not a call. Measured
// on 5.42.0, `my %h; $h{sort} = 1; $h{values} = 2; $h{-each} = 3` and
// `my %h = (values => 1, -each => 2)` compile; T1's
// foreach_implicit_live_array.t writes `bless { values => [10, 20] }`.
func TestArityRefusalLeavesAutoquotedWordsAlone(t *testing.T) {
	wantNoRefusal(t,
		"my %h; $h{sort} = 1; $h{values} = 2; $h{-each} = 3;\n",
		"my %h = (values => 1, -each => 2);\n",
		"my $r = bless { values => [10, 20] }, 'C';\n",
	)
}

// TestArityRefusalLeavesParenthesisedListsAlone: a parenthesised list
// written as one argument may count as several. Measured on 5.42.0, `pipe(my
// ($r, $w))`, `pipe((my $r, my $w))` and `link((1, 2))` compile, though pipe
// and link take two arguments, while `localtime((1, 2))` is "Too many
// arguments": how many it counts as is the builtin's, so no call with
// one is refused.
// perl.git's t/op/getppid.t writes `pipe my ($r, $w) or die`.
func TestArityRefusalLeavesParenthesisedListsAlone(t *testing.T) {
	wantNoRefusal(t,
		"pipe(my ($r, $w));\n", "pipe my ($r, $w) or die;\n",
		"pipe((my $r, my $w));\n", "link((1, 2));\n",
	)
}

// TestArityRefusalCodesDistinct: the arity refusal has a code of its own,
// listed in the inventory at a site no other code claims, so a corpus file
// can tell a constant in a `\$` slot, `sref(1)`, from a call whose arity
// perl refuses, `each()`.
func TestArityRefusalCodesDistinct(t *testing.T) {
	site, ok := RefusalSites[CallArity]
	if !ok {
		t.Fatalf("%s is not in RefusalSites", CallArity)
	}
	for code, other := range RefusalSites {
		if code != CallArity && other.Where == site.Where {
			t.Errorf("%s shares its site with %s", CallArity, code)
		}
	}
	for src, want := range map[string]RefusalCode{
		refScalarPrelude + "sref(1);\n": RefScalarSlot,
		"each();\n":                     CallArity,
	} {
		if got := refusedAs(src); len(got) != 1 || got[0] != want {
			t.Errorf("%q: refusals %v, want one %s", src, got, want)
		}
	}
}
