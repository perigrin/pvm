# PVM Build System Guide

## Introduction

PVM's modernized build system provides comprehensive development tools, automated code generation, performance monitoring, and CI/CD integration. Built following TypeScript-Go patterns, it offers a production-quality development experience with automated workflows and quality assurance.

## Overview

### Key Features

1. **Automated Tool Management**: Development dependencies managed through `tools.go`
2. **Code Generation**: Automated generation of repetitive code patterns
3. **Build Modes**: Development, release, and testing configurations
4. **Performance Monitoring**: Build performance tracking and optimization
5. **Quality Gates**: Automated linting, formatting, and security scanning
6. **CI/CD Integration**: Comprehensive automation workflows

### Build Architecture

```
Source Code
     ↓
Code Generation (go generate)
     ↓
Build Process (make targets)
     ↓
Quality Checks (lint, test, security)
     ↓
Performance Monitoring
     ↓
Artifacts & Deployment
```

## Makefile Targets

### Core Build Targets

#### Basic Building

```bash
# Build all components
make

# Build specific components
make pvm        # Core version manager
make psc        # Type checker with LSP
make pm        # Module installer
make pvx        # Script executor

# Clean build artifacts
make clean

# Full clean including caches
make clean-all
```

#### Development vs Release Builds

```bash
# Development build (with debug symbols, no embedding)
make build-dev

# Release build (optimized, embedded resources)
make build-release

# Cross-platform builds
make cross-compile

# Create release archives
make release
```

### Tool Management

#### Installation and Setup

```bash
# Install all development tools
make install-tools

# Check tool versions and availability
make check-tools

# Complete development environment setup
make setup

# Update development tools
make update-tools
```

#### Managed Tools

The build system automatically manages these development tools:

- **`stringer`**: Generates String() methods for enums
- **`moq`**: Creates mock implementations for interfaces
- **`gotestsum`**: Enhanced test output and reporting
- **`golangci-lint`**: Comprehensive Go linting
- **`govulncheck`**: Security vulnerability scanning
- **`tree-sitter-cli`**: Parser generator for PSC

### Code Generation

#### Generate All Code

```bash
# Run all code generation
make generate

# Verify generated code is up to date
make check-generate

# Force regeneration of all code
make generate-force
```

#### Specific Generation Targets

```bash
# Generate string methods for enums
make generate-strings

# Generate mocks for interfaces
make generate-mocks

# Generate error codes and messages
make generate-errors

# Generate diagnostic templates
make generate-diagnostics
```

### Testing Infrastructure

#### Basic Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with detailed output
make test-verbose

# Run specific package tests
make test PKG=./internal/parser
```

#### Enhanced Testing

```bash
# Run baseline tests (regression prevention)
make test-baselines

# Update baseline files
make test-baselines-update

# Run performance benchmarks
make test-performance

# Integration tests
make test-integration

# End-to-end tests
make test-e2e
```

#### Test Configuration

```bash
# Run tests with race detection
make test-race

# Run tests with short flag (skip slow tests)
make test-short

# Generate test coverage report
make coverage-html

# Upload coverage to codecov
make coverage-upload
```

### Quality Assurance

#### Code Quality

```bash
# Run all quality checks
make check

# Format code
make fmt

# Lint code
make lint

# Vet code for issues
make vet

# Check imports
make imports
```

#### Security Scanning

```bash
# Security vulnerability scan
make security

# Check for known vulnerabilities
make vulncheck

# Static security analysis
make security-static

# Dependency security audit
make security-deps
```

### Performance Monitoring

#### Performance Analysis

```bash
# Performance optimization
make optimize

# Profile application performance
make profile

# Generate performance analysis report
make performance-analysis

# Monitor build performance
make build-monitor
```

#### Benchmarking

```bash
# Run all benchmarks
make bench

# Compare benchmarks with baseline
make bench-compare

# Profile memory usage
make profile-memory

# Profile CPU usage
make profile-cpu
```

## Code Generation System

### Overview

PVM uses `go generate` directives throughout the codebase to automatically generate repetitive code patterns:

```go
//go:generate stringer -type=TokenType -output=token_type_string.go
//go:generate moq -out parser_mock.go . Parser
//go:generate go run ../../scripts/generate_errors.go ../../scripts/error_definitions.json
```

### Generated Code Types

#### 1. String Methods for Enums

**Purpose**: Automatic `String()` methods for enum types

**Configuration**: Uses `stringer` tool

**Example**:
```go
//go:generate stringer -type=AnnotationKind -output=annotation_kind_string.go

type AnnotationKind int

const (
    AnnotationVariable AnnotationKind = iota
    AnnotationFunction
    AnnotationMethod
)
```

**Generated Output**:
```go
// Code generated by "stringer -type=AnnotationKind"; DO NOT EDIT.

