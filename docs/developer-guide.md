# PVM Developer Guide

## Introduction

This guide provides comprehensive information for developers working with PVM's modernized compiler architecture. Whether you're contributing to PVM development or building tools that integrate with PVM, this guide covers the essential concepts and APIs.

## Getting Started

### Development Environment Setup

1. **Install development tools**:
   ```bash
   make install-tools
   ```

2. **Build the project**:
   ```bash
   make
   ```

3. **Run tests**:
   ```bash
   make test
   ```

4. **Verify code generation**:
   ```bash
   make check-generate
   ```

### Development Workflow

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature
   ```

2. **Make changes** following TDD principles
3. **Run comprehensive tests**:
   ```bash
   make test-all
   ```

4. **Update baselines if needed**:
   ```bash
   make test-baselines-update
   ```

5. **Check performance**:
   ```bash
   make performance-analysis
   ```

## Compiler Pipeline Architecture

### Overview

The PVM compiler follows a multi-stage pipeline:

```
Source → Scanner → Parser → Binder → Checker → Compiler → Output
```

Each stage has specific responsibilities and well-defined interfaces.

### Working with the Scanner

#### Basic Usage

```go
import "github.com/pvm/internal/scanner"

// Create a scanner for a file
scanner, err := scanner.NewScanner("script.pl")
if err != nil {
    log.Fatal(err)
}

// Tokenize the input
for {
    token, err := scanner.NextToken()
    if err != nil {
        break
    }
    fmt.Printf("Token: %s at %s\n", token.Type, token.Position)
}
```

#### Scanner Interface

```go
type Scanner interface {
    NextToken() (*Token, error)
    Peek() (*Token, error)
    Position() Position
    Reset() error
}

type Token struct {
    Type     TokenType
    Value    string
    Position Position
}
```

#### Performance Considerations

- Use `Peek()` for lookahead without consuming tokens
- Scanner maintains position state for error reporting
- Tree-sitter integration provides efficient tokenization

### Working with the Parser

#### Basic Usage

```go
import "github.com/pvm/internal/parser"

// Parse a file
parser := parser.NewParser()
ast, err := parser.ParseFile("script.pl")
if err != nil {
    log.Fatal(err)
}

// Work with the AST
for _, stmt := range ast.Statements {
    fmt.Printf("Statement: %T\n", stmt)
}
```

#### AST Navigation

```go
import "github.com/pvm/internal/astnav"

// Create a navigator
nav := astnav.NewNavigator(ast)

// Find nodes at a specific position
nodes := nav.FindNodeAt(position)

// Walk the AST
astnav.Walk(ast, func(node ast.Node) bool {
    fmt.Printf("Visiting: %T\n", node)
    return true // continue walking
})
```

#### AST Node Types

```go
// Expression types
type Expression interface {
    ast.Node
    ExpressionType() string
}

// Statement types
type Statement interface {
    ast.Node
    StatementType() string
}

// Common expressions
type VariableRef struct {
    Name     string
    Position ast.Position
}

type Assignment struct {
    Left  Expression
    Right Expression
    Position ast.Position
}
```

### Working with the Binder

#### Basic Usage

```go
import "github.com/pvm/internal/binder"

// Create a binder
binder := binder.NewBinder()

// Bind symbols in an AST
symbolTable, err := binder.BindFile(ast)
if err != nil {
    log.Fatal(err)
}

// Look up symbols
symbol, err := symbolTable.LookupSymbol("$variable", position)
if err != nil {
    log.Printf("Symbol not found: %v", err)
}
```

#### Symbol Types

```go
type Symbol struct {
    Name         string
    Kind         SymbolKind
    Type         *Type
    Position     ast.Position
    Scope        *Scope
    Declarations []*Declaration
    References   []*Reference
}

type SymbolKind int
const (
    SymbolVariable SymbolKind = iota
    SymbolSubroutine
    SymbolMethod
    SymbolPackage
    SymbolModule
)
```

#### Working with Scopes

```go
// Get the current scope
scope := symbolTable.GetScope(position)

// Find symbols in scope
symbols := scope.GetSymbols()

// Check if symbol exists in scope
if scope.HasSymbol("$var") {
    symbol := scope.GetSymbol("$var")
}

// Handle Perl scoping rules
switch varDecl.Type {
case "my":
    // Lexical scoping
    scope.AddLexicalSymbol(symbol)
case "our":
    // Package scoping
    scope.AddPackageSymbol(symbol)
case "local":
    // Dynamic scoping
    scope.AddLocalSymbol(symbol)
}
```

### Working with the Type Checker

#### Basic Usage

```go
import "github.com/pvm/internal/typechecker"

