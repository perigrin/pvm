// ABOUTME: Unified build command integrating all build functionality
// ABOUTME: Provides comprehensive build orchestration for Perl projects

package pvm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/build"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/project"
)

// BuildCommand implements the unified build command
func NewBuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build [target]",
		Short: "Build Perl projects with type checking and compilation",
		Long: `Build Perl projects with integrated type checking, compilation, and metadata generation.

The build command supports multiple modes:
- Distribution build: Creates CPAN-ready packages with metadata
- Inline build: Generates .pmc files for development
- Type check only: Validates types without compilation
- Watch mode: Continuous builds on file changes

Examples:
  pvm build                    # Default distribution build
  pvm build --inline          # Development build (.pmc files)
  pvm build --check-only      # Type check without compilation
  pvm build --watch           # Continuous build
  pvm build --clean           # Clean build artifacts first`,
		Args: cobra.MaximumNArgs(1),
		RunE: executeBuildCommand,
	}

	// Build mode flags
	cmd.Flags().Bool("check-only", false, "Only run type checking, don't compile")
	cmd.Flags().Bool("inline", false, "Generate .pmc files for development")
	cmd.Flags().Bool("watch", false, "Watch for changes and rebuild")
	cmd.Flags().Bool("clean", false, "Clean build artifacts before building")

	// Output and configuration flags
	cmd.Flags().String("output", "", "Output directory for build artifacts")
	cmd.Flags().String("mode", "", "Build mode (distribution, inline, both)")

	// Type checking flags
	cmd.Flags().Bool("strict", false, "Enable strict type checking")
	cmd.Flags().Bool("skip-typecheck", false, "Skip type checking")

	// Distribution flags
	cmd.Flags().Bool("skip-metadata", false, "Skip metadata generation")
	cmd.Flags().Bool("include-tests", true, "Include tests in distribution")
	cmd.Flags().Bool("include-scripts", true, "Include scripts in distribution")

	return cmd
}

// executeBuildCommand orchestrates the build process
func executeBuildCommand(cmd *cobra.Command, args []string) error {
	// Get current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Detect project context
	projectCtx, err := project.DetectProject(workingDir)
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}

	// Parse command line flags and merge with configuration
	buildOptions, err := parseBuildOptions(cmd, projectCtx)
	if err != nil {
		return fmt.Errorf("failed to parse build options: %w", err)
	}

	// Create build context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Execute the appropriate build strategy
	switch {
	case buildOptions.Watch:
		return executeWatchBuild(ctx, cmd, projectCtx, buildOptions)
	case buildOptions.CheckOnly:
		return executeTypeCheckBuild(ctx, cmd, projectCtx, buildOptions)
	case buildOptions.Inline:
		return executeInlineBuild(ctx, cmd, projectCtx, buildOptions)
	default:
		return executeDistributionBuild(ctx, cmd, projectCtx, buildOptions)
	}
}

// BuildOptions aggregates all build configuration
type BuildOptions struct {
	// Build modes
	CheckOnly bool
	Inline    bool
	Watch     bool
	Clean     bool

	// Configuration
	Mode      string
	OutputDir string

	// Type checking
	Strict        bool
	SkipTypeCheck bool

	// Distribution
	SkipMetadata   bool
	IncludeTests   bool
	IncludeScripts bool

	// Derived from project context
	ProjectRoot string
	SourceDirs  []string
}

// parseBuildOptions extracts build options from command flags and configuration
func parseBuildOptions(cmd *cobra.Command, projectCtx *project.ProjectContext) (*BuildOptions, error) {
	options := &BuildOptions{}

	// Parse command line flags
	var err error
	options.CheckOnly, err = cmd.Flags().GetBool("check-only")
	if err != nil {
		return nil, err
	}

	options.Inline, err = cmd.Flags().GetBool("inline")
	if err != nil {
		return nil, err
	}

	options.Watch, err = cmd.Flags().GetBool("watch")
	if err != nil {
		return nil, err
	}

	options.Clean, err = cmd.Flags().GetBool("clean")
	if err != nil {
		return nil, err
	}

	options.OutputDir, err = cmd.Flags().GetString("output")
	if err != nil {
		return nil, err
	}

	options.Mode, err = cmd.Flags().GetString("mode")
	if err != nil {
		return nil, err
	}

	options.Strict, err = cmd.Flags().GetBool("strict")
	if err != nil {
		return nil, err
	}

	options.SkipTypeCheck, err = cmd.Flags().GetBool("skip-typecheck")
	if err != nil {
		return nil, err
	}

	options.SkipMetadata, err = cmd.Flags().GetBool("skip-metadata")
	if err != nil {
		return nil, err
	}

	options.IncludeTests, err = cmd.Flags().GetBool("include-tests")
	if err != nil {
		return nil, err
	}

	options.IncludeScripts, err = cmd.Flags().GetBool("include-scripts")
	if err != nil {
		return nil, err
	}

	// Load build configuration and merge with flags
	workingDir := projectCtx.RootDir
	if workingDir == "" {
		workingDir, _ = os.Getwd()
	}

	buildConfig, err := config.GetProjectBuildConfig(workingDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load build configuration: %w", err)
	}

	// Merge configuration with command line flags (flags take precedence)
	if options.Mode == "" {
		options.Mode = buildConfig.Mode
	}

	if options.OutputDir == "" {
		options.OutputDir = buildConfig.OutputDir
	}

	// Set project context information
	options.ProjectRoot = projectCtx.RootDir
	if projectCtx.IsProject {
		options.SourceDirs = []string{
			filepath.Join(projectCtx.RootDir, "lib"),
			filepath.Join(projectCtx.RootDir, "script"),
			filepath.Join(projectCtx.RootDir, "t"),
		}
	} else {
		options.SourceDirs = []string{"lib", "script", "t"}
	}

	return options, nil
}

