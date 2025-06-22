// ABOUTME: Quality assurance tests to prevent ugly or broken UI output
// ABOUTME: Validates clean formatting, proper styling, and prevents UI regressions

package e2e

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/test/e2e/helpers"
)

// TestUIQuality_NoUglyOutput tests that UI output is clean and well-formatted
func TestUIQuality_NoUglyOutput(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test various commands to ensure clean, well-formatted output
	qualityTests := []struct {
		name     string
		command  []string
		validate func(t *testing.T, stdout, stderr string)
	}{
		{
			name:    "PVX help has clean formatting",
			command: []string{"pvx", "--help"},
			validate: func(t *testing.T, stdout, stderr string) {
				validateCleanOutput(t, stdout, "PVX help")
				validateNoUglyProgressOutput(t, stdout, "PVX help")
			},
		},
		{
			name:    "PVI help has clean formatting",
			command: []string{"pvi", "--help"},
			validate: func(t *testing.T, stdout, stderr string) {
				validateCleanOutput(t, stdout, "PVI help")
				validateNoUglyProgressOutput(t, stdout, "PVI help")
			},
		},
		{
			name:    "PSC help has clean formatting",
			command: []string{"psc", "--help"},
			validate: func(t *testing.T, stdout, stderr string) {
				validateCleanOutput(t, stdout, "PSC help")
				validateNoUglyProgressOutput(t, stdout, "PSC help")
			},
		},
		{
			name:    "Error messages are clean",
			command: []string{"pvx", "/nonexistent/file.pl"},
			validate: func(t *testing.T, stdout, stderr string) {
				// Errors should still be clean and readable
				validateCleanErrorOutput(t, stderr, "PVX error")
				validateNoUglyProgressOutput(t, stdout+stderr, "PVX error output")
			},
		},
	}

	for _, test := range qualityTests {
		t.Run(test.name, func(t *testing.T) {
			stdout, stderr, _ := env.RunPVM(test.command...)
			test.validate(t, stdout, stderr)
		})
	}
}

// validateNoUglyProgressOutput specifically checks for the ugly progress output pattern
func validateNoUglyProgressOutput(t *testing.T, output, context string) {
	lines := strings.Split(output, "\n")

	// Test 1: No excessive repetitive progress bars with file paths
	progressBarCount := 0
	percentageCount := 0
	longPathCount := 0

	for _, line := range lines {
		cleanLine := stripANSICodes(line)

		// Check for progress bar patterns like [==========>          ]
		if strings.Contains(cleanLine, "[=") && strings.Contains(cleanLine, ">") && strings.Contains(cleanLine, "]") {
			progressBarCount++
		}

		// Check for percentage patterns like "50.0%"
		if strings.Contains(cleanLine, "%") && (strings.Contains(cleanLine, ".0%") || regexp.MustCompile(`\d+\.\d+%`).MatchString(cleanLine)) {
			percentageCount++
		}

		// Check for very long file paths (like the perl pod files)
		if len(cleanLine) > 100 && (strings.Contains(cleanLine, ".pod") || strings.Contains(cleanLine, "/lib/") || strings.Contains(cleanLine, "/perl5/")) {
			longPathCount++
		}
	}

	// Allow some progress indication, but not excessive repetition
	if progressBarCount > 10 {
		t.Errorf("%s contains excessive progress bars (%d), looks like ugly build output", context, progressBarCount)
	}

	if percentageCount > 15 {
		t.Errorf("%s contains excessive percentage indicators (%d), looks like repetitive progress output", context, percentageCount)
	}

	if longPathCount > 5 {
		t.Errorf("%s contains many long file paths (%d), looks like verbose build logs", context, longPathCount)
	}

	// Test 2: No repetitive identical lines (sign of stuck progress output)
	lineMap := make(map[string]int)
	for _, line := range lines {
		cleanLine := strings.TrimSpace(stripANSICodes(line))
		if len(cleanLine) > 20 { // Only check substantial lines
			lineMap[cleanLine]++
		}
	}

	for line, count := range lineMap {
		if count > 5 {
			t.Errorf("%s has repetitive line appearing %d times (first 50 chars): %.50s", context, count, line)
			break
		}
	}

	// Test 3: No build/install artifacts in user-facing output
	uglyPatterns := []string{
		"DYLD_LIBRARY_PATH=",
		"installman --destdir=",
		"ERROR: Manual page installation",
		"./perl -Ilib -I.",
		"/usr/local/bin/perl",
		"lib/perl5/5.",
		"pods/perl",
	}

	for _, pattern := range uglyPatterns {
		if strings.Contains(output, pattern) {
			t.Errorf("%s contains build/install artifacts: %s", context, pattern)
		}
	}
}

