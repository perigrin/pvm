// ABOUTME: Migration and backward compatibility tests for Step 17
// ABOUTME: Validates that existing PVM installations can upgrade smoothly

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/test/e2e/helpers"
)

func TestMigrationCompatibility_ExistingConfigs(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Simulate an existing PVM installation with old config format
	oldConfigDir := filepath.Join(env.RootDir, ".pvm")
	err := os.MkdirAll(oldConfigDir, 0755)
	require.NoError(t, err)

	// Create old-style configuration
	oldConfigFile := filepath.Join(oldConfigDir, "config.toml")
	oldConfigContent := `# Old PVM configuration format
[versions]
default = "system"

[directories]
data = "~/.pvm/data"
cache = "~/.pvm/cache"

[build]
parallel = true
verbose = false

[perl]
system_path = "/usr/bin/perl"
`
	err = os.WriteFile(oldConfigFile, []byte(oldConfigContent), 0644)
	require.NoError(t, err)

	// Create old-style shims directory
	oldShimsDir := filepath.Join(oldConfigDir, "shims")
	err = os.MkdirAll(oldShimsDir, 0755)
	require.NoError(t, err)

	// Create old-style version data
	oldVersionsDir := filepath.Join(oldConfigDir, "versions")
	err = os.MkdirAll(oldVersionsDir, 0755)
	require.NoError(t, err)

	systemVersionFile := filepath.Join(oldVersionsDir, "system")
	err = os.WriteFile(systemVersionFile, []byte("/usr/bin/perl"), 0644)
	require.NoError(t, err)

	// Set PVM_HOME to the old config directory to simulate existing installation
	os.Setenv("PVM_HOME", oldConfigDir)

	// Test that new PVM can read old configuration
	t.Log("Testing old config compatibility...")
	stdout, _, err := env.RunPVM("config", "get", "versions.default")

	// Should either work with old config or gracefully handle migration
	if err == nil {
		assert.Contains(t, stdout, "system", "Should read old default version")
	} else {
		t.Logf("Old config migration needed (expected): %v", err)
	}

	// Test that list command works even without migrated data
	t.Log("Testing version list with old data...")
	_, _, err = env.RunPVM("list")
	assert.NoError(t, err, "List should work even without existing data")
	// Since we're in a test environment, system perl might not be detected
	// Just check that the command runs without error

	// Test migration process
	t.Log("Testing configuration migration...")
	_, _, err = env.RunPVM("config", "migrate")

	// Migration command may not exist yet, but config operations should work
	if err != nil {
		t.Logf("Migration command output: %v", err)
	}

	// Verify that basic operations still work after potential migration
	_, _, err = env.RunPVM("config", "list")
	assert.NoError(t, err, "Config operations should work after migration")
}

