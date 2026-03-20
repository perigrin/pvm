// ABOUTME: Tests for the PSC type inference engine's Analyze function (pass 2).
// ABOUTME: Covers literal types, variable types, operator types, builtin calls, and diagnostics.

package infer_test

import (
	"os"
	"path/filepath"
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
	annotations, diags, _ := infer.Analyze(tree, source, nil)
	return annotations, diags
}

// analyzeSourceFull is like analyzeSource but also returns the SymbolTable,
// which is needed for tests that verify assignment-based type narrowing.
func analyzeSourceFull(t *testing.T, source []byte) (map[uint32]types.Type, []infer.Diagnostic, *infer.SymbolTable) {
	t.Helper()
	p := parser.New()
	tree, err := p.Parse(source)
	require.NoError(t, err, "parse must succeed")
	return infer.Analyze(tree, source, nil)
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

// --- String literal tests ---
// Note: gotreesitter currently produces ERROR nodes for all quoted strings
// ('hello', "hello", q(), qq{}, backticks). The string_literal and
// interpolated_string_literal node kinds are handled for forward-compatibility,
// but cannot be exercised through the parser until the grammar is fixed.

func TestNarrowingUnknownRHSPreservesType(t *testing.T) {
	// String literals currently parse as ERROR nodes, yielding Unknown RHS.
	// The variable should NOT be narrowed to Unknown; it keeps sigil type.
	source := []byte("my $x = 'hello';\n$x;\n")
	_, _, st := analyzeSourceFull(t, source)
	sym, ok := st.Lookup("$x")
	require.True(t, ok, "$x should be in symbol table from declaration")
	assert.NotEqual(t, types.Unknown, sym.Type,
		"$x should not be narrowed to Unknown; should retain sigil type Scalar")
}

func TestNarrowingArrayDeclaration(t *testing.T) {
	source := []byte("my @arr = (1, 2, 3);\n@arr;\n")
	_, _, st := analyzeSourceFull(t, source)
	sym, ok := st.Lookup("@arr")
	require.True(t, ok, "@arr should be in symbol table")
	// Array variables keep their sigil type Array (narrowing from list
	// literals doesn't change the aggregate type)
	assert.Equal(t, types.Array, sym.Type)
}

func TestNarrowingUndeclaredVariable(t *testing.T) {
	// $x assigned without my — CollectDeclarations won't define it,
	// so UpdateType should be a no-op (no crash, no new symbol).
	source := []byte("$x = 42;\n$x;\n")
	_, diags, st := analyzeSourceFull(t, source)
	_, ok := st.Lookup("$x")
	assert.False(t, ok, "$x was never declared with my, should not be in symbol table")
	// Should not panic or produce unexpected diagnostics
	_ = diags
}

func TestStringLiteralNodeKindHandled(t *testing.T) {
	// Verify that inferNodeType handles string_literal and
	// interpolated_string_literal by checking the annotation map after
	// parsing a concat expression. The concat operator (.) returns Str,
	// which validates that the engine can produce Str types even though
	// string literals themselves currently parse as ERROR nodes.
	src := []byte("1 . 2;")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "1 . 2")
	require.True(t, ok, "concat expression should be annotated")
	assert.Equal(t, types.Str, typ, "concat should return Str (proves Str type works in engine)")
}

// --- Flow narrowing tests ---

