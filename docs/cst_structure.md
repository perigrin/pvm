# CST Structure Documentation for Typed Perl

## Overview

This document provides comprehensive documentation of the Concrete Syntax Tree (CST) structure produced by tree-sitter-typed-perl for typed Perl constructs. This information is essential for implementing the unified compiler architecture that works directly with CST nodes.

## Core Architecture

The tree-sitter-typed-perl grammar extends standard Perl syntax with type annotations. The CST preserves exact source structure including:

- **Precise positioning**: Every node has StartByte/EndByte positions
- **Complete syntax**: All punctuation and formatting preserved
- **Type annotations**: Dedicated nodes for type information
- **Error recovery**: ERROR nodes for unparseable constructs

## CST Node Types

### Source Structure Nodes

- `source_file` - Root node containing entire source
- `expression_statement` - Statement-level expressions
- `assignment_expression` - Assignment operations
- `block` - Code blocks `{ ... }`
- `statement` - General statement nodes

### Variable Declaration Nodes

#### Basic Structure
```
variable_declaration
├── my|field|our      # Declaration keyword
├── type_expression   # Type annotation (optional)
│   └── simple_type   # Basic type name
├── scalar|array|hash # Variable sigil and name
│   ├── $|@|%        # Variable sigil
│   └── varname      # Variable identifier
├── =                # Assignment operator (optional)
└── value            # Initial value (optional)
```

#### Node Types
- `variable_declaration` - Complete variable declaration
- `my`, `field`, `our` - Declaration keywords
- `type_expression` - Container for type annotations
- `simple_type` - Basic type identifier
- `scalar`, `array`, `hash` - Variable containers
- `varname` - Variable name identifier

### Type Annotation Nodes

#### Type Expression Structure
```
type_expression
├── simple_type       # Basic type: Int, Str, etc.
├── parameterized_type # Generic type: ArrayRef[Int]
├── union_type        # Union: Int|Str
└── intersection_type # Intersection: Object&Serializable
```

#### Type Assertion Structure
```
type_assertion_expression
├── scalar            # Variable being asserted
│   ├── $            # Sigil
│   └── varname      # Variable name
├── as               # Assertion keyword
└── type_expression  # Target type
    └── simple_type  # Type name
```

### Method Declaration Nodes

#### Basic Structure
```
method_declaration_statement
├── method            # Method keyword
├── bareword         # Method name
├── signature        # Parameter list
│   ├── (            # Opening paren
│   ├── mandatory_parameter # Typed parameter
│   │   ├── type_expression
│   │   │   └── simple_type
│   │   └── scalar
│   │       ├── $
│   │       └── varname
│   └── )            # Closing paren
├── ERROR            # Return type (currently not parsed)
│   └── -> Type      # Return type syntax
└── block            # Method body
```

## Typed Construct Patterns

### Pattern 1: Typed Variable Declaration

**Source:** `my Int $count = 42;`

**CST Structure:**
```
source_file [0:19]
└── expression_statement [0:18]
    └── assignment_expression [0:17]
        ├── variable_declaration [0:13]
        │   ├── my [0:2] "my"
        │   ├── type_expression [3:6]
        │   │   └── simple_type [3:6] "Int"
        │   └── scalar [7:13]
        │       ├── $ [7:8] "$"
        │       └── varname [8:13] "count"
        ├── = [14:15] "="
        └── number [16:18] "42"
```

**Key Characteristics:**
- `variable_declaration` contains both type and variable
- `type_expression` > `simple_type` path to type annotation
- `scalar` > `varname` path to variable name
- Assignment is separate from declaration

### Pattern 2: Field Declaration

**Source:** `field Str $name;`

**CST Structure:**
```
source_file [0:16]
└── expression_statement [0:15]
    └── variable_declaration [0:15]
        ├── field [0:5] "field"
        ├── type_expression [6:9]
        │   └── simple_type [6:9] "Str"
        └── scalar [10:15]
            ├── $ [10:11] "$"
            └── varname [11:15] "name"
```

**Key Characteristics:**
- Similar structure to variable declaration
- Uses `field` keyword instead of `my`
- Type annotation follows same pattern

### Pattern 3: Type Assertion

**Source:** `$value as Int`

**CST Structure:**
```
source_file [0:13]
└── expression_statement [0:13]
    └── type_assertion_expression [0:13]
        ├── scalar [0:6]
        │   ├── $ [0:1] "$"
        │   └── varname [1:6] "value"
        ├── as [7:9] "as"
        └── type_expression [10:13]
            └── simple_type [10:13] "Int"
```

