// ABOUTME: Tests that RenderMarkdownAsHelp surfaces all the markdown source's
// ABOUTME: content — particularly bullet lists, which were previously dropped silently

package cli

import (
	"bytes"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/data"
)

// renderToBuffer runs RenderMarkdownAsHelp into a bytes.Buffer-backed UI
// so tests can assert on the rendered content.
func renderToBuffer(t *testing.T, markdown string) string {
	t.Helper()
	var buf bytes.Buffer
	out := ui.NewOutput(&ui.UIContext{
		Writer:    &buf,
		ColorMode: ui.ColorNever,
	})
	RenderMarkdownAsHelp(markdown, out)
	return buf.String()
}

// TestRenderMarkdownAsHelp_RendersBulletListItems is the regression test
// for the bullet-dropping bug surfaced by issue #451. Before the fix,
// `RenderMarkdownAsHelp` silently dropped every "- foo" line with a TODO
// comment claiming list rendering would happen "separately" — but no
// separate rendering existed. The "Diagnostic Commands" section of
// troubleshooting.md was therefore invisible to users.
func TestRenderMarkdownAsHelp_RendersBulletListItems(t *testing.T) {
	markdown := `## Diagnostic Commands

- Check overall workspace health: ` + "`pvm workspace doctor`" + `
- View detailed workspace status: ` + "`pvm workspace status --json`"

	rendered := renderToBuffer(t, markdown)

	// The substantive content of each bullet should appear in the rendered
	// output. We're not picky about formatting (color, prefix, etc.) — just
	// that the text is present.
	mustContain := []string{
		"Check overall workspace health",
		"pvm workspace doctor",
		"View detailed workspace status",
		"pvm workspace status --json",
	}
	for _, want := range mustContain {
		if !strings.Contains(rendered, want) {
			t.Errorf("expected rendered output to contain %q, got:\n%s", want, rendered)
		}
	}
}

// TestTroubleshootingHelp_DiagnosticCommandsAreVisible is the issue-#451
// integration test: load the actual troubleshooting.md asset, render it
// the way `pvm help troubleshooting` does, and assert the user can see
// every command listed in the "Diagnostic Commands" section. This catches
// both (a) the markdown content drifting from reality and (b) the renderer
// regressing on bullet handling.
func TestTroubleshootingHelp_DiagnosticCommandsAreVisible(t *testing.T) {
	markdown, err := data.GetHelpTemplate("troubleshooting")
	if err != nil {
		t.Fatalf("load troubleshooting.md: %v", err)
	}

	rendered := renderToBuffer(t, markdown)

	// The "workspace health" line must point at `pvm workspace doctor`.
	// Both commands exist; this one specifically checks workspace health.
	// The drift from `pvm self doctor` (which checks PVM install, not
	// workspace) was the original report.
	if !strings.Contains(rendered, "Check overall workspace health") {
		t.Errorf("rendered help is missing the 'workspace health' diagnostic line; output:\n%s", rendered)
	}
	if !strings.Contains(rendered, "pvm workspace doctor") {
		t.Errorf("rendered help should mention 'pvm workspace doctor' for workspace health; output:\n%s", rendered)
	}

	// "Verify Perl version resolution" should land on `pvm perl resolve`
	// (NOT bare `pvm resolve`, which doesn't exist as a top-level command).
	if !strings.Contains(rendered, "pvm perl resolve") {
		t.Errorf("rendered help should mention 'pvm perl resolve'; output:\n%s", rendered)
	}
}
