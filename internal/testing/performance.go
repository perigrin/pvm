// ABOUTME: Performance monitoring and benchmarking infrastructure for regression detection
// ABOUTME: Provides tools for tracking performance metrics and detecting regressions

package basetesting

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

// BenchmarkResult represents the result of a benchmark
type BenchmarkResult struct {
	Name          string        `json:"name"`
	Iterations    int           `json:"iterations"`
	NsPerOp       int64         `json:"ns_per_op"`
	BytesPerOp    int64         `json:"bytes_per_op,omitempty"`
	AllocsPerOp   int64         `json:"allocs_per_op,omitempty"`
	Duration      time.Duration `json:"duration"`
	MemAllocBytes int64         `json:"mem_alloc_bytes,omitempty"`
	MemSysBytes   int64         `json:"mem_sys_bytes,omitempty"`
}

// PerformanceReport contains benchmark results and analysis
type PerformanceReport struct {
	Timestamp   time.Time         `json:"timestamp"`
	GitCommit   string            `json:"git_commit,omitempty"`
	GitBranch   string            `json:"git_branch,omitempty"`
	GoVersion   string            `json:"go_version"`
	GOOS        string            `json:"goos"`
	GOARCH      string            `json:"goarch"`
	Benchmarks  []BenchmarkResult `json:"benchmarks"`
	Regressions []Regression      `json:"regressions,omitempty"`
}

// Regression represents a performance regression
type Regression struct {
	BenchmarkName   string  `json:"benchmark_name"`
	Metric          string  `json:"metric"`
	OldValue        float64 `json:"old_value"`
	NewValue        float64 `json:"new_value"`
	PercentChange   float64 `json:"percent_change"`
	Severity        string  `json:"severity"`
	ThresholdBreach bool    `json:"threshold_breach"`
}

