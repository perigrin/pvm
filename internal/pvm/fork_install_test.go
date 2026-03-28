// ABOUTME: Tests for the fork install flow in the install command
// ABOUTME: Covers routing detection, error cases, and orchestration logic

package pvm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/perl"
)

// TestInstallCommand_ForkRouting verifies that the install command
// detects fork identifiers (contains "/") and routes accordingly.
func TestInstallCommand_ForkRouting(t *testing.T) {
	tests := []struct {
		name    string
		version string
		isFork  bool
	}{
		{
			name:    "plain version is not a fork",
			version: "5.40.2",
			isFork:  false,
		},
		{
			name:    "remote slash version is a fork",
			version: "mycompany/5.40.2",
			isFork:  true,
		},
		{
			name:    "remote slash fork at version is a fork",
			version: "mycompany/myfork@5.40.2",
			isFork:  true,
		},
		{
			name:    "latest is not a fork",
			version: "latest",
			isFork:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isForkInstall(tt.version)
			assert.Equal(t, tt.isFork, got)
		})
	}
}

// TestInstallCommand_ForkIdentifierParsing verifies that valid fork identifiers
// parse correctly and invalid ones return errors.
func TestInstallCommand_ForkIdentifierParsing(t *testing.T) {
	t.Run("valid remote/fork@version", func(t *testing.T) {
		fi, err := perl.ParseForkIdentifier("mycompany/myfork@5.40.2")
		require.NoError(t, err)
		assert.Equal(t, "mycompany", fi.Remote)
		assert.Equal(t, "myfork", fi.ForkName)
		assert.Equal(t, "5.40.2", fi.BaseVersion)
		assert.True(t, fi.IsFork())
	})

	t.Run("valid remote/version", func(t *testing.T) {
		fi, err := perl.ParseForkIdentifier("mycompany/5.40.2")
		require.NoError(t, err)
		assert.Equal(t, "mycompany", fi.Remote)
		assert.Equal(t, "", fi.ForkName)
		assert.Equal(t, "5.40.2", fi.BaseVersion)
		assert.True(t, fi.IsFork())
	})

	t.Run("invalid identifier returns error", func(t *testing.T) {
		_, err := perl.ParseForkIdentifier("origin/myfork@5.40.2")
		assert.Error(t, err, "origin is a reserved remote name")
	})
}

// TestFindRemoteInConfig tests looking up a remote by name in PVM config.
func TestFindRemoteInConfig(t *testing.T) {
	cfg := &config.Config{
		PVM: &config.PVMConfig{
			Remotes: []config.PVMRemoteConfig{
				{Name: "mycompany", URL: "https://example.com/perl.git", Type: "git"},
				{Name: "other", URL: "https://other.com/perl.git", Type: "git"},
			},
		},
	}

	t.Run("finds existing remote", func(t *testing.T) {
		r, found := findRemoteInConfig(cfg, "mycompany")
		assert.True(t, found)
		assert.Equal(t, "mycompany", r.Name)
		assert.Equal(t, "https://example.com/perl.git", r.URL)
	})

	t.Run("returns false for unknown remote", func(t *testing.T) {
		_, found := findRemoteInConfig(cfg, "unknown")
		assert.False(t, found)
	})

	t.Run("returns false when PVM config is nil", func(t *testing.T) {
		nilCfg := &config.Config{}
		_, found := findRemoteInConfig(nilCfg, "mycompany")
		assert.False(t, found)
	})
}

// TestMatchTagForVersion tests finding a matching git tag for a fork version.
func TestMatchTagForVersion(t *testing.T) {
	tags := []string{
		"myfork-5.40.2",
		"myfork-5.38.0",
		"v5.40.2",
		"v5.38.0",
		"unrelated-tag",
	}

	t.Run("finds fork tag for fork install", func(t *testing.T) {
		fi := &perl.ForkIdentifier{Remote: "mycompany", ForkName: "myfork", BaseVersion: "5.40.2"}
		tag, found := matchTagForVersion(tags, fi)
		assert.True(t, found)
		assert.Equal(t, "myfork-5.40.2", tag)
	})

	t.Run("finds v-prefixed tag for remote-only install", func(t *testing.T) {
		fi := &perl.ForkIdentifier{Remote: "mycompany", ForkName: "", BaseVersion: "5.40.2"}
		tag, found := matchTagForVersion(tags, fi)
		assert.True(t, found)
		assert.Equal(t, "v5.40.2", tag)
	})

	t.Run("returns false when no matching tag exists", func(t *testing.T) {
		fi := &perl.ForkIdentifier{Remote: "mycompany", ForkName: "myfork", BaseVersion: "5.42.0"}
		_, found := matchTagForVersion(tags, fi)
		assert.False(t, found)
	})

	t.Run("returns false for empty tag list", func(t *testing.T) {
		fi := &perl.ForkIdentifier{Remote: "mycompany", ForkName: "myfork", BaseVersion: "5.40.2"}
		_, found := matchTagForVersion([]string{}, fi)
		assert.False(t, found)
	})
}

// TestForkInstallPath tests that fork install paths are built correctly.
func TestForkInstallPath(t *testing.T) {
	t.Run("remote/fork-version path", func(t *testing.T) {
		fi := &perl.ForkIdentifier{Remote: "mycompany", ForkName: "myfork", BaseVersion: "5.40.2"}
		path := forkInstallSubpath(fi)
		assert.Equal(t, "mycompany/myfork-5.40.2", path)
	})

	t.Run("remote/version path (no fork name)", func(t *testing.T) {
		fi := &perl.ForkIdentifier{Remote: "mycompany", ForkName: "", BaseVersion: "5.40.2"}
		path := forkInstallSubpath(fi)
		assert.Equal(t, "mycompany/5.40.2", path)
	})
}

// TestBuildVersionForFork tests building the version string stored in the registry
// for a fork installation.
func TestBuildVersionForFork(t *testing.T) {
	t.Run("fork with name uses forkname-version", func(t *testing.T) {
		fi := &perl.ForkIdentifier{Remote: "mycompany", ForkName: "myfork", BaseVersion: "5.40.2"}
		v := buildVersionStringForFork(fi)
		assert.Equal(t, "myfork-5.40.2", v)
	})

	t.Run("fork without name uses base version", func(t *testing.T) {
		fi := &perl.ForkIdentifier{Remote: "mycompany", ForkName: "", BaseVersion: "5.40.2"}
		v := buildVersionStringForFork(fi)
		assert.Equal(t, "5.40.2", v)
	})
}