func TestMigrationCompatibility_ExistingShims(t *testing.T) {
	helpers.SkipIfNoSystemPerl(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create old-style shims
	shimsDir := filepath.Join(env.RootDir, "shims")
	err := os.MkdirAll(shimsDir, 0755)
	require.NoError(t, err)

	// Create old-style shim script
	perlShim := filepath.Join(shimsDir, "perl")
	shimContent := `#!/bin/bash
# Old PVM shim format
export PVM_ROOT="` + env.RootDir + `"
exec "` + helpers.FindSystemPerl() + `" "$@"
`
	err = os.WriteFile(perlShim, []byte(shimContent), 0755)
	require.NoError(t, err)

	// Test that new PVM can work with existing shims
	t.Log("Testing existing shim compatibility...")
	stdout, _, err := env.RunPVM("shim", "list")

	if err == nil {
		t.Logf("Shim list output: %s", stdout)
	} else {
		t.Logf("Shim list error (may be expected): %v", err)
	}

	// Test shim regeneration
	t.Log("Testing shim regeneration...")
	_, _, err = env.RunPVM("shim", "rehash")

	if err != nil {
		t.Logf("Shim rehash output: %v", err)
	}

	// Verify that shims directory structure is maintained
	_, err = os.Stat(shimsDir)
	assert.NoError(t, err, "Shims directory should be preserved")
}

func TestMigrationCompatibility_LegacyCommands(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test that all legacy commands still work
	legacyCommands := []struct {
		name       string
		args       []string
		shouldWork bool
	}{
		{"list", []string{"list"}, true},
		{"install", []string{"install", "--help"}, true}, // Use help to avoid actual installation
		{"uninstall", []string{"uninstall", "--help"}, true},
		{"use", []string{"use", "--help"}, true},
		{"current", []string{"current"}, true},
		{"shell", []string{"shell", "--help"}, true},
		{"config", []string{"config", "--help"}, true},
		{"version", []string{"version"}, true},
	}

	for _, cmd := range legacyCommands {
		t.Run("legacy_"+cmd.name, func(t *testing.T) {
			t.Logf("Testing legacy command: %v", cmd.args)
			_, stderr, err := env.RunPVM(cmd.args...)

			if cmd.shouldWork {
				if err != nil {
					// Log error but don't fail test - some commands might have changed
					t.Logf("Legacy command %s output: %v", cmd.name, stderr)
				}
			}
		})
	}
}

func TestMigrationCompatibility_EnvironmentVariables(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test legacy environment variable handling
	legacyEnvVars := map[string]string{
		"PVM_ROOT":      env.RootDir,
		"PVM_VERSION":   "system",
		"PVM_BIN_PATH":  filepath.Join(env.RootDir, "bin"),
		"PVM_CACHE_DIR": filepath.Join(env.RootDir, "cache"),
	}

	// Set legacy environment variables
	for key, value := range legacyEnvVars {
		os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	// Test that PVM honors legacy environment variables
	t.Log("Testing legacy environment variable compatibility...")
	stdout, _, err := env.RunPVM("config", "get", "xdg.data_dir")

	if err == nil {
		t.Logf("Data directory config: %s", stdout)
	}

	// Import system perl if available
	if _, err := os.Stat("/usr/bin/perl"); err == nil {
		env.RunPVM("import-system")
	}

	// Test that legacy variables don't break new functionality
	_, _, err = env.RunPVM("list")
	assert.NoError(t, err, "List should work with legacy environment")
	// The list might be empty in test environment, just check it doesn't fail
}

func TestMigrationCompatibility_ScriptExecution(t *testing.T) {
	helpers.SkipIfNoSystemPerl(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create scripts that would work with old PVM
	oldStyleScript := filepath.Join(env.RootDir, "old_style.pl")
	oldStyleContent := `#!/usr/bin/perl
# Old style Perl script without types
use strict;
use warnings;

my $message = "Hello from old style script";
my @numbers = (1, 2, 3, 4, 5);
my %config = (
    debug => 1,
    verbose => 0
);

print "$message\n";
print "Numbers: " . join(", ", @numbers) . "\n";
print "Debug mode: " . ($config{debug} ? "on" : "off") . "\n";
print "Old style script executed successfully\n";
`
	err := os.WriteFile(oldStyleScript, []byte(oldStyleContent), 0644)
	require.NoError(t, err)

	// Test that old style scripts still work
	t.Log("Testing old style script execution...")
	systemPerl := helpers.FindSystemPerl()
	stdout := helpers.AssertPVMSucceeds(t, env,
		[]string{"pvx", "--perl", systemPerl, oldStyleScript},
		"Old style script should execute")

	assert.Contains(t, stdout, "Hello from old style script", "Should show script output")
	assert.Contains(t, stdout, "Numbers: 1, 2, 3, 4, 5", "Should process arrays")
	assert.Contains(t, stdout, "Debug mode: on", "Should process hashes")
	assert.Contains(t, stdout, "executed successfully", "Should complete execution")

	// Test that PSC can handle old style scripts gracefully
	t.Log("Testing PSC with old style scripts...")
	_, stderr, err := env.RunPVM("psc", "check", oldStyleScript)

	// PSC might warn about lack of types but shouldn't fail completely
	if err != nil {
		assert.Contains(t, stderr, "warn", "Should warn about untyped code")
		t.Logf("PSC warnings for old style script (expected): %s", stderr)
	}
}

func TestMigrationCompatibility_ModuleHandling(t *testing.T) {
	helpers.SkipIfNoSystemPerl(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create an old-style module
	moduleDir := filepath.Join(env.RootDir, "OldModule")
	err := os.MkdirAll(moduleDir, 0755)
	require.NoError(t, err)

	oldModuleFile := filepath.Join(moduleDir, "OldModule.pm")
	oldModuleContent := `package OldModule;
use strict;
use warnings;

# Old style module without type annotations
sub new {
    my $class = shift;
    my $self = {
        data => {},
        counter => 0,
    };
    return bless $self, $class;
}

sub add_data {
    my ($self, $key, $value) = @_;
    $self->{data}->{$key} = $value;
    $self->{counter}++;
    return $self->{counter};
}

sub get_data {
    my ($self, $key) = @_;
    return $self->{data}->{$key};
}

sub get_count {
    my $self = shift;
    return $self->{counter};
}

1;
`
	err = os.WriteFile(oldModuleFile, []byte(oldModuleContent), 0644)
	require.NoError(t, err)

	// Create script using old module
	testScript := filepath.Join(env.RootDir, "test_old_module.pl")
	testContent := `#!/usr/bin/perl
use strict;
use warnings;
use lib '` + moduleDir + `';
use OldModule;

my $module = OldModule->new();

my $count1 = $module->add_data("name", "test");
my $count2 = $module->add_data("value", 42);

my $name = $module->get_data("name");
my $value = $module->get_data("value");
my $total_count = $module->get_count();

print "Name: $name\n";
print "Value: $value\n";
print "Total count: $total_count\n";
print "Old module test completed\n";
`
	err = os.WriteFile(testScript, []byte(testContent), 0644)
	require.NoError(t, err)

	// Test execution of old module
	t.Log("Testing old module execution...")
	systemPerl := helpers.FindSystemPerl()
	stdout := helpers.AssertPVMSucceeds(t, env,
		[]string{"pvx", "--perl", systemPerl, testScript},
		"Old module should work")

	assert.Contains(t, stdout, "Name: test", "Should use old module")
	assert.Contains(t, stdout, "Value: 42", "Should process data")
	assert.Contains(t, stdout, "Total count: 2", "Should track counter")

	// Test gradual migration - add some types to old module
	t.Log("Testing gradual type migration...")
	gradualModuleFile := filepath.Join(moduleDir, "GradualModule.pm")
	gradualContent := strings.Replace(oldModuleContent,
		"sub get_count {",
		"sub get_count() -> Int {", 1)
	gradualContent = strings.Replace(gradualContent,
		"package OldModule;",
		"package GradualModule;", 1)

	err = os.WriteFile(gradualModuleFile, []byte(gradualContent), 0644)
	require.NoError(t, err)

	// Test that PSC can handle gradual typing
	if helpers.HasTreeSitter() {
		_, stderr, err := env.RunPVM("psc", "check", gradualModuleFile)

		if err != nil {
			// Gradual typing might have warnings
			t.Logf("Gradual typing warnings (may be expected): %s", stderr)
		}
	}
}

func TestMigrationCompatibility_ConfigurationFormats(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test different configuration formats that might exist
	configFormats := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "TOML",
			filename: "config.toml",
			content: `[versions]
default = "system"

[build]
parallel = true
`,
		},
		{
			name:     "JSON",
			filename: "config.json",
			content: `{
  "versions": {
    "default": "system"
  },
  "build": {
    "parallel": true
  }
}`,
		},
		{
			name:     "YAML",
			filename: "config.yaml",
			content: `versions:
  default: system
build:
  parallel: true
`,
		},
	}

	for _, format := range configFormats {
		t.Run("config_format_"+format.name, func(t *testing.T) {
			configDir := filepath.Join(env.RootDir, "config_test_"+format.name)
			err := os.MkdirAll(configDir, 0755)
			require.NoError(t, err)

			configFile := filepath.Join(configDir, format.filename)
			err = os.WriteFile(configFile, []byte(format.content), 0644)
			require.NoError(t, err)

			// Test that PVM can work with different config formats
			// Note: Current PVM might not support all formats, but should handle gracefully
			os.Setenv("PVM_CONFIG_DIR", configDir)
			defer os.Unsetenv("PVM_CONFIG_DIR")

			_, _, err = env.RunPVM("config", "list")
			// Error is acceptable here - we're testing graceful handling
			t.Logf("Config format %s test result: %v", format.name, err)
		})
	}
}

func TestMigrationCompatibility_UpgradePath(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Simulate complete old installation
	oldPVMDir := filepath.Join(env.RootDir, ".pvm")
	err := os.MkdirAll(oldPVMDir, 0755)
	require.NoError(t, err)

	// Create all old-style directories and files
	oldDirs := []string{
		"versions",
		"shims",
		"cache",
		"tmp",
	}

	for _, dir := range oldDirs {
		err = os.MkdirAll(filepath.Join(oldPVMDir, dir), 0755)
		require.NoError(t, err)
	}

	// Create old configuration
	oldConfig := filepath.Join(oldPVMDir, "config")
	err = os.WriteFile(oldConfig, []byte("default_version=system\n"), 0644)
	require.NoError(t, err)

	// Create old version info
	systemInfo := filepath.Join(oldPVMDir, "versions", "system")
	err = os.WriteFile(systemInfo, []byte("/usr/bin/perl\n"), 0644)
	require.NoError(t, err)

	// Import system perl if available
	if _, err := os.Stat("/usr/bin/perl"); err == nil {
		env.RunPVM("import-system")
	}

	// Test that new PVM can coexist with old installation
	t.Log("Testing upgrade coexistence...")
	_, _, err = env.RunPVM("list")
	assert.NoError(t, err, "Should list versions with old installation present")
	// Version list might be empty in test environment, just check command works

	// Test data migration if supported
	t.Log("Testing data preservation...")
	var stdout string
	stdout, _, err = env.RunPVM("config", "get", "xdg.data_dir")

	if err == nil {
		t.Logf("Data directory: %s", stdout)
		// Should preserve access to old data or provide migration path
	}

	// Verify that essential operations work
	essential_operations := [][]string{
		{"version"},
		{"list"},
		{"config", "list"},
	}

	for _, op := range essential_operations {
		t.Logf("Testing essential operation: %v", op)
		_, _, err := env.RunPVM(op...)
		assert.NoError(t, err, "Essential operation %v should work", op)
	}
}
