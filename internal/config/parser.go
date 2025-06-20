// ABOUTME: Configuration parsing for the PVM Ecosystem
// ABOUTME: Reads and parses TOML configuration files

package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
	"tamarou.com/pvm/internal/errors"
)

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
	// Create a new configuration with default values
	config := NewDefaultConfig()

	// Parse the TOML data
	decoder := toml.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	err := decoder.Decode(config)
	if err != nil {
		// Check if it's an unknown field error
		if strings.Contains(err.Error(), "undecoded keys") {
			return nil, errors.NewConfigError("002",
				"Unknown field in configuration", err).
				WithLocation(source).
				WithHint("Check for typos in field names")
		}

		return nil, errors.NewConfigError("003",
			"Failed to parse TOML configuration", err).
			WithLocation(source)
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
	if validate {
		err = interpolationEngine.ValidateInterpolatedConfig(config)
		if err != nil {
			return nil, errors.NewConfigError("009",
				"Configuration validation failed after interpolation", err).
				WithLocation(source)
		}
	}

	return config, nil
}

// ParseString parses a TOML configuration from a string
func ParseString(data string) (*Config, error) {
	return ParseBytes([]byte(data), "string")
}

// MergeConfigs merges multiple configurations according to priority
// Later configs in the list take precedence over earlier ones
func MergeConfigs(configs ...*Config) *Config {
	// Start with default configuration
	result := NewDefaultConfig()

	// Merge each config in order
	for _, config := range configs {
		if config == nil {
			continue
		}

		// Merge PVM config
		if config.PVM != nil {
			mergePVMConfig(result.PVM, config.PVM)
		}

		// Merge PVX config
		if config.PVX != nil {
			mergePVXConfig(result.PVX, config.PVX)
		}

		// Merge PVI config
		if config.PVI != nil {
			mergePVIConfig(result.PVI, config.PVI)
		}

		// Merge PSC config
		if config.PSC != nil {
			mergePSCConfig(result.PSC, config.PSC)
		}

		// Merge Project config
		if config.Project != nil {
			mergeProjectConfig(result.Project, config.Project)
		}

		// Merge Build config
		if config.Build != nil {
			mergeBuildConfig(result.Build, config.Build)
		}
	}

	return result
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

func mergePVIConfig(target, source *PVIConfig) {
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
