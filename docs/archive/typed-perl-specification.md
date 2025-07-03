# Typed Perl: Comprehensive Type System Specification

This document provides a comprehensive specification for the Typed Perl type system. It is intended to be detailed enough for implementation by developers with limited prior type system experience.

## 1. Core Design Principles

### 1.1 Type System Philosophy

Typed Perl is a gradual, bidirectional type system with flow-sensitive analysis, designed to match Perl's dynamic nature while providing compile-time safety guarantees.

1. **Gradual** - Type annotations are optional and can be added incrementally
2. **Bidirectional** - Type information flows both from declarations to expressions and from usage to variables
3. **Flow-sensitive** - Type refinements based on runtime checks are tracked through control flow
4. **Zero runtime overhead** - Types are stripped after verification
5. **Separate from runtime** - Type checking happens in a separate process from Perl execution (like TypeScript/JavaScript relationship)

### 1.2 Type Checking Modes

1. **Strict Mode (Default)** - Expressions must have a specific type; inference to `Any` is an error
2. **Relaxed Mode (Optional)** - Expressions may default to `Any` when type cannot be determined

### 1.3 Type Inference Boundaries

1. **Global Inference** - Type information flows across function and module boundaries
2. **Principle** - If type can be inferred precisely, explicit annotation should not be required

### 1.4 Core Assumptions

1. **Strictness** - We assume `use strict` and `use warnings` are enabled (default as of Perl 5.36)
2. **Type Assertions** - Type casts check compatibility but don't convert values (no runtime access)
3. **Configuration** - The type checker will support configuration options for warnings, errors, and experimental features

## 2. Type Hierarchy and Relationships

### 2.1 Core Type Hierarchy

```
Unknown  (default state for unanalyzed expressions)
Any      (explicit polymorphic type)

├── Scalar
│   ├── Str
│   │   └── Num
│   │       └── Int
│   │           └── Bool
│   ├── Undef
│   └── Ref
│       ├── Object
│       ├── ScalarRef
│       ├── ArrayRef
│       ├── HashRef
│       ├── CodeRef
│       └── ...
├── List
│   ├── Array
│   └── Hash
├── Code
└── Glob
```

### 2.2 Type Relationships

1. **Subtyping** - If A is a subtype of B, values of type A can be used where B is expected
2. **Assignment Compatibility** - Value of type S can be assigned to variable of type T if:
   - S is a subtype of T
   - S can be safely coerced to T (non-lossy coercion)
   - T is a supertype of S

### 2.3 Type Operations

1. **Union Types (T1 | T2)** - Values that can be either type T1 or type T2
2. **Intersection Types (T1 & T2)** - Values that satisfy both type T1 and type T2
3. **Type Negation (!T)** - Values that are not of type T

### 2.4 Parameterized Types

1. **Container Types** - `ArrayRef[T]`, `HashRef[K, V]`, etc.
2. **Generic Constraints** - Type parameters can have constraints: `ArrayRef[T: Num]`
3. **Recursive Types** - Types that refer to themselves: `type Tree = { value: Int, children: ArrayRef[Tree] }`
4. **Variance Annotations**:
   - Covariant (`+T`): If S is a subtype of T, then Container[S] is a subtype of Container[T]
   - Contravariant (`-T`): If S is a subtype of T, then Container[T] is a subtype of Container[S]
   - Invariant (`T`): No subtyping relationship regardless of relationship between S and T
5. **Default Variance**:
   - `ArrayRef[+T]` - Covariant in read-only context, invariant when modified
   - `HashRef[+K, +V]` - Covariant in read-only context
   - Function parameters are contravariant
   - Function return types are covariant

### 2.5 Higher-Kinded Types

1. **Type Constructors** - Types that take other types as parameters
2. **Kind System** - Types have kinds: `Type`, `Type -> Type`, `(Type -> Type) -> Type`, etc.
3. **HKT Examples** - `Monad[F]`, `Functor[F]`, where F is itself a type constructor

### 2.6 Typing Approaches

1. **Structural Typing** - Types are compatible based on their structure/shape
2. **Nominal Typing** - Classes and roles are typed based on their inheritance relationships
3. **Hybrid Approach** - Use nominal typing for classes/roles, structural typing for hash-based objects and duck typing

## 3. Type Inference System

### 3.1 Bidirectional Type Checking

1. **Type Checking (↓)** - Propagate known types down to expressions
2. **Type Synthesis (↑)** - Generate types from expressions upward
3. **Synthesis Priority** - Generate most specific type possible

### 3.2 Operator-Based Type Inference

1. **String Operators** - `eq`, `ne`, `lt`, `gt`, `.` (concat) imply Str operands
2. **Numeric Operators** - `+`, `-`, `*`, `/`, `==`, `!=`, `<`, `>` imply Num operands
3. **Boolean Context** - Conditional expressions imply Bool result
4. **Reference Tests** - `ref($x) eq 'ARRAY'` implies Ref type for $x

### 3.3 Context Sensitivity

