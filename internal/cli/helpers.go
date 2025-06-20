// ABOUTME: Common CLI command helpers and utilities
// ABOUTME: Provides reusable patterns for command setup and error handling

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/modules"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/project"
)

// Error codes for CLI helpers
const (
	ErrHelperConfigFailed   = "CLI-4701" // Configuration loading failed
	ErrHelperProviderFailed = "CLI-4702" // Provider setup failed
	ErrHelperProjectFailed  = "CLI-4703" // Project context required but not available
	ErrHelperPerlNotFound   = "CLI-4704" // Perl interpreter not found
	ErrHelperInvalidFlags   = "CLI-4705" // Invalid command line flags
)

// CommandEnvironment contains the common environment setup for commands
type CommandEnvironment struct {
	// Config is the loaded effective configuration
	Config *config.Config

	// InstallationEnv contains the provider and resolver
	InstallationEnv *modules.InstallationEnvironment

	// ProjectContext contains project information (if in project)
	ProjectContext *project.ProjectContext

	// PerlPath is the resolved Perl interpreter path
	PerlPath string

	// UI provides the command UI interface
	UI *ui.Output
}

// SetupCommandEnvironment sets up the common environment for PVI commands
// Note: InstallationEnv will be nil and should be set by the caller using the provider builder
func SetupCommandEnvironment(cmd *cobra.Command, requireProject bool) (*CommandEnvironment, error) {
	// Load configuration
	cfg, err := config.LoadEffectiveConfig()
	if err != nil {
		return nil, errors.NewSystemError(
			ErrHelperConfigFailed,
			"Failed to load configuration",
			err)
	}

	// Setup project context if required
	var projectCtx *project.ProjectContext
	if requireProject {
		projectCtx, err = project.GetCurrentProject()
		if err != nil {
			return nil, errors.NewSystemError(
				ErrHelperProjectFailed,
				"Failed to detect project context",
				err)
		}
		if !projectCtx.IsProject {
			return nil, errors.NewSystemError(
				ErrHelperProjectFailed,
				"Not in a project directory. Use 'pvm project init' to create a project",
				nil)
		}
	} else {
		// Try to get project context but don't fail if not available
		projectCtx, _ = project.GetCurrentProject()
	}

	// Get UI interface
	ui := GetUI(cmd)

	env := &CommandEnvironment{
		Config:         cfg,
		ProjectContext: projectCtx,
		UI:             ui,
		// InstallationEnv is nil - should be set by caller
	}

	return env, nil
}

// ResolvePerlPath resolves the Perl path for the command environment
func (env *CommandEnvironment) ResolvePerlPath(perlPath string) error {
	if perlPath != "" {
		env.PerlPath = perlPath
		return nil
	}

	// Try to resolve perl path from system
	resolved, err := perl.GetCurrentPerlPath()
	if err != nil {
		return errors.NewSystemError(
			ErrHelperPerlNotFound,
			"Failed to resolve Perl interpreter path",
			err)
	}

	env.PerlPath = resolved
	return nil
}

// SetInstallationEnvironment sets the installation environment for the command
func (env *CommandEnvironment) SetInstallationEnvironment(installEnv *modules.InstallationEnvironment) {
	env.InstallationEnv = installEnv
}

// CreateInstaller creates a module installer from the command environment
func (env *CommandEnvironment) CreateInstaller() *modules.Installer {
	if env.InstallationEnv == nil {
		return nil // Installation environment not set
	}

	return modules.NewInstaller(
		env.InstallationEnv.Provider,
		env.InstallationEnv.Resolver,
		nil, // tracker - would be set based on specific command needs
		nil, // logger - would be configured based on verbosity settings
	)
}

// ParseCommonInstallFlags extracts common installation flags from a command
func ParseCommonInstallFlags(cmd *cobra.Command) (*CommonInstallFlags, error) {
	flags := &CommonInstallFlags{}

	var err error
	if flags.Source, err = cmd.Flags().GetString("source"); err != nil {
		return nil, errors.NewSystemError(ErrHelperInvalidFlags, "Invalid source flag", err)
	}
	if flags.NoCache, err = cmd.Flags().GetBool("no-cache"); err != nil {
		return nil, errors.NewSystemError(ErrHelperInvalidFlags, "Invalid no-cache flag", err)
	}
	if flags.SkipTests, err = cmd.Flags().GetBool("skip-tests"); err != nil {
		return nil, errors.NewSystemError(ErrHelperInvalidFlags, "Invalid skip-tests flag", err)
	}
	if flags.Force, err = cmd.Flags().GetBool("force"); err != nil {
		return nil, errors.NewSystemError(ErrHelperInvalidFlags, "Invalid force flag", err)
	}
	if flags.Verbose, err = cmd.Flags().GetBool("verbose"); err != nil {
		return nil, errors.NewSystemError(ErrHelperInvalidFlags, "Invalid verbose flag", err)
	}
	if flags.SkipDeps, err = cmd.Flags().GetBool("skip-deps"); err != nil {
		return nil, errors.NewSystemError(ErrHelperInvalidFlags, "Invalid skip-deps flag", err)
	}
	if flags.InstallDir, err = cmd.Flags().GetString("install-dir"); err != nil {
		return nil, errors.NewSystemError(ErrHelperInvalidFlags, "Invalid install-dir flag", err)
	}
	if flags.Version, err = cmd.Flags().GetString("version"); err != nil {
		return nil, errors.NewSystemError(ErrHelperInvalidFlags, "Invalid version flag", err)
	}
	if flags.PerlPath, err = cmd.Flags().GetString("perl-path"); err != nil {
		return nil, errors.NewSystemError(ErrHelperInvalidFlags, "Invalid perl-path flag", err)
	}

	return flags, nil
}

