# PVM TypeScript-Go Integration - Prompt Plan

## Project Context

PVM (Perl Version Manager) is a tool for managing Perl versions with an innovative typed-Perl extension. Based on Microsoft's TypeScript-Go architecture patterns, we're modernizing PVM's compiler pipeline and build system to achieve:
- **8x performance improvement** potential through better architecture
- **Enhanced developer experience** with improved tooling
- **Better error reporting** through symbol binding phase
- **Improved LSP features** with dedicated language service separation
- **Modern build system** with code generation and testing infrastructure

The project includes:
- **PVM**: Core version management
- **PSC**: Static type checker for typed-Perl
- **PVI**: Package installer with type awareness
- **PVX**: Isolated execution environment

## Blueprint Overview

**Goal**: Integrate Microsoft TypeScript-Go architectural patterns into PVM
**Strategy**: Incremental, tested approach with backward compatibility
**Target Outcome**: Modern, performant Perl toolchain with TypeScript-quality developer experience

### Target Architecture
1. **Compiler Pipeline**: Scanner → Parser → Binder → Checker → Compiler
2. **LSP Separation**: Language Service (business logic) + LSP Protocol (handler)
3. **Enhanced Build System**: Code generation, better testing, performance monitoring
4. **Symbol Binding Phase**: Dedicated symbol resolution before type checking
5. **Modern Testing**: Baseline testing, benchmarks, coverage integration

## Implementation Phases

### Phase 1: Foundation Architecture (Steps 1-4)
- Scanner extraction and AST consolidation
- Pipeline reorganization with backward compatibility

### Phase 2: Symbol Binding (Steps 5-8)
- Implement dedicated symbol resolution phase
- Update type checker to use symbol tables

### Phase 3: LSP Enhancement (Steps 9-11)
- Split LSP into language service + protocol handler
- Leverage symbol information for better features

### Phase 4: Build System Modernization (Steps 12-15)
- Add code generation infrastructure
- Implement baseline testing and performance monitoring

### Phase 5: Integration and Optimization (Steps 16-18)
- Performance optimization and integration testing
- Documentation and migration guide

---

## Step-by-Step Implementation Prompts

### Step 1: Scanner Extraction ✅ COMPLETED

```
You are modernizing PVM's architecture by integrating Microsoft TypeScript-Go patterns.

TASK: Extract lexical analysis into a dedicated scanner package, following TypeScript-Go's scanner/parser separation.

CONTEXT: Currently, PVM's parser package mixes scanning and parsing concerns. TypeScript-Go separates these for better modularity and performance.

REQUIREMENTS:
1. Create `internal/scanner/` package
2. Implement tree-sitter token wrapper that provides a clean Token interface
3. Add comprehensive token types and constants
4. Update parser to consume tokens instead of raw source text
5. Maintain backward compatibility with wrapper functions
6. Add thorough scanner unit tests

TECHNICAL DETAILS:
- Scanner should wrap tree-sitter's lexical analysis
- Provide Token interface with Position, Type, and Value
- Support incremental tokenization for LSP features
- Maintain all existing parser functionality through compatibility layer

DELIVERABLES:
- internal/scanner/ package with Token interface
- Updated parser consuming tokens
- Full backward compatibility maintained
- Comprehensive test coverage

SUCCESS CRITERIA:
- All existing parser tests pass unchanged
- Scanner can tokenize Perl source independently
- Foundation ready for incremental parsing improvements
- No breaking changes to public APIs

Focus on clean separation of concerns while maintaining compatibility.
```

### Step 2: AST Consolidation and Navigation ✅ COMPLETED

```
You are continuing the PVM architecture modernization from Step 1.

CONTEXT: Scanner extraction is complete. Now consolidate AST types and add navigation utilities following TypeScript-Go's astnav pattern.

TASK: Consolidate scattered AST types and create navigation utilities for better code organization.

REQUIREMENTS:
1. Create `internal/ast/` package and move all AST node types from parser, typechecker, typedef
2. Organize AST types logically (expressions.go, statements.go, types.go, etc.)
3. Create `internal/astnav/` package with navigation utilities
4. Implement AST traversal, search, and visitor pattern utilities
5. Update all imports across the codebase to use consolidated AST types
6. Add comprehensive AST manipulation utilities

TECHNICAL DETAILS:
- Consolidate AST types scattered across multiple packages
- Implement visitor pattern for AST traversal
- Add utilities: FindNodeAt, GetParent, GetChildren, WalkAST
- Ensure type safety and performance in navigation utilities
- Maintain all existing AST functionality

DELIVERABLES:
- internal/ast/ with consolidated node types
- internal/astnav/ with navigation utilities
- Updated imports throughout codebase
- Enhanced AST manipulation capabilities

SUCCESS CRITERIA:
- All AST types centralized and well-organized
- Navigation utilities support common use cases
- All existing functionality preserved
- Foundation ready for symbol binding and LSP improvements

Prioritize clean organization and useful utilities for future development.
```

