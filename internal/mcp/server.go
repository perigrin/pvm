// ABOUTME: MCP server implementation for PVM
// ABOUTME: Provides Model Context Protocol server with Perl code analysis capabilities

package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"tamarou.com/pvm/internal/config"
)

// Server represents the MCP server for PVM
type Server struct {
	config       *config.MCPConfig
	mcpServer    *server.MCPServer
	projects     map[string]*ProjectContext
	globalConfig *config.Config
}

// ProjectContext holds context information for a discovered Perl project
type ProjectContext struct {
	RootPath    string
	ProjectType string
	HasCpanfile bool
	HasVersion  bool
	PerlVersion string
}

// NewServer creates a new MCP server instance
func NewServer(cfg *config.Config) (*Server, error) {
	if cfg == nil || cfg.MCP == nil {
		return nil, fmt.Errorf("MCP configuration is required")
	}

	// Create MCP server instance
	mcpServer := server.NewMCPServer(
		"pvm-mcp-server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	pvmServer := &Server{
		config:       cfg.MCP,
		mcpServer:    mcpServer,
		projects:     make(map[string]*ProjectContext),
		globalConfig: cfg,
	}

	// Register tool groups
	if err := pvmServer.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	return pvmServer, nil
}

// Start starts the MCP server
func (s *Server) Start(ctx context.Context) error {
	// Auto-discover projects if enabled
	if s.config.AutoDiscoverProjects {
		if err := s.discoverProjects(); err != nil {
			return fmt.Errorf("failed to discover projects: %w", err)
		}
	}

	// Start the MCP server using stdio transport
	// Note: MCP typically uses stdio for communication with LLM clients
	return server.ServeStdio(s.mcpServer)
}

// Stop gracefully stops the MCP server
func (s *Server) Stop(ctx context.Context) error {
	// For now, the stop is handled by context cancellation
	// Future implementations may need cleanup logic here
	return nil
}

// discoverProjects automatically discovers Perl projects in the current directory tree
func (s *Server) discoverProjects() error {
	// Start from current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Walk up the directory tree looking for project indicators
	return s.discoverProjectsInPath(cwd)
}

// discoverProjectsInPath discovers projects starting from the given path
func (s *Server) discoverProjectsInPath(rootPath string) error {
	projectCtx := &ProjectContext{
		RootPath: rootPath,
	}

	// Check for cpanfile
	cpanfilePath := filepath.Join(rootPath, "cpanfile")
	if _, err := os.Stat(cpanfilePath); err == nil {
		projectCtx.HasCpanfile = true
		projectCtx.ProjectType = "cpan"
	}

	// Check for .perl-version
	versionFilePath := filepath.Join(rootPath, ".perl-version")
	if content, err := os.ReadFile(versionFilePath); err == nil {
		projectCtx.HasVersion = true
		projectCtx.PerlVersion = strings.TrimSpace(string(content))
	}

	// Check for other Perl project indicators
	if !projectCtx.HasCpanfile {
		// Look for .pm files or common Perl directory structures
		if s.hasPerlFiles(rootPath) {
			projectCtx.ProjectType = "perl"
		}
	}

	// Only register if we found project indicators
	if projectCtx.HasCpanfile || projectCtx.HasVersion || projectCtx.ProjectType != "" {
		s.projects[rootPath] = projectCtx
	}

	return nil
}

// hasPerlFiles checks if the directory contains Perl files or standard Perl directories
func (s *Server) hasPerlFiles(path string) bool {
	// Check for common Perl file extensions and directories
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files and common non-Perl directories
		if strings.HasPrefix(name, ".") ||
			name == "node_modules" ||
			name == "target" ||
			name == "build" {
			continue
		}

		if entry.IsDir() {
			// Check for common Perl directories
			if name == "lib" || name == "t" || name == "bin" || name == "script" {
				return true
			}
		} else {
			// Check for Perl file extensions
			if strings.HasSuffix(name, ".pl") ||
				strings.HasSuffix(name, ".pm") ||
				strings.HasSuffix(name, ".t") {
				return true
			}
		}
	}

	return false
}

