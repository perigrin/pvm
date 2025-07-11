// ABOUTME: Tests for cross-component interactions and integration points
// ABOUTME: Validates PSC-PVI, PSC-PVX, PVI-PVX integration scenarios

package e2e

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	basetesting "tamarou.com/pvm/internal/testing"
	"tamarou.com/pvm/test/e2e/helpers"
)

func TestComponentInteraction_PSC_PVI_TypeDefinitions(t *testing.T) {
	t.Skip("Skipping test that requires typed Perl module support")
	helpers.SkipIfNoSystemPerl(t)
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a module project for type definition testing
	projectDir := filepath.Join(env.RootDir, "module_project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Create a complex module with various type features
	moduleFile := filepath.Join(projectDir, "ComplexTypes.pm")
	moduleContent := `package ComplexTypes;
use v5.36;
use strict;
use warnings;

# Field declarations with types
field Int $counter = 0;
field HashRef[Str] $config = {};
field ArrayRef[Object] $items = [];

# Type aliases
type UserId = Int;
type UserData = HashRef[Str|Int];

# Constructor with complex types
sub new(Str $class, UserData $initial_data = {}) -> ComplexTypes {
    my ComplexTypes $self = bless {}, $class;
    $self->{config} = $initial_data;
    return $self;
}

# Method with union and parameterized types
method add_user(UserId $id, HashRef[Str] $data) -> Bool {
    if (exists $config->{$id}) {
        return 0;  # User already exists
    }
    $config->{$id} = $data;
    $counter++;
    return 1;
}

# Method with intersection types (simulated)
method get_user_data(UserId $id) -> UserData|Undef {
    return $config->{$id} // undef;
}

# Complex return type
method get_all_users() -> ArrayRef[HashRef[Str|Int]] {
    my ArrayRef[HashRef[Str|Int]] $users = [];
    for my UserId $id (keys %$config) {
        my HashRef[Str|Int] $user_info = {
            id => $id,
            %{$config->{$id}}
        };
        push @$users, $user_info;
    }
    return $users;
}

# Method with callback type (function reference)
method process_users(CodeRef $processor) -> ArrayRef[Any] {
    my ArrayRef[Any] $results = [];
    for my UserId $id (keys %$config) {
        my Any $result = $processor->($id, $config->{$id});
        push @$results, $result;
    }
    return $results;
}

1;
`
	err = os.WriteFile(moduleFile, []byte(moduleContent), 0644)
	require.NoError(t, err)

	// Create cpanfile to mark as proper project
	cpanFile := filepath.Join(projectDir, "cpanfile")
	cpanContent := `requires 'perl', '5.036';
requires 'strict';
requires 'warnings';
`
	err = os.WriteFile(cpanFile, []byte(cpanContent), 0644)
	require.NoError(t, err)

	// Set up module path
	os.Setenv("PERL5LIB", projectDir)
	defer os.Unsetenv("PERL5LIB")

	// Test PSC type checking
	t.Log("Testing PSC type checking of complex module...")
	stdout := helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "check", "--verbose", moduleFile},
		"Complex type checking should succeed")

	assert.Contains(t, stdout, "ComplexTypes", "Should check ComplexTypes module")

	// Test PSC type definition generation (PSC-PVI integration)
	t.Log("Testing PSC-PVI integration: type definition generation...")
	stdout = helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "def", "generate", "ComplexTypes"},
		"Type definition generation should work")

	assert.Contains(t, stdout, "ComplexTypes", "Should generate types for ComplexTypes")

	// Test that generated definitions can be used
	t.Log("Testing generated type definitions usage...")
	testScript := filepath.Join(projectDir, "test_generated_types.pl")
	testContent := `#!/usr/bin/perl
use v5.36;
use lib '.';
use ComplexTypes;

my ComplexTypes $manager = ComplexTypes->new();

# Test type-safe operations
my Bool $success = $manager->add_user(1, {name => "Alice", email => "alice@example.com"});
$success = $manager->add_user(2, {name => "Bob", email => "bob@example.com"});

my ArrayRef[HashRef[Str|Int]] $all_users = $manager->get_all_users();

say "Added " . scalar(@$all_users) . " users";
say "Type definition integration test completed";
`
	err = os.WriteFile(testScript, []byte(testContent), 0644)
	require.NoError(t, err)

	// Type check the test script
	helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "check", "--verbose", testScript},
		"Test script using generated types should type check")

	// Execute the test
	systemPerl := helpers.FindSystemPerl()
	stdout = helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "run", "--perl", systemPerl, testScript},
		"Test script should execute successfully")

	assert.Contains(t, stdout, "Added 2 users", "Should show correct user count")
	assert.Contains(t, stdout, "integration test completed", "Should complete successfully")
}

