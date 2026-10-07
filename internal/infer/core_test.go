// ABOUTME: Tests that inference reads builtin and operator types from CORE.pmt's declarations.
// ABOUTME: Each expectation is perl 5.42's measured behaviour, not a transcription of the table.

package infer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// lookupType is the type inference gave the variable name in src.
func lookupType(t *testing.T, src, name string) types.Type {
	t.Helper()
	_, _, st := analyzeSourceFull(t, []byte(src))
	sym, found := st.Lookup(name)
	require.True(t, found, "%s should be in the symbol table of %q", name, src)
	return sym.Type
}

// TestInferTypesBuiltinsFromCore: a builtin is typed by its CORE.pmt line,
// `sub hex (Str $string = $_) Int;`. Measured, `hex("ff")` is 255, created
// as a number.
func TestInferTypesBuiltinsFromCore(t *testing.T) {
	assert.Equal(t, types.Int, lookupType(t, `my $n = hex("ff");`, "$n"))
}

// TestInferCoreArityAcceptsPerlsCalls: a call perl compiles is not flagged
// for its arity. Measured on 5.42, each of these compiles and runs: push
// and unshift need only the array, join only the separator, and abs, int,
// uc, lc, ucfirst, lcfirst and reverse default their argument.
func TestInferCoreArityAcceptsPerlsCalls(t *testing.T) {
	for _, src := range []string{
		"my @a; push @a;",
		"my @a; push(@a);",
		"my @a; unshift @a;",
		`my $j = join(":");`,
		"my $n = abs();",
		"my $n = int();",
		"my $s = uc();",
		"my $s = lc();",
		"my $s = ucfirst();",
		"my $s = lcfirst();",
		"my @r = reverse();",
	} {
		_, diags := analyzeSource(t, []byte(src))
		for _, d := range diags {
			assert.NotEqual(t, infer.CodeArityMismatch, d.Code, "%q: %s", src, d.Message)
		}
	}
}

// TestInferScalarSlotTakesAnAggregateAsItsCount: CORE.pmt's `sub scalar
// (Scalar $expr) Scalar;` takes an aggregate in scalar context, as perl's `$`
// prototype does. Measured, `scalar(@a)` on (1,2,3) is 3 and `scalar(%h)` on
// one pair is 1 -- the idiom, not a type error.
func TestInferScalarSlotTakesAnAggregateAsItsCount(t *testing.T) {
	for _, src := range []string{
		"my @a = (1,2,3); my $n = scalar(@a);",
		"my %h = (a => 1); my $n = scalar(%h);",
	} {
		_, diags := analyzeSource(t, []byte(src))
		assert.Empty(t, diags, "%q", src)
	}
}

// TestInferUndeclaredNameHasNoCoreType: a call to a name CORE.pmt does not
// declare gets no builtin type, and no builtin's arity or argument checks.
func TestInferUndeclaredNameHasNoCoreType(t *testing.T) {
	src := []byte("my $n = frobnicate(1);")
	ann, diags := analyzeSource(t, src)
	got, _ := findNodeType(ann, src, "frobnicate(1)")
	assert.Equal(t, types.Unknown, got)
	assert.Empty(t, diags)
}

// TestInferCoreCoercionMismatchStillReported: a lossy coercion is reported
// through CORE.pmt's types. `bless "x"` passes a Str where `sub bless (Ref
// $ref, ...)` takes a Ref, which no Str coerces to; a Str where a Num is
// declared, abs's argument or +'s operand, is reported as the coercion it is.
func TestInferCoreCoercionMismatchStillReported(t *testing.T) {
	cases := []struct {
		src      string
		code     string
		severity infer.Severity
	}{
		{`my $o = bless "x";`, infer.CodeTypeMismatch, infer.Error},
		{`my $s = "abc"; my $n = abs($s);`, infer.CodeTypeMismatch, infer.Warning},
		{`my $s = "abc"; my $n = $s + 1;`, infer.CodeCoercionMismatch, infer.Warning},
	}
	for _, c := range cases {
		_, diags := analyzeSource(t, []byte(c.src))
		require.Len(t, diags, 1, "%q", c.src)
		assert.Equal(t, c.code, diags[0].Code, "%q: %s", c.src, diags[0].Message)
		assert.Equal(t, c.severity, diags[0].Severity, "%q: %s", c.src, diags[0].Message)
	}
}