// executeTypeCheckBuild performs type checking only
func executeTypeCheckBuild(ctx context.Context, cmd *cobra.Command, projectCtx *project.ProjectContext, options *BuildOptions) error {
	ui := cli.GetUI(cmd)
	ui.Status("Type checking Perl project...")

	// Create PSC builder
	pscBuilder, err := build.NewPSCBuilder(projectCtx)
	if err != nil {
		return fmt.Errorf("failed to create type checker: %w", err)
	}

	// Perform type checking
	start := time.Now()
	result, err := pscBuilder.TypeCheck(ctx, options.SourceDirs)
	if err != nil {
		return fmt.Errorf("type checking failed: %w", err)
	}

	duration := time.Since(start)

	// Report results
	ui.Info("Type checked %d files in %s", result.FilesChecked, duration.Round(time.Millisecond))

	if len(result.TypeErrors) > 0 {
		ui.Error("Found %d type errors:", len(result.TypeErrors))
		for _, typeErr := range result.TypeErrors {
			ui.Error("  %s:%d:%d - %s", typeErr.File, typeErr.Line, typeErr.Column, typeErr.Message)
		}
		return fmt.Errorf("type checking failed with %d errors", len(result.TypeErrors))
	}

	ui.Success("Type checking passed!")
	return nil
}

// executeInlineBuild performs development build with .pmc generation
func executeInlineBuild(ctx context.Context, cmd *cobra.Command, projectCtx *project.ProjectContext, options *BuildOptions) error {
	ui := cli.GetUI(cmd)
	ui.Status("Building inline development version...")

	// Create inline builder
	inlineBuilder, err := build.NewInlineBuilder(projectCtx)
	if err != nil {
		return fmt.Errorf("failed to create inline builder: %w", err)
	}

	// Clean if requested
	if options.Clean {
		ui.Status("Cleaning existing .pmc files...")
		targetDirs := []string{filepath.Join(options.ProjectRoot, "lib")}
		if err := inlineBuilder.Clean(ctx, targetDirs); err != nil {
			ui.Warning("Failed to clean .pmc files: %v", err)
		}
	}

	// Perform inline build
	start := time.Now()
	targetDirs := []string{filepath.Join(options.ProjectRoot, "lib")}
	result, err := inlineBuilder.Build(ctx, targetDirs)
	if err != nil {
		return fmt.Errorf("inline build failed: %w", err)
	}

	duration := time.Since(start)

	// Report results
	if !result.Success {
		ui.Error("Inline build failed after %s", duration.Round(time.Millisecond))
		if len(result.TypeErrors) > 0 {
			ui.Error("Type errors found:")
			for _, typeErr := range result.TypeErrors {
				ui.Error("  %s:%d:%d - %s", typeErr.File, typeErr.Line, typeErr.Column, typeErr.Message)
			}
		}
		if len(result.BuildErrors) > 0 {
			ui.Error("Build errors:")
			for _, buildErr := range result.BuildErrors {
				ui.Error("  %s", buildErr)
			}
		}
		return fmt.Errorf("inline build failed")
	}

	ui.Success("Inline build completed in %s", duration.Round(time.Millisecond))
	ui.Info("Processed %d files, generated %d .pmc files", result.FilesProcessed, result.PmcGenerated)
	ui.Success("Development build ready!")
	return nil
}

