// ABOUTME: Comprehensive tests for intersection type validation and contradiction detection
// ABOUTME: Tests semantic validation, conflict resolution, and intersection simplification

package typedef

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntersectionValidator_DisjointTypes(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := NewTypeHierarchy(storage)
	validator := NewIntersectionValidator(hierarchy)

	tests := []struct {
		name                   string
		intersectionType       string
		expectedContradictions int
		expectedSeverity       ContradictionSeverity
		expectedType           ContradictionType
	}{
		{
			name:                   "Int and Str are disjoint",
			intersectionType:       "Int&Str",
			expectedContradictions: 1,
			expectedSeverity:       Error,
			expectedType:           DisjointTypes,
		},
		{
			name:                   "Multiple scalar types",
			intersectionType:       "Int&Bool&Undef",
			expectedContradictions: 1,
			expectedSeverity:       Error,
			expectedType:           DisjointTypes,
		},
		{
			name:                   "Reference types conflict",
			intersectionType:       "ArrayRef&HashRef",
			expectedContradictions: 1,
			expectedSeverity:       Error,
			expectedType:           DisjointTypes,
		},
		{
			name:                   "Parameterized format",
			intersectionType:       "Intersection[Int,Str]",
			expectedContradictions: 1,
			expectedSeverity:       Error,
			expectedType:           DisjointTypes,
		},
		{
			name:                   "Compatible types",
			intersectionType:       "Serializable&Object",
			expectedContradictions: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contradictions, err := validator.ValidateIntersection(test.intersectionType)
			require.NoError(t, err)

			if test.expectedContradictions == 0 {
				assert.Empty(t, contradictions)
				return
			}

			assert.Len(t, contradictions, test.expectedContradictions)

			disjointContradictions := validator.GetContradictionsByType(contradictions, DisjointTypes)
			if test.expectedType == DisjointTypes {
				assert.NotEmpty(t, disjointContradictions)
				assert.Equal(t, test.expectedSeverity, disjointContradictions[0].Severity)
				assert.Contains(t, disjointContradictions[0].Explanation, "mutually exclusive")
				assert.Contains(t, disjointContradictions[0].Suggestion, "union type")
			}
		})
	}
}

func TestIntersectionValidator_SubtypeRedundancy(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := NewTypeHierarchy(storage)
	validator := NewIntersectionValidator(hierarchy)

	tests := []struct {
		name             string
		intersectionType string
		expectRedundancy bool
	}{
		{
			name:             "Int is subtype of Num",
			intersectionType: "Int&Num",
			expectRedundancy: true,
		},
		{
			name:             "Str is subtype of Scalar",
			intersectionType: "Str&Scalar",
			expectRedundancy: true,
		},
		{
			name:             "No redundancy in compatible types",
			intersectionType: "Serializable&Comparable",
			expectRedundancy: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contradictions, err := validator.ValidateIntersection(test.intersectionType)
			require.NoError(t, err)

			redundancyContradictions := validator.GetContradictionsByType(contradictions, SubtypeRedundancy)

			if test.expectRedundancy {
				assert.NotEmpty(t, redundancyContradictions)
				assert.Equal(t, Info, redundancyContradictions[0].Severity)
				assert.Contains(t, redundancyContradictions[0].Explanation, "subtype")
				assert.Contains(t, redundancyContradictions[0].Suggestion, "more specific type")
			} else {
				assert.Empty(t, redundancyContradictions)
			}
		})
	}
}

func TestIntersectionValidator_StructuralMismatch(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := NewTypeHierarchy(storage)
	validator := NewIntersectionValidator(hierarchy)

	tests := []struct {
		name             string
		intersectionType string
		expectMismatch   bool
	}{
		{
			name:             "Conflicting parameterizations",
			intersectionType: "ArrayRef[Int]&ArrayRef[Str]",
			expectMismatch:   true,
		},
		{
			name:             "Different parameterized types",
			intersectionType: "ArrayRef[Int]&HashRef[Str]",
			expectMismatch:   false,
		},
		{
			name:             "Compatible parameterized types",
			intersectionType: "Container[Int]&Serializable",
			expectMismatch:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contradictions, err := validator.ValidateIntersection(test.intersectionType)
			require.NoError(t, err)

			mismatchContradictions := validator.GetContradictionsByType(contradictions, StructuralMismatch)

			if test.expectMismatch {
				assert.NotEmpty(t, mismatchContradictions)
				assert.Equal(t, Error, mismatchContradictions[0].Severity)
				assert.Contains(t, mismatchContradictions[0].Explanation, "conflicting parameterizations")
			} else {
				assert.Empty(t, mismatchContradictions)
			}
		})
	}
}

