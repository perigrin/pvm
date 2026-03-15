// ABOUTME: Configuration types for the PVM Ecosystem
// ABOUTME: Defines structs matching the TOML configuration format

package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// Config represents the root configuration object
type Config struct {
	// PVM specific configuration
	PVM *PVMConfig `toml:"pvm" json:"pvm"`

	// PVX specific configuration
	PVX *PVXConfig `toml:"pvx" json:"pvx"`

	// PM specific configuration
	PM *PMConfig `toml:"pm" json:"pm"`

	// PSC specific configuration
	PSC *PSCConfig `toml:"psc" json:"psc"`

	// MCP specific configuration
	MCP *MCPConfig `toml:"mcp_server" json:"mcp_server"`

	// Project configuration (only in project config files)
	Project *ProjectConfig `toml:"project" json:"project"`

	// Build configuration
	Build *BuildConfig `toml:"build" json:"build"`

	// Dependencies configuration
	Dependencies *DependenciesConfig `toml:"dependencies" json:"dependencies"`

	// Documentation configuration
	Docs *DocsConfig `toml:"docs" json:"docs"`
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

	// Update configuration
	Update *PVMUpdateConfig `toml:"update" json:"update"`

	// Binary distribution configuration
	Binary *PVMBinaryConfig `toml:"binary" json:"binary"`

	// Shell integration configuration
	Shell *PVMShellConfig `toml:"shell" json:"shell"`
}

// PVMCustomMirrorConfig represents a custom mirror with authentication
type PVMCustomMirrorConfig struct {
	// Name is a human-readable identifier for this mirror
	Name string `toml:"name" json:"name"`

	// Type specifies the mirror type (github-releases, jsdelivr, cloudflare-r2, direct)
	Type string `toml:"type" json:"type"`

	// BaseURL is the base URL for the mirror
	BaseURL string `toml:"base_url" json:"base_url"`

	// Priority determines mirror selection order (lower = higher priority)
	Priority int `toml:"priority" json:"priority"`

	// Enabled specifies whether this mirror is active
	Enabled bool `toml:"enabled" json:"enabled"`

	// Timeout specifies request timeout for this mirror (e.g., "30s", "2m")
	Timeout string `toml:"timeout" json:"timeout"`

	// MaxRetries specifies maximum retry attempts for this mirror
	MaxRetries int `toml:"max_retries" json:"max_retries"`

	// HealthCheck specifies health check path relative to BaseURL
	HealthCheck string `toml:"health_check" json:"health_check"`

	// Authentication configuration for private mirrors
	Auth *PVMCustomMirrorAuth `toml:"auth" json:"auth"`

	// Headers specifies custom HTTP headers to send with requests
	Headers map[string]string `toml:"headers" json:"headers"`

	// URLTemplate specifies custom URL template for version/platform mapping
	// Variables: {version}, {platform}, {filename}, {ext}
	// Default: "{base_url}/perl-{version}/{filename}"
	URLTemplate string `toml:"url_template" json:"url_template"`

	// VersionMapping maps PVM versions to custom build versions
	// Example: {"5.38.0": "5.38.0-custom1", "5.40.0": "5.40.0-patched"}
	VersionMapping map[string]string `toml:"version_mapping" json:"version_mapping"`
}

// PVMCustomMirrorAuth represents authentication configuration for custom mirrors
type PVMCustomMirrorAuth struct {
	// Type specifies the authentication method
	// Valid values: "none", "basic", "bearer", "api-key", "oauth2"
	Type string `toml:"type" json:"type"`

	// Username for basic authentication
	Username string `toml:"username" json:"username"`

	// Password for basic authentication (consider using environment variables)
	Password string `toml:"password" json:"password"`

	// Token for bearer token authentication
	Token string `toml:"token" json:"token"`

	// APIKey for API key authentication
	APIKey string `toml:"api_key" json:"api_key"`

	// APIKeyHeader specifies the header name for API key (default: "X-API-Key")
	APIKeyHeader string `toml:"api_key_header" json:"api_key_header"`

	// OAuth2 configuration
	OAuth2 *PVMCustomMirrorOAuth2 `toml:"oauth2" json:"oauth2"`
}

// PVMCustomMirrorOAuth2 represents OAuth2 authentication configuration
type PVMCustomMirrorOAuth2 struct {
	// ClientID for OAuth2 authentication
	ClientID string `toml:"client_id" json:"client_id"`

	// ClientSecret for OAuth2 authentication
	ClientSecret string `toml:"client_secret" json:"client_secret"`

	// TokenURL for OAuth2 token endpoint
	TokenURL string `toml:"token_url" json:"token_url"`

	// Scopes for OAuth2 authentication
	Scopes []string `toml:"scopes" json:"scopes"`
}

