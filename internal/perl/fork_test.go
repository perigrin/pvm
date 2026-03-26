// ABOUTME: Tests for fork identifier parsing and fork-related types
// ABOUTME: Verifies parsing of remote/fork@version identifiers for custom Perl forks

package perl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseForkIdentifier(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantRemote  string
		wantFork    string
		wantVersion string
		wantErr     bool
	}{
		// Full form: remote/fork@version
		{
			name:        "full form with remote fork and version",
			input:       "mycompany/myfork@5.40.2",
			wantRemote:  "mycompany",
			wantFork:    "myfork",
			wantVersion: "5.40.2",
		},
		{
			name:        "full form with hyphenated names",
			input:       "my-company/my-fork@5.38.0",
			wantRemote:  "my-company",
			wantFork:    "my-fork",
			wantVersion: "5.38.0",
		},

		// Short form: remote/version (no fork name)
		{
			name:        "remote and version without fork name",
			input:       "mycompany/5.40.2",
			wantRemote:  "mycompany",
			wantFork:    "",
			wantVersion: "5.40.2",
		},

		// Bare version: just a version string (defaults to origin)
		{
			name:        "bare version defaults to origin",
			input:       "5.40.2",
			wantRemote:  "origin",
			wantFork:    "",
			wantVersion: "5.40.2",
		},
		{
			name:        "bare dev version",
			input:       "5.39.0",
			wantRemote:  "origin",
			wantFork:    "",
			wantVersion: "5.39.0",
		},
		{
			name:        "two-part version",
			input:       "5.40",
			wantRemote:  "origin",
			wantFork:    "",
			wantVersion: "5.40",
		},

		// Validation errors
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "origin as remote prefix is reserved",
			input:   "origin/myfork@5.40.2",
			wantErr: true,
		},
		{
			name:    "fork name perl is reserved",
			input:   "mycompany/perl@5.40.2",
			wantErr: true,
		},
		{
			name:    "invalid remote name with uppercase",
			input:   "MyCompany/myfork@5.40.2",
			wantErr: true,
		},
		{
			name:    "invalid remote name starting with hyphen",
			input:   "-company/myfork@5.40.2",
			wantErr: true,
		},
		{
			name:    "invalid fork name starting with digit",
			input:   "mycompany/1fork@5.40.2",
			wantErr: true,
		},
		{
			name:    "at sign without version",
			input:   "mycompany/myfork@",
			wantErr: true,
		},
		{
			name:    "invalid version in full form",
			input:   "mycompany/myfork@notaversion",
			wantErr: true,
		},
		{
			name:    "invalid version in short form",
			input:   "mycompany/notaversion",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseForkIdentifier(tt.input)

			if tt.wantErr {
				assert.Error(t, err, "expected error for input %q", tt.input)
				return
			}

			require.NoError(t, err, "unexpected error for input %q", tt.input)
			assert.Equal(t, tt.wantRemote, got.Remote)
			assert.Equal(t, tt.wantFork, got.ForkName)
			assert.Equal(t, tt.wantVersion, got.BaseVersion)
		})
	}
}

func TestForkIdentifierDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		id       ForkIdentifier
		expected string
	}{
		{
			name:     "full identifier with fork name",
			id:       ForkIdentifier{Remote: "mycompany", ForkName: "myfork", BaseVersion: "5.40.2"},
			expected: "mycompany/myfork-5.40.2",
		},
		{
			name:     "identifier without fork name",
			id:       ForkIdentifier{Remote: "mycompany", ForkName: "", BaseVersion: "5.40.2"},
			expected: "mycompany/5.40.2",
		},
		{
			name:     "origin remote with no fork (bare version)",
			id:       ForkIdentifier{Remote: "origin", ForkName: "", BaseVersion: "5.40.2"},
			expected: "5.40.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.id.DisplayName())
		})
	}
}

func TestForkIdentifierInstallPath(t *testing.T) {
	tests := []struct {
		name     string
		id       ForkIdentifier
		expected string
	}{
		{
			name:     "full identifier produces namespaced path",
			id:       ForkIdentifier{Remote: "mycompany", ForkName: "myfork", BaseVersion: "5.40.2"},
			expected: "mycompany/myfork-5.40.2",
		},
		{
			name:     "no fork name uses version directly",
			id:       ForkIdentifier{Remote: "mycompany", ForkName: "", BaseVersion: "5.40.2"},
			expected: "mycompany/5.40.2",
		},
		{
			name:     "origin remote produces flat path",
			id:       ForkIdentifier{Remote: "origin", ForkName: "", BaseVersion: "5.40.2"},
			expected: "5.40.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.id.InstallPath())
		})
	}
}

func TestForkIdentifierIsFork(t *testing.T) {
	tests := []struct {
		name     string
		id       ForkIdentifier
		expected bool
	}{
		{
			name:     "origin remote is not a fork",
			id:       ForkIdentifier{Remote: "origin", ForkName: "", BaseVersion: "5.40.2"},
			expected: false,
		},
		{
			name:     "non-origin remote is a fork",
			id:       ForkIdentifier{Remote: "mycompany", ForkName: "", BaseVersion: "5.40.2"},
			expected: true,
		},
		{
			name:     "non-origin with fork name is a fork",
			id:       ForkIdentifier{Remote: "mycompany", ForkName: "myfork", BaseVersion: "5.40.2"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.id.IsFork())
		})
	}
}
