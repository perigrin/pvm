// ABOUTME: Fork install flow for the pvm install command
// ABOUTME: Handles the "pvm install remote/fork@version" path via git clone cache

package pvm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/perl"
	"tamarou.com/pvm/internal/xdg"
)

// isForkInstall returns true when the version string identifies a fork install.
// Fork identifiers contain a "/" separating the remote name from the rest.
func isForkInstall(version string) bool {
	return strings.Contains(version, "/")
}

// findRemoteInConfig looks up a remote by name in the PVM configuration.
// Returns the matching PVMRemoteConfig and true when found, or zero-value and false otherwise.
func findRemoteInConfig(cfg *config.Config, name string) (config.PVMRemoteConfig, bool) {
	if cfg.PVM == nil {
		return config.PVMRemoteConfig{}, false
	}
	for _, r := range cfg.PVM.Remotes {
		if r.Name == name {
			return r, true
		}
	}
	return config.PVMRemoteConfig{}, false
}

// matchTagForVersion finds the git tag in tags that matches the fork identifier.
// For a fork with a ForkName it looks for "<forkname>-<baseversion>".
// For a remote-only (no ForkName) install it looks for "v<baseversion>".
// Returns the matching tag and true when found, or empty string and false otherwise.
func matchTagForVersion(tags []string, fi *perl.ForkIdentifier) (string, bool) {
	var candidate string
	if fi.ForkName != "" {
		candidate = fi.ForkName + "-" + fi.BaseVersion
	} else {
		candidate = "v" + fi.BaseVersion
	}
	for _, t := range tags {
		if t == candidate {
			return t, true
		}
	}
	return "", false
}

// forkInstallSubpath returns the path segment under the versions directory for a fork install.
// The format is "<remote>/<forkname>-<version>" or "<remote>/<version>" when no fork name.
func forkInstallSubpath(fi *perl.ForkIdentifier) string {
	if fi.ForkName != "" {
		return fi.Remote + "/" + fi.ForkName + "-" + fi.BaseVersion
	}
	return fi.Remote + "/" + fi.BaseVersion
}

// buildVersionStringForFork returns the version string used when registering a fork install.
// When a fork name is present the format is "<forkname>-<baseversion>"; otherwise the bare
// base version is used.
func buildVersionStringForFork(fi *perl.ForkIdentifier) string {
	if fi.ForkName != "" {
		return fi.ForkName + "-" + fi.BaseVersion
	}
	return fi.BaseVersion
}

