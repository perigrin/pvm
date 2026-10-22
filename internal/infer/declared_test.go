// ABOUTME: Tests that inference types a call to a declared sub from the signatures the parser resolved, RFC 0001 "Call sites".
// ABOUTME: The fixture modules are internal/parse/testdata/declarations/My/*.tpm, read through parse.OverlayDeclarations.

package infer_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// analyzeDeclared analyses src with the parse's fixture declarations
// beside the shipped ones, resolving `use` through load, which may be nil.
func analyzeDeclared(t *testing.T, src string, load parse.Loader) (typed, []infer.Diagnostic, *infer.SymbolTable) {
	t.Helper()
	t.Cleanup(parse.OverlayDeclarations(os.DirFS("../parse/testdata")))
	if load == nil {
		load = func(string) ([]byte, bool) { return nil, false }
	}
	r := infer.AnalyzeParsed(parse.ParseWithLoader([]byte(src), load), nil, infer.Options{})
	return typed{f: r.File, pos: r.Positions, ann: r.Annotations}, r.Diagnostics, r.Symbols
}

// declaredCallType is the type inference gives the call written as call in
// src, Unknown when it gives none, and src's diagnostics.
func declaredCallType(t *testing.T, src, call string) (types.Type, []infer.Diagnostic) {
	t.Helper()
	ty, diags, _ := analyzeDeclared(t, src, nil)
	got, _ := findNodeType(ty, []byte(src), call)
	return got, diags
}

// TestInferTypesCallFromDeclaration: My::Lib declares `sub join_all
// :export (Str $sep, List[Str] @parts) Str`, so a call to the imported
// join_all is a Str.
func TestInferTypesCallFromDeclaration(t *testing.T) {
	src := `use My::Lib; my $s = join_all(",", "a", "b");`
	got, diags := declaredCallType(t, src, `join_all(",", "a", "b")`)
	assert.Equal(t, types.Str, got)
	assert.Empty(t, diags)
}

// TestInferSelectsMultiAtCallSite: a multi's candidate is selected at the
// call by its context and its arguments' types. Measured on 5.42.0,
// `my $t = localtime(0)` is "Thu Jan  1 00:00:00 1970" and `my @l =
// localtime(0)` nine numbers; `my %h = (a => 1); my $k = each %h` is "a"
// and `my @a = (5); my $i = each @a` is 0. A call whose context the site
// cannot determine, a sub's return value, joins its candidates. My::Lib's
// `when_` is a library multi forking on context the same way, and joins
// them in a sub's return, with nothing reported.
func TestInferSelectsMultiAtCallSite(t *testing.T) {
	cases := []struct {
		src, call string
		want      types.Type
	}{
		{"my $t = localtime(0);", "localtime(0)", types.Str | types.Undef},
		{"my @l = localtime(0);", "localtime(0)", types.List},
		{"my %h = (a => 1); my $k = each %h;", "each %h", types.Str | types.Undef},
		{"my @a = (5); my $i = each @a;", "each @a", types.Int | types.Undef},
		{"sub f { return localtime($_[0]) }", "localtime($_[0])", types.Str | types.Undef | types.List},
		{"use My::Lib; my $w = when_();", "when_()", types.Str},
		{"use My::Lib; my @w = when_();", "when_()", types.List},
		{"use My::Lib; sub g { return when_() }", "when_()", types.Str | types.List},
	}
	for _, c := range cases {
		got, diags := declaredCallType(t, c.src, c.call)
		assert.Equal(t, c.want, got, "%q", c.src)
		assert.Empty(t, diags, "%q", c.src)
	}
}

// TestInferUndeclaredImportUntyped: a sub imported from a module with no
// .tpm has no declaration to type its call from.
func TestInferUndeclaredImportUntyped(t *testing.T) {
	load := func(name string) ([]byte, bool) {
		return []byte("package Plain;\nuse Exporter 'import';\nour @EXPORT = qw(plain);\nsub plain { 1 }\n1;\n"), name == "Plain"
	}
	src := "use Plain; my $x = plain(1);"
	ty, diags, _ := analyzeDeclared(t, src, load)
	got, _ := findNodeType(ty, []byte(src), "plain(1)")
	assert.Equal(t, types.Unknown, got)
	assert.Empty(t, diags)
}

