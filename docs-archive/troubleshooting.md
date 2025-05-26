# PVM MCP Server Troubleshooting Guide

This guide helps diagnose and resolve common issues with the PVM MCP Server.

## Quick Diagnostics

### Health Check Command

First, check if the server is operational:

```bash
# Basic health check
pvm mcp-server --health-check

# Detailed component status
pvm mcp-server --health-check --verbose
```

### Server Status

Check server status and metrics:

```json
{
  "tool": "health_check",
  "arguments": {}
}
```

Expected healthy response:
```json
{
  "status": "healthy",
  "components": {
    "embedding_store": {"status": "healthy"},
    "validator": {"status": "healthy"},
    "memory_manager": {"status": "healthy"},
    "sampling_client": {"status": "healthy"},
    "circuit_breakers": {"status": "healthy"}
  },
  "uptime": "00:05:23"
}
```

## Common Issues

### 1. Server Won't Start

#### Issue: Configuration Error
```
Error: MCP configuration is required
```

**Cause**: Missing or invalid configuration file.

**Solution**:
```bash
# Check if config file exists
ls -la pvm.toml ~/.pvm.toml

# Create minimal config
cat > pvm.toml << EOF
[mcp_server]
port = 3000
host = "localhost"
auto_discover_projects = true
embedding_provider = "local"
EOF

# Test with minimal config
pvm mcp-server --config pvm.toml
```

#### Issue: Port Already in Use
```
Error: listen tcp :3000: bind: address already in use
```

**Solution**:
```bash
# Find process using the port
lsof -i :3000

# Kill the process or use different port
pvm mcp-server --port 3001
```

#### Issue: Tree-sitter Build Failure
```
Error: failed to create code analyzer: tree-sitter not found
```

**Solution**:
```bash
# Build tree-sitter components
make tree-sitter

# Or install dependencies manually
cd tree-sitter-typed-perl
npm install
tree-sitter generate
```

### 2. Analysis Errors

#### Issue: Type Checking Fails
```
{
  "status": "error",
  "message": "type checker unavailable"
}
```

**Diagnostics**:
```json
{
  "tool": "health_check",
  "arguments": {
    "component": "validator"
  }
}
```

**Solutions**:
1. Check if tree-sitter parser is built correctly:
   ```bash
   make tree-sitter
   go test ./internal/parser/treesitter/...
   ```

2. Verify type definitions are available:
   ```bash
   ls -la internal/typedef/
   ```

3. Test with simple code first:
   ```json
   {
     "tool": "analyze_code",
     "arguments": {
       "code": "my Int $x = 42;",
       "analysis_type": "get_types"
     }
   }
   ```

#### Issue: Project Analysis Timeout
```
{
  "status": "timeout",
  "message": "project analysis timed out"
}
```

**Solutions**:
1. Increase timeout in configuration:
   ```toml
   [mcp_server]
   request_timeout = "120s"
   ```

2. Analyze smaller subsets:
   ```json
   {
     "tool": "analyze_code",
     "arguments": {
       "code": "specific_file_content",
       "analysis_type": "check_errors"
     }
   }
   ```

3. Check project size and complexity:
   ```bash
   find . -name "*.pl" -o -name "*.pm" | wc -l
   find . -name "*.pl" -o -name "*.pm" -exec wc -l {} + | tail -1
   ```

### 3. Search and Embedding Issues

#### Issue: Semantic Search Returns No Results
```
{
  "status": "success",
  "results": []
}
```

**Diagnostics**:
```json
{
  "tool": "get_metrics",
  "arguments": {
    "category": "circuit_breakers"
  }
}
```

**Solutions**:
1. Check embedding provider status:
   ```json
   {
     "tool": "health_check",
     "arguments": {
       "component": "embedding_store"
     }
   }
   ```

2. Verify project has been indexed:
   ```bash
   # Check if embedding cache has data
   ls -la ~/.local/share/pvm/embeddings/
   ```

3. Try different search methods:
   ```json
   {
     "tool": "search_code",
     "arguments": {
       "query": "your_search_term",
       "search_method": "pattern"
     }
   }
   ```

4. For OpenAI embeddings, check API key:
   ```bash
   echo $OPENAI_API_KEY
   # Should not be empty
   ```

#### Issue: Embedding API Failures
```
{
  "status": "error",
  "message": "embedding provider failed"
}
```

**Solutions**:
1. Check circuit breaker status:
   ```json
   {
     "tool": "circuit_breaker_control",
     "arguments": {
       "action": "status"
     }
   }
   ```

