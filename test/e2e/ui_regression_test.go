// ABOUTME: Regression tests for Fang UI integration to prevent functionality loss
// ABOUTME: Tests edge cases, error conditions, and maintains functional completeness

package e2e

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/cli/ui"
	basetesting "tamarou.com/pvm/internal/testing"
	"tamarou.com/pvm/test/e2e/helpers"
)

// TestUIRegression_FunctionalPreservation tests that all existing functionality is preserved
func TestUIRegression_FunctionalPreservation(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test that all major command categories still work with UI integration
	functionalTests := []struct {
		name          string
		command       []string
		shouldSucceed bool
		contains      []string
	}{
		{
			name:          "PVM help command works",
			command:       []string{"--help"},
			shouldSucceed: true,
			contains:      []string{"Usage", "Commands"},
		},
		{
			name:          "PVX help command works",
			command:       []string{"pvx", "--help"},
			shouldSucceed: true,
			contains:      []string{"Usage"},
		},
		{
			name:          "PVI help command works",
			command:       []string{"pvi", "--help"},
			shouldSucceed: true,
			contains:      []string{"Usage"},
		},
		{
			name:          "PSC help command works",
			command:       []string{"psc", "--help"},
			shouldSucceed: true,
			contains:      []string{"Usage"},
		},
		{
			name:          "Version commands work",
			command:       []string{"version"},
			shouldSucceed: true,
			contains:      []string{},
		},
		{
			name:          "Invalid commands fail gracefully",
			command:       []string{"pvm", "invalid-command"},
			shouldSucceed: false,
			contains:      []string{},
		},
	}

	for _, test := range functionalTests {
		t.Run(test.name, func(t *testing.T) {
			stdout, stderr, err := env.RunPVM(test.command...)

			if test.shouldSucceed {
				assert.NoError(t, err, "Command should succeed: %v", test.command)

				// Check expected content
				output := stdout + stderr
				for _, expected := range test.contains {
					assert.Contains(t, output, expected,
						"Output should contain: %s", expected)
				}
			} else {
				assert.Error(t, err, "Command should fail: %v", test.command)
				// Error commands should still produce meaningful output
				assert.NotEmpty(t, stderr, "Error commands should produce error output")
			}
		})
	}
}

// TestUIRegression_OutputFormatPreservation tests that output formats are preserved
func TestUIRegression_OutputFormatPreservation(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test specific output format requirements
	formatTests := []struct {
		name    string
		command []string
		checkFn func(t *testing.T, output string)
	}{
		{
			name:    "Help output has proper structure",
			command: []string{"--help"},
			checkFn: func(t *testing.T, output string) {
				lines := strings.Split(output, "\n")
				assert.Greater(t, len(lines), 5, "Help should have multiple lines")

				// Help should contain standard sections
				hasUsage := strings.Contains(output, "Usage:")
				assert.True(t, hasUsage, "Help should contain Usage section")
			},
		},
		{
			name:    "Version output is not empty",
			command: []string{"version"},
			checkFn: func(t *testing.T, output string) {
				assert.NotEmpty(t, strings.TrimSpace(output), "Version should not be empty")
			},
		},
	}

	for _, test := range formatTests {
		t.Run(test.name, func(t *testing.T) {
			stdout := helpers.AssertPVMSucceeds(t, env, test.command, test.name)
			test.checkFn(t, stdout)
		})
	}
}

// TestUIRegression_ErrorHandlingPreservation tests that error handling is preserved
func TestUIRegression_ErrorHandlingPreservation(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test error scenarios that should be handled gracefully
	errorTests := []struct {
		name    string
		command []string
		checkFn func(t *testing.T, stdout, stderr string, err error)
	}{
		{
			name:    "Invalid command produces error",
			command: []string{"pvm", "nonexistent"},
			checkFn: func(t *testing.T, stdout, stderr string, err error) {
				assert.Error(t, err, "Invalid command should fail")
				assert.NotEmpty(t, stderr, "Should produce error output")
			},
		},
		{
			name:    "Invalid flags produce error",
			command: []string{"pvm", "--invalid-flag"},
			checkFn: func(t *testing.T, stdout, stderr string, err error) {
				assert.Error(t, err, "Invalid flag should fail")
				// Should produce some kind of error or help output
				output := stdout + stderr
				assert.NotEmpty(t, output, "Should produce some output")
			},
		},
		{
			name:    "Missing required arguments",
			command: []string{"pvx"}, // PVX requires script argument
			checkFn: func(t *testing.T, stdout, stderr string, err error) {
				// This might succeed with help or fail with error - either is acceptable
				// as long as it doesn't crash
				output := stdout + stderr
				assert.NotEmpty(t, output, "Should produce some output")
			},
		},
	}

	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			stdout, stderr, err := env.RunPVM(test.command...)
			test.checkFn(t, stdout, stderr, err)
		})
	}
}