// TestInferScalarSortHasNoType: a sort in scalar context is not the List
// its list-context candidates return, and the call site gives it no type.
// Measured on 5.42.0, `my @x = (3, 1); my $n = sort @x` leaves $n undef,
// warning "Useless use of sort in scalar context".
func TestInferScalarSortHasNoType(t *testing.T) {
	got, _ := declaredCallType(t, "my @x = (3, 1); my $n = sort @x;", "sort @x")
	assert.NotEqual(t, types.List, got)
	assert.Equal(t, types.Unknown, got)
}

// TestInferNoCandidateArityIsError: My::Multi's two takes one argument or
// two, and has no prototype, so perl compiles `two(1, 2, 3)` and the
// declaration rules it out: an error at the call, which has no type. A
// call passing an array has no count the declaration can rule out.
func TestInferNoCandidateArityIsError(t *testing.T) {
	got, diags := declaredCallType(t, "use My::Multi; my $r = two(1, 2, 3);", "two(1, 2, 3)")
	assert.Equal(t, types.Unknown, got)
	require.Len(t, diags, 1)
	assert.Equal(t, infer.Error, diags[0].Severity)
	assert.Equal(t, infer.CodeArityMismatch, diags[0].Code, diags[0].Message)

	// An array flattens into the call's arguments, however many it holds.
	_, diags = declaredCallType(t, "use My::Multi; my @a = (1, 2, 3); my $r = two(@a, 4, 5);", "two(@a, 4, 5)")
	assert.Empty(t, diags)

	// Three arguments and an array are at least three, which no candidate
	// takes however many the array holds.
	got, diags = declaredCallType(t, "use My::Multi; my @a = (1); my $r = two(1, 2, 3, @a);", "two(1, 2, 3, @a)")
	assert.Equal(t, types.Unknown, got)
	require.Len(t, diags, 1)
	assert.Equal(t, infer.CodeArityMismatch, diags[0].Code, diags[0].Message)
}

// TestTooWideParametersNarrowed: two of audit group D's too-wide
// parameters (01a15ac0-6ccc) now reject a call shape perl itself dies
// on, where the old declaration -- Any for weaken, a bare slurpy List
// for crc32 -- admitted it without a diagnostic.
//
// Scalar::Util::weaken's $ref is now Ref|Undef: measured on 5.42.0,
// `weaken($n)` with $n holding a plain Int dies "Can't weaken a
// nonreference", which Any used to admit silently.
//
// Compress::Raw::Zlib::crc32's $buffer is now required (Str|ScalarRef|
// Undef), not a List accepting any arity: measured, `crc32()` dies
// "Usage: Compress::Raw::Zlib::crc32(buf, ...)", which the old bare
// List candidate's zero-or-more arity used to admit.
func TestTooWideParametersNarrowed(t *testing.T) {
	_, diags := declaredCallType(t, "use Scalar::Util qw(weaken); my $n = 5; weaken($n);", "weaken($n)")
	require.Len(t, diags, 1)
	assert.Equal(t, infer.Error, diags[0].Severity)
	assert.Equal(t, infer.CodeCoercionMismatch, diags[0].Code, diags[0].Message)

	_, diags = declaredCallType(t, "use Compress::Raw::Zlib qw(crc32); my $c = crc32();", "crc32()")
	require.Len(t, diags, 1)
	assert.Equal(t, infer.Error, diags[0].Severity)
	assert.Equal(t, infer.CodeArityMismatch, diags[0].Code, diags[0].Message)

	// weaken on an actual reference, or undef, still type-checks clean.
	_, diags = declaredCallType(t, "use Scalar::Util qw(weaken); my $r = {}; weaken($r);", "weaken($r)")
	assert.Empty(t, diags)
	_, diags = declaredCallType(t, "use Scalar::Util qw(weaken); my $u; weaken($u);", "weaken($u)")
	assert.Empty(t, diags)
}

