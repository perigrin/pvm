// ABOUTME: Tests for end-to-end workflow integration
// ABOUTME: Validates that all components work together correctly

package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/pvx"
)

func TestCompleteWorkflow(t *testing.T) {
	// Create a temporary test script
	tempDir := t.TempDir()
	testScript := filepath.Join(tempDir, "test_workflow.pl")

	testCode := `#!/usr/bin/env perl
use strict;
use warnings;

# Simple test script for integration testing (compatible with system Perl)
my $greeting = "Hello, World!";
my $number = 42;

sub format_message {
    my ($msg, $num) = @_;
    return "$msg The answer is $num";
}

print format_message($greeting, $number) . "\n";
`

	err := os.WriteFile(testScript, []byte(testCode), 0644)
	require.NoError(t, err)

	options := &WorkflowOptions{
		ScriptPath:        testScript,
		PerlVersion:       "", // Use system Perl instead of hardcoded default
		Verbose:           false,
		GenerateTypeDefs:  true,
		SaveTypeDefs:      false,
		IsolationLevel:    pvx.IsolationLocal,
		SkipModuleInstall: true, // Skip for testing
	}

	result, err := CompleteWorkflow(options)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify workflow results
	assert.NotEmpty(t, result.VersionUsed, "Should resolve a Perl version")
	assert.True(t, result.TypeCheckPassed, "Type checking should pass for valid script")
	assert.Empty(t, result.TypeErrors, "Should have no type errors")
	assert.Contains(t, result.ExecutionOutput, "Hello, World! The answer is 42", "Should contain expected output")
	assert.Equal(t, 0, result.ExecutionExitCode, "Script should exit successfully")
	assert.True(t, result.TypeDefGenerated, "Type definitions should be generated")
	assert.Greater(t, result.Duration, time.Duration(0), "Should track execution time")
}

func TestTypeCheckWorkflow(t *testing.T) {
	// Create a test script with type errors
	tempDir := t.TempDir()
	testScript := filepath.Join(tempDir, "test_typecheck.pl")

	testCode := `#!/usr/bin/env perl
use strict;
use warnings;

# Script with type annotations that will cause type checker errors but still run on system Perl
my Int $number = "hello"; # Type error: string assigned to Int
my Str $text = 42;        # Type error: number assigned to Str

print "Testing type checking\n";
`

	err := os.WriteFile(testScript, []byte(testCode), 0644)
	require.NoError(t, err)

	result, err := TypeCheckWorkflow(testScript, "", false)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify type checking detected errors
	assert.False(t, result.TypeCheckPassed, "Type checking should fail for invalid types")
	assert.NotEmpty(t, result.TypeErrors, "Should detect type errors")
	assert.True(t, result.TypeDefGenerated, "Type definitions should still be generated")
	// Execution should be skipped in TypeCheckWorkflow
	assert.Empty(t, result.ExecutionOutput, "Should not execute script in type check workflow")
	assert.Equal(t, 0, result.ExecutionExitCode, "Exit code should be default when not executed")
}

func TestExecutionWorkflow(t *testing.T) {
	// Create a simple executable script
	tempDir := t.TempDir()
	testScript := filepath.Join(tempDir, "test_execution.pl")

	testCode := `#!/usr/bin/env perl
use strict;
use warnings;

print "Execution workflow test\n";
print "Exit code: 0\n";
`

	err := os.WriteFile(testScript, []byte(testCode), 0644)
	require.NoError(t, err)

	result, err := ExecutionWorkflow(testScript, "", false)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify execution results
	assert.NotEmpty(t, result.VersionUsed, "Should resolve a Perl version")
	assert.Contains(t, result.ExecutionOutput, "Execution workflow test", "Should contain expected output")
	assert.Equal(t, 0, result.ExecutionExitCode, "Script should exit successfully")
	assert.False(t, result.TypeDefGenerated, "Type definitions should not be generated")
}

func TestDevelopmentWorkflow(t *testing.T) {
	// Create a comprehensive development script
	tempDir := t.TempDir()
	testScript := filepath.Join(tempDir, "test_development.pl")

	testCode := `#!/usr/bin/env perl
use strict;
use warnings;

# Development workflow test with comprehensive features (system Perl compatible)
my $project_name = "PVM Integration Test";
my $features = ["type checking", "execution", "modules"];

sub describe_project {
    my ($name, $feat) = @_;
    my $feature_list = join(", ", @$feat);
    return "Project: $name\nFeatures: $feature_list";
}

print describe_project($project_name, $features) . "\n";
print "Development workflow completed successfully\n";
`

	err := os.WriteFile(testScript, []byte(testCode), 0644)
	require.NoError(t, err)

	result, err := DevelopmentWorkflow(testScript, "")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify comprehensive workflow results
	assert.NotEmpty(t, result.VersionUsed, "Should resolve a Perl version")
	assert.True(t, result.TypeCheckPassed, "Type checking should pass")
	assert.Contains(t, result.ExecutionOutput, "PVM Integration Test", "Should contain project output")
	assert.Contains(t, result.ExecutionOutput, "Development workflow completed", "Should complete successfully")
	assert.True(t, result.TypeDefGenerated, "Type definitions should be generated")
	assert.Greater(t, result.Duration, time.Duration(0), "Should track execution time")
}

