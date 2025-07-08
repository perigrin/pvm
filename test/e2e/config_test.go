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

	// Run config show command
	stdout, stderr, err := env.RunPVM("config", "show")
	if err != nil {
		t.Fatalf("Configuration display failed\nError: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Check that output contains expected sections
	helpers.AssertStringContains(t, stdout, "[pvm]", "Config output missing PVM section")
	helpers.AssertStringContains(t, stdout, "default_perl", "Config output missing default_perl setting")
}

// TestConfigGetCommand tests the 'config get' command
func TestConfigGetCommand(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Get a config value
	stdout, stderr, err := env.RunPVM("config", "get", "pvm.download_mirror")
	if err != nil {
		t.Fatalf("Configuration value retrieval failed\nError: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Check that output contains a value (likely https://www.cpan.org/src/5.0)
	if stdout == "" {
		t.Error("Config get returned empty value for download_mirror")
	}
}

// TestConfigSetCommand tests the 'config set' command
func TestConfigSetCommand(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Set a config value
	testValue := "5.36.0"
	_, stderr, err := env.RunPVM("config", "set", "pvm.default_perl", testValue)
	if err != nil {
		t.Fatalf("Configuration value setting failed\nError: %v\nStderr: %s", err, stderr)
	}

	// Get the value to verify it was set
	stdout, stderr, err := env.RunPVM("config", "get", "pvm.default_perl")
	if err != nil {
		t.Fatalf("Configuration value retrieval failed\nError: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

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
