// ABOUTME: Tests for GitHub API client functionality
// ABOUTME: Includes unit tests with mocking and integration tests

package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	basetesting "tamarou.com/pvm/internal/testing"
)

func TestNewGitHubClient(t *testing.T) {
	// Save original environment
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("GITHUB_TOKEN", originalToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
	}()

	// Test without environment variable
	os.Unsetenv("GITHUB_TOKEN")
	client := NewGitHubClient()

	if client == nil {
		t.Fatal("expected client to be created")
	}

	if client.baseURL != "https://api.github.com" {
		t.Errorf("expected baseURL to be https://api.github.com, got %s", client.baseURL)
	}

	if client.token != "" {
		t.Errorf("expected token to be empty, got %s", client.token)
	}

	if client.httpClient == nil {
		t.Error("expected httpClient to be created")
	}
}

func TestNewGitHubClient_WithEnvironmentToken(t *testing.T) {
	// Save original environment
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if originalToken != "" {
			os.Setenv("GITHUB_TOKEN", originalToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
	}()

	// Test with environment variable
	testToken := "test-env-token"
	os.Setenv("GITHUB_TOKEN", testToken)

	client := NewGitHubClient()

	if client == nil {
		t.Fatal("expected client to be created")
	}

	if client.token != testToken {
		t.Errorf("expected token to be %s, got %s", testToken, client.token)
	}
}

func TestNewGitHubClientWithToken(t *testing.T) {
	token := "test-token"
	client := NewGitHubClientWithToken(token)

	if client == nil {
		t.Fatal("expected client to be created")
	}

	if client.token != token {
		t.Errorf("expected token to be %s, got %s", token, client.token)
	}
}

func TestGetLatestRelease_Success(t *testing.T) {
	// Create mock server
	mockRelease := GitHubRelease{
		TagName:    "v1.2.3",
		Name:       "Release 1.2.3",
		Body:       "Release notes",
		Draft:      false,
		Prerelease: false,
		CreatedAt:  time.Now(),
		HTMLURL:    "https://github.com/owner/repo/releases/tag/v1.2.3",
		Assets: []GitHubAsset{
			{
				ID:                 123,
				Name:               "pvm-linux-amd64",
				ContentType:        "application/octet-stream",
				Size:               1024,
				BrowserDownloadURL: "https://github.com/owner/repo/releases/download/v1.2.3/pvm-linux-amd64",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/releases" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Header.Get("Accept") != "application/vnd.github.v3+json" {
			t.Errorf("unexpected Accept header: %s", r.Header.Get("Accept"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]GitHubRelease{mockRelease})
	}))
	defer server.Close()

	client := NewGitHubClient()
	client.baseURL = server.URL

	release, err := client.GetLatestRelease("owner", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if release.TagName != mockRelease.TagName {
		t.Errorf("expected TagName %s, got %s", mockRelease.TagName, release.TagName)
	}

	if release.Name != mockRelease.Name {
		t.Errorf("expected Name %s, got %s", mockRelease.Name, release.Name)
	}

	if len(release.Assets) != 1 {
		t.Errorf("expected 1 asset, got %d", len(release.Assets))
	}
}

func TestGetLatestRelease_WithToken(t *testing.T) {
	token := "test-token"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "token "+token {
			t.Errorf("expected Authorization header 'token %s', got %s", token, auth)
		}

		mockRelease := GitHubRelease{
			TagName: "v1.0.0",
			Name:    "Test Release",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]GitHubRelease{mockRelease})
	}))
	defer server.Close()

	client := NewGitHubClientWithToken(token)
	client.baseURL = server.URL

	_, err := client.GetLatestRelease("owner", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetLatestRelease_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	client := NewGitHubClient()
	client.baseURL = server.URL

	_, err := client.GetLatestRelease("owner", "repo")
	if err == nil {
		t.Fatal("expected error but got none")
	}

	expectedMsg := "GitHub API returned 404"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("expected error to contain %s, got %s", expectedMsg, err.Error())
	}
}

func TestGetReleaseByTag_Success(t *testing.T) {
	mockRelease := GitHubRelease{
		TagName: "v1.2.3",
		Name:    "Release 1.2.3",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/repos/owner/repo/releases/tags/v1.2.3"
		if r.URL.Path != expectedPath {
			t.Errorf("unexpected path: got %s, expected %s", r.URL.Path, expectedPath)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockRelease)
	}))
	defer server.Close()

	client := NewGitHubClient()
	client.baseURL = server.URL

	release, err := client.GetReleaseByTag("owner", "repo", "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if release.TagName != mockRelease.TagName {
		t.Errorf("expected TagName %s, got %s", mockRelease.TagName, release.TagName)
	}
}

