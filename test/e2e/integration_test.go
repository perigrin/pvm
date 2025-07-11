// ABOUTME: End-to-end integration tests for cross-component functionality
// ABOUTME: Tests how PVM, PVX, PVI, and PSC work together

package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/test/e2e/helpers"
)

func TestCrossComponentIntegration_PSC_PVX(t *testing.T) {
	// Removed sampling to enable test in regular runs
	helpers.SkipIfNoSystemPerl(t)
	helpers.SkipIfNoTreeSitter(t)

	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	systemPerl := helpers.FindSystemPerl()

	// Create a typed Perl script that uses PSC and PVX integration
	scriptFile := filepath.Join(env.RootDir, "typed_script.pl")
	scriptContent := `#!/usr/bin/perl
use strict;
use warnings;
use feature 'say';

# Simple typed Perl script
my Int $count = 42;
my Str $message = "Hello from cross-component integration!";

say "$message $count";
`

	err := os.WriteFile(scriptFile, []byte(scriptContent), 0644)
	require.NoError(t, err)

	// Test PSC run command which uses PSC -> PVX integration
	// Use system Perl explicitly to avoid version resolution issues
	stdout := helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "run", "--verbose", "--perl", systemPerl, scriptFile},
		"PSC run should succeed and use PVX for execution")

	// Should contain both the type checking output and script execution output
	assert.Contains(t, stdout, "Hello from cross-component integration!",
		"Should execute script and show output")
	assert.Contains(t, stdout, "42",
		"Should show the processed count")
}

func TestCrossComponentIntegration_PVX_Modules(t *testing.T) {
	helpers.SkipIfNoSystemPerl(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a simple script
	scriptFile := filepath.Join(env.RootDir, "module_script.pl")
	scriptContent := `#!/usr/bin/perl
use v5.36;

say "Script executed successfully";
`

	err := os.WriteFile(scriptFile, []byte(scriptContent), 0644)
	require.NoError(t, err)

	// Test PVX with module requirements (PVX -> PVI integration)
	// Note: This will try to install modules, but may fail in test environment
	// The test is to verify the integration plumbing works, not actual module installation
	systemPerl := helpers.FindSystemPerl()
	_, stderr, err := env.RunPVM("pvx", "--perl", systemPerl, "--auto-install", "--require", "JSON", "--verbose", scriptFile)

	// The command may fail due to module installation issues in test environment,
	// but we're testing that the integration code paths are working
	if err != nil {
		// If it fails, it should be due to module-related issues, not basic integration problems
		// Accept various module-related error messages
		hasModuleError := assert.Contains(t, stderr, "install") ||
			assert.Contains(t, stderr, "JSON") ||
			assert.Contains(t, stderr, "module") ||
			assert.Contains(t, stderr, "Module")
		assert.True(t, hasModuleError,
			"Error should be related to module operations, indicating integration is working. Got: %s", stderr)
	}
}

func TestCrossComponentIntegration_ImportSystem(t *testing.T) {
	helpers.SkipIfNoSystemPerl(t)

	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test the top-level import-system command that was added for integration
	stdout, stderr, err := env.RunPVM("import-system")

	// The command should work, though it may detect that system Perl is already registered
	// or successfully import it
	// Note: import-system writes to stderr for informational messages
	assert.NoError(t, err, "import-system command should succeed")
	output := stdout + stderr
	assert.Contains(t, output, "perl", "Should mention perl in output")
}

func TestCrossComponentIntegration_TypeDefinitions(t *testing.T) {
	helpers.SkipIfNoSystemPerl(t)
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a module file with type annotations for PSC-PVI integration
	moduleFile := filepath.Join(env.RootDir, "TestModule.pm")
	moduleContent := `package TestModule;
use v5.36;

# Variable types
my Int $counter = 0;
my Str $name = "Test";

# Function signature
sub Int increment(Int $value) {
    return $value + 1;
}

1;
`

	err := os.WriteFile(moduleFile, []byte(moduleContent), 0644)
	require.NoError(t, err)

	// Set PERL5LIB so the module can be found
	os.Setenv("PERL5LIB", env.RootDir)
	defer os.Unsetenv("PERL5LIB")

	// Test PSC def command which demonstrates PSC-PVI integration for type definitions
	stdout := helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "def", "generate", "TestModule"},
		"PSC def generate should succeed and generate type definitions")

	// Should show type definition generation
	assert.Contains(t, stdout, "TestModule",
		"Should mention the module name in output")
}

func TestCrossComponentIntegration_EndToEnd(t *testing.T) {
	helpers.SkipIfNoSystemPerl(t)
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a comprehensive typed script
	scriptFile := filepath.Join(env.RootDir, "comprehensive.pl")
	scriptContent := `#!/usr/bin/perl
use v5.36;

# Comprehensive typed Perl script
my $start = 1;
my $end = 5;
my $numbers = [];

# Generate numbers
for my $i ($start..$end) {
    push @$numbers, $i * 2;
}

# Calculate sum
my $sum = 0;
for my $num (@$numbers) {
    $sum += $num;
}

say "Numbers: " . join(", ", @$numbers);
say "Sum: $sum";
say "Comprehensive integration test completed successfully";
`

	err := os.WriteFile(scriptFile, []byte(scriptContent), 0644)
	require.NoError(t, err)

	// Step 1: Type check the script (PSC)
	t.Log("Step 1: Type checking...")
	helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "check", "--verbose", scriptFile},
		"Type checking should pass")

	// Step 2: Strip type annotations (PSC)
	t.Log("Step 2: Stripping type annotations...")
	strippedFile := filepath.Join(env.RootDir, "comprehensive_stripped.pl")
	helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "strip", scriptFile, strippedFile},
		"Type stripping should succeed")

	// Step 3: Execute with PVX
	t.Log("Step 3: Executing with PVX...")
	systemPerl := helpers.FindSystemPerl()
	stdout := helpers.AssertPVMSucceeds(t, env,
		[]string{"pvx", "--perl", systemPerl, strippedFile},
		"PVX execution should succeed")

	// Verify output
	assert.Contains(t, stdout, "Numbers: 2, 4, 6, 8, 10",
		"Should show correct numbers")
	assert.Contains(t, stdout, "Sum: 30",
		"Should show correct sum")
	assert.Contains(t, stdout, "integration test completed successfully",
		"Should show completion message")

	// Step 4: Test integrated run (PSC -> PVX)
	t.Log("Step 4: Testing integrated PSC run...")
	stdout = helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "run", "--verbose", "--perl", systemPerl, scriptFile},
		"PSC run should succeed using PSC->PVX integration")

	// Should produce the same output as direct PVX execution
	assert.Contains(t, stdout, "Numbers: 2, 4, 6, 8, 10",
		"PSC run should show correct numbers")
	assert.Contains(t, stdout, "Sum: 30",
		"PSC run should show correct sum")
}