// executeDistributionBuild performs full distribution build
func executeDistributionBuild(ctx context.Context, cmd *cobra.Command, projectCtx *project.ProjectContext, options *BuildOptions) error {
	ui := cli.GetUI(cmd)
	ui.Status("Building CPAN distribution...")

	// Create distribution builder
	distBuilder, err := build.NewDistributionBuilder(projectCtx)
	if err != nil {
		return fmt.Errorf("failed to create distribution builder: %w", err)
	}

	// Clean if requested
	if options.Clean {
		ui.Status("Cleaning build directory...")
		if err := distBuilder.Clean(options.OutputDir); err != nil {
			ui.Warning("Failed to clean build directory: %v", err)
		}
	}

	// Prepare distribution build options
	buildOpts := &build.BuildOptions{
		OutputDir:       options.OutputDir,
		SourceDirs:      options.SourceDirs,
		CleanFirst:      options.Clean,
		SkipTypeCheck:   options.SkipTypeCheck,
		StrictTypeCheck: options.Strict,
		SkipMetadata:    options.SkipMetadata,
		IncludeTests:    options.IncludeTests,
		IncludeScripts:  options.IncludeScripts,
	}

	// Perform distribution build
	start := time.Now()
	result, err := distBuilder.Build(ctx, buildOpts)
	if err != nil {
		return fmt.Errorf("distribution build failed: %w", err)
	}

	duration := time.Since(start)

	// Report results
	if !result.Success {
		ui.Error("Distribution build failed after %s", duration.Round(time.Millisecond))
		if len(result.TypeErrors) > 0 {
			ui.Error("Type errors found:")
			for _, typeErr := range result.TypeErrors {
				ui.Error("  %s:%d:%d - %s", typeErr.File, typeErr.Line, typeErr.Column, typeErr.Message)
			}
		}
		if len(result.BuildErrors) > 0 {
			ui.Error("Build errors:")
			for _, buildErr := range result.BuildErrors {
				ui.Error("  %s", buildErr)
			}
		}
		return fmt.Errorf("distribution build failed")
	}

	ui.Success("Distribution build completed in %s", duration.Round(time.Millisecond))
	ui.Info("Processed %d files", result.FilesProcessed)
	if result.MetadataGenerated {
		ui.Success("Generated CPAN metadata files")
	}
	ui.Success("Distribution ready: %s", result.BuildDir)
	if result.DistributionName != "" {
		ui.Info("Distribution name: %s", result.DistributionName)
	}
	return nil
}

// executeWatchBuild performs continuous build with file watching
func executeWatchBuild(ctx context.Context, cmd *cobra.Command, projectCtx *project.ProjectContext, options *BuildOptions) error {
	ui := cli.GetUI(cmd)
	ui.Status("Starting watch mode...")

	// Create watcher configuration
	watchConfig := build.DefaultWatcherConfig()

	// Configure watch directories from build config
	workingDir := projectCtx.RootDir
	if workingDir == "" {
		workingDir, _ = os.Getwd()
	}

	watchDirs, err := config.GetBuildWatchDirs(workingDir)
	if err == nil {
		watchConfig.WatchDirs = watchDirs
	}

	// Configure based on build mode
	switch options.Mode {
	case "inline":
		watchConfig.EnableTypeCheck = true
		watchConfig.EnableInline = true
		watchConfig.EnableDist = false
	case "distribution":
		watchConfig.EnableTypeCheck = true
		watchConfig.EnableInline = false
		watchConfig.EnableDist = true
	case "both":
		watchConfig.EnableTypeCheck = true
		watchConfig.EnableInline = true
		watchConfig.EnableDist = true
	default:
		// Default based on inline flag
		if options.Inline {
			watchConfig.EnableInline = true
			watchConfig.EnableDist = false
		}
	}

	// Create and start watcher
	watcher, err := build.NewBuildWatcher(projectCtx, watchConfig)
	if err != nil {
		return fmt.Errorf("failed to create build watcher: %w", err)
	}

	if err := watcher.Start(); err != nil {
		return fmt.Errorf("failed to start build watcher: %w", err)
	}
	defer watcher.Stop()

	ui.Info("Watching directories: %v", watchConfig.WatchDirs)
	ui.Info("Press Ctrl+C to stop watching...")

	// Monitor build results
	for {
		select {
		case <-ctx.Done():
			return nil
		case result := <-watcher.Results():
			if result.Success {
				ui.Success("[%s] %s build completed (%s) - %d files",
					result.Timestamp.Format("15:04:05"),
					result.Type.String(),
					result.Duration.Round(time.Millisecond),
					len(result.Files))
			} else {
				ui.Error("[%s] %s build failed (%s): %v",
					result.Timestamp.Format("15:04:05"),
					result.Type.String(),
					result.Duration.Round(time.Millisecond),
					result.Error)
			}
		}
	}
}