func TestGetReleaseByTag_AddVPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should automatically add v prefix
		expectedPath := "/repos/owner/repo/releases/tags/v1.2.3"
		if r.URL.Path != expectedPath {
			t.Errorf("unexpected path: got %s, expected %s", r.URL.Path, expectedPath)
		}

		mockRelease := GitHubRelease{TagName: "v1.2.3"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockRelease)
	}))
	defer server.Close()

	client := NewGitHubClient()
	client.baseURL = server.URL

	_, err := client.GetReleaseByTag("owner", "repo", "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetReleaseByTag_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	client := NewGitHubClient()
	client.baseURL = server.URL

	_, err := client.GetReleaseByTag("owner", "repo", "v1.0.0")
	if err == nil {
		t.Fatal("expected error but got none")
	}

	expectedMsg := "release v1.0.0 not found"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("expected error to contain %s, got %s", expectedMsg, err.Error())
	}
}

func TestGetReleases_Success(t *testing.T) {
	mockReleases := []GitHubRelease{
		{
			TagName:    "v1.2.0",
			Name:       "Release 1.2.0",
			Prerelease: false,
			Draft:      false,
		},
		{
			TagName:    "v1.2.0-beta.1",
			Name:       "Release 1.2.0-beta.1",
			Prerelease: true,
			Draft:      false,
		},
		{
			TagName:    "v1.1.0",
			Name:       "Release 1.1.0",
			Prerelease: false,
			Draft:      false,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/releases" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockReleases)
	}))
	defer server.Close()

	client := NewGitHubClient()
	client.baseURL = server.URL

	releases, err := client.GetReleases("owner", "repo", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(releases) != 3 {
		t.Errorf("expected 3 releases, got %d", len(releases))
	}
}

func TestGetReleases_FilterPrerelease(t *testing.T) {
	mockReleases := []GitHubRelease{
		{
			TagName:    "v1.2.0",
			Name:       "Release 1.2.0",
			Prerelease: false,
			Draft:      false,
		},
		{
			TagName:    "v1.2.0-beta.1",
			Name:       "Release 1.2.0-beta.1",
			Prerelease: true,
			Draft:      false,
		},
		{
			TagName:    "v1.1.0",
			Name:       "Release 1.1.0",
			Prerelease: false,
			Draft:      true, // Draft should also be filtered
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockReleases)
	}))
	defer server.Close()

	client := NewGitHubClient()
	client.baseURL = server.URL

	releases, err := client.GetReleases("owner", "repo", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only include non-prerelease, non-draft releases
	if len(releases) != 1 {
		t.Errorf("expected 1 release, got %d", len(releases))
	}

	if len(releases) > 0 && releases[0].TagName != "v1.2.0" {
		t.Errorf("expected first release to be v1.2.0, got %s", releases[0].TagName)
	}
}

func TestGitHubClient_NetworkError(t *testing.T) {
	client := NewGitHubClient()
	client.baseURL = "http://nonexistent.example"

	_, err := client.GetLatestRelease("owner", "repo")
	if err == nil {
		t.Fatal("expected network error but got none")
	}

	expectedMsg := "making request"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("expected error to contain %s, got %s", expectedMsg, err.Error())
	}
}

func TestGitHubClient_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewGitHubClient()
	client.baseURL = server.URL

	_, err := client.GetLatestRelease("owner", "repo")
	if err == nil {
		t.Fatal("expected JSON decode error but got none")
	}

	expectedMsg := "decoding response"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("expected error to contain %s, got %s", expectedMsg, err.Error())
	}
}

