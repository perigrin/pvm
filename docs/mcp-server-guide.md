# PVM MCP Server User Guide

## Overview

The PVM MCP (Model Context Protocol) Server provides LLMs with advanced Perl code analysis, semantic search, and intelligent code generation capabilities. It integrates PVM's type system with embedding-based search and collaborative code generation.

## Features

- **Code Analysis**: Type checking, error detection, and type inference using PVM's advanced type system
- **Semantic Search**: Vector-based code search with type signature and pattern matching
- **Code Generation**: Collaborative code generation with LLM sampling and iterative refinement
- **Project-Scoped Operations**: Automatic project discovery and context-aware analysis
- **Production-Ready**: Health monitoring, circuit breakers, and performance optimization

## Quick Start

### 1. Prerequisites

- Go 1.21 or later
- PVM installed and configured
- Node.js and npm (for tree-sitter integration)

### 2. Build the MCP Server

```bash
# Build PVM with MCP server support
make build

# Or build just the MCP components
make pvm
```

### 3. Start the MCP Server

```bash
# Start the MCP server in the current directory
pvm mcp-server

# Start with custom configuration
pvm mcp-server --config /path/to/config.toml

# Start without auto-discovery
pvm mcp-server --no-auto-discover
```

### 4. Connect from an LLM Client

The MCP server uses stdio for communication. Most LLM clients that support MCP can connect by running:

```bash
pvm mcp-server
```

## Configuration

### Basic Configuration

Create a `pvm.toml` file in your project or home directory:

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
embedding_provider = "openai"  # openai | local
embedding_cache_size = "100MB"
embedding_model = "text-embedding-3-small"

# Generation settings
generation_memory_size = 50
enable_iterative_refinement = true

# Performance settings
max_concurrent_requests = 10
request_timeout = "30s"
```

### Embedding Providers

#### OpenAI Embeddings
```toml
[mcp_server]
embedding_provider = "openai"
embedding_model = "text-embedding-3-small"
```

Set the `OPENAI_API_KEY` environment variable.

#### Local Embeddings
```toml
[mcp_server]
embedding_provider = "local"
```

Uses deterministic local embeddings for development and testing.

### Performance Tuning

```toml
[mcp_server]
# Concurrent request limits
max_concurrent_requests = 20

# Request timeouts
request_timeout = "60s"

# Cache sizes
validation_cache_size = "100MB"
embedding_cache_size = "200MB"

# Memory settings
generation_memory_size = 100
```

## Available Tools

### Code Analysis Tools

#### `analyze_code`
Analyzes Perl code for types, errors, and inference.

**Parameters:**
- `code` (string): Perl code to analyze
- `analysis_type` (enum): Type of analysis
  - `get_types`: Extract type information
  - `check_errors`: Find type errors
  - `infer_types`: Infer missing types
  - `project_analysis`: Analyze entire project
  - `project_summary`: Get project summary
- `project_path` (optional): Project path for context

**Example:**
```json
{
  "tool": "analyze_code",
  "arguments": {
    "code": "my Int $x = 42; my Str $y = $x;",
    "analysis_type": "check_errors"
  }
}
```

### Search Tools

#### `search_code`
Search for code using semantic similarity, type signatures, or patterns.

**Parameters:**
- `query` (string): Search query
- `search_method` (enum): Search method
  - `similarity`: Vector similarity search
  - `type_signature`: Type-based search
  - `pattern`: Regex pattern search
- `project_path` (optional): Project scope

**Example:**
```json
{
  "tool": "search_code",
  "arguments": {
    "query": "function that calculates factorial",
    "search_method": "similarity",
    "project_path": "/path/to/project"
  }
}
```

### Generation Tools

#### `generate_code`
Generate Perl code using collaborative sampling.

**Parameters:**
- `generation_type` (enum): Type of code to generate
  - `function`: Generate a function
  - `class`: Generate a class
  - `test`: Generate test code
- `specification` (string): Description of what to generate
- `context` (optional): Additional context code
- `project_path` (optional): Project context
- `session_id` (optional): Memory session ID

**Example:**
```json
{
  "tool": "generate_code",
  "arguments": {
    "generation_type": "function",
    "specification": "Calculate factorial of a number with type checking",
    "context": "use v5.40; use experimental 'class';"
  }
}
```

### Advanced Generation Tools

#### `generate_tests_from_types`
Generate comprehensive test suites from type signatures.

#### `refactor_code`
Perform type-preserving code refactoring.

#### `generate_documentation`
Generate documentation from typed code.

#### `complete_code`
Provide intelligent code completion suggestions.

#### `batch_generate`
Handle multiple generation requests efficiently.

### Monitoring Tools

#### `health_check`
Get server health status and component status.

**Example:**
```json
{
  "tool": "health_check",
  "arguments": {
    "component": "embedding_store"
  }
}
```

#### `get_metrics`
Get server performance metrics and statistics.

**Example:**
```json
{
  "tool": "get_metrics",
  "arguments": {
    "category": "performance"
  }
}
```

#### `circuit_breaker_control`
Control circuit breaker state.

**Example:**
```json
{
  "tool": "circuit_breaker_control",
  "arguments": {
    "action": "status"
  }
}
```

### Memory Management

#### `memory_session`
Manage generation memory sessions for context continuity.

**Example:**
```json
{
  "tool": "memory_session",
  "arguments": {
    "action": "create",
    "session_id": "my_session"
  }
}
```

## Best Practices

### 1. Project Structure

Organize your Perl projects with clear structure:

```
project/
├── cpanfile                 # CPAN dependencies
├── .perl-version           # Perl version specification
├── lib/                    # Library modules
│   └── MyApp/
│       ├── Core.pm
│       └── Utils.pm
├── bin/                    # Executable scripts
├── t/                      # Tests
└── pvm.toml               # PVM configuration
```

### 2. Type Annotations

Use consistent type annotations for better analysis:

```perl
use v5.40;
use experimental 'class';