### Step 3: Pipeline Integration and Compatibility ✅ COMPLETED

```
You are continuing the PVM architecture modernization from Step 2.

CONTEXT: Scanner and AST consolidation are complete. Now integrate the new pipeline while maintaining full compatibility.

TASK: Update all components to use the new scanner→parser pipeline while preserving existing functionality.

REQUIREMENTS:
1. Update parser to output consolidated AST types using scanner tokens
2. Update typechecker and typedef to consume new AST format
3. Update compiler package to work with new pipeline
4. Update PSC commands to use new scanner/parser APIs
5. Update LSP implementation to leverage new AST navigation utilities
6. Ensure all existing functionality works identically
7. Add integration tests validating pipeline functionality

TECHNICAL DETAILS:
- Maintain all existing public APIs through compatibility wrappers
- Ensure performance is same or better than before
- Update error handling to work with new token-based parsing
- Verify all edge cases still work correctly

DELIVERABLES:
- All components using new scanner→parser pipeline
- Full backward compatibility maintained
- Integration tests validating functionality
- Performance baseline established

SUCCESS CRITERIA:
- Full test suite passes without modification
- Performance matches or exceeds baseline
- All components work with new AST structure
- LSP features improved with better AST navigation

Foundation phase complete - ready for symbol binding implementation.
```

### Step 4: Performance and Stability Validation ✅ COMPLETED

```
You are continuing the PVM architecture modernization from Step 3.

CONTEXT: The new scanner→parser pipeline is integrated. Now validate performance and stability before proceeding to symbol binding.

TASK: Comprehensive validation of the new pipeline architecture with performance benchmarking.

REQUIREMENTS:
1. Run comprehensive test suite and verify all tests pass
2. Benchmark parsing performance against baseline
3. Benchmark type checking performance against baseline
4. Test memory usage and identify any regressions
5. Validate LSP responsiveness with large files
6. Test error handling and edge cases thoroughly
7. Document performance characteristics and any improvements

TECHNICAL DETAILS:
- Use Go's testing package for benchmarks
- Test with realistic Perl codebases of varying sizes
- Profile memory allocation patterns
- Validate error messages are preserved
- Test incremental parsing capabilities

DELIVERABLES:
- Performance benchmark report
- Memory usage analysis
- Stability validation results
- Any optimizations needed for performance

SUCCESS CRITERIA:
- No performance regressions (target: same or better)
- Memory usage within 110% of baseline
- All functionality preserved exactly
- Foundation stable for symbol binding phase

This validation ensures solid foundation before adding symbol binding complexity.
```

### Step 5: Symbol Binding Architecture Design ✅ COMPLETED

```
You are continuing the PVM architecture modernization from Step 4.

CONTEXT: The scanner→parser pipeline is validated and stable. Now design the symbol binding phase following TypeScript-Go's binder architecture.

TASK: Design and implement the foundation for symbol binding that will resolve symbols before type checking.

REQUIREMENTS:
1. Create `internal/binder/` package with symbol resolution architecture
2. Design Symbol, Scope, and SymbolTable data structures
3. Implement scope chain management for Perl's lexical scoping rules
4. Handle basic variable declarations (my, our, state)
5. Implement subroutine and method symbol binding
6. Design integration points with type checker
7. Create comprehensive binder unit tests

TECHNICAL DETAILS:
- Design for Perl's complex scoping rules (lexical, dynamic, package)
- Handle symbol shadowing and closure capture
- Support incremental symbol resolution for LSP
- Plan for cross-module symbol resolution
- Ensure performance for large codebases

SYMBOL TYPES TO HANDLE:
- Scalar variables ($var)
- Array variables (@array)
- Hash variables (%hash)
- Subroutines (sub name)
- Methods (method name)
- Package symbols
- Imported symbols (basic)

DELIVERABLES:
- internal/binder/ package with core architecture
- Symbol, Scope, SymbolTable implementations
- Basic variable and subroutine binding
- Comprehensive test coverage

SUCCESS CRITERIA:
- Binder correctly identifies and resolves basic symbols
- Scope chains properly maintained for lexical scoping
- Foundation ready for integration with type checker
- Performance acceptable for medium-sized files

Focus on correct Perl scoping semantics and clean architecture.
```

### Step 6: Advanced Symbol Binding Features ✅ COMPLETED

