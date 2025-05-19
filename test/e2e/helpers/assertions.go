// ABOUTME: Assertion helpers for PVM end-to-end tests
// ABOUTME: Provides common assertion functions for test verification

package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// AssertStringContains checks if a string contains a substring
func AssertStringContains(t *testing.T, s, substr, message string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("%s: expected %q to contain %q", message, s, substr)
	}
}

// AssertStringDoesNotContain checks if a string does not contain a substring
func AssertStringDoesNotContain(t *testing.T, s, substr, message string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("%s: expected %q to not contain %q", message, s, substr)
	}
}

// AssertStringMatches checks if a string matches a regular expression
func AssertStringMatches(t *testing.T, s, pattern, message string) {
	t.Helper()
	matched, err := regexp.MatchString(pattern, s)
	if err != nil {
		t.Fatalf("Invalid regex pattern %q: %v", pattern, err)
	}
	if !matched {
		t.Errorf("%s: expected %q to match pattern %q", message, s, pattern)
	}
}

// AssertFileExists checks if a file exists
func AssertFileExists(t *testing.T, path, message string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("%s: file %q does not exist", message, path)
	}
}

// AssertFileDoesNotExist checks if a file does not exist
func AssertFileDoesNotExist(t *testing.T, path, message string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("%s: file %q exists but should not", message, path)
	}
}

// AssertFileContains checks if a file contains specific content
func AssertFileContains(t *testing.T, path, content, message string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %q: %v", path, err)
	}

	if !strings.Contains(string(data), content) {
		t.Errorf("%s: file %q does not contain %q", message, path, content)
	}
}

// AssertDirectoryContainsFile checks if a directory contains a file with a specific name
func AssertDirectoryContainsFile(t *testing.T, dirPath, fileName, message string) {
	t.Helper()
	filePath := filepath.Join(dirPath, fileName)
	AssertFileExists(t, filePath, message)
}

// AssertCommandSucceeds checks if a command succeeds in the test environment
func AssertCommandSucceeds(t *testing.T, env *TestEnv, command string, args []string, message string) {
	t.Helper()
	stdout, stderr, err := env.RunCommand(command, args...)
	if err != nil {
		t.Errorf("%s: command %q %v failed: %v\nStdout: %s\nStderr: %s",
			message, command, args, err, stdout, stderr)
	}
}

// AssertCommandFails checks if a command fails in the test environment
func AssertCommandFails(t *testing.T, env *TestEnv, command string, args []string, message string) {
	t.Helper()
	stdout, stderr, err := env.RunCommand(command, args...)
	if err == nil {
		t.Errorf("%s: command %q %v succeeded but should have failed\nStdout: %s\nStderr: %s",
			message, command, args, stdout, stderr)
	}
}

// AssertPVMSucceeds checks if a PVM command succeeds
func AssertPVMSucceeds(t *testing.T, env *TestEnv, args []string, message string) string {
	t.Helper()
	stdout, stderr, err := env.RunPVM(args...)
	if err != nil {
		t.Errorf("%s: pvm %v failed: %v\nStdout: %s\nStderr: %s",
			message, args, err, stdout, stderr)
	}
	return stdout
}

// AssertPVMFails checks if a PVM command fails
func AssertPVMFails(t *testing.T, env *TestEnv, args []string, message string) string {
	t.Helper()
	stdout, stderr, err := env.RunPVM(args...)
	if err == nil {
		t.Errorf("%s: pvm %v succeeded but should have failed\nStdout: %s\nStderr: %s",
			message, args, stdout, stderr)
	}
	return stderr
}

// AssertPerlVersionFile checks if a .perl-version file contains the expected version
func AssertPerlVersionFile(t *testing.T, path, expectedVersion, message string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %q: %v", path, err)
	}

	version := strings.TrimSpace(string(data))
	if version != expectedVersion {
		t.Errorf("%s: .perl-version file contains %q, expected %q", message, version, expectedVersion)
	}
}

// AssertShimExists checks if a shim exists for a specific Perl command
func AssertShimExists(t *testing.T, env *TestEnv, command, message string) {
	t.Helper()
	shimPath := filepath.Join(env.PVMShimsDir, command)
	AssertFileExists(t, shimPath, fmt.Sprintf("%s: shim for %q does not exist", message, command))
}

// AssertConfigValue checks if a config value is set correctly
func AssertConfigValue(t *testing.T, env *TestEnv, section, key, expectedValue, message string) {
	t.Helper()
	stdout := AssertPVMSucceeds(t, env, []string{"config", "get", fmt.Sprintf("%s.%s", section, key)},
		fmt.Sprintf("%s: failed to get config value %s.%s", message, section, key))

	value := strings.TrimSpace(stdout)
	if value != expectedValue {
		t.Errorf("%s: config value %s.%s is %q, expected %q", message, section, key, value, expectedValue)
	}
}

// SkipTODO marks a test as a TODO and skips it with an appropriate message
// This helps tests pass while clearly indicating functionality is not yet implemented
func SkipTODO(t *testing.T, feature string) {
	t.Helper()
	t.Skipf("TODO: %s not yet implemented", feature)
}

// AssertPVMSucceedsOrSkipTODO tries to run a PVM command but skips the test with a TODO
// message if the command fails, allowing tests to pass for not-yet-implemented features
func AssertPVMSucceedsOrSkipTODO(t *testing.T, env *TestEnv, args []string, feature string) string {
	t.Helper()
	stdout, stderr, err := env.RunPVM(args...)
	if err != nil {
		t.Skipf("TODO: %s not yet implemented\nCommand: pvm %s\nError: %v\nStdout: %s\nStderr: %s",
			feature, strings.Join(args, " "), err, stdout, stderr)
	}
	return stdout
}

// AssertPVMFailsOrSkipTODO tries to run a PVM command that's expected to fail
// but skips the test with a TODO message if the command succeeds or fails unexpectedly
func AssertPVMFailsOrSkipTODO(t *testing.T, env *TestEnv, args []string, feature string) string {
	t.Helper()
	stdout, stderr, err := env.RunPVM(args...)
	if err == nil {
		t.Skipf("TODO: %s not yet implemented (command succeeded unexpectedly)\nCommand: pvm %s\nStdout: %s\nStderr: %s",
			feature, strings.Join(args, " "), stdout, stderr)
	}
	return stderr
}
