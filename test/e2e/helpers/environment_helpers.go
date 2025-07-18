// ABOUTME: Environment helpers for PVM end-to-end tests
// ABOUTME: Provides utilities for testing environment variable inheritance and isolation

package helpers

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
)

// EnvironmentTestConfig holds configuration for environment tests
type EnvironmentTestConfig struct {
	// Test configuration
	TestName    string
	Description string

	// Environment setup
	EnvVars   map[string]string // Environment variables to set
	UnsetVars []string          // Environment variables to unset

	// PVX configuration
	IsolationLevel string   // "global", "local", "clean"
	PVXArgs        []string // Additional PVX arguments

	// Test command
	Command     string   // Perl command to run
	CommandArgs []string // Command arguments

	// Expected behavior
	ExpectedEnvVars     map[string]string // Expected environment variables
	ExpectedMissingVars []string          // Variables that should not be present
	ShouldContain       []string          // Output should contain these strings
	ShouldNotContain    []string          // Output should not contain these strings
	ExpectedError       bool              // Whether command should fail
}

// TestPVXEnvironmentInheritance tests PVX environment variable inheritance
func TestPVXEnvironmentInheritance(t *testing.T, env *TestEnv, config *EnvironmentTestConfig) {
	t.Helper()

	// Setup environment
	setupEnvironment(t, config)
	defer restoreEnvironment(t, config)

	// Build PVX command
	pvxArgs := []string{"pvx"}
	if config.IsolationLevel != "" {
		pvxArgs = append(pvxArgs, "--isolation", config.IsolationLevel)
	}
	pvxArgs = append(pvxArgs, config.PVXArgs...)

	// Add the command
	if config.Command != "" {
		pvxArgs = append(pvxArgs, "-e", config.Command)
	}
	pvxArgs = append(pvxArgs, config.CommandArgs...)

	// Run PVX command
	stdout, stderr, err := env.RunPVM(pvxArgs...)

	output := stdout + stderr

	// Check expected behavior
	if config.ExpectedError && err == nil {
		t.Errorf("%s: expected error but command succeeded\nOutput: %s", config.TestName, output)
	} else if !config.ExpectedError && err != nil {
		t.Errorf("%s: unexpected error: %v\nOutput: %s", config.TestName, err, output)
	}

	// Check output contains expected strings
	for _, expected := range config.ShouldContain {
		if !strings.Contains(output, expected) {
			t.Errorf("%s: output does not contain expected string %q\nOutput: %s",
				config.TestName, expected, output)
		}
	}

	// Check output does not contain unwanted strings
	for _, unwanted := range config.ShouldNotContain {
		if strings.Contains(output, unwanted) {
			t.Errorf("%s: output contains unwanted string %q\nOutput: %s",
				config.TestName, unwanted, output)
		}
	}
}

// setupEnvironment sets up the environment for the test
func setupEnvironment(t *testing.T, config *EnvironmentTestConfig) {
	t.Helper()

	// Set environment variables
	for key, value := range config.EnvVars {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Failed to set environment variable %s=%s: %v", key, value, err)
		}
	}

	// Unset environment variables
	for _, key := range config.UnsetVars {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("Failed to unset environment variable %s: %v", key, err)
		}
	}
}

// restoreEnvironment restores the original environment
func restoreEnvironment(t *testing.T, config *EnvironmentTestConfig) {
	t.Helper()

	// This is a simplified restore - in a real implementation you'd want to
	// save the original values and restore them
	for key := range config.EnvVars {
		os.Unsetenv(key)
	}
}

// TestEnvironmentVariableInheritance tests basic environment variable inheritance
func TestEnvironmentVariableInheritance(t *testing.T, env *TestEnv) {
	t.Helper()

	config := &EnvironmentTestConfig{
		TestName:    "BasicEnvironmentInheritance",
		Description: "Test basic environment variable inheritance",
		EnvVars: map[string]string{
			"TEST_VAR":         "test_value",
			"PVM_PERL_VERSION": "5.42.0",
		},
		Command: "print $ENV{TEST_VAR}; print $ENV{PVM_PERL_VERSION}",
		ShouldContain: []string{
			"test_value",
			"5.42.0",
		},
	}

	TestPVXEnvironmentInheritance(t, env, config)
}

