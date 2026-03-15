// ABOUTME: Installation metrics and telemetry for binary vs source performance tracking
// ABOUTME: Provides privacy-focused analytics with opt-out mechanism for installation monitoring

package perl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
)

// Installation method constants
const (
	MethodBinary  = "binary"
	MethodSource  = "source"
	MethodUnknown = "unknown"
)

// Metrics error codes
const (
	ErrMetricsStorageFailed   = "901" // Failed to store metrics data
	ErrMetricsRetrievalFailed = "902" // Failed to retrieve metrics data
	ErrMetricsCorrupted       = "903" // Metrics data is corrupted
	ErrMetricsCleanupFailed   = "904" // Failed to cleanup old metrics
)

// Metrics configuration constants
const (
	DefaultMetricsRetentionDays = 30
	MetricsFilePrefix           = "metrics-"
	MetricsFileSuffix           = ".json"
	MaxMetricsFileSize          = 1024 * 1024 // 1MB per metrics file
)

// TimedDuration represents a duration with start and end times
type TimedDuration struct {
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time,omitempty"`
	Duration  time.Duration `json:"duration"`
}

// InstallationMetrics contains comprehensive metrics for a single installation
type InstallationMetrics struct {
	// Timestamp when installation started
	Timestamp time.Time `json:"timestamp"`

	// Installation method (binary or source)
	Method string `json:"method"`

	// Perl version being installed
	Version string `json:"version"`

	// Platform triple (e.g., "linux-amd64")
	Platform string `json:"platform"`

	// Whether installation succeeded
	Success bool `json:"success"`

	// Generic error type (no sensitive details)
	ErrorType string `json:"error_type,omitempty"`

	// Non-sensitive error message for analytics
	ErrorMessage string `json:"-"` // Not serialized to maintain privacy

	// Timing information
	InstallStartTime time.Time     `json:"install_start_time"`
	TotalTime        TimedDuration `json:"total_time"`
	DownloadTime     TimedDuration `json:"download_time"`

	// Download metrics
	DownloadSize    int64 `json:"download_size"`
	DownloadSuccess bool  `json:"download_success"`

	// Performance metadata (no sensitive information)
	CacheHit         bool `json:"cache_hit"`
	ParallelDownload bool `json:"parallel_download"`
	ResumedDownload  bool `json:"resumed_download"`
}

// MetricsSummary provides aggregated metrics across installations
type MetricsSummary struct {
	// Total counts
	TotalInstallations  int `json:"total_installations"`
	BinaryInstallations int `json:"binary_installations"`
	SourceInstallations int `json:"source_installations"`

	// Success rates
	OverallSuccessRate float64 `json:"overall_success_rate"`
	BinarySuccessRate  float64 `json:"binary_success_rate"`
	SourceSuccessRate  float64 `json:"source_success_rate"`

	// Average timing
	AverageBinaryTime time.Duration `json:"average_binary_time"`
	AverageSourceTime time.Duration `json:"average_source_time"`

	// Data period
	FromDate time.Time `json:"from_date"`
	ToDate   time.Time `json:"to_date"`
}

// PerformanceComparison provides binary vs source performance analysis
type PerformanceComparison struct {
	// Speed comparison
	BinarySpeedupFactor float64 `json:"binary_speedup_factor"`

	// Success rate comparison
	BinarySuccessRate float64 `json:"binary_success_rate"`
	SourceSuccessRate float64 `json:"source_success_rate"`

	// Download metrics
	AverageBinaryDownloadSize int64         `json:"average_binary_download_size"`
	AverageBinaryDownloadTime time.Duration `json:"average_binary_download_time"`

	// Cache effectiveness
	BinaryCacheHitRate float64 `json:"binary_cache_hit_rate"`

	// Sample sizes
	BinarySampleSize int `json:"binary_sample_size"`
	SourceSampleSize int `json:"source_sample_size"`
}

// MetricsCollector handles collection and storage of installation metrics
type MetricsCollector struct {
	enabled       bool
	storageDir    string
	retentionDays int
}

// NewInstallationMetrics creates a new installation metrics instance
func NewInstallationMetrics() *InstallationMetrics {
	now := time.Now()
	return &InstallationMetrics{
		Timestamp:        now,
		InstallStartTime: now,
		Method:           MethodUnknown,
		TotalTime: TimedDuration{
			StartTime: now,
		},
		DownloadTime: TimedDuration{},
	}
}

