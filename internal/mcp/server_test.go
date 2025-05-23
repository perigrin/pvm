// ABOUTME: Unit tests for MCP server implementation
// ABOUTME: Tests server creation, configuration validation, and project discovery

package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
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

func TestHandleAnalyzeCode_ProjectAnalysis(t *testing.T) {
	// Create temporary test project
	tempDir, err := os.MkdirTemp("", "pvm-mcp-project-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := map[string]string{
		"module.pm": `package TestModule;
use strict;
use warnings;

type Counter = Int;

sub increment {
    my Counter $value = shift;
    return $value + 1;
}

1;`,
		"script.pl": `#!/usr/bin/env perl
use strict;
use warnings;
use TestModule;

type Counter = Str;  # Type conflict

my Counter $count = "10";
print TestModule::increment($count);
`,
	}

	for filename, content := range testFiles {
		if err := os.WriteFile(filepath.Join(tempDir, filename), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", filename, err)
		}
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

	tests := []struct {
		name         string
		analysisType string
		projectPath  string
		expectError  bool
	}{
		{
			name:         "project analysis",
			analysisType: "project_analysis",
			projectPath:  tempDir,
			expectError:  false,
		},
		{
			name:         "project summary",
			analysisType: "project_summary",
			projectPath:  tempDir,
			expectError:  false,
		},
		{
			name:         "project summary without analysis",
			analysisType: "project_summary",
			projectPath:  "/non/existent/path",
			expectError:  true,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			params := map[string]interface{}{
				"analysis_type": tt.analysisType,
				"project_path":  tt.projectPath,
			}

			request := mcp.CallToolRequest{
				Params: struct {
					Name      string    `json:"name"`
					Arguments any       `json:"arguments,omitempty"`
					Meta      *mcp.Meta `json:"_meta,omitempty"`
				}{
					Name:      "analyze_code",
					Arguments: params,
				},
			}

			// Call handler
			result, err := server.handleAnalyzeCode(ctx, request)

			if (err != nil) != tt.expectError {
				t.Errorf("handleAnalyzeCode() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError && result == nil {
				t.Error("Expected non-nil result for successful analysis")
			}
		})
	}
}

func TestHandleAnalyzeCode_SingleFile(t *testing.T) {
	// Create server
	cfg := &config.Config{
		MCP: &config.MCPConfig{
			Port:                 3000,
			Host:                 "localhost",
			AutoDiscoverProjects: false,
		},
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tests := []struct {
		name         string
		code         string
		analysisType string
		expectError  bool
	}{
		{
			name: "get types",
			code: `my Int $x = 42;
my Str $name = "test";`,
			analysisType: "get_types",
			expectError:  false,
		},
		{
			name:         "check errors",
			code:         `my Int $x = "not a number";`,
			analysisType: "check_errors",
			expectError:  false,
		},
		{
			name: "infer types",
			code: `my $x = 42;
my $y = "string";`,
			analysisType: "infer_types",
			expectError:  false,
		},
		{
			name:         "invalid analysis type",
			code:         `my $x = 42;`,
			analysisType: "invalid_type",
			expectError:  true,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			params := map[string]interface{}{
				"code":          tt.code,
				"analysis_type": tt.analysisType,
			}

			request := mcp.CallToolRequest{
				Params: struct {
					Name      string    `json:"name"`
					Arguments any       `json:"arguments,omitempty"`
					Meta      *mcp.Meta `json:"_meta,omitempty"`
				}{
					Name:      "analyze_code",
					Arguments: params,
				},
			}

			// Call handler
			result, err := server.handleAnalyzeCode(ctx, request)

			if (err != nil) != tt.expectError {
				t.Errorf("handleAnalyzeCode() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError && result == nil {
				t.Error("Expected non-nil result for successful analysis")
			}
		})
	}
}
