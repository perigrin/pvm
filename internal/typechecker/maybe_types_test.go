// ABOUTME: Comprehensive tests for Maybe type operations and safety analysis
// ABOUTME: Tests null safety, unwrapping operations, and integration with Perl idioms

package typechecker

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/typedef"
)

func TestMaybeTypeHandler_Basic(t *testing.T) {
	// Create test dependencies
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := typedef.NewTypeHierarchy(storage)
	checker := NewTypeCheckerLegacy(hierarchy, "test")
	handler := NewMaybeTypeHandler(checker)

	// Test basic Maybe type detection
	assert.True(t, handler.IsMaybeType("Maybe[Int]"))
	assert.True(t, handler.IsMaybeType("Maybe[ArrayRef[Str]]"))
	assert.False(t, handler.IsMaybeType("Int"))
	assert.False(t, handler.IsMaybeType("ArrayRef[Int]"))
	assert.False(t, handler.IsMaybeType("Maybe")) // Not properly formatted
}

func TestMaybeTypeHandler_ExtractParameter(t *testing.T) {
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := typedef.NewTypeHierarchy(storage)
	checker := NewTypeCheckerLegacy(hierarchy, "test")
	handler := NewMaybeTypeHandler(checker)

	// Test parameter extraction
	assert.Equal(t, "Int", handler.ExtractMaybeParameter("Maybe[Int]"))
	assert.Equal(t, "ArrayRef[Str]", handler.ExtractMaybeParameter("Maybe[ArrayRef[Str]]"))
	assert.Equal(t, "", handler.ExtractMaybeParameter("Int"))
	assert.Equal(t, "", handler.ExtractMaybeParameter("Maybe"))
}

func TestMaybeTypeHandler_CreateMaybeType(t *testing.T) {
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := typedef.NewTypeHierarchy(storage)
	checker := NewTypeCheckerLegacy(hierarchy, "test")
	handler := NewMaybeTypeHandler(checker)

	// Test Maybe type creation
	assert.Equal(t, "Maybe[Int]", handler.CreateMaybeType("Int"))
	assert.Equal(t, "Maybe[ArrayRef[Str]]", handler.CreateMaybeType("ArrayRef[Str]"))

	// Test double-wrapping prevention
	assert.Equal(t, "Maybe[Int]", handler.CreateMaybeType("Maybe[Int]"))
}

func TestMaybeTypeHandler_IsNullableType(t *testing.T) {
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := typedef.NewTypeHierarchy(storage)
	checker := NewTypeCheckerLegacy(hierarchy, "test")
	handler := NewMaybeTypeHandler(checker)

	// Test nullable type detection
	assert.True(t, handler.IsNullableType("Undef"))
	assert.True(t, handler.IsNullableType("Maybe[Int]"))
	assert.True(t, handler.IsNullableType("Int|Undef"))
	assert.True(t, handler.IsNullableType("Str|Int|Undef"))
	assert.False(t, handler.IsNullableType("Int"))
	assert.False(t, handler.IsNullableType("Str|Int"))
}

func TestMaybeTypeHandler_CheckUnsafeAccess(t *testing.T) {
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := typedef.NewTypeHierarchy(storage)
	checker := NewTypeCheckerLegacy(hierarchy, "test")
	handler := NewMaybeTypeHandler(checker)

	pos := ast.Position{Line: 10, Column: 5}

	// Test unsafe access detection
	err = handler.CheckUnsafeAccess("myVar", "Maybe[Int]", "direct", pos)
	assert.Error(t, err) // Should warn about unsafe access
	assert.Contains(t, err.Error(), "Unsafe access to Maybe type")

	// Check that the access was recorded
	accesses := handler.GetUnsafeAccesses()
	require.Len(t, accesses, 1)
	assert.Equal(t, "myVar", accesses[0].Variable)
	assert.Equal(t, "Maybe[Int]", accesses[0].MaybeType)
	assert.Equal(t, "direct", accesses[0].AccessType)

	// Test non-Maybe types don't trigger warnings
	err = handler.CheckUnsafeAccess("myVar", "Int", "direct", pos)
	assert.NoError(t, err)
}

func TestMaybeTypeHandler_CheckSafeUnwrapping(t *testing.T) {
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := typedef.NewTypeHierarchy(storage)
	checker := NewTypeCheckerLegacy(hierarchy, "test")
	handler := NewMaybeTypeHandler(checker)

	pos := ast.Position{Line: 10, Column: 5}

	// Test unsafe unwrapping
	resultType, err := handler.CheckSafeUnwrapping("myVar", "Maybe[Int]", pos, nil)
	assert.Error(t, err)                      // Should warn about unsafe unwrapping
	assert.Equal(t, "Maybe[Int]", resultType) // Type should remain wrapped

	// Test safe unwrapping with context
	safeContext := &MaybeFlowContext{
		Conditions: []MaybeCondition{
			{Type: "defined", Variable: "myVar", Positive: true},
		},
		RefinedTypes:  map[string]string{},
		SafeVariables: map[string]bool{},
	}

	resultType, err = handler.CheckSafeUnwrapping("myVar", "Maybe[Int]", pos, safeContext)
	assert.NoError(t, err)             // Should be safe with defined() check
	assert.Equal(t, "Int", resultType) // Type should be unwrapped

	// Test non-Maybe types pass through
	resultType, err = handler.CheckSafeUnwrapping("myVar", "Int", pos, nil)
	assert.NoError(t, err)
	assert.Equal(t, "Int", resultType)
}

