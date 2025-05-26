# LLM Integration Examples and Best Practices

This document provides practical examples and best practices for integrating the PVM MCP Server with Large Language Models.

## Integration Patterns

### 1. Code Analysis Workflow

Use the MCP server to analyze code before making suggestions or modifications.

```json
// Example: Analyze code for type errors before suggesting fixes
{
  "tool": "analyze_code",
  "arguments": {
    "code": "my Int $x = 'not a number'; my Str $y = $x + 1;",
    "analysis_type": "check_errors"
  }
}
```

**Response Pattern:**
```json
{
  "status": "error",
  "errors": [
    {
      "type": "type_mismatch",
      "line": 1,
      "message": "Cannot assign Str to Int variable",
      "suggestion": "Use Int literal: my Int $x = 42;"
    }
  ]
}
```

**LLM Integration:**
```
Based on the analysis, I found type errors in your code:
1. Line 1: Cannot assign string 'not a number' to Int variable $x
2. Line 1: Type mismatch in addition operation

Here's the corrected version:
```perl
my Int $x = 42;
my Str $y = ($x + 1).Str;
```

### 2. Semantic Code Search

Help users find relevant code examples and patterns.

```json
// Example: Find similar code patterns
{
  "tool": "search_code",
  "arguments": {
    "query": "class with field validation and error handling",
    "search_method": "similarity",
    "project_path": "/current/project"
  }
}
```

**LLM Integration:**
```
I found these relevant examples in your codebase:

1. `UserValidator` class (lib/App/User.pm):
   - Implements field validation with custom error messages
   - Uses type constraints for input validation

2. `ConfigParser` class (lib/App/Config.pm):
   - Has comprehensive error handling patterns
   - Shows how to validate and transform input data

Would you like me to show you how to adapt these patterns for your use case?
```

### 3. Incremental Code Generation

Generate code iteratively with context preservation.

```json
// Step 1: Start a memory session
{
  "tool": "memory_session",
  "arguments": {
    "action": "create",
    "session_id": "user_auth_feature"
  }
}

// Step 2: Generate the main class
{
  "tool": "generate_code",
  "arguments": {
    "generation_type": "class",
    "specification": "User authentication class with login/logout methods",
    "session_id": "user_auth_feature"
  }
}

// Step 3: Generate tests for the class
{
  "tool": "generate_tests_from_types",
  "arguments": {
    "code": "/* generated class code */",
    "framework": "Test2::V0",
    "session_id": "user_auth_feature"
  }
}
```

### 4. Code Refactoring Assistant

Help users refactor code while preserving types and functionality.

```json
// Example: Extract method refactoring
{
  "tool": "refactor_code",
  "arguments": {
    "code": "/* original code */",
    "refactoring_type": "extract_method",
    "target": "validation logic",
    "new_name": "validate_user_input"
  }
}
```

## Best Practices for LLMs

### 1. Always Validate Before Generating

```
Before I generate code for you, let me analyze your existing code to understand the context and types:

[Call analyze_code tool]

Based on the analysis, I can see you're using:
- Typed field declarations with validation
- Error handling with custom exception classes
- A specific coding style with camelCase methods

I'll generate code that follows these patterns...
```

### 2. Use Project Context

```
Let me search your project for similar patterns to ensure consistency:

[Call search_code tool with project_path]

I found similar implementations in your codebase. I'll base the new code on these established patterns...
```

### 3. Provide Type-Safe Suggestions

```
Here's a type-safe implementation based on your project's type system:

[Call generate_code tool]

The generated code includes:
- Proper type annotations matching your existing code
- Error handling that follows your project patterns
- Validation that's consistent with your type constraints
```

### 4. Iterative Refinement

```
Let me generate an initial version and then refine it based on analysis:

[Call generate_code tool]
[Call analyze_code tool on generated code]
[If errors found, call generate_code again with fixes]

