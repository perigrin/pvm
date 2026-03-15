// ABOUTME: Performance regression testing system for pool performance validation
// ABOUTME: Provides automated regression detection and performance baseline management

package monitoring

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
)

// RegressionTestSuite manages performance regression testing
type RegressionTestSuite struct {
	mu             sync.RWMutex
	tests          map[string]*RegressionTest
	history        []RegressionRunResult
	alertCallbacks []func(RegressionAlert)
	config         RegressionConfig
	benchmarkSuite *SimpleBenchmarkSuite
	maxHistory     int
}

// RegressionConfig contains regression testing configuration
type RegressionConfig struct {
	DefaultTolerance       float64       // Default regression tolerance percentage
	MaxConsecutiveFailures int           // Max failures before alerting
	TestInterval           time.Duration // How often to run regression tests
	BaselineUpdateStrategy string        // "manual", "auto", "rolling"
	AlertThreshold         float64       // Threshold for alerting
	EnableAutoBaseline     bool          // Automatically update baselines
	HistoryRetention       time.Duration // How long to keep test history
}

// RegressionRunResult contains results from a regression test run
type RegressionRunResult struct {
	Timestamp        time.Time
	TestResults      map[string]RegressionTestResult
	OverallStatus    string
	TotalTests       int
	PassedTests      int
	FailedTests      int
	RegressionCount  int
	PerformanceGrade string
}

// RegressionTestResult contains results for a single regression test
type RegressionTestResult struct {
	TestName           string
	Status             string
	CurrentMetrics     BenchmarkMetrics
	BaselineMetrics    BenchmarkMetrics
	PerformanceChange  float64
	MemoryChange       int64
	RegressionDetected bool
	Details            string
}

// RegressionAlert represents a performance regression alert
type RegressionAlert struct {
	Timestamp         time.Time
	TestName          string
	Severity          string
	RegressionPercent float64
	Description       string
	Recommendations   []string
	AffectedPools     []string
}

// NewRegressionTestSuite creates a new regression test suite
func NewRegressionTestSuite(benchmarkSuite *SimpleBenchmarkSuite, config RegressionConfig) *RegressionTestSuite {
	return &RegressionTestSuite{
		tests:          make(map[string]*RegressionTest),
		history:        make([]RegressionRunResult, 0),
		alertCallbacks: make([]func(RegressionAlert), 0),
		config:         config,
		benchmarkSuite: benchmarkSuite,
		maxHistory:     1000,
	}
}

// DefaultRegressionConfig returns sensible defaults for regression testing
func DefaultRegressionConfig() RegressionConfig {
	return RegressionConfig{
		DefaultTolerance:       5.0,           // 5% regression tolerance
		MaxConsecutiveFailures: 3,             // Alert after 3 consecutive failures
		TestInterval:           time.Hour * 6, // Run every 6 hours
		BaselineUpdateStrategy: "rolling",
		AlertThreshold:         10.0, // Alert on 10% regression
		EnableAutoBaseline:     true,
		HistoryRetention:       time.Hour * 24 * 30, // Keep 30 days of history
	}
}

// RegisterTest adds a regression test to the suite
func (rts *RegressionTestSuite) RegisterTest(test *RegressionTest) {
	rts.mu.Lock()
	defer rts.mu.Unlock()

	if test.Tolerance == 0 {
		test.Tolerance = rts.config.DefaultTolerance
	}
	if test.FailureThreshold == 0 {
		test.FailureThreshold = rts.config.MaxConsecutiveFailures
	}

	rts.tests[test.Name] = test
}

