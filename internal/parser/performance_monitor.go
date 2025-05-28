// ABOUTME: Performance monitoring and tracking infrastructure for Step 23
// ABOUTME: Provides continuous performance monitoring and automated regression detection

package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// PerformanceMetrics tracks detailed performance metrics over time
type PerformanceMetrics struct {
	TestName          string            `json:"test_name"`
	Timestamp         time.Time         `json:"timestamp"`
	ParseDuration     time.Duration     `json:"parse_duration"`
	MemoryUsage       int64             `json:"memory_usage"`
	AllocCount        int64             `json:"alloc_count"`
	GCCount           int64             `json:"gc_count"`
	ParserType        string            `json:"parser_type"`
	GitCommit         string            `json:"git_commit,omitempty"`
	BenchmarkScore    float64           `json:"benchmark_score"`
	ThroughputOpsPerSec float64         `json:"throughput_ops_per_sec"`
	CustomMetrics     map[string]float64 `json:"custom_metrics,omitempty"`
}

// PerformanceHistory stores historical performance data
type PerformanceHistory struct {
	TestName string               `json:"test_name"`
	Metrics  []PerformanceMetrics `json:"metrics"`
}

// PerformanceTrend represents performance trend analysis
type PerformanceTrend struct {
	TestName           string        `json:"test_name"`
	TrendDirection     string        `json:"trend_direction"` // "improving", "degrading", "stable"
	TrendSlope         float64       `json:"trend_slope"`
	ConfidenceLevel    float64       `json:"confidence_level"`
	RecentAvgDuration  time.Duration `json:"recent_avg_duration"`
	HistoricalAvgDuration time.Duration `json:"historical_avg_duration"`
	PercentChange      float64       `json:"percent_change"`
	LastUpdated        time.Time     `json:"last_updated"`
}

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	TestName      string    `json:"test_name"`
	AlertType     string    `json:"alert_type"` // "regression", "improvement", "anomaly"
	Severity      string    `json:"severity"`   // "low", "medium", "high", "critical"
	Message       string    `json:"message"`
	CurrentValue  float64   `json:"current_value"`
	BaselineValue float64   `json:"baseline_value"`
	Threshold     float64   `json:"threshold"`
	Timestamp     time.Time `json:"timestamp"`
}

