// ABOUTME: Tests for DisplayEnhancedError formatting
// ABOUTME: Verifies that error sections render on separate lines and reference real commands

package cli

import (
	"bytes"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/errors"
)

// newTestUI returns a UI Output writing into the given buffers, suitable for
// asserting on rendered output without color codes.
func newTestUI(stdout, stderr *bytes.Buffer) *ui.Output {
	return ui.NewOutput(&ui.UIContext{
		Writer:      stdout,
		ErrorWriter: stderr,
		ColorMode:   ui.ColorNever,
	})
}

// makeMissingShellIntegrationError builds the same error type that
// `pvm perl use system` raises when shell integration is missing.
func makeMissingShellIntegrationError() *errors.EnhancedError {
	shellErr := errors.NewMissingShellIntegrationError("zsh")
	shellErr.WithContext("command", "use")
	shellErr.WithContext("description", "Command requires active shell integration to modify current shell environment")
	shellErr.WithRecoveryAction("Examples after setup:")
	shellErr.WithRecoveryAction("  pvm use 5.38.0    # Use specific Perl version")
	shellErr.WithRecoveryAction("  pvm use system    # Fall back to system Perl")
	return shellErr.EnhancedError
}

func TestDisplayEnhancedErrorEachContextLineEndsWithNewline(t *testing.T) {
	// Without trailing newlines the output runs together as one line —
	// the user observed "Detected shell: zshℹ To fix this issue:" mashed
	// onto a single line.
	var stdout, stderr bytes.Buffer
	out := newTestUI(&stdout, &stderr)

	DisplayEnhancedError(out, makeMissingShellIntegrationError())

	combined := stdout.String() + stderr.String()

	// Each of these labels names a line; if any two appear on the same
	// physical line of output, formatting is broken.
	labels := []string{
		"Command:",
		"Issue:",
		"Detected shell:",
		"To fix this issue:",
	}

	lines := strings.Split(combined, "\n")
	type seen struct {
		label string
		line  int
	}
	var found []seen
	for i, line := range lines {
		for _, lbl := range labels {
			if strings.Contains(line, lbl) {
				found = append(found, seen{lbl, i})
			}
		}
	}

	for i := 0; i < len(found); i++ {
		for j := i + 1; j < len(found); j++ {
			if found[i].line == found[j].line {
				t.Errorf("labels %q and %q rendered on the same line %d; combined output:\n%s",
					found[i].label, found[j].label, found[i].line, combined)
			}
		}
	}
}

func TestDisplayEnhancedErrorReferencesRealDoctorCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	out := newTestUI(&stdout, &stderr)

	DisplayEnhancedError(out, makeMissingShellIntegrationError())

	combined := stdout.String() + stderr.String()

	// Search for "pvm doctor" not preceded by "self".
	lines := strings.Split(combined, "\n")
	for i, line := range lines {
		prefix, _, ok := strings.Cut(line, "pvm doctor")
		if !ok {
			continue
		}
		// Allow "pvm self doctor".
		if strings.HasSuffix(strings.TrimSpace(prefix), "pvm self") {
			continue
		}
		t.Errorf("line %d references nonexistent 'pvm doctor': %q (full output:\n%s)", i, line, combined)
	}
}