// PVMBinaryConfig represents configuration for binary distribution support
type PVMBinaryConfig struct {
	// DefaultInstallMethod specifies the default installation method
	// Valid values: "binary", "source", "prefer-binary"
	DefaultInstallMethod string `toml:"default_install_method" json:"default_install_method"`

	// BinaryMirrors specifies list of mirror URLs for binary downloads
	BinaryMirrors []string `toml:"binary_mirrors" json:"binary_mirrors"`

	// CustomMirrors specifies custom mirror configurations with authentication
	CustomMirrors []*PVMCustomMirrorConfig `toml:"custom_mirrors" json:"custom_mirrors"`

	// CacheRetentionDays specifies the binary cache cleanup policy in days
	CacheRetentionDays int `toml:"cache_retention_days" json:"cache_retention_days"`

	// MaxCacheSize specifies the maximum cache size in GB
	MaxCacheSize int `toml:"max_cache_size" json:"max_cache_size"`

	// VerifyChecksums specifies whether to verify checksums for binary downloads
	VerifyChecksums bool `toml:"verify_checksums" json:"verify_checksums"`

	// ParallelDownloads specifies whether to enable parallel downloading
	ParallelDownloads bool `toml:"parallel_downloads" json:"parallel_downloads"`

	// MaxRetries specifies maximum number of download retries
	MaxRetries int `toml:"max_retries" json:"max_retries"`

	// Timeout specifies timeout for download operations (e.g., "5m", "30s")
	Timeout string `toml:"timeout" json:"timeout"`

	// BandwidthLimit specifies bandwidth limit for downloads (e.g., "10MB", "1GB")
	BandwidthLimit string `toml:"bandwidth_limit" json:"bandwidth_limit"`
}

// PVMShellConfig represents configuration for shell integration
type PVMShellConfig struct {
	// ShowDirectoryChangeAlerts controls whether to show version alerts when changing directories
	ShowDirectoryChangeAlerts bool `toml:"show_directory_change_alerts" json:"show_directory_change_alerts"`
}

// PVMUpdateConfig represents configuration for PVM self-updater
type PVMUpdateConfig struct {
	// AutoUpdateEnabled specifies whether automatic update checking is enabled
	AutoUpdateEnabled bool `toml:"auto_update_enabled" json:"auto_update_enabled"`

	// AutoUpdateInterval specifies how often to check for updates (e.g., "24h", "1h")
	AutoUpdateInterval string `toml:"auto_update_interval" json:"auto_update_interval"`

	// Repository specifies the GitHub repository to check for updates
	Repository string `toml:"repository" json:"repository"`

	// Channel specifies the update channel (stable, beta, alpha, nightly, developer)
	Channel string `toml:"channel" json:"channel"`

	// GitHubToken specifies the GitHub token for higher API rate limits
	GitHubToken string `toml:"github_token" json:"github_token"`

	// BackupEnabled specifies whether to create backups before updating
	BackupEnabled bool `toml:"backup_enabled" json:"backup_enabled"`

	// AutoRollbackEnabled specifies whether to automatically rollback on failure
	AutoRollbackEnabled bool `toml:"auto_rollback_enabled" json:"auto_rollback_enabled"`

	// CheckPrerelease specifies whether to consider pre-release versions
	CheckPrerelease bool `toml:"check_prerelease" json:"check_prerelease"`

	// NotificationsEnabled specifies whether to show update notifications
	NotificationsEnabled bool `toml:"notifications_enabled" json:"notifications_enabled"`

	// SecurityUpdatesOnly specifies whether to only auto-update for security releases
	SecurityUpdatesOnly bool `toml:"security_updates_only" json:"security_updates_only"`

	// MaxRetries specifies maximum number of download retries
	MaxRetries int `toml:"max_retries" json:"max_retries"`

	// Timeout specifies timeout for download operations (e.g., "5m", "30s")
	Timeout string `toml:"timeout" json:"timeout"`

	// SkipChecksums specifies whether to skip checksum validation (not recommended)
	SkipChecksums bool `toml:"skip_checksums" json:"skip_checksums"`
}

