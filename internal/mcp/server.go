// ABOUTME: MCP server implementation for PVM
// ABOUTME: Provides Model Context Protocol server with Perl code analysis capabilities

package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/mcp/generation"
	"tamarou.com/pvm/internal/mcp/tools"
	"tamarou.com/pvm/internal/mcp/validation"
)

// Server represents the MCP server for PVM
type Server struct {
	config         *config.MCPConfig
	mcpServer      *server.MCPServer
	projects       map[string]*ProjectContext
	globalConfig   *config.Config
	metrics        *ServerMetrics
	logger         *log.Logger
	validator      *validation.Validator
	autoFixer      *validation.AutoFixer
	samplingClient *generation.SamplingClient
	codeAnalyzer   *tools.CodeAnalyzer
}

// ServerMetrics tracks server performance and usage statistics
type ServerMetrics struct {
	mu              sync.RWMutex
	RequestCount    int64            `json:"request_count"`
	ErrorCount      int64            `json:"error_count"`
	ToolUsageCount  map[string]int64 `json:"tool_usage_count"`
	AverageLatency  time.Duration    `json:"average_latency"`
	LastRequestTime time.Time        `json:"last_request_time"`
	StartTime       time.Time        `json:"start_time"`
}

// NewServerMetrics creates a new ServerMetrics instance
func NewServerMetrics() *ServerMetrics {
	return &ServerMetrics{
		ToolUsageCount: make(map[string]int64),
		StartTime:      time.Now(),
	}
}

// MetricsSnapshot represents a read-only snapshot of server metrics
type MetricsSnapshot struct {
	RequestCount    int64            `json:"request_count"`
	ErrorCount      int64            `json:"error_count"`
	ToolUsageCount  map[string]int64 `json:"tool_usage_count"`
	AverageLatency  time.Duration    `json:"average_latency"`
	LastRequestTime time.Time        `json:"last_request_time"`
	StartTime       time.Time        `json:"start_time"`
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

	// Create logger
	logger := log.NewLogger(log.LevelInfo, os.Stderr, "mcp-server")

	// Create validation cache
	cacheSize := cfg.MCP.ValidationCacheSize
	if cacheSize == "" {
		cacheSize = "50MB"
	}
	validationCache, err := validation.NewValidationCache(cacheSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create validation cache: %w", err)
	}

	// Create validator
	validator, err := validation.NewValidator(validationCache)
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}

	// Create sampling client
	samplingClient := generation.NewSamplingClient(cfg.MCP.EnableIterativeRefinement)

	// Create auto-fixer
	autoFixer := validation.NewAutoFixer(validator, samplingClient, cfg.MCP.AutoFixErrors)

	// Create code analyzer
	codeAnalyzer, err := tools.NewCodeAnalyzer(validator, autoFixer)
	if err != nil {
		return nil, fmt.Errorf("failed to create code analyzer: %w", err)
	}

	pvmServer := &Server{
		config:         cfg.MCP,
		mcpServer:      mcpServer,
		projects:       make(map[string]*ProjectContext),
		globalConfig:   cfg,
		metrics:        NewServerMetrics(),
		logger:         logger,
		validator:      validator,
		autoFixer:      autoFixer,
		samplingClient: samplingClient,
		codeAnalyzer:   codeAnalyzer,
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
	startTime := time.Now()
	defer s.recordToolUsage("analyze_code", startTime)

	// Log request
	s.logger.Debugf("handleAnalyzeCode called")

	// Validate and extract parameters
	code, err := request.RequireString("code")
	if err != nil {
		s.recordError()
		s.logger.Errorf("Failed to get 'code' parameter: %v", err)
		return nil, fmt.Errorf("missing or invalid 'code' parameter: %w", err)
	}

	analysisType, err := request.RequireString("analysis_type")
	if err != nil {
		s.recordError()
		s.logger.Errorf("Failed to get 'analysis_type' parameter: %v", err)
		return nil, fmt.Errorf("missing or invalid 'analysis_type' parameter: %w", err)
	}

	// Validate analysis type
	validTypes := map[string]bool{
		"get_types":    true,
		"check_errors": true,
		"infer_types":  true,
	}
	if !validTypes[analysisType] {
		s.recordError()
		s.logger.Errorf("Invalid analysis_type: %s", analysisType)
		return nil, fmt.Errorf("invalid analysis_type '%s', must be one of: get_types, check_errors, infer_types", analysisType)
	}

	// Get optional project path
	projectPath := request.GetString("project_path", "")

	s.logger.Infof("Analyzing code: type=%s, length=%d, project=%s", analysisType, len(code), projectPath)

	// Use the code analyzer
	analysisResult, err := s.codeAnalyzer.Analyze(ctx, code, analysisType, projectPath, s.config.AutoFixErrors)
	if err != nil {
		s.recordError()
		s.logger.Errorf("Analysis failed: %v", err)
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	// Convert to JSON-friendly format
	result := map[string]interface{}{
		"status":        analysisResult.Status,
		"analysis_type": analysisResult.AnalysisType,
		"valid":         analysisResult.Valid,
		"timestamp":     analysisResult.Timestamp,
	}

	// Add type-specific fields
	if len(analysisResult.TypeInfo) > 0 {
		result["type_info"] = analysisResult.TypeInfo
	}
	if len(analysisResult.Errors) > 0 {
		result["errors"] = analysisResult.Errors
	}
	if len(analysisResult.Warnings) > 0 {
		result["warnings"] = analysisResult.Warnings
	}
	if len(analysisResult.Fixes) > 0 {
		result["fixes"] = analysisResult.Fixes
	}
	if len(analysisResult.InferredTypes) > 0 {
		result["inferred_types"] = analysisResult.InferredTypes
	}

	return mcp.NewToolResultText(fmt.Sprintf("%v", result)), nil
}

func (s *Server) handleSearchCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	defer s.recordToolUsage("search_code", startTime)

	// Log request
	s.logger.Debugf("handleSearchCode called")

	// Validate and extract parameters
	query, err := request.RequireString("query")
	if err != nil {
		s.recordError()
		s.logger.Errorf("Failed to get 'query' parameter: %v", err)
		return nil, fmt.Errorf("missing or invalid 'query' parameter: %w", err)
	}

	searchMethod, err := request.RequireString("search_method")
	if err != nil {
		s.recordError()
		s.logger.Errorf("Failed to get 'search_method' parameter: %v", err)
		return nil, fmt.Errorf("missing or invalid 'search_method' parameter: %w", err)
	}

	// Validate search method
	validMethods := map[string]bool{
		"similarity":     true,
		"type_signature": true,
		"pattern":        true,
	}
	if !validMethods[searchMethod] {
		s.recordError()
		s.logger.Errorf("Invalid search_method: %s", searchMethod)
		return nil, fmt.Errorf("invalid search_method '%s', must be one of: similarity, type_signature, pattern", searchMethod)
	}

	// Get optional project path
	projectPath := request.GetString("project_path", "")

	s.logger.Infof("Searching code: method=%s, query=%s, project=%s", searchMethod, query, projectPath)

	// Placeholder implementation - will be implemented in Step 10
	result := map[string]interface{}{
		"status":        "not_implemented",
		"message":       "search_code tool will be implemented in Step 10",
		"query":         query,
		"search_method": searchMethod,
		"project_path":  projectPath,
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	}

	return mcp.NewToolResultText(fmt.Sprintf("Search code result: %v", result)), nil
}

