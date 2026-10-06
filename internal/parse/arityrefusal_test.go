// ABOUTME: Tests that a call whose arity perl refuses at compile time is refused, and every call perl compiles is not.
// ABOUTME: Each case was measured with perl 5.42.0's `perl -c`; the arity comes from the declarations, never a builtin's name.
package parse

import (
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

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

// TestArityRefusalDieDefaultIsRequired: RFC 0001 "A required argument
// defaults to `die`". sort's plain candidate is `(List @list = die)`, so a
// sort that writes no argument is refused by the same rule, and with the
// same code, as `grep()`. Measured on 5.42.0, `sort()` and bare `sort` are
// "Not enough arguments for sort", while `sort(())` and `my @e; sort @e`
// compile. The refusal reads requiredness from the declaration: arity.go
// names no sort.
func TestArityRefusalDieDefaultIsRequired(t *testing.T) {
	wantArityRefusal(t, "sort();\n", "my @r = sort;\n", "grep();\n")
	wantNoRefusal(t, "sort(());\n", "my @e; sort @e;\n")
	src, err := os.ReadFile("arity.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(src), `"sort"`) {
		t.Errorf("arity.go names sort; requiredness is the declaration's")
	}
}

// TestArityRefusalLeavesUnprototypedCallsAlone: a library sub with no
// prototype is a list operator to perl whatever a declaration says of it,
// so a call its `.pmt` candidates do not accept keeps its parse, and `psc
// check` reports it ("Call sites"). Measured on 5.42.0 with a module M
// exporting `sub f { }`, `use M; f(1, 2, 3)` compiles.
func TestArityRefusalLeavesUnprototypedCallsAlone(t *testing.T) {
	saved := declarations
	defer func() { declarations = saved }()
	none := func(string) ([]byte, bool) { return nil, false }
	for _, tc := range []struct {
		name string
		decl fstest.MapFS
		load Loader
	}{
		{".pm", fstest.MapFS{}, func(name string) ([]byte, bool) {
			return []byte("package M;\nour @EXPORT = qw(f);\nsub f { }\n1;\n"), name == "M"
		}},
		{".pmt", fstest.MapFS{"declarations/M.pmt": {Data: []byte(
			"package M;\nour @EXPORT = qw(f);\nmulti sub f (Int $x) Int;\nmulti sub f (Int $x, Int $y) Int;\n")}}, none},
	} {
		declarations = tc.decl
		for _, src := range []string{"use M; f();\n", "use M; f(1, 2, 3);\n"} {
			if got := refusalsIn(ParseWithLoader([]byte(src), tc.load)); len(got) > 0 {
				t.Errorf("%s: %q refused %v; perl compiles it", tc.name, src, got)
			}
		}
	}
}

// TestPrototypeArityPerlRefusesIsRefused: RFC 0001 "Refusing what perl
// refuses" holds for prototypes read from source too. Measured on 5.42.0
// with `perl -c`, each refused call dies "Not enough arguments for
// main::f" or "Too many arguments", and each kept one compiles: an
// optional slot, `_`, a block with an empty list, a slurpy, a call through
// `&`, which ignores the prototype, and a parenthesised list, which perl
// counts as one argument but a builtin may not.
func TestPrototypeArityPerlRefusesIsRefused(t *testing.T) {
	wantArityRefusal(t,
		"sub f ($$) {} f(1);\n", "sub f ($$) {} f(1, 2, 3);\n",
		"sub f ($$); f(1);\n", "sub f ($$) {} main::f(1);\n",
		"sub f ($) {} f;\n", "sub f (;$) {} f(1, 2);\n",
		"sub f ($;$) {} f(1, 2, 3);\n", "sub f () {} f(1);\n",
		"sub f (&@) {} f();\n", "sub f (&) {} f(sub {1}, 2);\n",
		"sub f (\\[$@%]) {} f();\n", "my @a; sub f (\\@) {} f(@a, 1);\n",
		"use constant PI => 3; PI(1);\n", "builtin::true(1);\n",
	)
	wantNoRefusal(t,
		"sub f (;$) {} f();\n", "sub f (_) {} f();\n", "sub f (&@) {} f { 1 } 2;\n",
		"my @a; sub f (@) {} f(@a, 1, 2);\n", "sub f ($$) {} &f(1);\n",
		"sub f ($$) {} f((1, 2));\n", "sub f ($$) {} f(1, 2);\n",
	)
}

// TestPrototypeArityFromModuleSource: a prototype read from a module's
// source, or derived from a `.pmt` signature, refuses as one written in
// the file does. Measured on 5.42.0 with a module M exporting `sub two
// ($$) { }`, `use M; two(1)` dies "Not enough arguments for M::two".
func TestPrototypeArityFromModuleSource(t *testing.T) {
	saved := declarations
	defer func() { declarations = saved }()
	none := func(string) ([]byte, bool) { return nil, false }
	for _, tc := range []struct {
		name string
		decl fstest.MapFS
		load Loader
	}{
		{".pm", fstest.MapFS{}, func(name string) ([]byte, bool) {
			return []byte("package M;\nour @EXPORT = qw(two);\nsub two ($$) { }\n1;\n"), name == "M"
		}},
		{".pmt", fstest.MapFS{"declarations/M.pmt": {Data: []byte(
			"package M;\nour @EXPORT = qw(two);\nsub two (Int $x, Int $y) Int;\n")}}, none},
	} {
		declarations = tc.decl
		if got := refusalsIn(ParseWithLoader([]byte("use M; two(1);\n"), tc.load)); len(got) != 1 || got[0] != CallArity {
			t.Errorf("%s: two(1) refusals %v, want one %s", tc.name, got, CallArity)
		}
		if got := refusalsIn(ParseWithLoader([]byte("use M; two(1, 2);\n"), tc.load)); len(got) > 0 {
			t.Errorf("%s: two(1, 2) refused %v", tc.name, got)
		}
	}
}

// TestSortComparatorBeforeParens: in `sort NAME (LIST)` the name is the
// comparator, not a call of the list. Measured on 5.42.0 with
// -MO=Deparse,-p, `sort foo (3, 1)` and `sort foo(3, 1)` are both `sort foo
// 3, 1`; perl.git t/op/sort.t sorts `cmp_as_string (1,5,4,7,3,2,3)` under a
// `($$)` prototype, which as a call would be "Too many arguments".
func TestSortComparatorBeforeParens(t *testing.T) {
	for _, src := range []string{
		"sub cmp_as_string ($$) {} my @b = sort cmp_as_string (1, 5, 4);\n",
		"sub foo ($$) {} my @b = sort foo(3, 1);\n",
	} {
		var sort *Node
		var walk func(*Node)
		walk = func(n *Node) {
			if n.Kind == Call && n.Text == "sort" {
				sort = n
			}
			for _, c := range n.Children {
				walk(c)
			}
		}
		root := Parse([]byte(src))
		walk(root)
		if got := refusalsIn(root); len(got) > 0 {
			t.Errorf("%q: refused %v; perl compiles it", src, got)
		}
		if sort == nil || len(sort.Children) != 2 || !sort.Children[0].Comparator {
			t.Errorf("%q: sort's first argument is not its comparator", src)
		}
	}
}

// TestArityRefusalLeavesIndirectObjectsAlone: a call whose first argument
// is a bare class name may be perl's indirect method call, to which no
// prototype applies, so its arity is not refused. Measured on 5.42.0 with
// -MO=Deparse,-p: under `package P; sub new ($$;$) {}`, `new IO::File` is
// `'IO::File'->new`, and under Test::More's `is ($$;$)`, `is
// Time::Moment->from_string("x")->to_string, "a", "b"` passes is three
// arguments (T1's indirect_object_constructor.t and
// time_moment_java_xs.t).
func TestArityRefusalLeavesIndirectObjectsAlone(t *testing.T) {
	wantNoRefusal(t,
		"package P; sub new ($$;$) {} my $fh = new IO::File;\n",
		"sub is ($$;$) {} package Time::Moment; package main; is Time::Moment->from_string(\"x\")->to_string, \"a\", \"b\";\n",
	)
}

// TestLooseBuiltinsRefuseTheirOtherSide: a builtin whose own parse is
// looser than its candidates on one side is still refused on the other.
// Measured on 5.42.0 with `perl -c`: `close($x, $x)` and `scalar($x, $x)`
// compile, but `scalar()`, `closedir()`, `fileno`, `getpeername()`,
// `getsockname()`, `readdir()`, `rewinddir()` and `telldir()` die "Not
// enough arguments"; `do()` compiles, but `do($x, $x)` dies "Too many
// arguments for do "file"".
func TestLooseBuiltinsRefuseTheirOtherSide(t *testing.T) {
	var srcs []string
	for _, call := range []string{
		"scalar()", "closedir()", "fileno", "getpeername()", "getsockname()",
		"readdir()", "rewinddir()", "telldir()", "do($x, $x)",
	} {
		srcs = append(srcs, "my $x; "+call+";\n")
	}
	wantArityRefusal(t, srcs...)
}

// TestArityRefusalLeavesImportedOverridesAlone: a builtin's name in a `use`
// list may be an override from a module this parse does not read, and a
// call to it is the override's. Measured on 5.42.0 with `perl -c`, each
// kept call compiles, while the refused ones die "Too many arguments": an
// override never takes the `CORE::` spelling, and `use POSIX ()` imports
// nothing.
func TestArityRefusalLeavesImportedOverridesAlone(t *testing.T) {
	wantNoRefusal(t,
		"use Time::HiRes qw(alarm); alarm(1, 0.5);\n",
		"use Time::HiRes qw(sleep); sleep(1, 2);\n",
		"use POSIX qw(abs); abs(1, 2);\n",
		"use POSIX qw(localtime); localtime(1, 2);\n",
		"use subs qw(each); each();\n",
		"use subs \"localtime\"; localtime(1, 2);\n",
	)
	wantArityRefusal(t,
		"use Time::HiRes qw(sleep); CORE::sleep(1, 2);\n",
		"use POSIX (); localtime(1, 2);\n",
	)
}