// PerformanceMonitor manages performance monitoring and regression detection
type PerformanceMonitor struct {
	ReportDir           string
	BaselineFile        string
	ThresholdFile       string
	RegressionThreshold float64 // Default threshold for detecting regressions (e.g., 10% = 0.1)
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(reportDir string) *PerformanceMonitor {
	return &PerformanceMonitor{
		ReportDir:           reportDir,
		BaselineFile:        filepath.Join(reportDir, "baseline.json"),
		ThresholdFile:       filepath.Join(reportDir, "thresholds.json"),
		RegressionThreshold: 0.1, // 10% threshold by default
	}
}

// RunBenchmarkSuite executes a benchmark suite and records results
func (pm *PerformanceMonitor) RunBenchmarkSuite(t *testing.T, suiteName string, benchmarks map[string]func(*testing.B)) {
	var results []BenchmarkResult

	for name, benchFunc := range benchmarks {
		result := pm.runSingleBenchmark(t, name, benchFunc)
		results = append(results, result)
	}

	// Create performance report
	report := PerformanceReport{
		Timestamp:  time.Now(),
		GoVersion:  "go1.21", // Would get from runtime in real implementation
		GOOS:       "darwin", // Would get from runtime
		GOARCH:     "amd64",  // Would get from runtime
		Benchmarks: results,
	}

	// Load baseline if exists and compare
	if baseline, err := pm.loadBaseline(); err == nil {
		report.Regressions = pm.detectRegressions(baseline, results)
	}

	// Save current report
	pm.saveReport(report)

	// Log results
	pm.logResults(t, results, report.Regressions)
}

// runSingleBenchmark executes a single benchmark function
func (pm *PerformanceMonitor) runSingleBenchmark(t *testing.T, name string, benchFunc func(*testing.B)) BenchmarkResult {
	result := testing.Benchmark(benchFunc)

	return BenchmarkResult{
		Name:        name,
		Iterations:  result.N,
		NsPerOp:     result.NsPerOp(),
		BytesPerOp:  result.AllocedBytesPerOp(),
		AllocsPerOp: result.AllocsPerOp(),
		Duration:    time.Duration(result.T),
	}
}

// loadBaseline loads the baseline performance data
func (pm *PerformanceMonitor) loadBaseline() ([]BenchmarkResult, error) {
	data, err := os.ReadFile(pm.BaselineFile)
	if err != nil {
		return nil, err
	}

	var report PerformanceReport
	err = json.Unmarshal(data, &report)
	if err != nil {
		return nil, err
	}

	return report.Benchmarks, nil
}

// saveReport saves the performance report
func (pm *PerformanceMonitor) saveReport(report PerformanceReport) error {
	err := os.MkdirAll(pm.ReportDir, 0755)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	// Save current report with timestamp
	timestamp := report.Timestamp.Format("2006-01-02-15-04-05")
	reportFile := filepath.Join(pm.ReportDir, fmt.Sprintf("report-%s.json", timestamp))
	err = os.WriteFile(reportFile, data, 0644)
	if err != nil {
		return err
	}

	// Update baseline if this is an update run
	if os.Getenv("UPDATE_PERFORMANCE_BASELINE") == "1" {
		return os.WriteFile(pm.BaselineFile, data, 0644)
	}

	return nil
}

// detectRegressions compares current results against baseline
func (pm *PerformanceMonitor) detectRegressions(baseline, current []BenchmarkResult) []Regression {
	var regressions []Regression

	// Create maps for easy lookup
	baselineMap := make(map[string]BenchmarkResult)
	for _, b := range baseline {
		baselineMap[b.Name] = b
	}

	for _, curr := range current {
		if base, exists := baselineMap[curr.Name]; exists {
			// Check various metrics for regressions
			regressions = append(regressions, pm.checkMetricRegression(curr.Name, "ns_per_op",
				float64(base.NsPerOp), float64(curr.NsPerOp))...)

			if base.BytesPerOp > 0 && curr.BytesPerOp > 0 {
				regressions = append(regressions, pm.checkMetricRegression(curr.Name, "bytes_per_op",
					float64(base.BytesPerOp), float64(curr.BytesPerOp))...)
			}

			if base.AllocsPerOp > 0 && curr.AllocsPerOp > 0 {
				regressions = append(regressions, pm.checkMetricRegression(curr.Name, "allocs_per_op",
					float64(base.AllocsPerOp), float64(curr.AllocsPerOp))...)
			}
		}
	}

	return regressions
}

// checkMetricRegression checks if a specific metric has regressed
func (pm *PerformanceMonitor) checkMetricRegression(benchName, metric string, oldValue, newValue float64) []Regression {
	if oldValue == 0 {
		return nil // Can't calculate percentage change
	}

	percentChange := (newValue - oldValue) / oldValue

	// Only report if it's a meaningful regression (positive change = worse performance)
	if percentChange > pm.RegressionThreshold {
		severity := "minor"
		if percentChange > 0.25 { // 25%
			severity = "major"
		} else if percentChange > 0.5 { // 50%
			severity = "critical"
		}

		regression := Regression{
			BenchmarkName:   benchName,
			Metric:          metric,
			OldValue:        oldValue,
			NewValue:        newValue,
			PercentChange:   percentChange,
			Severity:        severity,
			ThresholdBreach: percentChange > pm.RegressionThreshold,
		}

		return []Regression{regression}
	}

	return nil
}

// logResults logs benchmark results and regressions
func (pm *PerformanceMonitor) logResults(t *testing.T, results []BenchmarkResult, regressions []Regression) {
	t.Helper()

	// Log benchmark results
	t.Logf("=== Performance Results ===")
	for _, result := range results {
		t.Logf("%-30s %8d ops %8d ns/op", result.Name, result.Iterations, result.NsPerOp)
		if result.BytesPerOp > 0 {
			t.Logf("  %-26s %8d B/op", "", result.BytesPerOp)
		}
		if result.AllocsPerOp > 0 {
			t.Logf("  %-26s %8d allocs/op", "", result.AllocsPerOp)
		}
	}

	// Log regressions
	if len(regressions) > 0 {
		t.Logf("=== Performance Regressions Detected ===")
		for _, reg := range regressions {
			t.Errorf("REGRESSION in %s.%s: %.1f%% slower (%.0f -> %.0f) [%s]",
				reg.BenchmarkName, reg.Metric, reg.PercentChange*100,
				reg.OldValue, reg.NewValue, reg.Severity)
		}
	} else {
		t.Logf("No performance regressions detected")
	}
}

// GeneratePerformanceReport creates a human-readable performance report
func (pm *PerformanceMonitor) GeneratePerformanceReport() (string, error) {
	data, err := os.ReadFile(pm.BaselineFile)
	if err != nil {
		return "", fmt.Errorf("failed to read baseline: %w", err)
	}

	var report PerformanceReport
	err = json.Unmarshal(data, &report)
	if err != nil {
		return "", fmt.Errorf("failed to parse baseline: %w", err)
	}

	var output []string
	output = append(output, "# Performance Report")
	output = append(output, fmt.Sprintf("Generated: %s", report.Timestamp.Format("2006-01-02 15:04:05")))
	output = append(output, fmt.Sprintf("Go Version: %s", report.GoVersion))
	output = append(output, fmt.Sprintf("Platform: %s/%s", report.GOOS, report.GOARCH))
	output = append(output, "")

	// Sort benchmarks by name for consistent output
	sort.Slice(report.Benchmarks, func(i, j int) bool {
		return report.Benchmarks[i].Name < report.Benchmarks[j].Name
	})

	output = append(output, "## Benchmark Results")
	output = append(output, "")
	output = append(output, "| Benchmark | Iterations | ns/op | B/op | allocs/op |")
	output = append(output, "|-----------|------------|-------|------|-----------|")

	for _, bench := range report.Benchmarks {
		line := fmt.Sprintf("| %s | %d | %d | %d | %d |",
			bench.Name, bench.Iterations, bench.NsPerOp, bench.BytesPerOp, bench.AllocsPerOp)
		output = append(output, line)
	}

	if len(report.Regressions) > 0 {
		output = append(output, "")
		output = append(output, "## Performance Regressions")
		output = append(output, "")
		for _, reg := range report.Regressions {
			line := fmt.Sprintf("- **%s**: %s regression of %.1f%% in %s (%.0f -> %.0f)",
				reg.BenchmarkName, reg.Severity, reg.PercentChange*100, reg.Metric,
				reg.OldValue, reg.NewValue)
			output = append(output, line)
		}
	}

	return fmt.Sprintf("%s\n", fmt.Sprintf("%s", output)), nil
}

// BenchmarkHelper provides utilities for common benchmark patterns
type BenchmarkHelper struct {
	monitor *PerformanceMonitor
}

// NewBenchmarkHelper creates a new benchmark helper
func NewBenchmarkHelper(monitor *PerformanceMonitor) *BenchmarkHelper {
	return &BenchmarkHelper{monitor: monitor}
}

// BenchmarkParser benchmarks parsing operations
func (bh *BenchmarkHelper) BenchmarkParser(b *testing.B, name string, parseFunc func() error) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := parseFunc()
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}

// BenchmarkMemoryUsage measures memory usage of operations
func (bh *BenchmarkHelper) BenchmarkMemoryUsage(b *testing.B, name string, operation func()) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		operation()
	}
}
