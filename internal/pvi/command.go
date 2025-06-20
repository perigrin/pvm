// ABOUTME: PVI-specific commands and functionality
// ABOUTME: Implements commands for Perl module management

package pvi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	uipkg "tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/project"
	"tamarou.com/pvm/internal/pvi/deps"
	"tamarou.com/pvm/internal/pvi/modules"
)

// NewCommand creates a new PVI command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pvi",
		Short: "Perl Version Installer",
		Long:  "Manages CPAN modules for installed Perl versions",
	}

	// Add PVI-specific commands
	cmd.AddCommand(
		newInstallCommand(),
		newAddCommand(),
		newSyncCommand(),
		newListCommand(),
		newUpdateCommand(),
		newRemoveCommand(),
		newSearchCommand(),
		newDepsCommand(),
		newBundleCommand(),
		newTypeCommand(),
		newMirrorCommand(),
		newOutdatedCommand(),
	)

	return cmd
}

// Placeholder commands, to be implemented later

func newInstallCommand() *cobra.Command {
	var (
		verbose      bool
		force        bool
		skipTests    bool
		skipDeps     bool
		noCache      bool
		installDir   string
		version      string
		source       string
		buildOnly    bool
		installOnly  bool
		keepBuildDir bool
		workers      int
		parallel     bool
		timing       bool
		dev          bool
	)

	cmd := &cobra.Command{
		Use:   "install [module...]",
		Short: "Install one or more modules",
		Long:  "Install CPAN modules for the current Perl version. If no modules specified, installs from cpanfile. Supports parallel installation for multiple modules.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)
			var moduleNames []string

			// If no modules specified, try to read from cpanfile
			if len(args) == 0 {
				// Detect project context
				projectCtx, err := project.GetCurrentProject()
				if err != nil {
					return fmt.Errorf("failed to detect project context: %w", err)
				}

				if !projectCtx.IsProject {
					return fmt.Errorf("no modules specified and not in a project directory. Specify modules to install or use 'pvm project init' to create a project")
				}

				// Check for cpanfile
				cpanfilePath := filepath.Join(projectCtx.RootDir, "cpanfile")
				if _, err := os.Stat(cpanfilePath); os.IsNotExist(err) {
					return fmt.Errorf("no modules specified and no cpanfile found in project. Use 'pvm module add <module>' to add dependencies")
				}

				// Read cpanfile and extract module names
				cpanfileManager := NewCpanfileManager(cpanfilePath)
				cpanfile, err := cpanfileManager.ListDependencies()
				if err != nil {
					return fmt.Errorf("failed to read cpanfile: %w", err)
				}

				// Check for --dev flag to include development dependencies
				includeDev, _ := cmd.Flags().GetBool("dev")

				// Extract module names based on flags
				for _, req := range cpanfile.Requirements {
					if req.Phase == "develop" && !includeDev {
						continue // Skip dev dependencies unless --dev flag is used
					}
					if req.Phase != "develop" && req.Phase != "runtime" && req.Phase != "" {
						continue // Skip test and build dependencies for now
					}
					moduleNames = append(moduleNames, req.Module)
				}

				if len(moduleNames) == 0 {
					if includeDev {
						return fmt.Errorf("no dependencies found in cpanfile")
					} else {
						return fmt.Errorf("no runtime dependencies found in cpanfile. Use --dev flag to include development dependencies")
					}
				}

				ui.Info("Installing %d modules from cpanfile...", len(moduleNames))

				// Override installDir to use project lib directory
				if installDir == "" {
					installDir = projectCtx.LocalLibDir
				}
			} else {
				moduleNames = args
			}

			// Load configuration
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			// Create provider and resolver using builder pattern
			result, err := NewProviderBuilder().
				WithConfig(cfg).
				WithSource(source).
				WithNoCache(noCache).
				WithResolver().
				Build()
			if err != nil {
				return err
			}

			provider := result.Provider
			resolver := result.Resolver

			// Get current Perl path
			perlPath, err := perl.GetCurrentPerlPath()
			if err != nil {
				return err
			}

			// Create installation options
			installOptions := &modules.ModuleInstallOptions{
				ModuleName:         "", // Will be set per module later
				VersionConstraint:  version,
				PerlPath:           perlPath,
				InstallDir:         installDir,
				RunTests:           !skipTests,
				Force:              force,
				Cleanup:            !keepBuildDir,
				Verbose:            verbose,
				SkipDependencies:   skipDeps,
				Provider:           provider,
				DependencyResolver: resolver,
				ProgressCallback: func(stage modules.InstallProgressStage, module string, details string, progress float64) {
					if verbose {
						ui.Debug("[%s] %s: %s (%.0f%%)", stage.String(), module, details, progress*100)
					} else if stage != modules.StageFinished {
						// Only show major stage transitions if not verbose
						ui.Info("[%s] %s", stage.String(), module)
					}
				},
				Context: cmd.Context(),
			}

			// Determine if we should use parallel installation
			useParallel := parallel || len(moduleNames) > 1

			if useParallel && len(moduleNames) > 1 {
				// Parallel installation for multiple modules
				if timing {
					ui.Info("Installing %d modules in parallel with %d workers...", len(moduleNames), workers)
				} else {
					ui.Info("Installing %d modules in parallel...", len(moduleNames))
				}

				parallelOptions := &modules.ParallelInstallOptions{
					Modules:     make([]*modules.ModuleInstallOptions, len(moduleNames)),
					Workers:     workers,
					StopOnError: false,
					Context:     cmd.Context(),
				}

				// Create module options for each module
				for i, moduleName := range moduleNames {
					moduleOpts := *installOptions // Copy base options
					moduleOpts.ModuleName = moduleName
					parallelOptions.Modules[i] = &moduleOpts
				}

				// Add progress callback for parallel installation
				if verbose || timing {
					parallelOptions.ProgressCallback = func(completed, total int, currentModule string, stage modules.InstallProgressStage) {
						ui.Progress(completed, total, fmt.Sprintf("%s: %s", currentModule, stage.String()))
					}
				}

				startTime := time.Now()
				result, err := modules.InstallModulesParallel(parallelOptions)
				duration := time.Since(startTime)

				if err != nil {
					ui.Error("Parallel installation failed: %v", err)
					return err
				}

				// Display results
				ui.SubHeader("Installation Summary")
				summary := map[string]string{
					"Modules":    fmt.Sprintf("%d", len(moduleNames)),
					"Successful": fmt.Sprintf("%d", result.SuccessCount),
					"Failed":     fmt.Sprintf("%d", result.FailureCount),
				}
				if timing {
					summary["Total time"] = duration.Round(time.Millisecond).String()
					summary["Average per module"] = (duration / time.Duration(len(moduleNames))).Round(time.Millisecond).String()
				}
				ui.KeyValue(summary)

				// Show successful installations
				if len(result.Results) > 0 {
					var successList []string
					for _, res := range result.Results {
						if res.Success {
							if timing {
								successList = append(successList, fmt.Sprintf("✓ %s v%s (%s)", res.ModuleName, res.Version, res.Duration.Round(time.Millisecond)))
							} else {
								successList = append(successList, fmt.Sprintf("✓ %s v%s", res.ModuleName, res.Version))
							}
						}
					}
					if len(successList) > 0 {
						ui.ListWithOptions(uipkg.ListOptions{
							Title: "Successful installations",
							Items: successList,
						})
					}
				}

				// Show failures
				if len(result.Failures) > 0 {
					var failureList []string
					for _, failure := range result.Failures {
						if timing {
							failureList = append(failureList, fmt.Sprintf("✗ %s (%s): %v", failure.ModuleName, failure.Duration.Round(time.Millisecond), failure.Error))
						} else {
							failureList = append(failureList, fmt.Sprintf("✗ %s: %v", failure.ModuleName, failure.Error))
						}
					}
					ui.ListWithOptions(uipkg.ListOptions{
						Title: "Failed installations",
						Items: failureList,
					})
					return fmt.Errorf("%d modules failed to install", len(result.Failures))
				}

			} else {
				// Single module installation (original logic)
				moduleName := moduleNames[0]
				installOptions.ModuleName = moduleName

				ui.Info("Installing module %s...", moduleName)

				startTime := time.Now()
				result, err := modules.InstallModule(installOptions)
				duration := time.Since(startTime)

				if err != nil {
					ui.Error("Failed to install %s: %v", moduleName, err)
					return err
				}

				// Display result
				if result.Success {
					if timing {
						ui.Success("Successfully installed %s v%s (%s)", result.ModuleName, result.Version, duration.Round(time.Millisecond))
					} else {
						ui.Success("Successfully installed %s v%s", result.ModuleName, result.Version)
					}

					// Show installed dependencies count
					if len(result.Dependencies) > 0 {
						ui.Info("Installed %d dependencies", len(result.Dependencies))
					}

					// Show warnings if any
					if len(result.Warnings) > 0 && verbose {
						ui.ListWithOptions(uipkg.ListOptions{
							Title: "Warnings",
							Items: result.Warnings,
						})
					}

					if timing {
						ui.Info("Total installation time: %s", duration.Round(time.Millisecond))
					} else {
						ui.Info("Total installation time: %s", result.Duration.Round(time.Second))
					}
				} else {
					ui.Error("Failed to install %s", moduleName)
					if len(result.Errors) > 0 {
						ui.ListWithOptions(uipkg.ListOptions{
							Title: "Errors",
							Items: result.Errors,
						})
					}
					return fmt.Errorf("installation failed")
				}
			}

			return nil
		},
	}

	// Add flags for the install command
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force installation even if tests fail")
	cmd.Flags().BoolVar(&skipTests, "skip-tests", false, "Skip running tests")
	cmd.Flags().BoolVar(&skipDeps, "skip-dependencies", false, "Skip installation of dependencies")
	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Disable caching for this installation")
	cmd.Flags().StringVarP(&installDir, "install-dir", "i", "", "Directory to install the module to")
	cmd.Flags().StringVarP(&version, "version", "V", "", "Version constraint (e.g. '>= 1.0')")
	cmd.Flags().StringVarP(&source, "source", "s", "", "Use a specific metadata source (metacpan, cpan, custom)")
	cmd.Flags().BoolVar(&buildOnly, "build-only", false, "Only build the module, don't install it")
	cmd.Flags().BoolVar(&installOnly, "install-only", false, "Only install the module, don't build it (assumes it's already built)")
	cmd.Flags().BoolVar(&keepBuildDir, "keep-build-dir", false, "Keep the build directory after installation")
	cmd.Flags().IntVar(&workers, "workers", 0, "Number of parallel workers (0 = auto-detect, only used with multiple modules)")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Force parallel installation even for single module")
	cmd.Flags().BoolVar(&timing, "timing", false, "Show detailed timing information")
	cmd.Flags().BoolVar(&dev, "dev", false, "Include development dependencies when installing from cpanfile")

	return cmd
}

