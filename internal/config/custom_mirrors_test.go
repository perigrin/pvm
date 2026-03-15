// ABOUTME: Tests for custom mirror configuration support
// ABOUTME: Validates custom mirror configuration structure and validation logic

package config

import (
	"strings"
	"testing"
)

func TestPVMCustomMirrorConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *PVMCustomMirrorConfig
		expectError bool
		errorCount  int
	}{
		{
			name: "valid custom mirror config",
			config: &PVMCustomMirrorConfig{
				Name:        "Test Mirror",
				Type:        "github-releases",
				BaseURL:     "https://example.com/releases",
				Priority:    1,
				Enabled:     true,
				Timeout:     "30s",
				MaxRetries:  3,
				HealthCheck: "/health",
				Headers:     map[string]string{"X-Custom": "value"},
				URLTemplate: "{base_url}/perl-{version}/{filename}",
				VersionMapping: map[string]string{
					"5.38.0": "5.38.0-custom",
				},
			},
			expectError: false,
			errorCount:  0,
		},
		{
			name: "missing name",
			config: &PVMCustomMirrorConfig{
				Type:    "github-releases",
				BaseURL: "https://example.com",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "invalid type",
			config: &PVMCustomMirrorConfig{
				Name:    "Test",
				Type:    "invalid-type",
				BaseURL: "https://example.com",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "missing base URL",
			config: &PVMCustomMirrorConfig{
				Name: "Test",
				Type: "github-releases",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "invalid base URL",
			config: &PVMCustomMirrorConfig{
				Name:    "Test",
				Type:    "github-releases",
				BaseURL: "ftp://example.com",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "negative priority",
			config: &PVMCustomMirrorConfig{
				Name:     "Test",
				Type:     "github-releases",
				BaseURL:  "https://example.com",
				Priority: -1,
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "invalid timeout",
			config: &PVMCustomMirrorConfig{
				Name:    "Test",
				Type:    "github-releases",
				BaseURL: "https://example.com",
				Timeout: "invalid",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "negative max retries",
			config: &PVMCustomMirrorConfig{
				Name:       "Test",
				Type:       "github-releases",
				BaseURL:    "https://example.com",
				MaxRetries: -1,
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "empty header key",
			config: &PVMCustomMirrorConfig{
				Name:    "Test",
				Type:    "github-releases",
				BaseURL: "https://example.com",
				Headers: map[string]string{"": "value"},
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "empty header value",
			config: &PVMCustomMirrorConfig{
				Name:    "Test",
				Type:    "github-releases",
				BaseURL: "https://example.com",
				Headers: map[string]string{"key": ""},
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "empty version mapping key",
			config: &PVMCustomMirrorConfig{
				Name:           "Test",
				Type:           "github-releases",
				BaseURL:        "https://example.com",
				VersionMapping: map[string]string{"": "value"},
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "empty version mapping value",
			config: &PVMCustomMirrorConfig{
				Name:           "Test",
				Type:           "github-releases",
				BaseURL:        "https://example.com",
				VersionMapping: map[string]string{"key": ""},
			},
			expectError: true,
			errorCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.config.Validate()

			if tt.expectError && len(errors) == 0 {
				t.Errorf("Expected validation errors but got none")
			}

			if !tt.expectError && len(errors) > 0 {
				t.Errorf("Expected no validation errors but got: %v", errors)
			}

			if tt.errorCount > 0 && len(errors) != tt.errorCount {
				t.Errorf("Expected %d errors but got %d: %v", tt.errorCount, len(errors), errors)
			}
		})
	}
}

func TestPVMCustomMirrorAuth_Validate(t *testing.T) {
	tests := []struct {
		name        string
		auth        *PVMCustomMirrorAuth
		expectError bool
		errorCount  int
	}{
		{
			name: "valid none auth",
			auth: &PVMCustomMirrorAuth{
				Type: "none",
			},
			expectError: false,
			errorCount:  0,
		},
		{
			name: "valid basic auth",
			auth: &PVMCustomMirrorAuth{
				Type:     "basic",
				Username: "user",
				Password: "pass",
			},
			expectError: false,
			errorCount:  0,
		},
		{
			name: "valid bearer auth",
			auth: &PVMCustomMirrorAuth{
				Type:  "bearer",
				Token: "token123",
			},
			expectError: false,
			errorCount:  0,
		},
		{
			name: "valid api-key auth",
			auth: &PVMCustomMirrorAuth{
				Type:   "api-key",
				APIKey: "key123",
			},
			expectError: false,
			errorCount:  0,
		},
		{
			name: "valid oauth2 auth",
			auth: &PVMCustomMirrorAuth{
				Type: "oauth2",
				OAuth2: &PVMCustomMirrorOAuth2{
					ClientID:     "client123",
					ClientSecret: "secret123",
					TokenURL:     "https://example.com/token",
					Scopes:       []string{"read"},
				},
			},
			expectError: false,
			errorCount:  0,
		},
		{
			name: "invalid auth type",
			auth: &PVMCustomMirrorAuth{
				Type: "invalid",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "basic auth missing username",
			auth: &PVMCustomMirrorAuth{
				Type:     "basic",
				Password: "pass",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "basic auth missing password",
			auth: &PVMCustomMirrorAuth{
				Type:     "basic",
				Username: "user",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "bearer auth missing token",
			auth: &PVMCustomMirrorAuth{
				Type: "bearer",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "api-key auth missing key",
			auth: &PVMCustomMirrorAuth{
				Type: "api-key",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "oauth2 auth missing config",
			auth: &PVMCustomMirrorAuth{
				Type: "oauth2",
			},
			expectError: true,
			errorCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.auth.Validate()

			if tt.expectError && len(errors) == 0 {
				t.Errorf("Expected validation errors but got none")
			}

			if !tt.expectError && len(errors) > 0 {
				t.Errorf("Expected no validation errors but got: %v", errors)
			}

			if tt.errorCount > 0 && len(errors) != tt.errorCount {
				t.Errorf("Expected %d errors but got %d: %v", tt.errorCount, len(errors), errors)
			}
		})
	}
}

func TestPVMCustomMirrorOAuth2_Validate(t *testing.T) {
	tests := []struct {
		name        string
		oauth2      *PVMCustomMirrorOAuth2
		expectError bool
		errorCount  int
	}{
		{
			name: "valid oauth2 config",
			oauth2: &PVMCustomMirrorOAuth2{
				ClientID:     "client123",
				ClientSecret: "secret123",
				TokenURL:     "https://example.com/token",
				Scopes:       []string{"read", "write"},
			},
			expectError: false,
			errorCount:  0,
		},
		{
			name: "missing client ID",
			oauth2: &PVMCustomMirrorOAuth2{
				ClientSecret: "secret123",
				TokenURL:     "https://example.com/token",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "missing client secret",
			oauth2: &PVMCustomMirrorOAuth2{
				ClientID: "client123",
				TokenURL: "https://example.com/token",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "missing token URL",
			oauth2: &PVMCustomMirrorOAuth2{
				ClientID:     "client123",
				ClientSecret: "secret123",
			},
			expectError: true,
			errorCount:  1,
		},
		{
			name: "invalid token URL",
			oauth2: &PVMCustomMirrorOAuth2{
				ClientID:     "client123",
				ClientSecret: "secret123",
				TokenURL:     "ftp://example.com/token",
			},
			expectError: true,
			errorCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.oauth2.Validate()

			if tt.expectError && len(errors) == 0 {
				t.Errorf("Expected validation errors but got none")
			}

			if !tt.expectError && len(errors) > 0 {
				t.Errorf("Expected no validation errors but got: %v", errors)
			}

			if tt.errorCount > 0 && len(errors) != tt.errorCount {
				t.Errorf("Expected %d errors but got %d: %v", tt.errorCount, len(errors), errors)
			}
		})
	}
}

func TestPVMBinaryConfig_ValidateCustomMirrors(t *testing.T) {
	config := &PVMBinaryConfig{
		DefaultInstallMethod: "source",
		BinaryMirrors:        []string{"https://example.com"},
		CustomMirrors: []*PVMCustomMirrorConfig{
			{
				Name:    "Valid Mirror",
				Type:    "github-releases",
				BaseURL: "https://valid.com",
			},
			{
				Name: "Invalid Mirror",
				Type: "invalid-type",
			},
		},
	}

	errors := config.Validate()

	// Should have at least one error from the invalid mirror
	if len(errors) == 0 {
		t.Error("Expected validation errors for invalid custom mirror")
	}

	// Check that error mentions the custom mirror
	found := false
	for _, err := range errors {
		if strings.Contains(err.Error(), "CustomMirrors[1]") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error to reference CustomMirrors[1] but didn't find it")
	}
}

func TestDefaultConfig_HasEmptyCustomMirrors(t *testing.T) {
	config := NewDefaultConfig()

	if config.PVM == nil {
		t.Fatal("Expected PVM config to be initialized")
	}

	if config.PVM.Binary == nil {
		t.Fatal("Expected Binary config to be initialized")
	}

	if config.PVM.Binary.CustomMirrors == nil {
		t.Error("Expected CustomMirrors to be initialized as empty slice")
	}

	if len(config.PVM.Binary.CustomMirrors) != 0 {
		t.Error("Expected CustomMirrors to be empty by default")
	}
}