// RecordDownloadPhase records download timing and metrics
func (m *InstallationMetrics) RecordDownloadPhase(size int64, success bool) {
	now := time.Now()
	m.DownloadTime.EndTime = now
	m.DownloadTime.Duration = now.Sub(m.DownloadTime.StartTime)
	m.DownloadSize = size
	m.DownloadSuccess = success
}

// StartDownloadPhase marks the beginning of the download phase
func (m *InstallationMetrics) StartDownloadPhase() {
	m.DownloadTime.StartTime = time.Now()
}

// Complete finalizes the installation metrics
func (m *InstallationMetrics) Complete(success bool, errorMsg *string) {
	now := time.Now()
	m.TotalTime.EndTime = now
	m.TotalTime.Duration = now.Sub(m.TotalTime.StartTime)
	m.Success = success

	if errorMsg != nil {
		m.ErrorMessage = *errorMsg
		m.ErrorType = categorizeError(*errorMsg)
	}
}

// ToMap converts metrics to a map for serialization (privacy-safe)
func (m *InstallationMetrics) ToMap() map[string]interface{} {
	data := map[string]interface{}{
		"timestamp":        m.Timestamp.Unix(),
		"method":           m.Method,
		"version":          m.Version,
		"platform":         m.Platform,
		"success":          m.Success,
		"total_time":       m.TotalTime.Duration.Milliseconds(),
		"download_time":    m.DownloadTime.Duration.Milliseconds(),
		"download_size":    m.DownloadSize,
		"download_success": m.DownloadSuccess,
	}

	if m.ErrorType != "" {
		data["error_type"] = m.ErrorType
	}

	return data
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(storageDir string, enabled bool) *MetricsCollector {
	return &MetricsCollector{
		enabled:       enabled,
		storageDir:    storageDir,
		retentionDays: DefaultMetricsRetentionDays,
	}
}

// RecordInstallation records an installation's metrics
func (mc *MetricsCollector) RecordInstallation(metrics *InstallationMetrics) error {
	if !mc.enabled {
		log.Debugf("Metrics collection is disabled, skipping recording")
		return nil
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(mc.storageDir, 0755); err != nil {
		return errors.NewSystemError(ErrMetricsStorageFailed,
			"Failed to create metrics storage directory", err)
	}

	// Generate filename with timestamp and nanoseconds to ensure uniqueness
	filename := fmt.Sprintf("%s%d-%d%s",
		MetricsFilePrefix,
		metrics.Timestamp.Unix(),
		metrics.Timestamp.Nanosecond(),
		MetricsFileSuffix)
	filePath := filepath.Join(mc.storageDir, filename)

	// Serialize metrics (privacy-safe)
	data, err := json.MarshalIndent(metrics.ToMap(), "", "  ")
	if err != nil {
		return errors.NewSystemError(ErrMetricsStorageFailed,
			"Failed to serialize metrics", err)
	}

	// Write metrics file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return errors.NewSystemError(ErrMetricsStorageFailed,
			"Failed to write metrics file", err).WithLocation(filePath)
	}

	log.Debugf("Recorded installation metrics: %s", filePath)
	return nil
}

// GetSummary generates an aggregated summary of all metrics
func (mc *MetricsCollector) GetSummary() (*MetricsSummary, error) {
	if !mc.enabled {
		return &MetricsSummary{}, nil
	}

	metrics, err := mc.loadAllMetrics()
	if err != nil {
		return nil, err
	}

	if len(metrics) == 0 {
		return &MetricsSummary{}, nil
	}

	summary := &MetricsSummary{}
	binaryTimes := []time.Duration{}
	sourceTimes := []time.Duration{}
	binarySuccesses := 0
	sourceSuccesses := 0

	// Find date range
	summary.FromDate = metrics[0].Timestamp
	summary.ToDate = metrics[0].Timestamp

	for _, m := range metrics {
		if m.Timestamp.Before(summary.FromDate) {
			summary.FromDate = m.Timestamp
		}
		if m.Timestamp.After(summary.ToDate) {
			summary.ToDate = m.Timestamp
		}

		summary.TotalInstallations++

		switch m.Method {
		case MethodBinary:
			summary.BinaryInstallations++
			binaryTimes = append(binaryTimes, m.TotalTime.Duration)
			if m.Success {
				binarySuccesses++
			}
		case MethodSource:
			summary.SourceInstallations++
			sourceTimes = append(sourceTimes, m.TotalTime.Duration)
			if m.Success {
				sourceSuccesses++
			}
		}
	}

	// Calculate success rates
	if summary.BinaryInstallations > 0 {
		summary.BinarySuccessRate = float64(binarySuccesses) / float64(summary.BinaryInstallations)
	}
	if summary.SourceInstallations > 0 {
		summary.SourceSuccessRate = float64(sourceSuccesses) / float64(summary.SourceInstallations)
	}
	summary.OverallSuccessRate = float64(binarySuccesses+sourceSuccesses) / float64(summary.TotalInstallations)

	// Calculate average times
	summary.AverageBinaryTime = calculateAverageTime(binaryTimes)
	summary.AverageSourceTime = calculateAverageTime(sourceTimes)

	return summary, nil
}