1. **List vs. Scalar Context** - Functions may return different types based on context
2. **Union Type Approach** - Functions with context-dependent returns have union return types
3. **Context Resolution** - Assignment context determines which part of union applies:
   ```perl
   sub List[Int]|Str get_data() { ... }
   my @data = get_data();  # List[Int] part applies
   my $data = get_data();  # Str part applies
   ```

### 3.4 Flow-Sensitive Type Refinement

1. **Pattern Recognition** - Common validation patterns refine types in conditional branches:
   ```perl
   if ($x =~ /^\d+$/) {
       # $x refined from Str to Int here
   }
   ```

2. **Key Patterns**:
   - **Numeric Validation**: `/^\d+$/`, `looks_like_number()`, etc.
   - **Reference Checks**: `ref($x) eq 'TYPE'`
   - **Definedness**: `defined($x)`
   - **Boolean Tests**: `if($x)`, `unless($x)`

3. **Type Narrowing** - Types become more specific based on conditionals:
   ```perl
   my $data = get_input();  # Str

   if ($data =~ /^\d+$/) {
       say $data * 2;  # Int (narrowed from Str)
   }
   ```

4. **Extensibility** - Future versions will allow custom pattern registration
5. **Path-Sensitive Analysis** - Planned for future versions, will track relationships between variables

## 4. Type Syntax

### 4.1 Variable Type Annotations

```perl
my Type $variable;
my Type $variable = initial_value;
my Type @array;
my Type %hash;
```

### 4.2 Function Signatures and Return Types

```perl
sub ReturnType function_name(ParamType1 $param1, ParamType2 $param2) {
    # Implementation
}

method ReturnType method_name(ParamType $param) {
    # Implementation
}
```

### 4.3 Generic Type Syntax

```perl
# Parameterized types
my ArrayRef[Str] $names;
my HashRef[Str, Num] %grades;

# Nested generics
my ArrayRef[HashRef[Str, Num]] $records;

# Type variables
sub T identity<T>(T $value) {
    return $value;
}

# Variance annotations
type Box<+T> = { value: T };  # Covariant: Box[Dog] is a subtype of Box[Animal]
type Callback<-T> = CodeRef[(T) -> Void];  # Contravariant: Callback[Animal] is a subtype of Callback[Dog]
type Mutable<T> = { get: () -> T, set: (T) -> Void };  # Invariant: no subtyping relationship
```

### 4.4 Type Combinations

```perl
# Union types
my Str|Num $value;

# Intersection types
my Serializable & Printable $object;

# Type negation
my !Ref $primitive;
```

### 4.5 Type Aliases

```perl
type UserID = Str;
type PositiveInt = Int & {$_ > 0};
type Record = {name: Str, age: Int, email: Str};
type Tree<T> = {value: T, children: ArrayRef[Tree<T>]};
```

### 4.6 Type Assertions

```perl
my $value = get_data();
my $name = (Str) $value;  # Explicitly cast to Str
```

## 5. Type Definition Files

Type definition files allow specifying types for existing modules without modifying them.

### 5.1 File Format

Type definition files have a `.ptd` extension and define types for existing modules:

```perl
# DBI.ptd
package DBI {
    class DBI::db {
        method prepare(Str $query) -> DBI::st;
        method selectall_arrayref(Str $query) -> ArrayRef[ArrayRef[Scalar]];
    }

    class DBI::st {
        method execute(@params) -> Bool;
        method fetchrow_array() -> List;
    }
}
```

### 5.2 Module Import

Types from definition files are imported when the underlying Perl module is used:

```perl
use DBI;  # The type checker automatically applies type information from DBI.ptd
```

### 5.3 Type Information Management

The type system maintains its own registry of type information, separate from Perl's runtime:

1. **Parallel System** - Type information exists in a separate space from Perl's runtime
2. **Module Detection** - The checker detects `use Module` statements and applies corresponding types
3. **No Runtime Integration** - Type information has no access to Perl's runtime environment
4. **Import/Export** - Type import/export is tracked separately from Perl's symbol import/export

## 6. Implementation Requirements

### 6.1 Parser Extensions

1. **Grammar Extensions** - Extend Tree-sitter Perl grammar to recognize type syntax
2. **AST Enhancements** - Augment AST to include type annotations
3. **Source Mapping** - Maintain mapping between types and source locations

### 6.2 Type Checker

1. **Type Environment** - Track variables and their types in scope
2. **Constraint Collection** - Generate type constraints from expressions
3. **Constraint Solving** - Unify constraints to determine types
4. **Error Reporting** - Generate clear, actionable type error messages

### 6.3 Type Stripping

1. **AST Transformation** - Remove type annotations from AST
2. **Source Generation** - Output valid Perl code without type annotations

### 6.4 Configuration Options

1. **Type Checking Levels**:
   - `--strict` (default) - No implicit Any types
   - `--relaxed` - Allow implicit Any types
2. **Warning Controls**:
   - `--warn-lossy-coercion` - Warn on potentially lossy type coercions
   - `--warn-unused-types` - Warn on unnecessary type annotations
