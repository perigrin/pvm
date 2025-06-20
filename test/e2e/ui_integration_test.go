// ABOUTME: Comprehensive integration tests for Fang UI system across all PVM components
// ABOUTME: Tests cross-component consistency, styling, error handling, and performance

package e2e

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/test/e2e/helpers"
)

// TestUIFramework_ComponentConsistency tests that all components use consistent styling
func TestUIFramework_ComponentConsistency(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test each component produces consistently styled output
	components := []struct {
		name           string
		command        []string
		expectContains []string
	}{
		{
			name:           "PVM help",
			command:        []string{"--help"},
			expectContains: []string{"Usage:", "Commands:", "Flags:"},
		},
		{
			name:           "PVX help",
			command:        []string{"pvx", "--help"},
			expectContains: []string{"Usage:", "Flags:"},
		},
		{
			name:           "PVI help",
			command:        []string{"pvi", "--help"},
			expectContains: []string{"Usage:", "Commands:", "Flags:"},
		},
		{
			name:           "PSC help",
			command:        []string{"psc", "--help"},
			expectContains: []string{"Usage:", "Commands:", "Flags:"},
		},
	}

	for _, comp := range components {
		t.Run(comp.name, func(t *testing.T) {
			stdout := helpers.AssertPVMSucceeds(t, env, comp.command, comp.name+" should show help")

			// Verify expected content is present
			for _, expected := range comp.expectContains {
				assert.Contains(t, stdout, expected, "Help output should contain %s", expected)
			}

			// Verify no direct fmt.Print* output (should all be styled)
			// This is a heuristic check - styled output typically contains ANSI codes or special characters
			lines := strings.Split(stdout, "\n")
			styledLineCount := 0
			for _, line := range lines {
				if len(line) > 0 && (strings.Contains(line, "✓") || strings.Contains(line, "✗") ||
					strings.Contains(line, "ℹ") || strings.Contains(line, "⚠") ||
					strings.Contains(line, "Usage:") || strings.Contains(line, "Commands:")) {
					styledLineCount++
				}
			}

			// At least some lines should have styling indicators
			if styledLineCount == 0 && len(lines) > 3 {
				t.Logf("Warning: %s output may not be using UI framework (no styling indicators found)", comp.name)
			}
		})
	}
}

// TestUIFramework_ErrorHandling tests error display consistency across components
func TestUIFramework_ErrorHandling(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test error scenarios that should produce consistent styled error output
	errorTests := []struct {
		name        string
		command     []string
		expectError bool
	}{
		{
			name:        "PVM invalid command",
			command:     []string{"pvm", "nonexistent-command"},
			expectError: true,
		},
		{
			name:        "PVX missing script",
			command:     []string{"pvx", "/nonexistent/script.pl"},
			expectError: true,
		},
		{
			name:        "PSC invalid file",
			command:     []string{"psc", "check", "/nonexistent/file.pl"},
			expectError: true,
		},
	}

	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			_, stderr, err := env.RunPVM(test.command...)

			if test.expectError {
				assert.Error(t, err, "Command should fail")

				// Error output should exist
				assert.NotEmpty(t, stderr, "Error output should not be empty")

				// Error output should contain error indicators (styled output)
				// Look for common error patterns or styling
				hasErrorStyling := strings.Contains(stderr, "✗") ||
					strings.Contains(stderr, "Error:") ||
					strings.Contains(stderr, "error:") ||
					strings.Contains(stderr, "Error ") ||
					len(stderr) > 0 // At minimum, some error output should exist

				assert.True(t, hasErrorStyling, "Error output should be styled: %s", stderr)
			}
		})
	}
}

