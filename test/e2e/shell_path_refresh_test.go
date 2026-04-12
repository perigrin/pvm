// ABOUTME: End-to-end verification that `pvm use <version>` refreshes PATH
// ABOUTME: across bash, zsh, and fish in the current shell.

package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/test/e2e/helpers"
)

// TestPvmUseRefreshesPathAcrossShells verifies that after `pvm use X`, the
// activated Perl's bin directory is at the front of PATH in the same shell.
// Stronger than TestShellTemplatesRefreshPathAfterShUse (which only checks
// the template text) because this actually sources the template and inspects
// PATH after `pvm use` runs. Missing shells are skipped rather than failing.
func TestPvmUseRefreshesPathAcrossShells(t *testing.T) {
	helpers.SkipIfNotUnix(t)

	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// The test env seeds a .perl-version file matching the project's pinned
	// version. If that file is present, the resolver picks up the version on
	// shell init and PATH is already correct before `pvm use` runs, which
	// defeats the point of the test. Remove it so nothing implicit activates
	// a version — `pvm use` must be the thing that brings the bin into PATH.
	_ = os.Remove(filepath.Join(env.HomeDir, ".perl-version"))

	stdout, stderr, err := env.RunPVM("import-system")
	if err != nil {
		t.Fatalf("import-system failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	version := firstInstalledPerlVersion(t)
	if version == "" {
		t.Fatalf("import-system succeeded but no installed Perl version is registered")
	}
	expectedBin := filepath.Join(env.DataHome, "pvm", "versions", version, "bin")

	// The template's `[ -d "$new_perl_bin" ]` / `test -d ...` guards skip
	// adding the directory if it's missing; import-system doesn't populate a
	// version-local bin dir, so create one.
	if err := os.MkdirAll(expectedBin, 0o755); err != nil {
		t.Fatalf("create expected bin dir: %v", err)
	}

	// After sourcing the template, `pvm_init` calls `_pvm_update_perl_path`
	// which picks the initial (imported system) version. Then `pvm use
	// <targetVersion>` must swap PATH to point at the target version's bin
	// — the exact behavior the post-sh-use refresh provides.
	cases := []struct {
		shell     string
		scriptFmt string
	}{
		{
			shell: "bash",
			scriptFmt: `#!/bin/bash
set -e
export PVM_SKIP_NETWORK_CALLS=1
source "%s"
pvm use "%s"
echo "PATH_AFTER_USE=$PATH"
`,
		},
		{
			shell: "zsh",
			scriptFmt: `#!/usr/bin/env zsh
set -e
export PVM_SKIP_NETWORK_CALLS=1
source "%s"
pvm use "%s"
echo "PATH_AFTER_USE=$PATH"
`,
		},
		{
			shell: "fish",
			scriptFmt: `#!/usr/bin/env fish
set -gx PVM_SKIP_NETWORK_CALLS 1
source "%s"
pvm use "%s"
echo "PATH_AFTER_USE=$PATH"
`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.shell, func(t *testing.T) {
			if _, err := exec.LookPath(tc.shell); err != nil {
				t.Skipf("%s not available in PATH", tc.shell)
			}

			initScript, err := generateShellInitScript(env, tc.shell)
			if err != nil {
				t.Fatalf("generate %s init script: %v", tc.shell, err)
			}
			initPath := filepath.Join(env.HomeDir, "pvm_init."+tc.shell)
			if err := os.WriteFile(initPath, []byte(initScript), 0o644); err != nil {
				t.Fatalf("write init script: %v", err)
			}

			testPath := filepath.Join(env.HomeDir, "path_refresh."+tc.shell)
			content := fmt.Sprintf(tc.scriptFmt, initPath, version)
			if err := os.WriteFile(testPath, []byte(content), 0o755); err != nil {
				t.Fatalf("write test script: %v", err)
			}

			stdout, stderr, err := env.RunCommand(tc.shell, testPath)
			if err != nil {
				t.Fatalf("%s test script failed: %v\nstdout: %s\nstderr: %s", tc.shell, err, stdout, stderr)
			}
			output := stdout + stderr

			pathAfterUse := extractPathLine(output, "PATH_AFTER_USE=")
			if pathAfterUse == "" {
				t.Fatalf("missing PATH_AFTER_USE marker in output:\n%s", output)
			}
			if !pathHasFrontEntry(pathAfterUse, expectedBin) {
				t.Errorf("%s: %q should be at the front of PATH after `pvm use %s`\nPATH_AFTER_USE=%s\nfull output: %s",
					tc.shell, expectedBin, version, pathAfterUse, output)
			}
		})
	}
}

// firstInstalledPerlVersion returns the first 5.x Perl version from the
// registry, reading directly from the XDG data dirs the test env sets up.
// Returns "" if nothing is installed.
func firstInstalledPerlVersion(t *testing.T) string {
	t.Helper()
	versions, err := perl.GetInstalledVersions()
	if err != nil {
		t.Fatalf("perl.GetInstalledVersions: %v", err)
	}
	for _, v := range versions {
		if strings.HasPrefix(v.Version, "5.") {
			return v.Version
		}
	}
	return ""
}

// generateShellInitScript runs `pvm init <shell>` and returns its output.
// It uses the test env's PVM binary directly so the generated script points
// at it rather than the system's pvm.
func generateShellInitScript(env *helpers.TestEnv, shell string) (string, error) {
	cmd := exec.Command(env.PVMBinary, "init", shell)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func extractPathLine(output, prefix string) string {
	for _, line := range strings.Split(output, "\n") {
		if rest, ok := strings.CutPrefix(line, prefix); ok {
			return rest
		}
	}
	return ""
}

func pathHasFrontEntry(path, entry string) bool {
	parts := strings.Split(path, string(os.PathListSeparator))
	if len(parts) == 0 {
		return false
	}
	return parts[0] == entry
}
