---
category: typed-perl
subcategory: parameterized-types
tags:
    - whitespace
    - formatting
    - parameterized-types
type_check: true
---

# Whitespace Variations

Parameterized types with various whitespace patterns

```perl
my ArrayRef[ Int ] @spaced;
my HashRef[Str] %compact;
my Map[ Str , Int ] %loose;
my ArrayRef[
  Int
] @multiline;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 114 characters
  Type Annotations:
    VarAnnotation: @spaced :: ArrayRef[ Int ] at 1:1
    VarAnnotation: %compact :: HashRef[Str] at 2:1
    VarAnnotation: %loose :: Map[ Str , Int ] at 3:1
    VarAnnotation: @multiline :: ArrayRef[
  Int
] at 4:1
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
        array
          expression_stmt
            literal
          token
    token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 114 characters
  Type Annotations:
    VarAnnotation: @spaced :: ArrayRef[ Int ] at 1:1
    VarAnnotation: %compact :: HashRef[Str] at 2:1
    VarAnnotation: %loose :: Map[ Str , Int ] at 3:1
    VarAnnotation: @multiline :: ArrayRef[
  Int
] at 4:1
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
        array
          expression_stmt
            literal
          token
    token
}
```

# Expected Type Errors

```
(none)
```
