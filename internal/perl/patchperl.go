// ABOUTME: Devel::PatchPerl integration for cross-platform Perl source patching
// ABOUTME: Provides automatic patching of Perl source code for compatibility across platforms

package perl

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// PatchPerlOptions configures patchperl behavior
type PatchPerlOptions struct {
	// SourceDir is the Perl source directory to patch
	SourceDir string

	// PerlVersion is the Perl version being built
	PerlVersion string

	// Verbose enables verbose patchperl output
	Verbose bool

	// Context for command execution
	Context context.Context

	// Timeout for patchperl execution (default: 5 minutes)
	Timeout time.Duration
}

// PatchPerlResult contains information about the patching process
type PatchPerlResult struct {
	// Applied indicates whether patches were applied
	Applied bool

	// PatchCount is the number of patches applied
	PatchCount int

	// Output contains the patchperl command output
	Output string

	// Duration is how long patching took
	Duration time.Duration
}

// IsPatchPerlAvailable checks if Devel::PatchPerl is available via system perl
func IsPatchPerlAvailable() bool {
	// Try to run patchperl --help to check availability
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "perl", "-MDevel::PatchPerl", "-e", "print Devel::PatchPerl->VERSION")
	err := cmd.Run()
	return err == nil
}

// InstallPatchPerl installs Devel::PatchPerl using PVM's PM command
func InstallPatchPerl(verbose bool) error {
	if verbose {
		fmt.Println("Installing Devel::PatchPerl using PVM PM...")
	}

	// Find the PM command (should be in the same directory as the current executable or in PATH)
	pmCmd, err := findPMCommand()
	if err != nil {
		return fmt.Errorf("failed to find PM command: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Build PM install command
	args := []string{"install", "Devel::PatchPerl"}
	if !verbose {
		args = append(args, "--quiet")
	}

	cmd := exec.CommandContext(ctx, pmCmd, args...)

	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Printf("Running: %s %s\n", pmCmd, strings.Join(args, " "))
	}

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("PM install failed: %w", err)
	}

	if verbose {
		fmt.Println("Devel::PatchPerl installed successfully via PM")
	}

	return nil
}

// findPMCommand finds the PM command executable
func findPMCommand() (string, error) {
	// First try to find 'pm' in PATH
	if pmPath, err := exec.LookPath("pm"); err == nil {
		return pmPath, nil
	}

	// If not in PATH, try to find it relative to the current executable
	if execPath, err := os.Executable(); err == nil {
		// Check if pm is in the same directory as the current executable
		execDir := filepath.Dir(execPath)
		pmPath := filepath.Join(execDir, "pm")
		if _, err := os.Stat(pmPath); err == nil {
			return pmPath, nil
		}

		// Check for pm.exe on Windows
		pmExePath := filepath.Join(execDir, "pm.exe")
		if _, err := os.Stat(pmExePath); err == nil {
			return pmExePath, nil
		}

		// Check in build directory (during development)
		buildDir := filepath.Join(filepath.Dir(execDir), "build")
		buildPMPath := filepath.Join(buildDir, "pm")
		if _, err := os.Stat(buildPMPath); err == nil {
			return buildPMPath, nil
		}
	}

	return "", fmt.Errorf("PM command not found in PATH or relative to executable")
}

// ApplyPatchPerl applies Devel::PatchPerl patches to Perl source
func ApplyPatchPerl(options *PatchPerlOptions) (*PatchPerlResult, error) {
	if options == nil {
		return nil, fmt.Errorf("PatchPerlOptions cannot be nil")
	}

	// Set defaults
	if options.Context == nil {
		options.Context = context.Background()
	}
	if options.Timeout == 0 {
		options.Timeout = 5 * time.Minute
	}

	// Validate source directory
	if options.SourceDir == "" {
		return nil, fmt.Errorf("SourceDir is required")
	}

	// Check if source directory exists
	if _, err := os.Stat(options.SourceDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("source directory does not exist: %s", options.SourceDir)
	}

	startTime := time.Now()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(options.Context, options.Timeout)
	defer cancel()

	// Build patchperl command
	args := []string{
		"-MDevel::PatchPerl",
		"-e",
		"Devel::PatchPerl->patch_source()",
	}

	cmd := exec.CommandContext(ctx, "perl", args...)
	cmd.Dir = options.SourceDir

	// Capture output
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	result := &PatchPerlResult{
		Output:   outputStr,
		Duration: time.Since(startTime),
	}

	if err != nil {
		return result, fmt.Errorf("patchperl failed: %w\nOutput: %s", err, outputStr)
	}

	// Parse output to determine if patches were applied
	result.Applied = strings.Contains(outputStr, "applied") ||
		strings.Contains(outputStr, "patched") ||
		strings.Contains(outputStr, "Patching")

	// Try to count patches (this is approximate)
	result.PatchCount = strings.Count(outputStr, "patching") +
		strings.Count(outputStr, "Patching")

	if options.Verbose {
		fmt.Printf("PatchPerl completed in %v\n", result.Duration)
		if result.Applied {
			fmt.Printf("Applied %d patches\n", result.PatchCount)
		} else {
			fmt.Println("No patches needed")
		}
	}

	return result, nil
}

