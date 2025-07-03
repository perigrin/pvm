# Unified Compiler Architecture Documentation

## Overview

This document describes the unified compiler architecture implemented in PVM, which replaces the previous separate `CleanPerlCompiler` and `TypedPerlCompiler` classes with a single, more efficient implementation that works directly with tree-sitter's Concrete Syntax Tree (CST).

## Architectural Transformation

### Before: Legacy Architecture

```
Source Code
    ↓
Tree-sitter CST
    ↓
CST-to-AST Conversion (BUGGY)
    ↓
Separate Compiler Classes
├── CleanPerlCompiler (AST-based)
└── TypedPerlCompiler (AST-based)
    ↓
Output Code
```

**Problems with Legacy Architecture:**
- Buggy CST-to-AST conversion layer
- `VarDecl.LogicalVariables()` returned type names instead of variable names
- Code duplication between compiler implementations
- Loss of precise source positioning during conversion
- Inconsistent handling of type annotations

### After: Unified Architecture

```
Source Code
    ↓
Tree-sitter CST
    ↓
Unified PerlCompiler (CST-based)
├── Tree Transformation Rules
├── Type Annotation Removal (Optional)
└── Direct CST-to-Code Generation
    ↓
Output Code (Clean or Typed)
```

**Benefits of Unified Architecture:**
- ✅ Direct CST processing eliminates conversion bugs
- ✅ Single unified implementation reduces duplication
- ✅ Preserves exact source positioning and formatting
- ✅ Consistent type annotation handling
- ✅ Better performance with caching support
- ✅ Extensible transformation rule system

## Core Components

### 1. PerlCompiler (Unified Compiler)

**File:** `internal/compiler/perl_compiler.go`

The central component that replaces both legacy compilers:

```go
type PerlCompiler struct {
    target  Target                // TargetCleanPerl or TargetTypedPerl
    options CompilerOptions       // Compilation configuration
}
```

**Key Features:**
- **Target-Aware Compilation:** Single compiler handles both clean and typed output
- **CST-Based Processing:** Works directly with tree-sitter nodes
- **Automatic Version Pragmas:** Adds `use v5.36;` for clean Perl compatibility
- **Backward Compatibility:** Accepts both CST-based and legacy AST interfaces

**Example Usage:**
```go
// Create unified compiler for clean Perl
compiler := NewCleanPerlCompilerUnified()
result, err := compiler.CompileString(`my Int $count = 42;`)
// Result: "use v5.36;\nmy $count = 42;"

// Create unified compiler for typed Perl
typedCompiler := NewTypedPerlCompilerUnified()
result, err := typedCompiler.CompileString(`my Int $count = 42;`)
// Result: "my Int $count = 42;"
```

### 2. CST Analysis System

**File:** `internal/compiler/cst_analysis.go`

Provides utilities for understanding and navigating tree-sitter CST structure:

```go
type CSTAnalyzer struct {
    root    *sitter.Node
    content []byte
}
```

**Key Capabilities:**
- **Node Type Mapping:** Comprehensive mapping of tree-sitter node types
- **Type Annotation Extraction:** Finds and extracts type information
- **Variable Name Extraction:** Correctly identifies variable names
- **Pattern Recognition:** Identifies typed Perl constructs

**Node Types Supported:**
- `variable_declaration` - Variable declarations with optional types
- `type_assertion_expression` - Type assertions (`$value as Type`)
- `mandatory_parameter` - Method parameters with types
- `type_expression` - Type annotation containers

### 3. Tree Transformation Framework

**File:** `internal/compiler/transformation.go`

Provides declarative rules for transforming CST nodes:

```go
type TransformationRule interface {
    CanTransform(node *sitter.Node) bool
    Transform(node *sitter.Node, content []byte, transformer *CSTTransformer) (string, error)
    Description() string
}
```

**Built-in Transformation Rules:**
- `TypeExpressionRemovalRule` - Removes type expression nodes
- `VariableDeclarationCleanupRule` - Handles variable declarations
- `TypeAssertionCleanupRule` - Processes type assertions
- `MethodParameterCleanupRule` - Cleans method parameters
- `PreservationRule` - Preserves non-type syntax

**Transformation Examples:**