class Calculator {
    field Int $value = 0;

    method add(Int $n) : Int {
        $value += $n;
        return $value;
    }

    method multiply(Num $factor) : Num {
        return $value * $factor;
    }
}
```

### 3. Memory Session Management

Use memory sessions for complex generation workflows:

```json
// Start a session
{
  "tool": "memory_session",
  "arguments": {
    "action": "create",
    "session_id": "feature_development"
  }
}

// Generate related code with context
{
  "tool": "generate_code",
  "arguments": {
    "generation_type": "class",
    "specification": "User authentication class",
    "session_id": "feature_development"
  }
}

// Generate tests for the same feature
{
  "tool": "generate_code",
  "arguments": {
    "generation_type": "test",
    "specification": "Tests for user authentication",
    "session_id": "feature_development"
  }
}
```

### 4. Performance Optimization

- Use project-scoped operations when possible
- Enable caching for repeated operations
- Monitor performance metrics regularly
- Use appropriate timeout values

## Troubleshooting

### Common Issues

#### Server Won't Start

**Problem**: Server fails to start with configuration errors.

**Solution**: Check your configuration file for syntax errors and validate required fields:

```bash
# Check configuration syntax
pvm mcp-server --config /path/to/config.toml --validate

# Start with minimal configuration
pvm mcp-server --no-auto-discover
```

#### High Memory Usage

**Problem**: Server consumes too much memory.

**Solution**: Adjust cache sizes and limits:

```toml
[mcp_server]
validation_cache_size = "25MB"    # Reduce from default 50MB
embedding_cache_size = "50MB"     # Reduce from default 100MB
generation_memory_size = 25       # Reduce from default 50
max_concurrent_requests = 5       # Reduce from default 10
```

#### Slow Response Times

**Problem**: Analysis or generation takes too long.

**Solution**: Check performance metrics and optimize:

```json
{
  "tool": "get_metrics",
  "arguments": {
    "category": "performance"
  }
}
```

Look for:
- High queue length
- Circuit breakers in open state
- Resource limits being hit

#### Embedding Failures

**Problem**: Semantic search not working.

**Solution**: Check embedding provider configuration:

```bash
# For OpenAI embeddings
export OPENAI_API_KEY="your-api-key"

# Or switch to local embeddings
```

```toml
[mcp_server]
embedding_provider = "local"
```

#### Type Analysis Errors

**Problem**: Code analysis produces incorrect results.

**Solution**: Ensure type annotations are correct and check project context:

```json
{
  "tool": "analyze_code",
  "arguments": {
    "code": "your code here",
    "analysis_type": "project_analysis",
    "project_path": "/full/path/to/project"
  }
}
```

### Debug Mode

Enable debug logging by setting the log level:

```toml
[logging]
level = "debug"
```

### Health Monitoring

Regularly check server health:

```json
{
  "tool": "health_check",
  "arguments": {}
}
```

Look for:
- Component status (all should be "healthy")
- Circuit breaker states
- Resource usage

## Advanced Usage

### Custom Embedding Providers

You can extend the server to support additional embedding providers by implementing the embedding provider interface.

### Integration with CI/CD

Use the MCP server in automated workflows:

```bash
# Type check all files in a project
pvm mcp-server --batch-analyze /path/to/project

# Generate tests for new code
pvm mcp-server --generate-tests /path/to/new/code
```

### Performance Monitoring

Set up monitoring dashboards using the metrics endpoint:

```bash
# Get all metrics
curl -X POST localhost:3000/metrics

# Get specific category
curl -X POST localhost:3000/metrics -d '{"category": "circuit_breakers"}'
```

## Support

For support and bug reports:

- Check the troubleshooting section above
- Review the server logs for error messages
- Verify your configuration matches the examples
- Test with minimal configuration to isolate issues

## API Reference

For complete API documentation, see the [MCP Server API Reference](mcp-server-api.md).