func (i AnnotationKind) String() string {
    switch i {
    case AnnotationVariable:
        return "AnnotationVariable"
    case AnnotationFunction:
        return "AnnotationFunction"
    case AnnotationMethod:
        return "AnnotationMethod"
    default:
        return fmt.Sprintf("AnnotationKind(%d)", int(i))
    }
}
```

#### 2. Mock Generation

**Purpose**: Create test mocks for interfaces

**Configuration**: Uses `moq` tool

**Example**:
```go
//go:generate moq -out parser_mock.go . Parser

type Parser interface {
    ParseFile(filename string) (*ast.File, error)
    ParseExpression(input string) (ast.Expression, error)
}
```

**Generated Usage**:
```go
// In tests
mockParser := &ParserMock{
    ParseFileFunc: func(filename string) (*ast.File, error) {
        return &ast.File{}, nil
    },
}
```

#### 3. Error Code Generation

**Purpose**: Generate structured error codes and messages

**Configuration**: JSON definitions in `scripts/error_definitions.json`

**Definition File**:
```json
{
  "errors": [
    {
      "code": "PSC-E001",
      "severity": "error",
      "category": "UndefinedVariable",
      "template": "Undefined variable '{{.VariableName}}'",
      "help": "Variables must be declared before use with 'my', 'our', or 'state'",
      "suggestions": ["Did you mean: {{range .Suggestions}}{{.}}{{end}}"]
    }
  ]
}
```

**Generated Code**:
```go
// Error codes
const (
    PSCE001 = "PSC-E001"
    PSCE002 = "PSC-E002"
    // ...
)

// Error templates
var ErrorTemplates = map[string]ErrorTemplate{
    PSCE001: {
        Severity: "error",
        Template: "Undefined variable '{{.VariableName}}'",
        Help:     "Variables must be declared before use with 'my', 'our', or 'state'",
    },
}
```

#### 4. Diagnostic Message Generation

**Purpose**: Generate diagnostic message templates and formatters

**Configuration**: JSON definitions in `scripts/diagnostic_definitions.json`

**Definition File**:
```json
{
  "diagnostics": [
    {
      "code": "PSC-W001",
      "severity": "warning",
      "category": "Shadowing",
      "template": "Variable '{{.VariableName}}' shadows outer scope variable",
      "help": "Consider using a different name or accessing outer variable as needed",
      "format": "terminal"
    }
  ]
}
```

### Adding New Generated Code

#### 1. Add Generation Directive

```go
// In your Go file
//go:generate your-generator -args
```

#### 2. Update Makefile

```makefile
# Add to generate target
.PHONY: generate-your-feature
generate-your-feature:
	go generate ./path/to/your/package

generate: generate-your-feature
```

#### 3. Add to CI Verification

The `check-generate` target automatically verifies all generated code is up to date.

## Build Configuration

### Build Tags

PVM uses build tags for conditional compilation:

#### Development Mode

```bash
# Build with debug features
go build -tags debug

# Or use Makefile
make build-dev
```

**Debug Features**:
- Verbose logging
- Debug assertions
- Performance monitoring
- Memory profiling hooks

#### Release Mode

```bash
# Build for production
go build -tags release

# Or use Makefile
make build-release
```

**Release Features**:
- Optimized performance
- Embedded resources
- Minimal logging
- Security hardening

#### Testing Mode

```bash
# Build with test utilities
go build -tags testing

# Used automatically in tests
make test
```

**Testing Features**:
- Test helpers
- Mock implementations
- Debugging utilities
- Coverage instrumentation

### Cross-Platform Building

#### Supported Platforms

```bash
# Build for all supported platforms
make cross-compile

# Supported targets:
# - linux/amd64
# - linux/arm64
# - darwin/amd64
# - darwin/arm64
# - windows/amd64
```

#### Platform-Specific Configuration

```go
// Build tags for platform-specific code
//go:build darwin
// +build darwin

// macOS-specific implementation

//go:build linux
// +build linux

// Linux-specific implementation

//go:build windows
// +build windows

// Windows-specific implementation
```

## Performance Monitoring

### Build Performance

#### Monitoring Build Times

```bash
# Monitor build performance
make build-monitor

# View performance trends
cat .build-metrics/build-times.json
```

#### Performance Metrics

The build system tracks:

- **Compilation time**: Per package and total
- **Test execution time**: Per test and suite
- **Code generation time**: Per generator
- **Dependency resolution**: Module download times
- **Cache hit rates**: Build cache effectiveness

#### Performance Reports

```bash
# Generate comprehensive performance report
make performance-report

# View in browser
open .build-metrics/performance-report.html
```

### Runtime Performance

#### Profiling Integration

```bash
# Profile parser performance
make profile-parser

# Profile type checker performance
make profile-checker

# Profile LSP performance
make profile-lsp

