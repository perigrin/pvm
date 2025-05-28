// ABOUTME: Standalone tests for parser test framework that don't require tree-sitter
// ABOUTME: Validates the testing infrastructure independently of parser implementation

package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/ast"
)

// MockParser implements the Parser interface for testing
type MockParser struct {
	ShouldError bool
	ErrorMsg    string
}

func (m *MockParser) ParseString(code string) (*ast.AST, error) {
	if m.ShouldError {
		return nil, fmt.Errorf("%s", m.ErrorMsg)
	}
	
	// Return a minimal mock AST
	return &ast.AST{
		Path: "mock.pl",
		Source: code,
		TypeAnnotations: []*ast.TypeAnnotation{},
	}, nil
}

func (m *MockParser) ParseFile(filename string) (*ast.AST, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return m.ParseString(string(data))
}

func TestTestFramework_Core_Functionality(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()
	testDataDir := filepath.Join(tmpDir, "testdata")
	
	// Initialize framework with mock parser
	framework := NewParserTestFramework(testDataDir)
	framework.Parser = &MockParser{ShouldError: false}
	
	// Test test case generation
	testCase := framework.GenerateTestCase(
		"mock_test",
		`my $test = "value";`,
		"Mock test case",
		UntypedPerl,
		[]string{"mock", "variables"},
	)
	
	assert.Equal(t, "mock_test", testCase.Name)
	assert.Equal(t, UntypedPerl, testCase.Category)
	assert.False(t, testCase.ShouldError)
	
	// Test error case generation
	errorCase := framework.GenerateErrorTestCase(
		"mock_error",
		"invalid syntax",
		"Mock error case", 
		"syntax_error",
		ErrorCases,
		[]string{"errors"},
	)
	
	assert.True(t, errorCase.ShouldError)
	assert.Equal(t, "syntax_error", errorCase.ErrorType)
	
	// Test test case saving/loading
	testFile := filepath.Join(tmpDir, "test.json")
	err := framework.SaveTestCase(testCase, testFile)
	require.NoError(t, err)
	
	loadedCase, err := framework.LoadTestCase(testFile)
	require.NoError(t, err)
	assert.Equal(t, testCase.Name, loadedCase.Name)
	assert.Equal(t, testCase.Input, loadedCase.Input)
}

func TestTestFramework_Test_Execution(t *testing.T) {
	tmpDir := t.TempDir()
	framework := NewParserTestFramework(tmpDir)
	
	// Test successful parsing
	framework.Parser = &MockParser{ShouldError: false}
	successCase := framework.GenerateTestCase(
		"success_test",
		`my $var = 1;`,
		"Success test",
		UntypedPerl,
		[]string{"variables"},
	)
	
	result := framework.RunTestCase(t, successCase)
	assert.True(t, result, "Success case should pass")
	
	// Test error handling
	framework.Parser = &MockParser{ShouldError: true, ErrorMsg: "syntax_error"}
	errorCase := framework.GenerateErrorTestCase(
		"error_test",
		"invalid",
		"Error test",
		"syntax",
		ErrorCases,
		[]string{"errors"},
	)
	
	result = framework.RunTestCase(t, errorCase)
	assert.True(t, result, "Error case should pass when error is expected")
	
	// Test unexpected error
	unexpectedErrorCase := framework.GenerateTestCase(
		"unexpected_error",
		"code",
		"Should succeed but will fail",
		UntypedPerl,
		[]string{"test"},
	)
	
	result = framework.RunTestCase(t, unexpectedErrorCase)
	assert.False(t, result, "Should fail when unexpected error occurs")
}

func TestAccuracyMeasurement_Standalone(t *testing.T) {
	tmpDir := t.TempDir()
	testDataDir := filepath.Join(tmpDir, "testdata")
	baselineFile := filepath.Join(tmpDir, "baseline.json")
	reportDir := filepath.Join(tmpDir, "reports")
	
	// Create test case files
	err := os.MkdirAll(testDataDir, 0755)
	require.NoError(t, err)
	
	framework := NewParserTestFramework(testDataDir)
	framework.Parser = &MockParser{ShouldError: false}
	
	testCase := framework.GenerateTestCase(
		"standalone_test",
		`my $test = "value";`,
		"Standalone test",
		UntypedPerl,
		[]string{"variables"},
	)
	
	testFile := filepath.Join(testDataDir, "standalone_test.json")
	err = framework.SaveTestCase(testCase, testFile)
	require.NoError(t, err)
	
	// Test accuracy measurement
	am := NewAccuracyMeasurement(testDataDir, baselineFile, reportDir)
	am.Framework.Parser = &MockParser{ShouldError: false}
	
	baseline, err := am.MeasureCurrentAccuracy(t)
	require.NoError(t, err)
	require.NotNil(t, baseline)
	
	// Should have 100% accuracy with mock parser
	assert.Equal(t, 100.0, baseline.OverallAccuracy)
	assert.Greater(t, baseline.TestCoverage.TotalFeatures, 0)
	
	// Test baseline operations
	err = am.SaveBaseline(baseline)
	require.NoError(t, err)
	
	loadedBaseline, err := am.LoadBaseline()
	require.NoError(t, err)
	assert.Equal(t, baseline.OverallAccuracy, loadedBaseline.OverallAccuracy)
	
	// Test report generation
	err = am.GenerateReport(t, baseline)
	require.NoError(t, err)
	
	// Verify files were created
	jsonFile := filepath.Join(reportDir, "accuracy_report.json")
	textFile := filepath.Join(reportDir, "accuracy_report.txt")
	
	_, err = os.Stat(jsonFile)
	assert.NoError(t, err)
	
	_, err = os.Stat(textFile)
	assert.NoError(t, err)
}

