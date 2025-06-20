# Tree-sitter-typed-perl Grammar Enhancement Plan

## Project Overview

This plan implements the missing type annotation features in tree-sitter-typed-perl grammar (Issue #18) through test-driven development. The goal is to enable PSC static analysis by adding support for:

- Type declarations (`type MyType = Int|Str`)
- Union types (`Int|Str`)
- Intersection types (`Object&Serializable`)
- Negation types (`!Undef`)
- Parameterized types (`ArrayRef[Int]`)
- Type assertions (`$value as Int`)
- Complex method signatures with types
- Type constraints (where clauses)

## Current State Analysis

**Working Features:**
- Basic type annotations in variable declarations (`my Int $var`)
- Simple parameterized types in variable context
- Method signatures with basic types

**Missing Features:**
- Standalone type declarations (grammar has no `type_declaration` rule)
- Union/intersection operators (`|`, `&`, `!`)
- Complex type expressions with parentheses
- Type assertion syntax (`as` operator)
- Advanced parameterized type nesting
- Type constraints and where clauses

**Test Coverage:**
- ~80.6% tests passing (3073/3811)
- Many tests skip with "tree-sitter-typed-perl grammar doesn't support typed Perl features yet"
- Existing test corpus has placeholder tests for missing features

## Implementation Strategy

### Phase 1: Foundation (Type Expression Infrastructure)
Build the core type expression parsing infrastructure that all other features depend on.

### Phase 2: Basic Type Features (Union/Intersection/Negation)
Add support for combining types with operators.

### Phase 3: Advanced Type Features (Parameterized/Nested)
Handle complex type expressions with proper precedence.

### Phase 4: Type Declarations and Assertions
Add standalone type declarations and type assertion syntax.

### Phase 5: Integration and Validation
Ensure all features work together and integrate with PSC.

---

## Detailed Implementation Plan

### Prompt 1: Establish Type Expression Foundation

**Context:** Currently the grammar has basic type annotation support but lacks a comprehensive type expression system. We need to establish the foundation for parsing complex type expressions.

**Objective:** Create robust type expression parsing infrastructure that can handle precedence, associativity, and nesting.

```
You are implementing tree-sitter-typed-perl grammar enhancements for Issue #18. Your task is to establish the foundation for type expressions in the grammar.

CURRENT STATE:
- Basic type annotations work in variable declarations (my Int $var)
- Complex type expressions fail to parse
- No unified type expression system

TASK: Create comprehensive type expression infrastructure

REQUIREMENTS:
1. Add `type_expr` rule that handles:
   - Simple types (Int, Str, Bool)
   - Qualified types (Package::Type)
   - Parenthesized expressions ((Int|Str))

2. Establish precedence for type operators (prepare for |, &, !)

3. Create test cases covering:
   - Simple type expressions: Int, Str, CustomType
   - Qualified types: Package::Type, Foo::Bar::Baz
   - Parenthesized expressions: (Int), ((Str))

IMPLEMENTATION APPROACH:
1. Add `type_expr` rule to grammar.js
2. Update existing type annotation rules to use `type_expr`
3. Create test file: test/corpus/type_expressions_foundation.txt
4. Run tests and ensure no regressions

TESTING REQUIREMENTS:
- All existing tests must continue to pass
- New test cases must parse correctly with proper AST structure
- Use tree-sitter test to validate parsing

DELIVERABLES:
- Updated grammar.js with type_expr foundation
- Test file with comprehensive type expression cases
- Verified parsing with tree-sitter test
- Documentation of any breaking changes

Remember: This is the foundation - keep it simple and focused on basic type expression parsing. Complex operators come in later prompts.
```

### Prompt 2: Implement Union Type Syntax

**Context:** With type expression foundation in place, we can now add union type support (`Int|Str`). This is one of the most critical missing features.

**Objective:** Add union type parsing with proper precedence and associativity.

```
You are continuing tree-sitter-typed-perl grammar implementation. Your previous work established type_expr foundation. Now implement union type syntax.

CURRENT STATE:
- type_expr rule exists and handles basic types
- Union types (Int|Str) not yet supported
- Tests exist but skip due to missing grammar

TASK: Implement union type syntax (Int|Str)

REQUIREMENTS:
1. Add union type operator `|` to type_expr with proper precedence
2. Support chained unions: Int|Str|Bool
3. Support parenthesized unions: (Int|Str)|Bool
4. Maintain left-associativity: Int|Str|Bool = ((Int|Str)|Bool)

IMPLEMENTATION APPROACH:
1. Add `union_type` rule using prec.left for associativity
2. Update `type_expr` to include union_type
3. Add comprehensive test cases in test/corpus/union_types.txt:
   - Simple unions: Int|Str
   - Chained unions: Int|Str|Bool
   - Parenthesized: (Int|Str)|Bool
   - In variable context: my Int|Str $var

TESTING REQUIREMENTS:
- Run existing test suite - no regressions allowed
- New union type tests must pass 100%
- Test AST structure with tree-sitter parse --debug
- Verify precedence by testing: Int|Str|Bool parses as ((Int|Str)|Bool)

VALIDATION STEPS:
1. make tree-sitter (regenerate parser)
2. tree-sitter test -f union_types
3. make test (verify no regressions)
4. Test manual parsing: echo "my Int|Str \$var;" | tree-sitter parse

DELIVERABLES:
- Updated grammar.js with union_type rule
- Comprehensive union type test file
- AST validation showing proper precedence
- Performance check - no significant parsing slowdown

Focus on union types only - intersection and negation come in the next prompts.
```

### Prompt 3: Add Intersection and Negation Types

**Context:** Union types are now working. Next we need intersection (`Object&Serializable`) and negation (`!Undef`) types to complete the basic type operators.

**Objective:** Implement intersection and negation type operators with correct precedence relationships.

```
You are continuing tree-sitter-typed-perl grammar development. Union types (Int|Str) are now working. Implement intersection and negation type operators.

CURRENT STATE:
- Union types (Int|Str) working correctly
- Intersection types (Object&Serializable) not supported
- Negation types (!Undef) not supported
- Need proper precedence: ! > & > |

TASK: Add intersection (&) and negation (!) type operators

REQUIREMENTS:
1. Implement intersection types: Object&Serializable
2. Implement negation types: !Undef
3. Establish correct precedence: !Undef&Other|Another = ((!Undef)&Other)|Another
4. Support complex combinations: !(Int|Str)&Object

PRECEDENCE RULES (high to low):
- Negation (!): highest precedence, right-associative
- Intersection (&): medium precedence, left-associative
- Union (|): lowest precedence, left-associative

IMPLEMENTATION APPROACH:
1. Add `negation_type` rule with highest precedence
2. Add `intersection_type` rule between negation and union
3. Update type_expr precedence hierarchy
4. Create comprehensive test cases

TEST CASES REQUIRED:
```perl
# Basic intersection
my Object&Serializable $obj;

# Basic negation
my !Undef $not_null;

# Precedence testing
my !Int&Str|Bool $complex;  # Should parse as (((!Int)&Str)|Bool)
my !(Int|Str) $negated_union;
my !Int|Str $negated_first;  # Should parse as ((!Int)|Str)
```

TESTING REQUIREMENTS:
1. Create test/corpus/intersection_negation_types.txt
2. Test precedence with manual AST inspection
3. Verify all existing tests still pass
4. Test edge cases: !!, A&B&C, complex nesting

VALIDATION STEPS:
1. make tree-sitter
2. tree-sitter test -f intersection_negation_types
3. Test precedence: echo "my !Int&Str|Bool \$x;" | tree-sitter parse --debug
4. make test (ensure no regressions)

DELIVERABLES:
- grammar.js with intersection_type and negation_type rules
- Comprehensive test file covering all operator combinations
- Precedence validation showing correct AST structure
- Performance verification

Focus on getting operator precedence exactly right - this is critical for correct type checking later.
```

### Prompt 4: Implement Parameterized Types

**Context:** Basic type operators are working. Now we need parameterized types (`ArrayRef[Int]`, `HashRef[Str, Int]`) which are essential for container types.

**Objective:** Add parameterized type syntax with support for multiple parameters and nesting.

```
You are continuing tree-sitter-typed-perl grammar implementation. Basic type operators (|, &, !) are working. Now implement parameterized types.

CURRENT STATE:
- Type operators (union, intersection, negation) working
- Simple parameterized types may work in variable context
- Complex parameterized types not fully supported
- Nested parameterized types (ArrayRef[HashRef[Str, Int]]) fail

TASK: Implement comprehensive parameterized type support

REQUIREMENTS:
1. Single parameter types: ArrayRef[Int], Optional[Str]
2. Multiple parameter types: HashRef[Str, Int], Result[Success, Error]
3. Nested parameterized types: ArrayRef[HashRef[Str, Int]]
4. Parameters can be any type expression: ArrayRef[Int|Str]

IMPLEMENTATION APPROACH:
1. Add `parameterized_type` rule with proper bracket handling
2. Support parameter_list with comma separation
3. Allow full type_expr as parameters (enables nesting)
4. Handle whitespace properly in brackets

GRAMMAR STRUCTURE:
```javascript
parameterized_type: $ => seq(
  field('base_type', $.type_name),
  '[',
  field('parameters', $.type_parameter_list),
  ']'
),

type_parameter_list: $ => seq(
  $.type_expr,
  repeat(seq(',', $.type_expr))
),
```

TEST CASES REQUIRED:
```perl
# Single parameter
my ArrayRef[Int] $numbers;
my Optional[Str] $maybe_name;

# Multiple parameters
my HashRef[Str, Int] $scores;
my Result[Success, Error] $result;

# Nested parameterized
my ArrayRef[HashRef[Str, Int]] $complex_data;
my HashRef[Str, ArrayRef[Int]] $lookup;

# Parameters with operators
my ArrayRef[Int|Str] $mixed_array;
my Optional[!Undef] $definitely_something;
```

TESTING REQUIREMENTS:
1. Create test/corpus/parameterized_types.txt
2. Test all nesting levels (at least 3 deep)
3. Verify parameter parsing with type operators
4. Ensure no regression in existing tests

VALIDATION STEPS:
1. make tree-sitter
2. tree-sitter test -f parameterized_types
3. Test complex case: echo "my ArrayRef[HashRef[Str, Int|Bool]] \$data;" | tree-sitter parse
4. make test (full test suite)

DELIVERABLES:
- grammar.js with parameterized_type and type_parameter_list rules
- Comprehensive test file covering all parameterization scenarios
- AST validation for complex nested cases
- Performance testing for deeply nested types

This is critical for PSC type checking - ensure parameterized types integrate properly with type operators.
```

### Prompt 5: Add Type Assertion Syntax

**Context:** Core type expressions are complete. Now we need type assertion syntax (`$value as Int`) which is essential for runtime type checking.

**Objective:** Implement type assertion operator with proper precedence in expressions.

```
You are continuing tree-sitter-typed-perl grammar development. Core type expressions (unions, intersections, parameterized) are working. Now implement type assertion syntax.

CURRENT STATE:
- All type expression features working (|, &, !, parameterized)
- Type assertions ($value as Int) not supported
- Need to integrate with existing expression parsing

TASK: Implement type assertion syntax ($value as Int)

REQUIREMENTS:
1. Add `as` operator for type assertions
2. Support any expression as left operand: $var as Int, func() as Str
3. Support any type expression as right operand: $val as Int|Str
4. Proper precedence in expression hierarchy

PRECEDENCE CONSIDERATIONS:
- Should bind tighter than most operators but looser than postfix
- Suggested precedence: between ARROW and UNOP
- Left-associative: $a as Int as Str should error or warn

IMPLEMENTATION APPROACH:
1. Add `type_assertion_expression` rule to grammar
2. Integrate with existing expression precedence
3. Add to `_term` or appropriate expression level
4. Handle edge cases (parentheses, operator precedence)

GRAMMAR STRUCTURE:
```javascript
type_assertion_expression: $ => prec.left(TERMPREC.TYPE_AS, seq(
  field('expression', $._expr),
  'as',
  field('type', $.type_expr)
)),
```

TEST CASES REQUIRED:
```perl
# Basic assertions
my $val = $input as Int;
my $name = get_name() as Str;

# Complex type assertions
my $data = $result as ArrayRef[Int];
my $mixed = $value as Int|Str;

# In expressions
if ($input as Int > 42) { ... }
return $data as HashRef[Str, Int];

# Edge cases
my $nested = ($value as Int) + 10;
my $chain = func($x as Str) as Int;
```

TESTING REQUIREMENTS:
1. Create test/corpus/type_assertions.txt
2. Test precedence with complex expressions
3. Verify integration with existing expression parsing
4. Test edge cases and error conditions

VALIDATION STEPS:
1. make tree-sitter
2. tree-sitter test -f type_assertions
3. Test precedence: echo "my \$x = \$y as Int + 10;" | tree-sitter parse --debug
4. make test (ensure no expression parsing regressions)

DELIVERABLES:
- grammar.js with type_assertion_expression rule
- Comprehensive type assertion test file
- Precedence validation with complex expressions
- Integration verification with existing expression tests

Critical: Type assertions must integrate smoothly with existing expression parsing without breaking anything.
```

### Prompt 6: Implement Type Declarations

**Context:** Type expressions and assertions are complete. Now we need standalone type declarations (`type MyType = Int|Str`) which are fundamental for type aliasing.

**Objective:** Add standalone type declaration syntax as top-level statements.

```
You are continuing tree-sitter-typed-perl grammar implementation. Type expressions and assertions are working. Now implement standalone type declarations.

CURRENT STATE:
- All type expression features complete (|, &, !, parameterized, assertions)
- No support for standalone type declarations
- Need top-level statement: type MyType = Int|Str;

TASK: Implement standalone type declaration syntax

REQUIREMENTS:
1. Type declarations as statements: type MyType = Int|Str;
2. Support any type expression as definition: type Complex = ArrayRef[Int|Str];
3. Type names follow identifier rules
4. Proper statement termination (semicolon)

IMPLEMENTATION APPROACH:
1. Add `type_declaration` rule as a statement type
2. Add to top-level statement choices
3. Support full type_expr as the definition
4. Handle scope and visibility (if applicable)

GRAMMAR STRUCTURE:
```javascript
type_declaration: $ => seq(
  'type',
  field('name', $.type_name),
  '=',
  field('definition', $.type_expr),
  ';'
),

type_name: $ => $._identifier,
```

TEST CASES REQUIRED:
```perl
# Basic type declarations
type UserId = Int;
type UserName = Str;

# Union type declarations
type Flexible = Int|Str;
type Status = Success|Error|Pending;

# Complex type declarations
type UserData = HashRef[Str, Int|Str];
type ProcessResult = Result[Success, Error];

# Nested declarations
type ComplexData = ArrayRef[HashRef[Str, Int|Str]];
type Callback = CodeRef[Void, (Int, Str)];

# In package context
package MyPackage {
    type LocalType = Int;
}
```

TESTING REQUIREMENTS:
1. Create test/corpus/type_declarations.txt
2. Test as top-level statements
3. Test in package/block contexts
4. Verify proper statement parsing integration

VALIDATION STEPS:
1. make tree-sitter
2. tree-sitter test -f type_declarations
3. Test in context: echo "type MyType = Int|Str; my MyType \$var;" | tree-sitter parse
4. make test (verify statement parsing integration)

DELIVERABLES:
- grammar.js with type_declaration rule
- Comprehensive type declaration test file
- Statement integration verification
- Documentation of type declaration syntax

This completes the core type system - type declarations enable full type aliasing for PSC.
```

### Prompt 7: Add Complex Method Signature Support

**Context:** Core type system is complete. Now we need to enhance method signatures to support complex type annotations including return types.

**Objective:** Extend method signature parsing to handle complex typed parameters and return type annotations.

```
You are completing tree-sitter-typed-perl grammar implementation. Core type system is working. Now enhance method signatures with complex type support.

CURRENT STATE:
- Core type system complete (declarations, expressions, assertions)
- Basic method signatures may work
- Complex method signatures with return types not fully supported
- Need: method foo(Int $x, Str $y) -> Bool { ... }

TASK: Implement comprehensive method signature type support

REQUIREMENTS:
1. Complex parameter types: method foo(ArrayRef[Int] $data) { ... }
2. Return type annotations: method foo() -> Int { ... }
3. Multiple typed parameters with defaults
4. Optional parameters and slurpy parameters

CURRENT SIGNATURE PARSING:
Review existing signature parsing in grammar.js and extend it to handle:
- Full type expressions in parameters
- Return type syntax (-> Type)
- Optional and slurpy parameters with types

IMPLEMENTATION APPROACH:
1. Extend parameter parsing to use full type_expr
2. Add return type annotation support
3. Handle parameter defaults with types
4. Support slurpy parameters (@rest, %opts) with types

GRAMMAR ENHANCEMENTS:
```javascript
// Extend existing signature rules
typed_parameter: $ => seq(
  field('type', $.type_expr),
  field('variable', $._signature_scalar),
  optional(seq('=', field('default_value', $._expr)))
),

return_type_annotation: $ => seq(
  '->',
  field('return_type', $.type_expr)
),
```

TEST CASES REQUIRED:
```perl
# Complex parameter types
method process_data(ArrayRef[Int] $numbers, HashRef[Str, Int] $lookup) { ... }

# Return type annotations
method get_count() -> Int { return 42; }
method get_name() -> Str { return "example"; }

# Combined parameter and return types
method transform(ArrayRef[Int] $input) -> ArrayRef[Str] { ... }

# Optional and default parameters
method create_user(Str $name, Int $age = 0, Bool $active = 1) -> User { ... }

# Slurpy parameters (if supported)
method log_message(Str $message, %opts) { ... }
method sum_numbers(Int @numbers) -> Int { ... }
```

TESTING REQUIREMENTS:
1. Create test/corpus/complex_method_signatures.txt
2. Test parameter type parsing
3. Test return type parsing
4. Test integration with method body parsing

VALIDATION STEPS:
1. make tree-sitter
2. tree-sitter test -f complex_method_signatures
3. Test complex case: echo "method foo(ArrayRef[Int] \$data) -> Bool { return 1; }" | tree-sitter parse
4. make test (verify no method parsing regressions)

DELIVERABLES:
- Enhanced method signature parsing in grammar.js
- Comprehensive method signature test file
- Verification of parameter and return type integration
- Performance check for complex signatures

This enables PSC to perform comprehensive method signature analysis.
```

### Prompt 8: Add Type Constraints and Where Clauses

**Context:** Method signatures are enhanced. The final major feature is type constraints and where clauses for generic programming support.

**Objective:** Implement type constraints (where clauses) for generic type parameters.

```
You are completing the final major feature of tree-sitter-typed-perl grammar. All core type features are working. Now implement type constraints and where clauses.

CURRENT STATE:
- Complete type system (declarations, expressions, assertions)
- Enhanced method signatures with return types
- Type constraints (where clauses) not supported
- Need: type Container[T] where T: Serializable = ...;

TASK: Implement type constraints and where clauses

REQUIREMENTS:
1. Generic type parameters: type Container[T] = ...
2. Type constraints: type Container[T] where T: Serializable = ...
3. Multiple constraints: where T: Serialize, T: Clone
4. Constraint inheritance: where T: Base & Trait

NOTE: This is advanced functionality - implement basic where clause syntax first, full constraint checking comes later in PSC.

IMPLEMENTATION APPROACH:
1. Add generic type parameter support to type declarations
2. Add where clause syntax
3. Add constraint specifications
4. Handle multiple constraints and inheritance

GRAMMAR STRUCTURE:
```javascript
// Enhanced type declaration with constraints
type_declaration: $ => seq(
  'type',
  field('name', $.type_name),
  optional(field('parameters', $.type_parameter_declaration)),
  optional(field('constraints', $.where_clause)),
  '=',
  field('definition', $.type_expr),
  ';'
),

type_parameter_declaration: $ => seq(
  '[',
  sepBy1(',', $.type_parameter),
  ']'
),

where_clause: $ => seq(
  'where',
  sepBy1(',', $.type_constraint)
),

type_constraint: $ => seq(
  field('parameter', $.type_name),
  ':',
  field('bound', $.type_expr)
),
```

TEST CASES REQUIRED:
```perl
# Basic generic types
type Container[T] = ArrayRef[T];
type Pair[A, B] = HashRef[A, B];

# Simple constraints
type Sortable[T] where T: Ord = ArrayRef[T];
type Serializable[T] where T: Serialize = T;

# Multiple constraints
type Complex[T] where T: Clone, T: Serialize = Container[T];

# Constraint inheritance
type Advanced[T] where T: Base & Trait = T;

# In method signatures (future)
# method process[T](T $data) where T: Serializable -> T { ... }
```

TESTING REQUIREMENTS:
1. Create test/corpus/type_constraints.txt
2. Focus on parsing correctness, not semantic validation
3. Test constraint syntax variations
4. Verify integration with existing type declarations

VALIDATION STEPS:
1. make tree-sitter
2. tree-sitter test -f type_constraints
3. Test parsing: echo "type Container[T] where T: Serialize = ArrayRef[T];" | tree-sitter parse
4. make test (full integration test)

DELIVERABLES:
- Enhanced type declaration parsing with constraints
- Type constraint test file
- Generic type parameter support
- Documentation of constraint syntax

This completes the core grammar features - PSC can now parse all typed Perl constructs.
```

### Prompt 9: Integration Testing and Performance Optimization

**Context:** All major grammar features are implemented. Now we need comprehensive integration testing and performance optimization.

**Objective:** Ensure all features work together correctly and optimize parsing performance.

```
You are completing tree-sitter-typed-perl grammar implementation. All major features are implemented. Now perform comprehensive integration testing and optimization.

CURRENT STATE:
- All major type features implemented
- Individual feature tests passing
- Need comprehensive integration testing
- May have performance issues with complex types

TASK: Integration testing and performance optimization

REQUIREMENTS:
1. Test all features working together
2. Identify and fix performance bottlenecks
3. Ensure no regressions in existing functionality
4. Optimize for real-world code patterns

INTEGRATION TEST SCENARIOS:
1. Complex type declarations using all features
2. Method signatures with complex types and constraints
3. Type assertions with complex expressions
4. Nested parameterized types with unions and intersections
5. Real-world code patterns

TEST CASES REQUIRED:
```perl
# Complex integration example
type UserId = Int;
type UserName = Str;
type UserData = HashRef[Str, Int|Str];
type UserResult[T] where T: Serialize = Result[T, Error];

method create_user(
    UserName $name,
    ArrayRef[UserData] $data,
    Optional[Bool] $active = 1
) -> UserResult[User] {
    my $user = $data as UserData;
    return $user as UserResult[User];
}

# Stress test - deeply nested types
type DeepNested = ArrayRef[HashRef[Str, ArrayRef[HashRef[Str, Int|Str]]]];
method process_deep(DeepNested $data) -> DeepNested { ... }
```

PERFORMANCE OPTIMIZATION:
1. Profile parsing time for complex types
2. Optimize repetitive patterns in grammar
3. Ensure reasonable performance for deeply nested types
4. Test memory usage with large files

TESTING REQUIREMENTS:
1. Create test/corpus/integration_comprehensive.txt
2. Run performance tests on complex files
3. Test against real-world typed Perl code
4. Verify PSC integration still works

VALIDATION STEPS:
1. make tree-sitter
2. tree-sitter test (all tests must pass)
3. Performance test: time tree-sitter parse large_typed_file.pl
4. make test (full project test suite)
5. Test PSC integration: psc check sample_typed_file.pl

DELIVERABLES:
- Comprehensive integration test file
- Performance optimization report
- Any grammar fixes for edge cases discovered
- Documentation of performance characteristics

This ensures the grammar is production-ready for PSC integration.
```

### Prompt 10: Final Validation and Documentation

**Context:** Grammar implementation is complete and tested. Final step is comprehensive validation and documentation updates.

**Objective:** Perform final validation, update documentation, and prepare for PSC integration.

```
You are completing tree-sitter-typed-perl grammar implementation. All features are implemented and tested. Perform final validation and documentation updates.

CURRENT STATE:
- All grammar features implemented and integration tested
- Performance optimized
- Need final validation and documentation updates

TASK: Final validation and documentation completion

REQUIREMENTS:
1. Complete test suite validation (100% pass rate goal)
2. Update PARSING_FAILURE_PATTERNS.md
3. Update project documentation
4. Validate PSC integration readiness

VALIDATION CHECKLIST:
1. All tree-sitter tests pass: tree-sitter test
2. All project tests pass: make test
3. Grammar generates without errors: tree-sitter generate
4. No parsing regressions in existing code
5. PSC can use new grammar features

DOCUMENTATION UPDATES:
1. Update PARSING_FAILURE_PATTERNS.md - remove resolved patterns
2. Add new grammar features to documentation
3. Update build instructions if needed
4. Document any new dependencies or requirements

PSC INTEGRATION VERIFICATION:
1. Test PSC with new grammar features
2. Verify static analysis works with new constructs
3. Test error reporting for invalid syntax
4. Confirm performance is acceptable

FINAL TEST SCENARIOS:
```perl
# Everything working together
type Result[T, E] where T: Clone, E: Debug = {
    success: T,
    error: Optional[E]
};

method process_data[T](
    ArrayRef[T] $input,
    CodeRef[Bool, T] $filter
) -> Result[ArrayRef[T], ProcessError]
where T: Serialize + Clone {
    my $filtered = [];
    for my $item (@$input) {
        my $typed_item = $item as T;
        if ($filter->($typed_item)) {
            push @$filtered, $typed_item;
        }
    }
    return { success => $filtered, error => undef } as Result[ArrayRef[T], ProcessError];
}
```

DELIVERABLES:
1. Final test validation report
2. Updated PARSING_FAILURE_PATTERNS.md
3. Updated project documentation
4. PSC integration verification
5. Performance benchmarks
6. Any final grammar refinements

SUCCESS CRITERIA:
- make test shows 100% pass rate for new features
- No regressions in existing functionality
- PSC can successfully parse all new type constructs
- Grammar performance acceptable for production use

This completes Issue #18 - tree-sitter-typed-perl grammar is production-ready.
```

---

## Summary

This plan implements tree-sitter-typed-perl grammar enhancements through 10 iterative, test-driven prompts:

**Phase 1 (Foundation):** Prompts 1-2 establish type expression infrastructure and union types
**Phase 2 (Operators):** Prompts 3-4 add intersection/negation operators and parameterized types
**Phase 3 (Advanced):** Prompts 5-6 implement type assertions and declarations
**Phase 4 (Complete):** Prompts 7-8 add method signatures and constraints
**Phase 5 (Production):** Prompts 9-10 provide integration testing and final validation

Each prompt builds incrementally on previous work, maintains 100% test coverage, and includes comprehensive validation steps. The result enables PSC to perform meaningful static analysis on typed Perl code.
