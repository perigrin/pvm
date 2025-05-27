// ABOUTME: Performance regression detection tool for CI/CD monitoring
// ABOUTME: Compares current performance against baselines and detects regressions

//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// BenchmarkData represents performance data for a single benchmark
type BenchmarkData struct {
	Name        string  `json:"name"`
	OpsPerSec   int64   `json:"ops_per_sec"`
	NsPerOp     int64   `json:"ns_per_op"`
	MemoryMB    float64 `json:"memory_mb"`
	AllocsPerOp int64   `json:"allocs_per_op,omitempty"`
	Change      string  `json:"change,omitempty"`
}

// PerformanceReport represents the complete performance analysis
type PerformanceReport struct {
	Benchmarks  []BenchmarkData `json:"benchmarks"`
	Regressions []Regression    `json:"regressions,omitempty"`
}

// Regression represents a performance regression
type Regression struct {
	Component string `json:"component"`
	Change    string `json:"change"`
	Severity  string `json:"severity"`
}

// BaselineData represents historical performance data
type BaselineData struct {
	Name      string  `json:"name"`
	OpsPerSec int64   `json:"ops_per_sec"`
	MemoryMB  float64 `json:"memory_mb"`
}

const (
	// Performance regression thresholds
	OpsThresholdModerate = 0.10 // 10% decrease in ops/sec
	OpsThresholdSevere   = 0.25 // 25% decrease in ops/sec
	MemThresholdModerate = 0.20 // 20% increase in memory usage
	MemThresholdSevere   = 0.50 // 50% increase in memory usage
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run check_performance_regression.go <performance-report.json>")
	}

	reportFile := os.Args[1]

	// Read current performance report
	report, err := readPerformanceReport(reportFile)
	if err != nil {
		log.Fatalf("Failed to read performance report: %v", err)
	}

	// Read baseline data
	baselines, err := readBaselines()
	if err != nil {
		log.Printf("Warning: could not read baselines: %v", err)
		// Continue without baselines for first run
		baselines = make(map[string]BaselineData)
	}

	// Check for regressions
	regressions := checkRegressions(report.Benchmarks, baselines)

	// Update report with regression information
	report.Regressions = regressions

	// Update baselines with current data
	updateBaselines(report.Benchmarks)

	// Output updated report
	output, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal updated report: %v", err)
	}

	// Write back to file
	if err := os.WriteFile(reportFile, output, 0644); err != nil {
		log.Fatalf("Failed to write updated report: %v", err)
	}

	// Exit with error code if severe regressions found
	if hasSevereRegressions(regressions) {
		fmt.Println("❌ Severe performance regressions detected!")
		for _, reg := range regressions {
			if reg.Severity == "severe" {
				fmt.Printf("  - %s: %s\n", reg.Component, reg.Change)
			}
		}
		os.Exit(1)
	}

	if len(regressions) > 0 {
		fmt.Println("⚠️  Performance regressions detected:")
		for _, reg := range regressions {
			fmt.Printf("  - %s: %s (%s)\n", reg.Component, reg.Change, reg.Severity)
		}
	} else {
		fmt.Println("✅ No performance regressions detected")
	}
}

func readPerformanceReport(filename string) (*PerformanceReport, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var report PerformanceReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}

	return &report, nil
}

func readBaselines() (map[string]BaselineData, error) {
	baselineDir := "testdata/performance"
	baselineFile := filepath.Join(baselineDir, "baseline.json")

	data, err := os.ReadFile(baselineFile)
	if err != nil {
		return nil, err
	}

	var baselines []BaselineData
	if err := json.Unmarshal(data, &baselines); err != nil {
		return nil, err
	}

	baselineMap := make(map[string]BaselineData)
	for _, baseline := range baselines {
		baselineMap[baseline.Name] = baseline
	}

	return baselineMap, nil
}

func checkRegressions(benchmarks []BenchmarkData, baselines map[string]BaselineData) []Regression {
	var regressions []Regression

	for _, bench := range benchmarks {
		baseline, exists := baselines[bench.Name]
		if !exists {
			continue // No baseline to compare against
		}

		// Check operations per second regression
		if baseline.OpsPerSec > 0 {
			opsChange := float64(baseline.OpsPerSec-bench.OpsPerSec) / float64(baseline.OpsPerSec)

			if opsChange >= OpsThresholdSevere {
				regressions = append(regressions, Regression{
					Component: bench.Name,
					Change:    fmt.Sprintf("%.1f%% decrease in ops/sec", opsChange*100),
					Severity:  "severe",
				})
			} else if opsChange >= OpsThresholdModerate {
				regressions = append(regressions, Regression{
					Component: bench.Name,
					Change:    fmt.Sprintf("%.1f%% decrease in ops/sec", opsChange*100),
					Severity:  "moderate",
				})
			}
		}

		// Check memory usage regression
		if baseline.MemoryMB > 0 {
			memChange := (bench.MemoryMB - baseline.MemoryMB) / baseline.MemoryMB

			if memChange >= MemThresholdSevere {
				regressions = append(regressions, Regression{
					Component: bench.Name,
					Change:    fmt.Sprintf("%.1f%% increase in memory usage", memChange*100),
					Severity:  "severe",
				})
			} else if memChange >= MemThresholdModerate {
				regressions = append(regressions, Regression{
					Component: bench.Name,
					Change:    fmt.Sprintf("%.1f%% increase in memory usage", memChange*100),
					Severity:  "moderate",
				})
			}
		}
	}

	return regressions
}

func updateBaselines(benchmarks []BenchmarkData) {
	baselineDir := "testdata/performance"
	baselineFile := filepath.Join(baselineDir, "baseline.json")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(baselineDir, 0755); err != nil {
		log.Printf("Warning: could not create baseline directory: %v", err)
		return
	}

	// Convert benchmarks to baseline data
	var baselines []BaselineData
	for _, bench := range benchmarks {
		baselines = append(baselines, BaselineData{
			Name:      bench.Name,
			OpsPerSec: bench.OpsPerSec,
			MemoryMB:  bench.MemoryMB,
		})
	}

	// Write updated baselines
	data, err := json.MarshalIndent(baselines, "", "  ")
	if err != nil {
		log.Printf("Warning: could not marshal baselines: %v", err)
		return
	}

	if err := os.WriteFile(baselineFile, data, 0644); err != nil {
		log.Printf("Warning: could not write baselines: %v", err)
	}
}

func hasSevereRegressions(regressions []Regression) bool {
	for _, reg := range regressions {
		if reg.Severity == "severe" {
			return true
		}
	}
	return false
}
