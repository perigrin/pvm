// ABOUTME: GitHub client for fetching documentation with rate limiting and authentication
// ABOUTME: Supports both GitHub API and raw content access with retry logic and caching integration

package docs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/errors"
)

// GitHubDocsClient handles GitHub API and raw content access for documentation
type GitHubDocsClient struct {
	httpClient  *http.Client
	baseURL     string
	token       string
	rateLimiter *rateLimiter
	maxRetries  int
	timeout     time.Duration
}

// GitHubFile represents a file from GitHub Contents API
type GitHubFile struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Size        int    `json:"size"`
	Type        string `json:"type"` // "file" or "dir"
	DownloadURL string `json:"download_url"`
	Content     string `json:"content,omitempty"`
	Encoding    string `json:"encoding,omitempty"`
	SHA         string `json:"sha"`
}

// GitHubDirectory represents a directory listing from GitHub Contents API
type GitHubDirectory []GitHubFile

// rateLimiter implements a simple rate limiter for GitHub API requests
type rateLimiter struct {
	remaining int
	reset     time.Time
	limit     int
}

// NewGitHubDocsClient creates a new GitHub documentation client
func NewGitHubDocsClient(docsConfig *config.DocsConfig) *GitHubDocsClient {
	timeout := 30 * time.Second
	if docsConfig.Timeout != "" {
		if parsedTimeout, err := time.ParseDuration(docsConfig.Timeout); err == nil {
			timeout = parsedTimeout
		}
	}

	client := &GitHubDocsClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: "https://api.github.com",
		rateLimiter: &rateLimiter{
			remaining: 60, // GitHub default unauthenticated rate limit
			reset:     time.Now().Add(time.Hour),
			limit:     60,
		},
		maxRetries: docsConfig.MaxRetries,
		timeout:    timeout,
	}

	// Auto-detect GitHub token
	token := docsConfig.GitHubToken
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token != "" {
		client.token = token
		client.rateLimiter.limit = 5000 // Authenticated rate limit
		client.rateLimiter.remaining = 5000
	}

	return client
}

// GetFile retrieves a single file from GitHub
func (g *GitHubDocsClient) GetFile(ctx context.Context, owner, repo, branch, filePath string) (*GitHubFile, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", g.baseURL, owner, repo, filePath)
	if branch != "" {
		url += "?ref=" + branch
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.NewDocumentationError("001", "Failed to create GitHub API request", err)
	}

	resp, err := g.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.NewDocumentationError("004", fmt.Sprintf("Documentation file not found: %s", filePath), nil)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewDocumentationError("001", fmt.Sprintf("GitHub API request failed: %s", resp.Status), nil)
	}

	var file GitHubFile
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return nil, errors.NewDocumentationError("001", "Failed to decode GitHub API response", err)
	}

	return &file, nil
}

// GetDirectory lists files in a directory from GitHub
func (g *GitHubDocsClient) GetDirectory(ctx context.Context, owner, repo, branch, dirPath string) (GitHubDirectory, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", g.baseURL, owner, repo, dirPath)
	if branch != "" {
		url += "?ref=" + branch
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.NewDocumentationError("001", "Failed to create GitHub API request", err)
	}

	resp, err := g.doRequestWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return GitHubDirectory{}, nil // Empty directory
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewDocumentationError("001", fmt.Sprintf("GitHub API request failed: %s", resp.Status), nil)
	}

	var directory GitHubDirectory
	if err := json.NewDecoder(resp.Body).Decode(&directory); err != nil {
		return nil, errors.NewDocumentationError("001", "Failed to decode GitHub API response", err)
	}

	return directory, nil
}

// GetRawContent fetches raw file content from GitHub (bypasses API rate limits)
func (g *GitHubDocsClient) GetRawContent(ctx context.Context, owner, repo, branch, filePath string) ([]byte, http.Header, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, branch, filePath)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, errors.NewDocumentationError("001", "Failed to create raw content request", err)
	}

	// Add GitHub token if available (for private repos)
	if g.token != "" {
		req.Header.Set("Authorization", "token "+g.token)
	}

	req.Header.Set("User-Agent", "PVM-Docs/1.0")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, nil, errors.NewDocumentationError("001", "Failed to fetch raw content", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil, errors.NewDocumentationError("004", fmt.Sprintf("Documentation file not found: %s", filePath), nil)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, errors.NewDocumentationError("001", fmt.Sprintf("Raw content request failed: %s", resp.Status), nil)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, errors.NewDocumentationError("001", "Failed to read raw content", err)
	}

	return content, resp.Header, nil
}

