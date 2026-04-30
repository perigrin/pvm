// ABOUTME: Tests for checkRegistryIntegrity's classification of fresh-install state
// ABOUTME: A missing registry file with no installed versions is the expected steady state, not a warning

package pvm

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/cli/ui"
)

// quietUI returns a Quiet UI suitable for tests that don't care about live
// output. The doctor functions take a *ui.Output for status messaging; the
// tests assert on the issues/warnings slices instead.
func quietUI() *ui.Output {
	return ui.NewOutput(&ui.UIContext{
		Writer: os.Stdout,
		Quiet:  true,
	})
}

// withFreshXDGDataHome runs fn with XDG_DATA_HOME pointing at an empty
// tempdir, restoring the prior value afterward. fn receives the path that
// xdg.GetDirs will treat as the PVM data dir (XDG_DATA_HOME/pvm) — that's
// where registry.json and versions/ live, not at XDG_DATA_HOME itself.
func withFreshXDGDataHome(t *testing.T, fn func(pvmDataDir string)) {
	t.Helper()
	root := t.TempDir()
	prev, hadPrev := os.LookupEnv("XDG_DATA_HOME")
	if err := os.Setenv("XDG_DATA_HOME", root); err != nil {
		t.Fatalf("set XDG_DATA_HOME: %v", err)
	}
	defer func() {
		if hadPrev {
			_ = os.Setenv("XDG_DATA_HOME", prev)
		} else {
			_ = os.Unsetenv("XDG_DATA_HOME")
		}
	}()
	pvmDataDir := filepath.Join(root, "pvm")
	fn(pvmDataDir)
}

// TestCheckRegistryIntegrity_FreshInstallEmitsNoWarning is the regression
// test for issue #447. On a fresh install (no registry file, no versions
// directory), the doctor used to emit a warning ("Registry file doesn't
// exist (no versions installed)") that trained users to ignore doctor
// output. After the fix, this combination is treated as the expected
// steady state and emits nothing into the warnings slice.
func TestCheckRegistryIntegrity_FreshInstallEmitsNoWarning(t *testing.T) {
	withFreshXDGDataHome(t, func(pvmDataDir string) {
		// Sanity: on a fresh install the pvm data dir doesn't exist yet.
		// checkRegistryIntegrity must tolerate that without erroring or
		// warning.
		if _, err := os.Stat(pvmDataDir); !os.IsNotExist(err) {
			t.Fatalf("expected pvmDataDir not to exist on fresh install, got err=%v", err)
		}

		var issues, warnings []string
		err := checkRegistryIntegrity(quietUI(), &issues, &warnings)
		if err != nil {
			t.Fatalf("checkRegistryIntegrity: %v", err)
		}

		if len(issues) != 0 {
			t.Errorf("expected no issues on fresh install, got %v", issues)
		}
		if len(warnings) != 0 {
			t.Errorf("expected no warnings on fresh install, got %v", warnings)
		}
	})
}

// TestCheckRegistryIntegrity_RegistryMissingButVersionsExistStillFlagsIssue
// locks in the OTHER branch of the same conditional: when versions directory
// has installations but registry.json is missing, that's a real broken state
// and must continue to be reported as an issue (not silently dropped by an
// over-eager fix to issue #447).
func TestCheckRegistryIntegrity_RegistryMissingButVersionsExistStillFlagsIssue(t *testing.T) {
	withFreshXDGDataHome(t, func(pvmDataDir string) {
		// Create a versions directory under the pvm data dir with one fake
		// installation directory inside it. No registry.json.
		versionsDir := filepath.Join(pvmDataDir, "versions")
		fakeInstall := filepath.Join(versionsDir, "5.40.0")
		if err := os.MkdirAll(fakeInstall, 0755); err != nil {
			t.Fatalf("mkdir fake install: %v", err)
		}

		var issues, warnings []string
		err := checkRegistryIntegrity(quietUI(), &issues, &warnings)
		if err != nil {
			t.Fatalf("checkRegistryIntegrity: %v", err)
		}

		// The "registry missing but versions exist" path must still flag
		// an issue.
		foundIssue := false
		for _, i := range issues {
			if i == "Registry missing but versions directory contains installations" {
				foundIssue = true
				break
			}
		}
		if !foundIssue {
			t.Errorf("expected the 'Registry missing but versions directory contains installations' issue, got issues=%v warnings=%v", issues, warnings)
		}
	})
}