// GetPerformanceComparison generates binary vs source performance comparison
func (mc *MetricsCollector) GetPerformanceComparison() (*PerformanceComparison, error) {
	if !mc.enabled {
		return &PerformanceComparison{}, nil
	}

	metrics, err := mc.loadAllMetrics()
	if err != nil {
		return nil, err
	}

	comparison := &PerformanceComparison{}
	binaryMetrics := []*InstallationMetrics{}
	sourceMetrics := []*InstallationMetrics{}

	// Separate binary and source metrics
	for _, m := range metrics {
		switch m.Method {
		case MethodBinary:
			binaryMetrics = append(binaryMetrics, m)
		case MethodSource:
			sourceMetrics = append(sourceMetrics, m)
		}
	}

	comparison.BinarySampleSize = len(binaryMetrics)
	comparison.SourceSampleSize = len(sourceMetrics)

	if len(binaryMetrics) == 0 || len(sourceMetrics) == 0 {
		return comparison, nil
	}

	// Calculate success rates
	binarySuccesses := 0
	sourceSuccesses := 0
	binaryTimes := []time.Duration{}
	sourceTimes := []time.Duration{}
	totalDownloadSize := int64(0)
	totalDownloadTime := time.Duration(0)
	cacheHits := 0

	for _, m := range binaryMetrics {
		if m.Success {
			binarySuccesses++
			binaryTimes = append(binaryTimes, m.TotalTime.Duration)
		}
		totalDownloadSize += m.DownloadSize
		totalDownloadTime += m.DownloadTime.Duration
		if m.CacheHit {
			cacheHits++
		}
	}

	for _, m := range sourceMetrics {
		if m.Success {
			sourceSuccesses++
			sourceTimes = append(sourceTimes, m.TotalTime.Duration)
		}
	}

	comparison.BinarySuccessRate = float64(binarySuccesses) / float64(len(binaryMetrics))
	comparison.SourceSuccessRate = float64(sourceSuccesses) / float64(len(sourceMetrics))

	// Calculate speedup factor
	avgBinaryTime := calculateAverageTime(binaryTimes)
	avgSourceTime := calculateAverageTime(sourceTimes)

	if avgBinaryTime > 0 && avgSourceTime > 0 {
		comparison.BinarySpeedupFactor = float64(avgSourceTime) / float64(avgBinaryTime)
	}

	// Calculate download metrics
	if len(binaryMetrics) > 0 {
		comparison.AverageBinaryDownloadSize = totalDownloadSize / int64(len(binaryMetrics))
		comparison.AverageBinaryDownloadTime = totalDownloadTime / time.Duration(len(binaryMetrics))
		comparison.BinaryCacheHitRate = float64(cacheHits) / float64(len(binaryMetrics))
	}

	return comparison, nil
}

// CleanupOldMetrics removes metrics files older than retention period
func (mc *MetricsCollector) CleanupOldMetrics() error {
	if !mc.enabled {
		return nil
	}

	cutoffTime := time.Now().AddDate(0, 0, -mc.retentionDays)

	files, err := os.ReadDir(mc.storageDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No storage directory, nothing to clean
		}
		return errors.NewSystemError(ErrMetricsCleanupFailed,
			"Failed to read metrics directory", err)
	}

	removedCount := 0
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), MetricsFilePrefix) ||
			!strings.HasSuffix(file.Name(), MetricsFileSuffix) {
			continue
		}

		filePath := filepath.Join(mc.storageDir, file.Name())
		info, err := file.Info()
		if err != nil {
			log.Warnf("Failed to get file info for %s: %v", filePath, err)
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(filePath); err != nil {
				log.Warnf("Failed to remove old metrics file %s: %v", filePath, err)
			} else {
				removedCount++
				log.Debugf("Removed old metrics file: %s", filePath)
			}
		}
	}

	if removedCount > 0 {
		log.Infof("Cleaned up %d old metrics files", removedCount)
	}

	return nil
}

