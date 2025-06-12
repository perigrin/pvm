// ABOUTME: Comprehensive performance testing infrastructure for Step 23
// ABOUTME: Provides performance benchmarking and regression detection for parser improvements

package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"tamarou.com/pvm/internal/ast"
)

// PerformanceTest represents a single performance test configuration
type PerformanceTest struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	InputCode   string        `json:"input_code"`
	MaxDuration time.Duration `json:"max_duration"`
	MaxMemory   int64         `json:"max_memory"`
	Iterations  int           `json:"iterations"`
	Complexity  int           `json:"complexity"` // 1=simple, 2=medium, 3=complex, 4=stress
}

// Note: PerformanceResult is now defined in performance_monitor.go

// PerformanceBenchmarkBaseline represents baseline performance metrics for individual tests
type PerformanceBenchmarkBaseline struct {
	TestName         string        `json:"test_name"`
	BaselineDuration time.Duration `json:"baseline_duration"`
	BaselineMemory   int64         `json:"baseline_memory"`
	Timestamp        time.Time     `json:"timestamp"`
	ParserType       string        `json:"parser_type"`
}

// PerformanceReport aggregates performance test results
type PerformanceReport struct {
	TestResults []PerformanceResult            `json:"test_results"`
	Baselines   []PerformanceBenchmarkBaseline `json:"baselines"`
	Summary     PerformanceSummary             `json:"summary"`
	Timestamp   time.Time                      `json:"timestamp"`
}

// PerformanceSummary provides aggregate performance metrics
// Note: PerformanceSummary is now defined in performance_monitor.go

// PerformanceTestSuite manages performance testing infrastructure
type PerformanceTestSuite struct {
	TestDataDir string
	BaselineDir string
	ReportDir   string
	Parser      interface {
		ParseString(string) (*ast.AST, error)
		ParseFile(string) (*ast.AST, error)
	}
	RegressionThreshold float64 // Performance regression threshold (e.g., 1.20 = 20% slower)
}

// NewPerformanceTestSuite creates a new performance test suite
func NewPerformanceTestSuite(testDataDir, baselineDir, reportDir string) *PerformanceTestSuite {
	return &PerformanceTestSuite{
		TestDataDir:         testDataDir,
		BaselineDir:         baselineDir,
		ReportDir:           reportDir,
		RegressionThreshold: 1.20, // 20% slower is considered a regression
	}
}

