// ABOUTME: Tests for the PSC type inference engine's Analyze function (pass 2).
// ABOUTME: Covers literal types, variable types, operator types, builtin calls, and diagnostics.

package infer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/types"
)

// analyzeSource parses the given Perl source and runs both inference passes,
// returning the annotation map and diagnostics.
func analyzeSource(t *testing.T, source []byte) (map[uint32]types.Type, []infer.Diagnostic) {
	t.Helper()
	p := parser.New()
	tree, err := p.Parse(source)
	require.NoError(t, err, "parse must succeed")
	annotations, diags, _ := infer.Analyze(tree, source)
	return annotations, diags
}

// analyzeSourceFull is like analyzeSource but also returns the SymbolTable,
// which is needed for tests that verify assignment-based type narrowing.
func analyzeSourceFull(t *testing.T, source []byte) (map[uint32]types.Type, []infer.Diagnostic, *infer.SymbolTable) {
	t.Helper()
	p := parser.New()
	tree, err := p.Parse(source)
	require.NoError(t, err, "parse must succeed")
	return infer.Analyze(tree, source)
}

// findNodeType searches the annotation map for the first node whose source
// text matches want, returning its type. It iterates all byte offsets for
// which source[offset:] starts with want.
func findNodeType(annotations map[uint32]types.Type, source []byte, want string) (types.Type, bool) {
	wantBytes := []byte(want)
	wl := uint32(len(wantBytes))
	for offset, typ := range annotations {
		if offset+wl > uint32(len(source)) {
			continue
		}
		if string(source[offset:offset+wl]) == want {
			return typ, true
		}
	}
	return types.Unknown, false
}

// --- Basic result shape ---

func TestAnalyzeReturnsResults(t *testing.T) {
	annotations, diags := analyzeSource(t, []byte("42;"))
	assert.NotNil(t, annotations, "annotations map must not be nil")
	assert.NotNil(t, diags, "diagnostics slice must not be nil")
}

// --- Literal inference ---

func TestInferIntegerLiteral(t *testing.T) {
	src := []byte("42;")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "42")
	require.True(t, ok, "node for '42' should be annotated")
	assert.Equal(t, types.Int, typ, "integer literal should have type Int")
}

func TestInferFloatLiteral(t *testing.T) {
	src := []byte("3.14;")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "3.14")
	require.True(t, ok, "node for '3.14' should be annotated")
	assert.Equal(t, types.Num, typ, "float literal should have type Num")
}

func TestInferIntegerWithExponent(t *testing.T) {
	src := []byte("1e10;")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "1e10")
	require.True(t, ok, "node for '1e10' should be annotated")
	assert.Equal(t, types.Num, typ, "number with exponent should have type Num")
}

func TestInferHexLiteral(t *testing.T) {
	src := []byte("0xFF;")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "0xFF")
	require.True(t, ok, "node for '0xFF' should be annotated")
	assert.Equal(t, types.Int, typ, "hex literal should have type Int (not Num because 'e' is inside 0x prefix)")
}

// --- Variable inference ---

func TestInferScalarVariable(t *testing.T) {
	src := []byte("my $x = 1; $x;")
	annotations, _ := analyzeSource(t, src)

	// The $x in the declaration (offset 3) is visited before narrowing, so
	// it gets the sigil type Scalar. The $x reference (offset 11) is visited
	// after the assignment narrows $x to Int.
	declOffset := uint32(3) // "my " is 3 bytes
	refOffset := uint32(11) // "my $x = 1; " is 11 bytes

	declType, declOk := annotations[declOffset]
	require.True(t, declOk, "declaration $x at offset %d should be annotated", declOffset)
	assert.Equal(t, types.Scalar, declType, "declaration $x should have sigil type Scalar")

	refType, refOk := annotations[refOffset]
	require.True(t, refOk, "reference $x at offset %d should be annotated", refOffset)
	assert.Equal(t, types.Int, refType, "reference $x should have narrowed type Int")
}

func TestInferArrayVariable(t *testing.T) {
	src := []byte("my @arr; @arr;")
	annotations, _ := analyzeSource(t, src)

	found := false
	for offset, typ := range annotations {
		if int(offset) < len(src) && offset+4 <= uint32(len(src)) && string(src[offset:offset+4]) == "@arr" {
			assert.Equal(t, types.Array, typ, "array variable should have type Array")
			found = true
			break
		}
	}
	assert.True(t, found, "should find annotation for @arr")
}