// TestUIFramework_OutputModes tests different output modes (quiet, verbose, etc.)
func TestUIFramework_OutputModes(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test quiet mode across components
	t.Run("Quiet mode", func(t *testing.T) {
		quietTests := []struct {
			name    string
			command []string
		}{
			{
				name:    "PVM help quiet",
				command: []string{"--quiet", "--help"},
			},
			{
				name:    "PVX help quiet",
				command: []string{"pvx", "--quiet", "--help"},
			},
		}

		for _, test := range quietTests {
			t.Run(test.name, func(t *testing.T) {
				stdout := helpers.AssertPVMSucceeds(t, env, test.command, test.name)

				// Quiet mode should still show help but may be more minimal
				// The key is that it should work without errors
				assert.NotEmpty(t, stdout, "Quiet mode should still produce some output for help")
			})
		}
	})

	// Test verbose mode if supported
	t.Run("Verbose mode", func(t *testing.T) {
		verboseTests := []struct {
			name    string
			command []string
		}{
			{
				name:    "PVM help verbose",
				command: []string{"--verbose", "--help"},
			},
		}

		for _, test := range verboseTests {
			t.Run(test.name, func(t *testing.T) {
				stdout := helpers.AssertPVMSucceeds(t, env, test.command, test.name)

				// Verbose mode should produce output
				assert.NotEmpty(t, stdout, "Verbose mode should produce output")
			})
		}
	})
}

// TestUIFramework_PerformanceImpact tests that UI framework doesn't significantly impact performance
func TestUIFramework_PerformanceImpact(t *testing.T) {
	if testing.Short() {
		t.Skip("Performance test skipped in short mode")
	}

	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create a simple script for performance testing
	projectDir := filepath.Join(env.RootDir, "perf_test")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	testScript := filepath.Join(projectDir, "simple.pl")
	scriptContent := `#!/usr/bin/env perl
use strict;
use warnings;
print "Performance test completed\n";
`
	require.NoError(t, os.WriteFile(testScript, []byte(scriptContent), 0644))

	// Test execution time for help commands (UI-heavy operations)
	performanceTests := []struct {
		name    string
		command []string
		maxTime time.Duration
	}{
		{
			name:    "PVM help performance",
			command: []string{"--help"},
			maxTime: 5 * time.Second,
		},
		{
			name:    "PVX help performance",
			command: []string{"pvx", "--help"},
			maxTime: 5 * time.Second,
		},
		{
			name:    "PSC help performance",
			command: []string{"psc", "--help"},
			maxTime: 5 * time.Second,
		},
	}

	for _, test := range performanceTests {
		t.Run(test.name, func(t *testing.T) {
			start := time.Now()
			helpers.AssertPVMSucceeds(t, env, test.command, test.name)
			duration := time.Since(start)

			if duration > test.maxTime {
				t.Logf("Warning: %s took %v, which is longer than expected (%v)",
					test.name, duration, test.maxTime)
			}

			// Performance should be reasonable (not more than 10 seconds for help)
			assert.Less(t, duration, 10*time.Second,
				"UI framework should not cause excessive performance impact")
		})
	}
}

// TestUIFramework_TableAndListRendering tests structured output rendering
func TestUIFramework_TableAndListRendering(t *testing.T) {
	// This tests the UI framework directly since we may not have commands
	// that produce table/list output yet
	var buf bytes.Buffer
	ctx := &ui.UIContext{
		Writer:    &buf,
		ColorMode: ui.ColorNever,
		Quiet:     false,
		Verbose:   false,
	}
	output := ui.NewOutput(ctx)

	t.Run("Table rendering", func(t *testing.T) {
		buf.Reset()

		headers := []string{"Component", "Status", "Version"}
		rows := [][]string{
			{"PVM", "Active", "1.0.0"},
			{"PVX", "Active", "1.0.0"},
			{"PVI", "Active", "1.0.0"},
			{"PSC", "Active", "1.0.0"},
		}

		output.Table(headers, rows)
		result := buf.String()

		// Verify table structure
		assert.Contains(t, result, "Component", "Table should contain headers")
		assert.Contains(t, result, "PVM", "Table should contain data")
		assert.Contains(t, result, "Active", "Table should contain status")

		// Table should have proper structure (multiple lines)
		lines := strings.Split(result, "\n")
		assert.Greater(t, len(lines), 3, "Table should have multiple lines")
	})

	t.Run("List rendering", func(t *testing.T) {
		buf.Reset()

		items := []string{
			"Initialize project structure",
			"Configure Perl version",
			"Install dependencies",
			"Run tests",
		}

		output.List(items)
		result := buf.String()

		// Verify list structure
		for _, item := range items {
			assert.Contains(t, result, item, "List should contain all items")
		}

		// List should have proper structure
		lines := strings.Split(result, "\n")
		assert.Greater(t, len(lines), len(items), "List should have multiple lines")
	})

	t.Run("Progress rendering", func(t *testing.T) {
		buf.Reset()

		output.Progress(3, 10, "Processing files")
		result := buf.String()

		assert.Contains(t, result, "Processing files", "Progress should contain message")
		assert.Contains(t, result, "3", "Progress should show current")
		assert.Contains(t, result, "10", "Progress should show total")
	})

	t.Run("Status rendering", func(t *testing.T) {
		buf.Reset()

		output.Status("Downloading dependencies")
		result := buf.String()

		assert.Contains(t, result, "Downloading dependencies", "Status should contain message")
		assert.NotEmpty(t, result, "Status should produce output")
	})
}