// RegisterStandardTests registers the standard set of regression tests
func (rts *RegressionTestSuite) RegisterStandardTests() {
	simpleSuite := NewSimpleBenchmarkSuite()

	// Core pool performance test
	rts.RegisterTest(&RegressionTest{
		Name: "Core_Pool_Performance",
		TestFunction: func() BenchmarkResult {
			config := BenchmarkConfig{
				Operations:        5000,
				Concurrency:       4,
				ObjectSize:        64,
				AllocationPattern: "sequential",
				TestDuration:      time.Second * 5,
			}
			return simpleSuite.benchmarkCorePooled(config)
		},
		Tolerance:        5.0,
		Status:           "active",
		FailureThreshold: 3,
	})

	// Concurrent pool performance test
	rts.RegisterTest(&RegressionTest{
		Name: "Concurrent_Pool_Performance",
		TestFunction: func() BenchmarkResult {
			config := BenchmarkConfig{
				Operations:        10000,
				Concurrency:       8,
				ObjectSize:        32,
				AllocationPattern: "high_concurrency",
				TestDuration:      time.Second * 10,
			}
			return simpleSuite.benchmarkCorePooled(config)
		},
		Tolerance:        7.0,
		Status:           "active",
		FailureThreshold: 3,
	})

	// Memory efficiency test - small objects
	rts.RegisterTest(&RegressionTest{
		Name: "Memory_Efficiency_Small",
		TestFunction: func() BenchmarkResult {
			config := BenchmarkConfig{
				Operations:        2000,
				Concurrency:       4,
				ObjectSize:        64,
				AllocationPattern: "memory_efficiency",
				TestDuration:      time.Second * 10,
			}
			return simpleSuite.benchmarkMemoryPooled(config)
		},
		Tolerance:        4.0,
		Status:           "active",
		FailureThreshold: 2,
	})

	// Memory efficiency test - large objects
	rts.RegisterTest(&RegressionTest{
		Name: "Memory_Efficiency_Large",
		TestFunction: func() BenchmarkResult {
			config := BenchmarkConfig{
				Operations:        1000,
				Concurrency:       4,
				ObjectSize:        1024,
				AllocationPattern: "memory_efficiency",
				TestDuration:      time.Second * 15,
			}
			return simpleSuite.benchmarkMemoryPooled(config)
		},
		Tolerance:        8.0,
		Status:           "active",
		FailureThreshold: 3,
	})

	// Pool growth efficiency test
	rts.RegisterTest(&RegressionTest{
		Name: "Pool_Growth_Efficiency",
		TestFunction: func() BenchmarkResult {
			config := BenchmarkConfig{
				Operations:        50000,
				Concurrency:       1,
				ObjectSize:        32,
				AllocationPattern: "sequential",
				TestDuration:      time.Second * 20,
			}
			return simpleSuite.benchmarkCorePooled(config)
		},
		Tolerance:        6.0,
		Status:           "active",
		FailureThreshold: 3,
	})
}

// RunAllTests executes all registered regression tests
func (rts *RegressionTestSuite) RunAllTests(ctx context.Context) RegressionRunResult {
	rts.mu.Lock()
	defer rts.mu.Unlock()

	result := RegressionRunResult{
		Timestamp:   time.Now(),
		TestResults: make(map[string]RegressionTestResult),
		TotalTests:  len(rts.tests),
	}

	var regressionCount int
	var passedTests int
	var failedTests int

	// Run each test
	for name, test := range rts.tests {
		if test.Status != "active" {
			continue
		}

		testResult := rts.runSingleTest(ctx, test)
		result.TestResults[name] = testResult

		if testResult.Status == "passed" {
			passedTests++
		} else {
			failedTests++
		}

		if testResult.RegressionDetected {
			regressionCount++
		}

		// Update test state
		test.LastRun = time.Now()
		if testResult.Status == "failed" {
			test.ConsecutiveFailures++
		} else {
			test.ConsecutiveFailures = 0
		}

		// Check for alerts
		if test.ConsecutiveFailures >= test.FailureThreshold {
			rts.sendAlert(test, testResult)
		}

		// Update baseline if configured
		if rts.config.EnableAutoBaseline && rts.shouldUpdateBaseline(test, testResult) {
			test.Baseline = BenchmarkResult{
				PooledAllocations: testResult.CurrentMetrics,
				Timestamp:         time.Now(),
			}
		}
	}

	result.PassedTests = passedTests
	result.FailedTests = failedTests
	result.RegressionCount = regressionCount

	// Calculate overall status
	switch {
	case regressionCount > 0:
		result.OverallStatus = "regression_detected"
	case failedTests > 0:
		result.OverallStatus = "failures"
	default:
		result.OverallStatus = "passed"
	}

	// Calculate performance grade
	result.PerformanceGrade = rts.calculatePerformanceGrade(result)

	// Store result
	rts.history = append(rts.history, result)
	rts.trimHistory()

	return result
}

