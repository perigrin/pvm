// ABOUTME: Configuration types for the PVM Ecosystem
// ABOUTME: Defines structs matching the TOML configuration format

package config

import (
	"strings"
	"time"
)

// Config represents the root configuration object
type Config struct {
	// PVM specific configuration
	PVM *PVMConfig `toml:"pvm" json:"pvm"`

	// PVX specific configuration
	PVX *PVXConfig `toml:"pvx" json:"pvx"`

	// PVI specific configuration
	PVI *PVIConfig `toml:"pvi" json:"pvi"`

	// PSC specific configuration
	PSC *PSCConfig `toml:"psc" json:"psc"`

	// MCP specific configuration
	MCP *MCPConfig `toml:"mcp_server" json:"mcp_server"`

	// Project configuration (only in project config files)
	Project *ProjectConfig `toml:"project" json:"project"`

	// Build configuration
	Build *BuildConfig `toml:"build" json:"build"`
}

// PVMConfig represents configuration for the Perl Version Manager
type PVMConfig struct {
	// DefaultPerl is the default Perl version to use when none is specified
	DefaultPerl string `toml:"default_perl" json:"default_perl"`

	// BuildJobs specifies the number of parallel jobs to use during build
	BuildJobs int `toml:"build_jobs" json:"build_jobs"`

	// DownloadMirror specifies the mirror to use for downloading Perl source
	DownloadMirror string `toml:"download_mirror" json:"download_mirror"`

	// PatchesDir specifies the directory containing patches to apply during build
	PatchesDir string `toml:"patches_dir" json:"patches_dir"`

	// Compiler specifies the compiler to use for building Perl
	Compiler string `toml:"compiler" json:"compiler"`

	// RunTests specifies whether to run tests during Perl installation
	RunTests bool `toml:"run_tests" json:"run_tests"`

	// VersionAliases maps aliases to actual version strings
	VersionAliases map[string]string `toml:"version_aliases" json:"version_aliases"`
}

// PVXConfig represents configuration for the Perl Version eXecutor
type PVXConfig struct {
	// CacheModules specifies whether to cache modules for faster startup
	CacheModules bool `toml:"cache_modules" json:"cache_modules"`

	// CleanupAfter specifies whether to clean up temporary files after execution
	CleanupAfter bool `toml:"cleanup_after" json:"cleanup_after"`

	// IsolationLevel specifies the level of isolation for script execution
	// Valid values: "none", "low", "medium", "high"
	IsolationLevel string `toml:"isolation_level" json:"isolation_level"`

	// AlwaysInstallDeps specifies whether to automatically install missing dependencies
	AlwaysInstallDeps bool `toml:"always_install_deps" json:"always_install_deps"`

	// Timeout specifies the maximum execution time in seconds
	Timeout int `toml:"timeout" json:"timeout"`

	// MaxMemory specifies the maximum memory usage (e.g., "512MB")
	MaxMemory string `toml:"max_memory" json:"max_memory"`

	// IsolationReadOnlyPaths specifies paths that should be read-only in high isolation mode
	IsolationReadOnlyPaths []string `toml:"isolation_ro_paths" json:"isolation_ro_paths"`

	// IsolationReadWritePaths specifies paths that should be read-write in high isolation mode
	IsolationReadWritePaths []string `toml:"isolation_rw_paths" json:"isolation_rw_paths"`

	// IsolatedOutput specifies whether to create an isolated output directory
	IsolatedOutput bool `toml:"isolated_output" json:"isolated_output"`

	// SaveOutputDir specifies where to save isolated output files
	SaveOutputDir string `toml:"save_output_dir" json:"save_output_dir"`

	// PreserveEnvVars specifies which environment variables to preserve in isolation
	PreserveEnvVars []string `toml:"preserve_env_vars" json:"preserve_env_vars"`

	// AdditionalModulePaths specifies additional module paths to add to PERL5LIB
	AdditionalModulePaths []string `toml:"additional_module_paths" json:"additional_module_paths"`

	// CustomModulePath specifies a custom module installation path
	CustomModulePath string `toml:"custom_module_path" json:"custom_module_path"`
}

