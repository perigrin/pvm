// ABOUTME: Performance regression testing implementation for Step 23
// ABOUTME: Automated detection of performance regressions in parser improvements

package parser

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	basetesting "tamarou.com/pvm/internal/testing"
)

// BenchmarkParser_PerformanceRegression runs comprehensive performance benchmarks
func BenchmarkParser_PerformanceRegression(b *testing.B) {
	// Create test directories
	tempDir := b.TempDir()
	baselineDir := filepath.Join(tempDir, "baselines")
	reportDir := filepath.Join(tempDir, "reports")

	// Create performance test suite
	suite := NewPerformanceTestSuite("", baselineDir, reportDir)

	// Create parser for testing
	parser, err := NewParser()
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}
	suite.Parser = parser

	// Generate performance tests
	tests := suite.GeneratePerformanceTests()

	for _, test := range tests {
		b.Run(test.Name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := parser.ParseString(test.InputCode)
				if err != nil {
					b.Fatalf("Parse failed: %v", err)
				}
			}
		})
	}
}

// TestParser_PerformanceValidation validates parser performance meets requirements
func TestParser_PerformanceValidation(t *testing.T) {
	basetesting.SkipUnlessPerformance(t, "parser performance validation")

	// Create test directories
	tempDir := t.TempDir()
	baselineDir := filepath.Join(tempDir, "baselines")
	reportDir := filepath.Join(tempDir, "reports")

	// Create performance test suite
	suite := NewPerformanceTestSuite("", baselineDir, reportDir)

	// Create parser for testing
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	suite.Parser = parser

	// Run all performance tests
	report := suite.RunAllPerformanceTests(t)

	// Print report
	suite.PrintPerformanceReport(t, report)

	// Save report
	if err := suite.SaveReport(report); err != nil {
		t.Errorf("Failed to save performance report: %v", err)
	}

	// Validate performance requirements
	if report.Summary.FailedTests > 0 {
		t.Errorf("Performance validation failed: %d tests failed", report.Summary.FailedTests)
	}

	// Check for regressions
	if report.Summary.RegressionsFound > 0 {
		t.Errorf("Performance regressions detected: %d regressions found", report.Summary.RegressionsFound)
	}

	// Validate overall performance metrics
	maxAvgParseTime := 100 * time.Millisecond
	if report.Summary.AvgParseTime > maxAvgParseTime {
		t.Errorf("Average parse time %v exceeds maximum %v", report.Summary.AvgParseTime, maxAvgParseTime)
	}

	maxAvgMemory := int64(100 * 1024 * 1024) // 100MB average across all tests
	if report.Summary.AvgMemoryUsed > maxAvgMemory {
		t.Errorf("Average memory usage %d exceeds maximum %d", report.Summary.AvgMemoryUsed, maxAvgMemory)
	}
}

// TestParser_MemoryStability tests for memory leaks and stability
func TestParser_MemoryStability(t *testing.T) {
	basetesting.SkipUnlessLongRunning(t, "parser memory stability test")

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	testCodes := []string{
		"my Int $simple = 42;",
		"my ArrayRef[Str] $array = [];",
		"method Bool test(Int $a, Str $b) { return 1; }",
		"type MyType = Int|Str; my MyType $var = 123;",
		"class Test { field Int $id; method Test new() { return bless {}, __PACKAGE__; } }",
	}

	// Test for memory stability over many iterations
	const iterations = 1000
	initialMemory := getMemoryUsage()

	for i := 0; i < iterations; i++ {
		for _, code := range testCodes {
			_, err := parser.ParseString(code)
			if err != nil {
				t.Fatalf("Parse failed on iteration %d: %v", i, err)
			}
		}

		// Force GC every 100 iterations
		if i%100 == 0 {
			forceGC()
		}
	}

	finalMemory := getMemoryUsage()
	memoryGrowth := finalMemory - initialMemory

	t.Logf("Memory usage: initial=%d, final=%d, growth=%d", initialMemory, finalMemory, memoryGrowth)

	// Allow some memory growth but not excessive
	maxMemoryGrowth := int64(50 * 1024 * 1024) // 50MB
	if memoryGrowth > maxMemoryGrowth {
		t.Errorf("Excessive memory growth: %d bytes (max: %d)", memoryGrowth, maxMemoryGrowth)
	}
}

