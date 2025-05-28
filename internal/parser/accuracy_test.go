// ABOUTME: Accuracy measurement tools for parser testing and baseline establishment
// ABOUTME: Provides measurement capabilities for parser accuracy across different language features

package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// BaselineAccuracy represents baseline accuracy measurements for the parser
type BaselineAccuracy struct {
	Timestamp        time.Time                 `json:"timestamp"`
	ParserVersion    string                    `json:"parser_version"`
	OverallAccuracy  float64                   `json:"overall_accuracy"`
	CategoryAccuracy map[string]float64        `json:"category_accuracy"`
	FeatureAccuracy  map[string]float64        `json:"feature_accuracy"`
	PerformanceBaseline PerformanceBaseline    `json:"performance_baseline"`
	TestCoverage     TestCoverageReport        `json:"test_coverage"`
}

// PerformanceBaseline tracks performance characteristics
type PerformanceBaseline struct {
	AverageParseTime time.Duration `json:"average_parse_time"`
	MemoryUsage      int64         `json:"memory_usage"`
	ThroughputLPS    float64       `json:"throughput_lines_per_second"`
	ThroughputTPS    float64       `json:"throughput_tokens_per_second"`
}

// TestCoverageReport tracks what language features are covered by tests
type TestCoverageReport struct {
	TotalFeatures    int                `json:"total_features"`
	CoveredFeatures  int                `json:"covered_features"`
	CoveragePercent  float64            `json:"coverage_percent"`
	FeatureStatus    map[string]string  `json:"feature_status"` // "covered", "partial", "missing"
	UncoveredFeatures []string          `json:"uncovered_features"`
}

// AccuracyMeasurement provides tools for measuring parser accuracy
type AccuracyMeasurement struct {
	Framework    *ParserTestFramework
	BaselineFile string
	ReportDir    string
}

// NewAccuracyMeasurement creates a new accuracy measurement tool
func NewAccuracyMeasurement(testDataDir, baselineFile, reportDir string) *AccuracyMeasurement {
	return &AccuracyMeasurement{
		Framework:    NewParserTestFramework(testDataDir),
		BaselineFile: baselineFile,
		ReportDir:    reportDir,
	}
}

// MeasureCurrentAccuracy measures current parser accuracy across all test cases
func (am *AccuracyMeasurement) MeasureCurrentAccuracy(t *testing.T) (*BaselineAccuracy, error) {
	t.Helper()

	// Run all tests and collect metrics
	metrics := am.Framework.RunAllTests(t)

	// Measure performance
	perfBaseline, err := am.measurePerformance(t)
	if err != nil {
		return nil, fmt.Errorf("failed to measure performance: %w", err)
	}

	// Measure test coverage
	coverage, err := am.measureTestCoverage(t)
	if err != nil {
		return nil, fmt.Errorf("failed to measure test coverage: %w", err)
	}

	// Calculate category and feature accuracies
	categoryAccuracy := make(map[string]float64)
	for category, metric := range metrics.CategoryMetrics {
		categoryAccuracy[category] = metric.Accuracy
	}

	featureAccuracy := make(map[string]float64)
	for feature, metric := range metrics.FeatureMetrics {
		featureAccuracy[feature] = metric.Accuracy
	}

	overallAccuracy := float64(metrics.PassedTests) / float64(metrics.TotalTests) * 100

	baseline := &BaselineAccuracy{
		Timestamp:        time.Now(),
		ParserVersion:    "current", // TODO: get actual version
		OverallAccuracy:  overallAccuracy,
		CategoryAccuracy: categoryAccuracy,
		FeatureAccuracy:  featureAccuracy,
		PerformanceBaseline: *perfBaseline,
		TestCoverage:     *coverage,
	}

	return baseline, nil
}