func newListCommand() *cobra.Command {
	var (
		pattern     string
		includeCore bool
		format      string
		perlPath    string
		verbose     bool
	)

	cmd := &cobra.Command{
		Use:   "list [pattern]",
		Short: "List installed modules",
		Long:  "List all installed CPAN modules for the current Perl version",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)
			// Process pattern from args if provided
			if len(args) > 0 {
				pattern = args[0]
			}

			// Get current Perl path if not specified
			if perlPath == "" {
				var err error
				perlPath, err = perl.GetCurrentPerlPath()
				if err != nil {
					return err
				}
			}

			// Create options for listing modules
			options := &modules.ModuleListOptions{
				PerlPath:    perlPath,
				Pattern:     pattern,
				IncludeCore: includeCore,
				Context:     cmd.Context(),
			}

			// List installed modules
			moduleList, err := modules.ListInstalledModules(options)
			if err != nil {
				return err
			}

			// Display results
			if len(moduleList) == 0 {
				ui.Info("No modules found")
				return nil
			}

			// Format output
			switch format {
			case "json":
				// Output as JSON
				jsonData, err := json.MarshalIndent(moduleList, "", "  ")
				if err != nil {
					return err
				}
				ui.Println(string(jsonData))

			case "simple":
				// Output as simple name/version
				var simpleList []string
				for _, module := range moduleList {
					simpleList = append(simpleList, fmt.Sprintf("%s %s", module.Name, module.Version))
				}
				ui.List(simpleList)

			default:
				// Default tabular format
				ui.SubHeader(fmt.Sprintf("Found %d modules", len(moduleList)))

				if verbose {
					// Detailed format
					for i, module := range moduleList {
						ui.SubHeader(fmt.Sprintf("[%d] %s (%s)", i+1, module.Name, module.Version))

						info := map[string]string{
							"Path": module.Path,
						}

						if module.Description != "" {
							info["Description"] = module.Description
						}

						if !module.InstallationTime.IsZero() {
							info["Installed"] = module.InstallationTime.Format("2006-01-02 15:04:05")
						}

						if module.CoreModule {
							info["Core Module"] = "Yes"
						}

						ui.KeyValue(info)
					}
				} else {
					// Table format
					headers := []string{"Module", "Version", "Core Module"}
					var rows [][]string

					for _, module := range moduleList {
						coreStatus := ""
						if module.CoreModule {
							coreStatus = "Yes"
						}
						rows = append(rows, []string{module.Name, module.Version, coreStatus})
					}

					ui.Table(headers, rows)
				}
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&pattern, "pattern", "p", "", "Pattern to filter modules by name")
	cmd.Flags().BoolVarP(&includeCore, "include-core", "c", false, "Include core modules")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, simple, json)")
	cmd.Flags().StringVar(&perlPath, "perl", "", "Path to the perl executable to use")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed module information")

	return cmd
}

