# Unified Compiler Architecture Implementation Plan

## Project Overview

This plan implements a fundamental architectural refactoring to replace the current separate `CleanPerlCompiler` and `TypedPerlCompiler` classes with a single unified compiler that works directly with tree-sitter's Concrete Syntax Tree (CST). The unified compiler uses tree transformation to optionally remove type nodes, eliminating the buggy CST-to-AST conversion layer.

**Key Architectural Insight:**
- Current approach: CST → AST conversion → separate compilers (buggy)
- Target approach: CST → unified compiler with tree transformation (clean)

**Core Problem Being Solved:**
The current `VarDecl.LogicalVariables()` bug (returning type names instead of variable names) is a symptom of unnecessary CST-to-AST conversion. By working directly with tree-sitter CST, we eliminate the conversion layer entirely and create a more maintainable architecture.

## Architecture Analysis

**Current State:**
- Tree-sitter produces CST (Concrete Syntax Tree)
- Buggy conversion layer creates `*ast.VarDecl` structures
- Separate `CleanPerlCompiler` and `TypedPerlCompiler` classes
- `CleanPerlCompiler` has AST traversal bugs causing output like `my $Int;;`
- Code duplication between compiler implementations

**Target State:**
- Single `PerlCompiler` works directly with tree-sitter CST
- Tree transformation step optionally removes type nodes for clean output
- No CST-to-AST conversion layer
- Conditional compilation based on target: typed or clean Perl
- Unified codebase eliminates duplication and bugs

## Implementation Steps

### Phase 1: CST Analysis and Foundation (Steps 1-3)

#### Step 1: Analyze Current CST Structure and Tree-sitter Integration

```
Analyze the current tree-sitter CST structure to understand what nodes we're working with and how type information is represented in the concrete syntax tree.

Investigation tasks:
- Map out tree-sitter node types for typed Perl constructs
- Document how type annotations appear in the CST (variable declarations, method signatures, etc.)
- Identify which CST nodes need transformation for clean Perl output
- Analyze the current CST-to-AST conversion to understand what's being lost/corrupted
- Document tree-sitter navigation patterns for common compilation scenarios

Create comprehensive CST analysis:
- Document all tree-sitter node types used in typed Perl
- Map type annotation nodes to their positions in declarations
- Identify transformation patterns (what to keep vs remove for clean output)
- Create CST navigation utilities for common operations
- Establish patterns for working directly with tree-sitter nodes

Key files to create:
- `internal/compiler/cst_analysis.go` - CST structure documentation and utilities
- `internal/compiler/cst_navigation.go` - Navigation helpers for tree-sitter nodes
- `internal/compiler/cst_analysis_test.go` - CST analysis validation tests
- `docs/cst_structure.md` - Documentation of CST patterns for typed Perl

Success criteria:
- Complete mapping of tree-sitter node types for typed Perl constructs
- Clear understanding of type annotation representation in CST
- Navigation utilities enable easy CST traversal
- Documentation provides foundation for unified compiler implementation
- Analysis reveals exact transformation patterns needed for clean output

Test-driven approach:
- Create test cases with various typed Perl constructs
- Parse with tree-sitter and analyze resulting CST structure
- Document node relationships and type annotation positions
- Test navigation utilities with real CST examples
- Validate understanding with corpus test cases
```

#### Step 2: Design Tree Transformation System

```
Design the tree transformation system that will convert typed Perl CST to clean Perl CST by removing type annotation nodes while preserving all other syntax.

Design tree transformation framework:
- Create transformation rule interface for different node types
- Design type node identification and removal strategies
- Plan preservation of comments, whitespace, and formatting
- Create transformation pipeline for processing entire CST
- Design validation system to ensure semantic equivalence

The transformation system should be declarative, making it easy to specify which nodes to remove and how to handle edge cases.

Key files to create:
- `internal/compiler/transformation.go` - Tree transformation framework
- `internal/compiler/rules.go` - Transformation rules for different node types
- `internal/compiler/preservation.go` - Comment and formatting preservation
- `internal/compiler/transformation_test.go` - Transformation framework tests

Success criteria:
- Transformation rules can identify and remove type annotation nodes
- Preservation system maintains formatting and comments
- Pipeline processes entire CST systematically
- Validation ensures semantic equivalence between typed and clean output
- Framework is extensible for future transformation needs

Test-driven approach:
- Create test cases with type annotations to be removed
- Design transformation rules incrementally
- Test preservation of formatting and comments
- Validate semantic equivalence of transformed output
- Add edge case testing for complex type constructs
```

#### Step 3: Implement CST-to-Code Generation