// PerformanceMonitor manages performance monitoring and alerting
type PerformanceMonitor struct {
	DataDir         string
	AlertThresholds map[string]float64
	HistoryLimit    int // Maximum number of historical entries to keep
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(dataDir string) *PerformanceMonitor {
	return &PerformanceMonitor{
		DataDir: dataDir,
		AlertThresholds: map[string]float64{
			"regression_warning":  1.10, // 10% slower
			"regression_critical": 1.25, // 25% slower
			"memory_warning":      1.20, // 20% more memory
			"memory_critical":     1.50, // 50% more memory
		},
		HistoryLimit: 100, // Keep last 100 measurements
	}
}

// RecordMetrics records performance metrics for a test
func (pm *PerformanceMonitor) RecordMetrics(metrics PerformanceMetrics) error {
	history, err := pm.LoadHistory(metrics.TestName)
	if err != nil {
		// Create new history if doesn't exist
		history = &PerformanceHistory{
			TestName: metrics.TestName,
			Metrics:  []PerformanceMetrics{},
		}
	}

	// Add new metrics
	history.Metrics = append(history.Metrics, metrics)

	// Sort by timestamp
	sort.Slice(history.Metrics, func(i, j int) bool {
		return history.Metrics[i].Timestamp.Before(history.Metrics[j].Timestamp)
	})

	// Trim to history limit
	if len(history.Metrics) > pm.HistoryLimit {
		history.Metrics = history.Metrics[len(history.Metrics)-pm.HistoryLimit:]
	}

	return pm.SaveHistory(history)
}

// LoadHistory loads performance history for a test
func (pm *PerformanceMonitor) LoadHistory(testName string) (*PerformanceHistory, error) {
	historyFile := filepath.Join(pm.DataDir, "history", testName+".json")
	
	data, err := os.ReadFile(historyFile)
	if err != nil {
		return nil, err
	}

	var history PerformanceHistory
	err = json.Unmarshal(data, &history)
	return &history, err
}

// SaveHistory saves performance history for a test
func (pm *PerformanceMonitor) SaveHistory(history *PerformanceHistory) error {
	historyDir := filepath.Join(pm.DataDir, "history")
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	historyFile := filepath.Join(historyDir, history.TestName+".json")
	
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	return os.WriteFile(historyFile, data, 0644)
}

// AnalyzeTrend analyzes performance trends for a test
func (pm *PerformanceMonitor) AnalyzeTrend(testName string) (*PerformanceTrend, error) {
	history, err := pm.LoadHistory(testName)
	if err != nil {
		return nil, err
	}

	if len(history.Metrics) < 2 {
		return &PerformanceTrend{
			TestName:        testName,
			TrendDirection:  "insufficient_data",
			LastUpdated:     time.Now(),
		}, nil
	}

	// Calculate recent vs historical averages
	recentCount := minInt(10, len(history.Metrics))
	recentMetrics := history.Metrics[len(history.Metrics)-recentCount:]
	
	var recentAvg, historicalAvg time.Duration
	
	// Calculate recent average
	for _, m := range recentMetrics {
		recentAvg += m.ParseDuration
	}
	recentAvg /= time.Duration(len(recentMetrics))
	
	// Calculate historical average
	for _, m := range history.Metrics {
		historicalAvg += m.ParseDuration
	}
	historicalAvg /= time.Duration(len(history.Metrics))

	// Calculate percent change
	percentChange := (float64(recentAvg) - float64(historicalAvg)) / float64(historicalAvg) * 100

	// Determine trend direction
	var trendDirection string
	if percentChange > 5 {
		trendDirection = "degrading"
	} else if percentChange < -5 {
		trendDirection = "improving"
	} else {
		trendDirection = "stable"
	}

	// Calculate linear regression slope for more sophisticated trend analysis
	slope := pm.calculateTrendSlope(history.Metrics)

	return &PerformanceTrend{
		TestName:              testName,
		TrendDirection:        trendDirection,
		TrendSlope:            slope,
		ConfidenceLevel:       pm.calculateConfidence(history.Metrics),
		RecentAvgDuration:     recentAvg,
		HistoricalAvgDuration: historicalAvg,
		PercentChange:         percentChange,
		LastUpdated:           time.Now(),
	}, nil
}

// CheckForAlerts checks for performance alerts based on latest metrics
func (pm *PerformanceMonitor) CheckForAlerts(metrics PerformanceMetrics) ([]PerformanceAlert, error) {
	var alerts []PerformanceAlert

	history, err := pm.LoadHistory(metrics.TestName)
	if err != nil || len(history.Metrics) < 5 {
		// Need some history to establish baseline
		return alerts, nil
	}

	// Calculate baseline (average of last 10 historical measurements, excluding current)
	baselineCount := minInt(10, len(history.Metrics))
	baselineMetrics := history.Metrics[len(history.Metrics)-baselineCount:]
	
	var baselineDuration time.Duration
	var baselineMemory int64
	
	for _, m := range baselineMetrics {
		baselineDuration += m.ParseDuration
		baselineMemory += m.MemoryUsage
	}
	baselineDuration /= time.Duration(len(baselineMetrics))
	baselineMemory /= int64(len(baselineMetrics))

	// Check for duration regression
	durationRatio := float64(metrics.ParseDuration) / float64(baselineDuration)
	if durationRatio >= pm.AlertThresholds["regression_critical"] {
		alerts = append(alerts, PerformanceAlert{
			TestName:      metrics.TestName,
			AlertType:     "regression",
			Severity:      "critical",
			Message:       fmt.Sprintf("Critical performance regression: %.1f%% slower than baseline", (durationRatio-1)*100),
			CurrentValue:  float64(metrics.ParseDuration),
			BaselineValue: float64(baselineDuration),
			Threshold:     pm.AlertThresholds["regression_critical"],
			Timestamp:     time.Now(),
		})
	} else if durationRatio >= pm.AlertThresholds["regression_warning"] {
		alerts = append(alerts, PerformanceAlert{
			TestName:      metrics.TestName,
			AlertType:     "regression",
			Severity:      "warning",
			Message:       fmt.Sprintf("Performance regression detected: %.1f%% slower than baseline", (durationRatio-1)*100),
			CurrentValue:  float64(metrics.ParseDuration),
			BaselineValue: float64(baselineDuration),
			Threshold:     pm.AlertThresholds["regression_warning"],
			Timestamp:     time.Now(),
		})
	}

	// Check for memory regression
	memoryRatio := float64(metrics.MemoryUsage) / float64(baselineMemory)
	if memoryRatio >= pm.AlertThresholds["memory_critical"] {
		alerts = append(alerts, PerformanceAlert{
			TestName:      metrics.TestName,
			AlertType:     "regression",
			Severity:      "critical",
			Message:       fmt.Sprintf("Critical memory regression: %.1f%% more memory than baseline", (memoryRatio-1)*100),
			CurrentValue:  float64(metrics.MemoryUsage),
			BaselineValue: float64(baselineMemory),
			Threshold:     pm.AlertThresholds["memory_critical"],
			Timestamp:     time.Now(),
		})
	} else if memoryRatio >= pm.AlertThresholds["memory_warning"] {
		alerts = append(alerts, PerformanceAlert{
			TestName:      metrics.TestName,
			AlertType:     "regression",
			Severity:      "warning",
			Message:       fmt.Sprintf("Memory regression detected: %.1f%% more memory than baseline", (memoryRatio-1)*100),
			CurrentValue:  float64(metrics.MemoryUsage),
			BaselineValue: float64(baselineMemory),
			Threshold:     pm.AlertThresholds["memory_warning"],
			Timestamp:     time.Now(),
		})
	}

	// Check for significant improvements
	if durationRatio <= 0.8 {
		alerts = append(alerts, PerformanceAlert{
			TestName:      metrics.TestName,
			AlertType:     "improvement",
			Severity:      "low",
			Message:       fmt.Sprintf("Performance improvement: %.1f%% faster than baseline", (1-durationRatio)*100),
			CurrentValue:  float64(metrics.ParseDuration),
			BaselineValue: float64(baselineDuration),
			Threshold:     0.8,
			Timestamp:     time.Now(),
		})
	}

	return alerts, nil
}

// SaveAlerts saves performance alerts to a file
func (pm *PerformanceMonitor) SaveAlerts(alerts []PerformanceAlert) error {
	if len(alerts) == 0 {
		return nil
	}

	alertsDir := filepath.Join(pm.DataDir, "alerts")
	if err := os.MkdirAll(alertsDir, 0755); err != nil {
		return fmt.Errorf("failed to create alerts directory: %w", err)
	}

	alertsFile := filepath.Join(alertsDir, fmt.Sprintf("alerts_%s.json", 
		time.Now().Format("20060102_150405")))
	
	data, err := json.MarshalIndent(alerts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal alerts: %w", err)
	}

	return os.WriteFile(alertsFile, data, 0644)
}

// GeneratePerformanceReport generates a comprehensive performance report
func (pm *PerformanceMonitor) GeneratePerformanceReport() (*PerformanceReport, error) {
	historyDir := filepath.Join(pm.DataDir, "history")
	
	// Find all history files
	files, err := filepath.Glob(filepath.Join(historyDir, "*.json"))
	if err != nil {
		return nil, err
	}

	report := &PerformanceReport{
		Timestamp: time.Now(),
		Summary: PerformanceSummary{},
	}

	var allResults []PerformanceResult
	var allTrends []PerformanceTrend

	for _, file := range files {
		// Extract test name from filename
		testName := filepath.Base(file)
		testName = testName[:len(testName)-5] // Remove .json extension

		// Load history
		history, err := pm.LoadHistory(testName)
		if err != nil {
			continue
		}

		// Get latest result
		if len(history.Metrics) > 0 {
			latest := history.Metrics[len(history.Metrics)-1]
			result := PerformanceResult{
				TestName:      latest.TestName,
				ParseDuration: latest.ParseDuration,
				MemoryUsage:   latest.MemoryUsage,
				AllocCount:    latest.AllocCount,
				Success:       true,
				Timestamp:     latest.Timestamp,
				ParserType:    latest.ParserType,
			}
			allResults = append(allResults, result)
		}

		// Analyze trend
		trend, err := pm.AnalyzeTrend(testName)
		if err == nil {
			allTrends = append(allTrends, *trend)
		}
	}

	report.TestResults = allResults

	// Calculate summary
	pm.calculateSummaryFromResults(&report.Summary, allResults, allTrends)

	return report, nil
}

// Helper functions

func (pm *PerformanceMonitor) calculateTrendSlope(metrics []PerformanceMetrics) float64 {
	if len(metrics) < 2 {
		return 0
	}

	// Simple linear regression slope calculation
	n := float64(len(metrics))
	var sumX, sumY, sumXY, sumX2 float64

	for i, m := range metrics {
		x := float64(i)
		y := float64(m.ParseDuration)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Slope = (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0
	}

	return (n*sumXY - sumX*sumY) / denominator
}

func (pm *PerformanceMonitor) calculateConfidence(metrics []PerformanceMetrics) float64 {
	if len(metrics) < 3 {
		return 0
	}

	// Simple confidence based on data points and variance
	var sum, sumSquares float64
	for _, m := range metrics {
		duration := float64(m.ParseDuration)
		sum += duration
		sumSquares += duration * duration
	}

	mean := sum / float64(len(metrics))
	variance := (sumSquares - sum*sum/float64(len(metrics))) / float64(len(metrics)-1)
	stdDev := variance // Simplified

	// Higher confidence with more data points and lower variance
	confidence := float64(len(metrics)) / (1 + stdDev/mean)
	if confidence > 1 {
		confidence = 1
	}

	return confidence
}

func (pm *PerformanceMonitor) calculateSummaryFromResults(summary *PerformanceSummary, results []PerformanceResult, trends []PerformanceTrend) {
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

	// Count regressions from trends
	for _, trend := range trends {
		if trend.TrendDirection == "degrading" && trend.PercentChange > 10 {
			summary.RegressionsFound++
		}
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}