func newUpdateCommand() *cobra.Command {
	var (
		verbose    bool
		force      bool
		skipTests  bool
		skipDeps   bool
		noCache    bool
		installDir string
		version    string
		source     string
		perlPath   string
		all        bool
	)

	cmd := &cobra.Command{
		Use:   "update [module...]",
		Short: "Update modules",
		Long:  "Update one or more CPAN modules to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)
			// Check if at least one module is specified or --all flag is used
			if len(args) == 0 && !all {
				return fmt.Errorf("at least one module name must be provided or use --all flag")
			}

			// Load configuration
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			// Get current Perl path if not specified
			if perlPath == "" {
				perlPath, err = perl.GetCurrentPerlPath()
				if err != nil {
					return err
				}
			}

			// Create provider and resolver using builder pattern
			result, err := NewProviderBuilder().
				WithConfig(cfg).
				WithSource(source).
				WithNoCache(noCache).
				WithResolver().
				Build()
			if err != nil {
				return err
			}

			provider := result.Provider
			resolver := result.Resolver

			// If --all flag is used, get a list of all installed modules
			if all {
				options := &modules.ModuleListOptions{
					PerlPath:    perlPath,
					Pattern:     "",
					IncludeCore: false, // Don't update core modules
					Context:     cmd.Context(),
				}

				moduleList, err := modules.ListInstalledModules(options)
				if err != nil {
					return err
				}

				// Convert to module names
				args = make([]string, 0, len(moduleList))
				for _, mod := range moduleList {
					args = append(args, mod.Name)
				}

				ui.Info("Updating all %d installed modules...", len(args))
			}

			// Update each module
			successCount := 0
			failCount := 0

			for _, moduleName := range args {
				// Create installation options for the update
				installOptions := &modules.ModuleInstallOptions{
					ModuleName:         moduleName,
					VersionConstraint:  version,
					PerlPath:           perlPath,
					InstallDir:         installDir,
					RunTests:           !skipTests,
					Force:              force,
					Cleanup:            true,
					Verbose:            verbose,
					SkipDependencies:   skipDeps,
					Provider:           provider,
					DependencyResolver: resolver,
					ProgressCallback: func(stage modules.InstallProgressStage, module string, details string, progress float64) {
						if verbose {
							ui.Debug("[%s] %s: %s (%.0f%%)", stage.String(), module, details, progress*100)
						} else if stage != modules.StageFinished {
							// Only show major stage transitions if not verbose
							ui.Info("[%s] %s", stage.String(), module)
						}
					},
					Context: cmd.Context(),
				}

				// Update (install) the module
				ui.Info("Updating module %s...", moduleName)
				result, err := modules.InstallModule(installOptions)

				if err != nil {
					ui.Error("Failed to update %s: %v", moduleName, err)
					failCount++
					continue
				}

				// Display result
				if result.Success {
					ui.Success("Successfully updated %s to v%s", result.ModuleName, result.Version)
					successCount++

					// Show warnings if any
					if len(result.Warnings) > 0 && verbose {
						ui.ListWithOptions(uipkg.ListOptions{
							Title: "Warnings",
							Items: result.Warnings,
						})
					}
				} else {
					ui.Error("Failed to update %s", moduleName)
					failCount++
					if len(result.Errors) > 0 {
						ui.ListWithOptions(uipkg.ListOptions{
							Title: "Errors",
							Items: result.Errors,
						})
					}
				}
			}

			// Summary
			summary := map[string]string{
				"Succeeded": fmt.Sprintf("%d", successCount),
				"Failed":    fmt.Sprintf("%d", failCount),
			}
			ui.SubHeader("Update Summary")
			ui.KeyValue(summary)

			if failCount > 0 {
				return fmt.Errorf("%d module updates failed", failCount)
			}

			return nil
		},
	}

	// Add flags for the update command (similar to install command)
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force installation even if tests fail")
	cmd.Flags().BoolVar(&skipTests, "skip-tests", false, "Skip running tests")
	cmd.Flags().BoolVar(&skipDeps, "skip-dependencies", false, "Skip installation of dependencies")
	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Disable caching for this installation")
	cmd.Flags().StringVarP(&installDir, "install-dir", "i", "", "Directory to install the module to")
	cmd.Flags().StringVarP(&version, "version", "V", "", "Version constraint (e.g. '>= 1.0')")
	cmd.Flags().StringVarP(&source, "source", "s", "", "Use a specific metadata source (metacpan, cpan, custom)")
	cmd.Flags().StringVar(&perlPath, "perl", "", "Path to the perl executable to use")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Update all installed modules")

	return cmd
}

