// ABOUTME: Tests that inference reads builtin and operator types from CORE.pmt's declarations.
// ABOUTME: Each expectation is perl 5.42's measured behaviour, not a transcription of the table.

package infer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/infer"
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