// GeneratePerformanceTests creates a comprehensive set of performance tests
func (pts *PerformanceTestSuite) GeneratePerformanceTests() []*PerformanceTest {
	var tests []*PerformanceTest

	// Simple untyped Perl (baseline)
	tests = append(tests, &PerformanceTest{
		Name:        "simple_untyped_perl",
		Description: "Basic untyped Perl code without type annotations",
		InputCode: `my $simple = "hello";
my @array = (1, 2, 3);
sub simple_function { return 42; }
if ($simple) { print "hello"; }`,
		MaxDuration: 100 * time.Millisecond,
		MaxMemory:   1024 * 1024, // 1MB
		Iterations:  1000,
		Complexity:  1,
	})

	// Basic type annotations
	tests = append(tests, &PerformanceTest{
		Name:        "basic_type_annotations",
		Description: "Simple type annotations on variables and methods",
		InputCode: `my Int $typed_var = 42;
my ArrayRef[Str] @typed_array = ("a", "b");
method typed_method(Int $param) returns Str { return "$param"; }
field Bool $flag = 1;`,
		MaxDuration: 150 * time.Millisecond,
		MaxMemory:   2 * 1024 * 1024, // 2MB
		Iterations:  1000,
		Complexity:  2,
	})

	// Complex type expressions
	tests = append(tests, &PerformanceTest{
		Name:        "complex_type_expressions",
		Description: "Complex type expressions with unions, intersections, and parameterized types",
		InputCode: `my ArrayRef[HashRef[Int|Str]] @complex;
method complex_sig(
    ArrayRef[Object&Serializable] $input,
    CodeRef[Int, Bool|Str] $processor
) returns HashRef[ArrayRef[Int]|ErrorCode] { return {}; }
type ComplexType = ArrayRef[HashRef[Int|Str]|Boolean];
my ComplexType $var;`,
		MaxDuration: 500 * time.Millisecond,
		MaxMemory:   5 * 1024 * 1024, // 5MB
		Iterations:  500,
		Complexity:  3,
	})

	// Generate large program simulation
	var largeProgram strings.Builder
	largeProgram.WriteString("package LargeTestModule;\nuse strict;\nuse warnings;\n\n")

	// 1000+ variable declarations with types
	for i := 0; i < 1000; i++ {
		largeProgram.WriteString(fmt.Sprintf("my Int $var%d = %d;\n", i, i))
		if i%10 == 0 {
			largeProgram.WriteString(fmt.Sprintf("my ArrayRef[Str] @array%d = ();\n", i))
		}
		if i%20 == 0 {
			largeProgram.WriteString(fmt.Sprintf("my HashRef[Int] %%hash%d = {};\n", i))
		}
	}

	// 100+ method definitions with complex signatures
	for i := 0; i < 100; i++ {
		largeProgram.WriteString(fmt.Sprintf(`
method process%d(
    ArrayRef[Int] $data,
    Optional[CodeRef[Int, Str]] $transformer = undef
) returns Result[ArrayRef[Str], ErrorCode] {
    return Success->new([]);
}
`, i))
	}

	// 50+ class definitions with inheritance
	for i := 0; i < 50; i++ {
		largeProgram.WriteString(fmt.Sprintf(`
class TestClass%d {
    field Int $id = %d;
    field Str $name = "class%d";

    method new(Int $id, Str $name) returns TestClass%d {
        return bless {id => $id, name => $name}, __PACKAGE__;
    }
}
`, i, i, i, i))
	}

	tests = append(tests, &PerformanceTest{
		Name:        "large_program_simulation",
		Description: "Large program with 1000+ variables, 100+ methods, 50+ classes",
		InputCode:   largeProgram.String(),
		MaxDuration: 5 * time.Second,
		MaxMemory:   500 * 1024 * 1024, // 500MB for large programs
		Iterations:  10,
		Complexity:  4,
	})

	// Stress test patterns

	// Very deep type nesting
	deepNesting := "my "
	for i := 0; i < 10; i++ {
		deepNesting += "ArrayRef["
	}
	deepNesting += "Int"
	for i := 0; i < 10; i++ {
		deepNesting += "]"
	}
	deepNesting += " $deep_nested;"

	tests = append(tests, &PerformanceTest{
		Name:        "deep_type_nesting",
		Description: "Very deep type nesting (10+ levels)",
		InputCode:   deepNesting,
		MaxDuration: 1 * time.Second,
		MaxMemory:   10 * 1024 * 1024, // 10MB
		Iterations:  100,
		Complexity:  4,
	})

	// Very long union types
	unionTypes := "my "
	for i := 0; i < 20; i++ {
		if i > 0 {
			unionTypes += "|"
		}
		unionTypes += fmt.Sprintf("Type%d", i)
	}
	unionTypes += " $long_union;"

	tests = append(tests, &PerformanceTest{
		Name:        "long_union_types",
		Description: "Very long union types (20+ alternatives)",
		InputCode:   unionTypes,
		MaxDuration: 1 * time.Second,
		MaxMemory:   10 * 1024 * 1024, // 10MB
		Iterations:  100,
		Complexity:  4,
	})

	// Large method signatures
	var largeMethodSig strings.Builder
	largeMethodSig.WriteString("method large_method(")
	for i := 0; i < 50; i++ {
		if i > 0 {
			largeMethodSig.WriteString(", ")
		}
		largeMethodSig.WriteString(fmt.Sprintf("Int $param%d", i))
	}
	largeMethodSig.WriteString(") returns Bool { return 1; }")

	tests = append(tests, &PerformanceTest{
		Name:        "large_method_signature",
		Description: "Large method signature (50+ parameters)",
		InputCode:   largeMethodSig.String(),
		MaxDuration: 1 * time.Second,
		MaxMemory:   10 * 1024 * 1024, // 10MB
		Iterations:  100,
		Complexity:  4,
	})

	return tests
}