// TestAnyScalarArgumentKeepsArity: I8, an Any-typed scalar is not an
// aggregate, so it must not relax a call's arity the way an array, a
// hash, a list-returning call or a List-typed value does. List::Util's
// first is declared Any, its element type unknown; measured on 5.42.0,
// `use Cwd qw(abs_path); my $p = abs_path("x","y")` dies "Usage:
// Cwd::abs_path(pathsv=Nullsv)" -- abs_path takes at most one argument,
// whatever that argument's type -- so passing first's result as the
// extra argument must still be reported, not silently accepted as
// though first had flattened into the call's arguments.
func TestAnyScalarArgumentKeepsArity(t *testing.T) {
	src := `use List::Util qw(first); use Cwd qw(abs_path); my $v = first { 1 } @ARGV; my $r = abs_path($v, "y");`
	got, diags := declaredCallType(t, src, `abs_path($v, "y")`)
	assert.Equal(t, types.Unknown, got)
	require.Len(t, diags, 1)
	assert.Equal(t, infer.CodeArityMismatch, diags[0].Code, diags[0].Message)
}

// TestInferArityChoosesCandidate: CORE.tpm's select is a Str with one
// argument and a List with four. Measured on 5.42.0, `select(STDERR)` is
// "main::STDOUT" and `select(undef, undef, undef, 0)` is (0, 0). My::Multi's
// two is chosen by its arity the same way. An array or a hash flattens into
// as many arguments as it holds, so a call passing one has no arity to
// choose by, and joins the candidates taking at least its other arguments:
// measured on 5.42.0 with a module whose two returns 1 for one argument and
// "two" for two, `my @a = (1, 2); two(@a)` and `my %h = (a => 1); two(%h)`
// are "two", and `my @a = (1); two(@a)` is 1. A slurpy takes however many
// fixed arguments precede an array: perl 5.42 runs `join_all(1, 2, 3, 4,
// 5, @a)`. An array or a hash is one argument only where the prototype
// the parse resolved the call's sub to gives its position a slot: bag's
// candidates derive `\[@%]`, and measured on 5.42.0, `sub g (\[@%]) {
// ref $_[0] }` gives ARRAY for `g(@x)` and HASH for `g(%h)`; a `$`, `_`,
// `*`, `+` or `;$` slot likewise passes `@x` as one argument. Candidates
// that disagree on a position, as box's, d's and e's arities do, derive no
// prototype, and with none perl flattens: `sub d { scalar @_ } my @x =
// (1, 2, 3); d(@x)` is 3, so `d(@x)` and `e(@x)` join their candidates.
func TestInferArityChoosesCandidate(t *testing.T) {
	cases := []struct {
		src, call string
		want      types.Type
	}{
		{"open(my $fh, '<', '/dev/null'); my $h = select($fh);", "select($fh)", types.Str},
		{"my @r = select(undef, undef, undef, 0.25);", "select(undef, undef, undef, 0.25)", types.List},
		{"use My::Multi; my $r = two(1);", "two(1)", types.Int},
		{"use My::Multi; my $r = two(1, 2);", "two(1, 2)", types.Str},
		{"use My::Multi; my @a = (1, 2); my $r = two(@a);", "two(@a)", types.Int | types.Str},
		{"use My::Multi; my %h = (a => 1); my $r = two(%h);", "two(%h)", types.Int | types.Str},
		{"use My::Multi; my @a = (2); my $r = two(1, @a);", "two(1, @a)", types.Int | types.Str},
		{"use My::Lib; my @a = (1); my $s = join_all(1, 2, 3, 4, 5, @a);", "join_all(1, 2, 3, 4, 5, @a)", types.Str},
		{"use My::Multi; my @x = (1, 2, 3); my $r = box(@x);", "box(@x)", types.Int | types.Str},
		{"use My::Multi; my @x = (1); my @y = (2); my $r = box(@x, @y);", "box(@x, @y)", types.Int | types.Str},
		{"use My::Multi; my @x = (1, 2, 3); my $r = bag(@x);", "bag(@x)", types.Int},
		{"use My::Multi; my %h = (a => 1); my $r = bag(%h);", "bag(%h)", types.Str},
		{"use My::Multi; my @x = (1, 2, 3); my $r = d(@x);", "d(@x)", types.Int | types.Str},
		{"use My::Multi; my @x = (1, 2, 3); my $r = e(@x);", "e(@x)", types.Int | types.Str},
	}
	for _, c := range cases {
		got, diags := declaredCallType(t, c.src, c.call)
		assert.Equal(t, c.want, got, "%q", c.src)
		assert.Empty(t, diags, "%q", c.src)
	}
}

