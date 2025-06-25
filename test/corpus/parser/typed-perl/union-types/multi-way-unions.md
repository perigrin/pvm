---
category: typed-perl
subcategory: union-types
tags:
    - multi-way-unions
    - union-types
    - variable-declarations
type_check: true
---

# Multi Way Unions

Multi-way union types with three or more alternatives

```perl
my Int|Str|Bool $multi = "text";
my Num|ArrayRef|HashRef $complex;
my Int|Str|Bool|Undef $nullable = undef;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 107 characters
  Type Annotations:
    VarAnnotation: $multi :: Int|Str|Bool at 1:1
    VarAnnotation: $complex :: Num|ArrayRef|HashRef at 2:1
    VarAnnotation: $nullable :: Int|Str|Bool|Undef at 3:1
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
      var_decl
        variable
    token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 107 characters
  Type Annotations:
    VarAnnotation: $multi :: Int|Str|Bool at 1:1
    VarAnnotation: $complex :: Num|ArrayRef|HashRef at 2:1
    VarAnnotation: $nullable :: Int|Str|Bool|Undef at 3:1
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
      var_decl
        variable
    token
}
```

# Expected Type Errors

```
(none)
```
