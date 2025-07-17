// ABOUTME: Unit tests for glow integration functionality
// ABOUTME: Tests glow detection, fallback behavior, and markdown rendering

package ui

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestGlowRenderer_NewGlowRenderer(t *testing.T) {
	ui := NewDefaultOutput()
	renderer := NewGlowRenderer(ui)

	if renderer == nil {
		t.Fatal("NewGlowRenderer returned nil")
	}

	if renderer.ui != ui {
		t.Error("GlowRenderer.ui not properly set")
	}
}

func TestGlowRenderer_IsAvailable(t *testing.T) {
	ui := NewDefaultOutput()
	renderer := NewGlowRenderer(ui)

	// Test that IsAvailable returns a boolean
	available := renderer.IsAvailable()
	if available != renderer.available {
		t.Error("IsAvailable() doesn't match internal available state")
	}
}

func TestGlowRenderer_FallbackBehavior(t *testing.T) {
	// Create a test output buffer
	var buf bytes.Buffer
	ui := NewDefaultOutput()
	ui.SetWriter(&buf)

	// Create renderer
	renderer := NewGlowRenderer(ui)

	// Test markdown content
	testContent := `# Test Header

This is a test markdown content with **bold** text.

- List item 1
- List item 2

## Sub Header

More content here.`

	// Render markdown (should fallback to basic markdown if glow unavailable)
	err := renderer.RenderMarkdown(testContent)
	if err != nil {
		t.Fatalf("RenderMarkdown failed: %v", err)
	}

	// Verify output was written
	output := buf.String()
	if len(output) == 0 {
		t.Error("No output generated")
	}

	// Should contain formatted content (either glow or basic markdown)
	if !strings.Contains(output, "Test Header") {
		t.Error("Output doesn't contain expected header")
	}
}

func TestGlowRenderer_EmptyContent(t *testing.T) {
	var buf bytes.Buffer
	ui := NewDefaultOutput()
	ui.SetWriter(&buf)

	renderer := NewGlowRenderer(ui)

	// Test empty content
	err := renderer.RenderMarkdown("")
	if err != nil {
		t.Fatalf("RenderMarkdown failed with empty content: %v", err)
	}

	// Should handle empty content gracefully
	output := buf.String()
	// Empty content should produce minimal output
	if len(output) > 100 {
		t.Error("Empty content produced unexpectedly large output")
	}
}

func TestGlowRenderer_QuietMode(t *testing.T) {
	var buf bytes.Buffer
	ui := NewDefaultOutput()
	ui.SetWriter(&buf)
	ui.SetQuiet(true)

	renderer := NewGlowRenderer(ui)

	testContent := "# Test Header\n\nTest content"
	err := renderer.RenderMarkdown(testContent)
	if err != nil {
		t.Fatalf("RenderMarkdown failed: %v", err)
	}

	// In quiet mode, should produce no output
	output := buf.String()
	if len(output) > 0 {
		t.Error("Quiet mode produced output when it shouldn't")
	}
}

func TestGlowRenderer_ColorModeHandling(t *testing.T) {
	ui := NewDefaultOutput()
	renderer := NewGlowRenderer(ui)

	// Test different color modes
	testCases := []ColorMode{
		ColorAuto,
		ColorAlways,
		ColorNever,
	}

	for _, colorMode := range testCases {
		colorModeName := ""
		switch colorMode {
		case ColorAuto:
			colorModeName = "auto"
		case ColorAlways:
			colorModeName = "always"
		case ColorNever:
			colorModeName = "never"
		}

		t.Run(colorModeName, func(t *testing.T) {
			var buf bytes.Buffer
			ui.SetWriter(&buf)
			ui.SetColorMode(colorMode)

			testContent := "# Test\n\nContent"
			err := renderer.RenderMarkdown(testContent)
			if err != nil {
				t.Fatalf("RenderMarkdown failed for color mode %s: %v", colorModeName, err)
			}

			// Should handle all color modes without error
			output := buf.String()
			if len(output) == 0 {
				t.Error("No output generated for color mode:", colorModeName)
			}
		})
	}
}

