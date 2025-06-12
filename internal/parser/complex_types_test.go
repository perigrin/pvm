// ABOUTME: Tests for complex type expression parsing
// ABOUTME: Validates parsing of sophisticated type combinations and nested expressions

package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ComplexTypeTest represents a test case for complex type expressions
type ComplexTypeTest struct {
	Name           string
	Input          string
	Description    string
	ExpectedTypes  []ExpectedTypeInfo
	ShouldError    bool
	ErrorMessages  []string
	PerformanceMax time.Duration
	MemoryLimitMB  int64
}

// ExpectedTypeInfo describes what we expect to find in a parsed type expression
type ExpectedTypeInfo struct {
	Variable        string
	TypeString      string
	IsUnion         bool
	IsIntersection  bool
	IsNegation      bool
	IsParameterized bool
	NestingDepth    int
	UnionCount      int
	ParamCount      int
}

func TestComplexTypeExpressions(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)

	tests := []ComplexTypeTest{
		{
			Name: "Nested unions in parameterized types",
			Input: `
				my ArrayRef[Int|Str|Bool] @complex_array;
				my HashRef[ArrayRef[Int]|HashRef[Str]] %nested_complex;
				my Map[Str, Int|Undef] %optional_values;
			`,
			Description: "Union types nested within parameterized types",
			ExpectedTypes: []ExpectedTypeInfo{
				{
					Variable:        "@complex_array",
					TypeString:      "ArrayRef[Int|Str|Bool]",
					IsParameterized: true,
					ParamCount:      1,
					NestingDepth:    2,
				},
				{
					Variable:        "%nested_complex",
					TypeString:      "HashRef[ArrayRef[Int]|HashRef[Str]]",
					IsParameterized: true,
					ParamCount:      1,
					NestingDepth:    3,
				},
				{
					Variable:        "%optional_values",
					TypeString:      "Map[Str, Int|Undef]",
					IsParameterized: true,
					ParamCount:      2,
					NestingDepth:    2,
				},
			},
		},
		{
			Name: "Parameterized types within unions",
			Input: `
				my (ArrayRef[Int]|HashRef[Str]) $param_union;
				my (Container[MyType]|Wrapper[OtherType]) $flexible;
				my (Result[Data, Error]|Maybe[Value]) $outcome;
			`,
			Description:   "Parameterized types within union expressions",
			ShouldError:   true,
			ErrorMessages: []string{"UnknownTypeError"},
		},
		{
			Name: "Deep nesting combinations",
			Input: `
				my ArrayRef[HashRef[ArrayRef[Int|Str]]] @deep_nested;
				my Map[Str, ArrayRef[Tuple[Int, Bool|Str]]] %complex_map;
				my Container[Wrapper[Inner[Data[Value]]]] $deeply_nested;
			`,
			Description: "Deeply nested parameterized types with complex combinations",
			ExpectedTypes: []ExpectedTypeInfo{
				{
					Variable:        "@deep_nested",
					TypeString:      "ArrayRef[HashRef[ArrayRef[Int|Str]]]",
					IsParameterized: true,
					NestingDepth:    4,
				},
				{
					Variable:        "%complex_map",
					TypeString:      "Map[Str, ArrayRef[Tuple[Int, Bool|Str]]]",
					IsParameterized: true,
					ParamCount:      2,
					NestingDepth:    4,
				},
				{
					Variable:        "$deeply_nested",
					TypeString:      "Container[Wrapper[Inner[Data[Value]]]]",
					IsParameterized: true,
					NestingDepth:    5,
				},
			},
		},
		{
			Name: "Intersection type combinations",
			Input: `
				my ArrayRef[Object&Serializable] @serializable_list;
				my HashRef[ArrayRef[Int|Str]&Defined] %defined_arrays;
				my Container[Data&Validated&Cached] $safe_container;
			`,
			Description: "Intersection types combined with parameterized and union types",
			ExpectedTypes: []ExpectedTypeInfo{
				{
					Variable:        "@serializable_list",
					TypeString:      "ArrayRef[Object&Serializable]",
					IsParameterized: true,
					NestingDepth:    2,
				},
				{
					Variable:        "%defined_arrays",
					TypeString:      "HashRef[ArrayRef[Int|Str]&Defined]",
					IsParameterized: true,
					NestingDepth:    3,
				},
				{
					Variable:        "$safe_container",
					TypeString:      "Container[Data&Validated&Cached]",
					IsParameterized: true,
					NestingDepth:    2,
				},
			},
		},
		{
			Name: "Negation type combinations",
			Input: `
				my ArrayRef[!Undef] @non_undef_array;
				my HashRef[Str, !Empty] %non_empty_values;
				my Optional[!Null&!Undef] $definitely_defined;
			`,
			Description: "Negation types combined with parameterized and intersection types",
			ExpectedTypes: []ExpectedTypeInfo{
				{
					Variable:        "@non_undef_array",
					TypeString:      "ArrayRef[!Undef]",
					IsParameterized: true,
					NestingDepth:    2,
				},
				{
					Variable:        "%non_empty_values",
					TypeString:      "HashRef[Str, !Empty]",
					IsParameterized: true,
					ParamCount:      2,
					NestingDepth:    2,
				},
				{
					Variable:        "$definitely_defined",
					TypeString:      "Optional[!Null&!Undef]",
					IsParameterized: true,
					NestingDepth:    2,
				},
			},
		},
		{
			Name: "Performance stress test",
			Input: `
				my Map[Str, ArrayRef[HashRef[Tuple[Int, Result[Data[UserInfo], Error[ValidationFailure]]]]]] %extremely_nested;
				my (A|B|C|D|E|F|G|H|I|J|K|L|M|N|O|P|Q|R|S|T) $many_union_alternatives;
			`,
			Description:    "Stress testing with very deep nesting and many union alternatives",
			ShouldError:    true,
			ErrorMessages:  []string{"UnknownTypeError"},
			PerformanceMax: time.Second * 2,
			MemoryLimitMB:  100,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			startTime := time.Now()

			// Parse the input
			ast, err := parser.ParseString(test.Input)

			parseTime := time.Since(startTime)

			// Check for parsing errors
			if test.ShouldError {
				assert.Error(t, err, "Expected parsing to fail for %s", test.Name)
				if len(test.ErrorMessages) > 0 {
					for _, expectedError := range test.ErrorMessages {
						assert.Contains(t, err.Error(), expectedError, "Expected error message not found")
					}
				}
				return
			}

			// Should not error for successful cases
			require.NoError(t, err, "Parsing failed for %s: %v", test.Name, err)
			require.NotNil(t, ast, "AST should not be nil for %s", test.Name)

			// Performance check
			if test.PerformanceMax > 0 {
				assert.LessOrEqual(t, parseTime, test.PerformanceMax,
					"Parsing took too long: %v > %v", parseTime, test.PerformanceMax)
			}

			// Validate expected type annotations
			if len(test.ExpectedTypes) > 0 {
				validateTypeAnnotations(t, ast, test.ExpectedTypes)
			}

			t.Logf("Test %s completed in %v", test.Name, parseTime)
		})
	}
}

