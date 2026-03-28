// ABOUTME: End-to-end tests for PVM fork remote management (pvm remote add/list/remove)
// ABOUTME: Covers full lifecycle and error cases for fork remote sources

package e2e

import (
	"strings"
	"testing"

	"tamarou.com/pvm/test/e2e/helpers"
)

// TestRemoteList_ShowsOrigin verifies that 'pvm remote list' always shows
// the built-in 'origin' remote even when no user-defined remotes exist.
func TestRemoteList_ShowsOrigin(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	stdout, stderr, err := env.RunPVM("remote", "list")
	if err != nil {
		t.Fatalf("pvm remote list failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	helpers.AssertStringContains(t, stdout, "origin", "remote list must always include 'origin'")
}

// TestRemoteAddListRemove_Lifecycle exercises the full add/list/remove lifecycle
// for a user-defined remote.
func TestRemoteAddListRemove_Lifecycle(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Add a remote.
	stdout, stderr, err := env.RunPVM("remote", "add", "mycompany", "https://github.com/mycompany/perl.git")
	if err != nil {
		t.Fatalf("pvm remote add failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}
	output := stdout + stderr
	helpers.AssertStringContains(t, output, "mycompany", "remote add success message should mention remote name")

	// List remotes — both origin and mycompany must appear.
	stdout, stderr, err = env.RunPVM("remote", "list")
	if err != nil {
		t.Fatalf("pvm remote list after add failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}
	helpers.AssertStringContains(t, stdout, "origin", "remote list should include 'origin'")
	helpers.AssertStringContains(t, stdout, "mycompany", "remote list should include added remote")
	helpers.AssertStringContains(t, stdout, "https://github.com/mycompany/perl.git", "remote list should include URL")

	// Remove the remote.
	stdout, stderr, err = env.RunPVM("remote", "remove", "mycompany")
	if err != nil {
		t.Fatalf("pvm remote remove failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}
	output = stdout + stderr
	helpers.AssertStringContains(t, output, "mycompany", "remote remove success message should mention remote name")

	// List remotes — mycompany should no longer appear.
	stdout, stderr, err = env.RunPVM("remote", "list")
	if err != nil {
		t.Fatalf("pvm remote list after remove failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}
	helpers.AssertStringContains(t, stdout, "origin", "remote list should still include 'origin' after remove")
	if strings.Contains(stdout, "mycompany") {
		t.Errorf("remote list should not include 'mycompany' after it was removed, got:\n%s", stdout)
	}
}

// TestRemoteAdd_DuplicateNameIsRejected verifies that adding a remote with the
// same name twice returns an error.
func TestRemoteAdd_DuplicateNameIsRejected(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// First add must succeed.
	stdout, stderr, err := env.RunPVM("remote", "add", "mycompany", "https://github.com/mycompany/perl.git")
	if err != nil {
		t.Fatalf("first pvm remote add failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	// Second add with the same name must fail.
	stdout, stderr, err = env.RunPVM("remote", "add", "mycompany", "https://github.com/mycompany/perl2.git")
	if err == nil {
		t.Errorf("duplicate pvm remote add should have failed but succeeded\nStdout: %s\nStderr: %s", stdout, stderr)
	}
}

// TestRemoteAdd_OriginIsRejected verifies that 'origin' is a reserved name and
// cannot be added as a user-defined remote.
func TestRemoteAdd_OriginIsRejected(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	stdout, stderr, err := env.RunPVM("remote", "add", "origin", "https://example.com/perl.git")
	if err == nil {
		t.Errorf("pvm remote add origin should have failed (origin is reserved) but succeeded\nStdout: %s\nStderr: %s", stdout, stderr)
	}

	errOutput := stdout + stderr
	helpers.AssertStringContains(t, errOutput, "origin", "error output should mention 'origin'")
}

// TestRemoteAdd_InvalidNameIsRejected verifies that remote names not matching
// [a-z0-9][a-z0-9-]* are rejected.
func TestRemoteAdd_InvalidNameIsRejected(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	invalidNames := []string{
		"MyCompany",  // uppercase not allowed
		"my_company", // underscore not allowed
		"-mycompany", // cannot start with dash
		"",           // empty — cobra will catch arg count mismatch
	}

	for _, name := range invalidNames {
		if name == "" {
			// 'pvm remote add' requires exactly 2 args; empty name gives arg-count error.
			stdout, stderr, err := env.RunPVM("remote", "add")
			if err == nil {
				t.Errorf("pvm remote add (no args) should fail but succeeded\nStdout: %s\nStderr: %s", stdout, stderr)
			}
			continue
		}

		stdout, stderr, err := env.RunPVM("remote", "add", name, "https://example.com/perl.git")
		if err == nil {
			t.Errorf("pvm remote add %q should have failed (invalid name) but succeeded\nStdout: %s\nStderr: %s",
				name, stdout, stderr)
		}
	}
}

// TestRemoteRemove_OriginIsRejected verifies that 'origin' cannot be removed.
func TestRemoteRemove_OriginIsRejected(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	stdout, stderr, err := env.RunPVM("remote", "remove", "origin")
	if err == nil {
		t.Errorf("pvm remote remove origin should have failed (origin is reserved) but succeeded\nStdout: %s\nStderr: %s", stdout, stderr)
	}

	errOutput := stdout + stderr
	helpers.AssertStringContains(t, errOutput, "origin", "error output should mention 'origin'")
}

// TestRemoteRemove_UnknownNameIsRejected verifies that removing a remote that
// does not exist returns an error.
func TestRemoteRemove_UnknownNameIsRejected(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	stdout, stderr, err := env.RunPVM("remote", "remove", "doesnotexist")
	if err == nil {
		t.Errorf("pvm remote remove of nonexistent remote should have failed but succeeded\nStdout: %s\nStderr: %s", stdout, stderr)
	}
}

// TestRemoteAdd_MultipleRemotes verifies that multiple distinct remotes can be
// added and all appear in the list output.
func TestRemoteAdd_MultipleRemotes(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	remotes := []struct {
		name string
		url  string
	}{
		{"mycompany", "https://github.com/mycompany/perl.git"},
		{"staging", "https://git.internal/perl-fork.git"},
	}

	for _, r := range remotes {
		stdout, stderr, err := env.RunPVM("remote", "add", r.name, r.url)
		if err != nil {
			t.Fatalf("pvm remote add %s failed: %v\nStdout: %s\nStderr: %s", r.name, err, stdout, stderr)
		}
	}

	stdout, stderr, err := env.RunPVM("remote", "list")
	if err != nil {
		t.Fatalf("pvm remote list failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
	}

	helpers.AssertStringContains(t, stdout, "origin", "list should include origin")
	for _, r := range remotes {
		helpers.AssertStringContains(t, stdout, r.name, "list should include remote "+r.name)
		helpers.AssertStringContains(t, stdout, r.url, "list should include URL for "+r.name)
	}
}

// TestForkInstall_RequiresGitRemote is a placeholder for fork installation tests
// that require a real git remote. These cannot be run in unit test environments
// without a network-accessible git fixture.
//
// TODO: Write a test git fixture server or bare repository so that
// 'pvm install mycompany/myfork@5.40.2' can be exercised end-to-end.
func TestForkInstall_RequiresGitRemote(t *testing.T) {
	t.Skip("TODO: fork install requires a real git remote — needs a test git fixture")
}

// TestForkShellPath_RequiresForkInstall verifies that the shell PATH constructed
// for a fork version (versions/mycompany/myfork-5.40.2/bin) is correct.
//
// TODO: Enable once a test git fixture exists for 'pvm install mycompany/myfork@5.40.2'.
// The shell integration already handles '/' in fork version paths correctly:
// the bash template uses $PVM_DATA/pvm/versions/$current_version/bin, where
// $current_version is the DisplayName such as "mycompany/myfork-5.40.2", which
// produces the correct namespaced path.
func TestForkShellPath_RequiresForkInstall(t *testing.T) {
	t.Skip("TODO: fork shell path test requires a real git remote — needs a test git fixture")
}
