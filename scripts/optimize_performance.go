// ABOUTME: Performance optimization runner that applies and measures optimization effectiveness
// ABOUTME: Integrates profiling, optimization, and performance measurement for PVM components

//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/performance"
)

// OptimizationReport contains comprehensive optimization results
type OptimizationReport struct {
	Timestamp         time.Time                      `json:"timestamp"`
	BaselineMetrics   *PerformanceMetrics            `json:"baseline_metrics"`
	OptimizedMetrics  *PerformanceMetrics            `json:"optimized_metrics"`
	ImprovementFactor float64                        `json:"improvement_factor"`
	Recommendations   []string                       `json:"recommendations"`
	OptimizationStats *performance.OptimizationStats `json:"optimization_stats"`
}

// PerformanceMetrics represents performance measurements
type PerformanceMetrics struct {
	ParseTime        time.Duration `json:"parse_time"`
	MemoryUsage      uint64        `json:"memory_usage"`
	AllocationsCount int64         `json:"allocations_count"`
	OperationsPerSec float64       `json:"operations_per_sec"`
}

func main() {
	fmt.Println("🚀 Starting PVM Performance Optimization...")

	// Test content for optimization
	testContent := `my Int $count = 42;
my Str $name = "test";
my ArrayRef[Int] $numbers = [1, 2, 3, 4, 5];

sub Int add(Int $a, Int $b) {
    return $a + $b;
}

my $result = add($count, 10);`

	// Measure baseline performance
	fmt.Println("📊 Measuring baseline performance...")
	baselineMetrics, err := measureBaselinePerformance(testContent)
	if err != nil {
		log.Fatalf("Failed to measure baseline performance: %v", err)
	}

	// Apply optimizations and measure
	fmt.Println("⚡ Applying optimizations and measuring...")
	optimizedMetrics, optimizationStats, err := measureOptimizedPerformance(testContent)
	if err != nil {
		log.Fatalf("Failed to measure optimized performance: %v", err)
	}

	// Calculate improvement
	improvementFactor := calculateImprovement(baselineMetrics, optimizedMetrics)

	// Generate recommendations
	recommendations := generateRecommendations(baselineMetrics, optimizedMetrics, optimizationStats)

	// Create report
	report := &OptimizationReport{
		Timestamp:         time.Now(),
		BaselineMetrics:   baselineMetrics,
		OptimizedMetrics:  optimizedMetrics,
		ImprovementFactor: improvementFactor,
		Recommendations:   recommendations,
		OptimizationStats: optimizationStats,
	}

	// Output results
	fmt.Println("📈 Performance Optimization Results:")
	fmt.Printf("   Baseline parse time: %v\n", baselineMetrics.ParseTime)
	fmt.Printf("   Optimized parse time: %v\n", optimizedMetrics.ParseTime)
	fmt.Printf("   Improvement factor: %.2fx\n", improvementFactor)
	fmt.Printf("   Cache hit rate: %.1f%%\n", optimizationStats.ParseCacheHitRate*100)
	fmt.Printf("   Fast parse rate: %.1f%%\n", optimizationStats.FastParsePercentage)

	// Save detailed report
	reportFile := "performance-optimization-report.json"
	if err := saveReport(report, reportFile); err != nil {
		log.Printf("Warning: failed to save report: %v", err)
	} else {
		fmt.Printf("📄 Detailed report saved to: %s\n", reportFile)
	}

	// Show recommendations
	if len(recommendations) > 0 {
		fmt.Println("💡 Optimization Recommendations:")
		for i, rec := range recommendations {
			fmt.Printf("   %d. %s\n", i+1, rec)
		}
	}

	fmt.Println("✅ Performance optimization analysis complete!")
}

func measureBaselinePerformance(content string) (*PerformanceMetrics, error) {
	parser, err := parser.NewParser()
	if err != nil {
		return nil, err
	}

	iterations := 1000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		_, err := parser.ParseString(content)
		if err != nil {
			return nil, err
		}
	}

	totalTime := time.Since(start)
	avgTime := totalTime / time.Duration(iterations)
	opsPerSec := float64(iterations) / totalTime.Seconds()

	return &PerformanceMetrics{
		ParseTime:        avgTime,
		OperationsPerSec: opsPerSec,
		// Memory metrics would require more sophisticated measurement
	}, nil
}

func measureOptimizedPerformance(content string) (*PerformanceMetrics, *performance.OptimizationStats, error) {
	baseParser, err := parser.NewParser()
	if err != nil {
		return nil, nil, err
	}

	config := performance.DefaultOptimizationConfig()
	pipeline := performance.NewOptimizedPipeline(baseParser, config)
	defer pipeline.Shutdown()

	iterations := 1000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		_, err := pipeline.ProcessFile("test.pl", content)
		if err != nil {
			return nil, nil, err
		}
	}

	totalTime := time.Since(start)
	avgTime := totalTime / time.Duration(iterations)
	opsPerSec := float64(iterations) / totalTime.Seconds()

	stats := pipeline.GetOptimizationStats()

	return &PerformanceMetrics{
		ParseTime:        avgTime,
		OperationsPerSec: opsPerSec,
	}, stats, nil
}

func calculateImprovement(baseline, optimized *PerformanceMetrics) float64 {
	if baseline.ParseTime == 0 {
		return 1.0
	}
	return float64(baseline.ParseTime) / float64(optimized.ParseTime)
}

func generateRecommendations(baseline, optimized *PerformanceMetrics, stats *performance.OptimizationStats) []string {
	var recommendations []string

	// Check cache effectiveness
	if stats.ParseCacheHitRate < 0.5 {
		recommendations = append(recommendations,
			"Increase cache size - current hit rate is low")
	}

	// Check fast parser usage
	if stats.FastParsePercentage < 10 {
		recommendations = append(recommendations,
			"Optimize content patterns for fast parser usage")
	}

	// Check overall performance
	improvementFactor := calculateImprovement(baseline, optimized)
	if improvementFactor < 1.1 {
		recommendations = append(recommendations,
			"Consider enabling more aggressive optimizations")
	}

	// Memory recommendations
	if optimized.AllocationsCount > baseline.AllocationsCount*2 {
		recommendations = append(recommendations,
			"Review object pooling configuration - high allocation overhead")
	}

	// Operations per second analysis
	if optimized.OperationsPerSec < baseline.OperationsPerSec*0.9 {
		recommendations = append(recommendations,
			"Optimization overhead detected - consider selective optimization")
	}

	return recommendations
}

func saveReport(report *OptimizationReport, filename string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
