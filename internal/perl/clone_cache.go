// ABOUTME: Bare git clone cache for fork remote repositories
// ABOUTME: Provides CloneCache managing cached bare clones and tag-based version discovery

package perl

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// remotesSubdir is the subdirectory within the cache dir where bare clones are stored.
const remotesSubdir = "remotes"

// CloneCache manages cached bare git clones for fork remote repositories.
// Each remote gets its own bare clone under <cacheDir>/remotes/<remote_name>/.
type CloneCache struct {
	cacheDir string
}

// NewCloneCache creates a CloneCache rooted at cacheDir.
// The caller is responsible for ensuring cacheDir is the PVM cache directory
// (e.g. $XDG_CACHE_HOME/pvm).
func NewCloneCache(cacheDir string) *CloneCache {
	return &CloneCache{cacheDir: cacheDir}
}

// cloneDir returns the path to the bare clone directory for a named remote.
func (cc *CloneCache) cloneDir(remoteName string) string {
	return filepath.Join(cc.cacheDir, remotesSubdir, remoteName)
}

// EnsureClone guarantees a cached bare clone exists for the remote.
// On the first call it runs "git clone --bare --no-single-branch <url> <dir>".
// On subsequent calls (cache already present) it runs "git fetch --tags".
// Returns the path to the bare clone directory.
func (cc *CloneCache) EnsureClone(remote Remote) (string, error) {
	dir := cc.cloneDir(remote.Name)

	// Detect whether the bare clone already exists by checking for HEAD.
	headPath := filepath.Join(dir, "HEAD")
	_, err := os.Stat(headPath)
	if os.IsNotExist(err) {
		return dir, cc.initialClone(remote.URL, dir)
	}
	if err != nil {
		return dir, fmt.Errorf("stat bare clone HEAD %s: %w", headPath, err)
	}

	// Cache exists — fetch latest tags.
	return dir, cc.fetchTags(dir)
}

// initialClone performs "git clone --bare --no-single-branch <url> <destDir>".
func (cc *CloneCache) initialClone(url, destDir string) error {
	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(destDir), 0755); err != nil {
		return fmt.Errorf("create clone parent dir: %w", err)
	}
	cmd := exec.Command("git", "clone", "--bare", "--no-single-branch", url, destDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone --bare %s: %w\n%s", url, err, out)
	}
	return nil
}

// fetchTags runs "git fetch --tags" inside the bare clone directory.
func (cc *CloneCache) fetchTags(cloneDir string) error {
	cmd := exec.Command("git", "fetch", "--tags")
	cmd.Dir = cloneDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch --tags in %s: %w\n%s", cloneDir, err, out)
	}
	return nil
}

// RemoveCache deletes the cached bare clone for the named remote.
// Returns an error if the cache directory does not exist.
func (cc *CloneCache) RemoveCache(remoteName string) error {
	dir := cc.cloneDir(remoteName)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("no cache found for remote %q", remoteName)
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("remove cache for remote %q: %w", remoteName, err)
	}
	return nil
}

// ListTags returns all tag names from the cached bare clone for the named remote.
// The cache must already exist (call EnsureClone first).
func (cc *CloneCache) ListTags(remoteName string) ([]string, error) {
	dir := cc.cloneDir(remoteName)
	cmd := exec.Command("git", "tag", "--list")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git tag --list in %s: %w\n%s", dir, err, out)
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return []string{}, nil
	}

	tags := strings.Split(raw, "\n")
	// Trim any trailing whitespace from each tag line.
	result := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t != "" {
			result = append(result, t)
		}
	}
	return result, nil
}

