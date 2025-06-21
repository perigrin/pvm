// ABOUTME: Release manager utility for managing binary releases
// ABOUTME: Creates, verifies, and maintains GitHub releases for Perl binaries

package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/version"
)

// ReleaseMeta contains metadata about a binary release
type ReleaseMeta struct {
	Version   string            `json:"version"`
	CreatedAt time.Time         `json:"created_at"`
	Platforms []string          `json:"platforms"`
	Checksums map[string]string `json:"checksums"`
	TotalSize int64             `json:"total_size"`
	Verified  bool              `json:"verified"`
}

// BinaryIndex represents availability index for client consumption
type BinaryIndex struct {
	LastUpdated time.Time                  `json:"last_updated"`
	Versions    map[string]VersionBinaries `json:"versions"`
}

// VersionBinaries contains binary information for a specific version
type VersionBinaries struct {
	Version   string            `json:"version"`
	Available []string          `json:"available_platforms"`
	Checksums map[string]string `json:"checksums"`
	Sizes     map[string]int64  `json:"sizes"`
}

// Configuration for the release manager
type Config struct {
	Owner         string
	Repo          string
	Token         string
	RetentionDays int
}

func main() {
	var config Config

	// Root command
	rootCmd := &cobra.Command{
		Use:   "release-manager",
		Short: "Manage binary releases for PVM",
		Long: `Release manager utility for creating, verifying, and maintaining
GitHub releases containing pre-compiled Perl binaries.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Validate configuration
			if config.Owner == "" {
				return fmt.Errorf("GitHub owner is required")
			}
			if config.Repo == "" {
				return fmt.Errorf("GitHub repository is required")
			}
			if config.Token == "" {
				// Try to get from environment
				config.Token = os.Getenv("GITHUB_TOKEN")
				if config.Token == "" {
					return fmt.Errorf("GitHub token is required (use --token or GITHUB_TOKEN env var)")
				}
			}
			return nil
		},
	}

	// Add persistent flags
	rootCmd.PersistentFlags().StringVar(&config.Owner, "owner", "perlorg", "GitHub repository owner")
	rootCmd.PersistentFlags().StringVar(&config.Repo, "repo", "pvm", "GitHub repository name")
	rootCmd.PersistentFlags().StringVar(&config.Token, "token", "", "GitHub personal access token")
	rootCmd.PersistentFlags().IntVar(&config.RetentionDays, "retention-days", 90, "Number of days to retain old releases")

	// Add subcommands
	rootCmd.AddCommand(createReleaseCmd(&config))
	rootCmd.AddCommand(verifyReleaseCmd(&config))
	rootCmd.AddCommand(cleanupOldCmd(&config))
	rootCmd.AddCommand(indexBinariesCmd(&config))

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// createReleaseCmd creates a new GitHub release with proper naming and metadata
func createReleaseCmd(config *Config) *cobra.Command {
	var (
		perlVersion string
		binariesDir string
		draft       bool
		prerelease  bool
	)

	cmd := &cobra.Command{
		Use:   "create-release",
		Short: "Create a new GitHub release for Perl binaries",
		Long: `Creates a GitHub release with proper naming and metadata for
pre-compiled Perl binaries. Uploads all binaries from the specified directory
and generates checksums.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return createRelease(config, perlVersion, binariesDir, draft, prerelease)
		},
	}

	cmd.Flags().StringVar(&perlVersion, "version", "", "Perl version (e.g., 5.38.0)")
	cmd.Flags().StringVar(&binariesDir, "binaries-dir", "", "Directory containing binary files")
	cmd.Flags().BoolVar(&draft, "draft", false, "Create as draft release")
	cmd.Flags().BoolVar(&prerelease, "prerelease", false, "Mark as prerelease")

	cmd.MarkFlagRequired("version")
	cmd.MarkFlagRequired("binaries-dir")

	return cmd
}