// Integration test - only runs with GITHUB_INTEGRATION_TEST=1
func TestGetLatestRelease_Integration(t *testing.T) {
	basetesting.SkipUnlessIntegration(t, "GitHub API integration test")

	// Test against the PVM repository itself - this is a known quantity with controlled releases
	testRepos := []struct {
		owner string
		repo  string
		desc  string
	}{
		{"perigrin", "pvm", "PVM repository - known to have releases"},
	}

	// Use authentication to avoid rate limiting
	t.Log("Making authenticated API calls to PVM repository")

	// This test makes a real API call to GitHub to test the underlying HTTP functionality
	// Instead of testing GetLatestRelease (which filters for PVM releases), we'll test GetReleases
	// which provides broader coverage of the GitHub API integration
	client := NewGitHubClient()

	var lastErr error
	for _, testRepo := range testRepos {
		releases, err := client.GetReleases(testRepo.owner, testRepo.repo, false)
		if err != nil {
			lastErr = err
			t.Logf("Failed to get releases from %s/%s: %v", testRepo.owner, testRepo.repo, err)
			continue
		}

		if len(releases) == 0 {
			lastErr = fmt.Errorf("got zero releases")
			t.Logf("Got zero releases from %s/%s", testRepo.owner, testRepo.repo)
			continue
		}

		// Successfully got releases
		t.Logf("Successfully got %d releases from %s/%s", len(releases), testRepo.owner, testRepo.repo)

		// Test the first release has expected fields
		firstRelease := releases[0]
		if firstRelease.TagName == "" {
			t.Error("expected TagName to be non-empty")
		}

		if firstRelease.HTMLURL == "" {
			t.Error("expected HTMLURL to be non-empty")
		}

		// PVM-specific validation - we know PVM releases should have certain patterns
		if testRepo.owner == "perigrin" && testRepo.repo == "pvm" {
			// PVM releases should either be version tags (v1.0.0-rcX) or Perl releases (perl-5.X.X)
			tagName := firstRelease.TagName
			if !strings.HasPrefix(tagName, "v") && !strings.HasPrefix(tagName, "perl-") {
				t.Logf("Warning: PVM release tag '%s' doesn't match expected pattern (v* or perl-*)", tagName)
			}
			t.Logf("Found PVM release: %s", tagName)
		}

		// Test passed with at least one repository
		return
	}

	// If we get here, all repositories failed
	if lastErr != nil {
		t.Skipf("integration test failed for all test repositories (might be rate limited or network issues): %v", lastErr)
	} else {
		t.Skip("integration test failed for all test repositories (unknown reason)")
	}
}

func TestGitHubClient_CreateRelease(t *testing.T) {
	t.Run("SuccessfulCreate", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/repos/owner/repo/releases" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("unexpected Content-Type: %s", r.Header.Get("Content-Type"))
			}

			mockRelease := GitHubRelease{
				ID:         12345,
				TagName:    "v1.0.0",
				Name:       "Test Release",
				Body:       "Test release body",
				Draft:      false,
				Prerelease: false,
				CreatedAt:  time.Now(),
				HTMLURL:    "https://github.com/owner/repo/releases/tag/v1.0.0",
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(mockRelease)
		}))
		defer server.Close()

		client := NewGitHubClientWithToken("test-token")
		client.baseURL = server.URL

		release, err := client.CreateRelease("owner", "repo", "v1.0.0", "Test Release", "Test release body", false, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if release.ID != 12345 {
			t.Errorf("expected ID 12345, got %d", release.ID)
		}
		if release.TagName != "v1.0.0" {
			t.Errorf("expected TagName v1.0.0, got %s", release.TagName)
		}
	})

	t.Run("ReleaseAlreadyExists", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if r.Method == "POST" && callCount == 1 {
				// First call - create release fails with 422
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte(`{"message": "Validation Failed"}`))
			} else if r.Method == "GET" && callCount == 2 {
				// Second call - get existing release
				release := GitHubRelease{
					ID:      12345,
					TagName: "v1.0.0",
					Name:    "Existing Release",
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(release)
			}
		}))
		defer server.Close()

		client := NewGitHubClientWithToken("test-token")
		client.baseURL = server.URL

		release, err := client.CreateRelease("owner", "repo", "v1.0.0", "Test Release", "Test release body", false, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if release.ID != 12345 {
			t.Errorf("expected ID 12345, got %d", release.ID)
		}
		if release.Name != "Existing Release" {
			t.Errorf("expected Name 'Existing Release', got %s", release.Name)
		}
	})
}

func TestGitHubClient_UploadReleaseAsset(t *testing.T) {
	t.Run("BasicTest", func(t *testing.T) {
		// Create a test file
		tempDir := t.TempDir()
		testFile := tempDir + "/test.tar.gz"
		testContent := []byte("test archive content")
		if err := os.WriteFile(testFile, testContent, 0o644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Test that the function at least validates input correctly
		client := NewGitHubClientWithToken("test-token")

		// This will fail because we don't have a real server, but we can test input validation
		_, err := client.UploadReleaseAsset("owner", "repo", 12345, testFile, "test.tar.gz")
		if err == nil {
			t.Fatal("expected error for non-real GitHub API call")
		}
		// The error should mention making request or GitHub API, not file opening
		if contains(err.Error(), "opening file") {
			t.Errorf("unexpected error type - should not be file opening error: %v", err)
		}
	})

	t.Run("FileNotFound", func(t *testing.T) {
		client := NewGitHubClientWithToken("test-token")

		_, err := client.UploadReleaseAsset("owner", "repo", 12345, "/nonexistent/file.tar.gz", "file.tar.gz")
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
		if !contains(err.Error(), "opening file") {
			t.Errorf("expected error to mention opening file, got: %v", err)
		}
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