// PVIConfig represents configuration for the Perl Version Installer
type PVIConfig struct {
	// PreferredInstaller specifies the preferred installation method
	// Valid values: "auto", "cpanm", "cpan", "cpm"
	PreferredInstaller string `toml:"preferred_installer" json:"preferred_installer"`

	// DefaultMirror specifies the CPAN mirror to use
	DefaultMirror string `toml:"default_mirror" json:"default_mirror"`

	// AdditionalMirrors specifies backup CPAN mirrors to use if the default mirror fails
	AdditionalMirrors []string `toml:"additional_mirrors" json:"additional_mirrors"`

	// MetadataSource specifies the source for CPAN metadata
	// Valid values: "metacpan", "cpan"
	MetadataSource string `toml:"metadata_source" json:"metadata_source"`

	// MetadataURL specifies the URL for the metadata API
	// Used when MetadataSource is "custom"
	MetadataURL string `toml:"metadata_url" json:"metadata_url"`

	// CacheDir specifies the directory to use for caching metadata
	CacheDir string `toml:"cache_dir" json:"cache_dir"`

	// CacheTTL specifies the time-to-live for cached metadata in hours
	// Set to 0 to always refresh metadata
	CacheTTL int `toml:"cache_ttl" json:"cache_ttl"`

	// TestDuringInstall specifies whether to run tests during module installation
	TestDuringInstall bool `toml:"test_during_install" json:"test_during_install"`

	// CacheModules specifies whether to cache modules for faster installation
	CacheModules bool `toml:"cache_modules" json:"cache_modules"`

	// ForceReinstall specifies whether to force reinstallation of modules
	ForceReinstall bool `toml:"force_reinstall" json:"force_reinstall"`

	// CheckSignatures specifies whether to check module signatures
	CheckSignatures bool `toml:"check_signatures" json:"check_signatures"`

	// DisableNetwork specifies whether to disable network access (for testing)
	DisableNetwork bool `toml:"disable_network" json:"disable_network"`
}

// PSCConfig represents configuration for the Perl Script Compiler
type PSCConfig struct {
	// TypeDefinitionsPath specifies the path to type definitions
	TypeDefinitionsPath string `toml:"type_definitions_path" json:"type_definitions_path"`

	// StrictMode enables more rigorous type checking
	StrictMode bool `toml:"strict_mode" json:"strict_mode"`

	// WatchExclude specifies patterns to exclude from watch mode
	WatchExclude []string `toml:"watch_exclude" json:"watch_exclude"`

	// GenerateMissingTypes specifies whether to generate missing type definitions
	GenerateMissingTypes bool `toml:"generate_missing_types" json:"generate_missing_types"`

	// CheckBeforeRun specifies whether to check types before running a script
	CheckBeforeRun bool `toml:"check_before_run" json:"check_before_run"`
}

// MCPConfig represents configuration for the MCP Server
type MCPConfig struct {
	// Port specifies the port for the MCP server
	Port int `toml:"port" json:"port"`

	// Host specifies the host address for the MCP server
	Host string `toml:"host" json:"host"`

	// AutoDiscoverProjects specifies whether to automatically discover Perl projects
	AutoDiscoverProjects bool `toml:"auto_discover_projects" json:"auto_discover_projects"`

	// AutoFixErrors specifies whether to attempt auto-fixing errors via sampling
	AutoFixErrors bool `toml:"auto_fix_errors" json:"auto_fix_errors"`

	// ValidationCacheSize specifies the maximum size for validation cache
	ValidationCacheSize string `toml:"validation_cache_size" json:"validation_cache_size"`

	// EmbeddingProvider specifies which embedding provider to use
	// Valid values: "openai", "voyageai", "huggingface"
	EmbeddingProvider string `toml:"embedding_provider" json:"embedding_provider"`

	// EmbeddingCacheSize specifies the maximum size for embedding cache
	EmbeddingCacheSize string `toml:"embedding_cache_size" json:"embedding_cache_size"`

	// EmbeddingModel specifies the model to use (provider-specific)
	EmbeddingModel string `toml:"embedding_model" json:"embedding_model"`

	// GenerationMemorySize specifies the memory size for generation context
	GenerationMemorySize int `toml:"generation_memory_size" json:"generation_memory_size"`

	// EnableIterativeRefinement specifies whether to enable iterative refinement
	EnableIterativeRefinement bool `toml:"enable_iterative_refinement" json:"enable_iterative_refinement"`

	// MaxConcurrentRequests specifies the maximum number of concurrent requests
	MaxConcurrentRequests int `toml:"max_concurrent_requests" json:"max_concurrent_requests"`

	// RequestTimeout specifies the timeout for requests
	RequestTimeout time.Duration `toml:"request_timeout" json:"request_timeout"`
}

