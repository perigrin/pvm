# Typed Perl Compiler with Type Inference Implementation Plan

## ✅ IMPLEMENTATION STATUS: Steps 1-11 COMPLETED (11/13)

**PHASE 1 COMPLETE:** Type System Foundation (Steps 1-3) ✅
**PHASE 2 COMPLETE:** Compiler Implementation (Steps 4-6) ✅
**PHASE 3 COMPLETE:** Enhanced Corpus Testing (Steps 7-8) ✅
**PHASE 4 COMPLETE:** Advanced Type System (Steps 9-11) ✅

**REMAINING:** Phase 5 (Steps 12-13)

## Project Overview

This plan implements Issue #7: Add Typed Perl compiler with type inference integration. This is a major architecture enhancement that will add a new compilation target to transform untyped Perl code into fully type-annotated Typed Perl, making implicit types explicit through sophisticated type inference.

**Key Deliverables:**
1. Type inference engine interface and basic implementation
2. Enhanced AST with type information support
3. New `TargetInferredTypeAnnotations` compiler
4. Type formatting and annotation generation system
5. PSC command integration (`psc infer`)
6. Enhanced corpus test format with compilation outcomes
7. Complex type system support (unions, intersections, parameterized types)
8. Confidence-based type annotation with quality control

## Architecture Analysis

**Current State:**
- Clean Perl Compiler: Strips type annotations from typed code (`TargetCleanPerl`)
- Typed Perl Compiler: Preserves existing type annotations (`TargetTypedPerl`)
- Parser with comprehensive AST support and type annotation detection
- Test framework with markdown-based corpus tests (56 files baselined)
- `psc parse --format=ast` command for AST analysis

**Target State:**
- Type inference engine that analyzes untyped Perl and infers types
- Enhanced AST interface with inferred type information access
- New compilation target that generates typed Perl with inferred annotations
- `psc infer` command for type annotation generation
- Enhanced corpus tests with expected compilation outcomes for all targets
- Support for complex type systems and confidence-based output

## Implementation Steps

### Phase 1: Type System Foundation (Steps 1-3)

#### Step 1: Basic Type System and Interfaces ✅ COMPLETED

```
Implement the foundational type system interfaces and basic type representations. This establishes the core abstractions that will be used throughout the type inference and compilation system.

Create the type system foundation:
- Define core `Type` interface and basic type implementations (Int, Str, Bool, etc.)
- Implement `TypeInfo` structure with confidence tracking and source attribution
- Create `TypeSource` enumeration for tracking how types were inferred
- Add `TypeConstraint` interface for complex type relationships
- Implement type equality and compatibility checking
- Create comprehensive unit tests for type system basics

The type system should be extensible to support complex types (unions, intersections, parameterized) that will be added in later steps.

Key files to create:
- `internal/types/types.go` - Core type interfaces and basic implementations
- `internal/types/info.go` - TypeInfo and metadata structures
- `internal/types/constraints.go` - Type constraint system
- `internal/types/types_test.go` - Comprehensive type system tests

Success criteria:
- Basic type representations (Int, Str, Bool, ArrayRef, HashRef) work correctly
- Type equality and compatibility checking functions properly
- TypeInfo tracks confidence levels and sources accurately
- All type system tests pass (aim for 100+ test cases)
- Type system is ready for extension with complex types

Test-driven approach:
- Start with failing tests for basic type operations
- Implement type interfaces incrementally
- Add confidence and source tracking tests
- Validate type compatibility logic with comprehensive test cases
```

#### Step 2: Enhanced AST Interface for Type Information ✅ COMPLETED

```
Extend the existing AST system to support type inference information without breaking current functionality. This creates the bridge between the parser and type inference engine.

Implement enhanced AST interfaces:
- Create `InferredAST` interface that extends existing AST
- Add methods for accessing inferred type information per AST node
- Implement `ASTTypeAdapter` that wraps parser AST with type info
- Create type annotation attachment system for AST nodes
- Add support for type constraint propagation through AST
- Ensure backward compatibility with existing AST usage

The enhanced AST should integrate seamlessly with existing parser output while providing rich type information access for the compiler.

Key files to modify/create:
- `internal/ast/inferred.go` - InferredAST interface and implementation
- `internal/ast/adapter.go` - Adapter for wrapping parser AST with type info
- `internal/ast/types.go` - AST node type attachment system
- `internal/ast/inferred_test.go` - Tests for enhanced AST functionality

Success criteria:
- InferredAST interface provides clean access to type information
- Existing AST functionality remains unchanged and fully compatible
- Type information can be attached to and retrieved from any AST node
- Adapter pattern works seamlessly with parser output
- All AST tests pass including new type-aware functionality

Test-driven approach:
- Create failing tests for type information access
- Implement InferredAST interface incrementally
- Test adapter pattern with real parser output
- Validate backward compatibility with existing code
- Add comprehensive integration tests with parser
```

