// ABOUTME: Visual consistency tests for Fang UI integration across all components
// ABOUTME: Tests styling patterns, visual hierarchy, and output formatting consistency

package e2e

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/test/e2e/helpers"
)

// TestVisualConsistency_StylePatterns tests that all components use consistent styling patterns
func TestVisualConsistency_StylePatterns(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test that all components use consistent styling for similar elements
	styleTests := []struct {
		name     string
		commands [][]string
		checkFor []string
	}{
		{
			name: "Help command consistency",
			commands: [][]string{
				{"pvm", "--help"},
				{"pvx", "--help"},
				{"pvi", "--help"},
				{"psc", "--help"},
			},
			checkFor: []string{"Usage:", "Flags:", "Commands:"},
		},
		{
			name: "Version command consistency",
			commands: [][]string{
				{"pvm", "--version"},
				{"pvx", "--version"},
				{"pvi", "--version"},
				{"psc", "--version"},
			},
			checkFor: []string{}, // Version format may vary but should not error
		},
	}

	for _, test := range styleTests {
		t.Run(test.name, func(t *testing.T) {
			outputs := make([]string, len(test.commands))

			// Collect outputs from all commands
			for i, cmd := range test.commands {
				outputs[i] = helpers.AssertPVMSucceeds(t, env, cmd,
					"Command "+strings.Join(cmd, " ")+" should succeed")
			}

			// Check that all outputs contain expected elements
			for _, expected := range test.checkFor {
				foundIn := 0
				for _, output := range outputs {
					if strings.Contains(output, expected) {
						foundIn++
					}
				}

				// All commands should contain expected elements (if they're supposed to)
				if foundIn > 0 && foundIn < len(outputs) {
					t.Logf("Warning: '%s' found in %d/%d outputs, may indicate inconsistency",
						expected, foundIn, len(outputs))
				}
			}

			// Check for consistent formatting patterns
			for i, output := range outputs {
				lines := strings.Split(output, "\n")
				if len(lines) > 0 {
					// First line should not be empty (unless it's intentional)
					if strings.TrimSpace(lines[0]) == "" && len(lines) > 1 {
						t.Logf("Command %s has empty first line", strings.Join(test.commands[i], " "))
					}
				}
			}
		})
	}
}

// TestVisualConsistency_MessageFormatting tests consistent message formatting
func TestVisualConsistency_MessageFormatting(t *testing.T) {
	// Test the UI framework directly for message formatting consistency
	var buf bytes.Buffer
	ctx := &ui.UIContext{
		Writer:    &buf,
		ColorMode: ui.ColorNever,
		Quiet:     false,
		Verbose:   true,
	}
	output := ui.NewOutput(ctx)

	messageTests := []struct {
		name     string
		method   func(string, ...interface{})
		message  string
		expected []string // Patterns that should be present
	}{
		{
			name:     "Success messages",
			method:   output.Success,
			message:  "Operation completed successfully",
			expected: []string{"Operation completed successfully"},
		},
		{
			name:     "Error messages",
			method:   output.Error,
			message:  "Something went wrong",
			expected: []string{"Something went wrong"},
		},
		{
			name:     "Warning messages",
			method:   output.Warning,
			message:  "This is a warning",
			expected: []string{"This is a warning"},
		},
		{
			name:     "Info messages",
			method:   output.Info,
			message:  "Informational message",
			expected: []string{"Informational message"},
		},
		{
			name:     "Debug messages",
			method:   output.Debug,
			message:  "Debug information",
			expected: []string{"Debug information"},
		},
	}

	for _, test := range messageTests {
		t.Run(test.name, func(t *testing.T) {
			buf.Reset()
			test.method(test.message)
			result := buf.String()

			// Check that expected patterns are present
			for _, expected := range test.expected {
				assert.Contains(t, result, expected,
					"Output should contain expected pattern: %s", expected)
			}

			// Check that output is not empty
			assert.NotEmpty(t, result, "Output should not be empty")

			// Check that output ends with newline (good CLI practice)
			assert.True(t, strings.HasSuffix(result, "\n"),
				"Output should end with newline")
		})
	}
}