// ProjectConfig represents project-specific configuration (only in project config files)
type ProjectConfig struct {
	// Name is the project name
	Name string `toml:"name" json:"name"`

	// Version is the project version
	Version string `toml:"version" json:"version"`

	// PerlVersion is the required Perl version for this project
	PerlVersion string `toml:"perl_version" json:"perl_version"`

	// Description is a brief description of the project
	Description string `toml:"description" json:"description"`

	// Author information
	Author []string `toml:"author" json:"author"`

	// License specifies the project license
	License string `toml:"license" json:"license"`

	// Homepage URL for the project
	Homepage string `toml:"homepage" json:"homepage"`

	// BugTracker URL for issue reporting
	BugTracker string `toml:"bug_tracker" json:"bug_tracker"`

	// Repository URL
	Repository string `toml:"repository" json:"repository"`
}

// BuildConfig represents build system configuration
type BuildConfig struct {
	// Mode specifies the default build mode
	// Valid values: "distribution", "inline", "both"
	Mode string `toml:"mode" json:"mode"`

	// OutputDir specifies the build output directory
	OutputDir string `toml:"output_dir" json:"output_dir"`

	// CleanBeforeBuild specifies whether to clean output before building
	CleanBeforeBuild bool `toml:"clean_before_build" json:"clean_before_build"`

	// TypeCheck configuration
	TypeCheck *BuildTypeCheckConfig `toml:"typecheck" json:"typecheck"`

	// Files configuration
	Files *BuildFilesConfig `toml:"files" json:"files"`

	// Distribution configuration
	Distribution *BuildDistributionConfig `toml:"distribution" json:"distribution"`
}

// BuildTypeCheckConfig represents type checking configuration
type BuildTypeCheckConfig struct {
	// Strict enables strict type checking
	Strict bool `toml:"strict" json:"strict"`

	// Experimental enables experimental type features
	Experimental bool `toml:"experimental" json:"experimental"`

	// TargetPerl specifies the target Perl version for type checking
	TargetPerl string `toml:"target_perl" json:"target_perl"`
}

// BuildFilesConfig represents file handling configuration
type BuildFilesConfig struct {
	// Include specifies file patterns to include in builds
	Include []string `toml:"include" json:"include"`

	// Exclude specifies file patterns to exclude from builds
	Exclude []string `toml:"exclude" json:"exclude"`

	// WatchDirs specifies directories to watch in watch mode
	WatchDirs []string `toml:"watch_dirs" json:"watch_dirs"`
}