// TestInferArgTypesChooseSameArityCandidate: My::Multi's which takes an
// Int, returning an Int, or a CodeRef, returning a Str.
func TestInferArgTypesChooseSameArityCandidate(t *testing.T) {
	cases := []struct {
		src, call string
		want      types.Type
	}{
		{"use My::Multi; my $r = which(1);", "which(1)", types.Int},
		{"use My::Multi; my $r = which(sub { 1 });", "which(sub { 1 })", types.Str},
	}
	for _, c := range cases {
		got, diags := declaredCallType(t, c.src, c.call)
		assert.Equal(t, c.want, got, "%q", c.src)
		assert.Empty(t, diags, "%q", c.src)
	}
}

// TestInferUnimportedExportOkUntyped: My::OkOnly marks f `:export_ok`, so
// a bare `use My::OkOnly` does not import it and the call is to no
// declared sub; naming f in the use list does import it.
func TestInferUnimportedExportOkUntyped(t *testing.T) {
	got, _ := declaredCallType(t, "use My::OkOnly; my $y = f(1);", "f(1)")
	assert.Equal(t, types.Unknown, got)
	got, _ = declaredCallType(t, "use My::OkOnly qw(f); my $y = f(1);", "f(1)")
	assert.Equal(t, types.Int, got, "the fixture declares f, imported when asked for")
}

// TestInferLocalSubShadowsDeclaration: a file's own sub lone, defined
// before the call, is the one perl calls, not My::Multi's declared one.
// Measured on 5.42.0 with a module exporting `pick`, a local `sub pick`
// makes a later `pick(1)` call the local sub ("Subroutine pick
// redefined").
func TestInferLocalSubShadowsDeclaration(t *testing.T) {
	got, diags := declaredCallType(t, `use My::Multi; sub lone { return "s" } my $r = lone([]);`, "lone([])")
	assert.Equal(t, types.Str, got)
	assert.Empty(t, diags)
}

// TestInferTypesQualifiedCall: Carp.tpm declares `sub shortmess
// :export_ok (List[Str] @message) Str`, so a call by its qualified name,
// with Carp loaded and nothing imported, is a Str. Measured on 5.42.0,
// `use Carp (); my $s = Carp::shortmess("x")` leaves $s "x at -e line 1.\n".
func TestInferTypesQualifiedCall(t *testing.T) {
	src := `use Carp (); my $s = Carp::shortmess("x");`
	got, diags := declaredCallType(t, src, `Carp::shortmess("x")`)
	assert.Equal(t, types.Str, got)
	assert.Empty(t, diags)
}

// TestInferEncodeCheckAcceptsUndefLiteral: perl 5.42's Encode accepts a
// literal undef CHECK (measured: `Encode::encode("ascii", $s, undef)`
// returns the encoded string, the same as omitting it), so Encode.tpm's
// declared CHECK type must admit Undef or the call reports a spurious
// coercion-mismatch.
func TestInferEncodeCheckAcceptsUndefLiteral(t *testing.T) {
	src := `use Encode (); my $s = Encode::encode("ascii", "x", undef);`
	got, diags := declaredCallType(t, src, `Encode::encode("ascii", "x", undef)`)
	assert.Equal(t, types.Str, got)
	assert.Empty(t, diags, "a literal undef CHECK should not be reported as a coercion mismatch")
}

