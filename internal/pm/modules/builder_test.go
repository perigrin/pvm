// ABOUTME: Tests for the module builder functionality
// ABOUTME: Tests detectMakeCommand and related builder helper functions

package modules

import (
	"os/exec"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/perl"
)

// TestDetectMakeCommand verifies that detectMakeCommand returns a non-empty
// make command when queried against a real Perl installation.
func TestDetectMakeCommand(t *testing.T) {
	systemPerl, err := perl.DetectSystemPerl()
	if err != nil {
		t.Skipf("No system Perl available for testing: %v", err)
	}

	// Verify the perl binary is actually executable before proceeding
	if _, err := exec.LookPath(systemPerl.Path); err != nil {
		t.Skipf("System Perl binary not executable: %v", err)
	}

	cmd, err := detectMakeCommand(systemPerl.Path)
	if err != nil {
		t.Fatalf("detectMakeCommand returned unexpected error: %v", err)
	}

	if strings.TrimSpace(cmd) == "" {
		t.Error("detectMakeCommand returned an empty make command")
	}

	t.Logf("detectMakeCommand returned: %q", cmd)
}

// TestDetectMakeCommand_Fallback verifies that detectMakeCommand falls back to
// "make" and returns an error when given a nonexistent Perl path.
func TestDetectMakeCommand_Fallback(t *testing.T) {
	cmd, err := detectMakeCommand("/nonexistent/path/to/perl")
	if err == nil {
		t.Error("detectMakeCommand should return an error for a nonexistent perl path")
	}

	if cmd != "make" {
		t.Errorf("detectMakeCommand should fall back to 'make', got %q", cmd)
	}
}
