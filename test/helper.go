// ABOUTME: Contains helper functions for PVM testing
// ABOUTME: Provides common test utilities used across test suites

package test

import (
	"os"
	"path/filepath"
)

// GetProjectRoot returns the absolute path to the project root directory
func GetProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// If we're in the test directory, go up one level
	if filepath.Base(wd) == "test" {
		return filepath.Dir(wd), nil
	}

	return wd, nil
}

// TestHelper provides common testing utilities
type TestHelper struct {
	ProjectRoot string
}

// NewTestHelper creates a new test helper
func NewTestHelper() (*TestHelper, error) {
	root, err := GetProjectRoot()
	if err != nil {
		return nil, err
	}

	return &TestHelper{
		ProjectRoot: root,
	}, nil
}
