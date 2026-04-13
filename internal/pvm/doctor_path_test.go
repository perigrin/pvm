// ABOUTME: Unit tests for findStalePvmInstalls.
// ABOUTME: Drives the check with a fake PATH and fabricated pvm binaries.

package pvm

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFindStalePvmInstalls(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows binary naming differs; tested manually for now")
	}

	tmp := t.TempDir()

	// Build a fake two-directory layout with a pvm binary in each.
	dirA := filepath.Join(tmp, "a")
	dirB := filepath.Join(tmp, "b")
	dirC := filepath.Join(tmp, "c") // empty — should be ignored
	for _, d := range []string{dirA, dirB, dirC} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	for _, d := range []string{dirA, dirB} {
		path := filepath.Join(d, "pvm")
		if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatalf("write fake pvm in %s: %v", d, err)
		}
	}

	cases := []struct {
		name        string
		pathEnv     string
		active      string
		wantStale   []string
		wantNoStale bool
	}{
		{
			name:      "two installs, A active",
			pathEnv:   strings.Join([]string{dirA, dirB, dirC}, string(os.PathListSeparator)),
			active:    filepath.Join(dirA, "pvm"),
			wantStale: []string{filepath.Join(dirB, "pvm")},
		},
		{
			name:      "two installs, B active",
			pathEnv:   strings.Join([]string{dirA, dirB}, string(os.PathListSeparator)),
			active:    filepath.Join(dirB, "pvm"),
			wantStale: []string{filepath.Join(dirA, "pvm")},
		},
		{
			name:        "only the active install present",
			pathEnv:     strings.Join([]string{dirA, dirC}, string(os.PathListSeparator)),
			active:      filepath.Join(dirA, "pvm"),
			wantNoStale: true,
		},
		{
			name:        "empty PATH",
			pathEnv:     "",
			active:      filepath.Join(dirA, "pvm"),
			wantNoStale: true,
		},
		{
			name:        "PATH has duplicate entries of the active dir",
			pathEnv:     strings.Join([]string{dirA, dirA, dirC}, string(os.PathListSeparator)),
			active:      filepath.Join(dirA, "pvm"),
			wantNoStale: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := findStalePvmInstalls(tc.pathEnv, tc.active)
			if tc.wantNoStale {
				if len(got) != 0 {
					t.Errorf("expected no stale installs, got %v", got)
				}
				return
			}
			if len(got) != len(tc.wantStale) {
				t.Fatalf("expected %d stale installs, got %d: %v", len(tc.wantStale), len(got), got)
			}
			for i, want := range tc.wantStale {
				if got[i] != want {
					t.Errorf("stale[%d]: got %q, want %q", i, got[i], want)
				}
			}
		})
	}
}

func TestFindStalePvmInstalls_SymlinkDedup(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows symlink handling differs")
	}

	tmp := t.TempDir()
	realDir := filepath.Join(tmp, "real")
	linkDir := filepath.Join(tmp, "link")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	realBin := filepath.Join(realDir, "pvm")
	if err := os.WriteFile(realBin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write pvm: %v", err)
	}
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	// PATH contains both the real dir and the symlinked one. The pvm binary
	// is physically the same file; findStalePvmInstalls must not flag it
	// as stale.
	pathEnv := strings.Join([]string{realDir, linkDir}, string(os.PathListSeparator))
	got := findStalePvmInstalls(pathEnv, realBin)
	if len(got) != 0 {
		t.Errorf("expected symlinked duplicate to dedupe, got stale=%v", got)
	}
}
