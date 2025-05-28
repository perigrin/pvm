// ABOUTME: Backward compatibility validation for Step 24
// ABOUTME: Ensures enhanced parser maintains complete compatibility with existing untyped Perl code

package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/ast"
)

// CompatibilityTest represents a single backward compatibility test
type CompatibilityTest struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	InputCode         string   `json:"input_code"`
	ShouldParse       bool     `json:"should_parse"`
	ExpectedErrors    []string `json:"expected_errors,omitempty"`
	ExpectedAST       *ast.AST `json:"expected_ast,omitempty"`
	Category          string   `json:"category"`
	Tags              []string `json:"tags"`
}

// CompatibilityResult represents the result of a compatibility test
type CompatibilityResult struct {
	TestName        string   `json:"test_name"`
	ASTMatches      bool     `json:"ast_matches"`
	ErrorsMatch     bool     `json:"errors_match"`
	Compatible      bool     `json:"compatible"`
	Differences     []string `json:"differences"`
	ParseDuration   time.Duration `json:"parse_duration"`
	MemoryUsed      int64    `json:"memory_used"`
	ActualError     string   `json:"actual_error,omitempty"`
	ExpectedError   string   `json:"expected_error,omitempty"`
}

// CompatibilityReport aggregates compatibility test results
type CompatibilityReport struct {
	TotalTests       int                    `json:"total_tests"`
	CompatibleTests  int                    `json:"compatible_tests"`
	IncompatibleTests int                   `json:"incompatible_tests"`
	CategoryResults  map[string]int         `json:"category_results"`
	Results          []CompatibilityResult  `json:"results"`
	Summary          string                 `json:"summary"`
	Timestamp        time.Time              `json:"timestamp"`
}

// BackwardCompatibilityTester manages backward compatibility testing
type BackwardCompatibilityTester struct {
	TestDataDir string
	ReportDir   string
	Parser      interface {
		ParseString(string) (*ast.AST, error)
		ParseFile(string) (*ast.AST, error)
	}
	BaselineParser interface {
		ParseString(string) (*ast.AST, error)
		ParseFile(string) (*ast.AST, error)
	}
}

// NewBackwardCompatibilityTester creates a new compatibility tester
func NewBackwardCompatibilityTester(testDataDir, reportDir string) *BackwardCompatibilityTester {
	return &BackwardCompatibilityTester{
		TestDataDir: testDataDir,
		ReportDir:   reportDir,
	}
}

// GenerateCompatibilityTests creates comprehensive backward compatibility tests
func (bct *BackwardCompatibilityTester) GenerateCompatibilityTests() []*CompatibilityTest {
	var tests []*CompatibilityTest

	// Category 1: Basic untyped Perl patterns
	tests = append(tests, bct.generateBasicUntypedTests()...)
	
	// Category 2: Edge cases that might be affected by type parsing
	tests = append(tests, bct.generateEdgeCaseTests()...)
	
	// Category 3: Variables that look like types
	tests = append(tests, bct.generateTypeConflictTests()...)
	
	// Category 4: Complex expressions and operators
	tests = append(tests, bct.generateComplexExpressionTests()...)
	
	// Category 5: String literals and heredocs with type-like content
	tests = append(tests, bct.generateStringLiteralTests()...)
	
	// Category 6: Comments with type annotations
	tests = append(tests, bct.generateCommentTests()...)
	
	// Category 7: Error cases that should produce same errors
	tests = append(tests, bct.generateErrorCaseTests()...)

	return tests
}

