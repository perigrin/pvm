# Parser 100% Completion - Test-Driven Implementation Plan

## Project Context

This plan implements a systematic approach to achieve 100% parser test completion in the PVM (Perl Version Manager) project. Currently at 91.9% pass rate (73/899 failures), the goal is to eliminate all remaining parser test failures through incremental, test-driven development.

## Overall Blueprint

### Current State Analysis
- **Parser Status**: 73 failures out of 899 tests (91.9% pass rate)
- **Main Issue Categories**:
  1. Test expectation mismatches (40+ failures) - Parser works better than tests expect
  2. Complex type expression parsing (10+ failures) - Advanced nested types
  3. Grammar missing features (15+ failures) - `given/when`, edge cases
  4. Class/role declarations (8+ failures) - Major feature missing

### Target Architecture
- **Goal**: 100% parser test pass rate (0/899 failures)
- **Strategy**: Incremental improvements with comprehensive testing at each step
- **Approach**: Start with quick wins, progress to complex features
- **Validation**: Each step must maintain existing functionality while adding new capabilities

---

## Phase 1: Test Expectation Corrections (Quick Wins)

**Target**: Reduce 73 failures to ~35 failures (~95.5% pass rate)
**Strategy**: Fix test expectations that no longer match improved parser capabilities

### Step 1: Audit Simple Type Annotation Test Expectations

```
You are fixing parser test expectations that are now incorrect due to improved type annotation extraction.

CONTEXT: The PVM parser recently had a major breakthrough in type annotation extraction. Previously failing type annotation syntax now parses correctly, but many tests still expect these patterns to fail. This creates false negative test results.

PROBLEM: Multiple test files in `/internal/parser/testdata/typed-perl/` have `<!-- should_error: true -->` expectations for syntax that now correctly parses. These tests report "Expected error but parsing succeeded" failures.

TASK: Update test expectations for simple type annotations to reflect successful parsing.

TARGET TESTS:
- simple-annotations_basic_typed_variables
- simple-annotations_complex_assignments
- simple-annotations_mixed_typed_untyped
- simple-annotations_scoping_keywords
- simple-annotations_typed_arrays_hashes
- simple-annotations_whitespace_variations

REQUIREMENTS:
1. Examine the test content in `internal/parser/testdata/typed-perl/simple-annotations.md`
2. For each test case that shows "Expected error but parsing succeeded":
   - Verify the syntax is actually valid typed Perl
   - Remove `<!-- should_error: true -->` directive
   - Remove `<!-- expected_error: ... -->` directive
   - Ensure the test now expects successful parsing
3. Run targeted tests to verify fixes: `go test -v ./internal/parser -run TestRunMarkdownTestsByCategory`
4. Confirm error count reduction without introducing new failures

TECHNICAL DETAILS:
- Focus on basic patterns like `my Int $var = 42;`
- Include array and hash type annotations: `my ArrayRef[Str] @items;`
- Handle scoping keywords with types: `our Num $global;`
- Preserve complex cases that should still fail

EXPECTED OUTCOME:
- 6+ failing tests converted to passing
- No new test failures introduced
- Clear reduction in parser failure count
- Foundation for more complex type expression fixes

VALIDATION:
- Run full parser test suite after changes
- Verify overall failure count decreases
- Ensure no regressions in previously passing tests
```

### Step 2: Update Union Type Test Expectations

```
You are continuing the test expectation fixes from Step 1, focusing on union type syntax.

CONTEXT: Union type parsing has been significantly improved, but test expectations haven't been updated to reflect these improvements. Many union type patterns that now parse correctly still have error expectations.

TASK: Update test expectations for union type annotations to reflect successful parsing.

TARGET TESTS:
- union-types_simple_union_types
- union-types_custom_types_unions
- union-types_multi_way_unions
- Any additional union type tests showing "Expected error but parsing succeeded"

REQUIREMENTS:
1. Examine `internal/parser/testdata/typed-perl/union-types.md`
2. For each union type test showing success when error expected:
   - Verify syntax like `Int|Str` is valid
   - Remove error expectations for valid union syntax
   - Keep error expectations for truly invalid syntax (like `Int||Str`)
3. Test specifically: `go test -v ./internal/parser -run "TestRunMarkdownTestsByCategory.*union-types"`
4. Validate that basic union types like `my Int|Str $flexible;` now pass

TECHNICAL DETAILS:
- Valid patterns: `Int|Str`, `ArrayRef[Int]|HashRef[Str]`, `CustomType|Undef`
- Invalid patterns should still error: `Int||Str`, `Int|`, `|Str`
- Nested union contexts may need special handling
- Multi-way unions: `Type1|Type2|Type3`

EXPECTED OUTCOME:
- 4+ additional failing tests converted to passing
- Union type parsing validation confirmed working
- Clear progress toward target failure reduction

VALIDATION:
- Targeted union type test execution
- Verify union type extraction in AST
- Ensure complex union expressions still work correctly
```