func TestIntersectionValidator_ValueContradiction(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := NewTypeHierarchy(storage)
	validator := NewIntersectionValidator(hierarchy)

	tests := []struct {
		name                string
		intersectionType    string
		expectContradiction bool
	}{
		{
			name:                "Multiple literal values",
			intersectionType:    "42&'hello'",
			expectContradiction: true,
		},
		{
			name:                "String literals",
			intersectionType:    "'hello'&\"world\"",
			expectContradiction: true,
		},
		{
			name:                "Single literal with type",
			intersectionType:    "42&Int",
			expectContradiction: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contradictions, err := validator.ValidateIntersection(test.intersectionType)
			require.NoError(t, err)

			valueContradictions := validator.GetContradictionsByType(contradictions, ValueContradiction)

			if test.expectContradiction {
				assert.NotEmpty(t, valueContradictions)
				assert.Equal(t, Error, valueContradictions[0].Severity)
				assert.Contains(t, valueContradictions[0].Explanation, "cannot be satisfied simultaneously")
				assert.Contains(t, valueContradictions[0].Suggestion, "union type")
			} else {
				assert.Empty(t, valueContradictions)
			}
		})
	}
}

func TestIntersectionValidator_ParseIntersectionMembers(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := NewTypeHierarchy(storage)
	validator := NewIntersectionValidator(hierarchy)

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple infix format",
			input:    "Int&Str",
			expected: []string{"Int", "Str"},
		},
		{
			name:     "Multiple types",
			input:    "A&B&C",
			expected: []string{"A", "B", "C"},
		},
		{
			name:     "Parameterized format",
			input:    "Intersection[Int,Str,Bool]",
			expected: []string{"Int", "Str", "Bool"},
		},
		{
			name:     "Complex parameterized types",
			input:    "ArrayRef[Int]&HashRef[Str,Int]",
			expected: []string{"ArrayRef[Int]", "HashRef[Str,Int]"},
		},
		{
			name:     "Nested brackets",
			input:    "Container[ArrayRef[Int]]&Serializable",
			expected: []string{"Container[ArrayRef[Int]]", "Serializable"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			members, err := validator.parseIntersectionMembers(test.input)
			require.NoError(t, err)
			assert.Equal(t, test.expected, members)
		})
	}
}

func TestIntersectionValidator_SimplifyIntersection(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := NewTypeHierarchy(storage)
	validator := NewIntersectionValidator(hierarchy)

	tests := []struct {
		name          string
		input         string
		expectedType  string
		expectedInfos int // Number of Info-level contradictions
	}{
		{
			name:          "Remove subtype redundancy",
			input:         "Int&Num",
			expectedType:  "Int", // More specific type
			expectedInfos: 1,
		},
		{
			name:          "Keep compatible types",
			input:         "Serializable&Comparable",
			expectedType:  "Comparable&Serializable", // Sorted
			expectedInfos: 0,
		},
		{
			name:          "Single type",
			input:         "Int",
			expectedType:  "Int",
			expectedInfos: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			simplified, contradictions, err := validator.SimplifyIntersection(test.input)
			require.NoError(t, err)

			assert.Equal(t, test.expectedType, simplified)

			infoContradictions := validator.GetContradictionsBySeverity(contradictions, Info)
			assert.Len(t, infoContradictions, test.expectedInfos)
		})
	}
}

