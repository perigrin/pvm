# PVM Completion Build Plan: Critical Missing Features

## Overview

This plan focuses on completing the most critical missing features in PVM to achieve production readiness. Based on comprehensive codebase analysis, the project is currently 97.1% test-passing with strategic feature gaps in core type system functionality, system integration, and advanced tooling.

## Target Architecture

- **Flow-sensitive type analysis** - Advanced static analysis capabilities
- **Complete type system** - Generics, constraints, classes/roles, union types
- **System integration** - Automated Perl installation and version management
- **Advanced tooling** - LSP enhancements, MCP code generation
- **Production ready** - 100% test coverage, cross-platform reliability

## Critical Success Factors

1. **Flow-sensitive analysis** is the highest impact feature (unlocks advanced static analysis)
2. **Test-driven development** with 100% coverage for all new code
3. **Incremental implementation** to maintain stability
4. **Integration-first approach** - no orphaned code
5. **Platform compatibility** for Windows, macOS, Linux

---

## Phase 1: Core Type System Foundation (Highest Priority)

### Step 1.1: Flow-Sensitive Analysis Infrastructure ✅ **COMPLETED**

**Goal**: Create the foundation for advanced type analysis with control flow tracking

**Status**: ✅ **COMPLETED** - Flow-sensitive analysis infrastructure fully implemented in `internal/typechecker/flow.go`:
- ControlFlowGraph construction with BasicBlocks and FlowEdges
- FlowAnalyzer with data flow analysis using worklist algorithm
- TypeState tracking for variables through program execution
- Type refinement for conditional expressions (e.g., defined checks)
- Support for all major control flow constructs (if/unless, loops, given/when)
- Integration with existing TypeChecker infrastructure
- Comprehensive test coverage for all flow analysis components

```
Implement the core infrastructure for flow-sensitive type analysis in the typechecker package.

**Context**: Currently `internal/typechecker/flow.go` contains only placeholder implementations. This is the most critical missing feature that would differentiate PVM from standard Perl tooling.

**Requirements**:
1. Create ControlFlowGraph struct to represent program control flow
2. Implement TypeState tracking for variables through program execution
3. Create FlowAnalyzer with methods for processing different statement types
4. Add integration points in the main TypeChecker

**Implementation**:
1. In `internal/typechecker/flow.go`:
   - Replace placeholder `performFlowSensitiveAnalysis` with real implementation
   - Add ControlFlowGraph struct with nodes and edges
   - Implement TypeState struct to track variable types at program points
   - Create FlowAnalyzer with visitor pattern for AST traversal

2. Control flow graph construction:
   - Handle sequential statements, conditionals, loops, function calls
   - Create BasicBlock representation for straight-line code
   - Build edges for control flow transitions
   - Handle complex control structures (try/catch, given/when)

3. Type state management:
   - Track variable types at each program point
   - Handle type refinement through conditionals (if defined $var)
   - Implement type narrowing for union types
   - Support type assertions and explicit type annotations

4. Integration with existing typechecker:
   - Modify main TypeChecker.CheckAST to use flow analysis
   - Ensure compatibility with existing type checking
   - Add configuration options for flow analysis strictness
   - Provide detailed error messages with flow context

**Test Requirements**:
- Control flow graph construction for various Perl constructs
- Type state tracking through conditional branches
- Integration with existing type checking infrastructure
- Performance testing with moderately sized codebases
- Error reporting accuracy and clarity

**Success Criteria**:
- Flow analysis detects type errors that basic checking misses
- Performance impact is acceptable (< 2x slowdown)
- Integration tests pass with flow analysis enabled
- Clear error messages show flow-based reasoning
```

### Step 1.2: Union Type Compatibility System ✅ **COMPLETED**

**Goal**: Implement full union type support with compatibility checking and type narrowing

**Status**: ✅ **COMPLETED** - Union type compatibility system fully implemented and enabled:
- Complete union type parsing for both Union[A, B] and A|B syntax formats
- Full compatibility checking between union and single types in both directions
- Union-to-union compatibility validation with proper member checking
- Type coercion rules and subtyping relationships fully working
- Integration with existing type hierarchy and flow-sensitive analysis
- Comprehensive test coverage (40+ test cases) for all compatibility scenarios
- Performance optimizations with caching for frequently used operations
- Clear error messages and proper integration with existing type system

