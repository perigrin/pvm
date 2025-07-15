// ABOUTME: GitHub API client for fetching PVM release information
// ABOUTME: Handles release checking and asset enumeration for updates

package version

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	ID          int64         `json:"id"`
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	Draft       bool          `json:"draft"`
	Prerelease  bool          `json:"prerelease"`
	CreatedAt   time.Time     `json:"created_at"`
	PublishedAt *time.Time    `json:"published_at"`
	Assets      []GitHubAsset `json:"assets"`
	TarballURL  string        `json:"tarball_url"`
	ZipballURL  string        `json:"zipball_url"`
	HTMLURL     string        `json:"html_url"`
}

// GitHubAsset represents a release asset
type GitHubAsset struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	ContentType        string    `json:"content_type"`
	Size               int64     `json:"size"`
	DownloadCount      int64     `json:"download_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	BrowserDownloadURL string    `json:"browser_download_url"`
}

// GitHubClient handles GitHub API interactions
type GitHubClient struct {
	httpClient *http.Client
	baseURL    string
	token      string // Optional for higher rate limits
}

// NewGitHubClient creates a new GitHub API client with automatic token detection
func NewGitHubClient() *GitHubClient {
	client := &GitHubClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.github.com",
	}

	// Automatically use GITHUB_TOKEN environment variable if available
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		client.token = token
	}

	return client
}

// NewGitHubClientWithToken creates a new GitHub API client with authentication
func NewGitHubClientWithToken(token string) *GitHubClient {
	client := NewGitHubClient()
	client.token = token
	return client
}

// doRequestWithRetry executes an HTTP request with exponential backoff for rate limiting
func (g *GitHubClient) doRequestWithRetry(req *http.Request, maxRetries int) (*http.Response, error) {
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := g.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		// Check for rate limiting (403 with specific message)
		if resp.StatusCode == http.StatusForbidden {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			bodyStr := string(body)

			// Check if this is a rate limit error
			if strings.Contains(bodyStr, "rate limit exceeded") || strings.Contains(bodyStr, "API rate limit") {
				// Check if we can determine retry time from headers
				resetTime := resp.Header.Get("X-RateLimit-Reset")

				// Calculate backoff time
				var backoffTime time.Duration
				if resetTime != "" {
					if resetTimestamp, err := strconv.ParseInt(resetTime, 10, 64); err == nil {
						backoffTime = time.Until(time.Unix(resetTimestamp, 0))
						// Cap backoff at 5 minutes
						if backoffTime > 5*time.Minute {
							backoffTime = 5 * time.Minute
						}
					}
				}

				// Fall back to exponential backoff if no reset time
				if backoffTime <= 0 {
					backoffTime = time.Duration(math.Pow(2, float64(attempt))) * time.Second
				}

				// Don't retry on last attempt
				if attempt == maxRetries-1 {
					return nil, fmt.Errorf("GitHub API rate limit exceeded after %d attempts: %s", maxRetries, bodyStr)
				}

				// Wait before retry
				time.Sleep(backoffTime)
				continue
			}

			// Not a rate limit error, recreate response and return
			return &http.Response{
				StatusCode: resp.StatusCode,
				Header:     resp.Header,
				Body:       io.NopCloser(strings.NewReader(bodyStr)),
			}, nil
		}

		// Success or other error, return as-is
		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded")
}

// GetLatestRelease fetches the latest PVM release for the repository, filtering out non-PVM releases
func (g *GitHubClient) GetLatestRelease(owner, repo string) (*GitHubRelease, error) {
	// Always use GetReleases to avoid confusion with non-PVM releases
	// The /releases/latest endpoint may return Perl binary releases instead of PVM releases
	releases, err := g.GetReleases(owner, repo, true)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found for repository %s/%s", owner, repo)
	}

	// Filter for PVM releases only (exclude Perl binary releases)
	var pvmReleases []GitHubRelease
	for _, release := range releases {
		if g.isPVMRelease(release.TagName) {
			pvmReleases = append(pvmReleases, release)
		}
	}

	if len(pvmReleases) == 0 {
		return nil, fmt.Errorf("no PVM releases found for repository %s/%s", owner, repo)
	}

	// Find the latest PVM release (prefer stable over pre-release)
	var latestStable *GitHubRelease
	var latestPrerelease *GitHubRelease

	for i := range pvmReleases {
		release := &pvmReleases[i]
		if release.Draft {
			continue // Skip draft releases
		}

		if !release.Prerelease {
			// Stable release
			if latestStable == nil || release.CreatedAt.After(latestStable.CreatedAt) {
				latestStable = release
			}
		} else {
			// Pre-release
			if latestPrerelease == nil || release.CreatedAt.After(latestPrerelease.CreatedAt) {
				latestPrerelease = release
			}
		}
	}

	// Prefer stable release, fallback to latest pre-release
	if latestStable != nil {
		return latestStable, nil
	}
	if latestPrerelease != nil {
		return latestPrerelease, nil
	}

	return nil, fmt.Errorf("no published PVM releases found for repository %s/%s", owner, repo)
}

// isPVMRelease determines if a release tag represents a PVM release (not a Perl binary release)
func (g *GitHubClient) isPVMRelease(tagName string) bool {
	// PVM releases typically start with "v" (e.g., "v1.0.0", "v1.0.0-rc22")
	// Exclude Perl binary releases which start with "perl-" (e.g., "perl-5.38.0")
	if strings.HasPrefix(tagName, "perl-") {
		return false
	}
	if strings.HasPrefix(tagName, "v") {
		return true
	}
	// Also accept releases that start with "PVM" for older releases
	if strings.HasPrefix(tagName, "PVM") {
		return true
	}
	// Exclude any other patterns that don't look like PVM releases
	return false
}

// GetReleaseByTag fetches a specific release by tag
func (g *GitHubClient) GetReleaseByTag(owner, repo, tag string) (*GitHubRelease, error) {
	// Ensure tag has v prefix for consistency
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}

	url := fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", g.baseURL, owner, repo, tag)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if g.token != "" {
		req.Header.Set("Authorization", "token "+g.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.doRequestWithRetry(req, 3)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release %s not found", tag)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &release, nil
}

// GetReleases fetches all releases for the repository
func (g *GitHubClient) GetReleases(owner, repo string, includePrerelease bool) ([]GitHubRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases", g.baseURL, owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if g.token != "" {
		req.Header.Set("Authorization", "token "+g.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.doRequestWithRetry(req, 3)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Filter out prereleases if not requested
	if !includePrerelease {
		filtered := make([]GitHubRelease, 0, len(releases))
		for _, release := range releases {
			if !release.Prerelease && !release.Draft {
				filtered = append(filtered, release)
			}
		}
		releases = filtered
	}

	return releases, nil
}

// CreateRelease creates a new GitHub release
func (g *GitHubClient) CreateRelease(owner, repo, tag, name, body string, draft, prerelease bool) (*GitHubRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases", g.baseURL, owner, repo)

	// Prepare release data
	releaseData := map[string]interface{}{
		"tag_name":   tag,
		"name":       name,
		"body":       body,
		"draft":      draft,
		"prerelease": prerelease,
	}

	jsonData, err := json.Marshal(releaseData)
	if err != nil {
		return nil, fmt.Errorf("marshaling release data: %w", err)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if g.token != "" {
		req.Header.Set("Authorization", "token "+g.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnprocessableEntity {
		// Release may already exist, try to get it
		existing, err := g.GetReleaseByTag(owner, repo, tag)
		if err == nil {
			return existing, nil
		}
	}

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &release, nil
}

// UploadReleaseAsset uploads a file as an asset to a GitHub release
func (g *GitHubClient) UploadReleaseAsset(owner, repo string, releaseID int64, filePath, fileName string) (*GitHubAsset, error) {
	// Read the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	// Get file info for size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("getting file info: %w", err)
	}

	// Construct upload URL
	uploadURL := fmt.Sprintf("https://uploads.github.com/repos/%s/%s/releases/%d/assets", owner, repo, releaseID)
	uploadURL += "?name=" + url.QueryEscape(fileName)

	req, err := http.NewRequest("POST", uploadURL, file)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if g.token != "" {
		req.Header.Set("Authorization", "token "+g.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/gzip")
	req.ContentLength = fileInfo.Size()

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}

	var asset GitHubAsset
	if err := json.NewDecoder(resp.Body).Decode(&asset); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &asset, nil
}
