// ABOUTME: Configuration parsing for the PVM Ecosystem
// ABOUTME: Reads and parses TOML configuration files

package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
	"tamarou.com/pvm/internal/errors"
)

// ParseResult holds the result of parsing configuration with any warnings
type ParseResult struct {
	Config   *Config
	Warnings []string
}

// ParseOptions controls parsing behavior
type ParseOptions struct {
	StrictParsing bool // Whether to fail on unknown fields
	Validate      bool // Whether to validate after parsing
}

// DefaultParseOptions returns the default parsing options
func DefaultParseOptions() ParseOptions {
	return ParseOptions{
		StrictParsing: false, // Default to permissive parsing
		Validate:      true,
	}
}

// ParseFile reads and parses a TOML configuration file
func ParseFile(path string) (*Config, error) {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.NewConfigError("001",
			"Failed to read configuration file", err).
			WithLocation(path)
	}

	return ParseBytes(data, path)
}

// ParseFileWithWarnings reads and parses a TOML configuration file, returning warnings
func ParseFileWithWarnings(path string) (*ParseResult, error) {
	return ParseFileWithOptions(path, DefaultParseOptions())
}

// ParseFileWithOptions reads and parses a TOML configuration file with custom options
func ParseFileWithOptions(path string, options ParseOptions) (*ParseResult, error) {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.NewConfigError("001",
			"Failed to read configuration file", err).
			WithLocation(path)
	}

	return ParseBytesWithAdvancedOptions(data, path, options)
}

// ParseFileWithoutValidation parses a configuration file without validation
func ParseFileWithoutValidation(path string) (*Config, error) {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.NewConfigError("001",
			"Failed to read configuration file", err).
			WithLocation(path)
	}

	return ParseBytesWithOptions(data, path, false)
}

// ParseBytes parses a TOML configuration from a byte slice
func ParseBytes(data []byte, source string) (*Config, error) {
	return ParseBytesWithOptions(data, source, true)
}

// ParseBytesWithOptions parses a TOML configuration from a byte slice with options
func ParseBytesWithOptions(data []byte, source string, validate bool) (*Config, error) {
	options := ParseOptions{
		StrictParsing: false, // Maintain backward compatibility with permissive parsing
		Validate:      validate,
	}
	result, err := ParseBytesWithAdvancedOptions(data, source, options)
	if err != nil {
		return nil, err
	}
	return result.Config, nil
}

// ParseBytesWithAdvancedOptions parses a TOML configuration with advanced options and warning collection
func ParseBytesWithAdvancedOptions(data []byte, source string, options ParseOptions) (*ParseResult, error) {
	// Create a new configuration with default values
	config := NewDefaultConfig()
	warnings := []string{}

	// Parse the TOML data
	decoder := toml.NewDecoder(bytes.NewReader(data))

	// Only enable strict parsing if explicitly requested
	if options.StrictParsing {
		decoder.DisallowUnknownFields()
	}

	err := decoder.Decode(config)
	if err != nil {
		// Check if it's an unknown field error
		if strings.Contains(err.Error(), "undecoded keys") {
			if options.StrictParsing {
				// In strict mode, unknown fields are still errors
				return nil, errors.NewConfigError("002",
					"Unknown field in configuration", err).
					WithLocation(source).
					WithHint("Check for typos in field names or disable strict parsing")
			} else {
				// In permissive mode, this shouldn't happen, but handle it gracefully
				warnings = append(warnings, fmt.Sprintf("Unknown fields detected: %s", err.Error()))
			}
		} else {
			// All other parsing errors are still fatal
			return nil, errors.NewConfigError("003",
				"Failed to parse TOML configuration", err).
				WithLocation(source)
		}
	}

	// In permissive mode, collect unknown fields for warnings
	if !options.StrictParsing {
		unknownFields := collectUnknownFields(data, config)
		for _, field := range unknownFields {
			warnings = append(warnings, fmt.Sprintf("Unknown configuration field: %s", field))
		}
	}

	// Perform environment variable interpolation
	interpolationEngine := NewInterpolationEngine()
	config, err = interpolationEngine.InterpolateConfig(config)
	if err != nil {
		return nil, errors.NewConfigError("008",
			"Environment variable interpolation failed", err).
			WithLocation(source)
	}

	// Validate the configuration after interpolation (if enabled)
	if options.Validate {
		err = interpolationEngine.ValidateInterpolatedConfig(config)
		if err != nil {
			return nil, errors.NewConfigError("009",
				"Configuration validation failed after interpolation", err).
				WithLocation(source)
		}
	}

	return &ParseResult{
		Config:   config,
		Warnings: warnings,
	}, nil
}

