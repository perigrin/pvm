// ABOUTME: Performance report generation tool for CI/CD benchmarking
// ABOUTME: Parses benchmark output and generates structured performance reports

//go:build ignore

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
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
	Timestamp   time.Time       `json:"timestamp"`
	GitCommit   string          `json:"git_commit,omitempty"`
	GitBranch   string          `json:"git_branch,omitempty"`
	Benchmarks  []BenchmarkData `json:"benchmarks"`
	Regressions []Regression    `json:"regressions,omitempty"`
	Summary     Summary         `json:"summary"`
}

// Regression represents a performance regression
type Regression struct {
	Component string `json:"component"`
	Change    string `json:"change"`
	Severity  string `json:"severity"`
}

// Summary provides overall performance metrics
type Summary struct {
	TotalBenchmarks  int     `json:"total_benchmarks"`
	AverageOpsPerSec int64   `json:"average_ops_per_sec"`
	TotalMemoryMB    float64 `json:"total_memory_mb"`
	RegressionsCount int     `json:"regressions_count"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run generate_performance_report.go <benchmark-results.txt>")
	}

	inputFile := os.Args[1]

	benchmarks := func() []BenchmarkData {
		file, err := os.Open(inputFile)
		if err != nil {
			log.Fatalf("Failed to open input file: %v", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("Warning: failed to close file: %v", err)
			}
		}()

		// Parse benchmark results
		return parseBenchmarkResults(file)
	}()

	// Generate report
	report := PerformanceReport{
		Timestamp:  time.Now(),
		GitCommit:  os.Getenv("GITHUB_SHA"),
		GitBranch:  os.Getenv("GITHUB_REF_NAME"),
		Benchmarks: benchmarks,
		Summary:    generateSummary(benchmarks),
	}

	// Output JSON report
	output, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal report: %v", err)
	}

	fmt.Println(string(output))
}

func parseBenchmarkResults(file *os.File) []BenchmarkData {
	var benchmarks []BenchmarkData
	scanner := bufio.NewScanner(file)

	// Regex to match Go benchmark output
	// Example: BenchmarkTypeChecker_Performance-8   	     100	  10234567 ns/op	    1234 B/op	      12 allocs/op
	benchmarkRegex := regexp.MustCompile(`^Benchmark(\w+)-\d+\s+(\d+)\s+(\d+)\s+ns/op(?:\s+(\d+)\s+B/op)?(?:\s+(\d+)\s+allocs/op)?`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		matches := benchmarkRegex.FindStringSubmatch(line)
		if len(matches) >= 4 {
			name := matches[1]
			_, _ = strconv.ParseInt(matches[2], 10, 64) // iterations (not used)
			nsPerOp, _ := strconv.ParseInt(matches[3], 10, 64)

			var memoryBytes int64
			var allocsPerOp int64

			if len(matches) > 4 && matches[4] != "" {
				memoryBytes, _ = strconv.ParseInt(matches[4], 10, 64)
			}

			if len(matches) > 5 && matches[5] != "" {
				allocsPerOp, _ = strconv.ParseInt(matches[5], 10, 64)
			}

			// Calculate operations per second
			var opsPerSec int64
			if nsPerOp > 0 {
				opsPerSec = 1_000_000_000 / nsPerOp
			}

			benchmark := BenchmarkData{
				Name:        name,
				OpsPerSec:   opsPerSec,
				NsPerOp:     nsPerOp,
				MemoryMB:    float64(memoryBytes) / (1024 * 1024),
				AllocsPerOp: allocsPerOp,
			}

			benchmarks = append(benchmarks, benchmark)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Warning: error reading file: %v", err)
	}

	return benchmarks
}

func generateSummary(benchmarks []BenchmarkData) Summary {
	if len(benchmarks) == 0 {
		return Summary{}
	}

	var totalOps int64
	var totalMemory float64

	for _, b := range benchmarks {
		totalOps += b.OpsPerSec
		totalMemory += b.MemoryMB
	}

	return Summary{
		TotalBenchmarks:  len(benchmarks),
		AverageOpsPerSec: totalOps / int64(len(benchmarks)),
		TotalMemoryMB:    totalMemory,
		RegressionsCount: 0, // Will be filled by regression checker
	}
}
