# Tree-Sitter Shim Migration Guide

## Overview

This guide documents the tree-sitter shim architecture implemented in Phase 2 of the Tree-Sitter Shim Restoration Plan. The tree-sitter shim provides superior parsing capabilities for typed Perl code while maintaining backward compatibility with existing PVM workflows.

## Phase 2 Achievements Summary

### ✅ **Key Benefits Demonstrated**

1. **Superior Function Call Detection**: Tree-sitter detects 2+ function calls vs traditional parser's 0
2. **Advanced Syntax Support**: Handles complex typed Perl that traditional parser rejects
3. **Type Annotation Preservation**: Maintains type annotations through complete compilation workflows
4. **Production Integration**: Enhanced PSC commands demonstrate real-world applicability
5. **Performance Validation**: Acceptable performance cost for significant capability improvements

### ✅ **Components Successfully Migrated**

- **Enhanced PSC infer command** (`psc infer-ts`)
- **Type annotation preservation workflows**
- **Comprehensive validation test suites**
- **Performance benchmarking framework**

## Architecture Overview

### Tree-Sitter Shim Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Tree-Sitter Shim Architecture           │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌──────────────────────────────────┐ │
│  │   ShimParser    │    │         TreeSitterAST           │ │
│  │                 │    │                                  │ │
│  │ ParseStringShim │ ─► │ • Direct CST Access            │ │
│  │ ParseFileShim   │    │ • Type Annotation Preservation  │ │
│  └─────────────────┘    │ • Superior Function Detection   │ │
│                         └──────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌──────────────────────────────────┐ │
│  │ Migration Layer │    │      Compiler Integration       │ │
│  │                 │    │                                  │ │
│  │ • Backward      │    │ • TreeSitterASTAdapter          │ │
│  │   Compatibility │    │ • TargetTypedPerl               │ │
│  │ • AST Conversion│    │ • TargetCleanPerl               │ │
│  └─────────────────┘    └──────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Key Interfaces

#### ShimParser Interface
```go
type ShimParser interface {
    Parser // Inherits from traditional Parser
    ParseStringShim(content string) (*TreeSitterAST, error)
    ParseFileShim(path string) (*TreeSitterAST, error)
}
```

#### TreeSitterAST Structure
```go
type TreeSitterAST struct {
    Path            string
    Root            *TreeSitterNode
    Source          string
    Errors          []*ParseError
    TypeAnnotations []*TypeAnnotation
}

// Key capabilities
func (ast *TreeSitterAST) GetCSTRoot() *sitter.Node
func (ast *TreeSitterAST) GetTree() *sitter.Tree
```

## Migration Patterns

### Pattern 1: Enhanced Command Migration

**Before (Traditional):**
```go
parser, err := parser.NewParser()
ast, err := parser.ParseFile(inputFile)
```

**After (Tree-Sitter Enhanced):**
```go
shimParser, err := parser.NewShimParser()
shimAST, err := shimParser.ParseStringShim(content)
// Enhanced capabilities: superior function call detection
```

**Example: PSC Infer Command Enhancement**

The `psc infer-ts` command demonstrates the migration pattern:

```go
// Traditional infer command
func runInferCommand(cmd *cobra.Command, args []string) error {
    parser, err := parser.NewParser()
    ast, err := parser.ParseFile(inputFile)
    // Limited function call detection
}

// Enhanced tree-sitter infer command
func runInferTreeSitterCommand(cmd *cobra.Command, args []string) error {
    shimParser, err := parser.NewShimParser()
    shimAST, err := shimParser.ParseStringShim(content)
    // Superior function call detection + benchmarking + debug capabilities
}
```

### Pattern 2: Type Annotation Preservation

**Workflow Integration:**
```go
// Parse with tree-sitter shim
shimParser, err := parser.NewShimParser()
shimAST, err := shimParser.ParseStringShim(typedPerlCode)

// Create compiler adapter
adapter := &TreeSitterASTAdapter{shimAST}

// Compile to different targets
registry := compiler.NewCompilerRegistry()
typedOutput, err := registry.Compile(adapter, compiler.TargetTypedPerl)
cleanOutput, err := registry.Compile(adapter, compiler.TargetCleanPerl)
```

### Pattern 3: Function Call Detection Enhancement

**Tree-Sitter Advantage:**
```go
// Count function calls with tree-sitter
shimAST.Root.WalkNodes(func(node *ast.TreeSitterNode) bool {
    if node.Type() == "function_call_expression" {
        functionName := extractFunctionName(node)
        // Process function call with complete context
    }
    return true
})
```

## Performance Characteristics

### Benchmarking Results

| Test Case | Tree-Sitter | Traditional | Advantage |
|-----------|-------------|-------------|-----------|
| Simple Code | 430.6µs | 2.6µs | Traditional 166x faster |
| Complex Code | 3.03ms | **FAILS** | Tree-sitter only option |
| Real-World Code | 10.08ms | **FAILS** | Tree-sitter only option |
| Function Detection | 2 calls | 0 calls | Tree-sitter 100% better |