The final version passes type checking and follows your project conventions.
```

## Prompt Templates

### Code Analysis Prompt

```
I need to analyze this Perl code for type safety and potential issues:

```perl
[USER_CODE]
```

Please:
1. Check for type errors and inconsistencies
2. Identify potential runtime issues
3. Suggest improvements for type safety
4. Verify compatibility with the project's type system

[Use analyze_code tool with check_errors and infer_types]
```

### Code Generation Prompt

```
Generate a [TYPE] that [SPECIFICATION].

Requirements:
- Follow the existing project patterns and conventions
- Include proper type annotations
- Add comprehensive error handling
- Ensure type safety throughout

Context: [EXISTING_CODE_CONTEXT]

[Use memory_session to maintain context]
[Use generate_code tool with appropriate parameters]
[Use analyze_code to validate generated code]
```

### Refactoring Prompt

```
Help me refactor this code to improve [ASPECT]:

```perl
[ORIGINAL_CODE]
```

Goals:
- [SPECIFIC_GOALS]
- Maintain type safety
- Preserve existing functionality
- Follow project conventions

[Use analyze_code to understand current structure]
[Use refactor_code tool for transformations]
[Use analyze_code to verify refactored code]
```

## Error Handling Patterns

### Graceful Degradation

When MCP tools fail, provide helpful fallbacks:

```
I encountered an issue with code analysis, but I can still help based on the code structure I can see:

[Provide analysis based on visible patterns]

For more detailed analysis, you might want to check:
- Type annotation syntax
- Project configuration
- Available type definitions
```

### Tool-Specific Error Handling

```javascript
// Pseudo-code for LLM integration
async function analyzeCode(code) {
  try {
    const result = await mcpServer.call('analyze_code', {
      code: code,
      analysis_type: 'check_errors'
    });
    return result;
  } catch (error) {
    if (error.type === 'validation_timeout') {
      return "Analysis timed out - the code might be too complex. Try breaking it into smaller parts.";
    } else if (error.type === 'parser_error') {
      return "Syntax error detected. Please check your Perl syntax.";
    } else {
      return "Unable to analyze code at the moment. I can still help with general questions.";
    }
  }
}
```

## Performance Optimization

### Batch Operations

When possible, use batch operations for efficiency:

```json
// Instead of multiple individual requests
{
  "tool": "batch_generate",
  "arguments": {
    "requests": [
      {
        "type": "function",
        "specification": "Parse configuration file"
      },
      {
        "type": "function",
        "specification": "Validate configuration data"
      },
      {
        "type": "test",
        "specification": "Test configuration parsing"
      }
    ],
    "session_id": "config_feature"
  }
}
```

### Cache Management

Monitor and manage caching for better performance:

```json
// Check performance metrics
{
  "tool": "get_metrics",
  "arguments": {
    "category": "performance"
  }
}

// If cache hit rate is low, consider:
// - Using more specific search terms
// - Grouping related operations
// - Using memory sessions for context
```

### Circuit Breaker Awareness

Handle circuit breaker states gracefully:

```json
// Check circuit breaker status
{
  "tool": "circuit_breaker_control",
  "arguments": {
    "action": "status"
  }
}

// If circuits are open, inform user and suggest alternatives
```

## Integration Examples

### VS Code Extension Integration

```typescript
// Example VS Code extension using MCP server
import { MCPClient } from 'mcp-client';

class PVMCodeAnalyzer {
  private mcpClient: MCPClient;

  async analyzeDocument(document: vscode.TextDocument) {
    try {
      const result = await this.mcpClient.callTool('analyze_code', {
        code: document.getText(),
        analysis_type: 'check_errors',
        project_path: vscode.workspace.rootPath
      });

      return this.createDiagnostics(result);
    } catch (error) {
      vscode.window.showErrorMessage(`Analysis failed: ${error.message}`);
      return [];
    }
  }