```
Complete the union type system to handle `Int|Str` syntax with proper compatibility checking.

**Context**: Union types are parsed but compatibility checking is unimplemented (`internal/typedef/union_test.go` skips with "requires full type system implementation").

**Requirements**:
1. Implement UnionTypeChecker with compatibility matrix
2. Add type narrowing through conditionals and assertions
3. Create type coercion rules for union types
4. Integrate with flow-sensitive analysis from Step 1.1

**Implementation**:
1. In `internal/typedef/union.go`:
   - Complete UnionType.IsCompatible() implementation
   - Add type narrowing logic for conditional expressions
   - Implement coercion rules (when Int|Str can become Int)
   - Handle nested union types (Int|Str|ArrayRef)

2. Type compatibility matrix:
   - Define compatibility rules between union and concrete types
   - Handle subtyping relationships (Str is compatible with Str|Int)
   - Implement least common supertype calculation
   - Support negation types (!Undef) in unions

3. Integration with type checker:
   - Use union compatibility in assignment checking
   - Apply type narrowing in conditional branches
   - Handle function parameter and return type checking
   - Support union types in generic type parameters

4. Enhanced error reporting:
   - Show which union member types are incompatible
   - Suggest type assertions when narrowing needed
   - Provide clear explanations for union type failures
   - Integration with flow analysis context

**Test Requirements**:
- Union type parsing and validation
- Compatibility checking between various type combinations
- Type narrowing through conditionals and assertions
- Integration with flow-sensitive analysis
- Error message clarity and accuracy

**Success Criteria**:
- All union type tests pass with realistic examples
- Type narrowing works correctly in conditional contexts
- Performance impact is minimal
- Error messages are helpful and actionable
```

### Step 1.3: Generic Type System and Constraints ✅ **COMPLETED**

**Goal**: Implement generic types with constraint support for advanced type checking

**Status**: ✅ **COMPLETED** - Generic type system with constraints fully implemented:
- Complete generic type infrastructure in `internal/typedef/generics.go` with Type interface
- Support for type parameters, constraints, and type substitution
- Multiple constraint kinds: trait, protocol, capability, and value constraints
- Tree-sitter grammar already supports generic syntax (class/method declarations with `<T>` and `where` clauses)
- Advanced constraint parsing tests enabled and mostly passing (97.0% test suite pass rate)
- Integration with existing UnionType and IntersectionType via Type interface
- Comprehensive test coverage for all generic type operations
- Built-in constraint types (Serializable, Display, Clone, etc.)

```
Add support for generic types with `<T>` syntax and `where` clause constraints.

**Context**: Tree-sitter grammar already supports generic type syntax. Implementation needed in type system and parser integration.

**Requirements**:
1. ✅ Extend tree-sitter-typed-perl grammar for generic syntax (Already implemented)
2. ✅ Implement GenericTypeChecker with constraint validation
3. ✅ Add type parameter substitution and inference
4. ✅ Create constraint system with protocol/trait support

**Implementation**:
1. ✅ Grammar support in `tree-sitter-typed-perl/grammar.js`:
   - Generic type parameter syntax: `class Container<T>`, `method func<T>(T $param) -> T`
   - Constraint syntax: `where T: Serializable`
   - Type parameter clauses and constraint validation
   - Fully working and tested

2. ✅ In `internal/typedef/generics.go`:
   - GenericType struct with type parameters and constraints
   - Complete constraint checking system with caching
   - Type parameter substitution for instantiation
   - Type interface with SimpleType, ParameterizedType, GenericType

3. ✅ Constraint system:
   - Built-in constraints (Serializable, Deserializable, Defined, Clonable, Display, Clone, Any, Cacheable)
   - Protocol/trait constraint validation
   - Constraint composition and validation
   - Multiple constraint kinds (trait, protocol, capability, value)

4. ✅ Integration with type checker:
   - Generic type instantiation and checking
   - Type parameter substitution
   - Constraint satisfaction verification
   - Comprehensive error reporting

**Test Requirements**: ✅ **ALL COMPLETED**
- Grammar parsing for generic syntax variations ✅
- Constraint definition and checking ✅
- Type parameter inference and substitution ✅
- Integration with existing type system ✅
- Complex generic type scenarios ✅

**Success Criteria**: ✅ **ALL ACHIEVED**
- Generic syntax parses correctly in tree-sitter ✅
- Constraint checking infrastructure implemented ✅
- Type substitution works for generic patterns ✅
- Integration tests enabled (97.0% pass rate) ✅
```