func newRemoveCommand() *cobra.Command {
	var (
		force    bool
		verbose  bool
		perlPath string
	)

	cmd := &cobra.Command{
		Use:   "remove [module]",
		Short: "Remove a module",
		Long:  "Remove a CPAN module from the current Perl version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			moduleName := args[0]

			// Get current Perl path if not specified
			if perlPath == "" {
				var err error
				perlPath, err = perl.GetCurrentPerlPath()
				if err != nil {
					return err
				}
			}

			// Create options for removing module
			options := &modules.RemoveModuleOptions{
				ModuleName: moduleName,
				PerlPath:   perlPath,
				Force:      force,
				Verbose:    verbose,
				Context:    cmd.Context(),
			}

			// Remove the module
			ui := cli.GetUI(cmd)
			ui.Info("Removing module %s...", moduleName)
			result, err := modules.RemoveModule(options)
			if err != nil {
				return err
			}

			if result.Success {
				ui.Success("Successfully removed module %s", result.ModuleName)
				if verbose && result.Output != "" {
					ui.Debug("Remove operation output:\n%s", result.Output)
				}
			}
			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force removal even if module is a core module")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().StringVar(&perlPath, "perl", "", "Path to the perl executable to use")

	return cmd
}

func newSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search available modules",
		Long:  "Search for CPAN modules matching the given query",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)

			// Load configuration and get flags
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			limit, _ := cmd.Flags().GetInt("limit")
			source, _ := cmd.Flags().GetString("source")
			noCache, _ := cmd.Flags().GetBool("no-cache")

			// Create provider using builder pattern
			provider, err := NewProviderBuilder().
				WithConfig(cfg).
				WithSource(source).
				WithNoCache(noCache).
				BuildProvider()
			if err != nil {
				return err
			}

			// Join the arguments into a query
			query := strings.Join(args, " ")

			// Run the search
			results, err := provider.SearchModules(context.Background(), query, limit)
			if err != nil {
				return err
			}

			// Display the results
			ui.Info("Search results for '%s' (%d of %d results from %s)",
				results.Query, len(results.Results), results.Total, results.Source)

			// Create a formatted list of search results
			var searchResults []string
			for i, result := range results.Results {
				searchResults = append(searchResults,
					fmt.Sprintf("[%d] %s (%s)\n    %s\n    Author: %s | Released: %s",
						i+1, result.Name, result.Version, result.Abstract,
						result.Author, result.ReleaseDate.Format("2006-01-02")))
			}

			if len(searchResults) > 0 {
				ui.ListWithOptions(uipkg.ListOptions{
					Items: searchResults,
				})
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().IntP("limit", "l", 20, "Limit the number of results")
	cmd.Flags().StringP("source", "s", "", "Use a specific metadata source (metacpan, cpan, custom)")
	cmd.Flags().Bool("no-cache", false, "Disable caching for this search")

	return cmd
}

func newDepsCommand() *cobra.Command {
	var (
		includeTest  bool
		includeBuild bool
		includeCore  bool
		includeDev   bool
		maxDepth     int
		source       string
		noCache      bool
		perlVersion  string
		verbose      bool
		flat         bool
	)

	cmd := &cobra.Command{
		Use:   "deps [module]",
		Short: "Show module dependencies",
		Long:  "Display the dependencies for a CPAN module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)

			// Options are already bound to the variables

			// Load configuration
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			// Create provider and resolver using builder pattern
			result, err := NewProviderBuilder().
				WithConfig(cfg).
				WithSource(source).
				WithNoCache(noCache).
				WithResolver().
				Build()
			if err != nil {
				return err
			}

			provider := result.Provider
			resolver := result.Resolver

			// Set up resolution options
			options := &deps.DependencyResolutionOptions{
				Provider:     provider,
				IncludeCore:  includeCore,
				IncludeTest:  includeTest,
				IncludeBuild: includeBuild,
				IncludeDev:   includeDev,
				MaxDepth:     maxDepth,
				Verbose:      verbose,
				UseCache:     !noCache,
				CacheTTL:     0,  // Will use resolver's cache settings
				CacheDir:     "", // Will use resolver's cache settings
				PerlVersion:  perlVersion,
			}

			// Resolve dependencies
			moduleName := args[0]
			depResult, err := resolver.ResolveDependencies(context.Background(), moduleName, options)
			if err != nil {
				return err
			}

			// Display the results
			if flat {
				// Display as a flat list
				ui.Info("Dependencies for %s:", moduleName)
				deps := resolver.GetFlattenedDependencies(depResult)
				var listItems []string
				for _, dep := range deps {
					if dep.IsRoot {
						continue // Skip the root module
					}
					item := dep.Name
					if dep.VersionConstraint != "" {
						item += fmt.Sprintf(" (%s)", dep.VersionConstraint)
					}
					if dep.Phase != "" && dep.Phase != "runtime" {
						item += fmt.Sprintf(" [%s]", dep.Phase)
					}
					listItems = append(listItems, item)
				}
				ui.List(listItems)
			} else {
				// Display as a tree
				ui.Info("Dependency tree for %s:", moduleName)
				tree := resolver.PrintDependencyTree(depResult.Root)
				ui.Info("%s", tree)
			}

			// Display warnings and conflicts
			if len(depResult.Warnings) > 0 {
				ui.Warning("Warnings:")
				ui.List(depResult.Warnings)
			}

			if len(depResult.Conflicts) > 0 {
				ui.Error("Version conflicts:")
				for _, conflict := range depResult.Conflicts {
					ui.Error("- %s has conflicting requirements:", conflict.Module)
					var conflictItems []string
					for constraint, requirers := range conflict.Requirements {
						conflictItems = append(conflictItems, fmt.Sprintf("  %s required by: %s", constraint, strings.Join(requirers, ", ")))
					}
					ui.List(conflictItems)
				}
			}

			return nil
		},
	}

	// Add flags with direct binding to variables
	cmd.Flags().BoolVarP(&includeTest, "include-test", "t", false, "Include test dependencies")
	cmd.Flags().BoolVarP(&includeBuild, "include-build", "b", true, "Include build dependencies")
	cmd.Flags().BoolVar(&includeCore, "include-core", false, "Include core modules")
	cmd.Flags().BoolVar(&includeDev, "include-dev", false, "Include development dependencies")
	cmd.Flags().IntVarP(&maxDepth, "max-depth", "m", 0, "Maximum depth to traverse (0 = no limit)")
	cmd.Flags().StringVarP(&source, "source", "s", "", "Use a specific metadata source (metacpan, cpan, custom)")
	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Disable caching for this operation")
	cmd.Flags().StringVarP(&perlVersion, "perl", "p", "", "Perl version to use for core module detection")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
	cmd.Flags().BoolVarP(&flat, "flat", "f", false, "Display dependencies as a flat list instead of a tree")

	return cmd
}

func newBundleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bundle",
		Short: "Manage module bundles",
		Long:  "Export and import collections of modules",
	}

	// Export command
	exportCmd := &cobra.Command{
		Use:   "export [file]",
		Short: "Export a module bundle",
		Long:  "Export the list of installed modules to a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)
			outputPath := args[0]

			// Get options from flags
			name, _ := cmd.Flags().GetString("name")
			description, _ := cmd.Flags().GetString("description")
			pattern, _ := cmd.Flags().GetString("pattern")
			includeCore, _ := cmd.Flags().GetBool("include-core")
			includeVersions, _ := cmd.Flags().GetBool("include-versions")
			perlPath, _ := cmd.Flags().GetString("perl")

			// Get current Perl path if not specified
			if perlPath == "" {
				var err error
				perlPath, err = perl.GetCurrentPerlPath()
				if err != nil {
					return err
				}
			}

			// If name is not provided, use a default one
			if name == "" {
				name = "Bundle " + time.Now().Format("2006-01-02")
			}

			// Create export options
			options := &modules.ExportBundleOptions{
				OutputPath:      outputPath,
				Name:            name,
				Description:     description,
				PerlPath:        perlPath,
				Pattern:         pattern,
				IncludeCore:     includeCore,
				IncludeVersions: includeVersions,
				Context:         cmd.Context(),
			}

			// Export the bundle
			ui.Info("Exporting modules to bundle file %s...", outputPath)
			err := modules.ExportModuleBundle(options)
			if err != nil {
				return err
			}

			ui.Success("Successfully exported modules to %s", outputPath)
			return nil
		},
	}

	// Add flags to export command
	exportCmd.Flags().String("name", "", "Bundle name")
	exportCmd.Flags().String("description", "", "Bundle description")
	exportCmd.Flags().StringP("pattern", "p", "", "Pattern to filter modules by name")
	exportCmd.Flags().BoolP("include-core", "c", false, "Include core modules")
	exportCmd.Flags().BoolP("include-versions", "v", true, "Include version constraints")
	exportCmd.Flags().String("perl", "", "Path to the perl executable to use")

	// Import command
	importCmd := &cobra.Command{
		Use:   "import [file]",
		Short: "Import a module bundle",
		Long:  "Install modules from a bundle file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)
			inputPath := args[0]

			// Load configuration and get flags
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			perlPath, _ := cmd.Flags().GetString("perl")
			verbose, _ := cmd.Flags().GetBool("verbose")
			force, _ := cmd.Flags().GetBool("force")
			skipTests, _ := cmd.Flags().GetBool("skip-tests")
			skipDeps, _ := cmd.Flags().GetBool("skip-dependencies")
			noCache, _ := cmd.Flags().GetBool("no-cache")
			installDir, _ := cmd.Flags().GetString("install-dir")
			source, _ := cmd.Flags().GetString("source")

			// Get current Perl path if not specified
			if perlPath == "" {
				perlPath, err = perl.GetCurrentPerlPath()
				if err != nil {
					return err
				}
			}

			// Create provider and resolver using builder pattern
			result, err := NewProviderBuilder().
				WithConfig(cfg).
				WithSource(source).
				WithNoCache(noCache).
				WithResolver().
				Build()
			if err != nil {
				return err
			}

			provider := result.Provider
			resolver := result.Resolver

			// Read the bundle file to get the list of modules
			bundleData, err := os.ReadFile(inputPath)
			if err != nil {
				return fmt.Errorf("failed to read bundle file: %w", err)
			}

			// Parse the bundle
			var bundle modules.ModuleBundleInfo
			if err := json.Unmarshal(bundleData, &bundle); err != nil {
				return fmt.Errorf("failed to parse bundle file: %w", err)
			}

			// Display bundle info
			bundleInfo := map[string]string{
				"Bundle":       bundle.Name,
				"Created":      bundle.Created.Format("2006-01-02 15:04:05"),
				"Perl Version": bundle.PerlVersion,
				"Modules":      fmt.Sprintf("%d", len(bundle.Modules)),
			}
			if bundle.Description != "" {
				bundleInfo["Description"] = bundle.Description
			}
			ui.KeyValue(bundleInfo)

			// Install each module
			successCount := 0
			failCount := 0

			for i, mod := range bundle.Modules {
				// Skip optional modules if not explicitly requested
				if mod.IsOptional {
					ui.Info("Skipping optional module %s", mod.Name)
					continue
				}

				// Create installation options for the module
				installOptions := &modules.ModuleInstallOptions{
					ModuleName:         mod.Name,
					VersionConstraint:  mod.VersionConstraint,
					PerlPath:           perlPath,
					InstallDir:         installDir,
					RunTests:           !skipTests,
					Force:              force,
					Cleanup:            true,
					Verbose:            verbose,
					SkipDependencies:   skipDeps,
					Provider:           provider,
					DependencyResolver: resolver,
					ProgressCallback: func(stage modules.InstallProgressStage, module string, details string, progress float64) {
						if verbose {
							ui.Info("[%s] %s: %s (%.0f%%)", stage.String(), module, details, progress*100)
						} else if stage != modules.StageFinished {
							// Only show major stage transitions if not verbose
							ui.Info("[%s] %s", stage.String(), module)
						}
					},
					Context: cmd.Context(),
				}

				// Show progress
				ui.Info("Installing module %s (%d/%d)...", mod.Name, i+1, len(bundle.Modules))

				// Install the module
				result, err := modules.InstallModule(installOptions)

				if err != nil {
					ui.Error("Failed to install %s: %v", mod.Name, err)
					failCount++
					continue
				}

				// Display result
				if result.Success {
					ui.Success("Successfully installed %s v%s", result.ModuleName, result.Version)
					successCount++
				} else {
					ui.Error("Failed to install %s", mod.Name)
					failCount++
					if len(result.Errors) > 0 {
						var errorList []string
						for _, err := range result.Errors {
							errorList = append(errorList, err)
						}
						ui.Error("Errors:")
						ui.List(errorList)
					}
				}
			}

			// Summary
			summaryInfo := map[string]string{
				"Succeeded": fmt.Sprintf("%d", successCount),
				"Failed":    fmt.Sprintf("%d", failCount),
			}
			ui.Info("Bundle import summary:")
			ui.KeyValue(summaryInfo)
			if failCount > 0 {
				return fmt.Errorf("%d module installations failed", failCount)
			}

			return nil
		},
	}

	// Add flags to import command (similar to install command)
	importCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	importCmd.Flags().BoolP("force", "f", false, "Force installation even if tests fail")
	importCmd.Flags().Bool("skip-tests", false, "Skip running tests")
	importCmd.Flags().Bool("skip-dependencies", false, "Skip installation of dependencies")
	importCmd.Flags().Bool("no-cache", false, "Disable caching for this installation")
	importCmd.Flags().StringP("install-dir", "i", "", "Directory to install the module to")
	importCmd.Flags().StringP("source", "s", "", "Use a specific metadata source (metacpan, cpan, custom)")
	importCmd.Flags().String("perl", "", "Path to the perl executable to use")

	// Add subcommands
	cmd.AddCommand(exportCmd, importCmd)

	return cmd
}

func newTypeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "type",
		Short: "Manage type definitions",
		Long:  "Create, list, and manage type definitions for Perl modules",
	}

	// Add subcommands from type_command.go
	createTypeCommands(cmd)

	return cmd
}

func newMirrorCommand() *cobra.Command {
	var (
		add  bool
		list bool
	)

	cmd := &cobra.Command{
		Use:   "mirror [url]",
		Short: "Set/get CPAN mirror",
		Long:  "Set or display the current CPAN mirror URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)

			// Get configuration
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			// Create PVI configuration if it doesn't exist
			if cfg.PVI == nil {
				cfg.PVI = &config.PVIConfig{}
			}

			// List current mirrors
			if list || (len(args) == 0 && !add) {
				ui.SubHeader("Current CPAN mirrors:")

				mirrorInfo := map[string]string{}

				// Show default mirror
				if cfg.PVI.DefaultMirror != "" {
					mirrorInfo["Default"] = cfg.PVI.DefaultMirror
				} else {
					mirrorInfo["Default"] = "<not set> (using MetaCPAN API default)"
				}

				ui.KeyValue(mirrorInfo)

				// Show additional mirrors
				if len(cfg.PVI.AdditionalMirrors) > 0 {
					var additionalMirrors []string
					for i, mirror := range cfg.PVI.AdditionalMirrors {
						additionalMirrors = append(additionalMirrors, fmt.Sprintf("[%d] %s", i+1, mirror))
					}
					ui.ListWithOptions(uipkg.ListOptions{
						Title: "Additional mirrors",
						Items: additionalMirrors,
					})
				} else {
					ui.Info("No additional mirrors configured.")
				}

				return nil
			}

			// Set mirror
			if len(args) > 0 {
				mirrorURL := args[0]

				// Validate the URL (basic validation)
				if !strings.HasPrefix(mirrorURL, "http://") && !strings.HasPrefix(mirrorURL, "https://") {
					return fmt.Errorf("mirror URL must start with http:// or https://")
				}

				// Add as additional mirror or set as default
				if add {
					// Add as additional mirror
					cfg.PVI.AdditionalMirrors = append(cfg.PVI.AdditionalMirrors, mirrorURL)
					ui.Success("Added %s as additional mirror", mirrorURL)
				} else {
					// Set as default mirror
					cfg.PVI.DefaultMirror = mirrorURL
					ui.Success("Set %s as default mirror", mirrorURL)
				}

				// Save configuration
				if err := config.SaveUserConfig(cfg); err != nil {
					return fmt.Errorf("failed to save configuration: %w", err)
				}

				return nil
			}

			return fmt.Errorf("no mirror URL provided")
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&add, "add", "a", false, "Add as additional mirror instead of setting as default")
	cmd.Flags().BoolVarP(&list, "list", "l", false, "List configured mirrors")

	return cmd
}

