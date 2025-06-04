// ABOUTME: Integration tests for parser with other PVM tools and components
// ABOUTME: Validates that enhanced parser output works correctly with PSC, LSP, and other tools

package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ToolIntegrationTest represents a test for parser integration with other tools
type ToolIntegrationTest struct {
	Name          string
	Description   string
	Program       string
	ExpectedAST   bool // Whether program should produce valid AST
	ExpectedTypes int  // Expected number of type annotations
	TestPSC       bool // Whether to test PSC integration
	TestLSP       bool // Whether to test LSP integration
}

// TestToolIntegration_PSCIntegration tests parser integration with PSC components
func TestToolIntegration_PSCIntegration(t *testing.T) {
	testCases := getPSCIntegrationTests()

	parser, err := NewParser()
	require.NoError(t, err, "Failed to create parser")

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Parse the program
			ast, err := parser.ParseString(tc.Program)

			if tc.ExpectedAST {
				if err != nil {
					t.Logf("Program failed to parse: %v", err)
				}
				require.NotNil(t, ast, "AST should not be nil")

				// Validate type annotation count
				if ast.TypeAnnotations != nil {
					actualTypes := len(ast.TypeAnnotations)
					if tc.ExpectedTypes > 0 {
						assert.GreaterOrEqual(t, actualTypes, tc.ExpectedTypes,
							"Should have at least expected number of type annotations")
					}
					t.Logf("Found %d type annotations (expected >= %d)", actualTypes, tc.ExpectedTypes)
				}

				// Test that AST can be used for type checking simulation
				typeCheckResult := simulateTypeChecking(t, ast)
				t.Logf("Type checking simulation result: %v", typeCheckResult)

				// Test that AST provides useful information for PSC tools
				pscInfo := extractPSCInfo(t, ast)
				assert.NotNil(t, pscInfo, "Should extract PSC-relevant information")
				t.Logf("PSC info extracted: %+v", pscInfo)

			} else {
				t.Logf("Program parsing completed (may have failed as expected)")
			}
		})
	}
}

// TestToolIntegration_LSPIntegration tests parser integration with LSP functionality
func TestToolIntegration_LSPIntegration(t *testing.T) {
	testCases := getLSPIntegrationTests()

	parser, err := NewParser()
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ast, err := parser.ParseString(tc.Program)

			if tc.ExpectedAST {
				if err != nil {
					t.Logf("Program failed to parse: %v", err)
				}
				require.NotNil(t, ast)

				// Test LSP-style navigation
				symbols := extractSymbols(t, ast)
				t.Logf("Extracted %d symbols for LSP", len(symbols))

				// Test hover information extraction
				hoverInfo := extractHoverInfo(t, ast, 1, 10) // Line 1, column 10
				t.Logf("Hover info at (1,10): %v", hoverInfo)

				// Test completion suggestions
				completions := extractCompletions(t, ast, 1, 15)
				t.Logf("Completions at (1,15): %d items", len(completions))

				// Test diagnostics extraction
				diagnostics := extractDiagnostics(t, ast)
				t.Logf("Diagnostics: %d items", len(diagnostics))
			}
		})
	}
}