func TestIntersectionValidator_FormatContradictions(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := NewTypeHierarchy(storage)
	validator := NewIntersectionValidator(hierarchy)

	contradictions := []IntersectionContradiction{
		{
			Type:        DisjointTypes,
			Members:     []string{"Int", "Str"},
			Explanation: "Types Int and Str are mutually exclusive",
			Severity:    Error,
			Suggestion:  "Use a union type instead",
		},
		{
			Type:        SubtypeRedundancy,
			Members:     []string{"Int", "Num"},
			Explanation: "Int is a subtype of Num",
			Severity:    Info,
			Suggestion:  "Use Int alone",
		},
	}

	formatted := validator.FormatContradictions(contradictions)
	require.Len(t, formatted, 2)

	assert.Contains(t, formatted[0], "[ERROR]")
	assert.Contains(t, formatted[0], "mutually exclusive")
	assert.Contains(t, formatted[0], "Use a union type")

	assert.Contains(t, formatted[1], "[INFO]")
	assert.Contains(t, formatted[1], "subtype")
	assert.Contains(t, formatted[1], "Use Int alone")
}

func TestIntersectionValidator_ComplexIntersections(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := NewTypeHierarchy(storage)
	validator := NewIntersectionValidator(hierarchy)

	tests := []struct {
		name             string
		intersectionType string
		expectedErrors   int
		expectedWarnings int
		expectedInfos    int
	}{
		{
			name:             "Multiple contradictions",
			intersectionType: "Int&Str&Num", // Disjoint + redundancy
			expectedErrors:   1,             // Int&Str are disjoint
			expectedInfos:    0,             // Int&Num redundancy won't be detected due to disjoint error
		},
		{
			name:             "Complex parameterized intersection",
			intersectionType: "ArrayRef[Int]&ArrayRef[Str]&Serializable",
			expectedErrors:   1, // Conflicting parameterizations
			expectedInfos:    0,
		},
		{
			name:             "Valid complex intersection",
			intersectionType: "Serializable&Comparable&Cloneable",
			expectedErrors:   0,
			expectedInfos:    0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contradictions, err := validator.ValidateIntersection(test.intersectionType)
			require.NoError(t, err)

			errorContradictions := validator.GetContradictionsBySeverity(contradictions, Error)
			warningContradictions := validator.GetContradictionsBySeverity(contradictions, Warning)
			infoContradictions := validator.GetContradictionsBySeverity(contradictions, Info)

			assert.Len(t, errorContradictions, test.expectedErrors)
			assert.Len(t, warningContradictions, test.expectedWarnings)
			assert.Len(t, infoContradictions, test.expectedInfos)
		})
	}
}

func TestIntersectionValidator_EdgeCases(t *testing.T) {
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := NewTypeHierarchy(storage)
	validator := NewIntersectionValidator(hierarchy)

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "Empty intersection",
			input:       "",
			expectError: false,
		},
		{
			name:        "Single type",
			input:       "Int",
			expectError: false,
		},
		{
			name:        "Malformed parameterized type",
			input:       "Intersection[Int,Str",
			expectError: true,
		},
		{
			name:        "Complex nesting",
			input:       "Container[ArrayRef[HashRef[Str,Int]]]&Serializable",
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := validator.ValidateIntersection(test.input)

			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIntersectionValidator_Integration(t *testing.T) {
	// Test integration with the actual type hierarchy
	storage, err := NewStorageWithPath(t.TempDir())
	require.NoError(t, err)

	hierarchy := NewTypeHierarchy(storage)
	validator := NewIntersectionValidator(hierarchy)

	// Test with types that should be in the hierarchy
	contradictions, err := validator.ValidateIntersection("Int&Str")
	require.NoError(t, err)
	require.NotEmpty(t, contradictions)

	// Should detect Int and Str as disjoint
	disjointContradictions := validator.GetContradictionsByType(contradictions, DisjointTypes)
	assert.NotEmpty(t, disjointContradictions)
	assert.Equal(t, Error, disjointContradictions[0].Severity)

	// Test simplification
	simplified, _, err := validator.SimplifyIntersection("Int&Str")
	require.NoError(t, err)

	// Should remain the same since it's an error case
	assert.Equal(t, "Int&Str", simplified)
}
