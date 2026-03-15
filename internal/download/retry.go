// ABOUTME: Retry logic and error recovery for download operations
// ABOUTME: Implements exponential backoff and handles transient network errors

package download

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries      int           // Maximum number of retries (0 = no retries)
	InitialDelay    time.Duration // Initial delay between retries
	MaxDelay        time.Duration // Maximum delay between retries
	BackoffFactor   float64       // Multiplier for exponential backoff
	RetryableErrors []string      // Error patterns that should trigger retries
}

// DefaultRetryPolicy returns a sensible default retry policy
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []string{
			"timeout",
			"connection refused",
			"connection reset",
			"temporary failure",
			"network is unreachable",
			"no route to host",
		},
	}
}

// ShouldRetry determines if an error should trigger a retry
func (rp *RetryPolicy) ShouldRetry(err error, attempt int) bool {
	if attempt >= rp.MaxRetries {
		return false
	}

	// Don't retry context cancellation
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	// Check for specific retryable errors
	errStr := err.Error()
	for _, pattern := range rp.RetryableErrors {
		if containsIgnoreCase(errStr, pattern) {
			return true
		}
	}

	// Check for specific network errors
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout() || netErr.Temporary()
	}

	// Check for HTTP errors that might be retryable
	if urlErr, ok := err.(*url.Error); ok {
		if netErr, ok := urlErr.Err.(net.Error); ok {
			return netErr.Timeout() || netErr.Temporary()
		}
	}

	// Check for specific HTTP status codes that are retryable
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.IsRetryable()
	}

	return false
}

// CalculateDelay calculates the delay for a given retry attempt
func (rp *RetryPolicy) CalculateDelay(attempt int) time.Duration {
	if attempt == 0 {
		return 0
	}

	delay := float64(rp.InitialDelay)
	for i := 1; i < attempt; i++ {
		delay *= rp.BackoffFactor
	}

	delayDuration := time.Duration(delay)
	if delayDuration > rp.MaxDelay {
		delayDuration = rp.MaxDelay
	}

	return delayDuration
}

// HTTPError represents an HTTP error with additional retry information
type HTTPError struct {
	StatusCode int
	Status     string
	URL        string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s (URL: %s)", e.StatusCode, e.Status, e.URL)
}

// IsRetryable returns true if this HTTP error is retryable
func (e *HTTPError) IsRetryable() bool {
	switch e.StatusCode {
	case http.StatusRequestTimeout, // 408
		http.StatusTooManyRequests,     // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	default:
		return false
	}
}

// RetryableDownloader wraps a Downloader with retry logic
type RetryableDownloader struct {
	downloader *Downloader
	policy     *RetryPolicy
}

// NewRetryableDownloader creates a new downloader with retry capability
func NewRetryableDownloader(policy *RetryPolicy) *RetryableDownloader {
	if policy == nil {
		policy = DefaultRetryPolicy()
	}

	return &RetryableDownloader{
		downloader: NewDownloader(),
		policy:     policy,
	}
}

// Download downloads a file with retry logic
func (rd *RetryableDownloader) Download(opts *DownloadOptions) (*DownloadResult, error) {
	var lastErr error

	for attempt := 0; attempt <= rd.policy.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate delay and wait
			delay := rd.policy.CalculateDelay(attempt)
			if delay > 0 {
				select {
				case <-opts.Context.Done():
					return nil, opts.Context.Err()
				case <-time.After(delay):
				}
			}
		}

		// Attempt download
		result, err := rd.downloader.Download(opts)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if we should retry
		if !rd.policy.ShouldRetry(err, attempt) {
			break
		}
	}

	return nil, fmt.Errorf("download failed after %d attempts: %w", rd.policy.MaxRetries+1, lastErr)
}

// OfflineDetector detects if the system is offline
type OfflineDetector struct{}

// NewOfflineDetector creates a new offline detector
func NewOfflineDetector() *OfflineDetector {
	return &OfflineDetector{}
}

// IsOnline checks if the system has internet connectivity
func (od *OfflineDetector) IsOnline(ctx context.Context) bool {
	// Try to connect to a few reliable hosts
	hosts := []string{
		"8.8.8.8:53",     // Google DNS
		"1.1.1.1:53",     // Cloudflare DNS
		"github.com:443", // GitHub HTTPS
	}

	for _, host := range hosts {
		dialer := &net.Dialer{
			Timeout: 5 * time.Second,
		}

		conn, err := dialer.DialContext(ctx, "tcp", host)
		if err == nil {
			conn.Close()
			return true
		}
	}

	return false
}

// Helper functions

// containsIgnoreCase checks if a string contains a substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}
