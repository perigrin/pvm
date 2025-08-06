// ABOUTME: Test for version source display improvements (issue #196)
// ABOUTME: Verifies that environment variable names are displayed specifically

package current

import (
	"testing"

	"tamarou.com/pvm/internal/perl"
)

func TestVersionSourceDisplayImprovements(t *testing.T) {
	tests := []struct {
		name     string
		source   perl.ResolutionSource
		path     string
		expected string
	}{
		{
			name:     "PVM_PERL_VERSION environment variable",
			source:   perl.EnvironmentVariable,
			path:     "PVM_PERL_VERSION",
			expected: "set by PVM_PERL_VERSION",
		},
		{
			name:     "PLENV_VERSION environment variable",
			source:   perl.EnvironmentVariable,
			path:     "PLENV_VERSION",
			expected: "set by PLENV_VERSION",
		},
		{
			name:     "PERLBREW_PERL environment variable",
			source:   perl.EnvironmentVariable,
			path:     "PERLBREW_PERL",
			expected: "set by PERLBREW_PERL",
		},
		{
			name:     "Environment variable without specific name",
			source:   perl.EnvironmentVariable,
			path:     "",
			expected: "set by environment variable",
		},
		{
			name:     "Project version file with path",
			source:   perl.ProjectVersionFile,
			path:     "/project/.perl-version",
			expected: "set by /project/.perl-version",
		},
		{
			name:     "User config",
			source:   perl.UserConfig,
			path:     "/home/user/.pvm/config.toml",
			expected: "set by user configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSourceDescription(tt.source, tt.path)
			if result != tt.expected {
				t.Errorf("formatSourceDescription(%v, %q) = %q, want %q",
					tt.source, tt.path, result, tt.expected)
			}
		})
	}
}

func TestCurrentVersionInfoWithEnvironmentVariables(t *testing.T) {
	// Test that CurrentVersionInfo properly displays environment variable names
	tests := []struct {
		name         string
		info         *CurrentVersionInfo
		expectedDesc string
	}{
		{
			name: "PVM_PERL_VERSION in source description",
			info: &CurrentVersionInfo{
				Version:           "5.38.0",
				Source:            string(perl.EnvironmentVariable),
				SourceDescription: formatSourceDescription(perl.EnvironmentVariable, "PVM_PERL_VERSION"),
				Path:              "/usr/local/pvm/perls/5.38.0/bin/perl",
				IsAvailable:       true,
			},
			expectedDesc: "set by PVM_PERL_VERSION",
		},
		{
			name: "PLENV_VERSION in source description",
			info: &CurrentVersionInfo{
				Version:           "5.36.0",
				Source:            string(perl.EnvironmentVariable),
				SourceDescription: formatSourceDescription(perl.EnvironmentVariable, "PLENV_VERSION"),
				Path:              "/usr/local/plenv/perls/5.36.0/bin/perl",
				IsAvailable:       true,
			},
			expectedDesc: "set by PLENV_VERSION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.info.SourceDescription != tt.expectedDesc {
				t.Errorf("Expected source description %q, got %q",
					tt.expectedDesc, tt.info.SourceDescription)
			}

			// Test the formatted output
			output := tt.info.FormatOutput(false)
			expectedOutput := tt.info.Version + " (" + tt.expectedDesc + ")"
			if output != expectedOutput {
				t.Errorf("Expected formatted output %q, got %q",
					expectedOutput, output)
			}
		})
	}
}