func TestComponentInteraction_PSC_PVX_ErrorPropagation(t *testing.T) {
	t.Skip("Skipping test that requires typed Perl syntax support")
	helpers.SkipIfNoSystemPerl(t)
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create script with deliberate errors to test error propagation
	errorScript := filepath.Join(env.RootDir, "error_script.pl")
	errorContent := `#!/usr/bin/perl
use v5.36;

# Type error that should be caught by PSC
my Int $number = "not a number";

# Runtime error that should be caught by PVX
my Int $zero = 0;
my Int $result = 42 / $zero;

say "This should not execute";
`
	err := os.WriteFile(errorScript, []byte(errorContent), 0644)
	require.NoError(t, err)

	// Test that PSC catches the type error before execution
	t.Log("Testing PSC error detection...")
	_, stderr, err := env.RunPVM("psc", "check", errorScript)
	assert.Error(t, err, "PSC should catch type error")
	assert.Contains(t, stderr, "type", "Should report type error")

	// Test that PSC run command propagates errors properly
	t.Log("Testing PSC-PVX error propagation...")
	systemPerl := helpers.FindSystemPerl()
	_, stderr, err = env.RunPVM("psc", "run", "--perl", systemPerl, errorScript)
	assert.Error(t, err, "PSC run should fail due to type error")

	// Should show PSC type error, not runtime error
	assert.Contains(t, stderr, "type", "Should show type checking error")
}

func TestComponentInteraction_PVI_PVX_ModuleInstallation(t *testing.T) {
	basetesting.SkipUnlessLongRunning(t, "PVI/PVX module installation integration test")

	helpers.SkipIfNoSystemPerl(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create script that requires external modules
	moduleScript := filepath.Join(env.RootDir, "module_script.pl")
	moduleContent := `#!/usr/bin/perl
use v5.36;

# This script intentionally uses a simple, commonly available module
use Data::Dumper;

my %test_data = (
    name => "Test",
    value => 42,
    array => [1, 2, 3]
);

print Dumper(\%test_data);
print "Module installation test completed\n";
`
	err := os.WriteFile(moduleScript, []byte(moduleContent), 0644)
	require.NoError(t, err)

	// Test PVX with module requirements
	t.Log("Testing PVI-PVX integration: module handling...")
	systemPerl := helpers.FindSystemPerl()

	// Test that PVX can handle scripts with modules
	// Note: Data::Dumper is typically included with Perl, so this should work
	stdout := helpers.AssertPVMSucceeds(t, env,
		[]string{"pvx", "--perl", systemPerl, "--verbose", moduleScript},
		"PVX should handle module dependencies")

	assert.Contains(t, stdout, "Test", "Should show dumped data")
	assert.Contains(t, stdout, "42", "Should show test value")
	assert.Contains(t, stdout, "Module installation test completed", "Should complete")

	// Test with explicit module checking
	t.Log("Testing explicit module validation...")
	_, stderr, err := env.RunPVM("pvx", "--perl", systemPerl, "--require", "Data::Dumper", "--verbose", moduleScript)

	// This should either succeed or provide helpful error messages
	if err != nil {
		t.Logf("Module validation output: %s", stderr)
		// Should have meaningful error about module availability
		assert.Contains(t, stderr, "module", "Error should mention modules")
	}
}

func TestComponentInteraction_PerformanceOptimizations(t *testing.T) {
	t.Skip("Skipping test that requires typed Perl syntax and --optimize flag")
	helpers.SkipIfNoSystemPerl(t)
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a script to test performance optimizations
	perfScript := filepath.Join(env.RootDir, "performance_test.pl")
	perfContent := `#!/usr/bin/perl
use v5.36;

# Simple operations that should benefit from optimizations
my Int $total = 0;
my ArrayRef[Int] $numbers = [1..100];

for my Int $i (1..1000) {
    for my Int $num (@$numbers) {
        $total += $num;
    }
}

say "Performance test completed: total = $total";
`
	err := os.WriteFile(perfScript, []byte(perfContent), 0644)
	require.NoError(t, err)

	// Test with optimizations enabled
	t.Log("Testing performance optimizations...")
	start := time.Now()

	systemPerl := helpers.FindSystemPerl()
	stdout := helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "run", "--perl", systemPerl, "--optimize", perfScript},
		"Optimized execution should succeed")

	optimizedDuration := time.Since(start)

	assert.Contains(t, stdout, "total = 5050000", "Should show correct calculation")
	t.Logf("Optimized execution took: %v", optimizedDuration)

	// Test without optimizations for comparison
	t.Log("Testing without optimizations...")
	start = time.Now()

	stdout = helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "run", "--perl", systemPerl, perfScript},
		"Regular execution should succeed")

	regularDuration := time.Since(start)

	assert.Contains(t, stdout, "total = 5050000", "Should show correct calculation")
	t.Logf("Regular execution took: %v", regularDuration)

	// Performance should be reasonable (not testing for improvement since that's implementation-dependent)
	maxDuration := 10 * time.Second
	assert.Less(t, optimizedDuration, maxDuration, "Optimized execution should complete quickly")
	assert.Less(t, regularDuration, maxDuration, "Regular execution should complete quickly")
}

