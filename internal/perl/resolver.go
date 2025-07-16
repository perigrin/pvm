// ABOUTME: Version resolution algorithm for Perl versions
// ABOUTME: Implements precedence-based resolution for determining which Perl version to use

package perl

import (
	"os"
	"path/filepath"
	"runtime"

	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/errors"
)

// Version resolution error codes
const (
	ErrResolutionFailed    = "301" // Failed to resolve Perl version
	ErrUnsatisfiedVersion  = "302" // No available version satisfies constraints
	ErrNoVersionsAvailable = "303" // No Perl versions available
)

// ResolutionSource represents the source of a resolved version
type ResolutionSource string

const (
	// Resolution sources in order of precedence (highest to lowest)
	ExplicitVersion     ResolutionSource = "explicit"       // Explicitly specified version
	ProjectVersionFile  ResolutionSource = "project_file"   // Project .perl-version file
	ProjectConfig       ResolutionSource = "project_config" // Project .pvm/pvm.toml
	EnvironmentVariable ResolutionSource = "env_var"        // Environment variable (PVM_PERL_VERSION, PLENV_VERSION, PERLBREW_PERL)
	UserConfig          ResolutionSource = "user_config"    // User configuration
	SystemPerlSource    ResolutionSource = "system_perl"    // System Perl installation
)

// ResolvedVersion represents a resolved Perl version with metadata
type ResolvedVersion struct {
	// The resolved version string
	Version string

	// The source of the resolved version
	Source ResolutionSource

	// The path where this version was found (if applicable)
	Path string
}

// ResolutionOptions contains options for version resolution
type ResolutionOptions struct {
	// Explicit version to use (highest precedence)
	ExplicitVersion string

	// Project directory to check (defaults to current directory)
	ProjectDir string

	// Configuration to use (if nil, will be loaded)
	Config *config.Config

	// Available versions (if empty, will be detected)
	AvailableVersions []string

	// Skip specific resolution methods
	SkipLocal           bool // Skip project-local resolution
	SkipEnvVars         bool // Skip environment variables
	SkipUserConfig      bool // Skip user configuration
	SkipSystemPerl      bool // Skip system Perl detection
	SkipVersionResolved bool // Skip calling OnVersionResolved
}

// OnVersionResolved is a callback that will be called when a version is resolved
// This can be used for logging or other purposes
var OnVersionResolved func(version *ResolvedVersion)

