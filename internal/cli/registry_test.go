package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestCommandRegistry(t *testing.T) {
	registry := NewCommandRegistry()

	// Test empty registry
	_, ok := registry.Get("test")
	if ok {
		t.Error("Expected Get to return false for non-existent command")
	}

	// Test registration
	registry.Register("test", func() *cobra.Command {
		return &cobra.Command{Use: "test"}
	})

	// Test retrieval
	provider, ok := registry.Get("test")
	if !ok {
		t.Error("Expected Get to return true for registered command")
	}

	if provider == nil {
		t.Error("Expected provider to be non-nil")
	}

	// Test command creation
	cmd, ok := registry.CreateCommand("test")
	if !ok {
		t.Error("Expected CreateCommand to return true for registered command")
	}

	if cmd == nil {
		t.Error("Expected command to be non-nil")
	}

	if cmd.Use != "test" {
		t.Errorf("Expected command Use to be 'test', got %s", cmd.Use)
	}

	// Test non-existent command
	_, ok = registry.CreateCommand("nonexistent")
	if ok {
		t.Error("Expected CreateCommand to return false for non-existent command")
	}
}

func TestGlobalRegistry(t *testing.T) {
	// Ensure GlobalRegistry is initialized
	if GlobalRegistry == nil {
		t.Fatal("Expected GlobalRegistry to be initialized")
	}

	// Register a test command
	GlobalRegistry.Register("testglobal", func() *cobra.Command {
		return &cobra.Command{Use: "testglobal"}
	})

	// Test retrieval
	_, ok := GlobalRegistry.Get("testglobal")
	if !ok {
		t.Error("Expected Get to return true for registered command in GlobalRegistry")
	}
}