```
You are continuing the PVM architecture modernization from Step 5.

CONTEXT: Basic symbol binding architecture is implemented. Now add advanced Perl scoping features and edge cases.

TASK: Extend symbol binding to handle Perl's complex scoping scenarios and prepare for type checker integration.

REQUIREMENTS:
1. Implement package scope handling and package qualification
2. Add support for 'local' dynamic scoping
3. Handle closure variable capture correctly
4. Add module import/export symbol resolution (basic)
5. Implement typeglob handling for symbol table manipulation
6. Handle symbol aliasing and references
7. Add comprehensive test coverage for complex scoping scenarios

ADVANCED PERL FEATURES:
- 'our' variables and package scope inheritance
- 'local' dynamic scoping with restore semantics
- Closure variable capture and upvalue resolution
- Package qualification ($Package::var)
- Symbol table inheritance and manipulation
- Import/export symbol resolution

TECHNICAL DETAILS:
- Handle Perl's dynamic nature while providing static analysis benefits
- Ensure symbol resolution works across module boundaries
- Support both compile-time and runtime symbol creation patterns
- Optimize for common cases while handling edge cases correctly

DELIVERABLES:
- Advanced scoping features in binder
- Cross-module symbol resolution
- Comprehensive test suite for complex scenarios
- Documentation of supported and unsupported patterns

SUCCESS CRITERIA:
- Complex Perl scoping scenarios work correctly
- Module boundaries properly respected
- Performance remains acceptable for large codebases
- Edge cases documented and tested

Prepare binder for production use with type checker integration.
```

### Step 7: Type Checker Integration with Symbol Tables ✅ COMPLETED

```
You are continuing the PVM architecture modernization from Step 6.

CONTEXT: Symbol binding is complete with advanced Perl features. Now integrate symbol tables with the type checker.

TASK: Refactor the type checker to use symbol information from the binder phase.

REQUIREMENTS:
1. ✅ Update type checker to consume symbol tables from binder
2. ✅ Remove inline symbol resolution logic from type checker
3. ✅ Enhance error messages with symbol context and location information
4. ✅ Update inference engine to use symbol context for better type inference
5. ✅ Add type checking for symbol references with proper scope validation
6. ✅ Maintain all existing type checking accuracy while improving error quality
7. ✅ Provide backward compatibility wrapper for existing APIs

NEW PIPELINE:
- ✅ Source → Scanner → Parser → Binder → Checker → Compiler
- ✅ Type checker focuses purely on type analysis using symbol information

TECHNICAL DETAILS:
- ✅ Preserve all existing type checking logic and accuracy
- ✅ Enhance error messages: "Variable '$name' declared as Int at line 5, assigned Str at line 12"
- ✅ Use symbol scoping information for better type inference
- ✅ Handle type flow through different scopes correctly

DELIVERABLES:
- ✅ Refactored type checker using symbol tables
- ✅ Enhanced error messages with symbol context
- ✅ Backward compatibility maintained
- ✅ Integration tests validating improved error reporting

SUCCESS CRITERIA:
- ✅ All type checking accuracy preserved
- ✅ Error messages significantly more helpful with symbol context
- ✅ No regressions in type inference quality
- ✅ Foundation ready for enhanced LSP features

IMPLEMENTATION COMPLETED:
- Updated NewTypeChecker to accept symbol table parameter
- Integrated binder into type checking pipeline
- Added symbol-aware error reporting with suggestions
- Enhanced inference engine with symbol context
- Fixed pattern matching priorities for accurate type inference
- Added proper type compatibility rules for Perl (Int->Bool)
- All tests passing with improved type inference accuracy
```

### Step 8: Enhanced Error Reporting and Diagnostics ✅ COMPLETED

```
You are continuing the PVM architecture modernization from Step 7.

CONTEXT: Type checker now uses symbol tables. Now enhance error reporting to leverage symbol information for better diagnostics.

TASK: Implement comprehensive error reporting improvements using symbol binding information.

REQUIREMENTS:
1. Enhance error messages with precise symbol context and locations
2. Add "undefined variable" detection with suggestions
3. Implement symbol shadowing warnings
4. Add unused variable detection and warnings
5. Create cross-reference information for symbols in errors
6. Implement symbol usage tracking for better diagnostics
7. Add comprehensive diagnostic testing framework

ERROR IMPROVEMENTS:
- Before: "Type mismatch: expected Int, got Str"
- After: "Variable '$count' declared as Int at line 5, but assigned Str value 'hello' at line 12"
- Add: "Undefined variable '$typo' - did you mean '$type'?"
- Add: "Variable '$unused' declared but never used"
- Add: "Variable '$name' shadows outer scope variable at line 3"

TECHNICAL DETAILS:
- Use symbol table information for precise error locations
- Implement edit distance algorithm for variable name suggestions
- Track symbol usage patterns for unused variable detection
- Provide actionable suggestions in error messages

DELIVERABLES:
- Enhanced error reporting with symbol context
- Unused variable and shadowing detection
- Suggestion system for undefined variables
- Comprehensive diagnostic test suite

SUCCESS CRITERIA:
- Error messages provide actionable information with context
- Developers can quickly identify and fix symbol-related issues
- No false positive errors or warnings
- Error quality significantly improved over baseline

Prioritize actionable, helpful diagnostics that improve developer productivity.
```

### Step 9: LSP Architecture Separation ✅ COMPLETED

