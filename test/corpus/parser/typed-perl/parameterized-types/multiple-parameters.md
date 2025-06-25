---
category: typed-perl
subcategory: parameterized-types
tags:
    - multiple-parameters
    - Map
    - Tuple
    - Function
    - parameterized-types
type_check: true
---

# Multiple Parameters

Parameterized types with multiple type parameters

```perl
my Map[Str, Int] %mapping;
my Tuple[Int, Str, Bool] $triple;
my Function[Int, Str, Bool] $complex_func;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 103 characters
  Type Annotations:
    VarAnnotation: %mapping :: Map[Str, Int] at 1:1
    VarAnnotation: $triple :: Tuple[Int, Str, Bool] at 2:1
    VarAnnotation: $complex_func :: Function[Int, Str, Bool] at 3:1
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
                expression_stmt
                  literal
              expression_stmt
                literal
              type_expression
                expression_stmt
                  literal
            expression_stmt
              literal
        hash
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
  Source length: 103 characters
  Type Annotations:
    VarAnnotation: %mapping :: Map[Str, Int] at 1:1
    VarAnnotation: $triple :: Tuple[Int, Str, Bool] at 2:1
    VarAnnotation: $complex_func :: Function[Int, Str, Bool] at 3:1
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
                expression_stmt
                  literal
              expression_stmt
                literal
              type_expression
                expression_stmt
                  literal
            expression_stmt
              literal
        hash
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