// loadAllMetrics loads all metrics from storage directory
func (mc *MetricsCollector) loadAllMetrics() ([]*InstallationMetrics, error) {
	files, err := os.ReadDir(mc.storageDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*InstallationMetrics{}, nil
		}
		return nil, errors.NewSystemError(ErrMetricsRetrievalFailed,
			"Failed to read metrics directory", err)
	}

	metrics := []*InstallationMetrics{}

	for _, file := range files {
		if !strings.HasPrefix(file.Name(), MetricsFilePrefix) ||
			!strings.HasSuffix(file.Name(), MetricsFileSuffix) {
			continue
		}

		filePath := filepath.Join(mc.storageDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Warnf("Failed to read metrics file %s: %v", filePath, err)
			continue
		}

		var rawData map[string]interface{}
		if err := json.Unmarshal(data, &rawData); err != nil {
			log.Warnf("Failed to parse metrics file %s: %v", filePath, err)
			continue
		}

		metric, err := parseMetricsFromMap(rawData)
		if err != nil {
			log.Warnf("Failed to convert metrics from %s: %v", filePath, err)
			continue
		}

		metrics = append(metrics, metric)
	}

	// Sort by timestamp
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Timestamp.Before(metrics[j].Timestamp)
	})

	return metrics, nil
}

// parseMetricsFromMap converts a map to InstallationMetrics
func parseMetricsFromMap(data map[string]interface{}) (*InstallationMetrics, error) {
	m := &InstallationMetrics{}

	if timestamp, ok := data["timestamp"].(float64); ok {
		m.Timestamp = time.Unix(int64(timestamp), 0)
	}

	if method, ok := data["method"].(string); ok {
		m.Method = method
	}

	if version, ok := data["version"].(string); ok {
		m.Version = version
	}

	if platform, ok := data["platform"].(string); ok {
		m.Platform = platform
	}

	if success, ok := data["success"].(bool); ok {
		m.Success = success
	}

	if totalTime, ok := data["total_time"].(float64); ok {
		m.TotalTime.Duration = time.Duration(totalTime) * time.Millisecond
	}

	if downloadTime, ok := data["download_time"].(float64); ok {
		m.DownloadTime.Duration = time.Duration(downloadTime) * time.Millisecond
	}

	if downloadSize, ok := data["download_size"].(float64); ok {
		m.DownloadSize = int64(downloadSize)
	}

	if downloadSuccess, ok := data["download_success"].(bool); ok {
		m.DownloadSuccess = downloadSuccess
	}

	if errorType, ok := data["error_type"].(string); ok {
		m.ErrorType = errorType
	}

	return m, nil
}

// calculateAverageTime calculates the average of a slice of durations
func calculateAverageTime(times []time.Duration) time.Duration {
	if len(times) == 0 {
		return 0
	}

	total := time.Duration(0)
	for _, t := range times {
		total += t
	}

	return total / time.Duration(len(times))
}

// categorizeError converts specific error messages to generic categories for privacy
func categorizeError(errorMsg string) string {
	errorLower := strings.ToLower(errorMsg)

	switch {
	case strings.Contains(errorLower, "network") || strings.Contains(errorLower, "connection"):
		return "network_error"
	case strings.Contains(errorLower, "checksum") || strings.Contains(errorLower, "verify"):
		return "verification_error"
	case strings.Contains(errorLower, "permission") || strings.Contains(errorLower, "access"):
		return "permission_error"
	case strings.Contains(errorLower, "disk") || strings.Contains(errorLower, "space"):
		return "storage_error"
	case strings.Contains(errorLower, "timeout"):
		return "timeout_error"
	case strings.Contains(errorLower, "not found") || strings.Contains(errorLower, "404"):
		return "not_found_error"
	default:
		return "general_error"
	}
}