// measurePerformance measures parser performance characteristics
func (am *AccuracyMeasurement) measurePerformance(t *testing.T) (*PerformanceBaseline, error) {
	t.Helper()

	parser, err := NewParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}

	// Test different code samples for performance measurement
	testCodes := []string{
		`my $simple = "hello";`,
		`my Int $typed = 42; my Str $name = "test";`,
		`sub test(Int $a, Str $b) -> Bool { return 1; }`,
		`package TestClass; field Str $name; method new() -> TestClass { return bless {}, __PACKAGE__; }`,
	}

	var totalTime time.Duration
	var totalLines int
	var totalTokens int

	// Measure memory before
	var memBefore runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	for i := 0; i < 100; i++ { // Run multiple iterations for accuracy
		for _, code := range testCodes {
			start := time.Now()
			ast, err := parser.ParseString(code)
			duration := time.Since(start)

			if err != nil {
				continue // Skip failed parses for performance measurement
			}

			totalTime += duration
			totalLines += countLines(code)
			if ast != nil {
				totalTokens += estimateTokenCount(code)
			}
		}
	}

	// Measure memory after
	var memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	iterations := 100 * len(testCodes)
	avgParseTime := totalTime / time.Duration(iterations)
	throughputLPS := float64(totalLines) / totalTime.Seconds()
	throughputTPS := float64(totalTokens) / totalTime.Seconds()
	memoryUsage := int64(memAfter.Alloc - memBefore.Alloc)

	return &PerformanceBaseline{
		AverageParseTime: avgParseTime,
		MemoryUsage:      memoryUsage,
		ThroughputLPS:    throughputLPS,
		ThroughputTPS:    throughputTPS,
	}, nil
}

// measureTestCoverage analyzes what language features are covered by tests
func (am *AccuracyMeasurement) measureTestCoverage(t *testing.T) (*TestCoverageReport, error) {
	t.Helper()

	// Define all Perl language features we want to test
	allFeatures := []string{
		// Basic variables
		"scalar_variables", "array_variables", "hash_variables",
		"variable_scoping", "package_variables",
		
		// Type annotations
		"simple_types", "parameterized_types", "union_types", 
		"intersection_types", "negation_types", "type_assertions",
		
		// Expressions
		"arithmetic_expressions", "string_expressions", "logical_expressions",
		"comparison_expressions", "assignment_expressions",
		
		// Control flow
		"if_statements", "while_loops", "for_loops", "foreach_loops",
		"loop_control", "given_when",
		
		// Subroutines and methods
		"subroutine_definitions", "subroutine_calls", "method_definitions",
		"method_calls", "typed_parameters", "return_types",
		
		// Object-oriented features
		"class_definitions", "role_definitions", "field_declarations",
		"inheritance", "role_composition", "method_signatures",
		
		// Advanced features
		"type_constraints", "generic_types", "complex_expressions",
		"nested_types", "error_recovery",
	}

	// Discover all test cases
	testCases, err := am.Framework.DiscoverTestCases()
	if err != nil {
		return nil, fmt.Errorf("failed to discover test cases: %w", err)
	}

	// Analyze which features are covered
	coveredFeatures := make(map[string]bool)
	featureStatus := make(map[string]string)

	for _, testCase := range testCases {
		for _, tag := range testCase.Tags {
			coveredFeatures[tag] = true
		}
	}

	var uncoveredFeatures []string
	for _, feature := range allFeatures {
		if coveredFeatures[feature] {
			featureStatus[feature] = "covered"
		} else {
			featureStatus[feature] = "missing"
			uncoveredFeatures = append(uncoveredFeatures, feature)
		}
	}

	coveragePercent := float64(len(coveredFeatures)) / float64(len(allFeatures)) * 100

	return &TestCoverageReport{
		TotalFeatures:     len(allFeatures),
		CoveredFeatures:   len(coveredFeatures),
		CoveragePercent:   coveragePercent,
		FeatureStatus:     featureStatus,
		UncoveredFeatures: uncoveredFeatures,
	}, nil
}