func TestMaybeTypeHandler_InferMaybeType(t *testing.T) {
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := typedef.NewTypeHierarchy(storage)
	checker := NewTypeCheckerLegacy(hierarchy, "test")
	handler := NewMaybeTypeHandler(checker)

	context := &MaybeInferenceContext{
		VariableTypes:   map[string]string{},
		FunctionReturns: map[string]string{},
		CurrentScope:    "main",
	}

	// Test patterns that should return Maybe types
	tests := []struct {
		expression string
		expected   string
	}{
		{"shift @_", "Maybe[Scalar]"},
		{"pop @array", "Maybe[Scalar]"},
		{"$hash{$key}", "Maybe[Scalar]"},
		{"$array[0]", "Maybe[Scalar]"},
		{"split /,/, $string", "Maybe[ArrayRef]"},
		{"readline $fh", "Maybe[Scalar]"},
	}

	for _, test := range tests {
		result := handler.InferMaybeType(test.expression, context)
		assert.Equal(t, test.expected, result, "Expression: %s", test.expression)
	}

	// Test expressions that shouldn't be Maybe types
	nonMaybeExpressions := []string{
		"42",
		"'hello'",
		"1 + 2",
	}

	for _, expr := range nonMaybeExpressions {
		result := handler.InferMaybeType(expr, context)
		assert.Empty(t, result, "Expression should not be Maybe: %s", expr)
	}
}

func TestMaybeTypeHandler_PerlIdiomIntegration(t *testing.T) {
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := typedef.NewTypeHierarchy(storage)
	checker := NewTypeCheckerLegacy(hierarchy, "test")

	// Set up some Maybe type variables
	checker.VariableTypes["maybe_var"] = "Maybe[Str]"
	checker.VariableTypes["regular_var"] = "Str"

	handler := NewMaybeTypeHandler(checker)
	pos := ast.Position{Line: 15, Column: 10}

	// Test defined-or operator integration
	err = handler.CheckPerlIdiomIntegration("$result = $maybe_var // 'default'", pos)
	assert.NoError(t, err)
	assert.Equal(t, DefinedOrOperator, handler.PerlIdiomIntegrations["$result = $maybe_var // 'default'"])

	// Test logical-or assignment
	err = handler.CheckPerlIdiomIntegration("$maybe_var ||= 'default'", pos)
	assert.NoError(t, err)
	assert.Equal(t, LogicalOrAssign, handler.PerlIdiomIntegrations["$maybe_var ||= 'default'"])

	// Test exists() check
	err = handler.CheckPerlIdiomIntegration("if (exists($hash{$key}))", pos)
	assert.NoError(t, err)
	assert.Equal(t, ExistsCheck, handler.PerlIdiomIntegrations["if (exists($hash{$key}))"])

	// Test defined() check
	err = handler.CheckPerlIdiomIntegration("if (defined($maybe_var))", pos)
	assert.NoError(t, err)
	assert.Equal(t, DefinedCheck, handler.PerlIdiomIntegrations["if (defined($maybe_var))"])
}

func TestMaybeTypeHandler_SafetyReport(t *testing.T) {
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := typedef.NewTypeHierarchy(storage)
	checker := NewTypeCheckerLegacy(hierarchy, "test")
	checker.VariableTypes["maybe_var"] = "Maybe[Str]"

	handler := NewMaybeTypeHandler(checker)
	pos := ast.Position{Line: 20, Column: 8}

	// Generate some unsafe accesses
	_ = handler.CheckUnsafeAccess("maybe_var", "Maybe[Str]", "direct", pos)
	_ = handler.CheckUnsafeAccess("maybe_var2", "Maybe[Int]", "method_call", pos)

	// Add some Perl idiom usage
	_ = handler.CheckPerlIdiomIntegration("$maybe_var // 'default'", pos)

	// Generate safety report
	report := handler.GenerateSafetyReport()

	// Verify report contents
	assert.Equal(t, 2, report.UnsafeAccesses)
	assert.Equal(t, 0, report.WrapWarnings)
	assert.Equal(t, 1, report.PerlIdiomUsage)
	assert.Greater(t, report.SafetyScore, 0.0)
	assert.Less(t, report.SafetyScore, 100.0)

	// Check recommendations
	assert.NotEmpty(t, report.Recommendations)
	assert.Contains(t, report.Recommendations[0], "defined() checks")

	// Verify details are included
	assert.Len(t, report.UnsafeAccessDetails, 2)
	assert.Equal(t, "maybe_var", report.UnsafeAccessDetails[0].Variable)
	assert.Equal(t, "maybe_var2", report.UnsafeAccessDetails[1].Variable)
}

