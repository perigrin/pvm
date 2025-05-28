// ABOUTME: Comprehensive testing framework for parser accuracy and regression testing
// ABOUTME: Provides infrastructure for systematic testing of both untyped and typed Perl parsing

package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"tamarou.com/pvm/internal/ast"
)

// TestCategory represents different categories of parser tests
type TestCategory string

const (
	UntypedPerl TestCategory = "untyped-perl"
	TypedPerl   TestCategory = "typed-perl"
	ErrorCases  TestCategory = "error-cases"
)

// ParserTestCase represents a single test case for parser validation
type ParserTestCase struct {
	Name        string       `json:"name"`
	Category    TestCategory `json:"category"`
	Subcategory string       `json:"subcategory"`
	Input       string       `json:"input"`
	ExpectedAST *ast.AST     `json:"expected_ast,omitempty"`
	ShouldError bool         `json:"should_error"`
	ErrorType   string       `json:"error_type,omitempty"`
	Description string       `json:"description"`
	Tags        []string     `json:"tags"`
}

// AccuracyMetrics tracks parser accuracy across different dimensions
type AccuracyMetrics struct {
	TotalTests      int               `json:"total_tests"`
	PassedTests     int               `json:"passed_tests"`
	FailedTests     int               `json:"failed_tests"`
	CategoryMetrics map[string]Metric `json:"category_metrics"`
	FeatureMetrics  map[string]Metric `json:"feature_metrics"`
	ParsingTime     time.Duration     `json:"parsing_time"`
	MemoryUsage     int64             `json:"memory_usage"`
}

// Metric represents accuracy statistics for a specific dimension
type Metric struct {
	Total   int     `json:"total"`
	Passed  int     `json:"passed"`
	Failed  int     `json:"failed"`
	Accuracy float64 `json:"accuracy"`
}

// ParserTestFramework provides comprehensive testing infrastructure
type ParserTestFramework struct {
	TestDataDir string
	UpdateMode  bool
	Verbose     bool
	Parser      interface {
		ParseString(string) (*ast.AST, error)
		ParseFile(string) (*ast.AST, error)
	}
}

// NewParserTestFramework creates a new parser testing framework
func NewParserTestFramework(testDataDir string) *ParserTestFramework {
	return &ParserTestFramework{
		TestDataDir: testDataDir,
		UpdateMode:  os.Getenv("UPDATE_BASELINES") == "1",
		Verbose:     os.Getenv("VERBOSE_TESTS") == "1",
	}
}

// LoadTestCase loads a test case from a JSON file
func (f *ParserTestFramework) LoadTestCase(filePath string) (*ParserTestCase, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read test case file %s: %w", filePath, err)
	}

	var testCase ParserTestCase
	err = json.Unmarshal(data, &testCase)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test case file %s: %w", filePath, err)
	}

	return &testCase, nil
}

// SaveTestCase saves a test case to a JSON file
func (f *ParserTestFramework) SaveTestCase(testCase *ParserTestCase, filePath string) error {
	data, err := json.MarshalIndent(testCase, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal test case: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write test case file: %w", err)
	}

	return nil
}