// TestInferTypesOperatorsFromCore: an operator's result is its CORE.pmt
// declaration's. `sub <=> :infix(ORDERING) (Num $x, Num $y) Int|Undef;`:
// measured, `1 <=> "nan"` is undef. `x` forks on its left operand's
// parenthesis and its context: measured, `my @y = (1,2) x 2` is (1,2,1,2)
// and `my $s = "ab" x 2` is "abab".
func TestInferTypesOperatorsFromCore(t *testing.T) {
	cases := []struct {
		src, name string
		want      types.Type
	}{
		{"my $c = 1 <=> 2;", "$c", types.Int | types.Undef},
		{"my @y = (1,2) x 2;", "@y", types.List},
		{`my $s = "ab" x 2;`, "$s", types.Str},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, lookupType(t, c.src, c.name), "%q", c.src)
	}
}

// TestInferRepeatScalarContextIsStr: a parenthesised `x` in scalar context
// is the string repetition, not a list: measured, `my $x = (1,2) x 2` is
// "22", perl evaluating the parenthesised list in scalar context.
func TestInferRepeatScalarContextIsStr(t *testing.T) {
	assert.Equal(t, types.Str, lookupType(t, "my $x = (1,2) x 2;", "$x"))
}

// TestInferCoreListSlotsTakeEveryValue: push, unshift and join take a LIST,
// into which everything flattens -- measured on 5.42, scalars, arrays,
// hashes and references all append, and `join ":", @foo` joins the array's
// elements. keys, values and each take a hash or, since 5.12, an array.
// None of these calls is a mismatch.
func TestInferCoreListSlotsTakeEveryValue(t *testing.T) {
	for _, src := range []string{
		`my @a; my @b; my %h; push @a, 1, "s", {}, [], sub {1}, \*STDOUT, @b, %h;`,
		`my @a; my @b; my %h; unshift @a, 1, "s", {}, [], sub {1}, @b, %h;`,
		`my @foo; my $s = join ":", @foo;`,
		`my $s = join ":", "a", "b";`,
		`my @a; my %h; my @k = keys @a; my @v = values %h; my @e = each @a;`,
	} {
		_, diags := analyzeSource(t, []byte(src))
		assert.Empty(t, diags, "%q", src)
	}
}

// TestCoreListSlotExcludesCodeAndGlob: a LIST slot is not Any. Any is the
// escape hatch that disables checking; List still excludes the compiled CV
// and the typeglob, which no perl scalar holds, so the slot keeps
// constraining something.
func TestCoreListSlotExcludesCodeAndGlob(t *testing.T) {
	for _, name := range []string{"push", "unshift", "join"} {
		sigs := parse.CoreBuiltin(name)
		require.Len(t, sigs, 1, name)
		list := sigs[0].Params[1].Type
		assert.Equal(t, types.List, list, name)
		assert.False(t, types.TypeSatisfies(types.Code, list), "%s takes no bare CV", name)
		assert.False(t, types.TypeSatisfies(types.Glob, list), "%s takes no typeglob", name)
	}
}