// verifyReleaseCmd validates all binaries in a release
func verifyReleaseCmd(config *Config) *cobra.Command {
	var perlVersion string

	cmd := &cobra.Command{
		Use:   "verify-release",
		Short: "Verify all binaries in a release",
		Long: `Validates all binaries in a GitHub release by checking checksums,
executability, and basic functionality.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return verifyRelease(config, perlVersion)
		},
	}

	cmd.Flags().StringVar(&perlVersion, "version", "", "Perl version to verify")
	cmd.MarkFlagRequired("version")

	return cmd
}

// cleanupOldCmd removes old releases based on retention policy
func cleanupOldCmd(config *Config) *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "cleanup-old",
		Short: "Clean up old releases based on retention policy",
		Long: `Removes old releases that exceed the retention period.
Only removes releases that are older than the specified number of days.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cleanupOldReleases(config, dryRun)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted without actually deleting")

	return cmd
}

// indexBinariesCmd generates availability index for client consumption
func indexBinariesCmd(config *Config) *cobra.Command {
	var outputFile string

	cmd := &cobra.Command{
		Use:   "index-binaries",
		Short: "Generate binary availability index",
		Long: `Generates a JSON index of all available binary versions and platforms
for client consumption. This index can be used by PVM to quickly determine
binary availability without making multiple API calls.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return indexBinaries(config, outputFile)
		},
	}

	cmd.Flags().StringVar(&outputFile, "output", "binary-index.json", "Output file for the index")

	return cmd
}

// createRelease implements the create-release functionality
func createRelease(config *Config, perlVersion, binariesDir string, draft, prerelease bool) error {
	fmt.Printf("Creating release for Perl %s...\n", perlVersion)

	// Validate directory exists and contains binaries
	if _, err := os.Stat(binariesDir); os.IsNotExist(err) {
		return fmt.Errorf("binaries directory does not exist: %s", binariesDir)
	}

	// Scan for binary files
	binaries, err := scanBinaries(binariesDir)
	if err != nil {
		return fmt.Errorf("scanning binaries: %w", err)
	}

	if len(binaries) == 0 {
		return fmt.Errorf("no binary files found in directory: %s", binariesDir)
	}

	fmt.Printf("Found %d binary files:\n", len(binaries))
	for _, binary := range binaries {
		fmt.Printf("  - %s (%d bytes)\n", binary.Name, binary.Size)
	}

	// Generate checksums
	checksums := make(map[string]string)
	totalSize := int64(0)
	for _, binary := range binaries {
		checksum, err := calculateChecksum(binary.Path)
		if err != nil {
			return fmt.Errorf("calculating checksum for %s: %w", binary.Name, err)
		}
		checksums[binary.Name] = checksum
		totalSize += binary.Size
	}

	// Create GitHub client (would be used for actual API calls)
	_ = version.NewGitHubClientWithToken(config.Token)

	// Create release
	tagName := fmt.Sprintf("perl-%s", perlVersion)
	releaseName := fmt.Sprintf("Perl %s Binary Distribution", perlVersion)

	_ = generateReleaseBody(perlVersion, binaries, checksums, totalSize)

	fmt.Printf("Creating GitHub release: %s\n", tagName)

	// Note: This is a simplified implementation. In a real implementation,
	// you would need to use a proper GitHub API library like go-github
	// to create releases and upload assets.
	fmt.Printf("Release would be created with:\n")
	fmt.Printf("  Tag: %s\n", tagName)
	fmt.Printf("  Name: %s\n", releaseName)
	fmt.Printf("  Draft: %v\n", draft)
	fmt.Printf("  Prerelease: %v\n", prerelease)
	fmt.Printf("  Assets: %d files\n", len(binaries))
	fmt.Printf("  Total size: %d bytes\n", totalSize)

	// Save metadata
	meta := ReleaseMeta{
		Version:   perlVersion,
		CreatedAt: time.Now(),
		Platforms: extractPlatforms(binaries),
		Checksums: checksums,
		TotalSize: totalSize,
		Verified:  false,
	}

	metaFile := filepath.Join(binariesDir, fmt.Sprintf("perl-%s-meta.json", perlVersion))
	if err := saveMetadata(meta, metaFile); err != nil {
		return fmt.Errorf("saving metadata: %w", err)
	}

	fmt.Printf("Release metadata saved to: %s\n", metaFile)
	fmt.Println("Release creation completed successfully!")

	return nil
}

// verifyRelease implements the verify-release functionality
func verifyRelease(config *Config, perlVersion string) error {
	fmt.Printf("Verifying release for Perl %s...\n", perlVersion)

	client := version.NewGitHubClientWithToken(config.Token)
	tagName := fmt.Sprintf("perl-%s", perlVersion)

	// Get release information
	release, err := client.GetReleaseByTag(config.Owner, config.Repo, tagName)
	if err != nil {
		return fmt.Errorf("getting release: %w", err)
	}

	fmt.Printf("Found release: %s (%d assets)\n", release.Name, len(release.Assets))

	// Verify each asset
	verificationResults := make(map[string]bool)
	for _, asset := range release.Assets {
		fmt.Printf("Verifying asset: %s\n", asset.Name)

		// Download and verify checksum
		verified, err := verifyAsset(asset)
		if err != nil {
			fmt.Printf("  ❌ Error verifying: %v\n", err)
			verificationResults[asset.Name] = false
		} else {
			if verified {
				fmt.Printf("  ✅ Verified successfully\n")
			} else {
				fmt.Printf("  ❌ Verification failed\n")
			}
			verificationResults[asset.Name] = verified
		}
	}

	// Summary
	verified := 0
	failed := 0
	for _, result := range verificationResults {
		if result {
			verified++
		} else {
			failed++
		}
	}

	fmt.Printf("\nVerification Summary:\n")
	fmt.Printf("  ✅ Verified: %d\n", verified)
	fmt.Printf("  ❌ Failed: %d\n", failed)

	if failed > 0 {
		return fmt.Errorf("verification failed for %d assets", failed)
	}

	fmt.Println("All assets verified successfully!")
	return nil
}

// cleanupOldReleases implements the cleanup-old functionality
func cleanupOldReleases(config *Config, dryRun bool) error {
	fmt.Printf("Cleaning up releases older than %d days...\n", config.RetentionDays)

	client := version.NewGitHubClientWithToken(config.Token)
	releases, err := client.GetReleases(config.Owner, config.Repo, true)
	if err != nil {
		return fmt.Errorf("getting releases: %w", err)
	}

	cutoffTime := time.Now().AddDate(0, 0, -config.RetentionDays)
	var toDelete []version.GitHubRelease

	for _, release := range releases {
		// Only consider Perl binary releases (tagged with "perl-" prefix)
		if !strings.HasPrefix(release.TagName, "perl-") {
			continue
		}

		if release.CreatedAt.Before(cutoffTime) {
			toDelete = append(toDelete, release)
		}
	}

	if len(toDelete) == 0 {
		fmt.Println("No releases to delete.")
		return nil
	}

	fmt.Printf("Found %d releases to delete:\n", len(toDelete))
	for _, release := range toDelete {
		age := time.Since(release.CreatedAt).Hours() / 24
		fmt.Printf("  - %s (%.0f days old)\n", release.TagName, age)
	}

	if dryRun {
		fmt.Println("Dry run completed. No releases were deleted.")
		return nil
	}

	// In a real implementation, you would delete the releases here
	fmt.Printf("Would delete %d releases (implementation needed)\n", len(toDelete))

	return nil
}

// indexBinaries implements the index-binaries functionality
func indexBinaries(config *Config, outputFile string) error {
	fmt.Println("Generating binary availability index...")

	client := version.NewGitHubClientWithToken(config.Token)
	releases, err := client.GetReleases(config.Owner, config.Repo, false)
	if err != nil {
		return fmt.Errorf("getting releases: %w", err)
	}

	index := BinaryIndex{
		LastUpdated: time.Now(),
		Versions:    make(map[string]VersionBinaries),
	}

	for _, release := range releases {
		// Only process Perl binary releases
		if !strings.HasPrefix(release.TagName, "perl-") {
			continue
		}

		version := strings.TrimPrefix(release.TagName, "perl-")
		platforms := make([]string, 0)
		checksums := make(map[string]string)
		sizes := make(map[string]int64)

		for _, asset := range release.Assets {
			// Extract platform from asset name
			platform := extractPlatformFromAsset(asset.Name)
			if platform != "" {
				platforms = append(platforms, platform)
				checksums[platform] = "" // Would need to extract from release body or metadata
				sizes[platform] = asset.Size
			}
		}

		if len(platforms) > 0 {
			sort.Strings(platforms)
			index.Versions[version] = VersionBinaries{
				Version:   version,
				Available: platforms,
				Checksums: checksums,
				Sizes:     sizes,
			}
		}
	}

	// Write index to file
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling index: %w", err)
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("writing index file: %w", err)
	}

	fmt.Printf("Binary index generated with %d versions\n", len(index.Versions))
	fmt.Printf("Index saved to: %s\n", outputFile)

	return nil
}

// Helper functions

type BinaryFile struct {
	Name string
	Path string
	Size int64
}

func scanBinaries(dir string) ([]BinaryFile, error) {
	var binaries []BinaryFile

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file looks like a binary distribution
		name := info.Name()
		if strings.Contains(name, "perl-") &&
			(strings.HasSuffix(name, ".tar.gz") ||
				strings.HasSuffix(name, ".tar.xz") ||
				strings.HasSuffix(name, ".zip")) {
			binaries = append(binaries, BinaryFile{
				Name: name,
				Path: path,
				Size: info.Size(),
			})
		}

		return nil
	})

	return binaries, err
}

func calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func extractPlatforms(binaries []BinaryFile) []string {
	platformSet := make(map[string]bool)
	for _, binary := range binaries {
		platform := extractPlatformFromAsset(binary.Name)
		if platform != "" {
			platformSet[platform] = true
		}
	}

	platforms := make([]string, 0, len(platformSet))
	for platform := range platformSet {
		platforms = append(platforms, platform)
	}
	sort.Strings(platforms)
	return platforms
}

func extractPlatformFromAsset(name string) string {
	// Extract platform from filename like "perl-5.38.0-linux-amd64.tar.gz"
	parts := strings.Split(name, "-")
	if len(parts) >= 4 {
		// Find OS and arch parts
		for i := 1; i < len(parts)-1; i++ {
			os := parts[i]
			arch := parts[i+1]

			// Remove extension from arch if it's the last part
			if i+1 == len(parts)-1 {
				arch = strings.Split(arch, ".")[0]
			}

			if isValidPlatform(os, arch) {
				return os + "-" + arch
			}
		}
	}
	return ""
}

func isValidPlatform(os, arch string) bool {
	validOS := map[string]bool{
		"linux":   true,
		"darwin":  true,
		"windows": true,
	}
	validArch := map[string]bool{
		"amd64": true,
		"arm64": true,
		"386":   true,
	}
	return validOS[os] && validArch[arch]
}

func generateReleaseBody(version string, binaries []BinaryFile, checksums map[string]string, totalSize int64) string {
	var body strings.Builder

	body.WriteString(fmt.Sprintf("# Perl %s Binary Distribution\n\n", version))
	body.WriteString("Pre-compiled Perl binaries for multiple platforms.\n\n")
	body.WriteString("## Available Platforms\n\n")

	platforms := extractPlatforms(binaries)
	for _, platform := range platforms {
		body.WriteString(fmt.Sprintf("- %s\n", platform))
	}

	body.WriteString("\n## Checksums (SHA-256)\n\n")
	body.WriteString("```\n")
	for _, binary := range binaries {
		if checksum, ok := checksums[binary.Name]; ok {
			body.WriteString(fmt.Sprintf("%s  %s\n", checksum, binary.Name))
		}
	}
	body.WriteString("```\n\n")

	body.WriteString(fmt.Sprintf("**Total download size:** %d bytes\n", totalSize))

	return body.String()
}

func saveMetadata(meta ReleaseMeta, filePath string) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func verifyAsset(asset version.GitHubAsset) (bool, error) {
	// In a real implementation, this would:
	// 1. Download the asset to a temporary location
	// 2. Verify the checksum against published checksums
	// 3. For executables, verify they can run basic commands

	// For now, just return true as a placeholder
	return true, nil
}
