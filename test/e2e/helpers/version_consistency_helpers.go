// ABOUTME: Version consistency helpers for PVM end-to-end tests
// ABOUTME: Provides utilities for testing version consistency between PVM and PVX

package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// VersionConsistencyTestConfig holds configuration for version consistency tests
type VersionConsistencyTestConfig struct {
	// Test configuration
	TestName    string
	Description string

	// Version setup
	Version        string
	SetupMethod    string // "file", "env", "command", "config"
	PrecedenceTest bool   // Test precedence rules

	// Expected behavior
	ExpectedVersion string
	ExpectedError   bool

	// Test environment
	WorkingDir   string
	EnvVars      map[string]string
	ConfigValues map[string]string
}

// AssertVersionConsistency checks that PVM and PVX use the same version
func AssertVersionConsistency(t *testing.T, env *TestEnv, config *VersionConsistencyTestConfig) {
	t.Helper()

	// Setup test environment based on config
	setupVersionEnvironment(t, env, config)

	// Get version from PVM current
	pvmVersion, pvmErr := getPVMCurrentVersion(t, env)

	// Get version from PVX
	pvxVersion, pvxErr := getPVXVersion(t, env)

	// Check error consistency
	if config.ExpectedError {
		if pvmErr == nil || pvxErr == nil {
			t.Errorf("%s: expected both PVM and PVX to fail, but PVM err=%v, PVX err=%v",
				config.TestName, pvmErr, pvxErr)
		}
		return
	}

	// Check for unexpected errors
	if pvmErr != nil {
		t.Errorf("%s: PVM current failed: %v", config.TestName, pvmErr)
	}
	if pvxErr != nil {
		t.Errorf("%s: PVX version detection failed: %v", config.TestName, pvxErr)
	}

	// Check version consistency
	if pvmVersion != pvxVersion {
		t.Errorf("%s: version mismatch - PVM shows %q, PVX uses %q",
			config.TestName, pvmVersion, pvxVersion)
	}

	// Check against expected version if specified
	if config.ExpectedVersion != "" {
		if pvmVersion != config.ExpectedVersion {
			t.Errorf("%s: PVM version %q does not match expected %q",
				config.TestName, pvmVersion, config.ExpectedVersion)
		}
		if pvxVersion != config.ExpectedVersion {
			t.Errorf("%s: PVX version %q does not match expected %q",
				config.TestName, pvxVersion, config.ExpectedVersion)
		}
	}
}

// setupVersionEnvironment sets up the test environment based on the config
func setupVersionEnvironment(t *testing.T, env *TestEnv, config *VersionConsistencyTestConfig) {
	t.Helper()

	// Change to the test root directory first
	if err := os.Chdir(env.RootDir); err != nil {
		t.Fatalf("Failed to change to test root directory %s: %v", env.RootDir, err)
	}

	// Set working directory if specified
	if config.WorkingDir != "" {
		workDir := filepath.Join(env.RootDir, config.WorkingDir)
		if err := os.MkdirAll(workDir, 0755); err != nil {
			t.Fatalf("Failed to create working directory %s: %v", workDir, err)
		}
		if err := os.Chdir(workDir); err != nil {
			t.Fatalf("Failed to change to working directory %s: %v", workDir, err)
		}
	}

	// Set environment variables
	for key, value := range config.EnvVars {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Failed to set environment variable %s=%s: %v", key, value, err)
		}
	}

	// Setup version based on method
	switch config.SetupMethod {
	case "file":
		setupVersionFile(t, env, config)
	case "env":
		setupVersionEnv(t, env, config)
	case "command":
		setupVersionCommand(t, env, config)
	case "config":
		setupVersionConfig(t, env, config)
	default:
		t.Fatalf("Unknown setup method: %s", config.SetupMethod)
	}
}

