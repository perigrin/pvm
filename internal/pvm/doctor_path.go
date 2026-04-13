// ABOUTME: Helpers for the doctor command's PATH-related checks.
// ABOUTME: Separated for unit-test isolation from the main diagnostic flow.

package pvm

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// findStalePvmInstalls scans the entries of PATH looking for pvm executables
// beyond the one the shell resolves first. `activePath` is the canonical path
// returned by exec.LookPath("pvm"); any additional executable named pvm
// (or pvm.exe on Windows) on PATH is considered stale because only the
// first-in-PATH copy is ever invoked.
//
// Returns an empty slice when only the active path is present or PATH is
// empty. Entries are returned in the order they appear in PATH.
func findStalePvmInstalls(pathEnv, activePath string) []string {
	if pathEnv == "" {
		return nil
	}

	binName := "pvm"
	if runtime.GOOS == "windows" {
		binName = "pvm.exe"
	}

	// Resolve the active path through symlinks once so we can compare
	// canonical forms when different PATH entries point at the same file.
	activeCanonical := canonicalize(activePath)

	seen := map[string]struct{}{activeCanonical: {}}
	var stale []string

	for _, dir := range strings.Split(pathEnv, string(os.PathListSeparator)) {
		if dir == "" {
			continue
		}
		candidate := filepath.Join(dir, binName)
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}
		canonical := canonicalize(candidate)
		if _, already := seen[canonical]; already {
			continue
		}
		seen[canonical] = struct{}{}
		stale = append(stale, candidate)
	}
	return stale
}

// canonicalize returns the evaluated symlink form of a path, or the input
// path unchanged when resolution fails (e.g., permission errors on a parent
// directory). Doctor checks should never fail because of canonicalization.
func canonicalize(p string) string {
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil {
		return p
	}
	return resolved
}