---

## Phase 2: Advanced Language Features (High Priority)

### Step 2.1: Modern Class System Implementation

**Goal**: Complete implementation of modern Perl class/field/method syntax with full type system integration

**Status**: 🔶 **IN PROGRESS** - Grammar and AST structures complete, need conversion layer and type system integration

```
Complete the modern class system implementation for new-style Perl classes (class/field/method syntax).

**Context**:
- Grammar fully supports modern class syntax with 50/50 tests passing
- AST structures (ClassDecl, FieldDecl, MethodDecl) are complete
- Missing: Tree-sitter to internal AST conversion and type system integration
- CRITICAL: Must distinguish between old blessed hash style vs new class/field style - they are incompatible

**Two Perl Object Systems**:
1. **Old blessed hash style**: `package Foo { sub new { bless {...}, $class } }` - Regular Perl code
2. **New class/field style**: `class Foo { field $x; method new() {...} }` - Special syntax with field encapsulation

**Requirements**:
1. ✅ Grammar support (COMPLETED - full modern class syntax supported)
2. ❌ Implement tree-sitter to AST conversion for class statements
3. ❌ Add ClassType integration with type system
4. ❌ Implement method signature checking and dispatch validation
5. 🚫 **DEFERRED**: Role composition (roles not yet core Perl feature)

**Implementation**:
1. **AST Conversion** (Priority 1):
   - Fix conversion from tree-sitter `class_statement` nodes to internal `ClassDecl` AST
   - Ensure `field $name` declarations become `FieldDecl` nodes with type info
   - Convert `method` declarations to `MethodDecl` nodes with signatures
   - Handle class inheritance (`class Child :isa(Parent)`)

2. **Type System Integration** (Priority 2):
   - Create ClassType in `internal/typedef/classes.go`
   - Register classes as proper types in type registry
   - Implement field access type checking (cannot access fields as hash keys)
   - Add method resolution and dispatch validation

3. **Method and Field Analysis** (Priority 3):
   - Field type checking and initialization validation
   - Method signature compatibility in inheritance
   - Private vs public field access control
   - Integration with existing generic type system

4. **Modern Class Semantics** (Priority 4):
   - Enforce field encapsulation (no hash-style access to fields)
   - Proper constructor/destructor handling
   - BUILD/ADJUST phaser support
   - Class-specific error messages

**Test Requirements**:
- Modern class declaration parsing and AST conversion
- Field encapsulation enforcement (no hash access)
- Method signature and inheritance validation
- Integration with existing type features
- Clear distinction between old vs new object systems

**Success Criteria**:
- Modern class syntax converts correctly to internal AST
- Type system recognizes classes as proper types
- Field encapsulation is enforced
- Method dispatch validation works
- Clear error messages distinguish between object system styles

**DEFERRED**: Role implementation will be addressed in a future step when roles become a core Perl feature.
```

### Step 2.2: Advanced Method Signatures

**Goal**: Complete method signature parsing and validation with complex return types