**Key Characteristics:**
- Dedicated `type_assertion_expression` node type
- Variable and type connected by `as` keyword
- Type annotation in same structure as declarations

### Pattern 4: Method Parameter

**Source:** `method hello(Int $a) { ... }`

**CST Structure (Parameter Only):**
```
mandatory_parameter [12:18]
├── type_expression [12:15]
│   └── simple_type [12:15] "Int"
└── scalar [16:18]
    ├── $ [16:17] "$"
    └── varname [17:18] "a"
```

**Key Characteristics:**
- `mandatory_parameter` contains typed parameter
- Same type/variable structure as declarations
- Nested within `signature` node

## Current Grammar Issues

### Issue 1: Method Return Types

**Problem:** Prefix return type syntax should be supported

**Source:** `method Str hello() { ... }`

**CST:**
```
method_declaration_statement
├── method "method"
├── bareword "hello"
├── signature "() "
├── ERROR "-> Str"  # Should be return_type node
└── block "{ ... }"
```

**Impact:** Return type information is lost in ERROR node

### Issue 2: Type Declarations

**Problem:** Type declarations parsed as function calls

**Source:** `type UserId = Int;`

**CST:**
```
ambiguous_function_call_expression  # WRONG
├── function "type"
└── assignment_expression "UserId = Int"
```

**Should Be:**
```
type_declaration_statement
├── type "type"
├── type_name "UserId"
├── = "="
└── type_expression
    └── simple_type "Int"
```

### Issue 3: Complex Type Expressions

**Problem:** Union types parsed as binary expressions

**Source:** `my (Int|Str) $value;`

**CST:**
```
type_expression
└── binary_expression  # Should be union_type
    ├── simple_type "Int"
    ├── | "|"
    └── simple_type "Str"
```

## CST Navigation Patterns

### Finding Type Annotations

To find type annotations in any construct:
1. Look for `type_expression` descendants
2. Navigate to `simple_type` child for basic types
3. Handle complex types through `binary_expression` patterns

### Extracting Variable Names

To extract variable names from declarations:
1. Find `scalar`, `array`, or `hash` descendants
2. Navigate to `varname` child
3. Extract text content

### Identifying Typed Constructs

Key patterns for typed construct identification:
1. `variable_declaration` with `type_expression` descendant
2. `type_assertion_expression` nodes
3. `mandatory_parameter` with `type_expression`

## Transformation Requirements

For unified compiler implementation, transformations needed:

### Clean Perl Output

Remove type annotation nodes while preserving:
- Variable names and sigils
- Assignment operators and values
- Comments and whitespace
- All non-type syntax

### Type Preservation

For typed Perl output:
- Preserve all type annotation nodes exactly
- Maintain source positioning
- Keep ERROR nodes for debugging

## Memory Considerations

CST nodes contain:
- **Node references**: Parent/child relationships
- **Position data**: StartByte/EndByte for every node
- **Text content**: Available on demand
- **Type information**: String-based node types

For large codebases:
- CST can be memory-intensive
- Consider streaming processing for large files
- Cache frequently accessed navigation results

## Extension Points

The CST structure supports future enhancements:

### Additional Type Constructs
- Intersection types: `Object&Serializable`
- Negation types: `!Undef`
- Generic constraints: `T where T: Serialize`

### Better Error Recovery
- Partial type information from ERROR nodes
- Graceful handling of incomplete syntax

### Source Mapping
- Exact preservation of formatting
- Comment association with nodes
- Whitespace handling improvements

## Implementation Guidelines

### CST Processing Best Practices

1. **Always check for nil nodes** - Tree-sitter can return nil
2. **Use StartByte/EndByte for positioning** - More reliable than content
3. **Handle ERROR nodes gracefully** - Extract what information is available
4. **Cache navigation results** - CST traversal can be expensive
5. **Preserve exact text when possible** - Use node.Content() for output

### Performance Considerations

1. **Minimize CST walks** - Use targeted queries instead of full traversal
2. **Batch similar operations** - Process all nodes of same type together
3. **Consider memory usage** - Large CSTs can consume significant memory
4. **Profile navigation patterns** - Optimize frequently used paths

This documentation provides the foundation for implementing the unified compiler architecture that works directly with tree-sitter CST, eliminating the need for lossy CST-to-AST conversion.
