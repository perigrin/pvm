// ABOUTME: Strawberry Perl release discovery via the GitHub releases API
// ABOUTME: Provides FindStrawberryRelease to resolve a 3-part Perl version to a portable zip download URL

package perl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// networkCheckTimeout is the duration used when probing network reachability
// before making GitHub API calls.
const networkCheckTimeout = 2 * time.Second

// strawberryAsset represents a single asset attached to a GitHub release.
type strawberryAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// strawberryRelease is a minimal representation of a GitHub release object.
type strawberryRelease struct {
	TagName string            `json:"tag_name"`
	Assets  []strawberryAsset `json:"assets"`
}

// bestStrawberryAsset selects the portable 64-bit zip from assets that matches
// the given 3-part perlVersion (e.g. "5.38.2").  When multiple assets match
// (differing only in the 4th version component) the one with the highest 4th
// component is returned.  An error is returned when no matching asset is found.
func bestStrawberryAsset(perlVersion string, assets []strawberryAsset) (string, error) {
	// We need a 3-part version prefix to compare against asset names.
	prefix := "strawberry-perl-" + perlVersion + "."
	suffix := "-64bit-portable.zip"

	bestURL := ""
	bestFourth := -1

	for _, a := range assets {
		name := a.Name
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, suffix) {
			continue
		}
		// Extract the 4th version component from the name.
		// name looks like: strawberry-perl-5.38.2.2-64bit-portable.zip
		inner := name[len(prefix) : len(name)-len(suffix)]
		// inner should be the 4th component followed by nothing else
		// (it may contain a single integer).
		fourth, err := strconv.Atoi(inner)
		if err != nil {
			continue
		}
		if fourth > bestFourth {
			bestFourth = fourth
			bestURL = a.BrowserDownloadURL
		}
	}

	if bestURL == "" {
		return "", fmt.Errorf("no Strawberry Perl portable zip found for version %s", perlVersion)
	}
	return bestURL, nil
}

// FindStrawberryRelease queries the GitHub releases API for the StrawberryPerl
// repository and returns the download URL of the 64-bit portable zip for the
// given 3-part Perl version (e.g. "5.38.2").
func FindStrawberryRelease(perlVersion string) (string, error) {
	// Fetch up to 100 releases to cover older Perl versions
	const apiURL = "https://api.github.com/repos/StrawberryPerl/Perl-Dist-Strawberry/releases?per_page=100"

	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("strawberry: create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	// Support GitHub auth tokens for higher rate limits (e.g. in CI).
	// Check GH_TOKEN first (matches binary.go convention), then GITHUB_TOKEN.
	token := os.Getenv("GH_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("strawberry: github API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("strawberry: github API returned status %d", resp.StatusCode)
	}

	var releases []strawberryRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", fmt.Errorf("strawberry: decode releases: %w", err)
	}

	for _, rel := range releases {
		url, err := bestStrawberryAsset(perlVersion, rel.Assets)
		if err == nil {
			return url, nil
		}
	}

	return "", fmt.Errorf("strawberry: no release found for Perl %s", perlVersion)
}