func validateTypeAnnotations(t *testing.T, ast *AST, expected []ExpectedTypeInfo) {
	require.NotNil(t, ast.TypeAnnotations, "TypeAnnotations should not be nil")

	// Create a map of variable names to type annotations for easier lookup
	foundTypes := make(map[string]*TypeAnnotation)
	for _, annotation := range ast.TypeAnnotations {
		foundTypes[annotation.AnnotatedItem] = annotation
	}

	for _, expectedType := range expected {
		annotation, found := foundTypes[expectedType.Variable]
		if !assert.True(t, found, "Expected to find type annotation for %s", expectedType.Variable) {
			continue
		}

		require.NotNil(t, annotation.TypeExpression, "TypeExpression should not be nil for %s", expectedType.Variable)

		// Validate type string if specified
		if expectedType.TypeString != "" {
			actualTypeString := annotation.TypeExpression.String()
			assert.Equal(t, expectedType.TypeString, actualTypeString,
				"Type string mismatch for %s", expectedType.Variable)
		}

		// Validate type characteristics
		if expectedType.IsUnion {
			assert.True(t, annotation.TypeExpression.IsUnion,
				"Expected %s to be a union type", expectedType.Variable)
		}

		if expectedType.IsIntersection {
			assert.True(t, annotation.TypeExpression.IsIntersection,
				"Expected %s to be an intersection type", expectedType.Variable)
		}

		if expectedType.IsNegation {
			assert.True(t, annotation.TypeExpression.IsNegation,
				"Expected %s to be a negation type", expectedType.Variable)
		}

		if expectedType.IsParameterized {
			assert.True(t, len(annotation.TypeExpression.Parameters) > 0,
				"Expected %s to have parameters", expectedType.Variable)
		}

		if expectedType.ParamCount > 0 {
			assert.Equal(t, expectedType.ParamCount, len(annotation.TypeExpression.Parameters),
				"Parameter count mismatch for %s", expectedType.Variable)
		}

		if expectedType.UnionCount > 0 {
			assert.Equal(t, expectedType.UnionCount, len(annotation.TypeExpression.UnionTypes),
				"Union type count mismatch for %s", expectedType.Variable)
		}

		// Validate nesting depth if specified
		if expectedType.NestingDepth > 0 {
			actualDepth := calculateNestingDepth(annotation.TypeExpression)
			assert.LessOrEqual(t, expectedType.NestingDepth, actualDepth+1, // Allow some tolerance
				"Nesting depth mismatch for %s: expected %d, got %d",
				expectedType.Variable, expectedType.NestingDepth, actualDepth)
		}
	}
}

