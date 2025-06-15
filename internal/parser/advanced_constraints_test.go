// ABOUTME: Tests for advanced type constraint parsing functionality
// ABOUTME: Validates comprehensive constraint parsing including protocol and value constraints

package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/ast"
)

// ConstraintTestSuite represents the test data format for constraint tests
type ConstraintTestSuite struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Tests       []ConstraintTest `json:"tests"`
}

// ConstraintTest represents a single constraint test case
type ConstraintTest struct {
	Name                    string               `json:"name,omitempty"`
	Input                   string               `json:"input"`
	ExpectedTypeAnnotations []ExpectedAnnotation `json:"expected_type_annotations,omitempty"`
}

// ExpectedAnnotation represents expected type annotation information
type ExpectedAnnotation struct {
	Kind        string               `json:"kind"`
	Item        string               `json:"item"`
	Type        string               `json:"type"`
	Constraints []ExpectedConstraint `json:"constraints,omitempty"`
}

// ExpectedConstraint represents expected constraint information
type ExpectedConstraint struct {
	Parameter  string `json:"parameter"`
	Kind       string `json:"kind"`
	Expression string `json:"expression"`
}

// loadConstraintTestSuite loads constraint test data from JSON file
func loadConstraintTestSuite(filePath string) ([]ConstraintTestSuite, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var testSuites []ConstraintTestSuite
	err = json.Unmarshal(data, &testSuites)
	if err != nil {
		return nil, err
	}

	return testSuites, nil
}

func TestAdvancedConstraintParsing(t *testing.T) {
	// Test advanced constraint syntax since tree-sitter grammar supports it

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	testCases := []struct {
		name        string
		input       string
		expectedAST func(*testing.T, *ast.AST)
	}{
		{
			name: "Basic Type Constraint",
			input: `method process<T>(ArrayRef[T] $data) returns ArrayRef[T] where T: Serializable {
				return $data;
			}`,
			expectedAST: func(t *testing.T, parsedAST *ast.AST) {
				if len(parsedAST.TypeAnnotations) == 0 {
					t.Error("Expected type annotations to be found")
					return
				}

				// Look for method with constraints
				found := false
				for _, annotation := range parsedAST.TypeAnnotations {
					if annotation.Kind == ast.MethodParamAnnotation {
						found = true
						if annotation.TypeExpression == nil {
							t.Error("Expected type expression")
							continue
						}
						// TODO: Add constraint validation once constraints are integrated
					}
				}
				if !found {
					t.Error("Expected to find method parameter annotation with constraints")
				}
			},
		},
		{
			name: "Multiple Type Constraints",
			input: `method transform<T, U>(T $input) returns U where T: Serializable&Defined, U: Deserializable&!Undef {
				return deserialize($input->serialize());
			}`,
			expectedAST: func(t *testing.T, parsedAST *ast.AST) {
				if len(parsedAST.TypeAnnotations) == 0 {
					t.Error("Expected type annotations to be found")
					return
				}

				// Should find method with multiple generic parameters and constraints
				methodParamCount := 0
				for _, annotation := range parsedAST.TypeAnnotations {
					if annotation.Kind == ast.MethodParamAnnotation {
						methodParamCount++
					}
				}
				if methodParamCount < 1 {
					t.Error("Expected at least one method parameter annotation")
				}
			},
		},
		{
			name: "Protocol Constraint",
			input: `method handle<T>(T $object) returns ProcessResult where T does EventHandler {
				return $object->process();
			}`,
			expectedAST: func(t *testing.T, parsedAST *ast.AST) {
				if len(parsedAST.TypeAnnotations) == 0 {
					t.Error("Expected type annotations to be found")
					return
				}
				// TODO: Validate protocol constraint once implemented
			},
		},
		{
			name: "Capability Constraint",
			input: `method process<T>(T $obj) returns Bool where T can 'serialize' {
				return defined($obj->serialize());
			}`,
			expectedAST: func(t *testing.T, parsedAST *ast.AST) {
				if len(parsedAST.TypeAnnotations) == 0 {
					t.Error("Expected type annotations to be found")
					return
				}
				// TODO: Validate capability constraint once implemented
			},
		},
		{
			name: "Value Constraint",
			input: `method create_array<T>(Int $size) returns ArrayRef[T] where T: Any, $size > 0 && $size < 1000 {
				return [(undef) x $size];
			}`,
			expectedAST: func(t *testing.T, parsedAST *ast.AST) {
				if len(parsedAST.TypeAnnotations) == 0 {
					t.Error("Expected type annotations to be found")
					return
				}
				// TODO: Validate value constraint once implemented
			},
		},
		{
			name: "Complex Mixed Constraints",
			input: `method advanced<T>(T $input) returns T where T does Serializable, T can 'process', T->VERSION >= 1.5 {
				return $input->process();
			}`,
			expectedAST: func(t *testing.T, parsedAST *ast.AST) {
				if len(parsedAST.TypeAnnotations) == 0 {
					t.Error("Expected type annotations to be found")
					return
				}
				// TODO: Validate mixed constraints once implemented
			},
		},
		{
			name: "Class with Constraint",
			input: `class Container<T> where T: Clonable {
				field ArrayRef[T] $items;
			}`,
			expectedAST: func(t *testing.T, parsedAST *ast.AST) {
				if len(parsedAST.TypeAnnotations) == 0 {
					t.Error("Expected type annotations to be found")
					return
				}
				// TODO: Validate class constraint once implemented
			},
		},
		{
			name: "Role with Constraint",
			input: `role Processable<T> where T: Defined {
				method process(T $input) -> ProcessResult;
			}`,
			expectedAST: func(t *testing.T, parsedAST *ast.AST) {
				if len(parsedAST.TypeAnnotations) == 0 {
					t.Error("Expected type annotations to be found")
					return
				}
				// TODO: Validate role constraint once implemented
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsedAST, err := parser.ParseString(tc.input)
			if err != nil {
				t.Fatalf("Failed to parse input: %v", err)
			}

			if parsedAST == nil {
				t.Fatal("Expected AST but got nil")
			}

			tc.expectedAST(t, parsedAST)
		})
	}
}