// sortCandidate is the CORE.tpm sort candidate the call site selects for
// the first sort of src, and whether it selected one.
func sortCandidate(t *testing.T, src string) (types.Signature, bool) {
	t.Helper()
	sel := infer.SortSelection(src)
	if sel.Outcome != types.Selected {
		return types.Signature{}, false
	}
	return parse.CoreBuiltin("sort")[sel.Candidate], true
}

// TestInferSortCommaSelectsPlain: with a comma after it the routine's name
// is data, sorted among the rest, so the call writes no invocant and the
// plain candidate is selected. Measured on 5.42.0, with `sub byname { $a
// cmp $b }` and @x = qw(c a b), `sort "byname", @x` is (a, b, byname, c).
func TestInferSortCommaSelectsPlain(t *testing.T) {
	for _, call := range []string{`sort "byname", @x`, `sort @x`} {
		src := `my @x = qw(c a b); my @s = ` + call + `;`
		c, ok := sortCandidate(t, src)
		require.True(t, ok, "%q", src)
		assert.Nil(t, c.Invocant, "%q", src)
		require.Len(t, c.Params, 1, "%q", src)
		got, _ := declaredCallType(t, src, call)
		assert.Equal(t, types.List, got, "%q", src)
	}
}

// TestInferSortInvocantNotPlain: a comparison routine in the leading slot,
// named or held in a scalar, is the invocant the call writes, not data,
// so the invocant candidate is selected, and a block selects the block
// candidate. Measured on 5.42.0, with @x = qw(c a b), `sort byname @x` and
// `sort $by @x` with $by = "byname" are (a, b, c).
func TestInferSortInvocantNotPlain(t *testing.T) {
	for _, src := range []string{
		`sub byname { $a cmp $b } my @x = qw(c a b); my @s = sort byname @x;`,
		`my $by = "byname"; my @x = qw(c a b); my @s = sort $by @x;`,
	} {
		c, ok := sortCandidate(t, src)
		require.True(t, ok, "%q", src)
		assert.NotNil(t, c.Invocant, "%q", src)
	}
	c, ok := sortCandidate(t, `my @x = qw(c a b); my @s = sort { $a cmp $b } @x;`)
	require.True(t, ok)
	assert.Nil(t, c.Invocant)
	require.Len(t, c.Params, 2)
	assert.Equal(t, types.Code, c.Params[0].Type)
}

// TestInferDeleteSliceNotElement: CORE.tpm's delete takes a `@`-shaped
// slice or a `$`-shaped element, so the call site passes the operand's
// shape and a slice selects the List candidate. Measured on 5.42.0, with
// %h = (a => 1, b => 2) and @a = (5, 6, 7), `delete @h{qw(a b)}` is (1,
// 2), `delete %h{qw(a b)}` (a, 1, b, 2), `delete @a[0, 1]` (5, 6) and
// `delete %a[0, 1]` (0, 5, 1, 6), while `delete $h{a}` is 1.
func TestInferDeleteSliceNotElement(t *testing.T) {
	const decl = "my %h = (a => 1, b => 2); my @a = (5, 6, 7); "
	cases := []struct {
		call string
		want types.Type
	}{
		{"delete @h{qw(a b)}", types.List},
		{"delete %h{qw(a b)}", types.List},
		{"delete @a[0, 1]", types.List},
		{"delete %a[0, 1]", types.List},
	}
	for _, c := range cases {
		src := decl + "my @v = " + c.call + ";"
		got, diags := declaredCallType(t, src, c.call)
		assert.Equal(t, c.want, got, "%q", src)
		assert.Empty(t, diags, "%q", src)
	}
	got, _ := declaredCallType(t, decl+"my $v = delete $h{a};", "delete $h{a}")
	assert.Equal(t, types.Scalar, got)
}

