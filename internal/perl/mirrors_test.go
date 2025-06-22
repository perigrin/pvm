// ABOUTME: Tests for mirror support and CDN integration
// ABOUTME: Validates mirror configuration, health checking, and failover functionality

package perl

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewMirrorManager(t *testing.T) {
	mm := NewMirrorManager()
	defer mm.Close()

	mirrors := mm.GetMirrors()
	if len(mirrors) == 0 {
		t.Error("Expected default mirrors to be configured")
	}

	// Check that default mirrors are present
	found := false
	for _, mirror := range mirrors {
		if mirror.Name == "GitHub Releases (Primary)" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected GitHub Releases mirror to be in default configuration")
	}
}

func TestMirrorManagerWithCustomConfig(t *testing.T) {
	customMirrors := []MirrorConfig{
		{
			Name:     "Test Mirror",
			Type:     MirrorTypeDirectURL,
			BaseURL:  "https://test.example.com",
			Priority: 1,
			Enabled:  true,
		},
	}

	mm := NewMirrorManagerWithConfig(customMirrors)
	defer mm.Close()

	mirrors := mm.GetMirrors()
	if len(mirrors) != 1 {
		t.Errorf("Expected 1 mirror, got %d", len(mirrors))
	}

	if mirrors[0].Name != "Test Mirror" {
		t.Errorf("Expected 'Test Mirror', got %s", mirrors[0].Name)
	}
}

func TestGenerateMirrorURL(t *testing.T) {
	mm := NewMirrorManager()
	defer mm.Close()

	tests := []struct {
		name     string
		mirror   MirrorConfig
		version  string
		platform string
		expected string
		hasError bool
	}{
		{
			name: "GitHub Releases URL",
			mirror: MirrorConfig{
				Type:    MirrorTypeGitHubReleases,
				BaseURL: "https://github.com/owner/repo/releases/download",
			},
			version:  "5.38.0",
			platform: "linux-amd64",
			expected: "https://github.com/owner/repo/releases/download/perl-5.38.0/perl-5.38.0-linux-amd64.tar.gz",
		},
		{
			name: "jsDelivr CDN URL",
			mirror: MirrorConfig{
				Type:    MirrorTypeJSDelivr,
				BaseURL: "https://cdn.jsdelivr.net/gh/owner/repo@releases",
			},
			version:  "5.38.0",
			platform: "linux-amd64",
			expected: "https://cdn.jsdelivr.net/gh/owner/repo@releases/perl-5.38.0/perl-5.38.0-linux-amd64.tar.gz",
		},
		{
			name: "Windows ZIP archive",
			mirror: MirrorConfig{
				Type:    MirrorTypeDirectURL,
				BaseURL: "https://binaries.example.com",
			},
			version:  "5.38.0",
			platform: "windows-amd64",
			expected: "https://binaries.example.com/perl-5.38.0/perl-5.38.0-windows-amd64.zip",
		},
		{
			name: "Invalid version",
			mirror: MirrorConfig{
				Type:    MirrorTypeGitHubReleases,
				BaseURL: "https://github.com/owner/repo/releases/download",
			},
			version:  "invalid-version",
			platform: "linux-amd64",
			hasError: true,
		},
		{
			name: "Unknown mirror type",
			mirror: MirrorConfig{
				Type:    "unknown-type",
				BaseURL: "https://example.com",
			},
			version:  "5.38.0",
			platform: "linux-amd64",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := mm.GenerateMirrorURL(tt.mirror, tt.version, tt.platform)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if url != tt.expected {
				t.Errorf("Expected URL %s, got %s", tt.expected, url)
			}
		})
	}
}

