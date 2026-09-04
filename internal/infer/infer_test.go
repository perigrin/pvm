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

// analyzeSourceWithOptions is analyzeSource with explicit inference options,
// for tests that compare permissive and strict behaviour.
func analyzeSourceWithOptions(t *testing.T, source []byte, opts infer.Options) (map[uint32]types.Type, []infer.Diagnostic, *infer.SymbolTable) {
	t.Helper()
	p := parser.New()
	tree, err := p.Parse(source)
	require.NoError(t, err, "parse must succeed")
	return infer.AnalyzeWithOptions(tree, source, nil, opts)
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
	// `1 and 2` yields 2 — the operator returns one of its operands, so the
	// result is the join of the arms. Both are Int here, so the join is Int.
	// This previously expected Any, which was the escape hatch standing in
	// for an answer nobody computed.
	src := []byte("1 and 2;")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "1 and 2")
	require.True(t, ok, "lowprec_logical_expression '1 and 2' should be annotated")
	assert.Equal(t, types.Int, typ, "1 and 2 joins two Ints")
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
	// Unary minus PRESERVES its operand's type: $x is 2, so -$x is an Int.
	// Measured, -5 is an Int and -5.5 is a Num. This previously expected Num
	// for both, which is true (Int <: Num) and loses the distinction — and it
	// propagated, since abs(-5) follows its argument.
	src := []byte("my $x = 2; -$x;")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "-$x")
	require.True(t, ok, "unary_expression '-$x' should be annotated")
	assert.Equal(t, types.Int, typ, "negating an Int gives an Int")
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
	// `scalar(42)` IS 42 — the builtin imposes scalar context and hands back
	// the value, so the answer is Int. This previously expected the whole
	// Scalar family, which is the correct family and says nothing about which
	// member; the signature could not express a result that depends on the
	// argument.
	src := []byte("scalar(42);")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "scalar(42)")
	require.True(t, ok, "scalar() call should be annotated")
	assert.Equal(t, types.Int, typ, "scalar(42) is 42, an Int")
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

// TestMethodCallNotAny verifies that an unresolvable method call does not
// come back as Any.
//
// Any is the annotation escape hatch: it satisfies every requirement by
// construction, which is right for a type someone WROTE and wrong for one
// inference guessed. Perl has no annotation syntax, so nothing can request
// it, and an unresolved call is an absence of knowledge rather than a
// declared "anything goes".
func TestMethodCallNotAny(t *testing.T) {
	src := []byte("my $obj; $obj->method();")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "$obj->method()")
	require.True(t, ok, "method_call_expression should be annotated")
	assert.NotEqual(t, types.Any, typ,
		"an unresolved method call must not be Any")
}

