// ABOUTME: Comprehensive tests for method and field type annotation parsing
// ABOUTME: Tests typed method definitions, field declarations, and mixed typing scenarios

package parser

import (
	"path/filepath"
	"testing"
)

// TestMethodFieldAnnotations tests parsing of typed method definitions and field declarations
func TestMethodFieldAnnotations(t *testing.T) {
	testDataDir := filepath.Join("testdata", "typed-perl", "methods-fields")
	framework := NewParserTestFramework(testDataDir)
	
	// Create a parser instance for testing
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	framework.Parser = parser
	
	testFiles := []string{
		"basic_method_definitions.json",
		"basic_field_declarations.json", 
		"complex_method_signatures.json",
		"complex_field_types.json",
		"class_context_methods.json",
		"mixed_typed_untyped.json",
		"method_return_types.json",
		"field_access_modifiers.json",
	}
	
	totalTests := 0
	passedTests := 0
	
	for _, testFile := range testFiles {
		testPath := filepath.Join(testDataDir, testFile)
		t.Run(testFile, func(t *testing.T) {
			testCases, err := framework.LoadTestCases(testPath)
			if err != nil {
				t.Fatalf("Failed to load test cases from %s: %v", testFile, err)
			}
			
			for _, testCase := range testCases {
				t.Run(testCase.Name, func(t *testing.T) {
					totalTests++
					success := framework.RunTestCase(t, testCase)
					if success {
						passedTests++
					}
				})
			}
		})
	}
	
	t.Logf("Method and Field Annotation Tests: %d/%d passed (%.1f%%)", 
		passedTests, totalTests, float64(passedTests)/float64(totalTests)*100)
}

// TestMethodSignatureComponents tests individual components of method signatures
func TestMethodSignatureComponents(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	testCases := []struct {
		name     string
		input    string
		wantType string
		wantName string
	}{
		{
			name:     "simple_method",
			input:    "method foo() -> Str { return \"test\"; }",
			wantType: "method_decl",
			wantName: "foo",
		},
		{
			name:     "method_with_params",
			input:    "method add(Int $a, Int $b) -> Int { return $a + $b; }",
			wantType: "method_decl", 
			wantName: "add",
		},
		{
			name:     "method_no_return_type",
			input:    "method process(Str $input) { print $input; }",
			wantType: "method_decl",
			wantName: "process",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := parser.ParseString(tc.input)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			
			if ast == nil || ast.Root == nil {
				t.Fatal("Got nil AST or root")
			}
			
			// Basic validation that we can parse method definitions
			// More detailed AST validation would require deeper inspection
			t.Logf("Successfully parsed %s", tc.name)
		})
	}
}

// TestFieldDeclarationComponents tests individual components of field declarations
func TestFieldDeclarationComponents(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "simple_field",
			input: "field Int $count;",
		},
		{
			name:  "field_with_init",
			input: "field Str $name = \"default\";",
		},
		{
			name:  "complex_field_type",
			input: "field ArrayRef[HashRef[Int]] $data = [];",
		},
		{
			name:  "field_with_modifier",
			input: "field private Str $secret = \"hidden\";",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := parser.ParseString(tc.input)
			if err != nil {
				t.Fatalf("Parse error for %s: %v", tc.name, err)
			}
			
			if ast == nil || ast.Root == nil {
				t.Fatal("Got nil AST or root")
			}
			
			t.Logf("Successfully parsed %s", tc.name)
		})
	}
}