// CheckoutTag checks out a specific tag from the cached bare clone into destDir.
// destDir must already exist (or will be created). The result is a working tree
// containing the files from that tag's commit.
func (cc *CloneCache) CheckoutTag(remoteName, tag, destDir string) error {
	cloneDir := cc.cloneDir(remoteName)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("create checkout destDir %s: %w", destDir, err)
	}

	// "git --work-tree=<destDir> --git-dir=<bareClone> checkout <tag> -- ."
	// This checks out the tag's tree into destDir from a bare clone.
	cmd := exec.Command(
		"git",
		"--git-dir="+cloneDir,
		"--work-tree="+destDir,
		"checkout", tag, "--", ".",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout %s from %s: %w\n%s", tag, cloneDir, err, out)
	}
	return nil
}

// CheckGitAvailable returns nil if git is present in PATH, or a descriptive error otherwise.
func CheckGitAvailable() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git is required but was not found in PATH: %w", err)
	}
	return nil
}

// DiscoveredVersion represents a perl version found by scanning git tags on a remote.
type DiscoveredVersion struct {
	// Tag is the raw git tag string.
	Tag string
	// Version is the parsed PerlVersion extracted from the tag.
	Version PerlVersion
	// ForkName is the fork name prefix found in the tag (e.g. "myfork" from "myfork-5.40.2").
	// Empty for standard tags like "v5.40.2".
	ForkName string
}

// Tag patterns for perl version tags:
//
//	v5.40.2          — standard upstream tag
//	v5.40.2-1        — release-suffixed upstream tag
//	myfork-5.40.2    — fork-prefixed tag
//
// The patterns below are tried in order. The first match wins.
var (
	// vTagRe matches "v<version>" or "v<version>-<release>" (e.g. v5.40.2-1).
	vTagRe = regexp.MustCompile(`^v(\d+\.\d+(?:\.\d+)?(?:-RC\d+|_\d+)?)(?:-(\d+))?$`)

	// forkTagRe matches "<forkname>-<version>" where forkname starts with a letter.
	// The fork name must be separated from the version by a hyphen followed by a digit.
	forkTagRe = regexp.MustCompile(`^([a-z][a-z0-9-]*?)-(\d+\.\d+(?:\.\d+)?)$`)
)

// parseTagToDiscoveredVersion attempts to parse a git tag into a DiscoveredVersion.
// Returns (dv, true) if the tag contains a recognisable Perl version, (zero, false) otherwise.
func parseTagToDiscoveredVersion(tag string) (DiscoveredVersion, bool) {
	// Try "v<version>" or "v<version>-<N>" form first.
	if m := vTagRe.FindStringSubmatch(tag); m != nil {
		ver, err := ParseVersion(m[1])
		if err != nil {
			return DiscoveredVersion{}, false
		}
		return DiscoveredVersion{
			Tag:      tag,
			Version:  ver,
			ForkName: "",
		}, true
	}

	// Try "<forkname>-<version>" form.
	if m := forkTagRe.FindStringSubmatch(tag); m != nil {
		ver, err := ParseVersion(m[2])
		if err != nil {
			return DiscoveredVersion{}, false
		}
		return DiscoveredVersion{
			Tag:      tag,
			Version:  ver,
			ForkName: m[1],
		}, true
	}

	return DiscoveredVersion{}, false
}

// ListRemoteVersions fetches (or refreshes) the cached bare clone for remote and returns
// all tags that match a Perl version pattern. Non-version tags are silently ignored.
func ListRemoteVersions(remote Remote, cacheDir string) ([]DiscoveredVersion, error) {
	cc := NewCloneCache(cacheDir)

	if _, err := cc.EnsureClone(remote); err != nil {
		return nil, fmt.Errorf("ensure clone for remote %q: %w", remote.Name, err)
	}

	tags, err := cc.ListTags(remote.Name)
	if err != nil {
		return nil, fmt.Errorf("list tags for remote %q: %w", remote.Name, err)
	}

	var versions []DiscoveredVersion
	for _, tag := range tags {
		if dv, ok := parseTagToDiscoveredVersion(tag); ok {
			versions = append(versions, dv)
		}
	}
	return versions, nil
}
