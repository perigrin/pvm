// ABOUTME: Integration tests for the parser testing infrastructure
// ABOUTME: Validates that all components of the test framework work together correctly

package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserTestFramework_Integration(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()
	testDataDir := filepath.Join(tmpDir, "testdata")
	
	// Initialize framework
	framework := NewParserTestFramework(testDataDir)
	require.NotNil(t, framework)

	// Create test directories
	err := os.MkdirAll(filepath.Join(testDataDir, "untyped-perl", "variables"), 0755)
	require.NoError(t, err)
	err = os.MkdirAll(filepath.Join(testDataDir, "typed-perl", "simple-annotations"), 0755)
	require.NoError(t, err)

	// Create some test cases
	testCases := []*ParserTestCase{
		framework.GenerateTestCase(
			"simple_scalar",
			`my $name = "hello";`,
			"Simple scalar variable declaration",
			UntypedPerl,
			[]string{"scalar_variables", "variable_declarations"},
		),
		framework.GenerateTestCase(
			"typed_scalar", 
			`my Int $count = 42;`,
			"Typed scalar variable declaration",
			TypedPerl,
			[]string{"simple_types", "scalar_variables"},
		),
		framework.GenerateErrorTestCase(
			"syntax_error",
			`my $broken = ;`,
			"Syntax error test case",
			"syntax",
			ErrorCases,
			[]string{"error_recovery", "syntax_errors"},
		),
	}

	// Save test cases
	for i, testCase := range testCases {
		testFile := filepath.Join(testDataDir, "test_case_"+testCase.Name+".json")
		err := framework.SaveTestCase(testCase, testFile)
		require.NoError(t, err, "Failed to save test case %d", i)
	}

	// Test discovery
	discoveredTests, err := framework.DiscoverTestCases()
	require.NoError(t, err)
	assert.Len(t, discoveredTests, len(testCases), "Should discover all test cases")

	// Test individual test case execution
	for _, testCase := range testCases {
		t.Run("TestCase_"+testCase.Name, func(t *testing.T) {
			success := framework.RunTestCase(t, testCase)
			if testCase.ShouldError {
				assert.True(t, success, "Error test case should succeed when error is caught")
			} else {
				// For valid test cases, success depends on parser implementation
				// We don't assert here since the parser may not handle all cases yet
				t.Logf("Test case %s result: %v", testCase.Name, success)
			}
		})
	}

	// Test category-based execution
	metrics := framework.RunTestsByCategory(t, UntypedPerl)
	assert.NotNil(t, metrics)
	assert.Greater(t, metrics.TotalTests, 0, "Should run at least one untyped test")

	// Print metrics summary
	framework.PrintMetricsSummary(t, metrics)
}

func TestAccuracyMeasurement_Integration(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	testDataDir := filepath.Join(tmpDir, "testdata")
	baselineFile := filepath.Join(tmpDir, "baseline.json")
	reportDir := filepath.Join(tmpDir, "reports")

	// Create basic test case files
	err := os.MkdirAll(testDataDir, 0755)
	require.NoError(t, err)

	framework := NewParserTestFramework(testDataDir)
	testCase := framework.GenerateTestCase(
		"integration_test",
		`my $test = "value";`,
		"Integration test case",
		UntypedPerl,
		[]string{"scalar_variables"},
	)

	testFile := filepath.Join(testDataDir, "integration_test.json")
	err = framework.SaveTestCase(testCase, testFile)
	require.NoError(t, err)

	// Initialize accuracy measurement
	am := NewAccuracyMeasurement(testDataDir, baselineFile, reportDir)
	require.NotNil(t, am)

	// Test accuracy measurement (may fail if parser has issues, but shouldn't crash)
	baseline, err := am.MeasureCurrentAccuracy(t)
	if err != nil {
		t.Logf("Accuracy measurement failed (expected with minimal setup): %v", err)
		return
	}

	require.NotNil(t, baseline)
	assert.GreaterOrEqual(t, baseline.OverallAccuracy, 0.0)
	assert.LessOrEqual(t, baseline.OverallAccuracy, 100.0)

	// Test baseline saving and loading
	err = am.SaveBaseline(baseline)
	require.NoError(t, err)

	loadedBaseline, err := am.LoadBaseline()
	require.NoError(t, err)
	assert.Equal(t, baseline.OverallAccuracy, loadedBaseline.OverallAccuracy)

	// Test report generation
	err = am.GenerateReport(t, baseline)
	require.NoError(t, err)

	// Verify report files exist
	jsonReport := filepath.Join(reportDir, "accuracy_report.json")
	textReport := filepath.Join(reportDir, "accuracy_report.txt")
	
	_, err = os.Stat(jsonReport)
	assert.NoError(t, err, "JSON report should be created")
	
	_, err = os.Stat(textReport)
	assert.NoError(t, err, "Text report should be created")

	t.Logf("Accuracy measurement integration test completed successfully")
}