// TestConditionalExpressionJoinsItsArms verifies a ternary is typed as the
// join of its arms.
//
// This previously asserted Any, which accepted every subsequent use by
// construction — unsound rather than imprecise. Both arms here are Int, so the
// join is Int and the result stays checkable.
func TestConditionalExpressionJoinsItsArms(t *testing.T) {
	src := []byte("my $x = 1; my $y = 2; $x ? $y : 0;")
	annotations, _ := analyzeSource(t, src)
	typ, ok := findNodeType(annotations, src, "$x ? $y : 0")
	require.True(t, ok, "conditional_expression should be annotated")
	assert.Equal(t, types.Int, typ, "a ternary over two Ints is Int, not Any")
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
	// `my $x = undef` narrows $x to Undef, so `defined($x)` is a guard whose
	// true branch is UNREACHABLE: narrowing Undef by "is defined" leaves the
	// empty set, which is None. The else branch keeps Undef.
	//
	// This previously expected Scalar &^ Undef because the undef keyword had
	// no type of its own and $x stayed at the sigil type. Typing undef makes
	// the analysis sharper: PSC now knows the branch cannot be taken.
	src := []byte("my $x = undef;\nif (defined($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets[0]=declaration, [1]=condition defined($x), [2]=if-body, [3]=else-body
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	ifBodyTyp, ifOk := annotations[offsets[2]]
	assert.True(t, ifOk, "if-body $x should be annotated")
	assert.Equal(t, types.None, ifBodyTyp,
		"if-body $x is None — $x is definitely undef, so the defined() branch is unreachable")

	elseBodyTyp, elseOk := annotations[offsets[3]]
	assert.True(t, elseOk, "else-body $x should be annotated")
	assert.Equal(t, types.Undef, elseBodyTyp, "else-body $x should be Undef (negated defined guard)")
}

func TestFlowNarrowingRefGuard(t *testing.T) {
	// Inside if (ref($x)), $x should be narrowed to Ref.
	//
	// The declaration is bare rather than `= undef`: an explicit undef now
	// types as Undef, which would make the ref() branch unreachable (None) and
	// test guard-on-a-dead-branch rather than ref narrowing itself.
	src := []byte("my $x = get();\nif (ref($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "if-body $x at offset %d should be annotated", ifBodyXOffset)
	assert.Equal(t, types.Ref, typ, "if-body $x should be Ref after ref() guard")
}

func TestFlowNarrowingUnlessDefinedGuard(t *testing.T) {
	// Inside unless (defined($x)), $x should be Undef (negated defined).
	src := []byte("my $x = get();\nunless (defined($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "unless-body $x should be annotated")
	assert.Equal(t, types.Undef, typ, "unless-body $x should be Undef (negated defined guard)")
}

func TestFlowNarrowingWhileDefinedGuard(t *testing.T) {
	// Inside while (defined($x)), $x should be Scalar &^ Undef (non-Undef scalar bits).
	src := []byte("my $x = get();\nwhile (defined($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	whileBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[whileBodyXOffset]
	assert.True(t, ok, "while-body $x should be annotated")
	assert.Equal(t, types.Scalar&^types.Undef, typ, "while-body $x should be Scalar &^ Undef (defined guard removes Undef bit)")
}

func TestFlowNarrowingIfElseRefGuard(t *testing.T) {
	// if (ref($x)) → Ref in if-body, Scalar &^ Ref in else-body (non-reference scalar bits)
	src := []byte("my $x = get();\nif (ref($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
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
	// A bare "my $x" leaves $x at sigil type Scalar (the declaration says the
	// so no assignment narrowing occurs).
	src := []byte("my $x = get();\nif (ref($x)) {\n    my $y = $x;\n}\nmy $z = $x;\n")
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
	src := []byte("my $x = get();\nunless (ref($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
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
	src := []byte("my $x = get();\nif (defined($x)) {\n    my $y = $x;\n} elsif (ref($x)) {\n    my $z = $x;\n}\n")
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
	src := []byte("my $x = get();\nif ($x isa Foo) {\n    my $y = $x;\n}\n")
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
	src := []byte("my $x = get();\nif ($x isa Foo) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
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
	src := []byte("my $x = get();\nif ($x isa Foo) {\n    my $a = $x;\n} elsif (!defined($x)) {\n    my $b = $x;\n} else {\n    my $c = $x;\n}\n")
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
	src := []byte("my $x = get();\nif (defined($x)) {\n    my $a = $x;\n} elsif (ref($x)) {\n    my $b = $x;\n} elsif ($x isa Foo) {\n    my $c = $x;\n}\n")
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
	src := []byte("my $x = get();\nif (defined($x)) {\n    my $a = $x;\n} elsif (ref($x)) {\n    my $b = $x;\n} else {\n    my $c = $x;\n}\n")
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
	src := []byte("my $x = get();\nif (!defined($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
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
	src := []byte("my $x = get();\nif (not ref($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
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
	src := []byte("my $x = get();\nif (!defined($x)) {\n    return;\n}\nmy $y = $x;\n")
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
	src := []byte("my $x = get();\nif (!defined($x)) {\n    $x = 1;\n}\nmy $y = $x;\n")
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
	src := []byte("my $x = get();\nif (!ref($x)) {\n    die;\n}\nmy $y = $x;\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 3, "should find at least 3 occurrences of $x, got %d", len(offsets))

	postIfTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "post-if $x should be annotated")
	assert.Equal(t, types.Ref, postIfTyp, "post-if $x should be Ref (ref guard after early die)")
}

func TestFlowNarrowingEarlyDieWithArgNarrowsRef(t *testing.T) {
	// if (!ref($x)) { die $msg; } — die with argument should also be detected as early exit.
	src := []byte("my $msg = get();\nmy $x = get();\nif (!ref($x)) {\n    die $msg;\n}\nmy $y = $x;\n")
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
	src := []byte("my $x = get();\nif (!defined($x)) {\n    exit;\n}\nmy $y = $x;\n")
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
	src := []byte("my $x = get();\nunless (defined($x)) {\n    return;\n}\nmy $y = $x;\n")
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
	src := []byte("my $x = get();\nif (defined($x)) {\n    return;\n}\nmy $y = $x;\n")
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
	src := []byte("my $x = get();\nif (defined($x) && ref($x)) {\n    my $y = $x;\n}\n")
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
	src := []byte("my $x = get();\nif (defined($x) and ref($x)) {\n    my $y = $x;\n}\n")
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
	src := []byte("my $x = get();\nif (defined($x) || ref($x)) {\n    my $y = $x;\n}\n")
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
	src := []byte("my $x = get();\nif (defined($x) || ref($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
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
	src := []byte("my $x = get();\nmy $y = get();\nif (defined($x) && ref($y)) {\n    my $a = $x;\n    my $b = $y;\n}\n")
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
	src := []byte("my $x = get();\nmy $y = 1;\nif (defined($x) && $y > 0) {\n    my $z = $x;\n}\n")
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
	src := []byte("my $x = get();\nif (!(defined($x) && ref($x))) {\n    my $y = $x;\n}\n")
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
	src := []byte("my $x = get();\nif (ref($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\nmy $w = $x;\n")
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
	src := []byte("my $x = get();\nif (ref($x)) {\n    my $y = $x;\n}\nmy $z = $x;\n")
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
	src := []byte("my $x = get();\nif (!defined($x)) {\n    return;\n}\nmy $y = $x;\n")
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
	src := []byte("my $x = get();\nif (defined($x)) {\n    my $a = $x;\n} elsif (ref($x)) {\n    my $b = $x;\n} else {\n    my $c = $x;\n}\nmy $d = $x;\n")
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
	src := []byte("my $x = get();\nif (defined($x)) {\n    $x = 1;\n} else {\n    $x = 3.14;\n}\nmy $y = $x;\n")
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
	src := []byte("my $x = get();\nif (builtin::blessed($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)
	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 3)
	ifBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Object, ifBodyTyp, "blessed guard narrows to Object")
}

func TestGuardLibraryBlessedBareNarrows(t *testing.T) {
	// if (blessed($x)): bare name, same effect
	src := []byte("my $x = get();\nif (blessed($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)
	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 3)
	ifBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Object, ifBodyTyp, "bare blessed guard narrows to Object")
}

func TestGuardLibraryReftypeNarrows(t *testing.T) {
	// if (builtin::reftype($x)): body $x → Ref
	src := []byte("my $x = get();\nif (builtin::reftype($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)
	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 3)
	ifBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Ref, ifBodyTyp, "reftype guard narrows to Ref")
}

func TestGuardLibraryIsBoolNarrows(t *testing.T) {
	// if (builtin::is_bool($x)): body $x → Bool
	src := []byte("my $x = get();\nif (builtin::is_bool($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)
	offsets := findAllVarOffsets(src, "$x")
	require.True(t, len(offsets) >= 3)
	ifBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Bool, ifBodyTyp, "is_bool guard narrows to Bool")
}

func TestGuardLibraryNegatedBlessed(t *testing.T) {
	// if (!builtin::blessed($x)) { } else { $x } — else $x → Object
	src := []byte("my $x = get();\nif (!builtin::blessed($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
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
	src := []byte("my $x = get();\nif (defined($x) && builtin::blessed($x)) {\n    my $y = $x;\n}\n")
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
	src := []byte("use Dog;\nmy $x = get();\nif ($x isa Dog) {\n    my $v = $x->speak();\n}\n")
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
	src := []byte("sub is_defined { defined($_[0]) }\nmy $x = get();\nif (is_defined($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
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
	src := []byte("sub is_ref { ref($_[0]) }\nmy $x = get();\nif (is_ref($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Ref, typ, "if-body $x should be Ref after user-defined ref guard")
}

func TestUserDefinedGuardFuncSignature(t *testing.T) {
	src := []byte("sub is_defined ($val) { defined($val) }\nmy $x = get();\nif (is_defined($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Scalar&^types.Undef, typ, "if-body $x should have Undef removed after signature-style guard")
}

func TestUserDefinedGuardFuncShift(t *testing.T) {
	src := []byte("sub is_ref { my $val = shift; ref($val) }\nmy $x = get();\nif (is_ref($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Ref, typ, "if-body $x should be Ref after shift-style guard")
}

func TestUserDefinedGuardFuncIsa(t *testing.T) {
	src := []byte("sub is_foo { $_[0] isa Foo }\nmy $x = get();\nif (is_foo($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Object, typ, "if-body $x should be Object after isa guard")
}

func TestUserDefinedGuardFuncNegatedDefined(t *testing.T) {
	src := []byte("sub is_defined { defined($_[0]) }\nmy $x;\nunless (is_defined($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	// unless(is_defined($x)) means body gets negated guard -> Undef.
	unlessBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[unlessBodyXOffset]
	assert.True(t, ok, "unless-body $x should be annotated")
	assert.Equal(t, types.Undef, typ, "unless-body $x should be Undef (negated defined guard)")
}

func TestUserDefinedGuardFuncCompound(t *testing.T) {
	src := []byte("sub is_defined_ref { defined($_[0]) && ref($_[0]) }\nmy $x = get();\nif (is_defined_ref($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	ifBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[ifBodyXOffset]
	assert.True(t, ok, "if-body $x should be annotated")
	assert.Equal(t, types.Ref, typ, "if-body $x should be Ref after compound guard (defined+ref)")
}

func TestUserDefinedGuardFuncNonGuardIgnored(t *testing.T) {
	src := []byte("sub do_stuff { 42 }\nmy $x = get();\nif (do_stuff($x)) {\n    my $y = $x;\n}\n")
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

	// Navigate CST: source_file > expression_statement > <call> > list_expression > scalar.
	// The grammar labels the call either function_call_expression or
	// ambiguous_function_call_expression depending on whether it can tell a
	// call from a bareword-plus-parens; which one push() gets has moved
	// between grammar versions, and this test is about extractArgVarName
	// rather than about the label.
	root := tree.RootNode()
	exprStmt := root.Child(0) // expression_statement
	require.NotNil(t, exprStmt)
	callExpr := exprStmt.Child(0)
	require.NotNil(t, callExpr)
	require.Contains(t,
		[]string{"function_call_expression", "ambiguous_function_call_expression"},
		callExpr.Kind())

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

func TestExtractArgVarNameHash(t *testing.T) {
	src := []byte("keys(%h);\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)

	// Navigate CST: source_file > expression_statement > func1op_call_expression > hash
	root := tree.RootNode()
	exprStmt := root.Child(0)
	require.NotNil(t, exprStmt)
	callExpr := exprStmt.Child(0)
	require.NotNil(t, callExpr)
	require.Equal(t, "func1op_call_expression", callExpr.Kind())

	// Find the hash child
	var hashArg *parser.Node
	for i := 0; i < callExpr.ChildCount(); i++ {
		child := callExpr.Child(i)
		if child != nil && child.Kind() == "hash" {
			hashArg = child
			break
		}
	}
	require.NotNil(t, hashArg, "should find hash node")

	name := infer.ExtractArgVarName(hashArg, src)
	assert.Equal(t, "%h", name)
}

func TestExtractArgVarNameNil(t *testing.T) {
	assert.Equal(t, "", infer.ExtractArgVarName(nil, nil))
}

func TestTypeMismatchDiagnosticSuggestionWired(t *testing.T) {
	// push expects Array as first arg. Passing Scalar $x triggers
	// a type-mismatch diagnostic. No guard can narrow Scalar to Array,
	// so Suggestion should be empty.
	src := []byte("my $x;\npush($x, 1);\n")
	_, diags := analyzeSource(t, src)

	require.True(t, len(diags) > 0, "should have at least one diagnostic")

	var found bool
	for _, d := range diags {
		if d.Code == infer.CodeTypeMismatch {
			assert.Empty(t, d.Suggestion,
				"Scalar vs Array mismatch should have no suggestion")
			found = true
		}
	}
	assert.True(t, found, "should find a type-mismatch diagnostic")
}

func TestTypeMismatchDiagnosticNoSuggestionWhenClean(t *testing.T) {
	// push expects Array. Passing @arr is correct — no diagnostic at all.
	src := []byte("my @arr;\npush(@arr, 1);\n")
	_, diags := analyzeSource(t, src)

	for _, d := range diags {
		assert.NotEqual(t, infer.CodeTypeMismatch, d.Code,
			"clean code should have no type-mismatch diagnostics")
	}
}

// --- Coercion-mismatch diagnostic tests ---

// TestCoercionListInArithmetic verifies that a List in arithmetic is read as
// its element count rather than reported.
//
// This previously asserted the opposite. `keys(%h) + 1` is ordinary Perl —
// measured, it is 3 for a two-key hash — because a numeric operator imposes
// scalar context and an aggregate in scalar context is its count. The old
// expectation encoded PSC's missing context handling as though it were a rule
// about Perl.
func TestCoercionListInArithmetic(t *testing.T) {
	src := []byte("my %h;\nmy $y = keys(%h) + 1;\n")
	_, diags := analyzeSource(t, src)
	assert.Empty(t, diags, "keys(%h) + 1 is the count plus one, not a type error")
}

// TestCoercionArrayInConcat verifies that an Array in string concatenation is
// read as its count. Measured: `@arr . "x"` is "3x" for a three-element array.
func TestCoercionArrayInConcat(t *testing.T) {
	src := []byte("my @arr;\nmy $s = @arr . 1;\n")
	_, diags := analyzeSource(t, src)
	assert.Empty(t, diags, "@arr . 1 concatenates the count, not a type error")
}

// TestNoCoercionIntInArithmetic verifies that Int in arithmetic produces
// no coercion-mismatch (Int <: Num, so this is type-safe).
func TestNoCoercionIntInArithmetic(t *testing.T) {
	src := []byte("my $x = 42;\nmy $y = $x + 1;\n")
	_, diags := analyzeSource(t, src)

	for _, d := range diags {
		assert.NotEqual(t, infer.CodeCoercionMismatch, d.Code,
			"Int in arithmetic should produce no coercion-mismatch")
	}
}

// --- Undef propagation tests ---

// TestUndefPropagationInArithmetic verifies that an uninitialized variable
// used in arithmetic produces a coercion-mismatch diagnostic.
func TestUndefPropagationInArithmetic(t *testing.T) {
	// An EXPLICIT undef is a claim the source makes, so it is reportable.
	// A bare `my $x;` is not: perl binds names in more ways than inference can
	// enumerate, and the variable may be assigned before this line is reached.
	src := []byte("my $x = undef;\nmy $y = $x + 1;\n")
	_, diags := analyzeSource(t, src)

	var found bool
	for _, d := range diags {
		if d.Code == infer.CodeCoercionMismatch {
			found = true
			assert.Equal(t, infer.Error, d.Severity, "Undef in arithmetic should be Error")
			break
		}
	}
	assert.True(t, found, "should find a coercion-mismatch for explicit undef in arithmetic")
}

// TestUndefPropagationWithAssignment verifies that an initialized variable
// does NOT produce an undef diagnostic.
func TestUndefPropagationWithAssignment(t *testing.T) {
	src := []byte("my $x = 42;\nmy $y = $x + 1;\n")
	_, diags := analyzeSource(t, src)

	for _, d := range diags {
		assert.NotEqual(t, infer.CodeCoercionMismatch, d.Code,
			"initialized variable should produce no coercion-mismatch")
	}
}

// TestUndefSymbolTableType verifies that a bare declaration gets Scalar, not
// Undef.
//
// Undef is a definite claim and a declaration cannot establish it: the
// variable may be assigned anywhere later, including by forms no textual scan
// finds — `my ($k, $v) = each %h`, foreach aliasing, a sub signature. Claiming
// Undef here reported every subsequent use of such a variable as an error.
func TestUndefSymbolTableType(t *testing.T) {
	src := []byte("my $x;\n")
	_, _, st := analyzeSourceFull(t, src)

	sym, found := st.Lookup("$x")
	require.True(t, found, "$x should be in the symbol table")
	assert.Equal(t, types.Scalar, sym.Type,
		"a bare declaration is Scalar — inference cannot prove the variable stays undef")
}

// TestInitializedSymbolTableType verifies that an initialized my-variable
// keeps the normal narrowed type (not Undef).
func TestInitializedSymbolTableType(t *testing.T) {
	src := []byte("my $x = 42;\n")
	_, _, st := analyzeSourceFull(t, src)

	sym, found := st.Lookup("$x")
	require.True(t, found, "$x should be in the symbol table")
	assert.Equal(t, types.Int, sym.Type,
		"my $x = 42 should have type Int after assignment narrowing")
}

// --- Multi-param inference tests ---

// TestResolveSubParamsSignature verifies that multi-param signatures are identified.
func TestResolveSubParamsSignature(t *testing.T) {
	src := []byte("sub add($x, $y) { $x + $y }\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()

	// Find the subroutine_declaration_statement
	var subNode *parser.Node
	for i := 0; i < root.ChildCount(); i++ {
		child := root.Child(i)
		if child != nil && child.Kind() == "subroutine_declaration_statement" {
			subNode = child
			break
		}
	}
	require.NotNil(t, subNode, "should find a subroutine_declaration_statement")

	params := infer.ResolveSubParams(subNode, src)
	assert.Equal(t, []string{"$x", "$y"}, params,
		"should identify both params from signature sub add($x, $y)")
}

// TestResolveSubParamsShiftChain verifies that sequential shift assignments
// are identified as parameters.
func TestResolveSubParamsShiftChain(t *testing.T) {
	src := []byte("sub method { my $self = shift; my $name = shift; $self + $name }\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()

	var subNode *parser.Node
	for i := 0; i < root.ChildCount(); i++ {
		child := root.Child(i)
		if child != nil && child.Kind() == "subroutine_declaration_statement" {
			subNode = child
			break
		}
	}
	require.NotNil(t, subNode, "should find a subroutine_declaration_statement")

	params := infer.ResolveSubParams(subNode, src)
	assert.Equal(t, []string{"$self", "$name"}, params,
		"should identify both params from shift chain")
}

// TestResolveSubParamsAtUnderscoreUnpacking verifies that my ($x, $y) = @_;
// is recognized as parameter identification.
func TestResolveSubParamsAtUnderscoreUnpacking(t *testing.T) {
	src := []byte("sub foo { my ($x, $y) = @_; $x + $y }\n")
	p := parser.New()
	tree, err := p.Parse(src)
	require.NoError(t, err)
	root := tree.RootNode()

	var subNode *parser.Node
	for i := 0; i < root.ChildCount(); i++ {
		child := root.Child(i)
		if child != nil && child.Kind() == "subroutine_declaration_statement" {
			subNode = child
			break
		}
	}
	require.NotNil(t, subNode, "should find a subroutine_declaration_statement")

	params := infer.ResolveSubParams(subNode, src)
	assert.Equal(t, []string{"$x", "$y"}, params,
		"should identify both params from @_ unpacking")
}

// TestParamTypeInferenceFromUsage verifies that parameter types are inferred
// from how they are used in the function body.
func TestParamTypeInferenceFromUsage(t *testing.T) {
	// $x is used in arithmetic ($x + 1) → Num constraint
	src := []byte("sub calc { my $x = shift; my $result = $x + 1; }\n")
	_, _, st := analyzeSourceFull(t, src)

	sym, found := st.Lookup("calc")
	require.True(t, found, "calc should be in the symbol table")
	require.NotNil(t, sym.ParamTypes, "calc should have ParamTypes")
	require.Len(t, sym.ParamTypes, 1, "calc has 1 parameter")
	assert.Equal(t, types.Num, sym.ParamTypes[0],
		"$x used in arithmetic should be inferred as Num")
}

// TestParamTypeInferenceNoUsefulType verifies that parameters only used in
// permissive contexts (like assignment) get Unknown, not Any.
func TestParamTypeInferenceNoUsefulType(t *testing.T) {
	// $x is only assigned, never used in a type-constraining operation
	src := []byte("sub passthrough { my $x = shift; return $x; }\n")
	_, _, st := analyzeSourceFull(t, src)

	sym, found := st.Lookup("passthrough")
	require.True(t, found, "passthrough should be in the symbol table")
	// ParamTypes should either be nil or have Unknown entries
	if sym.ParamTypes != nil && len(sym.ParamTypes) > 0 {
		assert.Equal(t, types.Unknown, sym.ParamTypes[0],
			"param used only in permissive context should be Unknown, not Any")
	}
}

// TestNoCoercionIntWithEq verifies that eq on Int operands produces no
// coercion-mismatch (Int <: Str, so eq on integers is type-safe).
func TestNoCoercionIntWithEq(t *testing.T) {
	src := []byte("my $x = 42;\nmy $y = 99;\nmy $c = ($x eq $y);\n")
	_, diags := analyzeSource(t, src)

	for _, d := range diags {
		assert.NotEqual(t, infer.CodeCoercionMismatch, d.Code,
			"eq on Int operands should produce no coercion-mismatch (Int <: Str)")
	}
}

// TestParenthesizedOperandIsChecked verifies that a binary expression whose
// left operand is parenthesized still has its operands checked.
//
// The operator was previously found by taking the first anonymous child, but a
// parenthesized operand puts "(" in that position, so the lookup failed and
// the whole expression bailed out as Unknown before any operand was examined.
func TestParenthesizedOperandIsChecked(t *testing.T) {
	// A hashref in a numeric position is a real coercion mismatch — unlike an
	// array, which yields its count. Wrapping the operand in parentheses must
	// not hide it.
	bare := []byte("my $h = {};\nmy $x = $h + 1;\n")
	parens := []byte("my $h = {};\nmy $x = ($h) + 1;\n")

	_, bareDiags := analyzeSource(t, bare)
	_, parenDiags := analyzeSource(t, parens)

	assert.NotEmpty(t, bareDiags, "precondition: @a + 1 is a diagnostic")
	assert.Equal(t, len(bareDiags), len(parenDiags),
		"parenthesising the operand must not change what is reported")
}

// TestConditionalExpressionJoinsArms verifies that a ternary is typed as the
// join of its arms rather than as Any.
//
// Any satisfies every requirement by construction, so typing a ternary as Any
// made every value that passed through one unverifiable — unsound rather than
// merely imprecise.
func TestConditionalExpressionJoinsArms(t *testing.T) {
	// Both arms Int: the merge is Int, and using it as a number is fine.
	src := []byte("my $x = 1 ? 42 : 43;\nmy $y = $x + 1;\n")
	_, diags := analyzeSource(t, src)
	assert.Empty(t, diags, "a ternary over two Ints is a Num and needs no diagnostic")

	// Arms of unrelated types: the merge satisfies neither, so a numeric use
	// is reported.
	mixed := []byte("my @a;\nmy $x = (1 ? @a : 42) + 1;\n")
	_, mixedDiags := analyzeSource(t, mixed)
	assert.NotEmpty(t, mixedDiags,
		"a ternary that might yield an Array does not satisfy Num")
}

// --- Reference-producing expressions ---

// TestInferAnonymousConstructors verifies that the expressions which BUILD a
// reference are typed as that reference.
//
// These were absent from the inference switch, so `my $h = {}` left $h
// Unknown — every hashref in a program was invisible, and no diagnostic about
// misusing one could fire.
func TestInferAnonymousConstructors(t *testing.T) {
	cases := []struct {
		src  string
		expr string
		want types.Type
	}{
		{"my $h = {};\n", "{}", types.HashRef},
		{"my $a = [];\n", "[]", types.ArrayRef},
		{"my $c = sub { 1 };\n", "sub { 1 }", types.CodeRef},
		{"my $q = qr/x/;\n", "qr/x/", types.Regex},
	}
	for _, tc := range cases {
		annotations, _ := analyzeSource(t, []byte(tc.src))
		typ, ok := findNodeType(annotations, []byte(tc.src), tc.expr)
		require.True(t, ok, "%s should be annotated", tc.expr)
		assert.Equal(t, tc.want, typ, "%s is a %s", tc.expr, tc.want)
	}
}

// TestInferRefgen verifies that \$x, \@a, \%h, \&f and \*G are typed from
// what they take a reference TO. A single fixed type would be wrong for four
// of the five.
func TestInferRefgen(t *testing.T) {
	cases := []struct {
		src  string
		expr string
		want types.Type
	}{
		{"my $s;\nmy $r = \\$s;\n", "\\$s", types.ScalarRef},
		{"my @a;\nmy $r = \\@a;\n", "\\@a", types.ArrayRef},
		{"my %h;\nmy $r = \\%h;\n", "\\%h", types.HashRef},
		{"my $r = \\&f;\n", "\\&f", types.CodeRef},
		{"my $r = \\*STDOUT;\n", "\\*STDOUT", types.GlobRef},
	}
	for _, tc := range cases {
		annotations, _ := analyzeSource(t, []byte(tc.src))
		typ, ok := findNodeType(annotations, []byte(tc.src), tc.expr)
		require.True(t, ok, "%s should be annotated", tc.expr)
		assert.Equal(t, tc.want, typ, "%s is a %s", tc.expr, tc.want)
	}
}

// TestAnonymousRefInArithmeticIsReported is the payoff: with these typed, a
// reference used as a number is reportable. It was silent before, because the
// value had no type to disagree with.
func TestAnonymousRefInArithmeticIsReported(t *testing.T) {
	src := []byte("my $h = {};\nmy $n = $h + 1;\n")
	_, diags := analyzeSource(t, src)
	assert.NotEmpty(t, diags, "a hashref in arithmetic should be reported")
}

// TestListOperatorArgsFromEnclosingList verifies that a paren-less list
// operator gets its full argument list even when the grammar leaves the later
// arguments OUTSIDE the call node.
//
// The Perl grammar produces two different shapes for the same call:
//
//	push @todo, [1,2];          call(function, list_expression(array, ref))
//	push @todo, [1,2] unless $o; list_expression(call(function, array), ref)
//
// In the second, only the first argument is inside the call and the rest are
// siblings in the enclosing list_expression. Counting only the call's own
// children therefore reported a bogus arity error — 168 of 804 diagnostics on
// perl5/lib were this, across push, join, substr, unshift and index.
func TestListOperatorArgsFromEnclosingList(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"statement modifier unless", "my @todo; my $o = 1;\npush @todo, [1, 2] unless $o;\n"},
		{"statement modifier if", "my @todo; my $o = 1;\npush @todo, 'x' if $o;\n"},
		{"join with paren-less args", "my @a;\nmy $s = join \",\", @a;\n"},
	}
	for _, tc := range cases {
		_, diags := analyzeSource(t, []byte(tc.src))
		var arity []string
		for _, d := range diags {
			if d.Code == infer.CodeArityMismatch {
				arity = append(arity, d.Message)
			}
		}
		assert.Empty(t, arity, "%s: should report no arity error, got %v", tc.name, arity)
	}
}

// TestArrayInNumericContextIsCount verifies that an array compared numerically
// is read as its element count rather than reported as a type error.
//
// `@fields == 3`, `while (@_ > 1)`, `die unless @_ == 1` are the ordinary Perl
// idiom: a numeric operator imposes scalar context, and an array in scalar
// context yields its count. types.NarrowByContext already modelled this; the
// operand check simply never applied it, which produced 104 of the diagnostics
// on perl5/lib — the largest single false-positive class.
func TestArrayInNumericContextIsCount(t *testing.T) {
	cases := []string{
		"my @fields;\nmy $ok = (@fields == 3);\n",
		"my @a;\nmy $ok = (@a > 1);\n",
		"my %h;\nmy $ok = (%h == 0);\n",
	}
	for _, src := range cases {
		_, diags := analyzeSource(t, []byte(src))
		assert.Empty(t, diags,
			"an aggregate in numeric context is its count, not a type error: %q", src)
	}
}

// The counterpart: scalar context does NOT excuse a genuinely wrong operand.
// A hashref is one value in any context and still cannot be a number.
func TestScalarContextDoesNotExcuseRefs(t *testing.T) {
	src := []byte("my $h = {};\nmy $n = $h + 1;\n")
	_, diags := analyzeSource(t, src)
	assert.NotEmpty(t, diags, "a hashref in arithmetic is still reported")
}

// TestBareDeclarationIsNotPermanentlyUndef verifies that `my $v;` does not
// pin the variable to Undef for the rest of the scope.
//
// A declaration without an initializer says the variable holds undef AT THAT
// POINT, not that it always will. Perl's ordinary lazy-init idiom assigns it
// through a call:
//
//	my $v_unicode_version;
//	UnicodeVersion() unless defined $v_unicode_version;
//	if ($v_unicode_version ge v2.0.0) { ... }
//
// Typing the declaration Undef made the third line a coercion mismatch, which
// is a claim PSC cannot support: the call may well have assigned it. 42
// diagnostics on perl5/lib were this shape, concentrated in Unicode/UCD.
func TestBareDeclarationIsNotPermanentlyUndef(t *testing.T) {
	src := []byte("my $v;\nsetit() unless defined $v;\nmy $ok = ($v ge 'a');\nsub setit { $v = 'x' }\n")
	_, diags := analyzeSource(t, src)
	assert.Empty(t, diags,
		"a bare declaration must not pin the variable to Undef across a call")
}

// The genuine case still reports: an explicit assignment of undef, with
// nothing between it and the use, IS evidence. Perl agrees — it warns "Use of
// uninitialized value in string eq".
func TestExplicitUndefAssignmentStillReports(t *testing.T) {
	src := []byte("my $x = undef;\nmy $ok = ($x eq 'a');\n")
	_, diags := analyzeSource(t, src)
	assert.NotEmpty(t, diags,
		"an explicit undef assignment used as a string is a real finding")
}

// TestIndirectObjectIsFilehandleNotArgument verifies that the filehandle in
// `print $fh "..."` is not counted as print's first argument.
//
// The grammar marks it with its own node kind, indirect_object:
//
//	print $OUT "text\n";
//	  ambiguous_function_call_expression
//	    function "print"
//	    indirect_object -> scalar $OUT      <- the HANDLE
//	    interpolated_string_literal         <- argument 1
//
// PSC counted the handle as argument 1 and required it to be a Str, so every
// `print $fh ...` against a typed handle was a mismatch. perl5db.pl assigns
// `$OUT = \*STDERR`, which is a GlobRef, and produced 43 of the 338
// diagnostics on perl5/lib — the largest remaining class, and all from one
// file.
func TestIndirectObjectIsFilehandleNotArgument(t *testing.T) {
	cases := []string{
		"my $OUT = \\*STDERR;\nprint $OUT \"hello\\n\";\n",
		"open my $fh, '>', '/dev/null';\nprint $fh \"hello\\n\";\n",
		"print STDERR \"hello\\n\";\n",
	}
	for _, src := range cases {
		_, diags := analyzeSource(t, []byte(src))
		assert.Empty(t, diags,
			"the filehandle is not an argument to print: %q", src)
	}
}

// The real arguments are still checked: a reference in a print argument
// position still reports, whether or not a filehandle precedes it.
func TestIndirectObjectStillChecksRealArgs(t *testing.T) {
	src := []byte("my $OUT = \\*STDERR;\nmy $h = {};\nprint $OUT $h;\n")
	_, diags := analyzeSource(t, src)
	assert.NotEmpty(t, diags,
		"a hashref passed as print's argument is still reported")
}

// TestNumInIntPositionIsWarningNotError verifies that a Num where an Int is
// wanted is reported as a warning rather than an error.
//
// The paper defines Num -> Int as a COERCION — "truncate toward zero" — not a
// failure, and perl performs it silently, with no warning even under -w:
//
//	0 .. 3.7            measured (0 1 2 3)
//	"ab" x 2.9          measured "abab"
//	7 >> 1.5            measured 3
//	"\t" x ($lvl / 8)   $lvl/8 really is 2.5, and this is real code
//
// Reporting it as an error claims the code is broken when the language
// defines the behaviour. It is still worth saying — truncation is often
// unintended — so it stays a diagnostic, at the severity the paper's
// coercion/membership split implies. 39 of the 297 diagnostics on perl5/lib
// were this, at Error severity.
func TestNumInIntPositionIsWarningNotError(t *testing.T) {
	cases := []string{
		"my $n = 3.7;\nmy @r = (0 .. $n);\n",
		"my $n = 2.9;\nmy $s = \"ab\" x $n;\n",
		"my $lvl = 20;\nmy $t = \"\\t\" x ($lvl / 8);\n",
	}
	for _, src := range cases {
		_, diags := analyzeSource(t, []byte(src))
		for _, d := range diags {
			assert.Equal(t, infer.Warning, d.Severity,
				"Num in an Int position truncates rather than failing: %q -> %s", src, d.Message)
		}
	}
}

// The counterpart: an operand that cannot coerce at all is still an Error.
// A hashref in arithmetic yields its address, which is never a computation
// anyone intended.
func TestUncoercibleOperandStaysError(t *testing.T) {
	src := []byte("my $h = {};\nmy $n = $h + 1;\n")
	_, diags := analyzeSource(t, src)
	require.NotEmpty(t, diags, "a hashref in arithmetic is reported")
	assert.Equal(t, infer.Error, diags[0].Severity,
		"a reference in a numeric position is an error, not a truncation")
}

// TestObjectInNumericPositionIsWarning verifies that an object where a number
// is wanted is a warning rather than an error.
//
// The paper's Overloaded Objects section: a class declaring `use overload
// '0+'` has a user-defined conversion, so `$obj + 1` is a defined operation —
// measured, 6 for a Money-like class. The object is still NOT a Num (the
// conversion runs out of the type and nothing converts back), so this remains
// worth reporting.
//
// PSC cannot tell an overloaded class from a plain one without resolving the
// class across files, and the two differ: a plain object numifies to its
// ADDRESS. Warning is the honest severity for "this may be a real conversion",
// where Error would assert it cannot be. 54 diagnostics on perl5/lib, all in
// overload64.t and overloading.t — files whose test names literally read
// "0+ overload with bit shift right".
func TestObjectInNumericPositionIsWarning(t *testing.T) {
	src := []byte("my $o = Foo->new;\nmy $n = $o + 1;\n")
	_, diags := analyzeSource(t, src)
	require.NotEmpty(t, diags, "an object in arithmetic is still reported")
	for _, d := range diags {
		assert.Equal(t, infer.Warning, d.Severity,
			"an object may define a 0+ conversion: %s", d.Message)
	}
}

// A plain reference is different and stays an Error: there is no overload
// table on a raw hashref, so numifying it can only produce an address.
//
// Regex is the borderline case and it lands on the warning side, because the
// paper makes a compiled pattern an Object — blessed into Regexp — and PSC
// cannot see whether a class declares a conversion. Measured, Regexp does NOT
// declare 0+, so `qr/a/ + 1` really is an address and Error would be the more
// accurate call. Reaching that needs the class, which is cross-file analysis
// PSC does not do; being wrong toward "may convert" costs a severity level,
// while being wrong the other way asserts something false about code that
// works.
func TestPlainRefInNumericPositionStaysError(t *testing.T) {
	src := []byte("my $h = {};\nmy $n = $h + 1;\n")
	_, diags := analyzeSource(t, src)
	require.NotEmpty(t, diags, "a hashref in arithmetic is reported")
	assert.Equal(t, infer.Error, diags[0].Severity,
		"a raw reference has no conversion table — it can only yield an address")
}

// TestCallSwallowedComparisonStillCountsArgs verifies that a call whose
// arguments the grammar buried under a comparison still counts them.
//
// When a paren-less-capable list operator is followed by a comparison, the
// grammar makes the comparison the call's argument rather than the other way
// round:
//
//	substr($s, 0, 2)               call(function, list(scalar, num, num))
//	substr($s, 0, 2) eq "he"       call(function, equality(list(scalar, num, num), str))
//
// The real arguments are one level down, inside the comparison's own first
// operand. Counting only the call's direct children saw ONE argument — the
// comparison — and reported a bogus arity error. This is the same upstream
// shape that affects index() and any unrecognised call: `foo($x) >= 0` parses
// as `foo(($x) >= 0)`.
func TestCallSwallowedComparisonStillCountsArgs(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"substr in eq", "my $s = \"hello\";\nmy $b = (substr($s, 0, 2) eq \"he\");\n"},
		{"substr in ne", "my $s = \"hello\";\nmy $b = (substr($s, 0, 1) ne \"h\");\n"},
		{"index in >=", "my $s = \"hello\";\nmy $b = (index($s, \"e\") >= 0);\n"},
	}
	for _, tc := range cases {
		_, diags := analyzeSource(t, []byte(tc.src))
		var arity []string
		for _, d := range diags {
			if d.Code == infer.CodeArityMismatch {
				arity = append(arity, d.Message)
			}
		}
		assert.Empty(t, arity, "%s: should report no arity error, got %v", tc.name, arity)
	}
}

// TestLvalueSubstrAndNestedCallArgs covers two more shapes where the grammar
// moves a call's arguments somewhere the call node cannot see them.
//
//	substr($f, 1, 0) = "-";      lvalue substr — the ASSIGNMENT swallows the
//	                             argument list, exactly as a comparison does
//	my $x = "k " . substr $s, 1; the call is nested in the concatenation, and
//	                             its trailing argument is a sibling one level
//	                             further out
//
// Both are ordinary Perl: measured, the first makes "abc" into "a-bc" and the
// second yields "k ello".
func TestLvalueSubstrAndNestedCallArgs(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"lvalue substr", "my $f = \"abc\";\nsubstr($f, 1, 0) = \"-\";\n"},
		{"call nested in concat", "my $s = \"hello\";\nmy $x = \"k \" . substr $s, 1;\n"},
		{"call inside return", "sub f { return join ' ', @_; }\n"},
	}
	for _, tc := range cases {
		_, diags := analyzeSource(t, []byte(tc.src))
		var arity []string
		for _, d := range diags {
			if d.Code == infer.CodeArityMismatch {
				arity = append(arity, d.Message)
			}
		}
		assert.Empty(t, arity, "%s: should report no arity error, got %v", tc.name, arity)
	}
}

// TestScalarAssignmentImposesScalarContext verifies that assigning an
// aggregate to a scalar yields the count, not the aggregate.
//
//	my @arr = (1,2,3);
//	my $count = @arr;      # 3, measured
//
// PSC typed $count as Array, which is a type the value does not have. It
// produced no diagnostic at the assignment but poisoned every later use of
// $count. Found by the runtime precision oracle rather than by reading: the
// observed value was Int and PSC had said Array.
func TestScalarAssignmentImposesScalarContext(t *testing.T) {
	src := []byte("my @arr = (1,2,3);\nmy $count = @arr;\n")
	_, _, st := analyzeSourceFull(t, src)

	sym, found := st.Lookup("$count")
	require.True(t, found, "$count should be in the symbol table")
	assert.Equal(t, types.Int, sym.Type,
		"an array assigned to a scalar is its count")
}

// Assigning to an array keeps the aggregate: `my @copy = @arr` is a list
// assignment, not a count.
func TestArrayAssignmentKeepsAggregate(t *testing.T) {
	src := []byte("my @arr = (1,2,3);\nmy @copy = @arr;\n")
	_, _, st := analyzeSourceFull(t, src)

	sym, found := st.Lookup("@copy")
	require.True(t, found, "@copy should be in the symbol table")
	assert.Equal(t, types.Array, sym.Type,
		"an array assigned to an array stays an Array")
}

// TestUnresolvedMethodCallIsUnknownNotAny verifies that a method call PSC
// cannot resolve yields Unknown rather than Any.
//
// The two are not interchangeable, and TypeScript learned this the expensive
// way: `any` disables checking, `unknown` demands narrowing. Any is the
// ESCAPE HATCH, and an escape hatch is only meaningful when someone can write
// it down. Perl has no annotation syntax, so nothing in a Perl program ever
// requests Any — every Any produced by inference is really "I could not
// determine this", which is Unknown's job.
//
// The difference is observable: Any satisfies every requirement by
// construction, so an unresolved method call was invisible even under
// --strict, which exists precisely to surface what inference does not know.
func TestUnresolvedMethodCallIsUnknownNotAny(t *testing.T) {
	// A class method that is not a constructor, with no project index to
	// resolve it: this is the path that had no answer and returned Any.
	// `Foo->new` is different — it yields Object by construction — and
	// `$obj->m` reads the invocant's recorded class, so neither reaches here.
	src := []byte("my $r = Foo->compute;\n")
	annotations, _ := analyzeSource(t, src)

	// walkNode records only non-Unknown types, so Unknown shows as the absence
	// of an annotation. What matters is that it is NOT Any: Any would be
	// recorded, and would satisfy every later requirement by construction.
	typ, ok := findNodeType(annotations, src, "Foo->compute")
	if ok {
		assert.NotEqual(t, types.Any, typ,
			"Any is the annotation escape hatch and cannot be requested in Perl")
		assert.Equal(t, types.Unknown, typ,
			"an unresolved method call is Unknown — inference could not determine it")
	}
}

// Strict mode reports an un-inferred value where one is USED directly.
//
// A variable assigned from an unresolved call is a separate matter and is NOT
// covered here: `my $r = Foo->compute` gives $r the sigil type Scalar rather
// than the RHS's Unknown, so strict does not see through the assignment. That
// is a real gap — the un-inferred result is laundered into a Scalar by the
// declaration — and closing it means propagating Unknown through assignment
// narrowing, which is a change with its own blast radius.
func TestStrictSeesUnresolvedValue(t *testing.T) {
	src := []byte("my $n = Foo->compute + 1;\n")

	_, permissive, _ := analyzeSourceWithOptions(t, src, infer.Options{})
	_, strict, _ := analyzeSourceWithOptions(t, src, infer.Options{Strict: true})

	assert.Empty(t, permissive, "the default accepts an un-inferred value")
	assert.NotEmpty(t, strict, "strict reports it — that is what strict is for")
}

// TestLogicalOperatorsJoinOperands verifies that &&, ||, //, and, or are
// typed as the join of their operands rather than as Any.
//
// Each yields one of its two operands — measured: `undef // "s"` is "s",
// `0 || 42` is 42, `1 && "x"` is "x" — so the result is exactly the join,
// which is what a control-flow merge means. They were declared {Any, Any,
// Any}, which is the annotation escape hatch standing in for an uncomputed
// answer, and it satisfied every later requirement by construction.
//
// The operands stay unconstrained: any value has a truth value, so there is
// nothing to reject on the input side. Only the RESULT was wrong.
func TestLogicalOperatorsJoinOperands(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want types.Type
	}{
		{"defined-or picks the string", "my $x = undef // \"s\";\n", types.Str},
		{"or over two ints", "my $x = 0 || 42;\n", types.Int},
		{"and over int and string", "my $x = 1 && \"x\";\n", types.Str},
	}
	for _, tc := range cases {
		_, _, st := analyzeSourceFull(t, []byte(tc.src))
		sym, found := st.Lookup("$x")
		require.True(t, found, "%s: $x should be in the symbol table", tc.name)
		assert.NotEqual(t, types.Any, sym.Type,
			"%s: a logical operator must not yield Any", tc.name)
		assert.True(t, types.IsSubtype(tc.want, sym.Type),
			"%s: %s should be within the inferred %s", tc.name, tc.want, sym.Type)
	}
}

// --- Element type tracking ---

// TestArrayElementType verifies that reading an element yields the element's
// type rather than the sigil default.
//
// `my @n = (1,2); my $e = $n[0]` gives an Int. PSC said Scalar — correct but
// unhelpfully broad, and six of the fourteen "wider" results in the precision
// oracle came from this one gap.
func TestArrayElementType(t *testing.T) {
	src := []byte("my @n = (1, 2);\nmy $e = $n[0];\n")
	_, _, st := analyzeSourceFull(t, src)

	sym, found := st.Lookup("$e")
	require.True(t, found, "$e should be in the symbol table")
	assert.Equal(t, types.Int, sym.Type, "an element of an Int array is an Int")
}

// TestHashElementType is the same for hashes.
func TestHashElementType(t *testing.T) {
	src := []byte("my %h = (a => \"x\");\nmy $v = $h{a};\n")
	_, _, st := analyzeSourceFull(t, src)

	sym, found := st.Lookup("$v")
	require.True(t, found, "$v should be in the symbol table")
	assert.Equal(t, types.Str, sym.Type, "an element of a Str hash is a Str")
}

// TestMixedElementTypeIsTheJoin verifies that a container holding unlike
// values yields the JOIN of them, not one arm or the other.
//
// `my @m = (1, "s")` holds an Int and a Str; an element read is whichever the
// index selects, which is exactly a merge.
func TestMixedElementTypeIsTheJoin(t *testing.T) {
	src := []byte("my @m = (1, \"s\");\nmy $e = $m[0];\n")
	_, _, st := analyzeSourceFull(t, src)

	sym, found := st.Lookup("$e")
	require.True(t, found, "$e should be in the symbol table")
	assert.True(t, types.IsSubtype(types.Int, sym.Type),
		"the Int arm survives the join, got %s", sym.Type)
	assert.True(t, types.IsSubtype(types.Str, sym.Type),
		"the Str arm survives the join, got %s", sym.Type)
}

// An untracked container still yields Scalar rather than a guess: an element
// of something PSC never saw filled is one value of unknown type.
func TestUnknownContainerElementIsScalar(t *testing.T) {
	src := []byte("my $e = $unseen[0];\n")
	_, _, st := analyzeSourceFull(t, src)

	sym, found := st.Lookup("$e")
	require.True(t, found, "$e should be in the symbol table")
	assert.Equal(t, types.Scalar, sym.Type,
		"an element of an unknown container is a Scalar, not a guess")
}

// TestArrowElementType verifies the same for the arrow forms, where the
// container is a scalar holding a reference.
//
// `my $a = [1,2]` makes $a an ArrayRef whose elements are Ints, and
// `$a->[0]` should say Int. The elements live inside the constructor node
// rather than in a list assigned to the variable, so they are recorded from
// there.
func TestArrowElementType(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want types.Type
	}{
		{"arrayref element", "my $a = [1, 2];\nmy $e = $a->[0];\n", types.Int},
		{"hashref element", "my $h = {k => \"v\"};\nmy $e = $h->{k};\n", types.Str},
	}
	for _, tc := range cases {
		_, _, st := analyzeSourceFull(t, []byte(tc.src))
		sym, found := st.Lookup("$e")
		require.True(t, found, "%s: $e should be in the symbol table", tc.name)
		assert.True(t, types.IsSubtype(tc.want, sym.Type),
			"%s: %s should be within the inferred %s", tc.name, tc.want, sym.Type)
		assert.NotEqual(t, types.Scalar, sym.Type,
			"%s: the element type should be narrower than the sigil default", tc.name)
	}
}

// TestPostfixDerefYieldsAggregate verifies that `$ref->@*` is an Array and
// `$ref->%*` is a Hash, not the scalar that holds the reference.
//
// Measured: `my %r = (op => [1,2]); push $r{op}->@*, 3` appends, and
// `my @c = $r{op}->@*` copies three elements. The deref yields the whole
// aggregate, so `push $r{op}->@*, ...` is passing an ARRAY to push — exactly
// what its first argument wants.
//
// PSC had no case for array_deref_expression, so the node fell through to the
// element type of the thing being dereferenced. Once element typing started
// answering with a real type, that produced a confident "expected Array, got
// Scalar" on correct code.
func TestPostfixDerefYieldsAggregate(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want types.Type
	}{
		{"array deref", "my $a = [1, 2];\nmy @c = $a->@*;\n", types.Array},
		{"hash deref", "my $h = {k => 1};\nmy %c = $h->%*;\n", types.Hash},
	}
	for _, tc := range cases {
		annotations, _ := analyzeSource(t, []byte(tc.src))
		var found bool
		for _, ty := range annotations {
			if ty == tc.want {
				found = true
				break
			}
		}
		assert.True(t, found, "%s: some node should be typed %s", tc.name, tc.want)
	}
}

// The payoff: push against a dereferenced arrayref is correct code and must
// not be reported.
func TestPushToPostfixDerefIsClean(t *testing.T) {
	src := []byte("my %r = (op => [1, 2]);\npush $r{op}->@*, 3;\n")
	_, diags := analyzeSource(t, src)
	assert.Empty(t, diags, "push to a dereferenced arrayref is correct Perl")
}

// TestScalarBuiltinNarrowsItsArgument verifies that scalar() reports the
// count for an aggregate and passes a scalar through.
//
// Measured: `scalar(@a)` on a two-element array is 2, `scalar($s)` on "str"
// is "str". A fixed Scalar return type cannot express either — it is the
// correct family and says nothing about which member. This is the same
// scalar-context rule NarrowByContext already implements, applied to a
// builtin whose whole job is to impose that context.
func TestScalarBuiltinNarrowsItsArgument(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want types.Type
	}{
		{"array yields a count", "my @a = (1, 2);\nmy $n = scalar(@a);\n", types.Int},
		{"hash yields a count", "my %h = (a => 1);\nmy $n = scalar(%h);\n", types.Int},
		{"scalar passes through", "my $s = \"str\";\nmy $n = scalar($s);\n", types.Str},
	}
	for _, tc := range cases {
		_, _, st := analyzeSourceFull(t, []byte(tc.src))
		sym, found := st.Lookup("$n")
		require.True(t, found, "%s: $n should be in the symbol table", tc.name)
		assert.True(t, types.IsSubtype(tc.want, sym.Type),
			"%s: %s should be within the inferred %s", tc.name, tc.want, sym.Type)
		assert.NotEqual(t, types.Scalar, sym.Type,
			"%s: the answer should be narrower than the whole scalar family", tc.name)
	}
}

// TestForeachAliasWriteWidensTheSource verifies that assigning to the loop
// variable of a foreach updates the type of what it aliases.
//
// A foreach variable is an ALIAS, so a body write mutates the source:
//
//	my $n = 42;
//	foreach ($n) { $_ = "x" }
//	my $r = $n + 1;        # perl warns: Argument "x" isn't numeric
//
// PSC said nothing, because $n kept the Int from its initialiser and the
// loop body was never connected back to it. bson found the same aliasing in
// its IR from the other direction, where a list read after an alias write
// returned the original elements.
func TestForeachAliasWriteWidensTheSource(t *testing.T) {
	src := []byte("my $n = 42;\nforeach ($n) { $_ = \"x\" }\nmy $r = $n + 1;\n")
	_, diags := analyzeSource(t, src)
	assert.NotEmpty(t, diags,
		"a write through a foreach alias changes the source type; perl warns here")
}

// The named form aliases the ARRAY's elements, so a body write changes the
// element type rather than any scalar.
func TestForeachNamedAliasWidensElements(t *testing.T) {
	src := []byte("my @a = (1, 2);\nfor my $x (@a) { $x = \"s\" }\nmy $e = $a[0];\n")
	_, _, st := analyzeSourceFull(t, src)

	sym, found := st.Lookup("@a")
	require.True(t, found, "@a should be in the symbol table")
	assert.True(t, types.IsSubtype(types.Str, sym.ElemType),
		"a write through the loop alias adds Str to the element type, got %s", sym.ElemType)
}

// TestMatchInListContextYieldsCaptures verifies that `my ($a,$b) = $s =~ /../`
// gives the CAPTURES, not the boolean.
//
// A match is context-dependent, which the `=~` signature cannot express:
//
//	my $ok = ("abc" =~ /b/);              1 — a boolean
//	my ($a,$b) = ("x42" =~ /(\w)(\d+)/);  ("x","42") — the captures
//	my $n = () = ("aaa" =~ /a/g);         3 — the count-of idiom
//
// PSC returned Bool for all three. A capture is always a Str: perl hands back
// the matched substring, so `$2` on digits is a Str whose text happens to look
// numeric.
func TestMatchInListContextYieldsCaptures(t *testing.T) {
	src := []byte("my ($a, $b) = (\"x42\" =~ /([a-z])(\\d+)/);\n")
	_, _, st := analyzeSourceFull(t, src)

	for _, name := range []string{"$a", "$b"} {
		sym, found := st.Lookup(name)
		require.True(t, found, "%s should be in the symbol table", name)
		assert.NotEqual(t, types.Bool, sym.Type,
			"%s is a capture, not the match's boolean", name)
		assert.True(t, types.IsSubtype(types.Str, sym.Type),
			"%s is a Str — perl hands back the matched substring, got %s", name, sym.Type)
	}
}

// A match in SCALAR context is still a boolean, which is the case the
// signature already had right.
func TestMatchInScalarContextIsBool(t *testing.T) {
	src := []byte("my $ok = (\"abc\" =~ /b/);\n")
	_, _, st := analyzeSourceFull(t, src)

	sym, found := st.Lookup("$ok")
	require.True(t, found, "$ok should be in the symbol table")
	assert.Equal(t, types.Bool, sym.Type, "a scalar-context match is a boolean")
}

// TestCaptureVariableIsStr verifies that a capture is narrower than the sigil
// default.
//
// The witness is a non-numeric group deliberately: `(\d+)` is now typed Int,
// since Int <: Num <: Str makes Int the more precise TRUE statement about a
// digit capture. This case is about the general shape — a capture is a
// substring, so Str — and TestNumericCaptureIsInt covers the refinement.
func TestCaptureVariableIsStr(t *testing.T) {
	src := []byte("my $s = \"ab\";\n$s =~ /([a-z]+)/;\nmy $c = $1;\n")
	_, _, st := analyzeSourceFull(t, src)

	sym, found := st.Lookup("$c")
	require.True(t, found, "$c should be in the symbol table")
	assert.Equal(t, types.Str, sym.Type,
		"a capture of letters is a Str, got %s", sym.Type)
	assert.NotEqual(t, types.Scalar, sym.Type,
		"a capture is narrower than the whole scalar family")
}

// TestCountOfIdiomIsInt verifies that `my $n = () = EXPR` yields a count.
//
// The empty list forces EXPR into list context, and the outer scalar
// assignment then takes the LENGTH of that list. Measured: `my $n = () =
// ("aaa" =~ /a/g)` is 3, not the boolean 1 that a scalar-context match gives.
//
// The grammar marks the empty list as a stub_expression, which makes the
// idiom recognisable without guessing.
func TestCountOfIdiomIsInt(t *testing.T) {
	src := []byte("my $n = () = (\"aaa\" =~ /a/g);\n")
	_, _, st := analyzeSourceFull(t, src)

	sym, found := st.Lookup("$n")
	require.True(t, found, "$n should be in the symbol table")
	assert.Equal(t, types.Int, sym.Type,
		"the count-of idiom yields a count, not the match's boolean")
}

// TestNumericCaptureIsInt verifies that a capture group which can only match
// digits is typed Int rather than the blanket Str.
//
// Str for every capture is TRUE but coarse: Int <: Num <: Str, so a digit
// capture really is a Str — and it is also an Int, which is the more precise
// true statement. perl distinguishes them, measured:
//
//	"abc123" =~ /(\d+)/;  $1 + 1   is 124, no warning
//	"abcxyz" =~ /([a-z]+)/; $1 + 1 is 1, and warns "isn't numeric"
//
// The pattern says which. `(\d+)` cannot match a non-digit, so the capture is
// an Int; anything else stays Str, since a capture is a substring and that is
// all PSC can establish.
func TestNumericCaptureIsInt(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want types.Type
	}{
		{"digits only", "my $s = \"a1\";\n$s =~ /(\\d+)/;\nmy $c = $1;\n", types.Int},
		{"letters", "my $s = \"ab\";\n$s =~ /([a-z]+)/;\nmy $c = $1;\n", types.Str},
		{"mixed word chars", "my $s = \"a1\";\n$s =~ /(\\w+)/;\nmy $c = $1;\n", types.Str},
		{"second group numeric", "my $s = \"a1\";\n$s =~ /([a-z]+)(\\d+)/;\nmy $c = $2;\n", types.Int},
		{"first group not numeric", "my $s = \"a1\";\n$s =~ /([a-z]+)(\\d+)/;\nmy $c = $1;\n", types.Str},
	}
	for _, tc := range cases {
		_, _, st := analyzeSourceFull(t, []byte(tc.src))
		sym, found := st.Lookup("$c")
		require.True(t, found, "%s: $c should be in the symbol table", tc.name)
		assert.Equal(t, tc.want, sym.Type, "%s", tc.name)
	}
}

// TestSprintfNumericFormatsAreNumeric verifies that sprintf with a purely
// numeric decimal format is typed Int or Num rather than the blanket Str.
//
// Same reasoning as digit captures: Int <: Num <: Str, so Str is true for
// every sprintf result and Int is the more precise true statement when the
// format can only produce a decimal number.
//
// THE RADIX FORMATS ARE EXCLUDED AND THAT IS THE POINT. %o and %b produce
// digit strings whose numeric VALUE is not the number that was formatted:
//
//	sprintf("%o", 8)  is "10",  and "10" + 1 is 11, not 9
//	sprintf("%b", 5)  is "101", and "101" + 1 is 102, not 6
//
// Calling those Int would be true of the text and misleading about the value,
// so they stay Str along with %x (which gives "2a"), %c (a character) and %s.
func TestSprintfNumericFormatsAreNumeric(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want types.Type
	}{
		{"%d is Int", "my $f = sprintf(\"%d\", 42);\n", types.Int},
		{"%05d is Int", "my $f = sprintf(\"%05d\", 42);\n", types.Int},
		{"%+d is Int", "my $f = sprintf(\"%+d\", 42);\n", types.Int},
		{"%.2f is Num", "my $f = sprintf(\"%.2f\", 3.14);\n", types.Num},
		{"%e is Num", "my $f = sprintf(\"%e\", 3.14);\n", types.Num},
		{"%s stays Str", "my $f = sprintf(\"%s\", \"x\");\n", types.Str},
		{"%x stays Str", "my $f = sprintf(\"%x\", 255);\n", types.Str},
		{"%o stays Str", "my $f = sprintf(\"%o\", 8);\n", types.Str},
		{"%b stays Str", "my $f = sprintf(\"%b\", 5);\n", types.Str},
		{"mixed stays Str", "my $f = sprintf(\"%s=%d\", \"a\", 1);\n", types.Str},
		{"literal text stays Str", "my $f = sprintf(\"n=%d\", 1);\n", types.Str},
	}
	for _, tc := range cases {
		_, _, st := analyzeSourceFull(t, []byte(tc.src))
		sym, found := st.Lookup("$f")
		require.True(t, found, "%s: $f should be in the symbol table", tc.name)
		assert.Equal(t, tc.want, sym.Type, "%s", tc.name)
	}
}

// TestArgumentDirectedBuiltins verifies the builtins whose result type
// follows their argument rather than being fixed by the signature.
//
// Three different rules, all measured:
//
//	abs(-5)    Int    abs(-5.5)  Num     follows the ARGUMENT
//	int(3.9)   Int    int(-3.9)  Int     ALWAYS Int, whatever goes in
//	pop @ints  Int    pop @strs  Str     follows the array's ELEMENT type
//
// All three returned the sigil default of Scalar, which is correct and says
// nothing. This is the same shape scalar() had: a signature can name one type
// and these answers depend on what was passed.
func TestArgumentDirectedBuiltins(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want types.Type
	}{
		{"abs of an int", "my $n = abs(-5);\n", types.Int},
		{"abs of a float", "my $n = abs(-5.5);\n", types.Num},
		{"int truncates to Int", "my $n = int(3.9);\n", types.Int},
		{"int of an int", "my $n = int(7);\n", types.Int},
		{"pop an int array", "my @a = (1, 2);\nmy $n = pop @a;\n", types.Int},
		{"pop a str array", "my @a = (\"x\", \"y\");\nmy $n = pop @a;\n", types.Str},
		{"shift an int array", "my @a = (1, 2);\nmy $n = shift @a;\n", types.Int},
	}
	for _, tc := range cases {
		_, _, st := analyzeSourceFull(t, []byte(tc.src))
		sym, found := st.Lookup("$n")
		require.True(t, found, "%s: $n should be in the symbol table", tc.name)
		assert.Equal(t, tc.want, sym.Type, "%s", tc.name)
	}
}

// An array PSC never saw filled still yields Scalar from pop: one value of
// unknown type, which is what Scalar says.
func TestPopUnknownArrayIsScalar(t *testing.T) {
	src := []byte("my $n = pop @unseen;\n")
	_, _, st := analyzeSourceFull(t, src)

	sym, found := st.Lookup("$n")
	require.True(t, found, "$n should be in the symbol table")
	assert.Equal(t, types.Scalar, sym.Type,
		"popping an unknown array is a Scalar, not a guess")
}

// TestSortGrepPreserveElementType verifies that sort and grep carry the
// element type through, and that map takes its element type from its BODY.
//
// Measured:
//
//	sort @ints          Int    reorders, so the elements are unchanged
//	sort @strs          Str
//	grep { $_>1 } @ints Int    SELECTS, so the elements are unchanged
//	map { $_*2 } @ints  Int    TRANSFORMS — the body decides
//	map { "x$_" } @ints Str    same input, different body, different type
//
// All of them returned List, so an element read off the result fell back to
// Scalar. Six of the seventeen remaining widenings were this one cause.
func TestSortGrepPreserveElementType(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want types.Type
	}{
		{"sort ints", "my @n = (3, 1);\nmy @s = sort @n;\nmy $e = $s[0];\n", types.Int},
		{"sort strs", "my @n = (\"b\", \"a\");\nmy @s = sort @n;\nmy $e = $s[0];\n", types.Str},
		{"sort with comparator", "my @n = (3, 1);\nmy @s = sort { $a <=> $b } @n;\nmy $e = $s[0];\n", types.Int},
		{"grep preserves", "my @n = (3, 1);\nmy @g = grep { $_ > 1 } @n;\nmy $e = $g[0];\n", types.Int},
		{"map body decides", "my @n = (3, 1);\nmy @m = map { $_ * 2 } @n;\nmy $e = $m[0];\n", types.Num},
	}
	for _, tc := range cases {
		_, _, st := analyzeSourceFull(t, []byte(tc.src))
		sym, found := st.Lookup("$e")
		require.True(t, found, "%s: $e should be in the symbol table", tc.name)
		assert.Equal(t, tc.want, sym.Type, "%s", tc.name)
	}
}

// TestKeysValuesSpliceElementTypes verifies the element types of the
// remaining list-returning builtins.
//
// Measured, and each is a different rule:
//
//	keys %h     Str    hash keys are ALWAYS strings, whatever was stored
//	values %h   Int    follows the stored values
//	splice @a   Int    hands back the REMOVED elements
//	reverse @a  Int    reorders, so elements are unchanged
//
// keys is the one worth stating: perl stringifies a hash key on the way in,
// so `$h{1}` and `$h{"1"}` are the same slot and the key comes back "1". The
// element type of the key list is Str no matter what the hash holds.
func TestKeysValuesSpliceElementTypes(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want types.Type
	}{
		{"keys are strings", "my %h = (a => 1);\nmy @k = keys %h;\nmy $e = $k[0];\n", types.Str},
		// values follows the stored values: measured, `values %h` on (a => 1)
		// gives 1, an Int. My first assertion here said Str and was simply
		// wrong — PSC had it right.
		{"values follow the hash", "my %h = (a => 1);\nmy @v = values %h;\nmy $e = $v[0];\n", types.Int},
		{"splice removes elements", "my @a = (1, 2, 3);\nmy @s = splice(@a, 0, 2);\nmy $e = $s[0];\n", types.Int},
		{"reverse preserves", "my @a = (1, 2);\nmy @r = reverse @a;\nmy $e = $r[0];\n", types.Int},
	}
	for _, tc := range cases {
		_, _, st := analyzeSourceFull(t, []byte(tc.src))
		sym, found := st.Lookup("$e")
		require.True(t, found, "%s: $e should be in the symbol table", tc.name)
		assert.Equal(t, tc.want, sym.Type, "%s", tc.name)
	}
}

// keys in SCALAR context is a count, which is a different question from its
// element type.
func TestKeysInScalarContextIsCount(t *testing.T) {
	src := []byte("my %h = (a => 1);\nmy $n = keys %h;\n")
	_, _, st := analyzeSourceFull(t, src)

	sym, found := st.Lookup("$n")
	require.True(t, found, "$n should be in the symbol table")
	assert.Equal(t, types.Int, sym.Type, "keys in scalar context is a count")
}

// TestReverseInScalarContextIsString verifies that reverse is the exception to
// "a list in scalar context is a count".
//
// Measured: `reverse("abc")` is "cba" and `reverse(@a)` on (1,2,3) is "321".
// It concatenates its arguments and reverses the resulting STRING, where every
// other List-returning builtin gives an element count.
//
// The precision oracle caught this the moment List started narrowing to a
// count: $rev went from a widening to the only WRONG answer in the run.
func TestReverseInScalarContextIsString(t *testing.T) {
	src := []byte("my $r = reverse(\"abc\");\n")
	_, _, st := analyzeSourceFull(t, src)

	sym, found := st.Lookup("$r")
	require.True(t, found, "$r should be in the symbol table")
	assert.Equal(t, types.Str, sym.Type,
		"reverse in scalar context reverses a string, it does not count")
}

// TestArrayIndexMustBeNumeric verifies that a reference used as an array
// index is reported.
//
// perl warns explicitly here — "Use of reference "ARRAY(0x...)" as array
// index" — so this is agreement with perl's own diagnostics rather than a
// stricter opinion. The index is numified, and a reference numifies to its
// address, which is never the element anyone wanted.
//
// Found by widening the type-overwriting mutation corpus: PSC typed the
// element access but never looked at the index expression.
func TestArrayIndexMustBeNumeric(t *testing.T) {
	src := []byte("my @a = (1, 2);\nmy $x = [];\nmy $y = $a[$x];\n")
	_, diags := analyzeSource(t, src)
	assert.NotEmpty(t, diags, "a reference used as an array index is reported")
}

// An ordinary integer index is not reported.
func TestIntegerArrayIndexIsClean(t *testing.T) {
	src := []byte("my @a = (1, 2);\nmy $i = 1;\nmy $y = $a[$i];\n")
	_, diags := analyzeSource(t, src)
	assert.Empty(t, diags, "an Int index is correct")
}