### Step 3: Fix Parameterized Type Test Expectations

```
You are continuing test expectation fixes from Step 2, focusing on parameterized type syntax.

CONTEXT: Parameterized type parsing (like `ArrayRef[Int]`, `HashRef[Str]`) has been working correctly, but test expectations may not reflect this. These are fundamental to typed Perl support.

TASK: Update test expectations for parameterized type annotations.

TARGET TESTS:
- parameterized-types_custom_parameterized
- parameterized-types_method_signatures (if expectation mismatch)
- Any other parameterized type tests showing unexpected success

REQUIREMENTS:
1. Review `internal/parser/testdata/typed-perl/parameterized-types.md`
2. Identify parameterized type patterns that should parse successfully
3. Update expectations for valid syntax like:
   - `ArrayRef[Int]`
   - `HashRef[String]`
   - `CodeRef[Str, Bool]`
   - `Optional[MyType]`
4. Keep error expectations for malformed syntax
5. Test: `go test -v ./internal/parser -run "TestRunMarkdownTestsByCategory.*parameterized-types"`

TECHNICAL DETAILS:
- Single parameter: `ArrayRef[Int]`
- Multiple parameters: `CodeRef[Str, Bool]`
- Nested parameterized: `ArrayRef[HashRef[Int]]`
- Custom parameterized types: `Container[MyType]`
- Invalid: missing brackets, unmatched brackets

EXPECTED OUTCOME:
- 2-3 additional tests converted from failing to passing
- Parameterized type parsing confirmed working
- Foundation established for complex nested types

VALIDATION:
- Test parameterized type extraction in AST
- Verify nested parameterized types work
- Ensure type parameter information preserved
```

### Step 4: Validate Phase 1 Results and Measure Impact

```
You are completing Phase 1 of the parser improvement plan by validating all test expectation fixes.

CONTEXT: Steps 1-3 have updated test expectations for simple annotations, union types, and parameterized types. This should result in a significant reduction in parser test failures.

TASK: Validate the cumulative impact of Phase 1 changes and ensure no regressions.

REQUIREMENTS:
1. Run the complete parser test suite: `go test -v ./internal/parser -count=1`
2. Count remaining failures and calculate improvement from baseline of 73 failures
3. Categorize remaining failures by type to plan Phase 2
4. Verify no new failures were introduced in previously passing tests
5. Document the success rate improvement

VALIDATION STEPS:
- Full test suite execution with failure analysis
- Comparison to baseline: 73 failures → target ~35 failures
- Categorization of remaining failure types
- Performance regression check (ensure no significant slowdown)
- Sanity test: verify basic type annotation extraction still works

ANALYSIS REQUIRED:
- Calculate exact improvement: old_failures - new_failures
- Identify dominant remaining failure categories
- Assess readiness for Phase 2 (complex type expressions)
- Document any unexpected issues or patterns

SUCCESS CRITERIA:
- Major reduction in test failures (target: 50%+ improvement)
- No new regressions introduced
- Clear categorization of remaining issues
- Confirmed foundation for Phase 2 work

EXPECTED OUTCOME:
- ~40 failures eliminated through expectation fixes
- Parser failure rate reduced from 73 to ~35 (95%+ pass rate achieved)
- Clear roadmap for remaining work established
```

---

## Phase 2: Complex Type Expression Enhancements

**Target**: Reduce ~35 failures to ~25 failures (~97% pass rate)
**Strategy**: Improve parsing of advanced type expressions and method signatures

### Step 5: Enhance Complex Method Signature Parsing

