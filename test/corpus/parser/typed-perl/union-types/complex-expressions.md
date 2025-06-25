---
category: typed-perl
subcategory: union-types
tags:
    - complex-expressions
    - parameterized-types
    - union-types
    - fields
type_check: true
---

# Complex Expressions

Union types within complex type expressions like parameterized types

```perl
my ArrayRef[Int|Str] @mixed_array;
field HashRef[Int|Bool] $mixed_hash;
my CodeRef[Int|Str, Bool|Undef] $flexible_function;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 123 characters
  Type Annotations:
    VarAnnotation: @mixed_array :: ArrayRef[Int|Str] at 1:1
    VarAnnotation: $mixed_hash :: HashRef[Int|Bool] at 2:1
    VarAnnotation: $flexible_function :: CodeRef[Int|Str, Bool|Undef] at 3:1
    FieldAnnotation: $mixed_hash :: HashRef[Int|Bool] at 2:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      variable_declaration
        token
        type_expression
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                union_type
                  type_expression
                    expression_stmt
                      literal
                  expression_stmt
                    literal
                  type_expression
                    expression_stmt
                      literal
            expression_stmt
              literal
        array
          expression_stmt
            literal
          token
    token
    expression_statement
      variable_declaration
        token
        type_expression
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                union_type
                  type_expression
                    expression_stmt
                      literal
                  expression_stmt
                    literal
                  type_expression
                    expression_stmt
                      literal
            expression_stmt
              literal
        scalar
          token
          token
    token
    expression_statement
      variable_declaration
        token
        type_expression
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                union_type
                  type_expression
                    expression_stmt
                      literal
                  expression_stmt
                    literal
                  type_expression
                    expression_stmt
                      literal
              expression_stmt
                literal
              type_expression
                union_type
                  type_expression
                    expression_stmt
                      literal
                  expression_stmt
                    literal
                  type_expression
                    expression_stmt
                      literal
            expression_stmt
              literal
        scalar
          token
          token
    token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 123 characters
  Type Annotations:
    VarAnnotation: @mixed_array :: ArrayRef[Int|Str] at 1:1
    VarAnnotation: $mixed_hash :: HashRef[Int|Bool] at 2:1
    VarAnnotation: $flexible_function :: CodeRef[Int|Str, Bool|Undef] at 3:1
    FieldAnnotation: $mixed_hash :: HashRef[Int|Bool] at 2:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      variable_declaration
        token
        type_expression
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                union_type
                  type_expression
                    expression_stmt
                      literal
                  expression_stmt
                    literal
                  type_expression
                    expression_stmt
                      literal
            expression_stmt
              literal
        array
          expression_stmt
            literal
          token
    token
    expression_statement
      variable_declaration
        token
        type_expression
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                union_type
                  type_expression
                    expression_stmt
                      literal
                  expression_stmt
                    literal
                  type_expression
                    expression_stmt
                      literal
            expression_stmt
              literal
        scalar
          token
          token
    token
    expression_statement
      variable_declaration
        token
        type_expression
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                union_type
                  type_expression
                    expression_stmt
                      literal
                  expression_stmt
                    literal
                  type_expression
                    expression_stmt
                      literal
              expression_stmt
                literal
              type_expression
                union_type
                  type_expression
                    expression_stmt
                      literal
                  expression_stmt
                    literal
                  type_expression
                    expression_stmt
                      literal
            expression_stmt
              literal
        scalar
          token
          token
    token
}
```

# Expected Type Errors

```
(none)
```
