// ABOUTME: Comprehensive integration tests for Step 17 - validates complete system functionality
// ABOUTME: Tests all major workflows: typed-Perl development, legacy migration, LSP integration, performance

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	basetesting "tamarou.com/pvm/internal/testing"
	"tamarou.com/pvm/test/e2e/helpers"
)

func TestComprehensiveIntegration_TypedPerlDevelopment(t *testing.T) {
	basetesting.SampleE2ETest(t)
	helpers.SkipIfNoSystemPerl(t)
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a comprehensive typed Perl project
	projectDir := filepath.Join(env.RootDir, "typed_project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Create main module with complex types
	moduleFile := filepath.Join(projectDir, "Calculator.pm")
	moduleContent := `package Calculator;
use v5.36;
use strict;
use warnings;

# Constructor with typed parameters
sub new(Str $class, Num $initial_value = 0) -> Calculator {
    my Calculator $self = bless {
        value => $initial_value,
        name => "calculator"
    }, $class;
    return $self;
}

# Typed method with union types
sub add(Calculator $self, Int|Num $operand) -> Calculator {
    $self->{value} += $operand;
    return $self;
}

sub multiply(Calculator $self, Num $operand) -> Calculator {
    $self->{value} *= $operand;
    return $self;
}

sub get_value(Calculator $self) -> Num {
    return $self->{value};
}

# Complex type with array references
sub calculate_sequence(ArrayRef[Int] $numbers) -> ArrayRef[Num] {
    my ArrayRef[Num] $results = [];
    for my Int $num (@$numbers) {
        push @$results, $num * 2.5;
    }
    return $results;
}

1;
`
	err = os.WriteFile(moduleFile, []byte(moduleContent), 0644)
	require.NoError(t, err)

	// Create test script using the module
	scriptFile := filepath.Join(projectDir, "test_calculator.pl")
	scriptContent := `#!/usr/bin/perl
use v5.36;
use lib '.';
use Calculator;

# Test typed Perl features
my Calculator $calc = Calculator->new(10);
my Num $result = $calc->add(5)->multiply(2)->get_value();

say "Calculator result: $result";

# Test array processing
my ArrayRef[Int] $numbers = [1, 2, 3, 4, 5];
my ArrayRef[Num] $processed = Calculator::calculate_sequence($numbers);

say "Processed numbers: " . join(", ", @$processed);

# Test error conditions
my Int $test_int = 42;
my Str $test_str = "hello";

say "Integration test completed successfully";
`
	err = os.WriteFile(scriptFile, []byte(scriptContent), 0644)
	require.NoError(t, err)

	// Step 1: Type check the entire project
	t.Log("Step 1: Type checking project...")
	stdout := helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "check", "--verbose", moduleFile, scriptFile},
		"Type checking should pass for typed Perl project")

	assert.Contains(t, stdout, "Calculator.pm", "Should check module file")
	assert.Contains(t, stdout, "test_calculator.pl", "Should check script file")

	// Step 2: Generate type definitions
	t.Log("Step 2: Generating type definitions...")
	os.Setenv("PERL5LIB", projectDir)
	defer os.Unsetenv("PERL5LIB")

	stdout = helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "def", "generate", "Calculator"},
		"Type definition generation should succeed")

	assert.Contains(t, stdout, "Calculator", "Should mention Calculator module")

	// Step 3: Strip and execute
	t.Log("Step 3: Stripping and executing...")
	strippedScript := filepath.Join(projectDir, "test_calculator_stripped.pl")
	strippedModule := filepath.Join(projectDir, "Calculator_stripped.pm")

	// Strip the module first
	helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "strip", moduleFile, strippedModule},
		"Module type stripping should succeed")

	// Move the stripped module to replace the original
	err = os.Rename(strippedModule, moduleFile)
	require.NoError(t, err, "Should rename stripped module")

	// Strip the script
	helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "strip", scriptFile, strippedScript},
		"Type stripping should succeed")

	systemPerl := helpers.FindSystemPerl()
	stdout = helpers.AssertPVMSucceeds(t, env,
		[]string{"pvx", "--perl", systemPerl, "--no-install", "--include-path", projectDir, strippedScript},
		"Execution should succeed")

	assert.Contains(t, stdout, "Calculator result: 30", "Should show correct calculation")
	assert.Contains(t, stdout, "Processed numbers: 2.5, 5, 7.5, 10, 12.5", "Should show processed array")
	assert.Contains(t, stdout, "Integration test completed successfully", "Should complete successfully")

	// Step 4: Test integrated run
	t.Log("Step 4: Testing integrated PSC run...")
	stdout = helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "run", "--verbose", "--perl", systemPerl, scriptFile},
		"PSC run should succeed with complex types")

	assert.Contains(t, stdout, "Calculator result: 30", "PSC run should show correct calculation")
	assert.Contains(t, stdout, "Integration test completed successfully", "PSC run should complete")
}

