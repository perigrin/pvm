// ABOUTME: Core logic for determining and displaying currently active Perl version
// ABOUTME: Provides API for showing active version with source attribution

package current

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/perl"
)

// CurrentVersionInfo represents information about the currently active Perl version
type CurrentVersionInfo struct {
	// The resolved version string
	Version string

	// The source of the resolved version
	Source string

	// Human-readable description of the source
	SourceDescription string

	// The path where this version was found (if applicable)
	Path string

	// Whether this is a valid/available version
	IsAvailable bool
}

// GetCurrentVersion retrieves information about the currently active Perl version
func GetCurrentVersion() (*CurrentVersionInfo, error) {
	// Use the existing resolver to get the current version
	resolved, err := perl.ResolveVersion(nil)
	if err != nil {
		return nil, enhanceResolutionError(err)
	}

	if resolved == nil {
		return &CurrentVersionInfo{
			Version:           "none",
			Source:            "none",
			SourceDescription: "no version configured",
			Path:              "",
			IsAvailable:       false,
		}, nil
	}

	// Convert resolver source to user-friendly description
	sourceDesc := formatSourceDescription(resolved.Source, resolved.SourcePath)

	return &CurrentVersionInfo{
		Version:           resolved.Version,
		Source:            string(resolved.Source),
		SourceDescription: sourceDesc,
		Path:              resolved.Path,
		IsAvailable:       true,
	}, nil
}

// formatSourceDescription converts resolver source types to human-readable descriptions
func formatSourceDescription(source perl.ResolutionSource, path string) string {
	switch source {
	case perl.ExplicitVersion:
		return "explicitly specified"
	case perl.ProjectVersionFile:
		if path != "" {
			return fmt.Sprintf("set by %s", path)
		}
		return "set by .perl-version file"
	case perl.ProjectConfig:
		if path != "" {
			return fmt.Sprintf("set by %s", path)
		}
		return "set by project configuration"
	case perl.EnvironmentVariable:
		return "set by environment variable"
	case perl.UserConfig:
		return "set by user configuration"
	case perl.SystemPerlSource:
		if path != "" {
			return fmt.Sprintf("system Perl at %s", path)
		}
		return "system Perl"
	default:
		return string(source)
	}
}

// FormatOutput formats the current version information for display
func (c *CurrentVersionInfo) FormatOutput(bare bool) string {
	if !c.IsAvailable {
		if bare {
			return ""
		}
		return "No Perl version is currently active"
	}

	if bare {
		return c.Version
	}

	return fmt.Sprintf("%s (%s)", c.Version, c.SourceDescription)
}

// GetShortSource returns a short version of the source for compact display
func (c *CurrentVersionInfo) GetShortSource() string {
	if !c.IsAvailable {
		return "none"
	}

	switch perl.ResolutionSource(c.Source) {
	case perl.ExplicitVersion:
		return "explicit"
	case perl.ProjectVersionFile:
		return ".perl-version"
	case perl.ProjectConfig:
		return "project config"
	case perl.EnvironmentVariable:
		return "env var"
	case perl.UserConfig:
		return "user config"
	case perl.SystemPerlSource:
		return "system"
	default:
		return c.Source
	}
}

// GetDetailedInfo returns detailed information about the current version
func (c *CurrentVersionInfo) GetDetailedInfo() map[string]string {
	info := map[string]string{
		"version":     c.Version,
		"source":      c.GetShortSource(),
		"description": c.SourceDescription,
		"available":   fmt.Sprintf("%t", c.IsAvailable),
	}

	if c.Path != "" {
		info["path"] = c.Path
	}

	return info
}

// ValidateCurrentVersion checks if the current version is properly configured and available
func ValidateCurrentVersion() (*CurrentVersionInfo, []string, error) {
	current, err := GetCurrentVersion()
	if err != nil {
		return nil, nil, err
	}

	var warnings []string

	if !current.IsAvailable {
		warnings = append(warnings, "No Perl version is currently configured")
		return current, warnings, nil
	}

	// Check if the resolved version is actually installed
	if current.Version != "" {
		installed, err := perl.IsVersionInstalled(current.Version)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Could not verify if version %s is installed: %v", current.Version, err))
		} else if !installed {
			warnings = append(warnings, fmt.Sprintf("Version %s is configured but not installed", current.Version))
		}
	}

	// Check for common configuration issues
	if perl.ResolutionSource(current.Source) == perl.EnvironmentVariable {
		warnings = append(warnings, "Version is set by environment variable - this may override project settings")
	}

	return current, warnings, nil
}

// CompareWithSystem compares the current version with system Perl
func (c *CurrentVersionInfo) CompareWithSystem() (string, error) {
	if !c.IsAvailable {
		return "No current version to compare", nil
	}

	systemPerl, err := perl.DetectSystemPerl()
	if err != nil {
		return "", fmt.Errorf("failed to detect system Perl: %w", err)
	}

	if systemPerl == nil {
		return "System Perl not found", nil
	}

	if c.Version == systemPerl.Version {
		return fmt.Sprintf("Current version %s matches system Perl", c.Version), nil
	}

	// Parse versions for comparison if possible
	currentVer, err1 := perl.ParseVersion(c.Version)
	systemVer, err2 := perl.ParseVersion(systemPerl.Version)

	if err1 != nil || err2 != nil {
		return fmt.Sprintf("Current: %s, System: %s", c.Version, systemPerl.Version), nil
	}

	comparison := currentVer.Compare(systemVer)
	var relationship string
	switch {
	case comparison > 0:
		relationship = "newer than"
	case comparison < 0:
		relationship = "older than"
	default:
		relationship = "same as"
	}

	return fmt.Sprintf("Current %s is %s system %s", c.Version, relationship, systemPerl.Version), nil
}

// enhanceResolutionError provides more helpful error messages for version resolution failures
func enhanceResolutionError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Provide helpful context for common error scenarios
	if strings.Contains(errMsg, "no .perl-version file found") {
		return fmt.Errorf("failed to resolve current Perl version: %w\n\nSuggestion: Create a .perl-version file or set a global version with 'pvm global <version>'", err)
	}

	if strings.Contains(errMsg, "not available") {
		return fmt.Errorf("failed to resolve current Perl version: %w\n\nSuggestion: Install the required version with 'pvm install <version>' or check available versions with 'pvm available'", err)
	}

	if strings.Contains(errMsg, "no versions available") {
		return fmt.Errorf("failed to resolve current Perl version: %w\n\nSuggestion: Install a Perl version with 'pvm install 5.38.0' or import system Perl with 'pvm import-system'", err)
	}

	// For other errors, just add a generic helpful message
	return fmt.Errorf("failed to resolve current Perl version: %w\n\nFor help with version resolution, run 'pvm help' or 'pvm resolve' to debug", err)
}

// ShouldShowDirectoryChangeAlerts checks if directory change alerts should be shown
func ShouldShowDirectoryChangeAlerts() bool {
	cfg, err := config.LoadEffectiveConfig()
	if err != nil {
		// If we can't load config, default to showing alerts
		return true
	}

	if cfg.PVM == nil || cfg.PVM.Shell == nil {
		// If shell config is not set, default to showing alerts
		return true
	}

	return cfg.PVM.Shell.ShowDirectoryChangeAlerts
}
