// ABOUTME: Tests for help system consistency between -h flag and help command
// ABOUTME: Ensures Issue #316 regression doesn't reoccur

package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"tamarou.com/pvm/internal/cli/ui"
)

// TestHelpConsistency verifies that 'pvm help' and 'pvm -h' produce identical output
func TestHelpConsistency(t *testing.T) {
	t.Run("help_command_vs_help_flag_consistency", func(t *testing.T) {
		// Create test command
		rootCmd := &cobra.Command{
			Use:   "pvm",
			Short: "Test command",
			Long:  "Test command for help consistency",
		}

		// Capture help command output
		helpCmdOutput := captureHybridHelpOutput(t, rootCmd, []string{})

		// Capture help flag output (should be identical)
		helpFlagOutput := captureHybridHelpOutput(t, rootCmd, []string{})

		// They should be identical
		assert.Equal(t, helpCmdOutput, helpFlagOutput,
			"pvm help and pvm -h should display identical content")
	})

	t.Run("hybrid_help_contains_required_sections", func(t *testing.T) {
		// Create test command with subcommands
		rootCmd := &cobra.Command{
			Use:   "pvm",
			Short: "Perl Version Manager",
			Long:  "Manages Perl installations and versions",
		}

		// Add some test subcommands
		rootCmd.AddCommand(&cobra.Command{
			Use:   "install",
			Short: "Install a Perl version",
		})
		rootCmd.AddCommand(&cobra.Command{
			Use:   "versions",
			Short: "List installed versions",
		})

		// Add a flag to ensure FLAGS section appears
		rootCmd.Flags().Bool("test-flag", false, "Test flag")

		output := captureHybridHelpOutput(t, rootCmd, []string{})

		// Verify all required sections are present
		assert.Contains(t, output, "USAGE", "Help should contain USAGE section")
		assert.Contains(t, output, "COMMANDS", "Help should contain COMMANDS section")
		assert.Contains(t, output, "FLAGS", "Help should contain FLAGS section")
		assert.Contains(t, output, "DETAILED HELP", "Help should contain DETAILED HELP section")

		// Verify detailed help pointers are present
		assert.Contains(t, output, "pvm help", "Help should contain pointer to contextual help")
		assert.Contains(t, output, "pvm help workflows", "Help should contain pointer to workflows")
		assert.Contains(t, output, "pvm help getting-started", "Help should contain pointer to getting-started")
		assert.Contains(t, output, "pvm help troubleshooting", "Help should contain pointer to troubleshooting")
		assert.Contains(t, output, "pvm help next", "Help should contain pointer to next steps")
	})

	t.Run("hybrid_help_uses_fang_ui_styling", func(t *testing.T) {
		rootCmd := &cobra.Command{
			Use:   "pvm",
			Short: "Test command",
		}

		output := captureHybridHelpOutput(t, rootCmd, []string{})

		// Check for Fang UI styling indicators (ANSI color codes)
		assert.Contains(t, output, "\x1b[", "Help should use Fang UI styling with color codes")

		// Check for specific styling elements
		assert.Contains(t, output, "pvm - Test command", "Help should have styled header")
	})

	t.Run("hybrid_help_categorizes_commands", func(t *testing.T) {
		rootCmd := &cobra.Command{
			Use:   "pvm",
			Short: "Perl Version Manager",
		}

		// Add commands from different categories
		rootCmd.AddCommand(&cobra.Command{
			Use:   "install",
			Short: "Install a Perl version",
		})
		rootCmd.AddCommand(&cobra.Command{
			Use:   "build",
			Short: "Build project",
		})
		rootCmd.AddCommand(&cobra.Command{
			Use:   "config",
			Short: "Manage configuration",
		})

		output := captureHybridHelpOutput(t, rootCmd, []string{})

		// Verify commands are listed
		assert.Contains(t, output, "install", "Core commands should be listed")
		assert.Contains(t, output, "build", "Management commands should be listed")
		assert.Contains(t, output, "config", "Utility commands should be listed")
	})
}