| Input (Typed Perl) | Clean Perl Output | Typed Perl Output |
|---|---|---|
| `my Int $count = 42;` | `my $count = 42;` | `my Int $count = 42;` |
| `$value as Str` | `$value` | `$value as Str` |
| `field ArrayRef[Int] $items;` | `field $items;` | `field ArrayRef[Int] $items;` |

### 4. Performance Optimizations

**Files:** `internal/compiler/optimization.go`, `internal/compiler/caching.go`

Comprehensive performance enhancements:

#### Compilation Caching
```go
type CompilationCache struct {
    cleanCache  map[string]*CachedResult
    typedCache  map[string]*CachedResult
    // ... thread-safe caching with LRU eviction
}
```

**Cache Features:**
- SHA256-based cache keys for content integrity
- Separate caches for clean and typed Perl
- LRU eviction when cache reaches capacity
- Thread-safe concurrent access
- Performance monitoring and statistics

#### Optimized CST Transformation
```go
type OptimizedCSTTransformer struct {
    *CSTTransformer
    nodeCache     map[*sitter.Node]string
    stringBuilder *strings.Builder
    // ... memory pooling and efficient string building
}
```

**Optimization Features:**
- Node result caching for repeated transformations
- Memory pooling for string builders and byte slices
- Efficient bounds checking for text extraction
- Reusable transformation instances

#### Performance Results

Based on benchmarks in `internal/compiler/benchmark_test.go`:

| Metric | Value | Improvement |
|---|---|---|
| Basic Compilation | ~258μs/op | Baseline |
| Cached Compilation | ~40μs/op | 84% faster |
| Memory Usage | ~11KB/op | Efficient |
| Cache Hit Ratio | 99-100% | Excellent |
| Parallel Performance | ~76μs/op | Thread-safe |

### 5. Compiler Registry Integration

**File:** `internal/compiler/compiler.go`

Updated registry system using unified compilers:

```go
func NewCompilerRegistry() *CompilerRegistry {
    registry := &CompilerRegistry{
        compilers: make(map[Target]Compiler),
    }

    // Register unified compilers (CST-based)
    registry.Register(NewCleanPerlCompilerUnified())
    registry.Register(NewTypedPerlCompilerUnified())

    return registry
}
```

**Registry Benefits:**
- Transparent migration to unified architecture
- Maintains existing API compatibility
- Supports both optimized and standard compilers
- Provides aggregated performance statistics

## Type Annotation Handling

### Supported Type Constructs

The unified compiler handles all typed Perl constructs:

#### Basic Types
```perl
my Int $number = 42;
my Str $text = "hello";
my Bool $flag = 1;
```

#### Complex Types
```perl
my ArrayRef[Int] $numbers = [1, 2, 3];
my HashRef[Str] $config = {key => "value"};
my Maybe[Object] $optional = undef;
```

#### Union and Intersection Types
```perl
my Union[Int, Str, Bool] $flexible = 42;
my Object&Serializable $obj = get_object();
my !Undef $required = get_value();
```

#### Type Assertions
```perl
my $value = get_data();
my $typed = $value as ArrayRef[Str];
```

#### Method Signatures
```perl
method Bool process(Int $a, Str $b) {
    return $a > length($b);
}
```

#### Field Declarations
```perl
field Int $count = 0;
field ArrayRef[Str] $items = [];
```

### Transformation Logic

The transformation system uses a rule-based approach:

1. **Identify Constructs:** Pattern matching identifies typed constructs
2. **Apply Rules:** Transformation rules process each construct
3. **Preserve Context:** Comments, whitespace, and formatting preserved
4. **Generate Output:** Clean or typed Perl based on target

## Error Handling and Diagnostics

### Error Types

The unified compiler provides structured error handling:

```go
type CompilerError struct {
    Code    string
    Message string
    Cause   error
}
```

**Error Categories:**
- `ErrInvalidAST` - AST validation failures
- `ErrCompilationFailed` - Compilation process errors
- `ErrUnsupportedTarget` - Unknown compilation targets

### Error Recovery

The compiler handles various error scenarios gracefully:

- **Syntax Errors:** Tree-sitter error recovery preserves partial structure
- **Invalid Type Annotations:** Graceful degradation to untyped constructs
- **Memory Issues:** Automatic cache cleanup and memory management
- **Concurrent Access:** Thread-safe error reporting