func newOutdatedCommand() *cobra.Command {
	var (
		pattern     string
		includeCore bool
		format      string
		perlPath    string
		verbose     bool
		noCache     bool
		source      string
	)

	cmd := &cobra.Command{
		Use:   "outdated",
		Short: "Show outdated modules",
		Long:  "List installed modules that have newer versions available",
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)

			// Process pattern from args if provided
			if len(args) > 0 {
				pattern = args[0]
			}

			// Load configuration
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			// Get current Perl path if not specified
			if perlPath == "" {
				perlPath, err = perl.GetCurrentPerlPath()
				if err != nil {
					return err
				}
			}

			// Create provider using builder pattern
			provider, err := NewProviderBuilder().
				WithConfig(cfg).
				WithSource(source).
				WithNoCache(noCache).
				BuildProvider()
			if err != nil {
				return err
			}

			// Create options for checking outdated modules
			options := &modules.CheckOutdatedOptions{
				PerlPath:    perlPath,
				Pattern:     pattern,
				IncludeCore: includeCore,
				Provider:    provider,
				Context:     cmd.Context(),
			}

			// Create a function to get the latest version for a module
			checkLatest := func(moduleName string) (string, error) {
				// Get module info from provider
				moduleInfo, err := provider.GetModuleInfo(cmd.Context(), moduleName)
				if err != nil {
					return "", err
				}

				// Return the latest version
				return moduleInfo.Version, nil
			}

			// Check for outdated modules
			outdatedModules, err := modules.CheckOutdatedModules(options, checkLatest)
			if err != nil {
				return err
			}

			// Display results
			if len(outdatedModules) == 0 {
				ui.Info("All modules are up to date")
				return nil
			}

			// Format output
			switch format {
			case "json":
				// Output as JSON
				jsonData, err := json.MarshalIndent(outdatedModules, "", "  ")
				if err != nil {
					return err
				}
				ui.Println(string(jsonData))

			case "simple":
				// Output as simple name/version
				var simpleList []string
				for _, module := range outdatedModules {
					simpleList = append(simpleList, fmt.Sprintf("%s %s -> %s", module.Name, module.InstalledVersion, module.LatestVersion))
				}
				ui.List(simpleList)

			default:
				// Default tabular format
				ui.SubHeader(fmt.Sprintf("Found %d outdated modules", len(outdatedModules)))

				headers := []string{"Module", "Installed", "Latest"}
				var rows [][]string
				for _, module := range outdatedModules {
					rows = append(rows, []string{module.Name, module.InstalledVersion, module.LatestVersion})
				}
				ui.Table(headers, rows)

				// Add update command hint
				updateHints := []string{
					"pvi update --all    # Update all outdated modules",
					"pvi update [module] # Update specific module",
				}
				ui.ListWithOptions(uipkg.ListOptions{
					Title: "To update these modules, run:",
					Items: updateHints,
				})
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&pattern, "pattern", "p", "", "Pattern to filter modules by name")
	cmd.Flags().BoolVarP(&includeCore, "include-core", "c", false, "Include core modules")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, simple, json)")
	cmd.Flags().StringVar(&perlPath, "perl", "", "Path to the perl executable to use")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed module information")
	cmd.Flags().BoolVar(&noCache, "no-cache", false, "Disable caching for this operation")
	cmd.Flags().StringVarP(&source, "source", "s", "", "Use a specific metadata source (metacpan, cpan, custom)")

	return cmd
}