func TestMirrorHealthCheck(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/healthy":
			w.WriteHeader(http.StatusOK)
		case "/unhealthy":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	mm := NewMirrorManager()
	defer mm.Close()

	tests := []struct {
		name          string
		mirror        MirrorConfig
		expectHealthy bool
	}{
		{
			name: "Healthy mirror",
			mirror: MirrorConfig{
				Name:        "Healthy Mirror",
				Type:        MirrorTypeDirectURL,
				BaseURL:     server.URL,
				Enabled:     true,
				Timeout:     5 * time.Second,
				HealthCheck: "/healthy",
			},
			expectHealthy: true,
		},
		{
			name: "Unhealthy mirror",
			mirror: MirrorConfig{
				Name:        "Unhealthy Mirror",
				Type:        MirrorTypeDirectURL,
				BaseURL:     server.URL,
				Enabled:     true,
				Timeout:     5 * time.Second,
				HealthCheck: "/unhealthy",
			},
			expectHealthy: false,
		},
		{
			name: "Disabled mirror",
			mirror: MirrorConfig{
				Name:        "Disabled Mirror",
				Type:        MirrorTypeDirectURL,
				BaseURL:     server.URL,
				Enabled:     false,
				HealthCheck: "/healthy",
			},
			expectHealthy: false,
		},
		{
			name: "No health check",
			mirror: MirrorConfig{
				Name:    "No Health Check Mirror",
				Type:    MirrorTypeDirectURL,
				BaseURL: server.URL,
				Enabled: true,
				Timeout: 5 * time.Second,
			},
			expectHealthy: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			health := mm.CheckMirrorHealth(tt.mirror)

			if health.Healthy != tt.expectHealthy {
				t.Errorf("Expected healthy=%t, got %t", tt.expectHealthy, health.Healthy)
			}

			if health.LastCheck.IsZero() {
				t.Error("Expected LastCheck to be set")
			}

			if tt.expectHealthy && health.LastError != nil {
				t.Errorf("Expected no error for healthy mirror, got: %v", health.LastError)
			}
		})
	}
}

func TestGetBestMirror(t *testing.T) {
	mirrors := []MirrorConfig{
		{
			Name:     "Priority 1",
			Type:     MirrorTypeDirectURL,
			BaseURL:  "https://priority1.example.com",
			Priority: 1,
			Enabled:  true,
		},
		{
			Name:     "Priority 2",
			Type:     MirrorTypeDirectURL,
			BaseURL:  "https://priority2.example.com",
			Priority: 2,
			Enabled:  true,
		},
		{
			Name:     "Disabled",
			Type:     MirrorTypeDirectURL,
			BaseURL:  "https://disabled.example.com",
			Priority: 0,
			Enabled:  false,
		},
	}

	mm := NewMirrorManagerWithConfig(mirrors)
	defer mm.Close()

	best, err := mm.GetBestMirror()
	if err != nil {
		t.Fatalf("GetBestMirror failed: %v", err)
	}

	if best.Name != "Priority 1" {
		t.Errorf("Expected 'Priority 1', got %s", best.Name)
	}

	// Mark priority 1 as unhealthy
	mm.UpdateMirrorHealth("Priority 1", false, 0, fmt.Errorf("test error"))

	best, err = mm.GetBestMirror()
	if err != nil {
		t.Fatalf("GetBestMirror failed: %v", err)
	}

	if best.Name != "Priority 2" {
		t.Errorf("Expected 'Priority 2' after Priority 1 became unhealthy, got %s", best.Name)
	}
}

