# Union Type System Implementation Blueprint

## Overview

We're implementing a pragmatic union type system for PVM's Typed Perl that:
- Extends the existing working type system
- Uses trait-based capabilities derived from Perl's overload.pm core operations
- Computes trait intersections for union types (lazy + caching)
- Resolves operations to concrete result types
- Maintains Perl's coercion philosophy while enabling future strictness

## High-Level Architecture

```
Union Type (ArrayRef|Str)
    ↓
Trait Intersection Computation (lazy)
    ↓
Operation Validation (+ trait check)
    ↓
Result Type Resolution (Num from + trait)
```

## Implementation Phases

### Phase 1: Foundation (Steps 1-3)
- Define trait data structures based on overload.pm core
- Extend type system to support traits
- Implement basic trait intersection algorithm

### Phase 2: Union Types (Steps 4-6)
- Add union type representation
- Implement union type parsing
- Wire trait intersection into union type checking

### Phase 3: Integration (Steps 7-9)
- Integrate with existing type checker
- Add comprehensive testing
- Performance optimization (caching)

---

## Detailed Implementation Steps

### Step 1: Core Trait System Foundation ✅ COMPLETED

```text
Implement the fundamental trait data structures based on Perl's overload.pm core operations.

Create the basic trait system in the existing type system:
- Define a Trait struct with operation symbol and result type
- Identify overload.pm's core operations (the minimal set that others derive from)
- Create trait definitions for fundamental operations: "", 0+, bool, <=>, cmp, +, -, *, /, %, **, &, |, ^, ~, <<, >>, eq, ne, lt, le, gt, ge
- Add trait storage to existing type objects
- Define default traits for existing basic types (Int, Str, Num, Bool, ArrayRef, HashRef)

Requirements:
- Use TDD approach - write tests first for trait definitions and basic operations
- Follow existing codebase patterns for data structures
- Ensure traits can be easily queried and manipulated
- Create comprehensive unit tests for trait assignment and queries
- Test that basic types have expected default traits

Focus on simple, correct implementation. This foundation will support all subsequent union type work.

IMPLEMENTATION STATUS: ✅ COMPLETED
- Created comprehensive trait system in internal/traits/ package
- Implemented Trait struct and TraitSet with full operations
- Defined all 23 core overload.pm operations with correct result types
- Added default traits for all basic types (Int, Str, Num, Bool, ArrayRef, HashRef)
- Comprehensive test suite with 100% coverage
- TDD approach with tests written first
- All tests passing
```

### Step 2: Trait Operation Resolution ✅ COMPLETED

```text
Implement the operation-to-result-type resolution system using the trait foundation from Step 1.

Build the system that determines what happens when operations are performed:
- Create an operation resolver that takes a type and operation symbol
- Implement trait lookup to determine if a type supports an operation
- Add result type resolution (what type does the operation produce)
- Handle operation validation (does this type support this operation?)
- Create comprehensive error messages for unsupported operations

Requirements:
- Extend the existing type checker to use trait-based operation validation
- Write extensive tests covering all core operations on all basic types
- Ensure proper error handling and reporting for invalid operations
- Test edge cases like operations on unknown or invalid types
- Verify that operation resolution matches Perl's actual behavior

This step establishes the foundation for checking operations on any type, which will be essential for union type intersection calculations.

IMPLEMENTATION STATUS: ✅ COMPLETED
- Created comprehensive OperationResolver with trait-based validation
- Implemented operation support checking and result type resolution
- Added binary and unary operation handling
- Comprehensive error reporting for unsupported operations
- Support for custom trait assignment and unknown type handling
- Extensive test suite covering all operations on all basic types
- Advanced features: operation sequences, type comparisons, operation info
- All tests passing
```

### Step 3: Trait Intersection Algorithm

```text
Implement the core algorithm that computes trait intersections for multiple types.

Create the intersection computation system:
- Design algorithm to find common traits across multiple types
- Implement lazy computation (compute when needed, not at creation time)
- Add trait intersection caching for performance
- Handle edge cases (empty intersections, single types, etc.)
- Create utilities for intersection operations and queries

Requirements:
- Write comprehensive unit tests for intersection computation
- Test various combinations of types and their trait intersections
- Verify lazy computation works correctly and efficiently
- Test caching behavior and cache invalidation
- Ensure intersection results are consistent and correct

This algorithm is the heart of union type capability determination. It must be rock-solid before we add union types themselves.
```

### Step 4: Union Type Data Structure

