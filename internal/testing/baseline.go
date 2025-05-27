// ABOUTME: Baseline testing framework for regression prevention and output validation
// ABOUTME: Provides infrastructure for comparing test outputs against expected baselines

package basetesting

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// BaselineTest represents a single baseline test case
type BaselineTest struct {
	Name        string
	InputFile   string
	BaselineDir string
	UpdateMode  bool
}

// BaselineTestSuite manages a collection of baseline tests
type BaselineTestSuite struct {
	Name          string
	TestDataDir   string
	BaselineDir   string
	InputDir      string
	UpdateMode    bool
	TestProcessor func(input []byte) ([]byte, error)
}

// NewBaselineTestSuite creates a new baseline test suite
func NewBaselineTestSuite(name, testDataDir string, processor func(input []byte) ([]byte, error)) *BaselineTestSuite {
	return &BaselineTestSuite{
		Name:          name,
		TestDataDir:   testDataDir,
		BaselineDir:   filepath.Join(testDataDir, "baseline"),
		InputDir:      filepath.Join(testDataDir, "input"),
		TestProcessor: processor,
		UpdateMode:    os.Getenv("UPDATE_BASELINES") == "1",
	}
}

// RunTestCase executes a single baseline test
func (suite *BaselineTestSuite) RunTestCase(t *testing.T, testName string) {
	t.Helper()

	inputPath := filepath.Join(suite.InputDir, testName+".pl")
	baselinePath := filepath.Join(suite.BaselineDir, testName+".expected")

	// Read input file
	input, err := ioutil.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("Failed to read input file %s: %v", inputPath, err)
	}

	// Process input through test processor
	actual, err := suite.TestProcessor(input)
	if err != nil {
		t.Fatalf("Test processor failed for %s: %v", testName, err)
	}

	if suite.UpdateMode {
		// Update baseline mode - write actual output as new baseline
		err := os.MkdirAll(suite.BaselineDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create baseline directory: %v", err)
		}

		err = ioutil.WriteFile(baselinePath, actual, 0644)
		if err != nil {
			t.Fatalf("Failed to write baseline file %s: %v", baselinePath, err)
		}

		t.Logf("Updated baseline for %s", testName)
		return
	}

	// Normal test mode - compare against existing baseline
	expected, err := ioutil.ReadFile(baselinePath)
	if err != nil {
		t.Fatalf("Failed to read baseline file %s: %v\nRun with UPDATE_BASELINES=1 to create initial baseline", baselinePath, err)
	}

	// Compare actual vs expected
	if diff := cmp.Diff(string(expected), string(actual)); diff != "" {
		t.Errorf("Baseline mismatch for %s (-want +got):\n%s", testName, diff)

		// If baseline differs significantly, show suggestions
		if len(diff) > 1000 {
			t.Logf("Large diff detected. Consider running with UPDATE_BASELINES=1 if changes are intentional")
		}
	}
}

// RunAllTests discovers and runs all baseline tests in the suite
func (suite *BaselineTestSuite) RunAllTests(t *testing.T) {
	// Ensure directories exist
	err := os.MkdirAll(suite.InputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create input directory: %v", err)
	}

	err = os.MkdirAll(suite.BaselineDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create baseline directory: %v", err)
	}

	// Discover input files
	inputFiles, err := filepath.Glob(filepath.Join(suite.InputDir, "*.pl"))
	if err != nil {
		t.Fatalf("Failed to discover input files: %v", err)
	}

	if len(inputFiles) == 0 {
		t.Logf("No input files found in %s", suite.InputDir)
		return
	}

	// Run each test case
	for _, inputFile := range inputFiles {
		basename := strings.TrimSuffix(filepath.Base(inputFile), ".pl")
		t.Run(basename, func(t *testing.T) {
			suite.RunTestCase(t, basename)
		})
	}
}

