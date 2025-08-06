// ABOUTME: Test suite for PVX -E (execute with features) flag functionality
// ABOUTME: Ensures -E flag properly enables Perl feature bundles like say, state, etc.
package e2e

import (
	"strings"
	"testing"

	"tamarou.com/pvm/test/e2e/helpers"
)

// TestPVXExecuteFeatures tests the -E flag functionality that enables Perl features
func TestPVXExecuteFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping -E flag tests in short mode")
	}

	env := helpers.NewTestEnv(t)
	defer env.Cleanup()
	
	// Test 1: -E flag should enable 'say' feature
	t.Run("EnablesSayFeature", func(t *testing.T) {
		// This should work with -E (features enabled)
		stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{"pvx", "--execute-features", "say 'hello from -E'"},
			"PVX -E flag with say feature")
		
		expected := "hello from -E"
		actual := strings.TrimSpace(stdout)
		if actual != expected {
			t.Errorf("Expected %q, got %q", expected, actual)
		}
	})
	
	// Test 2: -e flag should fail with 'say' feature
	t.Run("PlainEFlagFailsWithSay", func(t *testing.T) {
		// This should fail with -e (no features)
		stdout, stderr, err := env.RunPVM("pvx", "-e", "say 'hello from -e'")
		if err == nil {
			t.Fatal("Expected PVX -e with say to fail, but it succeeded")
		}
		
		// Check that it failed with a compilation error (either in stdout or stderr)
		combinedOutput := stdout + stderr
		if !strings.Contains(combinedOutput, "Do you need to predeclare say") && 
		   !strings.Contains(combinedOutput, "exit code 255") {
			t.Errorf("Expected Perl compilation error or exit code 255, got stdout: %s, stderr: %s", stdout, stderr)
		}
	})
	
	// Test 3: -E should enable 'state' variables
	t.Run("EnablesStateVariables", func(t *testing.T) {
		// Test state variable with -E (should work)
		stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{"pvx", "--execute-features", "state $x = 0; say ++$x; say ++$x"},
			"PVX -E flag with state variables")
		
		lines := strings.Split(strings.TrimSpace(stdout), "\n")
		if len(lines) != 2 {
			t.Fatalf("Expected 2 lines of output, got %d: %v", len(lines), lines)
		}
		
		// Should output 1, 2 (state variable incrementing)
		if lines[0] != "1" || lines[1] != "2" {
			t.Errorf("Expected state increment '1' and '2', got: %v", lines)
		}
	})
	
	// Test 4: Version output with features
	t.Run("VersionOutputWithFeatures", func(t *testing.T) {
		// Test that -E works for version display using say
		stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{"pvx", "--execute-features", "say $^V"},
			"PVX -E flag version output")
		
		versionStr := strings.TrimSpace(stdout)
		if !strings.HasPrefix(versionStr, "v") {
			t.Errorf("Expected version string starting with 'v', got: %s", versionStr)
		}
	})
}

// TestPVXExecuteFeaturesEdgeCases tests edge cases and error conditions for -E flag
func TestPVXExecuteFeaturesEdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping -E flag edge case tests in short mode")
	}

	env := helpers.NewTestEnv(t)
	defer env.Cleanup()
	
	// Test 1: Minimal code with -E
	t.Run("MinimalCodeWithFeatures", func(t *testing.T) {
		// Minimal code should succeed but produce no output
		stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{"pvx", "--execute-features", ";"},
			"PVX -E flag with minimal code")
		
		if len(strings.TrimSpace(stdout)) != 0 {
			t.Errorf("Expected no output for minimal code, got: %q", stdout)
		}
	})
	
	// Test 2: Complex feature usage  
	t.Run("ComplexFeatureUsage", func(t *testing.T) {
		// Test combining multiple modern features in one line to avoid parsing issues
		code := "state @items = qw(a b c); say \"Item: $_\" for @items"
		
		stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env,
			[]string{"pvx", "--execute-features", code},
			"PVX -E flag with complex features")
		
		// Check that it contains expected patterns
		if !strings.Contains(stdout, "Item: a") || !strings.Contains(stdout, "Item: b") {
			t.Errorf("Output doesn't contain expected items: %s", stdout)
		}
	})
}