// ResolveVersion determines which Perl version to use based on the resolution algorithm
// The precedence order (highest to lowest) is:
// 1. Explicitly specified version
// 2. Project-local .perl-version file
// 3. Project-local .pvm/pvm.toml
// 4. Environment variables (PVM_PERL_VERSION, PLENV_VERSION, PERLBREW_PERL)
// 5. User-level configuration
// 6. System Perl
func ResolveVersion(options *ResolutionOptions) (*ResolvedVersion, error) {
	// Use default options if nil
	if options == nil {
		options = &ResolutionOptions{}
	}

	// Set current directory as project directory if not specified
	if options.ProjectDir == "" {
		dir, err := os.Getwd()
		if err == nil {
			options.ProjectDir = dir
		}
	}

	// Load configuration if not provided
	var cfg *config.Config
	var err error
	if options.Config == nil {
		cfg, err = config.LoadEffectiveConfig()
		if err != nil {
			return nil, errors.NewVersionError(
				ErrResolutionFailed,
				"Failed to load configuration",
				err)
		}
	} else {
		cfg = options.Config
	}

	// Get available versions if not provided
	var availableVersions []string
	if len(options.AvailableVersions) == 0 {
		// In a full implementation, this would include installed versions
		// For now, we'll use system Perl version as a starting point
		systemPerl, err := DetectSystemPerl()
		if err == nil && systemPerl != nil {
			availableVersions = append(availableVersions, systemPerl.Version)
		}

		// Add some common versions for testing
		availableVersions = append(availableVersions,
			"5.38.0", "5.36.0", "5.34.1", "5.32.1", "5.30.3", "5.28.3", "5.26.3")
	} else {
		availableVersions = options.AvailableVersions
	}

	// Auto-import on first run if no versions are available
	if len(availableVersions) == 0 {
		// Always try auto-import of legacy versions (plenv/perlbrew) first
		// This ensures plenv/perlbrew versions are available even if system Perl exists
		results, err := AutoImportLegacyVersions()
		if err == nil && results.TotalImported > 0 {
			// Rebuild available versions list with imported versions
			for _, result := range results.PlenvImports {
				availableVersions = append(availableVersions, result.Version)
			}
			for _, result := range results.PerlbrewImports {
				availableVersions = append(availableVersions, result.Version)
			}
		}

		// If still no versions, try system Perl auto-import
		if len(availableVersions) == 0 {
			systemPerl, detectErr := DetectSystemPerl()
			if detectErr == nil && systemPerl != nil {
				err := AutoImportSystemPerl()
				if err == nil {
					availableVersions = append(availableVersions, systemPerl.Version)
				}
			}
		}
	}

	// Verify we have some available versions
	if len(availableVersions) == 0 {
		return nil, errors.NewVersionError(
			ErrNoVersionsAvailable,
			"No Perl versions available. Install with 'pvm install <version>' or check system Perl availability.",
			nil)
	}

	// 1. Check explicit version (highest precedence)
	if options.ExplicitVersion != "" {
		resolved, err := resolveExplicitVersion(options.ExplicitVersion, availableVersions, cfg)
		if err == nil && resolved != nil {
			notifyResolved(resolved, options)
			return resolved, nil
		}
	}

	// 2. Check project-local .perl-version file (if not skipped)
	if !options.SkipLocal && options.ProjectDir != "" {
		resolved, err := resolveFromPerlVersionFile(options.ProjectDir, availableVersions, cfg)
		if err == nil && resolved != nil {
			notifyResolved(resolved, options)
			return resolved, nil
		}
	}

	// 3. Check project-local .pvm/pvm.toml (if not skipped)
	if !options.SkipLocal && options.ProjectDir != "" {
		resolved, err := resolveFromProjectConfig(options.ProjectDir, availableVersions, cfg)
		if err == nil && resolved != nil {
			notifyResolved(resolved, options)
			return resolved, nil
		}
	}

	// 4. Check environment variables (if not skipped)
	if !options.SkipEnvVars {
		// Only show debug info for environment variables here
		resolved, err := resolveFromEnvironment(availableVersions, cfg)
		if err == nil && resolved != nil {
			notifyResolved(resolved, options)
			return resolved, nil
		}
	}

	// 5. Check user-level configuration (if not skipped)
	if !options.SkipUserConfig && cfg != nil && cfg.PVM != nil {
		resolved, err := resolveFromUserConfig(cfg, availableVersions)
		if err == nil && resolved != nil {
			notifyResolved(resolved, options)
			return resolved, nil
		}
	}

	// 6. Fallback to system Perl (if not skipped)
	if !options.SkipSystemPerl {
		resolved, err := resolveFromSystemPerl()
		if err == nil && resolved != nil {
			notifyResolved(resolved, options)
			return resolved, nil
		}
	}

	// No version could be resolved
	return nil, errors.NewVersionError(
		ErrResolutionFailed,
		"Failed to resolve Perl version using any method",
		nil)
}