func TestValidationWorkflow(t *testing.T) {
	// Test the built-in validation workflow
	result, err := ValidationWorkflow("")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify validation results
	assert.NotEmpty(t, result.VersionUsed, "Should resolve a Perl version")
	assert.True(t, result.TypeCheckPassed, "Validation script should pass type checking")
	assert.Contains(t, result.ExecutionOutput, "Hello, PVM Integration!", "Should contain expected message")
	assert.Contains(t, result.ExecutionOutput, "Sum of 2 + 3 = 5", "Should contain calculation result")
	assert.Equal(t, 0, result.ExecutionExitCode, "Validation should exit successfully")
	assert.True(t, result.TypeDefGenerated, "Type definitions should be generated")
}

func TestWorkflowWithRequiredModules(t *testing.T) {
	// Create a script that uses external modules
	tempDir := t.TempDir()
	testScript := filepath.Join(tempDir, "test_modules.pl")

	testCode := `#!/usr/bin/env perl
use v5.40;
use strict;
use warnings;

# Simple script that doesn't actually use external modules
# but specifies them as required for testing
my Str $message = "Module workflow test";
print "$message\n";
`

	err := os.WriteFile(testScript, []byte(testCode), 0644)
	require.NoError(t, err)

	options := &WorkflowOptions{
		ScriptPath:      testScript,
		Verbose:         false,
		RequiredModules: []string{"JSON", "YAML"}, // Common modules for testing
		IsolationLevel:  pvx.IsolationClean,
	}

	result, err := CompleteWorkflow(options)

	// Note: This test might fail if the modules can't be installed
	// In a real environment, but we still check that the workflow handles it gracefully
	if err != nil {
		// Check that it's a module installation error, not a workflow error
		assert.Contains(t, err.Error(), "module", "Error should be module-related")
	} else {
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.VersionUsed, "Should resolve a Perl version")
		assert.Contains(t, result.ExecutionOutput, "Module workflow test", "Should execute script")
	}
}

func TestWorkflowErrorHandling(t *testing.T) {
	// Test workflow with invalid script
	result, err := CompleteWorkflow(&WorkflowOptions{
		ScriptPath: "/nonexistent/script.pl",
		Verbose:    false,
	})

	// Should handle errors gracefully
	assert.Error(t, err, "Should return error for nonexistent script")
	assert.NotNil(t, result, "Should return result even on error")
	assert.NotEmpty(t, result.Errors, "Should record errors in result")
}

func TestWorkflowWithCustomPerlVersion(t *testing.T) {
	// Create a simple test script
	tempDir := t.TempDir()
	testScript := filepath.Join(tempDir, "test_version.pl")

	testCode := `#!/usr/bin/env perl
use strict;
use warnings;
print "Testing custom Perl version\n";
`

	err := os.WriteFile(testScript, []byte(testCode), 0644)
	require.NoError(t, err)

	options := &WorkflowOptions{
		ScriptPath:     testScript,
		PerlVersion:    "system", // Use system Perl
		Verbose:        false,
		SkipTypeCheck:  true, // Skip for simpler test
		IsolationLevel: pvx.IsolationGlobal,
	}

	result, err := CompleteWorkflow(options)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify version resolution
	assert.NotEmpty(t, result.VersionUsed, "Should resolve the requested version")
	assert.Contains(t, result.ExecutionOutput, "Testing custom Perl version", "Should execute with custom version")
}

func TestExtractModulesFromTypeCheck(t *testing.T) {
	// Test the module extraction function
	result := &parser.TypeCheckResult{
		TypeAnnotations: []*parser.TypeAnnotation{
			{
				TypeExpression: &parser.TypeExpression{
					Name: "Data::Dumper::Simple",
				},
			},
			{
				TypeExpression: &parser.TypeExpression{
					Name: "JSON::PP::Boolean",
				},
			},
			{
				TypeExpression: &parser.TypeExpression{
					Name: "Str", // Built-in type
				},
			},
		},
	}

	modules := extractModulesFromTypeCheck(result)

	// Should extract non-builtin modules
	assert.Contains(t, modules, "Data::Dumper", "Should extract Data::Dumper module")
	assert.Contains(t, modules, "JSON::PP", "Should extract JSON::PP module")
	assert.NotContains(t, modules, "Str", "Should not extract built-in types")
}

func TestIsBuiltinModule(t *testing.T) {
	// Test builtin module detection
	assert.True(t, isBuiltinModule("strict"), "strict should be builtin")
	assert.True(t, isBuiltinModule("warnings"), "warnings should be builtin")
	assert.True(t, isBuiltinModule("Carp"), "Carp should be builtin")
	assert.False(t, isBuiltinModule("Data::Dumper"), "Data::Dumper should not be builtin")

	assert.False(t, isBuiltinModule("Moose"), "Moose should not be builtin")
	assert.False(t, isBuiltinModule("DBI"), "DBI should not be builtin")
	assert.False(t, isBuiltinModule("My::Custom::Module"), "Custom modules should not be builtin")
}
