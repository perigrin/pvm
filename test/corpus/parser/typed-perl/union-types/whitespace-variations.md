---
category: typed-perl
subcategory: union-types
tags:
    - whitespace
    - formatting
    - union-types
type_check: true
---

# Whitespace Variations

Union types with different whitespace formatting variations

```perl
my Int | Str $spaced;
my Int|Str|Bool $compact;
my  Num  |  Str  |  Bool  $extra_spaced;
my Int|
    Str|
    Bool $multiline;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 126 characters
  Type Annotations:
    VarAnnotation: $spaced :: Int | Str at 1:1
    VarAnnotation: $compact :: Int|Str|Bool at 2:1
    VarAnnotation: $extra_spaced :: Num  |  Str  |  Bool at 3:1
    VarAnnotation: $multiline :: Int|
    Str|
    Bool at 4:1
  Root: source_file
  Tree Structure:
  source_file
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
      variable_declaration
        token
        type_expression
          union_type
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
          union_type
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
          union_type
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
  Source length: 126 characters
  Type Annotations:
    VarAnnotation: $spaced :: Int | Str at 1:1
    VarAnnotation: $compact :: Int|Str|Bool at 2:1
    VarAnnotation: $extra_spaced :: Num  |  Str  |  Bool at 3:1
    VarAnnotation: $multiline :: Int|
    Str|
    Bool at 4:1
  Root: source_file
  Tree Structure:
  source_file
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
      variable_declaration
        token
        type_expression
          union_type
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
          union_type
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
          union_type
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