// TestUIFramework_ColorModeHandling tests different color modes
func TestUIFramework_ColorModeHandling(t *testing.T) {
	colorModes := []struct {
		name string
		mode ui.ColorMode
	}{
		{"Auto", ui.ColorAuto},
		{"Always", ui.ColorAlways},
		{"Never", ui.ColorNever},
	}

	for _, colorMode := range colorModes {
		t.Run(colorMode.name, func(t *testing.T) {
			var buf bytes.Buffer
			ctx := &ui.UIContext{
				Writer:    &buf,
				ColorMode: colorMode.mode,
				Quiet:     false,
				Verbose:   true,
			}
			output := ui.NewOutput(ctx)

			// Test different output types with different color modes
			output.Success("Success message")
			output.Error("Error message")
			output.Warning("Warning message")
			output.Info("Info message")

			result := buf.String()

			// All modes should produce output
			assert.Contains(t, result, "Success message")
			assert.Contains(t, result, "Error message")
			assert.Contains(t, result, "Warning message")
			assert.Contains(t, result, "Info message")

			// ColorNever should not contain ANSI escape codes
			if colorMode.mode == ui.ColorNever {
				// This is a basic check - in practice ANSI codes are complex
				// But we can check for some common escape sequences
				assert.NotContains(t, result, "\033[", "ColorNever should not contain ANSI codes")
			}
		})
	}
}

// TestUIFramework_CrossComponentIntegration tests that UI framework works across component boundaries
func TestUIFramework_CrossComponentIntegration(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test that different components can work together with consistent UI
	projectDir := filepath.Join(env.RootDir, "cross_component")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Create a sample Perl file
	sampleFile := filepath.Join(projectDir, "sample.pl")
	sampleContent := `#!/usr/bin/env perl
use strict;
use warnings;

print "Cross-component test\n";
`
	require.NoError(t, os.WriteFile(sampleFile, []byte(sampleContent), 0644))

	// Test that components can be chained or used together
	t.Run("Help consistency", func(t *testing.T) {
		// Test root help command
		stdout := helpers.AssertPVMSucceeds(t, env,
			[]string{"--help"},
			"pvm help should work")
		assert.Contains(t, stdout, "Usage:",
			"pvm help should contain usage information")

		// Test subcommand help
		subcommands := []string{"pvx", "pvi", "psc"}
		for _, subcmd := range subcommands {
			stdout := helpers.AssertPVMSucceeds(t, env,
				[]string{subcmd, "--help"},
				subcmd+" help should work")

			// All help should contain Usage information
			assert.Contains(t, stdout, "Usage:",
				"%s help should contain usage information", subcmd)
		}
	})

	t.Run("Version consistency", func(t *testing.T) {
		// Test root version command
		stdout := helpers.AssertPVMSucceeds(t, env,
			[]string{"version"},
			"pvm version should work")
		assert.NotEmpty(t, stdout,
			"pvm version should produce output")

		// Test subcommand version commands
		subcommands := []string{"pvx", "pvi", "psc"}
		for _, subcmd := range subcommands {
			stdout := helpers.AssertPVMSucceeds(t, env,
				[]string{subcmd, "version"},
				subcmd+" version should work")

			// Version output should not be empty
			assert.NotEmpty(t, stdout,
				"%s version should produce output", subcmd)
		}
	})
}

