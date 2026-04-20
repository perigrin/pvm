// ABOUTME: Tests for component detection and command routing in the cli package
// ABOUTME: Verifies that DetectComponent resolves the correct PVM/PVX/PM/PSC component from the invoked binary name

package cli

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestDetectComponent(t *testing.T) {
	// Save original args
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	testCases := []struct {
		name     string
		args     []string
		expected string
	}{
		{"PVM", []string{"/path/to/pvm"}, ComponentPVM},
		{"PVX", []string{"/path/to/pvx"}, ComponentPVX},
		{"PM", []string{"/path/to/pm"}, ComponentPM},
		{"PSC", []string{"/path/to/psc"}, ComponentPSC},
		{"PVM with Windows extension", []string{"/path/to/pvm.exe"}, ComponentPVM},
		{"Unknown", []string{"/path/to/unknown"}, ComponentPVM}, // Defaults to PVM
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Args = tc.args

			component := DetectComponent()
			if component != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, component)
			}
		})
	}
}

func TestGetComponentDescription(t *testing.T) {
	testCases := []struct {
		component string
		expected  string
	}{
		{ComponentPVM, DescriptionPVM},
		{ComponentPVX, DescriptionPVX},
		{ComponentPM, DescriptionPM},
		{ComponentPSC, DescriptionPSC},
		{"unknown", "Unknown component"},
	}

	for _, tc := range testCases {
		t.Run(tc.component, func(t *testing.T) {
			desc := GetComponentDescription(tc.component)
			if desc != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, desc)
			}
		})
	}
}

func TestCreateRootCommand(t *testing.T) {
	// Register a test command in the global registry
	GlobalRegistry.Register("test", func() *cobra.Command {
		cmd := &cobra.Command{Use: "test"}
		cmd.AddCommand(&cobra.Command{Use: "subcommand"})
		return cmd
	})

	rootCmd := CreateRootCommand("test")

	// Check that the root command has the correct name
	if rootCmd.Use != "test" {
		t.Errorf("Expected root command Use to be 'test', got %s", rootCmd.Use)
	}

	// Check that version information is added to the long description
	if rootCmd.Long == "" || rootCmd.Long == GetComponentDescription("test") {
		t.Error("Expected version information to be added to long description")
	}

	// Check that commands from the registry are added
	hasSubcommand := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "subcommand" {
			hasSubcommand = true
			break
		}
	}

	if !hasSubcommand {
		t.Error("Expected subcommand from registry to be added to root command")
	}
}