// runSingleTest executes a single regression test
func (rts *RegressionTestSuite) runSingleTest(ctx context.Context, test *RegressionTest) RegressionTestResult {
	// Run the test function
	currentResult := test.TestFunction()
	currentMetrics := currentResult.PooledAllocations

	result := RegressionTestResult{
		TestName:       test.Name,
		CurrentMetrics: currentMetrics,
		Status:         "passed",
	}

	// Compare against baseline if it exists
	if !test.Baseline.Timestamp.IsZero() {
		baselineMetrics := test.Baseline.PooledAllocations
		result.BaselineMetrics = baselineMetrics

		// Calculate performance change
		if baselineMetrics.TotalTime > 0 {
			timeDiff := currentMetrics.TotalTime - baselineMetrics.TotalTime
			result.PerformanceChange = float64(timeDiff) / float64(baselineMetrics.TotalTime) * 100
		}

		// Calculate memory change
		result.MemoryChange = currentMetrics.MemoryAllocated - baselineMetrics.MemoryAllocated

		// Check for regression
		if result.PerformanceChange > test.Tolerance {
			result.RegressionDetected = true
			result.Status = "failed"
			result.Details = fmt.Sprintf("Performance regression of %.2f%% detected (tolerance: %.2f%%)",
				result.PerformanceChange, test.Tolerance)
		}

		// Check for significant memory increase
		memoryIncreasePercent := float64(result.MemoryChange) / float64(baselineMetrics.MemoryAllocated) * 100
		if memoryIncreasePercent > test.Tolerance {
			result.RegressionDetected = true
			result.Status = "failed"
			if result.Details != "" {
				result.Details += "; "
			}
			result.Details += fmt.Sprintf("Memory usage regression of %.2f%% detected", memoryIncreasePercent)
		}
	} else {
		// No baseline - set current as baseline
		test.Baseline = currentResult
		result.Details = "Baseline established"
	}

	return result
}

// shouldUpdateBaseline determines if baseline should be updated
func (rts *RegressionTestSuite) shouldUpdateBaseline(test *RegressionTest, result RegressionTestResult) bool {
	switch rts.config.BaselineUpdateStrategy {
	case "manual":
		return false
	case "auto":
		// Update if performance improved significantly
		return result.PerformanceChange < -5.0 // 5% improvement
	case "rolling":
		// Update if within tolerance and enough time has passed
		return !result.RegressionDetected &&
			time.Since(test.Baseline.Timestamp) > time.Hour*24 // Update daily
	default:
		return false
	}
}

// sendAlert sends a regression alert
func (rts *RegressionTestSuite) sendAlert(test *RegressionTest, result RegressionTestResult) {
	severity := "medium"
	if result.PerformanceChange > rts.config.AlertThreshold {
		severity = "high"
	}
	if result.PerformanceChange > rts.config.AlertThreshold*2 {
		severity = "critical"
	}

	alert := RegressionAlert{
		Timestamp:         time.Now(),
		TestName:          test.Name,
		Severity:          severity,
		RegressionPercent: result.PerformanceChange,
		Description:       result.Details,
		Recommendations:   rts.generateRecommendations(test, result),
		AffectedPools:     rts.getAffectedPools(test.Name),
	}

	// Send to all registered callbacks
	for _, callback := range rts.alertCallbacks {
		go callback(alert)
	}
}

// generateRecommendations generates recommendations for addressing regressions
func (rts *RegressionTestSuite) generateRecommendations(test *RegressionTest, result RegressionTestResult) []string {
	var recommendations []string

	if result.PerformanceChange > 10 {
		recommendations = append(recommendations, "Investigate recent code changes for performance impact")
		recommendations = append(recommendations, "Run detailed profiling to identify bottlenecks")
	}

	if result.MemoryChange > 1024*1024 { // 1MB increase
		recommendations = append(recommendations, "Check for memory leaks or excessive allocations")
		recommendations = append(recommendations, "Review pool sizing and growth strategies")
	}

	if test.ConsecutiveFailures >= 2 {
		recommendations = append(recommendations, "Consider updating performance expectations")
		recommendations = append(recommendations, "Review test configuration and baseline validity")
	}

	return recommendations
}