// TestInferSplitPatternTakesRegexOrStr: split's pattern is a compiled Regex
// or a plain Str, which perl compiles into one: measured on 5.42, `split
// /:/, $x`, `split ":", $x`, `split $sep, $x` and `split qr/:/, $x` all give
// ("a","b","c") for "a:b:c". A reference there is still a mismatch.
func TestInferSplitPatternTakesRegexOrStr(t *testing.T) {
	for _, src := range []string{
		`my $x = "a:b:c"; my @f = split ":", $x;`,
		`my $x = "a:b:c"; my @f = split qr/:/, $x;`,
	} {
		_, diags := analyzeSource(t, []byte(src))
		assert.Empty(t, diags, "%q", src)
	}
	_, diags := analyzeSource(t, []byte(`my $h = {}; my $x = "a:b:c"; my @f = split $h, $x;`))
	assert.NotEmpty(t, diags, "a reference is not a pattern")
}

// TestInferReverseFollowsContext: reverse is two `:context` candidates, a
// list in list context and a string in scalar context. Measured on 5.42,
// with @x = ("ab","cd"), `my @r = reverse @x` is ("cd","ab") and `my $s =
// reverse @x` is "dcba". Where the call's context is unknown, its type is
// the join of the two, Str ⊔ List = List.
func TestInferReverseFollowsContext(t *testing.T) {
	assert.Equal(t, types.List, lookupType(t, "my @x; my @r = reverse @x;", "@r"))
	assert.Equal(t, types.Str, lookupType(t, "my @x; my $s = reverse @x;", "$s"))

	src := []byte("my @x; frobnicate(reverse @x);")
	ann, _ := analyzeSource(t, src)
	got, _ := findNodeType(ann, src, "reverse @x")
	assert.Equal(t, types.List, got)
}

// TestInferContextBuiltinsFollowContext: the other builtins CORE.pmt
// declares as `:context` candidates take the candidate their assignment's
// context selects. Measured on 5.42: `my $n = keys %h` is the count, `my $t
// = localtime 0` is "Thu Jan  1 00:00:00 1970" where the list form has nine
// elements, `my $p = readpipe("echo hi")` is "hi\n", and `my $st = stat
// "/"` is 1.
func TestInferContextBuiltinsFollowContext(t *testing.T) {
	cases := []struct {
		src, name string
		want      types.Type
	}{
		{"my %h; my $n = keys %h;", "$n", types.Int},
		{"my %h; my @k = keys %h;", "@k", types.List},
		{"my %h; my $n = values %h;", "$n", types.Int},
		{"my $t = localtime(0);", "$t", types.Str | types.Undef},
		{"my @l = localtime(0);", "@l", types.List},
		{"my $t = gmtime(0);", "$t", types.Str | types.Undef},
		{"my $p = readpipe('echo');", "$p", types.Str | types.Undef},
		{"my @p = readpipe('echo');", "@p", types.List},
		{"my $s = stat('/');", "$s", types.Boolean},
		{"my @s = stat('/');", "@s", types.List},
		{"my $s = lstat('/');", "$s", types.Boolean},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, lookupType(t, c.src, c.name), "%q", c.src)
	}
}

// TestInferMultiBuiltinNarrowsByArityAndShape: a multi with no `:context`
// is cut to the candidates its call's arity and operand shapes take before
// their returns are joined, without consulting argument types. Measured on
// 5.42: `my $d = delete $h{a}` is the deleted value, "x"; `my @s = delete
// @g{qw(a b)}` is (1, 2); `my $o = select(STDERR)` is "main::STDOUT", as is
// `select()`; four-argument select in list context has two elements.
func TestInferMultiBuiltinNarrowsByArityAndShape(t *testing.T) {
	cases := []struct {
		src, name string
		want      types.Type
	}{
		{"my %h; my $d = delete $h{a};", "$d", types.Scalar},
		{"my @a; my $d = delete $a[1];", "$d", types.Scalar},
		{"my %h; my @d = delete @h{qw(a b)};", "@d", types.List},
		{"my $o = select(STDERR);", "$o", types.Str},
		{"my $o = select();", "$o", types.Str},
		{"my @r = select(undef, undef, undef, 0.1);", "@r", types.List},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, lookupType(t, c.src, c.name), "%q", c.src)
	}
}