// BuildDistributionConfig represents distribution build configuration
type BuildDistributionConfig struct {
	// IncludeTests specifies whether to include tests in distribution
	IncludeTests bool `toml:"include_tests" json:"include_tests"`

	// IncludeScripts specifies whether to include scripts in distribution
	IncludeScripts bool `toml:"include_scripts" json:"include_scripts"`

	// Installer specifies the installer type to generate
	// Valid values: "ExtUtils::MakeMaker", "Module::Build", "Module::Install"
	Installer string `toml:"installer" json:"installer"`
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
			AdditionalMirrors:  []string{"https://cpan.mirror.co.uk", "https://www.cpan.org"},
			MetadataSource:     "metacpan",
			MetadataURL:        "https://api.metacpan.org/v1",
			CacheDir:           "$XDG_CACHE_HOME/pvm/cpan",
			CacheTTL:           24,
			TestDuringInstall:  false,
			CacheModules:       true,
			CheckSignatures:    true,
			DisableNetwork:     false,
		},
		PSC: &PSCConfig{
			TypeDefinitionsPath:  "$XDG_DATA_HOME/pvm/type_definitions",
			StrictMode:           false,
			WatchExclude:         []string{"**/node_modules/**", "**/.git/**", "**/local/**"},
			GenerateMissingTypes: true,
			CheckBeforeRun:       true,
		},
		MCP: &MCPConfig{
			Port:                      3000,
			Host:                      "localhost",
			AutoDiscoverProjects:      true,
			AutoFixErrors:             true,
			ValidationCacheSize:       "50MB",
			EmbeddingProvider:         "openai",
			EmbeddingCacheSize:        "100MB",
			EmbeddingModel:            "text-embedding-3-small",
			GenerationMemorySize:      50,
			EnableIterativeRefinement: true,
			MaxConcurrentRequests:     10,
			RequestTimeout:            30 * time.Second,
		},
		Project: &ProjectConfig{
			PerlVersion: "5.36",
			License:     "perl_5",
		},
		Build: &BuildConfig{
			Mode:             "distribution",
			OutputDir:        "build",
			CleanBeforeBuild: true,
			TypeCheck: &BuildTypeCheckConfig{
				Strict:       false,
				Experimental: false,
				TargetPerl:   "5.36",
			},
			Files: &BuildFilesConfig{
				Include:   []string{"lib/**/*.pm"},
				Exclude:   []string{"local/**", "build/**", "**/.git/**"},
				WatchDirs: []string{"lib", "script", "t"},
			},
			Distribution: &BuildDistributionConfig{
				IncludeTests:   true,
				IncludeScripts: true,
				Installer:      "ExtUtils::MakeMaker",
			},
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

	// Add MCP validation
	if c.MCP != nil {
		errors = append(errors, c.MCP.Validate()...)
	}

	// Add Project validation
	if c.Project != nil {
		errors = append(errors, c.Project.Validate()...)
	}

	// Add Build validation
	if c.Build != nil {
		errors = append(errors, c.Build.Validate()...)
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

	// Validate MetadataSource
	validSources := map[string]bool{
		"metacpan": true,
		"cpan":     true,
		"custom":   true,
	}

	if c.MetadataSource != "" && !validSources[c.MetadataSource] {
		errors = append(errors, ValidateError("MetadataSource must be one of: metacpan, cpan, custom"))
	}

	// Validate MetadataURL if source is custom
	if c.MetadataSource == "custom" && c.MetadataURL == "" {
		errors = append(errors, ValidateError("MetadataURL must be specified when MetadataSource is 'custom'"))
	}

	// Validate CacheTTL
	if c.CacheTTL < 0 {
		errors = append(errors, ValidateError("CacheTTL cannot be negative"))
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

// Validate checks if the MCP configuration is valid
func (c *MCPConfig) Validate() []error {
	var errors []error

	// Validate Port
	if c.Port < 1 || c.Port > 65535 {
		errors = append(errors, ValidateError("Port must be between 1 and 65535"))
	}

	// Validate Host
	if c.Host == "" {
		errors = append(errors, ValidateError("Host cannot be empty"))
	}

	// Validate EmbeddingProvider
	validProviders := map[string]bool{
		"openai":      true,
		"voyageai":    true,
		"huggingface": true,
	}

	if c.EmbeddingProvider != "" && !validProviders[c.EmbeddingProvider] {
		errors = append(errors, ValidateError("EmbeddingProvider must be one of: openai, voyageai, huggingface"))
	}

	// Validate cache sizes (must have valid memory format)
	if c.ValidationCacheSize != "" && !isValidMemoryFormat(c.ValidationCacheSize) {
		errors = append(errors, ValidateError("ValidationCacheSize must be in format like '50MB' or '2GB'"))
	}

	if c.EmbeddingCacheSize != "" && !isValidMemoryFormat(c.EmbeddingCacheSize) {
		errors = append(errors, ValidateError("EmbeddingCacheSize must be in format like '100MB' or '2GB'"))
	}

	// Validate GenerationMemorySize
	if c.GenerationMemorySize < 0 {
		errors = append(errors, ValidateError("GenerationMemorySize cannot be negative"))
	}

	// Validate MaxConcurrentRequests
	if c.MaxConcurrentRequests < 1 {
		errors = append(errors, ValidateError("MaxConcurrentRequests must be at least 1"))
	}

	// Validate RequestTimeout (must be positive duration)
	if c.RequestTimeout <= 0 {
		errors = append(errors, ValidateError("RequestTimeout must be positive"))
	}

	return errors
}

// ValidateError creates a validation error
type ValidateError string

// Error implements the error interface
func (e ValidateError) Error() string {
	return string(e)
}

// Validate checks if the Project configuration is valid
func (c *ProjectConfig) Validate() []error {
	var errors []error

	// Validate PerlVersion format if provided
	if c.PerlVersion != "" {
		// Basic validation - should be in format like "5.36" or "5.36.0"
		if !strings.Contains(c.PerlVersion, ".") {
			errors = append(errors, ValidateError("PerlVersion should be in format like '5.36' or '5.36.0'"))
		}
	}

	// Validate License if provided
	validLicenses := map[string]bool{
		"perl_5":       true,
		"artistic_2":   true,
		"apache_2_0":   true,
		"mit":          true,
		"bsd_3_clause": true,
		"gpl_1":        true,
		"gpl_2":        true,
		"gpl_3":        true,
		"lgpl_2_1":     true,
		"lgpl_3_0":     true,
	}

	if c.License != "" && !validLicenses[c.License] {
		errors = append(errors, ValidateError("License must be a valid license identifier"))
	}

	return errors
}

// Validate checks if the Build configuration is valid
func (c *BuildConfig) Validate() []error {
	var errors []error

	// Validate Mode
	validModes := map[string]bool{
		"distribution": true,
		"inline":       true,
		"both":         true,
	}

	if c.Mode != "" && !validModes[c.Mode] {
		errors = append(errors, ValidateError("Build mode must be one of: distribution, inline, both"))
	}

	// Validate OutputDir
	if c.OutputDir == "" {
		errors = append(errors, ValidateError("Build OutputDir cannot be empty"))
	}

	// Validate TypeCheck configuration
	if c.TypeCheck != nil {
		errors = append(errors, c.TypeCheck.Validate()...)
	}

	// Validate Files configuration
	if c.Files != nil {
		errors = append(errors, c.Files.Validate()...)
	}

	// Validate Distribution configuration
	if c.Distribution != nil {
		errors = append(errors, c.Distribution.Validate()...)
	}

	return errors
}

// Validate checks if the BuildTypeCheck configuration is valid
func (c *BuildTypeCheckConfig) Validate() []error {
	var errors []error

	// Validate TargetPerl format if provided
	if c.TargetPerl != "" {
		// Basic validation - should be in format like "5.36" or "5.36.0"
		if !strings.Contains(c.TargetPerl, ".") {
			errors = append(errors, ValidateError("TypeCheck TargetPerl should be in format like '5.36' or '5.36.0'"))
		}
	}

	return errors
}

// Validate checks if the BuildFiles configuration is valid
func (c *BuildFilesConfig) Validate() []error {
	var errors []error

	// Validate Include patterns
	for _, pattern := range c.Include {
		if pattern == "" {
			errors = append(errors, ValidateError("Build files include patterns cannot be empty"))
			break
		}
	}

	// Validate Exclude patterns
	for _, pattern := range c.Exclude {
		if pattern == "" {
			errors = append(errors, ValidateError("Build files exclude patterns cannot be empty"))
			break
		}
	}

	// Validate WatchDirs
	for _, dir := range c.WatchDirs {
		if dir == "" {
			errors = append(errors, ValidateError("Build files watch directories cannot be empty"))
			break
		}
	}

	return errors
}

// Validate checks if the BuildDistribution configuration is valid
func (c *BuildDistributionConfig) Validate() []error {
	var errors []error

	// Validate Installer
	validInstallers := map[string]bool{
		"ExtUtils::MakeMaker": true,
		"Module::Build":       true,
		"Module::Install":     true,
	}

	if c.Installer != "" && !validInstallers[c.Installer] {
		errors = append(errors, ValidateError("Build distribution installer must be one of: ExtUtils::MakeMaker, Module::Build, Module::Install"))
	}

	return errors
}

// Helper functions for validation

func isValidMemoryFormat(memory string) bool {
	return strings.HasSuffix(memory, "MB") || strings.HasSuffix(memory, "GB") ||
		strings.HasSuffix(memory, "KB") || strings.HasSuffix(memory, "TB")
}