// getAffectedPools identifies which pools might be affected by a test
func (rts *RegressionTestSuite) getAffectedPools(testName string) []string {
	// Map test names to affected pools
	poolMapping := map[string][]string{
		"AST_Allocation_Performance": {"ast_expressions", "ast_statements", "ast_types"},
		"Scanner_Token_Performance":  {"scanner_tokens", "scanner_iterators"},
		"TypeChecker_Performance":    {"type_objects", "inference_contexts"},
		"Core_Pool_Performance":      {"core_pools"},
		"Memory_Efficiency_Test":     {"all_pools"},
	}

	if pools, exists := poolMapping[testName]; exists {
		return pools
	}
	return []string{"unknown"}
}

// calculatePerformanceGrade calculates an overall performance grade
func (rts *RegressionTestSuite) calculatePerformanceGrade(result RegressionRunResult) string {
	if result.TotalTests == 0 {
		return "N/A"
	}

	passRate := float64(result.PassedTests) / float64(result.TotalTests)
	regressionRate := float64(result.RegressionCount) / float64(result.TotalTests)

	if passRate == 1.0 && regressionRate == 0 {
		return "A+"
	}
	if passRate >= 0.9 && regressionRate <= 0.1 {
		return "A"
	}
	if passRate >= 0.8 && regressionRate <= 0.2 {
		return "B"
	}
	if passRate >= 0.7 && regressionRate <= 0.3 {
		return "C"
	}
	if passRate >= 0.6 {
		return "D"
	}
	return "F"
}

// trimHistory removes old history entries
func (rts *RegressionTestSuite) trimHistory() {
	cutoff := time.Now().Add(-rts.config.HistoryRetention)

	// Find first entry to keep
	keepIndex := 0
	for i, entry := range rts.history {
		if entry.Timestamp.After(cutoff) {
			keepIndex = i
			break
		}
	}

	// Trim old entries
	if keepIndex > 0 {
		rts.history = rts.history[keepIndex:]
	}

	// Also enforce max history length
	if len(rts.history) > rts.maxHistory {
		excess := len(rts.history) - rts.maxHistory
		rts.history = rts.history[excess:]
	}
}

// AddAlertCallback adds a callback for regression alerts
func (rts *RegressionTestSuite) AddAlertCallback(callback func(RegressionAlert)) {
	rts.mu.Lock()
	defer rts.mu.Unlock()
	rts.alertCallbacks = append(rts.alertCallbacks, callback)
}

// GetTestStatus returns the current status of all tests
func (rts *RegressionTestSuite) GetTestStatus() map[string]RegressionTestStatus {
	rts.mu.RLock()
	defer rts.mu.RUnlock()

	status := make(map[string]RegressionTestStatus)

	for name, test := range rts.tests {
		status[name] = RegressionTestStatus{
			Name:                name,
			Status:              test.Status,
			LastRun:             test.LastRun,
			ConsecutiveFailures: test.ConsecutiveFailures,
			Tolerance:           test.Tolerance,
			HasBaseline:         !test.Baseline.Timestamp.IsZero(),
			BaselineAge:         time.Since(test.Baseline.Timestamp),
		}
	}

	return status
}

// RegressionTestStatus contains status information for a regression test
type RegressionTestStatus struct {
	Name                string        `json:"name"`
	Status              string        `json:"status"`
	LastRun             time.Time     `json:"last_run"`
	ConsecutiveFailures int           `json:"consecutive_failures"`
	Tolerance           float64       `json:"tolerance"`
	HasBaseline         bool          `json:"has_baseline"`
	BaselineAge         time.Duration `json:"baseline_age"`
}

// GetHistory returns regression test history
func (rts *RegressionTestSuite) GetHistory() []RegressionRunResult {
	rts.mu.RLock()
	defer rts.mu.RUnlock()

	history := make([]RegressionRunResult, len(rts.history))
	copy(history, rts.history)
	return history
}