```
You are continuing the PVM architecture modernization from Step 8.

CONTEXT: Symbol binding and enhanced error reporting are complete. Now restructure LSP following TypeScript-Go's separation pattern.

TASK: Split LSP implementation into language service (business logic) and protocol handler (LSP protocol) following TypeScript-Go architecture.

REQUIREMENTS:
1. Create `internal/ls/` package for language service business logic
2. Refactor existing `internal/lsp/` to handle only protocol concerns
3. Implement language service methods using symbol tables and AST navigation
4. Provide clean interface between language service and protocol handler
5. Leverage symbol binding for accurate goto definition and references
6. Use AST navigation utilities for better hover and completion features
7. Maintain full backward compatibility with existing LSP functionality

ARCHITECTURE SEPARATION:
```go
// internal/ls/ - Language Service (business logic)
type LanguageService struct {
    parser  *parser.Parser
    binder  *binder.Binder
    checker *checker.TypeChecker
}

func (ls *LanguageService) GetDefinition(uri string, pos Position) (*Definition, error)
func (ls *LanguageService) GetHover(uri string, pos Position) (*Hover, error)

// internal/lsp/ - LSP Protocol Handler
type Server struct {
    ls *ls.LanguageService
}
```

DELIVERABLES:
- internal/ls/ with language service business logic
- Refactored internal/lsp/ for protocol handling only
- Clean separation of concerns
- All existing LSP functionality preserved

SUCCESS CRITERIA:
- LSP features work identically to before
- Clean separation enables future feature development
- Symbol-based features ready for enhancement
- Architecture matches TypeScript-Go patterns

Foundation ready for enhanced LSP features using symbol information.

IMPLEMENTATION COMPLETED:
- Created internal/ls/ package with LanguageService that handles business logic
- Refactored internal/lsp/ to use LanguageService for all analysis operations
- Implemented clean separation: LSP handles protocol, LanguageService handles features
- Added conversion functions between LSP and LanguageService types
- Maintained backward compatibility - all existing LSP functionality preserved
- Language service integrates with symbol tables and AST navigation utilities
- Foundation ready for enhanced symbol-aware LSP features in next steps
```

### Step 10: Enhanced LSP Features with Symbol Information

```
You are continuing the PVM architecture modernization from Step 9.

CONTEXT: LSP architecture is separated. Now enhance LSP features using symbol binding and AST navigation.

TASK: Implement enhanced LSP features leveraging symbol tables and AST navigation utilities.

REQUIREMENTS:
1. Implement accurate goto definition using symbol resolution
2. Add find all references functionality using symbol tables
3. Enhance hover information with symbol details and type information
4. Improve autocompletion with symbol-aware suggestions
5. Add rename symbol capability across files
6. Implement document symbol outline using AST navigation
7. Add workspace symbol search functionality

ENHANCED FEATURES:
- Goto definition: accurate across modules and scopes
- Find references: all uses of symbols including across files
- Hover: show symbol declaration, type, and documentation
- Completion: context-aware suggestions based on available symbols
- Rename: safely rename symbols with scope awareness
- Outline: document symbol hierarchy
- Workspace search: find symbols across entire project

TECHNICAL DETAILS:
- Use symbol tables for accurate cross-reference information
- Leverage AST navigation for efficient symbol searches
- Ensure performance is acceptable for large projects
- Handle incremental updates for responsive experience

DELIVERABLES:
- Enhanced goto definition and find references
- Improved hover and completion features
- Rename symbol and workspace search functionality
- Performance optimizations for large codebases

SUCCESS CRITERIA:
- LSP features significantly improved over baseline
- Goto definition works accurately across modules
- Find references comprehensive and fast
- Rename operations safe and complete
- Performance acceptable for large Perl projects

Focus on accuracy and performance for production-quality LSP experience.
```

### Step 11: LSP Performance Optimization and Testing

```
You are continuing the PVM architecture modernization from Step 10.

CONTEXT: Enhanced LSP features are implemented. Now optimize performance and validate with comprehensive testing.

TASK: Optimize LSP performance and create comprehensive testing framework for LSP functionality.

REQUIREMENTS:
1. Profile LSP operations and identify performance bottlenecks
2. Implement incremental parsing and symbol resolution for file changes
3. Add caching for expensive operations (symbol lookups, type checking)
4. Optimize memory usage for large projects
5. Create comprehensive LSP integration test suite
6. Benchmark LSP responsiveness against targets (<100ms for common operations)
7. Add stress testing with large Perl codebases

PERFORMANCE TARGETS:
- Goto definition: <50ms
- Find references: <200ms
- Hover information: <25ms
- Completions: <100ms
- Symbol resolution: <100ms for typical files
- Memory usage: <500MB for large projects

TECHNICAL DETAILS:
- Implement incremental symbol resolution
- Cache symbol tables and type information
- Use background processing for expensive operations
- Optimize data structures for LSP access patterns

DELIVERABLES:
- Performance-optimized LSP implementation
- Comprehensive LSP test suite
- Performance benchmark results
- Memory usage analysis and optimizations

SUCCESS CRITERIA:
- LSP performance meets or exceeds targets
- Memory usage acceptable for large projects
- All LSP features tested comprehensively
- Performance regressions detected automatically
- Production-ready LSP implementation

LSP architecture modernization complete - ready for build system improvements.
```

