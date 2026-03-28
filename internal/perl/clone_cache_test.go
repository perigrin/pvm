// ABOUTME: Tests for bare git clone caching and tag-based version discovery
// ABOUTME: Verifies CloneCache operations and remote version listing for fork remotes

package perl

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeTestRepo creates a bare git repo at destDir, seeded with the provided tags.
// It initializes a non-bare repo, commits a file, and tags it, then creates a bare
// clone so EnsureClone can fetch from it.
func makeTestRepo(t *testing.T, tags []string) (repoURL string) {
	t.Helper()

	// Work dir for initial commits and tags
	work := t.TempDir()

	// Initialise git repo (git version 2.28+ supports --initial-branch)
	runGit(t, work, "init", "-b", "main")
	runGit(t, work, "config", "user.email", "test@pvm.test")
	runGit(t, work, "config", "user.name", "PVM Test")

	// Commit a placeholder file so we have something to tag
	placeholder := filepath.Join(work, "README")
	require.NoError(t, os.WriteFile(placeholder, []byte("pvm test repo"), 0644))
	runGit(t, work, "add", ".")
	runGit(t, work, "commit", "-m", "initial commit")

	// Create all requested tags
	for _, tag := range tags {
		runGit(t, work, "tag", tag)
	}

	// The URL for cloning is just the local file path
	return work
}

// runGit runs a git command in dir and fails the test on error.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
}

// TestCheckGitAvailable verifies that git is found in PATH during tests.
func TestCheckGitAvailable(t *testing.T) {
	err := CheckGitAvailable()
	assert.NoError(t, err, "git must be available in PATH for these tests to run")
}

// TestNewCloneCache verifies that NewCloneCache stores the cache directory.
func TestNewCloneCache(t *testing.T) {
	cacheDir := t.TempDir()
	cc := NewCloneCache(cacheDir)
	assert.NotNil(t, cc)
	assert.Equal(t, cacheDir, cc.cacheDir)
}

// TestEnsureCloneCreatesBareClon verifies the first access clones the remote as a bare repo.
func TestEnsureCloneCreatesBareClon(t *testing.T) {
	repoURL := makeTestRepo(t, []string{"v5.40.0"})
	cacheDir := t.TempDir()

	remote := Remote{Name: "myremote", URL: repoURL, Type: "git"}
	cc := NewCloneCache(cacheDir)

	clonePath, err := cc.EnsureClone(remote)
	require.NoError(t, err)

	// The clone directory must exist and contain the bare repo marker
	_, err = os.Stat(filepath.Join(clonePath, "HEAD"))
	assert.NoError(t, err, "bare clone must contain HEAD file")
}

// TestEnsureClonePathLocation verifies the bare clone lands in the expected location.
func TestEnsureClonePathLocation(t *testing.T) {
	repoURL := makeTestRepo(t, nil)
	cacheDir := t.TempDir()

	remote := Remote{Name: "acme", URL: repoURL, Type: "git"}
	cc := NewCloneCache(cacheDir)

	clonePath, err := cc.EnsureClone(remote)
	require.NoError(t, err)

	expectedPath := filepath.Join(cacheDir, "remotes", "acme")
	assert.Equal(t, expectedPath, clonePath)
}

// TestEnsureCloneFetchesOnSubsequentCall verifies that a second call fetches (not re-clones).
func TestEnsureCloneFetchesOnSubsequentCall(t *testing.T) {
	repoURL := makeTestRepo(t, []string{"v5.38.0"})
	cacheDir := t.TempDir()

	remote := Remote{Name: "upstream", URL: repoURL, Type: "git"}
	cc := NewCloneCache(cacheDir)

	// First call — clone
	clonePath, err := cc.EnsureClone(remote)
	require.NoError(t, err)

	// Add a new tag to the source repo and verify subsequent EnsureClone sees it
	runGit(t, repoURL, "tag", "v5.40.0")

	// Second call — fetch
	clonePath2, err := cc.EnsureClone(remote)
	require.NoError(t, err)
	assert.Equal(t, clonePath, clonePath2, "path must be stable across calls")

	// The newly created tag must now be visible in the cache
	tags, err := cc.ListTags("upstream")
	require.NoError(t, err)
	assert.Contains(t, tags, "v5.40.0", "fetch must bring new tags into the cache")
}

// TestRemoveCache verifies that the cached clone directory is deleted.
func TestRemoveCache(t *testing.T) {
	repoURL := makeTestRepo(t, nil)
	cacheDir := t.TempDir()

	remote := Remote{Name: "todelete", URL: repoURL, Type: "git"}
	cc := NewCloneCache(cacheDir)

	_, err := cc.EnsureClone(remote)
	require.NoError(t, err)

	// Confirm it exists
	cloneDir := filepath.Join(cacheDir, "remotes", "todelete")
	_, err = os.Stat(cloneDir)
	require.NoError(t, err, "clone must exist before remove")

	// Remove it
	err = cc.RemoveCache("todelete")
	require.NoError(t, err)

	// Confirm it is gone
	_, err = os.Stat(cloneDir)
	assert.True(t, os.IsNotExist(err), "clone directory must be deleted after RemoveCache")
}

// TestRemoveCacheNonExistent verifies RemoveCache returns an error for unknown names.
func TestRemoveCacheNonExistent(t *testing.T) {
	cacheDir := t.TempDir()
	cc := NewCloneCache(cacheDir)
	err := cc.RemoveCache("doesnotexist")
	assert.Error(t, err)
}