// TestInferConditionMatchNotList: a condition is evaluated in scalar
// context, where a match is a Boolean, not the List of its captures.
// Measured on 5.42.0, a sub reporting wantarray says "scalar" as the
// condition of if, elsif, unless, a statement modifier, while, a C-style
// for, a ternary, and as the operand of `!` and of `not`.
func TestInferConditionMatchNotList(t *testing.T) {
	const decl = "my $s = 'abc'; "
	cases := []struct{ src, match string }{
		{"if ($s =~ /b/) { print 1 }", "$s =~ /b/"},
		{"if (0) { print 0 } elsif ($s =~ /b/) { print 1 }", "$s =~ /b/"},
		{"unless ($s =~ /b/) { print 1 }", "$s =~ /b/"},
		{"print 1 if $s =~ /b/;", "$s =~ /b/"},
		{"while ($s =~ /b/) { last }", "$s =~ /b/"},
		{"for (; $s =~ /b/; ) { last }", "$s =~ /b/"},
		{"my $r = $s =~ /b/ ? 1 : 0;", "$s =~ /b/"},
		// The parenthesised match spans its parentheses.
		{"my $r = !($s =~ /b/);", "($s =~ /b/)"},
		{"my $r = not $s =~ /b/;", "$s =~ /b/"},
	}
	for _, c := range cases {
		got, diags := declaredCallType(t, decl+c.src, c.match)
		assert.Equal(t, types.Boolean, got, "%q", c.src)
		assert.Empty(t, diags, "%q", c.src)
	}
}

// TestInferVoidMatchIsUntypedNotError: no `=~` candidate answers void
// context, yet a match or substitution in void context is ordinary perl,
// so it is untyped there and never reported. perl 5.42 -w is silent on
// `$s =~ s/a/b/g; print $s` and on `$s =~ /X/;`, in a file or a sub's
// block. A sub's last statement is its value, not void: `sub f { $_[0] =~
// s/b/c/g; }` returns the count.
func TestInferVoidMatchIsUntypedNotError(t *testing.T) {
	for _, match := range []string{"$s =~ s/a/b/g", "$s =~ /X/"} {
		for _, src := range []string{
			`my $s = "aXa"; ` + match + `; print $s;`,
			`my $s = "aXa"; sub f { ` + match + `; return 1 }`,
		} {
			got, diags := declaredCallType(t, src, match)
			assert.Equal(t, types.Unknown, got, "%q", src)
			assert.Empty(t, diags, "%q", src)
		}
	}
	got, _ := declaredCallType(t, `sub f { $_[0] =~ s/b/c/g; }`, "$_[0] =~ s/b/c/g")
	assert.Equal(t, types.Str, got)
}