3. **Error Controls**:
   - `--allow-unknown-modules` - Don't error on modules without type definitions
   - `--allow-any-cast` - Allow casting to Any without warning
4. **Performance Options**:
   - `--fast` - Skip some advanced checks for faster performance
   - `--thorough` - Perform all possible checks
5. **Experimental Features**:
   - `--enable-experimental=<feature>` - Enable experimental type system features

### 6.5 Performance Requirements

1. **Priority** - Correctness and completeness over performance in initial implementation
2. **Incremental Analysis** - Support for analyzing changes without full reprocessing (future)
3. **Caching** - Type information should be cached where appropriate

## 7. Type Checking Rules

### 7.1 Variables and Assignment

1. **Declaration** - Variables with explicit type annotations must satisfy that type
2. **Assignment** - Right-hand side must be compatible with variable's type
3. **Inference** - Variables without annotations get type from initialization or usage

### 7.2 Operators

1. **Arithmetic** - Operands must be compatible with Num, result is Num
2. **String** - Operands must be compatible with Str, result is Str
3. **Comparison** - Type determined by operator (e.g., `eq` requires Str, returns Bool)
4. **Logical** - Operands evaluated in Boolean context, result is Bool
5. **Regular Expression** - Match operators (`=~`, `!~`) expect (Str, Regexp) and return Bool

### 7.3 Access Patterns

1. **Array Access** - Array indices are always typed as PositiveInt, even when negative integers are used
2. **Hash Access** - Hash keys are always typed as Str
3. **Match Variables** - After regex match, variables like $1, $2 are typed as Str

### 7.4 Function Calls

1. **Arguments** - Must be compatible with parameter types
2. **Return Value** - Must be compatible with declared return type
3. **Context** - Return type may depend on caller context (list vs. scalar)

### 7.5 Control Flow

1. **Conditionals** - Condition expressions evaluated in Boolean context
2. **Loops** - Iteration variables take type from collection elements
3. **Type Refinement** - Types refined in branches based on conditions

### 7.6 Object-Oriented Code

1. **Method Calls** - Object must have method, arguments must match signature
2. **Inheritance** - Subtypes satisfy supertype requirements (Liskov Substitution)
3. **Roles/Traits** - Objects satisfy role requirements if they implement required methods

## 8. Error Handling

### 8.1 Error Categories

1. **Type Mismatch** - Value does not match expected type
2. **Undefined Method** - Method not available on object
3. **Invalid Operation** - Operation not valid for operand types
4. **Ambiguous Type** - Cannot determine specific type (defaults to error in strict mode)

### 8.2 Error Reporting

1. **Location** - Source file, line, and column
2. **Context** - Expression that caused the error
3. **Expected vs. Actual** - What type was expected vs. what was found
4. **Suggestions** - Potential fixes or clarifications

## 9. Future Extensions

### 9.1 Path-Sensitive Analysis

Future versions will track relationships between variables, understanding that conditions on one variable may imply conditions on related variables.

### 9.2 Custom Validation Patterns

A future extension will allow registering custom validation patterns for type refinement.

### 9.3 Performance Optimizations

Later versions will focus on performance improvements once the core functionality is stable and correct.

## Appendix A: Type Compatibility Rules

| Type A | Type B | Compatible? | Notes |
|--------|--------|-------------|-------|
| Int    | Num    | Yes         | Int is a subtype of Num |
| Num    | Str    | Yes         | Num is a subtype of Str |
| Int    | Str    | Yes         | Int is a subtype of Str |
| Str    | Num    | Conditional | Only if validated (e.g., `/^\d+$/`) |
| ArrayRef[A] | ArrayRef[B] | Conditional | If A is compatible with B (covariant) |
| HashRef[A,B] | HashRef[C,D] | Conditional | If A is compatible with C and B with D (covariant) |
| CodeRef[(A)->R] | CodeRef[(B)->S] | Conditional | If B is compatible with A (contravariant) and R is compatible with S (covariant) |
| A | A\|B | Yes | A is a subtype of A\|B |
| A&B | A | Yes | A&B is a subtype of A |
| A | !A | No | Types are disjoint |
| Box<+A> | Box<+B> | Yes | If A is a subtype of B (covariant) |
| Callback<-A> | Callback<-B> | Yes | If B is a subtype of A (contravariant) |
| Container<A> | Container<B> | No | Invariant, no subtyping regardless of A and B relationship |

## Appendix B: Operator Type Rules

| Operator | Operand Types | Result Type |
|----------|---------------|-------------|
| +, -, *, / | Num, Num | Num |
| . (concat) | Str, Str | Str |
| eq, ne, lt, gt | Str, Str | Bool |
| ==, !=, <, > | Num, Num | Bool |
| &&, \|\|, ! | Any, Any | Bool |
| =~ | Str, Regexp | Bool |

This specification provides a comprehensive blueprint for implementing the Typed Perl type system. The design combines features from modern type systems while respecting Perl's dynamic and pragmatic nature.