// RunPerformanceTest executes a single performance test
func (pts *PerformanceTestSuite) RunPerformanceTest(t *testing.T, test *PerformanceTest) *PerformanceResult {
	t.Helper()

	if pts.Parser == nil {
		t.Fatal("No parser configured for performance test suite")
	}

	// Force garbage collection before test
	runtime.GC()

	var totalDuration time.Duration
	var maxMemoryUsage int64
	var totalAllocations int64

	result := &PerformanceResult{
		TestName:   test.Name,
		Timestamp:  time.Now(),
		ParserType: "enhanced", // Could be determined dynamically
	}

	// Run multiple iterations to get stable measurements
	for i := 0; i < test.Iterations; i++ {
		// Measure memory before
		var m1, m2 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Run the parse operation
		start := time.Now()
		_, err := pts.Parser.ParseString(test.InputCode)
		duration := time.Since(start)

		// Measure memory after
		runtime.ReadMemStats(&m2)

		if err != nil {
			result.Success = false
			result.Error = err
			return result
		}

		totalDuration += duration
		allocDiff := m2.TotalAlloc - m1.TotalAlloc
		totalAllocations += int64(allocDiff)

		if int64(allocDiff) > maxMemoryUsage {
			maxMemoryUsage = int64(allocDiff)
		}

		// Check if we're exceeding limits
		if duration > test.MaxDuration {
			result.Success = false
			result.Error = fmt.Errorf("parse duration %v exceeded maximum %v", duration, test.MaxDuration)
			return result
		}
	}

	// Calculate averages
	result.ParseDuration = totalDuration / time.Duration(test.Iterations)
	result.MemoryUsage = maxMemoryUsage
	result.AllocCount = totalAllocations / int64(test.Iterations)
	result.Success = true

	// Check memory limits
	if result.MemoryUsage > test.MaxMemory {
		result.Success = false
		result.Error = fmt.Errorf("memory usage %d exceeded maximum %d", result.MemoryUsage, test.MaxMemory)
	}

	return result
}

// RunAllPerformanceTests executes all performance tests
func (pts *PerformanceTestSuite) RunAllPerformanceTests(t *testing.T) *PerformanceReport {
	tests := pts.GeneratePerformanceTests()

	report := &PerformanceReport{
		TestResults: make([]PerformanceResult, 0, len(tests)),
		Timestamp:   time.Now(),
	}

	// Load existing baselines if available
	baselines, err := pts.LoadBaselines()
	if err != nil {
		t.Logf("Warning: Could not load baselines: %v", err)
		baselines = make([]PerformanceBenchmarkBaseline, 0)
	}
	report.Baselines = baselines

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := pts.RunPerformanceTest(t, test)
			report.TestResults = append(report.TestResults, *result)

			// Check for regression
			if pts.CheckForRegression(result, baselines) {
				report.Summary.RegressionsFound++
				t.Logf("REGRESSION DETECTED in test %s", test.Name)
			}
		})
	}

	// Calculate summary
	pts.calculateSummary(&report.Summary, report.TestResults)

	return report
}

// CheckForRegression compares current result against baseline
func (pts *PerformanceTestSuite) CheckForRegression(result *PerformanceResult, baselines []PerformanceBenchmarkBaseline) bool {
	for _, baseline := range baselines {
		if baseline.TestName == result.TestName && baseline.ParserType == result.ParserType {
			// Check duration regression
			durationRatio := float64(result.ParseDuration) / float64(baseline.BaselineDuration)
			if durationRatio > pts.RegressionThreshold {
				return true
			}

			// Check memory regression
			memoryRatio := float64(result.MemoryUsage) / float64(baseline.BaselineMemory)
			if memoryRatio > pts.RegressionThreshold {
				return true
			}

			break
		}
	}
	return false
}