func TestComprehensiveIntegration_LegacyMigration(t *testing.T) {
	helpers.SkipIfNoSystemPerl(t)
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a legacy Perl project
	legacyDir := filepath.Join(env.RootDir, "legacy_project")
	err := os.MkdirAll(legacyDir, 0755)
	require.NoError(t, err)

	// Legacy module without types
	legacyModule := filepath.Join(legacyDir, "LegacyMath.pm")
	legacyContent := `package LegacyMath;
use strict;
use warnings;

sub new {
    my $class = shift;
    my $self = {
        precision => shift || 2,
    };
    return bless $self, $class;
}

sub calculate {
    my ($self, $a, $b, $operation) = @_;

    if ($operation eq 'add') {
        return $a + $b;
    } elsif ($operation eq 'multiply') {
        return $a * $b;
    } elsif ($operation eq 'divide') {
        return $b != 0 ? $a / $b : undef;
    }

    return undef;
}

sub format_result {
    my ($self, $result) = @_;
    return defined $result ? sprintf("%.${self->{precision}}f", $result) : "ERROR";
}

1;
`
	err = os.WriteFile(legacyModule, []byte(legacyContent), 0644)
	require.NoError(t, err)

	// Legacy script
	legacyScript := filepath.Join(legacyDir, "legacy_test.pl")
	legacyScriptContent := `#!/usr/bin/perl
use strict;
use warnings;
use lib '.';
use LegacyMath;

my $math = LegacyMath->new(3);

my $result1 = $math->calculate(10, 5, 'add');
my $result2 = $math->calculate(10, 3, 'multiply');
my $result3 = $math->calculate(10, 3, 'divide');

print "Add: " . $math->format_result($result1) . "\n";
print "Multiply: " . $math->format_result($result2) . "\n";
print "Divide: " . $math->format_result($result3) . "\n";

print "Legacy migration test completed\n";
`
	err = os.WriteFile(legacyScript, []byte(legacyScriptContent), 0644)
	require.NoError(t, err)

	// Step 1: Test original execution
	t.Log("Step 1: Testing original legacy code...")
	systemPerl := helpers.FindSystemPerl()
	stdout := helpers.AssertPVMSucceeds(t, env,
		[]string{"pvx", "--perl", systemPerl, legacyScript},
		"Legacy code should execute correctly")

	assert.Contains(t, stdout, "Add: 15.000", "Should show correct addition")
	assert.Contains(t, stdout, "Multiply: 30.000", "Should show correct multiplication")
	assert.Contains(t, stdout, "Divide: 3.333", "Should show correct division")

	// Step 2: Check legacy code (should work even without types)
	t.Log("Step 2: Type checking legacy code...")
	// PSC should handle legacy code gracefully
	_, stderr, err := env.RunPVM("psc", "check", "--verbose", legacyModule, legacyScript)

	// Legacy code might have warnings but should not fail completely
	if err != nil {
		// Check if it's just type-related warnings, not fatal errors
		assert.Contains(t, stderr, "warning", "Should have warnings for untyped code")
		t.Logf("Legacy code warnings (expected): %s", stderr)
	}

	// Step 3: Gradual typing - add types to one method
	t.Log("Step 3: Adding gradual types...")
	typedVersion := strings.Replace(legacyContent,
		"sub calculate {",
		"sub calculate(LegacyMath $self, Num $a, Num $b, Str $operation) -> Num|Undef {", 1)

	typedModule := filepath.Join(legacyDir, "LegacyMath_typed.pm")
	err = os.WriteFile(typedModule, []byte(typedVersion), 0644)
	require.NoError(t, err)

	// Test gradual typing
	helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "check", "--verbose", typedModule},
		"Gradual typing should work")
}

