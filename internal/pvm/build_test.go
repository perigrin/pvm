// ABOUTME: Tests for the unified build command functionality
// ABOUTME: Validates build orchestration with different modes and configurations

package pvm

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/project"
)

func TestNewBuildCommand(t *testing.T) {
	cmd := NewBuildCommand()

	if cmd.Use != "build [target]" {
		t.Errorf("Expected Use to be 'build [target]', got %s", cmd.Use)
	}

	if cmd.Short != "Build Perl projects with type checking and compilation" {
		t.Errorf("Expected Short description about Perl projects, got %s", cmd.Short)
	}

	// Check that command has the expected flags
	expectedFlags := []string{
		"check-only", "inline", "watch", "clean",
		"output", "mode", "strict", "skip-typecheck",
		"skip-metadata", "include-tests", "include-scripts",
	}

	for _, flag := range expectedFlags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag %s to be present", flag)
		}
	}

	// Check that the check subcommand is present
	checkCmd := findSubcommand(cmd, "check")
	if checkCmd == nil {
		t.Error("Expected 'check' subcommand to be present in build command")
	}
}

func TestNewBuildCheckCommand(t *testing.T) {
	cmd := newBuildCheckCommand()

	if cmd.Use != "check [files...]" {
		t.Errorf("Expected Use to be 'check [files...]', got %s", cmd.Use)
	}

	if cmd.Short != "Type check Perl files with type annotations" {
		t.Errorf("Expected Short description about type checking, got %s", cmd.Short)
	}

	// Check that command has the expected flags
	expectedFlags := []string{"perl", "verbose"}
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

// Helper function to find subcommands
func findSubcommand(parent *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}

func TestParseBuildOptions(t *testing.T) {
	tests := []struct {
		name        string
		flags       map[string]interface{}
		projectCtx  *project.ProjectContext
		expectError bool
	}{
		{
			name: "default options",
			flags: map[string]interface{}{
				"check-only":      false,
				"inline":          false,
				"watch":           false,
				"clean":           false,
				"output":          "",
				"mode":            "",
				"strict":          false,
				"skip-typecheck":  false,
				"skip-metadata":   false,
				"include-tests":   true,
				"include-scripts": true,
			},
			projectCtx: &project.ProjectContext{
				IsProject: false,
				RootDir:   ".",
			},
			expectError: false,
		},
		{
			name: "inline mode",
			flags: map[string]interface{}{
				"check-only":      false,
				"inline":          true,
				"watch":           false,
				"clean":           false,
				"output":          "",
				"mode":            "inline",
				"strict":          false,
				"skip-typecheck":  false,
				"skip-metadata":   false,
				"include-tests":   true,
				"include-scripts": true,
			},
			projectCtx: &project.ProjectContext{
				IsProject: true,
				RootDir:   ".",
			},
			expectError: false,
		},
		{
			name: "watch mode",
			flags: map[string]interface{}{
				"check-only":      false,
				"inline":          false,
				"watch":           true,
				"clean":           false,
				"output":          "/tmp/build",
				"mode":            "distribution",
				"strict":          true,
				"skip-typecheck":  false,
				"skip-metadata":   false,
				"include-tests":   true,
				"include-scripts": true,
			},
			projectCtx: &project.ProjectContext{
				IsProject: true,
				RootDir:   ".",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test command with flags
			cmd := &cobra.Command{}
			for flag, value := range tt.flags {
				switch v := value.(type) {
				case bool:
					cmd.Flags().Bool(flag, v, "")
				case string:
					cmd.Flags().String(flag, v, "")
				}
			}

			// Parse the flags
			cmd.ParseFlags([]string{})

			options, err := parseBuildOptions(cmd, tt.projectCtx)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err == nil {
				// Verify basic options parsing
				if options.CheckOnly != tt.flags["check-only"] {
					t.Errorf("CheckOnly mismatch: expected %v, got %v", tt.flags["check-only"], options.CheckOnly)
				}
				if options.Inline != tt.flags["inline"] {
					t.Errorf("Inline mismatch: expected %v, got %v", tt.flags["inline"], options.Inline)
				}
				if options.Watch != tt.flags["watch"] {
					t.Errorf("Watch mismatch: expected %v, got %v", tt.flags["watch"], options.Watch)
				}

				// Verify project root is set correctly
				if options.ProjectRoot != tt.projectCtx.RootDir {
					t.Errorf("ProjectRoot mismatch: expected %s, got %s", tt.projectCtx.RootDir, options.ProjectRoot)
				}

				// Verify source directories are set correctly based on project context
				if tt.projectCtx.IsProject {
					expectedLibDir := filepath.Join(tt.projectCtx.RootDir, "lib")
					found := false
					for _, dir := range options.SourceDirs {
						if dir == expectedLibDir {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected lib directory %s not found in SourceDirs: %v", expectedLibDir, options.SourceDirs)
					}
				}
			}
		})
	}
}

func TestBuildOptionsStructure(t *testing.T) {
	// Test that BuildOptions has all expected fields
	options := &BuildOptions{
		CheckOnly:      true,
		Inline:         true,
		Watch:          true,
		Clean:          true,
		Mode:           "distribution",
		OutputDir:      "/tmp/build",
		Strict:         true,
		SkipTypeCheck:  false,
		SkipMetadata:   false,
		IncludeTests:   true,
		IncludeScripts: true,
		ProjectRoot:    "/tmp/project",
		SourceDirs:     []string{"lib", "script", "t"},
	}

	// Basic structural verification
	if !options.CheckOnly {
		t.Error("CheckOnly field not accessible")
	}
	if !options.Inline {
		t.Error("Inline field not accessible")
	}
	if !options.Watch {
		t.Error("Watch field not accessible")
	}
	if !options.Clean {
		t.Error("Clean field not accessible")
	}
	if options.Mode != "distribution" {
		t.Error("Mode field not accessible")
	}
	if options.OutputDir != "/tmp/build" {
		t.Error("OutputDir field not accessible")
	}
	if len(options.SourceDirs) != 3 {
		t.Error("SourceDirs field not accessible or wrong length")
	}
}

func TestBuildCommandIntegration(t *testing.T) {
	// This test verifies that the build command can be created and flags can be parsed
	// without actually executing builds (which would require a real project setup)

	// Test setting various flag combinations
	testCases := [][]string{
		{"--check-only"},
		{"--inline"},
		{"--watch"},
		{"--clean"},
		{"--mode", "distribution"},
		{"--mode", "inline"},
		{"--mode", "both"},
		{"--output", "/tmp/test-build"},
		{"--strict"},
		{"--skip-typecheck"},
		{"--skip-metadata"},
		{"--include-tests=false"},
		{"--include-scripts=false"},
		{"--inline", "--clean"},
		{"--mode", "distribution", "--strict", "--output", "/tmp/build"},
	}

	for i, args := range testCases {
		t.Run(fmt.Sprintf("flags_case_%d", i), func(t *testing.T) {
			// Create a fresh command for each test
			testCmd := NewBuildCommand()

			err := testCmd.ParseFlags(args)
			if err != nil {
				t.Errorf("Failed to parse flags %v: %v", args, err)
			}

			// Verify that the flags were parsed successfully by checking a few
			if testCmd.Flags().Changed("check-only") {
				checkOnly, _ := testCmd.Flags().GetBool("check-only")
				if !checkOnly {
					t.Error("check-only flag should be true when set")
				}
			}

			if testCmd.Flags().Changed("mode") {
				mode, _ := testCmd.Flags().GetString("mode")
				if mode == "" {
					t.Error("mode flag should not be empty when set")
				}
			}
		})
	}
}

func TestBuildCommandHelp(t *testing.T) {
	cmd := NewBuildCommand()

	// Test that help command works
	helpCmd := &cobra.Command{Use: "help"}
	helpCmd.AddCommand(cmd)

	// Verify that help text contains expected information
	helpOutput := cmd.Long
	expectedPhrases := []string{
		"Build Perl projects",
		"type checking",
		"compilation",
		"Distribution build",
		"Inline build",
		"Watch mode",
		"pvm build",
	}

	for _, phrase := range expectedPhrases {
		if !contains(helpOutput, phrase) {
			t.Errorf("Help text should contain '%s'", phrase)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			indexOfSubstring(s, substr) >= 0)))
}

func indexOfSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