func TestMetrics_Calculation(t *testing.T) {
	framework := NewParserTestFramework(t.TempDir())
	framework.Parser = &MockParser{ShouldError: false}
	
	metrics := &AccuracyMetrics{
		CategoryMetrics: make(map[string]Metric),
		FeatureMetrics:  make(map[string]Metric),
	}
	
	// Simulate test results
	testCases := []*ParserTestCase{
		framework.GenerateTestCase("test1", "code1", "desc1", UntypedPerl, []string{"feature1"}),
		framework.GenerateTestCase("test2", "code2", "desc2", TypedPerl, []string{"feature1", "feature2"}),
		framework.GenerateTestCase("test3", "code3", "desc3", UntypedPerl, []string{"feature2"}),
	}
	
	// Simulate test results: 2 pass, 1 fail
	framework.updateMetrics(metrics, testCases[0], true)
	framework.updateMetrics(metrics, testCases[1], true)
	framework.updateMetrics(metrics, testCases[2], false)
	
	framework.calculateAccuracyPercentages(metrics)
	
	// Check overall metrics
	assert.Equal(t, 3, metrics.TotalTests)
	assert.Equal(t, 2, metrics.PassedTests)
	assert.Equal(t, 1, metrics.FailedTests)
	
	// Check category metrics
	untypedMetric := metrics.CategoryMetrics["untyped-perl"]
	assert.Equal(t, 2, untypedMetric.Total)
	assert.Equal(t, 1, untypedMetric.Passed)
	assert.Equal(t, 50.0, untypedMetric.Accuracy)
	
	typedMetric := metrics.CategoryMetrics["typed-perl"]
	assert.Equal(t, 1, typedMetric.Total)
	assert.Equal(t, 1, typedMetric.Passed)
	assert.Equal(t, 100.0, typedMetric.Accuracy)
	
	// Check feature metrics
	feature1Metric := metrics.FeatureMetrics["feature1"]
	assert.Equal(t, 2, feature1Metric.Total)
	assert.Equal(t, 2, feature1Metric.Passed)
	assert.Equal(t, 100.0, feature1Metric.Accuracy)
	
	feature2Metric := metrics.FeatureMetrics["feature2"]
	assert.Equal(t, 2, feature2Metric.Total)
	assert.Equal(t, 1, feature2Metric.Passed)
	assert.Equal(t, 50.0, feature2Metric.Accuracy)
}

func TestTestCase_Serialization(t *testing.T) {
	framework := NewParserTestFramework(t.TempDir())
	
	original := framework.GenerateTestCase(
		"serialization_test",
		`my Int $test = 42;`,
		"Test serialization",
		TypedPerl,
		[]string{"serialization", "types"},
	)
	
	// Test JSON serialization
	data, err := json.Marshal(original)
	require.NoError(t, err)
	
	var deserialized ParserTestCase
	err = json.Unmarshal(data, &deserialized)
	require.NoError(t, err)
	
	assert.Equal(t, original.Name, deserialized.Name)
	assert.Equal(t, original.Input, deserialized.Input)
	assert.Equal(t, original.Category, deserialized.Category)
	assert.Equal(t, original.Tags, deserialized.Tags)
	assert.Equal(t, original.ShouldError, deserialized.ShouldError)
}

func TestTestFramework_Directory_Operations(t *testing.T) {
	tmpDir := t.TempDir()
	testDataDir := filepath.Join(tmpDir, "testdata")
	
	framework := NewParserTestFramework(testDataDir)
	
	// Create multiple test cases in different subdirectories  
	categories := []string{"cat1", "cat2"}
	for i, cat := range categories {
		subDir := filepath.Join(testDataDir, cat)
		err := os.MkdirAll(subDir, 0755)
		require.NoError(t, err)
		
		testCase := framework.GenerateTestCase(
			fmt.Sprintf("test_%d", i),
			fmt.Sprintf("my $var%d = %d;", i, i),
			fmt.Sprintf("Test %d", i),
			UntypedPerl,
			[]string{cat},
		)
		
		testFile := filepath.Join(subDir, fmt.Sprintf("test_%d.json", i))
		err = framework.SaveTestCase(testCase, testFile)
		require.NoError(t, err)
	}
	
	// Test discovery
	testCases, err := framework.DiscoverTestCases()
	require.NoError(t, err)
	assert.Len(t, testCases, len(categories))
	
	// Verify discovered test cases
	for _, testCase := range testCases {
		assert.NotEmpty(t, testCase.Name)
		assert.NotEmpty(t, testCase.Input)
		assert.Contains(t, []string{"test_0", "test_1"}, testCase.Name)
	}
}

// Benchmark the test framework performance
func BenchmarkTestFramework_Operations(b *testing.B) {
	tmpDir := b.TempDir()
	framework := NewParserTestFramework(tmpDir)
	framework.Parser = &MockParser{ShouldError: false}
	
	testCase := framework.GenerateTestCase(
		"benchmark",
		`my $benchmark = "test";`,
		"Benchmark test",
		UntypedPerl,
		[]string{"benchmark"},
	)
	
	b.Run("TestCaseExecution", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			t := &testing.T{} // Mock testing.T for benchmark
			framework.RunTestCase(t, testCase)
		}
	})
	
	b.Run("TestCaseSerialization", func(b *testing.B) {
		testFile := filepath.Join(tmpDir, "bench.json")
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			err := framework.SaveTestCase(testCase, testFile)
			if err != nil {
				b.Fatal(err)
			}
			
			_, err = framework.LoadTestCase(testFile)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}