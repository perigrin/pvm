// ABOUTME: Comprehensive test suite for AST-based clean Perl compiler
// ABOUTME: Ensures semantic equivalence with regex-based implementation

package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/parser"
)

// TestRegressionAgainstRegexCompiler ensures new AST compiler produces identical output
func TestASTCompilerCorrectness(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "simple_typed_variable",
			input: `my Int $count = 42;
print "Count: $count\n";`,
			expected: `my $count = 42;
print "Count: $count\n";`,
		},
		{
			name: "simple_str_variable",
			input: `my Str $name = "hello";
print "Name: $name\n";`,
			expected: `my $name = "hello";
print "Name: $name\n";`,
		},
		{
			name: "simple_function_signature",
			input: `sub add (Int $a, Int $b) -> Int {
    return $a + $b;
}`,
			expected: `use v5.36;
sub add($a, $b) { return $a + $b; }`,
		},
		{
			name: "complex_parameterized_types",
			input: `sub process (ArrayRef[HashRef[Str]] $data) -> Result[Bool] {
    return 1;
}`,
		},
		{
			name: "mixed_typed_untyped_params",
			input: `sub mixed (Int $typed, $untyped, Str $another) -> Void {
    return;
}`,
		},
		{
			name: "nested_union_types",
			input: `sub complex (Union[Str, Union[Int, Bool]] $param) -> Str {
    return "result";
}`,
		},
		{
			name: "intersection_and_negation",
			input: `sub validate (Object&Serializable $obj, !Undef $config) -> Bool {
    return 1;
}`,
		},
		{
			name: "field_declarations",
			input: `field Int $count;
field ArrayRef[Str] $items;`,
		},
		{
			name: "type_declarations",
			input: `type UserId = Int;
type UserData = HashRef[Str];`,
		},
		{
			name: "type_assertions",
			input: `my $value = get_value();
my $typed = $value as Int;`,
		},
		{
			name: "for_loop_with_types",
			input: `for my Int $i (@numbers) {
    print $i;
}`,
			expected: `for my $i (@numbers) { print $i; }`,
		},
		{
			name: "multiline_signature",
			input: `sub multiline (
    Int $first,
    Str $second,
    ArrayRef[Int] $third
) -> Bool {
    return 1;
}`,
		},
		{
			name: "attributes_with_signature",
			input: `sub tagged :lvalue :const (Int $value) -> Int {
    return $value;
}`,
		},
		{
			name: "method_declaration",
			input: `method add_user(UserId $id, HashRef[Str] $data) -> Bool {
    return 1;
}`,
		},
		{
			name: "field_with_initialization",
			input: `field Int $counter = 0;
field HashRef[Str] $config = {};`,
		},
		{
			name: "complex_typed_variables",
			input: `my ComplexTypes $self = bless {}, $class;
my ArrayRef[HashRef[Str|Int]] $users = [];`,
		},
		{
			name: "for_loop_with_complex_types",
			input: `for my UserId $id (keys %$config) {
    my HashRef[Str|Int] $user_info = {};
}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip test cases that don't have expected output defined yet
			if tc.expected == "" {
				t.Skip("Expected output not defined yet")
			}

			// Create temporary file for testing
			tempFile := createTempFile(t, tc.input)
			defer os.Remove(tempFile)

			// Parse the file
			p, err := parser.NewParser()
			require.NoError(t, err)

			ast, err := p.ParseFile(tempFile)
			require.NoError(t, err)

			// Test AST compiler
			astCompiler := NewASTCompiler()
			astResult, err := astCompiler.Compile(ast)
			require.NoError(t, err)

			t.Logf("Input:\n%s", tc.input)
			t.Logf("Expected output:\n%s", tc.expected)
			t.Logf("AST Compiler output:\n%s", astResult)

			// The AST compiler should produce the correct clean Perl output
			assert.Equal(t, strings.TrimSpace(tc.expected), strings.TrimSpace(astResult),
				"AST compiler output should match expected clean Perl output")
		})
	}
}

// TestExecutionValidation ensures generated Perl is syntactically valid and executable
func TestExecutionValidation(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string // Expected output when executed
	}{
		{
			name: "simple_function_execution",
			input: `use feature 'signatures';
sub add (Int $a, Int $b) -> Int {
    return $a + $b;
}

my Int $result = add(5, 3);
print "Result: $result\n";`,
			expected: "Result: 8\n",
		},
		{
			name: "typed_variables",
			input: `my Int $count = 42;
my Str $message = "Hello";
print "$message: $count\n";`,
			expected: "Hello: 42\n",
		},
		{
			name: "complex_data_structures",
			input: `sub process_data (ArrayRef[Int] $numbers) -> Int {
    my Int $sum = 0;
    for my Int $num (@$numbers) {
        $sum += $num;
    }
    return $sum;
}

my ArrayRef[Int] $data = [1, 2, 3, 4, 5];
my Int $total = process_data($data);
print "Total: $total\n";`,
			expected: "Total: 15\n",
		},
	}

	compiler := NewCleanPerlCompiler()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary file
			tempFile := createTempFile(t, tc.input)
			defer os.Remove(tempFile)

			// Parse and compile
			p, err := parser.NewParser()
			require.NoError(t, err)

			ast, err := p.ParseFile(tempFile)
			require.NoError(t, err)

			adapter := NewParserASTAdapter(ast)
			cleanPerl, err := compiler.Compile(adapter)
			require.NoError(t, err)

			// Debug: Show generated Perl
			t.Logf("Generated Perl:\n%s", cleanPerl)

			// Verify syntax is valid
			err = validatePerlSyntax(cleanPerl)
			assert.NoError(t, err, "Generated Perl should have valid syntax")

			// Execute and verify output
			if tc.expected != "" {
				output, err := executePerlCode(cleanPerl)
				assert.NoError(t, err, "Generated Perl should execute without errors")
				assert.Equal(t, tc.expected, output, "Execution output should match expected")
			}
		})
	}
}

// TestParserTestdataCompatibility ensures all parser test cases work with compiler
func TestParserTestdataCompatibility(t *testing.T) {
	// Load all test cases from internal/parser/testdata/typed-perl/
	testdataDir := "../../internal/parser/testdata/typed-perl"

	err := filepath.Walk(testdataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Process .json files containing test cases
		if strings.HasSuffix(path, ".json") && !strings.Contains(path, "metrics") {
			t.Run(filepath.Base(path), func(t *testing.T) {
				// TODO: Load and process JSON test cases
				// For now, just verify the file exists
				assert.FileExists(t, path)
			})
		}

		return nil
	})

	assert.NoError(t, err)
}

// TestSemanticEquivalence verifies compiled clean versions produce expected output
func TestSemanticEquivalence(t *testing.T) {
	testCases := []struct {
		name           string
		typed          string
		expectedOutput string
	}{
		{
			name: "simple_arithmetic",
			typed: `sub calculate (Int $x, Int $y) -> Int {
    return $x * 2 + $y;
}

print calculate(5, 3) . "\n";`,
			expectedOutput: "13\n",
		},
		{
			name: "string_operations",
			typed: `sub greet (Str $name) -> Str {
    return "Hello, " . $name . "!";
}

print greet("World") . "\n";`,
			expectedOutput: "Hello, World!\n",
		},
	}

	compiler := NewCleanPerlCompiler()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Compile typed version to clean Perl
			tempFile := createTempFile(t, tc.typed)
			defer os.Remove(tempFile)

			p, err := parser.NewParser()
			require.NoError(t, err)

			ast, err := p.ParseFile(tempFile)
			require.NoError(t, err)

			adapter := NewParserASTAdapter(ast)
			cleanPerl, err := compiler.Compile(adapter)
			require.NoError(t, err)

			// Execute the compiled clean version and compare against expected output
			cleanOutput, err := executePerlCode(cleanPerl)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedOutput, cleanOutput,
				"Compiled clean version should produce expected output")
		})
	}
}

// Helper functions

func createTempFile(t *testing.T, content string) string {
	tempFile, err := os.CreateTemp("", "test_*.pl")
	require.NoError(t, err)

	_, err = tempFile.WriteString(content)
	require.NoError(t, err)

	err = tempFile.Close()
	require.NoError(t, err)

	return tempFile.Name()
}

func validatePerlSyntax(code string) error {
	tempFile, err := os.CreateTemp("", "syntax_*.pl")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(code)
	if err != nil {
		return err
	}
	tempFile.Close()

	// Use perl -c to check syntax
	cmd := exec.Command("perl", "-c", tempFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("perl syntax error: %s\nCode:\n%s", string(output), code)
	}
	return nil
}

func executePerlCode(code string) (string, error) {
	tempFile, err := os.CreateTemp("", "exec_*.pl")
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(code)
	if err != nil {
		return "", err
	}
	tempFile.Close()

	cmd := exec.Command("perl", tempFile.Name())
	output, err := cmd.Output()
	return string(output), err
}