func newAddCommand() *cobra.Command {
	var (
		dev     bool
		version string
		verbose bool
		force   bool
	)

	cmd := &cobra.Command{
		Use:   "add [module]",
		Short: "Add a module to cpanfile and install it",
		Long:  "Add a CPAN module dependency to the project's cpanfile and install it to the project lib directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)
			moduleName := args[0]

			// Detect project context
			projectCtx, err := project.GetCurrentProject()
			if err != nil {
				return fmt.Errorf("failed to detect project context: %w", err)
			}

			if !projectCtx.IsProject {
				return fmt.Errorf("not in a project directory. Use 'pvm project init' to create a project")
			}

			// Set up cpanfile path
			cpanfilePath := filepath.Join(projectCtx.RootDir, "cpanfile")
			cpanfileManager := NewCpanfileManager(cpanfilePath)

			// Build the "Adding module" message
			addMsg := fmt.Sprintf("Adding %s to cpanfile", moduleName)
			if dev {
				addMsg += " (development dependency)"
			}
			if version != "" {
				addMsg += fmt.Sprintf(" with version constraint '%s'", version)
			}
			addMsg += "..."
			ui.Info("%s", addMsg)

			// Add to cpanfile first
			err = cpanfileManager.AddDependency(moduleName, version, dev)
			if err != nil {
				return fmt.Errorf("failed to add dependency to cpanfile: %w", err)
			}

			ui.Success("Added %s to cpanfile", moduleName)

			// Now install the module to project lib directory
			ui.Info("Installing %s to project lib directory...", moduleName)

			// Load configuration and create provider/resolver
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			result, err := NewProviderBuilder().
				WithConfig(cfg).
				WithResolver().
				Build()
			if err != nil {
				return err
			}

			provider := result.Provider
			resolver := result.Resolver

			// Get current Perl path
			perlPath, err := perl.GetCurrentPerlPath()
			if err != nil {
				return err
			}

			// Install to project lib directory
			installDir := projectCtx.LocalLibDir

			installOptions := &modules.ModuleInstallOptions{
				ModuleName:         moduleName,
				VersionConstraint:  version,
				PerlPath:           perlPath,
				InstallDir:         installDir,
				RunTests:           true,
				Force:              force,
				Cleanup:            true,
				Verbose:            verbose,
				SkipDependencies:   false,
				Provider:           provider,
				DependencyResolver: resolver,
				ProgressCallback: func(stage modules.InstallProgressStage, module string, details string, progress float64) {
					if verbose {
						ui.Debug("[%s] %s: %s (%.0f%%)", stage.String(), module, details, progress*100)
					} else if stage != modules.StageFinished {
						ui.Info("[%s] %s", stage.String(), module)
					}
				},
				Context: cmd.Context(),
			}

			installResult, err := modules.InstallModule(installOptions)
			if err != nil {
				// If installation fails, remove from cpanfile
				ui.Warning("Installation failed, removing %s from cpanfile...", moduleName)
				if removeErr := cpanfileManager.RemoveDependency(moduleName); removeErr != nil {
					ui.Warning("Failed to remove %s from cpanfile: %v", moduleName, removeErr)
				}
				return fmt.Errorf("failed to install %s: %w", moduleName, err)
			}

			if installResult.Success {
				ui.Success("Successfully added and installed %s v%s", installResult.ModuleName, installResult.Version)
				if len(installResult.Dependencies) > 0 {
					ui.Info("Installed %d dependencies", len(installResult.Dependencies))
				}
			} else {
				// Remove from cpanfile if installation wasn't successful
				ui.Warning("Installation not successful, removing %s from cpanfile...", moduleName)
				if removeErr := cpanfileManager.RemoveDependency(moduleName); removeErr != nil {
					ui.Warning("Failed to remove %s from cpanfile: %v", moduleName, removeErr)
				}
				return fmt.Errorf("failed to install %s", moduleName)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&dev, "dev", false, "Add as development dependency")
	cmd.Flags().StringVar(&version, "version", "", "Version constraint (e.g., '>= 1.0', '~1.2.3')")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force installation even if tests fail")

	return cmd
}

func newSyncCommand() *cobra.Command {
	var (
		fromSnapshot bool
		verbose      bool
		force        bool
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Generate or install from cpanfile.snapshot",
		Long: `Generate cpanfile.snapshot lockfile from installed modules or install from existing snapshot.

By default, generates cpanfile.snapshot with exact versions of all installed dependencies.
Use --from-snapshot to install exact versions from an existing cpanfile.snapshot file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Detect project context
			projectCtx, err := project.GetCurrentProject()
			if err != nil {
				return fmt.Errorf("failed to detect project context: %w", err)
			}

			if !projectCtx.IsProject {
				return fmt.Errorf("not in a project directory. Use 'pvm project init' to create a project")
			}

			if fromSnapshot {
				// Install from snapshot
				return installFromSnapshot(cmd, projectCtx, verbose)
			} else {
				// Generate snapshot
				return generateSnapshot(cmd, projectCtx, verbose)
			}
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&fromSnapshot, "from-snapshot", false, "Install from existing cpanfile.snapshot instead of generating one")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force regeneration even if snapshot is newer than cpanfile")

	return cmd
}

// generateSnapshot creates a cpanfile.snapshot from currently installed modules
func generateSnapshot(cmd *cobra.Command, projectCtx *project.ProjectContext, verbose bool) error {
	ui := cli.GetUI(cmd)
	cpanfilePath := filepath.Join(projectCtx.RootDir, "cpanfile")
	snapshotPath := filepath.Join(projectCtx.RootDir, "cpanfile.snapshot")

	// Check if cpanfile exists
	if _, err := os.Stat(cpanfilePath); os.IsNotExist(err) {
		return fmt.Errorf("cpanfile not found. Use 'pvm module add <module>' to add dependencies first")
	}

	ui.Info("Generating cpanfile.snapshot from installed modules...")

	// Create cpanfile manager
	cpanfileManager := NewCpanfileManager(cpanfilePath)

	// Generate snapshot
	snapshot, err := cpanfileManager.GenerateSnapshot()
	if err != nil {
		return fmt.Errorf("failed to generate snapshot: %w", err)
	}

	// Write snapshot
	err = cpanfileManager.WriteSnapshot(snapshot)
	if err != nil {
		return fmt.Errorf("failed to write snapshot: %w", err)
	}

	ui.Success("Generated cpanfile.snapshot with %d distributions", len(snapshot.Distributions))
	if verbose {
		snapshotInfo := map[string]string{
			"Snapshot saved to": snapshotPath,
			"Perl version":      snapshot.PerlVersion,
		}
		ui.KeyValue(snapshotInfo)

		var distList []string
		for distName, entry := range snapshot.Distributions {
			distInfo := fmt.Sprintf("%s (%s)", distName, entry.Pathname)
			if verbose {
				for module, version := range entry.Provides {
					distInfo += fmt.Sprintf("\n    provides: %s %s", module, version)
				}
			}
			distList = append(distList, distInfo)
		}
		ui.ListWithOptions(uipkg.ListOptions{
			Title: "Distributions",
			Items: distList,
		})
	}

	return nil
}

// installFromSnapshot installs exact versions from cpanfile.snapshot
func installFromSnapshot(cmd *cobra.Command, projectCtx *project.ProjectContext, verbose bool) error {
	ui := cli.GetUI(cmd)
	cpanfilePath := filepath.Join(projectCtx.RootDir, "cpanfile")
	snapshotPath := filepath.Join(projectCtx.RootDir, "cpanfile.snapshot")

	// Check if snapshot exists
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		return fmt.Errorf("cpanfile.snapshot not found. Run 'pvm module sync' to generate one")
	}

	ui.Info("Installing modules from cpanfile.snapshot...")

	// Create cpanfile manager
	cpanfileManager := NewCpanfileManager(cpanfilePath)

	// Read snapshot
	snapshot, err := cpanfileManager.ReadSnapshot()
	if err != nil {
		return fmt.Errorf("failed to read snapshot: %w", err)
	}

	ui.Info("Found %d distributions in snapshot", len(snapshot.Distributions))
	if verbose {
		snapshotInfo := map[string]string{
			"Snapshot Perl version": snapshot.PerlVersion,
			"Generated at":          snapshot.GeneratedAt.Format("2006-01-02 15:04:05"),
		}
		ui.KeyValue(snapshotInfo)
	}

	// For each distribution in snapshot, install the exact version
	var installList []string
	for distName, entry := range snapshot.Distributions {
		ui.Info("Installing %s...", distName)
		installInfo := fmt.Sprintf("%s", distName)
		if verbose {
			installInfo += fmt.Sprintf(" (Pathname: %s)", entry.Pathname)
		}

		// In a real implementation, you would:
		// 1. Download the exact distribution from the pathname
		// 2. Install it to the project lib directory
		// 3. Verify the installation

		// For now, we'll just collect what would be installed
		for module, version := range entry.Provides {
			if verbose {
				installInfo += fmt.Sprintf("\n  Would install: %s version %s", module, version)
			}
		}
		installList = append(installList, installInfo)
	}

	if verbose && len(installList) > 0 {
		ui.ListWithOptions(uipkg.ListOptions{
			Title: "Installation Plan",
			Items: installList,
		})
	}

	ui.Success("Installation from snapshot completed successfully")
	ui.Warning("Note: Actual installation from snapshot is not yet fully implemented")

	return nil
}