// collectUnknownFields analyzes the TOML data to find unknown fields
func collectUnknownFields(data []byte, config *Config) []string {
	var unknownFields []string

	// Parse into a generic map to find all keys
	var genericData map[string]interface{}
	err := toml.Unmarshal(data, &genericData)
	if err != nil {
		return unknownFields // If we can't parse generically, return empty list
	}

	// Check top-level sections
	knownSections := map[string]bool{
		"pvm":        true,
		"pvx":        true,
		"pvi":        true,
		"psc":        true,
		"mcp_server": true,
		"project":    true,
		"build":      true,
	}

	for key := range genericData {
		if !knownSections[key] {
			unknownFields = append(unknownFields, key)
		}
	}

	return unknownFields
}

// ParseString parses a TOML configuration from a string
func ParseString(data string) (*Config, error) {
	return ParseBytes([]byte(data), "string")
}

// ParseStringWithWarnings parses a TOML configuration from a string, returning warnings
func ParseStringWithWarnings(data string) (*ParseResult, error) {
	return ParseBytesWithAdvancedOptions([]byte(data), "string", DefaultParseOptions())
}

// Helper functions for merging configurations

func mergePVMConfig(target, source *PVMConfig) {
	// For all fields, source takes precedence over target

	// String fields
	if source.DefaultPerl != "" {
		target.DefaultPerl = source.DefaultPerl
	}

	// Integer fields (always merge, unlike the old code that only merged non-zero values)
	// This allows zero values to override non-zero values in the target
	target.BuildJobs = source.BuildJobs

	if source.DownloadMirror != "" {
		target.DownloadMirror = source.DownloadMirror
	}

	if source.PatchesDir != "" {
		target.PatchesDir = source.PatchesDir
	}

	if source.Compiler != "" {
		target.Compiler = source.Compiler
	}

	// Boolean fields (always merge)
	target.RunTests = source.RunTests

	// Merge maps (add or replace entries)
	if source.VersionAliases != nil {
		if target.VersionAliases == nil {
			target.VersionAliases = make(map[string]string)
		}

		for key, value := range source.VersionAliases {
			target.VersionAliases[key] = value
		}
	}

	// Merge remote sources — additive, same-name entry from source wins
	if len(source.Remotes) > 0 {
		target.Remotes = mergeRemoteConfigs(target.Remotes, source.Remotes)
	}

	// Merge Update configuration
	if source.Update != nil {
		if target.Update == nil {
			target.Update = &PVMUpdateConfig{}
		}
		mergePVMUpdateConfig(target.Update, source.Update)
	}
}

