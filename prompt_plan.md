# PVM Ecosystem Build Plan - Test-Driven Development to 100% Parser Completion

## Project Overview

The PVM Ecosystem is a comprehensive Perl development toolchain featuring:
- **PVM (Perl Version Manager)** - Manages Perl installations and versions
- **PSC (Perl Script Compiler)** - Static type checker with enhanced LSP features
- **PVI (Perl Version Installer)** - Module installer with type-aware dependency management
- **PVX (Perl Version eXecutor)** - Isolated script execution with dependency resolution

**Current Status**: 94.1% test pass rate (59/899 parser failures remaining)
**Goal**: Achieve 100% parser test completion through targeted, test-driven improvements

## Architecture Foundation

### Modernized Compiler Pipeline
```
Source → Scanner → Parser → Binder → Checker → Compiler → Output
```

Key components already implemented:
- `internal/scanner/` - Lexical analysis and tokenization ✅
- `internal/parser/` - AST generation with tree-sitter integration ✅ (needs completion)
- `internal/ast/` - Consolidated AST node types and navigation ✅
- `internal/binder/` - Symbol resolution and scope management ✅
- `internal/typechecker/` - Type analysis using symbol information ✅
- `internal/compiler/` - Code generation to multiple targets ✅

### Tree-sitter Integration
- Custom `tree-sitter-typed-perl` grammar extending standard Perl
- Supports typed variable declarations, union types, parameterized types
- Self-contained build with Go and tree-sitter CLI dependencies

---

## Phase 1: Parser Completion to 100% (Critical Path)

**Current Reality**: 59 parser failures identified, concentrated in 6 specific patterns
**Target**: 0 parser failures (100% pass rate)
**Timeline**: 8 focused steps

### Step 1: Complex Method Signature Parsing

```
Fix the core issue affecting multiple test categories: complex method signatures with advanced type expressions.

CONTEXT: Most remaining parser failures (4 out of 6) are related to method signature parsing with complex types. The parser handles basic method signatures but fails on nested parameterized types, union types in parameters, and complex return types.

PROBLEM: Tests like `complex-types_complex_method_signatures` and `parameterized-types_method_signatures` fail with "UnknownTypeError: Syntax error in type expression" when parsing method signatures containing:
- Complex parameter types: `ArrayRef[HashRef[Int|Str]] $data`  
- Complex return types: `-> Result[Map[Str, Array[Item]], ProcessingError]`
- Union types in method signatures: `method func(Int|Str $param) -> Bool|Error`

TASK: Enhance method signature parsing in the tree-sitter grammar and parser logic.

TARGET FAILURES:
- complex-types_complex_method_signatures
- parameterized-types_method_signatures
- union-types_method_signatures_unions (partially)

IMPLEMENTATION:
1. **Analyze failing patterns**: Read test files to identify exact syntax causing failures
2. **Enhance grammar**: Extend `tree-sitter-typed-perl/grammar.js` method signature rules
3. **Improve parser logic**: Update `internal/parser/treesitter/perl.go` type expression handling
4. **Test incrementally**: Write focused tests for each pattern before implementing

TECHNICAL REQUIREMENTS:
- Support deeply nested type expressions in method parameters
- Handle union types within method signatures: `(Int|Str) $param`
- Parse complex return type expressions: `-> ArrayRef[HashRef[CustomType]]`
- Maintain backward compatibility with existing simple method signatures
- Generate proper AST nodes for complex method signatures

VALIDATION:
- Run targeted tests: `go test -v ./internal/parser -run "complex.*method|parameterized.*method|union.*method"`
- Verify no regressions in existing method signature parsing
- Confirm AST contains complete type information for method signatures
- Performance test with deeply nested method signature expressions

EXPECTED OUTCOME:
- 3 major failing test categories converted to passing
- Foundation for handling any complex type expression in method context
- Improved production readiness for typed method signatures
```

### Step 2: Union Types in Nested Contexts

