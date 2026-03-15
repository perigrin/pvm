// ABOUTME: Configuration accessors for the PVM Ecosystem
// ABOUTME: Provides helper functions for accessing specific configuration values

package config

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GetString returns a string configuration value, with environment variable expansion
func (c *Config) GetString(section, key string) string {
	var value string

	switch section {
	case "pvm":
		if c.PVM != nil {
			value = getPVMString(c.PVM, key)
		}
	case "pvx":
		if c.PVX != nil {
			value = getPVXString(c.PVX, key)
		}
	case "pm":
		if c.PM != nil {
			value = getPVIString(c.PM, key)
		}
	case "psc":
		if c.PSC != nil {
			value = getPSCString(c.PSC, key)
		}
	}

	// Expand environment variables in the value
	return expandEnvironmentVariables(value)
}

// GetInt returns an integer configuration value
func (c *Config) GetInt(section, key string) int {
	switch section {
	case "pvm":
		if c.PVM != nil {
			return getPVMInt(c.PVM, key)
		}
	case "pvx":
		if c.PVX != nil {
			return getPVXInt(c.PVX, key)
		}
	}
	return 0
}

// GetBool returns a boolean configuration value
func (c *Config) GetBool(section, key string) bool {
	switch section {
	case "pvm":
		if c.PVM != nil {
			return getPVMBool(c.PVM, key)
		}
	case "pvx":
		if c.PVX != nil {
			return getPVXBool(c.PVX, key)
		}
	case "pm":
		if c.PM != nil {
			return getPVIBool(c.PM, key)
		}
	case "psc":
		if c.PSC != nil {
			return getPSCBool(c.PSC, key)
		}
	}
	return false
}

// GetStringSlice returns a string slice configuration value
func (c *Config) GetStringSlice(section, key string) []string {
	if section == "psc" && c.PSC != nil {
		return getPSCStringSlice(c.PSC, key)
	}
	return nil
}

// GetStringMap returns a string map configuration value
func (c *Config) GetStringMap(section, key string) map[string]string {
	if section == "pvm" && c.PVM != nil {
		return getPVMStringMap(c.PVM, key)
	}
	return nil
}

// Helper functions for accessing specific fields

func getPVMString(c *PVMConfig, key string) string {
	switch key {
	case "default_perl":
		return c.DefaultPerl
	case "download_mirror":
		return c.DownloadMirror
	case "patches_dir":
		return c.PatchesDir
	case "compiler":
		return c.Compiler
	}
	return ""
}

func getPVXString(c *PVXConfig, key string) string {
	switch key {
	case "isolation_level":
		return c.IsolationLevel
	case "max_memory":
		return c.MaxMemory
	}
	return ""
}

func getPVIString(c *PMConfig, key string) string {
	switch key {
	case "preferred_installer":
		return c.PreferredInstaller
	case "default_mirror":
		return c.DefaultMirror
	}
	return ""
}

func getPSCString(c *PSCConfig, key string) string {
	if key == "type_definitions_path" {
		return c.TypeDefinitionsPath
	}
	return ""
}

func getPVMInt(c *PVMConfig, key string) int {
	if key == "build_jobs" {
		return c.BuildJobs
	}
	return 0
}

func getPVXInt(c *PVXConfig, key string) int {
	if key == "timeout" {
		return c.Timeout
	}
	return 0
}

func getPVMBool(c *PVMConfig, key string) bool {
	if key == "run_tests" {
		return c.RunTests
	}
	return false
}

func getPVXBool(c *PVXConfig, key string) bool {
	switch key {
	case "cache_modules":
		return c.CacheModules
	case "cleanup_after":
		return c.CleanupAfter
	case "always_install_deps":
		return c.AlwaysInstallDeps
	}
	return false
}

func getPVIBool(c *PMConfig, key string) bool {
	switch key {
	case "test_during_install":
		return c.TestDuringInstall
	case "cache_modules":
		return c.CacheModules
	case "force_reinstall":
		return c.ForceReinstall
	case "check_signatures":
		return c.CheckSignatures
	}
	return false
}

func getPSCBool(c *PSCConfig, key string) bool {
	switch key {
	case "strict_mode":
		return c.StrictMode
	case "generate_missing_types":
		return c.GenerateMissingTypes
	case "check_before_run":
		return c.CheckBeforeRun
	}
	return false
}

func getPSCStringSlice(c *PSCConfig, key string) []string {
	if key == "watch_exclude" {
		return c.WatchExclude
	}
	return nil
}

func getPVMStringMap(c *PVMConfig, key string) map[string]string {
	if key == "version_aliases" {
		return c.VersionAliases
	}
	return nil
}