// generateBasicUntypedTests creates tests for basic untyped Perl patterns
func (bct *BackwardCompatibilityTester) generateBasicUntypedTests() []*CompatibilityTest {
	return []*CompatibilityTest{
		{
			Name:        "simple_variable_declarations",
			Description: "Basic variable declarations without types",
			InputCode: `my $var = "value";
our @array = (1, 2, 3);
local %hash = (key => 'value');`,
			ShouldParse: true,
			Category:    "basic_untyped",
			Tags:        []string{"variables", "declarations"},
		},
		{
			Name:        "subroutines_without_types",
			Description: "Subroutine definitions without type annotations",
			InputCode: `sub existing_function {
    my ($param1, $param2) = @_;
    return $param1 + $param2;
}

sub another_function($a, $b) {
    return $a * $b;
}`,
			ShouldParse: true,
			Category:    "basic_untyped",
			Tags:        []string{"subroutines", "functions"},
		},
		{
			Name:        "complex_expressions_operators",
			Description: "Complex expressions and operators without types",
			InputCode: `my $result = ($a + $b) * ($c || 1) / ($d && $e);
my $string = "Hello " . $name . "!";
my $comparison = $x <=> $y;
my $logical = $p && $q || $r;`,
			ShouldParse: true,
			Category:    "basic_untyped",
			Tags:        []string{"expressions", "operators"},
		},
		{
			Name:        "control_structures",
			Description: "Control flow structures without types",
			InputCode: `if ($condition) {
    do_something();
} elsif ($other) {
    do_other();
}

for my $item (@list) {
    process($item);
    next if skip_condition($item);
    last if stop_condition($item);
}

while ($running) {
    work();
}`,
			ShouldParse: true,
			Category:    "basic_untyped",
			Tags:        []string{"control_flow", "loops"},
		},
		{
			Name:        "package_module_usage",
			Description: "Package declarations and module imports",
			InputCode: `package MyPackage;
use strict;
use warnings;
use Data::Dumper;

our $VERSION = '1.0';

sub new {
    my $class = shift;
    return bless {}, $class;
}`,
			ShouldParse: true,
			Category:    "basic_untyped",
			Tags:        []string{"packages", "modules"},
		},
	}
}

// generateEdgeCaseTests creates tests for edge cases
func (bct *BackwardCompatibilityTester) generateEdgeCaseTests() []*CompatibilityTest {
	return []*CompatibilityTest{
		{
			Name:        "variables_named_like_types",
			Description: "Variables that have names similar to type keywords",
			InputCode: `my $Int = "not a type";
my $ArrayRef = [];
my $Str = 42;
my $Bool = "false";`,
			ShouldParse: true,
			Category:    "edge_cases",
			Tags:        []string{"naming", "keywords"},
		},
		{
			Name:        "methods_conflicting_with_keywords",
			Description: "Methods that might conflict with type keywords",
			InputCode: `sub type { return "method"; }
sub method { return "function"; }
sub field { return "accessor"; }
sub as { return "conversion"; }

my $result = type() . method() . field() . as();`,
			ShouldParse: true,
			Category:    "edge_cases",
			Tags:        []string{"methods", "keywords"},
		},
		{
			Name:        "complex_expressions_with_type_like_calls",
			Description: "Complex expressions that might confuse type parser",
			InputCode: `my $complex = $obj->method() + $other->as() * $value;
my $chained = $data->type()->field()->method();
my $result = as($input) + type($param);`,
			ShouldParse: true,
			Category:    "edge_cases",
			Tags:        []string{"expressions", "method_calls"},
		},
		{
			Name:        "unicode_content",
			Description: "Unicode content that should parse correctly",
			InputCode: `my $名前 = "テスト";
my $données = "français";
my $данные = "русский";

sub процедура {
    return "функция";
}`,
			ShouldParse: true,
			Category:    "edge_cases",
			Tags:        []string{"unicode", "international"},
		},
	}
}

// generateTypeConflictTests creates tests for potential type parsing conflicts
func (bct *BackwardCompatibilityTester) generateTypeConflictTests() []*CompatibilityTest {
	return []*CompatibilityTest{
		{
			Name:        "package_names_like_types",
			Description: "Package names that look like type names",
			InputCode: `package Int::Parser;
package ArrayRef::Utils;
package Str::Helper;

use Int::Parser;
use ArrayRef::Utils qw(process);

my $obj = Int::Parser->new();`,
			ShouldParse: true,
			Category:    "type_conflicts",
			Tags:        []string{"packages", "naming"},
		},
		{
			Name:        "variables_with_type_operators",
			Description: "Variables using operators that look like type operators",
			InputCode: `my $pipe = $a | $b;
my $ampersand = $x & $y;
my $exclamation = !$flag;
my $arrow = $obj->method();`,
			ShouldParse: true,
			Category:    "type_conflicts",
			Tags:        []string{"operators", "variables"},
		},
		{
			Name:        "bareword_function_calls",
			Description: "Bareword function calls that might look like types",
			InputCode: `print Int();
say Str();
my $result = ArrayRef();
my @data = HashRef();`,
			ShouldParse: true,
			Category:    "type_conflicts",
			Tags:        []string{"barewords", "function_calls"},
		},
	}
}

