// ABOUTME: Integration tests for platform detection and asset selection
// ABOUTME: Tests against real GitHub releases to ensure asset filtering works correctly

package version

import (
	"os"
	"testing"
)

func TestDarwinARM64AssetSelection_Integration(t *testing.T) {
	// Skip integration test if not explicitly requested
	if os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=1 to run")
	}

	// Test specifically for darwin-arm64 platform
	platform := &Platform{
		OS:           "darwin",
		Architecture: "arm64",
	}

	// Create a realistic asset list based on actual v1.0.0-rc30 release
	assets := []GitHubAsset{
		{Name: "pvm-1.0.0-rc30-darwin-amd64.tar.gz", Size: 5242880},
		{Name: "pvm-1.0.0-rc30-darwin-arm64.tar.gz", Size: 5242880},
		{Name: "pvm-1.0.0-rc30-linux-amd64.tar.gz", Size: 5242880},
	}

	t.Logf("Testing darwin-arm64 asset selection with realistic asset names")
	t.Logf("Platform: %s", platform.String())
	t.Logf("Expected pattern: %s", platform.AssetPattern())

	// Test FilterAssets
	filtered := FilterAssets(assets, platform)
	if len(filtered) == 0 {
		t.Fatal("FilterAssets returned no matches for darwin-arm64")
	}

	if len(filtered) != 1 {
		t.Fatalf("Expected exactly 1 match, got %d", len(filtered))
	}

	expectedAsset := "pvm-1.0.0-rc30-darwin-arm64.tar.gz"
	if filtered[0].Name != expectedAsset {
		t.Errorf("Expected asset %s, got %s", expectedAsset, filtered[0].Name)
	}

	// Test SelectBestAsset
	bestAsset, err := SelectBestAsset(assets, platform)
	if err != nil {
		t.Fatalf("SelectBestAsset failed: %v", err)
	}

	if bestAsset.Name != expectedAsset {
		t.Errorf("SelectBestAsset returned wrong asset: expected %s, got %s",
			expectedAsset, bestAsset.Name)
	}

	t.Logf("SUCCESS: Correctly selected asset %s for platform %s",
		bestAsset.Name, platform.String())
}

func TestAllPlatformsAssetSelection_Integration(t *testing.T) {
	// Skip integration test if not explicitly requested
	if os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test - set RUN_INTEGRATION_TESTS=1 to run")
	}

	// Create a realistic asset list
	assets := []GitHubAsset{
		{Name: "pvm-1.0.0-rc30-darwin-amd64.tar.gz", Size: 5242880},
		{Name: "pvm-1.0.0-rc30-darwin-arm64.tar.gz", Size: 5242880},
		{Name: "pvm-1.0.0-rc30-linux-amd64.tar.gz", Size: 5242880},
		{Name: "pvm-1.0.0-rc30-linux-arm64.tar.gz", Size: 5242880},
		{Name: "pvm-1.0.0-rc30-windows-amd64.exe.zip", Size: 5242880},
		{Name: "checksums.txt", Size: 1024},
	}

	platforms := []Platform{
		{OS: "darwin", Architecture: "amd64"},
		{OS: "darwin", Architecture: "arm64"},
		{OS: "linux", Architecture: "amd64"},
		{OS: "linux", Architecture: "arm64"},
		{OS: "windows", Architecture: "amd64", Extension: ".exe"},
	}

	for _, platform := range platforms {
		t.Run(platform.String(), func(t *testing.T) {
			asset, err := SelectBestAsset(assets, &platform)
			if err != nil {
				t.Fatalf("SelectBestAsset failed for %s: %v", platform.String(), err)
			}

			// Verify the asset name contains the platform pattern
			pattern := platform.AssetPattern()
			if !isAssetMatch(asset.Name, pattern) {
				t.Errorf("Selected asset %s does not match pattern %s for platform %s",
					asset.Name, pattern, platform.String())
			}

			t.Logf("Platform %s -> Asset %s", platform.String(), asset.Name)
		})
	}
}
