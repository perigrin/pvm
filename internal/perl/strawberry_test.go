// ABOUTME: Tests for Strawberry Perl release discovery via GitHub API
// ABOUTME: Verifies version matching logic and API query behavior

package perl

import (
	"net"
	"testing"
)

// networkAvailable checks if network connectivity is available by attempting
// a short TCP dial to a well-known host.
func networkAvailable() bool {
	conn, err := net.DialTimeout("tcp", "api.github.com:443", networkCheckTimeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// TestParseStrawberryVersion tests the version-matching logic that selects the
// correct Strawberry Perl asset from a release's asset list.  No network
// access is needed.
func TestParseStrawberryVersion(t *testing.T) {
	tests := []struct {
		name        string
		perlVersion string // 3-part Perl version (e.g. "5.38.2")
		assets      []strawberryAsset
		wantURL     string
		wantErr     bool
	}{
		{
			name:        "exact 4-part match",
			perlVersion: "5.38.2",
			assets: []strawberryAsset{
				{Name: "strawberry-perl-5.38.2.2-64bit-portable.zip", BrowserDownloadURL: "https://example.com/5.38.2.2.zip"},
			},
			wantURL: "https://example.com/5.38.2.2.zip",
		},
		{
			name:        "picks highest 4th component when multiple match",
			perlVersion: "5.38.2",
			assets: []strawberryAsset{
				{Name: "strawberry-perl-5.38.2.1-64bit-portable.zip", BrowserDownloadURL: "https://example.com/5.38.2.1.zip"},
				{Name: "strawberry-perl-5.38.2.3-64bit-portable.zip", BrowserDownloadURL: "https://example.com/5.38.2.3.zip"},
				{Name: "strawberry-perl-5.38.2.2-64bit-portable.zip", BrowserDownloadURL: "https://example.com/5.38.2.2.zip"},
			},
			wantURL: "https://example.com/5.38.2.3.zip",
		},
		{
			name:        "ignores non-portable zip assets",
			perlVersion: "5.38.2",
			assets: []strawberryAsset{
				{Name: "strawberry-perl-5.38.2.2-64bit.zip", BrowserDownloadURL: "https://example.com/non-portable.zip"},
				{Name: "strawberry-perl-5.38.2.2-64bit-portable.zip", BrowserDownloadURL: "https://example.com/portable.zip"},
			},
			wantURL: "https://example.com/portable.zip",
		},
		{
			name:        "ignores assets for different 3-part version",
			perlVersion: "5.38.0",
			assets: []strawberryAsset{
				{Name: "strawberry-perl-5.38.2.2-64bit-portable.zip", BrowserDownloadURL: "https://example.com/5.38.2.2.zip"},
			},
			wantErr: true,
		},
		{
			name:        "no assets returns error",
			perlVersion: "5.38.2",
			assets:      []strawberryAsset{},
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := bestStrawberryAsset(tc.perlVersion, tc.assets)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got URL %q", got)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got != tc.wantURL {
				t.Errorf("got URL %q, want %q", got, tc.wantURL)
			}
		})
	}
}

// TestFindStrawberryRelease_NotFound verifies that an unknown version returns
// an error.  This requires network access to the GitHub API.
func TestFindStrawberryRelease_NotFound(t *testing.T) {
	if !networkAvailable() {
		t.Skip("network not available")
	}

	_, err := FindStrawberryRelease("99.99.99")
	if err == nil {
		t.Error("expected error for non-existent version but got nil")
	}
}

// TestFindStrawberryRelease verifies that a real, known version can be
// resolved to a download URL.  This requires network access.
func TestFindStrawberryRelease(t *testing.T) {
	if !networkAvailable() {
		t.Skip("network not available")
	}

	url, err := FindStrawberryRelease("5.38.2")
	if err != nil {
		t.Fatalf("FindStrawberryRelease(\"5.38.2\") returned error: %v", err)
	}
	if url == "" {
		t.Fatal("FindStrawberryRelease returned empty URL")
	}
	// The URL must be a GitHub release download URL containing the version
	// prefix and the portable-zip suffix.
	if !containsAll(url, "5.38.2", "64bit-portable.zip") {
		t.Errorf("URL %q does not look like a portable zip for 5.38.2", url)
	}
}

// containsAll returns true when s contains every substr in parts.
func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		found := false
		for i := 0; i <= len(s)-len(p); i++ {
			if s[i:i+len(p)] == p {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