// GetTrends analyzes performance trends over time
func (rts *RegressionTestSuite) GetTrends() map[string]PerformanceTrend {
	rts.mu.RLock()
	defer rts.mu.RUnlock()

	trends := make(map[string]PerformanceTrend)

	if len(rts.history) < 2 {
		return trends
	}

	// Analyze trends for each test
	for testName := range rts.tests {
		var dataPoints []TrendDataPoint

		for _, run := range rts.history {
			if result, exists := run.TestResults[testName]; exists {
				dataPoints = append(dataPoints, TrendDataPoint{
					Timestamp:         run.Timestamp,
					PerformanceChange: result.PerformanceChange,
					MemoryChange:      result.MemoryChange,
				})
			}
		}

		if len(dataPoints) >= 2 {
			trends[testName] = rts.calculateTrend(dataPoints)
		}
	}

	return trends
}

// PerformanceTrend represents performance trend analysis
type PerformanceTrend struct {
	Direction       string    `json:"direction"`
	AverageChange   float64   `json:"average_change"`
	Volatility      float64   `json:"volatility"`
	RecentTrend     string    `json:"recent_trend"`
	LastImprovement time.Time `json:"last_improvement"`
	LastRegression  time.Time `json:"last_regression"`
	StabilityScore  float64   `json:"stability_score"`
}

// TrendDataPoint represents a single data point in trend analysis
type TrendDataPoint struct {
	Timestamp         time.Time
	PerformanceChange float64
	MemoryChange      int64
}

// calculateTrend analyzes trend from data points
func (rts *RegressionTestSuite) calculateTrend(dataPoints []TrendDataPoint) PerformanceTrend {
	if len(dataPoints) < 2 {
		return PerformanceTrend{Direction: "insufficient_data"}
	}

	// Sort by timestamp
	sort.Slice(dataPoints, func(i, j int) bool {
		return dataPoints[i].Timestamp.Before(dataPoints[j].Timestamp)
	})

	// Calculate average change
	var totalChange float64
	var improvements, regressions int
	var lastImprovement, lastRegression time.Time

	for _, point := range dataPoints {
		totalChange += point.PerformanceChange

		if point.PerformanceChange < -1.0 { // 1% improvement threshold
			improvements++
			lastImprovement = point.Timestamp
		} else if point.PerformanceChange > 1.0 { // 1% regression threshold
			regressions++
			lastRegression = point.Timestamp
		}
	}

	avgChange := totalChange / float64(len(dataPoints))

	// Calculate volatility (standard deviation)
	var variance float64
	for _, point := range dataPoints {
		diff := point.PerformanceChange - avgChange
		variance += diff * diff
	}
	volatility := variance / float64(len(dataPoints))

	// Determine direction
	direction := "stable"
	if avgChange > 2.0 {
		direction = "degrading"
	} else if avgChange < -2.0 {
		direction = "improving"
	}

	// Analyze recent trend (last 5 data points)
	recentTrend := "stable"
	if len(dataPoints) >= 5 {
		recentPoints := dataPoints[len(dataPoints)-5:]
		var recentTotal float64
		for _, point := range recentPoints {
			recentTotal += point.PerformanceChange
		}
		recentAvg := recentTotal / 5.0

		if recentAvg > 3.0 {
			recentTrend = "degrading"
		} else if recentAvg < -3.0 {
			recentTrend = "improving"
		}
	}

	// Calculate stability score (lower volatility = higher stability)
	stabilityScore := 1.0 / (1.0 + volatility/10.0)

	return PerformanceTrend{
		Direction:       direction,
		AverageChange:   avgChange,
		Volatility:      volatility,
		RecentTrend:     recentTrend,
		LastImprovement: lastImprovement,
		LastRegression:  lastRegression,
		StabilityScore:  stabilityScore,
	}
}

// StartScheduledTesting starts running regression tests on a schedule
func (rts *RegressionTestSuite) StartScheduledTesting(ctx context.Context) {
	ticker := time.NewTicker(rts.config.TestInterval)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				result := rts.RunAllTests(ctx)
				if result.RegressionCount > 0 {
					log.Printf("Regression detected in %d tests during scheduled run", result.RegressionCount)
				}

			case <-ctx.Done():
				return
			}
		}
	}()
}
