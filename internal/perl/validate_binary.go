// ABOUTME: Binary installation validation functionality
// ABOUTME: Provides validation tools for binary Perl installations

package perl

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"tamarou.com/pvm/internal/errors"
)

// Binary validation error codes
const (
	ErrBinaryValidationFailed  = "801" // Binary validation failed
	ErrBinaryExecutionFailed   = "802" // Binary execution test failed
	ErrBinaryVersionMismatch   = "803" // Binary version doesn't match expected
	ErrBinaryIncompleteInstall = "804" // Binary installation is incomplete
)

// BinaryValidationResult contains the results of binary installation validation
type BinaryValidationResult struct {
	// Whether the installation is valid
	Valid bool

	// List of validation warnings (non-fatal issues)
	Warnings []string

	// Detected Perl version
	DetectedVersion string

	// Path to the Perl executable
	PerlExecutable string

	// Installation completeness score (0.0 to 1.0)
	CompletenessScore float64

	// Performance benchmark results (if requested)
	BenchmarkResults *BinaryBenchmarkResults
}

// BinaryBenchmarkResults contains performance benchmark data
type BinaryBenchmarkResults struct {
	// Time to start Perl and print version
	StartupTime time.Duration

	// Time to execute a simple script
	SimpleScriptTime time.Duration

	// Time to load a common module
	ModuleLoadTime time.Duration
}

// ValidateBinaryInstallation validates a binary Perl installation
func ValidateBinaryInstallation(installPath string) (bool, []string, error) {
	result, err := ValidateBinaryInstallationDetailed(installPath, false)
	if err != nil {
		return false, nil, err
	}

	return result.Valid, result.Warnings, nil
}

// ValidateBinaryInstallationDetailed performs comprehensive validation of a binary installation
func ValidateBinaryInstallationDetailed(installPath string, includeBenchmarks bool) (*BinaryValidationResult, error) {
	result := &BinaryValidationResult{
		Valid:             true,
		Warnings:          make([]string, 0),
		CompletenessScore: 0.0,
	}

	// Check if installation directory exists
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		return &BinaryValidationResult{
				Valid: false,
			}, errors.NewSystemError(ErrBinaryValidationFailed,
				"Installation directory does not exist", err).
				WithLocation(installPath)
	}

	// Validate directory structure
	score, warnings := validateDirectoryStructure(installPath)
	result.CompletenessScore += score * 0.3 // 30% weight
	result.Warnings = append(result.Warnings, warnings...)

	// Validate Perl executable
	perlExe := filepath.Join(installPath, "bin", "perl")
	if runtime.GOOS == "windows" {
		perlExe = filepath.Join(installPath, "bin", "perl.exe")
	}

	execScore, execWarnings, err := validatePerlExecutable(perlExe)
	if err != nil {
		result.Valid = false
		return result, err
	}
	result.CompletenessScore += execScore * 0.4 // 40% weight
	result.Warnings = append(result.Warnings, execWarnings...)
	result.PerlExecutable = perlExe

	// Test Perl execution and get version
	versionScore, version, versionWarnings, err := testPerlExecution(perlExe)
	if err != nil {
		result.Valid = false
		return result, err
	}
	result.CompletenessScore += versionScore * 0.3 // 30% weight
	result.DetectedVersion = version
	result.Warnings = append(result.Warnings, versionWarnings...)

	// Run performance benchmarks if requested
	if includeBenchmarks {
		benchmarks, err := runBinaryBenchmarks(perlExe)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Benchmark failed: %v", err))
		} else {
			result.BenchmarkResults = benchmarks
		}
	}

	// Overall validation
	if result.CompletenessScore < 0.5 {
		result.Valid = false
	}

	return result, nil
}

// ValidateDirectoryStructure checks if the installation has the expected directory structure
func ValidateDirectoryStructure(installPath string) (float64, []string) {
	return validateDirectoryStructure(installPath)
}

