// ABOUTME: Unit tests for MCP server implementation
// ABOUTME: Tests server creation, configuration validation, and project discovery

package mcp

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"tamarou.com/pvm/internal/config"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "nil MCP config",
			config: &config.Config{
				MCP: nil,
			},
			wantErr: true,
		},
		{
			name: "valid config",
			config: &config.Config{
				MCP: &config.MCPConfig{
					Port:                 3000,
					Host:                 "localhost",
					AutoDiscoverProjects: true,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := NewServer(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && server == nil {
				t.Error("NewServer() returned nil server for valid config")
			}
		})
	}
}

func TestProjectDiscovery(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pvm-mcp-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	}()

	// Create test files
	cpanfileContent := "requires 'Test::More';\n"
	if err := os.WriteFile(filepath.Join(tempDir, "cpanfile"), []byte(cpanfileContent), 0644); err != nil {
		t.Fatalf("Failed to create cpanfile: %v", err)
	}

	versionContent := "5.38.0\n"
	if err := os.WriteFile(filepath.Join(tempDir, ".perl-version"), []byte(versionContent), 0644); err != nil {
		t.Fatalf("Failed to create .perl-version: %v", err)
	}

	// Create a .pm file
	pmContent := "package Test;\n1;\n"
	if err := os.WriteFile(filepath.Join(tempDir, "Test.pm"), []byte(pmContent), 0644); err != nil {
		t.Fatalf("Failed to create Test.pm: %v", err)
	}

	// Create server
	cfg := &config.Config{
		MCP: &config.MCPConfig{
			Port:                 3000,
			Host:                 "localhost",
			AutoDiscoverProjects: true,
		},
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test project discovery
	if err := server.discoverProjectsInPath(tempDir); err != nil {
		t.Fatalf("Failed to discover projects: %v", err)
	}

	// Verify project was discovered
	projects := server.GetProjects()
	if len(projects) == 0 {
		t.Error("No projects discovered")
		return
	}

	project, exists := projects[tempDir]
	if !exists {
		t.Error("Project not found in discovered projects")
		return
	}

	// Verify project details
	if !project.HasCpanfile {
		t.Error("Expected project to have cpanfile")
	}

	if !project.HasVersion {
		t.Error("Expected project to have version file")
	}

	if project.PerlVersion != "5.38.0" {
		t.Errorf("Expected Perl version 5.38.0, got %s", project.PerlVersion)
	}

	if project.ProjectType != "cpan" {
		t.Errorf("Expected project type 'cpan', got %s", project.ProjectType)
	}
}

func TestHasPerlFiles(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "pvm-mcp-test-perl")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	}()

	// Create server
	cfg := &config.Config{
		MCP: &config.MCPConfig{
			Port: 3000,
			Host: "localhost",
		},
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test empty directory
	if server.hasPerlFiles(tempDir) {
		t.Error("Empty directory should not have Perl files")
	}

	// Create a Perl file
	pmContent := "package Test;\n1;\n"
	if err := os.WriteFile(filepath.Join(tempDir, "Test.pm"), []byte(pmContent), 0644); err != nil {
		t.Fatalf("Failed to create Test.pm: %v", err)
	}

	// Test directory with Perl file
	if !server.hasPerlFiles(tempDir) {
		t.Error("Directory with .pm file should have Perl files")
	}

	// Test with common Perl directory
	libDir := filepath.Join(tempDir, "lib")
	if err := os.Mkdir(libDir, 0755); err != nil {
		t.Fatalf("Failed to create lib directory: %v", err)
	}

	if !server.hasPerlFiles(tempDir) {
		t.Error("Directory with lib subdirectory should have Perl files")
	}
}

func TestMCPConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.MCPConfig
		wantError bool
	}{
		{
			name: "valid config",
			config: &config.MCPConfig{
				Port:                  3000,
				Host:                  "localhost",
				EmbeddingProvider:     "openai",
				ValidationCacheSize:   "50MB",
				EmbeddingCacheSize:    "100MB",
				RequestTimeout:        "30s",
				MaxConcurrentRequests: 10,
				GenerationMemorySize:  50,
			},
			wantError: false,
		},
		{
			name: "invalid port",
			config: &config.MCPConfig{
				Port: 0,
				Host: "localhost",
			},
			wantError: true,
		},
		{
			name: "invalid embedding provider",
			config: &config.MCPConfig{
				Port:              3000,
				Host:              "localhost",
				EmbeddingProvider: "invalid",
			},
			wantError: true,
		},
		{
			name: "invalid memory format",
			config: &config.MCPConfig{
				Port:                3000,
				Host:                "localhost",
				ValidationCacheSize: "invalid",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.config.Validate()
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("MCPConfig.Validate() hasError = %v, wantError %v, errors = %v", hasError, tt.wantError, errors)
			}
		})
	}
}

func TestServerMetrics(t *testing.T) {
	// Create server
	cfg := &config.Config{
		MCP: &config.MCPConfig{
			Port:                 3000,
			Host:                 "localhost",
			AutoDiscoverProjects: true,
		},
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test initial metrics
	metrics := server.GetMetrics()
	if metrics.RequestCount != 0 {
		t.Errorf("Expected initial RequestCount to be 0, got %d", metrics.RequestCount)
	}
	if metrics.ErrorCount != 0 {
		t.Errorf("Expected initial ErrorCount to be 0, got %d", metrics.ErrorCount)
	}
	if len(metrics.ToolUsageCount) != 0 {
		t.Errorf("Expected initial ToolUsageCount to be empty, got %v", metrics.ToolUsageCount)
	}

	// Test recordError
	server.recordError()
	metrics = server.GetMetrics()
	if metrics.ErrorCount != 1 {
		t.Errorf("Expected ErrorCount to be 1 after recordError(), got %d", metrics.ErrorCount)
	}

	// Test recordToolUsage
	server.recordToolUsage("test_tool", time.Now().Add(-10*time.Millisecond))
	metrics = server.GetMetrics()
	if metrics.RequestCount != 1 {
		t.Errorf("Expected RequestCount to be 1 after recordToolUsage(), got %d", metrics.RequestCount)
	}
	if metrics.ToolUsageCount["test_tool"] != 1 {
		t.Errorf("Expected ToolUsageCount['test_tool'] to be 1, got %d", metrics.ToolUsageCount["test_tool"])
	}
	if metrics.AverageLatency == 0 {
		t.Error("Expected AverageLatency to be set after recordToolUsage()")
	}
}