// TestParser_PerformanceBaselines establishes or updates performance baselines
func TestParser_PerformanceBaselines(t *testing.T) {
	// Only run when explicitly requested
	if os.Getenv("UPDATE_PERFORMANCE_BASELINES") != "1" {
		t.Skip("Skipping baseline update (set UPDATE_PERFORMANCE_BASELINES=1 to run)")
	}

	// Create test directories
	tempDir := t.TempDir()
	baselineDir := filepath.Join(tempDir, "baselines")
	reportDir := filepath.Join(tempDir, "reports")

	// Create performance test suite
	suite := NewPerformanceTestSuite("", baselineDir, reportDir)

	// Create parser for testing
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	suite.Parser = parser

	// Run all performance tests
	report := suite.RunAllPerformanceTests(t)

	// Save baselines
	if err := suite.SaveBaselines(report.TestResults); err != nil {
		t.Fatalf("Failed to save performance baselines: %v", err)
	}

	t.Logf("Performance baselines updated successfully")
	suite.PrintPerformanceReport(t, report)
}

// TestParser_StressTest performs stress testing with extreme inputs
func TestParser_StressTest(t *testing.T) {
	basetesting.SkipUnlessStress(t, "parser stress test with extreme inputs")

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	stressTests := []struct {
		name      string
		code      string
		maxTime   time.Duration
		maxMemory int64
	}{
		{
			name:      "deeply_nested_types",
			code:      generateDeeplyNestedTypes(20), // 20 levels deep
			maxTime:   5 * time.Second,
			maxMemory: 100 * 1024 * 1024, // 100MB
		},
		{
			name:      "very_long_union",
			code:      generateLongUnionType(100), // 100 union members
			maxTime:   5 * time.Second,
			maxMemory: 50 * 1024 * 1024, // 50MB
		},
		{
			name:      "massive_method_signature",
			code:      generateMassiveMethodSignature(200), // 200 parameters
			maxTime:   5 * time.Second,
			maxMemory: 50 * 1024 * 1024, // 50MB
		},
		{
			name:      "large_class_hierarchy",
			code:      generateLargeClassHierarchy(100), // 100 classes
			maxTime:   10 * time.Second,
			maxMemory: 200 * 1024 * 1024, // 200MB
		},
	}

	for _, test := range stressTests {
		t.Run(test.name, func(t *testing.T) {
			// Measure memory before
			memBefore := getMemoryUsage()

			// Parse with timeout
			start := time.Now()
			_, err := parser.ParseString(test.code)
			duration := time.Since(start)

			// Measure memory after
			memAfter := getMemoryUsage()
			memUsed := memAfter - memBefore

			if err != nil {
				t.Logf("Stress test %s failed (this may be expected): %v", test.name, err)
			}

			t.Logf("Stress test %s: duration=%v, memory=%d bytes", test.name, duration, memUsed)

			// Check time limits
			if duration > test.maxTime {
				t.Errorf("Stress test %s exceeded time limit: %v > %v", test.name, duration, test.maxTime)
			}

			// Check memory limits
			if memUsed > test.maxMemory {
				t.Errorf("Stress test %s exceeded memory limit: %d > %d", test.name, memUsed, test.maxMemory)
			}
		})
	}
}

// Helper functions for stress test generation

func generateDeeplyNestedTypes(depth int) string {
	result := "my "
	for i := 0; i < depth; i++ {
		result += "ArrayRef["
	}
	result += "Int"
	for i := 0; i < depth; i++ {
		result += "]"
	}
	result += " $deeply_nested;"
	return result
}

func generateLongUnionType(members int) string {
	result := "my "
	for i := 0; i < members; i++ {
		if i > 0 {
			result += "|"
		}
		result += "Type" + string(rune('A'+(i%26)))
		if i >= 26 {
			result += string(rune('0' + ((i / 26) % 10)))
		}
	}
	result += " $long_union;"
	return result
}

func generateMassiveMethodSignature(paramCount int) string {
	result := "method massive_method("
	for i := 0; i < paramCount; i++ {
		if i > 0 {
			result += ", "
		}
		result += "Int $param" + string(rune('0'+(i%10)))
		if i >= 10 {
			result += string(rune('A' + ((i / 10) % 26)))
		}
	}
	result += ") -> Bool { return 1; }"
	return result
}

func generateLargeClassHierarchy(classCount int) string {
	result := ""
	for i := 0; i < classCount; i++ {
		result += "class TestClass" + string(rune('A'+(i%26)))
		if i >= 26 {
			result += string(rune('0' + ((i / 26) % 10)))
		}
		result += " {\n"
		result += "    field Int $id" + string(rune('0'+(i%10))) + ";\n"
		result += "    field Str $name;\n"
		result += "    method Self new() { return bless {}, __PACKAGE__; }\n"
		result += "}\n\n"
	}
	return result
}

// Helper functions for memory measurement

func getMemoryUsage() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Alloc)
}

func forceGC() {
	runtime.GC()
	runtime.GC() // Call twice to ensure cleanup
}
