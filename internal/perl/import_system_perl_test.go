// ABOUTME: Tests for error paths in ImportSystemPerl function
// ABOUTME: Provides comprehensive error path testing for ImportSystemPerl

package perl

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestImportSystemPerlErrorPaths tests all error paths in ImportSystemPerl
func TestImportSystemPerlErrorPaths(t *testing.T) {
	// Save original functions to restore later
	originalDetectSystemPerl := DetectSystemPerl
	originalRegisterVersion := registerVersion
	defer func() {
		DetectSystemPerl = originalDetectSystemPerl
		registerVersion = originalRegisterVersion
	}()

	// Test 1: DetectSystemPerl fails
	t.Run("DetectSystemPerlFails", func(t *testing.T) {
		// Mock DetectSystemPerl to fail
		DetectSystemPerl = func() (*SystemPerl, error) {
			return nil, errors.New("mock DetectSystemPerl error")
		}

		// Call function - should return error
		err := ImportSystemPerl()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock DetectSystemPerl error")
	})

	// Test 2: RegisterVersion fails
	t.Run("RegisterVersionFails", func(t *testing.T) {
		// Mock DetectSystemPerl to succeed
		DetectSystemPerl = func() (*SystemPerl, error) {
			return &SystemPerl{
				Path:    "/usr/bin/perl",
				Version: "5.38.0",
			}, nil
		}

		// Mock RegisterVersion to fail
		registerVersion = func(versionInfo VersionInfo) error {
			return errors.New("mock RegisterVersion error")
		}

		// Call function - should return error
		err := ImportSystemPerl()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock RegisterVersion error")
	})
}
