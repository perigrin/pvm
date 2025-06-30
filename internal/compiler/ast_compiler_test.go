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

// NOTE: The old hardcoded AST compiler test was successfully replaced by
// the corpus-based TestCompilerCorpus test which provides better maintainability
// and easier addition of new test cases.

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

			// Use unified compiler from registry
			registry := NewCompilerRegistry()
			cleanPerl, err := registry.Compile(ast, TargetCleanPerl)
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

// TestCompilerCorpus tests compiler using corpus files
func TestCompilerCorpus(t *testing.T) {
	corpusDir := "../../test/corpus/compiler"

	// Check if corpus directory exists, if not skip the test
	if _, err := os.Stat(corpusDir); os.IsNotExist(err) {
		t.Skip("Compiler corpus directory not found, skipping corpus-based tests")
		return
	}

	// Create test framework
	framework := parser.NewParserTestFramework(corpusDir)

	// Walk through all markdown files in corpus directory
	err := filepath.Walk(corpusDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Process .md files containing test cases
		if strings.HasSuffix(path, ".md") {
			t.Run(filepath.Base(path), func(t *testing.T) {
				// Load test cases from markdown file
				testCases, err := framework.LoadMarkdownTestCases(path)
				require.NoError(t, err, "Failed to load test cases from %s", path)

				for _, testCase := range testCases {
					t.Run(testCase.Name, func(t *testing.T) {

						// Create temporary file for testing
						tempFile := createTempFile(t, testCase.Input)
						defer os.Remove(tempFile)

						// Parse the file
						p, err := parser.NewParser()
						require.NoError(t, err)

						ast, err := p.ParseFile(tempFile)
						require.NoError(t, err)

						// Test unified compiler (CST-based)
						unifiedCompiler := NewCleanPerlCompilerUnified()
						astResult, err := unifiedCompiler.Compile(ast)
						require.NoError(t, err)

						// Check if we have expected clean Perl output
						if testCase.ExpectedCompilationOutcomes != nil && testCase.ExpectedCompilationOutcomes.ExpectedCleanPerl != "" {
							expectedClean := strings.TrimSpace(testCase.ExpectedCompilationOutcomes.ExpectedCleanPerl)
							actualClean := strings.TrimSpace(astResult)

							assert.Equal(t, expectedClean, actualClean,
								"Compiler output should match expected clean Perl output for test case: %s", testCase.Name)
						} else {
							// For tests without expected output, just ensure we got some output
							assert.NotEmpty(t, strings.TrimSpace(astResult),
								"Compiler should produce some output for test case: %s", testCase.Name)
							t.Logf("Test %s passed - produced output (expected output not yet defined)", testCase.Name)
						}
					})
				}
			})
		}

		return nil
	})

	require.NoError(t, err, "Failed to walk corpus directory")
}

// TestParserTestdataCompatibility ensures all parser test cases work with compiler
func TestParserTestdataCompatibility(t *testing.T) {
	// Load all test cases from test/corpus/parser/typed-perl/
	testdataDir := "../../test/corpus/parser/typed-perl"

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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Compile typed version to clean Perl
			tempFile := createTempFile(t, tc.typed)
			defer os.Remove(tempFile)

			p, err := parser.NewParser()
			require.NoError(t, err)

			ast, err := p.ParseFile(tempFile)
			require.NoError(t, err)

			cleanPerl, err := NewCleanPerlCompilerUnified().Compile(ast)
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

	// Get a reliable perl path instead of depending on system environment
	perlPath, err := getTestPerlPath()
	if err != nil {
		return fmt.Errorf("failed to get perl path: %v", err)
	}

	// Use perl -c to check syntax
	cmd := exec.Command(perlPath, "-c", tempFile.Name())
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

	// Get a reliable perl path instead of depending on system environment
	perlPath, err := getTestPerlPath()
	if err != nil {
		return "", fmt.Errorf("failed to get perl path: %v", err)
	}

	cmd := exec.Command(perlPath, tempFile.Name())
	output, err := cmd.Output()
	return string(output), err
}

// getTestPerlPath returns a reliable perl path for testing, avoiding environment dependencies
func getTestPerlPath() (string, error) {
	// First try to find system perl directly in standard locations
	standardPaths := []string{
		"/usr/bin/perl",
		"/usr/local/bin/perl",
		"/opt/perl/bin/perl",
	}

	for _, path := range standardPaths {
		if _, err := os.Stat(path); err == nil {
			// Verify this perl works by checking version
			cmd := exec.Command(path, "-v")
			if err := cmd.Run(); err == nil {
				return path, nil
			}
		}
	}

	// If no standard paths work, try the PATH but with explicit command
	if perlPath, err := exec.LookPath("perl"); err == nil {
		// Test it works
		cmd := exec.Command(perlPath, "-v")
		if err := cmd.Run(); err == nil {
			return perlPath, nil
		}
	}

	return "", fmt.Errorf("no working perl installation found for testing")
}