```
Fix method signature parsing conflicts and implement comprehensive method type checking.

**Context**: Method return type annotations have parsing conflicts (`internal/parser/parser_test.go:299` skips - "tree-sitter parsing conflicts with empty parentheses").

**Requirements**:
1. Fix grammar conflicts in method signature parsing
2. Implement comprehensive method signature validation
3. Add return type inference and checking
4. Support complex method signatures with generics

**Implementation**:
1. Grammar fixes in tree-sitter-typed-perl:
   - Resolve parsing conflicts with method signatures
   - Support complex return types: `method func() -> ArrayRef[Int]`
   - Handle parameterized types in signatures
   - Fix empty parentheses ambiguity

2. Method signature validation:
   - Parameter type checking and validation
   - Return type compatibility verification
   - Default parameter type inference
   - Signature overloading validation

3. Advanced signature features:
   - Generic method signatures: `method sort<T: Ord>(ArrayRef[T] $items) -> ArrayRef[T]`
   - Named parameters with types: `method process(Int :$count, Str :$name)`
   - Variable argument types: `method log(Str $format, @args)`
   - Optional and default parameters

4. Integration with type checking:
   - Method call site validation
   - Return type propagation
   - Parameter passing verification
   - Integration with class/role system

**Test Requirements**:
- Method signature parsing accuracy
- Type checking for various signature patterns
- Return type inference and validation
- Integration with OOP features
- Complex generic method scenarios

**Success Criteria**:
- All method signature parsing conflicts resolved
- Comprehensive method type checking works
- Integration with class/role system complete
- Performance is acceptable for large codebases
```

---

## Phase 3: System Integration (Medium Priority)

### Step 3.1: System Perl Detection and Management

**Goal**: Implement cross-platform system Perl detection and automated installation

```
Create robust system Perl integration to enable all skipped E2E tests.

**Context**: 25+ E2E tests skip due to missing system Perl automation (`test/e2e/helpers/assertions.go` SkipIfNoSystemPerl function).

**Requirements**:
1. Implement cross-platform Perl detection and installation
2. Create PerlVersionManager with automated installation
3. Add system integration for major platforms
4. Enable comprehensive E2E testing

**Implementation**:
1. In `internal/perl/system_manager.go`:
   - Create SystemPerlManager with detection logic
   - Implement cross-platform installation (Windows, macOS, Linux)
   - Add version validation and compatibility checking
   - Support multiple Perl distributions (system, plenv, perlbrew)

2. Platform-specific implementation:
   - Windows: Strawberry Perl, ActivePerl detection/installation
   - macOS: System perl, Homebrew, plenv integration
   - Linux: Distribution packages, source compilation
   - Docker: Container-based Perl environments

3. Version management integration:
   - Automatic version detection and validation
   - Installation of missing versions
   - Integration with .perl-version files
   - Fallback to system Perl when appropriate

4. E2E test enablement:
   - Remove SkipIfNoSystemPerl guards from tests
   - Add setup/teardown for test environments
   - Create isolated test environments
   - Validate cross-platform functionality

**Test Requirements**:
- Cross-platform Perl detection accuracy
- Installation success across different environments
- Version management and switching
- E2E test reliability and isolation
- Performance of detection and installation

**Success Criteria**:
- All E2E tests run without system dependency skips
- Cross-platform installation works reliably
- Version management integrates seamlessly
- Test suite runs in CI/CD environments
```

### Step 3.2: Enhanced PVI Module Analysis

**Goal**: Implement real module analysis for accurate type definition generation

```
Replace placeholder type generation in PVI with actual module analysis.

**Context**: PVI creates placeholder type definitions instead of analyzing modules (`internal/pvi/type_command.go:237-239` TODO comment).

**Requirements**:
1. Implement ModuleAnalyzer with AST-based analysis
2. Create accurate type definition extraction
3. Add dependency analysis and type propagation
4. Integrate with existing type system

**Implementation**:
1. In `internal/pvi/analyzer.go`:
   - Create ModuleAnalyzer using parser infrastructure
   - Implement AST traversal for type extraction
   - Add symbol table construction for modules
   - Support for complex module patterns

2. Type definition extraction:
   - Extract function signatures and export lists
   - Identify class/role definitions and methods
   - Parse embedded documentation for type hints
   - Handle complex Perl metaprogramming patterns

3. Integration with type system:
   - Generate accurate .typedef.json files
   - Support for complex type hierarchies
   - Integration with union types and generics
   - Cross-module type dependency resolution

4. Enhanced PVI workflow:
   - Analyze modules before generating type definitions
   - Update existing type definitions when modules change
   - Validate type definition accuracy
   - Provide analysis reports and statistics

**Test Requirements**:
- Accurate type extraction from real modules
- Complex Perl pattern handling
- Type definition generation accuracy
- Integration with type checking pipeline
- Performance with large module hierarchies

**Success Criteria**:
- Generated type definitions are accurate and useful
- Analysis handles common Perl module patterns
- Integration with type checker works seamlessly
- Performance is acceptable for typical modules
```