func mergePVMUpdateConfig(target, source *PVMUpdateConfig) {
	// For all fields, source takes precedence over target

	// Boolean fields (always merge)
	target.AutoUpdateEnabled = source.AutoUpdateEnabled
	target.BackupEnabled = source.BackupEnabled
	target.AutoRollbackEnabled = source.AutoRollbackEnabled
	target.CheckPrerelease = source.CheckPrerelease
	target.NotificationsEnabled = source.NotificationsEnabled
	target.SecurityUpdatesOnly = source.SecurityUpdatesOnly
	target.SkipChecksums = source.SkipChecksums

	// String fields
	if source.AutoUpdateInterval != "" {
		target.AutoUpdateInterval = source.AutoUpdateInterval
	}

	if source.Repository != "" {
		target.Repository = source.Repository
	}

	if source.Channel != "" {
		target.Channel = source.Channel
	}

	if source.GitHubToken != "" {
		target.GitHubToken = source.GitHubToken
	}

	if source.Timeout != "" {
		target.Timeout = source.Timeout
	}

	// Integer fields (always merge)
	target.MaxRetries = source.MaxRetries
}

func mergePVXConfig(target, source *PVXConfig) {
	// For all fields, source takes precedence over target

	// String fields
	if source.IsolationLevel != "" {
		target.IsolationLevel = source.IsolationLevel
	}

	if source.MaxMemory != "" {
		target.MaxMemory = source.MaxMemory
	}

	// Integer fields
	target.Timeout = source.Timeout

	// Boolean fields (always merge)
	target.CacheModules = source.CacheModules
	target.CleanupAfter = source.CleanupAfter
	target.AlwaysInstallDeps = source.AlwaysInstallDeps
}

func mergePMConfig(target, source *PMConfig) {
	// For all fields, source takes precedence over target

	// String fields
	if source.PreferredInstaller != "" {
		target.PreferredInstaller = source.PreferredInstaller
	}

	if source.DefaultMirror != "" {
		target.DefaultMirror = source.DefaultMirror
	}

	// Boolean fields (always merge)
	target.TestDuringInstall = source.TestDuringInstall
	target.ForceReinstall = source.ForceReinstall
	target.CacheModules = source.CacheModules
	target.CheckSignatures = source.CheckSignatures
}

func mergePSCConfig(target, source *PSCConfig) {
	// For all fields, source takes precedence over target

	// String fields
	if source.TypeDefinitionsPath != "" {
		target.TypeDefinitionsPath = source.TypeDefinitionsPath
	}

	// Boolean fields (always merge)
	target.StrictMode = source.StrictMode
	target.GenerateMissingTypes = source.GenerateMissingTypes
	target.CheckBeforeRun = source.CheckBeforeRun

	// Merge arrays (replace entirely)
	if source.WatchExclude != nil {
		target.WatchExclude = make([]string, len(source.WatchExclude))
		copy(target.WatchExclude, source.WatchExclude)
	}

	// Merge nested LSP configuration
	if source.LSP != nil {
		if target.LSP == nil {
			target.LSP = &PSCLSPConfig{}
		}
		mergePSCLSPConfig(target.LSP, source.LSP)
	}
}

func mergePSCLSPConfig(target, source *PSCLSPConfig) {
	// For all fields, source takes precedence over target

	// String fields
	if source.LogFile != "" {
		target.LogFile = source.LogFile
	}
	if source.LogLevel != "" {
		target.LogLevel = source.LogLevel
	}
	if source.DefaultMode != "" {
		target.DefaultMode = source.DefaultMode
	}

	// Integer fields (always merge, allows zero to override)
	target.TCPPort = source.TCPPort
	target.MaxCacheSize = source.MaxCacheSize

	// Duration fields (always merge)
	target.RequestTimeout = source.RequestTimeout
	target.DiagnosticsDelay = source.DiagnosticsDelay

	// Boolean fields (always merge)
	target.Verbose = source.Verbose
	target.EnableHover = source.EnableHover
	target.EnableCompletion = source.EnableCompletion
	target.EnableDefinition = source.EnableDefinition
	target.EnableReferences = source.EnableReferences
	target.EnableFormatting = source.EnableFormatting
	target.WorkspaceSymbols = source.WorkspaceSymbols
	target.CrossFileAnalysis = source.CrossFileAnalysis

	// Merge arrays (replace entirely)
	if source.ExcludePatterns != nil {
		target.ExcludePatterns = make([]string, len(source.ExcludePatterns))
		copy(target.ExcludePatterns, source.ExcludePatterns)
	}
	if source.IncludeDirectories != nil {
		target.IncludeDirectories = make([]string, len(source.IncludeDirectories))
		copy(target.IncludeDirectories, source.IncludeDirectories)
	}
}

