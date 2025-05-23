# MCP Server Implementation Blueprint

## Project Overview

Implement a Model Context Protocol (MCP) server as a PVM subcommand (`pvm mcp-server`) that provides LLMs with:
- Perl code analysis using PVM's type system
- Semantic code search via embeddings
- Intelligent code generation with collaborative sampling
- Rich context awareness and project-scoped operations

## High-Level Architecture

The implementation spans four main components:
1. **MCP Server Foundation** - Basic server setup, configuration, and tool registration
2. **Code Analysis Engine** - Leverage PVM's existing type checker and parser
3. **Embedding System** - Pluggable embedding providers with intelligent caching
4. **Generation Engine** - Collaborative code generation using MCP sampling

## Detailed Step-by-Step Blueprint

### Phase 1: Foundation (Steps 1-3)
**Goal**: Establish basic MCP server infrastructure with PVM integration

### Phase 2: Analysis Engine (Steps 4-6)
**Goal**: Implement code analysis tools leveraging existing PVM capabilities

### Phase 3: Embedding System (Steps 7-10)
**Goal**: Build semantic search with pluggable embedding providers

### Phase 4: Generation Engine (Steps 11-13)
**Goal**: Implement collaborative code generation with sampling

### Phase 5: Integration & Testing (Steps 14-15)
**Goal**: Wire everything together with comprehensive testing

---

## Implementation Steps

### Step 1: MCP Server Command Foundation ✅ COMPLETED

```text
Create the basic MCP server subcommand infrastructure for PVM.

Requirements:
- Add `mcp-server` subcommand to PVM CLI using cobra
- Implement basic MCP protocol server using the official MCP Go SDK
- Integrate with PVM's existing configuration system
- Support auto-discovery of Perl projects in current directory tree
- Handle graceful startup/shutdown with proper error reporting

Implementation Details:
- Create `internal/mcp/server.go` with MCP server setup
- Add `cmd/pvm/mcp.go` for the subcommand
- Use PVM's config system for MCP-specific settings
- Implement project discovery logic (look for cpanfile, .perl-version files)
- Return PVM-style errors for consistency

Testing:
- Unit tests for project discovery
- Integration tests for basic server startup/shutdown
- Configuration loading tests
- Error handling verification

Acceptance Criteria:
- ✅ `pvm mcp-server` starts successfully
- ✅ Server responds to MCP capability requests
- ✅ Projects are auto-discovered correctly
- ✅ Configuration loads from PVM config files
- ✅ Graceful error handling for missing dependencies

COMPLETED: All requirements implemented and tested successfully. MCP server foundation is working.
```

### Step 2: Basic Tool Registration and Request Handling ✅ COMPLETED

```text
Implement the MCP tool registration system and basic request routing.

Requirements:
- Register three main tool groups: analyze_code, search_code, generate_code
- Implement request validation and parameter parsing
- Set up project context management (separate contexts per project)
- Implement basic error responses using PVM error format
- Add request logging and basic metrics

Implementation Details:
- Create `internal/mcp/tools/` package structure
- Implement tool schema definitions for MCP protocol
- Add request router that validates and dispatches to appropriate handlers
- Create project context manager for scoped operations
- Integrate PVM's error system for consistent error responses

Testing:
- Tool registration verification
- Parameter validation tests
- Project context isolation tests
- Error response format validation
- Request routing correctness

Acceptance Criteria:
- ✅ All three tool groups are properly registered with MCP
- ✅ Invalid requests return proper error responses
- ✅ Project contexts are correctly isolated
- ✅ Tool schemas are valid according to MCP specification
- ✅ Logging captures all request/response cycles

COMPLETED: Enhanced tool handlers with proper validation, logging, and metrics tracking. All tools now include comprehensive error handling and performance monitoring.
```

### Step 3: Configuration and Validation Framework ✅ COMPLETED