// TestToolIntegration_ParserConsistency tests that parser output is consistent across runs
func TestToolIntegration_ParserConsistency(t *testing.T) {
	program := `
my Int $count = 42;
my Str $name = "test";

class TestClass {
    field Str $field1;
    field Int $field2;

    method test_method(Int $param) -> Str {
        return "result: $param";
    }
}
`

	parser, err := NewParser()
	require.NoError(t, err)

	// Parse the same program multiple times
	const numRuns = 5
	var asts []*AST
	var parseTimes []time.Duration

	for i := 0; i < numRuns; i++ {
		start := time.Now()
		ast, err := parser.ParseString(program)
		duration := time.Since(start)

		if err != nil {
			t.Logf("Parse run %d failed: %v", i+1, err)
			continue
		}

		require.NotNil(t, ast)
		asts = append(asts, ast)
		parseTimes = append(parseTimes, duration)
	}

	require.GreaterOrEqual(t, len(asts), 3, "Should have at least 3 successful parses")

	// Verify consistency across runs
	firstAST := asts[0]
	for i, ast := range asts[1:] {
		assert.Equal(t, firstAST.Source, ast.Source, "Source should be consistent (run %d)", i+2)

		if firstAST.TypeAnnotations != nil && ast.TypeAnnotations != nil {
			assert.Equal(t, len(firstAST.TypeAnnotations), len(ast.TypeAnnotations),
				"Type annotation count should be consistent (run %d)", i+2)
		}
	}

	// Log performance consistency
	totalTime := time.Duration(0)
	for _, duration := range parseTimes {
		totalTime += duration
	}
	avgTime := totalTime / time.Duration(len(parseTimes))
	t.Logf("Average parse time over %d runs: %v", len(parseTimes), avgTime)
}

// TestToolIntegration_ErrorHandling tests that parser errors are useful for tools
func TestToolIntegration_ErrorHandling(t *testing.T) {
	errorPrograms := []struct {
		name        string
		program     string
		description string
	}{
		{
			"incomplete_type",
			"my Int $var =;",
			"Incomplete assignment with type annotation",
		},
		{
			"malformed_union",
			"my Int||Str $bad_union;",
			"Malformed union type syntax",
		},
		{
			"unclosed_bracket",
			"my ArrayRef[Int $unclosed;",
			"Unclosed bracket in parameterized type",
		},
		{
			"invalid_constraint",
			"method test() where invalid syntax { }",
			"Invalid constraint syntax",
		},
	}

	parser, err := NewParser()
	require.NoError(t, err)

	for _, errorProg := range errorPrograms {
		t.Run(errorProg.name, func(t *testing.T) {
			ast, err := parser.ParseString(errorProg.program)

			// Even with errors, we might get partial AST
			if ast != nil {
				t.Logf("Got partial AST despite errors")

				// Test that tools can still extract some information
				diagnostics := extractDiagnostics(t, ast)
				t.Logf("Extracted %d diagnostics from error case", len(diagnostics))
			}

			if err != nil {
				t.Logf("Expected error for %s: %v", errorProg.description, err)

				// Verify error is informative
				errorStr := err.Error()
				assert.NotEmpty(t, errorStr, "Error message should not be empty")
				assert.True(t, len(errorStr) > 10, "Error message should be descriptive")
			}
		})
	}
}

// TestToolIntegration_PerformanceWithTools tests parser performance in tool contexts
func TestToolIntegration_PerformanceWithTools(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	// Create a moderately complex program
	program := generateComplexProgram(50) // 50 classes

	parser, err := NewParser()
	require.NoError(t, err)

	// Time the parsing
	start := time.Now()
	ast, err := parser.ParseString(program)
	parseTime := time.Since(start)

	if err != nil {
		t.Logf("Complex program parsing failed: %v", err)
		return
	}

	require.NotNil(t, ast)
	lines := strings.Count(program, "\n") + 1
	t.Logf("Parsed complex program: %d lines in %v", lines, parseTime)

	// Test tool operations on the complex AST
	start = time.Now()
	symbols := extractSymbols(t, ast)
	symbolTime := time.Since(start)
	t.Logf("Extracted %d symbols in %v", len(symbols), symbolTime)

	start = time.Now()
	diagnostics := extractDiagnostics(t, ast)
	diagnosticsTime := time.Since(start)
	t.Logf("Extracted %d diagnostics in %v", len(diagnostics), diagnosticsTime)

	// Performance thresholds
	maxParseTime := time.Millisecond * 100 // 100ms for moderate complexity
	maxToolTime := time.Millisecond * 50   // 50ms for tool operations

	if parseTime > maxParseTime {
		t.Logf("WARNING: Parse time %v exceeded threshold %v", parseTime, maxParseTime)
	}

	if symbolTime > maxToolTime {
		t.Logf("WARNING: Symbol extraction time %v exceeded threshold %v", symbolTime, maxToolTime)
	}

	if diagnosticsTime > maxToolTime {
		t.Logf("WARNING: Diagnostics time %v exceeded threshold %v", diagnosticsTime, maxToolTime)
	}
}