// ApplyPatchPerlWithAutoInstall applies patches, installing Devel::PatchPerl if needed
func ApplyPatchPerlWithAutoInstall(options *PatchPerlOptions) (*PatchPerlResult, error) {
	// Check if PatchPerl is available
	if !IsPatchPerlAvailable() {
		if options.Verbose {
			fmt.Println("Devel::PatchPerl not found, installing...")
		}

		err := InstallPatchPerl(options.Verbose)
		if err != nil {
			return nil, fmt.Errorf("failed to install Devel::PatchPerl: %w", err)
		}

		// Verify installation worked
		if !IsPatchPerlAvailable() {
			return nil, fmt.Errorf("Devel::PatchPerl installation failed - not available after install")
		}

		if options.Verbose {
			fmt.Println("Devel::PatchPerl installed successfully")
		}
	}

	return ApplyPatchPerl(options)
}

// GetPatchPerlInfo returns information about the available PatchPerl installation
func GetPatchPerlInfo() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "perl", "-MDevel::PatchPerl", "-e",
		"print 'Version: ' . Devel::PatchPerl->VERSION . \"\\n\"")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get PatchPerl info: %w", err)
	}

	return string(output), nil
}

// ShouldUsePatchPerl determines if PatchPerl should be used for a given Perl version
func ShouldUsePatchPerl(perlVersion string) bool {
	// PatchPerl is beneficial for most Perl versions, especially older ones
	// Only skip for very new development versions or if explicitly disabled

	parsedVersion, err := ParseVersion(perlVersion)
	if err != nil {
		// If we can't parse version, err on the side of using PatchPerl
		return true
	}

	// Use PatchPerl for all stable versions
	// Skip only for development versions (odd minor numbers >= 5.37)
	if parsedVersion.Major == 5 && parsedVersion.Minor >= 37 && parsedVersion.Minor%2 == 1 {
		return false // Skip for development versions like 5.37.x, 5.39.x
	}

	return true
}

// ValidatePatchPerlEnvironment checks if the environment is suitable for PatchPerl
func ValidatePatchPerlEnvironment() error {
	// Check if perl is available
	if _, err := exec.LookPath("perl"); err != nil {
		return fmt.Errorf("perl not found in PATH - required for PatchPerl: %w", err)
	}

	// Check if we can execute perl
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "perl", "-e", "print 'OK'")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("perl execution failed - required for PatchPerl: %w", err)
	}

	return nil
}

// ApplyPatchPerlSafely is a convenience function that applies PatchPerl with error recovery
func ApplyPatchPerlSafely(sourceDir, perlVersion string, verbose bool) error {
	// Validate environment first
	if err := ValidatePatchPerlEnvironment(); err != nil {
		if verbose {
			fmt.Printf("Warning: PatchPerl environment validation failed: %v\n", err)
			fmt.Println("Skipping PatchPerl - build may fail on older Perl versions")
		}
		return nil // Non-fatal - let build attempt to continue
	}

	// Don't use PatchPerl if not beneficial
	if !ShouldUsePatchPerl(perlVersion) {
		if verbose {
			fmt.Printf("Skipping PatchPerl for Perl %s (not needed)\n", perlVersion)
		}
		return nil
	}

	options := &PatchPerlOptions{
		SourceDir:   sourceDir,
		PerlVersion: perlVersion,
		Verbose:     verbose,
		Context:     context.Background(),
		Timeout:     5 * time.Minute,
	}

	result, err := ApplyPatchPerlWithAutoInstall(options)
	if err != nil {
		if verbose {
			fmt.Printf("Warning: PatchPerl failed: %v\n", err)
			fmt.Println("Continuing build - may fail for older Perl versions")
		}
		// Return as warning, not fatal error
		return nil
	}

	if verbose && result.Applied {
		fmt.Printf("PatchPerl applied %d patches successfully\n", result.PatchCount)
	}

	return nil
}