```text
Implement comprehensive configuration system and input validation using PVM's type checker.

Requirements:
- Extend PVM config to support MCP server settings
- Implement code validation using existing PVM type checker
- Add auto-fix functionality with configurable behavior
- Create sampling infrastructure for LLM collaboration
- Set up validation caching for performance

Implementation Details:
- Add MCP section to PVM configuration schema
- Create validation service that wraps PVM's type checker
- Implement sampling client for MCP protocol
- Add validation result caching with smart invalidation
- Configure auto-fix behavior (default enabled, user configurable)

Testing:
- Configuration loading and validation
- Type checker integration tests
- Sampling request/response cycle tests
- Cache hit/miss behavior verification
- Auto-fix functionality testing

Acceptance Criteria:
- ✅ MCP settings integrate seamlessly with PVM config
- ✅ All Perl code inputs are validated before processing
- ✅ Auto-fix attempts work via sampling when enabled
- ✅ Validation results are cached for repeated requests
- ✅ Configuration changes are respected without restart

COMPLETED: All requirements implemented and tested successfully. Validation framework is working with caching and auto-fix support.
```

### Step 4: Code Analysis Tool Implementation

```text
Implement the analyze_code tool with full type analysis capabilities.

Requirements:
- Support three analysis types: get_types, check_errors, infer_types
- Return structured JSON responses to minimize token usage
- Leverage PVM's existing parser for code block extraction
- Implement auto-fix workflow with sampling
- Handle edge cases and malformed code gracefully

Implementation Details:
- Create `internal/mcp/tools/analyze.go` with analysis logic
- Integrate with PVM's type system for type extraction
- Implement error checking and type inference workflows
- Add auto-fix logic that samples LLM for corrections
- Return minimal, structured data optimized for LLM consumption

Testing:
- Type extraction accuracy tests
- Error detection and reporting tests
- Type inference correctness verification
- Auto-fix workflow integration tests
- Edge case handling (malformed code, missing types)

Acceptance Criteria:
- Accurately extracts type information from Perl code
- Detects and reports type errors with PVM consistency
- Infers types correctly using PVM's inference engine
- Auto-fix workflow successfully corrects common errors
- Responses are minimal and LLM-friendly
```

### Step 5: Project-Scoped Analysis and Context Management

```text
Enhance analysis tools with project-aware context and cross-file type resolution.

Requirements:
- Implement project-scoped type lookups and imports
- Add support for analyzing code in context of entire project
- Handle module dependencies and type imports correctly
- Optimize analysis performance with smart caching
- Provide project-level type summaries and statistics

Implementation Details:
- Create project context builder that maps all types and modules
- Implement cross-file type resolution using existing PVM logic
- Add project-level caching for type definitions and imports
- Create type summary generation for project overview
- Optimize for repeated analyses within same project

Testing:
- Cross-file type resolution accuracy
- Project context building correctness
- Cache performance and invalidation
- Type summary completeness
- Large project handling performance

Acceptance Criteria:
- Code analysis works correctly with imported types
- Project context provides accurate cross-file type information
- Performance is acceptable for projects with 100+ files
- Type summaries accurately represent project structure
- Context updates correctly when files change
```

### Step 6: Advanced Analysis Features

```text
Add sophisticated analysis capabilities including flow-sensitive analysis and type refinement.

Requirements:
- Integrate PVM's flow-sensitive type analysis
- Add support for conditional type refinement
- Implement type compatibility checking
- Add code quality analysis (complexity, type coverage)
- Support analysis of type annotations and their correctness

Implementation Details:
- Integrate with PVM's flow-sensitive analysis engine
- Add conditional type refinement detection and reporting
- Implement type compatibility matrix generation
- Create code quality metrics based on type information
- Add type annotation validation and suggestions

Testing:
- Flow-sensitive analysis accuracy tests
- Type refinement detection verification
- Compatibility checking correctness
- Code quality metrics validation
- Type annotation analysis tests

Acceptance Criteria:
- Flow-sensitive analysis provides accurate results
- Type refinement is correctly detected and reported
- Compatibility checking matches PVM's behavior
- Code quality metrics are meaningful and actionable
- Type annotation analysis helps improve code quality
```

