# PVM Contributor Guide

## Welcome to PVM Development

This guide helps new and experienced contributors understand PVM's modernized architecture and development practices. PVM follows TypeScript-Go patterns for clean, maintainable, and performant code.

## Getting Started

### Prerequisites

- **Go 1.21+**: Required for building PVM
- **Node.js 16+**: Required for tree-sitter (PSC component)
- **Git**: Version control
- **Make**: Build automation
- **Editor with Go support**: VS Code, GoLand, Vim, or Emacs

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/pvm/pvm.git
cd pvm

# Set up development environment
make setup

# Verify installation
make test
```

### Development Environment

```bash
# Install development tools
make install-tools

# Install pre-commit hooks
make install-hooks

# Build in development mode
make build-dev

# Run tests with coverage
make test-coverage
```

## Architecture Overview

### Compiler Pipeline

PVM follows a multi-stage compiler architecture:

```
Source → Scanner → Parser → Binder → Checker → Compiler → Output
```

Each stage has specific responsibilities:

- **Scanner** (`internal/scanner/`): Lexical analysis and tokenization
- **Parser** (`internal/parser/`): AST generation from tokens
- **Binder** (`internal/binder/`): Symbol resolution and scope management
- **Checker** (`internal/typechecker/`): Type analysis using symbol information
- **Compiler** (`internal/compiler/`): Code generation to target formats

### Language Server Architecture

Clean separation between business logic and protocol handling:

- **Language Service** (`internal/ls/`): Core language features
- **LSP Server** (`internal/lsp/`): Protocol communication

### Component Structure

```
pvm/
├── cmd/                    # Command entry points
│   ├── pvm/               # Main version manager
│   ├── psc/               # Type checker and LSP
│   ├── pm/               # Module installer
│   └── pvx/               # Script executor
├── internal/              # Internal packages
│   ├── ast/               # Consolidated AST types
│   ├── astnav/           # AST navigation utilities
│   ├── binder/           # Symbol binding and resolution
│   ├── compiler/         # Code generation
│   ├── diagnostics/      # Enhanced error reporting
│   ├── ls/               # Language service
│   ├── lsp/              # LSP protocol handler
│   ├── parser/           # Parsing and tree-sitter integration
│   ├── scanner/          # Lexical analysis
│   └── typechecker/      # Type checking and inference
├── docs/                 # Documentation
├── test/                 # Test fixtures and helpers
└── scripts/              # Build and development scripts
```

## Development Workflow

### Creating a Feature

#### 1. Plan the Feature

1. **Create an issue**: Describe the feature and requirements
2. **Design discussion**: Discuss architecture and approach
3. **Break down work**: Identify affected components
4. **Write tests first**: Following TDD principles

#### 2. Implementation Process

```bash
# Create feature branch
git checkout -b feature/your-feature-name

# Make changes following TDD
# 1. Write failing test
# 2. Implement minimum code to pass
# 3. Refactor while keeping tests green
# 4. Repeat

# Generate any needed code
make generate

# Run tests frequently
make test

# Check code quality
make check
```

#### 3. Testing Requirements

All contributions must include comprehensive tests:

```bash
# Unit tests for new functionality
go test ./internal/yourpackage/

# Integration tests if applicable
make test-integration

# Update baselines if behavior changes
make test-baselines-update

# Performance tests for performance-critical code
make test-performance
```

#### 4. Documentation Updates

- Update relevant documentation in `docs/`
- Add code comments for complex logic
- Update API documentation for public interfaces
- Add examples for new features

### Code Quality Standards

#### Go Code Style

1. **Follow Go conventions**: Use `gofmt`, `goimports`, `golint`
2. **Package documentation**: Every package needs a doc comment
3. **Function documentation**: Document exported functions
4. **Error handling**: Handle errors explicitly, don't ignore them
5. **Meaningful names**: Choose clear, descriptive names

**Example Good Code**:
```go
// Package binder provides symbol resolution and scope management
// for the PVM compiler pipeline.
package binder

