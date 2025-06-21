// ABOUTME: Unit tests for current version functionality
// ABOUTME: Tests version information retrieval, formatting, and error handling

package current

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCurrentVersionInfo_FormatOutput(t *testing.T) {
	tests := []struct {
		name     string
		info     *CurrentVersionInfo
		bare     bool
		expected string
	}{
		{
			name: "available version with source",
			info: &CurrentVersionInfo{
				Version:           "5.38.0",
				Source:            "user_config",
				SourceDescription: "set by user configuration",
				Path:              "",
				IsAvailable:       true,
			},
			bare:     false,
			expected: "5.38.0 (set by user configuration)",
		},
		{
			name: "available version bare output",
			info: &CurrentVersionInfo{
				Version:           "5.38.0",
				Source:            "user_config",
				SourceDescription: "set by user configuration",
				Path:              "",
				IsAvailable:       true,
			},
			bare:     true,
			expected: "5.38.0",
		},
		{
			name: "no version available",
			info: &CurrentVersionInfo{
				Version:           "none",
				Source:            "none",
				SourceDescription: "no version configured",
				Path:              "",
				IsAvailable:       false,
			},
			bare:     false,
			expected: "No Perl version is currently active",
		},
		{
			name: "no version available bare",
			info: &CurrentVersionInfo{
				Version:           "none",
				Source:            "none",
				SourceDescription: "no version configured",
				Path:              "",
				IsAvailable:       false,
			},
			bare:     true,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.info.FormatOutput(tt.bare)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("FormatOutput() = %q, want to contain %q", result, tt.expected)
			}
		})
	}
}

func TestCurrentVersionInfo_GetShortSource(t *testing.T) {
	tests := []struct {
		name     string
		info     *CurrentVersionInfo
		expected string
	}{
		{
			name: "user config source",
			info: &CurrentVersionInfo{
				Source:      "user_config",
				IsAvailable: true,
			},
			expected: "user config",
		},
		{
			name: "project version file source",
			info: &CurrentVersionInfo{
				Source:      "project_file",
				IsAvailable: true,
			},
			expected: ".perl-version",
		},
		{
			name: "system source",
			info: &CurrentVersionInfo{
				Source:      "system_perl",
				IsAvailable: true,
			},
			expected: "system",
		},
		{
			name: "not available",
			info: &CurrentVersionInfo{
				Source:      "none",
				IsAvailable: false,
			},
			expected: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.info.GetShortSource()
			if result != tt.expected {
				t.Errorf("GetShortSource() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatCurrentVersion(t *testing.T) {
	info := &CurrentVersionInfo{
		Version:           "5.38.0",
		Source:            "user_config",
		SourceDescription: "set by user configuration",
		Path:              "/home/user/.pvm/config.toml",
		IsAvailable:       true,
	}

	tests := []struct {
		name     string
		options  *DisplayOptions
		validate func(string, error)
	}{
		{
			name:    "default format",
			options: DefaultDisplayOptions(),
			validate: func(output string, err error) {
				if err != nil {
					t.Errorf("FormatCurrentVersion() error = %v", err)
					return
				}
				if !strings.Contains(output, "5.38.0") {
					t.Errorf("Expected version in output, got: %s", output)
				}
				if !strings.Contains(output, "set by user configuration") {
					t.Errorf("Expected source description in output, got: %s", output)
				}
			},
		},
		{
			name:    "bare format",
			options: BareDisplayOptions(),
			validate: func(output string, err error) {
				if err != nil {
					t.Errorf("FormatCurrentVersion() error = %v", err)
					return
				}
				if output != "5.38.0" {
					t.Errorf("Expected bare version, got: %s", output)
				}
			},
		},
		{
			name: "JSON format",
			options: func() *DisplayOptions {
				opts := DefaultDisplayOptions()
				opts.Format = FormatJSON
				return opts
			}(),
			validate: func(output string, err error) {
				if err != nil {
					t.Errorf("FormatCurrentVersion() error = %v", err)
					return
				}

				var jsonData map[string]interface{}
				if err := json.Unmarshal([]byte(output), &jsonData); err != nil {
					t.Errorf("Expected valid JSON, got error: %v", err)
					return
				}

				if jsonData["version"] != "5.38.0" {
					t.Errorf("Expected version 5.38.0 in JSON, got: %v", jsonData["version"])
				}

				if jsonData["available"] != true {
					t.Errorf("Expected available=true in JSON, got: %v", jsonData["available"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := FormatCurrentVersion(info, tt.options)
			tt.validate(output, err)
		})
	}
}

func TestDisplayOptions(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() *DisplayOptions
		expected DisplayOptions
	}{
		{
			name: "default options",
			fn:   DefaultDisplayOptions,
			expected: DisplayOptions{
				Format:         FormatDefault,
				ShowPath:       false,
				ShowWarnings:   true,
				ShowComparison: false,
				Validate:       false,
			},
		},
		{
			name: "bare options",
			fn:   BareDisplayOptions,
			expected: DisplayOptions{
				Format:         FormatBare,
				ShowPath:       false,
				ShowWarnings:   false,
				ShowComparison: false,
				Validate:       false,
			},
		},
		{
			name: "detailed options",
			fn:   DetailedDisplayOptions,
			expected: DisplayOptions{
				Format:         FormatDetailed,
				ShowPath:       true,
				ShowWarnings:   true,
				ShowComparison: true,
				Validate:       true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if result.Format != tt.expected.Format {
				t.Errorf("Format = %v, want %v", result.Format, tt.expected.Format)
			}
			if result.ShowPath != tt.expected.ShowPath {
				t.Errorf("ShowPath = %v, want %v", result.ShowPath, tt.expected.ShowPath)
			}
			if result.ShowWarnings != tt.expected.ShowWarnings {
				t.Errorf("ShowWarnings = %v, want %v", result.ShowWarnings, tt.expected.ShowWarnings)
			}
			if result.ShowComparison != tt.expected.ShowComparison {
				t.Errorf("ShowComparison = %v, want %v", result.ShowComparison, tt.expected.ShowComparison)
			}
			if result.Validate != tt.expected.Validate {
				t.Errorf("Validate = %v, want %v", result.Validate, tt.expected.Validate)
			}
		})
	}
}

func TestVersionStatus_String(t *testing.T) {
	tests := []struct {
		status   VersionStatus
		expected string
	}{
		{StatusOK, "OK"},
		{StatusNotConfigured, "Not Configured"},
		{StatusNotInstalled, "Not Installed"},
		{StatusInvalid, "Invalid"},
		{VersionStatus(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.status.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}