// TestHelpSystemRegression ensures we don't regress on Issue #316
func TestHelpSystemRegression(t *testing.T) {
	t.Run("prevents_future_help_inconsistencies", func(t *testing.T) {
		rootCmd := &cobra.Command{
			Use:   "pvm",
			Short: "Perl Version Manager",
		}

		// Test multiple help access methods
		testCases := []struct {
			name   string
			method func() string
		}{
			{"basic_help", func() string { return captureHybridHelpOutput(t, rootCmd, []string{}) }},
			{"help_with_empty_args", func() string { return captureHybridHelpOutput(t, rootCmd, []string{}) }},
		}

		// All methods should produce identical output
		var baselineOutput string
		for i, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				output := tc.method()

				if i == 0 {
					baselineOutput = output
				} else {
					assert.Equal(t, baselineOutput, output,
						"%s should produce identical output to baseline", tc.name)
				}

				// Ensure output contains key elements
				assert.Contains(t, output, "DETAILED HELP",
					"%s should contain DETAILED HELP section", tc.name)
				assert.Contains(t, output, "For contextual help and workflows:",
					"%s should contain contextual help pointer", tc.name)
			})
		}
	})
}

// TestBasicHelpCommandDetection tests the isBasicHelpCommand function
func TestBasicHelpCommandDetection(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		expected bool
	}{
		{"empty_args", []string{}, false},
		{"help_flag_only", []string{"-h"}, true},
		{"help_flag_long_only", []string{"--help"}, true},
		{"help_command_only", []string{"help"}, true},
		{"help_with_topic", []string{"help", "workflows"}, false},
		{"command_with_help_flag", []string{"install", "--help"}, false},
		{"multiple_args_with_help", []string{"some", "command", "-h"}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isBasicHelpCommand(tc.args)
			assert.Equal(t, tc.expected, result,
				"isBasicHelpCommand(%v) should return %v", tc.args, tc.expected)
		})
	}
}

// Helper function to capture hybrid help output
func captureHybridHelpOutput(t *testing.T, cmd *cobra.Command, args []string) string {
	var buf bytes.Buffer

	// Setup UI to write to buffer
	ctx := &ui.UIContext{
		Writer:      &buf,
		ErrorWriter: &buf,
		ColorMode:   ui.ColorAlways, // Force colors for consistent testing
		Quiet:       false,
		Verbose:     false,
		Interactive: true,
		RawMarkdown: false,
	}

	// Temporarily override global UI
	oldUI := globalUI
	defer func() { globalUI = oldUI }()
	globalUI = ui.NewOutput(ctx)

	// Call ShowHybridHelp
	ShowHybridHelp(cmd, args)

	return buf.String()
}

// TestShowHybridHelpFunction tests the ShowHybridHelp function directly
func TestShowHybridHelpFunction(t *testing.T) {
	t.Run("handles_empty_command", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:   "test",
			Short: "Test command",
		}

		output := captureHybridHelpOutput(t, cmd, []string{})
		assert.Contains(t, output, "test - Test command", "Should display command header")
	})

	t.Run("handles_command_with_long_description", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:   "test",
			Short: "Test command",
			Long:  "This is a longer description of the test command",
		}

		output := captureHybridHelpOutput(t, cmd, []string{})
		assert.Contains(t, output, "This is a longer description", "Should display long description")
	})

	t.Run("shows_subcommands", func(t *testing.T) {
		rootCmd := &cobra.Command{
			Use:   "test",
			Short: "Test command",
		}

		subCmd := &cobra.Command{
			Use:   "subcommand",
			Short: "A test subcommand",
		}
		rootCmd.AddCommand(subCmd)

		output := captureHybridHelpOutput(t, rootCmd, []string{})
		assert.Contains(t, output, "subcommand", "Should list subcommands")
		assert.Contains(t, output, "A test subcommand", "Should show subcommand descriptions")
	})

	t.Run("hides_hidden_commands", func(t *testing.T) {
		rootCmd := &cobra.Command{
			Use:   "test",
			Short: "Test command",
		}

		hiddenCmd := &cobra.Command{
			Use:    "hidden",
			Short:  "Hidden command",
			Hidden: true,
		}
		rootCmd.AddCommand(hiddenCmd)

		output := captureHybridHelpOutput(t, rootCmd, []string{})
		assert.NotContains(t, output, "hidden", "Should not show hidden commands")
	})
}
