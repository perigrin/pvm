// ABOUTME: Tests for type definition functionality
// ABOUTME: Verifies TypeDefinition storage and validation

package typedef

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTypeDefinitionJSON tests marshaling and unmarshaling of TypeDefinition
func TestTypeDefinitionJSON(t *testing.T) {
	// Create a test type definition
	typeDef := &TypeDefinition{
		Module:     "Test::Module",
		Version:    "1.0.0",
		Generated:  time.Now().Round(time.Second), // Round to seconds for JSON comparison
		Maintainer: "Test User",
		Source:     "test",
		Types: []TypeInfo{
			{
				Name:        "TestType",
				Description: "A test type",
				Kind:        "class",
				Methods: []MethodInfo{
					{
						Name:        "test_method",
						Description: "A test method",
						Parameters: []ParamInfo{
							{
								Name:        "param1",
								Type:        "Str",
								Description: "A string parameter",
							},
						},
						Returns: []ReturnInfo{
							{
								Type:        "Bool",
								Description: "Success indicator",
							},
						},
					},
				},
				Properties: []PropInfo{
					{
						Name:        "test_property",
						Type:        "Int",
						Description: "A test property",
					},
				},
			},
		},
		Packages: []PackageInfo{
			{
				Name:        "Test::Package",
				Description: "A test package",
			},
		},
		Subs: []SubInfo{
			{
				Name:        "test_sub",
				Description: "A test subroutine",
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(typeDef)
	require.NoError(t, err, "Failed to marshal TypeDefinition")

	// Unmarshal from JSON
	var typeDefFromJSON TypeDefinition
	err = json.Unmarshal(data, &typeDefFromJSON)
	require.NoError(t, err, "Failed to unmarshal TypeDefinition")

	// Compare the original and unmarshaled type definitions
	assert.Equal(t, typeDef.Module, typeDefFromJSON.Module, "Module name should match")
	assert.Equal(t, typeDef.Version, typeDefFromJSON.Version, "Version should match")
	assert.Equal(t, typeDef.Generated.Unix(), typeDefFromJSON.Generated.Unix(), "Generated time should match")
	assert.Equal(t, typeDef.Maintainer, typeDefFromJSON.Maintainer, "Maintainer should match")
	assert.Equal(t, typeDef.Source, typeDefFromJSON.Source, "Source should match")
	assert.Equal(t, len(typeDef.Types), len(typeDefFromJSON.Types), "Types length should match")
	assert.Equal(t, len(typeDef.Packages), len(typeDefFromJSON.Packages), "Packages length should match")
	assert.Equal(t, len(typeDef.Subs), len(typeDefFromJSON.Subs), "Subs length should match")
}

// TestTypeStorage tests the Storage operations
func TestTypeStorage(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a storage with the temporary directory
	storage, err := NewStorageWithPath(tempDir)
	require.NoError(t, err, "Failed to create storage")

	// Verify the directory path
	assert.Equal(t, tempDir, storage.DirectoryPath, "Storage directory path should match")

	// Create a test type definition
	typeDef := &TypeDefinition{
		Module:     "Test::Storage",
		Version:    "1.0.0",
		Generated:  time.Now(),
		Maintainer: "Test User",
		Source:     "test",
	}

	// Save the type definition
	err = storage.Save(typeDef)
	require.NoError(t, err, "Failed to save type definition")

	// Verify the file was created
	expectedFilePath := filepath.Join(tempDir, "Test-Storage.ptd")
	_, err = os.Stat(expectedFilePath)
	assert.NoError(t, err, "Type definition file should exist")

	// Load the type definition
	loadedTypeDef, err := storage.Load("Test::Storage")
	require.NoError(t, err, "Failed to load type definition")

	// Compare the original and loaded type definitions
	assert.Equal(t, typeDef.Module, loadedTypeDef.Module, "Module name should match")
	assert.Equal(t, typeDef.Version, loadedTypeDef.Version, "Version should match")
	assert.Equal(t, typeDef.Generated.Unix(), loadedTypeDef.Generated.Unix(), "Generated time should match")
	assert.Equal(t, typeDef.Maintainer, loadedTypeDef.Maintainer, "Maintainer should match")
	assert.Equal(t, typeDef.Source, loadedTypeDef.Source, "Source should match")

	// List type definitions
	modules, err := storage.List()
	require.NoError(t, err, "Failed to list type definitions")
	assert.Len(t, modules, 1, "Should have one type definition")
	assert.Equal(t, "Test::Storage", modules[0], "Module name should match")

	// Delete the type definition
	err = storage.Delete("Test::Storage")
	require.NoError(t, err, "Failed to delete type definition")

	// Verify the file was deleted
	_, err = os.Stat(expectedFilePath)
	assert.True(t, os.IsNotExist(err), "Type definition file should not exist")

	// List type definitions again
	modules, err = storage.List()
	require.NoError(t, err, "Failed to list type definitions")
	assert.Len(t, modules, 0, "Should have no type definitions")
}

// TestValidation tests the type definition validation
func TestValidation(t *testing.T) {
	// Test with a valid type definition
	validTypeDef := &TypeDefinition{
		Module:     "Valid::Module",
		Version:    "1.0.0",
		Generated:  time.Now(),
		Maintainer: "Test User",
		Source:     "test",
	}
	err := validateTypeDefinition(validTypeDef)
	assert.NoError(t, err, "Valid type definition should pass validation")

	// Test with missing module name
	invalidTypeDef1 := &TypeDefinition{
		// Module:     "", // Missing
		Version:    "1.0.0",
		Generated:  time.Now(),
		Maintainer: "Test User",
		Source:     "test",
	}
	err = validateTypeDefinition(invalidTypeDef1)
	assert.Error(t, err, "Missing module name should fail validation")

	// Test with missing version
	invalidTypeDef2 := &TypeDefinition{
		Module: "Invalid::Module",
		// Version:    "", // Missing
		Generated:  time.Now(),
		Maintainer: "Test User",
		Source:     "test",
	}
	err = validateTypeDefinition(invalidTypeDef2)
	assert.Error(t, err, "Missing version should fail validation")

	// Test with missing source
	invalidTypeDef3 := &TypeDefinition{
		Module:     "Invalid::Module",
		Version:    "1.0.0",
		Generated:  time.Now(),
		Maintainer: "Test User",
		// Source:     "", // Missing
	}
	err = validateTypeDefinition(invalidTypeDef3)
	assert.Error(t, err, "Missing source should fail validation")

	// Test with missing generated time
	invalidTypeDef4 := &TypeDefinition{
		Module:  "Invalid::Module",
		Version: "1.0.0",
		// Generated:  time.Time{}, // Missing
		Maintainer: "Test User",
		Source:     "test",
	}
	err = validateTypeDefinition(invalidTypeDef4)
	assert.NoError(t, err, "Missing generated time should be auto-set")
	assert.False(t, invalidTypeDef4.Generated.IsZero(), "Generated time should be auto-set")
}

// TestFilenameConversion tests the module name to filename conversion
func TestFilenameConversion(t *testing.T) {
	testCases := []struct {
		moduleName string
		filename   string
	}{
		{"Test::Module", "Test-Module.ptd"},
		{"Path::Tiny", "Path-Tiny.ptd"},
		{"SingleName", "SingleName.ptd"},
		{"Multi::Part::Name", "Multi-Part-Name.ptd"},
	}

	for _, tc := range testCases {
		t.Run(tc.moduleName, func(t *testing.T) {
			// Test moduleToFilename
			filename := moduleToFilename(tc.moduleName)
			assert.Equal(t, tc.filename, filename, "Filename should match expected")

			// Test filenameToModule
			moduleName := filenameToModule(tc.filename)
			assert.Equal(t, tc.moduleName, moduleName, "Module name should match expected")
		})
	}
}
