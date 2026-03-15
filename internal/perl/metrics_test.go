// ABOUTME: Tests for installation metrics and telemetry functionality
// ABOUTME: Validates metrics collection, storage, and privacy compliance

package perl

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInstallationMetrics_NewInstallationMetrics(t *testing.T) {
	metrics := NewInstallationMetrics()

	if metrics == nil {
		t.Fatal("NewInstallationMetrics() returned nil")
	}

	if metrics.InstallStartTime.IsZero() {
		t.Error("Expected InstallStartTime to be set")
	}

	if metrics.Method == "" {
		t.Error("Expected Method to be initialized")
	}
}

func TestInstallationMetrics_RecordDownloadPhase(t *testing.T) {
	metrics := NewInstallationMetrics()

	// Simulate download phase
	time.Sleep(10 * time.Millisecond)
	metrics.RecordDownloadPhase(1024*1024, true) // 1MB successful download

	if metrics.DownloadTime.Duration == 0 {
		t.Error("Expected DownloadTime to be recorded")
	}

	if metrics.DownloadSize != 1024*1024 {
		t.Errorf("Expected DownloadSize 1048576, got %d", metrics.DownloadSize)
	}

	if !metrics.DownloadSuccess {
		t.Error("Expected DownloadSuccess to be true")
	}
}

func TestInstallationMetrics_Complete(t *testing.T) {
	metrics := NewInstallationMetrics()
	metrics.Method = MethodBinary

	// Simulate installation process
	time.Sleep(5 * time.Millisecond)
	metrics.Complete(true, nil)

	if metrics.TotalTime.Duration == 0 {
		t.Error("Expected TotalTime to be recorded")
	}

	if !metrics.Success {
		t.Error("Expected Success to be true")
	}

	if metrics.ErrorMessage != "" {
		t.Errorf("Expected no error message, got: %s", metrics.ErrorMessage)
	}
}

func TestInstallationMetrics_CompleteWithError(t *testing.T) {
	metrics := NewInstallationMetrics()
	testError := "installation failed"

	metrics.Complete(false, &testError)

	if metrics.Success {
		t.Error("Expected Success to be false")
	}

	if metrics.ErrorMessage != testError {
		t.Errorf("Expected ErrorMessage '%s', got '%s'", testError, metrics.ErrorMessage)
	}
}

func TestMetricsCollector_NewMetricsCollector(t *testing.T) {
	tempDir := t.TempDir()
	collector := NewMetricsCollector(tempDir, true) // metrics enabled

	if collector == nil {
		t.Fatal("NewMetricsCollector() returned nil")
	}

	if !collector.enabled {
		t.Error("Expected collector to be enabled")
	}

	if collector.storageDir != tempDir {
		t.Errorf("Expected storageDir %s, got %s", tempDir, collector.storageDir)
	}
}

func TestMetricsCollector_RecordInstallation(t *testing.T) {
	tempDir := t.TempDir()
	collector := NewMetricsCollector(tempDir, true)

	metrics := NewInstallationMetrics()
	metrics.Method = MethodBinary
	metrics.Version = "5.38.0"
	metrics.Platform = "linux-amd64"
	metrics.Complete(true, nil)

	err := collector.RecordInstallation(metrics)
	if err != nil {
		t.Fatalf("RecordInstallation failed: %v", err)
	}

	// Verify metrics file was created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	if len(files) == 0 {
		t.Error("Expected metrics file to be created")
	}
}

func TestMetricsCollector_Disabled(t *testing.T) {
	tempDir := t.TempDir()
	collector := NewMetricsCollector(tempDir, false) // metrics disabled

	metrics := NewInstallationMetrics()
	metrics.Complete(true, nil)

	err := collector.RecordInstallation(metrics)
	if err != nil {
		t.Fatalf("RecordInstallation failed: %v", err)
	}

	// Verify no metrics file was created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	if len(files) > 0 {
		t.Error("Expected no metrics files when collector is disabled")
	}
}