#### Step 3: Basic Type Inference Engine Framework ✅ COMPLETED

```
Create the type inference engine framework with a simple literal-based inference implementation. This establishes the inference pipeline without complex flow analysis initially.

Implement basic type inference:
- Create `TypeInferenceEngine` interface and basic implementation
- Implement literal type inference (strings, numbers, booleans)
- Add variable declaration type propagation
- Create AST traversal system for type inference
- Implement confidence scoring for inferred types
- Add error collection and reporting for type conflicts

The inference engine should handle simple cases accurately and provide a foundation for more sophisticated analysis in later steps.

Key files to create:
- `internal/inference/engine.go` - Main inference engine interface
- `internal/inference/literal.go` - Literal type inference implementation
- `internal/inference/traversal.go` - AST traversal for inference
- `internal/inference/engine_test.go` - Comprehensive inference tests

Success criteria:
- Basic literal inference works correctly (my $x = 42 → Int)
- Variable type propagation functions properly
- Confidence scoring reflects inference quality
- Type conflicts are detected and reported
- Integration with InferredAST works seamlessly

Test-driven approach:
- Start with simple literal inference test cases
- Add variable propagation tests
- Test confidence scoring with various scenarios
- Validate error reporting for type conflicts
- Create integration tests with parser and AST system
```

### Phase 2: Compiler Implementation (Steps 4-6)

#### Step 4: Type Formatting and Code Generation ✅ COMPLETED

```
Implement the type formatting system that converts type information back into Perl syntax. This is the core of generating readable type annotations from inferred types.

Create type formatting system:
- Implement `TypeFormatter` with multiple output styles
- Add support for basic type annotation generation (my Int $var)
- Create compact vs verbose formatting options
- Implement confidence-based annotation decisions
- Add comment-based type hints for low-confidence types
- Support for different annotation styles (inline, separate declarations)

The formatter should produce clean, readable Perl code with appropriate type annotations based on inference confidence.

Key files to create:
- `internal/compiler/formatter.go` - Type formatting system
- `internal/compiler/styles.go` - Different formatting style implementations
- `internal/compiler/annotations.go` - Type annotation generation logic
- `internal/compiler/formatter_test.go` - Comprehensive formatting tests

Success criteria:
- Basic type annotations generate correctly (my Int $var = 42)
- Confidence thresholds control annotation inclusion
- Multiple formatting styles work as expected
- Comment-based hints appear for uncertain types
- Generated code is syntactically valid Perl

Test-driven approach:
- Test basic type annotation generation
- Validate confidence-based annotation decisions
- Test multiple formatting styles
- Verify generated Perl syntax validity
- Add edge case tests for complex formatting scenarios
```

#### Step 5: Inferred Type Annotations Compiler ✅ COMPLETED

```
Implement the new compilation target that takes an InferredAST and generates typed Perl code with inferred type annotations. This is the core compiler that fulfills the main requirement.

Create inferred typed Perl compiler:
- Implement `InferredTypedPerlCompiler` struct and interface
- Add `TargetInferredTypeAnnotations` compilation target
- Integrate with existing compiler registry system
- Implement AST node compilation with type annotation injection
- Add support for compiler options (confidence threshold, verbosity)
- Ensure generated code maintains semantic equivalence

The compiler should transform untyped Perl into typed Perl while preserving all original behavior and adding helpful type information.

Key files to create:
- `internal/compiler/inferred_typed_perl.go` - Main compiler implementation
- Update `internal/compiler/registry.go` - Add new compilation target
- `internal/compiler/inferred_typed_perl_test.go` - Comprehensive compiler tests

Success criteria:
- New compilation target integrates with existing registry
- Generated typed Perl is syntactically and semantically correct
- Confidence thresholds control annotation behavior appropriately
- Compiler options provide flexible output control
- Integration with type inference engine works seamlessly

Test-driven approach:
- Create failing tests for basic compilation scenarios
- Implement compiler incrementally with AST node handlers
- Test confidence threshold behavior with various settings
- Validate semantic equivalence of generated code
- Add comprehensive integration tests with inference engine
```