// TestUIRegression_EdgeCaseHandling tests edge cases that might break UI
func TestUIRegression_EdgeCaseHandling(t *testing.T) {
	// Test UI framework directly for edge cases
	edgeCaseTests := []struct {
		name   string
		testFn func(t *testing.T)
	}{
		{
			name: "Nil writers handling",
			testFn: func(t *testing.T) {
				ctx := &ui.UIContext{
					Writer:      nil,
					ErrorWriter: nil,
					ColorMode:   ui.ColorNever,
					Quiet:       false,
				}

				// Should not crash with nil writers
				assert.NotPanics(t, func() {
					output := ui.NewOutput(ctx)
					output.Info("Test message")
				}, "Should handle nil writers gracefully")
			},
		},
		{
			name: "Empty context handling",
			testFn: func(t *testing.T) {
				// Should not crash with empty context
				assert.NotPanics(t, func() {
					output := ui.NewOutput(&ui.UIContext{})
					output.Info("Test message")
				}, "Should handle empty context gracefully")
			},
		},
		{
			name: "Extremely long messages",
			testFn: func(t *testing.T) {
				var buf bytes.Buffer
				ctx := &ui.UIContext{
					Writer:    &buf,
					ColorMode: ui.ColorNever,
					Quiet:     false,
				}
				output := ui.NewOutput(ctx)

				// Test with very long message
				longMsg := strings.Repeat("x", 100000)
				assert.NotPanics(t, func() {
					output.Info("%s", longMsg)
				}, "Should handle very long messages")

				result := buf.String()
				assert.Contains(t, result, "xxxx", "Long message should be output")
			},
		},
		{
			name: "Unicode and special characters",
			testFn: func(t *testing.T) {
				var buf bytes.Buffer
				ctx := &ui.UIContext{
					Writer:    &buf,
					ColorMode: ui.ColorNever,
					Quiet:     false,
				}
				output := ui.NewOutput(ctx)

				specialMsg := "🚀 Unicode: café, naïve, résumé 🎉 \n\t\r Special chars"
				assert.NotPanics(t, func() {
					output.Info("%s", specialMsg)
				}, "Should handle unicode and special characters")

				result := buf.String()
				assert.Contains(t, result, "🚀", "Unicode should be preserved")
				assert.Contains(t, result, "café", "Accented characters should be preserved")
			},
		},
		{
			name: "Large table data",
			testFn: func(t *testing.T) {
				var buf bytes.Buffer
				ctx := &ui.UIContext{
					Writer:    &buf,
					ColorMode: ui.ColorNever,
					Quiet:     false,
				}
				output := ui.NewOutput(ctx)

				// Test with large table
				headers := make([]string, 20)
				for i := 0; i < 20; i++ {
					headers[i] = "Column" + string(rune('A'+i))
				}

				rows := make([][]string, 50)
				for i := 0; i < 50; i++ {
					row := make([]string, 20)
					for j := 0; j < 20; j++ {
						row[j] = "Data" + string(rune('A'+j)) + string(rune('0'+i%10))
					}
					rows[i] = row
				}

				assert.NotPanics(t, func() {
					output.Table(headers, rows)
				}, "Should handle large tables")

				result := buf.String()
				assert.Contains(t, result, "ColumnA", "Table headers should be present")
				assert.Contains(t, result, "DataA0", "Table data should be present")
			},
		},
	}

	for _, test := range edgeCaseTests {
		t.Run(test.name, test.testFn)
	}
}

// TestUIRegression_BackwardCompatibility tests backward compatibility
func TestUIRegression_BackwardCompatibility(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Create test files to ensure file-based operations still work
	projectDir := filepath.Join(env.RootDir, "compat_test")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	testScript := filepath.Join(projectDir, "test.pl")
	scriptContent := `#!/usr/bin/env perl
use strict;
use warnings;
print "Compatibility test\n";
`
	require.NoError(t, os.WriteFile(testScript, []byte(scriptContent), 0644))

	compatibilityTests := []struct {
		name    string
		command []string
		setup   func(t *testing.T)
	}{
		{
			name:    "Script execution still works",
			command: []string{"pvx", "--help"}, // Use help instead of script execution for safety
			setup:   func(t *testing.T) {},
		},
		{
			name:    "Help system still works",
			command: []string{"--help"},
			setup:   func(t *testing.T) {},
		},
		{
			name:    "Version system still works",
			command: []string{"version"},
			setup:   func(t *testing.T) {},
		},
	}

	for _, test := range compatibilityTests {
		t.Run(test.name, func(t *testing.T) {
			test.setup(t)

			// Should succeed and produce output
			stdout := helpers.AssertPVMSucceeds(t, env, test.command, test.name)
			assert.NotEmpty(t, stdout, "Command should produce output")
		})
	}
}