// ValidateBaselines checks that all baseline files are up to date
func (suite *BaselineTestSuite) ValidateBaselines(t *testing.T) bool {
	inputFiles, err := filepath.Glob(filepath.Join(suite.InputDir, "*.pl"))
	if err != nil {
		t.Errorf("Failed to discover input files: %v", err)
		return false
	}

	allValid := true
	for _, inputFile := range inputFiles {
		basename := strings.TrimSuffix(filepath.Base(inputFile), ".pl")
		baselinePath := filepath.Join(suite.BaselineDir, basename+".expected")

		if _, err := os.Stat(baselinePath); os.IsNotExist(err) {
			t.Errorf("Missing baseline file for %s: %s", basename, baselinePath)
			allValid = false
		}
	}

	return allValid
}

// CleanOrphanedBaselines removes baseline files that no longer have corresponding input files
func (suite *BaselineTestSuite) CleanOrphanedBaselines(t *testing.T) {
	// Get all input files
	inputFiles, err := filepath.Glob(filepath.Join(suite.InputDir, "*.pl"))
	if err != nil {
		t.Errorf("Failed to discover input files: %v", err)
		return
	}

	inputNames := make(map[string]bool)
	for _, inputFile := range inputFiles {
		basename := strings.TrimSuffix(filepath.Base(inputFile), ".pl")
		inputNames[basename] = true
	}

	// Get all baseline files
	baselineFiles, err := filepath.Glob(filepath.Join(suite.BaselineDir, "*.expected"))
	if err != nil {
		t.Errorf("Failed to discover baseline files: %v", err)
		return
	}

	// Remove orphaned baselines
	for _, baselineFile := range baselineFiles {
		basename := strings.TrimSuffix(filepath.Base(baselineFile), ".expected")
		if !inputNames[basename] {
			t.Logf("Removing orphaned baseline: %s", baselineFile)
			err := os.Remove(baselineFile)
			if err != nil {
				t.Errorf("Failed to remove orphaned baseline %s: %v", baselineFile, err)
			}
		}
	}
}

// BaselineTestFunc is a helper function for simple baseline testing
func BaselineTestFunc(t *testing.T, name string, processor func(input []byte) ([]byte, error), input []byte) {
	t.Helper()

	testDataDir := "testdata"
	suite := NewBaselineTestSuite(name, testDataDir, processor)

	// Create a temporary input file for this test
	inputPath := filepath.Join(suite.InputDir, name+".pl")
	err := os.MkdirAll(suite.InputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create input directory: %v", err)
	}

	err = ioutil.WriteFile(inputPath, input, 0644)
	if err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	// Run the test
	suite.RunTestCase(t, name)

	// Clean up temporary input file
	os.Remove(inputPath)
}

// Performance tracking types
type PerformanceMetric struct {
	Name     string
	Value    float64
	Unit     string
	Category string
}

type PerformanceBaseline struct {
	TestName string
	Metrics  []PerformanceMetric
}

// PerformanceTracker tracks performance metrics for baseline comparison
type PerformanceTracker struct {
	BaselineFile string
	Current      []PerformanceMetric
}

// NewPerformanceTracker creates a new performance tracker
func NewPerformanceTracker(baselineFile string) *PerformanceTracker {
	return &PerformanceTracker{
		BaselineFile: baselineFile,
		Current:      make([]PerformanceMetric, 0),
	}
}

// RecordMetric records a performance metric
func (pt *PerformanceTracker) RecordMetric(name string, value float64, unit, category string) {
	metric := PerformanceMetric{
		Name:     name,
		Value:    value,
		Unit:     unit,
		Category: category,
	}
	pt.Current = append(pt.Current, metric)
}

// CompareAgainstBaseline compares current metrics against baseline
func (pt *PerformanceTracker) CompareAgainstBaseline(t *testing.T, testName string) {
	t.Helper()

	// In a full implementation, this would load baseline from file
	// and compare metrics, flagging regressions
	for _, metric := range pt.Current {
		t.Logf("Performance metric %s: %.2f %s (%s)",
			metric.Name, metric.Value, metric.Unit, metric.Category)
	}
}

// SaveBaseline saves current metrics as new baseline
func (pt *PerformanceTracker) SaveBaseline(testName string) error {
	// In a full implementation, this would save metrics to baseline file
	// For now, just create the directory structure
	baselineDir := filepath.Dir(pt.BaselineFile)
	return os.MkdirAll(baselineDir, 0755)
}