// validateCleanOutput checks for common UI problems
func validateCleanOutput(t *testing.T, output, context string) {
	// Test 1: No raw ANSI escape sequences visible in text (but allow actual ANSI codes)
	assert.NotContains(t, output, "\\033[", "%s should not contain visible ANSI escape sequences", context)
	assert.NotContains(t, output, "\\x1b[", "%s should not contain visible ANSI escape sequences", context)

	// Test 2: Allow ANSI color codes (they're expected), but warn if they seem malformed
	// Note: [38;2; and [m are actually valid ANSI codes, so we should be more careful here
	// Only flag them if they appear to be escaped or malformed
	if strings.Contains(output, "\\[38;2;") || strings.Contains(output, "\\[m") {
		t.Errorf("%s contains escaped ANSI codes that should be processed", context)
	}

	// Test 3: No excessively long lines (> 200 chars without breaks)
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		// Remove ANSI codes for length calculation
		cleanLine := stripANSICodes(line)
		if len(cleanLine) > 200 {
			t.Errorf("%s line %d is too long (%d chars): %.50s...", context, i+1, len(cleanLine), cleanLine)
		}
	}

	// Test 4: No excessive blank lines (more than 3 consecutive)
	consecutiveBlankLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			consecutiveBlankLines++
			if consecutiveBlankLines > 3 {
				t.Errorf("%s has more than 3 consecutive blank lines", context)
				break
			}
		} else {
			consecutiveBlankLines = 0
		}
	}

	// Test 5: Proper line endings (no mixed line endings)
	assert.NotContains(t, output, "\r\n", "%s should not contain Windows line endings", context)

	// Test 6: No control characters except newlines and tabs
	for i, r := range output {
		if r < 32 && r != '\n' && r != '\t' && r != 27 { // Allow newline, tab, and ESC (for ANSI)
			t.Errorf("%s contains control character %d at position %d", context, int(r), i)
			break
		}
	}

	// Test 7: Content should exist (not just whitespace)
	cleanContent := stripANSICodes(output)
	cleanContent = strings.TrimSpace(cleanContent)
	assert.NotEmpty(t, cleanContent, "%s should contain actual content", context)
}

// validateCleanErrorOutput checks error output for quality
func validateCleanErrorOutput(t *testing.T, output, context string) {
	if output == "" {
		return // No error output is fine
	}

	// Basic clean output checks
	validateCleanOutput(t, output, context)

	// Error-specific checks
	// Test 1: Should contain proper error indicators
	cleanOutput := stripANSICodes(output)
	hasErrorIndicator := strings.Contains(cleanOutput, "Error") ||
		strings.Contains(cleanOutput, "error") ||
		strings.Contains(cleanOutput, "✗") ||
		strings.Contains(cleanOutput, "failed")
	assert.True(t, hasErrorIndicator, "%s should contain proper error indicators", context)

	// Test 2: Should not be excessively verbose (< 1000 chars for simple errors)
	if len(cleanOutput) > 1000 {
		t.Logf("Warning: %s is quite verbose (%d chars)", context, len(cleanOutput))
	}
}