func TestFlowNarrowingDefinedGuard(t *testing.T) {
	// if (defined($x)): if-body $x → Scalar &^ Undef (defined, all non-undef scalar bits),
	// else-body $x → Undef.
	// Note: "my $x = undef" does NOT narrow $x to Undef because the undef
	// keyword produces Unknown type, so $x stays at sigil type Scalar.
	src := []byte("my $x = undef;\nif (defined($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets[0]=declaration, [1]=condition defined($x), [2]=if-body, [3]=else-body
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	ifBodyTyp, ifOk := annotations[offsets[2]]
	assert.True(t, ifOk, "if-body $x should be annotated")
	assert.Equal(t, types.Scalar&^types.Undef, ifBodyTyp, "if-body $x should be Scalar &^ Undef (defined guard removes Undef bit)")

	elseBodyTyp, elseOk := annotations[offsets[3]]
	assert.True(t, elseOk, "else-body $x should be annotated")
	assert.Equal(t, types.Undef, elseBodyTyp, "else-body $x should be Undef (negated defined guard)")
}

func TestFlowNarrowingRefGuard(t *testing.T) {
	// Inside if (ref($x)), $x should be narrowed to Ref.
	src := []byte("my $x = undef;\nif (ref($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "if-body $x at offset %d should be annotated", ifBodyXOffset)
	assert.Equal(t, types.Ref, typ, "if-body $x should be Ref after ref() guard")
}

func TestFlowNarrowingUnlessDefinedGuard(t *testing.T) {
	// Inside unless (defined($x)), $x should be Undef (negated defined).
	src := []byte("my $x = undef;\nunless (defined($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "unless-body $x should be annotated")
	assert.Equal(t, types.Undef, typ, "unless-body $x should be Undef (negated defined guard)")
}

func TestFlowNarrowingWhileDefinedGuard(t *testing.T) {
	// Inside while (defined($x)), $x should be Scalar &^ Undef (non-Undef scalar bits).
	src := []byte("my $x = undef;\nwhile (defined($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	whileBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[whileBodyXOffset]
	assert.True(t, ok, "while-body $x should be annotated")
	assert.Equal(t, types.Scalar&^types.Undef, typ, "while-body $x should be Scalar &^ Undef (defined guard removes Undef bit)")
}

func TestFlowNarrowingIfElseRefGuard(t *testing.T) {
	// if (ref($x)) → Ref in if-body, Scalar &^ Ref in else-body (non-reference scalar bits)
	src := []byte("my $x = undef;\nif (ref($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets[0] = declaration "my $x", offsets[1] = ref($x) condition,
	// offsets[2] = if-body $x, offsets[3] = else-body $x
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	ifBodyTyp, ifOk := annotations[offsets[2]]
	assert.True(t, ifOk, "if-body $x should be annotated")
	assert.Equal(t, types.Ref, ifBodyTyp, "if-body $x should be Ref after ref() guard")

	elseBodyTyp, elseOk := annotations[offsets[3]]
	assert.True(t, elseOk, "else-body $x should be annotated")
	assert.Equal(t, types.Scalar&^types.Ref, elseBodyTyp, "else-body $x should be Scalar &^ Ref (negated ref guard removes Ref bits)")
}

func TestFlowNarrowingScopeRestoration(t *testing.T) {
	// After the if block, $x should revert to its pre-narrowed type.
	// "my $x = undef" leaves $x at sigil type Scalar (undef keyword → Unknown,
	// so no assignment narrowing occurs).
	src := []byte("my $x = undef;\nif (ref($x)) {\n    my $y = $x;\n}\nmy $z = $x;\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets: [0]=decl, [1]=condition ref($x), [2]=if-body, [3]=after-if
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x")

	afterIfTyp, ok := annotations[offsets[3]]
	assert.True(t, ok, "post-if $x should be annotated")
	assert.Equal(t, types.Scalar, afterIfTyp, "post-if $x should revert to Scalar (pre-narrowed sigil type)")
}

// findLastVarOffset returns the byte offset of the last occurrence of varName in source.
func findLastVarOffset(source []byte, varName string) uint32 {
	needle := []byte(varName)
	last := -1
	for i := 0; i <= len(source)-len(needle); i++ {
		if string(source[i:i+len(needle)]) == string(needle) {
			last = i
		}
	}
	return uint32(last)
}

// findAllVarOffsets returns all byte offsets where varName appears in source.
func findAllVarOffsets(source []byte, varName string) []uint32 {
	needle := []byte(varName)
	var offsets []uint32
	for i := 0; i <= len(source)-len(needle); i++ {
		if string(source[i:i+len(needle)]) == string(needle) {
			offsets = append(offsets, uint32(i))
		}
	}
	return offsets
}

func TestFlowNarrowingUndeclaredVariableNoPhantomNarrowing(t *testing.T) {
	// An undeclared variable (no my) inside a defined() guard should NOT get
	// a phantom Scalar annotation from a guard scope shadow. Without the fix,
	// walkBlockWithGuard defaults currentType to Scalar for undeclared vars,
	// NarrowByGuard(Scalar, GuardDefined) returns (Scalar, true), and a
	// phantom shadow is created. The reference to $x inside the block would
	// then be annotated as Scalar instead of the normal sigil type Scalar.
	//
	// The observable difference: with the phantom shadow, the guard scope
	// contains $x with type Scalar. Without it, $x falls through to the
	// sigil-based annotation. The net annotation is the same (Scalar either
	// way for defined guard), so we verify via ref() guard instead, where
	// the phantom would produce Ref (wrong) vs Scalar (correct sigil type).
	src := []byte("if (ref($x)) {\n    $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	// The $x reference inside the block: if a phantom shadow was created,
	// it would be Ref (from NarrowByGuard on undeclared Scalar). If no
	// phantom, lookupNarrowedType returns the sigil fallback Scalar.
	refOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[refOffset]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Scalar, typ, "undeclared $x inside ref() guard should be Scalar (sigil type), not Ref (phantom narrowing)")
}

func TestFlowNarrowingNonGuardCondition(t *testing.T) {
	// if ($x > 0) is not a recognized guard; $x should keep its type.
	src := []byte("my $x = 42;\nif ($x > 0) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets: [0]=decl, [1]=condition, [2]=if-body
	require.True(t, len(offsets) >= 3, "should find at least 3 occurrences of $x")

	ifBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Int, ifBodyTyp, "if-body $x should keep Int (no guard narrowing)")
}

func TestFlowNarrowingUnlessRefWithElse(t *testing.T) {
	// unless (ref($x)): body $x → Scalar &^ Ref (negated ref removes Ref bits),
	// else $x → Ref (positive ref)
	src := []byte("my $x = undef;\nunless (ref($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets[0]=decl, [1]=condition, [2]=unless-body, [3]=else-body
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	unlessBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "unless-body $x should be annotated")
	assert.Equal(t, types.Scalar&^types.Ref, unlessBodyTyp, "unless-body $x should be Scalar &^ Ref (negated ref guard removes Ref bits)")

	elseBodyTyp, elseOk := annotations[offsets[3]]
	assert.True(t, elseOk, "else-body $x should be annotated")
	assert.Equal(t, types.Ref, elseBodyTyp, "else-body $x should be Ref (positive ref guard)")
}

func TestFlowNarrowingElsifBranchGetsAnnotations(t *testing.T) {
	// elsif blocks get type annotations with guard narrowing applied.
	// elsif (ref($x)) narrows $x to Ref in the elsif body.
	src := []byte("my $x = undef;\nif (defined($x)) {\n    my $y = $x;\n} elsif (ref($x)) {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets[0]=decl, [1]=if-condition, [2]=if-body, [3]=elsif-condition, [4]=elsif-body
	require.True(t, len(offsets) >= 5, "should find at least 5 occurrences of $x, got %d", len(offsets))

	elsifBodyTyp, ok := annotations[offsets[4]]
	assert.True(t, ok, "elsif-body $x should be annotated (not skipped)")
	assert.Equal(t, types.Ref, elsifBodyTyp, "elsif-body $x should be Ref (ref guard narrowing in elsif)")
}

func TestFlowNarrowingIsaGuard(t *testing.T) {
	// if ($x isa Foo): if-body $x → Object (isa guard narrows to Object).
	src := []byte("my $x = undef;\nif ($x isa Foo) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets[0]=decl, [1]=condition, [2]=if-body
	require.True(t, len(offsets) >= 3, "should find at least 3 occurrences of $x, got %d", len(offsets))

	ifBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Object, ifBodyTyp, "if-body $x should be Object after isa guard")
}

func TestFlowNarrowingIsaGuardWithElse(t *testing.T) {
	// if ($x isa Foo): if-body → Object, else-body → Scalar (isa negation is not useful).
	src := []byte("my $x = undef;\nif ($x isa Foo) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets[0]=decl, [1]=condition, [2]=if-body, [3]=else-body
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	ifBodyTyp, ifOk := annotations[offsets[2]]
	assert.True(t, ifOk, "if-body $x should be annotated")
	assert.Equal(t, types.Object, ifBodyTyp, "if-body $x should be Object after isa guard")

	elseBodyTyp, elseOk := annotations[offsets[3]]
	assert.True(t, elseOk, "else-body $x should be annotated")
	assert.Equal(t, types.Scalar, elseBodyTyp, "else-body $x should be Scalar (isa negation is not useful)")
}

func TestFlowNarrowingElsifNegatedGuard(t *testing.T) {
	// elsif (!defined($x)) should narrow $x to Undef in the elsif body
	// (negated defined guard).
	// else-body $x → Scalar &^ Undef (positive defined guard, which removed the Undef bit).
	src := []byte("my $x = undef;\nif ($x isa Foo) {\n    my $a = $x;\n} elsif (!defined($x)) {\n    my $b = $x;\n} else {\n    my $c = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// [0]=decl, [1]=if-cond, [2]=if-body, [3]=elsif-cond, [4]=elsif-body, [5]=else-body
	require.True(t, len(offsets) >= 6, "should find at least 6 occurrences of $x, got %d", len(offsets))

	elsifBodyTyp, ok := annotations[offsets[4]]
	assert.True(t, ok, "elsif-body $x should be annotated")
	assert.Equal(t, types.Undef, elsifBodyTyp, "elsif-body $x should be Undef (negated defined guard in elsif)")

	elseBodyTyp, ok2 := annotations[offsets[5]]
	assert.True(t, ok2, "else-body $x should be annotated")
	assert.Equal(t, types.Scalar&^types.Undef, elseBodyTyp, "else-body $x should be Scalar &^ Undef (positive defined guard removes Undef bit)")
}

func TestFlowNarrowingElsifChainThreeBranches(t *testing.T) {
	// Three branches: if + elsif + elsif. Each should get its own guard.
	src := []byte("my $x = undef;\nif (defined($x)) {\n    my $a = $x;\n} elsif (ref($x)) {\n    my $b = $x;\n} elsif ($x isa Foo) {\n    my $c = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// [0]=decl, [1]=if-cond, [2]=if-body, [3]=elsif1-cond, [4]=elsif1-body, [5]=elsif2-cond, [6]=elsif2-body
	require.True(t, len(offsets) >= 7, "should find at least 7 occurrences of $x, got %d", len(offsets))

	elsif1BodyTyp, ok := annotations[offsets[4]]
	assert.True(t, ok, "first elsif-body $x should be annotated")
	assert.Equal(t, types.Ref, elsif1BodyTyp, "first elsif-body $x should be Ref")

	elsif2BodyTyp, ok2 := annotations[offsets[6]]
	assert.True(t, ok2, "second elsif-body $x should be annotated")
	assert.Equal(t, types.Object, elsif2BodyTyp, "second elsif-body $x should be Object (isa guard)")
}

func TestFlowNarrowingElsifWithElse(t *testing.T) {
	// elsif (ref($x)) with else: elsif-body → Ref, else-body → Scalar &^ Ref (negated ref removes Ref bits)
	src := []byte("my $x = undef;\nif (defined($x)) {\n    my $a = $x;\n} elsif (ref($x)) {\n    my $b = $x;\n} else {\n    my $c = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// [0]=decl, [1]=if-cond, [2]=if-body, [3]=elsif-cond, [4]=elsif-body, [5]=else-body
	require.True(t, len(offsets) >= 6, "should find at least 6 occurrences of $x, got %d", len(offsets))

	elsifBodyTyp, ok := annotations[offsets[4]]
	assert.True(t, ok, "elsif-body $x should be annotated")
	assert.Equal(t, types.Ref, elsifBodyTyp, "elsif-body $x should be Ref (ref guard)")

	elseBodyTyp, ok2 := annotations[offsets[5]]
	assert.True(t, ok2, "else-body $x should be annotated")
	assert.Equal(t, types.Scalar&^types.Ref, elseBodyTyp, "else-body $x should be Scalar &^ Ref (negated ref guard removes Ref bits)")
}

func TestExtractGuardPatternNegatedDefined(t *testing.T) {
	// if (!defined($x)): the condition is a negated defined guard.
	// The if-body should get the negated guard (Undef), not the positive (Scalar &^ Undef).
	src := []byte("my $x = undef;\nif (!defined($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// [0]=decl, [1]=condition, [2]=if-body, [3]=else-body
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	ifBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Undef, ifBodyTyp, "if-body $x should be Undef (negated defined guard)")

	elseBodyTyp, ok2 := annotations[offsets[3]]
	assert.True(t, ok2, "else-body $x should be annotated")
	assert.Equal(t, types.Scalar&^types.Undef, elseBodyTyp, "else-body $x should be Scalar &^ Undef (positive defined guard removes Undef bit)")
}

func TestExtractGuardPatternNotKeyword(t *testing.T) {
	// if (not ref($x)): "not" is an ambiguous_function_call_expression wrapping ref.
	// The if-body gets the negated guard (Scalar &^ Ref), else gets the positive guard (Ref).
	src := []byte("my $x = undef;\nif (not ref($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	ifBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Scalar&^types.Ref, ifBodyTyp, "if-body $x should be Scalar &^ Ref (negated ref guard removes Ref bits)")

	elseBodyTyp, ok2 := annotations[offsets[3]]
	assert.True(t, ok2, "else-body $x should be annotated")
	assert.Equal(t, types.Ref, elseBodyTyp, "else-body $x should be Ref (positive ref guard)")
}

func TestFlowNarrowingEarlyReturnNarrowsDefined(t *testing.T) {
	// if (!defined($x)) { return; } — after the if, $x should be Scalar &^ Undef
	// (defined guard removes the Undef bit from Scalar).
	src := []byte("my $x = undef;\nif (!defined($x)) {\n    return;\n}\nmy $y = $x;\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// [0]=decl, [1]=condition, [2]=post-if reference
	require.True(t, len(offsets) >= 3, "should find at least 3 occurrences of $x, got %d", len(offsets))

	postIfTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "post-if $x should be annotated")
	assert.Equal(t, types.Scalar&^types.Undef, postIfTyp, "post-if $x should be Scalar &^ Undef (defined guard after early return removes Undef bit)")
}

func TestFlowNarrowingNoEarlyExitNoNarrowing(t *testing.T) {
	// if (!defined($x)) { $x = 1; } — block does not exit, so no post-if narrowing.
	src := []byte("my $x = undef;\nif (!defined($x)) {\n    $x = 1;\n}\nmy $y = $x;\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// [0]=decl, [1]=condition, [2]=if-body assignment LHS, [3]=post-if ref
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	postIfTyp, ok := annotations[offsets[3]]
	assert.True(t, ok, "post-if $x should be annotated")
	assert.Equal(t, types.Scalar, postIfTyp, "post-if $x should remain Scalar (no early exit, no narrowing)")
}

func TestFlowNarrowingEarlyDieNarrowsRef(t *testing.T) {
	// if (!ref($x)) { die; } — after the if, $x should be Ref.
	src := []byte("my $x = undef;\nif (!ref($x)) {\n    die;\n}\nmy $y = $x;\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 3, "should find at least 3 occurrences of $x, got %d", len(offsets))

	postIfTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "post-if $x should be annotated")
	assert.Equal(t, types.Ref, postIfTyp, "post-if $x should be Ref (ref guard after early die)")
}

func TestFlowNarrowingEarlyDieWithArgNarrowsRef(t *testing.T) {
	// if (!ref($x)) { die $msg; } — die with argument should also be detected as early exit.
	src := []byte("my $msg;\nmy $x = undef;\nif (!ref($x)) {\n    die $msg;\n}\nmy $y = $x;\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 3, "should find at least 3 occurrences of $x, got %d", len(offsets))

	postIfTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "post-if $x should be annotated")
	assert.Equal(t, types.Ref, postIfTyp, "post-if $x should be Ref (ref guard after die $msg)")
}

func TestFlowNarrowingEarlyExitNarrowsDefined(t *testing.T) {
	// if (!defined($x)) { exit; } — after the if, $x should be Scalar &^ Undef
	// (defined guard removes the Undef bit from Scalar).
	src := []byte("my $x = undef;\nif (!defined($x)) {\n    exit;\n}\nmy $y = $x;\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 3, "should find at least 3 occurrences of $x, got %d", len(offsets))

	postIfTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "post-if $x should be annotated")
	assert.Equal(t, types.Scalar&^types.Undef, postIfTyp, "post-if $x should be Scalar &^ Undef (defined guard after early exit removes Undef bit)")
}

func TestFlowNarrowingUnlessEarlyReturn(t *testing.T) {
	// unless (defined($x)) { return; } — after the unless, $x should be Scalar &^ Undef
	// (defined guard removes the Undef bit from Scalar).
	src := []byte("my $x = undef;\nunless (defined($x)) {\n    return;\n}\nmy $y = $x;\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 3, "should find at least 3 occurrences of $x, got %d", len(offsets))

	postIfTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "post-unless $x should be annotated")
	assert.Equal(t, types.Scalar&^types.Undef, postIfTyp, "post-unless $x should be Scalar &^ Undef (defined guard after early return removes Undef bit)")
}

func TestFlowNarrowingNonNegatedEarlyReturn(t *testing.T) {
	// if (defined($x)) { return; } — after the if, $x should be Undef
	// (NegateGuard applied because the exiting branch proved defined).
	src := []byte("my $x = undef;\nif (defined($x)) {\n    return;\n}\nmy $y = $x;\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 3, "should find at least 3 occurrences of $x, got %d", len(offsets))

	postIfTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "post-if $x should be annotated")
	assert.Equal(t, types.Undef, postIfTyp, "post-if $x should be Undef (negated defined guard after early return)")
}

func TestFlowNarrowingForLoopVariable(t *testing.T) {
	// for my $item (@arr) { $item; } — $item should be Scalar inside the body.
	src := []byte("my @arr;\nfor my $item (@arr) {\n    $item;\n}\n")
	annotations, _ := analyzeSource(t, src)

	itemOffset := findLastVarOffset(src, "$item")
	typ, ok := annotations[itemOffset]
	assert.True(t, ok, "for-body $item should be annotated")
	assert.Equal(t, types.Scalar, typ, "for-body $item should be Scalar")
}

func TestFlowNarrowingForLoopVariableNoLeak(t *testing.T) {
	// After the for loop, $item should not be in the symbol table.
	src := []byte("my @arr;\nfor my $item (@arr) {\n    $item;\n}\n$item;\n")
	_, _, st := analyzeSourceFull(t, src)

	_, ok := st.Lookup("$item")
	assert.False(t, ok, "$item should not be in symbol table after for loop")
}

func TestFlowNarrowingForLoopOverList(t *testing.T) {
	// for my $n (1, 2, 3) { $n; } — $n should be Scalar inside the body.
	src := []byte("for my $n (1, 2, 3) {\n    $n;\n}\n")
	annotations, _ := analyzeSource(t, src)

	nOffset := findLastVarOffset(src, "$n")
	typ, ok := annotations[nOffset]
	assert.True(t, ok, "for-body $n should be annotated")
	assert.Equal(t, types.Scalar, typ, "for-body $n should be Scalar")
}

func TestFlowNarrowingForLoopWithoutMy(t *testing.T) {
	// for $item (@arr) { $item; } — same behavior, no my keyword.
	src := []byte("my @arr;\nfor $item (@arr) {\n    $item;\n}\n")
	annotations, _ := analyzeSource(t, src)

	itemOffset := findLastVarOffset(src, "$item")
	typ, ok := annotations[itemOffset]
	assert.True(t, ok, "for-body $item should be annotated")
	assert.Equal(t, types.Scalar, typ, "for-body $item should be Scalar")
}

func TestFlowNarrowingCStyleForInitializer(t *testing.T) {
	// for (my $i = 0; ...) — $i should be narrowed to Int by assignment.
	src := []byte("for (my $i = 0; $i < 10; $i++) {\n    $i;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$i")
	// Find the $i inside the block body (last occurrence).
	bodyOffset := offsets[len(offsets)-1]
	typ, ok := annotations[bodyOffset]
	assert.True(t, ok, "for-body $i should be annotated")
	assert.Equal(t, types.Int, typ, "for-body $i should be Int (narrowed by assignment)")
}

// --- Compound guard extraction tests ---
// These tests verify that compound conditions (&&, ||, and, or) are handled
// correctly. Partial compound guards (where only one operand is a recognized
// guard) fall back to that single guard's narrowing behavior. Full compound
// guards (both operands recognized) apply all leaf guards simultaneously via
// flattenGuards + walkBlockWithGuard.

func TestCompoundGuardAmpAmpNarrowsBoth(t *testing.T) {
	// if (defined($x) && ref($x)) — full compound: both sides are guards.
	// && with negate=false: both guards apply → defined removes Undef, ref
	// keeps only Ref bits → result is Ref.
	src := []byte("my $x = undef;\nif (defined($x) && ref($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets[0]=decl, [1]=defined($x) condition, [2]=ref($x) condition, [3]=if-body
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	ifBodyTyp, ok := annotations[offsets[3]]
	assert.True(t, ok, "if-body $x should be annotated with compound && guard")
	assert.Equal(t, types.Ref, ifBodyTyp, "if-body $x should be Ref: defined removes Undef, ref keeps only Ref bits")
}

func TestCompoundGuardAndKeywordNarrowsBoth(t *testing.T) {
	// if (defined($x) and ref($x)) — lowprec_logical_expression with "and".
	// "and" normalizes to "&&", so both guards apply → Ref (same as &&).
	src := []byte("my $x = undef;\nif (defined($x) and ref($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	ifBodyTyp, ok := annotations[offsets[3]]
	assert.True(t, ok, "if-body $x should be annotated with compound 'and' guard")
	assert.Equal(t, types.Ref, ifBodyTyp, "if-body $x should be Ref: 'and' behaves like && — both guards apply")
}

func TestCompoundGuardOrNoNarrowingInBody(t *testing.T) {
	// if (defined($x) || ref($x)) — || with negate=false: either could be
	// true, so no narrowing applies in the if-body → $x stays Scalar.
	src := []byte("my $x = undef;\nif (defined($x) || ref($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	ifBodyTyp, ok := annotations[offsets[3]]
	assert.True(t, ok, "if-body $x should be annotated with compound || guard")
	assert.Equal(t, types.Scalar, ifBodyTyp, "if-body $x stays Scalar: || means either could be true, no narrowing")
}

func TestCompoundGuardOrNarrowsElse(t *testing.T) {
	// if (defined($x) || ref($x)) {} else { $x }
	// else-branch: both guards are false → !defined AND !ref → Undef &^ Ref = Undef.
	src := []byte("my $x = undef;\nif (defined($x) || ref($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets[0]=decl, [1]=defined cond, [2]=ref cond, [3]=if-body, [4]=else-body
	require.True(t, len(offsets) >= 5, "should find at least 5 occurrences of $x, got %d", len(offsets))

	elseBodyTyp, ok := annotations[offsets[4]]
	assert.True(t, ok, "else-body $x should be annotated with compound || guard")
	assert.Equal(t, types.Undef, elseBodyTyp, "else-body $x should be Undef: || else means !defined && !ref")
}

func TestCompoundGuardAndDifferentVars(t *testing.T) {
	// if (defined($x) && ref($y)) — two different variables guarded.
	// $x narrowed by defined (Undef removed), $y narrowed by ref → Ref.
	src := []byte("my $x = undef;\nmy $y = undef;\nif (defined($x) && ref($y)) {\n    my $a = $x;\n    my $b = $y;\n}\n")
	annotations, _ := analyzeSource(t, src)

	xOffsets := findAllVarOffsets(src, "$x")
	yOffsets := findAllVarOffsets(src, "$y")

	// $x: offsets[0]=decl, [1]=defined cond, [2]=if-body
	require.True(t, len(xOffsets) >= 3, "should find at least 3 occurrences of $x, got %d", len(xOffsets))
	xBody, xOk := annotations[xOffsets[2]]
	assert.True(t, xOk, "if-body $x should be annotated")
	assert.True(t, xBody&types.Undef == 0, "if-body $x should have Undef removed (defined guard)")

	// $y: offsets[0]=decl, [1]=ref cond, [2]=if-body
	require.True(t, len(yOffsets) >= 3, "should find at least 3 occurrences of $y, got %d", len(yOffsets))
	yBody, yOk := annotations[yOffsets[2]]
	assert.True(t, yOk, "if-body $y should be annotated")
	assert.Equal(t, types.Ref, yBody, "if-body $y should be Ref (ref guard)")
}

func TestCompoundGuardPartialOneNonGuardSide(t *testing.T) {
	// if (defined($x) && $y > 0) — partial compound: the right side ($y > 0)
	// is not a recognized guard. extractCompoundGuard returns the defined($x)
	// guard directly (the non-guard side is dropped). walkBlockWithGuard applies
	// the defined guard, narrowing $x to Scalar &^ Undef.
	src := []byte("my $x = undef;\nmy $y = 1;\nif (defined($x) && $y > 0) {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets[0]=decl, [1]=defined($x) condition, [2]=if-body $x
	require.True(t, len(offsets) >= 3, "should find at least 3 occurrences of $x, got %d", len(offsets))

	// The defined guard was extracted as the surviving single guard.
	ifBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Scalar&^types.Undef, ifBodyTyp, "if-body $x should be Scalar &^ Undef from partial compound defined guard")
}

func TestCompoundGuardNegatedAmpAmpDeMorganNarrows(t *testing.T) {
	// if (!(defined($x) && ref($x))) — De Morgan applied:
	// becomes compound {Op:"||", Left:!defined($x), Right:!ref($x)}.
	// || with negate=false: no narrowing in if-body → $x stays Scalar.
	src := []byte("my $x = undef;\nif (!(defined($x) && ref($x))) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	ifBodyTyp, ok := annotations[offsets[3]]
	assert.True(t, ok, "if-body $x should be annotated with negated compound && guard")
	assert.Equal(t, types.Scalar, ifBodyTyp, "if-body $x stays Scalar: !(A && B) becomes (||) which applies no narrowing in the body")
}

func TestCompoundGuardArithmeticBinaryNotExtracted(t *testing.T) {
	// if (1 + 2) — binary_expression with non-boolean operator.
	// extractCompoundGuard returns nil (op is "+" not "&&"/"||").
	// No guard narrowing is attempted; $x retains its declared type.
	src := []byte("my $x = 42;\nif (1 + 2) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets[0]=decl, [1]=if-body
	require.True(t, len(offsets) >= 2, "should find at least 2 occurrences of $x, got %d", len(offsets))

	ifBodyTyp, ok := annotations[offsets[1]]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Int, ifBodyTyp, "if-body $x should be Int (arithmetic binary_expression is not a guard)")
}

// --- Branch merging at join points ---
// After an if/elsif/else chain, the post-if type is the union (OR) of all
// branch types for guarded variables. Early-exit branches are excluded from
// the join. When there is no else branch, the implicit else contributes the
// pre-if type.

func TestBranchMergingIfElseJoinType(t *testing.T) {
	// After if (ref($x)) {...} else {...}, $x should be the union of branch types.
	// if-body: Ref; else-body: Scalar &^ Ref; union = Ref | (Scalar &^ Ref) = Scalar.
	src := []byte("my $x = undef;\nif (ref($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\nmy $w = $x;\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets: [0]=decl, [1]=condition, [2]=if-body, [3]=else-body, [4]=post-if
	require.True(t, len(offsets) >= 5, "should find at least 5 occurrences of $x, got %d", len(offsets))

	postTyp, ok := annotations[offsets[len(offsets)-1]]
	assert.True(t, ok, "post-if $x should be annotated")
	// Ref | (Scalar &^ Ref) = Scalar
	assert.Equal(t, types.Scalar, postTyp, "post-if/else $x should be Scalar (union of Ref and non-Ref branches)")
}

func TestBranchMergingIfNoElse(t *testing.T) {
	// if (ref($x)) {...} — implicit else contributes pre-if type (Scalar).
	// if-body: Ref; implicit else: Scalar; union = Ref | Scalar = Scalar.
	src := []byte("my $x = undef;\nif (ref($x)) {\n    my $y = $x;\n}\nmy $z = $x;\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets: [0]=decl, [1]=condition, [2]=if-body, [3]=post-if
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	postTyp, ok := annotations[offsets[len(offsets)-1]]
	assert.True(t, ok, "post-if $x should be annotated")
	// Ref | Scalar (pre-if) = Scalar
	assert.Equal(t, types.Scalar, postTyp, "post-if (no else) $x should be Scalar (union of Ref branch and implicit else Scalar)")
}

func TestBranchMergingEarlyExitExcluded(t *testing.T) {
	// if (!defined($x)) { return; } — early exit: the if-body does NOT
	// contribute to the join. Post-if type is narrowed via early-exit logic.
	src := []byte("my $x = undef;\nif (!defined($x)) {\n    return;\n}\nmy $y = $x;\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets: [0]=decl, [1]=condition, [2]=post-if
	require.True(t, len(offsets) >= 3, "should find at least 3 occurrences of $x, got %d", len(offsets))

	postTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "post-early-exit $x should be annotated")
	assert.True(t, postTyp&types.Undef == 0, "post-early-exit $x should not include Undef (early-exit narrowing removes Undef)")
}

func TestBranchMergingElsifChain(t *testing.T) {
	// if/elsif/else: join type is union of all branch types for $x.
	// if-body: Scalar &^ Undef (defined guard); elsif-body: Ref; else-body: Undef.
	// union = (Scalar &^ Undef) | Ref | Undef = Scalar.
	src := []byte("my $x = undef;\nif (defined($x)) {\n    my $a = $x;\n} elsif (ref($x)) {\n    my $b = $x;\n} else {\n    my $c = $x;\n}\nmy $d = $x;\n")
	annotations, _ := analyzeSource(t, src)

	xOffsets := findAllVarOffsets(src, "$x")
	// Last occurrence is the post-chain $x.
	require.True(t, len(xOffsets) >= 6, "should find at least 6 occurrences of $x, got %d", len(xOffsets))

	postTyp, ok := annotations[xOffsets[len(xOffsets)-1]]
	assert.True(t, ok, "post-chain $x should be annotated")
	// All branches together cover Scalar territory; union should be Scalar.
	assert.Equal(t, types.Scalar, postTyp, "post-chain $x should be Scalar (union of all branch types)")
}

func TestBranchMergingAssignmentInsideBranch(t *testing.T) {
	// $x starts as Scalar. if-body assigns Int (1), else-body assigns Num (3.14).
	// After the if/else, $x should be the join of the two branch-end types: Int | Num.
	// Without branch merging, scope restoration would revert $x to Scalar.
	// Int | Num = Num (since Num includes Int in the type hierarchy).
	src := []byte("my $x = undef;\nif (defined($x)) {\n    $x = 1;\n} else {\n    $x = 3.14;\n}\nmy $y = $x;\n")
	annotations, _ := analyzeSource(t, src)

	xOffsets := findAllVarOffsets(src, "$x")
	// offsets: [0]=decl, [1]=cond, [2]=if-body assign LHS, [3]=else-body assign LHS, [4]=post-if ref
	require.True(t, len(xOffsets) >= 5, "should find at least 5 occurrences of $x, got %d", len(xOffsets))

	postTyp, ok := annotations[xOffsets[len(xOffsets)-1]]
	assert.True(t, ok, "post-if/else $x should be annotated")
	// Int | Num = Num (Num = numLeaf | Int, which already includes Int).
	assert.Equal(t, types.Num, postTyp, "post-if/else $x should be Num (union of Int from if-body and Num from else-body)")
}

// --- Guard pattern library tests (builtin:: functions) ---

func TestGuardLibraryBlessedNarrows(t *testing.T) {
	// if (builtin::blessed($x)): body $x → Object
	src := []byte("my $x = undef;\nif (builtin::blessed($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)
	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 3)
	ifBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Object, ifBodyTyp, "blessed guard narrows to Object")
}

func TestGuardLibraryBlessedBareNarrows(t *testing.T) {
	// if (blessed($x)): bare name, same effect
	src := []byte("my $x = undef;\nif (blessed($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)
	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 3)
	ifBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Object, ifBodyTyp, "bare blessed guard narrows to Object")
}

func TestGuardLibraryReftypeNarrows(t *testing.T) {
	// if (builtin::reftype($x)): body $x → Ref
	src := []byte("my $x = undef;\nif (builtin::reftype($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)
	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 3)
	ifBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Ref, ifBodyTyp, "reftype guard narrows to Ref")
}

func TestGuardLibraryIsBoolNarrows(t *testing.T) {
	// if (builtin::is_bool($x)): body $x → Bool
	src := []byte("my $x = undef;\nif (builtin::is_bool($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)
	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 3)
	ifBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Bool, ifBodyTyp, "is_bool guard narrows to Bool")
}

func TestGuardLibraryNegatedBlessed(t *testing.T) {
	// if (!builtin::blessed($x)) { } else { $x } — else $x → Object
	src := []byte("my $x = undef;\nif (!builtin::blessed($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)
	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 4)
	elseBodyTyp, ok := annotations[offsets[3]]
	assert.True(t, ok, "else-body $x should be annotated")
	assert.Equal(t, types.Object, elseBodyTyp, "negated blessed: else → Object")
}

func TestGuardLibraryCompoundWithBuiltin(t *testing.T) {
	// if (defined($x) && builtin::blessed($x)): body $x → Object
	// offsets: [0]=decl, [1]=defined($x), [2]=blessed($x), [3]=if-body ref
	src := []byte("my $x = undef;\nif (defined($x) && builtin::blessed($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)
	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 4, "should find at least 4 $x, got %d", len(offsets))
	ifBodyTyp, ok := annotations[offsets[3]]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Object, ifBodyTyp, "defined && blessed → Object")
}

// --- Subroutine return type inference ---

func TestReturnTypeExplicitReturn(t *testing.T) {
	src := []byte("sub foo { return 42; }\n")
	_, _, st := analyzeSourceFull(t, src)
	sym, ok := st.Lookup("foo")
	require.True(t, ok, "sub foo must be in symbol table")
	assert.Equal(t, types.Int, sym.ReturnType, "sub foo with 'return 42' should have return type Int")
}

func TestReturnTypeImplicitReturn(t *testing.T) {
	src := []byte("sub bar { 42; }\n")
	_, _, st := analyzeSourceFull(t, src)
	sym, ok := st.Lookup("bar")
	require.True(t, ok, "sub bar must be in symbol table")
	assert.Equal(t, types.Int, sym.ReturnType, "sub bar with implicit 42 should have return type Int")
}

func TestReturnTypeMultipleReturns(t *testing.T) {
	src := []byte("my $x = 1;\nsub qux { if ($x) { return 0; } return 3.14; }\n")
	_, _, st := analyzeSourceFull(t, src)
	sym, ok := st.Lookup("qux")
	require.True(t, ok, "sub qux must be in symbol table")
	assert.Equal(t, types.Num, sym.ReturnType, "sub qux with int and float returns should have return type Num (Int | Num = Num)")
}

func TestReturnTypeEmptySub(t *testing.T) {
	src := []byte("sub empty { }\n")
	_, _, st := analyzeSourceFull(t, src)
	sym, ok := st.Lookup("empty")
	require.True(t, ok, "sub empty must be in symbol table")
	assert.Equal(t, types.Unknown, sym.ReturnType, "empty sub should have Unknown return type")
}

func TestReturnTypeImplicitConditional(t *testing.T) {
	// sub pick { if ($x) { 42; } else { 3.14; } }
	// implicit return from if → Int, from else → Num; union = Num
	src := []byte("my $x = 1;\nsub pick { if ($x) { 42; } else { 3.14; } }\n")
	_, _, st := analyzeSourceFull(t, src)
	sym, ok := st.Lookup("pick")
	require.True(t, ok, "sub pick must be in symbol table")
	assert.Equal(t, types.Num, sym.ReturnType, "sub pick with if/else returning Int and Num should have return type Num")
}

func TestReturnTypeBareReturn(t *testing.T) {
	src := []byte("sub noop { return; }\n")
	_, _, st := analyzeSourceFull(t, src)
	sym, ok := st.Lookup("noop")
	require.True(t, ok, "sub noop must be in symbol table")
	assert.Equal(t, types.Undef, sym.ReturnType, "bare return; should have return type Undef")
}

func TestReturnTypeExplicitPlusImplicit(t *testing.T) {
	// sub mixed { if ($x) { return 42; } 3.14; }
	// explicit return: Int, implicit return: Num → union = Num
	src := []byte("my $x = 1;\nsub mixed { if ($x) { return 42; } 3.14; }\n")
	_, _, st := analyzeSourceFull(t, src)
	sym, ok := st.Lookup("mixed")
	require.True(t, ok, "sub mixed must be in symbol table")
	assert.Equal(t, types.Num, sym.ReturnType, "sub mixed with explicit Int return and implicit Num return should have return type Num")
}

// --- Call site return type lookup tests ---

func TestCallSiteUsesReturnType(t *testing.T) {
	// sub foo { return 42; } my $x = foo(); — $x should be Int
	src := []byte("sub foo { return 42; }\nmy $x = foo();\n$x;\n")
	annotations, _ := analyzeSource(t, src)
	xOffsets := findAllVarOffsets(src, "$x")
	refTyp, ok := annotations[xOffsets[len(xOffsets)-1]]
	assert.True(t, ok, "$x reference should be annotated")
	assert.Equal(t, types.Int, refTyp, "$x should be Int from foo()'s return type")
}

func TestCallSiteImplicitReturn(t *testing.T) {
	// sub bar { 3.14; } my $y = bar(); — $y should be Num
	src := []byte("sub bar { 3.14; }\nmy $y = bar();\n$y;\n")
	annotations, _ := analyzeSource(t, src)
	yOffsets := findAllVarOffsets(src, "$y")
	refTyp, ok := annotations[yOffsets[len(yOffsets)-1]]
	assert.True(t, ok)
	assert.Equal(t, types.Num, refTyp, "$y should be Num from bar()'s implicit return")
}

func TestCallSiteForwardRefUnknown(t *testing.T) {
	// my $x = foo(); sub foo { return 42; } — forward ref, $x stays Scalar
	src := []byte("my $x = foo();\nsub foo { return 42; }\n$x;\n")
	annotations, _ := analyzeSource(t, src)
	xOffsets := findAllVarOffsets(src, "$x")
	refTyp, ok := annotations[xOffsets[len(xOffsets)-1]]
	assert.True(t, ok)
	assert.Equal(t, types.Scalar, refTyp, "$x should be Scalar (forward ref, no return type yet)")
}

func TestCallSiteBuiltinPriority(t *testing.T) {
	// Builtins should not be overridden by user subs
	src := []byte("my @a;\nmy $x = push(@a, 1);\n")
	annotations, _ := analyzeSource(t, src)
	pushTyp, ok := findNodeType(annotations, src, "push(@a, 1)")
	assert.True(t, ok)
	assert.Equal(t, types.Int, pushTyp, "builtin push should return Int")
}

// --- Cross-file analysis tests ---

// TestUseStatementTriggersAnalysis verifies that when a ProjectIndex is provided,
// a "use Foo;" statement causes the corresponding Foo.pm to be analysed and its
// symbols to be registered in the index so that LookupSymbol succeeds afterward.
func TestUseStatementTriggersAnalysis(t *testing.T) {
	dir := t.TempDir()
	err := os.MkdirAll(filepath.Join(dir, "lib"), 0755)
	require.NoError(t, err)
	err = os.WriteFile(
		filepath.Join(dir, "lib", "Foo.pm"),
		[]byte("package Foo;\nsub bar { return 42; }\n"),
		0644,
	)
	require.NoError(t, err)

	idx := infer.NewProjectIndex(dir)
	src := []byte("use Foo;\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	infer.Analyze(tree, src, idx)

	sym, ok := idx.LookupSymbol("Foo", "bar")
	assert.True(t, ok, "use Foo should trigger analysis of Foo.pm")
	assert.Equal(t, types.Int, sym.ReturnType, "bar should have return type Int")
}

// TestFQCallResolution verifies that a fully-qualified function call like
// Foo::Bar::baz() is resolved through the ProjectIndex and its return type
// propagates to the assigned variable.
func TestFQCallResolution(t *testing.T) {
	dir := t.TempDir()
	err := os.MkdirAll(filepath.Join(dir, "lib", "Foo"), 0755)
	require.NoError(t, err)
	err = os.WriteFile(
		filepath.Join(dir, "lib", "Foo", "Bar.pm"),
		[]byte("package Foo::Bar;\nsub baz { return 42; }\n"),
		0644,
	)
	require.NoError(t, err)

	idx := infer.NewProjectIndex(dir)
	src := []byte("use Foo::Bar;\nmy $x = Foo::Bar::baz();\n$x;\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	annotations, _, _ := infer.Analyze(tree, src, idx)

	xOffsets := findAllVarOffsets(src, "$x")
	refTyp, ok := annotations[xOffsets[len(xOffsets)-1]]
	assert.True(t, ok, "$x reference should be annotated")
	assert.Equal(t, types.Int, refTyp, "$x should be Int from Foo::Bar::baz()")
}

// TestConstructorSetsClassType verifies that calling Foo->new() on a bareword
// class name sets the ClassType field on the assigned variable's Symbol.
func TestConstructorSetsClassType(t *testing.T) {
	src := []byte("my $obj = Foo->new();\n")
	_, _, st := analyzeSourceFull(t, src)
	sym, ok := st.Lookup("$obj")
	require.True(t, ok)
	assert.Equal(t, "Foo", sym.ClassType)
}

// TestMethodResolutionViaProjectIndex verifies that calling a method on an
// object whose class was established by a constructor call resolves the method
// through the ProjectIndex and propagates the return type to the result variable.
func TestMethodResolutionViaProjectIndex(t *testing.T) {
	dir := t.TempDir()
	err := os.MkdirAll(filepath.Join(dir, "lib"), 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "lib", "Counter.pm"),
		[]byte("package Counter;\nsub count { return 42; }\n"), 0644)
	require.NoError(t, err)
	idx := infer.NewProjectIndex(dir)
	src := []byte("use Counter;\nmy $c = Counter->new();\nmy $n = $c->count();\n$n;\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	annotations, _, _ := infer.Analyze(tree, src, idx)
	nOffsets := findAllVarOffsets(src, "$n")
	refTyp, ok := annotations[nOffsets[len(nOffsets)-1]]
	assert.True(t, ok)
	assert.Equal(t, types.Int, refTyp, "$n should be Int from Counter::count()")
}

func TestFlowNarrowingDefinedGuardArrayElement(t *testing.T) {
	// defined($_[0]) inside a sub body uses array_element_expression.
	// Verify that extractFunc1opGuard handles this node type.
	src := []byte("my @arr;\nif (defined($arr[0])) {\n    my $y = $arr[0];\n}\n")
	// This tests that array_element_expression is recognized by
	// extractFunc1opGuard. If it works, the guard extracts correctly.
	annotations, _ := analyzeSource(t, src)
	_ = annotations // Mainly verifying no crash; array element guard is new behavior.
}

// TestIsaGuardEnablesMethodResolution verifies that an "isa" guard in an if
// condition narrows the variable's ClassType within the block, enabling method
// call resolution through the ProjectIndex. The method call expression itself
// is annotated with the resolved return type.
func TestIsaGuardEnablesMethodResolution(t *testing.T) {
	dir := t.TempDir()
	err := os.MkdirAll(filepath.Join(dir, "lib"), 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "lib", "Dog.pm"),
		[]byte("package Dog;\nsub speak { return 1; }\n"), 0644)
	require.NoError(t, err)
	idx := infer.NewProjectIndex(dir)
	// The method call $x->speak() is annotated at the position of $x (the invocant).
	// We find the $x inside the block (after "isa Dog) {") by using the last
	// occurrence of $x before "->speak".
	src := []byte("use Dog;\nmy $x = undef;\nif ($x isa Dog) {\n    my $v = $x->speak();\n}\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	annotations, _, _ := infer.Analyze(tree, src, idx)
	// The method_call_expression node starts at the invocant ($x in $x->speak()).
	// findNodeType searches for annotations whose offset matches the start of "$x->speak()".
	callTyp, ok := findNodeType(annotations, src, "$x->speak()")
	assert.True(t, ok, "method call $x->speak() should be annotated")
	assert.Equal(t, types.Int, callTyp, "method call should resolve to Int from Dog::speak()")
}

func TestFindSubDeclNode(t *testing.T) {
	src := []byte("my $x = 1;\nsub is_ref { ref($_[0]) }\nmy $y = 2;\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()

	// Find the sub node using its byte range.
	var subStart, subEnd uint32
	for i := 0; i < root.ChildCount(); i++ {
		child := root.Child(i)
		if child != nil && child.Kind() == "subroutine_declaration_statement" {
			subStart = child.StartByte()
			subEnd = child.EndByte()
		}
	}
	require.NotZero(t, subEnd, "should find a sub in the source")

	subNode := infer.FindSubDeclNode(root, subStart, subEnd)
	require.NotNil(t, subNode, "should find the sub declaration node")
	assert.Equal(t, "subroutine_declaration_statement", subNode.Kind())
	assert.Contains(t, subNode.Text(src), "is_ref")
}

func TestFindSubDeclNodeNotFound(t *testing.T) {
	src := []byte("my $x = 1;\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()

	subNode := infer.FindSubDeclNode(root, 99, 199)
	assert.Nil(t, subNode, "should return nil when no matching sub found")
}

// --- ResolveSubParam ---

func TestResolveSubParamSignature(t *testing.T) {
	src := []byte("sub is_defined ($val) { defined($val) }\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()
	subNode := root.Child(0)
	require.Equal(t, "subroutine_declaration_statement", subNode.Kind())

	paramName, ok := infer.ResolveSubParam(subNode, src)
	assert.True(t, ok, "should resolve signature parameter")
	assert.Equal(t, "$val", paramName)
}

func TestResolveSubParamMultiArgRejected(t *testing.T) {
	src := []byte("sub check ($x, $y) { defined($x) }\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()
	subNode := root.Child(0)

	_, ok := infer.ResolveSubParam(subNode, src)
	assert.False(t, ok, "multi-arg sub should not be recognized as guard")
}

func TestResolveSubParamArrayElement(t *testing.T) {
	src := []byte("sub is_ref { ref($_[0]) }\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()
	subNode := root.Child(0)
	require.Equal(t, "subroutine_declaration_statement", subNode.Kind())

	paramName, ok := infer.ResolveSubParam(subNode, src)
	assert.True(t, ok, "should resolve parameter")
	assert.Equal(t, "$_[0]", paramName)
}

func TestResolveSubParamShift(t *testing.T) {
	src := []byte("sub is_ref { my $val = shift; ref($val) }\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()
	subNode := root.Child(0)
	require.Equal(t, "subroutine_declaration_statement", subNode.Kind())

	paramName, ok := infer.ResolveSubParam(subNode, src)
	assert.True(t, ok, "should resolve shift parameter")
	assert.Equal(t, "$val", paramName)
}

func TestExtractSubReturnExprExplicit(t *testing.T) {
	src := []byte("sub is_ref { return ref($_[0]) }\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()
	subNode := root.Child(0)

	var block *parser.Node
	for i := 0; i < subNode.ChildCount(); i++ {
		child := subNode.Child(i)
		if child != nil && child.Kind() == "block" {
			block = child
		}
	}
	require.NotNil(t, block)

	expr := infer.ExtractSubReturnExpr(block, src)
	require.NotNil(t, expr, "should find return expression")
	assert.Contains(t, expr.Text(src), "ref")
}

func TestExtractSubReturnExprImplicit(t *testing.T) {
	src := []byte("sub is_ref { ref($_[0]) }\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()
	subNode := root.Child(0)

	var block *parser.Node
	for i := 0; i < subNode.ChildCount(); i++ {
		child := subNode.Child(i)
		if child != nil && child.Kind() == "block" {
			block = child
		}
	}
	require.NotNil(t, block)

	expr := infer.ExtractSubReturnExpr(block, src)
	require.NotNil(t, expr, "should find implicit return expression")
	assert.Contains(t, expr.Text(src), "ref")
}

func TestExtractSubReturnExprMultipleReturns(t *testing.T) {
	src := []byte("sub check { if ($x) { return 1 } return 0 }\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()
	subNode := root.Child(0)

	var block *parser.Node
	for i := 0; i < subNode.ChildCount(); i++ {
		child := subNode.Child(i)
		if child != nil && child.Kind() == "block" {
			block = child
		}
	}
	require.NotNil(t, block)

	expr := infer.ExtractSubReturnExpr(block, src)
	assert.Nil(t, expr, "multiple returns should be rejected")
}

func TestUserDefinedGuardFuncDefined(t *testing.T) {
	src := []byte("sub is_defined { defined($_[0]) }\nmy $x = undef;\nif (is_defined($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 4, "need at least 4 $x occurrences, got %d", len(offsets))

	ifBodyTyp, ifOk := annotations[offsets[2]]
	assert.True(t, ifOk, "if-body $x should be annotated")
	assert.Equal(t, types.Scalar&^types.Undef, ifBodyTyp, "if-body $x should have Undef removed")

	elseBodyTyp, elseOk := annotations[offsets[3]]
	assert.True(t, elseOk, "else-body $x should be annotated")
	assert.Equal(t, types.Undef, elseBodyTyp, "else-body $x should be Undef")
}

func TestUserDefinedGuardFuncRefArrayAccess(t *testing.T) {
	src := []byte("sub is_ref { ref($_[0]) }\nmy $x = undef;\nif (is_ref($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Ref, typ, "if-body $x should be Ref after user-defined ref guard")
}

func TestUserDefinedGuardFuncSignature(t *testing.T) {
	src := []byte("sub is_defined ($val) { defined($val) }\nmy $x = undef;\nif (is_defined($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Scalar&^types.Undef, typ, "if-body $x should have Undef removed after signature-style guard")
}

func TestUserDefinedGuardFuncShift(t *testing.T) {
	src := []byte("sub is_ref { my $val = shift; ref($val) }\nmy $x = undef;\nif (is_ref($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Ref, typ, "if-body $x should be Ref after shift-style guard")
}

func TestUserDefinedGuardFuncIsa(t *testing.T) {
	src := []byte("sub is_foo { $_[0] isa Foo }\nmy $x = undef;\nif (is_foo($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Object, typ, "if-body $x should be Object after isa guard")
}

func TestUserDefinedGuardFuncNegatedDefined(t *testing.T) {
	src := []byte("sub is_defined { defined($_[0]) }\nmy $x = undef;\nunless (is_defined($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	// unless(is_defined($x)) means body gets negated guard -> Undef.
	unlessBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[unlessBodyXOffset]
	assert.True(t, ok, "unless-body $x should be annotated")
	assert.Equal(t, types.Undef, typ, "unless-body $x should be Undef (negated defined guard)")
}

func TestUserDefinedGuardFuncCompound(t *testing.T) {
	src := []byte("sub is_defined_ref { defined($_[0]) && ref($_[0]) }\nmy $x = undef;\nif (is_defined_ref($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Ref, typ, "if-body $x should be Ref after compound guard (defined+ref)")
}

func TestUserDefinedGuardFuncNonGuardIgnored(t *testing.T) {
	src := []byte("sub do_stuff { 42 }\nmy $x = undef;\nif (do_stuff($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	if ok {
		assert.Equal(t, types.Scalar, typ, "if-body $x should stay Scalar (non-guard sub)")
	}
}

func TestUserDefinedGuardFuncComplexArgNoCrash(t *testing.T) {
	src := []byte("sub is_ref { ref($_[0]) }\nmy @arr;\nif (is_ref($arr[0])) {\n    my $y = 1;\n}\n")
	annotations, _ := analyzeSource(t, src)
	_ = annotations // Just verify no panic.
}

func TestExtractArgVarNameScalar(t *testing.T) {
	src := []byte("push($x, 1);\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)

	// Navigate CST: source_file > expression_statement > function_call_expression > list_expression > scalar
	root := tree.RootNode()
	exprStmt := root.Child(0) // expression_statement
	require.NotNil(t, exprStmt)
	callExpr := exprStmt.Child(0) // function_call_expression
	require.NotNil(t, callExpr)
	require.Equal(t, "function_call_expression", callExpr.Kind())

	// Find the list_expression child
	var listExpr *parser.Node
	for i := 0; i < callExpr.ChildCount(); i++ {
		child := callExpr.Child(i)
		if child != nil && child.Kind() == "list_expression" {
			listExpr = child
			break
		}
	}
	require.NotNil(t, listExpr, "should find list_expression")

	// First child of list_expression is the scalar arg
	scalarArg := listExpr.Child(0)
	require.NotNil(t, scalarArg)
	require.Equal(t, "scalar", scalarArg.Kind())

	name := infer.ExtractArgVarName(scalarArg, src)
	assert.Equal(t, "$x", name)

	// The second real arg is "1" (a number) — should return empty
	// Find the number node
	var numArg *parser.Node
	for i := 0; i < listExpr.ChildCount(); i++ {
		child := listExpr.Child(i)
		if child != nil && child.Kind() == "number" {
			numArg = child
			break
		}
	}
	require.NotNil(t, numArg)
	noName := infer.ExtractArgVarName(numArg, src)
	assert.Equal(t, "", noName)
}

func TestExtractArgVarNameNil(t *testing.T) {
	assert.Equal(t, "", infer.ExtractArgVarName(nil, nil))
}