### Step 12: Build System Foundation and Tool Management

```
You are continuing the PVM architecture modernization from Step 11.

CONTEXT: Core architecture modernization is complete. Now modernize the build system following TypeScript-Go patterns.

TASK: Establish modern build system foundation with automated tool management and build configuration.

REQUIREMENTS:
1. Create tools.go for development dependency management
2. Add build tags for debug/release/development builds
3. Implement automated tool installation and updates
4. Add build performance monitoring and tracking
5. Create development vs production build targets with different optimizations
6. Establish code generation infrastructure foundation
7. Update Makefile with enhanced build targets and tool integration

BUILD CONFIGURATION:
```makefile
# Tool installation
.PHONY: install-tools
install-tools:
    go install github.com/matryer/moq@latest
    go install golang.org/x/tools/cmd/stringer@latest
    go install gotest.tools/gotestsum@latest

# Build modes
build-dev: BUILD_TAGS += debug,noembed
build-release: BUILD_TAGS += release,embed
```

TECHNICAL DETAILS:
- Establish build tags for conditional compilation
- Create tool dependency management with tools.go
- Add performance monitoring for build processes
- Prepare infrastructure for code generation

DELIVERABLES:
- Enhanced Makefile with modern build targets
- Automated tool management system
- Build performance monitoring
- Foundation for code generation

SUCCESS CRITERIA:
- Developers can set up environment with single command
- Build performance is monitored and tracked
- Development and production builds properly differentiated
- Tool dependencies managed automatically

Foundation ready for code generation and testing improvements.
```

### Step 13: Code Generation Infrastructure

```
You are continuing the PVM architecture modernization from Step 12.

CONTEXT: Build system foundation is established. Now implement comprehensive code generation following TypeScript-Go patterns.

TASK: Implement automated code generation for repetitive code patterns using go generate.

REQUIREMENTS:
1. Add go generate directives throughout codebase for automated code generation
2. Implement AST node string generation using stringer
3. Add mock generation for interfaces using moq
4. Create error code generation from structured definitions
5. Implement diagnostic message generation
6. Add generation verification in CI to ensure generated code is up to date
7. Create generation scripts for custom code patterns

CODE GENERATION TARGETS:
```go
// AST node string methods
//go:generate stringer -type=NodeType -output=node_string.go

// Mock generation for testing
//go:generate moq -out=mocks_test.go . Parser TypeChecker

// Error code generation
//go:generate go run scripts/generate_errors.go

// Diagnostic message generation
//go:generate go run scripts/generate_diagnostics.go
```

TECHNICAL DETAILS:
- Eliminate manual maintenance of repetitive code
- Ensure generated code is always up to date
- Integrate generation into build process
- Add CI verification of generated code

DELIVERABLES:
- Comprehensive code generation system
- Generated string methods, mocks, and error codes
- CI integration for generated code verification
- Documentation for extending code generation

SUCCESS CRITERIA:
- No manually maintained repetitive code
- Generated code always up to date
- CI catches outdated generated code
- Easy to extend generation for new patterns

Build system ready for enhanced testing infrastructure.
```

### Step 14: Baseline Testing and Performance Monitoring

```
You are continuing the PVM architecture modernization from Step 13.

CONTEXT: Code generation is implemented. Now establish comprehensive testing infrastructure with baseline testing and performance monitoring.

TASK: Implement baseline testing system and performance monitoring following TypeScript-Go patterns.

REQUIREMENTS:
1. Implement baseline testing framework for regression prevention
2. Integrate gotestsum for enhanced test output and reporting
3. Add benchmark testing infrastructure with performance regression detection
4. Implement coverage reporting and analysis
5. Create integration test framework for end-to-end validation
6. Add performance monitoring for critical operations
7. Establish CI integration for comprehensive test reporting

BASELINE TESTING:
```go
func TestTypeChecker_Baselines(t *testing.T) {
    testCases := []string{"simple_types", "union_types", "intersection_types"}
    for _, tc := range testCases {
        input := readFile("testdata/input/" + tc + ".pl")
        expected := readFile("testdata/baseline/" + tc + ".expected")
        result := runTypeChecker(input)
        if diff := cmp.Diff(expected, result); diff != "" {
            t.Errorf("Baseline mismatch (-want +got):\n%s", diff)
        }
    }
}
```

TECHNICAL DETAILS:
- Compare actual output against expected baselines
- Track performance regressions automatically
- Provide clear test output for debugging
- Integrate with CI for automated validation

DELIVERABLES:
- Baseline testing framework with comprehensive test cases
- Enhanced test output with gotestsum integration
- Performance benchmarking and regression detection
- Coverage reporting and analysis tools

SUCCESS CRITERIA:
- Regression detection prevents breaking changes
- Test output is clear and actionable
- Performance regressions caught automatically
- Coverage tracks testing comprehensiveness

Testing infrastructure complete - ready for CI/CD integration.
```

### Step 15: CI/CD Integration and Automation

