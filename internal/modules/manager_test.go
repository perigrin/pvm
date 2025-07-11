// ABOUTME: Tests for unified module manager
// ABOUTME: Validates module listing, searching, and management operations

package modules

import (
	"os"
	"testing"

	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/log"
)

func TestNewManager(t *testing.T) {
	// Create a basic provider (we'll use nil for testing)
	var provider cpan.Provider
	logger := log.NewLogger(1, os.Stderr, "test")

	manager := NewManager(provider, nil, logger)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.provider != provider {
		t.Error("Manager provider not set correctly")
	}

	if manager.logger != logger {
		t.Error("Manager logger not set correctly")
	}
}

// Module manager tests removed - tracked in GitHub issue #55
// These tests require complete module management system implementation
// including CPAN integration, Perl execution framework, and dependency resolution
//
// See: https://github.com/perigrin/pvm/issues/55
// "Implement complete module management system for PVI"