func (s *Server) handleGenerateCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	defer s.recordToolUsage("generate_code", startTime)

	// Log request
	s.logger.Debugf("handleGenerateCode called")

	// Validate and extract parameters
	generationType, err := request.RequireString("generation_type")
	if err != nil {
		s.recordError()
		s.logger.Errorf("Failed to get 'generation_type' parameter: %v", err)
		return nil, fmt.Errorf("missing or invalid 'generation_type' parameter: %w", err)
	}

	specification, err := request.RequireString("specification")
	if err != nil {
		s.recordError()
		s.logger.Errorf("Failed to get 'specification' parameter: %v", err)
		return nil, fmt.Errorf("missing or invalid 'specification' parameter: %w", err)
	}

	// Validate generation type
	validTypes := map[string]bool{
		"function": true,
		"class":    true,
		"test":     true,
	}
	if !validTypes[generationType] {
		s.recordError()
		s.logger.Errorf("Invalid generation_type: %s", generationType)
		return nil, fmt.Errorf("invalid generation_type '%s', must be one of: function, class, test", generationType)
	}

	// Get optional parameters
	context := request.GetString("context", "")
	projectPath := request.GetString("project_path", "")

	s.logger.Infof("Generating code: type=%s, spec_length=%d, project=%s", generationType, len(specification), projectPath)

	// Placeholder implementation - will be implemented in Step 12
	result := map[string]interface{}{
		"status":          "not_implemented",
		"message":         "generate_code tool will be implemented in Step 12",
		"generation_type": generationType,
		"specification":   specification,
		"context":         context,
		"project_path":    projectPath,
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
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

// recordToolUsage records metrics for tool usage
func (s *Server) recordToolUsage(toolName string, startTime time.Time) {
	duration := time.Since(startTime)

	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()

	// Increment request count
	atomic.AddInt64(&s.metrics.RequestCount, 1)

	// Increment tool-specific usage count
	s.metrics.ToolUsageCount[toolName]++

	// Update average latency (simple moving average)
	if s.metrics.AverageLatency == 0 {
		s.metrics.AverageLatency = duration
	} else {
		// Simple weighted average - could be improved with more sophisticated algorithms
		s.metrics.AverageLatency = (s.metrics.AverageLatency + duration) / 2
	}

	// Update last request time
	s.metrics.LastRequestTime = time.Now()

	s.logger.Debugf("Tool usage recorded: %s took %v", toolName, duration)
}

// recordError records an error occurrence
func (s *Server) recordError() {
	atomic.AddInt64(&s.metrics.ErrorCount, 1)
}

// GetMetrics returns a snapshot of the current metrics
func (s *Server) GetMetrics() MetricsSnapshot {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	// Create a snapshot to avoid data races
	snapshot := MetricsSnapshot{
		RequestCount:    atomic.LoadInt64(&s.metrics.RequestCount),
		ErrorCount:      atomic.LoadInt64(&s.metrics.ErrorCount),
		ToolUsageCount:  make(map[string]int64),
		AverageLatency:  s.metrics.AverageLatency,
		LastRequestTime: s.metrics.LastRequestTime,
		StartTime:       s.metrics.StartTime,
	}

	// Copy tool usage counts
	for tool, count := range s.metrics.ToolUsageCount {
		snapshot.ToolUsageCount[tool] = count
	}

	return snapshot
}