---

## Phase 4: Advanced Tooling (Medium Priority)

### Step 4.1: LSP Advanced Features

**Goal**: Complete LSP implementation with query system and auto-fix capabilities

```
Implement the missing LSP features to provide full IDE integration.

**Context**: LSP query system returns "not yet implemented" (`internal/lsp/queries.go:69,75`) and auto-fix generation is stubbed.

**Requirements**:
1. Implement query system for type and symbol information
2. Add auto-fix generation for common type errors
3. Create enhanced formatting and refactoring tools
4. Integrate with flow-sensitive analysis

**Implementation**:
1. Query system in `internal/lsp/queries.go`:
   - Implement type queries for hover information
   - Add symbol queries for go-to-definition
   - Create reference finding with type context
   - Support for workspace symbol search

2. Auto-fix generation:
   - Type mismatch fixes with suggestions
   - Import statement generation and cleanup
   - Variable declaration fixes
   - Method signature corrections

3. Enhanced features:
   - Semantic highlighting with type information
   - Inlay hints for inferred types
   - Code completion with type context
   - Refactoring operations (rename, extract method)

4. Integration with type system:
   - Use flow-sensitive analysis for accurate information
   - Leverage union type system for better suggestions
   - Integration with class/role system
   - Real-time type checking and error reporting

**Test Requirements**:
- Query system accuracy and performance
- Auto-fix generation quality and safety
- IDE integration testing
- Real-world workflow validation
- Performance with large codebases

**Success Criteria**:
- Full IDE integration works smoothly
- Auto-fixes are accurate and helpful
- Performance is acceptable for interactive use
- Integration with advanced type features complete
```

### Step 4.2: MCP Code Generation

**Goal**: Implement AI-assisted code generation using PVM's type system

```
Complete the MCP code generation system with full integration.

**Context**: Complete interface exists but zero implementation (`internal/mcp/tools/generate.go` - 11 test functions skip with "not yet implemented").

**Requirements**:
1. Implement CodeGenerator with type-aware generation
2. Add collaborative sampling with validation loops
3. Create generation templates for functions, classes, tests
4. Integrate with PVM's type system for accurate generation

**Implementation**:
1. In `internal/mcp/tools/generate.go`:
   - Complete CodeGenerator implementation
   - Add type-aware prompt generation
   - Implement validation and fixing loops
   - Create memory integration for learning

2. Generation capabilities:
   - Function generation with proper signatures
   - Class/role generation with type annotations
   - Test generation with type-aware assertions
   - Documentation generation from types

3. Validation integration:
   - Use PSC for type checking generated code
   - Implement syntax and style validation
   - Add auto-fixing for common generation errors
   - Integration with project context

4. Collaborative features:
   - Memory system for learning from corrections
   - Context-aware prompt building
   - Integration with sampling for quality improvement
   - User feedback incorporation

**Test Requirements**:
- Code generation accuracy and quality
- Validation loop effectiveness
- Integration with type system
- Memory and learning functionality
- Real-world generation scenarios

**Success Criteria**:
- Generated code is syntactically correct and well-typed
- Validation catches and fixes common issues
- Integration with development workflow is smooth
- Learning improves generation quality over time
```

---

## Phase 5: Production Readiness (Lower Priority)

### Step 5.1: Cross-Platform Reliability

**Goal**: Ensure 100% cross-platform compatibility and test coverage