2. Reset circuit breaker if needed:
   ```json
   {
     "tool": "circuit_breaker_control",
     "arguments": {
       "action": "reset",
       "breaker_name": "embedding_provider"
     }
   }
   ```

3. Switch to local embeddings temporarily:
   ```toml
   [mcp_server]
   embedding_provider = "local"
   ```

### 4. Performance Issues

#### Issue: High Memory Usage
```
Memory usage: 2.5GB (exceeds 2GB limit)
```

**Diagnostics**:
```json
{
  "tool": "get_metrics",
  "arguments": {
    "category": "performance"
  }
}
```

**Solutions**:
1. Reduce cache sizes:
   ```toml
   [mcp_server]
   validation_cache_size = "25MB"
   embedding_cache_size = "50MB"
   generation_memory_size = 25
   ```

2. Clear caches:
   ```bash
   rm -rf ~/.cache/pvm/validation/
   rm -rf ~/.local/share/pvm/embeddings/
   ```

3. Restart server periodically:
   ```bash
   # Add to cron for regular restarts
   0 2 * * * pkill -f "pvm mcp-server" && sleep 5 && pvm mcp-server &
   ```

#### Issue: Slow Response Times
```
Average response time: 15 seconds (expected < 5s)
```

**Diagnostics**:
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
- Low cache hit rates
- Resource exhaustion

**Solutions**:
1. Increase concurrent request limit:
   ```toml
   [mcp_server]
   max_concurrent_requests = 20
   ```

2. Optimize cache usage:
   ```toml
   [mcp_server]
   validation_cache_size = "100MB"  # Increase cache
   ```

3. Profile performance:
   ```bash
   go test -bench=. -benchmem ./internal/mcp/...
   ```

### 5. Generation Issues

#### Issue: Code Generation Produces Invalid Code
```
{
  "status": "success",
  "generated_code": "invalid syntax",
  "validation_result": {
    "valid": false,
    "errors": ["syntax error"]
  }
}
```

**Solutions**:
1. Enable auto-fix:
   ```toml
   [mcp_server]
   auto_fix_errors = true
   enable_iterative_refinement = true
   ```

2. Provide better context:
   ```json
   {
     "tool": "generate_code",
     "arguments": {
       "generation_type": "function",
       "specification": "detailed specification here",
       "context": "existing code context",
       "project_path": "/full/project/path"
     }
   }
   ```

3. Use memory sessions for complex generation:
   ```json
   {
     "tool": "memory_session",
     "arguments": {
       "action": "create",
       "session_id": "generation_session"
     }
   }
   ```

#### Issue: Generation Memory Errors
```
{
  "status": "error",
  "message": "generation memory full"
}
```

**Solutions**:
1. Increase memory size:
   ```toml
   [mcp_server]
   generation_memory_size = 100
   ```

2. Clear old sessions:
   ```json
   {
     "tool": "memory_session",
     "arguments": {
       "action": "clear",
       "session_id": "old_session"
     }
   }
   ```

3. Monitor memory usage:
   ```json
   {
     "tool": "memory_session",
     "arguments": {
       "action": "stats",
       "session_id": "current_session"
     }
   }
   ```

## Advanced Debugging

### Enable Debug Logging

```toml
[logging]
level = "debug"
output = "/tmp/pvm-mcp-debug.log"

[mcp_server]
log_tool_usage = true
enable_metrics = true
```

### Performance Profiling

```bash
# Run with profiling
go build -o pvm-debug ./cmd/pvm
./pvm-debug mcp-server --cpuprofile=cpu.prof --memprofile=mem.prof

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

### Component Testing

Test individual components:

```bash
# Test validation cache
go test ./internal/mcp/validation -run TestValidationCache -v

# Test circuit breakers
go test ./internal/mcp -run TestStep14_CircuitBreaker -v

# Test embedding system
go test ./internal/mcp/embeddings -run TestEmbeddingStore -v

# Test performance manager
go test ./internal/mcp -run TestStep14_PerformanceManager -v
```

### Database Debugging

For embedding store issues:

```bash
# Check chromem database
ls -la ~/.local/share/pvm/embeddings/collections/

# Test database connectivity
go run -c "
import 'github.com/philippgille/chromem-go'
db := chromem.NewDB()
collections := db.ListCollections()
fmt.Println('Collections:', len(collections))
"
```

## Monitoring and Alerts

### Health Monitoring Script

```bash
#!/bin/bash
# health-monitor.sh

