// ABOUTME: Validation functions for the parser test framework
// ABOUTME: Simple tests that validate infrastructure without requiring working tree-sitter parser

package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/ast"
)

// ValidateTestFrameworkInfrastructure runs basic validation of the test framework components
func ValidateTestFrameworkInfrastructure(t *testing.T) bool {
	t.Helper()
	
	allPassed := true

	// Test 1: Framework creation
	framework := NewParserTestFramework("testdata")
	if framework == nil {
		t.Error("Failed to create test framework")
		allPassed = false
	}

	// Test 2: Test case generation
	testCase := framework.GenerateTestCase(
		"validation_test",
		`my $test = "value";`,
		"Validation test case",
		UntypedPerl,
		[]string{"validation"},
	)
	if testCase == nil {
		t.Error("Failed to generate test case")
		allPassed = false
	}
	if testCase.Name != "validation_test" {
		t.Error("Test case name not set correctly")
		allPassed = false
	}

	// Test 3: Error test case generation  
	errorCase := framework.GenerateErrorTestCase(
		"error_validation",
		"invalid syntax",
		"Error validation test",
		"syntax_error",
		ErrorCases,
		[]string{"errors"},
	)
	if errorCase == nil {
		t.Error("Failed to generate error test case")
		allPassed = false
	}
	if !errorCase.ShouldError {
		t.Error("Error test case should have ShouldError = true")
		allPassed = false
	}

	// Test 4: JSON serialization
	data, err := json.Marshal(testCase)
	if err != nil {
		t.Errorf("Failed to serialize test case: %v", err)
		allPassed = false
	}

	var deserialized ParserTestCase
	err = json.Unmarshal(data, &deserialized)
	if err != nil {
		t.Errorf("Failed to deserialize test case: %v", err)
		allPassed = false
	}

	if deserialized.Name != testCase.Name {
		t.Error("Serialization/deserialization failed")
		allPassed = false
	}

	// Test 5: File operations
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	
	err = framework.SaveTestCase(testCase, testFile)
	if err != nil {
		t.Errorf("Failed to save test case: %v", err)
		allPassed = false
	}

	loadedCase, err := framework.LoadTestCase(testFile)
	if err != nil {
		t.Errorf("Failed to load test case: %v", err)
		allPassed = false
	}

	if loadedCase.Name != testCase.Name {
		t.Error("Save/load test case failed")
		allPassed = false
	}

	// Test 6: Accuracy metrics structure
	metrics := &AccuracyMetrics{
		CategoryMetrics: make(map[string]Metric),
		FeatureMetrics:  make(map[string]Metric),
	}

	// Simulate adding metrics
	framework.updateMetrics(metrics, testCase, true)
	framework.updateMetrics(metrics, errorCase, false)
	framework.calculateAccuracyPercentages(metrics)

	if metrics.TotalTests != 2 {
		t.Errorf("Expected 2 total tests, got %d", metrics.TotalTests)
		allPassed = false
	}

	if metrics.PassedTests != 1 {
		t.Errorf("Expected 1 passed test, got %d", metrics.PassedTests)
		allPassed = false
	}

	if len(metrics.CategoryMetrics) == 0 {
		t.Error("No category metrics generated")
		allPassed = false
	}

	// Test 7: AST validation
	mockAST := &ast.AST{
		Path:            "mock.pl",
		Source:          "my $test = 1;",
		TypeAnnotations: []*ast.TypeAnnotation{},
	}

	if !framework.ValidateAST(t, mockAST, "validation_test") {
		t.Error("Valid mock AST failed validation")
		allPassed = false
	}

	if framework.ValidateAST(t, nil, "nil_test") {
		t.Error("Nil AST should fail validation")
		allPassed = false
	}

	return allPassed
}