// SaveBaselines saves current results as performance baselines
func (pts *PerformanceTestSuite) SaveBaselines(results []PerformanceResult) error {
	baselines := make([]PerformanceBenchmarkBaseline, len(results))

	for i, result := range results {
		baselines[i] = PerformanceBenchmarkBaseline{
			TestName:         result.TestName,
			BaselineDuration: result.ParseDuration,
			BaselineMemory:   result.MemoryUsage,
			Timestamp:        result.Timestamp,
			ParserType:       result.ParserType,
		}
	}

	// Ensure baseline directory exists
	if err := os.MkdirAll(pts.BaselineDir, 0755); err != nil {
		return fmt.Errorf("failed to create baseline directory: %w", err)
	}

	baselineFile := filepath.Join(pts.BaselineDir, "performance_baselines.json")
	data, err := json.MarshalIndent(baselines, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal baselines: %w", err)
	}

	return os.WriteFile(baselineFile, data, 0644)
}

// LoadBaselines loads existing performance baselines
func (pts *PerformanceTestSuite) LoadBaselines() ([]PerformanceBenchmarkBaseline, error) {
	baselineFile := filepath.Join(pts.BaselineDir, "performance_baselines.json")

	data, err := os.ReadFile(baselineFile)
	if err != nil {
		return nil, err
	}

	var baselines []PerformanceBenchmarkBaseline
	err = json.Unmarshal(data, &baselines)
	return baselines, err
}

// SaveReport saves the performance report to a file
func (pts *PerformanceTestSuite) SaveReport(report *PerformanceReport) error {
	// Ensure report directory exists
	if err := os.MkdirAll(pts.ReportDir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	reportFile := filepath.Join(pts.ReportDir, fmt.Sprintf("performance_report_%s.json",
		report.Timestamp.Format("20060102_150405")))

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	return os.WriteFile(reportFile, data, 0644)
}

// calculateSummary calculates summary statistics for the performance report
func (pts *PerformanceTestSuite) calculateSummary(summary *PerformanceSummary, results []PerformanceResult) {
	summary.TotalTests = len(results)

	for _, result := range results {
		if result.Success {
			summary.PassedTests++
		} else {
			summary.FailedTests++
		}

		summary.TotalParseTime += result.ParseDuration
		summary.TotalMemoryUsed += result.MemoryUsage
	}

	if summary.TotalTests > 0 {
		summary.AvgParseTime = summary.TotalParseTime / time.Duration(summary.TotalTests)
		summary.AvgMemoryUsed = summary.TotalMemoryUsed / int64(summary.TotalTests)
	}
}

// PrintPerformanceReport prints a summary of the performance report
func (pts *PerformanceTestSuite) PrintPerformanceReport(t *testing.T, report *PerformanceReport) {
	t.Helper()

	t.Logf("=== Performance Test Report ===")
	t.Logf("Total Tests: %d", report.Summary.TotalTests)
	t.Logf("Passed Tests: %d", report.Summary.PassedTests)
	t.Logf("Failed Tests: %d", report.Summary.FailedTests)
	t.Logf("Regressions Found: %d", report.Summary.RegressionsFound)
	t.Logf("Total Parse Time: %v", report.Summary.TotalParseTime)
	t.Logf("Average Parse Time: %v", report.Summary.AvgParseTime)
	t.Logf("Total Memory Used: %d bytes", report.Summary.TotalMemoryUsed)
	t.Logf("Average Memory Used: %d bytes", report.Summary.AvgMemoryUsed)

	t.Logf("\nDetailed Results:")
	for _, result := range report.TestResults {
		status := "PASS"
		if !result.Success {
			status = "FAIL"
		}
		t.Logf("  %s [%s]: Duration=%v, Memory=%d bytes, Error=%s",
			result.TestName, status, result.ParseDuration, result.MemoryUsage, result.Error)
	}
}