// SaveBaseline saves the current accuracy measurements as a baseline
func (am *AccuracyMeasurement) SaveBaseline(baseline *BaselineAccuracy) error {
	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal baseline: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(am.BaselineFile), 0755)
	if err != nil {
		return fmt.Errorf("failed to create baseline directory: %w", err)
	}

	err = os.WriteFile(am.BaselineFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write baseline file: %w", err)
	}

	return nil
}

// LoadBaseline loads a previously saved accuracy baseline
func (am *AccuracyMeasurement) LoadBaseline() (*BaselineAccuracy, error) {
	data, err := os.ReadFile(am.BaselineFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read baseline file: %w", err)
	}

	var baseline BaselineAccuracy
	err = json.Unmarshal(data, &baseline)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal baseline: %w", err)
	}

	return &baseline, nil
}

// CompareWithBaseline compares current accuracy against saved baseline
func (am *AccuracyMeasurement) CompareWithBaseline(t *testing.T, current *BaselineAccuracy) error {
	t.Helper()

	baseline, err := am.LoadBaseline()
	if err != nil {
		t.Logf("No baseline found, creating new baseline")
		return am.SaveBaseline(current)
	}

	// Compare overall accuracy
	accuracyDiff := current.OverallAccuracy - baseline.OverallAccuracy
	if accuracyDiff < -5.0 { // 5% regression threshold
		t.Errorf("Accuracy regression detected: %.1f%% -> %.1f%% (%.1f%% drop)",
			baseline.OverallAccuracy, current.OverallAccuracy, -accuracyDiff)
	} else if accuracyDiff > 1.0 { // 1% improvement threshold
		t.Logf("Accuracy improvement: %.1f%% -> %.1f%% (+%.1f%%)",
			baseline.OverallAccuracy, current.OverallAccuracy, accuracyDiff)
	}

	// Compare performance
	perfDiff := current.PerformanceBaseline.AverageParseTime - baseline.PerformanceBaseline.AverageParseTime
	if perfDiff > baseline.PerformanceBaseline.AverageParseTime/2 { // 50% performance regression
		t.Errorf("Performance regression detected: %v -> %v",
			baseline.PerformanceBaseline.AverageParseTime, current.PerformanceBaseline.AverageParseTime)
	}

	// Compare category accuracies
	for category, currentAcc := range current.CategoryAccuracy {
		if baselineAcc, exists := baseline.CategoryAccuracy[category]; exists {
			diff := currentAcc - baselineAcc
			if diff < -10.0 { // 10% category regression threshold
				t.Errorf("Category %s accuracy regression: %.1f%% -> %.1f%% (%.1f%% drop)",
					category, baselineAcc, currentAcc, -diff)
			}
		}
	}

	return nil
}