func TestMetricsCollector_GetSummary(t *testing.T) {
	tempDir := t.TempDir()
	collector := NewMetricsCollector(tempDir, true)

	// Record some test metrics
	binaryMetrics := NewInstallationMetrics()
	binaryMetrics.Method = MethodBinary
	// Simulate time passing before completion
	time.Sleep(1 * time.Millisecond)
	binaryMetrics.Complete(true, nil)
	// Override the duration for testing
	binaryMetrics.TotalTime.Duration = 30 * time.Second

	sourceMetrics := NewInstallationMetrics()
	sourceMetrics.Method = MethodSource
	// Simulate time passing before completion
	time.Sleep(1 * time.Millisecond)
	sourceMetrics.Complete(true, nil)
	// Override the duration for testing
	sourceMetrics.TotalTime.Duration = 10 * time.Minute

	collector.RecordInstallation(binaryMetrics)
	collector.RecordInstallation(sourceMetrics)

	summary, err := collector.GetSummary()
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}

	if summary.TotalInstallations != 2 {
		t.Errorf("Expected 2 total installations, got %d", summary.TotalInstallations)
	}

	if summary.BinaryInstallations != 1 {
		t.Errorf("Expected 1 binary installation, got %d", summary.BinaryInstallations)
	}

	if summary.SourceInstallations != 1 {
		t.Errorf("Expected 1 source installation, got %d", summary.SourceInstallations)
	}

	if summary.AverageBinaryTime == 0 {
		t.Error("Expected AverageBinaryTime to be calculated")
	}

	if summary.AverageSourceTime == 0 {
		t.Error("Expected AverageSourceTime to be calculated")
	}
}

func TestMetricsCollector_CleanupOldMetrics(t *testing.T) {
	tempDir := t.TempDir()
	collector := NewMetricsCollector(tempDir, true)

	// Create an old metrics file
	oldFile := filepath.Join(tempDir, "metrics-old.json")
	oldFileContent := `{"timestamp":"2020-01-01T00:00:00Z","method":"binary"}`
	err := os.WriteFile(oldFile, []byte(oldFileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create old metrics file: %v", err)
	}

	// Set file modification time to be old
	oldTime := time.Now().AddDate(0, 0, -31) // 31 days ago
	err = os.Chtimes(oldFile, oldTime, oldTime)
	if err != nil {
		t.Fatalf("Failed to set old file time: %v", err)
	}

	err = collector.CleanupOldMetrics()
	if err != nil {
		t.Fatalf("CleanupOldMetrics failed: %v", err)
	}

	// Verify old file was removed
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Expected old metrics file to be removed")
	}
}

func TestInstallationMetrics_PrivacyCompliance(t *testing.T) {
	metrics := NewInstallationMetrics()
	metrics.Version = "5.38.0"
	metrics.Platform = "linux-amd64"
	metrics.Method = MethodBinary

	// Verify no personally identifiable information is stored
	data := metrics.ToMap()

	sensitiveFields := []string{"hostname", "username", "ip", "mac", "path", "home"}
	for _, field := range sensitiveFields {
		if _, exists := data[field]; exists {
			t.Errorf("Metrics contain potentially sensitive field: %s", field)
		}
	}

	// Verify only approved fields are present
	approvedFields := map[string]bool{
		"timestamp":        true,
		"method":           true,
		"version":          true,
		"platform":         true,
		"success":          true,
		"total_time":       true,
		"download_time":    true,
		"download_size":    true,
		"download_success": true,
		"error_type":       true, // Generic error type, not specific messages
	}

	for field := range data {
		if !approvedFields[field] {
			t.Errorf("Metrics contain non-approved field: %s", field)
		}
	}
}

func TestPerformanceComparison(t *testing.T) {
	tempDir := t.TempDir()
	collector := NewMetricsCollector(tempDir, true)

	// Record binary installation (fast)
	binaryMetrics := NewInstallationMetrics()
	binaryMetrics.Method = MethodBinary
	time.Sleep(1 * time.Millisecond)
	binaryMetrics.Complete(true, nil)
	binaryMetrics.TotalTime.Duration = 30 * time.Second

	// Record source installation (slow)
	sourceMetrics := NewInstallationMetrics()
	sourceMetrics.Method = MethodSource
	time.Sleep(1 * time.Millisecond)
	sourceMetrics.Complete(true, nil)
	sourceMetrics.TotalTime.Duration = 10 * time.Minute

	collector.RecordInstallation(binaryMetrics)
	collector.RecordInstallation(sourceMetrics)

	comparison, err := collector.GetPerformanceComparison()
	if err != nil {
		t.Fatalf("GetPerformanceComparison failed: %v", err)
	}

	if comparison.BinarySpeedupFactor <= 1.0 {
		t.Errorf("Expected binary to be faster than source, got speedup factor: %f", comparison.BinarySpeedupFactor)
	}

	if comparison.BinarySuccessRate != 1.0 {
		t.Errorf("Expected 100%% binary success rate, got %f", comparison.BinarySuccessRate)
	}

	if comparison.SourceSuccessRate != 1.0 {
		t.Errorf("Expected 100%% source success rate, got %f", comparison.SourceSuccessRate)
	}
}
