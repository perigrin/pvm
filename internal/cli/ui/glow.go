// ABOUTME: Built-in glamour integration for enhanced markdown rendering
// ABOUTME: Provides consistent markdown styling with optional raw output fallback

package ui

import (
	"github.com/charmbracelet/glamour"
)

// GlowRenderer handles built-in glamour integration for markdown rendering
type GlowRenderer struct {
	renderer *glamour.TermRenderer
	ui       *Output
}

// NewGlowRenderer creates a new glamour-based renderer
func NewGlowRenderer(ui *Output) *GlowRenderer {
	renderer, err := createGlamourRenderer(ui)
	if err != nil {
		// If glamour setup fails, we'll handle this in RenderMarkdown
		ui.Debug("failed to create glamour renderer: %v", err)
	}

	return &GlowRenderer{
		renderer: renderer,
		ui:       ui,
	}
}

// createGlamourRenderer creates a glamour terminal renderer with appropriate settings
func createGlamourRenderer(ui *Output) (*glamour.TermRenderer, error) {
	options := []glamour.TermRendererOption{}

	// Configure styling based on color mode
	switch ui.context.ColorMode {
	case ColorNever:
		// Use no styling for plain text output
		options = append(options, glamour.WithStandardStyle("notty"))
	case ColorAlways, ColorAuto:
		// Use automatic style detection (dark/light based on terminal)
		options = append(options, glamour.WithAutoStyle())
	}

	// Set word wrap to reasonable terminal width
	options = append(options, glamour.WithWordWrap(80))

	return glamour.NewTermRenderer(options...)
}

// RenderMarkdown renders markdown content using glamour, with fallback to basic markdown
func (g *GlowRenderer) RenderMarkdown(content string) error {
	// Check if raw markdown is requested via configuration
	if g.shouldUseRawMarkdown() {
		g.ui.Markdown(content)
		return nil
	}

	// Attempt to render with glamour
	if g.renderer != nil {
		return g.renderWithGlamour(content)
	}

	// Fall back to basic markdown rendering
	g.ui.Debug("glamour renderer unavailable, falling back to basic markdown")
	g.ui.Markdown(content)
	return nil
}

// renderWithGlamour renders markdown content using the glamour library
func (g *GlowRenderer) renderWithGlamour(content string) error {
	output, err := g.renderer.Render(content)
	if err != nil {
		// If glamour fails, fall back to basic markdown
		g.ui.Debug("glamour rendering failed: %v, falling back to basic markdown", err)
		g.ui.Markdown(content)
		return nil
	}

	// Write glamour output to the UI writer
	if !g.ui.context.Quiet {
		g.ui.safeWrite(g.ui.context.Writer, output)
	}

	return nil
}

// shouldUseRawMarkdown checks if raw markdown output is requested
func (g *GlowRenderer) shouldUseRawMarkdown() bool {
	// Check the RawMarkdown flag from the UI context
	return g.ui.context.RawMarkdown
}

// IsAvailable returns whether glamour rendering is available (false only if initialization failed)
func (g *GlowRenderer) IsAvailable() bool {
	return g.renderer != nil
}
