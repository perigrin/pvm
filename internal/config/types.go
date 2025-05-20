// ABOUTME: Configuration types for the PVM Ecosystem
// ABOUTME: Defines structs matching the TOML configuration format

package config

import (
	"strings"
)

// Config represents the root configuration object
type Config struct {
	// PVM specific configuration
	PVM *PVMConfig `toml:"pvm"`

	// PVX specific configuration
	PVX *PVXConfig `toml:"pvx"`

	// PVI specific configuration
	PVI *PVIConfig `toml:"pvi"`

	// PSC specific configuration
	PSC *PSCConfig `toml:"psc"`
}

// PVMConfig represents configuration for the Perl Version Manager
type PVMConfig struct {
	// DefaultPerl is the default Perl version to use when none is specified
	DefaultPerl string `toml:"default_perl"`

	// BuildJobs specifies the number of parallel jobs to use during build
	BuildJobs int `toml:"build_jobs"`

	// DownloadMirror specifies the mirror to use for downloading Perl source
	DownloadMirror string `toml:"download_mirror"`

	// PatchesDir specifies the directory containing patches to apply during build
	PatchesDir string `toml:"patches_dir"`

	// Compiler specifies the compiler to use for building Perl
	Compiler string `toml:"compiler"`

	// RunTests specifies whether to run tests during Perl installation
	RunTests bool `toml:"run_tests"`

	// VersionAliases maps aliases to actual version strings
	VersionAliases map[string]string `toml:"version_aliases"`
}

// PVXConfig represents configuration for the Perl Version eXecutor
type PVXConfig struct {
	// CacheModules specifies whether to cache modules for faster startup
	CacheModules bool `toml:"cache_modules"`

	// CleanupAfter specifies whether to clean up temporary files after execution
	CleanupAfter bool `toml:"cleanup_after"`

	// IsolationLevel specifies the level of isolation for script execution
	// Valid values: "none", "low", "medium", "high"
	IsolationLevel string `toml:"isolation_level"`

	// AlwaysInstallDeps specifies whether to automatically install missing dependencies
	AlwaysInstallDeps bool `toml:"always_install_deps"`

	// Timeout specifies the maximum execution time in seconds
	Timeout int `toml:"timeout"`

	// MaxMemory specifies the maximum memory usage (e.g., "512MB")
	MaxMemory string `toml:"max_memory"`

	// IsolationReadOnlyPaths specifies paths that should be read-only in high isolation mode
	IsolationReadOnlyPaths []string `toml:"isolation_ro_paths"`

	// IsolationReadWritePaths specifies paths that should be read-write in high isolation mode
	IsolationReadWritePaths []string `toml:"isolation_rw_paths"`

	// IsolatedOutput specifies whether to create an isolated output directory
	IsolatedOutput bool `toml:"isolated_output"`

	// SaveOutputDir specifies where to save isolated output files
	SaveOutputDir string `toml:"save_output_dir"`

	// PreserveEnvVars specifies which environment variables to preserve in isolation
	PreserveEnvVars []string `toml:"preserve_env_vars"`

	// AdditionalModulePaths specifies additional module paths to add to PERL5LIB
	AdditionalModulePaths []string `toml:"additional_module_paths"`

	// CustomModulePath specifies a custom module installation path
	CustomModulePath string `toml:"custom_module_path"`
}

// PVIConfig represents configuration for the Perl Version Installer
type PVIConfig struct {
	// PreferredInstaller specifies the preferred installation method
	// Valid values: "auto", "cpanm", "cpan", "cpm"
	PreferredInstaller string `toml:"preferred_installer"`

	// DefaultMirror specifies the CPAN mirror to use
	DefaultMirror string `toml:"default_mirror"`

	// TestDuringInstall specifies whether to run tests during module installation
	TestDuringInstall bool `toml:"test_during_install"`

	// CacheModules specifies whether to cache modules for faster installation
	CacheModules bool `toml:"cache_modules"`

	// ForceReinstall specifies whether to force reinstallation of modules
	ForceReinstall bool `toml:"force_reinstall"`

	// CheckSignatures specifies whether to check module signatures
	CheckSignatures bool `toml:"check_signatures"`
}

// PSCConfig represents configuration for the Perl Script Compiler
type PSCConfig struct {
	// TypeDefinitionsPath specifies the path to type definitions
	TypeDefinitionsPath string `toml:"type_definitions_path"`

	// StrictMode enables more rigorous type checking
	StrictMode bool `toml:"strict_mode"`

	// WatchExclude specifies patterns to exclude from watch mode
	WatchExclude []string `toml:"watch_exclude"`

	// GenerateMissingTypes specifies whether to generate missing type definitions
	GenerateMissingTypes bool `toml:"generate_missing_types"`

	// CheckBeforeRun specifies whether to check types before running a script
	CheckBeforeRun bool `toml:"check_before_run"`
}

