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
	"tamarou.com/pvm/internal/mcp/embeddings"
	"tamarou.com/pvm/internal/mcp/generation"
	"tamarou.com/pvm/internal/mcp/tools"
	"tamarou.com/pvm/internal/mcp/validation"
)

// Server represents the MCP server for PVM
type Server struct {
	config            *config.MCPConfig
	mcpServer         *server.MCPServer
	projects          map[string]*ProjectContext
	globalConfig      *config.Config
	metrics           *ServerMetrics
	logger            *log.Logger
	validator         *validation.Validator
	autoFixer         *validation.AutoFixer
	samplingClient    *generation.SamplingClient
	memoryManager     *generation.MemoryManager
	codeAnalyzer      *tools.CodeAnalyzer
	projectAnalyzer   *tools.ProjectAnalyzer
	embeddingStore    *embeddings.EmbeddingStore
	embeddingManager  *embeddings.CollectionManager
	codeExtractor     *embeddings.Extractor
	codeSearcher      *tools.CodeSearcher
	codeGenerator     *tools.CodeGenerator
	advancedGenerator *tools.AdvancedGenerator
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

	// Create memory manager
	memorySize := cfg.MCP.GenerationMemorySize
	if memorySize == 0 {
		memorySize = 50 // Default memory size
	}
	memoryManager := generation.NewMemoryManager(memorySize)

	// Create auto-fixer
	autoFixer := validation.NewAutoFixer(validator, samplingClient, cfg.MCP.AutoFixErrors)

	// Create code analyzer
	codeAnalyzer, err := tools.NewCodeAnalyzer(validator, autoFixer)
	if err != nil {
		return nil, fmt.Errorf("failed to create code analyzer: %w", err)
	}

	// Create project analyzer
	projectAnalyzer, err := tools.NewProjectAnalyzer(codeAnalyzer, validator)
	if err != nil {
		return nil, fmt.Errorf("failed to create project analyzer: %w", err)
	}

	// Create embedding provider
	embeddingConfig := embeddings.EmbeddingConfig{
		Provider:   cfg.MCP.EmbeddingProvider,
		Model:      cfg.MCP.EmbeddingModel,
		Dimensions: 384, // Default dimensions
		MaxTokens:  512, // Default max tokens
		BatchSize:  100, // Default batch size
	}

	embeddingProvider, err := embeddings.NewProvider(embeddingConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding provider: %w", err)
	}

	// Create embedding store
	embeddingStore, err := embeddings.NewEmbeddingStore(embeddings.StoreConfig{
		Provider: embeddingProvider,
		Logger:   logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding store: %w", err)
	}

	// Create collection manager
	embeddingManager := embeddings.NewCollectionManager(embeddingStore, logger)

	// Create code extractor
	codeExtractor, err := embeddings.NewExtractor()
	if err != nil {
		return nil, fmt.Errorf("failed to create code extractor: %w", err)
	}

	// Create code searcher
	codeSearcher := tools.NewCodeSearcher(embeddingStore, embeddingManager, codeExtractor, logger)

	// Create code generator
	codeGenerator := tools.NewCodeGenerator(validator, autoFixer, samplingClient, memoryManager, logger)

	// Create advanced generator
	advancedGenerator := tools.NewAdvancedGenerator(validator, autoFixer, samplingClient, memoryManager, logger)

	pvmServer := &Server{
		config:            cfg.MCP,
		mcpServer:         mcpServer,
		projects:          make(map[string]*ProjectContext),
		globalConfig:      cfg,
		metrics:           NewServerMetrics(),
		logger:            logger,
		validator:         validator,
		autoFixer:         autoFixer,
		samplingClient:    samplingClient,
		memoryManager:     memoryManager,
		codeAnalyzer:      codeAnalyzer,
		projectAnalyzer:   projectAnalyzer,
		embeddingStore:    embeddingStore,
		embeddingManager:  embeddingManager,
		codeExtractor:     codeExtractor,
		codeSearcher:      codeSearcher,
		codeGenerator:     codeGenerator,
		advancedGenerator: advancedGenerator,
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
	// Clean up memory manager
	if s.memoryManager != nil {
		s.memoryManager.Close()
	}

	// Close embedding store if it has cleanup methods
	// (Note: Current implementation doesn't require explicit cleanup)

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

	// Register memory management tools
	if err := s.registerMemoryTools(); err != nil {
		return fmt.Errorf("failed to register memory tools: %w", err)
	}

	// Register advanced generation tools
	if err := s.registerAdvancedGenerationTools(); err != nil {
		return fmt.Errorf("failed to register advanced generation tools: %w", err)
	}

	return nil
}

// registerAnalyzeTools registers code analysis tools
func (s *Server) registerAnalyzeTools() error {
	// Create analyze_code tool using the correct API
	analyzeCodeTool := mcp.NewTool("analyze_code",
		mcp.WithDescription("Analyze Perl code for types, errors, and inference"),
		mcp.WithString("code",
			mcp.Description("Perl code to analyze (required for single-file analysis)")),
		mcp.WithString("analysis_type",
			mcp.Required(),
			mcp.Description("Type of analysis: get_types, check_errors, infer_types, project_analysis, project_summary"),
			mcp.Enum("get_types", "check_errors", "infer_types", "project_analysis", "project_summary")),
		mcp.WithString("project_path",
			mcp.Description("Project path (required for project_analysis and project_summary)")),
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
		mcp.WithString("session_id",
			mcp.Description("Optional generation session ID for memory continuity")),
	)

	s.mcpServer.AddTool(generateCodeTool, s.handleGenerateCode)
	return nil
}

// registerMemoryTools registers generation memory management tools
func (s *Server) registerMemoryTools() error {
	// Create memory_session tool for managing generation memory
	memorySessionTool := mcp.NewTool("memory_session",
		mcp.WithDescription("Manage generation memory sessions for maintaining context during code generation"),
		mcp.WithString("action",
			mcp.Required(),
			mcp.Description("Action to perform: create, get, clear, stats"),
			mcp.Enum("create", "get", "clear", "stats")),
		mcp.WithString("session_id",
			mcp.Required(),
			mcp.Description("Unique identifier for the generation session")),
		mcp.WithString("type_choice_name",
			mcp.Description("Variable/function name for type choice operations")),
		mcp.WithString("type_choice_value",
			mcp.Description("Type value for type choice operations")),
		mcp.WithString("naming_pattern_type",
			mcp.Description("Pattern type for naming convention operations")),
		mcp.WithString("naming_pattern_value",
			mcp.Description("Naming convention value for pattern operations")),
		mcp.WithString("decision_type",
			mcp.Description("Type of decision being recorded")),
		mcp.WithString("decision_context",
			mcp.Description("Context for the decision")),
		mcp.WithString("decision_choice",
			mcp.Description("Choice made for the decision")),
		mcp.WithString("decision_rationale",
			mcp.Description("Rationale for the decision")),
	)

	s.mcpServer.AddTool(memorySessionTool, s.handleMemorySession)
	return nil
}

// registerAdvancedGenerationTools registers advanced code generation tools
func (s *Server) registerAdvancedGenerationTools() error {
	// Test generation from types
	testGenTool := mcp.NewTool("generate_tests_from_types",
		mcp.WithDescription("Generate comprehensive test suites from type signatures"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("Code to generate tests for")),
		mcp.WithObject("type_signatures",
			mcp.Description("Map of function/method names to type signatures")),
		mcp.WithString("framework",
			mcp.Description("Test framework to use (Test2::V0, Test::More, Test::Most)"),
			mcp.Enum("Test2::V0", "Test::More", "Test::Most")),
		mcp.WithString("session_id",
			mcp.Description("Optional generation session ID for memory continuity")),
	)
	s.mcpServer.AddTool(testGenTool, s.handleGenerateTestsFromTypes)

	// Code refactoring
	refactorTool := mcp.NewTool("refactor_code",
		mcp.WithDescription("Perform type-preserving code refactoring"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("Code to refactor")),
		mcp.WithString("refactoring_type",
			mcp.Required(),
			mcp.Description("Type of refactoring to perform"),
			mcp.Enum("extract_method", "rename", "inline")),
		mcp.WithString("target",
			mcp.Required(),
			mcp.Description("Target element to refactor")),
		mcp.WithString("new_name",
			mcp.Description("New name (required for rename operations)")),
		mcp.WithString("session_id",
			mcp.Description("Optional generation session ID")),
	)
	s.mcpServer.AddTool(refactorTool, s.handleRefactorCode)

	// Documentation generation
	docGenTool := mcp.NewTool("generate_documentation",
		mcp.WithDescription("Generate documentation from typed code"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("Code to document")),
		mcp.WithString("style",
			mcp.Required(),
			mcp.Description("Documentation style"),
			mcp.Enum("pod", "markdown", "inline")),
		mcp.WithString("session_id",
			mcp.Description("Optional generation session ID")),
	)
	s.mcpServer.AddTool(docGenTool, s.handleGenerateDocumentation)

	// Code completion
	completionTool := mcp.NewTool("complete_code",
		mcp.WithDescription("Provide intelligent code completion suggestions"),
		mcp.WithString("partial_code",
			mcp.Required(),
			mcp.Description("Incomplete code to complete")),
		mcp.WithNumber("cursor_position",
			mcp.Required(),
			mcp.Description("Cursor position in the code")),
		mcp.WithString("context",
			mcp.Description("Surrounding code context")),
		mcp.WithString("session_id",
			mcp.Description("Optional generation session ID")),
	)
	s.mcpServer.AddTool(completionTool, s.handleCompleteCode)

	// Batch generation
	batchGenTool := mcp.NewTool("batch_generate",
		mcp.WithDescription("Handle multiple generation requests efficiently"),
		mcp.WithArray("requests",
			mcp.Required(),
			mcp.Description("Array of generation requests")),
		mcp.WithString("session_id",
			mcp.Description("Shared session ID for all requests")),
	)
	s.mcpServer.AddTool(batchGenTool, s.handleBatchGenerate)

	return nil
}

// Tool handler implementations

func (s *Server) handleAnalyzeCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	defer s.recordToolUsage("analyze_code", startTime)

	// Log request
	s.logger.Debugf("handleAnalyzeCode called")

	// Validate and extract parameters
	analysisType, err := request.RequireString("analysis_type")
	if err != nil {
		s.recordError()
		s.logger.Errorf("Failed to get 'analysis_type' parameter: %v", err)
		return nil, fmt.Errorf("missing or invalid 'analysis_type' parameter: %w", err)
	}

	// Validate analysis type
	validTypes := map[string]bool{
		"get_types":        true,
		"check_errors":     true,
		"infer_types":      true,
		"project_analysis": true,
		"project_summary":  true,
	}
	if !validTypes[analysisType] {
		s.recordError()
		s.logger.Errorf("Invalid analysis_type: %s", analysisType)
		return nil, fmt.Errorf("invalid analysis_type '%s', must be one of: get_types, check_errors, infer_types, project_analysis, project_summary", analysisType)
	}

	// Handle project-wide analysis types
	if analysisType == "project_analysis" || analysisType == "project_summary" {
		projectPath := request.GetString("project_path", "")
		if projectPath == "" {
			// Try to use current working directory
			cwd, err := os.Getwd()
			if err != nil {
				s.recordError()
				return nil, fmt.Errorf("project_path is required for %s", analysisType)
			}
			projectPath = cwd
		}

		if analysisType == "project_analysis" {
			s.logger.Infof("Performing project analysis for: %s", projectPath)
			projectAnalysis, err := s.projectAnalyzer.AnalyzeProject(ctx, projectPath)
			if err != nil {
				s.recordError()
				s.logger.Errorf("Project analysis failed: %v", err)
				return nil, fmt.Errorf("project analysis failed: %w", err)
			}
			return mcp.NewToolResultText(fmt.Sprintf("%v", projectAnalysis)), nil
		} else { // project_summary
			s.logger.Infof("Getting project summary for: %s", projectPath)
			summary, err := s.projectAnalyzer.GetProjectSummary(projectPath)
			if err != nil {
				s.recordError()
				s.logger.Errorf("Failed to get project summary: %v", err)
				return nil, fmt.Errorf("failed to get project summary: %w", err)
			}
			return mcp.NewToolResultText(fmt.Sprintf("%v", summary)), nil
		}
	}

	// For single-file analysis types, code parameter is required
	code, err := request.RequireString("code")
	if err != nil {
		s.recordError()
		s.logger.Errorf("Failed to get 'code' parameter: %v", err)
		return nil, fmt.Errorf("missing or invalid 'code' parameter: %w", err)
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
	if projectPath == "" {
		// Try to use current working directory
		cwd, err := os.Getwd()
		if err != nil {
			s.recordError()
			return nil, fmt.Errorf("project_path is required for search")
		}
		projectPath = cwd
	}

	s.logger.Infof("Searching code: method=%s, query=%s, project=%s", searchMethod, query, projectPath)

	// Create search request
	searchRequest := tools.SearchRequest{
		Query:         query,
		Method:        searchMethod,
		ProjectPath:   projectPath,
		TopK:          20,  // Default top K
		MinSimilarity: 0.3, // Default minimum similarity
	}

	// Perform search
	searchResponse, err := s.codeSearcher.Search(ctx, searchRequest)
	if err != nil {
		s.recordError()
		s.logger.Errorf("Search failed: %v", err)
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("%v", searchResponse)), nil
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
	sessionID := request.GetString("session_id", fmt.Sprintf("gen_%d", time.Now().UnixNano()))

	s.logger.Infof("Generating code: type=%s, spec_length=%d, project=%s, session=%s",
		generationType, len(specification), projectPath, sessionID)

	// Use the code generator to perform collaborative generation
	generationRequest := tools.GenerationRequest{
		Type:          generationType,
		Specification: specification,
		Context:       context,
		ProjectPath:   projectPath,
		SessionID:     sessionID,
	}

	generationResult, err := s.codeGenerator.Generate(generationRequest)
	if err != nil {
		s.recordError()
		s.logger.Errorf("Code generation failed: %v", err)
		return nil, fmt.Errorf("code generation failed: %w", err)
	}

	// Convert to MCP response format
	result := map[string]interface{}{
		"status":            generationResult.Status,
		"generated_code":    generationResult.GeneratedCode,
		"validation_result": generationResult.ValidationResult,
		"memory_context":    generationResult.MemoryContext,
		"iterations":        generationResult.Iterations,
		"decisions":         generationResult.Decisions,
		"message":           generationResult.Message,
		"generation_type":   generationType,
		"specification":     specification,
		"context":           context,
		"project_path":      projectPath,
		"session_id":        sessionID,
		"timestamp":         generationResult.Timestamp,
	}

	return mcp.NewToolResultText(fmt.Sprintf("Generate code result: %v", result)), nil
}

func (s *Server) handleMemorySession(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	defer s.recordToolUsage("memory_session", startTime)

	// Log request
	s.logger.Debugf("handleMemorySession called")

	// Validate and extract required parameters
	action, err := request.RequireString("action")
	if err != nil {
		s.recordError()
		s.logger.Errorf("Failed to get 'action' parameter: %v", err)
		return nil, fmt.Errorf("missing or invalid 'action' parameter: %w", err)
	}

	sessionID, err := request.RequireString("session_id")
	if err != nil {
		s.recordError()
		s.logger.Errorf("Failed to get 'session_id' parameter: %v", err)
		return nil, fmt.Errorf("missing or invalid 'session_id' parameter: %w", err)
	}

	s.logger.Infof("Memory session action: %s for session: %s", action, sessionID)

	var result map[string]interface{}

	switch action {
	case "create":
		// Create a new memory session
		memory := s.memoryManager.CreateSession(sessionID)
		result = map[string]interface{}{
			"status":     "success",
			"action":     "create",
			"session_id": sessionID,
			"message":    "Memory session created successfully",
			"context":    memory.GetContext(),
		}

	case "get":
		// Get memory session context and recent decisions
		memory := s.memoryManager.GetSession(sessionID)
		result = map[string]interface{}{
			"status":           "success",
			"action":           "get",
			"session_id":       sessionID,
			"context":          memory.GetContext(),
			"recent_decisions": memory.GetRecentDecisions(30), // Last 30 minutes
			"stats":            memory.SessionStats(),
		}

		// Handle optional type choice operations
		if typeChoiceName := request.GetString("type_choice_name", ""); typeChoiceName != "" {
			if typeChoiceValue := request.GetString("type_choice_value", ""); typeChoiceValue != "" {
				// Set type choice
				memory.SetTypeChoice(typeChoiceName, typeChoiceValue)
				result["type_choice_set"] = map[string]string{
					"name": typeChoiceName,
					"type": typeChoiceValue,
				}
			} else {
				// Get type choice
				if typeStr, exists := memory.GetTypeChoice(typeChoiceName); exists {
					result["type_choice"] = map[string]string{
						"name": typeChoiceName,
						"type": typeStr,
					}
				}
			}
		}

		// Handle optional naming pattern operations
		if namingPatternType := request.GetString("naming_pattern_type", ""); namingPatternType != "" {
			if namingPatternValue := request.GetString("naming_pattern_value", ""); namingPatternValue != "" {
				// Set naming pattern
				memory.SetNamingPattern(namingPatternType, namingPatternValue)
				result["naming_pattern_set"] = map[string]string{
					"pattern_type": namingPatternType,
					"convention":   namingPatternValue,
				}
			} else {
				// Get naming pattern
				if pattern, exists := memory.GetNamingPattern(namingPatternType); exists {
					result["naming_pattern"] = map[string]string{
						"pattern_type": namingPatternType,
						"convention":   pattern,
					}
				}
			}
		}

		// Handle optional decision recording
		if decisionType := request.GetString("decision_type", ""); decisionType != "" {
			decisionContext := request.GetString("decision_context", "")
			decisionChoice := request.GetString("decision_choice", "")
			decisionRationale := request.GetString("decision_rationale", "")

			if decisionContext != "" && decisionChoice != "" {
				memory.AddDecision(decisionType, decisionContext, decisionChoice, decisionRationale)
				result["decision_recorded"] = map[string]string{
					"type":      decisionType,
					"context":   decisionContext,
					"choice":    decisionChoice,
					"rationale": decisionRationale,
				}
			}
		}

	case "clear":
		// Clear memory session
		s.memoryManager.ClearSession(sessionID)
		result = map[string]interface{}{
			"status":     "success",
			"action":     "clear",
			"session_id": sessionID,
			"message":    "Memory session cleared successfully",
		}

	case "stats":
		// Get memory session statistics
		memory := s.memoryManager.GetSession(sessionID)
		result = map[string]interface{}{
			"status":     "success",
			"action":     "stats",
			"session_id": sessionID,
			"stats":      memory.SessionStats(),
		}

	default:
		s.recordError()
		s.logger.Errorf("Invalid action: %s", action)
		return nil, fmt.Errorf("invalid action '%s', must be one of: create, get, clear, stats", action)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Memory session result: %v", result)), nil
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

// Advanced generation tool handlers

func (s *Server) handleGenerateTestsFromTypes(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	defer s.recordToolUsage("generate_tests_from_types", startTime)

	// Extract parameters
	code, err := request.RequireString("code")
	if err != nil {
		s.recordError()
		return nil, fmt.Errorf("missing or invalid 'code' parameter: %w", err)
	}

	// Get type signatures (optional)
	typeSigs := make(map[string]string)
	args := request.GetArguments()
	if typeSigsRaw, exists := args["type_signatures"]; exists {
		if typeSigsMap, ok := typeSigsRaw.(map[string]interface{}); ok {
			for k, v := range typeSigsMap {
				if str, ok := v.(string); ok {
					typeSigs[k] = str
				}
			}
		}
	}

	framework := request.GetString("framework", "Test2::V0")
	sessionID := request.GetString("session_id", fmt.Sprintf("test_gen_%d", time.Now().UnixNano()))

	// Create request
	testGenRequest := tools.TestGenerationRequest{
		Code:        code,
		TypeSigs:    typeSigs,
		Framework:   framework,
		SessionID:   sessionID,
		ProjectPath: "",
	}

	// Generate tests
	result, err := s.advancedGenerator.GenerateTestsFromTypes(ctx, testGenRequest)
	if err != nil {
		s.recordError()
		return nil, fmt.Errorf("test generation failed: %w", err)
	}

	// Convert to MCP response
	response := map[string]interface{}{
		"status":         result.Status,
		"generated_code": result.GeneratedCode,
		"iterations":     result.Iterations,
		"message":        result.Message,
		"session_id":     sessionID,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	}

	return mcp.NewToolResultText(fmt.Sprintf("Test generation result: %v", response)), nil
}

func (s *Server) handleRefactorCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	defer s.recordToolUsage("refactor_code", startTime)

	// Extract parameters
	code, err := request.RequireString("code")
	if err != nil {
		s.recordError()
		return nil, fmt.Errorf("missing or invalid 'code' parameter: %w", err)
	}

	refactoringType, err := request.RequireString("refactoring_type")
	if err != nil {
		s.recordError()
		return nil, fmt.Errorf("missing or invalid 'refactoring_type' parameter: %w", err)
	}

	target, err := request.RequireString("target")
	if err != nil {
		s.recordError()
		return nil, fmt.Errorf("missing or invalid 'target' parameter: %w", err)
	}

	newName := request.GetString("new_name", "")
	sessionID := request.GetString("session_id", fmt.Sprintf("refactor_%d", time.Now().UnixNano()))

	// Create request
	refactorRequest := tools.RefactoringRequest{
		Code:            code,
		RefactoringType: refactoringType,
		Target:          target,
		NewName:         newName,
		SessionID:       sessionID,
	}

	// Perform refactoring
	result, err := s.advancedGenerator.RefactorCode(ctx, refactorRequest)
	if err != nil {
		s.recordError()
		return nil, fmt.Errorf("refactoring failed: %w", err)
	}

	// Convert to MCP response
	response := map[string]interface{}{
		"status":          result.Status,
		"refactored_code": result.GeneratedCode,
		"iterations":      result.Iterations,
		"message":         result.Message,
		"session_id":      sessionID,
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
	}

	if result.ValidationResult != nil {
		response["validation_result"] = result.ValidationResult
	}

	return mcp.NewToolResultText(fmt.Sprintf("Refactoring result: %v", response)), nil
}

func (s *Server) handleGenerateDocumentation(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	defer s.recordToolUsage("generate_documentation", startTime)

	// Extract parameters
	code, err := request.RequireString("code")
	if err != nil {
		s.recordError()
		return nil, fmt.Errorf("missing or invalid 'code' parameter: %w", err)
	}

	style, err := request.RequireString("style")
	if err != nil {
		s.recordError()
		return nil, fmt.Errorf("missing or invalid 'style' parameter: %w", err)
	}

	sessionID := request.GetString("session_id", fmt.Sprintf("doc_%d", time.Now().UnixNano()))

	// Create request
	docRequest := tools.DocumentationRequest{
		Code:      code,
		Style:     style,
		SessionID: sessionID,
	}

	// Generate documentation
	result, err := s.advancedGenerator.GenerateDocumentation(ctx, docRequest)
	if err != nil {
		s.recordError()
		return nil, fmt.Errorf("documentation generation failed: %w", err)
	}

	// Convert to MCP response
	response := map[string]interface{}{
		"status":                  result.Status,
		"generated_documentation": result.GeneratedCode,
		"style":                   style,
		"iterations":              result.Iterations,
		"message":                 result.Message,
		"session_id":              sessionID,
		"timestamp":               time.Now().UTC().Format(time.RFC3339),
	}

	return mcp.NewToolResultText(fmt.Sprintf("Documentation generation result: %v", response)), nil
}

func (s *Server) handleCompleteCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	defer s.recordToolUsage("complete_code", startTime)

	// Extract parameters
	partialCode, err := request.RequireString("partial_code")
	if err != nil {
		s.recordError()
		return nil, fmt.Errorf("missing or invalid 'partial_code' parameter: %w", err)
	}

	// Get cursor position from arguments
	args := request.GetArguments()
	cursorPosRaw, exists := args["cursor_position"]
	if !exists {
		s.recordError()
		return nil, fmt.Errorf("missing 'cursor_position' parameter")
	}

	var cursorPos int
	switch v := cursorPosRaw.(type) {
	case float64:
		cursorPos = int(v)
	case int:
		cursorPos = v
	default:
		s.recordError()
		return nil, fmt.Errorf("invalid 'cursor_position' parameter type: %T", cursorPosRaw)
	}

	context := request.GetString("context", "")
	sessionID := request.GetString("session_id", fmt.Sprintf("complete_%d", time.Now().UnixNano()))

	// Create request
	completionRequest := tools.CompletionRequest{
		PartialCode: partialCode,
		CursorPos:   cursorPos,
		Context:     context,
		SessionID:   sessionID,
	}

	// Complete code
	result, err := s.advancedGenerator.CompleteCode(ctx, completionRequest)
	if err != nil {
		s.recordError()
		return nil, fmt.Errorf("code completion failed: %w", err)
	}

	// Convert to MCP response
	response := map[string]interface{}{
		"status":     result.Status,
		"completion": result.GeneratedCode,
		"message":    result.Message,
		"session_id": sessionID,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	if result.ValidationResult != nil {
		response["validation_result"] = result.ValidationResult
	}

	return mcp.NewToolResultText(fmt.Sprintf("Code completion result: %v", response)), nil
}

func (s *Server) handleBatchGenerate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()
	defer s.recordToolUsage("batch_generate", startTime)

	// Extract requests array
	args := request.GetArguments()
	requestsRaw, exists := args["requests"]
	if !exists {
		s.recordError()
		return nil, fmt.Errorf("missing 'requests' parameter")
	}

	requestsArray, ok := requestsRaw.([]interface{})
	if !ok {
		s.recordError()
		return nil, fmt.Errorf("'requests' parameter must be an array")
	}

	// Convert to generation requests
	var requests []tools.GenerationRequest
	for i, reqRaw := range requestsArray {
		if reqMap, ok := reqRaw.(map[string]interface{}); ok {
			genType, _ := reqMap["type"].(string)
			spec, _ := reqMap["specification"].(string)
			context, _ := reqMap["context"].(string)
			projectPath, _ := reqMap["project_path"].(string)

			if genType == "" || spec == "" {
				s.recordError()
				return nil, fmt.Errorf("invalid request at index %d: missing type or specification", i)
			}

			requests = append(requests, tools.GenerationRequest{
				Type:          genType,
				Specification: spec,
				Context:       context,
				ProjectPath:   projectPath,
			})
		} else {
			s.recordError()
			return nil, fmt.Errorf("invalid request format at index %d", i)
		}
	}

	sessionID := request.GetString("session_id", fmt.Sprintf("batch_%d", time.Now().UnixNano()))

	// Create batch request
	batchRequest := tools.BatchGenerationRequest{
		Requests:  requests,
		SessionID: sessionID,
	}

	// Perform batch generation
	results, err := s.advancedGenerator.BatchGenerate(ctx, batchRequest)
	if err != nil {
		s.recordError()
		// Note: err contains info about partial failures, results may still have successful items
	}

	// Convert results to response format
	var responseResults []map[string]interface{}
	for _, result := range results {
		resMap := map[string]interface{}{
			"status":         result.Status,
			"generated_code": result.GeneratedCode,
			"iterations":     result.Iterations,
			"message":        result.Message,
		}
		if result.ValidationResult != nil {
			resMap["validation_result"] = result.ValidationResult
		}
		responseResults = append(responseResults, resMap)
	}

	// Build response
	response := map[string]interface{}{
		"results":    responseResults,
		"total":      len(requests),
		"session_id": sessionID,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	if err != nil {
		response["errors"] = err.Error()
	}

	return mcp.NewToolResultText(fmt.Sprintf("Batch generation result: %v", response)), nil
}
