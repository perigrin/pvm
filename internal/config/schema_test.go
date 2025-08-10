// ABOUTME: Tests for configuration schema validation
// ABOUTME: Validates schema-based configuration validation and error detection

package config

import (
	"strings"
	"testing"
	"time"
)

func TestSchemaValidator(t *testing.T) {
	sv := NewSchemaValidator()

	t.Run("ValidConfiguration", func(t *testing.T) {
		// Create a valid configuration
		config := NewDefaultConfig()

		// Validate should pass with no errors
		errors := sv.ValidateConfig(config)
		if len(errors) > 0 {
			t.Errorf("Expected no validation errors for default config, got: %v", errors)
		}
	})

	t.Run("InvalidStringField", func(t *testing.T) {
		config := NewDefaultConfig()
		// Set invalid isolation level
		config.PVX.IsolationLevel = "invalid"

		errors := sv.ValidateConfig(config)
		hasIsolationError := false
		for _, err := range errors {
			if strings.Contains(err.Error(), "isolation_level") && strings.Contains(err.Error(), "one of") {
				hasIsolationError = true
				break
			}
		}
		if !hasIsolationError {
			t.Error("Expected validation error for invalid isolation level")
		}
	})

	t.Run("InvalidIntegerField", func(t *testing.T) {
		config := NewDefaultConfig()
		// Set invalid port (too high)
		config.MCP.Port = 70000

		errors := sv.ValidateConfig(config)
		hasPortError := false
		for _, err := range errors {
			if strings.Contains(err.Error(), "port") && strings.Contains(err.Error(), "at most") {
				hasPortError = true
				break
			}
		}
		if !hasPortError {
			t.Error("Expected validation error for invalid port")
		}
	})

	t.Run("PatternValidation", func(t *testing.T) {
		config := NewDefaultConfig()

		// Test URL pattern
		config.PVM.DownloadMirror = "not-a-url"
		errors := sv.ValidateConfig(config)
		hasURLError := false
		for _, err := range errors {
			if strings.Contains(err.Error(), "download_mirror") && strings.Contains(err.Error(), "pattern") {
				hasURLError = true
				break
			}
		}
		if !hasURLError {
			t.Error("Expected validation error for invalid URL pattern")
		}

		// Test memory pattern
		config.PVX.MaxMemory = "invalid-memory"
		errors = sv.ValidateConfig(config)
		hasMemoryError := false
		for _, err := range errors {
			if strings.Contains(err.Error(), "max_memory") && strings.Contains(err.Error(), "pattern") {
				hasMemoryError = true
				break
			}
		}
		if !hasMemoryError {
			t.Error("Expected validation error for invalid memory pattern")
		}
	})

	t.Run("RangeValidation", func(t *testing.T) {
		config := NewDefaultConfig()

		// Test negative timeout
		config.PVX.Timeout = -1
		errors := sv.ValidateConfig(config)
		hasTimeoutError := false
		for _, err := range errors {
			if strings.Contains(err.Error(), "timeout") && strings.Contains(err.Error(), "at least") {
				hasTimeoutError = true
				break
			}
		}
		if !hasTimeoutError {
			t.Error("Expected validation error for negative timeout")
		}

		// Test excessive build jobs
		config.PVM.BuildJobs = 100
		errors = sv.ValidateConfig(config)
		hasBuildJobsError := false
		for _, err := range errors {
			if strings.Contains(err.Error(), "build_jobs") && strings.Contains(err.Error(), "at most") {
				hasBuildJobsError = true
				break
			}
		}
		if !hasBuildJobsError {
			t.Error("Expected validation error for excessive build jobs")
		}
	})

	t.Run("EnumValidation", func(t *testing.T) {
		config := NewDefaultConfig()

		// Test invalid installer
		config.PM.PreferredInstaller = "invalid-installer"
		errors := sv.ValidateConfig(config)
		hasInstallerError := false
		for _, err := range errors {
			if strings.Contains(err.Error(), "preferred_installer") && strings.Contains(err.Error(), "one of") {
				hasInstallerError = true
				break
			}
		}
		if !hasInstallerError {
			t.Error("Expected validation error for invalid installer")
		}

		// Test invalid embedding provider
		config.MCP.EmbeddingProvider = "invalid-provider"
		errors = sv.ValidateConfig(config)
		hasProviderError := false
		for _, err := range errors {
			if strings.Contains(err.Error(), "embedding_provider") && strings.Contains(err.Error(), "one of") {
				hasProviderError = true
				break
			}
		}
		if !hasProviderError {
			t.Error("Expected validation error for invalid embedding provider")
		}
	})

	t.Run("DurationValidation", func(t *testing.T) {
		config := NewDefaultConfig()

		// Test zero or negative duration
		config.MCP.RequestTimeout = 0
		errors := sv.ValidateConfig(config)
		hasDurationError := false
		for _, err := range errors {
			if strings.Contains(err.Error(), "request_timeout") && strings.Contains(err.Error(), "positive") {
				hasDurationError = true
				break
			}
		}
		if !hasDurationError {
			t.Error("Expected validation error for zero duration")
		}
	})

	t.Run("ValidMemoryFormats", func(t *testing.T) {
		config := NewDefaultConfig()

		// Test valid memory formats
		validMemoryFormats := []string{"512MB", "2GB", "1024KB", "1TB"}
		for _, format := range validMemoryFormats {
			config.PVX.MaxMemory = format
			errors := sv.ValidateConfig(config)
			hasMemoryError := false
			for _, err := range errors {
				if strings.Contains(err.Error(), "max_memory") {
					hasMemoryError = true
					break
				}
			}
			if hasMemoryError {
				t.Errorf("Unexpected validation error for valid memory format '%s'", format)
			}
		}
	})

	t.Run("ValidURLFormats", func(t *testing.T) {
		config := NewDefaultConfig()

		// Test valid URL formats
		validURLs := []string{
			"http://example.com",
			"https://example.com",
			"https://api.example.com/v1",
			"http://localhost:8080",
		}
		for _, url := range validURLs {
			config.PVM.DownloadMirror = url
			errors := sv.ValidateConfig(config)
			hasURLError := false
			for _, err := range errors {
				if strings.Contains(err.Error(), "download_mirror") {
					hasURLError = true
					break
				}
			}
			if hasURLError {
				t.Errorf("Unexpected validation error for valid URL '%s'", url)
			}
		}
	})

	t.Run("ValidVersionFormats", func(t *testing.T) {
		config := NewDefaultConfig()

		// Test valid version formats
		validVersions := []string{
			"5.38.0",
			"5.36.1",
			"5.40",
			"latest",
			"stable",
		}
		for _, version := range validVersions {
			config.PVM.DefaultPerl = version
			errors := sv.ValidateConfig(config)
			hasVersionError := false
			for _, err := range errors {
				if strings.Contains(err.Error(), "default_perl") {
					hasVersionError = true
					break
				}
			}
			if hasVersionError {
				t.Errorf("Unexpected validation error for valid version '%s'", version)
			}
		}
	})

	t.Run("MultipleErrors", func(t *testing.T) {
		config := NewDefaultConfig()

		// Set multiple invalid values
		config.PVX.IsolationLevel = "invalid"
		config.PVM.BuildJobs = -1
		config.MCP.Port = 70000
		config.PM.PreferredInstaller = "invalid-installer"

		errors := sv.ValidateConfig(config)
		if len(errors) < 4 {
			t.Errorf("Expected at least 4 validation errors, got %d", len(errors))
		}
	})

	t.Run("NilSections", func(t *testing.T) {
		config := &Config{
			// Only set some sections, leave others nil
			PVM: &PVMConfig{
				DefaultPerl: "5.38.0",
				BuildJobs:   4,
			},
		}

		// Should not crash with nil sections
		errors := sv.ValidateConfig(config)
		// Some errors might be expected due to nil sections, but shouldn't crash
		if len(errors) > 10 {
			t.Errorf("Too many validation errors for config with nil sections: %d", len(errors))
		}
	})

	t.Run("EdgeCaseValues", func(t *testing.T) {
		config := NewDefaultConfig()

		// Test boundary values
		config.MCP.Port = 1                             // minimum valid port
		config.PVM.BuildJobs = 1                        // minimum valid build jobs
		config.PVX.Timeout = 0                          // minimum valid timeout
		config.MCP.MaxConcurrentRequests = 1            // minimum valid concurrent requests
		config.MCP.RequestTimeout = 1 * time.Nanosecond // minimum positive duration

		errors := sv.ValidateConfig(config)
		// These should all be valid boundary values
		for _, err := range errors {
			if strings.Contains(err.Error(), "at least") {
				t.Errorf("Boundary value validation error: %v", err)
			}
		}

		// Test maximum boundary values
		config.MCP.Port = 65535                // maximum valid port
		config.PVM.BuildJobs = 64              // maximum recommended build jobs
		config.PVX.Timeout = 3600              // maximum recommended timeout
		config.MCP.MaxConcurrentRequests = 100 // maximum recommended concurrent requests

		errors = sv.ValidateConfig(config)
		// These should all be valid boundary values
		for _, err := range errors {
			if strings.Contains(err.Error(), "at most") || strings.Contains(err.Error(), "exceed") {
				t.Errorf("Boundary value validation error: %v", err)
			}
		}
	})
}