// NewDefaultConfig creates a new configuration with default values
func NewDefaultConfig() *Config {
	return &Config{
		PVM: &PVMConfig{
			DefaultPerl:    "5.38.0",
			BuildJobs:      4,
			DownloadMirror: "https://www.cpan.org/src/5.0",
			RunTests:       true,
			VersionAliases: map[string]string{
				"latest": "5.38.0",
				"stable": "5.36.0",
			},
		},
		PVX: &PVXConfig{
			CacheModules:            true,
			CleanupAfter:            true,
			IsolationLevel:          "medium",
			AlwaysInstallDeps:       true,
			Timeout:                 300,
			MaxMemory:               "512MB",
			IsolationReadOnlyPaths:  []string{"/usr", "/bin", "/lib"},
			IsolationReadWritePaths: []string{},
			IsolatedOutput:          false,
			SaveOutputDir:           "$PWD/output",
			PreserveEnvVars:         []string{"TERM", "DISPLAY"},
			AdditionalModulePaths:   []string{},
			CustomModulePath:        "",
		},
		PVI: &PVIConfig{
			PreferredInstaller: "auto",
			DefaultMirror:      "https://cpan.metacpan.org",
			TestDuringInstall:  false,
			CacheModules:       true,
			CheckSignatures:    true,
		},
		PSC: &PSCConfig{
			TypeDefinitionsPath:  "$XDG_DATA_HOME/pvm/type_definitions",
			StrictMode:           false,
			WatchExclude:         []string{"**/node_modules/**", "**/.git/**", "**/local/**"},
			GenerateMissingTypes: true,
			CheckBeforeRun:       true,
		},
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() []error {
	var errors []error

	// Add PVM validation
	if c.PVM != nil {
		errors = append(errors, c.PVM.Validate()...)
	}

	// Add PVX validation
	if c.PVX != nil {
		errors = append(errors, c.PVX.Validate()...)
	}

	// Add PVI validation
	if c.PVI != nil {
		errors = append(errors, c.PVI.Validate()...)
	}

	// Add PSC validation
	if c.PSC != nil {
		errors = append(errors, c.PSC.Validate()...)
	}

	return errors
}

// Validate checks if the PVM configuration is valid
func (c *PVMConfig) Validate() []error {
	var errors []error

	// Validate BuildJobs
	if c.BuildJobs < 1 {
		errors = append(errors, ValidateError("BuildJobs must be at least 1"))
	}

	// Validate DownloadMirror
	if c.DownloadMirror == "" {
		errors = append(errors, ValidateError("DownloadMirror cannot be empty"))
	}

	return errors
}

// Validate checks if the PVX configuration is valid
func (c *PVXConfig) Validate() []error {
	var errors []error

	// Validate IsolationLevel
	validLevels := map[string]bool{
		"none":   true,
		"low":    true,
		"medium": true,
		"high":   true,
	}

	if c.IsolationLevel != "" && !validLevels[c.IsolationLevel] {
		errors = append(errors, ValidateError("IsolationLevel must be one of: none, low, medium, high"))
	}

	// Validate Timeout
	if c.Timeout < 0 {
		errors = append(errors, ValidateError("Timeout cannot be negative"))
	}

	// Validate Read-Only Paths
	for _, path := range c.IsolationReadOnlyPaths {
		if path == "" {
			errors = append(errors, ValidateError("IsolationReadOnlyPaths cannot contain empty paths"))
			break
		}
	}

	// Validate Read-Write Paths
	for _, path := range c.IsolationReadWritePaths {
		if path == "" {
			errors = append(errors, ValidateError("IsolationReadWritePaths cannot contain empty paths"))
			break
		}
	}

	// Validate SaveOutputDir if IsolatedOutput is true
	if c.IsolatedOutput && c.SaveOutputDir == "" {
		errors = append(errors, ValidateError("SaveOutputDir must be specified when IsolatedOutput is enabled"))
	}

	// Validate Custom Module Path
	if c.CustomModulePath != "" {
		// Check if it contains valid path characters (simplified version)
		if strings.Contains(c.CustomModulePath, "\\") && !strings.Contains(c.CustomModulePath, "\\\\") {
			errors = append(errors, ValidateError("CustomModulePath contains invalid character: '\\'"))
		}
	}

	return errors
}

// Validate checks if the PVI configuration is valid
func (c *PVIConfig) Validate() []error {
	var errors []error

	// Validate PreferredInstaller
	validInstallers := map[string]bool{
		"auto":  true,
		"cpanm": true,
		"cpan":  true,
		"cpm":   true,
	}

	if c.PreferredInstaller != "" && !validInstallers[c.PreferredInstaller] {
		errors = append(errors, ValidateError("PreferredInstaller must be one of: auto, cpanm, cpan, cpm"))
	}

	// Validate DefaultMirror
	if c.DefaultMirror == "" {
		errors = append(errors, ValidateError("DefaultMirror cannot be empty"))
	}

	return errors
}

// Validate checks if the PSC configuration is valid
func (c *PSCConfig) Validate() []error {
	var errors []error

	// Validate TypeDefinitionsPath
	if c.TypeDefinitionsPath == "" {
		errors = append(errors, ValidateError("TypeDefinitionsPath cannot be empty"))
	}

	return errors
}

// ValidateError creates a validation error
type ValidateError string

// Error implements the error interface
func (e ValidateError) Error() string {
	return string(e)
}