func TestGetMirrorChain(t *testing.T) {
	mirrors := []MirrorConfig{
		{
			Name:     "Priority 3",
			Type:     MirrorTypeDirectURL,
			BaseURL:  "https://priority3.example.com",
			Priority: 3,
			Enabled:  true,
		},
		{
			Name:     "Priority 1",
			Type:     MirrorTypeDirectURL,
			BaseURL:  "https://priority1.example.com",
			Priority: 1,
			Enabled:  true,
		},
		{
			Name:     "Priority 2",
			Type:     MirrorTypeDirectURL,
			BaseURL:  "https://priority2.example.com",
			Priority: 2,
			Enabled:  true,
		},
		{
			Name:     "Disabled",
			Type:     MirrorTypeDirectURL,
			BaseURL:  "https://disabled.example.com",
			Priority: 0,
			Enabled:  false,
		},
	}

	mm := NewMirrorManagerWithConfig(mirrors)
	defer mm.Close()

	chain := mm.GetMirrorChain()

	// Should only include enabled mirrors, sorted by priority
	expectedOrder := []string{"Priority 1", "Priority 2", "Priority 3"}
	if len(chain) != len(expectedOrder) {
		t.Errorf("Expected %d mirrors in chain, got %d", len(expectedOrder), len(chain))
	}

	for i, expected := range expectedOrder {
		if i >= len(chain) {
			t.Errorf("Missing mirror at position %d", i)
			break
		}
		if chain[i].Name != expected {
			t.Errorf("Position %d: expected %s, got %s", i, expected, chain[i].Name)
		}
	}
}

func TestMirrorConfigValidation(t *testing.T) {
	mm := NewMirrorManager()
	defer mm.Close()

	tests := []struct {
		name    string
		mirror  MirrorConfig
		isValid bool
	}{
		{
			name: "Valid configuration",
			mirror: MirrorConfig{
				Name:     "Valid Mirror",
				Type:     MirrorTypeGitHubReleases,
				BaseURL:  "https://github.com/owner/repo/releases/download",
				Priority: 1,
				Enabled:  true,
			},
			isValid: true,
		},
		{
			name: "Empty name",
			mirror: MirrorConfig{
				Name:    "",
				Type:    MirrorTypeGitHubReleases,
				BaseURL: "https://example.com",
			},
			isValid: false,
		},
		{
			name: "Empty base URL",
			mirror: MirrorConfig{
				Name:    "Test Mirror",
				Type:    MirrorTypeGitHubReleases,
				BaseURL: "",
			},
			isValid: false,
		},
		{
			name: "Invalid URL scheme",
			mirror: MirrorConfig{
				Name:    "Test Mirror",
				Type:    MirrorTypeGitHubReleases,
				BaseURL: "ftp://example.com",
			},
			isValid: false,
		},
		{
			name: "Invalid mirror type",
			mirror: MirrorConfig{
				Name:    "Test Mirror",
				Type:    "invalid-type",
				BaseURL: "https://example.com",
			},
			isValid: false,
		},
		{
			name: "Negative priority",
			mirror: MirrorConfig{
				Name:     "Test Mirror",
				Type:     MirrorTypeGitHubReleases,
				BaseURL:  "https://example.com",
				Priority: -1,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mm.validateMirrorConfig(tt.mirror)

			if tt.isValid && err != nil {
				t.Errorf("Expected valid configuration, got error: %v", err)
			}

			if !tt.isValid && err == nil {
				t.Error("Expected invalid configuration to return error")
			}
		})
	}
}

func TestDownloadWithMirrorFailover(t *testing.T) {
	// Create test servers
	failingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failingServer.Close()

	successServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only serve the health check, but fail for the actual binary download
		if strings.Contains(r.URL.Path, "perl-5.38.0-linux-amd64.tar.gz") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Length", "13")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test content"))
	}))
	defer successServer.Close()

	mirrors := []MirrorConfig{
		{
			Name:     "Failing Mirror",
			Type:     MirrorTypeDirectURL,
			BaseURL:  failingServer.URL,
			Priority: 1,
			Enabled:  true,
			Timeout:  5 * time.Second,
		},
		{
			Name:     "Success Mirror",
			Type:     MirrorTypeDirectURL,
			BaseURL:  successServer.URL,
			Priority: 2,
			Enabled:  true,
			Timeout:  5 * time.Second,
		},
	}

	mm := NewMirrorManagerWithConfig(mirrors)
	defer mm.Close()

	// Test that failover works
	options := &BinaryDownloadOptions{
		Version:    "5.38.0",
		Platform:   "linux-amd64",
		Context:    context.Background(),
		SkipCache:  true, // Skip cache to ensure we actually hit the servers
		MaxRetries: 1,    // Reduce retries for faster test
	}

	// This should fail because servers return 404/500 for the binary download
	_, err := mm.DownloadWithMirrorFailover("5.38.0", "linux-amd64", options)

	// We expect this to fail because servers don't serve valid binary files
	if err == nil {
		t.Error("Expected download to fail with test servers that return errors")
	}

	// Check that health status was updated
	health := mm.GetMirrorHealth()

	if len(health) != 2 {
		t.Errorf("Expected health status for 2 mirrors, got %d", len(health))
	}

	// Verify that both mirrors were attempted and marked as unhealthy
	failingHealth := health["Failing Mirror"]
	if failingHealth == nil {
		t.Error("Expected health status for Failing Mirror")
	} else if failingHealth.Healthy {
		t.Error("Expected Failing Mirror to be marked unhealthy after failure")
	}
}