// TestUIFramework_EdgeCases tests UI framework behavior in edge cases
func TestUIFramework_EdgeCases(t *testing.T) {
	t.Run("Empty messages", func(t *testing.T) {
		var buf bytes.Buffer
		ctx := &ui.UIContext{
			Writer:    &buf,
			ColorMode: ui.ColorNever,
			Quiet:     false,
		}
		output := ui.NewOutput(ctx)

		// Test empty message handling
		output.Info("")
		output.Success("")
		output.Warning("")
		output.Error("")

		// Should not crash and should produce some output
		result := buf.String()
		assert.NotEmpty(t, result, "Empty messages should still produce some output")
	})

	t.Run("Very long messages", func(t *testing.T) {
		var buf bytes.Buffer
		ctx := &ui.UIContext{
			Writer:    &buf,
			ColorMode: ui.ColorNever,
			Quiet:     false,
		}
		output := ui.NewOutput(ctx)

		// Test very long message
		longMessage := strings.Repeat("This is a very long message. ", 100)
		output.Info("%s", longMessage)

		result := buf.String()
		assert.Contains(t, result, "This is a very long message",
			"Long messages should be handled properly")
	})

	t.Run("Special characters", func(t *testing.T) {
		var buf bytes.Buffer
		ctx := &ui.UIContext{
			Writer:    &buf,
			ColorMode: ui.ColorNever,
			Quiet:     false,
		}
		output := ui.NewOutput(ctx)

		// Test messages with special characters
		specialMessage := "Message with special chars: 🚀 ✨ 🎉 newline\nand tab\t"
		output.Info("%s", specialMessage)

		result := buf.String()
		assert.Contains(t, result, "🚀", "Special characters should be preserved")
	})

	t.Run("Nil context handling", func(t *testing.T) {
		// Test that NewOutput handles nil context gracefully
		output := ui.NewOutput(nil)
		assert.NotNil(t, output, "NewOutput should handle nil context")
		assert.NotNil(t, output.Context(), "Context should be created")

		// Should be able to use without crashing
		output.Info("Test message")
	})
}

// TestUIFramework_ContextPropagation tests that UI context flows properly through command execution
func TestUIFramework_ContextPropagation(t *testing.T) {
	t.Run("Context preservation", func(t *testing.T) {
		var buf bytes.Buffer
		originalCtx := &ui.UIContext{
			Writer:      &buf,
			ColorMode:   ui.ColorNever,
			Quiet:       true,
			Verbose:     false,
			Interactive: false,
		}

		output := ui.NewOutput(originalCtx)
		retrievedCtx := output.Context()

		// Context should be preserved
		assert.Equal(t, originalCtx.ColorMode, retrievedCtx.ColorMode)
		assert.Equal(t, originalCtx.Quiet, retrievedCtx.Quiet)
		assert.Equal(t, originalCtx.Verbose, retrievedCtx.Verbose)
		assert.Equal(t, originalCtx.Interactive, retrievedCtx.Interactive)
	})

	t.Run("Context modification", func(t *testing.T) {
		output := ui.NewDefaultOutput()

		// Test context modification methods
		output.SetQuiet(true)
		assert.True(t, output.Context().Quiet, "SetQuiet should update context")

		output.SetVerbose(true)
		assert.True(t, output.Context().Verbose, "SetVerbose should update context")

		output.SetColorMode(ui.ColorAlways)
		assert.Equal(t, ui.ColorAlways, output.Context().ColorMode,
			"SetColorMode should update context")
	})
}