// validateDirectoryStructure checks if the installation has the expected directory structure
func validateDirectoryStructure(installPath string) (float64, []string) {
	var score float64
	var warnings []string

	// Expected directories and their importance weights
	expectedDirs := map[string]float64{
		"bin":   0.4,  // Most important - contains executables
		"lib":   0.3,  // Perl libraries
		"man":   0.15, // Manual pages (optional)
		"share": 0.15, // Shared data (optional)
	}

	// Optional directories that don't affect score but might generate warnings
	optionalDirs := []string{"include", "etc"}

	for dir, weight := range expectedDirs {
		dirPath := filepath.Join(installPath, dir)
		if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
			score += weight
		} else {
			if weight >= 0.3 { // Important directories
				warnings = append(warnings, fmt.Sprintf("Missing important directory: %s", dir))
			} else {
				warnings = append(warnings, fmt.Sprintf("Missing optional directory: %s", dir))
			}
		}
	}

	// Check for unexpected directories that might indicate a problem
	entries, err := os.ReadDir(installPath)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				name := entry.Name()
				_, isExpected := expectedDirs[name]
				isOptional := false
				for _, optional := range optionalDirs {
					if name == optional {
						isOptional = true
						break
					}
				}

				if !isExpected && !isOptional && !strings.HasPrefix(name, ".") {
					warnings = append(warnings, fmt.Sprintf("Unexpected directory found: %s", name))
				}
			}
		}
	}

	return score, warnings
}

// validatePerlExecutable checks if the Perl executable exists and has proper permissions
func validatePerlExecutable(perlPath string) (float64, []string, error) {
	var score float64
	var warnings []string

	// Check if file exists
	info, err := os.Stat(perlPath)
	if os.IsNotExist(err) {
		return 0.0, warnings, errors.NewSystemError(ErrBinaryValidationFailed,
			"Perl executable not found", err).
			WithLocation(perlPath)
	}
	if err != nil {
		return 0.0, warnings, errors.NewSystemError(ErrBinaryValidationFailed,
			"Failed to stat Perl executable", err).
			WithLocation(perlPath)
	}

	score += 0.5 // File exists

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		warnings = append(warnings, "Perl executable is not a regular file")
	} else {
		score += 0.2
	}

	// Check permissions (Unix only)
	if runtime.GOOS != "windows" {
		if info.Mode()&0111 == 0 {
			return score, warnings, errors.NewSystemError(ErrBinaryValidationFailed,
				"Perl executable is not executable", nil).
				WithLocation(perlPath)
		}
		score += 0.3 // Proper permissions
	} else {
		score += 0.3 // On Windows, assume permissions are OK if file exists
	}

	return score, warnings, nil
}

// testPerlExecution tests if Perl can be executed and returns version information
func testPerlExecution(perlPath string) (float64, string, []string, error) {
	var score float64
	var warnings []string

	// Test basic execution with version flag
	cmd := exec.Command(perlPath, "-v")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return 0.0, "", warnings, errors.NewSystemError(ErrBinaryExecutionFailed,
				fmt.Sprintf("Perl execution failed with exit code %d", exitErr.ExitCode()),
				err).WithLocation(perlPath)
		}
		return 0.0, "", warnings, errors.NewSystemError(ErrBinaryExecutionFailed,
			"Failed to execute Perl", err).
			WithLocation(perlPath)
	}

	score += 0.5 // Perl executes successfully

	// Extract version from output
	version := extractVersionFromOutput(string(output))
	if version == "" {
		warnings = append(warnings, "Could not extract version from Perl output")
	} else {
		score += 0.3 // Version detected
	}

	// Test simple script execution
	cmd = exec.Command(perlPath, "-e", "print 'Hello World'")
	output, err = cmd.Output()
	switch {
	case err != nil:
		warnings = append(warnings, fmt.Sprintf("Simple script execution failed: %v", err))
	case string(output) != "Hello World":
		warnings = append(warnings, "Simple script output is incorrect")
	default:
		score += 0.2 // Simple script works
	}

	return score, version, warnings, nil
}