# Analyze profiles
go tool pprof cpu.prof
```

#### Benchmark Tracking

```bash
# Run benchmarks and track results
make bench-track

# Compare with previous runs
make bench-compare

# Generate benchmark report
make bench-report
```

## CI/CD Integration

### GitHub Actions Workflows

#### Cross-Platform Testing

```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go-version: [1.21, 1.22]

    runs-on: ${{ matrix.os }}

    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}

    - name: Install tools
      run: make install-tools

    - name: Verify code generation
      run: make check-generate

    - name: Run tests
      run: make test-all

    - name: Upload coverage
      uses: codecov/codecov-action@v3
```

#### Security Scanning

```yaml
# .github/workflows/security.yml
name: Security

on: [push, pull_request]

jobs:
  security:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: 1.22

    - name: Run Gosec
      run: make security-static

    - name: Run vulnerability check
      run: make vulncheck

    - name: Run dependency audit
      run: make security-deps
```

#### Performance Monitoring

```yaml
# .github/workflows/performance.yml
name: Performance

on: [push, pull_request]

jobs:
  performance:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: 1.22

    - name: Run benchmarks
      run: make bench-compare

    - name: Check performance regressions
      run: make performance-check

    - name: Comment PR with results
      if: github.event_name == 'pull_request'
      run: make performance-comment
```

### Quality Gates

#### Coverage Thresholds

```yaml
# Quality gate configuration
quality:
  coverage:
    minimum: 70%
    target: 85%

  performance:
    regression_threshold: 10%

  security:
    fail_on_vulnerability: true
```

#### Automated Checks

```bash
# Run all quality gates locally
make quality-gates

# Individual checks
make coverage-check     # Verify coverage meets threshold
make performance-check  # Check for performance regressions
make security-check     # Verify no security issues
make generate-check     # Verify generated code is up to date
```

## Development Workflow

### Daily Development

#### Setup

```bash
# One-time setup
make setup

# Daily workflow
git checkout -b feature/my-feature
make build-dev
```

#### Development Cycle

```bash
# Make changes
# ...

# Generate code if needed
make generate

# Run tests
make test

# Check quality
make check

# Commit changes
git add .
git commit -m "feature: implement my feature"
```

#### Pre-commit Hooks

```bash
# Install pre-commit hooks
make install-hooks

# Hooks will automatically run:
# - Code formatting
# - Linting
# - Basic tests
# - Generate verification
```

### Release Process

#### Preparation

```bash
# Prepare release
make release-prepare VERSION=v1.2.3

# This will:
# - Update version numbers
# - Generate changelog
# - Run full test suite
# - Build all platforms
# - Create release archives
```

#### Release Workflow

```bash
# Tag release
git tag v1.2.3

# Push tag (triggers CI)
git push origin v1.2.3

# CI will:
# - Build all platforms
# - Run comprehensive tests
# - Perform security scanning
# - Create GitHub release
# - Upload artifacts
```

## Troubleshooting

### Common Build Issues

#### Tool Installation Failures

```bash
# Check tool status
make check-tools

# Reinstall tools
make install-tools

# Manual tool installation
go install golang.org/x/tools/cmd/stringer@latest
npm install -g tree-sitter-cli
```

#### Code Generation Issues

```bash
# Check for outdated generated code
make check-generate

# Force regeneration
make generate-force

# Debug generation
make generate-debug
```

#### Performance Issues

```bash
# Profile build performance
make build-monitor

# Check cache usage
make cache-stats

# Clean caches
make clean-cache
```

### Debug Mode

```bash
# Enable debug output
export PVM_BUILD_DEBUG=1

# Run with debug information
make build-dev

# Check build logs
tail -f .build-metrics/build.log
```

## Best Practices

### Code Generation

1. **Keep generators simple**: Focus on specific patterns
2. **Use JSON configuration**: For complex template-based generation
3. **Verify in CI**: Always check generated code is up to date
4. **Document patterns**: Explain when and how to add new generation

### Testing

1. **Use baseline tests**: For regression prevention
2. **Monitor performance**: Track benchmark results over time
3. **Test across platforms**: Ensure compatibility
4. **Mock external dependencies**: For reliable testing

### Performance

1. **Profile before optimizing**: Use data to guide decisions
2. **Monitor build times**: Keep build performance acceptable
3. **Use caching effectively**: Leverage Go build cache and custom caches
4. **Set performance budgets**: Define acceptable performance thresholds

### Quality

1. **Use linting consistently**: Automate code quality checks
2. **Monitor security**: Regular vulnerability scanning
3. **Maintain coverage**: Track test coverage trends
4. **Review generated code**: Understand what's being generated

The PVM build system provides a comprehensive, modern development experience with automated workflows, quality assurance, and performance monitoring that scales from individual development to large team collaboration.