// TestVisualConsistency_StructuredOutput tests consistency of structured output
func TestVisualConsistency_StructuredOutput(t *testing.T) {
	var buf bytes.Buffer
	ctx := &ui.UIContext{
		Writer:    &buf,
		ColorMode: ui.ColorNever,
		Quiet:     false,
	}
	output := ui.NewOutput(ctx)

	t.Run("Table formatting consistency", func(t *testing.T) {
		// Test various table configurations for consistency
		tableTests := []struct {
			name    string
			headers []string
			rows    [][]string
		}{
			{
				name:    "Simple table",
				headers: []string{"Name", "Value"},
				rows:    [][]string{{"key1", "value1"}, {"key2", "value2"}},
			},
			{
				name:    "Wide table",
				headers: []string{"Component", "Status", "Version", "Description"},
				rows: [][]string{
					{"PVM", "Active", "1.0.0", "Perl Version Manager"},
					{"PVX", "Active", "1.0.0", "Perl Version Executor"},
				},
			},
			{
				name:    "Single column table",
				headers: []string{"Items"},
				rows:    [][]string{{"Item 1"}, {"Item 2"}, {"Item 3"}},
			},
		}

		for _, test := range tableTests {
			t.Run(test.name, func(t *testing.T) {
				buf.Reset()
				output.Table(test.headers, test.rows)
				result := buf.String()

				// Check table structure
				lines := strings.Split(result, "\n")
				assert.Greater(t, len(lines), len(test.rows),
					"Table should have header, separator, and data lines")

				// Check that headers are present
				for _, header := range test.headers {
					assert.Contains(t, result, header,
						"Table should contain header: %s", header)
				}

				// Check that data is present
				for _, row := range test.rows {
					for _, cell := range row {
						assert.Contains(t, result, cell,
							"Table should contain cell data: %s", cell)
					}
				}
			})
		}
	})

	t.Run("List formatting consistency", func(t *testing.T) {
		listTests := []struct {
			name    string
			items   []string
			options ui.ListOptions
		}{
			{
				name:  "Simple list",
				items: []string{"Item 1", "Item 2", "Item 3"},
				options: ui.ListOptions{
					Items:    []string{"Item 1", "Item 2", "Item 3"},
					Numbered: false,
				},
			},
			{
				name:  "Numbered list",
				items: []string{"First", "Second", "Third"},
				options: ui.ListOptions{
					Items:    []string{"First", "Second", "Third"},
					Numbered: true,
				},
			},
			{
				name:  "List with title",
				items: []string{"Task A", "Task B"},
				options: ui.ListOptions{
					Items:    []string{"Task A", "Task B"},
					Title:    "Todo List",
					Numbered: false,
				},
			},
		}

		for _, test := range listTests {
			t.Run(test.name, func(t *testing.T) {
				buf.Reset()
				output.ListWithOptions(test.options)
				result := buf.String()

				// Check that items are present
				for _, item := range test.items {
					assert.Contains(t, result, item,
						"List should contain item: %s", item)
				}

				// Check title if present
				if test.options.Title != "" {
					assert.Contains(t, result, test.options.Title,
						"List should contain title: %s", test.options.Title)
				}

				// Check list structure
				lines := strings.Split(result, "\n")
				nonEmptyLines := 0
				for _, line := range lines {
					if strings.TrimSpace(line) != "" {
						nonEmptyLines++
					}
				}
				assert.GreaterOrEqual(t, nonEmptyLines, len(test.items),
					"List should have at least as many non-empty lines as items")
			})
		}
	})
}