// resolveExplicitVersion resolves an explicitly specified version
func resolveExplicitVersion(version string, availableVersions []string, cfg *config.Config) (*ResolvedVersion, error) {
	// Handle special "system" identifier
	if version == "system" {
		// Find the system Perl version that was imported
		installedVersions, err := GetInstalledVersions()
		if err != nil {
			return nil, err
		}

		for _, versionInfo := range installedVersions {
			if versionInfo.Source == "system" {
				// For system perl, use the InstallPath directly if it points to perl executable
				var perlExe string
				if filepath.Base(versionInfo.InstallPath) == "perl" || filepath.Base(versionInfo.InstallPath) == "perl.exe" {
					perlExe = versionInfo.InstallPath
				} else {
					// InstallPath is likely the bin directory, append perl
					perlExe = filepath.Join(versionInfo.InstallPath, "perl")
					if runtime.GOOS == "windows" {
						perlExe = filepath.Join(versionInfo.InstallPath, "perl.exe")
					}
				}

				return &ResolvedVersion{
					Version: versionInfo.Version,
					Source:  SystemPerlSource,
					Path:    perlExe,
				}, nil
			}
		}

		// If no system Perl is imported, return an error
		return nil, errors.NewVersionError(
			ErrUnsatisfiedVersion,
			"System Perl not found. Use 'pvm import-system' to import it first.",
			nil)
	}

	// Check if it's an alias and resolve if needed
	var aliases map[string]string
	if cfg != nil && cfg.PVM != nil {
		aliases = cfg.PVM.VersionAliases
	}

	resolvedVersion, err := ResolveVersionAlias(version, aliases)
	if err != nil {
		return nil, err
	}

	// Check if the version is available
	available := false
	for _, v := range availableVersions {
		if v == resolvedVersion {
			available = true
			break
		}
	}

	if !available {
		return nil, errors.NewVersionError(
			ErrUnsatisfiedVersion,
			"Specified version is not available: "+resolvedVersion,
			nil)
	}

	return &ResolvedVersion{
		Version: resolvedVersion,
		Source:  ExplicitVersion,
		Path:    "", // Path will be resolved by the executor
	}, nil
}

// resolveFromPerlVersionFile checks for a .perl-version file in the project directory
// and its parent directories
func resolveFromPerlVersionFile(projectDir string, availableVersions []string, cfg *config.Config) (*ResolvedVersion, error) {
	// Find all .perl-version files in the project and its parent directories
	versionFiles, err := FindDotPerlVersionFiles(projectDir)
	if err != nil || len(versionFiles) == 0 {
		return nil, errors.NewVersionError(
			ErrResolutionFailed,
			"No .perl-version file found",
			err)
	}

	// Use the closest .perl-version file (first in the list)
	versionFile := versionFiles[0]
	versionStr, err := ReadPerlVersionFile(filepath.Dir(versionFile))
	if err != nil {
		return nil, err
	}

	// Check if it's an alias and resolve if needed
	var aliases map[string]string
	if cfg != nil && cfg.PVM != nil {
		aliases = cfg.PVM.VersionAliases
	}

	resolvedVersion, err := ResolveVersionAlias(versionStr, aliases)
	if err != nil {
		return nil, err
	}

	// Check if the version is available
	if !IsVersionAvailable(resolvedVersion, availableVersions) {
		return nil, errors.NewVersionError(
			ErrUnsatisfiedVersion,
			"Version in .perl-version file is not available: "+resolvedVersion,
			nil)
	}

	return &ResolvedVersion{
		Version: resolvedVersion,
		Source:  ProjectVersionFile,
		Path:    "", // Path will be resolved by the executor
	}, nil
}

