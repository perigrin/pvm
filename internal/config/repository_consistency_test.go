// ABOUTME: Tests to ensure repository configuration consistency across the codebase
// ABOUTME: Prevents accidental reverts to incorrect repository names like pvm-dev

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"tamarou.com/pvm/internal/updater"
	"tamarou.com/pvm/internal/version"
)

const expectedRepository = "perigrin/pvm"
const expectedBinaryMirror = "https://github.com/perigrin/pvm/releases/download"

func TestRepositoryConsistency_DefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	t.Run("PVM Update Repository", func(t *testing.T) {
		assert.NotNil(t, cfg.PVM, "PVM config should not be nil")
		assert.NotNil(t, cfg.PVM.Update, "PVM Update config should not be nil")
		assert.Equal(t, expectedRepository, cfg.PVM.Update.Repository,
			"Default update repository should be %s, not pvm-dev", expectedRepository)
	})

	t.Run("PVM Binary Mirrors", func(t *testing.T) {
		assert.NotNil(t, cfg.PVM, "PVM config should not be nil")
		assert.NotNil(t, cfg.PVM.Binary, "PVM Binary config should not be nil")
		assert.NotEmpty(t, cfg.PVM.Binary.BinaryMirrors, "Binary mirrors should not be empty")

		// Check that at least one mirror points to the correct repository
		found := false
		for _, mirror := range cfg.PVM.Binary.BinaryMirrors {
			if mirror == expectedBinaryMirror {
				found = true
				break
			}
		}
		assert.True(t, found, "Binary mirrors should include %s, not pvm-dev", expectedBinaryMirror)
	})
}

func TestRepositoryConsistency_UpdaterDefaults(t *testing.T) {
	t.Run("Updater Default Options", func(t *testing.T) {
		opts := updater.DefaultUpdateOptions()
		assert.NotNil(t, opts, "Default update options should not be nil")
		assert.Equal(t, expectedRepository, opts.Repository,
			"Updater default repository should be %s, not pvm-dev", expectedRepository)
	})
}

func TestRepositoryConsistency_VersionDefaults(t *testing.T) {
	t.Run("Version Check Default Options", func(t *testing.T) {
		opts := version.DefaultCheckOptions()
		assert.NotNil(t, opts, "Default check options should not be nil")
		assert.Equal(t, expectedRepository, opts.Repository,
			"Version check default repository should be %s, not pvm-dev", expectedRepository)
	})
}

func TestRepositoryConsistency_NoDevReferences(t *testing.T) {
	t.Run("Config Should Not Reference pvm-dev", func(t *testing.T) {
		cfg := NewDefaultConfig()

		// Check that no configuration references pvm-dev
		if cfg.PVM != nil && cfg.PVM.Update != nil {
			assert.NotContains(t, cfg.PVM.Update.Repository, "pvm-dev",
				"Update repository should not reference pvm-dev")
		}

		if cfg.PVM != nil && cfg.PVM.Binary != nil {
			for i, mirror := range cfg.PVM.Binary.BinaryMirrors {
				assert.NotContains(t, mirror, "pvm-dev",
					"Binary mirror %d should not reference pvm-dev: %s", i, mirror)
			}
		}
	})

	t.Run("Updater Should Not Reference pvm-dev", func(t *testing.T) {
		opts := updater.DefaultUpdateOptions()
		assert.NotContains(t, opts.Repository, "pvm-dev",
			"Updater repository should not reference pvm-dev")
	})

	t.Run("Version Check Should Not Reference pvm-dev", func(t *testing.T) {
		opts := version.DefaultCheckOptions()
		assert.NotContains(t, opts.Repository, "pvm-dev",
			"Version check repository should not reference pvm-dev")
	})
}

func TestRepositoryConsistency_AllPointToSameRepo(t *testing.T) {
	t.Run("All Components Use Same Repository", func(t *testing.T) {
		cfg := NewDefaultConfig()
		updaterOpts := updater.DefaultUpdateOptions()
		versionOpts := version.DefaultCheckOptions()

		repositories := []string{}

		if cfg.PVM != nil && cfg.PVM.Update != nil {
			repositories = append(repositories, cfg.PVM.Update.Repository)
		}

		repositories = append(repositories, updaterOpts.Repository)
		repositories = append(repositories, versionOpts.Repository)

		// All repositories should be the same
		for i, repo := range repositories {
			assert.Equal(t, expectedRepository, repo,
				"Repository %d should be %s but got %s", i, expectedRepository, repo)
		}

		// Verify they're all consistent with each other
		if len(repositories) > 1 {
			firstRepo := repositories[0]
			for i := 1; i < len(repositories); i++ {
				assert.Equal(t, firstRepo, repositories[i],
					"All repositories should be consistent. Repository 0: %s, Repository %d: %s",
					firstRepo, i, repositories[i])
			}
		}
	})
}

func TestRepositoryConsistency_BreakageDetection(t *testing.T) {
	t.Run("Detect Common Regression Cases", func(t *testing.T) {
		cfg := NewDefaultConfig()

		// These are the exact patterns that would indicate a regression
		problematicPatterns := []string{
			"pvm-dev",
			"your-username",
			"example.com",
			"localhost",
			"127.0.0.1",
		}

		repositories := []string{
			cfg.PVM.Update.Repository,
		}

		// Add binary mirrors to check
		if cfg.PVM != nil && cfg.PVM.Binary != nil {
			repositories = append(repositories, cfg.PVM.Binary.BinaryMirrors...)
		}

		for _, repo := range repositories {
			for _, pattern := range problematicPatterns {
				assert.NotContains(t, repo, pattern,
					"Repository configuration contains problematic pattern '%s': %s", pattern, repo)
			}
		}
	})
}

func TestRepositoryConsistency_ValidGitHubFormat(t *testing.T) {
	t.Run("Repository Format Validation", func(t *testing.T) {
		cfg := NewDefaultConfig()
		updaterOpts := updater.DefaultUpdateOptions()
		versionOpts := version.DefaultCheckOptions()

		repositories := []string{
			cfg.PVM.Update.Repository,
			updaterOpts.Repository,
			versionOpts.Repository,
		}

		for i, repo := range repositories {
			// Check format is owner/repo
			assert.Regexp(t, `^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`, repo,
				"Repository %d should be in owner/repo format: %s", i, repo)

			// Check it's specifically perigrin/pvm
			assert.Equal(t, expectedRepository, repo,
				"Repository %d should be exactly %s", i, expectedRepository)
		}
	})

	t.Run("Binary Mirror URL Format Validation", func(t *testing.T) {
		cfg := NewDefaultConfig()

		if cfg.PVM != nil && cfg.PVM.Binary != nil {
			for i, mirror := range cfg.PVM.Binary.BinaryMirrors {
				// Should start with https://github.com/perigrin/pvm
				assert.Regexp(t, `^https://github\.com/perigrin/pvm/`, mirror,
					"Binary mirror %d should point to correct GitHub repository: %s", i, mirror)

				// Should not contain pvm-dev
				assert.NotContains(t, mirror, "pvm-dev",
					"Binary mirror %d should not reference pvm-dev: %s", i, mirror)
			}
		}
	})
}
