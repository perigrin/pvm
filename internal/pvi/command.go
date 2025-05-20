// ABOUTME: PVI-specific commands and functionality
// ABOUTME: Implements commands for Perl module management

package pvi

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/pvi/deps"
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
	return &cobra.Command{
		Use:   "install [module]",
		Short: "Install a module",
		Long:  "Install a CPAN module for the current Perl version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Install command not yet implemented")
		},
	}
}

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed modules",
		Long:  "List all installed CPAN modules for the current Perl version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("List command not yet implemented")
		},
	}
}

func newUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update [module...]",
		Short: "Update modules",
		Long:  "Update one or more CPAN modules to the latest version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Update command not yet implemented")
		},
	}
}

func newRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [module]",
		Short: "Remove a module",
		Long:  "Remove a CPAN module from the current Perl version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Remove command not yet implemented")
		},
	}
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

	cmd.AddCommand(
		&cobra.Command{
			Use:   "export [file]",
			Short: "Export a module bundle",
			Long:  "Export the list of installed modules to a file",
			Run: func(cmd *cobra.Command, args []string) {
				cmd.Println("Bundle export not yet implemented")
			},
		},
		&cobra.Command{
			Use:   "import [file]",
			Short: "Import a module bundle",
			Long:  "Install modules from a bundle file",
			Run: func(cmd *cobra.Command, args []string) {
				cmd.Println("Bundle import not yet implemented")
			},
		},
	)

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
	return &cobra.Command{
		Use:   "mirror [url]",
		Short: "Set/get CPAN mirror",
		Long:  "Set or display the current CPAN mirror URL",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Mirror command not yet implemented")
		},
	}
}

func newOutdatedCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "outdated",
		Short: "Show outdated modules",
		Long:  "List installed modules that have newer versions available",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Outdated command not yet implemented")
		},
	}
}