// setupVersionFile creates a .perl-version file
func setupVersionFile(t *testing.T, env *TestEnv, config *VersionConsistencyTestConfig) {
	t.Helper()

	// Get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Create .perl-version file in current directory
	versionFile := filepath.Join(currentDir, ".perl-version")
	if err := os.WriteFile(versionFile, []byte(config.Version), 0644); err != nil {
		t.Fatalf("Failed to create .perl-version file at %s: %v", versionFile, err)
	}

	// Also create it in the test root for consistency
	rootVersionFile := filepath.Join(env.RootDir, ".perl-version")
	if err := os.WriteFile(rootVersionFile, []byte(config.Version), 0644); err != nil {
		t.Fatalf("Failed to create root .perl-version file at %s: %v", rootVersionFile, err)
	}

	// Debug output
	t.Logf("Created .perl-version file at %s with content: %q", versionFile, config.Version)
	t.Logf("Created .perl-version file at %s with content: %q", rootVersionFile, config.Version)
	t.Logf("Current working directory: %s", currentDir)
}

// setupVersionEnv sets the PVM_PERL_VERSION environment variable
func setupVersionEnv(t *testing.T, env *TestEnv, config *VersionConsistencyTestConfig) {
	t.Helper()

	if err := os.Setenv("PVM_PERL_VERSION", config.Version); err != nil {
		t.Fatalf("Failed to set PVM_PERL_VERSION: %v", err)
	}
}

// setupVersionCommand runs pvm use to set the version
func setupVersionCommand(t *testing.T, env *TestEnv, config *VersionConsistencyTestConfig) {
	t.Helper()

	_, _, err := env.RunPVM("use", config.Version)
	if err != nil {
		t.Fatalf("Failed to run pvm use %s: %v", config.Version, err)
	}
}

// setupVersionConfig sets version via configuration
func setupVersionConfig(t *testing.T, env *TestEnv, config *VersionConsistencyTestConfig) {
	t.Helper()

	// Set default version in config
	_, _, err := env.RunPVM("config", "set", "perl.default_version", config.Version)
	if err != nil {
		t.Fatalf("Failed to set default version config: %v", err)
	}
}