// ValidateAccuracyMeasurementInfrastructure validates the accuracy measurement components
func ValidateAccuracyMeasurementInfrastructure(t *testing.T) bool {
	t.Helper()
	
	allPassed := true
	tmpDir := t.TempDir()

	// Test accuracy measurement creation
	am := NewAccuracyMeasurement(
		filepath.Join(tmpDir, "testdata"),
		filepath.Join(tmpDir, "baseline.json"),
		filepath.Join(tmpDir, "reports"),
	)

	if am == nil {
		t.Error("Failed to create accuracy measurement")
		allPassed = false
	}

	// Test baseline structure
	baseline := &BaselineAccuracy{
		OverallAccuracy:  85.5,
		CategoryAccuracy: map[string]float64{"test": 90.0},
		FeatureAccuracy:  map[string]float64{"feature1": 80.0},
		TestCoverage: TestCoverageReport{
			TotalFeatures:   10,
			CoveredFeatures: 8,
			CoveragePercent: 80.0,
		},
	}

	// Test baseline save/load
	err := am.SaveBaseline(baseline)
	if err != nil {
		t.Errorf("Failed to save baseline: %v", err)
		allPassed = false
	}

	loadedBaseline, err := am.LoadBaseline()
	if err != nil {
		t.Errorf("Failed to load baseline: %v", err)
		allPassed = false
	}

	if loadedBaseline.OverallAccuracy != baseline.OverallAccuracy {
		t.Error("Baseline save/load failed")
		allPassed = false
	}

	// Test report generation
	err = am.GenerateReport(t, baseline)
	if err != nil {
		t.Errorf("Failed to generate report: %v", err)
		allPassed = false
	}

	// Verify reports were created
	jsonReport := filepath.Join(tmpDir, "reports", "accuracy_report.json")
	textReport := filepath.Join(tmpDir, "reports", "accuracy_report.txt")

	if _, err := os.Stat(jsonReport); err != nil {
		t.Error("JSON report was not created")
		allPassed = false
	}

	if _, err := os.Stat(textReport); err != nil {
		t.Error("Text report was not created")
		allPassed = false
	}

	// Test performance measurement structure (without actual parsing)
	perf := &PerformanceBaseline{
		AverageParseTime: 1000000, // 1ms in nanoseconds  
		MemoryUsage:      1024,
		ThroughputLPS:    100.0,
		ThroughputTPS:    500.0,
	}

	if perf.AverageParseTime <= 0 {
		t.Error("Performance baseline not set correctly")
		allPassed = false
	}

	return allPassed
}

// ValidateTestDataStructure validates the test data directory structure
func ValidateTestDataStructure(t *testing.T, testDataDir string) bool {
	t.Helper()
	
	allPassed := true

	// Check that required directories exist
	requiredDirs := []string{
		"input",
		"untyped-perl",
		"untyped-perl/variables", 
		"untyped-perl/expressions",
		"untyped-perl/control-flow",
		"untyped-perl/subroutines",
		"untyped-perl/packages",
		"typed-perl",
		"typed-perl/simple-annotations",
		"typed-perl/methods-fields",
		"typed-perl/assertions",
		"typed-perl/union-types",
		"typed-perl/parameterized-types",
		"error-cases",
	}

	for _, dir := range requiredDirs {
		fullPath := filepath.Join(testDataDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Required directory %s does not exist", fullPath)
			allPassed = false
		}
	}

	// Check that input files exist if they were created
	inputFiles := []string{"simple_variables.pl", "type_annotations.pl"}
	for _, file := range inputFiles {
		fullPath := filepath.Join(testDataDir, "input", file)
		if _, err := os.Stat(fullPath); err == nil {
			// File exists, validate it has content
			data, err := os.ReadFile(fullPath)
			if err != nil {
				t.Errorf("Cannot read input file %s: %v", fullPath, err)
				allPassed = false
			} else if len(strings.TrimSpace(string(data))) == 0 {
				t.Errorf("Input file %s is empty", fullPath)
				allPassed = false
			}
		}
	}

	return allPassed
}

// PrintInfrastructureStatus prints the status of the test framework infrastructure
func PrintInfrastructureStatus(t *testing.T, testDataDir string) {
	t.Helper()
	
	t.Log("=== Parser Test Framework Infrastructure Status ===")
	
	// Check framework components
	frameworkValid := ValidateTestFrameworkInfrastructure(t)
	t.Logf("Test Framework: %s", status(frameworkValid))
	
	accuracyValid := ValidateAccuracyMeasurementInfrastructure(t)
	t.Logf("Accuracy Measurement: %s", status(accuracyValid))
	
	structureValid := ValidateTestDataStructure(t, testDataDir)
	t.Logf("Test Data Structure: %s", status(structureValid))
	
	overall := frameworkValid && accuracyValid && structureValid
	t.Logf("Overall Infrastructure: %s", status(overall))
	
	if overall {
		t.Log("✅ All infrastructure components are ready for Step 2")
	} else {
		t.Log("❌ Some infrastructure components need attention")
	}
}

func status(valid bool) string {
	if valid {
		return "✅ READY"
	}
	return "❌ NEEDS WORK"
}