```
Implement the core CST-to-code generation system that converts tree-sitter CST directly to Perl source code, bypassing the AST conversion layer entirely.

Create CST code generation:
- Implement CST node visitor pattern for code generation
- Create text reconstruction from CST preserving original formatting
- Handle whitespace, comments, and source positioning accurately
- Support for both typed and clean Perl output modes
- Implement efficient string building and memory management

The code generator should produce output that exactly matches the original source (for typed mode) or properly cleaned source (for clean mode).

Key files to create:
- `internal/compiler/cst_generator.go` - CST-to-code generation system
- `internal/compiler/text_reconstruction.go` - Text reconstruction with formatting
- `internal/compiler/visitor.go` - CST visitor pattern implementation
- `internal/compiler/cst_generator_test.go` - Code generation tests

Success criteria:
- CST visitor can traverse all tree-sitter node types systematically
- Text reconstruction preserves original formatting exactly
- Code generation handles both typed and clean modes correctly
- Memory usage is efficient for large CST structures
- Generated code maintains semantic and syntactic correctness

Test-driven approach:
- Test text reconstruction with various formatting scenarios
- Validate visitor pattern with complex CST structures
- Test both typed and clean output modes
- Verify memory efficiency with large test cases
- Add comprehensive formatting preservation tests
```

### Phase 2: Unified Compiler Implementation (Steps 4-6)

#### Step 4: Create Unified PerlCompiler Class

```
Implement the unified PerlCompiler class that replaces both CleanPerlCompiler and TypedPerlCompiler with a single implementation using tree transformation.

Create unified compiler:
- Implement `PerlCompiler` struct with target-aware compilation
- Integrate CST code generation and tree transformation systems
- Support both `TargetCleanPerl` and `TargetTypedPerl` targets
- Implement compiler options for controlling transformation behavior
- Add proper error handling and validation throughout the pipeline

The unified compiler should provide the same interface as existing compilers while internally using the superior CST-based approach.

Key files to create:
- `internal/compiler/perl_compiler.go` - Unified PerlCompiler implementation
- Update `internal/compiler/types.go` - Unified compiler interface
- `internal/compiler/perl_compiler_test.go` - Comprehensive unified compiler tests

Success criteria:
- Single compiler supports both clean and typed Perl targets
- Tree transformation correctly removes type annotations for clean output
- Typed output preserves all type information exactly
- Compiler options provide flexible control over output
- Interface compatibility maintained with existing compiler registry

Test-driven approach:
- Test compilation of same input to both targets
- Validate tree transformation removes only type annotations
- Test preservation of all non-type syntax elements
- Verify interface compatibility with existing systems
- Add comprehensive edge case testing
```

#### Step 5: Update Compiler Registry and Integration

