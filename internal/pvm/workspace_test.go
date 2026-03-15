// ABOUTME: Tests for workspace command functionality
// ABOUTME: Validates workspace management and initialization features

package pvm

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewWorkspaceCommand(t *testing.T) {
	cmd := newWorkspaceCommand()

	if cmd.Use != "workspace" {
		t.Errorf("Expected Use to be 'workspace', got %s", cmd.Use)
	}

	if cmd.Short != "Workspace management commands" {
		t.Errorf("Expected Short description about workspace management, got %s", cmd.Short)
	}

	// Check that the validate subcommand is present
	validateCmd := findSubcommandInWorkspace(cmd, "validate")
	if validateCmd == nil {
		t.Error("Expected 'validate' subcommand to be present in workspace command")
	}
}

func TestNewWorkspaceValidateCommand(t *testing.T) {
	cmd := newWorkspaceValidateCommand()

	if cmd.Use != "validate [script]" {
		t.Errorf("Expected Use to be 'validate [script]', got %s", cmd.Use)
	}

	if cmd.Short != "Validate complete workspace setup" {
		t.Errorf("Expected Short description about workspace validation, got %s", cmd.Short)
	}

	// Check that command has the expected flags
	expectedFlags := []string{"verbose"}
	for _, flag := range expectedFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %s to be present", flag)
		}
	}

	// Check that RunE function is set
	if cmd.RunE == nil {
		t.Error("Expected RunE function to be set")
	}
}

// Helper function to find subcommands in workspace
func findSubcommandInWorkspace(parent *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}