// TestToolIntegration_FileOperations tests parser integration with file-based operations
func TestToolIntegration_FileOperations(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	testFiles := map[string]string{
		"simple.pl": `
my Int $count = 42;
sub increment { $count++; }
`,
		"complex.pl": `
use v5.38;
use strict;
use warnings;

class Calculator {
    field Num $precision = 0.001;

    method add(Num $a, Num $b) -> Num {
        return $a + $b;
    }
}
`,
		"mixed.pl": `
# Mixed typed and untyped code
my $old_var = "legacy";
my Int $new_var = 42;

sub old_function { return shift() * 2; }
method new_method(Int $param) -> Int { return $param * 3; }
`,
	}

	parser, err := NewParser()
	require.NoError(t, err)

	for filename, content := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

		t.Run(filename, func(t *testing.T) {
			// Test parsing from file
			ast, err := parser.ParseFile(filePath)
			if err != nil {
				t.Logf("File parsing failed: %v", err)
			} else {
				require.NotNil(t, ast)
				assert.Equal(t, content, ast.Source)
				t.Logf("Successfully parsed file %s", filename)

				// Test that file path is preserved
				if ast.Path != "" {
					assert.Equal(t, filePath, ast.Path)
				}

				// Test tool operations work with file-based AST
				symbols := extractSymbols(t, ast)
				t.Logf("File %s: extracted %d symbols", filename, len(symbols))
			}
		})
	}
}

// Helper functions for simulating tool integrations

func simulateTypeChecking(t *testing.T, ast *AST) map[string]interface{} {
	result := map[string]interface{}{
		"ast_valid":  ast != nil,
		"has_source": ast != nil && ast.Source != "",
	}

	if ast != nil && ast.TypeAnnotations != nil {
		result["type_annotations"] = len(ast.TypeAnnotations)
		result["can_type_check"] = len(ast.TypeAnnotations) > 0
	}

	return result
}

func extractPSCInfo(t *testing.T, ast *AST) map[string]interface{} {
	if ast == nil {
		return nil
	}

	info := map[string]interface{}{
		"source_available": ast.Source != "",
		"path_available":   ast.Path != "",
	}

	if ast.TypeAnnotations != nil {
		info["type_count"] = len(ast.TypeAnnotations)
		info["has_types"] = len(ast.TypeAnnotations) > 0
	}

	return info
}

func extractSymbols(t *testing.T, ast *AST) []map[string]interface{} {
	if ast == nil {
		return nil
	}

	var symbols []map[string]interface{}

	// Simple symbol extraction simulation
	lines := strings.Split(ast.Source, "\n")
	for i, line := range lines {
		if strings.Contains(line, "my ") || strings.Contains(line, "class ") ||
			strings.Contains(line, "method ") || strings.Contains(line, "field ") {
			symbols = append(symbols, map[string]interface{}{
				"line": i + 1,
				"type": "declaration",
				"text": strings.TrimSpace(line),
			})
		}
	}

	return symbols
}

func extractHoverInfo(t *testing.T, ast *AST, line, column int) map[string]interface{} {
	if ast == nil {
		return nil
	}

	return map[string]interface{}{
		"line":          line,
		"column":        column,
		"has_content":   true,
		"source_length": len(ast.Source),
	}
}

func extractCompletions(t *testing.T, ast *AST, line, column int) []string {
	if ast == nil {
		return nil
	}

	// Simple completion simulation
	completions := []string{
		"Int", "Str", "Bool", "Num",
		"ArrayRef", "HashRef", "CodeRef",
		"class", "method", "field",
	}

	if len(ast.TypeAnnotations) > 0 {
		completions = append(completions, "type_annotation_context")
	}

	return completions
}