// TestMooseNamedListNoFalseArity: `use Moose qw(has);` installs has and
// meta alone, not blessed (TestMooseNamedListInstallsNamesAndMeta,
// internal/parse) -- so a later blessed() is an ordinary call, not one
// to a sub whose `($)` prototype refuses it for taking too few
// arguments. crochet:review round 2 (quality S15, issue 01a15efc) found
// the false call_arity this guards: the shipped Moose.tpm used to
// install all 13 names from any non-empty list. Needs the signature's
// named parameters and :always, backed out (01a15f6c) until the export
// syntax is redesigned (01a15f6b): until then Moose over-installs on a
// named list again, and this false call_arity is back, an accepted,
// tracked state (01a15f6b).
func TestMooseNamedListNoFalseArity(t *testing.T) {
	t.Skip("backed out with :always (01a15f6c) until the export syntax is redesigned (01a15f6b)")
	root := parse.ParseWithLoader([]byte("use Moose qw(has);\nblessed();\n"), func(string) ([]byte, bool) { return nil, false })
	var codes []parse.RefusalCode
	var walk func(*parse.Node)
	walk = func(n *parse.Node) {
		if n.Kind == parse.Unknown {
			codes = append(codes, n.Refusal)
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(root)
	assert.NotContains(t, codes, parse.CallArity, "blessed() refused for arity though qw(has) never imports it")
}

// TestCwdGetcwdIgnoresExtraArgs: Cwd's subs that ignore trailing
// arguments take them without an arity error, and the ones perl dies on
// (on Linux and macOS) still report. Measured on Linux, 5.42.0, perl -w: cwd(1,2,3),
// getcwd(1,2,3), fastgetcwd(1,2,3), fast_abs_path('.',2,3),
// fast_realpath('.',2,3) and Cwd::chdir('/',2) all run silently, while
// fastcwd(1) dies "Usage: Cwd::fastcwd()" and abs_path('.',2) and
// realpath('.',2) die "Usage: Cwd::abs_path(pathsv=Nullsv)".
func TestCwdGetcwdIgnoresExtraArgs(t *testing.T) {
	for _, call := range []string{"cwd(1, 2, 3)", "getcwd(1, 2, 3)", "fastgetcwd(1, 2, 3)",
		"fast_abs_path('.', 2, 3)", "fast_realpath('.', 2, 3)", "chdir('/', 2)"} {
		_, diags := declaredCallType(t, "use Cwd qw(cwd getcwd fastgetcwd fast_abs_path fast_realpath chdir); my $r = "+call+";", call)
		assert.Empty(t, diags, call)
	}
	for _, call := range []string{"fastcwd(1)", "abs_path('.', 2)", "realpath('.', 2)"} {
		_, diags := declaredCallType(t, "use Cwd qw(fastcwd abs_path realpath); my $r = "+call+";", call)
		require.Len(t, diags, 1, call)
		assert.Equal(t, infer.CodeArityMismatch, diags[0].Code, diags[0].Message)
	}
}

// TestIgnoredArgumentsNoFalseArity: subs whose bodies never read their
// arguments take extra ones without an arity error. Measured on 5.42.0
// with perl -w, each call below runs silently: FileHandle::pipe and
// DynaLoader::dl_load_flags read no @_, ExtUtils::Manifest's manicheck,
// filecheck and fullcheck call their helpers with no arguments, and
// skipcheck copies its first argument into a lexical it never uses
// (`skipcheck(undef, {}, 3)` runs, as does `skipcheck()`). mkmanifest
// is declared from its source, which never reads @_; it is not run,
// since it writes files.
func TestIgnoredArgumentsNoFalseArity(t *testing.T) {
	for src, call := range map[string]string{
		"use FileHandle; my @r = FileHandle::pipe(1, 2);":                        "FileHandle::pipe(1, 2)",
		"use DynaLoader; my $r = DynaLoader::dl_load_flags(1, 2);":               "DynaLoader::dl_load_flags(1, 2)",
		"use ExtUtils::Manifest qw(manicheck); my @r = manicheck(1, 2);":         "manicheck(1, 2)",
		"use ExtUtils::Manifest qw(filecheck); my @r = filecheck(1, 2);":         "filecheck(1, 2)",
		"use ExtUtils::Manifest qw(fullcheck); my @r = fullcheck(1, 2);":         "fullcheck(1, 2)",
		"use ExtUtils::Manifest qw(mkmanifest); mkmanifest(1, 2);":               "mkmanifest(1, 2)",
		"use ExtUtils::Manifest qw(skipcheck); my @r = skipcheck(undef, {}, 3);": "skipcheck(undef, {}, 3)",
		"use ExtUtils::Manifest qw(skipcheck); my @r = skipcheck('x');":          "skipcheck('x')",
		"use ExtUtils::Manifest qw(skipcheck); my @r = skipcheck();":             "skipcheck()",
	} {
		_, diags := declaredCallType(t, src, call)
		assert.Empty(t, diags, call)
	}
}
