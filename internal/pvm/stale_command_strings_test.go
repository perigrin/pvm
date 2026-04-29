// ABOUTME: Corpus-wide regression test guarding against stale 'pvm doctor' command suggestions
// ABOUTME: Scans production Go source for the bare phrase, allowing only approved namespaced or marker-comment contexts

package pvm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoStalePVMDoctorReferences scans the production source tree for the
// literal phrase "pvm doctor" and rejects every occurrence that is not in
// an approved context. The doctor command lives only under "pvm self
// doctor" and "pvm workspace doctor"; bare "pvm doctor" suggestions tell
// users to run a command that returns "unknown command".
//
// Approved contexts:
//   - "pvm self doctor"     — the real top-level form
//   - "pvm workspace doctor" — the workspace variant
//   - shell/config.go marker matching for legacy rc files (with a comment
//     explaining the back-compat reason)
//   - test files (which legitimately reference the bare form to assert it
//     is absent from production strings)
func TestNoStalePVMDoctorReferences(t *testing.T) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	scanRoot := filepath.Join(repoRoot, "internal")

	type hit struct {
		path string
		line int
		text string
	}
	var hits []hit

	err = filepath.WalkDir(scanRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// Production Go only — skip tests and non-Go files. Markdown
		// help assets are checked separately if needed.
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for i, line := range strings.Split(string(data), "\n") {
			if !strings.Contains(line, "pvm doctor") {
				continue
			}
			if isApprovedContext(line) {
				continue
			}
			hits = append(hits, hit{path: path, line: i + 1, text: strings.TrimSpace(line)})
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", scanRoot, err)
	}

	for _, h := range hits {
		rel, _ := filepath.Rel(repoRoot, h.path)
		t.Errorf("%s:%d references nonexistent 'pvm doctor': %s", rel, h.line, h.text)
	}
}

// isApprovedContext reports whether a source line that mentions "pvm
// doctor" is in one of the approved forms.
func isApprovedContext(line string) bool {
	// Strip every approved namespaced form first; if "pvm doctor" remains
	// in what's left, the line has a bare reference.
	stripped := line
	stripped = strings.ReplaceAll(stripped, "pvm self doctor", "")
	stripped = strings.ReplaceAll(stripped, "pvm workspace doctor", "")
	if !strings.Contains(stripped, "pvm doctor") {
		return true
	}

	// Back-compat marker for legacy shell rc files. The full marker
	// quote ("Added by 'pvm doctor --fix'") and the bare "pvm doctor
	// --fix" form are allowed because removePVMInit must keep matching
	// them to clean up rc files written by older PVM versions.
	if strings.Contains(line, "Added by 'pvm doctor --fix'") ||
		strings.Contains(line, `"pvm doctor --fix"`) {
		return true
	}

	return false
}

// findRepoRoot walks upward from the package's working directory until it
// finds the go.mod file at the repository root.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
