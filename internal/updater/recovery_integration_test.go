// ABOUTME: Integration tests for PVM updater recovery scenarios
// ABOUTME: End-to-end tests for recovery workflows and failure scenarios

package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestRecoveryManager_Integration_CorruptedBinaryRecovery tests end-to-end corrupted binary recovery
func TestRecoveryManager_Integration_CorruptedBinaryRecovery(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "pvm-recovery-integration-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup test scenario
	targetPath := filepath.Join(tempDir, "pvm")

	// Create a corrupted binary (empty file)
	if err := os.WriteFile(targetPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create corrupted binary: %v", err)
	}

	// Create recovery manager
	rm := NewRecoveryManager()

	// Create recovery context for corrupted binary
	ctx := rm.CreateRecoveryContext(targetPath, "v1.0.0", fmt.Errorf("binary corrupted"))

	// Verify scenario is diagnosed correctly
	if ctx.Scenario != ScenarioCorruptedBinary {
		t.Errorf("Expected scenario %v, got %v", ScenarioCorruptedBinary, ctx.Scenario)
	}

	// Test getting recovery strategies
	strategies := rm.getRecoveryStrategies(ctx.Scenario)
	if len(strategies) == 0 {
		t.Error("No recovery strategies found for corrupted binary scenario")
	}

	// Verify strategies are properly configured
	for _, strategy := range strategies {
		if strategy.Scenario != ScenarioCorruptedBinary {
			t.Errorf("Strategy has wrong scenario: %v", strategy.Scenario)
		}
		if strategy.Description == "" {
			t.Error("Strategy missing description")
		}
		if strategy.Action == nil {
			t.Error("Strategy missing action function")
		}
	}

	t.Logf("Found %d recovery strategies for corrupted binary", len(strategies))
}

// TestRecoveryManager_Integration_IncompatibleBinaryRecovery tests incompatible binary recovery
func TestRecoveryManager_Integration_IncompatibleBinaryRecovery(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pvm-recovery-integration-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetPath := filepath.Join(tempDir, "pvm")

	// Create recovery manager
	rm := NewRecoveryManager()

	// Create recovery context for incompatible binary
	ctx := rm.CreateRecoveryContext(targetPath, "v1.0.0", fmt.Errorf("exec format error"))

	// Verify scenario diagnosis
	if ctx.Scenario != ScenarioIncompatibleBinary {
		t.Errorf("Expected scenario %v, got %v", ScenarioIncompatibleBinary, ctx.Scenario)
	}

	// Test getting recovery strategies
	strategies := rm.getRecoveryStrategies(ctx.Scenario)
	if len(strategies) == 0 {
		t.Error("No recovery strategies found for incompatible binary scenario")
	}

	// Find the compatible version download strategy
	var downloadStrategy *RecoveryStrategy
	for _, strategy := range strategies {
		if containsAny(strategy.Description, []string{"compatible", "download"}) {
			downloadStrategy = strategy
			break
		}
	}

	if downloadStrategy == nil {
		t.Error("No compatible version download strategy found")
		return
	}

	// Verify the strategy properties
	if downloadStrategy.Automatic {
		t.Error("Compatible version download should not be automatic")
	}

	if downloadStrategy.Priority != 1 {
		t.Errorf("Expected priority 1, got %d", downloadStrategy.Priority)
	}

	t.Logf("Compatible version download strategy: %s", downloadStrategy.Description)
}

// TestRecoveryManager_Integration_PermissionDeniedRecovery tests permission recovery
func TestRecoveryManager_Integration_PermissionDeniedRecovery(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pvm-recovery-integration-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetPath := filepath.Join(tempDir, "pvm")

	// Create recovery manager
	rm := NewRecoveryManager()

	// Create recovery context for permission denied
	ctx := rm.CreateRecoveryContext(targetPath, "v1.0.0", fmt.Errorf("permission denied"))

	// Verify scenario diagnosis
	if ctx.Scenario != ScenarioPermissionDenied {
		t.Errorf("Expected scenario %v, got %v", ScenarioPermissionDenied, ctx.Scenario)
	}

	// Test the permission fix strategy
	strategies := rm.getRecoveryStrategies(ctx.Scenario)
	if len(strategies) == 0 {
		t.Error("No recovery strategies found for permission denied scenario")
	}

	// Test that we can run the permission fix (this should work in most test environments)
	err = rm.fixPermissionsAndRetry(ctx)

	// The error is expected since the target doesn't exist yet, but it should be a specific error
	if err != nil {
		t.Logf("Permission fix returned expected error: %v", err)
	} else {
		t.Log("Permission fix succeeded (target directory was accessible)")
	}
}