// Symbol represents a named entity in the source code with its
// type information and scope context.
type Symbol struct {
    Name         string
    Kind         SymbolKind
    Type         *Type
    Position     ast.Position
    Scope        *Scope
    Declarations []*Declaration
    References   []*Reference
}

// LookupSymbol finds a symbol by name at the given position,
// respecting Perl's scoping rules.
func (st *SymbolTable) LookupSymbol(name string, pos ast.Position) (*Symbol, error) {
    if name == "" {
        return nil, errors.New("symbol name cannot be empty")
    }

    scope := st.getScopeAt(pos)
    if scope == nil {
        return nil, fmt.Errorf("no scope found at position %v", pos)
    }

    return scope.lookup(name)
}
```

#### Architecture Principles

1. **Separation of concerns**: Each package has a single responsibility
2. **Interface-based design**: Use interfaces for testability
3. **Dependency injection**: Avoid global state
4. **Error handling**: Return errors, don't panic
5. **Performance awareness**: Consider performance implications

#### Testing Best Practices

1. **Test-driven development**: Write tests before implementation
2. **Comprehensive coverage**: Aim for >80% test coverage
3. **Unit tests**: Test individual functions and methods
4. **Integration tests**: Test component interactions
5. **Baseline tests**: Prevent regressions with expected outputs
6. **Performance tests**: Benchmark critical operations

**Example Test Structure**:
```go
func TestSymbolTable_LookupSymbol(t *testing.T) {
    tests := []struct {
        name     string
        setup    func() *SymbolTable
        symbol   string
        position ast.Position
        want     *Symbol
        wantErr  bool
    }{
        {
            name: "find local variable",
            setup: func() *SymbolTable {
                st := NewSymbolTable()
                // Setup test data
                return st
            },
            symbol:   "$var",
            position: ast.Position{Line: 5, Column: 10},
            want:     &Symbol{Name: "$var", Kind: SymbolVariable},
            wantErr:  false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            st := tt.setup()
            got, err := st.LookupSymbol(tt.symbol, tt.position)

            if (err != nil) != tt.wantErr {
                t.Errorf("LookupSymbol() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("LookupSymbol() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Contributing to Specific Components

### Scanner/Parser Components

**Location**: `internal/scanner/`, `internal/parser/`

**Key Concepts**:
- Tree-sitter integration for parsing
- Token-based lexical analysis
- Fast parser for simple patterns
- Full parser for complex constructs

**Contribution Areas**:
- New Perl syntax support
- Performance optimizations
- Error recovery improvements
- Parser caching enhancements

**Example Contribution**:
```go
// Add support for new Perl syntax
func (p *Parser) parseNewSyntax() (ast.Expression, error) {
    // Implementation following existing patterns
    if !p.scanner.Match(TokenKeyword, "new_syntax") {
        return nil, p.expectError("new_syntax")
    }

    // Parse syntax elements
    expr, err := p.parseExpression()
    if err != nil {
        return nil, err
    }

    return &ast.NewSyntaxExpr{
        Expression: expr,
        Position:   p.scanner.Position(),
    }, nil
}
```

### Symbol Binding (Binder)

**Location**: `internal/binder/`

**Key Concepts**:
- Perl scoping rules (lexical, package, dynamic)
- Symbol resolution across modules
- Scope chain management
- Cross-reference tracking

**Contribution Areas**:
- Advanced scoping features
- Module import/export handling
- Performance optimizations
- Symbol table serialization

**Example Contribution**:
```go
// Add support for new symbol type
func (b *Binder) bindMethodSymbol(node *ast.MethodDecl) error {
    symbol := &Symbol{
        Name:     node.Name,
        Kind:     SymbolMethod,
        Position: node.Position,
        Scope:    b.currentScope,
    }

    // Handle method-specific binding logic
    if node.IsStatic {
        symbol.Flags |= SymbolStatic
    }

    return b.currentScope.AddSymbol(symbol)
}
```

### Type Checking

**Location**: `internal/typechecker/`

**Key Concepts**:
- Flow-sensitive type analysis
- Union and intersection types
- Type inference and compatibility
- Symbol-aware checking

**Contribution Areas**:
- New type features
- Improved inference algorithms
- Performance optimizations
- Enhanced error messages

**Example Contribution**:
```go
// Add support for new type checking rule
func (tc *TypeChecker) checkNewTypeRule(node ast.Node, expectedType *Type) error {
    actualType, err := tc.inferType(node)
    if err != nil {
        return err
    }

    if !tc.isCompatible(actualType, expectedType) {
        return &TypeError{
            Position: node.GetPosition(),
            Expected: expectedType,
            Actual:   actualType,
            Message:  "type compatibility check failed",
        }
    }

    return nil
}
```

### LSP Implementation

**Location**: `internal/ls/`, `internal/lsp/`

**Key Concepts**:
- Language service vs protocol separation
- Symbol-aware LSP features
- Caching and performance optimization
- Real-time diagnostics

**Contribution Areas**:
- New LSP features
- Performance improvements
- Editor-specific optimizations
- Enhanced user experience

**Example Contribution**:
```go
// Add new LSP feature
func (ls *LanguageService) GetNewFeature(uri string, pos Position) (*NewFeatureResult, error) {
    // Parse and analyze the document
    doc, err := ls.getDocument(uri)
    if err != nil {
        return nil, err
    }

    // Use symbol information for accurate results
    symbol, err := ls.binder.ResolveSymbolAt(doc.AST, pos)
    if err != nil {
        return nil, err
    }

    // Generate feature-specific result
    return &NewFeatureResult{
        Symbol:   symbol,
        Location: symbol.Position,
        // Additional result data
    }, nil
}
```

### Enhanced Diagnostics

**Location**: `internal/diagnostics/`

**Key Concepts**:
- Context-aware error messages
- Symbol information integration
- Actionable suggestions
- Error code management

**Contribution Areas**:
- New diagnostic types
- Improved error messages
- Performance optimizations
- Integration with other components

**Example Contribution**:
```go
// Add new diagnostic type
func (de *DiagnosticEngine) detectNewIssue(ast *ast.File, symbols *binder.SymbolTable) []*Diagnostic {
    var diagnostics []*Diagnostic

    // Walk AST looking for new issue pattern
    astnav.Walk(ast, func(node ast.Node) bool {
        if issue := de.checkForNewIssue(node, symbols); issue != nil {
            diagnostic := &Diagnostic{
                Kind:        DiagnosticWarning,
                Code:        "PSC-W003",
                Message:     issue.Description,
                Position:    node.GetPosition(),
                Symbol:      issue.Symbol,
                Suggestion:  issue.FixSuggestion,
                HelpMessage: "How to fix this issue...",
            }
            diagnostics = append(diagnostics, diagnostic)
        }
        return true
    })

    return diagnostics
}
```

## Code Generation

### Adding Generated Code

PVM uses `go generate` for automated code generation:

#### 1. String Methods for Enums

```go
//go:generate stringer -type=YourEnum -output=your_enum_string.go

type YourEnum int

const (
    YourEnumValue1 YourEnum = iota
    YourEnumValue2
    YourEnumValue3
)
```

#### 2. Mock Generation

```go
//go:generate moq -out your_interface_mock.go . YourInterface

type YourInterface interface {
    Method1(arg string) error
    Method2() (int, error)
}
```

#### 3. Custom Generators

Create custom generators for complex patterns:

```go
//go:generate go run ../../scripts/generate_your_feature.go config.json

// Implementation in scripts/generate_your_feature.go
func main() {
    configFile := os.Args[1]
    // Read configuration
    // Generate code based on templates
    // Write output files
}
```

### Code Generation Best Practices

1. **Keep generators simple**: Focus on specific, repetitive patterns
2. **Use configuration files**: JSON/YAML for complex generation
3. **Generate to separate files**: Don't mix generated and hand-written code
4. **Add build verification**: Use `make check-generate` in CI
5. **Document generation**: Explain when and how to regenerate

## Performance Considerations

### Profiling and Optimization

#### Identifying Bottlenecks

```bash
# Profile specific components
make profile-parser
make profile-binder
make profile-checker

# Analyze results
go tool pprof cpu.prof
```

#### Common Optimization Patterns

1. **Object pooling**: For frequently allocated objects
```go
var nodePool = sync.Pool{
    New: func() interface{} {
        return &ast.Node{}
    },
}

func getNode() *ast.Node {
    return nodePool.Get().(*ast.Node)
}

func putNode(n *ast.Node) {
    n.Reset()
    nodePool.Put(n)
}
```

2. **String interning**: For commonly used strings
```go
type StringInterner struct {
    mu    sync.RWMutex
    cache map[string]string
}

func (si *StringInterner) Intern(s string) string {
    si.mu.RLock()
    if interned, ok := si.cache[s]; ok {
        si.mu.RUnlock()
        return interned
    }
    si.mu.RUnlock()

    si.mu.Lock()
    defer si.mu.Unlock()
    si.cache[s] = s
    return s
}
```

3. **Caching expensive operations**:
```go
type CachedChecker struct {
    cache map[string]*TypeInfo
    mu    sync.RWMutex
}

func (cc *CachedChecker) CheckFile(file *ast.File) (*TypeInfo, error) {
    key := file.Hash()

    cc.mu.RLock()
    if cached, ok := cc.cache[key]; ok {
        cc.mu.RUnlock()
        return cached, nil
    }
    cc.mu.RUnlock()

    result, err := cc.doCheck(file)
    if err != nil {
        return nil, err
    }

    cc.mu.Lock()
    cc.cache[key] = result
    cc.mu.Unlock()

    return result, nil
}
```

### Performance Testing

Add benchmarks for performance-critical code:

```go
func BenchmarkSymbolLookup(b *testing.B) {
    st := setupLargeSymbolTable()
    pos := ast.Position{Line: 100, Column: 50}

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := st.LookupSymbol("$common_var", pos)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkParseFile(b *testing.B) {
    source := loadLargeSourceFile()
    parser := NewParser()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := parser.ParseString(source)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Documentation Guidelines

### Code Documentation

1. **Package documentation**: Describe package purpose and main concepts
2. **Function documentation**: Document exported functions with examples
3. **Complex logic**: Add comments for non-obvious code
4. **Examples**: Include usage examples for public APIs

**Example Package Documentation**:
```go
// Package binder implements symbol resolution and scope management
// for the PVM compiler pipeline.
//
// The binder is responsible for building symbol tables by analyzing
// AST nodes and resolving symbol references according to Perl's
// scoping rules. It supports:
//
//   - Lexical scoping (my variables)
//   - Package scoping (our variables)
//   - Dynamic scoping (local variables)
//   - Cross-module symbol resolution
//
// Basic usage:
//
//   binder := binder.NewBinder()
//   symbolTable, err := binder.BindFile(ast)
//   if err != nil {
//       log.Fatal(err)
//   }
//
//   symbol, err := symbolTable.LookupSymbol("$var", position)
package binder
```

### User Documentation

Update user-facing documentation when adding features:

1. **Architecture overview**: Update if architecture changes
2. **User guides**: Add new features to relevant guides
3. **API documentation**: Document new public APIs
4. **Migration guides**: Update for breaking changes

## Debugging and Troubleshooting

### Debug Mode

Enable detailed debugging output:

```bash
# Enable debug logging
export PVM_DEBUG=1

# Component-specific debugging
export PVM_PARSER_DEBUG=1
export PVM_BINDER_DEBUG=1
export PVM_CHECKER_DEBUG=1
export PVM_LSP_DEBUG=1

# Run with debugging
psc --debug check script.pl
```

### Common Issues

#### Build Issues

```bash
# Clean and rebuild
make clean && make

# Update dependencies
go mod tidy

# Regenerate code
make generate

# Check tool versions
make check-tools
```

#### Test Failures

```bash
# Run specific test with verbose output
go test -v ./internal/binder -run TestSpecificFunction

# Run with race detection
go test -race ./...

# Update baselines if behavior changed intentionally
make test-baselines-update
```

#### Performance Issues

```bash
# Profile the problematic component
make profile-component

# Check for memory leaks
go test -memprofile=mem.prof ./...

# Analyze allocation patterns
go tool pprof mem.prof
```

## Review Process

### Pull Request Guidelines

1. **Small, focused changes**: Keep PRs small and focused on one feature
2. **Clear description**: Explain what changes and why
3. **Test coverage**: Include comprehensive tests
4. **Documentation**: Update relevant documentation
5. **Performance impact**: Note any performance implications

### Code Review Checklist

**For Reviewers**:
- [ ] Code follows Go conventions and project style
- [ ] Tests are comprehensive and pass
- [ ] Documentation is updated
- [ ] Performance impact is acceptable
- [ ] Error handling is appropriate
- [ ] API design is consistent
- [ ] Security implications are considered

**For Contributors**:
- [ ] All tests pass locally
- [ ] Code is formatted (`make fmt`)
- [ ] Linting passes (`make lint`)
- [ ] Generated code is up to date (`make check-generate`)
- [ ] Documentation is updated
- [ ] Commit messages are clear and descriptive

## Release Process

### Version Management

PVM uses semantic versioning:

- **Major versions**: Breaking changes
- **Minor versions**: New features (backward compatible)
- **Patch versions**: Bug fixes

### Release Checklist

1. **Update version numbers**: In relevant files
2. **Update CHANGELOG**: Document changes
3. **Run full test suite**: Ensure everything works
4. **Update documentation**: Reflect any changes
5. **Create release tag**: Follow version format
6. **Build release artifacts**: For all platforms
7. **Publish release**: With release notes

## Getting Help

### Resources

1. **Documentation**: Comprehensive guides in `docs/`
2. **Code examples**: Test files and examples
3. **GitHub Issues**: Bug reports and feature requests
4. **Discussions**: Architecture and design discussions
5. **Code review**: Learn from experienced contributors

### Community Guidelines

1. **Be respectful**: Treat everyone with respect
2. **Be helpful**: Help others learn and contribute
3. **Be patient**: Allow time for discussion and review
4. **Be collaborative**: Work together toward common goals

### Asking for Help

When asking for help:

1. **Be specific**: Describe exactly what you're trying to do
2. **Show your work**: Include code, error messages, and steps taken
3. **Be patient**: Allow time for responses
4. **Follow up**: Let others know if their suggestions helped

## Advanced Topics

### Adding New Language Features

1. **Parse new syntax**: Update parser with new grammar rules
2. **Add AST nodes**: Create new node types for the feature
3. **Update binder**: Handle new symbols and scoping rules
4. **Update checker**: Implement type checking logic
5. **Add LSP support**: Provide language server features
6. **Update compiler**: Generate appropriate output
7. **Add diagnostics**: Provide helpful error messages

### Performance Optimization

1. **Profile first**: Identify actual bottlenecks
2. **Measure impact**: Benchmark before and after changes
3. **Consider trade-offs**: Balance performance vs maintainability
4. **Test thoroughly**: Ensure optimizations don't break functionality
5. **Document changes**: Explain optimization decisions

### Testing Strategy

1. **Unit tests**: Test individual functions and methods
2. **Integration tests**: Test component interactions
3. **End-to-end tests**: Test complete workflows
4. **Performance tests**: Benchmark critical operations
5. **Regression tests**: Prevent breaking existing functionality

Contributing to PVM is rewarding and helps advance Perl development tooling. This guide provides the foundation for effective contributions, and the community is always ready to help new contributors succeed.
