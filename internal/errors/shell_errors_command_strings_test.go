// ABOUTME: Regression tests asserting shell-integration error recovery actions
// ABOUTME: only reference real PVM commands (pvm self doctor, not pvm doctor)

package errors

import (
	"strings"
	"testing"
)

// staleDoctorReference reports any recovery action that mentions the dead
// top-level "pvm doctor" command. The doctor command lives under "pvm self",
// so a bare "pvm doctor" suggestion would tell users to run a command that
// does not exist.
func staleDoctorReference(actions []string) (int, string) {
	for i, action := range actions {
		if !strings.Contains(action, "pvm doctor") {
			continue
		}
		if strings.Contains(action, "pvm self doctor") {
			continue
		}
		return i, action
	}
	return -1, ""
}

func TestShellIntegrationErrorRecoveryActionsReferenceRealCommands(t *testing.T) {
	cases := []struct {
		name string
		err  *ShellIntegrationError
	}{
		{"missing-shell-integration-bash", NewMissingShellIntegrationError("bash")},
		{"missing-shell-integration-zsh", NewMissingShellIntegrationError("zsh")},
		{"missing-shell-integration-fish", NewMissingShellIntegrationError("fish")},
		{"shell-config-missing-zsh", NewShellConfigMissingError("zsh")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actions := tc.err.RecoveryActions()
			if len(actions) == 0 {
				t.Fatalf("expected at least one recovery action, got none")
			}
			if i, bad := staleDoctorReference(actions); i >= 0 {
				t.Errorf("recovery action %d references nonexistent 'pvm doctor': %q", i, bad)
			}
		})
	}
}

func TestDirectoryGuidanceReferencesRealCommand(t *testing.T) {
	// addDirectoryGuidance is exercised by the "directory missing" error code
	err := NewShellIntegrationError(
		ErrShellDirectoryMissing,
		"Required directories are missing",
		nil,
		"zsh",
	)
	if i, bad := staleDoctorReference(err.RecoveryActions()); i >= 0 {
		t.Errorf("directory-missing recovery action %d references nonexistent 'pvm doctor': %q", i, bad)
	}
}