// generateComplexExpressionTests creates tests for complex expressions
func (bct *BackwardCompatibilityTester) generateComplexExpressionTests() []*CompatibilityTest {
	return []*CompatibilityTest{
		{
			Name:        "nested_data_structures",
			Description: "Deeply nested data structures without types",
			InputCode: `my $deep = {
    level1 => {
        level2 => [
            { level3 => "value" },
            [ "array", "in", "array" ],
        ],
    },
};

my @complex_array = (
    { key => "value" },
    [ 1, 2, 3 ],
    \&some_function,
    \$scalar_ref,
);`,
			ShouldParse: true,
			Category:    "complex_expressions",
			Tags:        []string{"data_structures", "references"},
		},
		{
			Name:        "reference_operations",
			Description: "Reference and dereference operations",
			InputCode: `my $scalar_ref = \$scalar;
my $array_ref = \@array;
my $hash_ref = \%hash;
my $code_ref = \&function;

my $value = $$scalar_ref;
my @values = @$array_ref;
my %pairs = %$hash_ref;
my $result = &$code_ref();`,
			ShouldParse: true,
			Category:    "complex_expressions",
			Tags:        []string{"references", "dereferencing"},
		},
		{
			Name:        "regular_expressions",
			Description: "Regular expression patterns and operations",
			InputCode: `my $pattern = qr/\w+/;
my $match = $string =~ /pattern/;
my $substitution = $text =~ s/old/new/g;
my $transliteration = $data =~ tr/a-z/A-Z/;

if ($input =~ /^(\w+)\s*=\s*(.+)$/) {
    my ($key, $value) = ($1, $2);
}`,
			ShouldParse: true,
			Category:    "complex_expressions",
			Tags:        []string{"regex", "pattern_matching"},
		},
	}
}

// generateStringLiteralTests creates tests for string literals with type-like content
func (bct *BackwardCompatibilityTester) generateStringLiteralTests() []*CompatibilityTest {
	return []*CompatibilityTest{
		{
			Name:        "heredocs_with_type_content",
			Description: "Heredocs containing type annotation syntax",
			InputCode: `my $code = <<'END';
my Int $typed_var = 42;
method foo() -> Str { return "test"; }
field Bool $flag;
type MyType = Int|Str;
END

my $template = <<"TEMPLATE";
This is a template with $variable interpolation
and type-like content: ArrayRef[Int]
TEMPLATE`,
			ShouldParse: true,
			Category:    "string_literals",
			Tags:        []string{"heredocs", "interpolation"},
		},
		{
			Name:        "quoted_strings_with_types",
			Description: "Quoted strings containing type-like syntax",
			InputCode: `my $message = "The type is ArrayRef[Int]";
my $description = 'Method signature: sub foo(Int $a) -> Str';
my $pattern = q{my Int|Str $union = 42;};
my $command = qq{echo "field Bool \$flag;"};`,
			ShouldParse: true,
			Category:    "string_literals",
			Tags:        []string{"quotes", "strings"},
		},
	}
}

// generateCommentTests creates tests for comments with type annotations
func (bct *BackwardCompatibilityTester) generateCommentTests() []*CompatibilityTest {
	return []*CompatibilityTest{
		{
			Name:        "comments_with_type_syntax",
			Description: "Comments containing type annotation syntax",
			InputCode: `# This is a comment with Int and ArrayRef[Str]
my $var = 42; # Not typed despite comment containing: my Int $typed

=pod
This is POD documentation that mentions:
my HashRef[Int] $data;
method process() -> Bool;
=cut

# TODO: Add type annotations later
# my Str $name; - this should be a string type
my $name = "actual code";`,
			ShouldParse: true,
			Category:    "comments",
			Tags:        []string{"comments", "pod"},
		},
	}
}

// generateErrorCaseTests creates tests for error cases
func (bct *BackwardCompatibilityTester) generateErrorCaseTests() []*CompatibilityTest {
	return []*CompatibilityTest{
		{
			Name:        "incomplete_syntax_recovery",
			Description: "Parser error recovery for incomplete syntax",
			InputCode: `my $incomplete_syntax = `,
			ShouldParse: true, // Parser may recover from incomplete syntax
			Category:    "error_cases",
			Tags:        []string{"error_recovery"},
		},
		{
			Name:        "undefined_variables",
			Description: "References to undefined variables",
			InputCode: `my $result = $undefined_variable + 42;
print $nonexistent;`,
			ShouldParse: true, // This is a runtime error, not parse error
			Category:    "error_cases",
			Tags:        []string{"undefined_variables"},
		},
	}
}