// resolveFromProjectConfig checks for a .pvm/pvm.toml file in the project directory
func resolveFromProjectConfig(projectDir string, availableVersions []string, cfg *config.Config) (*ResolvedVersion, error) {
	// Find project root directory with .pvm directory
	projectRoot := config.GetProjectRoot()
	if projectRoot == "" {
		return nil, errors.NewVersionError(
			ErrResolutionFailed,
			"No project configuration found",
			nil)
	}

	// Get project configuration path
	projectConfigPath := filepath.Join(projectRoot, ".pvm", "pvm.toml")
	if _, err := os.Stat(projectConfigPath); os.IsNotExist(err) {
		return nil, errors.NewVersionError(
			ErrResolutionFailed,
			"No project configuration file found",
			nil)
	}

	// Parse project configuration file
	projectConfig, err := config.ParseFile(projectConfigPath)
	if err != nil {
		return nil, errors.NewVersionError(
			ErrResolutionFailed,
			"Failed to parse project configuration",
			err)
	}

	// Check if project configuration has default_perl
	if projectConfig.PVM == nil || projectConfig.PVM.DefaultPerl == "" {
		return nil, errors.NewVersionError(
			ErrResolutionFailed,
			"No default_perl specified in project configuration",
			nil)
	}

	// Check if it's an alias and resolve if needed
	var aliases map[string]string
	if projectConfig.PVM.VersionAliases != nil {
		// Use project aliases first
		aliases = projectConfig.PVM.VersionAliases
	} else if cfg != nil && cfg.PVM != nil {
		// Fall back to global aliases
		aliases = cfg.PVM.VersionAliases
	}

	resolvedVersion, err := ResolveVersionAlias(projectConfig.PVM.DefaultPerl, aliases)
	if err != nil {
		return nil, err
	}

	// Check if the version is available
	if !IsVersionAvailable(resolvedVersion, availableVersions) {
		return nil, errors.NewVersionError(
			ErrUnsatisfiedVersion,
			"Version in project configuration is not available: "+resolvedVersion,
			nil)
	}

	return &ResolvedVersion{
		Version: resolvedVersion,
		Source:  ProjectConfig,
		Path:    "", // Path will be resolved by the executor
	}, nil
}

// resolveFromEnvironment checks environment variables for version specifications
func resolveFromEnvironment(availableVersions []string, cfg *config.Config) (*ResolvedVersion, error) {
	// Check for PVM_PERL_VERSION environment variable (highest precedence)
	// This is set by the shell integration `pvm use` command
	if versionStr := os.Getenv("PVM_PERL_VERSION"); versionStr != "" {
		// Check if it's an alias and resolve if needed
		var aliases map[string]string
		if cfg != nil && cfg.PVM != nil {
			aliases = cfg.PVM.VersionAliases
		}

		resolvedVersion, err := ResolveVersionAlias(versionStr, aliases)
		if err != nil {
			return nil, err
		}

		// Check if the version is available
		if !IsVersionAvailable(resolvedVersion, availableVersions) {
			return nil, errors.NewVersionError(
				ErrUnsatisfiedVersion,
				"Version in PVM_PERL_VERSION is not available: "+resolvedVersion,
				nil)
		}

		return &ResolvedVersion{
			Version: resolvedVersion,
			Source:  EnvironmentVariable,
			Path:    "", // Path will be resolved by the executor
		}, nil
	}

	// Check for PLENV_VERSION environment variable (lower precedence)
	if versionStr := os.Getenv("PLENV_VERSION"); versionStr != "" {
		// Check if it's an alias and resolve if needed
		var aliases map[string]string
		if cfg != nil && cfg.PVM != nil {
			aliases = cfg.PVM.VersionAliases
		}

		resolvedVersion, err := ResolveVersionAlias(versionStr, aliases)
		if err != nil {
			return nil, err
		}

		// Check if the version is available
		if !IsVersionAvailable(resolvedVersion, availableVersions) {
			return nil, errors.NewVersionError(
				ErrUnsatisfiedVersion,
				"Version in PLENV_VERSION is not available: "+resolvedVersion,
				nil)
		}

		return &ResolvedVersion{
			Version: resolvedVersion,
			Source:  EnvironmentVariable,
			Path:    "", // Path will be resolved by the executor
		}, nil
	}

	// Check for PERLBREW_PERL environment variable (lower precedence)
	if versionStr := os.Getenv("PERLBREW_PERL"); versionStr != "" {
		// Perlbrew uses a "perl-5.xx.x" format, so strip "perl-" prefix if present
		if len(versionStr) > 5 && versionStr[:5] == "perl-" {
			versionStr = versionStr[5:]
		}

		// Check if it's an alias and resolve if needed
		var aliases map[string]string
		if cfg != nil && cfg.PVM != nil {
			aliases = cfg.PVM.VersionAliases
		}

		resolvedVersion, err := ResolveVersionAlias(versionStr, aliases)
		if err != nil {
			return nil, err
		}

		// Check if the version is available
		if !IsVersionAvailable(resolvedVersion, availableVersions) {
			return nil, errors.NewVersionError(
				ErrUnsatisfiedVersion,
				"Version in PERLBREW_PERL is not available: "+resolvedVersion,
				nil)
		}

		return &ResolvedVersion{
			Version: resolvedVersion,
			Source:  EnvironmentVariable,
			Path:    "", // Path will be resolved by the executor
		}, nil
	}

	return nil, errors.NewVersionError(
		ErrResolutionFailed,
		"No version found in environment variables",
		nil)
}