```
Fix union type parsing when used in nested expression contexts beyond simple variable declarations.

CONTEXT: Basic union types (Int|Str) work in variable declarations, but fail when used in nested contexts like method parameters, parameterized types, and complex expressions.

PROBLEM: Test `union-types_nested_contexts` fails because union types can't be parsed correctly when they appear in complex nested expressions like:
- `ArrayRef[Int|Str]` (union inside parameterized type)
- `method func((Int|Str) $param)` (parenthesized union in method)
- `$var as (Success|Error)` (union in type assertion)

TASK: Enhance union type parsing to handle all nested expression contexts.

TARGET FAILURES:
- union-types_nested_contexts
- Remaining union-types_method_signatures_unions issues

IMPLEMENTATION:
1. **Study nested union patterns**: Examine failing test cases for specific syntax patterns
2. **Grammar precedence**: Fix operator precedence for union vs parameterized types
3. **Parenthesized unions**: Add support for `(Type1|Type2)` expressions
4. **Context handling**: Ensure union types work in all expression contexts

TECHNICAL REQUIREMENTS:
- Parse `ArrayRef[Int|Str]` as parameterized type with union parameter
- Support parenthesized union expressions: `(Type1|Type2)`
- Handle union types in type assertions: `$value as (Int|Str)`
- Maintain proper precedence: union vs intersection vs parameterized types
- Generate accurate AST for nested union type information

VALIDATION:
- Test union types in various nested contexts
- Verify parameterized types with union parameters work
- Ensure parenthesized unions parse correctly
- Confirm AST preserves union type structure in nested contexts

EXPECTED OUTCOME:
- Union types work in all expression contexts
- 1-2 remaining union type failures eliminated
- Robust foundation for complex type expressions
```

### Step 3: Complex Type Assertions

```
Implement support for type assertion expressions with complex type patterns.

CONTEXT: Type assertions ($value as Type) work for simple types but fail when the type expression involves complex nested types, unions, or parameterized types.

PROBLEM: Test `complex-types_complex_type_assertions` fails when type assertions use complex type expressions like:
- `$data as ArrayRef[HashRef[Int|Str]]`
- `$result as (Success|Error)&Detailed`
- `$object as MyClass[T]`

TASK: Enhance type assertion parsing to support any valid type expression.

TARGET FAILURES:
- complex-types_complex_type_assertions

IMPLEMENTATION:
1. **Analyze assertion patterns**: Identify specific type assertion syntax causing failures
2. **Extend assertion grammar**: Allow any valid type expression after `as` keyword
3. **Integration testing**: Ensure type assertions work with Step 1 and 2 improvements
4. **AST generation**: Proper type assertion nodes for type checker integration

TECHNICAL REQUIREMENTS:
- Support type assertions with parameterized types: `$var as ArrayRef[Int]`
- Handle union types in assertions: `$var as (Int|Str)`
- Parse intersection types: `$var as (Type1&Type2)`
- Work with nested expressions: `$data->method() as ProcessedType`
- Generate AST nodes compatible with type checker

VALIDATION:
- Test type assertions with all complex type patterns
- Verify assertions work in expression contexts
- Ensure AST contains correct type assertion information
- No regressions in simple type assertions

EXPECTED OUTCOME:
- Complete type assertion support for production code
- 1 major failure category eliminated
- Enhanced type safety capabilities for PVM users
```

### Step 4: Generic Class Declarations

```
Fix the final major category: generic class declarations with type parameters.

CONTEXT: Basic class declarations work, but generic classes with type parameters fail. This is the last major category of failures.

PROBLEM: Test `classes-roles_generic_class_declarations` fails because the parser doesn't properly handle:
- `class Container[T] { ... }`
- `class HashMap[K, V] { ... }`
- `class Cache[T: Serializable] { ... }` (with constraints)

TASK: Implement complete generic class declaration support.

TARGET FAILURES:
- classes-roles_generic_class_declarations

IMPLEMENTATION:
1. **Study generic class syntax**: Examine the failing test patterns
2. **Grammar extension**: Add type parameter list support to class declarations
3. **Constraint syntax**: Support type parameter constraints if needed
4. **Method integration**: Ensure generic methods work within generic classes

TECHNICAL REQUIREMENTS:
- Parse `class Name[T] { ... }` syntax
- Support multiple type parameters: `[T, U, V]`
- Handle type constraints if present: `[T: Constraint]`
- Generate proper AST for generic class information
- Ensure compatibility with existing class declaration parsing

VALIDATION:
- Test various generic class declaration patterns
- Verify type parameter information preserved in AST
- Ensure generic methods work within generic classes
- No regressions in basic class declarations

EXPECTED OUTCOME:
- Complete object-oriented programming support
- Final major failure category eliminated
- Production-ready generic class support
```