### Step 7: Embedding System with chromem-go Integration

```text
Integrate chromem-go as the vector database and implement embedding providers.

Requirements:
- Integrate chromem-go for vector storage and similarity search
- Implement embedding provider wrapper for chromem-go
- Support OpenAI embeddings (primary provider)
- Support local embeddings with sentence-transformers
- Configure persistence and collection management

Implementation Details:
- Add chromem-go dependency to go.mod
- Create `internal/mcp/embeddings/` package with chromem integration
- Implement EmbeddingStore that wraps chromem.DB
- Create OpenAI embedding function for chromem-go
- Add local embedding option using go-sentence-transformers
- Configure persistent storage in XDG data directory
- Implement collection-per-project organization

Testing:
- Vector storage and retrieval tests
- Similarity search accuracy tests
- Persistence and recovery tests
- Multi-project isolation tests
- Embedding provider switching tests

Acceptance Criteria:
- Code embeddings are stored and retrieved correctly
- Similarity search returns relevant results
- Database persists between server restarts
- Projects are isolated in separate collections
- Graceful fallback when embedding provider is unavailable
```

### Step 8: Code Block Extraction and Document Preparation

```text
Implement code block extraction and prepare documents for chromem-go storage.

Requirements:
- Use PVM's Perl parser to identify meaningful code blocks
- Extract rich metadata for each code block
- Create chromem.Document structures with proper metadata
- Include file context and type information
- Batch processing for efficient embedding generation

Implementation Details:
- Create `internal/mcp/embeddings/extractor.go` for code extraction
- Parse Perl files into AST using PVM's parser
- Extract functions, methods, classes, and significant blocks
- Build chromem.Document with:
  - ID: project/file/block identifier
  - Content: the code block text
  - Metadata: type info, context, imports, location
- Implement batch document creation for efficiency
- Add incremental extraction for modified files only

Testing:
- Code block extraction coverage tests
- Document metadata completeness tests
- Batch processing performance tests
- Incremental update correctness tests
- Large file handling tests

Acceptance Criteria:
- All significant code blocks are extracted
- Documents contain complete metadata
- Batch processing handles 100+ files efficiently
- Incremental updates work correctly
- Type information is preserved in metadata
```

### Step 9: Collection Management and Optimization

```text
Implement chromem-go collection management and search optimization.

Requirements:
- Create per-project collections in chromem-go
- Implement efficient document updates and deletions
- Add metadata-based filtering for searches
- Optimize embedding generation with batching
- Monitor collection statistics and performance

Implementation Details:
- Create `internal/mcp/embeddings/manager.go` for collection ops
- Use chromem.DB.CreateCollection for project isolation
- Implement document lifecycle management:
  - Add new documents when files are created
  - Update documents when code changes
  - Remove documents when files are deleted
- Add metadata filters for search refinement:
  - Filter by type (function, class, method)
  - Filter by file path patterns
  - Filter by type annotations
- Batch embedding requests to reduce API calls
- Track collection metrics (size, document count, etc.)

Testing:
- Multi-collection isolation tests
- Document lifecycle management tests
- Metadata filtering accuracy tests
- Batch processing efficiency tests
- Performance benchmarks with large collections

Acceptance Criteria:
- Projects are fully isolated in separate collections
- Document updates are handled efficiently
- Searches can be filtered by metadata
- Embedding generation is optimized with batching
- Collection metrics are accurately tracked
```

### Step 10: Search Tool Implementation with chromem-go