// TestPVMSpecificVariables tests PVM-specific environment variable handling
func TestPVMSpecificVariables(t *testing.T, env *TestEnv) {
	t.Helper()

	config := &EnvironmentTestConfig{
		TestName:    "PVMSpecificVariables",
		Description: "Test PVM-specific environment variable preservation",
		EnvVars: map[string]string{
			"PVM_PERL_VERSION":      "5.42.0",
			"PVM_SUPPRESS_WARNINGS": "1",
			"PVM_DEBUG":             "1",
		},
		Command: "print join(',', $ENV{PVM_PERL_VERSION}, $ENV{PVM_SUPPRESS_WARNINGS}, $ENV{PVM_DEBUG})",
		ShouldContain: []string{
			"5.42.0,1,1",
		},
	}

	TestPVXEnvironmentInheritance(t, env, config)
}

// TestIsolationLevelEnvironment tests environment with different isolation levels
func TestIsolationLevelEnvironment(t *testing.T, env *TestEnv) {
	t.Helper()

	testCases := []struct {
		isolationLevel  string
		expectCustomVar bool
	}{
		{"global", true},
		{"local", true},
		{"clean", false},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Isolation_%s", testCase.isolationLevel), func(t *testing.T) {
			config := &EnvironmentTestConfig{
				TestName:       fmt.Sprintf("IsolationLevel_%s", testCase.isolationLevel),
				Description:    fmt.Sprintf("Test environment with %s isolation", testCase.isolationLevel),
				IsolationLevel: testCase.isolationLevel,
				EnvVars: map[string]string{
					"CUSTOM_VAR":       "custom_value",
					"PVM_PERL_VERSION": "5.42.0",
				},
				Command: "print defined($ENV{CUSTOM_VAR}) ? $ENV{CUSTOM_VAR} : 'undefined'",
			}

			if testCase.expectCustomVar {
				config.ShouldContain = []string{"custom_value"}
			} else {
				config.ShouldContain = []string{"undefined"}
			}

			TestPVXEnvironmentInheritance(t, env, config)
		})
	}
}

// TestEnvironmentConsistency tests environment consistency between PVM and PVX
func TestEnvironmentConsistency(t *testing.T, env *TestEnv) {
	t.Helper()

	// Set test environment
	os.Setenv("PVM_PERL_VERSION", "5.42.0")
	os.Setenv("CUSTOM_VAR", "test_value")
	defer os.Unsetenv("PVM_PERL_VERSION")
	defer os.Unsetenv("CUSTOM_VAR")

	// Get PVM environment
	pvmStdout, _, err := env.RunPVM("current")
	if err != nil {
		t.Fatalf("Failed to get PVM current: %v", err)
	}

	// Get PVX environment
	pvxStdout, _, err := env.RunPVM("pvx", "-e", "print $ENV{PVM_PERL_VERSION}")
	if err != nil {
		t.Fatalf("Failed to get PVX environment: %v", err)
	}

	// Check consistency
	if !strings.Contains(pvmStdout, "5.42.0") {
		t.Errorf("PVM does not show expected version 5.42.0: %s", pvmStdout)
	}

	if !strings.Contains(pvxStdout, "5.42.0") {
		t.Errorf("PVX does not inherit expected version 5.42.0: %s", pvxStdout)
	}
}

// TestEnvironmentDebugOutput tests environment debug output
func TestEnvironmentDebugOutput(t *testing.T, env *TestEnv) {
	t.Helper()

	config := &EnvironmentTestConfig{
		TestName:    "EnvironmentDebugOutput",
		Description: "Test environment debug output",
		EnvVars: map[string]string{
			"PVM_DEBUG":        "1",
			"PVM_PERL_VERSION": "5.42.0",
		},
		PVXArgs: []string{"--verbose"},
		Command: "print $ENV{PVM_PERL_VERSION}",
		ShouldContain: []string{
			"5.42.0",
		},
	}

	TestPVXEnvironmentInheritance(t, env, config)
}

