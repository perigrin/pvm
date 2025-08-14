# PVM Type System Reference

PVM provides a comprehensive type system for Perl with modern developer experience.

## Basic Type Syntax

Type annotations follow the pattern: `my Type $variable = value;`

```
my Int $count = 42;
my Str $name = "example";
my Bool $flag = 1;         # Only accepts: 1, 0, "", undef
my Num $price = 19.99;
my ArrayRef $items = [];
my HashRef $config = {};
```

## Parameterized Types

Types can be parameterized to specify the types of contained elements:

```
my ArrayRef[Int] @numbers;             # Array of integers
my HashRef[Str] %values;               # Hash with string values
my Map[Str, Int] %mapping;             # Key-value constraints
my Maybe[Str] $optional_name;          # Optional string value
my ArrayRef[HashRef[Str]] @records;    # Nested parameterized types
```

## Union & Intersection Types

Combine types for flexible type constraints:

```
my Int|Str $flexible;                  # Either integer or string
my ArrayRef|HashRef $collection;       # Array or hash reference
my Object&Serializable $obj;           # Object that implements Serializable
my !Undef $required;                   # Any type except undefined
```

## Types::Standard Migration

### Important: PVM uses Types::Standard-compliant type names

```
# ❌ OLD (no longer supported)
my HashRef[Str, Int] %old;

# ✅ NEW (Types::Standard compliant)
my Map[Str, Int] %new;                 # Correct key-value mapping
my HashRef[Int] %values;               # Hash with integer values only
```

## Bool Type Validation

💡 Bool type has strict validation rules:

**Accepted values:** 1, 0, "", undef
**Rejected values:** "true", "false", 2, -1, "yes", "no"
**Conversion:** Use explicit conversion for other values: !!$value

## Type Hierarchy Overview

PVM's type system follows this hierarchy:

```
Any
├── Defined
│   ├── Value
│   │   ├── Str
│   │   ├── Num (Int, Rat)
│   │   └── Bool
│   └── Ref
│       ├── ScalarRef
│       ├── ArrayRef[T]
│       ├── HashRef[T]
│       ├── CodeRef
│       └── Object
└── Undef
```

## Common Type Patterns

Frequently used type combinations:

```
Optional values:     Maybe[Str] $name
Collections:         ArrayRef[HashRef[Str]] @records
Callbacks:           CodeRef $handler
Configuration:       Map[Str, Any] %config
Flexible input:      Str|ArrayRef[Str] $input
Type assertions:     $value as Int
```

## Getting More Help

💡 For interactive examples, use 'pvm dev' to get real-time type checking feedback.

**Command:** pvm help workflows - Development workflows with types
**Command:** pvm build --check-only - Type-check your code
**Command:** pvm workspace doctor - Diagnose type-related issues
