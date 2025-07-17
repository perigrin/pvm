// ABOUTME: Glow integration utility for enhanced markdown rendering
// ABOUTME: Provides graceful fallback to basic markdown when glow is unavailable

package ui

import (
	"bytes"
	"os/exec"
	"strings"
)

// GlowRenderer handles glow integration with graceful fallback
type GlowRenderer struct {
	available bool
	ui        *Output
}

// NewGlowRenderer creates a new glow renderer with auto-detection
func NewGlowRenderer(ui *Output) *GlowRenderer {
	return &GlowRenderer{
		available: isGlowAvailable(),
		ui:        ui,
	}
}

// isGlowAvailable checks if glow is available on the system
func isGlowAvailable() bool {
	_, err := exec.LookPath("glow")
	return err == nil
}

// RenderMarkdown renders markdown content using glow if available, otherwise falls back to basic markdown
func (g *GlowRenderer) RenderMarkdown(content string) error {
	if g.available {
		return g.renderWithGlow(content)
	}
	// Fall back to existing markdown rendering
	g.ui.Markdown(content)
	return nil
}

// renderWithGlow renders markdown content using the glow command
func (g *GlowRenderer) renderWithGlow(content string) error {
	// Prepare glow command with appropriate flags
	cmd := exec.Command("glow", "-")

	// Apply color mode settings
	switch g.ui.context.ColorMode {
	case ColorNever:
		cmd.Args = append(cmd.Args, "--style", "notty")
	case ColorAlways:
		cmd.Args = append(cmd.Args, "--style", "auto")
	case ColorAuto:
		// Let glow auto-detect
		cmd.Args = append(cmd.Args, "--style", "auto")
	}

	// Set up input/output
	cmd.Stdin = strings.NewReader(content)
	var output bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &stderr

	// Execute glow
	err := cmd.Run()
	if err != nil {
		// If glow fails, fall back to basic markdown
		g.ui.Debug("glow execution failed: %v, falling back to basic markdown", err)
		g.ui.Markdown(content)
		return nil
	}

	// Write glow output to the UI writer
	if !g.ui.context.Quiet {
		g.ui.safeWrite(g.ui.context.Writer, output.String())
	}

	return nil
}

// IsAvailable returns whether glow is available on the system
func (g *GlowRenderer) IsAvailable() bool {
	return g.available
}

// checkGlowVersion gets the glow version for debugging purposes
func checkGlowVersion() (string, error) {
	cmd := exec.Command("glow", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// SuggestGlowInstallation provides installation suggestions for glow
func SuggestGlowInstallation(ui *Output) {
	ui.Info("For enhanced markdown display, consider installing glow:")
	ui.Info("  • macOS: brew install glow")
	ui.Info("  • Linux: Download from https://github.com/charmbracelet/glow/releases")
	ui.Info("  • Go: go install github.com/charmbracelet/glow@latest")
}