### Step 5: Remaining Failure Analysis and Resolution

```
Identify and systematically eliminate any remaining parser failures through targeted analysis.

CONTEXT: After Steps 1-4, we should have eliminated the 6 known major failure patterns. This step addresses any remaining edge cases or newly discovered issues.

PROBLEM: There may be additional parser failures not covered by the major categories, or edge cases within the categories that need specific attention.

TASK: Complete systematic elimination of all remaining parser failures.

IMPLEMENTATION:
1. **Current status check**: Run full parser test suite to get updated failure count
2. **Failure categorization**: Group any remaining failures by pattern/type
3. **Targeted fixes**: Address each remaining failure with minimal, focused changes
4. **Edge case handling**: Fix unusual but valid Perl syntax patterns
5. **Grammar completion**: Final grammar extensions for edge cases

TECHNICAL APPROACH:
- Run: `go test -v ./internal/parser -count=1` for complete status
- For each remaining failure:
  - Examine the specific test case and expected vs actual behavior
  - Determine if it's a grammar, parsing logic, or test expectation issue
  - Implement minimal fix targeting that specific pattern
  - Validate fix doesn't break existing functionality

VALIDATION:
- After each fix: full parser test suite to ensure no regressions
- Confirm failure count decreases with each change
- Final validation: 0 parser failures, 100% pass rate

EXPECTED OUTCOME:
- All remaining parser failures eliminated
- 100% parser test pass rate achieved
- Complete typed Perl language support
```

### Step 6: Performance and Integration Validation

```
Validate that 100% parser completion doesn't introduce performance regressions and integrates properly with the overall system.

CONTEXT: After achieving 100% parser test completion, we need to ensure the improvements don't negatively impact performance or break integration with other components.

TASK: Comprehensive validation of parser improvements across the entire PVM ecosystem.

IMPLEMENTATION:
1. **Performance benchmarking**: Compare parsing performance before/after improvements
2. **Memory usage analysis**: Ensure complex type parsing doesn't cause memory issues
3. **Integration testing**: Verify parser improvements work with PSC, PVI, PVX
4. **Real-world testing**: Test against actual typed Perl codebases
5. **Regression testing**: Full system test suite to catch any regressions

TECHNICAL VALIDATION:
- Performance benchmarks: complex type parsing should remain under production thresholds
- Memory profiling: no memory leaks or excessive allocation in complex type handling
- Cross-component testing: PSC type checking, PVI dependency analysis, PVX execution
- End-to-end workflows: complete development workflows with enhanced parser

VALIDATION COMMANDS:
- `make test` - Full system test suite
- Performance profiling with complex typed Perl files
- Integration testing with real-world typed Perl projects
- Memory usage monitoring during complex type parsing

EXPECTED OUTCOME:
- 100% parser completion maintained
- No performance regressions introduced
- All PVM ecosystem components work with enhanced parser
- System ready for production use with complete typed Perl support
```

---

## Phase 2: Post-Parser System Enhancement

### Step 7: Enhanced LSP Integration