// Create a type checker
checker := typechecker.NewTypeChecker(symbolTable)

// Check types in a file
typeInfo, err := checker.CheckFile(ast)
if err != nil {
    log.Fatal(err)
}

// Get type information for expressions
exprType := typeInfo.GetExpressionType(expression)
```

#### Type System

```go
// Basic types
type Type struct {
    Kind TypeKind
    Name string
    // Union types: Int|Str
    Union []Type
    // Intersection types: Object&Serializable
    Intersection []Type
    // Parameterized types: ArrayRef[Int]
    Parameters []Type
}

type TypeKind int
const (
    TypeScalar TypeKind = iota
    TypeArray
    TypeHash
    TypeSubroutine
    TypeUnion
    TypeIntersection
)
```

#### Type Inference

```go
// Infer type from expression
inferredType, err := checker.InferType(expression)
if err != nil {
    log.Printf("Type inference failed: %v", err)
}

// Check type compatibility
compatible := checker.IsCompatible(sourceType, targetType)

// Get type constraints
constraints := checker.GetConstraints(symbol)
```

### Working with the Compiler

#### Basic Usage

```go
import "github.com/pvm/internal/compiler"

// Create a compiler registry
registry := compiler.NewCompilerRegistry()

// Create an AST adapter
adapter := compiler.NewParserASTAdapter(ast)

// Compile to different targets
cleanPerl, err := registry.Compile(adapter, compiler.TargetCleanPerl)
if err != nil {
    log.Fatal(err)
}

typedPerl, err := registry.Compile(adapter, compiler.TargetTypedPerl)
if err != nil {
    log.Fatal(err)
}
```

#### Compilation Targets

```go
// Available targets
const (
    TargetCleanPerl TargetType = "clean_perl"  // Standard Perl without types
    TargetTypedPerl TargetType = "typed_perl"  // Perl with type annotations
)

// Register custom targets
registry.RegisterCompiler(TargetJavaScript, &JavaScriptCompiler{})
```

## Language Service Integration

### Basic LSP Features

```go
import "github.com/pvm/internal/ls"

// Create a language service
ls := ls.NewLanguageService(parser, binder, checker)

// Implement LSP features
definition, err := ls.GetDefinition(uri, position)
references, err := ls.FindReferences(uri, position)
hover, err := ls.GetHover(uri, position)
completions, err := ls.GetCompletions(uri, position)
```

### Enhanced Features

```go
// Rename symbol across files
edits, err := ls.RenameSymbol(uri, position, newName)

// Get document symbols
symbols, err := ls.GetDocumentSymbols(uri)

// Search workspace symbols
results, err := ls.GetWorkspaceSymbols(query)
```

## Build System Integration

### Code Generation

#### Adding Generated Code

1. **Add go:generate directive**:
   ```go
   //go:generate stringer -type=MyEnum -output=myenum_string.go
   ```

2. **Update Makefile target**:
   ```makefile
   .PHONY: generate
   generate:
       go generate ./...
   ```

3. **Verify in CI**:
   ```bash
   make check-generate
   ```

#### Mock Generation

```go
// Add mock generation directive
//go:generate moq -out my_interface_mock.go . MyInterface

type MyInterface interface {
    DoSomething(arg string) error
}
```

### Testing Framework

#### Baseline Testing

```go
import "github.com/pvm/internal/testing/baseline"

func TestMyFeature_Baselines(t *testing.T) {
    baseline.RunTest(t, baseline.TestCase{
        Name:     "simple_case",
        Input:    "my Int $x = 42;",
        Expected: "expected_output.txt",
        Runner:   myFeatureRunner,
    })
}
```

#### Performance Testing

```go
import "github.com/pvm/internal/testing/performance"

func BenchmarkMyFeature(b *testing.B) {
    monitor := performance.NewMonitor()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        monitor.Start("operation")
        result := myFeature(input)
        monitor.End("operation")
    }

    monitor.Report()
}
```

## Error Handling and Diagnostics

### Enhanced Diagnostics

```go
import "github.com/pvm/internal/diagnostics"

// Create enhanced diagnostic engine
engine := diagnostics.NewEnhancedDiagnosticEngine(binder)

// Generate diagnostics
diagnostics, err := engine.GenerateDiagnostics(ast, symbolTable)