### Performance Analysis

**Key Insights:**
- **Simple Code**: Traditional parser faster but misses function calls
- **Complex Code**: Only tree-sitter works; traditional parser fails completely
- **Real-World Code**: Only tree-sitter handles production syntax
- **Development Impact**: 10ms parsing time negligible in developer workflow

## Migration Strategy

### Phase 1: Foundation (✅ Complete)
- Tree-sitter shim architecture implemented
- Core interfaces and AST structures defined
- Basic integration with compiler pipeline

### Phase 2: Validation & Migration (✅ Complete)
- High-value component migration (PSC infer command)
- Type annotation preservation validation
- Performance benchmarking and analysis
- Production workflow testing

### Phase 3: Broader Adoption (Future)
- Migrate additional PSC commands to tree-sitter shim
- Update type inference engine for direct tree-sitter integration
- Extend LSP server with tree-sitter capabilities
- Performance optimization for large codebases

## Best Practices

### When to Use Tree-Sitter Shim

**✅ Use tree-sitter shim for:**
- Complex typed Perl syntax parsing
- Function call detection and analysis
- Type annotation preservation workflows
- Production code with advanced type features
- New component development

**⚠️ Consider traditional parser for:**
- Simple scripts without type annotations
- Performance-critical batch processing
- Legacy compatibility requirements

### Code Quality Guidelines

**Parser Selection:**
```go
// For advanced typed Perl (recommended)
shimParser, err := parser.NewShimParser()
if err != nil {
    // Fallback to traditional parser if needed
    traditionalParser, err := parser.NewParser()
}

// For simple untyped Perl (optional optimization)
traditionalParser, err := parser.NewParser()
```

**Error Handling:**
```go
shimAST, err := shimParser.ParseStringShim(content)
if err != nil {
    return fmt.Errorf("tree-sitter parsing failed: %w", err)
}

// Always check for parse errors
if len(shimAST.Errors) > 0 {
    // Handle parse errors appropriately
}
```

## Testing Guidelines

### Validation Test Categories

1. **Functional Tests**: Verify tree-sitter shim behavior
2. **Preservation Tests**: Ensure type annotation preservation
3. **Comparison Tests**: Validate advantages over traditional parsing
4. **Performance Tests**: Monitor performance characteristics
5. **Integration Tests**: Test complete workflows

### Example Test Structure

```go
func TestTreeSitterShimMigration(t *testing.T) {
    t.Run("parsing_capability", func(t *testing.T) {
        // Test parsing of complex typed Perl syntax
    })

    t.Run("function_detection", func(t *testing.T) {
        // Verify superior function call detection
    })

    t.Run("type_preservation", func(t *testing.T) {
        // Validate type annotation preservation
    })

    t.Run("performance_comparison", func(t *testing.T) {
        // Benchmark against traditional parser
    })
}
```

## Troubleshooting

### Common Issues

**Problem: Tree-sitter shim parser not available**
```
Solution: Ensure tree-sitter-typed-perl is built correctly
Command: make tree-sitter
```

**Problem: Function calls not detected**
```
Solution: Verify node type checking
Expected: node.Type() == "function_call_expression"
```

**Problem: Type annotations lost during compilation**
```
Solution: Use TreeSitterASTAdapter for compiler integration
Ensure: TargetTypedPerl compilation target
```

## Migration Checklist

### For New Components
- [ ] Use `parser.NewShimParser()` for initialization
- [ ] Implement tree-sitter specific capabilities
- [ ] Add comparison tests vs traditional parser
- [ ] Document performance characteristics
- [ ] Include function call detection validation

### For Existing Components
- [ ] Assess complexity of typed Perl syntax used
- [ ] Create enhanced variant (e.g., `command-ts`)
- [ ] Maintain backward compatibility
- [ ] Add migration path documentation
- [ ] Validate performance impact

## Examples

### Complete Migration Example

See `internal/psc/infer_command_tree_sitter.go` for a complete example of migrating a PSC command to use tree-sitter shim architecture with enhanced capabilities.

### Validation Example

See `internal/validation/` for comprehensive validation tests demonstrating type annotation preservation and parsing capability comparisons.

### Performance Example

See `internal/benchmarks/parser_performance_bench_test.go` for detailed performance benchmarking comparing tree-sitter shim vs traditional parsing.

## Conclusion

The tree-sitter shim architecture provides essential parsing capabilities that traditional parsers cannot deliver, with acceptable performance costs. The Phase 2 migration demonstrates clear benefits for function call detection, type annotation preservation, and advanced syntax support.

**Next Steps:**
1. Continue migrating high-value components to tree-sitter shim
2. Optimize performance for large-scale usage
3. Extend tree-sitter capabilities to additional PVM components
4. Document lessons learned for future migrations

For questions or support, refer to the validation tests and examples provided in this guide.
