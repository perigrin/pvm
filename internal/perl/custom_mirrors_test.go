// ABOUTME: Tests for custom mirror integration with configuration system
// ABOUTME: Validates configuration conversion and mirror functionality

package perl

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tamarou.com/pvm/internal/config"
)

func TestNewMirrorManagerFromConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.Config
		expected int // Expected number of mirrors (including defaults)
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: 3, // Just default mirrors
		},
		{
			name: "config with no custom mirrors",
			config: &config.Config{
				PVM: &config.PVMConfig{
					Binary: &config.PVMBinaryConfig{
						CustomMirrors: []*config.PVMCustomMirrorConfig{},
					},
				},
			},
			expected: 3, // Just default mirrors
		},
		{
			name: "config with custom mirrors",
			config: &config.Config{
				PVM: &config.PVMConfig{
					Binary: &config.PVMBinaryConfig{
						CustomMirrors: []*config.PVMCustomMirrorConfig{
							{
								Name:    "Custom Mirror 1",
								Type:    "direct",
								BaseURL: "https://custom1.example.com",
								Enabled: true,
							},
							{
								Name:    "Custom Mirror 2",
								Type:    "github-releases",
								BaseURL: "https://custom2.example.com",
								Enabled: true,
							},
						},
					},
				},
			},
			expected: 5, // 3 defaults + 2 custom
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mm := NewMirrorManagerFromConfig(tt.config)
			defer mm.Close()

			mirrors := mm.GetMirrors()
			if len(mirrors) != tt.expected {
				t.Errorf("Expected %d mirrors, got %d", tt.expected, len(mirrors))
			}
		})
	}
}

func TestConvertCustomMirrorConfig(t *testing.T) {
	customMirror := &config.PVMCustomMirrorConfig{
		Name:        "Test Mirror",
		Type:        "direct",
		BaseURL:     "https://test.example.com",
		Priority:    1,
		Enabled:     true,
		Timeout:     "45s",
		MaxRetries:  5,
		HealthCheck: "/health",
		Headers:     map[string]string{"X-Custom": "test"},
		URLTemplate: "{base_url}/downloads/{version}/{filename}",
		VersionMapping: map[string]string{
			"5.38.0": "5.38.0-custom",
		},
		Auth: &config.PVMCustomMirrorAuth{
			Type:  "bearer",
			Token: "test-token",
		},
	}

	mirror := convertCustomMirrorConfig(customMirror)

	if mirror.Name != "Test Mirror" {
		t.Errorf("Expected name 'Test Mirror', got %s", mirror.Name)
	}

	if mirror.Type != "direct" {
		t.Errorf("Expected type 'direct', got %s", mirror.Type)
	}

	if mirror.Timeout != 45*time.Second {
		t.Errorf("Expected timeout 45s, got %v", mirror.Timeout)
	}

	if mirror.Auth == nil {
		t.Error("Expected auth to be configured")
	} else {
		if mirror.Auth.Type != "bearer" {
			t.Errorf("Expected auth type 'bearer', got %s", mirror.Auth.Type)
		}
		if mirror.Auth.Token != "test-token" {
			t.Errorf("Expected token 'test-token', got %s", mirror.Auth.Token)
		}
	}

	if mirror.URLTemplate != "{base_url}/downloads/{version}/{filename}" {
		t.Errorf("Expected URL template, got %s", mirror.URLTemplate)
	}

	if len(mirror.VersionMapping) != 1 {
		t.Errorf("Expected 1 version mapping, got %d", len(mirror.VersionMapping))
	}
}

func TestGenerateURLFromTemplate(t *testing.T) {
	mm := NewMirrorManager()
	defer mm.Close()

	tests := []struct {
		name     string
		template string
		baseURL  string
		version  string
		platform string
		filename string
		ext      string
		expected string
	}{
		{
			name:     "basic template",
			template: "{base_url}/downloads/{version}/{filename}",
			baseURL:  "https://example.com",
			version:  "5.38.0",
			platform: "linux-amd64",
			filename: "perl-5.38.0-linux-amd64.tar.gz",
			ext:      ".tar.gz",
			expected: "https://example.com/downloads/5.38.0/perl-5.38.0-linux-amd64.tar.gz",
		},
		{
			name:     "template with platform and ext",
			template: "{base_url}/v{version}/{platform}/perl{ext}",
			baseURL:  "https://example.com",
			version:  "5.38.0",
			platform: "darwin-arm64",
			filename: "perl-5.38.0-darwin-arm64.tar.gz",
			ext:      ".tar.gz",
			expected: "https://example.com/v5.38.0/darwin-arm64/perl.tar.gz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mm.generateURLFromTemplate(tt.template, tt.baseURL, tt.version, tt.platform, tt.filename, tt.ext)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestMirrorVersionMapping(t *testing.T) {
	mirror := MirrorConfig{
		Name:    "Test Mirror",
		Type:    "direct",
		BaseURL: "https://example.com",
		VersionMapping: map[string]string{
			"5.38.0": "5.38.0-custom",
			"5.40.0": "5.40.0-patched",
		},
	}

	mm := NewMirrorManager()
	defer mm.Close()

	tests := []struct {
		version  string
		platform string
		expected string
	}{
		{
			version:  "5.38.0",
			platform: "linux-amd64",
			expected: "https://example.com/perl-5.38.0-custom/perl-5.38.0-custom-linux-amd64.tar.gz",
		},
		{
			version:  "5.40.0",
			platform: "darwin-arm64",
			expected: "https://example.com/perl-5.40.0-patched/perl-5.40.0-patched-darwin-arm64.tar.gz",
		},
		{
			version:  "5.36.0",
			platform: "linux-amd64",
			expected: "https://example.com/perl-5.36.0/perl-5.36.0-linux-amd64.tar.gz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.version+"-"+tt.platform, func(t *testing.T) {
			url, err := mm.GenerateMirrorURL(mirror, tt.version, tt.platform)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if url != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, url)
			}
		})
	}
}