// Format for display
for _, diag := range diagnostics {
    formatted := diag.Format(diagnostics.FormatOptions{
        Color:       true,
        ShowContext: true,
    })
    fmt.Println(formatted)
}
```

### Custom Diagnostic Types

1. **Define error code** in `scripts/diagnostic_definitions.json`:
   ```json
   {
     "code": "PSC-E003",
     "severity": "error",
     "template": "Custom error: {message}",
     "help": "Custom help message"
   }
   ```

2. **Generate code**:
   ```bash
   make generate
   ```

3. **Use in diagnostics**:
   ```go
   diagnostic := &diagnostics.Diagnostic{
       Code:    diagnostics.PSCE003,
       Message: "Custom error occurred",
       Pos:     position,
   }
   ```

## Performance Optimization

### Profiling

```bash
# Profile parser performance
make profile-parser

# Profile type checker performance
make profile-checker

# Analyze results
go tool pprof cpu.prof
```

### Optimization Techniques

#### Object Pooling

```go
import "github.com/pvm/internal/memory"

// Use object pools for frequently allocated objects
nodePool := memory.NewNodePool()
defer nodePool.Close()

// Get from pool
node := nodePool.GetNode()
defer nodePool.PutNode(node)
```

#### Caching

```go
import "github.com/pvm/internal/cache"

// Create multi-level cache
cache := cache.NewMultiLevelCache()

// Cache expensive operations
key := cache.KeyFor(input)
if result, ok := cache.Get(key); ok {
    return result
}

result := expensiveOperation(input)
cache.Set(key, result)
```

### Performance Monitoring

```go
import "github.com/pvm/internal/performance"

// Monitor operation performance
monitor := performance.NewMonitor()
monitor.Start("type_checking")
defer monitor.End("type_checking")

// Track metrics
monitor.Counter("symbols_resolved").Inc()
monitor.Histogram("parse_time").Observe(duration)
```

## Contributing Guidelines

### Code Style

1. **Follow Go conventions**: Use `gofmt`, `goimports`, `golint`
2. **Add package comments**: Every package should have a doc comment
3. **Use meaningful names**: Prefer clarity over brevity
4. **Handle errors explicitly**: Don't ignore errors

### Testing Requirements

1. **Write tests first**: Follow TDD principles
2. **Test all public APIs**: Ensure comprehensive coverage
3. **Add integration tests**: Test component interactions
4. **Update baselines**: When behavior changes intentionally

### Performance Considerations

1. **Profile before optimizing**: Use data to guide optimization
2. **Optimize hot paths**: Focus on high-impact areas
3. **Monitor regressions**: Use performance tests in CI
4. **Document trade-offs**: Explain performance decisions

### Documentation

1. **Update user docs**: For user-facing changes
2. **Update API docs**: For interface changes
3. **Add examples**: Show real usage patterns
4. **Update migration guides**: For breaking changes

## Troubleshooting

### Common Issues

#### Build Failures

```bash
# Clean and rebuild
make clean && make

# Update dependencies
go mod tidy

# Regenerate code
make generate
```

#### Test Failures

```bash
# Run specific test
go test -v ./internal/parser -run TestSpecificFunction

# Update baselines
make test-baselines-update

# Check race conditions
go test -race ./...
```

#### Performance Issues

```bash
# Profile the issue
make profile

# Check for memory leaks
go test -memprofile=mem.prof

# Analyze allocation patterns
go tool pprof mem.prof
```

### Debug Mode

```bash
# Enable debug logging
export PVM_DEBUG=1

# Run with debug output
psc --debug check script.pl

# Analyze debug output
psc --debug --verbose check script.pl 2>&1 | grep "symbol"
```

## Next Steps

### Learning Path

1. **Start with examples**: Use the test files as reference
2. **Read the architecture overview**: Understand the big picture
3. **Implement a simple feature**: Practice with the APIs
4. **Contribute to PVM**: Join the development community

### Resources

- **API Documentation**: Generated docs in `docs/api/`
- **Test Examples**: Real usage in `internal/*/test.go` files
- **Performance Reports**: Benchmarks in `.build-metrics/`
- **Architecture Diagrams**: Visual guides in `docs/diagrams/`

### Community

- **GitHub Issues**: Report bugs and request features
- **Discussions**: Ask questions and share ideas
- **Pull Requests**: Contribute code and documentation
- **Code Reviews**: Learn from experienced contributors

This guide provides the foundation for effective PVM development. As you work with the system, refer back to these patterns and principles to maintain consistency and quality in your contributions.
