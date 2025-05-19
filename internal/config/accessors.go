// ABOUTME: Configuration accessors for the PVM Ecosystem
// ABOUTME: Provides helper functions for accessing specific configuration values

package config

import (
	"os"
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
	case "pvi":
		if c.PVI != nil {
			value = getPVIString(c.PVI, key)
		}
	case "psc":
		if c.PSC != nil {
			value = getPSCString(c.PSC, key)
		}
	}

	// Expand environment variables if value starts with $
	if strings.HasPrefix(value, "$") {
		// Strip the $ and get the environment variable
		envVar := value[1:]

		// Check for complex expressions like ${VAR}
		if strings.HasPrefix(envVar, "{") && strings.HasSuffix(envVar, "}") {
			envVar = envVar[1 : len(envVar)-1]
		}

		if envValue := os.Getenv(envVar); envValue != "" {
			return envValue
		}
	}

	return value
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
	case "pvi":
		if c.PVI != nil {
			return getPVIBool(c.PVI, key)
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

func getPVIString(c *PVIConfig, key string) string {
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

func getPVIBool(c *PVIConfig, key string) bool {
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