```
You are continuing the PVM architecture modernization from Step 14.

CONTEXT: Testing infrastructure is complete. Now integrate enhanced build system with CI/CD following TypeScript-Go patterns.

TASK: Update CI/CD workflows to leverage new build system capabilities and ensure comprehensive automation.

REQUIREMENTS:
1. Update GitHub Actions to use new build targets and tool management
2. Add code generation verification in CI pipeline
3. Implement comprehensive coverage reporting with codecov integration
4. Add performance regression detection in CI
5. Integrate security scanning and vulnerability detection
6. Add artifact collection and deployment automation
7. Create performance monitoring dashboards

ENHANCED CI WORKFLOW:
```yaml
jobs:
  test:
    steps:
    - name: Install tools
      run: make install-tools
    - name: Verify code generation
      run: make check-generate
    - name: Run comprehensive tests
      run: make test-all
    - name: Performance regression check
      run: make bench-compare
    - name: Security scan
      run: govulncheck ./...
```

TECHNICAL DETAILS:
- Verify generated code is up to date
- Track performance regressions over time
- Integrate security scanning into pipeline
- Provide comprehensive build artifacts

DELIVERABLES:
- Enhanced GitHub Actions with new build system
- Code generation verification in CI
- Performance regression detection
- Security scanning integration

SUCCESS CRITERIA:
- CI catches all classes of regressions
- Performance tracked over time
- Security vulnerabilities detected automatically
- Build artifacts properly collected and deployed

CI/CD modernization complete - ready for final integration.
```

### Step 16: Performance Optimization and Profiling

```
You are continuing the PVM architecture modernization from Step 15.

CONTEXT: All architectural components are implemented. Now optimize performance across the entire system.

TASK: Comprehensive performance optimization and profiling of the modernized PVM architecture.

REQUIREMENTS:
1. Profile all major components (scanner, parser, binder, checker) for performance bottlenecks
2. Optimize memory allocation patterns and data structures
3. Implement caching strategies for expensive operations
4. Add performance monitoring throughout the pipeline
5. Benchmark against original PVM implementation
6. Optimize for common use cases while maintaining correctness
7. Create performance regression testing framework

OPTIMIZATION TARGETS:
- Parsing performance: Match or exceed baseline
- Type checking: Leverage symbol pre-resolution for speed
- LSP responsiveness: <100ms for common operations
- Memory usage: Efficient allocation patterns
- Startup time: Fast initialization for CLI tools

TECHNICAL DETAILS:
- Use Go profiling tools (pprof) for bottleneck identification
- Optimize data structures for access patterns
- Implement object pooling for frequently allocated objects
- Add performance counters and monitoring

DELIVERABLES:
- Performance optimization report with before/after benchmarks
- Optimized implementations of critical path components
- Performance monitoring infrastructure
- Regression testing for performance

SUCCESS CRITERIA:
- Overall performance matches or exceeds original implementation
- Memory usage optimized for large codebases
- LSP performance meets responsiveness targets
- Performance regressions detected and prevented

System performance optimized for production use.
```

### Step 17: Integration Testing and Validation

```
You are continuing the PVM architecture modernization from Step 16.

CONTEXT: Performance optimization is complete. Now validate the entire modernized system through comprehensive integration testing.

TASK: Comprehensive end-to-end testing and validation of the complete modernized PVM system.

REQUIREMENTS:
1. Create comprehensive integration test suite covering all major workflows
2. Validate backward compatibility with existing PVM usage patterns
3. Test all components working together (PSC, PVI, PVX, PVM with new architecture)
4. Validate LSP integration with real editors and development environments
5. Test performance with large, realistic Perl codebases
6. Validate error handling and edge cases throughout the pipeline
7. Create migration testing for users upgrading from old PVM

INTEGRATION SCENARIOS:
- Complete typed-Perl project development workflow
- Legacy Perl project migration and type checking
- CI/CD pipeline integration with new build system
- LSP usage in popular editors (VS Code, vim, emacs)
- Cross-platform compatibility (Linux, macOS, Windows)
- Large codebase performance and memory usage

TECHNICAL DETAILS:
- Test with realistic Perl projects of varying sizes
- Validate all features work with new architecture
- Ensure no regressions in functionality or compatibility
- Test upgrade path from existing PVM installations

DELIVERABLES:
- Comprehensive integration test suite
- Backward compatibility validation report
- Performance validation with large codebases
- Migration testing results

SUCCESS CRITERIA:
- All existing PVM functionality preserved
- New features work reliably in real-world scenarios
- Performance targets met with realistic workloads
- Upgrade path validated and documented

System validated and ready for production deployment.
```

### Step 18: Documentation and Migration Guide