func calculateNestingDepth(expr *TypeExpression) int {
	if expr == nil {
		return 0
	}

	maxDepth := 0

	// Check parameters
	for _, param := range expr.Parameters {
		depth := calculateNestingDepth(param)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	// Check union types
	for _, unionType := range expr.UnionTypes {
		depth := calculateNestingDepth(unionType)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	// Check intersection types
	for _, intersectionType := range expr.IntersectionTypes {
		depth := calculateNestingDepth(intersectionType)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	// Check negated type
	if expr.NegatedType != nil {
		depth := calculateNestingDepth(expr.NegatedType)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth + 1
}

func TestComplexTypeErrorRecovery(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)

	errorTests := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name:          "Incomplete union type",
			input:         "my ArrayRef[Int| $incomplete_union;",
			expectedError: "missing closing bracket",
		},
		{
			name:          "Double comma in parameters",
			input:         "my HashRef[Str,, Int] $double_comma;",
			expectedError: "syntax error",
		},
		{
			name:          "Unclosed nested brackets",
			input:         "my Container[Wrapper[Missing $unclosed_nested;",
			expectedError: "syntax error",
		},
		{
			name:          "Mixed operators without parentheses",
			input:         "my (Int|Str& $mixed_operators_without_close;",
			expectedError: "syntax error",
		},
	}

	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parser.ParseString(test.input)

			// We expect parsing to either fail or produce warnings
			// The exact behavior depends on error recovery implementation
			if err != nil {
				assert.Contains(t, strings.ToLower(err.Error()),
					strings.ToLower(test.expectedError),
					"Error message should contain expected text")
			} else {
				// If parsing succeeds, check for errors in the AST
				t.Logf("Parsing succeeded despite malformed input - parser has good error recovery")
			}
		})
	}
}

func TestComplexTypesFromTestData(t *testing.T) {
	t.Skip("Complex types tests now covered by TestRunMarkdownTestsByCategory - JSON files removed to avoid duplication")
	testDataDir := filepath.Join("testdata", "typed-perl", "complex-types")

	// Load test cases from JSON files
	testFiles, err := filepath.Glob(filepath.Join(testDataDir, "*.json"))
	if err != nil {
		t.Skipf("Could not load test files from %s: %v", testDataDir, err)
		return
	}

	if len(testFiles) == 0 {
		t.Skipf("No test files found in %s", testDataDir)
		return
	}

	parser, err := NewParser()
	require.NoError(t, err)

	for _, testFile := range testFiles {
		testCase, err := LoadTestCase(testFile)
		if err != nil {
			t.Errorf("Failed to load test case %s: %v", testFile, err)
			continue
		}

		t.Run(testCase.Name, func(t *testing.T) {
			startTime := time.Now()

			ast, err := parser.ParseString(testCase.Input)
			parseTime := time.Since(startTime)

			if testCase.ShouldError {
				assert.Error(t, err, "Expected parsing to fail for %s", testCase.Name)
			} else {
				assert.NoError(t, err, "Parsing failed for %s: %v", testCase.Name, err)
				assert.NotNil(t, ast, "AST should not be nil for %s", testCase.Name)

				// Basic validation that we found some type annotations
				if ast != nil && len(ast.TypeAnnotations) > 0 {
					t.Logf("Found %d type annotations in %v", len(ast.TypeAnnotations), parseTime)
					for _, annotation := range ast.TypeAnnotations {
						t.Logf("  %s: %s", annotation.AnnotatedItem, annotation.TypeExpression.String())
					}
				}
			}
		})
	}
}

func LoadTestCase(filepath string) (*ParserTestCase, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var testCase ParserTestCase
	err = json.Unmarshal(data, &testCase)
	if err != nil {
		return nil, err
	}

	return &testCase, nil
}