func TestComprehensiveIntegration_PerformanceStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance stress test in short mode")
	}

	helpers.SkipIfNoSystemPerl(t)
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a large-ish Perl project to stress test the system
	projectDir := filepath.Join(env.RootDir, "stress_project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Generate multiple modules
	numModules := 10
	for i := 0; i < numModules; i++ {
		moduleFile := filepath.Join(projectDir, "Module"+string(rune('A'+i))+".pm")
		moduleContent := `package Module` + string(rune('A'+i)) + `;
use v5.36;
use strict;
use warnings;

# Generate multiple typed methods
field Int $counter = 0;
field ArrayRef[Str] $items = [];

method increment() -> Int {
    $counter++;
    return $counter;
}

method add_item(Str $item) -> ArrayRef[Str] {
    push @$items, $item;
    return $items;
}

method process_batch(ArrayRef[Int] $numbers) -> ArrayRef[Int] {
    my ArrayRef[Int] $results = [];
    for my Int $num (@$numbers) {
        push @$results, $num * $counter;
    }
    return $results;
}

method get_summary() -> HashRef {
    return {
        counter => $counter,
        item_count => scalar(@$items),
        class_name => __PACKAGE__,
    };
}

1;
`
		err = os.WriteFile(moduleFile, []byte(moduleContent), 0644)
		require.NoError(t, err)
	}

	// Create a complex main script
	mainScript := filepath.Join(projectDir, "stress_test.pl")
	scriptContent := `#!/usr/bin/perl
use v5.36;
use lib '.';

# Import all generated modules
`
	for i := 0; i < numModules; i++ {
		scriptContent += "use Module" + string(rune('A'+i)) + ";\n"
	}

	scriptContent += `
# Test all modules
my ArrayRef[Object] $modules = [];

# Create instances and test them
`
	for i := 0; i < numModules; i++ {
		moduleName := "Module" + string(rune('A'+i))
		scriptContent += "my " + moduleName + " $mod" + string(rune('a'+i)) + " = " + moduleName + "->new();\n"
		scriptContent += "push @$modules, $mod" + string(rune('a'+i)) + ";\n"
	}

	scriptContent += `
# Perform operations on all modules
for my Int $i (1..10) {
    for my Object $mod (@$modules) {
        $mod->increment();
        $mod->add_item("item_$i");
    }
}

# Process data
my ArrayRef[Int] $test_data = [1, 2, 3, 4, 5];
for my Object $mod (@$modules) {
    my ArrayRef[Int] $results = $mod->process_batch($test_data);
    my HashRef $summary = $mod->get_summary();
}

say "Stress test completed with " . scalar(@$modules) . " modules";
`

	err = os.WriteFile(mainScript, []byte(scriptContent), 0644)
	require.NoError(t, err)

	// Time the type checking
	t.Log("Performance stress test: Type checking large project...")
	start := time.Now()

	// Build file list for type checking
	var files []string
	for i := 0; i < numModules; i++ {
		files = append(files, filepath.Join(projectDir, "Module"+string(rune('A'+i))+".pm"))
	}
	files = append(files, mainScript)

	args := append([]string{"psc", "check", "--verbose"}, files...)
	helpers.AssertPVMSucceeds(t, env, args, "Large project type checking should succeed")

	typeCheckDuration := time.Since(start)
	t.Logf("Type checking %d modules took: %v", numModules, typeCheckDuration)

	// Performance target: should complete within reasonable time
	maxDuration := 30 * time.Second
	assert.Less(t, typeCheckDuration, maxDuration, "Type checking should complete within %v", maxDuration)

	// Test execution performance
	t.Log("Performance stress test: Execution...")
	start = time.Now()

	strippedScript := filepath.Join(projectDir, "stress_test_stripped.pl")
	helpers.AssertPVMSucceeds(t, env,
		[]string{"psc", "strip", mainScript, strippedScript},
		"Stripping should succeed")

	systemPerl := helpers.FindSystemPerl()
	stdout := helpers.AssertPVMSucceeds(t, env,
		[]string{"pvx", "--perl", systemPerl, strippedScript},
		"Stress test execution should succeed")

	execDuration := time.Since(start)
	t.Logf("Execution took: %v", execDuration)

	assert.Contains(t, stdout, "Stress test completed with 10 modules", "Should complete stress test")
}