```text
Implement the search_code tool using chromem-go's search capabilities.

Requirements:
- Support three search methods: similarity, type_signature, pattern
- Use chromem-go's Query method for similarity search
- Combine vector search with metadata filtering
- Provide ranked results with relevance scores
- Support cross-project and project-scoped searches

Implementation Details:
- Update `internal/mcp/tools/search.go` to use chromem-go
- Implement similarity search:
  - Use chromem.Collection.Query for vector similarity
  - Apply metadata filters for refinement
  - Return top-k results with scores
- Implement type signature search:
  - Extract type signatures from query
  - Use metadata filtering on type annotations
  - Combine with text search for accuracy
- Implement pattern search:
  - Use regex on document content
  - Leverage AST metadata for structured search
  - Combine with similarity for better results
- Add result post-processing:
  - Merge results from multiple search methods
  - Re-rank based on combined scores
  - Format results for LLM consumption

Testing:
- Search accuracy tests for each method
- Performance tests with 1000+ documents
- Relevance scoring validation
- Cross-project search isolation
- Query edge case handling

Acceptance Criteria:
- Similarity search leverages chromem-go effectively
- Type searches accurately match signatures
- Pattern searches work with complex queries
- Results are relevant and well-ranked
- Performance meets requirements at scale
```

### Step 11: Generation Memory System

```text
Implement small memory system for maintaining context during code generation tasks.

Requirements:
- Maintain context within single generation task
- Store type choices, naming conventions, and decisions
- Clear memory when generation task completes
- Provide context querying for generation decisions
- Keep memory size bounded and efficient

Implementation Details:
- Create generation session manager with scoped memory
- Implement context storage for types, names, and patterns
- Add memory querying interface for generation tools
- Create automatic cleanup on task completion
- Design memory structure for efficient access and updates

Testing:
- Memory persistence within generation sessions
- Proper cleanup between different generation tasks
- Context querying accuracy and performance
- Memory size bounds and enforcement
- Concurrent session isolation

Acceptance Criteria:
- Memory correctly maintains context during generation
- Memory is completely cleared between generation tasks
- Context queries provide relevant information efficiently
- Memory usage stays within configured bounds
- Multiple concurrent generations are properly isolated
```

### Step 12: Collaborative Code Generation with Sampling

```text
Implement collaborative code generation using MCP sampling for LLM interaction.

Requirements:
- Support three generation types: function, class, test
- Use sampling to collaborate with LLM on design decisions
- Integrate with generation memory for context continuity
- Validate generated code using PVM's type checker
- Provide iterative refinement through sampling

Implementation Details:
- Create `internal/mcp/tools/generate.go` with generation logic
- Implement sampling workflows for each generation type
- Add template systems for common Perl patterns
- Integrate type validation and auto-correction
- Create iterative refinement loops with LLM feedback

Testing:
- Generation quality across different code types
- Sampling workflow correctness and reliability
- Memory integration during generation sessions
- Type validation and correction effectiveness
- Iterative refinement convergence

Acceptance Criteria:
- Generated code follows Perl best practices and PVM typing
- Sampling collaboration produces meaningful results
- Memory provides helpful context throughout generation
- Type validation catches errors early in generation
- Iterative refinement improves code quality
```

### Step 13: Advanced Generation Features

```text
Add sophisticated generation capabilities including test generation and refactoring.

Requirements:
- Generate comprehensive test suites from type signatures
- Support code refactoring with type preservation
- Add documentation generation from typed code
- Implement code completion and suggestion features
- Support batch generation operations

Implementation Details:
- Create test generation templates and logic
- Implement type-preserving refactoring algorithms
- Add documentation generators using type information
- Create completion engine for partial code
- Add batch processing for multiple generation requests

Testing:
- Test generation coverage and quality
- Refactoring correctness and type preservation
- Documentation accuracy and completeness
- Completion relevance and accuracy
- Batch operation performance and reliability

Acceptance Criteria:
- Generated tests provide good coverage of type constraints
- Refactoring maintains type correctness throughout
- Generated documentation is accurate and helpful
- Code completion suggestions are contextually relevant
- Batch operations handle large requests efficiently
```

### Step 14: Integration and Performance Optimization