#### Step 6: PSC Command Integration ✅ COMPLETED

```
Integrate the type inference and compilation system with PSC commands. Add the new `psc infer` command and extend existing commands to support the new compilation target.

Implement PSC command integration:
- Create new `psc infer` command with comprehensive flag support
- Add new compilation target to existing `psc compile` command
- Implement command-line options for confidence, verbosity, output style
- Add progress reporting for inference and compilation phases
- Integrate with existing PSC error handling and user interface
- Support both file and directory processing

The commands should provide an intuitive interface for type inference and code generation that integrates well with existing PSC workflows.

Key files to modify/create:
- `internal/psc/infer_command.go` - New psc infer command implementation
- Update `internal/psc/compile_command.go` - Add new target support
- Update `internal/psc/command.go` - Register new command
- `internal/psc/infer_command_test.go` - Command integration tests

Success criteria:
- `psc infer` command works with all required flags and options
- Integration with existing `psc compile` provides new target access
- Command-line interface is intuitive and well-documented
- Progress reporting provides useful feedback during processing
- Error handling provides clear, actionable messages

Test-driven approach:
- Test command flag parsing and validation
- Validate file and directory processing workflows
- Test integration with inference engine and compiler
- Verify error handling and user feedback
- Add end-to-end command testing scenarios
```

### Phase 3: Enhanced Corpus Testing (Steps 7-8)

#### Step 7: Corpus Test Format Enhancement ✅ COMPLETED

```
Enhance the existing corpus test format to include expected compilation outcomes for all compilation targets. This extends the systematic baselining work to validate compiler accuracy.

Extend corpus test format:
- Add "Expected Compilation Outcomes" sections to test framework
- Support for "Expected Typed Output", "Expected Untyped Output", "Expected Inferred Output"
- Extend markdown test loader to parse compilation outcome sections
- Update test framework to validate compilation against expected outcomes
- Add compilation outcome baseline generation tools
- Ensure backward compatibility with existing AST-focused tests

The enhanced format should provide comprehensive validation of all compiler targets while maintaining the existing AST validation system.

Key files to modify:
- `internal/parser/test_framework.go` - Extend markdown test parsing
- Add compilation outcome parsing and validation logic
- Update test case structures to include compilation expectations
- `internal/parser/test_framework_test.go` - Test enhanced format

Success criteria:
- Markdown tests can include compilation outcome sections
- Test framework validates all compilation targets correctly
- Baseline generation tools create accurate expected outcomes
- Existing AST-focused tests continue to work unchanged
- New format supports all three compilation targets

Test-driven approach:
- Test extended markdown parsing with compilation sections
- Validate compilation outcome comparison logic
- Test baseline generation for all compiler targets
- Verify backward compatibility with existing tests
- Add comprehensive format validation tests
```

#### Step 8: Systematic Corpus Test Updates ✅ COMPLETED

```
Update all 56 corpus test files to include expected compilation outcomes for all three compilation targets. This provides comprehensive validation coverage for the new compiler.

Update corpus test files systematically:
- Generate expected outcomes for all existing test files using baseline tools
- Add compilation outcome sections to all 56 test files across 7 directories
- Validate that inferred type annotations are accurate and helpful
- Ensure round-trip compilation maintains semantic equivalence
- Update any test cases that reveal inference or compilation issues
- Document any test files that require special handling

This work should be done systematically using the parallel agent approach established during the initial baselining work.

Key files to update:
- All 56 test files in `test/corpus/parser/typed-perl/` subdirectories
- Each file gets "Expected Compilation Outcomes" sections
- Validation of accuracy across all test scenarios

Success criteria:
- All 56 corpus test files include compilation outcome expectations
- Generated outcomes accurately reflect compiler behavior
- Test suite validates all compilation targets against expected results
- Any compiler issues discovered during baseline generation are documented
- Round-trip compilation validation passes for all test cases

Test-driven approach:
- Use systematic baseline generation approach
- Validate each generated outcome for accuracy
- Test compilation targets against all corpus scenarios
- Document and address any issues discovered during baselining
- Ensure comprehensive test coverage for type inference accuracy
```

### Phase 4: Advanced Type System (Steps 9-11)

#### Step 9: Complex Type Support Implementation