// getXDGFallback returns the XDG Base Directory specification fallback for XDG environment variables
func getXDGFallback(envVar string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		return ""
	}

	switch envVar {
	case "XDG_CACHE_HOME":
		return filepath.Join(homeDir, ".cache")
	case "XDG_DATA_HOME":
		return filepath.Join(homeDir, ".local", "share")
	case "XDG_CONFIG_HOME":
		return filepath.Join(homeDir, ".config")
	case "XDG_STATE_HOME":
		return filepath.Join(homeDir, ".local", "state")
	default:
		return ""
	}
}

// expandEnvironmentVariables expands environment variables in configuration values
func expandEnvironmentVariables(value string) string {
	if value == "" {
		return value
	}

	// Handle cases where the entire value is a single variable like $VAR or ${VAR}
	if strings.HasPrefix(value, "$") && !strings.Contains(value[1:], "$") {
		envVar := value[1:]
		// Check for complex expressions like ${VAR}
		if strings.HasPrefix(envVar, "{") && strings.HasSuffix(envVar, "}") {
			envVar = envVar[1 : len(envVar)-1]
			// Entire value is ${VAR}
			envValue, exists := os.LookupEnv(envVar)
			if exists {
				return envValue
			}
			// Try XDG fallback for unset XDG variables
			if fallback := getXDGFallback(envVar); fallback != "" {
				return fallback
			}
			return value
		}
		// Check if this is a simple $VAR without any other characters
		if !strings.ContainsAny(envVar, "/\\:. ") {
			envValue, exists := os.LookupEnv(envVar)
			if exists {
				return envValue
			}
			// Try XDG fallback for unset XDG variables
			if fallback := getXDGFallback(envVar); fallback != "" {
				return fallback
			}
			return value
		}
	}

	// Handle embedded variables like /path/$VAR/subdir
	re := regexp.MustCompile(`\$\{([^}]+)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)
	expanded := re.ReplaceAllStringFunc(value, func(match string) string {
		var envVar string
		if strings.HasPrefix(match, "${") {
			// ${VAR} format
			envVar = match[2 : len(match)-1]
		} else {
			// $VAR format
			envVar = match[1:]
		}

		envValue, exists := os.LookupEnv(envVar)
		if exists {
			return envValue
		}
		// Try XDG fallback for unset XDG variables
		if fallback := getXDGFallback(envVar); fallback != "" {
			return fallback
		}
		return match // Return original if env var not found
	})
	return expanded
}

// GetStringWithDefault returns a string configuration value with a default fallback
func (c *Config) GetStringWithDefault(section, key, defaultValue string) string {
	value := c.GetString(section, key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetIntWithDefault returns an integer configuration value with a default fallback
func (c *Config) GetIntWithDefault(section, key string, defaultValue int) int {
	value := c.GetInt(section, key)
	if value == 0 {
		return defaultValue
	}
	return value
}

// GetBoolWithDefault returns a boolean configuration value with a default fallback
func (c *Config) GetBoolWithDefault(section, key string, defaultValue bool) bool {
	// Note: for boolean, we can't distinguish between false and unset
	// This is a limitation of the current implementation
	return c.GetBool(section, key)
}

// HasSection returns true if the specified section exists in the configuration
func (c *Config) HasSection(section string) bool {
	switch section {
	case "pvm":
		return c.PVM != nil
	case "pvx":
		return c.PVX != nil
	case "pm":
		return c.PM != nil
	case "psc":
		return c.PSC != nil
	}
	return false
}

// HasKey returns true if the specified key exists in the given section
func (c *Config) HasKey(section, key string) bool {
	switch section {
	case "pvm":
		if c.PVM != nil {
			return hasPVMKey(c.PVM, key)
		}
	case "pvx":
		if c.PVX != nil {
			return hasPVXKey(c.PVX, key)
		}
	case "pm":
		if c.PM != nil {
			return hasPVIKey(c.PM, key)
		}
	case "psc":
		if c.PSC != nil {
			return hasPSCKey(c.PSC, key)
		}
	}
	return false
}

// Helper functions to check if keys exist

func hasPVMKey(c *PVMConfig, key string) bool {
	switch key {
	case "default_perl", "download_mirror", "patches_dir", "compiler",
		"build_jobs", "run_tests", "version_aliases":
		return true
	}
	return false
}

func hasPVXKey(c *PVXConfig, key string) bool {
	switch key {
	case "isolation_level", "max_memory", "timeout", "cache_modules",
		"cleanup_after", "always_install_deps", "isolated_output",
		"save_output_dir", "custom_module_path":
		return true
	}
	return false
}

func hasPVIKey(c *PMConfig, key string) bool {
	switch key {
	case "preferred_installer", "default_mirror", "additional_mirrors",
		"metadata_source", "metadata_url", "cache_dir", "cache_ttl",
		"test_during_install", "cache_modules", "force_reinstall",
		"check_signatures", "disable_network":
		return true
	}
	return false
}

func hasPSCKey(c *PSCConfig, key string) bool {
	switch key {
	case "type_definitions_path", "strict_mode", "watch_exclude",
		"generate_missing_types", "check_before_run":
		return true
	}
	return false
}