// registerTools registers the three main tool groups with the MCP server
func (s *Server) registerTools() error {
	// Register analyze_code tool group
	if err := s.registerAnalyzeTools(); err != nil {
		return fmt.Errorf("failed to register analyze tools: %w", err)
	}

	// Register search_code tool group
	if err := s.registerSearchTools(); err != nil {
		return fmt.Errorf("failed to register search tools: %w", err)
	}

	// Register generate_code tool group
	if err := s.registerGenerateTools(); err != nil {
		return fmt.Errorf("failed to register generate tools: %w", err)
	}

	return nil
}

// registerAnalyzeTools registers code analysis tools
func (s *Server) registerAnalyzeTools() error {
	// Create analyze_code tool using the correct API
	analyzeCodeTool := mcp.NewTool("analyze_code",
		mcp.WithDescription("Analyze Perl code for types, errors, and inference"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("Perl code to analyze")),
		mcp.WithString("analysis_type",
			mcp.Required(),
			mcp.Description("Type of analysis: get_types, check_errors, infer_types"),
			mcp.Enum("get_types", "check_errors", "infer_types")),
		mcp.WithString("project_path",
			mcp.Description("Optional project path for context")),
	)

	s.mcpServer.AddTool(analyzeCodeTool, s.handleAnalyzeCode)
	return nil
}

// registerSearchTools registers code search tools
func (s *Server) registerSearchTools() error {
	searchCodeTool := mcp.NewTool("search_code",
		mcp.WithDescription("Search for code using semantic similarity, type signatures, or patterns"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query")),
		mcp.WithString("search_method",
			mcp.Required(),
			mcp.Description("Search method: similarity, type_signature, pattern"),
			mcp.Enum("similarity", "type_signature", "pattern")),
		mcp.WithString("project_path",
			mcp.Description("Optional project path to search within")),
	)

	s.mcpServer.AddTool(searchCodeTool, s.handleSearchCode)
	return nil
}

// registerGenerateTools registers code generation tools
func (s *Server) registerGenerateTools() error {
	generateCodeTool := mcp.NewTool("generate_code",
		mcp.WithDescription("Generate Perl code using collaborative sampling with LLM"),
		mcp.WithString("generation_type",
			mcp.Required(),
			mcp.Description("Type of code to generate: function, class, test"),
			mcp.Enum("function", "class", "test")),
		mcp.WithString("specification",
			mcp.Required(),
			mcp.Description("Specification or description of what to generate")),
		mcp.WithString("context",
			mcp.Description("Optional context code")),
		mcp.WithString("project_path",
			mcp.Description("Optional project path for context")),
	)

	s.mcpServer.AddTool(generateCodeTool, s.handleGenerateCode)
	return nil
}

// Tool handler implementations (placeholders for now)

func (s *Server) handleAnalyzeCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Placeholder implementation
	// This will be implemented in Step 4
	code, _ := request.RequireString("code")
	analysisType, _ := request.RequireString("analysis_type")

	result := map[string]interface{}{
		"status":        "not_implemented",
		"message":       "analyze_code tool will be implemented in Step 4",
		"code_length":   len(code),
		"analysis_type": analysisType,
	}

	return mcp.NewToolResultText(fmt.Sprintf("Analyze code result: %v", result)), nil
}

func (s *Server) handleSearchCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Placeholder implementation
	// This will be implemented in Step 10
	query, _ := request.RequireString("query")
	searchMethod, _ := request.RequireString("search_method")

	result := map[string]interface{}{
		"status":        "not_implemented",
		"message":       "search_code tool will be implemented in Step 10",
		"query":         query,
		"search_method": searchMethod,
	}

	return mcp.NewToolResultText(fmt.Sprintf("Search code result: %v", result)), nil
}

func (s *Server) handleGenerateCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Placeholder implementation
	// This will be implemented in Step 12
	generationType, _ := request.RequireString("generation_type")
	specification, _ := request.RequireString("specification")

	result := map[string]interface{}{
		"status":          "not_implemented",
		"message":         "generate_code tool will be implemented in Step 12",
		"generation_type": generationType,
		"specification":   specification,
	}

	return mcp.NewToolResultText(fmt.Sprintf("Generate code result: %v", result)), nil
}

// GetProjects returns the discovered projects
func (s *Server) GetProjects() map[string]*ProjectContext {
	return s.projects
}

// GetConfig returns the MCP configuration
func (s *Server) GetConfig() *config.MCPConfig {
	return s.config
}