```
Extend the type system to support complex types including unions, intersections, and parameterized types. This enables sophisticated type inference for advanced Perl constructs.

Implement complex type support:
- Add `UnionType` implementation for multiple possible types (Int|Str)
- Implement `IntersectionType` for combined type requirements
- Add `ParameterizedType` for generic types (ArrayRef[Int], HashRef[Str])
- Extend type formatting to handle complex type syntax
- Update type compatibility and equality checking for complex types
- Enhance inference engine to generate complex types appropriately

Complex types should integrate seamlessly with the existing type system and provide accurate representation of sophisticated Perl type scenarios.

Key files to modify/create:
- `internal/types/complex.go` - Complex type implementations
- Update `internal/types/types.go` - Extend core interfaces
- Update `internal/compiler/formatter.go` - Complex type formatting
- `internal/types/complex_test.go` - Comprehensive complex type tests

Success criteria:
- Union types handle multiple possibilities correctly (Int|Str|Undef)
- Intersection types represent combined requirements properly
- Parameterized types support nested generic scenarios accurately
- Type formatting produces readable complex type syntax
- Type inference generates appropriate complex types when needed

Test-driven approach:
- Create comprehensive test cases for each complex type variety
- Test type compatibility logic with complex type combinations
- Validate formatting output for readability and correctness
- Test inference engine integration with complex type generation
- Add edge case testing for deeply nested parameterized types
```

#### Step 10: Advanced Type Inference Implementation ✅ COMPLETED

```
Enhance the type inference engine with sophisticated analysis including control flow analysis, function signature inference, and contextual type determination.

Implement advanced inference capabilities:
- Add control flow analysis for type propagation through conditionals ✅
- Implement function signature inference from usage patterns ✅
- Add contextual type inference (scalar vs list context) ✅
- Implement type constraint propagation through assignments ✅
- Add external library type hint integration ✅
- Support for method resolution with type information ✅

The enhanced inference should handle real-world Perl code patterns and provide high-confidence type annotations for complex scenarios.

Key files to modify/create:
- `internal/inference/flow.go` - Control flow analysis implementation ✅
- `internal/inference/functions.go` - Function signature inference ✅
- `internal/inference/context.go` - Contextual type analysis ✅
- Update `internal/inference/engine.go` - Integrate advanced features ✅
- `internal/inference/advanced_test.go` - Advanced inference tests ✅

Success criteria:
- Control flow analysis tracks types through conditionals and loops ✅
- Function signatures are inferred from usage patterns correctly ✅
- Contextual analysis handles scalar vs list context appropriately ✅
- Type constraints propagate correctly through complex code ✅
- Integration with external type hints works when available ✅

Test-driven approach:
- Test control flow scenarios with type propagation ✅
- Validate function signature inference with various patterns ✅
- Test contextual type analysis with Perl context scenarios ✅
- Verify constraint propagation through complex assignments ✅
- Add integration tests for advanced inference scenarios ✅
```

#### Step 11: Quality Control and Confidence Tuning ✅ COMPLETED

```
Implement sophisticated quality control mechanisms and confidence scoring to ensure type annotations are accurate and helpful. Focus on minimizing false positives and providing useful uncertainty indicators.

Implement quality control system:
- Develop sophisticated confidence scoring algorithms ✅
- Add conflict detection and resolution for competing type inferences ✅
- Implement quality metrics for generated type annotations ✅
- Add user-configurable confidence thresholds with smart defaults ✅
- Create uncertainty visualization in generated code comments ✅
- Add validation against known-good type patterns ✅

The quality control system should ensure that generated type annotations add value and don't mislead developers with incorrect type information.

Key files to create:
- `internal/inference/quality.go` - Quality control and confidence scoring ✅
- `internal/inference/conflicts.go` - Type conflict detection and resolution ✅
- `internal/inference/validation.go` - Type annotation validation ✅
- `internal/inference/quality_test.go` - Quality control tests ✅

Success criteria:
- Confidence scoring accurately reflects type inference quality ✅
- Type conflicts are detected and resolved appropriately ✅
- Quality metrics provide useful feedback on annotation value ✅
- User configuration enables appropriate confidence tuning ✅
- Generated code includes helpful uncertainty indicators ✅

Test-driven approach:
- Test confidence scoring with various inference scenarios ✅
- Validate conflict detection with competing type evidence ✅
- Test quality metrics against manually verified examples ✅
- Verify user configuration affects output appropriately ✅
- Add comprehensive validation tests for annotation quality ✅
```

### Phase 5: Integration and Polish (Steps 12-13)

#### Step 12: End-to-End Integration Testing