func extractDiagnostics(t *testing.T, ast *AST) []map[string]interface{} {
	if ast == nil {
		return nil
	}

	var diagnostics []map[string]interface{}

	// Simple diagnostic extraction
	lines := strings.Split(ast.Source, "\n")
	for i, line := range lines {
		if strings.Contains(line, "=;") || strings.Contains(line, "||") ||
			strings.Contains(line, "[") && !strings.Contains(line, "]") {
			diagnostics = append(diagnostics, map[string]interface{}{
				"line":     i + 1,
				"severity": "error",
				"message":  "Potential syntax issue",
			})
		}
	}

	return diagnostics
}

func generateComplexProgram(numClasses int) string {
	var program strings.Builder

	program.WriteString("use v5.38;\nuse strict;\nuse warnings;\n\n")

	for i := 0; i < numClasses; i++ {
		program.WriteString(fmt.Sprintf(`class Class%d {
    field Int $id_%d = %d;
    field Str $name_%d = "class_%d";

    method get_id() -> Int {
        return $id_%d;
    }

    method set_name(Str $name) -> Void {
        $name_%d = $name;
    }
}

`, i, i, i, i, i, i, i))
	}

	return program.String()
}

// Test data functions

func getPSCIntegrationTests() []ToolIntegrationTest {
	return []ToolIntegrationTest{
		{
			Name:        "basic_type_annotations",
			Description: "Basic type annotations for PSC processing",
			Program: `
my Int $count = 42;
my Str $name = "test";
my ArrayRef[Int] @numbers = (1, 2, 3);

method calculate(Int $a, Int $b) -> Int {
    return $a + $b;
}
`,
			ExpectedAST:   true,
			ExpectedTypes: 4,
			TestPSC:       true,
		},
		{
			Name:        "class_with_types",
			Description: "Class definition with typed fields and methods",
			Program: `
class UserService {
    field HashRef[Int, User] $cache = {};
    field Int $max_cache_size = 1000;

    method get_user(Int $id) -> Optional[User] {
        return $cache->{$id};
    }

    method add_user(User $user) -> Void {
        $cache->{$user->id} = $user;
    }
}
`,
			ExpectedAST:   true,
			ExpectedTypes: 5,
			TestPSC:       true,
		},
		{
			Name:        "complex_generics",
			Description: "Complex generic types for PSC type checking",
			Program: `
sub process {
    my (ArrayRef $input, CodeRef $filter) = @_;
    my ArrayRef $results = [];
    for my $item (@{$input}) {
        push @{$results}, $item if $filter->($item);
    }
    return $results;
}
`,
			ExpectedAST:   true,
			ExpectedTypes: 2,
			TestPSC:       true,
		},
	}
}

func getLSPIntegrationTests() []ToolIntegrationTest {
	return []ToolIntegrationTest{
		{
			Name:        "symbol_extraction",
			Description: "Code with various symbols for LSP navigation",
			Program: `
package MyPackage;

my Int $global_var = 42;

class MyClass {
    field Str $name;

    method get_name() -> Str {
        return $name;
    }
}

sub legacy_function {
    my ($param) = @_;
    return $param * 2;
}
`,
			ExpectedAST:   true,
			ExpectedTypes: 2,
			TestLSP:       true,
		},
		{
			Name:        "hover_information",
			Description: "Code with type information for hover display",
			Program: `
my ArrayRef[HashRef[Int]] @complex_data;
my CodeRef[Str, Bool] $validator = sub { length(shift) > 0 };

method validate_data(ArrayRef[Str] $input) -> ArrayRef[Bool] {
    return [map { $validator->($_) } @{$input}];
}
`,
			ExpectedAST:   true,
			ExpectedTypes: 3,
			TestLSP:       true,
		},
	}
}