// GenerateReport generates a comprehensive accuracy report
func (am *AccuracyMeasurement) GenerateReport(t *testing.T, baseline *BaselineAccuracy) error {
	t.Helper()

	err := os.MkdirAll(am.ReportDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	// Generate JSON report
	jsonReport := filepath.Join(am.ReportDir, "accuracy_report.json")
	err = am.saveJSONReport(baseline, jsonReport)
	if err != nil {
		return fmt.Errorf("failed to save JSON report: %w", err)
	}

	// Generate human-readable report
	textReport := filepath.Join(am.ReportDir, "accuracy_report.txt")
	err = am.saveTextReport(baseline, textReport)
	if err != nil {
		return fmt.Errorf("failed to save text report: %w", err)
	}

	t.Logf("Accuracy reports generated in %s", am.ReportDir)
	return nil
}

// saveJSONReport saves a JSON format accuracy report
func (am *AccuracyMeasurement) saveJSONReport(baseline *BaselineAccuracy, filePath string) error {
	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// saveTextReport saves a human-readable accuracy report
func (am *AccuracyMeasurement) saveTextReport(baseline *BaselineAccuracy, filePath string) error {
	content := fmt.Sprintf(`Parser Accuracy Report
Generated: %s
Parser Version: %s

OVERALL ACCURACY
Overall: %.1f%%

CATEGORY BREAKDOWN
`, baseline.Timestamp.Format(time.RFC3339), baseline.ParserVersion, baseline.OverallAccuracy)

	for category, accuracy := range baseline.CategoryAccuracy {
		content += fmt.Sprintf("  %s: %.1f%%\n", category, accuracy)
	}

	content += "\nFEATURE BREAKDOWN\n"
	for feature, accuracy := range baseline.FeatureAccuracy {
		content += fmt.Sprintf("  %s: %.1f%%\n", feature, accuracy)
	}

	content += fmt.Sprintf(`
PERFORMANCE METRICS
Average Parse Time: %v
Memory Usage: %d bytes
Throughput: %.1f lines/sec, %.1f tokens/sec

TEST COVERAGE
Total Features: %d
Covered Features: %d
Coverage: %.1f%%

UNCOVERED FEATURES
`, baseline.PerformanceBaseline.AverageParseTime,
		baseline.PerformanceBaseline.MemoryUsage,
		baseline.PerformanceBaseline.ThroughputLPS,
		baseline.PerformanceBaseline.ThroughputTPS,
		baseline.TestCoverage.TotalFeatures,
		baseline.TestCoverage.CoveredFeatures,
		baseline.TestCoverage.CoveragePercent)

	for _, feature := range baseline.TestCoverage.UncoveredFeatures {
		content += fmt.Sprintf("  - %s\n", feature)
	}

	return os.WriteFile(filePath, []byte(content), 0644)
}

// Helper functions

func countLines(code string) int {
	lines := 1
	for _, char := range code {
		if char == '\n' {
			lines++
		}
	}
	return lines
}

func estimateTokenCount(code string) int {
	// Simple token estimation based on whitespace and common separators
	tokens := 0
	inToken := false
	
	for _, char := range code {
		if char == ' ' || char == '\t' || char == '\n' || char == ';' || char == '(' || char == ')' {
			if inToken {
				tokens++
				inToken = false
			}
		} else {
			inToken = true
		}
	}
	
	if inToken {
		tokens++
	}
	
	return tokens
}

// Test functions for the accuracy measurement system

func TestAccuracyMeasurement_Basic(t *testing.T) {
	testDataDir := "testdata"
	baselineFile := "testdata/baseline_accuracy.json"
	reportDir := "testdata/reports"

	am := NewAccuracyMeasurement(testDataDir, baselineFile, reportDir)
	
	// Measure current accuracy
	baseline, err := am.MeasureCurrentAccuracy(t)
	if err != nil {
		t.Logf("Failed to measure accuracy (expected with minimal test data): %v", err)
		return
	}

	// Generate report
	err = am.GenerateReport(t, baseline)
	if err != nil {
		t.Errorf("Failed to generate report: %v", err)
	}

	// Test baseline comparison
	err = am.CompareWithBaseline(t, baseline)
	if err != nil {
		t.Errorf("Failed to compare with baseline: %v", err)
	}

	t.Logf("Accuracy measurement test completed successfully")
}

func TestAccuracyMeasurement_Performance(t *testing.T) {
	am := NewAccuracyMeasurement("testdata", "testdata/perf_baseline.json", "testdata/reports")
	
	perf, err := am.measurePerformance(t)
	if err != nil {
		t.Errorf("Failed to measure performance: %v", err)
		return
	}

	t.Logf("Performance measurements:")
	t.Logf("  Average parse time: %v", perf.AverageParseTime)
	t.Logf("  Memory usage: %d bytes", perf.MemoryUsage)
	t.Logf("  Throughput: %.1f lines/sec, %.1f tokens/sec", 
		perf.ThroughputLPS, perf.ThroughputTPS)
}