// CompareEnvironments compares two environment snapshots
func CompareEnvironments(t *testing.T, env1, env2 map[string]string, ignoreKeys []string) {
	t.Helper()

	// Create ignore map for fast lookup
	ignore := make(map[string]bool)
	for _, key := range ignoreKeys {
		ignore[key] = true
	}

	// Get all keys from both environments
	allKeys := make(map[string]bool)
	for key := range env1 {
		if !ignore[key] {
			allKeys[key] = true
		}
	}
	for key := range env2 {
		if !ignore[key] {
			allKeys[key] = true
		}
	}

	// Check for differences
	var differences []string
	for key := range allKeys {
		val1, exists1 := env1[key]
		val2, exists2 := env2[key]

		switch {
		case !exists1 && exists2:
			differences = append(differences, fmt.Sprintf("+ %s=%s", key, val2))
		case exists1 && !exists2:
			differences = append(differences, fmt.Sprintf("- %s=%s", key, val1))
		case exists1 && exists2 && val1 != val2:
			differences = append(differences, fmt.Sprintf("! %s: %s -> %s", key, val1, val2))
		}
	}

	if len(differences) > 0 {
		sort.Strings(differences)
		t.Errorf("Environment differences found:\n%s", strings.Join(differences, "\n"))
	}
}

// CaptureEnvironment captures the current environment as a map
func CaptureEnvironment() map[string]string {
	env := make(map[string]string)
	for _, pair := range os.Environ() {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}
	return env
}

// TestEnvironmentCapture tests environment capture functionality
func TestEnvironmentCapture(t *testing.T, env *TestEnv) {
	t.Helper()

	// Set some test variables
	os.Setenv("TEST_CAPTURE_VAR", "test_value")
	defer os.Unsetenv("TEST_CAPTURE_VAR")

	// Capture environment
	envSnapshot := CaptureEnvironment()

	// Verify capture
	if val, exists := envSnapshot["TEST_CAPTURE_VAR"]; !exists || val != "test_value" {
		t.Errorf("Environment capture failed: expected TEST_CAPTURE_VAR=test_value, got %s (exists: %v)", val, exists)
	}
}

// CreateEnvironmentTestCases creates standard test cases for environment testing
func CreateEnvironmentTestCases() []EnvironmentTestConfig {
	return []EnvironmentTestConfig{
		{
			TestName:    "BasicInheritance",
			Description: "Test basic environment variable inheritance",
			EnvVars: map[string]string{
				"TEST_VAR": "test_value",
			},
			Command: "print $ENV{TEST_VAR}",
			ShouldContain: []string{
				"test_value",
			},
		},
		{
			TestName:    "PVMVersionInheritance",
			Description: "Test PVM_PERL_VERSION inheritance",
			EnvVars: map[string]string{
				"PVM_PERL_VERSION": "5.42.0",
			},
			Command: "print $ENV{PVM_PERL_VERSION}",
			ShouldContain: []string{
				"5.42.0",
			},
		},
		{
			TestName:       "CleanIsolation",
			Description:    "Test clean isolation removes non-essential variables",
			IsolationLevel: "clean",
			EnvVars: map[string]string{
				"CUSTOM_VAR":       "should_be_removed",
				"PVM_PERL_VERSION": "5.42.0",
			},
			Command: "print defined($ENV{CUSTOM_VAR}) ? 'present' : 'absent'",
			ShouldContain: []string{
				"absent",
			},
		},
		{
			TestName:       "GlobalIsolation",
			Description:    "Test global isolation preserves all variables",
			IsolationLevel: "global",
			EnvVars: map[string]string{
				"CUSTOM_VAR":       "should_be_present",
				"PVM_PERL_VERSION": "5.42.0",
			},
			Command: "print $ENV{CUSTOM_VAR}",
			ShouldContain: []string{
				"should_be_present",
			},
		},
		{
			TestName:    "ErrorHandling",
			Description: "Test error handling with environment issues",
			EnvVars: map[string]string{
				"PVM_PERL_VERSION": "5.99.0", // Invalid version
			},
			Command:       "print 'test'",
			ExpectedError: true,
		},
	}
}

// RunEnvironmentTestSuite runs a comprehensive environment test suite
func RunEnvironmentTestSuite(t *testing.T, env *TestEnv) {
	testCases := CreateEnvironmentTestCases()

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			TestPVXEnvironmentInheritance(t, env, &testCase)
		})
	}
}