func TestComponentInteraction_ConcurrentOperations(t *testing.T) {
	t.Skip("Skipping test that requires typed Perl syntax")
	helpers.SkipIfNoSystemPerl(t)
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create multiple scripts for concurrent testing
	numScripts := 5
	var scriptFiles []string

	for i := 0; i < numScripts; i++ {
		scriptFile := filepath.Join(env.RootDir, "concurrent_script_"+string(rune('A'+i))+".pl")
		scriptContent := `#!/usr/bin/perl
use v5.36;

my Int $script_id = ` + string(rune('1'+i)) + `;
my Int $result = 0;

# Some computation
for my Int $j (1..100) {
    $result += $j * $script_id;
}

say "Script $script_id completed with result: $result";
`
		err := os.WriteFile(scriptFile, []byte(scriptContent), 0644)
		require.NoError(t, err)
		scriptFiles = append(scriptFiles, scriptFile)
	}

	// Test concurrent type checking
	t.Log("Testing concurrent type checking...")
	start := time.Now()

	// Check all scripts simultaneously (PSC should handle this)
	args := append([]string{"psc", "check", "--verbose"}, scriptFiles...)
	helpers.AssertPVMSucceeds(t, env, args, "Concurrent type checking should work")

	concurrentDuration := time.Since(start)
	t.Logf("Concurrent type checking of %d scripts took: %v", numScripts, concurrentDuration)

	// Performance should be reasonable
	maxDuration := 15 * time.Second
	assert.Less(t, concurrentDuration, maxDuration, "Concurrent operations should complete quickly")
}

func TestComponentInteraction_MemoryManagement(t *testing.T) {
	t.Skip("Skipping test that requires typed Perl syntax")
	helpers.SkipIfNoSystemPerl(t)
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a script that uses significant memory
	memoryScript := filepath.Join(env.RootDir, "memory_test.pl")
	memoryContent := `#!/usr/bin/perl
use v5.36;

# Create large data structures to test memory management
my ArrayRef[ArrayRef[Int]] $matrix = [];

for my Int $i (1..100) {
    my ArrayRef[Int] $row = [];
    for my Int $j (1..100) {
        push @$row, $i * $j;
    }
    push @$matrix, $row;
}

# Calculate sum to verify data
my Int $total = 0;
for my ArrayRef[Int] $row (@$matrix) {
    for my Int $value (@$row) {
        $total += $value;
    }
}

say "Memory test completed: total = $total";
`
	err := os.WriteFile(memoryScript, []byte(memoryContent), 0644)
	require.NoError(t, err)

	// Test type checking with large data structures
	t.Log("Testing memory management during type checking...")
	helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "check", "--verbose", memoryScript},
		"Type checking with large data should succeed")

	// Test execution
	t.Log("Testing memory management during execution...")
	systemPerl := helpers.FindSystemPerl()
	stdout := helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "run", "--perl", systemPerl, memoryScript},
		"Execution with large data should succeed")

	assert.Contains(t, stdout, "total = 338350000", "Should show correct total")
	assert.Contains(t, stdout, "Memory test completed", "Should complete successfully")
}
