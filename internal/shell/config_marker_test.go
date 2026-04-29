// ABOUTME: Round-trip tests for the PVM shell-config marker comment
// ABOUTME: Ensures both legacy 'pvm doctor --fix' and current 'pvm self doctor --fix' markers are removed cleanly

package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const userPreambleLine = "# user content above\n"

// writeFile writes content with a parent dir check; t.Helper() makes failures
// point at the caller.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func TestRemovePVMInitStripsLegacyMarker(t *testing.T) {
	// Existing user shell configs may still have the legacy "pvm doctor --fix"
	// marker. Removal must continue to work for them.
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	content := userPreambleLine +
		"\n" +
		"# PVM (Perl Version Manager) initialization\n" +
		"# Added by 'pvm doctor --fix'\n" +
		"eval \"$(pvm init)\"\n"
	writeFile(t, rc, content)

	cm := &ConfigManager{}
	if err := cm.removePVMInit(rc); err != nil {
		t.Fatalf("removePVMInit: %v", err)
	}

	got := readFile(t, rc)
	for _, banned := range []string{"pvm init", "PVM (Perl Version Manager)", "pvm doctor --fix"} {
		if strings.Contains(got, banned) {
			t.Errorf("expected %q to be stripped, file is now:\n%s", banned, got)
		}
	}
	if !strings.Contains(got, "# user content above") {
		t.Errorf("user content was incorrectly removed; file is now:\n%s", got)
	}
}

func TestRemovePVMInitStripsCurrentMarker(t *testing.T) {
	// New installs write the "pvm self doctor --fix" marker. Removal must
	// recognize it.
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	content := userPreambleLine +
		"\n" +
		"# PVM (Perl Version Manager) initialization\n" +
		"# Added by 'pvm self doctor --fix'\n" +
		"eval \"$(pvm init)\"\n"
	writeFile(t, rc, content)

	cm := &ConfigManager{}
	if err := cm.removePVMInit(rc); err != nil {
		t.Fatalf("removePVMInit: %v", err)
	}

	got := readFile(t, rc)
	for _, banned := range []string{"pvm init", "PVM (Perl Version Manager)", "pvm self doctor --fix"} {
		if strings.Contains(got, banned) {
			t.Errorf("expected %q to be stripped, file is now:\n%s", banned, got)
		}
	}
	if !strings.Contains(got, "# user content above") {
		t.Errorf("user content was incorrectly removed; file is now:\n%s", got)
	}
}

func TestAppendPVMInitWritesCurrentMarker(t *testing.T) {
	// Newly-written shell configs should reference the real command,
	// not the dead "pvm doctor --fix".
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	writeFile(t, rc, userPreambleLine)

	cm := &ConfigManager{}
	cfg := &ShellConfig{InitCommand: `eval "$(pvm init)"`}
	if err := cm.appendPVMInit(rc, cfg); err != nil {
		t.Fatalf("appendPVMInit: %v", err)
	}

	got := readFile(t, rc)
	if !strings.Contains(got, "Added by 'pvm self doctor --fix'") {
		t.Errorf("expected new marker, got:\n%s", got)
	}
	// We tolerate the legacy marker only on existing configs we read; we
	// never write it on new appends.
	if strings.Contains(got, "Added by 'pvm doctor --fix'") {
		t.Errorf("appendPVMInit must not write the dead 'pvm doctor --fix' marker; got:\n%s", got)
	}
}