// getPVMCurrentVersion gets the current version from PVM
func getPVMCurrentVersion(t *testing.T, env *TestEnv) (string, error) {
	t.Helper()

	stdout, stderr, err := env.RunPVM("current")
	if err != nil {
		return "", fmt.Errorf("pvm current failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Debug output
	t.Logf("PVM current stdout: %q", stdout)
	t.Logf("PVM current stderr: %q", stderr)

	// Extract version from output - check both stdout and stderr
	combinedOutput := stdout + stderr
	version := extractVersionFromPVMOutput(combinedOutput)
	if version == "" {
		return "", fmt.Errorf("could not extract version from PVM output: stdout=%q, stderr=%q", stdout, stderr)
	}

	return version, nil
}

// getPVXVersion gets the version used by PVX
func getPVXVersion(t *testing.T, env *TestEnv) (string, error) {
	t.Helper()

	stdout, stderr, err := env.RunPVM("pvx", "-e", "print $^V")
	if err != nil {
		return "", fmt.Errorf("pvx version detection failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Debug output
	t.Logf("PVX version stdout: %q", stdout)
	t.Logf("PVX version stderr: %q", stderr)

	// Extract version from Perl's $^V output
	version := extractVersionFromPerlOutput(stdout)
	if version == "" {
		return "", fmt.Errorf("could not extract version from PVX output: %q", stdout)
	}

	return version, nil
}

// extractVersionFromPVMOutput extracts version from PVM current output
func extractVersionFromPVMOutput(output string) string {
	// Handle different PVM output formats
	output = strings.TrimSpace(output)

	// Format: "5.42.0 (set by .perl-version file)"
	// Format: "5.42.0"
	// Format: "5.38.2 (system Perl at /usr/bin/perl)"
	// Format: "system (5.38.2)"
	// Format: "system"

	// Try to extract version number pattern first
	versionPattern := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
	matches := versionPattern.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1]
	}

	// Handle "system" case without version number
	if strings.Contains(output, "system") && !strings.Contains(output, "(") {
		return "system"
	}

	// If no version found, return empty string
	return ""
}

// extractVersionFromPerlOutput extracts version from Perl's $^V output
func extractVersionFromPerlOutput(output string) string {
	// Perl $^V output format varies by version
	// v5.42.0 or similar
	output = strings.TrimSpace(output)

	// Remove leading 'v' if present
	if strings.HasPrefix(output, "v") {
		output = output[1:]
	}

	// Extract version pattern
	versionPattern := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
	matches := versionPattern.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// CreateVersionHierarchy creates a directory structure with multiple .perl-version files
func CreateVersionHierarchy(t *testing.T, env *TestEnv, versions map[string]string) {
	t.Helper()

	for dir, version := range versions {
		fullDir := filepath.Join(env.RootDir, dir)
		if err := os.MkdirAll(fullDir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", fullDir, err)
		}

		versionFile := filepath.Join(dir, ".perl-version")
		if err := env.CreateFile(versionFile, version); err != nil {
			t.Fatalf("Failed to create .perl-version file in %s: %v", dir, err)
		}
	}
}

// TestVersionResolutionPerformance measures the performance of version resolution
func TestVersionResolutionPerformance(t *testing.T, env *TestEnv, description string) time.Duration {
	t.Helper()

	start := time.Now()

	// Run version resolution
	_, _, err := env.RunPVM("current")
	if err != nil {
		t.Fatalf("Version resolution failed during performance test: %v", err)
	}

	duration := time.Since(start)

	// Log performance for analysis
	t.Logf("Version resolution performance (%s): %v", description, duration)

	// Warn if resolution takes too long
	if duration > 5*time.Second {
		t.Logf("WARNING: Version resolution took %v, which may be too slow", duration)
	}

	return duration
}

// ValidateVersionResolutionTrace validates that version resolution debugging works
func ValidateVersionResolutionTrace(t *testing.T, env *TestEnv, expectedSources []string) {
	t.Helper()

	// Run PVM with debug output
	os.Setenv("PVM_DEBUG", "1")
	defer os.Unsetenv("PVM_DEBUG")

	stdout, stderr, err := env.RunPVM("current")
	if err != nil {
		t.Fatalf("Version resolution failed with debug enabled: %v", err)
	}

	// Check debug output for expected sources
	debugOutput := stdout + stderr
	for _, source := range expectedSources {
		if !strings.Contains(debugOutput, source) {
			t.Errorf("Debug output does not contain expected source %q\nOutput: %s", source, debugOutput)
		}
	}
}

// CreateVersionResolutionTestCases creates standard test cases for version resolution
func CreateVersionResolutionTestCases() []VersionConsistencyTestConfig {
	return []VersionConsistencyTestConfig{
		{
			TestName:        "BasicPerlVersionFile",
			Description:     "Test basic .perl-version file resolution",
			Version:         "5.42.0",
			SetupMethod:     "file",
			ExpectedVersion: "5.42.0",
		},
		{
			TestName:        "EnvironmentVariablePrecedence",
			Description:     "Test PVM_PERL_VERSION environment variable precedence",
			Version:         "5.40.0",
			SetupMethod:     "env",
			ExpectedVersion: "5.40.0",
			EnvVars: map[string]string{
				"PVM_PERL_VERSION": "5.40.0",
			},
		},
		{
			TestName:        "CommandLinePrecedence",
			Description:     "Test command line version precedence",
			Version:         "5.38.0",
			SetupMethod:     "command",
			ExpectedVersion: "5.38.0",
		},
		{
			TestName:        "ConfigurationFallback",
			Description:     "Test configuration fallback",
			Version:         "5.36.0",
			SetupMethod:     "config",
			ExpectedVersion: "5.36.0",
		},
		{
			TestName:      "InvalidVersionError",
			Description:   "Test invalid version error handling",
			Version:       "5.99.0",
			SetupMethod:   "file",
			ExpectedError: true,
		},
	}
}

// RunVersionConsistencyTestSuite runs a comprehensive version consistency test suite
func RunVersionConsistencyTestSuite(t *testing.T, env *TestEnv) {
	testCases := CreateVersionResolutionTestCases()

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			AssertVersionConsistency(t, env, &testCase)
		})
	}
}