## Integration Points

### Parser Integration

The unified compiler integrates seamlessly with the parser:

```go
// Parse file
parser, _ := parser.NewParser()
ast, _ := parser.ParseFile("script.pl")

// Compile with unified compiler
registry := NewCompilerRegistry()
cleanCode, _ := registry.Compile(ast, TargetCleanPerl)
```

### PSC Command Integration

All PSC commands use the unified compiler:

- `psc strip` - Uses `TargetCleanPerl` for type annotation removal
- `psc run` - Compiles to clean Perl for execution
- `psc check` - Uses `TargetTypedPerl` for type preservation

### Backwards Compatibility

The unified compiler maintains compatibility with existing code:

- **AST Interface:** Accepts both new CST-based and legacy ASTs
- **Registry API:** Maintains existing `CompilerRegistry` interface
- **Output Format:** Produces equivalent output with improved quality

## Testing Strategy

### Test Coverage

Comprehensive test suite ensures reliability:

- **Unit Tests:** Individual component testing
- **Integration Tests:** End-to-end workflow validation
- **Performance Tests:** Benchmarking and regression detection
- **Stress Tests:** Large codebase and concurrent access testing

### Test Categories

1. **Functional Tests** (`*_test.go`)
   - Basic compilation correctness
   - Type transformation accuracy
   - Error handling validation

2. **Performance Tests** (`benchmark_test.go`)
   - Compilation speed benchmarks
   - Memory usage analysis
   - Cache effectiveness measurement

3. **Integration Tests** (`integration_*_test.go`)
   - Parser compatibility
   - PSC command integration
   - Real-world scenario testing

4. **Corpus Validation** (`corpus_validation_test.go`)
   - Large-scale test case validation
   - Regression prevention
   - Output quality assurance

## Future Enhancements

### Planned Improvements

1. **Additional Compilation Targets**
   - JavaScript transpilation
   - WebAssembly generation
   - Documentation extraction

2. **Enhanced Type System**
   - Generic type parameters
   - Type constraints and bounds
   - Advanced inference capabilities

3. **Performance Optimizations**
   - Parallel CST processing
   - Streaming compilation for large files
   - Advanced caching strategies

4. **Developer Tools**
   - Language server integration
   - IDE plugin support
   - Enhanced error diagnostics

### Extension Points

The architecture provides several extension points:

- **Custom Transformation Rules:** Add new type handling logic
- **Alternative Backends:** Support additional output formats
- **Performance Plugins:** Implement custom optimization strategies
- **Diagnostic Extensions:** Enhanced error reporting and analysis

## Migration Considerations

### Breaking Changes

The unified architecture introduces minimal breaking changes:

- **Legacy Compilers Deprecated:** `NewCleanPerlCompiler()` and `NewTypedPerlCompiler()` marked deprecated
- **Output Differences:** Clean Perl now includes `use v5.36;` pragma
- **Performance Characteristics:** Different memory and timing profiles

### Migration Path

Recommended migration steps:

1. **Update Compiler Creation:**
   ```go
   // Old
   compiler := NewCleanPerlCompiler()

   // New
   compiler := NewCleanPerlCompilerUnified()
   ```

2. **Verify Output Compatibility:**
   - Test compilation results with existing test suites
   - Validate generated Perl executes correctly

3. **Leverage Performance Features:**
   - Consider using caching compilers for improved performance
   - Monitor cache effectiveness with provided statistics

4. **Update Documentation:**
   - Reference new compiler architecture in project documentation
   - Update examples to use unified compiler APIs

## Conclusion

The unified compiler architecture represents a significant improvement in PVM's compilation capabilities. By eliminating the buggy CST-to-AST conversion layer and implementing a direct CST-based approach, the system now provides:

- **Higher Reliability:** No more variable name bugs or conversion issues
- **Better Performance:** Caching and optimization provide substantial speed improvements
- **Improved Maintainability:** Single unified implementation reduces code duplication
- **Enhanced Extensibility:** Rule-based transformation system supports future enhancements

The architecture successfully achieves all design goals while maintaining backward compatibility and providing a solid foundation for future development.