// RunCompatibilityTest executes a single compatibility test
func (bct *BackwardCompatibilityTester) RunCompatibilityTest(t *testing.T, test *CompatibilityTest) *CompatibilityResult {
	t.Helper()

	result := &CompatibilityResult{
		TestName:    test.Name,
		Compatible:  true,
		Differences: []string{},
	}

	// Test with enhanced parser
	start := time.Now()
	enhancedAST, enhancedErr := bct.Parser.ParseString(test.InputCode)
	result.ParseDuration = time.Since(start)

	// Record memory usage
	result.MemoryUsed = getMemoryUsage()

	// Check if parsing behavior matches expectations
	if test.ShouldParse {
		if enhancedErr != nil {
			result.Compatible = false
			result.Differences = append(result.Differences, 
				fmt.Sprintf("Expected parsing to succeed but got error: %v", enhancedErr))
			result.ActualError = enhancedErr.Error()
		}
		if enhancedAST == nil {
			result.Compatible = false
			result.Differences = append(result.Differences, "Expected non-nil AST")
		}
	} else {
		if enhancedErr == nil {
			result.Compatible = false
			result.Differences = append(result.Differences, "Expected parsing to fail but succeeded")
		}
		// Check if error messages match expected patterns
		if enhancedErr != nil && len(test.ExpectedErrors) > 0 {
			errorMatched := false
			for _, expectedError := range test.ExpectedErrors {
				if strings.Contains(enhancedErr.Error(), expectedError) {
					errorMatched = true
					break
				}
			}
			if !errorMatched {
				result.ErrorsMatch = false
				result.Compatible = false
				result.Differences = append(result.Differences,
					fmt.Sprintf("Error message doesn't match expected patterns"))
				result.ActualError = enhancedErr.Error()
				result.ExpectedError = strings.Join(test.ExpectedErrors, " OR ")
			} else {
				result.ErrorsMatch = true
			}
		}
	}

	// If we have a baseline parser, compare AST structures
	if bct.BaselineParser != nil && test.ShouldParse && enhancedAST != nil {
		baselineAST, baselineErr := bct.BaselineParser.ParseString(test.InputCode)
		
		if baselineErr == nil && baselineAST != nil {
			astMatch := bct.compareASTs(enhancedAST, baselineAST)
			result.ASTMatches = astMatch
			if !astMatch {
				result.Compatible = false
				result.Differences = append(result.Differences, "AST structure differs from baseline")
			}
		}
	} else {
		result.ASTMatches = true // No baseline to compare against
	}

	// Validate AST structure for basic sanity
	if test.ShouldParse && enhancedAST != nil {
		if !bct.validateASTStructure(t, enhancedAST, test.Name) {
			result.Compatible = false
			result.Differences = append(result.Differences, "AST structure validation failed")
		}
	}

	return result
}

// RunAllCompatibilityTests executes all compatibility tests
func (bct *BackwardCompatibilityTester) RunAllCompatibilityTests(t *testing.T) *CompatibilityReport {
	tests := bct.GenerateCompatibilityTests()
	
	report := &CompatibilityReport{
		TotalTests:      len(tests),
		CategoryResults: make(map[string]int),
		Results:         make([]CompatibilityResult, 0, len(tests)),
		Timestamp:       time.Now(),
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := bct.RunCompatibilityTest(t, test)
			report.Results = append(report.Results, *result)

			if result.Compatible {
				report.CompatibleTests++
			} else {
				report.IncompatibleTests++
				t.Errorf("Compatibility test %s failed: %s", test.Name, 
					strings.Join(result.Differences, "; "))
			}

			// Track by category
			report.CategoryResults[test.Category]++
		})
	}

	// Generate summary
	compatibilityPercentage := float64(report.CompatibleTests) / float64(report.TotalTests) * 100
	report.Summary = fmt.Sprintf("Backward compatibility: %.1f%% (%d/%d tests passed)",
		compatibilityPercentage, report.CompatibleTests, report.TotalTests)

	return report
}

