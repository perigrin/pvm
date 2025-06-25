---
category: typed-perl
subcategory: union-types
tags:
    - simple-unions
    - union-types
    - variable-declarations
type_check: true
---

# Simple Union Types

Simple union type expressions with two types

```perl
my Int|Str $flexible = 42;
my Bool|Undef $maybe_flag;
my Num|Str $mixed_value = "text";
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 87 characters
  Type Annotations:
    VarAnnotation: $flexible :: Int|Str at 1:1
    VarAnnotation: $maybe_flag :: Bool|Undef at 2:1
    VarAnnotation: $mixed_value :: Num|Str at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      var_decl
        variable
    token
    expression_statement
      variable_declaration
        token
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
        scalar
          token
          token
    token
    expression_statement
      var_decl
        variable
    token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 87 characters
  Type Annotations:
    VarAnnotation: $flexible :: Int|Str at 1:1
    VarAnnotation: $maybe_flag :: Bool|Undef at 2:1
    VarAnnotation: $mixed_value :: Num|Str at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      var_decl
        variable
    token
    expression_statement
      variable_declaration
        token
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
        scalar
          token
          token
    token
    expression_statement
      var_decl
        variable
    token
}
```

# Expected Type Errors

```
(none)
```
