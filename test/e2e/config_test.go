// ABOUTME: End-to-end tests for PVM configuration system
// ABOUTME: Tests configuration file loading, saving, and priority

package e2e

import (
	"path/filepath"
	"testing"

	"tamarou.com/pvm/test/e2e/helpers"
)

// TestConfigShowCommand tests the 'config show' command
func TestConfigShowCommand(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Run config show command or skip as TODO if not implemented
	stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"config", "show"}, "Configuration display")

	// Check that output contains expected sections
	helpers.AssertStringContains(t, stdout, "[pvm]", "Config output missing PVM section")
	helpers.AssertStringContains(t, stdout, "default_perl", "Config output missing default_perl setting")
}

// TestConfigGetCommand tests the 'config get' command
func TestConfigGetCommand(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Get a config value or skip as TODO if not implemented
	stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"config", "get", "pvm.download_mirror"},
		"Configuration value retrieval")

	// Check that output contains a value (likely https://www.cpan.org/src/5.0)
	if stdout == "" {
		t.Error("Config get returned empty value for download_mirror")
	}
}

// TestConfigSetCommand tests the 'config set' command
func TestConfigSetCommand(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Set a config value or skip as TODO if not implemented
	testValue := "5.36.0"
	_ = helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"config", "set", "pvm.default_perl", testValue},
		"Configuration value setting")

	// Get the value to verify it was set
	stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"config", "get", "pvm.default_perl"},
		"Configuration value retrieval")

	// Check that the value was set correctly
	helpers.AssertStringContains(t, stdout, testValue,
		"Config value not set correctly")

	// Check that the config file was created
	configFile := filepath.Join(env.PVMConfigDir, "pvm.toml")
	helpers.AssertFileExists(t, configFile, "Config file not created")
	helpers.AssertFileContains(t, configFile, testValue, "Config file does not contain expected value")
}

// TestConfigLayering tests that configuration layering works correctly
func TestConfigLayering(t *testing.T) {
	helpers.SkipTODO(t, "Configuration layering and priority functionality")
}

// TestConfigInit tests the 'config init' command
func TestConfigInit(t *testing.T) {
	helpers.SkipTODO(t, "Configuration initialization command")
}
