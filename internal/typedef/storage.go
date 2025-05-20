// ABOUTME: Type definition storage and retrieval functionality
// ABOUTME: Manages saving and loading type definitions

package typedef

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/xdg"
)

// PSC Error codes
const (
	ErrTypeSaveError    = "701" // Error saving type definition
	ErrTypeLoadError    = "702" // Error loading type definition
	ErrTypeInvalidError = "703" // Error validating type definition
)

// Storage handles storing and retrieving type definitions
type Storage struct {
	// DirectoryPath is the path to the type definitions directory
	DirectoryPath string
}

// NewStorage creates a new Storage using the XDG directories
func NewStorage() (*Storage, error) {
	// Get the XDG directories
	dirs, err := xdg.GetDirs()
	if err != nil {
		return nil, err
	}

	// Ensure the type definitions directory exists
	if err := dirs.EnsureDirs(); err != nil {
		return nil, err
	}

	return &Storage{
		DirectoryPath: dirs.TypeDefinitionsDir,
	}, nil
}

// NewStorageWithPath creates a new Storage with a custom directory path
func NewStorageWithPath(dirPath string) (*Storage, error) {
	// Ensure the directory exists
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, errors.NewSystemError("001",
			"Failed to create type definitions directory", err).
			WithLocation(dirPath)
	}

	return &Storage{
		DirectoryPath: dirPath,
	}, nil
}

// Save saves a type definition to disk
func (s *Storage) Save(typeDef *TypeDefinition) error {
	// Validate the type definition
	if err := validateTypeDefinition(typeDef); err != nil {
		return errors.NewTypeError(
			ErrTypeInvalidError,
			"Type definition validation failed",
			err,
		)
	}

	// Create the filename from the module name
	filename := moduleToFilename(typeDef.Module)
	filePath := filepath.Join(s.DirectoryPath, filename)

	// Marshal the type definition to JSON
	data, err := json.MarshalIndent(typeDef, "", "  ")
	if err != nil {
		return errors.NewTypeError(
			ErrTypeSaveError,
			fmt.Sprintf("Failed to marshal type definition for %s", typeDef.Module),
			err,
		)
	}

	// Write the file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return errors.NewTypeError(
			ErrTypeSaveError,
			fmt.Sprintf("Failed to write type definition file for %s", typeDef.Module),
			err,
		).WithLocation(filePath)
	}

	return nil
}

// Load loads a type definition from disk
func (s *Storage) Load(moduleName string) (*TypeDefinition, error) {
	// Create the filename from the module name
	filename := moduleToFilename(moduleName)
	filePath := filepath.Join(s.DirectoryPath, filename)

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, errors.NewTypeError(
			ErrTypeLoadError,
			fmt.Sprintf("Type definition not found for module %s", moduleName),
			err,
		).WithLocation(filePath)
	}

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.NewTypeError(
			ErrTypeLoadError,
			fmt.Sprintf("Failed to read type definition file for %s", moduleName),
			err,
		).WithLocation(filePath)
	}

	// Unmarshal the type definition
	var typeDef TypeDefinition
	if err := json.Unmarshal(data, &typeDef); err != nil {
		return nil, errors.NewTypeError(
			ErrTypeLoadError,
			fmt.Sprintf("Failed to parse type definition for %s", moduleName),
			err,
		).WithLocation(filePath)
	}

	return &typeDef, nil
}

// List returns a list of all type definitions
func (s *Storage) List() ([]string, error) {
	// Read the directory
	files, err := os.ReadDir(s.DirectoryPath)
	if err != nil {
		return nil, errors.NewTypeError(
			ErrTypeLoadError,
			"Failed to read type definitions directory",
			err,
		).WithLocation(s.DirectoryPath)
	}

	// Extract module names from filenames
	var modules []string
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".ptd") {
			continue
		}

		// Convert filename back to module name
		moduleName := filenameToModule(file.Name())
		modules = append(modules, moduleName)
	}

	return modules, nil
}

// Delete removes a type definition
func (s *Storage) Delete(moduleName string) error {
	// Create the filename from the module name
	filename := moduleToFilename(moduleName)
	filePath := filepath.Join(s.DirectoryPath, filename)

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.NewTypeError(
			ErrTypeLoadError,
			fmt.Sprintf("Type definition not found for module %s", moduleName),
			err,
		).WithLocation(filePath)
	}

	// Remove the file
	if err := os.Remove(filePath); err != nil {
		return errors.NewTypeError(
			ErrTypeSaveError,
			fmt.Sprintf("Failed to delete type definition for %s", moduleName),
			err,
		).WithLocation(filePath)
	}

	return nil
}

// Helper functions

// moduleToFilename converts a module name to a filename
// e.g., "Path::Tiny" -> "Path-Tiny.ptd"
func moduleToFilename(moduleName string) string {
	// Replace :: with -
	filename := strings.ReplaceAll(moduleName, "::", "-")
	// Add .ptd extension
	return filename + ".ptd"
}

// filenameToModule converts a filename to a module name
// e.g., "Path-Tiny.ptd" -> "Path::Tiny"
func filenameToModule(filename string) string {
	// Remove .ptd extension
	moduleName := strings.TrimSuffix(filename, ".ptd")
	// Replace - with ::
	return strings.ReplaceAll(moduleName, "-", "::")
}

// validateTypeDefinition validates a type definition
func validateTypeDefinition(typeDef *TypeDefinition) error {
	// Basic validation
	if typeDef.Module == "" {
		return TypeDefError("Module name is required")
	}

	if typeDef.Version == "" {
		return TypeDefError("Module version is required")
	}

	// Zero time is fine for Generated, we can default it
	if typeDef.Generated.IsZero() {
		typeDef.Generated = time.Now()
	}

	// Source is required
	if typeDef.Source == "" {
		return TypeDefError("Source is required")
	}

	// Validate types, packages, subs, and methods would go here
	// For now, we'll keep it simple

	return nil
}
