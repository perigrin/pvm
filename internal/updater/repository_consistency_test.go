// ABOUTME: Tests to ensure updater repository configuration consistency
// ABOUTME: Prevents regressions in auto-update and updater default configurations

package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const expectedRepository = "perigrin/pvm"

func TestAutoUpdateDefaults_RepositoryConsistency(t *testing.T) {
	t.Run("Auto Update Default Config Repository", func(t *testing.T) {
		defaults := DefaultAutoUpdateConfig()
		assert.Equal(t, expectedRepository, defaults.Repository,
			"Auto update default repository should be %s, not pvm-dev", expectedRepository)
	})
}

func TestAutoUpdateDefaults_NoDevReferences(t *testing.T) {
	t.Run("Auto Update Should Not Reference pvm-dev", func(t *testing.T) {
		defaults := DefaultAutoUpdateConfig()
		assert.NotContains(t, defaults.Repository, "pvm-dev",
			"Auto update repository should not reference pvm-dev")
		assert.NotContains(t, defaults.Repository, "your-username",
			"Auto update repository should not reference placeholder text")
	})
}

func TestUpdaterDefaults_RepositoryConsistency(t *testing.T) {
	t.Run("Updater Default Options Repository", func(t *testing.T) {
		opts := DefaultUpdateOptions()
		assert.Equal(t, expectedRepository, opts.Repository,
			"Updater default repository should be %s, not pvm-dev", expectedRepository)
	})
}

func TestUpdaterDefaults_NoDevReferences(t *testing.T) {
	t.Run("Updater Should Not Reference pvm-dev", func(t *testing.T) {
		opts := DefaultUpdateOptions()
		assert.NotContains(t, opts.Repository, "pvm-dev",
			"Updater repository should not reference pvm-dev")
		assert.NotContains(t, opts.Repository, "your-username",
			"Updater repository should not reference placeholder text")
	})
}

func TestUpdaterConsistency_CrossPackage(t *testing.T) {
	t.Run("Auto Update and Updater Use Same Repository", func(t *testing.T) {
		autoUpdateDefaults := DefaultAutoUpdateConfig()
		updaterDefaults := DefaultUpdateOptions()

		assert.Equal(t, autoUpdateDefaults.Repository, updaterDefaults.Repository,
			"Auto update and updater should use the same repository. AutoUpdate: %s, Updater: %s",
			autoUpdateDefaults.Repository, updaterDefaults.Repository)
	})
}