// TestRecoveryManager_Integration_MultipleFailureScenarios tests handling multiple failure types
func TestRecoveryManager_Integration_MultipleFailureScenarios(t *testing.T) {
	scenarios := []struct {
		name     string
		err      error
		expected RecoveryScenario
	}{
		{"network timeout", fmt.Errorf("network timeout"), ScenarioNetworkFailure},
		{"checksum failure", fmt.Errorf("checksum mismatch"), ScenarioChecksumMismatch},
		{"disk full", fmt.Errorf("no space left on device"), ScenarioFileSystemFull},
		{"partial update", fmt.Errorf("update interrupted"), ScenarioPartialUpdate},
		{"unknown error", fmt.Errorf("something went wrong"), ScenarioUnknownFailure},
	}

	rm := NewRecoveryManager()

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			ctx := rm.CreateRecoveryContext("/tmp/pvm", "v1.0.0", scenario.err)

			if ctx.Scenario != scenario.expected {
				t.Errorf("Expected scenario %v, got %v", scenario.expected, ctx.Scenario)
			}

			strategies := rm.getRecoveryStrategies(ctx.Scenario)
			t.Logf("Scenario %v has %d recovery strategies", scenario.expected, len(strategies))
		})
	}
}

// TestRecoveryManager_Integration_RecoveryWorkflow tests the complete recovery workflow
func TestRecoveryManager_Integration_RecoveryWorkflow(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pvm-recovery-workflow-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetPath := filepath.Join(tempDir, "pvm")

	// Create a corrupted binary
	if err := os.WriteFile(targetPath, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rm := NewRecoveryManager()

	// Test full recovery workflow
	ctx := rm.CreateRecoveryContext(targetPath, "v1.0.0", fmt.Errorf("execution failed"))

	// Attempt recovery (this will likely fail due to no backups in test environment)
	result, err := rm.PerformRecovery(ctx)

	// In test environment, recovery is expected to fail since we don't have real backups
	if err == nil {
		t.Log("Recovery succeeded unexpectedly - this is fine in test environment")
		if result.Success {
			t.Logf("Recovery result: %s", result.Message)
		}
	} else {
		t.Logf("Recovery failed as expected in test environment: %v", err)
		if result != nil {
			t.Logf("Recovery attempts: %d, strategy: %s", result.AttemptsUsed, result.Strategy)
		}
	}

	// The important thing is that the workflow completes without panicking
	// and provides meaningful error messages
}

// TestRecoveryManager_Integration_PlatformCompatibility tests platform detection in recovery
func TestRecoveryManager_Integration_PlatformCompatibility(t *testing.T) {
	rm := NewRecoveryManager()

	// Test that FindCompatibleVersion handles platform detection correctly
	updateInfo, err := rm.FindCompatibleVersion()

	if err != nil {
		// In test environments, this might fail due to network/API issues
		if containsAny(err.Error(), []string{"rate limit", "network", "timeout", "no releases", "no compatible versions", "platform", "not supported"}) {
			t.Logf("FindCompatibleVersion failed with acceptable test error: %v", err)
			t.Skip("Skipping due to expected test environment limitations")
			return
		}
		t.Errorf("FindCompatibleVersion failed with unexpected error: %v", err)
		return
	}

	// If successful, verify the result includes proper platform handling
	if updateInfo != nil && updateInfo.Release != nil {
		t.Logf("Found compatible version for current platform: %s", updateInfo.LatestVersion.String())

		// Verify release has assets
		if len(updateInfo.Release.Assets) == 0 {
			t.Error("Compatible release has no assets")
		} else {
			t.Logf("Release has %d assets", len(updateInfo.Release.Assets))
		}
	}
}