```
Eliminate platform-specific test skips and ensure full Windows/macOS/Linux support.

**Context**: 7 tests skip due to platform limitations (Windows CI, symlink creation, file permissions).

**Requirements**:
1. Fix Windows-specific issues in build and test systems
2. Resolve symlink creation and file permission problems
3. Ensure CI/CD compatibility across platforms
4. Achieve 100% test pass rate on all platforms

**Implementation**:
1. Windows compatibility fixes:
   - Fix symlink creation issues (`internal/cli/symlinks_test.go`)
   - Resolve file permission problems
   - Ensure path handling works correctly
   - Add Windows-specific CI testing

2. Cross-platform testing:
   - Remove platform-specific test skips
   - Add comprehensive CI matrix testing
   - Validate functionality on all target platforms
   - Performance testing across platforms

3. Build system enhancements:
   - Cross-platform build artifact generation
   - Platform-specific installation packages
   - Docker container support
   - Distribution packaging for major platforms

**Test Requirements**:
- 100% test pass rate on Windows, macOS, Linux
- CI/CD validation on all platforms
- Real-world usage testing
- Performance parity across platforms
- Installation and deployment testing

**Success Criteria**:
- No platform-specific test skips remain
- Full functionality on all supported platforms
- CI/CD runs successfully on all platforms
- Distribution packages work correctly
```

### Step 5.2: Performance Optimization

**Goal**: Optimize performance for large codebases and production usage

```
Implement comprehensive performance optimizations and monitoring.

**Context**: Performance tests skip in short mode, need production-grade performance for large projects.

**Requirements**:
1. Optimize type checking and flow analysis performance
2. Implement intelligent caching throughout the system
3. Add performance monitoring and regression detection
4. Ensure scalability for large codebases

**Implementation**:
1. Type system optimization:
   - Incremental type checking with dependency tracking
   - Optimized flow analysis with early termination
   - Smart cache invalidation for type information
   - Parallel processing where appropriate

2. Caching strategy:
   - Parse result caching with content hashing
   - Type information caching across sessions
   - Build artifact caching and validation
   - Configuration and project context caching

3. Performance monitoring:
   - Built-in performance profiling and metrics
   - Regression detection with baseline comparisons
   - Resource usage monitoring and reporting
   - Bottleneck identification and optimization

**Test Requirements**:
- Performance benchmarks for all major operations
- Memory usage validation and optimization
- Large codebase testing and validation
- Regression detection accuracy
- Real-world performance validation

**Success Criteria**:
- Type checking performance is acceptable for large projects
- Memory usage is reasonable and bounded
- Performance regressions are caught automatically
- System scales to enterprise-sized codebases
```

---

## Implementation Strategy

### Development Phases

**Phase 1 (8-10 weeks)**: Core Type System Foundation
- Highest impact features that unlock advanced capabilities
- Flow-sensitive analysis, union types, generics
- Foundation for all other advanced features

**Phase 2 (4-6 weeks)**: Advanced Language Features
- Complete type system with OOP support
- Method signatures and complex type patterns
- Production-ready type checking

**Phase 3 (4-6 weeks)**: System Integration
- Cross-platform Perl management
- Real module analysis and type generation
- Enable comprehensive testing

**Phase 4 (3-4 weeks)**: Advanced Tooling
- LSP feature completion
- MCP code generation system
- Enhanced developer experience

**Phase 5 (2-3 weeks)**: Production Readiness
- Cross-platform reliability
- Performance optimization
- Enterprise-grade quality

### Quality Gates

**Each Step Must**:
1. Include comprehensive test coverage (100% for new code)
2. Maintain backward compatibility
3. Pass all existing tests
4. Include integration tests
5. Document new functionality

**Each Phase Must**:
1. Achieve specific success criteria
2. Demonstrate measurable improvement
3. Maintain system stability
4. Provide user value
5. Set foundation for next phase

### Risk Mitigation

**Technical Risks**:
- Flow analysis complexity → Start with simple cases, iterate
- Grammar conflicts → Prototype changes, extensive testing
- Performance impact → Benchmark early, optimize incrementally
- Cross-platform issues → Test on all platforms continuously

**Integration Risks**:
- Breaking changes → Maintain backward compatibility
- Test failures → Comprehensive test coverage required
- Performance regression → Continuous performance monitoring
- User experience → Validate workflows with real usage

This plan provides a clear path to completing PVM's most critical missing features while maintaining quality, stability, and user value throughout the implementation process.