// extractVersionFromOutput extracts Perl version from the output of 'perl -v'
func extractVersionFromOutput(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Look for lines containing version information
		if strings.Contains(line, "This is perl") || strings.Contains(line, "version") {
			// Try to extract version number (e.g., "v5.38.0", "5.38.0")
			words := strings.Fields(line)
			for _, word := range words {
				// Remove common punctuation and parentheses
				cleanWord := strings.Trim(word, "(),")

				if strings.HasPrefix(cleanWord, "v") && len(cleanWord) > 1 {
					version := strings.TrimPrefix(cleanWord, "v")
					if isValidVersionPattern(version) {
						return version
					}
				}

				// Look for patterns like "5.xx.x"
				if isValidVersionPattern(cleanWord) {
					return cleanWord
				}
			}
		}
	}
	return ""
}

// isValidVersionPattern checks if a string looks like a valid Perl version
func isValidVersionPattern(s string) bool {
	if len(s) < 3 { // Minimum "5.0"
		return false
	}

	// Must start with a digit (typically 5 for Perl)
	if s[0] < '0' || s[0] > '9' {
		return false
	}

	// Must contain at least one dot
	if !strings.Contains(s, ".") {
		return false
	}

	// Check format: digits, dots, and maybe more digits
	for i, char := range s {
		if char != '.' && (char < '0' || char > '9') {
			return false
		}
		// First character after dot should be a digit
		if i > 0 && s[i-1] == '.' && (char < '0' || char > '9') {
			return false
		}
	}

	// Shouldn't end with a dot
	if strings.HasSuffix(s, ".") {
		return false
	}

	return true
}

// runBinaryBenchmarks runs performance benchmarks on the Perl installation
func runBinaryBenchmarks(perlPath string) (*BinaryBenchmarkResults, error) {
	results := &BinaryBenchmarkResults{}

	// Benchmark 1: Startup time
	start := time.Now()
	cmd := exec.Command(perlPath, "-e", "exit 0")
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("startup benchmark failed: %w", err)
	}
	results.StartupTime = time.Since(start)

	// Benchmark 2: Simple script execution
	start = time.Now()
	cmd = exec.Command(perlPath, "-e", "for my $i (1..1000) { $i * $i }")
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("simple script benchmark failed: %w", err)
	}
	results.SimpleScriptTime = time.Since(start)

	// Benchmark 3: Module loading (use a common core module)
	start = time.Now()
	cmd = exec.Command(perlPath, "-e", "use strict; use warnings; use File::Spec;")
	err = cmd.Run()
	if err != nil {
		// Not a critical failure - some installations might not have all modules
		results.ModuleLoadTime = time.Duration(-1) // Indicate failure
	} else {
		results.ModuleLoadTime = time.Since(start)
	}

	return results, nil
}

// ValidateExpectedVersion checks if the installed version matches the expected version
func ValidateExpectedVersion(installPath, expectedVersion string) error {
	result, err := ValidateBinaryInstallationDetailed(installPath, false)
	if err != nil {
		return err
	}

	if !result.Valid {
		return errors.NewSystemError(ErrBinaryValidationFailed,
			"Binary installation validation failed", nil).
			WithLocation(installPath)
	}

	if result.DetectedVersion == "" {
		return errors.NewSystemError(ErrBinaryVersionMismatch,
			"Could not detect Perl version", nil).
			WithLocation(installPath)
	}

	// Parse both versions for comparison
	expected, err := ParseVersion(expectedVersion)
	if err != nil {
		return errors.NewVersionError(ErrBinaryVersionMismatch,
			fmt.Sprintf("Invalid expected version format: %s", expectedVersion), err)
	}

	detected, err := ParseVersion(result.DetectedVersion)
	if err != nil {
		return errors.NewVersionError(ErrBinaryVersionMismatch,
			fmt.Sprintf("Invalid detected version format: %s", result.DetectedVersion), err)
	}

	if expected.String() != detected.String() {
		return errors.NewVersionError(ErrBinaryVersionMismatch,
			fmt.Sprintf("Version mismatch: expected %s, got %s", expected.String(), detected.String()),
			nil).WithLocation(installPath)
	}

	return nil
}
