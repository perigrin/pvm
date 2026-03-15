// ABOUTME: Type definitions for current version functionality
// ABOUTME: Provides structured data types for version information and formatting options

package current

// OutputFormat represents different output formatting options
type OutputFormat int

const (
	// FormatDefault shows version with source attribution
	FormatDefault OutputFormat = iota
	// FormatBare shows only the version string
	FormatBare
	// FormatJSON shows structured JSON output
	FormatJSON
	// FormatDetailed shows comprehensive version information
	FormatDetailed
)

// DisplayOptions controls how current version information is displayed
type DisplayOptions struct {
	// Format specifies the output format
	Format OutputFormat

	// ShowPath includes file paths in output when available
	ShowPath bool

	// ShowWarnings includes validation warnings in output
	ShowWarnings bool

	// ShowComparison includes comparison with system Perl
	ShowComparison bool

	// Validate performs validation checks on the current version
	Validate bool
}

// DefaultDisplayOptions returns sensible default display options
func DefaultDisplayOptions() *DisplayOptions {
	return &DisplayOptions{
		Format:         FormatDefault,
		ShowPath:       false,
		ShowWarnings:   true,
		ShowComparison: false,
		Validate:       false,
	}
}

// BareDisplayOptions returns options for bare/scripting output
func BareDisplayOptions() *DisplayOptions {
	return &DisplayOptions{
		Format:         FormatBare,
		ShowPath:       false,
		ShowWarnings:   false,
		ShowComparison: false,
		Validate:       false,
	}
}

// DetailedDisplayOptions returns options for detailed output
func DetailedDisplayOptions() *DisplayOptions {
	return &DisplayOptions{
		Format:         FormatDetailed,
		ShowPath:       true,
		ShowWarnings:   true,
		ShowComparison: true,
		Validate:       true,
	}
}

// VersionStatus represents the status of the current version configuration
type VersionStatus int

const (
	// StatusOK indicates the version is properly configured and available
	StatusOK VersionStatus = iota
	// StatusNotConfigured indicates no version is configured
	StatusNotConfigured
	// StatusNotInstalled indicates the configured version is not installed
	StatusNotInstalled
	// StatusInvalid indicates the configuration is invalid
	StatusInvalid
)

// String returns a human-readable string for the version status
func (s VersionStatus) String() string {
	switch s {
	case StatusOK:
		return "OK"
	case StatusNotConfigured:
		return "Not Configured"
	case StatusNotInstalled:
		return "Not Installed"
	case StatusInvalid:
		return "Invalid"
	default:
		return "Unknown"
	}
}

// CurrentVersionSummary provides a comprehensive summary of the current version state
type CurrentVersionSummary struct {
	// Version information
	Version *CurrentVersionInfo

	// Status of the current configuration
	Status VersionStatus

	// Any warnings or issues found
	Warnings []string

	// Comparison with system Perl (if requested)
	SystemComparison string

	// Additional context information
	Context map[string]string
}