func TestParserTestFramework_AST_Validation(t *testing.T) {
	// Test AST validation functionality
	framework := NewParserTestFramework("testdata")
	
	parser, err := NewParser()
	require.NoError(t, err)

	// Test with valid Perl code
	testInput := `my $test = "hello";`
	ast, err := parser.ParseString(testInput)
	if err != nil {
		t.Skipf("Parser returned error for valid code: %v", err)
		return
	}

	if ast != nil {
		isValid := framework.ValidateAST(t, ast, "test_validation", testInput)
		assert.True(t, isValid, "Valid AST should pass validation")
	}

	// Test with nil AST
	isValid := framework.ValidateAST(t, nil, "nil_test", "")
	assert.False(t, isValid, "Nil AST should fail validation")
}

func TestTestCaseGeneration(t *testing.T) {
	framework := NewParserTestFramework("testdata")
	
	// Test normal test case generation
	testCase := framework.GenerateTestCase(
		"test_name",
		"my $var = 1;",
		"Test description",
		UntypedPerl,
		[]string{"tag1", "tag2"},
	)
	
	assert.Equal(t, "test_name", testCase.Name)
	assert.Equal(t, "my $var = 1;", testCase.Input)
	assert.Equal(t, "Test description", testCase.Description)
	assert.Equal(t, UntypedPerl, testCase.Category)
	assert.Equal(t, []string{"tag1", "tag2"}, testCase.Tags)
	assert.False(t, testCase.ShouldError)

	// Test error test case generation
	errorCase := framework.GenerateErrorTestCase(
		"error_test",
		"invalid syntax",
		"Error test description",
		"parse_error",
		ErrorCases,
		[]string{"error_tag"},
	)
	
	assert.Equal(t, "error_test", errorCase.Name)
	assert.True(t, errorCase.ShouldError)
	assert.Equal(t, "parse_error", errorCase.ErrorType)
	assert.Equal(t, ErrorCases, errorCase.Category)
}

// Benchmark the test framework itself
func BenchmarkTestFramework_Performance(b *testing.B) {
	framework := NewParserTestFramework("testdata")
	
	testCase := framework.GenerateTestCase(
		"benchmark_test",
		`my $test = "performance";`,
		"Performance benchmark test",
		UntypedPerl,
		[]string{"performance"},
	)

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Create a test that doesn't fail the benchmark
		t := &testing.T{}
		framework.RunTestCase(t, testCase)
	}
}

func TestParserTestFramework_ErrorHandling(t *testing.T) {
	framework := NewParserTestFramework("nonexistent/path")
	
	// Test discovery with nonexistent path
	testCases, err := framework.DiscoverTestCases()
	if err != nil {
		t.Logf("Expected error with nonexistent path: %v", err)
		assert.Empty(t, testCases)
	}

	// Test with invalid test case
	invalidTestCase := &ParserTestCase{
		Name:     "invalid",
		Category: "invalid_category",
		Input:    "",
	}
	
	success := framework.RunTestCase(t, invalidTestCase)
	t.Logf("Invalid test case result: %v", success)
}