// TestUIQuality_ProperStyling tests that Fang styling is applied correctly
func TestUIQuality_ProperStyling(t *testing.T) {
	var buf bytes.Buffer
	ctx := &ui.UIContext{
		Writer:    &buf,
		ColorMode: ui.ColorAlways, // Force colors for testing
		Quiet:     false,
		Verbose:   true,
	}
	output := ui.NewOutput(ctx)

	stylingTests := []struct {
		name     string
		action   func()
		validate func(t *testing.T, result string)
	}{
		{
			name: "Success messages have proper styling",
			action: func() {
				output.Success("Operation completed successfully")
			},
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "Operation completed successfully", "Should contain message")
				// Should have some styling (ANSI codes or icons)
				hasStyle := strings.Contains(result, "✓") || containsANSICodes(result)
				assert.True(t, hasStyle, "Success message should be styled")
			},
		},
		{
			name: "Error messages have proper styling",
			action: func() {
				output.Error("Something went wrong")
			},
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "Something went wrong", "Should contain message")
				// Should have error styling
				hasStyle := strings.Contains(result, "✗") || containsANSICodes(result)
				assert.True(t, hasStyle, "Error message should be styled")
			},
		},
		{
			name: "Tables have proper structure",
			action: func() {
				headers := []string{"Component", "Status"}
				rows := [][]string{
					{"PVM", "Active"},
					{"PVX", "Active"},
				}
				output.Table(headers, rows)
			},
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "Component", "Should contain headers")
				assert.Contains(t, result, "PVM", "Should contain data")

				// Should have table structure
				lines := strings.Split(result, "\n")
				nonEmptyLines := 0
				for _, line := range lines {
					if strings.TrimSpace(line) != "" {
						nonEmptyLines++
					}
				}
				assert.GreaterOrEqual(t, nonEmptyLines, 3, "Table should have header, separator, and data lines")
			},
		},
		{
			name: "Lists have proper formatting",
			action: func() {
				items := []string{"First item", "Second item", "Third item"}
				output.List(items)
			},
			validate: func(t *testing.T, result string) {
				for _, item := range []string{"First item", "Second item", "Third item"} {
					assert.Contains(t, result, item, "Should contain list item")
				}

				// Should have list structure with bullets or numbers
				hasBullets := strings.Contains(result, "•") ||
					strings.Contains(result, "1.") ||
					strings.Contains(result, "-")
				assert.True(t, hasBullets, "List should have proper bullet points or numbers")
			},
		},
	}

	for _, test := range stylingTests {
		t.Run(test.name, func(t *testing.T) {
			buf.Reset()
			test.action()
			result := buf.String()
			test.validate(t, result)
		})
	}
}

// TestUIQuality_NoRegressionPatterns tests against specific known bad patterns
func TestUIQuality_NoRegressionPatterns(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test commands to ensure they don't produce known bad patterns
	regressionTests := []struct {
		name        string
		command     []string
		badPatterns []string // Patterns that should NOT appear in output
	}{
		{
			name:    "Help output doesn't contain debug info",
			command: []string{"pvx", "--help"},
			badPatterns: []string{
				"DEBUG:",
				"[DEBUG]",
				"panic:",
				"goroutine",
				"runtime.main",
				"stack trace",
			},
		},
		{
			name:    "Error output is clean",
			command: []string{"pvx", "/nonexistent.pl"},
			badPatterns: []string{
				"panic:",
				"goroutine",
				"runtime.main",
				"stack trace",
				"nil pointer",
				"index out of range",
			},
		},
	}

	for _, test := range regressionTests {
		t.Run(test.name, func(t *testing.T) {
			stdout, stderr, _ := env.RunPVM(test.command...)
			output := stdout + stderr

			for _, badPattern := range test.badPatterns {
				assert.NotContains(t, output, badPattern,
					"Output should not contain bad pattern: %s", badPattern)
			}
		})
	}
}

// TestUIQuality_ReadableOutput tests that output is human-readable
func TestUIQuality_ReadableOutput(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test that help output is readable and well-structured
	readabilityTests := []struct {
		name    string
		command []string
	}{
		{
			name:    "PVX help readability",
			command: []string{"pvx", "--help"},
		},
		{
			name:    "PSC help readability",
			command: []string{"psc", "--help"},
		},
	}

	for _, test := range readabilityTests {
		t.Run(test.name, func(t *testing.T) {
			stdout := helpers.AssertPVMSucceeds(t, env, test.command, test.name)

			// Check for good readability markers
			cleanOutput := stripANSICodes(stdout)

			// Should have clear sections
			hasStructure := strings.Contains(cleanOutput, "Usage:") ||
				strings.Contains(cleanOutput, "Commands:") ||
				strings.Contains(cleanOutput, "Flags:") ||
				strings.Contains(cleanOutput, "Examples:")
			assert.True(t, hasStructure, "Output should have clear structure")

			// Should not be wall of text (should have reasonable line breaks)
			lines := strings.Split(cleanOutput, "\n")
			assert.Greater(t, len(lines), 3, "Output should have multiple lines")

			// Should not have extremely long lines in help text
			for i, line := range lines {
				if len(strings.TrimSpace(line)) > 150 {
					t.Logf("Warning: Line %d is quite long (%d chars): %.50s...",
						i+1, len(line), strings.TrimSpace(line))
				}
			}
		})
	}
}