```
Leverage the 100% complete parser to enhance Language Server Protocol features.

CONTEXT: With complete parser functionality, we can now provide TypeScript-quality LSP features including accurate symbol navigation, type-aware autocompletion, and precise error reporting.

TASK: Implement advanced LSP features using the enhanced parser capabilities.

IMPLEMENTATION:
1. **Symbol-aware navigation**: Implement goto definition, find references using complete AST
2. **Type-aware autocompletion**: Provide intelligent code completion based on type information
3. **Advanced diagnostics**: Enhanced error messages with type context and suggestions
4. **Real-time type checking**: Integration with enhanced parser for live error detection

TECHNICAL REQUIREMENTS:
- Utilize complete AST information for accurate symbol resolution
- Implement caching for performance with large codebases
- Provide contextual error messages and quick fixes
- Support for complex type expressions in LSP features

EXPECTED OUTCOME:
- TypeScript-quality development experience for Perl
- Production-ready LSP server with advanced features
- Enhanced developer productivity with intelligent tooling
```

### Step 8: Advanced Type System Features

```
Build upon the complete parser to implement advanced type system features and optimizations.

CONTEXT: With 100% parser completion, we can now implement advanced type system features that were previously impossible due to parsing limitations.

TASK: Implement advanced type checking, inference, and optimization features.

IMPLEMENTATION:
1. **Flow-sensitive analysis**: Advanced type checking with control flow analysis
2. **Type inference engine**: Intelligent type inference for untyped code
3. **Generic type system**: Complete support for generic types and constraints
4. **Performance optimizations**: Optimize type checking for large codebases

TECHNICAL REQUIREMENTS:
- Leverage complete AST for accurate type analysis
- Implement incremental type checking for performance
- Support for complex type relationships and constraints
- Integration with existing typechecker infrastructure

EXPECTED OUTCOME:
- Production-ready advanced type system
- Intelligent type inference capabilities
- Performance optimized for large codebases
- Foundation for future type system enhancements
```

---

## Implementation Guidelines

### Test-Driven Development Requirements

Every step MUST follow strict TDD practices:
1. **Write failing tests first** - Create specific tests for the functionality being implemented
2. **Implement minimal solution** - Write just enough code to make the test pass
3. **Refactor safely** - Improve code while maintaining all tests passing
4. **Validate completely** - Run full test suite to ensure no regressions

### Parser-Specific TDD Workflow

For parser improvements:
1. **Identify failing pattern** - Examine specific test case causing failure
2. **Create focused test** - Write minimal test reproducing the issue
3. **Grammar first** - Update tree-sitter grammar if needed
4. **Parser logic** - Enhance parsing logic to handle new patterns
5. **AST generation** - Ensure proper AST nodes are created
6. **Integration test** - Verify improvement works in context

### Build and Test Commands

**Critical**: All tests MUST pass 100% before committing any changes.

```bash
# Build all components
make

# Test complete system
make test

# Test parser specifically
go test ./internal/parser

# Build tree-sitter after grammar changes
make tree-sitter

# Performance testing
make benchmark
```

### Risk Mitigation

1. **Grammar Changes**: Always backup grammar before changes, test extensively
2. **Parser Logic**: Incremental changes with validation at each step
3. **Performance**: Benchmark before/after each major change
4. **Integration**: Cross-component testing after parser improvements

### Success Metrics

**Phase 1 Success Criteria:**
- Parser test suite: 0/899 failures (100% pass rate)
- No performance regressions in parser performance
- All existing functionality preserved
- Complete typed Perl language support achieved

**Overall Success Criteria:**
- System test suite: >95% pass rate maintained
- Performance benchmarks meet production requirements
- Real-world typed Perl codebases parse and process correctly
- Enhanced LSP features provide TypeScript-quality experience

### Completion Validation

Final validation checklist:
- [ ] Parser test suite: 100% pass rate
- [ ] Full system test suite: >95% pass rate
- [ ] Performance benchmarks within acceptable ranges
- [ ] Integration tests with all PVM components passing
- [ ] Real-world typed Perl code parsing successfully
- [ ] Enhanced LSP features working with complete parser
- [ ] Documentation updated to reflect completed features

This plan provides a systematic, test-driven approach to completing the PVM ecosystem with focus on achieving 100% parser completion as the critical foundation for all advanced features.