// TestListTags verifies that tags in the cached repo are returned.
func TestListTags(t *testing.T) {
	tags := []string{"v5.38.0", "v5.40.0", "v5.40.2"}
	repoURL := makeTestRepo(t, tags)
	cacheDir := t.TempDir()

	remote := Remote{Name: "tagtest", URL: repoURL, Type: "git"}
	cc := NewCloneCache(cacheDir)
	_, err := cc.EnsureClone(remote)
	require.NoError(t, err)

	listed, err := cc.ListTags("tagtest")
	require.NoError(t, err)

	for _, tag := range tags {
		assert.Contains(t, listed, tag, "ListTags must return tag %s", tag)
	}
}

// TestListTagsEmptyRepo verifies ListTags returns an empty slice for a repo with no tags.
func TestListTagsEmptyRepo(t *testing.T) {
	repoURL := makeTestRepo(t, nil)
	cacheDir := t.TempDir()

	remote := Remote{Name: "notagsremote", URL: repoURL, Type: "git"}
	cc := NewCloneCache(cacheDir)
	_, err := cc.EnsureClone(remote)
	require.NoError(t, err)

	listed, err := cc.ListTags("notagsremote")
	require.NoError(t, err)
	assert.Empty(t, listed)
}

// TestCheckoutTag verifies a tagged commit is checked out to a working directory.
func TestCheckoutTag(t *testing.T) {
	repoURL := makeTestRepo(t, []string{"v5.40.2"})
	cacheDir := t.TempDir()

	remote := Remote{Name: "checkouttest", URL: repoURL, Type: "git"}
	cc := NewCloneCache(cacheDir)
	_, err := cc.EnsureClone(remote)
	require.NoError(t, err)

	destDir := t.TempDir()
	err = cc.CheckoutTag("checkouttest", "v5.40.2", destDir)
	require.NoError(t, err)

	// The README committed in makeTestRepo must be present
	_, err = os.Stat(filepath.Join(destDir, "README"))
	assert.NoError(t, err, "checked-out directory must contain committed files")
}

// TestCheckoutTagInvalidTag verifies an error is returned for a non-existent tag.
func TestCheckoutTagInvalidTag(t *testing.T) {
	repoURL := makeTestRepo(t, nil)
	cacheDir := t.TempDir()

	remote := Remote{Name: "badtagremote", URL: repoURL, Type: "git"}
	cc := NewCloneCache(cacheDir)
	_, err := cc.EnsureClone(remote)
	require.NoError(t, err)

	destDir := t.TempDir()
	err = cc.CheckoutTag("badtagremote", "v99.99.99", destDir)
	assert.Error(t, err, "CheckoutTag must fail for non-existent tag")
}

// TestListRemoteVersions verifies that perl-version tags are discovered and non-version tags filtered.
func TestListRemoteVersions(t *testing.T) {
	tags := []string{
		"v5.40.2",        // standard perl tag
		"v5.38.0",        // standard perl tag
		"v5.40.2-1",      // versioned release tag
		"myfork-5.40.2",  // fork-prefixed tag
		"random-tag",     // not a perl version — must be ignored
		"release-v1.2.3", // not a perl version — must be ignored
	}
	repoURL := makeTestRepo(t, tags)
	cacheDir := t.TempDir()

	remote := Remote{Name: "versiontest", URL: repoURL, Type: "git"}

	versions, err := ListRemoteVersions(remote, cacheDir)
	require.NoError(t, err)

	// Build a map for easy lookup
	found := map[string]DiscoveredVersion{}
	for _, dv := range versions {
		found[dv.Tag] = dv
	}

	// Standard tags must be present
	assert.Contains(t, found, "v5.40.2")
	assert.Contains(t, found, "v5.38.0")
	assert.Contains(t, found, "v5.40.2-1")

	// Fork-prefixed tag must be present
	assert.Contains(t, found, "myfork-5.40.2")

	// Non-version tags must be absent
	assert.NotContains(t, found, "random-tag")
	assert.NotContains(t, found, "release-v1.2.3")
}

// TestDiscoveredVersionFields verifies that tag parsing populates DiscoveredVersion correctly.
func TestDiscoveredVersionFields(t *testing.T) {
	tests := []struct {
		tag         string
		wantVersion string
		wantFork    string
	}{
		{"v5.40.2", "5.40.2", ""},
		{"v5.38.0", "5.38.0", ""},
		{"v5.40.2-1", "5.40.2", ""}, // release suffix stripped for version
		{"myfork-5.40.2", "5.40.2", "myfork"},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			dv, ok := parseTagToDiscoveredVersion(tt.tag)
			require.True(t, ok, "tag %q must be recognised as a perl version tag", tt.tag)
			assert.Equal(t, tt.tag, dv.Tag)
			assert.Equal(t, tt.wantVersion, dv.Version.String())
			assert.Equal(t, tt.wantFork, dv.ForkName)
		})
	}
}

// TestParseTagToDiscoveredVersionIgnoresNonVersionTags verifies unrecognised tags return false.
func TestParseTagToDiscoveredVersionIgnoresNonVersionTags(t *testing.T) {
	ignored := []string{
		"random-tag",
		"release-v1.2.3",
		"latest",
		"HEAD",
		"",
	}

	for _, tag := range ignored {
		t.Run(tag, func(t *testing.T) {
			_, ok := parseTagToDiscoveredVersion(tag)
			assert.False(t, ok, "tag %q must not be recognised as a perl version tag", tag)
		})
	}
}
