// ABOUTME: Integration tests for metrics collection with installation functions
// ABOUTME: Validates metrics are properly collected during binary installations

package perl

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestInstallFromBinary_MetricsIntegration(t *testing.T) {
	tempDir := t.TempDir()
	collector := NewMetricsCollector(tempDir, true)

	// Test metrics collection during installation (will fail since we don't have real binaries)
	options := &BinaryInstallOptions{
		Version:          "5.38.0",
		Platform:         "linux-amd64",
		InstallDir:       t.TempDir(),
		Context:          context.Background(),
		MetricsCollector: collector,
	}

	// This will fail but should still record metrics
	_, err := InstallFromBinary(options)
	if err == nil {
		t.Error("Expected installation to fail without real binary")
	}

	// Check that metrics were recorded
	summary, err := collector.GetSummary()
	if err != nil {
		t.Fatalf("Failed to get metrics summary: %v", err)
	}

	if summary.TotalInstallations != 1 {
		t.Errorf("Expected 1 installation recorded, got %d", summary.TotalInstallations)
	}

	if summary.BinaryInstallations != 1 {
		t.Errorf("Expected 1 binary installation recorded, got %d", summary.BinaryInstallations)
	}

	if summary.BinarySuccessRate != 0.0 {
		t.Errorf("Expected 0%% success rate for failed installation, got %f", summary.BinarySuccessRate)
	}
}

func TestInstallFromBinary_MetricsDisabled(t *testing.T) {
	tempDir := t.TempDir()
	collector := NewMetricsCollector(tempDir, false) // disabled

	options := &BinaryInstallOptions{
		Version:          "5.38.0",
		Platform:         "linux-amd64",
		InstallDir:       t.TempDir(),
		Context:          context.Background(),
		MetricsCollector: collector,
	}

	// This will fail but should not record metrics
	_, err := InstallFromBinary(options)
	if err == nil {
		t.Error("Expected installation to fail without real binary")
	}

	// Check that no metrics were recorded
	summary, err := collector.GetSummary()
	if err != nil {
		t.Fatalf("Failed to get metrics summary: %v", err)
	}

	if summary.TotalInstallations != 0 {
		t.Errorf("Expected 0 installations recorded when disabled, got %d", summary.TotalInstallations)
	}
}

func TestInstallFromBinary_NoMetricsCollector(t *testing.T) {
	// Test that installation works without metrics collector
	options := &BinaryInstallOptions{
		Version:          "5.38.0",
		Platform:         "linux-amd64",
		InstallDir:       t.TempDir(),
		Context:          context.Background(),
		MetricsCollector: nil, // no collector
	}

	// This should not panic and should handle nil collector gracefully
	_, err := InstallFromBinary(options)
	if err == nil {
		t.Error("Expected installation to fail without real binary")
	}
	// The test is that it doesn't panic
}

func TestMetricsCollector_IntegrationWithRealFiles(t *testing.T) {
	// Create a temporary metrics directory
	tempDir := t.TempDir()
	collector := NewMetricsCollector(tempDir, true)

	// Manually create and record some test metrics
	metrics1 := NewInstallationMetrics()
	metrics1.Method = MethodBinary
	metrics1.Version = "5.38.0"
	metrics1.Platform = "linux-amd64"
	metrics1.StartDownloadPhase()
	time.Sleep(1 * time.Millisecond)                 // Simulate download time
	metrics1.RecordDownloadPhase(50*1024*1024, true) // 50MB download
	time.Sleep(1 * time.Millisecond)                 // Simulate total time
	metrics1.Complete(true, nil)
	// Override with test duration
	metrics1.TotalTime.Duration = 30 * time.Second

	metrics2 := NewInstallationMetrics()
	metrics2.Method = MethodSource
	metrics2.Version = "5.38.0"
	metrics2.Platform = "linux-amd64"
	time.Sleep(1 * time.Millisecond) // Simulate total time
	metrics2.Complete(true, nil)
	// Override with test duration
	metrics2.TotalTime.Duration = 10 * time.Minute // Slow source build

	// Record the metrics
	err := collector.RecordInstallation(metrics1)
	if err != nil {
		t.Fatalf("Failed to record binary metrics: %v", err)
	}

	err = collector.RecordInstallation(metrics2)
	if err != nil {
		t.Fatalf("Failed to record source metrics: %v", err)
	}

	// Test that metrics files were created
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read metrics directory: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 metrics files, got %d", len(files))
	}

	// Test summary generation
	summary, err := collector.GetSummary()
	if err != nil {
		t.Fatalf("Failed to get summary: %v", err)
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

	// Test performance comparison
	comparison, err := collector.GetPerformanceComparison()
	if err != nil {
		t.Fatalf("Failed to get performance comparison: %v", err)
	}

	if comparison.BinarySpeedupFactor <= 1.0 {
		t.Errorf("Expected binary to be faster than source, got speedup factor: %f", comparison.BinarySpeedupFactor)
	}

	if comparison.AverageBinaryDownloadSize != 50*1024*1024 {
		t.Errorf("Expected 50MB average download size, got %d", comparison.AverageBinaryDownloadSize)
	}

	// Test cleanup
	err = collector.CleanupOldMetrics()
	if err != nil {
		t.Fatalf("Failed to cleanup metrics: %v", err)
	}
}