func TestInferHashVariable(t *testing.T) {
	src := []byte("my %h; %h;")
	annotations, _ := analyzeSource(t, src)

	found := false
	for offset, typ := range annotations {
		if int(offset) < len(src) && offset+2 <= uint32(len(src)) && string(src[offset:offset+2]) == "%h" {
			assert.Equal(t, types.Hash, typ, "hash variable should have type Hash")
			found = true
			break
		}
	}
	assert.True(t, found, "should find annotation for %h")
}

func TestInferArraylen(t *testing.T) {
	src := []byte("$#arr;")
	annotations, _ := analyzeSource(t, src)
	// arraylen node covers the entire "$#arr"
	typ, ok := findNodeType(annotations, src, "$#arr")
	require.True(t, ok, "node for '$#arr' should be annotated")
	assert.Equal(t, types.Int, typ, "arraylen should have type Int")
}

// --- Binary operator inference ---

func TestInferBinaryAddition(t *testing.T) {
	src := []byte("1 + 2;")
	annotations, _ := analyzeSource(t, src)
	// The binary_expression node covers "1 + 2"
	typ, ok := findNodeType(annotations, src, "1 + 2")
	require.True(t, ok, "binary_expression '1 + 2' should be annotated")
	assert.Equal(t, types.Num, typ, "addition should have result type Num")
}

func TestInferEqualityExpression(t *testing.T) {
	src := []byte("1 == 2;")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "1 == 2")
	require.True(t, ok, "equality_expression '1 == 2' should be annotated")
	assert.Equal(t, types.Bool, typ, "equality should return Bool")
}

func TestInferRelationalExpression(t *testing.T) {
	src := []byte("1 < 2;")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "1 < 2")
	require.True(t, ok, "relational_expression '1 < 2' should be annotated")
	assert.Equal(t, types.Bool, typ, "less-than should return Bool")
}

func TestInferStringConcatExpression(t *testing.T) {
	src := []byte("1 . 2;")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "1 . 2")
	require.True(t, ok, "binary_expression '1 . 2' should be annotated")
	assert.Equal(t, types.Str, typ, "string concat should return Str")
}

func TestInferLowprecLogicalExpression(t *testing.T) {
	src := []byte("1 and 2;")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "1 and 2")
	require.True(t, ok, "lowprec_logical_expression '1 and 2' should be annotated")
	assert.Equal(t, types.Any, typ, "low-prec logical should return Any")
}

// --- Unary operator inference ---

func TestInferUnaryNot(t *testing.T) {
	src := []byte("!1;")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "!1")
	require.True(t, ok, "unary_expression '!1' should be annotated")
	assert.Equal(t, types.Bool, typ, "logical not should return Bool")
}

func TestInferUnaryMinus(t *testing.T) {
	src := []byte("my $x = 2; -$x;")
	annotations, _ := analyzeSource(t, src)
	// "-$x" starts at offset 11
	typ, ok := findNodeType(annotations, src, "-$x")
	require.True(t, ok, "unary_expression '-$x' should be annotated")
	assert.Equal(t, types.Num, typ, "unary minus should return Num")
}

// --- Builtin function call inference ---

func TestInferBuiltinReturnType(t *testing.T) {
	src := []byte("my @arr; push(@arr, 1);")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "push(@arr, 1)")
	require.True(t, ok, "push call should be annotated")
	assert.Equal(t, types.Int, typ, "push should return Int")
}

func TestInferFunc1opReturnType(t *testing.T) {
	src := []byte("scalar(42);")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "scalar(42)")
	require.True(t, ok, "scalar() call should be annotated")
	assert.Equal(t, types.Scalar, typ, "scalar() should return Scalar")
}

func TestInferAmbiguousFunctionCallReturnType(t *testing.T) {
	src := []byte("my @a; push @a, 1;")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "push @a, 1")
	require.True(t, ok, "ambiguous push call should be annotated")
	assert.Equal(t, types.Int, typ, "push should return Int")
}

// --- Diagnostic tests ---