  async generateCode(specification: string) {
    const session = `vscode_${Date.now()}`;

    try {
      // Create memory session
      await this.mcpClient.callTool('memory_session', {
        action: 'create',
        session_id: session
      });

      // Generate code
      const result = await this.mcpClient.callTool('generate_code', {
        generation_type: 'function',
        specification: specification,
        session_id: session
      });

      return result.generated_code;
    } finally {
      // Clean up session
      await this.mcpClient.callTool('memory_session', {
        action: 'clear',
        session_id: session
      });
    }
  }
}
```

### CLI Tool Integration

```bash
#!/bin/bash
# Example CLI tool using MCP server

# Start MCP server in background
pvm mcp-server &
MCP_PID=$!

# Function to analyze files
analyze_file() {
  local file=$1
  echo "Analyzing $file..."

  # Call MCP server via JSON-RPC
  curl -s -X POST localhost:3000 \
    -H "Content-Type: application/json" \
    -d "{
      \"tool\": \"analyze_code\",
      \"arguments\": {
        \"code\": \"$(cat "$file")\",
        \"analysis_type\": \"check_errors\"
      }
    }" | jq '.errors[]'
}

# Cleanup on exit
trap "kill $MCP_PID" EXIT

# Analyze all Perl files
find . -name "*.pl" -o -name "*.pm" | while read file; do
  analyze_file "$file"
done
```

### GitHub Actions Integration

```yaml
# .github/workflows/perl-analysis.yml
name: Perl Code Analysis

on: [push, pull_request]

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup PVM
        run: |
          # Install PVM and dependencies
          make install

      - name: Start MCP Server
        run: |
          pvm mcp-server &
          echo $! > mcp.pid

      - name: Analyze Code
        run: |
          # Run analysis on all Perl files
          find . -name "*.pl" -o -name "*.pm" | \
          xargs -I {} pvm analyze-file {}

      - name: Generate Coverage Report
        run: |
          pvm generate-coverage-report

      - name: Stop MCP Server
        run: |
          kill $(cat mcp.pid)
```

## Debugging and Monitoring

### Health Check Integration

```json
// Regular health monitoring
{
  "tool": "health_check",
  "arguments": {}
}

// Component-specific checks
{
  "tool": "health_check",
  "arguments": {
    "component": "embedding_store"
  }
}
```

### Performance Monitoring

```json
// Monitor performance metrics
{
  "tool": "get_metrics",
  "arguments": {
    "category": "performance"
  }
}

// Check for issues:
// - High error rates
// - Slow response times
// - Resource exhaustion
// - Circuit breaker trips
```

### Logging Integration

Configure appropriate logging levels:

```toml
[logging]
level = "info"  # debug, info, warn, error

[mcp_server]
# Enable performance logging
enable_metrics = true
log_tool_usage = true
```

## Advanced Patterns

### Multi-Project Analysis

```json
// Analyze code across multiple projects
{
  "tool": "analyze_code",
  "arguments": {
    "code": "use MyApp::Utils;",
    "analysis_type": "project_analysis",
    "project_path": "/workspace"  // Parent directory
  }
}
```

### Custom Type Definitions

```json
// Generate code with custom type constraints
{
  "tool": "generate_code",
  "arguments": {
    "generation_type": "class",
    "specification": "Email validator with custom EmailAddress type",
    "context": "type EmailAddress = Str where { /^[^@]+@[^@]+$/ }"
  }
}
```

### Collaborative Refactoring

```json
// Multi-step refactoring with LLM collaboration
{
  "tool": "memory_session",
  "arguments": {
    "action": "create",
    "session_id": "refactor_project"
  }
}

// Store refactoring decisions
{
  "tool": "memory_session",
  "arguments": {
    "action": "get",
    "session_id": "refactor_project",
    "decision_type": "naming_convention",
    "decision_choice": "camelCase_for_methods"
  }
}
```

This comprehensive guide should help LLM developers effectively integrate with the PVM MCP Server and provide better Perl development assistance to users.