func TestMirrorAuthentication(t *testing.T) {
	// Create test server to capture requests
	var capturedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mm := NewMirrorManager()
	defer mm.Close()

	tests := []struct {
		name         string
		auth         *MirrorAuth
		expectHeader string
		expectValue  string
	}{
		{
			name: "basic auth",
			auth: &MirrorAuth{
				Type:     "basic",
				Username: "user",
				Password: "pass",
			},
			expectHeader: "Authorization",
			expectValue:  "Basic dXNlcjpwYXNz", // base64("user:pass")
		},
		{
			name: "bearer token",
			auth: &MirrorAuth{
				Type:  "bearer",
				Token: "secret-token",
			},
			expectHeader: "Authorization",
			expectValue:  "Bearer secret-token",
		},
		{
			name: "api key",
			auth: &MirrorAuth{
				Type:   "api-key",
				APIKey: "api-secret",
			},
			expectHeader: "X-API-Key",
			expectValue:  "api-secret",
		},
		{
			name: "custom api key header",
			auth: &MirrorAuth{
				Type:         "api-key",
				APIKey:       "api-secret",
				APIKeyHeader: "X-Custom-Key",
			},
			expectHeader: "X-Custom-Key",
			expectValue:  "api-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mirror := MirrorConfig{
				Name:        "Test Mirror",
				Type:        "direct",
				BaseURL:     server.URL,
				HealthCheck: "/health",
				Auth:        tt.auth,
				Timeout:     1 * time.Second,
				Enabled:     true,
			}

			// Reset captured headers
			capturedHeaders = nil

			// Perform health check which should include auth headers
			health := mm.CheckMirrorHealth(mirror)
			if health.LastError != nil {
				t.Fatalf("Health check failed: %v", health.LastError)
			}

			// Check if auth header was set correctly
			if capturedHeaders == nil {
				t.Fatal("No headers captured")
			}

			value := capturedHeaders.Get(tt.expectHeader)
			if value != tt.expectValue {
				t.Errorf("Expected header %s=%s, got %s", tt.expectHeader, tt.expectValue, value)
			}
		})
	}
}

func TestMirrorURLTemplate(t *testing.T) {
	mirror := MirrorConfig{
		Name:        "Custom Template Mirror",
		Type:        "direct",
		BaseURL:     "https://example.com",
		URLTemplate: "{base_url}/custom/{version}/{platform}.{ext}",
	}

	mm := NewMirrorManager()
	defer mm.Close()

	url, err := mm.GenerateMirrorURL(mirror, "5.38.0", "linux-amd64")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "https://example.com/custom/5.38.0/linux-amd64..tar.gz"
	if url != expected {
		t.Errorf("Expected %s, got %s", expected, url)
	}
}

func TestConvertAuthConfig(t *testing.T) {
	tests := []struct {
		name     string
		auth     *config.PVMCustomMirrorAuth
		expected *MirrorAuth
	}{
		{
			name: "basic auth",
			auth: &config.PVMCustomMirrorAuth{
				Type:     "basic",
				Username: "testuser",
				Password: "testpass",
			},
			expected: &MirrorAuth{
				Type:     "basic",
				Username: "testuser",
				Password: "testpass",
			},
		},
		{
			name: "api key with default header",
			auth: &config.PVMCustomMirrorAuth{
				Type:   "api-key",
				APIKey: "secret-key",
			},
			expected: &MirrorAuth{
				Type:         "api-key",
				APIKey:       "secret-key",
				APIKeyHeader: "X-API-Key",
			},
		},
		{
			name: "oauth2",
			auth: &config.PVMCustomMirrorAuth{
				Type: "oauth2",
				OAuth2: &config.PVMCustomMirrorOAuth2{
					ClientID:     "client123",
					ClientSecret: "secret123",
					TokenURL:     "https://auth.example.com/token",
					Scopes:       []string{"read", "write"},
				},
			},
			expected: &MirrorAuth{
				Type:               "oauth2",
				OAuth2ClientID:     "client123",
				OAuth2ClientSecret: "secret123",
				OAuth2TokenURL:     "https://auth.example.com/token",
				OAuth2Scopes:       []string{"read", "write"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertAuthConfig(tt.auth)

			if result.Type != tt.expected.Type {
				t.Errorf("Expected type %s, got %s", tt.expected.Type, result.Type)
			}

			if result.Username != tt.expected.Username {
				t.Errorf("Expected username %s, got %s", tt.expected.Username, result.Username)
			}

			if result.APIKeyHeader != tt.expected.APIKeyHeader {
				t.Errorf("Expected API key header %s, got %s", tt.expected.APIKeyHeader, result.APIKeyHeader)
			}

			if result.OAuth2ClientID != tt.expected.OAuth2ClientID {
				t.Errorf("Expected OAuth2 client ID %s, got %s", tt.expected.OAuth2ClientID, result.OAuth2ClientID)
			}
		})
	}
}
