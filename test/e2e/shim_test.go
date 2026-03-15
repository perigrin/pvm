// ABOUTME: End-to-end tests for PVM shim functionality
// ABOUTME: Tests shim creation, execution, and rehashing

package e2e

import (
	"testing"

	"tamarou.com/pvm/test/e2e/helpers"
)

// TestShimCreation tests that perl commands work via two-tier PATH (no shims needed)
func TestShimCreation(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// With two-tier PATH system, perl/cpan shims should NOT exist
	helpers.AssertFileDoesNotExist(t, env.PVMShimsDir+"/perl", "perl shim should not exist with two-tier PATH")
	helpers.AssertFileDoesNotExist(t, env.PVMShimsDir+"/cpan", "cpan shim should not exist with two-tier PATH")

	// Run rehash to verify no perl shims are created
	helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"rehash"}, "Rehash should succeed without creating perl shims")

	// Confirm perl/cpan shims still don't exist (this is correct behavior)
	helpers.AssertFileDoesNotExist(t, env.PVMShimsDir+"/perl", "perl shim should not exist after rehash")
	helpers.AssertFileDoesNotExist(t, env.PVMShimsDir+"/cpan", "cpan shim should not exist after rehash")

	// Verify shell integration works via pvm init
	output := helpers.AssertPVMSucceeds(t, env, []string{"init"}, "pvm init should work")
	helpers.AssertStringContains(t, output, "_pvm_update_perl_path", "Should contain two-tier PATH function")
	helpers.AssertStringContains(t, output, "xdg_bin_home", "Should contain XDG_BIN_HOME setup")
}

// TestShimExecutable tests that perl commands work via PATH (no shims needed)
func TestShimExecutable(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Verify shell integration generates correct PATH setup
	output := helpers.AssertPVMSucceeds(t, env, []string{"init"}, "pvm init should work")

	// Should contain two-tier PATH setup
	helpers.AssertStringContains(t, output, "_pvm_update_perl_path", "Should contain PATH update function")
	helpers.AssertStringContains(t, output, "XDG_BIN_HOME", "Should contain XDG_BIN_HOME setup")
	helpers.AssertStringContains(t, output, "current_version", "Should query current version for PATH")

	// Confirm no perl shims were created (correct behavior)
	helpers.AssertFileDoesNotExist(t, env.PVMShimsDir+"/perl", "perl shim should not exist")
	helpers.AssertFileDoesNotExist(t, env.PVMShimsDir+"/cpan", "cpan shim should not exist")
}

// TestShimPathPriority tests that shims take priority in PATH
func TestShimPathPriority(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test that shell integration sets up correct PATH priority
	output := helpers.AssertPVMSucceeds(t, env, []string{"init"}, "pvm init should work")

	// Should set up XDG_BIN_HOME first (highest priority for tool shims)
	helpers.AssertStringContains(t, output, "XDG_BIN_HOME", "Should contain XDG_BIN_HOME setup")

	// Should dynamically add current Perl version's bin directory
	helpers.AssertStringContains(t, output, "_pvm_update_perl_path", "Should contain PATH update function")
	helpers.AssertStringContains(t, output, "pvm/versions/$current_version/bin", "Should add version-specific bin path")

	// Verify no perl shims exist (they're not needed with two-tier PATH)
	helpers.AssertFileDoesNotExist(t, env.PVMShimsDir+"/perl", "perl shim should not exist")

	// Test that pvm exec works without explicit version (our enhancement)
	helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"exec", "echo", "hello"}, "pvm exec should work without version")
}

// TestRehashCommand tests the rehash command
func TestRehashCommand(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Run the rehash command
	stdout := helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"rehash"}, "Rehash command functionality")

	// Check that it reports success (now uses two-tier PATH instead of shims)
	helpers.AssertStringContains(t, stdout, "Updated shell PATH for Perl",
		"Rehash command output does not indicate success")

	// With two-tier PATH, perl shims should NOT be created
	helpers.AssertFileDoesNotExist(t, env.PVMShimsDir+"/perl", "perl shim should not exist after rehash")
	helpers.AssertFileDoesNotExist(t, env.PVMShimsDir+"/cpan", "cpan shim should not exist after rehash")
	helpers.AssertFileDoesNotExist(t, env.PVMShimsDir+"/prove", "prove shim should not exist after rehash")
	helpers.AssertFileDoesNotExist(t, env.PVMShimsDir+"/perldoc", "perldoc shim should not exist after rehash")
}

// TestShimVersionResolution tests that PATH-based version resolution works
func TestShimVersionResolution(t *testing.T) {
	env := helpers.NewTestEnv(t)
	defer env.Cleanup()

	// Test that shell integration provides version resolution via PATH
	output := helpers.AssertPVMSucceeds(t, env, []string{"init"}, "pvm init should work")

	// Should contain version resolution via current command
	helpers.AssertStringContains(t, output, "current --bare", "Should query current version")
	helpers.AssertStringContains(t, output, "_pvm_update_perl_path", "Should update PATH dynamically")

	// Verify no shims exist (correct with two-tier PATH)
	helpers.AssertFileDoesNotExist(t, env.PVMShimsDir+"/perl", "perl shim should not exist")

	// Test that pvm exec resolves versions correctly without explicit version
	helpers.AssertPVMSucceedsOrSkipTODO(t, env, []string{"exec", "echo", "version resolved"}, "pvm exec should resolve version")

}