// GetRawContentWithCache fetches raw content with cache headers for conditional requests
func (g *GitHubDocsClient) GetRawContentWithCache(ctx context.Context, owner, repo, branch, filePath string, cacheHeaders http.Header) ([]byte, http.Header, bool, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, branch, filePath)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, false, errors.NewDocumentationError("001", "Failed to create raw content request", err)
	}

	// Add GitHub token if available
	if g.token != "" {
		req.Header.Set("Authorization", "token "+g.token)
	}

	// Add cache headers for conditional request
	if cacheHeaders != nil {
		if etag := cacheHeaders.Get("If-None-Match"); etag != "" {
			req.Header.Set("If-None-Match", etag)
		}
		if modified := cacheHeaders.Get("If-Modified-Since"); modified != "" {
			req.Header.Set("If-Modified-Since", modified)
		}
	}

	req.Header.Set("User-Agent", "PVM-Docs/1.0")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, nil, false, errors.NewDocumentationError("001", "Failed to fetch raw content", err)
	}
	defer resp.Body.Close()

	// Handle 304 Not Modified
	if resp.StatusCode == http.StatusNotModified {
		return nil, resp.Header, true, nil // Content not modified, use cache
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil, false, errors.NewDocumentationError("004", fmt.Sprintf("Documentation file not found: %s", filePath), nil)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, false, errors.NewDocumentationError("001", fmt.Sprintf("Raw content request failed: %s", resp.Status), nil)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, false, errors.NewDocumentationError("001", "Failed to read raw content", err)
	}

	return content, resp.Header, false, nil
}

// ListDocumentationFiles recursively lists all documentation files in a repository
func (g *GitHubDocsClient) ListDocumentationFiles(ctx context.Context, owner, repo, branch, docsPath string) ([]string, error) {
	var files []string

	err := g.walkDirectory(ctx, owner, repo, branch, docsPath, &files)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// walkDirectory recursively walks a directory tree and collects file paths
func (g *GitHubDocsClient) walkDirectory(ctx context.Context, owner, repo, branch, dirPath string, files *[]string) error {
	directory, err := g.GetDirectory(ctx, owner, repo, branch, dirPath)
	if err != nil {
		return err
	}

	for _, item := range directory {
		if item.Type == "file" {
			// Only include markdown files and other documentation formats
			if g.isDocumentationFile(item.Name) {
				*files = append(*files, item.Path)
			}
		} else if item.Type == "dir" {
			// Recursively walk subdirectories
			if err := g.walkDirectory(ctx, owner, repo, branch, item.Path, files); err != nil {
				return err
			}
		}
	}

	return nil
}

// isDocumentationFile checks if a file is a documentation file
func (g *GitHubDocsClient) isDocumentationFile(filename string) bool {
	ext := strings.ToLower(path.Ext(filename))
	docExts := []string{".md", ".markdown", ".rst", ".txt", ".adoc", ".asciidoc"}

	for _, docExt := range docExts {
		if ext == docExt {
			return true
		}
	}

	return false
}

// doRequestWithRetry executes an HTTP request with retry logic and rate limiting
func (g *GitHubDocsClient) doRequestWithRetry(req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= g.maxRetries; attempt++ {
		// Check rate limit before making request
		if err := g.checkRateLimit(); err != nil {
			return nil, err
		}

		// Add authentication header
		if g.token != "" {
			req.Header.Set("Authorization", "token "+g.token)
		}

		req.Header.Set("User-Agent", "PVM-Docs/1.0")
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := g.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt < g.maxRetries {
				backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
				time.Sleep(backoff)
				continue
			}
			break
		}

		// Update rate limit info from response headers
		g.updateRateLimit(resp)

		// Handle rate limit exceeded
		if resp.StatusCode == http.StatusTooManyRequests {
			resetTime := g.getRateLimitReset(resp)
			if !resetTime.IsZero() && time.Until(resetTime) < 10*time.Minute {
				resp.Body.Close()
				time.Sleep(time.Until(resetTime) + time.Second)
				continue
			}
		}

		// Return successful response or client/server errors that shouldn't be retried
		if resp.StatusCode < 500 {
			return resp, nil
		}

		// Close response body and retry for server errors
		resp.Body.Close()
		lastErr = fmt.Errorf("server error: %s", resp.Status)

		if attempt < g.maxRetries {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			time.Sleep(backoff)
		}
	}

	return nil, errors.NewDocumentationError("003", "GitHub API request failed after retries", lastErr)
}

// checkRateLimit checks if we have remaining API calls
func (g *GitHubDocsClient) checkRateLimit() error {
	if g.rateLimiter.remaining <= 0 && time.Now().Before(g.rateLimiter.reset) {
		waitTime := time.Until(g.rateLimiter.reset)
		if waitTime > 10*time.Minute {
			return errors.NewDocumentationError("003", "GitHub API rate limit exceeded for extended period", nil)
		}
		time.Sleep(waitTime + time.Second)
	}
	return nil
}

// updateRateLimit updates rate limit info from response headers
func (g *GitHubDocsClient) updateRateLimit(resp *http.Response) {
	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		if r, err := strconv.Atoi(remaining); err == nil {
			g.rateLimiter.remaining = r
		}
	}

	if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
		if r, err := strconv.ParseInt(reset, 10, 64); err == nil {
			g.rateLimiter.reset = time.Unix(r, 0)
		}
	}

	if limit := resp.Header.Get("X-RateLimit-Limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			g.rateLimiter.limit = l
		}
	}
}

// getRateLimitReset extracts rate limit reset time from response
func (g *GitHubDocsClient) getRateLimitReset(resp *http.Response) time.Time {
	if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
		if r, err := strconv.ParseInt(reset, 10, 64); err == nil {
			return time.Unix(r, 0)
		}
	}
	return time.Time{}
}

// GetRateLimit returns current rate limit status
func (g *GitHubDocsClient) GetRateLimit() (remaining, limit int, reset time.Time) {
	return g.rateLimiter.remaining, g.rateLimiter.limit, g.rateLimiter.reset
}
