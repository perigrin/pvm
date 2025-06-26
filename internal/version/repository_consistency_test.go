// ABOUTME: Tests to ensure version package repository configuration consistency
// ABOUTME: Prevents regressions in version check default configurations

package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const expectedRepository = "perigrin/pvm"

func TestVersionDefaults_RepositoryConsistency(t *testing.T) {
	t.Run("Version Check Default Options Repository", func(t *testing.T) {
		opts := DefaultCheckOptions()
		assert.Equal(t, expectedRepository, opts.Repository,
			"Version check default repository should be %s, not pvm-dev", expectedRepository)
	})
}

func TestVersionDefaults_NoDevReferences(t *testing.T) {
	t.Run("Version Check Should Not Reference pvm-dev", func(t *testing.T) {
		opts := DefaultCheckOptions()
		assert.NotContains(t, opts.Repository, "pvm-dev",
			"Version check repository should not reference pvm-dev")
		assert.NotContains(t, opts.Repository, "your-username",
			"Version check repository should not reference placeholder text")
		assert.NotContains(t, opts.Repository, "example.com",
			"Version check repository should not reference example domains")
	})
}

func TestVersionDefaults_ValidFormat(t *testing.T) {
	t.Run("Repository Format Should Be Valid GitHub Repo", func(t *testing.T) {
		opts := DefaultCheckOptions()

		// Should be in owner/repo format
		assert.Regexp(t, `^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`, opts.Repository,
			"Repository should be in owner/repo format: %s", opts.Repository)

		// Should be specifically the expected repository
		assert.Equal(t, expectedRepository, opts.Repository,
			"Repository should be exactly %s", expectedRepository)
	})
}