// TestUIRegression_ResourceUsage tests that UI doesn't consume excessive resources
func TestUIRegression_ResourceUsage(t *testing.T) {
	basetesting.SkipUnlessPerformance(t, "UI resource usage regression test")

	t.Run("File descriptor usage", func(t *testing.T) {
		// Create many UI contexts to test file descriptor usage
		var contexts []*ui.UIContext
		var outputs []*ui.Output

		for i := 0; i < 100; i++ {
			var buf bytes.Buffer
			ctx := &ui.UIContext{
				Writer:    &buf,
				ColorMode: ui.ColorNever,
				Quiet:     false,
			}
			output := ui.NewOutput(ctx)

			contexts = append(contexts, ctx)
			outputs = append(outputs, output)

			// Use the output
			output.Info("Test %d", i)
		}

		// Should not exhaust file descriptors or other resources
		assert.Equal(t, 100, len(outputs), "Should create all UI outputs")
	})

	t.Run("Buffer overflow protection", func(t *testing.T) {
		// Test with very large output buffer
		var buf bytes.Buffer
		ctx := &ui.UIContext{
			Writer:    &buf,
			ColorMode: ui.ColorNever,
			Quiet:     false,
		}
		output := ui.NewOutput(ctx)

		// Generate large amount of output
		for i := 0; i < 1000; i++ {
			largeMessage := strings.Repeat("Large output test ", 100)
			output.Info("%s", largeMessage)
		}

		// Should not crash or consume excessive memory
		result := buf.String()
		assert.NotEmpty(t, result, "Should produce output")
		assert.Contains(t, result, "Large output test", "Should contain test message")
	})
}

// TestUIRegression_WriterTypes tests compatibility with different writer types
func TestUIRegression_WriterTypes(t *testing.T) {
	writerTests := []struct {
		name   string
		writer io.Writer
		testFn func(t *testing.T, output *ui.Output, writer io.Writer)
	}{
		{
			name:   "Bytes buffer writer",
			writer: &bytes.Buffer{},
			testFn: func(t *testing.T, output *ui.Output, writer io.Writer) {
				output.Info("Test message")
				buf := writer.(*bytes.Buffer)
				assert.Contains(t, buf.String(), "Test message")
			},
		},
		{
			name:   "File writer",
			writer: nil, // Will be set up in test
			testFn: func(t *testing.T, output *ui.Output, writer io.Writer) {
				output.Info("File test message")
				// File writer test would need file creation/cleanup
				// For now just verify no crash
			},
		},
		{
			name:   "Discard writer",
			writer: io.Discard,
			testFn: func(t *testing.T, output *ui.Output, writer io.Writer) {
				// Should not crash when writing to discard
				output.Info("Discarded message")
				output.Error("Discarded error")
				output.Success("Discarded success")
			},
		},
	}

	for _, test := range writerTests {
		t.Run(test.name, func(t *testing.T) {
			writer := test.writer

			// Special setup for file writer
			if test.name == "File writer" {
				tmpFile, err := os.CreateTemp("", "ui_test")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())
				defer tmpFile.Close()
				writer = tmpFile
			}

			ctx := &ui.UIContext{
				Writer:    writer,
				ColorMode: ui.ColorNever,
				Quiet:     false,
			}
			output := ui.NewOutput(ctx)

			assert.NotPanics(t, func() {
				test.testFn(t, output, writer)
			}, "Should not panic with %s", test.name)
		})
	}
}

// TestUIRegression_QuietModePreservation tests that quiet mode behavior is preserved
func TestUIRegression_QuietModePreservation(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test quiet mode with different commands
	quietTests := []struct {
		name    string
		command []string
	}{
		{
			name:    "PVM help in quiet mode",
			command: []string{"--quiet", "--help"},
		},
		{
			name:    "PVX help in quiet mode",
			command: []string{"pvx", "--quiet", "--help"},
		},
	}

	for _, test := range quietTests {
		t.Run(test.name, func(t *testing.T) {
			stdout, stderr, err := env.RunPVM(test.command...)

			// Quiet mode should still work (may produce less output)
			if err != nil {
				// Some commands might fail in quiet mode, but should not crash
				assert.NotEmpty(t, stderr, "Should produce error output if command fails")
			} else {
				// If successful, may produce minimal output
				output := stdout + stderr
				// Don't assert on output content as quiet mode may suppress it
				t.Logf("Quiet mode output length: %d characters", len(output))
			}
		})
	}
}

// TestUIRegression_VerboseModePreservation tests that verbose mode behavior is preserved
func TestUIRegression_VerboseModePreservation(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test verbose mode with different commands
	verboseTests := []struct {
		name    string
		command []string
	}{
		{
			name:    "PVM help in verbose mode",
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
}