```
You are implementing Phase 2 of the parser improvement plan, focusing on complex method signature parsing.

CONTEXT: Phase 1 successfully reduced parser failures from 73 to ~35 through test expectation fixes. Phase 2 targets remaining parsing issues with complex type expressions, particularly in method signatures.

PROBLEM: Tests like `complex-types_complex_method_signatures` and `methods-fields_*` show "UnknownTypeError: Syntax error in type expression" for advanced method signatures with complex parameter and return types.

TASK: Improve the parser's ability to handle complex method signatures with advanced type expressions.

TARGET FAILURES:
- complex-types_complex_method_signatures
- parameterized-types_method_signatures
- Any method signature parsing failures

REQUIREMENTS:
1. Examine the failing test cases in `internal/parser/testdata/typed-perl/complex-types.md`
2. Identify specific method signature patterns causing parsing failures
3. Enhance type expression parsing in `internal/parser/treesitter/perl.go` to handle:
   - Complex parameter types: `ArrayRef[HashRef[Int|Str]] $data`
   - Complex return types: `-> Result[Map[Str, Array[Item]], ProcessingError>`
   - Multiple complex parameters in method signatures
   - Optional and slurpy parameters with complex types

IMPLEMENTATION APPROACH:
1. Add test-driven development: write failing tests for specific signature patterns
2. Enhance `processMethodSignature()` function to handle complex types
3. Improve type expression recursion for deeply nested types
4. Add support for method return type parsing with complex expressions
5. Ensure backward compatibility with existing simple signatures

EXPECTED OUTCOME:
- 3-5 method signature tests converted from failing to passing
- Improved support for production-ready typed method signatures
- Foundation for advanced type system features

VALIDATION:
- Test complex method signatures individually
- Verify return type parsing accuracy
- Ensure simple method signatures still work
- Performance check for nested type expression parsing
```

### Step 6: Fix Nested Union Type Context Parsing

```
You are continuing Phase 2 improvements, focusing on union types in nested contexts.

CONTEXT: Basic union types work from Phase 1, but complex union type expressions in nested contexts (like method parameters, return types, and data structures) still fail with parsing errors.

PROBLEM: Tests like `union-types_nested_contexts` show "UnknownTypeError: Syntax error in type expression" when union types appear in complex nested expressions.

TASK: Enhance union type parsing to handle nested contexts correctly.

TARGET FAILURES:
- union-types_nested_contexts
- Any complex type expressions involving unions in parameters/returns

REQUIREMENTS:
1. Examine failing nested union type patterns
2. Enhance union type parsing to handle:
   - Union types in method parameters: `method func((Int|Str) $param)`
   - Union types in return types: `-> (Success|Error)`
   - Union types in parameterized types: `ArrayRef[Int|Str]`
   - Parenthesized union expressions: `(Type1|Type2)`
   - Union types in complex data structures

IMPLEMENTATION:
1. Write targeted tests for nested union type contexts
2. Improve union type parsing logic in type expression handlers
3. Add support for parenthesized union expressions
4. Handle precedence correctly for union vs parameterized types
5. Ensure proper AST generation for nested union types

TECHNICAL DETAILS:
- Parse `(Int|Str)` as parenthesized union expression
- Handle `ArrayRef[Int|Str]` as parameterized type with union parameter
- Support `method func(Int|Str $x) -> Bool|Error`
- Maintain proper type information in AST for complex unions

EXPECTED OUTCOME:
- 2-3 nested union type tests converted to passing
- Robust union type support in all expression contexts
- Improved type checking foundation for complex expressions

VALIDATION:
- Test union types in various nested contexts
- Verify AST contains correct union type information
- Ensure simple union types still work correctly
```

### Step 7: Implement Complex Type Assertion Support

```
You are continuing Phase 2, implementing support for complex type assertion expressions.

CONTEXT: Basic type assertions may work, but complex type assertions with parameterized types, unions, and nested expressions are failing.

PROBLEM: Tests like `complex-types_complex_type_assertions` fail with parsing errors when type assertions involve complex type expressions.

TASK: Enhance type assertion parsing to support complex type expressions.

TARGET FAILURES:
- complex-types_complex_type_assertions
- Any type assertion tests with complex types

REQUIREMENTS:
1. Identify failing type assertion patterns like:
   - `$data as ArrayRef[HashRef[Int|Str]]`
   - `$result as (Success|Error)&Detailed`
   - `$object as MyClass[T]`
2. Enhance type assertion parsing logic
3. Support type assertions in complex expressions and statement contexts
4. Ensure proper AST generation for complex type assertions

IMPLEMENTATION:
1. Write tests for complex type assertion patterns
2. Improve `processTypeAssertion()` function to handle complex types
3. Add support for type assertions with:
   - Parameterized types: `$var as ArrayRef[Int]`
   - Union types: `$var as (Int|Str)`
   - Intersection types: `$var as (Type1&Type2)`
   - Nested expressions: `$data->transform() as ProcessedData`

TECHNICAL DETAILS:
- Parse `as Type` syntax with complex type expressions
- Handle type assertions in method calls and complex expressions
- Maintain proper precedence for type assertion operators
- Generate accurate AST nodes for type checker integration

EXPECTED OUTCOME:
- 2+ type assertion tests converted to passing
- Complete type assertion support for production code
- Enhanced type safety capabilities for PVM users

VALIDATION:
- Test type assertions with various complex types
- Verify AST contains correct type assertion information
- Ensure type assertions work in method calls and expressions
```

### Step 8: Optimize Performance for Complex Type Expressions

```
You are completing Phase 2 by addressing performance issues with complex type expression parsing.

CONTEXT: Previous steps in Phase 2 added support for complex type expressions, but this may have introduced performance concerns for deeply nested or repeatedly used complex types.

PROBLEM: Tests like `complex-types_stress_testing` or performance-related failures may indicate that complex type parsing is too slow for production use.

TASK: Optimize parsing performance for complex type expressions while maintaining functionality.

TARGET ISSUES:
- Performance stress test failures
- Any timeouts or slowdowns in complex type parsing
- Memory usage issues with deeply nested types

REQUIREMENTS:
1. Profile current parsing performance for complex type expressions
2. Identify bottlenecks in type expression parsing logic
3. Implement optimizations:
   - Caching for repeated type expressions
   - Optimized recursion for nested types
   - Efficient AST node creation
   - Memory usage optimization
4. Ensure optimizations don't break functionality

IMPLEMENTATION:
1. Add performance benchmarks for complex type parsing
2. Profile parsing of deeply nested type expressions
3. Implement caching for commonly used type patterns
4. Optimize recursive type expression parsing algorithms
5. Add memory usage monitoring for complex type ASTs

TECHNICAL DETAILS:
- Benchmark parsing time for expressions like `ArrayRef[HashRef[Union[Type1|Type2]]]`
- Cache type expression AST nodes for reuse
- Optimize tree-sitter node traversal for type expressions
- Monitor memory allocation during complex type parsing

EXPECTED OUTCOME:
- Performance stress tests pass
- Complex type parsing meets production performance requirements
- Foundation established for Phase 3 grammar extensions

VALIDATION:
- Run performance benchmarks before/after optimizations
- Verify no functionality regressions
- Test parsing speed on large files with many complex types
- Memory usage stays within acceptable bounds
```

---

## Phase 3: Tree-sitter Grammar Extensions

**Target**: Reduce ~25 failures to ~10 failures (~98.5% pass rate)
**Strategy**: Add missing Perl language constructs to tree-sitter grammar

### Step 9: Implement given/when Control Flow Grammar Support

```
You are beginning Phase 3 of the parser improvement plan, adding missing Perl language constructs to the tree-sitter grammar.

CONTEXT: Phase 2 achieved ~97% pass rate by fixing complex type expressions. Phase 3 targets fundamental Perl language features missing from the tree-sitter grammar, particularly `given/when` control flow statements.

PROBLEM: Multiple control flow tests fail with "parse error (ERROR nodes detected)" because the tree-sitter-typed-perl grammar doesn't support Perl's `given/when` syntax (introduced in Perl 5.10).

TASK: Extend the tree-sitter-typed-perl grammar to support `given/when` control flow constructs.

TARGET FAILURES:
- control-flow_given_when_basic
- control-flow_given_when_arrays
- control-flow_given_when_complex_condition
- control-flow_given_when_nested
- control-flow_given_when_ranges
- control-flow_given_when_regex
- control-flow_given_no_default
- All other `given/when` related failures

REQUIREMENTS:
1. Study Perl's `given/when` syntax specification (Perl 5.10+)
2. Extend `tree-sitter-typed-perl/grammar.js` to support:
   - `given (EXPR) { ... }` statements
   - `when (PATTERN) { ... }` clauses
   - `default { ... }` clauses
   - Nested `given/when` constructs
   - Complex when patterns (arrays, ranges, regex)
3. Ensure grammar changes don't break existing functionality
4. Add comprehensive test coverage for new grammar rules

IMPLEMENTATION APPROACH:
1. First add basic `given/when` grammar rules using TDD:
   - Write test for simple `given ($var) { when (1) { ... } }`
   - Implement minimal grammar to pass the test
   - Iterate to add more complex patterns
2. Extend grammar incrementally:
   - Basic given statement structure
   - Simple when clauses with literal patterns
   - Complex when patterns (arrays, ranges, regex)
   - Default clause support
   - Nested given/when structures
3. Rebuild tree-sitter parser after each grammar change
4. Test against real Perl code using given/when

TECHNICAL DETAILS:
- Add to grammar.js: `given_statement`, `when_clause`, `default_clause`
- Handle when patterns: literals, arrays, ranges, regex, expressions
- Support break/continue semantics (fall-through control)
- Ensure proper precedence and associativity
- Maintain compatibility with existing control flow parsing

EXPECTED OUTCOME:
- 12+ given/when control flow tests converted to passing
- Foundation for complete Perl 5.10+ language support
- Robust control flow parsing for production typed Perl code

VALIDATION:
- Test all given/when patterns individually
- Verify nested given/when structures work
- Ensure no regression in existing control flow parsing
- Performance test for complex given/when structures
```

### Step 10: Handle Edge Case Variable Declarations

```
You are continuing Phase 3 grammar extensions, focusing on edge case variable declarations.

CONTEXT: The grammar improvements from Step 9 addressed control flow, but some variable declaration edge cases still cause parse errors.

PROBLEM: Tests like `variables_variable_edge_cases` fail with "unexpected token" errors for unusual but valid Perl variable declaration patterns.

TASK: Extend grammar support for edge case variable declarations and improve error recovery.

TARGET FAILURES:
- variables_variable_edge_cases
- Any variable declaration parsing errors

REQUIREMENTS:
1. Examine failing variable declaration patterns
2. Extend grammar to handle edge cases like:
   - Incomplete declarations: `my %;` (legal in Perl)
   - Unusual sigil patterns
   - Complex variable scope declarations
3. Improve error recovery for malformed variable declarations
4. Ensure backward compatibility with standard variable declarations

IMPLEMENTATION:
1. Write tests for specific edge case patterns that should parse
2. Extend variable declaration grammar rules in `grammar.js`
3. Add error recovery rules for incomplete declarations
4. Handle unusual but valid Perl variable declaration syntax
5. Test against edge cases from Perl specification

TECHNICAL DETAILS:
- Support incomplete variable declarations where allowed by Perl
- Handle unusual sigil combinations and patterns
- Add proper error recovery without breaking parser
- Maintain compatibility with typed variable declarations

EXPECTED OUTCOME:
- 1-2 variable edge case tests converted to passing
- More robust variable declaration parsing
- Better error recovery for malformed code

VALIDATION:
- Test edge case variable patterns
- Verify standard variable declarations still work
- Ensure typed variable declarations unaffected
```

### Step 11: Add Complex Control Flow Pattern Support

```
You are continuing Phase 3, adding support for complex control flow patterns beyond basic given/when.

CONTEXT: Basic given/when support from Step 9 covered simple cases, but complex control flow patterns like state machines and event loops still fail.

PROBLEM: Tests like `control-flow_state_machine_loop` and `control-flow_event_loop` fail with parse errors for complex nested control flow.

TASK: Extend grammar and parsing to support complex control flow patterns.

TARGET FAILURES:
- control-flow_state_machine_loop
- control-flow_event_loop
- control-flow_when_break
- control-flow_when_continue

REQUIREMENTS:
1. Analyze complex control flow patterns that are failing
2. Extend grammar to support:
   - Complex nested loop structures
   - State machine patterns with given/when
   - Event loop constructs
   - Break/continue semantics in when clauses
3. Ensure proper parsing of real-world control flow code

IMPLEMENTATION:
1. Study failing test patterns to understand required syntax
2. Extend grammar rules for complex control flow combinations
3. Add support for break/continue in when clauses
4. Handle nested loop/given/when combinations
5. Test against complex real-world Perl control flow code

EXPECTED OUTCOME:
- 3-4 complex control flow tests converted to passing
- Production-ready control flow parsing
- Foundation for complete Perl language support

VALIDATION:
- Test complex nested control flow structures
- Verify performance with deeply nested constructs
- Ensure simple control flow still works correctly
```

### Step 12: Validate Phase 3 Grammar Extensions

```
You are completing Phase 3 by validating all grammar extension work and measuring impact.

CONTEXT: Steps 9-11 added significant grammar extensions for control flow and edge cases. This should eliminate most grammar-related parsing failures.

TASK: Validate grammar extensions and prepare for Phase 4 (class/role declarations).

REQUIREMENTS:
1. Rebuild tree-sitter parser with all grammar changes
2. Run complete parser test suite and measure improvement
3. Verify no regressions in existing functionality
4. Document remaining failures for Phase 4
5. Performance test to ensure grammar extensions don't slow parsing

VALIDATION STEPS:
- Complete grammar rebuild: `make tree-sitter`
- Full parser test suite: `go test -v ./internal/parser -count=1`
- Compare to Phase 2 baseline: ~25 failures → target ~10 failures
- Performance regression test
- Categorize remaining failures for Phase 4

EXPECTED OUTCOME:
- ~15 failures eliminated by grammar extensions
- Parser failure rate reduced to ~10 failures (~98.5% pass rate)
- Ready for Phase 4 class/role implementation
- No performance regressions

SUCCESS CRITERIA:
- Major reduction in grammar-related failures
- No regressions in type annotation or complex type parsing
- Clear path to Phase 4 implementation
- Performance remains acceptable for production use
```

---

## Phase 4: Class and Role Declaration Implementation

**Target**: Reduce ~10 failures to 0 failures (100% pass rate)
**Strategy**: Implement full class and role declaration support

### Step 13: Implement Basic Class Declaration Grammar

```
You are beginning Phase 4, the final phase to achieve 100% parser completion by implementing class and role declaration support.

CONTEXT: Phase 3 achieved ~98.5% pass rate through grammar extensions. Phase 4 targets the final ~10 failures related to class and role declarations, which are major missing features in the typed Perl specification.

PROBLEM: Tests like `classes-roles_basic_role_declarations` fail because the tree-sitter grammar doesn't support `class` and `role` keywords and their associated syntax.

TASK: Add basic class declaration support to the tree-sitter grammar and parser.

TARGET FAILURES:
- classes-roles_basic_role_declarations
- Any basic class declaration parsing failures

REQUIREMENTS:
1. Study the class declaration syntax in `internal/parser/testdata/typed-perl/classes-roles.md`
2. Extend `tree-sitter-typed-perl/grammar.js` to support:
   - `class Name { ... }` declarations
   - `role Name { ... }` declarations
   - Basic class/role body parsing
   - Field declarations within classes: `field Type $name;`
   - Method declarations within classes
3. Implement parser logic to handle class/role AST nodes
4. Use test-driven development approach

IMPLEMENTATION APPROACH:
1. Start with minimal class declaration support:
   - Add `class` and `role` keywords to grammar
   - Basic class/role declaration structure
   - Empty class body parsing
2. Incrementally add class body support:
   - Field declarations with types
   - Method declarations
   - Access modifiers (if basic tests require them)
3. Extend parser logic to generate appropriate AST nodes
4. Test each increment against failing tests

TECHNICAL DETAILS:
- Add grammar rules: `class_declaration`, `role_declaration`
- Support class body with fields and methods
- Parse field declarations: `field Type $name;`
- Generate AST nodes for class/role information
- Ensure compatibility with existing type system

EXPECTED OUTCOME:
- 2+ basic class/role declaration tests converted to passing
- Foundation for advanced class features
- Core object-oriented typed Perl support

VALIDATION:
- Test basic class declaration parsing
- Verify field declarations work within classes
- Ensure method declarations in classes parse correctly
- No regression in existing parsing functionality
```

### Step 14: Add Generic Class and Role Parameter Support

```
You are continuing Phase 4, adding support for generic (parameterized) classes and roles.

CONTEXT: Basic class declarations from Step 13 provide the foundation. Now implement generic class support for patterns like `class Container[T] { ... }`.

PROBLEM: Tests like `classes-roles_generic_class_declarations` fail because generic class syntax isn't supported.

TASK: Extend class/role declarations to support type parameters (generics).

TARGET FAILURES:
- classes-roles_generic_class_declarations
- Any generic class/role related failures

REQUIREMENTS:
1. Extend grammar to support generic syntax:
   - `class Name[T] { ... }`
   - `role Name[T, U] { ... }`
   - Type parameter constraints: `class Name[T: Constraint] { ... }`
   - Multiple type parameters
2. Update parser to handle generic class AST nodes
3. Support generic methods within generic classes
4. Ensure type parameter information is preserved in AST

IMPLEMENTATION:
1. Extend grammar rules for parameterized classes/roles
2. Add type parameter list parsing
3. Support type constraints on parameters
4. Update AST generation for generic classes
5. Test against failing generic class patterns

TECHNICAL DETAILS:
- Grammar: `class_declaration` with optional type parameter list
- Parse `[T]`, `[T, U]`, `[T: Constraint]` syntax
- Handle generic methods: `method func[U](U $param) -> U`
- Generate AST with type parameter information
- Support constraint syntax parsing

EXPECTED OUTCOME:
- 2+ generic class tests converted to passing
- Full generic class/role declaration support
- Foundation for advanced type system features

VALIDATION:
- Test various generic class declaration patterns
- Verify type parameter information in AST
- Ensure generic methods work within generic classes
- Check constraint syntax parsing
```

### Step 15: Implement Field Access Modifiers and Advanced Features

```
You are continuing Phase 4, implementing field access modifiers and advanced class features.

CONTEXT: Basic and generic class support from Steps 13-14 covers core functionality. Now add advanced features like access modifiers, field visibility, and method constraints.

PROBLEM: Tests like `methods-fields_field_access_modifiers` fail because advanced class features aren't implemented.

TASK: Add support for field access modifiers and advanced class/role features.

TARGET FAILURES:
- methods-fields_field_access_modifiers
- Any advanced class feature failures

REQUIREMENTS:
1. Implement access modifier support:
   - `field private Type $name;`
   - `field protected Type $name;`
   - `field public Type $name;`
   - `method private name() { ... }`
2. Add readonly field support: `field readonly Type $name;`
3. Support method access modifiers
4. Handle constructor/destructor special methods if needed

IMPLEMENTATION:
1. Extend grammar for access modifier keywords
2. Add field access modifier parsing
3. Support method access modifiers
4. Implement readonly field declarations
5. Test against advanced class feature patterns

TECHNICAL DETAILS:
- Add keywords: `private`, `protected`, `public`, `readonly`
- Extend field declaration grammar: `field [modifier] Type $name`
- Method modifier grammar: `method [modifier] name()`
- Generate AST with access modifier information
- Ensure modifiers work with typed fields/methods

EXPECTED OUTCOME:
- 2+ advanced class feature tests converted to passing
- Complete field access modifier support
- Production-ready class declaration functionality

VALIDATION:
- Test all access modifier combinations
- Verify readonly field declarations
- Ensure modifier information preserved in AST
- Check method access modifiers work correctly
```

### Step 16: Implement Role Composition and Final Class Features

```
You are completing Phase 4 and the entire parser improvement plan by implementing role composition and final class/role features.

CONTEXT: Steps 13-15 implemented core class/role functionality. This final step addresses role composition, inheritance, and any remaining class/role features needed for 100% test completion.

PROBLEM: Tests like `classes-roles_role_composition_conflicts` fail because role composition and advanced inheritance features aren't implemented.

TASK: Implement role composition, inheritance, and complete class/role feature support.

TARGET FAILURES:
- classes-roles_role_composition_conflicts
- Any remaining class/role related failures
- All remaining parser failures to achieve 100%

REQUIREMENTS:
1. Implement role composition syntax:
   - `class Name with Role1, Role2 { ... }`
   - `role Name with OtherRole { ... }`
2. Add inheritance support: `class Child extends Parent { ... }`
3. Handle method conflict resolution in role composition
4. Support multiple role composition patterns
5. Complete any remaining class/role features for 100% completion

IMPLEMENTATION:
1. Extend grammar for role composition: `with Role1, Role2`
2. Add inheritance syntax: `extends Parent`
3. Support multiple inheritance/composition patterns
4. Implement method conflict resolution parsing
5. Complete final parser features for 100% test pass rate

FINAL VALIDATION:
1. Run complete parser test suite: `go test -v ./internal/parser -count=1`
2. Verify 0 failures, 100% pass rate achieved
3. Performance regression test
4. Complete integration test to ensure no system regressions
5. Document the achievement of 100% parser completion

TECHNICAL DETAILS:
- Grammar: `class_declaration` with `extends` and `with` clauses
- Parse multiple role composition: `with Role1, Role2, Role3`
- Handle inheritance hierarchy: `extends Parent with Role1, Role2`
- Generate complete AST for object-oriented features
- Ensure all class/role information available for type checking

EXPECTED OUTCOME:
- ALL remaining parser failures eliminated
- 100% parser test pass rate achieved (0/899 failures)
- Complete typed Perl class/role declaration support
- Production-ready object-oriented parsing capabilities

SUCCESS CRITERIA:
- Parser test suite: 100% pass rate (0 failures)
- No performance regressions
- Complete class/role feature support
- Foundation for advanced type checking and analysis
- Ready for production typed Perl development
```

---

## Final Integration and Validation

### Step 17: Complete System Integration and Performance Validation

```
You are completing the parser 100% achievement by performing final system integration and performance validation.

CONTEXT: Step 16 achieved 100% parser test completion. This final step ensures the parser improvements integrate correctly with the entire PVM system and don't introduce regressions in other components.

TASK: Validate complete system integration and performance after achieving 100% parser completion.

REQUIREMENTS:
1. Run complete PVM test suite: `make test`
2. Verify no regressions in other packages due to parser changes
3. Performance test entire system with new parser capabilities
4. Validate PSC, PVI, PVX components work with enhanced parser
5. Integration test with real-world typed Perl code
6. Document the complete achievement and system impact

INTEGRATION VALIDATION:
- Full system test suite execution
- Cross-component integration verification
- Real-world typed Perl file processing
- Performance benchmarking vs. baseline
- Memory usage analysis for complex parsing
- Documentation updates for 100% achievement

EXPECTED OUTCOME:
- 100% parser completion confirmed
- No system regressions introduced
- Performance remains acceptable for production
- Complete typed Perl language support achieved
- PVM ready for advanced typed Perl development

SUCCESS CRITERIA:
- Make test shows overall improvement in system test pass rate
- Parser: 100% pass rate maintained
- No performance regressions in any component
- Real-world typed Perl code parses and processes correctly
- System ready for production use with full typed Perl support
```

---

## Implementation Guidelines

### Test-Driven Development Requirements
- Every step must include comprehensive tests before implementation
- Each change must be validated against existing functionality
- Performance regression testing after each significant change
- Integration testing to ensure no system-wide regressions

### Risk Mitigation
- **Grammar Changes**: Comprehensive regression testing, git backup points
- **Parser Logic Changes**: Incremental implementation with validation
- **Performance Impact**: Benchmarking before/after each phase
- **System Integration**: Cross-component testing after major changes

### Success Metrics by Phase
- **Phase 1**: 73 → 35 failures (95%+ pass rate)
- **Phase 2**: 35 → 25 failures (97%+ pass rate)
- **Phase 3**: 25 → 10 failures (98.5%+ pass rate)
- **Phase 4**: 10 → 0 failures (100% pass rate)

### Completion Criteria
- Parser test suite: 0/899 failures (100% pass rate)
- No regressions in system test suite
- Performance benchmarks meet production requirements
- Complete typed Perl language feature support
- Integration with all PVM components confirmed

This plan provides a systematic, test-driven approach to achieving 100% parser completion through incremental improvements, comprehensive testing, and careful validation at each step.