func TestMaybeTypeHandler_SafetyScore(t *testing.T) {
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := typedef.NewTypeHierarchy(storage)
	checker := NewTypeCheckerLegacy(hierarchy, "test")
	checker.VariableTypes["maybe_var"] = "Maybe[Str]"

	// Test perfect score (no issues, good practices)
	handler := NewMaybeTypeHandler(checker)
	pos := ast.Position{Line: 25, Column: 12}
	_ = handler.CheckPerlIdiomIntegration("$maybe_var // 'default'", pos)

	report := handler.GenerateSafetyReport()
	assert.Equal(t, 100.0, report.SafetyScore)

	// Test good score (no issues, no practices)
	handler2 := NewMaybeTypeHandler(checker)
	report2 := handler2.GenerateSafetyReport()
	assert.Equal(t, 90.0, report2.SafetyScore)

	// Test mixed score (some issues, some practices)
	handler3 := NewMaybeTypeHandler(checker)
	_ = handler3.CheckUnsafeAccess("maybe_var", "Maybe[Str]", "direct", pos)
	_ = handler3.CheckPerlIdiomIntegration("$maybe_var // 'default'", pos)

	report3 := handler3.GenerateSafetyReport()
	assert.Greater(t, report3.SafetyScore, 0.0)
	assert.Less(t, report3.SafetyScore, 100.0)
}

func TestMaybeTypeHandler_FlowSensitiveAnalysis(t *testing.T) {
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := typedef.NewTypeHierarchy(storage)
	checker := NewTypeCheckerLegacy(hierarchy, "test")
	handler := NewMaybeTypeHandler(checker)

	// Test different safe contexts
	testCases := []struct {
		name       string
		context    *MaybeFlowContext
		expectSafe bool
	}{
		{
			name: "defined() check makes it safe",
			context: &MaybeFlowContext{
				Conditions: []MaybeCondition{
					{Type: "defined", Variable: "myVar", Positive: true},
				},
				RefinedTypes:  map[string]string{},
				SafeVariables: map[string]bool{},
			},
			expectSafe: true,
		},
		{
			name: "not_undef check makes it safe",
			context: &MaybeFlowContext{
				Conditions: []MaybeCondition{
					{Type: "not_undef", Variable: "myVar", Positive: true},
				},
				RefinedTypes:  map[string]string{},
				SafeVariables: map[string]bool{},
			},
			expectSafe: true,
		},
		{
			name: "refined type makes it safe",
			context: &MaybeFlowContext{
				Conditions: []MaybeCondition{},
				RefinedTypes: map[string]string{
					"myVar": "Int", // Not a Maybe type
				},
				SafeVariables: map[string]bool{},
			},
			expectSafe: true,
		},
		{
			name: "no safe context",
			context: &MaybeFlowContext{
				Conditions:    []MaybeCondition{},
				RefinedTypes:  map[string]string{},
				SafeVariables: map[string]bool{},
			},
			expectSafe: false,
		},
		{
			name: "negative condition is not safe",
			context: &MaybeFlowContext{
				Conditions: []MaybeCondition{
					{Type: "defined", Variable: "myVar", Positive: false},
				},
				RefinedTypes:  map[string]string{},
				SafeVariables: map[string]bool{},
			},
			expectSafe: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isSafe := handler.isInSafeContext("myVar", tc.context)
			assert.Equal(t, tc.expectSafe, isSafe)
		})
	}
}

func TestMaybeTypeHandler_Integration(t *testing.T) {
	// This test simulates a complete flow with a TypeChecker
	storage, err := typedef.NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := typedef.NewTypeHierarchy(storage)
	checker := NewTypeCheckerLegacy(hierarchy, "test")

	// Set up variable types
	checker.VariableTypes["param"] = "Maybe[Str]"
	checker.VariableTypes["result"] = "Str"

	handler := NewMaybeTypeHandler(checker)
	pos := ast.Position{Line: 30, Column: 15}

	// Simulate unsafe access
	_ = handler.CheckUnsafeAccess("param", "Maybe[Str]", "direct", pos)

	// Simulate good Perl idiom usage
	_ = handler.CheckPerlIdiomIntegration("$result = $param // 'default'", pos)

	// Generate comprehensive report
	report := handler.GenerateSafetyReport()

	// Verify the integration works end-to-end
	assert.Equal(t, 1, report.UnsafeAccesses)
	assert.Equal(t, 1, report.PerlIdiomUsage)
	assert.NotEmpty(t, report.Recommendations)
	assert.Greater(t, report.SafetyScore, 0.0)

	// Verify specific recommendations
	foundDefineCheckRec := false
	foundDefinedOrRec := false
	for _, rec := range report.Recommendations {
		if strings.Contains(rec, "defined() checks") {
			foundDefineCheckRec = true
		}
		if strings.Contains(rec, "// operator") {
			foundDefinedOrRec = true
		}
	}
	assert.True(t, foundDefineCheckRec, "Should recommend defined() checks")
	assert.True(t, foundDefinedOrRec, "Should recommend // operator usage")
}
