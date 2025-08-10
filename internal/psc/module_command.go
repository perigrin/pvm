// ABOUTME: PSC module management integration command
// ABOUTME: Provides type-aware module management for PSC projects

package psc

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/cli/progress"
	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/modules"
	"tamarou.com/pvm/internal/pm"
)

// newModuleCommand creates a module management command for PSC
func newModuleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module",
		Short: "Type-aware module management",
		Long:  "Manage CPAN modules with enhanced type checking integration for PSC projects",
	}

	cmd.AddCommand(
		newModuleInstallCommand(),
		newModuleAnalyzeCommand(),
		newModuleTypeDefsCommand(),
	)

	return cmd
}

// newModuleInstallCommand creates a command to install modules with type definitions
func newModuleInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [modules...]",
		Short: "Install modules with type definitions",
		Long:  "Install CPAN modules and generate/install corresponding type definitions for enhanced static analysis",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)

			// Get flags
			generateTypeDefs, _ := cmd.Flags().GetBool("generate-types")
			skipTests, _ := cmd.Flags().GetBool("skip-tests")
			verbose, _ := cmd.Flags().GetBool("verbose")
			_, _ = cmd.Flags().GetString("perl") // unused for now

			ui.Status("Installing modules with type-aware integration")

			// Load configuration
			cfg, err := config.LoadEffectiveConfig()
			if err != nil {
				ui.Warning("Failed to load configuration, using defaults: %v", err)
			}

			// Create provider using builder pattern
			source := "metacpan"
			if cfg != nil && cfg.PM != nil && cfg.PM.MetadataSource != "" {
				source = cfg.PM.MetadataSource
			}

			providerResult, err := pm.NewProviderBuilder().
				WithConfig(cfg).
				WithSource(source).
				WithResolver().
				Build()
			if err != nil {
				return fmt.Errorf("failed to create CPAN provider: %w", err)
			}

			// Create progress tracker
			tracker := progress.NewNullTracker()

			// Create logger
			logger := log.NewLogger(log.LevelInfo, os.Stderr, "PSC-Module")

			// Create installer using extracted packages
			installer := modules.NewInstaller(
				providerResult.Provider,
				providerResult.Resolver,
				tracker,
				logger,
			)

			// Set up install options
			installOptions := modules.InstallOptions{
				PerlPath:          "", // Will be resolved
				VersionConstraint: "",
				Force:             false,
				RunTests:          !skipTests,
				SkipDependencies:  false,
				Verbose:           verbose,
				Cleanup:           true,
				Context:           context.Background(),
			}

			// Install modules
			ctx := context.Background()
			var installResults []*modules.InstallResult

			if len(args) == 1 {
				// Single module installation
				result, err := installer.InstallModule(ctx, args[0], installOptions)
				if err != nil {
					return fmt.Errorf("failed to install module %s: %w", args[0], err)
				}
				installResults = []*modules.InstallResult{result}
			} else {
				// Batch installation
				installResults, err = installer.InstallBatch(ctx, args, installOptions)
				if err != nil {
					return fmt.Errorf("failed to install modules: %w", err)
				}
			}

			// Report results
			successful := 0
			for _, result := range installResults {
				if result.Success {
					successful++
					ui.Success("Installed %s v%s", result.ModuleName, result.Version)
				} else {
					errorMsg := "unknown error"
					if len(result.Errors) > 0 {
						errorMsg = result.Errors[0]
					}
					ui.Error("Failed to install %s: %s", result.ModuleName, errorMsg)
				}
			}

			// Generate type definitions if requested
			if generateTypeDefs && successful > 0 {
				ui.Status("Generating type definitions for installed modules")

				for _, result := range installResults {
					if result.Success {
						err := generateTypeDefinitionForModule(result.ModuleName, ui)
						if err != nil {
							ui.Warning("Failed to generate type definitions for %s: %v", result.ModuleName, err)
						} else {
							ui.Info("Generated type definitions for %s", result.ModuleName)
						}
					}
				}
			}

			ui.Success("Module installation complete: %d successful, %d failed",
				successful, len(installResults)-successful)

			return nil
		},
	}

	// Add flags
	cmd.Flags().Bool("generate-types", true, "Generate type definitions after installation")
	cmd.Flags().Bool("skip-tests", false, "Skip running tests during installation")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.Flags().StringP("perl", "p", "", "Target Perl version")

	return cmd
}