```
You are completing the PVM architecture modernization from Step 17.

CONTEXT: The modernized PVM system is validated and ready. Now create comprehensive documentation and migration guidance.

TASK: Create documentation for the modernized architecture and provide migration guidance for users and contributors.

REQUIREMENTS:
1. Update architecture documentation reflecting TypeScript-Go integration
2. Create developer guide for working with new compiler pipeline
3. Document new LSP capabilities and usage
4. Create migration guide for existing PVM users
5. Document new build system capabilities and usage
6. Create contributor guide for new architecture
7. Update all existing documentation to reflect architectural changes

DOCUMENTATION SECTIONS:
- Architecture Overview: New pipeline and component separation
- Developer Guide: Working with scanner, binder, checker components
- LSP Guide: Enhanced features and performance characteristics
- Build System: New capabilities, code generation, testing
- Migration Guide: Upgrading from old PVM versions
- Performance Guide: Optimization techniques and monitoring
- Troubleshooting: Common issues and solutions

TECHNICAL DETAILS:
- Include architectural diagrams showing new pipeline
- Provide code examples for common development tasks
- Document performance characteristics and targets
- Include troubleshooting guide for common issues

DELIVERABLES:
- Complete architectural documentation
- Developer and user migration guides
- Updated contributor documentation
- Performance and troubleshooting guides

SUCCESS CRITERIA:
- Documentation enables successful adoption of modernized PVM
- Migration path clear for existing users
- Contributors can effectively work with new architecture
- Performance characteristics well documented

PVM architecture modernization project complete with full documentation.
```

---

## Phase Planning Prompts

### Generate Updated Plan for Phase 2: Symbol Binding (Steps 5-8)

```
You are continuing the PVM architecture modernization after completing Phase 1 (Foundation Architecture).

CONTEXT: Phase 1 is complete with scanner extraction, AST consolidation, pipeline integration, and performance validation. The foundation is stable and ready for symbol binding implementation.

TASK: Generate a detailed updated plan for Phase 2: Symbol Binding implementation based on current codebase state and lessons learned from Phase 1.

REQUIREMENTS:
1. Analyze current codebase state after Phase 1 completion
2. Review existing symbol resolution logic in typechecker and identify integration points
3. Update Step 5-8 prompts based on actual Phase 1 implementation details
4. Identify any new challenges or opportunities discovered during Phase 1
5. Provide specific guidance for Perl scoping semantics implementation
6. Update success criteria based on Phase 1 performance characteristics
7. Create transition plan from current type checker to symbol-aware architecture

ANALYSIS NEEDED:
- Current symbol resolution patterns in internal/typechecker/
- Integration points with consolidated AST from Phase 1
- Performance implications of symbol binding phase
- Perl-specific scoping challenges (lexical, dynamic, package)
- LSP integration requirements for symbol-aware features

DELIVERABLES:
- Updated Step 5: Symbol Binding Architecture Design prompt
- Updated Step 6: Advanced Symbol Binding Features prompt
- Updated Step 7: Type Checker Integration prompt
- Updated Step 8: Enhanced Error Reporting prompt
- Transition strategy from Phase 1 foundation
- Risk assessment and mitigation strategies

SUCCESS CRITERIA:
- Prompts reflect actual codebase state after Phase 1
- Clear integration path with Phase 1 foundation
- Realistic timeline and complexity estimates
- Perl scoping semantics properly addressed
- Performance targets aligned with Phase 1 baseline

Generate prompts that build effectively on Phase 1 foundation.
```

### Generate Updated Plan for Phase 3: LSP Enhancement (Steps 9-11)

```
You are continuing the PVM architecture modernization planning after Phase 2: Symbol Binding.

CONTEXT: Phase 1 (Foundation) and Phase 2 (Symbol Binding) are planned. Now generate updated prompts for Phase 3: LSP Enhancement based on the symbol-aware architecture.

TASK: Generate detailed updated plan for Phase 3: LSP Enhancement that leverages symbol binding and AST navigation from previous phases.

REQUIREMENTS:
1. Analyze current LSP implementation in internal/lsp/
2. Design separation strategy into language service + protocol handler
3. Plan integration with symbol tables from Phase 2
4. Leverage AST navigation utilities from Phase 1
5. Update LSP feature enhancement priorities based on symbol capabilities
6. Plan performance optimization strategy for symbol-aware LSP
7. Design testing framework for enhanced LSP features

ENHANCED FEATURES TO PLAN:
- Symbol-aware goto definition and find references
- Cross-module symbol resolution for LSP
- Real-time symbol-based completions
- Rename operations with scope awareness
- Workspace symbol search functionality
- Document outline based on symbol tables
- Hover information with symbol context

DELIVERABLES:
- Updated Step 9: LSP Architecture Separation prompt
- Updated Step 10: Enhanced LSP Features prompt
- Updated Step 11: LSP Performance Optimization prompt
- Integration strategy with Phase 2 symbol tables
- Performance targets for symbol-aware LSP operations
- Testing framework design for LSP enhancements

SUCCESS CRITERIA:
- LSP enhancements leverage symbol binding effectively
- Performance targets realistic for symbol-aware operations
- Clear separation of language service vs protocol concerns
- Integration path with existing editor configurations
- Comprehensive testing strategy for LSP functionality

Generate prompts that maximize the benefits of symbol-aware architecture.
```

