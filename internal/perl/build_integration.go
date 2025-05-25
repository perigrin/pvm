// ABOUTME: Integration of enhanced build system with existing PVM commands
// ABOUTME: Provides backward compatibility while using new features

package perl

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
)

// BuildPerlEnhanced is the main entry point for the enhanced build system
// It provides a compatible interface with the existing BuildPerl function
func BuildPerlEnhanced(options *BuildOptions) (*BuildResult, error) {
	// Create logger
	logger := log.NewLogger(log.LevelInfo, os.Stderr, "build")

	// Create build manager
	bm, err := NewBuildManager(logger)
	if err != nil {
		return nil, err
	}

	// Use enhanced build
	return bm.BuildEnhanced(options)
}

// EnableEnhancedBuild enables the enhanced build system globally
var EnableEnhancedBuild = false

// BuildPerlWithFallback attempts enhanced build with fallback to original
func BuildPerlWithFallback(options *BuildOptions) (*BuildResult, error) {
	if EnableEnhancedBuild {
		// Try enhanced build first
		result, err := BuildPerlEnhanced(options)
		if err == nil {
			return result, nil
		}

		// Log the error but continue with fallback
		log.Warningf("Enhanced build failed, falling back to standard build: %v", err)
	}

	// Use standard build
	return BuildPerl(options)
}

// MigrateToEnhancedBuild migrates existing build configurations to enhanced format
func MigrateToEnhancedBuild(options *BuildOptions) *BuildOptions {
	if options == nil {
		options = &BuildOptions{}
	}

	// Set enhanced defaults if not specified
	if options.MaxRetries == 0 {
		options.MaxRetries = 3
	}

	if options.TestTimeout == 0 {
		options.TestTimeout = 30 * time.Minute
	}

	// Enable verbose mode for better debugging
	if options.ProgressCallback != nil && !options.Verbose {
		options.Verbose = true
	}

	return options
}

// BuildReport generates a detailed build report
type BuildReport struct {
	Result          *BuildResult
	SystemInfo      SystemInfo
	Dependencies    *DependencyInfo
	ConfigureOutput string
	TestResults     TestSummary
	Warnings        []string
	Errors          []error
}

// TestSummary contains test execution summary
type TestSummary struct {
	TotalTests   int
	PassedTests  int
	FailedTests  int
	SkippedTests int
	Duration     time.Duration
}

// GenerateBuildReport creates a detailed report of the build process
func GenerateBuildReport(ctx *BuildContext, result *BuildResult) *BuildReport {
	report := &BuildReport{
		Result:       result,
		SystemInfo:   getSystemInfo(),
		Dependencies: ctx.SystemDeps,
		Warnings:     []string{},
		Errors:       []error{},
	}

	// Add any warnings collected during build
	if ctx.SystemDeps != nil && len(ctx.SystemDeps.Optional) > 0 {
		report.Warnings = append(report.Warnings,
			fmt.Sprintf("Optional dependencies missing: %v", ctx.SystemDeps.Optional))
	}

	return report
}

// getSystemInfo collects system information (reuse from build_cache.go)
func getSystemInfo() SystemInfo {
	info := SystemInfo{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		CPUCount: runtime.NumCPU(),
	}

	if compiler, version := detectCompiler(); compiler != "" {
		info.Compiler = compiler
		info.CompilerVer = version
	}

	return info
}

// CheckBuildRequirements performs pre-build checks
func CheckBuildRequirements(version string) error {
	// Check version availability
	// First get available versions
	availableVersions, err := GetAvailableVersions()
	if err != nil {
		return errors.NewVersionError(
			ErrInvalidBuildOptions,
			"Failed to get available versions",
			err,
		)
	}

	available := IsVersionAvailable(version, availableVersions)

	if !available {
		return errors.NewVersionError(
			ErrInvalidBuildOptions,
			fmt.Sprintf("Version %s is not available for download", version),
			nil,
		)
	}

	// Check system dependencies
	depChecker, err := NewDependencyChecker()
	if err != nil {
		return err
	}

	deps, err := depChecker.CheckBuildDependencies()
	if err != nil {
		return err
	}

	if len(deps.Missing) > 0 {
		depErr := errors.NewSystemError(
			ErrBuildDirFailed,
			fmt.Sprintf("Missing required dependencies: %s", strings.Join(deps.Missing, ", ")),
			nil,
		)
		depErr.WithDetail(fmt.Sprintf("Missing dependencies: %v", deps.Missing))
		if hint, ok := deps.InstallHint[runtime.GOOS]; ok {
			depErr.WithHint(hint)
		}
		return depErr
	}

	return nil
}

// OptimizeBuildOptions optimizes build options based on system capabilities
func OptimizeBuildOptions(options *BuildOptions) *BuildOptions {
	if options == nil {
		options = &BuildOptions{}
	}

	// Optimize parallel jobs based on system
	if options.BuildJobs == 0 {
		cpus := runtime.NumCPU()

		// Consider memory constraints
		if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
			// Apple Silicon can be memory constrained
			if cpus > 4 {
				options.BuildJobs = 4
			} else {
				options.BuildJobs = cpus
			}
		} else {
			options.BuildJobs = cpus
		}
	}

	// Enable cleanup by default for CI environments
	if os.Getenv("CI") != "" && !options.CleanupBuildDir {
		options.CleanupBuildDir = true
	}

	// Set reasonable test timeout
	if options.TestTimeout == 0 {
		if os.Getenv("CI") != "" {
			options.TestTimeout = 60 * time.Minute // More time in CI
		} else {
			options.TestTimeout = 30 * time.Minute
		}
	}

	return options
}

// CompareBuilds compares two build results for debugging
func CompareBuilds(standard, enhanced *BuildResult) map[string]interface{} {
	comparison := map[string]interface{}{
		"version_match":        standard.Version == enhanced.Version,
		"install_path_match":   standard.InstallPath == enhanced.InstallPath,
		"duration_diff":        enhanced.Duration - standard.Duration,
		"duration_improvement": 100.0 * (standard.Duration - enhanced.Duration) / standard.Duration,
	}

	// Compare stage timings if available
	if len(standard.Stages) > 0 && len(enhanced.Stages) > 0 {
		stageDiffs := make(map[string]time.Duration)
		for stage, stdTime := range standard.Stages {
			if enhTime, ok := enhanced.Stages[stage]; ok {
				stageDiffs[stage.String()] = enhTime - stdTime
			}
		}
		comparison["stage_diffs"] = stageDiffs
	}

	return comparison
}