func TestDiagnosticArityMismatch(t *testing.T) {
	src := []byte("push();")
	_, diags := analyzeSource(t, src)
	require.NotEmpty(t, diags, "should emit at least one diagnostic for push() with no args")

	var found bool
	for _, d := range diags {
		if d.Code == infer.CodeArityMismatch {
			found = true
			assert.Equal(t, infer.Error, d.Severity, "arity mismatch should be an Error")
			break
		}
	}
	assert.True(t, found, "should find an arity-mismatch diagnostic for push()")
}

func TestDiagnosticTypeMismatch(t *testing.T) {
	src := []byte("my $x = 1;\npush($x, 1);\n")
	_, diags := analyzeSource(t, src)
	require.NotEmpty(t, diags, "should emit at least one diagnostic for push($x, 1)")

	var found bool
	for _, d := range diags {
		if d.Code == infer.CodeTypeMismatch {
			found = true
			assert.Equal(t, infer.Error, d.Severity, "type mismatch should be an Error")
			break
		}
	}
	assert.True(t, found, "should find a type-mismatch diagnostic for push($x, 1) where $x is Scalar not Array")
}

func TestMethodCallAnnotatedAny(t *testing.T) {
	src := []byte("my $obj; $obj->method();")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "$obj->method()")
	require.True(t, ok, "method_call_expression should be annotated")
	assert.Equal(t, types.Any, typ, "method call return type should be Any")
}

func TestConditionalExpressionAnnotatedAny(t *testing.T) {
	src := []byte("my $x = 1; my $y = 2; $x ? $y : 0;")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "$x ? $y : 0")
	require.True(t, ok, "conditional_expression should be annotated")
	assert.Equal(t, types.Any, typ, "conditional expression should have type Any")
}

// --- Assignment narrowing tests ---

func TestNarrowingIntegerAssignment(t *testing.T) {
	src := []byte("my $x = 42;")
	_, _, st := analyzeSourceFull(t, src)

	sym, ok := st.Lookup("$x")
	require.True(t, ok, "$x should be in the symbol table")
	assert.Equal(t, types.Int, sym.Type, "$x should be narrowed to Int after 'my $x = 42'")
}

func TestNarrowingFloatAssignment(t *testing.T) {
	src := []byte("my $n = 3.14;")
	_, _, st := analyzeSourceFull(t, src)

	sym, ok := st.Lookup("$n")
	require.True(t, ok, "$n should be in the symbol table")
	assert.Equal(t, types.Num, sym.Type, "$n should be narrowed to Num after 'my $n = 3.14'")
}

func TestNarrowingReassignment(t *testing.T) {
	src := []byte("my $x = 42; $x = 3.14;")
	_, _, st := analyzeSourceFull(t, src)

	sym, ok := st.Lookup("$x")
	require.True(t, ok, "$x should be in the symbol table")
	assert.Equal(t, types.Num, sym.Type, "$x should be narrowed to Num after reassignment with 3.14")
}

// --- Narrowed variable annotation tests ---

func TestNarrowedVariableAnnotation(t *testing.T) {
	// After "my $x = 42;", a subsequent reference to $x should be annotated as
	// Int (the narrowed type), not Scalar (the sigil type).
	src := []byte("my $x = 42; $x;")
	annotations, _ := analyzeSource(t, src)

	// The second $x reference starts at byte offset 12 ("my $x = 42; " is 12 bytes)
	refOffset := uint32(12)
	typ, ok := annotations[refOffset]
	require.True(t, ok, "the $x reference node at offset %d should be annotated", refOffset)
	assert.Equal(t, types.Int, typ, "$x reference should be annotated as Int after narrowing")
}

func TestNarrowedVariableInBuiltinCall(t *testing.T) {
	// chr() expects Int. After "my $n = 3.14;", $n is narrowed to Num.
	// TypeSatisfies(Num, Int) is false, so a type-mismatch diagnostic
	// should be produced. Without narrowing, $n would be Scalar (polymorphic)
	// and no diagnostic would fire.
	src := []byte("my $n = 3.14; chr($n);")
	_, diags := analyzeSource(t, src)

	var found bool
	for _, d := range diags {
		if d.Code == infer.CodeTypeMismatch {
			found = true
			break
		}
	}
	assert.True(t, found, "chr(Num) should produce a type-mismatch diagnostic because chr expects Int")
}
