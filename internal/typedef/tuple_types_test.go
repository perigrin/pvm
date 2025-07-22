// ABOUTME: Test tuple type support in typedef system
// ABOUTME: Validates tuple type validation and hierarchy integration

package typedef

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTupleTypeSupport(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := NewTypeHierarchy(storage)

	tests := []struct {
		name      string
		tupleType string
		params    []string
		expectErr bool
	}{
		{
			name:      "Basic two-element tuple",
			tupleType: "Tuple",
			params:    []string{"Int", "Str"},
			expectErr: false,
		},
		{
			name:      "Three-element tuple",
			tupleType: "Tuple",
			params:    []string{"Int", "Int", "Str"},
			expectErr: false,
		},
		{
			name:      "Complex nested tuple",
			tupleType: "Tuple",
			params:    []string{"Tuple[Int, Str]", "Bool"},
			expectErr: false,
		},
		{
			name:      "Tuple with parameterized types",
			tupleType: "Tuple",
			params:    []string{"ArrayRef[Int]", "HashRef[Str]"},
			expectErr: false,
		},
		{
			name:      "Single element tuple - should fail",
			tupleType: "Tuple",
			params:    []string{"Int"},
			expectErr: true,
		},
		{
			name:      "Empty tuple - should fail",
			tupleType: "Tuple",
			params:    []string{},
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := hierarchy.CreateParameterizedType(test.tupleType, test.params)

			if test.expectErr {
				assert.Error(t, err, "Expected error for invalid tuple: %v", test.params)
			} else {
				assert.NoError(t, err, "Should not error for valid tuple: %v", test.params)
				assert.NotEmpty(t, result, "Result should not be empty")

				// Check that result has proper tuple format
				expectedFormat := "Tuple[" + joinWithCommas(test.params) + "]"
				assert.Equal(t, expectedFormat, result, "Tuple format should match expected")

				// Validate the created tuple type
				err = hierarchy.ValidateType(result)
				assert.NoError(t, err, "Tuple type should be valid: %s", result)
			}
		})
	}
}

func TestTupleTypeHierarchyIntegration(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := NewTypeHierarchy(storage)

	// Test that Tuple is recognized as a parameterized type
	assert.True(t, hierarchy.IsBuiltinType("Tuple[Int, Str]"), "Tuple should be recognized as builtin parameterized type")

	// Test tuple type validation
	tupleType := "Tuple[Int, Str, Bool]"
	err = hierarchy.ValidateType(tupleType)
	assert.NoError(t, err, "Complex tuple type should validate successfully")

	// Test nested tuple validation
	nestedTuple := "Tuple[Tuple[Int, Int], Str]"
	err = hierarchy.ValidateType(nestedTuple)
	assert.NoError(t, err, "Nested tuple type should validate successfully")
}

func TestTupleTypeCompatibility(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := NewTypeHierarchy(storage)

	tests := []struct {
		name        string
		sourceType  string
		targetType  string
		shouldMatch bool
	}{
		{
			name:        "Same tuple types",
			sourceType:  "Tuple[Int, Str]",
			targetType:  "Tuple[Int, Str]",
			shouldMatch: true,
		},
		{
			name:        "Different tuple types",
			sourceType:  "Tuple[Int, Str]",
			targetType:  "Tuple[Str, Int]",
			shouldMatch: false,
		},
		{
			name:        "Different tuple lengths",
			sourceType:  "Tuple[Int, Str]",
			targetType:  "Tuple[Int, Str, Bool]",
			shouldMatch: false,
		},
		{
			name:        "Nested tuple compatibility",
			sourceType:  "Tuple[Tuple[Int, Int], Str]",
			targetType:  "Tuple[Tuple[Int, Int], Str]",
			shouldMatch: true,
		},
		{
			name:        "Tuple to Any",
			sourceType:  "Tuple[Int, Str]",
			targetType:  "Any",
			shouldMatch: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := hierarchy.CheckTypeCompatibility(test.sourceType, test.targetType)

			if test.shouldMatch {
				assert.NoError(t, err, "Types should be compatible: %s -> %s", test.sourceType, test.targetType)
			} else {
				assert.Error(t, err, "Types should not be compatible: %s -> %s", test.sourceType, test.targetType)
			}
		})
	}
}

func TestTupleTypeErrorMessages(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := NewTypeHierarchy(storage)

	tests := []struct {
		name          string
		tupleParams   []string
		expectedError string
	}{
		{
			name:          "Single parameter error",
			tupleParams:   []string{"Int"},
			expectedError: "Tuple requires at least two type parameters",
		},
		{
			name:          "Empty parameters error",
			tupleParams:   []string{},
			expectedError: "Tuple requires at least two type parameters",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := hierarchy.CreateParameterizedType("Tuple", test.tupleParams)
			assert.Error(t, err, "Should error for invalid tuple parameters")
			assert.Contains(t, err.Error(), test.expectedError, "Error message should contain expected text")
		})
	}
}

func TestTupleTypeWithInvalidMemberTypes(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := NewTypeHierarchy(storage)

	tests := []struct {
		name        string
		tupleType   string
		expectValid bool
	}{
		{
			name:        "Valid tuple with builtin types",
			tupleType:   "Tuple[Int, Str]",
			expectValid: true,
		},
		{
			name:        "Tuple with malformed member type",
			tupleType:   "Tuple[invalid_type, Str]", // lowercase type name is invalid
			expectValid: false,
		},
		{
			name:        "Tuple with malformed parameterized member",
			tupleType:   "Tuple[ArrayRef[], Str]",
			expectValid: false,
		},
		{
			name:        "Valid tuple with union member",
			tupleType:   "Tuple[Int|Str, Bool]",
			expectValid: true,
		},
		{
			name:        "Valid tuple with intersection member",
			tupleType:   "Tuple[Serializable&Comparable, Str]",
			expectValid: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := hierarchy.ValidateType(test.tupleType)

			if test.expectValid {
				assert.NoError(t, err, "Tuple type should be valid: %s", test.tupleType)
			} else {
				assert.Error(t, err, "Tuple type should be invalid: %s", test.tupleType)
			}
		})
	}
}

// Helper function to join strings with commas and spaces
func joinWithCommas(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += ", " + strs[i]
	}
	return result
}