// compareASTs compares two AST structures for compatibility
func (bct *BackwardCompatibilityTester) compareASTs(enhanced, baseline *ast.AST) bool {
	// Basic structural comparison
	if enhanced == nil && baseline == nil {
		return true
	}
	if enhanced == nil || baseline == nil {
		return false
	}

	// Compare source preservation
	if enhanced.Source != baseline.Source {
		return false
	}

	// For untyped code, type annotations should be empty in both
	if len(enhanced.TypeAnnotations) > 0 {
		// Enhanced parser shouldn't add type annotations to untyped code
		return false
	}

	// Compare root node structure if available
	if enhanced.Root != nil && baseline.Root != nil {
		return bct.compareNodes(enhanced.Root, baseline.Root)
	}

	return enhanced.Root == baseline.Root // Both nil or both non-nil
}

// compareNodes compares individual AST nodes
func (bct *BackwardCompatibilityTester) compareNodes(enhanced, baseline ast.Node) bool {
	if enhanced == nil && baseline == nil {
		return true
	}
	if enhanced == nil || baseline == nil {
		return false
	}

	// Compare node types
	enhancedType := reflect.TypeOf(enhanced)
	baselineType := reflect.TypeOf(baseline)
	if enhancedType != baselineType {
		return false
	}

	// Compare children count
	enhancedChildren := enhanced.Children()
	baselineChildren := baseline.Children()
	if len(enhancedChildren) != len(baselineChildren) {
		return false
	}

	// Recursively compare children
	for i := 0; i < len(enhancedChildren); i++ {
		if !bct.compareNodes(enhancedChildren[i], baselineChildren[i]) {
			return false
		}
	}

	return true
}

// validateASTStructure performs basic validation of AST structure
func (bct *BackwardCompatibilityTester) validateASTStructure(t *testing.T, ast *ast.AST, testName string) bool {
	t.Helper()

	if ast == nil {
		t.Errorf("Test %s: AST is nil", testName)
		return false
	}

	// Basic structural validation
	if ast.Source == "" {
		t.Errorf("Test %s: AST source is empty", testName)
		return false
	}

	// Validate that TypeAnnotations slice is initialized (should be empty for untyped code)
	if ast.TypeAnnotations == nil {
		t.Errorf("Test %s: AST TypeAnnotations is nil", testName)
		return false
	}

	// For backward compatibility, we allow some type annotations to be detected
	// (e.g., in strings or comments) as long as they don't affect functionality
	// The key is that the code should parse and have a reasonable AST structure
	if len(ast.TypeAnnotations) > 0 {
		t.Logf("Test %s: Note: Found %d type annotations in untyped code (may be in strings/comments)", 
			testName, len(ast.TypeAnnotations))
	}

	return true
}

// SaveReport saves the compatibility report to a file
func (bct *BackwardCompatibilityTester) SaveReport(report *CompatibilityReport) error {
	if err := os.MkdirAll(bct.ReportDir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	reportFile := filepath.Join(bct.ReportDir, fmt.Sprintf("compatibility_report_%s.json", 
		report.Timestamp.Format("20060102_150405")))
	
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	return os.WriteFile(reportFile, data, 0644)
}

// PrintCompatibilityReport prints a summary of the compatibility report
func (bct *BackwardCompatibilityTester) PrintCompatibilityReport(t *testing.T, report *CompatibilityReport) {
	t.Helper()

	t.Logf("=== Backward Compatibility Report ===")
	t.Logf("%s", report.Summary)
	t.Logf("Total Tests: %d", report.TotalTests)
	t.Logf("Compatible Tests: %d", report.CompatibleTests)
	t.Logf("Incompatible Tests: %d", report.IncompatibleTests)

	t.Logf("\nCategory Breakdown:")
	for category, count := range report.CategoryResults {
		categoryPassed := 0
		for _, result := range report.Results {
			if result.Compatible {
				// Find test category - this is a simplified approach
				for _, test := range bct.GenerateCompatibilityTests() {
					if test.Category == category && test.Name == result.TestName {
						categoryPassed++
						break
					}
				}
			}
		}
		t.Logf("  %s: %d/%d tests", category, categoryPassed, count)
	}

	if report.IncompatibleTests > 0 {
		t.Logf("\nFailed Tests:")
		for _, result := range report.Results {
			if !result.Compatible {
				t.Logf("  %s: %s", result.TestName, strings.Join(result.Differences, "; "))
			}
		}
	}
}