```
Update the compiler registry system to use the unified PerlCompiler and remove the legacy separate compiler classes.

Update compiler integration:
- Modify `CompilerRegistry` to use unified `PerlCompiler` for both targets
- Remove legacy `CleanPerlCompiler` and `TypedPerlCompiler` classes
- Update all compiler instantiation points throughout codebase
- Ensure PSC commands work with unified compiler
- Add migration validation to prevent regressions

The registry should transparently use the unified compiler while maintaining all existing functionality and interfaces.

Key files to modify:
- `internal/compiler/registry.go` - Update to use unified compiler
- Remove `internal/compiler/clean_perl.go` - Legacy clean compiler
- Remove `internal/compiler/typed_perl.go` - Legacy typed compiler
- Update `internal/psc/*.go` - PSC command integration points

Success criteria:
- Compiler registry seamlessly uses unified compiler
- All PSC commands work with new compiler architecture
- Legacy compiler classes completely removed
- No functionality regression from unified approach
- Performance equivalent or better than separate compilers

Test-driven approach:
- Test registry with both compilation targets
- Validate PSC command integration
- Test removal of legacy classes doesn't break anything
- Verify performance with benchmarks
- Add migration validation tests
```

#### Step 6: Fix Corpus Validation and Test Integration

```
Update the corpus validation tests to work with the unified compiler and validate that the VarDecl.LogicalVariables() bug is eliminated.

Fix corpus validation:
- Update corpus validation tests to use unified compiler
- Verify that `my Int $count = 42;` compiles to `my $count = 42;` correctly
- Validate all corpus test cases pass with unified compiler
- Fix any corpus expectations that were based on buggy behavior
- Add comprehensive regression testing for the original bug

The corpus validation should prove that the unified compiler eliminates the variable name bug and produces correct output.

Key files to modify:
- `internal/parser/corpus_validation_test.go` - Update for unified compiler
- Update corpus test expectations if needed based on correct behavior
- Add specific regression tests for the original bug

Success criteria:
- All corpus validation tests pass with unified compiler
- Basic typed variables test produces correct clean Perl output
- Variable names appear correctly in generated code (not type names)
- No regression in any existing corpus test cases
- New architecture eliminates the LogicalVariables() bug entirely

Test-driven approach:
- Run corpus validation with unified compiler
- Specifically test the failing case: `my Int $count = 42;`
- Validate correct output: `my $count = 42;`
- Test all corpus cases for regression
- Add permanent regression tests for the original bug
```

### Phase 3: Validation and Migration (Steps 7-9)

#### Step 7: Performance and Memory Optimization

```
Optimize the unified compiler for performance and memory usage, ensuring it meets or exceeds the performance of the legacy separate compilers.

Implement performance optimizations:
- Profile CST traversal and identify bottlenecks
- Optimize string building and memory allocation patterns
- Add efficient caching for repeated compilation operations
- Implement parallel processing for large codebases where appropriate
- Add performance benchmarks and regression testing

The optimized compiler should handle large Perl codebases efficiently while maintaining correctness.

Key files to create/modify:
- `internal/compiler/optimization.go` - Performance optimization utilities
- `internal/compiler/caching.go` - Compilation result caching
- `internal/compiler/benchmark_test.go` - Performance benchmarks

Success criteria:
- Compilation performance meets or exceeds legacy compilers
- Memory usage is efficient for large CST structures
- Caching provides meaningful performance improvements
- Benchmarks demonstrate acceptable performance characteristics
- No performance regressions from unified approach

Test-driven approach:
- Create performance benchmarks for various codebase sizes
- Profile and optimize critical path operations
- Test memory usage patterns with large inputs
- Validate caching effectiveness with repeated operations
- Add performance regression testing
```

#### Step 8: Comprehensive Integration Testing

```
Perform comprehensive end-to-end integration testing to ensure the unified compiler works correctly with all existing PVM/PSC functionality.

Comprehensive integration testing:
- Test all PSC commands with unified compiler
- Validate integration with parser and existing tools
- Test edge cases and error handling scenarios
- Verify backward compatibility with existing workflows
- Add stress testing with large, complex Perl codebases

The integration testing should prove the unified compiler is a drop-in replacement for the legacy architecture.

Key files to create:
- `internal/compiler/integration_test.go` - Comprehensive integration tests
- Add edge case tests throughout codebase
- Update existing integration tests for unified architecture

Success criteria:
- All existing PSC functionality works with unified compiler
- Parser integration maintains all existing capabilities
- Error handling provides clear, actionable messages
- Complex Perl constructs compile correctly
- Stress testing validates robustness

Test-driven approach:
- Test every PSC command with unified compiler
- Add comprehensive edge case testing
- Test error scenarios and recovery
- Validate with real-world Perl codebases
- Add stress tests for performance validation
```

#### Step 9: Documentation and Migration Guide

```
Create comprehensive documentation for the unified compiler architecture and provide migration guidance for any external users.

Create documentation:
- Document the unified compiler architecture and design decisions
- Create troubleshooting guide for common issues
- Document performance characteristics and optimization features
- Provide migration guide for any breaking changes
- Update all existing documentation to reflect new architecture

The documentation should help users understand the benefits of the unified approach and how to use it effectively.

Key files to create/update:
- `docs/compiler_architecture.md` - Unified compiler design documentation
- `docs/migration_guide.md` - Migration guidance
- Update existing docs throughout codebase
- Add inline documentation to new code

Success criteria:
- Architecture documentation clearly explains design decisions
- Troubleshooting guide helps users resolve common issues
- Migration guide provides clear transition path
- All documentation is accurate and up-to-date
- Code documentation supports maintainability

Test-driven approach:
- Validate documentation accuracy against implementation
- Test troubleshooting guide with real scenarios
- Verify migration guide with actual migration scenarios
- Review documentation for completeness and clarity
- Add documentation validation to CI pipeline
```

## Prompt Structure

Each step above provides a complete prompt that:
1. Clearly states the objective and builds incrementally on CST-first architecture
2. Eliminates the buggy CST-to-AST conversion layer progressively
3. Maintains backward compatibility throughout the migration
4. Includes specific implementation guidance for tree-sitter integration
5. Defines measurable success criteria for each architectural change
6. Emphasizes comprehensive testing to prevent regressions
7. Provides validation strategies for the unified approach

The prompts are designed for execution in sequence, with each step building toward the unified compiler architecture while maintaining system stability.

## Success Metrics

**Technical Success:**
- VarDecl.LogicalVariables() bug completely eliminated
- Single unified compiler replaces separate implementations
- CST-to-AST conversion layer removed entirely
- Performance equivalent or better than legacy compilers
- All corpus validation tests pass with correct output

**Architectural Success:**
- Clean separation between tree transformation and code generation
- Unified codebase eliminates duplication between compilers
- Tree-sitter CST used directly without lossy conversion
- Extensible architecture supports future enhancements
- Maintainable design with clear responsibilities

**User Experience Success:**
- No functionality regression from architectural change
- All PSC commands work transparently with unified compiler
- Better error messages and debugging information
- Improved performance for large codebases
- Foundation for future compiler enhancements

## Architecture Integration

The unified compiler integrates with existing PVM architecture:
- **Parser Integration**: Works directly with tree-sitter CST output
- **PSC Commands**: Drop-in replacement for existing compiler usage
- **Test Framework**: Enhanced corpus validation with correct behavior
- **Performance**: Optimized CST processing with caching
- **Extensibility**: Foundation for future compilation targets

This architectural change eliminates a major source of bugs while creating a more maintainable and extensible foundation for future compiler development.
