package tree_sitter_typed_perl

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

var (
	// Library handle and function pointer cached for performance
	libHandle uintptr
	langFunc  func() uintptr
	initOnce  sync.Once
	initError error
)

// Get the tree-sitter Language for this grammar.
func Language() unsafe.Pointer {
	initOnce.Do(func() {
		initError = initializeLibrary()
	})

	if initError != nil {
		panic(fmt.Sprintf("failed to initialize tree-sitter library: %v", initError))
	}

	return unsafe.Pointer(langFunc())
}

// initializeLibrary loads the shared library and registers the function
func initializeLibrary() error {
	libPath, err := getLibraryPath()
	if err != nil {
		return fmt.Errorf("failed to find library: %w", err)
	}

	// Load the shared library (platform-specific)
	lib, err := loadLibrary(libPath)
	if err != nil {
		return fmt.Errorf("failed to load library %s: %w", libPath, err)
	}

	libHandle = lib

	// Register the tree_sitter_typed_perl function
	purego.RegisterLibFunc(&langFunc, lib, "tree_sitter_typed_perl")

	return nil
}

// loadLibrary handles platform-specific library loading
func loadLibrary(path string) (uintptr, error) {
	return loadLibraryPlatform(path)
}

// getLibraryPath determines the correct library path for the current platform
func getLibraryPath() (string, error) {
	var filename string
	switch runtime.GOOS {
	case "windows":
		filename = "tree-sitter-typed-perl.dll"
	case "darwin":
		filename = "libtree-sitter-typed-perl.dylib"
	case "linux":
		filename = "libtree-sitter-typed-perl.so"
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	// Get the current working directory to resolve absolute paths
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Search paths in order of preference (with absolute resolution)
	searchPaths := []string{
		// Try to find the project root and use absolute path
		findProjectRoot(cwd, filename),
		// Relative to bindings directory (development)
		filepath.Join("..", "..", filename),
		// Direct relative path
		filename,
		// Build output directory
		filepath.Join("build", filename),
	}

	// For development/testing, try the relative path first
	for _, path := range searchPaths {
		if path != "" && fileExists(path) {
			return path, nil
		}
	}

	// If not found, return the first non-empty path and let dlopen handle the error
	for _, path := range searchPaths {
		if path != "" {
			return path, nil
		}
	}

	return searchPaths[1], nil // fallback to relative path
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// findProjectRoot searches up the directory tree to find the project root
// and returns the path to the shared library if found
func findProjectRoot(startDir, filename string) string {
	dir := startDir
	for {
		// Check if this directory contains the tree-sitter-typed-perl subdirectory
		libPath := filepath.Join(dir, "tree-sitter-typed-perl", filename)
		if fileExists(libPath) {
			return libPath
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}
	return ""
}