### Generate Updated Plan for Phase 4: Build System Modernization (Steps 12-15)

```
You are continuing the PVM architecture modernization planning after Phase 3: LSP Enhancement.

CONTEXT: Phases 1-3 will establish modern compiler architecture with symbol binding and enhanced LSP. Now plan Phase 4: Build System Modernization to support the new architecture.

TASK: Generate detailed updated plan for Phase 4: Build System Modernization following TypeScript-Go patterns for code generation and testing.

REQUIREMENTS:
1. Analyze current Makefile and build system capabilities
2. Plan code generation infrastructure for repetitive patterns
3. Design baseline testing framework for regression prevention
4. Plan CI/CD integration with new architecture components
5. Design performance monitoring and regression detection
6. Plan tool dependency management and automation
7. Create development vs production build optimization strategy

BUILD SYSTEM ENHANCEMENTS TO PLAN:
- go generate directives for AST node methods, mocks, error codes
- Baseline testing framework for parser, binder, checker outputs
- Performance benchmarking and regression detection
- Enhanced CI/CD with comprehensive coverage reporting
- Tool dependency management (tree-sitter-cli, testing tools)
- Cross-platform build optimization
- Code generation verification in CI

DELIVERABLES:
- Updated Step 12: Build System Foundation prompt
- Updated Step 13: Code Generation Infrastructure prompt
- Updated Step 14: Baseline Testing Framework prompt
- Updated Step 15: CI/CD Integration prompt
- Tool dependency management strategy
- Performance monitoring framework design
- Code generation templates and scripts

SUCCESS CRITERIA:
- Build system supports new architecture components
- Code generation eliminates manual repetitive patterns
- Baseline testing prevents regressions effectively
- CI/CD catches all classes of issues automatically
- Development experience significantly improved
- Performance regressions detected and prevented

Generate prompts that establish production-quality build infrastructure.
```

### Generate Updated Plan for Phase 5: Integration and Optimization (Steps 16-18)

```
You are completing the PVM architecture modernization planning after Phase 4: Build System Modernization.

CONTEXT: Phases 1-4 will establish complete modern architecture with scanner/parser/binder/checker pipeline, symbol-aware LSP, and modern build system. Now plan Phase 5: Final integration and optimization.

TASK: Generate detailed updated plan for Phase 5: Integration and Optimization to bring the complete system to production readiness.

REQUIREMENTS:
1. Plan comprehensive performance optimization across all components
2. Design end-to-end integration testing framework
3. Plan backward compatibility validation strategy
4. Design migration guide and documentation framework
5. Plan production readiness validation and stress testing
6. Design performance monitoring and observability
7. Create comprehensive user and developer documentation

INTEGRATION AND OPTIMIZATION TO PLAN:
- Cross-component performance optimization and profiling
- Large codebase integration testing and validation
- Migration path testing for existing PVM users
- LSP integration testing with popular editors
- Performance stress testing and memory optimization
- Comprehensive documentation and migration guides
- Production deployment and monitoring strategy

DELIVERABLES:
- Updated Step 16: Performance Optimization prompt
- Updated Step 17: Integration Testing prompt
- Updated Step 18: Documentation and Migration prompt
- Production readiness validation framework
- Performance monitoring and observability design
- Migration testing strategy for existing users
- Comprehensive documentation framework

SUCCESS CRITERIA:
- System performance meets or exceeds all targets
- Integration testing validates real-world usage
- Migration path is smooth for existing users
- Documentation enables successful adoption
- Production monitoring ensures system health
- TypeScript-Go integration benefits fully realized

Generate prompts that ensure production-ready system with full TypeScript-Go benefits.
```

---

## Success Criteria for Complete Project

The PVM TypeScript-Go integration is successful when:

1. **Architecture Modernization**: Clean separation of compiler phases (Scanner → Parser → Binder → Checker)
2. **Performance**: Meets or exceeds baseline performance with potential for significant improvements
3. **Developer Experience**: Enhanced LSP features with symbol-aware functionality
4. **Error Quality**: Significantly improved error messages with symbol context
5. **Build System**: Modern code generation, testing, and CI/CD integration
6. **Backward Compatibility**: All existing functionality preserved during migration
7. **Documentation**: Comprehensive guides for users, developers, and contributors
8. **Production Ready**: Validated with real-world workloads and edge cases

The final result should be a modernized Perl toolchain that provides TypeScript-quality developer experience while maintaining Perl's flexibility and the unique advantages of PVM's typed-Perl approach.

## Key Benefits Achieved

- **8x Performance Potential**: Foundation for significant performance improvements
- **Enhanced LSP**: Symbol-aware goto definition, find references, rename, completions
- **Better Error Messages**: Context-aware diagnostics with symbol information
- **Modern Build System**: Code generation, baseline testing, performance monitoring
- **Clean Architecture**: Maintainable, extensible compiler pipeline
- **Production Quality**: Comprehensive testing, CI/CD, and performance validation

This modernization positions PVM as a best-in-class development tool for Perl with modern language server capabilities and development experience.