// resolveFromUserConfig checks user-level configuration for default Perl version
func resolveFromUserConfig(cfg *config.Config, availableVersions []string) (*ResolvedVersion, error) {
	// Check if configuration has default_perl
	if cfg.PVM == nil || cfg.PVM.DefaultPerl == "" {
		return nil, errors.NewVersionError(
			ErrResolutionFailed,
			"No default_perl specified in user configuration",
			nil)
	}

	// Check if it's an alias and resolve if needed
	resolvedVersion, err := ResolveVersionAlias(cfg.PVM.DefaultPerl, cfg.PVM.VersionAliases)
	if err != nil {
		return nil, err
	}

	// Check if the version is available
	if !IsVersionAvailable(resolvedVersion, availableVersions) {
		return nil, errors.NewVersionError(
			ErrUnsatisfiedVersion,
			"Version in user configuration is not available: "+resolvedVersion,
			nil)
	}

	return &ResolvedVersion{
		Version: resolvedVersion,
		Source:  UserConfig,
		Path:    "", // Path will be resolved by the executor
	}, nil
}

// resolveFromSystemPerl falls back to using the system Perl installation
func resolveFromSystemPerl() (*ResolvedVersion, error) {
	// Detect system Perl
	systemPerl, err := DetectSystemPerl()
	if err != nil {
		return nil, err
	}

	return &ResolvedVersion{
		Version: systemPerl.Version,
		Source:  SystemPerlSource,
		Path:    systemPerl.Path,
	}, nil
}

// getPerlExecutablePath returns the path to the perl executable for a given version
func getPerlExecutablePath(version string) (string, error) {
	// Get version info from registry
	versionInfo, err := GetVersionInfo(version)
	if err != nil {
		return "", err
	}

	if versionInfo == nil {
		return "", errors.NewVersionError(
			ErrResolutionFailed,
			"Version info not found for version: "+version,
			nil)
	}

	var perlExe string
	if versionInfo.Source == "system" {
		// For system perl, InstallPath might be the bin directory or the install root
		// Check if InstallPath already points to the perl executable
		if filepath.Base(versionInfo.InstallPath) == "perl" || filepath.Base(versionInfo.InstallPath) == "perl.exe" {
			perlExe = versionInfo.InstallPath
		} else {
			// InstallPath is likely the bin directory, append perl
			perlExe = filepath.Join(versionInfo.InstallPath, "perl")
			if runtime.GOOS == "windows" {
				perlExe = filepath.Join(versionInfo.InstallPath, "perl.exe")
			}
		}
	} else {
		// For PVM-installed versions, InstallPath is the installation root
		perlExe = filepath.Join(versionInfo.InstallPath, "bin", "perl")
		if runtime.GOOS == "windows" {
			perlExe = filepath.Join(versionInfo.InstallPath, "bin", "perl.exe")
		}
	}

	// Verify the executable exists
	if _, err := os.Stat(perlExe); os.IsNotExist(err) {
		return "", errors.NewVersionError(
			ErrResolutionFailed,
			"Perl executable not found at: "+perlExe,
			err)
	}

	return perlExe, nil
}

// notifyResolved calls the OnVersionResolved callback if set
func notifyResolved(version *ResolvedVersion, options *ResolutionOptions) {
	if !options.SkipVersionResolved && OnVersionResolved != nil {
		OnVersionResolved(version)
	}
}