// PVXConfig represents configuration for the Perl Version eXecutor
type PVXConfig struct {
	// CacheModules specifies whether to cache modules for faster startup
	CacheModules bool `toml:"cache_modules" json:"cache_modules"`

	// CleanupAfter specifies whether to clean up temporary files after execution
	CleanupAfter bool `toml:"cleanup_after" json:"cleanup_after"`

	// IsolationLevel specifies the level of isolation for script execution
	// Valid values: "global", "local", "clean"
	IsolationLevel string `toml:"isolation_level" json:"isolation_level"`

	// AlwaysInstallDeps specifies whether to automatically install missing dependencies
	AlwaysInstallDeps bool `toml:"always_install_deps" json:"always_install_deps"`

	// AutoInstallPerl specifies whether to automatically install missing Perl versions
	AutoInstallPerl bool `toml:"auto_install_perl" json:"auto_install_perl"`

	// Timeout specifies the maximum execution time in seconds
	Timeout int `toml:"timeout" json:"timeout"`

	// MaxMemory specifies the maximum memory usage (e.g., "512MB")
	MaxMemory string `toml:"max_memory" json:"max_memory"`

	// IsolationReadOnlyPaths specifies paths that should be read-only in clean isolation mode
	IsolationReadOnlyPaths []string `toml:"isolation_ro_paths" json:"isolation_ro_paths"`

	// IsolationReadWritePaths specifies paths that should be read-write in clean isolation mode
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

// PMConfig represents configuration for the Perl Version Installer
type PMConfig struct {
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

	// Backup configuration for cpanfile operations
	Backup *PMBackupConfig `toml:"backup" json:"backup"`
}

// PMBackupConfig represents backup configuration for cpanfile operations
type PMBackupConfig struct {
	// CpanfileBackup specifies the backup mode for cpanfile operations
	// Valid values: "off", "local", "cache"
	CpanfileBackup string `toml:"cpanfile_backup" json:"cpanfile_backup"`

	// RetentionDays specifies how many days to keep backups before cleanup
	RetentionDays int `toml:"retention_days" json:"retention_days"`

	// MaxBackups specifies the maximum number of backups to keep per project
	MaxBackups int `toml:"max_backups" json:"max_backups"`
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

	// LSP server configuration
	LSP *PSCLSPConfig `toml:"lsp" json:"lsp"`
}

// PSCLSPConfig represents configuration for PSC LSP server
type PSCLSPConfig struct {
	// Logging configuration
	LogFile  string `toml:"log_file" json:"log_file"`
	LogLevel string `toml:"log_level" json:"log_level"` // debug, info, warn, error
	Verbose  bool   `toml:"verbose" json:"verbose"`

	// Communication settings
	DefaultMode string `toml:"default_mode" json:"default_mode"` // stdio, tcp
	TCPPort     int    `toml:"tcp_port" json:"tcp_port"`

	// Feature toggles
	EnableHover      bool `toml:"enable_hover" json:"enable_hover"`
	EnableCompletion bool `toml:"enable_completion" json:"enable_completion"`
	EnableDefinition bool `toml:"enable_definition" json:"enable_definition"`
	EnableReferences bool `toml:"enable_references" json:"enable_references"`
	EnableFormatting bool `toml:"enable_formatting" json:"enable_formatting"`

	// Performance settings
	MaxCacheSize     int           `toml:"max_cache_size" json:"max_cache_size"`
	RequestTimeout   time.Duration `toml:"request_timeout" json:"request_timeout"`
	DiagnosticsDelay time.Duration `toml:"diagnostics_delay" json:"diagnostics_delay"`

	// Project-specific settings
	WorkspaceSymbols   bool     `toml:"workspace_symbols" json:"workspace_symbols"`
	CrossFileAnalysis  bool     `toml:"cross_file_analysis" json:"cross_file_analysis"`
	ExcludePatterns    []string `toml:"exclude_patterns" json:"exclude_patterns"`
	IncludeDirectories []string `toml:"include_directories" json:"include_directories"`
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

	// Sampling configuration for real MCP integration
	Sampling *MCPSamplingConfig `toml:"sampling" json:"sampling"`

	// Authentication configuration
	Auth *MCPAuthConfig `toml:"auth" json:"auth"`

	// Circuit breaker configuration
	CircuitBreaker *MCPCircuitBreakerConfig `toml:"circuit_breaker" json:"circuit_breaker"`
}

// MCPSamplingConfig represents configuration for MCP sampling
type MCPSamplingConfig struct {
	// Mode specifies the sampling mode: "real" or "mock"
	Mode string `toml:"mode" json:"mode"`

	// Enabled specifies whether sampling is enabled
	Enabled bool `toml:"enabled" json:"enabled"`

	// Endpoint specifies the MCP server endpoint for real mode
	Endpoint string `toml:"endpoint" json:"endpoint"`

	// MaxTokens specifies the maximum tokens for sampling requests
	MaxTokens int `toml:"max_tokens" json:"max_tokens"`

	// Temperature specifies the temperature for sampling
	Temperature float64 `toml:"temperature" json:"temperature"`

	// TopP specifies the top-p value for sampling
	TopP float64 `toml:"top_p" json:"top_p"`

	// SystemPrompt specifies the default system prompt
	SystemPrompt string `toml:"system_prompt" json:"system_prompt"`

	// MaxRetries specifies the maximum number of retries for failed requests
	MaxRetries int `toml:"max_retries" json:"max_retries"`

	// Timeout specifies the timeout for sampling requests
	Timeout time.Duration `toml:"timeout" json:"timeout"`
}

// MCPAuthConfig represents authentication configuration for MCP
type MCPAuthConfig struct {
	// Type specifies the authentication type: "none", "apikey", "oauth"
	Type string `toml:"type" json:"type"`

	// APIKey specifies the API key for apikey authentication
	APIKey string `toml:"api_key" json:"api_key"`

	// TokenFile specifies the file containing the authentication token
	TokenFile string `toml:"token_file" json:"token_file"`

	// OAuth configuration
	OAuth *MCPOAuthConfig `toml:"oauth" json:"oauth"`
}

// MCPOAuthConfig represents OAuth configuration for MCP
type MCPOAuthConfig struct {
	// ClientID specifies the OAuth client ID
	ClientID string `toml:"client_id" json:"client_id"`

	// ClientSecret specifies the OAuth client secret
	ClientSecret string `toml:"client_secret" json:"client_secret"`

	// AuthURL specifies the authorization server URL
	AuthURL string `toml:"auth_url" json:"auth_url"`

	// TokenURL specifies the token endpoint URL
	TokenURL string `toml:"token_url" json:"token_url"`

	// Scopes specifies the OAuth scopes to request
	Scopes []string `toml:"scopes" json:"scopes"`
}

// MCPCircuitBreakerConfig represents circuit breaker configuration
type MCPCircuitBreakerConfig struct {
	// Enabled specifies whether the circuit breaker is enabled
	Enabled bool `toml:"enabled" json:"enabled"`

	// FailureThreshold specifies the failure threshold for opening the circuit
	FailureThreshold int `toml:"failure_threshold" json:"failure_threshold"`

	// Timeout specifies the timeout for circuit breaker operations
	Timeout time.Duration `toml:"timeout" json:"timeout"`

	// ResetTimeout specifies the reset timeout for the circuit breaker
	ResetTimeout time.Duration `toml:"reset_timeout" json:"reset_timeout"`
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

// DependenciesConfig represents dependency management configuration
type DependenciesConfig struct {
	// Cpanfile specifies the path to the cpanfile for dependency management
	Cpanfile string `toml:"cpanfile" json:"cpanfile"`

	// LocalLib specifies the local library directory for project dependencies
	LocalLib string `toml:"local_lib" json:"local_lib"`
}

// Validate checks if the Dependencies configuration is valid
func (c *DependenciesConfig) Validate() []error {
	var errors []error

	// Validate LocalLib path
	if c.LocalLib != "" {
		// Check for path traversal
		if strings.Contains(c.LocalLib, "..") {
			errors = append(errors, ValidateError("LocalLib cannot contain path traversal sequences (..)"))
		}

		// Check for absolute paths
		if filepath.IsAbs(c.LocalLib) {
			errors = append(errors, ValidateError("LocalLib must be a relative path"))
		}

		// Check for empty path components
		if strings.Contains(c.LocalLib, "//") {
			errors = append(errors, ValidateError("LocalLib cannot contain empty path components"))
		}

		// Check for null bytes and control characters
		for _, r := range c.LocalLib {
			if r == 0 || (r < 32 && r != '\t') {
				errors = append(errors, ValidateError("LocalLib cannot contain null bytes or control characters"))
				break
			}
		}

		// Length validation
		if len(c.LocalLib) > 255 {
			errors = append(errors, ValidateError("LocalLib path too long (max 255 characters)"))
		}
	}

	return errors
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

// DocsConfig represents configuration for external documentation access
type DocsConfig struct {
	// Repository specifies the GitHub repository containing documentation
	// Format: "owner/repo" (e.g., "perigrin/pvm")
	Repository string `toml:"repository" json:"repository"`

	// Branch specifies the Git branch or tag to fetch documentation from
	Branch string `toml:"branch" json:"branch"`

	// GitHubToken specifies the GitHub token for API access and rate limiting
	// Can be set via environment variable GITHUB_TOKEN
	GitHubToken string `toml:"github_token" json:"github_token"`

	// CacheDir specifies the directory for caching external documentation
	CacheDir string `toml:"cache_dir" json:"cache_dir"`

	// CacheTTL specifies the cache time-to-live in hours
	// 0 = always refresh, -1 = never expire
	CacheTTL int `toml:"cache_ttl" json:"cache_ttl"`

	// AutoUpdate specifies whether to automatically update documentation cache
	AutoUpdate bool `toml:"auto_update" json:"auto_update"`

	// UpdateInterval specifies how often to update documentation (e.g., "24h", "1h")
	UpdateInterval string `toml:"update_interval" json:"update_interval"`

	// MaxRetries specifies maximum retry attempts for failed requests
	MaxRetries int `toml:"max_retries" json:"max_retries"`

	// Timeout specifies timeout for documentation requests (e.g., "30s", "2m")
	Timeout string `toml:"timeout" json:"timeout"`

	// OfflineMode specifies whether to operate in offline-only mode
	// When true, only cached/embedded documentation is available
	OfflineMode bool `toml:"offline_mode" json:"offline_mode"`

	// PreferEmbedded specifies whether to prefer embedded docs over external
	PreferEmbedded bool `toml:"prefer_embedded" json:"prefer_embedded"`

	// DocumentSources specifies additional documentation sources
	DocumentSources []*DocsSourceConfig `toml:"document_sources" json:"document_sources"`
}

// DocsSourceConfig represents a documentation source configuration
type DocsSourceConfig struct {
	// Name is a human-readable identifier for this source
	Name string `toml:"name" json:"name"`

	// Type specifies the source type (github, url, local)
	Type string `toml:"type" json:"type"`

	// Repository specifies the GitHub repository (for github type)
	Repository string `toml:"repository" json:"repository"`

	// Branch specifies the branch/tag (for github type)
	Branch string `toml:"branch" json:"branch"`

	// BaseURL specifies the base URL (for url type)
	BaseURL string `toml:"base_url" json:"base_url"`

	// Path specifies the local path (for local type)
	Path string `toml:"path" json:"path"`

	// Priority determines source selection order (lower = higher priority)
	Priority int `toml:"priority" json:"priority"`

	// Enabled specifies whether this source is active
	Enabled bool `toml:"enabled" json:"enabled"`

	// Categories specifies which documentation categories this source provides
	Categories []string `toml:"categories" json:"categories"`
}

// NewDefaultConfig creates a new configuration with default values
func NewDefaultConfig() *Config {
	cfg := &Config{
		PVM: &PVMConfig{
			DefaultPerl:    "5.40.2",
			BuildJobs:      4,
			DownloadMirror: "https://www.cpan.org/src/5.0",
			RunTests:       true,
			VersionAliases: map[string]string{
				"latest": "5.40.2",
				"stable": "5.36.0",
			},
			Update: &PVMUpdateConfig{
				AutoUpdateEnabled:    false,
				AutoUpdateInterval:   "24h",
				Repository:           "perigrin/pvm",
				Channel:              "stable",
				GitHubToken:          "",
				BackupEnabled:        true,
				AutoRollbackEnabled:  true,
				CheckPrerelease:      false,
				NotificationsEnabled: true,
				SecurityUpdatesOnly:  false,
				MaxRetries:           3,
				Timeout:              "5m",
				SkipChecksums:        false,
			},
			Binary: &PVMBinaryConfig{
				DefaultInstallMethod: "source",
				BinaryMirrors: []string{
					"https://github.com/perigrin/pvm/releases/download",
				},
				CustomMirrors:      []*PVMCustomMirrorConfig{},
				CacheRetentionDays: 30,
				MaxCacheSize:       5,
				VerifyChecksums:    true,
				ParallelDownloads:  true,
				MaxRetries:         3,
				Timeout:            "10m",
				BandwidthLimit:     "",
			},
			Shell: &PVMShellConfig{
				ShowDirectoryChangeAlerts: true, // Default to showing alerts
			},
		},
		PVX: &PVXConfig{
			CacheModules:            true,
			CleanupAfter:            true,
			IsolationLevel:          "clean",
			AlwaysInstallDeps:       true,
			AutoInstallPerl:         false, // Conservative default - require explicit opt-in
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
		PM: &PMConfig{
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
			Backup: &PMBackupConfig{
				CpanfileBackup: "off", // Default to no backups to avoid directory clutter
				RetentionDays:  30,    // Clean up backups after 30 days
				MaxBackups:     10,    // Limit backup storage usage
			},
		},
		PSC: &PSCConfig{
			TypeDefinitionsPath:  "$XDG_DATA_HOME/pvm/type_definitions",
			StrictMode:           false,
			WatchExclude:         []string{"**/node_modules/**", "**/.git/**", "**/local/**"},
			GenerateMissingTypes: true,
			CheckBeforeRun:       true,
			LSP: &PSCLSPConfig{
				// Logging configuration - reasonable defaults
				LogLevel: "info",
				Verbose:  false,

				// Communication settings - prefer stdio for editor integration
				DefaultMode: "stdio",
				TCPPort:     9999,

				// Feature toggles - enable all features by default
				EnableHover:      true,
				EnableCompletion: true,
				EnableDefinition: true,
				EnableReferences: true,
				EnableFormatting: true,

				// Performance settings - reasonable defaults
				MaxCacheSize:     1000,
				RequestTimeout:   5 * time.Second,
				DiagnosticsDelay: 500 * time.Millisecond,

				// Project-specific settings - sensible defaults
				WorkspaceSymbols:   true,
				CrossFileAnalysis:  true,
				ExcludePatterns:    []string{"**/test_data/**", "**/temp/**", "**/.git/**"},
				IncludeDirectories: []string{"lib", "script"},
			},
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
		Dependencies: &DependenciesConfig{
			Cpanfile: "cpanfile",
			LocalLib: "local", // Changed default from "lib" to "local"
		},
		Docs: &DocsConfig{
			Repository:      "perigrin/pvm",
			Branch:          "pu",
			GitHubToken:     "",
			CacheDir:        "$XDG_CACHE_HOME/pvm/docs",
			CacheTTL:        24, // 24 hours
			AutoUpdate:      true,
			UpdateInterval:  "24h",
			MaxRetries:      3,
			Timeout:         "30s",
			OfflineMode:     false,
			PreferEmbedded:  false,
			DocumentSources: []*DocsSourceConfig{},
		},
	}

	// Expand environment variables in paths
	cfg.PM.CacheDir = expandEnvironmentVariables(cfg.PM.CacheDir)
	cfg.PSC.TypeDefinitionsPath = expandEnvironmentVariables(cfg.PSC.TypeDefinitionsPath)
	cfg.Docs.CacheDir = expandEnvironmentVariables(cfg.Docs.CacheDir)

	return cfg
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

	// Add PM validation
	if c.PM != nil {
		errors = append(errors, c.PM.Validate()...)
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

	// Add Dependencies validation
	if c.Dependencies != nil {
		errors = append(errors, c.Dependencies.Validate()...)
	}

	// Add Docs validation
	if c.Docs != nil {
		errors = append(errors, c.Docs.Validate()...)
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

	// Validate Update configuration
	if c.Update != nil {
		errors = append(errors, c.Update.Validate()...)
	}

	// Validate Binary configuration
	if c.Binary != nil {
		errors = append(errors, c.Binary.Validate()...)
	}

	if c.Shell != nil {
		errors = append(errors, c.Shell.Validate()...)
	}

	return errors
}

// Validate checks if the PVMUpdate configuration is valid
func (c *PVMUpdateConfig) Validate() []error {
	var errors []error

	// Validate AutoUpdateInterval
	if c.AutoUpdateInterval != "" {
		if _, err := time.ParseDuration(c.AutoUpdateInterval); err != nil {
			errors = append(errors, ValidateError("AutoUpdateInterval must be a valid duration (e.g., '24h', '1h')"))
		}
	}

	// Validate Repository
	if c.Repository == "" {
		errors = append(errors, ValidateError("Repository cannot be empty"))
	}

	// Validate Channel
	validChannels := map[string]bool{
		"stable":    true,
		"beta":      true,
		"alpha":     true,
		"nightly":   true,
		"developer": true,
	}

	if c.Channel != "" && !validChannels[c.Channel] {
		errors = append(errors, ValidateError("Channel must be one of: stable, beta, alpha, nightly, developer"))
	}

	// Validate MaxRetries
	if c.MaxRetries < 0 {
		errors = append(errors, ValidateError("MaxRetries cannot be negative"))
	}

	// Validate Timeout
	if c.Timeout != "" {
		if _, err := time.ParseDuration(c.Timeout); err != nil {
			errors = append(errors, ValidateError("Timeout must be a valid duration (e.g., '5m', '30s')"))
		}
	}

	return errors
}

// Validate checks if the PVMBinary configuration is valid
func (c *PVMBinaryConfig) Validate() []error {
	var errors []error

	// Validate DefaultInstallMethod
	validMethods := map[string]bool{
		"binary":        true,
		"source":        true,
		"prefer-binary": true,
	}

	if c.DefaultInstallMethod != "" && !validMethods[c.DefaultInstallMethod] {
		errors = append(errors, ValidateError("DefaultInstallMethod must be one of: binary, source, prefer-binary"))
	}

	// Validate BinaryMirrors
	for _, mirror := range c.BinaryMirrors {
		if mirror == "" {
			errors = append(errors, ValidateError("BinaryMirrors cannot contain empty URLs"))
			break
		}
	}

	// Validate CacheRetentionDays
	if c.CacheRetentionDays < 0 {
		errors = append(errors, ValidateError("CacheRetentionDays cannot be negative"))
	}

	// Validate MaxCacheSize
	if c.MaxCacheSize < 0 {
		errors = append(errors, ValidateError("MaxCacheSize cannot be negative"))
	}

	// Validate MaxRetries
	if c.MaxRetries < 0 {
		errors = append(errors, ValidateError("MaxRetries cannot be negative"))
	}

	// Validate Timeout
	if c.Timeout != "" {
		if _, err := time.ParseDuration(c.Timeout); err != nil {
			errors = append(errors, ValidateError("Timeout must be a valid duration (e.g., '5m', '30s')"))
		}
	}

	// Validate BandwidthLimit format if provided
	if c.BandwidthLimit != "" && !isValidMemoryFormat(c.BandwidthLimit) {
		errors = append(errors, ValidateError("BandwidthLimit must be in format like '10MB' or '1GB'"))
	}

	// Validate CustomMirrors
	for i, mirror := range c.CustomMirrors {
		if mirror != nil {
			if mirrorErrors := mirror.Validate(); len(mirrorErrors) > 0 {
				for _, err := range mirrorErrors {
					errors = append(errors, ValidateError(fmt.Sprintf("CustomMirrors[%d]: %s", i, err.Error())))
				}
			}
		}
	}

	return errors
}

// Validate checks if the PVMShell configuration is valid
func (c *PVMShellConfig) Validate() []error {
	// No validation needed for boolean fields - they're always valid
	return []error{}
}

// Validate checks if the PVX configuration is valid
func (c *PVXConfig) Validate() []error {
	var errors []error

	// Validate IsolationLevel
	validLevels := map[string]bool{
		"global": true,
		"local":  true,
		"clean":  true,
	}

	if c.IsolationLevel != "" && !validLevels[c.IsolationLevel] {
		errors = append(errors, ValidateError("IsolationLevel must be one of: global, local, clean"))
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

// Validate checks if the PM configuration is valid
func (c *PMConfig) Validate() []error {
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

	// Validate Backup configuration
	if c.Backup != nil {
		errors = append(errors, c.Backup.Validate()...)
	}

	return errors
}

// Validate checks if the PMBackup configuration is valid
func (c *PMBackupConfig) Validate() []error {
	var errors []error

	// Validate CpanfileBackup mode
	validModes := map[string]bool{
		"off":   true,
		"local": true,
		"cache": true,
	}

	if c.CpanfileBackup != "" && !validModes[c.CpanfileBackup] {
		errors = append(errors, ValidateError("CpanfileBackup must be one of: off, local, cache"))
	}

	// Validate RetentionDays
	if c.RetentionDays < 0 {
		errors = append(errors, ValidateError("RetentionDays cannot be negative"))
	}

	// Validate MaxBackups
	if c.MaxBackups < 0 {
		errors = append(errors, ValidateError("MaxBackups cannot be negative"))
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

	// Validate LSP configuration
	if c.LSP != nil {
		errors = append(errors, c.LSP.Validate()...)
	}

	return errors
}

// Validate checks if the PSCLSPConfig is valid
func (c *PSCLSPConfig) Validate() []error {
	var errors []error

	// Validate LogLevel
	if c.LogLevel != "" {
		validLevels := map[string]bool{
			"debug": true,
			"info":  true,
			"warn":  true,
			"error": true,
		}
		if !validLevels[c.LogLevel] {
			errors = append(errors, ValidateError("LSP log_level must be one of: debug, info, warn, error"))
		}
	}

	// Validate DefaultMode
	if c.DefaultMode != "" {
		validModes := map[string]bool{
			"stdio": true,
			"tcp":   true,
		}
		if !validModes[c.DefaultMode] {
			errors = append(errors, ValidateError("LSP default_mode must be one of: stdio, tcp"))
		}
	}

	// Validate TCPPort
	if c.TCPPort != 0 && (c.TCPPort < 1 || c.TCPPort > 65535) {
		errors = append(errors, ValidateError("LSP tcp_port must be between 1 and 65535"))
	}

	// Validate MaxCacheSize
	if c.MaxCacheSize < 0 {
		errors = append(errors, ValidateError("LSP max_cache_size cannot be negative"))
	}

	// Validate RequestTimeout (must be positive if specified)
	if c.RequestTimeout < 0 {
		errors = append(errors, ValidateError("LSP request_timeout cannot be negative"))
	}

	// Validate DiagnosticsDelay (must be positive if specified)
	if c.DiagnosticsDelay < 0 {
		errors = append(errors, ValidateError("LSP diagnostics_delay cannot be negative"))
	}

	// Validate LogFile path (basic validation)
	if c.LogFile != "" {
		// Check for null bytes and control characters
		for _, r := range c.LogFile {
			if r == 0 || (r < 32 && r != '\t') {
				errors = append(errors, ValidateError("LSP log_file cannot contain null bytes or control characters"))
				break
			}
		}
	}

	// Validate ExcludePatterns
	for _, pattern := range c.ExcludePatterns {
		if pattern == "" {
			errors = append(errors, ValidateError("LSP exclude_patterns cannot contain empty patterns"))
			break
		}
	}

	// Validate IncludeDirectories
	for _, dir := range c.IncludeDirectories {
		if dir == "" {
			errors = append(errors, ValidateError("LSP include_directories cannot contain empty directories"))
			break
		}
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

// Validate checks if the Docs configuration is valid
func (c *DocsConfig) Validate() []error {
	var errors []error

	// Validate Repository format
	if c.Repository == "" {
		errors = append(errors, ValidateError("Docs repository cannot be empty"))
	} else if !strings.Contains(c.Repository, "/") {
		errors = append(errors, ValidateError("Docs repository must be in format 'owner/repo'"))
	}

	// Validate Branch
	if c.Branch == "" {
		errors = append(errors, ValidateError("Docs branch cannot be empty"))
	}

	// Validate CacheTTL
	if c.CacheTTL < -1 {
		errors = append(errors, ValidateError("Docs CacheTTL cannot be less than -1"))
	}

	// Validate UpdateInterval
	if c.UpdateInterval != "" {
		if _, err := time.ParseDuration(c.UpdateInterval); err != nil {
			errors = append(errors, ValidateError("Docs UpdateInterval must be a valid duration (e.g., '24h', '1h')"))
		}
	}

	// Validate MaxRetries
	if c.MaxRetries < 0 {
		errors = append(errors, ValidateError("Docs MaxRetries cannot be negative"))
	}

	// Validate Timeout
	if c.Timeout != "" {
		if _, err := time.ParseDuration(c.Timeout); err != nil {
			errors = append(errors, ValidateError("Docs Timeout must be a valid duration (e.g., '30s', '2m')"))
		}
	}

	// Validate DocumentSources
	for i, source := range c.DocumentSources {
		if source != nil {
			if sourceErrors := source.Validate(); len(sourceErrors) > 0 {
				for _, err := range sourceErrors {
					errors = append(errors, ValidateError(fmt.Sprintf("DocumentSources[%d]: %s", i, err.Error())))
				}
			}
		}
	}

	return errors
}

// Validate checks if the DocsSource configuration is valid
func (c *DocsSourceConfig) Validate() []error {
	var errors []error

	// Validate Name
	if c.Name == "" {
		errors = append(errors, ValidateError("Name cannot be empty"))
	}

	// Validate Type
	validTypes := map[string]bool{
		"github": true,
		"url":    true,
		"local":  true,
	}
	if c.Type != "" && !validTypes[c.Type] {
		errors = append(errors, ValidateError("Type must be one of: github, url, local"))
	}

	// Validate type-specific fields
	switch c.Type {
	case "github":
		if c.Repository == "" {
			errors = append(errors, ValidateError("Repository is required for github source type"))
		} else if !strings.Contains(c.Repository, "/") {
			errors = append(errors, ValidateError("Repository must be in format 'owner/repo' for github source type"))
		}
		if c.Branch == "" {
			errors = append(errors, ValidateError("Branch is required for github source type"))
		}
	case "url":
		if c.BaseURL == "" {
			errors = append(errors, ValidateError("BaseURL is required for url source type"))
		} else if !strings.HasPrefix(c.BaseURL, "http://") && !strings.HasPrefix(c.BaseURL, "https://") {
			errors = append(errors, ValidateError("BaseURL must start with http:// or https://"))
		}
	case "local":
		if c.Path == "" {
			errors = append(errors, ValidateError("Path is required for local source type"))
		}
	}

	// Validate Priority
	if c.Priority < 0 {
		errors = append(errors, ValidateError("Priority cannot be negative"))
	}

	// Validate Categories (optional, no validation needed for empty)
	for _, category := range c.Categories {
		if category == "" {
			errors = append(errors, ValidateError("Categories cannot contain empty strings"))
			break
		}
	}

	return errors
}

// Helper functions for validation

func isValidMemoryFormat(memory string) bool {
	return strings.HasSuffix(memory, "MB") || strings.HasSuffix(memory, "GB") ||
		strings.HasSuffix(memory, "KB") || strings.HasSuffix(memory, "TB")
}

// Validate checks if the PVMCustomMirrorConfig is valid
func (c *PVMCustomMirrorConfig) Validate() []error {
	var errors []error

	// Validate Name
	if c.Name == "" {
		errors = append(errors, ValidateError("Name cannot be empty"))
	}

	// Validate Type
	validTypes := map[string]bool{
		"github-releases": true,
		"jsdelivr":        true,
		"cloudflare-r2":   true,
		"direct":          true,
	}
	if c.Type != "" && !validTypes[c.Type] {
		errors = append(errors, ValidateError("Type must be one of: github-releases, jsdelivr, cloudflare-r2, direct"))
	}

	// Validate BaseURL
	if c.BaseURL == "" {
		errors = append(errors, ValidateError("BaseURL cannot be empty"))
	} else if !strings.HasPrefix(c.BaseURL, "http://") && !strings.HasPrefix(c.BaseURL, "https://") {
		errors = append(errors, ValidateError("BaseURL must start with http:// or https://"))
	}

	// Validate Priority
	if c.Priority < 0 {
		errors = append(errors, ValidateError("Priority cannot be negative"))
	}

	// Validate Timeout format if provided
	if c.Timeout != "" {
		if _, err := time.ParseDuration(c.Timeout); err != nil {
			errors = append(errors, ValidateError("Timeout must be a valid duration (e.g., '30s', '2m')"))
		}
	}

	// Validate MaxRetries
	if c.MaxRetries < 0 {
		errors = append(errors, ValidateError("MaxRetries cannot be negative"))
	}

	// Validate Headers
	for key, value := range c.Headers {
		if key == "" {
			errors = append(errors, ValidateError("Header keys cannot be empty"))
			break
		}
		if value == "" {
			errors = append(errors, ValidateError("Header values cannot be empty"))
			break
		}
	}

	// Validate VersionMapping
	for key, value := range c.VersionMapping {
		if key == "" {
			errors = append(errors, ValidateError("VersionMapping keys cannot be empty"))
			break
		}
		if value == "" {
			errors = append(errors, ValidateError("VersionMapping values cannot be empty"))
			break
		}
	}

	// Validate Auth configuration
	if c.Auth != nil {
		errors = append(errors, c.Auth.Validate()...)
	}

	return errors
}

// Validate checks if the PVMCustomMirrorAuth is valid
func (c *PVMCustomMirrorAuth) Validate() []error {
	var errors []error

	// Validate Type
	validTypes := map[string]bool{
		"none":    true,
		"basic":   true,
		"bearer":  true,
		"api-key": true,
		"oauth2":  true,
	}
	if c.Type != "" && !validTypes[c.Type] {
		errors = append(errors, ValidateError("Auth type must be one of: none, basic, bearer, api-key, oauth2"))
	}

	// Validate type-specific fields
	switch c.Type {
	case "basic":
		if c.Username == "" {
			errors = append(errors, ValidateError("Username is required for basic authentication"))
		}
		if c.Password == "" {
			errors = append(errors, ValidateError("Password is required for basic authentication"))
		}
	case "bearer":
		if c.Token == "" {
			errors = append(errors, ValidateError("Token is required for bearer authentication"))
		}
	case "api-key":
		if c.APIKey == "" {
			errors = append(errors, ValidateError("APIKey is required for api-key authentication"))
		}
		// APIKeyHeader has default, so it's optional
	case "oauth2":
		if c.OAuth2 == nil {
			errors = append(errors, ValidateError("OAuth2 configuration is required for oauth2 authentication"))
		} else {
			errors = append(errors, c.OAuth2.Validate()...)
		}
	}

	return errors
}

// Validate checks if the PVMCustomMirrorOAuth2 is valid
func (c *PVMCustomMirrorOAuth2) Validate() []error {
	var errors []error

	// Validate ClientID
	if c.ClientID == "" {
		errors = append(errors, ValidateError("ClientID cannot be empty for OAuth2"))
	}

	// Validate ClientSecret
	if c.ClientSecret == "" {
		errors = append(errors, ValidateError("ClientSecret cannot be empty for OAuth2"))
	}

	// Validate TokenURL
	if c.TokenURL == "" {
		errors = append(errors, ValidateError("TokenURL cannot be empty for OAuth2"))
	} else if !strings.HasPrefix(c.TokenURL, "http://") && !strings.HasPrefix(c.TokenURL, "https://") {
		errors = append(errors, ValidateError("TokenURL must start with http:// or https://"))
	}

	// Scopes are optional, no validation needed

	return errors
}