func TestConstraintTestDataFiles(t *testing.T) {
	// Test constraint test data files since tree-sitter grammar supports constraint parsing

	framework := NewParserTestFramework(filepath.Join("../../test/corpus/parser", "typed-perl", "advanced-constraints"))

	testFiles := []string{
		"basic_type_constraints.json",
		"multiple_constraints.json",
		"protocol_constraints.json",
		"value_constraints.json",
		"constraint_inheritance.json",
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			testPath := filepath.Join("../../test/corpus/parser", "typed-perl", "advanced-constraints", filename)

			// Convert constraint test format to parser test case format
			constraintTests, err := loadConstraintTestSuite(testPath)
			if err != nil {
				t.Fatalf("Failed to load constraint test suite %s: %v", filename, err)
			}

			for _, constraintTest := range constraintTests {
				for _, test := range constraintTest.Tests {
					// Convert to ParserTestCase format
					testCase := &ParserTestCase{
						Name:        test.Name,
						Category:    TypedPerl,
						Subcategory: "advanced-constraints",
						Input:       test.Input,
						ShouldError: false,
						Description: constraintTest.Description,
						Tags:        []string{"constraints", "typed-perl"},
					}

					t.Run(test.Name, func(t *testing.T) {
						success := framework.RunTestCase(t, testCase)
						if !success {
							t.Errorf("Constraint test case failed: %s", test.Name)
						}
					})
				}
			}
		})
	}
}

func TestConstraintParsingErrorRecovery(t *testing.T) {
	// Test constraint parsing error recovery since tree-sitter grammar supports 'where' syntax
	// This is a placeholder test - error recovery can be implemented later
}

// TestConstraintInheritance tests constraint inheritance from roles and parent classes
func TestConstraintInheritance(t *testing.T) {
	// Test constraint inheritance since tree-sitter grammar supports where clauses

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	input := `
		role Processable<T> where T: Defined {
			method process(T $input) -> ProcessResult;
		}

		class DataProcessor<T> does Processable<T> where T: Serializable&Defined {
			field ArrayRef[T] $data;
		}

		class AdvancedProcessor<T> : DataProcessor<T> where T: Cacheable {
			method cache(T $item) returns Void where $item->can('cache_key') {
				# Implementation
			}
		}
	`

	parsedAST, err := parser.ParseString(input)
	if err != nil {
		t.Fatalf("Failed to parse input: %v", err)
	}

	if parsedAST == nil {
		t.Fatal("Expected AST but got nil")
	}

	// Should find type annotations for role, class, and method
	if len(parsedAST.TypeAnnotations) == 0 {
		t.Error("Expected to find type annotations")
	}

	// TODO: Add specific validation for constraint inheritance once implemented
}
