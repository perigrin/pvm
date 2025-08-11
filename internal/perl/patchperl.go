// ABOUTME: Devel::PatchPerl integration for cross-platform Perl source patching
// ABOUTME: Provides automatic patching of Perl source code for compatibility across platforms

package perl

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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

// ModuleInstaller is an interface for installing Perl modules
// This allows us to inject the PM functionality without creating import cycles
type ModuleInstaller interface {
	InstallModule(moduleName string, verbose bool) error
}

// Default module installer that provides user guidance
type DefaultModuleInstaller struct{}

func (d *DefaultModuleInstaller) InstallModule(moduleName string, verbose bool) error {
	if verbose {
		fmt.Printf("%s not found - please install it using PM\n", moduleName)
	}

	return fmt.Errorf(`%s is required but not installed.

Please install it using PVM's PM command:
  pm install %s

Or if you prefer to skip patching, use the --no-patchperl flag:
  pvm install --no-patchperl <version>`, moduleName, moduleName)
}

// moduleInstaller is the current module installer implementation
// It can be set by other packages to provide actual installation functionality
var moduleInstaller ModuleInstaller = &DefaultModuleInstaller{}

// SetModuleInstaller allows other packages to provide module installation functionality
func SetModuleInstaller(installer ModuleInstaller) {
	moduleInstaller = installer
}

// InstallPatchPerl installs Devel::PatchPerl using the configured module installer
func InstallPatchPerl(verbose bool) error {
	return moduleInstaller.InstallModule("Devel::PatchPerl", verbose)
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