// installFork orchestrates the full fork install flow for a parsed ForkIdentifier.
// It validates the remote, clones or fetches the repo, checks out the requested tag into a
// temporary directory, optionally reads the .pvm-fork.toml manifest, builds Perl from that
// source, and registers the result in the version registry.
func installFork(cmd *cobra.Command, fi *perl.ForkIdentifier, cfg *config.Config) error {
	ui := cli.GetUI(cmd)

	// Step 1: Look up the remote in configuration.
	remoteCfg, found := findRemoteInConfig(cfg, fi.Remote)
	if !found {
		return fmt.Errorf("remote %q not configured. Run 'pvm remote add %s <url>' first.", fi.Remote, fi.Remote)
	}

	// Step 2: Verify git is available.
	if err := perl.CheckGitAvailable(); err != nil {
		return fmt.Errorf("git is required for fork installation: %w", err)
	}

	// Step 3: Get XDG directories for cache and install paths.
	dirs, err := xdg.GetDirs()
	if err != nil {
		return fmt.Errorf("failed to get XDG directories: %w", err)
	}

	remote := perl.Remote{
		Name: remoteCfg.Name,
		URL:  remoteCfg.URL,
		Type: remoteCfg.Type,
	}

	// Step 4: Clone or fetch the remote repository into the cache.
	ui.Info("Fetching fork repository '%s'...", fi.Remote)
	cc := perl.NewCloneCache(dirs.CacheDir)
	if _, err := cc.EnsureClone(remote); err != nil {
		return fmt.Errorf("failed to fetch remote %q: %w", fi.Remote, err)
	}

	// Step 5: List tags and find the matching one for the requested version.
	tags, err := cc.ListTags(fi.Remote)
	if err != nil {
		return fmt.Errorf("failed to list tags for remote %q: %w", fi.Remote, err)
	}

	tag, found := matchTagForVersion(tags, fi)
	if !found {
		// Build a helpful error message showing available tags.
		return buildNoTagError(fi, tags)
	}

	// Step 6: Check out the tag into a temporary directory.
	checkoutDir, err := os.MkdirTemp("", "pvm-fork-checkout-*")
	if err != nil {
		return fmt.Errorf("failed to create checkout directory: %w", err)
	}
	defer os.RemoveAll(checkoutDir)

	ui.Info("Checking out tag '%s'...", tag)
	if err := cc.CheckoutTag(fi.Remote, tag, checkoutDir); err != nil {
		return fmt.Errorf("failed to check out tag %q: %w", tag, err)
	}

	// Step 7: Read the optional .pvm-fork.toml manifest.
	manifest, err := perl.ParseForkManifest(checkoutDir)
	if err != nil {
		return fmt.Errorf("failed to read fork manifest: %w", err)
	}
	if manifest != nil {
		ui.Info("Fork: %s — %s", manifest.Name, manifest.Description)
	}

	// Step 8: Determine the install directory under the versions directory.
	installDir := filepath.Join(dirs.VersionsDir, forkInstallSubpath(fi))

	// Step 9: Build perl from the checked-out source.
	buildJobs, _ := cmd.Flags().GetInt("jobs")
	runTests, _ := cmd.Flags().GetBool("test")
	noPatchPerl, _ := cmd.Flags().GetBool("no-patchperl")

	var configureFlags []string
	if manifest != nil {
		configureFlags = manifest.ConfigureFlags
	}

	ui.Info("Building Perl %s from fork source...", fi.BaseVersion)

	var currentStage perl.BuildProgressStage
	progressCallback := func(stage perl.BuildProgressStage, details string, progress float64) {
		if stage != currentStage {
			ui.Header(stage.String())
			currentStage = stage
		}
		if details != "" {
			if stage == perl.StageCompile || stage == perl.StageTest {
				if strings.Contains(details, "ERROR") ||
					strings.Contains(details, "WARNING") ||
					strings.Contains(details, "warning:") ||
					strings.Contains(details, "error:") ||
					strings.Contains(details, "Done") ||
					strings.Contains(details, "All tests successful") {
					logCompileDetails(ui, details)
				}
			} else {
				ui.Info("%s", details)
			}
		}
	}

	buildOptions := &perl.BuildOptions{
		Version:          fi.BaseVersion,
		InstallDir:       installDir,
		BuildJobs:        buildJobs,
		RunTests:         runTests,
		CleanupBuildDir:  true,
		ProgressCallback: progressCallback,
		Context:          cmd.Context(),
		NoPatchPerl:      noPatchPerl,
		SourceDir:        checkoutDir,
	}

	// Apply configure flags from manifest if present.
	if len(configureFlags) > 0 {
		buildOptions.ConfigureOptions = configureFlags
	}

	result, err := perl.BuildPerl(buildOptions)
	if err != nil {
		return fmt.Errorf("fork build failed: %w", err)
	}

	// Step 10: Register the fork installation in the version registry.
	versionStr := buildVersionStringForFork(fi)
	versionInfo := perl.VersionInfo{
		Version:     versionStr,
		InstallPath: result.InstallPath,
		InstallTime: time.Now(),
		Source:      "pvm",
		Remote:      fi.Remote,
		ForkName:    fi.ForkName,
		BaseVersion: fi.BaseVersion,
	}
	if err := perl.RegisterVersion(versionInfo); err != nil {
		return fmt.Errorf("failed to register fork installation: %w", err)
	}

	ui.Success("Fork installation completed successfully!")
	ui.Info("Perl %s (%s) installed at: %s", fi.BaseVersion, fi.DisplayName(), result.InstallPath)
	ui.Info("Total installation time: %s", result.Duration.Round(time.Second))

	return nil
}

// buildNoTagError constructs a descriptive error message when no matching tag is found.
// It lists the tags discovered on the remote to help the user correct the version string.
func buildNoTagError(fi *perl.ForkIdentifier, tags []string) error {
	var wanted string
	if fi.ForkName != "" {
		wanted = fi.ForkName + "-" + fi.BaseVersion
	} else {
		wanted = "v" + fi.BaseVersion
	}

	// Filter to show only tags that look like version tags (simple heuristic).
	var versionTags []string
	for _, t := range tags {
		if strings.HasPrefix(t, "v") || (len(t) > 0 && t[0] >= 'a' && t[0] <= 'z') {
			versionTags = append(versionTags, t)
		}
	}

	if len(versionTags) == 0 {
		return fmt.Errorf("no tag %q found in remote %q (remote has no version tags)", wanted, fi.Remote)
	}

	available := strings.Join(versionTags, ", ")
	return fmt.Errorf("no tag %q found in remote %q. Available tags: %s", wanted, fi.Remote, available)
}
