package cli

import (
	"strings"
	"testing"
)

func TestNewRootCommand(t *testing.T) {
	name := "test"
	desc := "Test command"
	cmd := NewRootCommand(name, desc)

	if cmd.Use != name {
		t.Errorf("Expected command name to be %s, got %s", name, cmd.Use)
	}

	if cmd.Short != desc {
		t.Errorf("Expected command description to be %s, got %s", desc, cmd.Short)
	}

	// Check that verbose flag is added
	flag := cmd.PersistentFlags().Lookup("verbose")
	if flag == nil {
		t.Error("Expected verbose flag to be added to root command")
	}

	// Check that version command is added
	var hasVersion bool
	for _, subcmd := range cmd.Commands() {
		if subcmd.Use == "version" {
			hasVersion = true
			break
		}
	}
	if !hasVersion {
		t.Error("Expected version command to be added to root command")
	}
}

func TestVersionCommand(t *testing.T) {
	// Just verify the command is created correctly
	verCmd := newVersionCommand("test")

	if verCmd.Use != "version" {
		t.Errorf("Expected command Use to be 'version', got %s", verCmd.Use)
	}

	if verCmd.RunE == nil {
		t.Error("Expected RunE function to be set")
	}
}

func TestHandleErrorFunc(t *testing.T) {
	// Just create a CLI error and test its format
	err := NewError(PrefixPVM, CategoryUserInput, "001", "Test error", nil)
	err = err.WithDetail("Some detail").WithHint("Some hint")

	// Check error type
	if err.Prefix() != PrefixPVM {
		t.Errorf("Expected prefix %s, got %s", PrefixPVM, err.Prefix())
	}

	if err.Code() != "001" {
		t.Errorf("Expected code 001, got %s", err.Code())
	}

	// Check error formatting
	errStr := err.Error()
	expected := "PVM-001: Test error"
	if !strings.Contains(errStr, expected) {
		t.Errorf("Expected error string to contain '%s', got: %s", expected, errStr)
	}
}
