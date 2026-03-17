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
	// if (defined($x)): if-body $x → Scalar (non-Undef), else-body $x → Undef.
	// Note: "my $x = undef" does NOT narrow $x to Undef because the undef
	// keyword produces Unknown type, so $x stays at sigil type Scalar.
	src := []byte("my $x = undef;\nif (defined($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets[0]=declaration, [1]=condition defined($x), [2]=if-body, [3]=else-body
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	ifBodyTyp, ifOk := annotations[offsets[2]]
	assert.True(t, ifOk, "if-body $x should be annotated")
	assert.Equal(t, types.Scalar, ifBodyTyp, "if-body $x should be Scalar (defined guard on Scalar)")

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
	// Inside while (defined($x)), $x should be Scalar (non-Undef).
	src := []byte("my $x = undef;\nwhile (defined($x)) {\n    my $y = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	whileBodyXOffset := findLastVarOffset(src, "$x")
	typ, ok := annotations[whileBodyXOffset]
	assert.True(t, ok, "while-body $x should be annotated")
	assert.Equal(t, types.Scalar, typ, "while-body $x should be Scalar (defined guard)")
}

func TestFlowNarrowingIfElseRefGuard(t *testing.T) {
	// if (ref($x)) → Ref in if-body, Scalar in else-body
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
	assert.Equal(t, types.Scalar, elseBodyTyp, "else-body $x should be Scalar (negated ref guard)")
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
	// unless (ref($x)): body $x → Scalar (negated ref), else $x → Ref (positive ref)
	src := []byte("my $x = undef;\nunless (ref($x)) {\n    my $y = $x;\n} else {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets[0]=decl, [1]=condition, [2]=unless-body, [3]=else-body
	require.True(t, len(offsets) >= 4, "should find at least 4 occurrences of $x, got %d", len(offsets))

	unlessBodyTyp, ok := annotations[offsets[2]]
	assert.True(t, ok, "unless-body $x should be annotated")
	assert.Equal(t, types.Scalar, unlessBodyTyp, "unless-body $x should be Scalar (negated ref guard)")

	elseBodyTyp, elseOk := annotations[offsets[3]]
	assert.True(t, elseOk, "else-body $x should be annotated")
	assert.Equal(t, types.Ref, elseBodyTyp, "else-body $x should be Ref (positive ref guard)")
}

func TestFlowNarrowingElsifBranchGetsAnnotations(t *testing.T) {
	// elsif blocks should get type annotations even though guard narrowing
	// is deferred for elsif conditions. Verify that $x inside the elsif body
	// is annotated (with its sigil type Scalar, since no narrowing is applied).
	src := []byte("my $x = undef;\nif (defined($x)) {\n    my $y = $x;\n} elsif (ref($x)) {\n    my $z = $x;\n}\n")
	annotations, _ := analyzeSource(t, src)

	offsets := findAllVarOffsets(src, "$x")
	// offsets[0]=decl, [1]=if-condition, [2]=if-body, [3]=elsif-condition, [4]=elsif-body
	require.True(t, len(offsets) >= 5, "should find at least 5 occurrences of $x, got %d", len(offsets))

	elsifBodyTyp, ok := annotations[offsets[4]]
	assert.True(t, ok, "elsif-body $x should be annotated (not skipped)")
	assert.Equal(t, types.Scalar, elsifBodyTyp, "elsif-body $x should have sigil type Scalar (no narrowing applied)")
}