check_health() {
  local response=$(curl -s -X POST localhost:3000 -d '{"tool":"health_check","arguments":{}}')
  local status=$(echo "$response" | jq -r '.status')

  if [ "$status" != "healthy" ]; then
    echo "ALERT: MCP Server unhealthy - $response"
    # Send alert notification
    # restart service if needed
  fi
}

# Run every minute
while true; do
  check_health
  sleep 60
done
```

### Performance Monitoring

```bash
#!/bin/bash
# perf-monitor.sh

monitor_performance() {
  local metrics=$(curl -s -X POST localhost:3000 -d '{"tool":"get_metrics","arguments":{"category":"performance"}}')
  local avg_response=$(echo "$metrics" | jq -r '.performance_metrics.metrics.avg_response_time_ms')

  if [ "$avg_response" -gt 5000 ]; then
    echo "ALERT: High response time - ${avg_response}ms"
  fi
}
```

### Resource Monitoring

```bash
#!/bin/bash
# resource-monitor.sh

monitor_resources() {
  local pid=$(pgrep -f "pvm mcp-server")
  if [ -n "$pid" ]; then
    local memory=$(ps -p "$pid" -o rss= | awk '{print $1/1024}')
    local cpu=$(ps -p "$pid" -o %cpu= | awk '{print $1}')

    echo "Memory: ${memory}MB, CPU: ${cpu}%"

    if [ "${memory%.*}" -gt 1000 ]; then
      echo "ALERT: High memory usage - ${memory}MB"
    fi
  fi
}
```

## Recovery Procedures

### Automatic Recovery

```bash
#!/bin/bash
# auto-recover.sh

recover_service() {
  echo "Attempting recovery..."

  # Stop service
  pkill -f "pvm mcp-server"
  sleep 5

  # Clear caches
  rm -rf ~/.cache/pvm/validation/*
  rm -rf ~/.cache/pvm/embeddings/*

  # Restart service
  pvm mcp-server &
  sleep 10

  # Verify recovery
  local status=$(curl -s -X POST localhost:3000 -d '{"tool":"health_check","arguments":{}}' | jq -r '.status')
  if [ "$status" = "healthy" ]; then
    echo "Recovery successful"
  else
    echo "Recovery failed"
    exit 1
  fi
}
```

### Manual Recovery Steps

1. **Stop the server gracefully**:
   ```bash
   # Send SIGTERM for graceful shutdown
   pkill -TERM -f "pvm mcp-server"
   sleep 10

   # Force kill if needed
   pkill -KILL -f "pvm mcp-server"
   ```

2. **Clear problematic state**:
   ```bash
   # Clear caches
   rm -rf ~/.cache/pvm/
   rm -rf ~/.local/share/pvm/embeddings/

   # Reset configuration to defaults
   cp pvm.toml pvm.toml.backup
   cat > pvm.toml << EOF
   [mcp_server]
   port = 3000
   embedding_provider = "local"
   validation_cache_size = "25MB"
   generation_memory_size = 25
   EOF
   ```

3. **Restart with minimal configuration**:
   ```bash
   pvm mcp-server --config pvm.toml
   ```

4. **Gradually restore functionality**:
   ```bash
   # Test basic functionality
   curl -X POST localhost:3000 -d '{"tool":"health_check","arguments":{}}'

   # Enable features incrementally
   # Update configuration step by step
   ```

## Getting Help

### Log Collection

When reporting issues, collect these logs:

```bash
# System logs
journalctl -u pvm-mcp-server --since "1 hour ago"

# Application logs
tail -100 /tmp/pvm-mcp-debug.log

# Configuration
cat pvm.toml

# System information
uname -a
go version
df -h
free -h
```

### Diagnostic Information

```bash
# Generate diagnostic report
cat > diagnostic-report.txt << EOF
# PVM MCP Server Diagnostic Report
Date: $(date)
Version: $(pvm --version)

# Configuration
$(cat pvm.toml)

# Health Status
$(curl -s -X POST localhost:3000 -d '{"tool":"health_check","arguments":{}}')

# Performance Metrics
$(curl -s -X POST localhost:3000 -d '{"tool":"get_metrics","arguments":{}}')

# System Resources
Memory: $(free -h | grep Mem)
Disk: $(df -h .)
CPU: $(nproc) cores

# Recent Logs
$(tail -50 /tmp/pvm-mcp-debug.log)
EOF
```

### Support Channels

1. **Check documentation**: Review user guide and API reference
2. **Search issues**: Look for similar problems in issue tracker
3. **Create minimal reproduction**: Simplify the problem case
4. **Provide diagnostic information**: Include logs and configuration
5. **Describe expected vs actual behavior**: Be specific about the issue
