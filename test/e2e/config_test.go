// ABOUTME: End-to-end tests for PVM configuration system
// ABOUTME: Tests configuration file loading, saving, and priority

package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/test/e2e/helpers"
)

// TestConfigShowCommand tests the 'config show' command
func TestConfigShowCommand(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Run config show command
	stdout := helpers.AssertPVMSucceeds(t, env, []string{"config", "show"}, "Config show command failed")

	// Check that output contains expected sections
	helpers.AssertStringContains(t, stdout, "[pvm]", "Config output missing PVM section")
	helpers.AssertStringContains(t, stdout, "default_perl", "Config output missing default_perl setting")
}

// TestConfigGetCommand tests the 'config get' command
func TestConfigGetCommand(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Get a config value
	stdout := helpers.AssertPVMSucceeds(t, env, []string{"config", "get", "pvm.download_mirror"},
		"Config get command failed")

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
	helpers.AssertPVMSucceeds(t, env, []string{"config", "set", "pvm.default_perl", testValue},
		"Config set command failed")

	// Get the value to verify it was set
	stdout := helpers.AssertPVMSucceeds(t, env, []string{"config", "get", "pvm.default_perl"},
		"Config get command failed")

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
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create system config (lowest priority)
	systemConfigDir := filepath.Join(env.RootDir, "etc", "pvm")
	err := os.MkdirAll(systemConfigDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create system config directory: %v", err)
	}

	systemConfigFile := filepath.Join(systemConfigDir, "pvm.toml")
	systemConfig := `[pvm]
default_perl = "5.30.0"
download_mirror = "https://system-mirror.example.com"
`
	err = os.WriteFile(systemConfigFile, []byte(systemConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create system config file: %v", err)
	}

	// Create user config (middle priority)
	userConfigFile := filepath.Join(env.PVMConfigDir, "pvm.toml")
	userConfig := `[pvm]
default_perl = "5.32.0"
`
	err = os.WriteFile(userConfigFile, []byte(userConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create user config file: %v", err)
	}

	// Create project config (highest priority)
	projectDir := filepath.Join(env.HomeDir, "project")
	err = os.MkdirAll(filepath.Join(projectDir, ".pvm"), 0755)
	if err != nil {
		t.Fatalf("Failed to create project config directory: %v", err)
	}

	projectConfigFile := filepath.Join(projectDir, ".pvm", "pvm.toml")
	projectConfig := `[pvm]
default_perl = "5.34.0"
`
	err = os.WriteFile(projectConfigFile, []byte(projectConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create project config file: %v", err)
	}

	// Create a script to run PVM with custom config paths
	testScript := filepath.Join(env.HomeDir, "test_config.sh")
	scriptContent := `#!/bin/bash
export PVM_SYSTEM_CONFIG_PATH="` + systemConfigDir + `"
export PVM_CONFIG_PATH="` + env.PVMConfigDir + `"
export PVM_PROJECT_CONFIG_PATH="` + filepath.Join(projectDir, ".pvm") + `"

# In global context, should use user config
cd "` + env.HomeDir + `"
"` + env.PVMBinary + `" config get pvm.default_perl
echo "Download mirror: $("` + env.PVMBinary + `" config get pvm.download_mirror)"

# In project context, should use project config
cd "` + projectDir + `"
"` + env.PVMBinary + `" config get pvm.default_perl
echo "Download mirror: $("` + env.PVMBinary + `" config get pvm.download_mirror)"
`
	err = os.WriteFile(testScript, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Run the test script
	stdout, stderr, err := env.RunCommand("bash", testScript)
	if err != nil {
		t.Fatalf("Failed to run test script: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Check the results
	// In home dir, should use user config for default_perl
	helpers.AssertStringContains(t, stdout, "5.32.0", "User config not applied correctly")

	// In project dir, should use project config for default_perl
	helpers.AssertStringContains(t, stdout, "5.34.0", "Project config not applied correctly")

	// For download_mirror, should fall back to system config in both cases
	helpers.AssertStringContains(t, stdout, "https://system-mirror.example.com",
		"System config fallback not working")
}

// TestConfigInit tests the 'config init' command
func TestConfigInit(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Run config init command
	stdout := helpers.AssertPVMSucceeds(t, env, []string{"config", "init", "--user"},
		"Config init command failed")
	helpers.AssertStringContains(t, stdout, "Created user configuration",
		"Config init output does not indicate success")

	// Check that config file was created
	configFile := filepath.Join(env.PVMConfigDir, "pvm.toml")
	helpers.AssertFileExists(t, configFile, "Config file not created")

	// Check that it contains default values
	helpers.AssertFileContains(t, configFile, "default_perl", "Config file missing default_perl setting")

	// Test project config init
	projectDir := filepath.Join(env.HomeDir, "project")
	err := os.MkdirAll(projectDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Change to project directory and run init
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	err = os.Chdir(projectDir)
	if err != nil {
		t.Fatalf("Failed to change to project directory: %v", err)
	}

	stdout = helpers.AssertPVMSucceeds(t, env, []string{"config", "init", "--project"},
		"Config init project command failed")
	helpers.AssertStringContains(t, stdout, "Created project configuration",
		"Config init project output does not indicate success")

	// Check that project config file was created
	projectConfigFile := filepath.Join(projectDir, ".pvm", "pvm.toml")
	helpers.AssertFileExists(t, projectConfigFile, "Project config file not created")
}
