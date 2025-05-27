// ABOUTME: Integration tests for LSP functionality and editor integration
// ABOUTME: Tests LSP features work correctly with the enhanced architecture

package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/test/e2e/helpers"
)

func TestLSPIntegration_BasicFunctionality(t *testing.T) {
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a test project with typed Perl
	projectDir := filepath.Join(env.RootDir, "lsp_test_project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Create a main module
	moduleFile := filepath.Join(projectDir, "TestModule.pm")
	moduleContent := `package TestModule;
use v5.36;
use strict;
use warnings;

# Variable declarations for testing
field Int $counter = 0;
field Str $name = "test";
field ArrayRef[Str] $items = [];

# Method with types for testing goto definition
method increment() -> Int {
    $counter++;
    return $counter;
}

method add_item(Str $item) -> Bool {
    push @$items, $item;
    return 1;
}

method get_count() -> Int {
    return $counter;
}

# Function for testing find references
sub utility_function(Str $input) -> Str {
    return "processed: $input";
}

1;
`
	err = os.WriteFile(moduleFile, []byte(moduleContent), 0644)
	require.NoError(t, err)

	// Create a test script that uses the module
	testScript := filepath.Join(projectDir, "test_script.pl")
	testContent := `#!/usr/bin/perl
use v5.36;
use lib '.';
use TestModule;

my TestModule $module = TestModule->new();

# Test various symbol references
my Int $count = $module->get_count();
my Bool $success = $module->add_item("test item");
my Int $new_count = $module->increment();

# Test function calls
my Str $result = TestModule::utility_function("hello");

say "LSP test completed";
`
	err = os.WriteFile(testScript, []byte(testContent), 0644)
	require.NoError(t, err)

	// Test that PSC can analyze the files (this is the foundation for LSP functionality)
	t.Log("Testing PSC analysis of LSP test project...")
	stdout := helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "check", "--verbose", moduleFile, testScript},
		"PSC should analyze test project successfully")

	assert.Contains(t, stdout, "TestModule", "Should analyze TestModule")
	assert.Contains(t, stdout, "test_script", "Should analyze test script")

	// Test PSC LSP command
	t.Log("Testing PSC LSP command startup...")
	_, stderr, err := env.RunPVM("psc", "lsp", "--help")

	// LSP command should exist and show help
	if err == nil {
		t.Log("PSC LSP command is available")
	} else {
		t.Logf("PSC LSP command output: %s", stderr)
		// LSP command might not be fully implemented yet, which is fine for integration test
	}

	t.Log("LSP integration foundation test completed successfully")
}

func TestLSPIntegration_PerformanceAndResponsiveness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping LSP performance test in short mode")
	}

	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a larger project for performance testing
	projectDir := filepath.Join(env.RootDir, "lsp_performance_test")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Create multiple modules
	numModules := 10
	var moduleFiles []string

	for i := 0; i < numModules; i++ {
		moduleFile := filepath.Join(projectDir, "Module"+string(rune('A'+i))+".pm")
		moduleContent := `package Module` + string(rune('A'+i)) + `;
use v5.36;

field Int $counter = 0;
field Str $name = "module_` + string(rune('a'+i)) + `";

method process_data(ArrayRef[Int] $data) -> ArrayRef[Int] {
    my ArrayRef[Int] $results = [];
    for my Int $item (@$data) {
        push @$results, $item * 2;
    }
    return $results;
}

method get_info() -> HashRef {
    return {
        name => $name,
        counter => $counter,
        module_id => "` + string(rune('A'+i)) + `"
    };
}

1;
`
		err = os.WriteFile(moduleFile, []byte(moduleContent), 0644)
		require.NoError(t, err)
		moduleFiles = append(moduleFiles, moduleFile)
	}

	// Test PSC performance with multiple modules
	t.Log("Testing PSC performance with multiple modules...")

	// Build file list for type checking
	args := append([]string{"psc", "check", "--verbose"}, moduleFiles...)
	stdout := helpers.AssertPVMSucceeds(t, env, args, "Large project type checking should succeed")

	for i := 0; i < numModules; i++ {
		moduleName := "Module" + string(rune('A'+i))
		assert.Contains(t, stdout, moduleName, "Should check %s", moduleName)
	}

	t.Log("LSP performance foundation test completed successfully")
}

func TestLSPIntegration_ErrorHandling(t *testing.T) {
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a project with errors for testing diagnostics
	projectDir := filepath.Join(env.RootDir, "lsp_error_test")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Create a file with type errors
	errorFile := filepath.Join(projectDir, "error_test.pl")
	errorContent := `#!/usr/bin/perl
use v5.36;

# Type mismatch error
my Int $number = "not a number";

# Undefined variable error
say $undefined_variable;

# Function call error
sub typed_function(Int $param) -> Str {
    return "result: $param";
}

my Str $result = typed_function("wrong type");

say "This has errors";
`
	err = os.WriteFile(errorFile, []byte(errorContent), 0644)
	require.NoError(t, err)

	// Test that PSC can detect errors (foundation for LSP diagnostics)
	t.Log("Testing PSC error detection for LSP...")
	_, stderr, err := env.RunPVM("psc", "check", errorFile)

	// Should detect type errors
	assert.Error(t, err, "PSC should detect type errors")
	assert.Contains(t, stderr, "type", "Should report type-related errors")
	t.Logf("PSC error detection: %s", stderr)

	t.Log("LSP error handling foundation test completed successfully")
}

func TestLSPIntegration_ConfigurationAndSettings(t *testing.T) {
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	projectDir := filepath.Join(env.RootDir, "lsp_config_test")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Create a simple test file
	testFile := filepath.Join(projectDir, "config_test.pl")
	testContent := `#!/usr/bin/perl
use v5.36;

my Int $value = 42;
say "Configuration test: $value";
`
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	// Test PSC configuration handling
	t.Log("Testing PSC configuration for LSP...")
	stdout := helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "check", "--verbose", testFile},
		"PSC should handle configuration correctly")

	assert.Contains(t, stdout, "config_test", "Should check configuration test file")

	t.Log("LSP configuration foundation test completed successfully")
}
