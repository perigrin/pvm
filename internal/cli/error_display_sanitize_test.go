// ABOUTME: Tests for terminal-output sanitization in DisplayEnhancedError
// ABOUTME: Ensures attacker-influenceable bytes (PATH entries, env vars) cannot inject ANSI escapes or fake lines

package cli

import (
	"bytes"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/errors"
)

// TestDisplayEnhancedError_SanitizesContextValuesWithControlBytes is the
// regression test for issue #455. A malicious PATH entry (or any context
// value derived from environment input) could contain ANSI escape codes,
// newlines, or control bytes; %v formatting forwards them verbatim to
// the terminal. After the fix, control bytes in context values should
// be rendered escaped or stripped, never interpreted by the terminal.
func TestDisplayEnhancedError_SanitizesContextValuesWithControlBytes(t *testing.T) {
	cases := []struct {
		name      string
		conflicts []string
	}{
		{
			name: "ANSI clear-screen escape in PATH entry",
			// \033[2J = clear screen, \033[H = cursor home — a malicious
			// PATH entry could blank PVM's output and reposition the
			// cursor before injecting fake text.
			conflicts: []string{"plenv (\x1b[2J\x1b[H/tmp/evil)"},
		},
		{
			name: "embedded newline in PATH entry",
			// A newline in a PATH entry would let the attacker inject a
			// new line of fake output that looks like PVM said it.
			conflicts: []string{"plenv (/tmp/evil\nFake PVM message: success)"},
		},
		{
			name: "carriage-return-based overwrite",
			// \r alone moves the cursor to column 0 without advancing —
			// next characters overwrite the existing line.
			conflicts: []string{"plenv (/tmp/evil\rOVERWRITTEN)"},
		},
		{
			name: "NUL byte truncation attempt",
			// Some terminals stop at NUL; attacker could hide content.
			conflicts: []string{"plenv (/tmp/before\x00hidden-after)"},
		},
		{
			name:      "DEL character (0x7f)",
			conflicts: []string{"plenv (/tmp/evil\x7f)"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			out := newTestUI(&stdout, &stderr)

			shellErr := errors.NewShellIntegrationError(
				errors.ErrShellVersionManagerConflict,
				"Version manager conflicts detected",
				nil,
				"zsh",
			)
			shellErr.WithContext("detected_conflicts", tc.conflicts)

			DisplayEnhancedError(out, shellErr.EnhancedError)

			combined := stdout.String() + stderr.String()
			assertNoUnescapedControlBytes(t, combined)
		})
	}
}

// TestDisplayEnhancedError_NewlineInValueDoesNotInjectExtraLine asserts
// that a newline embedded in a context value does not inject a fresh
// line of output that could mimic PVM's own messages. Counted by line
// number — a single conflict entry should occupy one line, regardless
// of newlines inside its content.
func TestDisplayEnhancedError_NewlineInValueDoesNotInjectExtraLine(t *testing.T) {
	var stdout, stderr bytes.Buffer
	out := newTestUI(&stdout, &stderr)

	shellErr := errors.NewShellIntegrationError(
		errors.ErrShellVersionManagerConflict,
		"Version manager conflicts detected",
		nil,
		"zsh",
	)
	// Embedded newline tries to make this look like two separate items.
	shellErr.WithContext("detected_conflicts", []string{"plenv (/tmp/evil\nFake PVM message: success)"})

	DisplayEnhancedError(out, shellErr.EnhancedError)

	combined := stdout.String() + stderr.String()
	// "Fake PVM message" must not appear at the start of any rendered
	// line. If it does, an embedded newline broke out of its container
	// and the user sees what looks like PVM's own output.
	for _, line := range strings.Split(combined, "\n") {
		trimmed := strings.TrimLeft(line, " \t]")
		if strings.HasPrefix(trimmed, "Fake PVM message") {
			t.Errorf("attacker-controlled newline injected a standalone line; output:\n%s", combined)
		}
	}
}

// TestDisplayEnhancedError_SanitizesRecoveryActionsWithControlBytes covers
// the other side: recovery actions interpolate paths derived from $HOME
// (e.g. "Add 'eval ...' to ~/.zshrc"). If HOME contains control bytes,
// they reach the terminal through the recovery-action string. Same policy.
func TestDisplayEnhancedError_SanitizesRecoveryActionsWithControlBytes(t *testing.T) {
	var stdout, stderr bytes.Buffer
	out := newTestUI(&stdout, &stderr)

	shellErr := errors.NewMissingShellIntegrationError("zsh")
	shellErr.WithRecoveryAction("Create config file: touch /home/\x1b[31mevil\x1b[0m/.zshrc")

	DisplayEnhancedError(out, shellErr.EnhancedError)

	combined := stdout.String() + stderr.String()
	assertNoUnescapedControlBytes(t, combined)
}

// TestDisplayEnhancedError_PreservesLegitimateContent ensures the
// sanitizer doesn't break normal output: tabs and newlines remain valid
// (the renderer relies on \n for line breaks), printable ASCII passes
// through, and Unicode characters (e.g. the ✓ glyph in Success messages)
// are not mangled.
func TestDisplayEnhancedError_PreservesLegitimateContent(t *testing.T) {
	var stdout, stderr bytes.Buffer
	out := newTestUI(&stdout, &stderr)

	DisplayEnhancedError(out, makeMissingShellIntegrationError())

	combined := stdout.String() + stderr.String()

	// Ordinary content from makeMissingShellIntegrationError must survive.
	for _, want := range []string{"Command: use", "Detected shell: zsh", "Examples after setup:"} {
		if !strings.Contains(combined, want) {
			t.Errorf("expected sanitization to preserve %q, got:\n%s", want, combined)
		}
	}
}

// assertNoUnescapedControlBytes scans s for raw control bytes that should
// have been stripped or escaped by the sanitizer. Permits \n (line
// separator the renderer relies on) and \t (tab) since they're benign
// for terminal display. ESC (0x1b) is the most dangerous one — it's the
// lead byte for ANSI escape sequences. Color codes from the rendering
// library itself stay because tests use ColorNever.
func assertNoUnescapedControlBytes(t *testing.T, s string) {
	t.Helper()
	for i, r := range s {
		switch r {
		case '\n', '\t':
			continue
		}
		if r < 0x20 || r == 0x7f {
			t.Errorf("control byte %#x found at offset %d in rendered output:\n%q", r, i, s)
			return
		}
	}
}
