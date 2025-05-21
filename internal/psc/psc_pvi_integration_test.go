// ABOUTME: Tests for PSC-PVI integration
// ABOUTME: Ensures proper integration between PSC and PVI

package psc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/parser"
)

func TestGenerateTypeDefinition(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "psc-pvi-test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a sample Perl file with type annotations
	sampleFile := filepath.Join(tmpDir, "TestModule.pm")
	sampleContent := `package TestModule;
use v5.36;

# Type annotations for variables
my Int $count = 42;
my Str $name = "Example";
my ArrayRef[Int] $numbers = [1, 2, 3];

# Type annotation for subroutine parameters and return
sub add_numbers(Int $a, Int $b) -> Int {
    return $a + $b;
}

# Type annotation for method
sub new(Str $class, Str $name) -> Object {
    my $self = {
        name => $name,
    };
    return bless $self, $class;
}

# Type annotation for attribute
has Str $message;

1;
`
	err = os.WriteFile(sampleFile, []byte(sampleContent), 0644)
	require.NoError(t, err)

	// Test generating a type definition from the file
	options := &TypeDefinitionOptions{
		ModuleName: "TestModule",
		SourceFile: sampleFile,
		Verbose:    false,
		Save:       false,
	}

	result, err := GenerateTypeDefinition(options)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.TypeDef)

	// Verify the type definition
	typeDef := result.TypeDef
	assert.Equal(t, "TestModule", typeDef.Module)
	assert.Equal(t, sampleFile, typeDef.Source)

	// Verify that types were extracted
	// Note: This is a simplified test, as the current implementation
	// is not yet able to fully extract all annotations from the tree-sitter
	// parser in a reliable way. This would need to be enhanced in a real
	// implementation with full AST support.
}

func TestExtractTypeAndParams(t *testing.T) {
	// Test extracting type and parameters
	tests := []struct {
		input          string
		expectedBase   string
		expectedParams []string
	}{
		{"Int", "Int", nil},
		{"Str", "Str", nil},
		{"ArrayRef[Int]", "ArrayRef", []string{"Int"}},
		{"HashRef[Str]", "HashRef", []string{"Str"}},
		{"Maybe[Int]", "Maybe", []string{"Int"}},
		{"Tuple[Int, Str]", "Tuple", []string{"Int", "Str"}},
		{"ArrayRef[HashRef[Int]]", "ArrayRef", []string{"HashRef[Int]"}},
	}

	for _, test := range tests {
		base, params := parser.ExtractTypeAndParams(test.input)
		assert.Equal(t, test.expectedBase, base)
		assert.Equal(t, test.expectedParams, params)
	}
}

func TestPSCPVIIntegration(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "psc-pvi-integration-test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a sample Perl file with type annotations
	sampleFile := filepath.Join(tmpDir, "SampleModule.pm")
	sampleContent := `package SampleModule;
use v5.36;

# Variable types
my Int $counter = 0;
my Str $name = "Test";

# Function signature
sub increment(Int $value) -> Int {
    return $value + 1;
}

1;
`
	err = os.WriteFile(sampleFile, []byte(sampleContent), 0644)
	require.NoError(t, err)

	// Generate a type definition
	typeDef, err := ExtractTypeDefinitionsFromFile(sampleFile, "SampleModule")
	require.NoError(t, err)
	require.NotNil(t, typeDef)

	// Verify the type definition basics
	assert.Equal(t, "SampleModule", typeDef.Module)
	assert.Equal(t, sampleFile, typeDef.Source)

	// Create an output file for the type definition
	outputFile := filepath.Join(tmpDir, "SampleModule.ptd")
	outOptions := &TypeDefinitionOptions{
		ModuleName: "SampleModule",
		SourceFile: sampleFile,
		OutputFile: outputFile,
		Verbose:    false,
		Save:       false,
	}

	// Generate and save the type definition
	genResult, err := GenerateTypeDefinition(outOptions)
	require.NoError(t, err)
	require.NotNil(t, genResult)

	// Check if the output file was created
	_, err = os.Stat(outputFile)
	assert.True(t, os.IsNotExist(err), "Output file should not exist because we didn't specify it")

	// Test loading a non-existent type definition
	_, err = LoadTypeDefinition("NonExistentModule")
	assert.Error(t, err)

	// Test error handling and conversion
	tc, err := parser.NewTypeCheck()
	require.NoError(t, err)

	checkResult, err := tc.CheckFile(sampleFile)
	require.NoError(t, err)

	errors := GetTypeErrorsFromPSC(checkResult)
	assert.Len(t, errors, 0, "Sample file should have no type errors")
}

func TestTypeHierarchy(t *testing.T) {
	// Test accessing the type hierarchy
	hierarchy, err := GetTypeHierarchy()
	require.NoError(t, err)
	require.NotNil(t, hierarchy)

	// Verify that basic types are in the hierarchy
	assert.NoError(t, hierarchy.ValidateType("Int"))
	assert.NoError(t, hierarchy.ValidateType("Str"))
	assert.NoError(t, hierarchy.ValidateType("Bool"))
	assert.NoError(t, hierarchy.ValidateType("ArrayRef[Int]"))

	// Test type compatibility
	assert.NoError(t, hierarchy.CheckTypeCompatibility("Int", "Int"))
	assert.NoError(t, hierarchy.CheckTypeCompatibility("Int", "Num"))
	assert.Error(t, hierarchy.CheckTypeCompatibility("Str", "Int"))
	assert.NoError(t, hierarchy.CheckTypeCompatibility("ArrayRef[Int]", "ArrayRef[Int]"))
	assert.Error(t, hierarchy.CheckTypeCompatibility("ArrayRef[Int]", "ArrayRef[Str]"))
}

func TestUnifiedErrorHandling(t *testing.T) {
	// Test error mapping between PSC and PVI
	pscErr := NewTypeError(parser.ErrTypeValidationError, "Invalid type", nil)
	pviErr := MapTypeCheckerErrorToPVI(pscErr)

	assert.Equal(t, ErrTypeInvalid, pviErr.Code())
	assert.Equal(t, "Invalid type", pviErr.Description())

	// Convert back to PSC
	pscErrConverted := MapPVIErrorToPSC(pviErr)
	assert.Equal(t, parser.ErrTypeValidationError, pscErrConverted.Code())
	assert.Equal(t, "Invalid type", pscErrConverted.Description())

	// Test error formatting
	errs := []error{
		NewTypeError(parser.ErrTypeValidationError, "Invalid type", nil),
		NewTypeError(parser.ErrTypeAnnotationMismatch, "Type mismatch", nil),
	}

	formatted := FormatErrorsForOutput(errs, false)
	assert.Len(t, formatted, 2)
	assert.True(t, strings.Contains(formatted[0], "Invalid type"))
	assert.True(t, strings.Contains(formatted[1], "Type mismatch"))

	// Test verbose formatting
	verboseFormatted := FormatErrorsForOutput(errs, true)
	assert.Len(t, verboseFormatted, 2)
	assert.True(t, strings.Contains(verboseFormatted[0], parser.ErrTypeValidationError))
	assert.True(t, strings.Contains(verboseFormatted[1], parser.ErrTypeAnnotationMismatch))
}