```text
Integrate all components and optimize performance for production use.

Requirements:
- Integrate all tools into cohesive MCP server
- Optimize performance for concurrent requests
- Add comprehensive error handling and recovery
- Implement health checks and monitoring
- Add graceful degradation for component failures

Implementation Details:
- Create integrated server with all tools properly wired
- Add connection pooling and request queuing
- Implement circuit breakers for external dependencies
- Add health check endpoints and monitoring metrics
- Create fallback modes for component failures

Testing:
- End-to-end integration tests covering all workflows
- Performance tests under load with concurrent requests
- Failure recovery and graceful degradation tests
- Health check accuracy and monitoring coverage
- Resource usage optimization verification

Acceptance Criteria:
- All tools work together seamlessly in integrated server
- Performance scales to handle 10+ concurrent LLM sessions
- System recovers gracefully from component failures
- Health checks accurately reflect system status
- Resource usage is optimized for production deployment
```

### Step 15: Comprehensive Testing and Documentation

```text
Add comprehensive test suite and complete documentation for the MCP server.

Requirements:
- Create integration tests for all major workflows
- Add performance benchmarks and regression tests
- Write comprehensive user documentation
- Create LLM integration examples and best practices
- Add troubleshooting guides and debugging tools

Implementation Details:
- Create end-to-end test suite covering all use cases
- Implement performance benchmarks for key operations
- Write user guide with configuration and usage examples
- Create LLM prompt examples and integration patterns
- Add debugging tools and troubleshooting documentation

Testing:
- Full system integration tests with real projects
- Performance regression test suite
- Documentation accuracy and completeness verification
- Example code correctness and functionality
- Troubleshooting guide effectiveness

Acceptance Criteria:
- Test suite provides comprehensive coverage of all functionality
- Performance benchmarks establish baseline expectations
- Documentation enables users to successfully deploy and use the system
- Examples demonstrate best practices for LLM integration
- Troubleshooting guides help users resolve common issues
```

## Implementation Notes

### Dependencies
- MCP Go SDK for protocol implementation
- Existing PVM components (type checker, parser, config system)
- Embedding provider APIs (OpenAI, VoyageAI)
- HuggingFace transformers for local embeddings

### File Structure
```
internal/mcp/
├── server.go              # Main MCP server implementation
├── config.go              # MCP-specific configuration
├── project.go             # Project discovery and context management
├── tools/
│   ├── analyze.go          # Code analysis tool
│   ├── search.go           # Code search tool
│   └── generate.go         # Code generation tool
├── embeddings/
│   ├── provider.go         # Embedding provider interface
│   ├── openai.go           # OpenAI embeddings
│   ├── voyage.go           # VoyageAI embeddings
│   ├── huggingface.go      # Local HF embeddings
│   └── cache.go            # Embedding cache system
├── generation/
│   ├── memory.go           # Generation memory system
│   ├── sampling.go         # MCP sampling client
│   └── templates.go        # Code generation templates
└── validation/
    ├── validator.go        # Code validation service
    └── auto_fix.go         # Auto-fix functionality

cmd/pvm/
└── mcp.go                  # MCP server subcommand
```

### Configuration Schema Extension
```toml
[mcp_server]
# Server settings
port = 3000
host = "localhost"
auto_discover_projects = true

# Analysis settings
auto_fix_errors = true
validation_cache_size = "50MB"

# Embedding settings
embedding_provider = "openai"  # openai | voyageai | huggingface
embedding_cache_size = "100MB"
embedding_model = "text-embedding-3-small"  # provider-specific

# Generation settings
generation_memory_size = 50
enable_iterative_refinement = true

# Performance settings
max_concurrent_requests = 10
request_timeout = "30s"
```

This blueprint provides a comprehensive, step-by-step approach to implementing a sophisticated MCP server that leverages PVM's existing capabilities while adding powerful LLM integration features. Each step builds incrementally on the previous ones, ensuring safe and testable development progression.