// newModuleAnalyzeCommand creates a command to analyze module dependencies for type checking
func newModuleAnalyzeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "analyze [directory]",
		Short: "Analyze module dependencies for type checking",
		Long:  "Analyze a Perl project to identify module dependencies and their type definition availability",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ui := cli.GetUI(cmd)

			projectDir := "."
			if len(args) > 0 {
				projectDir = args[0]
			}

			ui.Status("Analyzing project dependencies")

			// Create project analyzer
			analyzer := &ProjectAnalyzer{
				RootDir: projectDir,
			}

			// Analyze dependencies
			dependencies, err := analyzer.AnalyzeDependencies()
			if err != nil {
				return fmt.Errorf("failed to analyze dependencies: %w", err)
			}

			// Display results
			ui.Header("Project Dependency Analysis")
			ui.Info("Found %d module dependencies", len(dependencies))

			for _, dep := range dependencies {
				status := "✓ Available"
				if !dep.HasTypeDefinitions {
					status = "✗ No type definitions"
				}
				ui.Printf("  %s - %s", dep.ModuleName, status)
			}

			// Suggest improvements
			missingTypeDefs := 0
			for _, dep := range dependencies {
				if !dep.HasTypeDefinitions {
					missingTypeDefs++
				}
			}

			if missingTypeDefs > 0 {
				ui.Warning("%d modules lack type definitions", missingTypeDefs)
				ui.Info("Use 'psc module typedefs --generate' to create type definitions")
			} else {
				ui.Success("All dependencies have type definitions available")
			}

			return nil
		},
	}
}

// newModuleTypeDefsCommand creates a command to manage type definitions for modules
func newModuleTypeDefsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "typedefs",
		Short: "Manage type definitions for modules",
		Long:  "Generate and manage type definitions for installed modules",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "generate [modules...]",
			Short: "Generate type definitions for modules",
			Args:  cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				ui := cli.GetUI(cmd)

				ui.Status("Generating type definitions for modules")

				successful := 0
				for _, moduleName := range args {
					err := generateTypeDefinitionForModule(moduleName, ui)
					if err != nil {
						ui.Error("Failed to generate type definitions for %s: %v", moduleName, err)
					} else {
						ui.Success("Generated type definitions for %s", moduleName)
						successful++
					}
				}

				ui.Info("Generated type definitions for %d/%d modules", successful, len(args))
				return nil
			},
		},
		&cobra.Command{
			Use:   "check [modules...]",
			Short: "Check type definition availability",
			Args:  cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				ui := cli.GetUI(cmd)

				for _, moduleName := range args {
					available := checkTypeDefinitionAvailability(moduleName)
					if available {
						ui.Success("%s - Type definitions available", moduleName)
					} else {
						ui.Warning("%s - No type definitions found", moduleName)
					}
				}

				return nil
			},
		},
	)

	return cmd
}

// Helper functions

// generateTypeDefinitionForModule generates type definitions for a module
func generateTypeDefinitionForModule(moduleName string, ui *ui.Output) error {
	// This would integrate with PSC's type definition generation system
	// For now, this is a placeholder that would call the actual generator
	ui.Debug("Generating type definitions for %s", moduleName)

	// In a real implementation, this would:
	// 1. Find the installed module
	// 2. Parse its interface
	// 3. Generate appropriate type definitions
	// 4. Store them in the type definition system

	return nil
}

// checkTypeDefinitionAvailability checks if type definitions exist for a module
func checkTypeDefinitionAvailability(moduleName string) bool {
	// This would check PSC's type definition storage
	// For now, return false as a placeholder
	return false
}

// DependencyInfo represents information about a module dependency
type DependencyInfo struct {
	ModuleName         string
	Version            string
	HasTypeDefinitions bool
	TypeDefPath        string
}

// AnalyzeDependencies analyzes project dependencies
func (pa *ProjectAnalyzer) AnalyzeDependencies() ([]*DependencyInfo, error) {
	// This would integrate with the existing ProjectAnalyzer
	// For now, return empty slice as placeholder
	return []*DependencyInfo{}, nil
}