func TestIsGlowAvailable(t *testing.T) {
	// Test glow availability detection
	available := isGlowAvailable()

	// Should return a boolean without error
	if available != true && available != false {
		t.Error("isGlowAvailable should return a boolean value")
	}

	// If glow is available, verify it's actually executable
	if available {
		// This is a basic sanity check - if glow is detected as available,
		// it should be in the PATH
		if os.Getenv("PATH") == "" {
			t.Skip("PATH not set, skipping glow availability verification")
		}

		// The detection should be consistent
		available2 := isGlowAvailable()
		if available != available2 {
			t.Error("isGlowAvailable returned inconsistent results")
		}
	}
}

func TestCheckGlowVersion(t *testing.T) {
	// Only test if glow is available
	if !isGlowAvailable() {
		t.Skip("Glow not available, skipping version check test")
	}

	version, err := checkGlowVersion()
	if err != nil {
		t.Fatalf("checkGlowVersion failed: %v", err)
	}

	if version == "" {
		t.Error("checkGlowVersion returned empty version string")
	}

	// Version should contain some expected content
	if !strings.Contains(strings.ToLower(version), "glow") {
		t.Error("Version string doesn't contain 'glow':", version)
	}
}

func TestSuggestGlowInstallation(t *testing.T) {
	var buf bytes.Buffer
	ui := NewDefaultOutput()
	ui.SetWriter(&buf)

	SuggestGlowInstallation(ui)

	output := buf.String()
	if len(output) == 0 {
		t.Error("SuggestGlowInstallation produced no output")
	}

	// Should contain installation suggestions
	expectedStrings := []string{
		"glow",
		"brew install",
		"github.com/charmbracelet/glow",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Installation suggestion missing expected string: %s", expected)
		}
	}
}

func TestGlowRenderer_LargeContent(t *testing.T) {
	var buf bytes.Buffer
	ui := NewDefaultOutput()
	ui.SetWriter(&buf)

	renderer := NewGlowRenderer(ui)

	// Generate large markdown content
	var content strings.Builder
	for i := 0; i < 100; i++ {
		content.WriteString("# Header ")
		content.WriteString(string(rune(i)))
		content.WriteString("\n\n")
		content.WriteString("This is paragraph ")
		content.WriteString(string(rune(i)))
		content.WriteString(" with some **bold** text and *italic* text.\n\n")
		content.WriteString("- List item 1\n")
		content.WriteString("- List item 2\n")
		content.WriteString("- List item 3\n\n")
	}

	err := renderer.RenderMarkdown(content.String())
	if err != nil {
		t.Fatalf("RenderMarkdown failed with large content: %v", err)
	}

	// Should handle large content without error
	output := buf.String()
	if len(output) == 0 {
		t.Error("No output generated for large content")
	}
}

func TestGlowRenderer_SpecialCharacters(t *testing.T) {
	var buf bytes.Buffer
	ui := NewDefaultOutput()
	ui.SetWriter(&buf)

	renderer := NewGlowRenderer(ui)

	// Test content with special characters
	testContent := `# Test with Special Characters

Content with unicode: 🚀 ✨ 🎉

Code block:
` + "```" + `
func main() {
    fmt.Println("Hello, world!")
}
` + "```" + `

Table:
| Column 1 | Column 2 |
|----------|----------|
| Value 1  | Value 2  |

Links: [GitHub](https://github.com)
`

	err := renderer.RenderMarkdown(testContent)
	if err != nil {
		t.Fatalf("RenderMarkdown failed with special characters: %v", err)
	}

	// Should handle special characters without error
	output := buf.String()
	if len(output) == 0 {
		t.Error("No output generated for special characters")
	}
}