// TestUIQuality_NoUglyProgressPatterns specifically tests against the ugly progress pattern
func TestUIQuality_NoUglyProgressPatterns(t *testing.T) {
	// Test that our UI framework doesn't produce the ugly progress output pattern
	var buf bytes.Buffer
	ctx := &ui.UIContext{
		Writer:    &buf,
		ColorMode: ui.ColorNever,
		Quiet:     false,
	}
	output := ui.NewOutput(ctx)

	t.Run("Progress indicators are clean", func(t *testing.T) {
		buf.Reset()

		// Simulate multiple progress updates (like what might happen during installation)
		for i := 1; i <= 10; i++ {
			output.Progress(i, 10, "Processing files")
		}

		result := buf.String()
		validateNoUglyProgressOutput(t, result, "UI framework progress")

		// Should be clean progress output
		assert.Contains(t, result, "Processing files", "Should contain progress message")

		// Should not look like the ugly pattern
		assert.NotContains(t, result, "[====================>", "Should not contain ASCII progress bars")
		assert.NotContains(t, result, ".pod", "Should not contain file extensions")
		assert.NotContains(t, result, "/lib/perl5/", "Should not contain internal paths")
	})

	t.Run("Long operations display cleanly", func(t *testing.T) {
		buf.Reset()

		// Simulate processing many files (like installation might do)
		files := []string{
			"config.yaml",
			"main.pl",
			"utils.pm",
			"tests.t",
		}

		for i, file := range files {
			output.Progress(i+1, len(files), "Installing "+file)
		}

		result := buf.String()
		validateNoUglyProgressOutput(t, result, "File processing progress")

		// Should show progress without ugly repetition
		lines := strings.Split(result, "\n")
		nonEmptyLines := 0
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				nonEmptyLines++
			}
		}

		// Should have reasonable number of lines (not one per file with repetitive formatting)
		assert.LessOrEqual(t, nonEmptyLines, len(files)+2, "Should not have excessive output lines")
	})

	t.Run("Status updates are concise", func(t *testing.T) {
		buf.Reset()

		// Test that status updates don't become repetitive walls of text
		statuses := []string{
			"Downloading packages",
			"Extracting archives",
			"Configuring installation",
			"Installing files",
			"Cleaning up",
		}

		for _, status := range statuses {
			output.Status(status)
		}

		result := buf.String()
		validateNoUglyProgressOutput(t, result, "Status updates")

		// Should contain all statuses
		for _, status := range statuses {
			assert.Contains(t, result, status, "Should contain status: %s", status)
		}
	})
}

// TestUIQuality_ConsistentBranding tests that output has consistent branding/styling
func TestUIQuality_ConsistentBranding(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Collect output from different components
	components := []struct {
		name    string
		command []string
	}{
		{"PVX", []string{"pvx", "--help"}},
		{"PVI", []string{"pvi", "--help"}},
		{"PSC", []string{"psc", "--help"}},
	}

	outputs := make(map[string]string)
	for _, comp := range components {
		stdout := helpers.AssertPVMSucceeds(t, env, comp.command, comp.name+" help")
		outputs[comp.name] = stdout
	}

	// Check for consistent patterns across components
	t.Run("Consistent help structure", func(t *testing.T) {
		usageCount := 0
		for name, output := range outputs {
			cleanOutput := stripANSICodes(output)
			if strings.Contains(cleanOutput, "Usage:") {
				usageCount++
			} else {
				t.Logf("%s help may be missing 'Usage:' section", name)
			}
		}

		// Most components should have usage information
		assert.Greater(t, usageCount, 0, "At least some components should have Usage sections")
	})

	t.Run("Consistent styling approach", func(t *testing.T) {
		styledCount := 0
		for name, output := range outputs {
			// Check if output appears to be styled (has ANSI codes or special characters)
			hasANSI := containsANSICodes(output)
			hasIcons := strings.Contains(output, "✓") || strings.Contains(output, "✗") ||
				strings.Contains(output, "ℹ") || strings.Contains(output, "⚠")

			if hasANSI || hasIcons {
				styledCount++
			} else {
				t.Logf("%s output may not be using Fang styling", name)
			}
		}

		// Most components should be using styling
		assert.Greater(t, styledCount, 0, "At least some components should use Fang styling")
	})
}

// Helper function to strip ANSI color codes for clean text analysis
func stripANSICodes(s string) string {
	// Remove ANSI escape sequences
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(s, "")
}

// Helper function to check if string contains ANSI codes
func containsANSICodes(s string) bool {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.MatchString(s)
}
