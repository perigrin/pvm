// ABOUTME: PVI-specific commands and functionality
// ABOUTME: Implements commands for Perl module management

package pvi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/perl"
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
	)

	cmd := &cobra.Command{
		Use:   "install [module]",
		Short: "Install a module",
		Long:  "Install a CPAN module for the current Perl version",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			moduleName := args[0]

			// Get configuration
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			// Get options from flags and config
			if source == "" && cfg.PVI != nil {
				source = cfg.PVI.MetadataSource
			}

			// Default to metacpan if still not set
			if source == "" {
				source = "metacpan"
			}

			// Build the provider options
			var providerOpts []cpan.ProviderOption

			// Add options based on configuration
			if cfg.PVI != nil {
				// Set base URL if custom source and URL provided
				if source == "custom" && cfg.PVI.MetadataURL != "" {
					providerOpts = append(providerOpts, cpan.WithBaseURL(cfg.PVI.MetadataURL))
				}

				// Set cache settings if caching is enabled and not disabled by flag
				if cfg.PVI.CacheModules && !noCache && cfg.PVI.CacheDir != "" {
					providerOpts = append(providerOpts, cpan.WithCache(cfg.PVI.CacheDir, cfg.PVI.CacheTTL))
				}

				// Set mirrors
				if cfg.PVI.DefaultMirror != "" {
					providerOpts = append(providerOpts, cpan.WithMirror(cfg.PVI.DefaultMirror))
				}
				if len(cfg.PVI.AdditionalMirrors) > 0 {
					providerOpts = append(providerOpts, cpan.WithAdditionalMirrors(cfg.PVI.AdditionalMirrors))
				}

				// Set network access
				providerOpts = append(providerOpts, cpan.WithDisableNetwork(cfg.PVI.DisableNetwork))
			}

			// Create the provider
			provider, err := cpan.NewProvider(source, providerOpts...)
			if err != nil {
				return err
			}

			// Create the dependency resolver
			var cacheDir string
			var cacheTTL int

			if cfg.PVI != nil && !noCache {
				cacheDir = cfg.PVI.CacheDir
				cacheTTL = cfg.PVI.CacheTTL
			}

			resolver, err := deps.NewDefaultResolver(cacheDir, cacheTTL)
			if err != nil {
				return err
			}

			// Get current Perl path
			perlPath, err := perl.GetCurrentPerlPath()
			if err != nil {
				return err
			}

			// Create installation options
			installOptions := &modules.ModuleInstallOptions{
				ModuleName:         moduleName,
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
						cmd.Printf("[%s] %s: %s (%.0f%%)\n", stage.String(), module, details, progress*100)
					} else if stage != modules.StageFinished {
						// Only show major stage transitions if not verbose
						cmd.Printf("[%s] %s\n", stage.String(), module)
					}
				},
				Context: cmd.Context(),
			}

			// Install the module
			cmd.Printf("Installing module %s...\n", moduleName)
			result, err := modules.InstallModule(installOptions)

			if err != nil {
				cmd.Printf("Failed to install %s: %v\n", moduleName, err)
				return err
			}

			// Display result
			if result.Success {
				cmd.Printf("Successfully installed %s v%s\n", result.ModuleName, result.Version)

				// Show installed dependencies count
				if len(result.Dependencies) > 0 {
					cmd.Printf("Installed %d dependencies\n", len(result.Dependencies))
				}

				// Show warnings if any
				if len(result.Warnings) > 0 && verbose {
					cmd.Println("Warnings:")
					for _, warning := range result.Warnings {
						cmd.Printf("  - %s\n", warning)
					}
				}

				cmd.Printf("Total installation time: %s\n", result.Duration.Round(time.Second))
			} else {
				cmd.Printf("Failed to install %s\n", moduleName)
				if len(result.Errors) > 0 {
					cmd.Println("Errors:")
					for _, err := range result.Errors {
						cmd.Printf("  - %s\n", err)
					}
				}
				return fmt.Errorf("installation failed")
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
				cmd.Println("No modules found")
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
				cmd.Println(string(jsonData))

			case "simple":
				// Output as simple name/version
				for _, module := range moduleList {
					cmd.Printf("%s %s\n", module.Name, module.Version)
				}

			default:
				// Default tabular format
				cmd.Printf("Found %d modules\n\n", len(moduleList))

				if verbose {
					// Detailed format
					for i, module := range moduleList {
						cmd.Printf("[%d] %s (%s)\n", i+1, module.Name, module.Version)
						if module.Description != "" {
							cmd.Printf("    Description: %s\n", module.Description)
						}
						cmd.Printf("    Path: %s\n", module.Path)
						if !module.InstallationTime.IsZero() {
							cmd.Printf("    Installed: %s\n", module.InstallationTime.Format("2006-01-02 15:04:05"))
						}
						if module.CoreModule {
							cmd.Printf("    Core Module: Yes\n")
						}
						cmd.Println()
					}
				} else {
					// Table format
					cmd.Printf("%-40s %-15s %-10s\n", "Module", "Version", "Core Module")
					cmd.Printf("%s\n", strings.Repeat("-", 68))

					for _, module := range moduleList {
						coreStatus := ""
						if module.CoreModule {
							coreStatus = "Yes"
						}
						cmd.Printf("%-40s %-15s %-10s\n", module.Name, module.Version, coreStatus)
					}
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
			// Check if at least one module is specified or --all flag is used
			if len(args) == 0 && !all {
				return fmt.Errorf("at least one module name must be provided or use --all flag")
			}

			// Get configuration
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			// Get options from flags and config
			if source == "" && cfg.PVI != nil {
				source = cfg.PVI.MetadataSource
			}

			// Default to metacpan if still not set
			if source == "" {
				source = "metacpan"
			}

			// Get current Perl path if not specified
			if perlPath == "" {
				perlPath, err = perl.GetCurrentPerlPath()
				if err != nil {
					return err
				}
			}

			// Build the provider options
			var providerOpts []cpan.ProviderOption

			// Add options based on configuration
			if cfg.PVI != nil {
				// Set base URL if custom source and URL provided
				if source == "custom" && cfg.PVI.MetadataURL != "" {
					providerOpts = append(providerOpts, cpan.WithBaseURL(cfg.PVI.MetadataURL))
				}

				// Set cache settings if caching is enabled and not disabled by flag
				if cfg.PVI.CacheModules && !noCache && cfg.PVI.CacheDir != "" {
					providerOpts = append(providerOpts, cpan.WithCache(cfg.PVI.CacheDir, cfg.PVI.CacheTTL))
				}

				// Set mirrors
				if cfg.PVI.DefaultMirror != "" {
					providerOpts = append(providerOpts, cpan.WithMirror(cfg.PVI.DefaultMirror))
				}
				if len(cfg.PVI.AdditionalMirrors) > 0 {
					providerOpts = append(providerOpts, cpan.WithAdditionalMirrors(cfg.PVI.AdditionalMirrors))
				}

				// Set network access
				providerOpts = append(providerOpts, cpan.WithDisableNetwork(cfg.PVI.DisableNetwork))
			}

			// Create the provider
			provider, err := cpan.NewProvider(source, providerOpts...)
			if err != nil {
				return err
			}

			// Create the dependency resolver
			var cacheDir string
			var cacheTTL int

			if cfg.PVI != nil && !noCache {
				cacheDir = cfg.PVI.CacheDir
				cacheTTL = cfg.PVI.CacheTTL
			}

			resolver, err := deps.NewDefaultResolver(cacheDir, cacheTTL)
			if err != nil {
				return err
			}

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

				cmd.Printf("Updating all %d installed modules...\n", len(args))
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
							cmd.Printf("[%s] %s: %s (%.0f%%)\n", stage.String(), module, details, progress*100)
						} else if stage != modules.StageFinished {
							// Only show major stage transitions if not verbose
							cmd.Printf("[%s] %s\n", stage.String(), module)
						}
					},
					Context: cmd.Context(),
				}

				// Update (install) the module
				cmd.Printf("Updating module %s...\n", moduleName)
				result, err := modules.InstallModule(installOptions)

				if err != nil {
					cmd.Printf("Failed to update %s: %v\n", moduleName, err)
					failCount++
					continue
				}

				// Display result
				if result.Success {
					cmd.Printf("Successfully updated %s to v%s\n", result.ModuleName, result.Version)
					successCount++

					// Show warnings if any
					if len(result.Warnings) > 0 && verbose {
						cmd.Println("Warnings:")
						for _, warning := range result.Warnings {
							cmd.Printf("  - %s\n", warning)
						}
					}
				} else {
					cmd.Printf("Failed to update %s\n", moduleName)
					failCount++
					if len(result.Errors) > 0 {
						cmd.Println("Errors:")
						for _, err := range result.Errors {
							cmd.Printf("  - %s\n", err)
						}
					}
				}
			}

			// Summary
			cmd.Printf("\nUpdate summary: %d succeeded, %d failed\n", successCount, failCount)
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
			cmd.Printf("Removing module %s...\n", moduleName)
			err := modules.RemoveModule(options)
			if err != nil {
				return err
			}

			cmd.Printf("Successfully removed module %s\n", moduleName)
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
			// Get configuration
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			// Get options from flags
			limit, _ := cmd.Flags().GetInt("limit")
			source, _ := cmd.Flags().GetString("source")
			noCache, _ := cmd.Flags().GetBool("no-cache")

			// If source is not provided, use the one from the configuration
			if source == "" && cfg.PVI != nil {
				source = cfg.PVI.MetadataSource
			}

			// Default to metacpan if still not set
			if source == "" {
				source = "metacpan"
			}

			// Build the provider options
			var providerOpts []cpan.ProviderOption

			// Add options based on configuration
			if cfg.PVI != nil {
				// Set base URL if custom source and URL provided
				if source == "custom" && cfg.PVI.MetadataURL != "" {
					providerOpts = append(providerOpts, cpan.WithBaseURL(cfg.PVI.MetadataURL))
				}

				// Set cache settings if caching is enabled and not disabled by flag
				if cfg.PVI.CacheModules && !noCache && cfg.PVI.CacheDir != "" {
					providerOpts = append(providerOpts, cpan.WithCache(cfg.PVI.CacheDir, cfg.PVI.CacheTTL))
				}

				// Set mirrors
				if cfg.PVI.DefaultMirror != "" {
					providerOpts = append(providerOpts, cpan.WithMirror(cfg.PVI.DefaultMirror))
				}
				if len(cfg.PVI.AdditionalMirrors) > 0 {
					providerOpts = append(providerOpts, cpan.WithAdditionalMirrors(cfg.PVI.AdditionalMirrors))
				}

				// Set network access
				providerOpts = append(providerOpts, cpan.WithDisableNetwork(cfg.PVI.DisableNetwork))
			}

			// Create the provider
			provider, err := cpan.NewProvider(source, providerOpts...)
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
			cmd.Printf("Search results for '%s' (%d of %d results from %s):\n\n",
				results.Query, len(results.Results), results.Total, results.Source)

			for i, result := range results.Results {
				cmd.Printf("[%d] %s (%s)\n", i+1, result.Name, result.Version)
				cmd.Printf("    %s\n", result.Abstract)
				cmd.Printf("    Author: %s | Released: %s\n\n",
					result.Author, result.ReleaseDate.Format("2006-01-02"))
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
			// Options are already bound to the variables

			// Get configuration
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			// If source is not provided, use the one from the configuration
			if source == "" && cfg.PVI != nil {
				source = cfg.PVI.MetadataSource
			}

			// Default to metacpan if still not set
			if source == "" {
				source = "metacpan"
			}

			// Build the provider options
			var providerOpts []cpan.ProviderOption

			// Add options based on configuration
			if cfg.PVI != nil {
				// Set base URL if custom source and URL provided
				if source == "custom" && cfg.PVI.MetadataURL != "" {
					providerOpts = append(providerOpts, cpan.WithBaseURL(cfg.PVI.MetadataURL))
				}

				// Set cache settings if caching is enabled and not disabled by flag
				if cfg.PVI.CacheModules && !noCache && cfg.PVI.CacheDir != "" {
					providerOpts = append(providerOpts, cpan.WithCache(cfg.PVI.CacheDir, cfg.PVI.CacheTTL))
				}

				// Set mirrors
				if cfg.PVI.DefaultMirror != "" {
					providerOpts = append(providerOpts, cpan.WithMirror(cfg.PVI.DefaultMirror))
				}
				if len(cfg.PVI.AdditionalMirrors) > 0 {
					providerOpts = append(providerOpts, cpan.WithAdditionalMirrors(cfg.PVI.AdditionalMirrors))
				}

				// Set network access
				providerOpts = append(providerOpts, cpan.WithDisableNetwork(cfg.PVI.DisableNetwork))
			}

			// Create the provider
			provider, err := cpan.NewProvider(source, providerOpts...)
			if err != nil {
				return err
			}

			// Create the dependency resolver
			var cacheDir string
			var cacheTTL int

			if cfg.PVI != nil && !noCache {
				cacheDir = cfg.PVI.CacheDir
				cacheTTL = cfg.PVI.CacheTTL
			}

			resolver, err := deps.NewDefaultResolver(cacheDir, cacheTTL)
			if err != nil {
				return err
			}

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
				CacheTTL:     cacheTTL,
				CacheDir:     cacheDir,
				PerlVersion:  perlVersion,
			}

			// Resolve dependencies
			moduleName := args[0]
			result, err := resolver.ResolveDependencies(context.Background(), moduleName, options)
			if err != nil {
				return err
			}

			// Display the results
			if flat {
				// Display as a flat list
				cmd.Printf("Dependencies for %s:\n\n", moduleName)
				deps := resolver.GetFlattenedDependencies(result)
				for _, dep := range deps {
					if dep.IsRoot {
						continue // Skip the root module
					}
					cmd.Printf("%s", dep.Name)
					if dep.VersionConstraint != "" {
						cmd.Printf(" (%s)", dep.VersionConstraint)
					}
					if dep.Phase != "" && dep.Phase != "runtime" {
						cmd.Printf(" [%s]", dep.Phase)
					}
					cmd.Println()
				}
			} else {
				// Display as a tree
				cmd.Printf("Dependency tree for %s:\n\n", moduleName)
				tree := resolver.PrintDependencyTree(result.Root)
				cmd.Println(tree)
			}

			// Display warnings and conflicts
			if len(result.Warnings) > 0 {
				cmd.Println("\nWarnings:")
				for _, warning := range result.Warnings {
					cmd.Printf("- %s\n", warning)
				}
			}

			if len(result.Conflicts) > 0 {
				cmd.Println("\nVersion conflicts:")
				for _, conflict := range result.Conflicts {
					cmd.Printf("- %s has conflicting requirements:\n", conflict.Module)
					for constraint, requirers := range conflict.Requirements {
						cmd.Printf("  - %s required by: %s\n", constraint, strings.Join(requirers, ", "))
					}
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
			cmd.Printf("Exporting modules to bundle file %s...\n", outputPath)
			err := modules.ExportModuleBundle(options)
			if err != nil {
				return err
			}

			cmd.Printf("Successfully exported modules to %s\n", outputPath)
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
			inputPath := args[0]

			// Get configuration
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			// Get options from flags
			perlPath, _ := cmd.Flags().GetString("perl")
			verbose, _ := cmd.Flags().GetBool("verbose")
			force, _ := cmd.Flags().GetBool("force")
			skipTests, _ := cmd.Flags().GetBool("skip-tests")
			skipDeps, _ := cmd.Flags().GetBool("skip-dependencies")
			noCache, _ := cmd.Flags().GetBool("no-cache")
			installDir, _ := cmd.Flags().GetString("install-dir")
			source, _ := cmd.Flags().GetString("source")

			// Get options from flags and config
			if source == "" && cfg.PVI != nil {
				source = cfg.PVI.MetadataSource
			}

			// Default to metacpan if still not set
			if source == "" {
				source = "metacpan"
			}

			// Get current Perl path if not specified
			if perlPath == "" {
				perlPath, err = perl.GetCurrentPerlPath()
				if err != nil {
					return err
				}
			}

			// Build the provider options
			var providerOpts []cpan.ProviderOption

			// Add options based on configuration
			if cfg.PVI != nil {
				// Set base URL if custom source and URL provided
				if source == "custom" && cfg.PVI.MetadataURL != "" {
					providerOpts = append(providerOpts, cpan.WithBaseURL(cfg.PVI.MetadataURL))
				}

				// Set cache settings if caching is enabled and not disabled by flag
				if cfg.PVI.CacheModules && !noCache && cfg.PVI.CacheDir != "" {
					providerOpts = append(providerOpts, cpan.WithCache(cfg.PVI.CacheDir, cfg.PVI.CacheTTL))
				}

				// Set mirrors
				if cfg.PVI.DefaultMirror != "" {
					providerOpts = append(providerOpts, cpan.WithMirror(cfg.PVI.DefaultMirror))
				}
				if len(cfg.PVI.AdditionalMirrors) > 0 {
					providerOpts = append(providerOpts, cpan.WithAdditionalMirrors(cfg.PVI.AdditionalMirrors))
				}

				// Set network access
				providerOpts = append(providerOpts, cpan.WithDisableNetwork(cfg.PVI.DisableNetwork))
			}

			// Create the provider
			provider, err := cpan.NewProvider(source, providerOpts...)
			if err != nil {
				return err
			}

			// Create the dependency resolver
			var cacheDir string
			var cacheTTL int

			if cfg.PVI != nil && !noCache {
				cacheDir = cfg.PVI.CacheDir
				cacheTTL = cfg.PVI.CacheTTL
			}

			resolver, err := deps.NewDefaultResolver(cacheDir, cacheTTL)
			if err != nil {
				return err
			}

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
			cmd.Printf("Bundle: %s\n", bundle.Name)
			if bundle.Description != "" {
				cmd.Printf("Description: %s\n", bundle.Description)
			}
			cmd.Printf("Created: %s\n", bundle.Created.Format("2006-01-02 15:04:05"))
			cmd.Printf("Perl Version: %s\n", bundle.PerlVersion)
			cmd.Printf("Modules: %d\n\n", len(bundle.Modules))

			// Install each module
			successCount := 0
			failCount := 0

			for i, mod := range bundle.Modules {
				// Skip optional modules if not explicitly requested
				if mod.IsOptional {
					cmd.Printf("Skipping optional module %s\n", mod.Name)
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
							cmd.Printf("[%s] %s: %s (%.0f%%)\n", stage.String(), module, details, progress*100)
						} else if stage != modules.StageFinished {
							// Only show major stage transitions if not verbose
							cmd.Printf("[%s] %s\n", stage.String(), module)
						}
					},
					Context: cmd.Context(),
				}

				// Show progress
				cmd.Printf("Installing module %s (%d/%d)...\n", mod.Name, i+1, len(bundle.Modules))

				// Install the module
				result, err := modules.InstallModule(installOptions)

				if err != nil {
					cmd.Printf("Failed to install %s: %v\n", mod.Name, err)
					failCount++
					continue
				}

				// Display result
				if result.Success {
					cmd.Printf("Successfully installed %s v%s\n", result.ModuleName, result.Version)
					successCount++
				} else {
					cmd.Printf("Failed to install %s\n", mod.Name)
					failCount++
					if len(result.Errors) > 0 {
						cmd.Println("Errors:")
						for _, err := range result.Errors {
							cmd.Printf("  - %s\n", err)
						}
					}
				}
			}

			// Summary
			cmd.Printf("\nBundle import summary: %d succeeded, %d failed\n", successCount, failCount)
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
	return &cobra.Command{
		Use:   "type [module]",
		Short: "Manage type definitions for a module",
		Long:  "Generate, install, or view type definitions for a module",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Type command not yet implemented")
		},
	}
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
				cmd.Println("Current CPAN mirrors:")

				// Show default mirror
				if cfg.PVI.DefaultMirror != "" {
					cmd.Printf("Default: %s\n", cfg.PVI.DefaultMirror)
				} else {
					cmd.Println("Default: <not set> (using MetaCPAN API default)")
				}

				// Show additional mirrors
				if len(cfg.PVI.AdditionalMirrors) > 0 {
					cmd.Println("\nAdditional mirrors:")
					for i, mirror := range cfg.PVI.AdditionalMirrors {
						cmd.Printf("[%d] %s\n", i+1, mirror)
					}
				} else {
					cmd.Println("\nNo additional mirrors configured.")
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
					cmd.Printf("Added %s as additional mirror\n", mirrorURL)
				} else {
					// Set as default mirror
					cfg.PVI.DefaultMirror = mirrorURL
					cmd.Printf("Set %s as default mirror\n", mirrorURL)
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
			// Process pattern from args if provided
			if len(args) > 0 {
				pattern = args[0]
			}

			// Get configuration
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				return err
			}

			// Get options from flags and config
			if source == "" && cfg.PVI != nil {
				source = cfg.PVI.MetadataSource
			}

			// Default to metacpan if still not set
			if source == "" {
				source = "metacpan"
			}

			// Get current Perl path if not specified
			if perlPath == "" {
				perlPath, err = perl.GetCurrentPerlPath()
				if err != nil {
					return err
				}
			}

			// Build the provider options
			var providerOpts []cpan.ProviderOption

			// Add options based on configuration
			if cfg.PVI != nil {
				// Set base URL if custom source and URL provided
				if source == "custom" && cfg.PVI.MetadataURL != "" {
					providerOpts = append(providerOpts, cpan.WithBaseURL(cfg.PVI.MetadataURL))
				}

				// Set cache settings if caching is enabled and not disabled by flag
				if cfg.PVI.CacheModules && !noCache && cfg.PVI.CacheDir != "" {
					providerOpts = append(providerOpts, cpan.WithCache(cfg.PVI.CacheDir, cfg.PVI.CacheTTL))
				}

				// Set mirrors
				if cfg.PVI.DefaultMirror != "" {
					providerOpts = append(providerOpts, cpan.WithMirror(cfg.PVI.DefaultMirror))
				}
				if len(cfg.PVI.AdditionalMirrors) > 0 {
					providerOpts = append(providerOpts, cpan.WithAdditionalMirrors(cfg.PVI.AdditionalMirrors))
				}

				// Set network access
				providerOpts = append(providerOpts, cpan.WithDisableNetwork(cfg.PVI.DisableNetwork))
			}

			// Create the provider
			provider, err := cpan.NewProvider(source, providerOpts...)
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
				cmd.Println("All modules are up to date")
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
				cmd.Println(string(jsonData))

			case "simple":
				// Output as simple name/version
				for _, module := range outdatedModules {
					cmd.Printf("%s %s -> %s\n", module.Name, module.InstalledVersion, module.LatestVersion)
				}

			default:
				// Default tabular format
				cmd.Printf("Found %d outdated modules\n\n", len(outdatedModules))
				cmd.Printf("%-40s %-15s %-15s\n", "Module", "Installed", "Latest")
				cmd.Printf("%s\n", strings.Repeat("-", 72))

				for _, module := range outdatedModules {
					cmd.Printf("%-40s %-15s %-15s\n", module.Name, module.InstalledVersion, module.LatestVersion)
				}

				// Add update command hint
				cmd.Println("\nTo update these modules, run:")
				cmd.Println("  pvi update --all    # Update all outdated modules")
				cmd.Println("  pvi update [module] # Update specific module")
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