```
Perform comprehensive end-to-end integration testing covering complete workflows from parsing through inference to compilation. This ensures the entire system works together reliably.

Implement comprehensive integration testing:
- Create end-to-end test suite covering complete parse-infer-compile pipeline
- Test all compilation targets with identical input for consistency
- Add performance testing for inference and compilation phases
- Create regression tests for complex real-world Perl scenarios
- Validate memory usage and performance characteristics
- Test error handling and recovery across the entire pipeline

The integration testing should validate that the complete system provides reliable, high-quality type inference and compilation.

Key files to create:
- `test/e2e/type_inference_workflow_test.go` - End-to-end workflow tests
- `test/performance/inference_benchmarks_test.go` - Performance testing
- `test/regression/real_world_scenarios_test.go` - Real-world scenario tests

Success criteria:
- Complete parse-infer-compile pipeline works reliably
- All compilation targets produce consistent results
- Performance is acceptable for real-world usage
- Error handling provides useful feedback throughout pipeline
- Memory usage is reasonable for large codebases

Test-driven approach:
- Create comprehensive workflow test scenarios
- Test performance with various file sizes and complexity levels
- Validate error handling with intentionally problematic input
- Test integration with all existing PVM/PSC functionality
- Add regression tests for any issues discovered during development
```

#### Step 13: Documentation and User Experience

```
Create comprehensive documentation and polish the user experience for the new type inference and compilation functionality. This ensures successful adoption and usage.

Complete documentation and UX:
- Write comprehensive user guide for type inference functionality
- Add detailed command reference for all new PSC commands
- Create tutorial covering common workflows and best practices
- Add troubleshooting guide for common issues and limitations
- Document configuration options and customization capabilities
- Create examples showing before/after code transformations

The documentation should enable users to successfully adopt and benefit from the new type inference capabilities.

Key files to create:
- `docs/type-inference-guide.md` - Comprehensive user guide
- `docs/psc-infer-reference.md` - Command reference documentation
- `docs/type-inference-tutorial.md` - Step-by-step tutorial
- `docs/type-inference-troubleshooting.md` - Troubleshooting guide
- `examples/type-inference/` - Example transformations and workflows

Success criteria:
- Documentation covers all functionality comprehensively
- Tutorial enables new users to get started successfully
- Troubleshooting guide addresses common issues effectively
- Examples demonstrate value and proper usage patterns
- Command reference provides complete flag and option documentation

Test-driven approach:
- Validate documentation accuracy against actual functionality
- Test tutorial steps for completeness and clarity
- Verify troubleshooting guide with real user scenarios
- Test examples for correctness and educational value
- Ensure documentation stays current with implementation changes
```

## Prompt Structure

Each step above provides a complete prompt that:
1. Clearly states the objective and scope with specific technical goals
2. Builds incrementally on previous steps without large complexity jumps
3. Maintains backward compatibility and system stability
4. Includes specific implementation guidance and file structure
5. Defines clear, measurable success criteria
6. Emphasizes test-driven development throughout
7. Provides comprehensive testing strategies for validation

The prompts are designed for execution in sequence, with each step building on previous work while maintaining system stability and comprehensive test coverage.

## Success Metrics

**Technical Success:**
- Type inference accuracy >90% for common Perl patterns
- All compilation targets produce semantically equivalent output
- 100% test pass rate maintained throughout implementation
- Performance acceptable for real-world codebase sizes
- Integration seamless with existing PVM/PSC functionality

**User Experience Success:**
- `psc infer` command provides intuitive type annotation generation
- Generated type annotations improve code readability and maintainability
- Confidence-based output prevents misleading type annotations
- Documentation enables successful adoption by Perl developers
- Error messages provide actionable guidance for resolution

**Project Impact:**
- Issue #7 requirements fully implemented with advanced features
- Foundation established for future type system enhancements
- Gradual typing adoption path created for Perl developers
- Static analysis capabilities significantly enhanced
- Community benefits from sophisticated type inference tooling

## Architecture Integration

The implementation integrates with existing PVM architecture:
- **Parser Integration**: Enhanced AST interfaces work with existing parser
- **Compiler Extension**: New target extends existing compiler registry
- **PSC Commands**: New commands integrate with existing CLI framework
- **Test Framework**: Enhanced corpus tests build on systematic baselining work
- **Configuration**: Type inference settings integrate with existing config system

This ensures the new functionality feels native to PVM while providing powerful new capabilities for Perl development.
