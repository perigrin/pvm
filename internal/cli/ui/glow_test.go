// ABOUTME: Unit tests for glamour integration functionality
// ABOUTME: Tests built-in markdown rendering, raw markdown mode, and fallback behavior

package ui

import (
	"bytes"
	"strconv"
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

	// Test that IsAvailable returns true (glamour is built-in)
	available := renderer.IsAvailable()
	if !available {
		t.Error("IsAvailable() should return true with built-in glamour")
	}
}

func TestGlowRenderer_StyledMarkdown(t *testing.T) {
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

	// Render markdown using glamour styling
	err := renderer.RenderMarkdown(testContent)
	if err != nil {
		t.Fatalf("RenderMarkdown failed: %v", err)
	}

	// Verify output was written
	output := buf.String()
	if len(output) == 0 {
		t.Error("No output generated")
	}

	// Should contain formatted content with glamour styling
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

func TestGlowRenderer_RawMarkdownMode(t *testing.T) {
	var buf bytes.Buffer
	ui := NewDefaultOutput()
	ui.SetWriter(&buf)

	// Enable raw markdown mode
	ui.context.RawMarkdown = true

	renderer := NewGlowRenderer(ui)

	testContent := `# Test Header

This is **bold** text and *italic* text.

- List item
- Another item`

	err := renderer.RenderMarkdown(testContent)
	if err != nil {
		t.Fatalf("RenderMarkdown failed in raw mode: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("No output generated in raw markdown mode")
	}

	// In raw mode, should contain the original markdown text
	if !strings.Contains(output, "Test Header") {
		t.Error("Raw mode output doesn't contain expected header")
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
		content.WriteString(strconv.Itoa(i))
		content.WriteString("\n\n")
		content.WriteString("This is paragraph ")
		content.WriteString(strconv.Itoa(i))
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