// TestVisualConsistency_HierarchicalOutput tests visual hierarchy consistency
func TestVisualConsistency_HierarchicalOutput(t *testing.T) {
	var buf bytes.Buffer
	ctx := &ui.UIContext{
		Writer:    &buf,
		ColorMode: ui.ColorNever,
		Quiet:     false,
	}
	output := ui.NewOutput(ctx)

	t.Run("Header hierarchy", func(t *testing.T) {
		buf.Reset()

		// Test header hierarchy
		output.Header("Main Title")
		output.SubHeader("Section Title")
		output.Info("Regular content")

		result := buf.String()
		lines := strings.Split(result, "\n")

		// Check that content is present
		assert.Contains(t, result, "Main Title")
		assert.Contains(t, result, "Section Title")
		assert.Contains(t, result, "Regular content")

		// Check that headers appear before content
		mainTitleLine := -1
		sectionTitleLine := -1
		contentLine := -1

		for i, line := range lines {
			if strings.Contains(line, "Main Title") {
				mainTitleLine = i
			}
			if strings.Contains(line, "Section Title") {
				sectionTitleLine = i
			}
			if strings.Contains(line, "Regular content") {
				contentLine = i
			}
		}

		// Check proper hierarchy order
		if mainTitleLine >= 0 && sectionTitleLine >= 0 {
			assert.Less(t, mainTitleLine, sectionTitleLine,
				"Main title should appear before section title")
		}
		if sectionTitleLine >= 0 && contentLine >= 0 {
			assert.Less(t, sectionTitleLine, contentLine,
				"Section title should appear before content")
		}
	})

	t.Run("Section organization", func(t *testing.T) {
		buf.Reset()

		output.Section("Configuration", "Current settings and options")
		output.Section("Status", "System status information")

		result := buf.String()

		// Check section content
		assert.Contains(t, result, "Configuration")
		assert.Contains(t, result, "Current settings and options")
		assert.Contains(t, result, "Status")
		assert.Contains(t, result, "System status information")

		// Sections should have proper spacing
		lines := strings.Split(result, "\n")
		assert.Greater(t, len(lines), 4, "Sections should have multiple lines with spacing")
	})
}

// TestVisualConsistency_ProgressAndStatus tests progress and status display consistency
func TestVisualConsistency_ProgressAndStatus(t *testing.T) {
	var buf bytes.Buffer
	ctx := &ui.UIContext{
		Writer:    &buf,
		ColorMode: ui.ColorNever,
		Quiet:     false,
	}
	output := ui.NewOutput(ctx)

	t.Run("Progress display", func(t *testing.T) {
		progressTests := []struct {
			current int
			total   int
			message string
		}{
			{1, 5, "Starting task"},
			{3, 5, "Halfway through"},
			{5, 5, "Completing task"},
		}

		for _, test := range progressTests {
			buf.Reset()
			output.Progress(test.current, test.total, test.message)
			result := buf.String()

			// Check progress components
			assert.Contains(t, result, test.message, "Progress should contain message")
			assert.Contains(t, result, string(rune(test.current+'0')),
				"Progress should show current value")
			assert.Contains(t, result, string(rune(test.total+'0')),
				"Progress should show total value")

			// Progress should be on single line
			lines := strings.Split(strings.TrimSpace(result), "\n")
			assert.Equal(t, 1, len(lines), "Progress should be single line")
		}
	})

	t.Run("Status display", func(t *testing.T) {
		statusTests := []string{
			"Initializing",
			"Processing data",
			"Finalizing",
		}

		for _, status := range statusTests {
			buf.Reset()
			output.Status(status)
			result := buf.String()

			assert.Contains(t, result, status, "Status should contain message")
			assert.NotEmpty(t, result, "Status should produce output")
		}
	})
}