func mergeProjectConfig(target, source *ProjectConfig) {
	// For all fields, source takes precedence over target

	// String fields
	if source.Name != "" {
		target.Name = source.Name
	}
	if source.Version != "" {
		target.Version = source.Version
	}
	if source.PerlVersion != "" {
		target.PerlVersion = source.PerlVersion
	}
	if source.Description != "" {
		target.Description = source.Description
	}
	if source.License != "" {
		target.License = source.License
	}
	if source.Homepage != "" {
		target.Homepage = source.Homepage
	}
	if source.BugTracker != "" {
		target.BugTracker = source.BugTracker
	}
	if source.Repository != "" {
		target.Repository = source.Repository
	}

	// Array fields (replace entirely)
	if source.Author != nil {
		target.Author = make([]string, len(source.Author))
		copy(target.Author, source.Author)
	}
}

func mergeBuildConfig(target, source *BuildConfig) {
	// For all fields, source takes precedence over target

	// String fields
	if source.Mode != "" {
		target.Mode = source.Mode
	}
	if source.OutputDir != "" {
		target.OutputDir = source.OutputDir
	}

	// Boolean fields (always merge)
	target.CleanBeforeBuild = source.CleanBeforeBuild

	// Nested config fields
	if source.TypeCheck != nil {
		if target.TypeCheck == nil {
			target.TypeCheck = &BuildTypeCheckConfig{}
		}
		mergeBuildTypeCheckConfig(target.TypeCheck, source.TypeCheck)
	}

	if source.Files != nil {
		if target.Files == nil {
			target.Files = &BuildFilesConfig{}
		}
		mergeBuildFilesConfig(target.Files, source.Files)
	}

	if source.Distribution != nil {
		if target.Distribution == nil {
			target.Distribution = &BuildDistributionConfig{}
		}
		mergeBuildDistributionConfig(target.Distribution, source.Distribution)
	}
}

func mergeBuildTypeCheckConfig(target, source *BuildTypeCheckConfig) {
	// Boolean fields (always merge)
	target.Strict = source.Strict
	target.Experimental = source.Experimental

	// String fields
	if source.TargetPerl != "" {
		target.TargetPerl = source.TargetPerl
	}
}

func mergeBuildFilesConfig(target, source *BuildFilesConfig) {
	// Array fields (replace entirely)
	if source.Include != nil {
		target.Include = make([]string, len(source.Include))
		copy(target.Include, source.Include)
	}
	if source.Exclude != nil {
		target.Exclude = make([]string, len(source.Exclude))
		copy(target.Exclude, source.Exclude)
	}
	if source.WatchDirs != nil {
		target.WatchDirs = make([]string, len(source.WatchDirs))
		copy(target.WatchDirs, source.WatchDirs)
	}
}

func mergeBuildDistributionConfig(target, source *BuildDistributionConfig) {
	// Boolean fields (always merge)
	target.IncludeTests = source.IncludeTests
	target.IncludeScripts = source.IncludeScripts

	// String fields
	if source.Installer != "" {
		target.Installer = source.Installer
	}
}

// SaveToFile saves a configuration to a file
func SaveToFile(config *Config, path string) error {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.NewConfigError("005",
			"Failed to create configuration directory", err).
			WithLocation(dir)
	}

	// Marshal the configuration to TOML
	data, err := toml.Marshal(config)
	if err != nil {
		return errors.NewConfigError("006",
			"Failed to marshal configuration", err)
	}

	// Write the file
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return errors.NewConfigError("007",
			"Failed to write configuration file", err).
			WithLocation(path)
	}

	return nil
}