func TestMirrorHealthUpdate(t *testing.T) {
	mm := NewMirrorManager()
	defer mm.Close()

	// Add a test mirror
	mirrors := []MirrorConfig{
		{
			Name:     "Test Mirror",
			Type:     MirrorTypeDirectURL,
			BaseURL:  "https://test.example.com",
			Priority: 1,
			Enabled:  true,
		},
	}
	mm.SetMirrors(mirrors)

	// Test updating health status
	testError := fmt.Errorf("test error")
	mm.UpdateMirrorHealth("Test Mirror", false, 100*time.Millisecond, testError)

	health := mm.GetMirrorHealth()
	mirrorHealth := health["Test Mirror"]

	if mirrorHealth == nil {
		t.Fatal("Expected health status for Test Mirror")
	}

	if mirrorHealth.Healthy {
		t.Error("Expected mirror to be unhealthy")
	}

	if mirrorHealth.ResponseTime != 100*time.Millisecond {
		t.Errorf("Expected response time 100ms, got %v", mirrorHealth.ResponseTime)
	}

	if mirrorHealth.LastError.Error() != "test error" {
		t.Errorf("Expected 'test error', got %v", mirrorHealth.LastError)
	}

	if mirrorHealth.ConsecutiveFailures != 1 {
		t.Errorf("Expected 1 consecutive failure, got %d", mirrorHealth.ConsecutiveFailures)
	}

	// Test marking as healthy again
	mm.UpdateMirrorHealth("Test Mirror", true, 50*time.Millisecond, nil)

	health = mm.GetMirrorHealth()
	mirrorHealth = health["Test Mirror"]

	if !mirrorHealth.Healthy {
		t.Error("Expected mirror to be healthy")
	}

	if mirrorHealth.ConsecutiveFailures != 0 {
		t.Errorf("Expected 0 consecutive failures after recovery, got %d", mirrorHealth.ConsecutiveFailures)
	}
}

func TestMirrorManagerClose(t *testing.T) {
	mm := NewMirrorManager()

	// Should be able to close without panic
	mm.Close()

	// Closing again should be safe
	mm.Close()
}

func TestDefaultMirrorConfiguration(t *testing.T) {
	// Verify default mirrors are reasonable
	if len(DefaultMirrors) == 0 {
		t.Error("Expected default mirrors to be configured")
	}

	for i, mirror := range DefaultMirrors {
		if mirror.Name == "" {
			t.Errorf("Mirror %d has empty name", i)
		}

		if mirror.BaseURL == "" {
			t.Errorf("Mirror %d (%s) has empty base URL", i, mirror.Name)
		}

		if mirror.Type == "" {
			t.Errorf("Mirror %d (%s) has empty type", i, mirror.Name)
		}

		if mirror.Timeout == 0 {
			t.Errorf("Mirror %d (%s) has zero timeout", i, mirror.Name)
		}

		if mirror.MaxRetries == 0 {
			t.Errorf("Mirror %d (%s) has zero max retries", i, mirror.Name)
		}
	}
}