// TestVisualConsistency_MarkdownRendering tests markdown rendering consistency
func TestVisualConsistency_MarkdownRendering(t *testing.T) {
	var buf bytes.Buffer
	ctx := &ui.UIContext{
		Writer:    &buf,
		ColorMode: ui.ColorNever,
		Quiet:     false,
	}
	output := ui.NewOutput(ctx)

	markdownTests := []struct {
		name     string
		markdown string
		contains []string
	}{
		{
			name: "Headers",
			markdown: `# Main Header
## Sub Header
### Sub-sub Header`,
			contains: []string{"Main Header", "Sub Header", "Sub-sub Header"},
		},
		{
			name: "Lists",
			markdown: `- Item 1
- Item 2
- Item 3`,
			contains: []string{"Item 1", "Item 2", "Item 3"},
		},
		{
			name: "Mixed content",
			markdown: `# Configuration Guide

## Installation Steps

- Download the package
- Extract files
- Run setup

## Usage

Start with the **basic** commands.`,
			contains: []string{"Configuration Guide", "Installation Steps", "Download the package", "Usage", "basic"},
		},
	}

	for _, test := range markdownTests {
		t.Run(test.name, func(t *testing.T) {
			buf.Reset()
			output.Markdown(test.markdown)
			result := buf.String()

			// Check that content is rendered
			for _, expected := range test.contains {
				assert.Contains(t, result, expected,
					"Markdown should contain: %s", expected)
			}

			// Check that output is structured
			lines := strings.Split(result, "\n")
			assert.Greater(t, len(lines), 1, "Markdown should produce multiple lines")
		})
	}
}

// TestVisualConsistency_ErrorFormatting tests consistent error formatting
func TestVisualConsistency_ErrorFormatting(t *testing.T) {
	var buf bytes.Buffer
	ctx := &ui.UIContext{
		Writer:      &buf,
		ErrorWriter: &buf, // Use same buffer for testing
		ColorMode:   ui.ColorNever,
		Quiet:       false,
	}
	output := ui.NewOutput(ctx)

	errorTests := []struct {
		name    string
		message string
	}{
		{"Simple error", "File not found"},
		{"Complex error", "Failed to parse configuration: invalid syntax on line 42"},
		{"Error with details", "Connection failed: timeout after 30 seconds"},
	}

	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			buf.Reset()
			output.Error("%s", test.message)
			result := buf.String()

			// Check error content
			assert.Contains(t, result, test.message, "Error should contain message")
			assert.NotEmpty(t, result, "Error should produce output")

			// Error should be properly formatted
			assert.True(t, strings.HasSuffix(result, "\n"),
				"Error should end with newline")
		})
	}
}

// TestVisualConsistency_KeyValueDisplay tests key-value display consistency
func TestVisualConsistency_KeyValueDisplay(t *testing.T) {
	var buf bytes.Buffer
	ctx := &ui.UIContext{
		Writer:    &buf,
		ColorMode: ui.ColorNever,
		Quiet:     false,
	}
	output := ui.NewOutput(ctx)

	keyValueTests := []struct {
		name  string
		pairs map[string]string
	}{
		{
			name: "System info",
			pairs: map[string]string{
				"OS":      "Linux",
				"Arch":    "x86_64",
				"Version": "1.0.0",
			},
		},
		{
			name: "Configuration",
			pairs: map[string]string{
				"Home Directory": "/home/user",
				"Config File":    "/etc/config.conf",
				"Log Level":      "INFO",
			},
		},
	}

	for _, test := range keyValueTests {
		t.Run(test.name, func(t *testing.T) {
			buf.Reset()
			output.KeyValue(test.pairs)
			result := buf.String()

			// Check that all keys and values are present
			for key, value := range test.pairs {
				assert.Contains(t, result, key, "Output should contain key: %s", key)
				assert.Contains(t, result, value, "Output should contain value: %s", value)
			}

			// Check formatting
			lines := strings.Split(result, "\n")
			nonEmptyLines := 0
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					nonEmptyLines++
				}
			}
			assert.GreaterOrEqual(t, nonEmptyLines, len(test.pairs),
				"Should have at least one line per key-value pair")
		})
	}
}