// RunTestCase executes a single test case and returns the result
func (f *ParserTestFramework) RunTestCase(t *testing.T, testCase *ParserTestCase) bool {
	t.Helper()

	if f.Parser == nil {
		t.Errorf("No parser configured for test framework")
		return false
	}

	startTime := time.Now()
	ast, err := f.Parser.ParseString(testCase.Input)
	parseTime := time.Since(startTime)

	if testCase.ShouldError {
		if err == nil {
			t.Errorf("Test %s: Expected error but parsing succeeded", testCase.Name)
			return false
		}
		if testCase.ErrorType != "" && !strings.Contains(err.Error(), testCase.ErrorType) {
			t.Errorf("Test %s: Expected error type '%s' but got: %v", 
				testCase.Name, testCase.ErrorType, err)
			return false
		}
		if f.Verbose {
			t.Logf("Test %s: Successfully caught expected error: %v", testCase.Name, err)
		}
		return true
	}

	if err != nil {
		t.Errorf("Test %s: Unexpected parsing error: %v", testCase.Name, err)
		return false
	}

	if ast == nil {
		t.Errorf("Test %s: Parser returned nil AST", testCase.Name)
		return false
	}
	
	// Debug: log what we got from the parser
	if f.Verbose {
		t.Logf("Test %s: AST Source='%s', TypeAnnotations count=%d", 
			testCase.Name, ast.Source, len(ast.TypeAnnotations))
	}

	// Log performance metrics
	if f.Verbose {
		t.Logf("Test %s: Parse time: %v", testCase.Name, parseTime)
	}

	// If we have an expected AST, compare it
	if testCase.ExpectedAST != nil {
		if !f.CompareASTs(t, testCase.ExpectedAST, ast, testCase.Name) {
			return false
		}
	}

	// Validate AST structure is reasonable
	if !f.ValidateAST(t, ast, testCase.Name) {
		return false
	}

	return true
}

// CompareASTs compares two AST structures for equivalence
func (f *ParserTestFramework) CompareASTs(t *testing.T, expected, actual *ast.AST, testName string) bool {
	t.Helper()

	// Convert ASTs to comparable string representations
	expectedStr := expected.String()
	actualStr := actual.String()

	if expectedStr != actualStr {
		if f.UpdateMode {
			t.Logf("Test %s: AST mismatch - updating baseline in update mode", testName)
			return true
		}

		diff := cmp.Diff(expectedStr, actualStr)
		t.Errorf("Test %s: AST mismatch (-expected +actual):\n%s", testName, diff)
		return false
	}

	return true
}

// ValidateAST performs basic validation of AST structure
func (f *ParserTestFramework) ValidateAST(t *testing.T, ast *ast.AST, testName string) bool {
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

	// Validate that TypeAnnotations slice is initialized (can be empty)
	if ast.TypeAnnotations == nil {
		t.Errorf("Test %s: AST TypeAnnotations is nil", testName)
		return false
	}

	// Root can be nil for simple cases, so we don't require it
	// Additional validation can be added here

	return true
}

// DiscoverTestCases finds all test cases in the test data directory
func (f *ParserTestFramework) DiscoverTestCases() ([]*ParserTestCase, error) {
	var testCases []*ParserTestCase

	err := filepath.Walk(f.TestDataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".json") {
			testCase, err := f.LoadTestCase(path)
			if err != nil {
				return fmt.Errorf("failed to load test case %s: %w", path, err)
			}
			testCases = append(testCases, testCase)
		}

		return nil
	})

	return testCases, err
}

// RunAllTests executes all discovered test cases and returns accuracy metrics
func (f *ParserTestFramework) RunAllTests(t *testing.T) *AccuracyMetrics {
	testCases, err := f.DiscoverTestCases()
	if err != nil {
		t.Fatalf("Failed to discover test cases: %v", err)
	}

	metrics := &AccuracyMetrics{
		CategoryMetrics: make(map[string]Metric),
		FeatureMetrics:  make(map[string]Metric),
	}

	startTime := time.Now()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			success := f.RunTestCase(t, testCase)
			f.updateMetrics(metrics, testCase, success)
		})
	}

	metrics.ParsingTime = time.Since(startTime)
	f.calculateAccuracyPercentages(metrics)

	return metrics
}

// RunTestsByCategory runs tests for a specific category
func (f *ParserTestFramework) RunTestsByCategory(t *testing.T, category TestCategory) *AccuracyMetrics {
	testCases, err := f.DiscoverTestCases()
	if err != nil {
		t.Fatalf("Failed to discover test cases: %v", err)
	}

	metrics := &AccuracyMetrics{
		CategoryMetrics: make(map[string]Metric),
		FeatureMetrics:  make(map[string]Metric),
	}

	for _, testCase := range testCases {
		if testCase.Category == category {
			t.Run(testCase.Name, func(t *testing.T) {
				success := f.RunTestCase(t, testCase)
				f.updateMetrics(metrics, testCase, success)
			})
		}
	}

	f.calculateAccuracyPercentages(metrics)
	return metrics
}

