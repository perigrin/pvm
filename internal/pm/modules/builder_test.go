// ABOUTME: Tests for the module builder functionality
// ABOUTME: Tests detectMakeCommand and related builder helper functions

package modules

import (
	"os/exec"
	"runtime"
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

// TestBuildPLCommand verifies that buildPLCommand returns the correct command
// for the current platform with no extra arguments.
func TestBuildPLCommand(t *testing.T) {
	cmd := buildPLCommand("/usr/bin/perl")
	if runtime.GOOS == "windows" {
		if len(cmd) != 2 || cmd[0] != "/usr/bin/perl" || cmd[1] != "Build" {
			t.Errorf("Expected [/usr/bin/perl Build], got %v", cmd)
		}
	} else {
		if len(cmd) != 1 || cmd[0] != "./Build" {
			t.Errorf("Expected [./Build], got %v", cmd)
		}
	}
}

// TestBuildPLCommand_WithArgs verifies that buildPLCommand correctly threads
// extra arguments after the base command on both platforms.
func TestBuildPLCommand_WithArgs(t *testing.T) {
	cmd := buildPLCommand("/usr/bin/perl", "install", "--verbose")
	if runtime.GOOS == "windows" {
		expected := []string{"/usr/bin/perl", "Build", "install", "--verbose"}
		if len(cmd) != 4 {
			t.Fatalf("Expected 4 elements, got %d: %v", len(cmd), cmd)
		}
		for i, v := range expected {
			if cmd[i] != v {
				t.Errorf("cmd[%d] = %q, want %q", i, cmd[i], v)
			}
		}
	} else {
		expected := []string{"./Build", "install", "--verbose"}
		if len(cmd) != 3 {
			t.Fatalf("Expected 3 elements, got %d: %v", len(cmd), cmd)
		}
		for i, v := range expected {
			if cmd[i] != v {
				t.Errorf("cmd[%d] = %q, want %q", i, cmd[i], v)
			}
		}
	}
}