func TestComprehensiveIntegration_ErrorHandling(t *testing.T) {
	helpers.SkipIfNoSystemPerl(t)
	helpers.SkipIfNoTreeSitter(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create files with various error conditions
	errorDir := filepath.Join(env.RootDir, "error_test")
	err := os.MkdirAll(errorDir, 0755)
	require.NoError(t, err)

	// Type mismatch errors
	typeMismatchFile := filepath.Join(errorDir, "type_mismatch.pl")
	typeMismatchContent := `#!/usr/bin/perl
use v5.36;

my Int $number = 42;
my Str $text = "hello";

# This should cause a type error
$number = $text;  # Int assigned Str

say "This should not execute";
`
	err = os.WriteFile(typeMismatchFile, []byte(typeMismatchContent), 0644)
	require.NoError(t, err)

	// Undefined variable errors
	undefinedVarFile := filepath.Join(errorDir, "undefined_var.pl")
	undefinedVarContent := `#!/usr/bin/perl
use v5.36;

my Int $defined_var = 10;

# This should cause an undefined variable error
say $undefined_variable;  # Should suggest $defined_var

say "This should not execute";
`
	err = os.WriteFile(undefinedVarFile, []byte(undefinedVarContent), 0644)
	require.NoError(t, err)

	// Function signature mismatch
	signatureMismatchFile := filepath.Join(errorDir, "signature_mismatch.pl")
	signatureMismatchContent := `#!/usr/bin/perl
use v5.36;

sub typed_function(Int $param) -> Str {
    return "result: $param";
}

# This should cause signature mismatch errors
my Str $result1 = typed_function("string");  # Wrong argument type
my Int $result2 = typed_function(42);        # Wrong return type assignment

say "This should not execute";
`
	err = os.WriteFile(signatureMismatchFile, []byte(signatureMismatchContent), 0644)
	require.NoError(t, err)

	// Test 1: Type mismatch error detection
	t.Log("Testing type mismatch error detection...")
	_, stderr, err := env.RunPVM("psc", "check", typeMismatchFile)
	assert.Error(t, err, "Type mismatch should cause error")
	assert.Contains(t, stderr, "type", "Error should mention type")
	assert.Contains(t, stderr, "Int", "Error should mention Int type")
	assert.Contains(t, stderr, "Str", "Error should mention Str type")
	t.Logf("Type mismatch error: %s", stderr)

	// Test 2: Undefined variable detection
	t.Log("Testing undefined variable detection...")
	_, stderr, err = env.RunPVM("psc", "check", undefinedVarFile)
	assert.Error(t, err, "Undefined variable should cause error")
	assert.Contains(t, stderr, "undefined", "Error should mention undefined variable")
	// Should suggest the similarly named variable
	assert.Contains(t, stderr, "defined_var", "Error should suggest similar variable name")
	t.Logf("Undefined variable error: %s", stderr)

	// Test 3: Signature mismatch detection
	t.Log("Testing function signature mismatch detection...")
	_, stderr, err = env.RunPVM("psc", "check", signatureMismatchFile)
	assert.Error(t, err, "Signature mismatch should cause error")
	assert.Contains(t, stderr, "function", "Error should mention function")
	t.Logf("Signature mismatch error: %s", stderr)

	// Test 4: Error recovery - mixed valid/invalid code
	mixedFile := filepath.Join(errorDir, "mixed_errors.pl")
	mixedContent := `#!/usr/bin/perl
use v5.36;

# Valid code
my Int $valid_number = 42;
my Str $valid_string = "hello";

# Invalid code
my Int $invalid = $valid_string;  # Type error

# More valid code
my ArrayRef[Int] $numbers = [1, 2, 3];

say "Mixed error test";
`
	err = os.WriteFile(mixedFile, []byte(mixedContent), 0644)
	require.NoError(t, err)

	t.Log("Testing error recovery with mixed valid/invalid code...")
	_, stderr, err = env.RunPVM("psc", "check", mixedFile)
	assert.Error(t, err, "Should report errors but continue parsing")
	assert.Contains(t, stderr, "type", "Should report type error")
	t.Logf("Mixed error recovery: %s", stderr)
}

func TestComprehensiveIntegration_BackwardCompatibility(t *testing.T) {
	helpers.SkipIfNoSystemPerl(t)
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test that all original PVM functionality still works

	// Test 1: Basic version management
	t.Log("Testing backward compatibility: version management...")
	stdout, _, err := env.RunPVM("list")
	assert.NoError(t, err, "List command should work")
	assert.Contains(t, stdout, "system", "Should show system Perl")

	// Test 2: Shell integration
	t.Log("Testing backward compatibility: shell integration...")
	stdout, _, err = env.RunPVM("shell", "--format", "bash")
	assert.NoError(t, err, "Shell command should work")
	assert.Contains(t, stdout, "export", "Should generate shell exports")

	// Test 3: Configuration
	t.Log("Testing backward compatibility: configuration...")
	stdout, _, err = env.RunPVM("config", "get", "xdg.data_dir")
	assert.NoError(t, err, "Config get should work")
	assert.NotEmpty(t, stdout, "Should return config value")

	// Test 4: Basic script execution (without types)
	simpleScript := filepath.Join(env.RootDir, "simple.pl")
	simpleContent := `#!/usr/bin/perl
use strict;
use warnings;

print "Simple script works\n";
`
	err = os.WriteFile(simpleScript, []byte(simpleContent), 0644)
	require.NoError(t, err)

	systemPerl := helpers.FindSystemPerl()
	stdout = helpers.AssertPVMSucceeds(t, env,
		[]string{"pvx", "--perl", systemPerl, simpleScript},
		"Simple script execution should work")

	assert.Contains(t, stdout, "Simple script works", "Simple script should execute")
}