// updateMetrics updates accuracy metrics based on test results
func (f *ParserTestFramework) updateMetrics(metrics *AccuracyMetrics, testCase *ParserTestCase, success bool) {
	metrics.TotalTests++
	if success {
		metrics.PassedTests++
	} else {
		metrics.FailedTests++
	}

	// Update category metrics
	categoryKey := string(testCase.Category)
	catMetric := metrics.CategoryMetrics[categoryKey]
	catMetric.Total++
	if success {
		catMetric.Passed++
	} else {
		catMetric.Failed++
	}
	metrics.CategoryMetrics[categoryKey] = catMetric

	// Update feature metrics based on tags
	for _, tag := range testCase.Tags {
		featureMetric := metrics.FeatureMetrics[tag]
		featureMetric.Total++
		if success {
			featureMetric.Passed++
		} else {
			featureMetric.Failed++
		}
		metrics.FeatureMetrics[tag] = featureMetric
	}
}

// calculateAccuracyPercentages calculates accuracy percentages for all metrics
func (f *ParserTestFramework) calculateAccuracyPercentages(metrics *AccuracyMetrics) {
	for key, metric := range metrics.CategoryMetrics {
		if metric.Total > 0 {
			metric.Accuracy = float64(metric.Passed) / float64(metric.Total) * 100
			metrics.CategoryMetrics[key] = metric
		}
	}

	for key, metric := range metrics.FeatureMetrics {
		if metric.Total > 0 {
			metric.Accuracy = float64(metric.Passed) / float64(metric.Total) * 100
			metrics.FeatureMetrics[key] = metric
		}
	}
}

// GenerateTestCase creates a test case from input code and expected behavior
func (f *ParserTestFramework) GenerateTestCase(name, input, description string, category TestCategory, tags []string) *ParserTestCase {
	return &ParserTestCase{
		Name:        name,
		Category:    category,
		Input:       input,
		Description: description,
		Tags:        tags,
		ShouldError: false,
	}
}

// GenerateErrorTestCase creates a test case that expects parsing to fail
func (f *ParserTestFramework) GenerateErrorTestCase(name, input, description, errorType string, category TestCategory, tags []string) *ParserTestCase {
	return &ParserTestCase{
		Name:        name,
		Category:    category,
		Input:       input,
		Description: description,
		Tags:        tags,
		ShouldError: true,
		ErrorType:   errorType,
	}
}

// SaveMetricsReport saves accuracy metrics to a JSON file
func (f *ParserTestFramework) SaveMetricsReport(metrics *AccuracyMetrics, filePath string) error {
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write metrics file: %w", err)
	}

	return nil
}

// PrintMetricsSummary prints a summary of accuracy metrics
func (f *ParserTestFramework) PrintMetricsSummary(t *testing.T, metrics *AccuracyMetrics) {
	t.Helper()

	overallAccuracy := float64(metrics.PassedTests) / float64(metrics.TotalTests) * 100

	t.Logf("=== Parser Accuracy Report ===")
	t.Logf("Overall: %d/%d tests passed (%.1f%% accuracy)", 
		metrics.PassedTests, metrics.TotalTests, overallAccuracy)
	t.Logf("Parse time: %v", metrics.ParsingTime)

	t.Logf("\nCategory Breakdown:")
	for category, metric := range metrics.CategoryMetrics {
		t.Logf("  %s: %d/%d (%.1f%%)", category, metric.Passed, metric.Total, metric.Accuracy)
	}

	if len(metrics.FeatureMetrics) > 0 {
		t.Logf("\nFeature Breakdown:")
		for feature, metric := range metrics.FeatureMetrics {
			t.Logf("  %s: %d/%d (%.1f%%)", feature, metric.Passed, metric.Total, metric.Accuracy)
		}
	}
}