```text
Extend the existing type system to support union type representation and basic operations.

Add union type support to the type system:
- Define UnionType struct/class that holds multiple member types
- Integrate union types into the existing type hierarchy
- Add basic union type operations (creation, member access, equality)
- Implement union type string representation for debugging/display
- Add validation for union type construction (no empty unions, handle duplicates)

Requirements:
- Follow existing type system patterns and interfaces
- Write comprehensive tests for union type creation and basic operations
- Test union type validation and error cases
- Ensure union types integrate cleanly with existing type infrastructure
- Test string representation and debugging output

This step adds the basic data structure without any complex logic. The union type should be a first-class citizen in the type system.
```

### Step 5: Union Type Capability System

```text
Integrate the trait intersection algorithm with union types to determine their capabilities.

Wire together the intersection algorithm with union type checking:
- Add capability computation to union types using trait intersection
- Implement operation support checking for union types
- Add result type computation for operations on union types
- Create caching system for computed capabilities
- Handle capability queries efficiently

Requirements:
- Union types should report capabilities as intersection of member capabilities
- Operation checking should use trait intersection results
- Result types should be computed correctly based on operation traits
- Add comprehensive tests covering various union type capability scenarios
- Test performance and caching behavior

This step makes union types fully functional for capability checking while maintaining the lazy computation design.
```

### Step 6: Union Type Parsing Integration

```text
Extend the existing parser to recognize and create union type syntax.

Add union type parsing to the tree-sitter integration:
- Extend grammar to recognize union type syntax (Type1|Type2|Type3)
- Add parsing logic for union type declarations
- Integrate union type creation with existing type parsing
- Handle complex union type expressions and nested unions
- Add proper error handling for malformed union type syntax

Requirements:
- Build on existing tree-sitter parser integration
- Write extensive parsing tests for various union type syntaxes
- Test error handling for invalid union type declarations
- Ensure parsing integrates seamlessly with existing type declaration parsing
- Test complex expressions and edge cases

The parser should now be able to handle union types anywhere regular types are used, creating proper UnionType objects.
```

### Step 7: Type Checker Integration

```text
Integrate union types with the existing type checking system for full functionality.

Wire union types into the complete type checking workflow:
- Modify assignment checking to handle union types
- Add union type compatibility checking (can Type1|Type2 be assigned to Type3?)
- Implement operation validation using union type capabilities
- Add proper error reporting for union type mismatches
- Integrate with existing error reporting system

Requirements:
- Extend existing type checker without breaking current functionality
- Write comprehensive integration tests covering assignment and operations
- Test type compatibility rules with union types
- Ensure error messages are clear and helpful
- Test interaction with existing type checking features

The type checker should now fully support union types in all contexts where regular types work.
```

### Step 8: Comprehensive Testing and Edge Cases

```text
Create comprehensive test suite covering all union type functionality and edge cases.

Build exhaustive test coverage:
- Add integration tests covering real-world union type scenarios
- Test complex union type expressions and nested operations
- Add performance tests for trait intersection and caching
- Test edge cases (empty unions, single-member unions, recursive unions)
- Create regression tests for existing functionality

Requirements:
- Achieve high test coverage for all new union type code
- Test realistic Perl code patterns using union types
- Verify performance characteristics meet requirements
- Test all error conditions and edge cases
- Ensure no regressions in existing type system functionality

This step ensures the union type system is robust and ready for production use.
```

### Step 9: Performance Optimization and Caching

```text
Optimize union type performance through intelligent caching and algorithm improvements.

Implement performance enhancements:
- Optimize trait intersection computation for common cases
- Add intelligent caching for capability lookups
- Implement cache invalidation strategies
- Add performance monitoring and metrics
- Optimize memory usage for union type storage

Requirements:
- Profile existing performance and establish benchmarks
- Implement caching without changing external behavior
- Write tests to verify cache correctness and invalidation
- Measure performance improvements and memory usage
- Ensure optimizations don't introduce bugs

The final system should be both correct and performant for real-world usage.
```

---

## Integration Notes

Each step builds directly on the previous ones:
- Step 1 creates the trait foundation
- Step 2 builds operation resolution on top of traits
- Step 3 adds intersection computation for multiple types
- Step 4 adds union type data structures
- Step 5 combines intersection with union types
- Step 6 integrates with parsing
- Step 7 integrates with type checking
- Step 8 ensures robustness
- Step 9 optimizes performance

No code should be left orphaned - each step integrates immediately with the previous work and the existing codebase.

## Testing Strategy

- **Unit tests** for each component in isolation
- **Integration tests** between components
- **End-to-end tests** with real Perl code
- **Performance tests** for caching and algorithms
- **Regression tests** to ensure existing functionality is preserved

This approach ensures steady progress with strong foundations at each step.
