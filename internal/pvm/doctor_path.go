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
// returned by exec.LookPath("pvm") — the binary shells will invoke. Any
// additional pvm (or pvm.exe on Windows) on PATH is returned so the caller
// can warn about it. Entries are returned in the order they appear in PATH.
//
// The caller is responsible for disambiguating the "running" binary
// (os.Executable()) in the warning message — findStalePvmInstalls treats all
// non-active copies uniformly because the doctor's job is to tell the user
// "these extras exist"; deciding which extra is which is presentation-layer
// work that has access to all three paths (active, running, extras).
func findStalePvmInstalls(pathEnv, activePath string) []string {
	if pathEnv == "" {
		return nil
	}

	binName := pvmBinaryName()

	seen := map[string]struct{}{}
	if activePath != "" {
		seen[canonicalize(activePath)] = struct{}{}
	}
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

// pvmBinaryName returns the platform-specific executable name. Factored out
// so tests can exercise the Windows branch on non-Windows hosts by injecting
// the name via a parameterized helper (see doctor_path_test.go).
func pvmBinaryName() string {
	if runtime.GOOS == "windows" {
		return "pvm.exe"
	}
	return "pvm"
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