// CommonInstallFlags contains the common installation flags
type CommonInstallFlags struct {
	Source     string
	NoCache    bool
	SkipTests  bool
	Force      bool
	Verbose    bool
	SkipDeps   bool
	InstallDir string
	Version    string
	PerlPath   string
}

// ToInstallOptions converts common flags to InstallOptions
func (f *CommonInstallFlags) ToInstallOptions(perlPath string) modules.InstallOptions {
	return modules.InstallOptions{
		PerlPath:          perlPath,
		VersionConstraint: f.Version,
		InstallDir:        f.InstallDir,
		Force:             f.Force,
		RunTests:          !f.SkipTests,
		SkipDependencies:  f.SkipDeps,
		Verbose:           f.Verbose,
		Cleanup:           true, // Always clean up by default
	}
}

// DisplayInstallationResults shows installation results in a standardized format
func DisplayInstallationResults(ui *ui.Output, results []*modules.InstallResult) {
	if len(results) == 0 {
		ui.Warning("No installation results to display")
		return
	}

	successful := 0
	failed := 0
	totalDuration := int64(0)

	for _, result := range results {
		totalDuration += result.Duration.Nanoseconds()

		if result.Success {
			successful++
			ui.Success("✓ Successfully installed %s", result.ModuleName)
			if result.Version != "" {
				ui.Info("  Version: %s", result.Version)
			}
			if len(result.Dependencies) > 0 {
				ui.Info("  Dependencies: %d installed", len(result.Dependencies))
			}
			if len(result.Warnings) > 0 {
				for _, warning := range result.Warnings {
					ui.Warning("  Warning: %s", warning)
				}
			}
		} else {
			failed++
			ui.Error("✗ Failed to install %s", result.ModuleName)
			for _, err := range result.Errors {
				ui.Error("  Error: %s", err)
			}
		}

		if result.Duration > 0 {
			ui.Debug("  Duration: %v", result.Duration)
		}
	}

	// Summary
	ui.Info("")
	ui.Info("Installation Summary:")
	ui.Info("  Total modules: %d", len(results))
	ui.Info("  Successful: %d", successful)
	if failed > 0 {
		ui.Info("  Failed: %d", failed)
	}
	if totalDuration > 0 {
		avgDuration := totalDuration / int64(len(results))
		ui.Info("  Average time per module: %v", avgDuration)
	}
}

// DisplayModuleList shows a list of modules in a standardized format
func DisplayModuleList(ui *ui.Output, modules []*modules.InstalledModule, format string) {
	if len(modules) == 0 {
		ui.Info("No modules found")
		return
	}

	switch format {
	case "table":
		ui.Info("%-30s %-15s %s", "Module", "Version", "Path")
		ui.Info("%-30s %-15s %s", "------", "-------", "----")
		for _, mod := range modules {
			path := mod.Path
			if len(path) > 50 {
				path = "..." + path[len(path)-47:]
			}
			ui.Info("%-30s %-15s %s", mod.Name, mod.Version, path)
		}
	case "list":
		for _, mod := range modules {
			if mod.Version != "" {
				ui.Info("%s (%s)", mod.Name, mod.Version)
			} else {
				ui.Info("%s", mod.Name)
			}
		}
	case "json":
		// Would output JSON format - simplified for this implementation
		ui.Info("JSON output not implemented in this helper")
	default:
		// Default to simple list
		for _, mod := range modules {
			ui.Info("%s", mod.Name)
		}
	}

	ui.Info("")
	ui.Info("Total: %d modules", len(modules))
}

// ValidateModuleNames performs basic validation on module names
func ValidateModuleNames(modules []string) error {
	if len(modules) == 0 {
		return errors.NewSystemError(
			ErrHelperInvalidFlags,
			"No modules specified",
			nil)
	}

	for _, module := range modules {
		if module == "" {
			return errors.NewSystemError(
				ErrHelperInvalidFlags,
				"Empty module name provided",
				nil)
		}

		// Basic module name validation - could be enhanced
		if len(module) > 255 {
			return errors.NewSystemError(
				ErrHelperInvalidFlags,
				fmt.Sprintf("Module name too long: %s", module),
				nil)
		}
	}

	return nil
}
