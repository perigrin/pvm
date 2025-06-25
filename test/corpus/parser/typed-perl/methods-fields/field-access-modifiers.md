---
category: typed-perl
subcategory: methods-fields
tags:
    - field-modifiers
    - field-visibility
    - access-control
type_check: true
---

# Field Access Modifiers

Field declarations with access modifiers and visibility keywords

```perl
# Different field visibility patterns
field Int $public_field = 1;
field private Str $private_field = "secret";
field protected Bool $protected_field = 0;
field readonly ArrayRef[Int] $readonly_field = [1, 2, 3];
field static HashRef[Str] $class_field = {};
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 257 characters
  Type Annotations:
    VarAnnotation: $public_field :: Int at 2:1
    VarAnnotation: $private_field :: private at 3:1
    VarAnnotation: $protected_field :: protected at 4:1
    VarAnnotation: $readonly_field :: ArrayRef[Int] at 5:1
    VarAnnotation: $class_field :: HashRef[Str] at 6:1
    FieldAnnotation: $public_field :: Int at 2:1
    FieldAnnotation: Str :: private at 3:1
    FieldAnnotation: Bool :: protected at 4:1
    FieldAnnotation: ArrayRef[Int] :: readonly at 5:1
    FieldAnnotation: HashRef[Str] :: static at 6:1
  Root: source_file
  Tree Structure:
  source_file
    expression_stmt
      literal
    expression_statement
      assignment_expression
        variable_declaration
          token
          type_expression
            expression_stmt
              literal
          scalar
            token
            token
        token
        token
    token
    expression_statement
      assignment_expression
        variable_declaration
          token
          type_expression
            expression_stmt
              literal
          expression_stmt
            literal
          scalar
            token
            token
        token
        interpolated_string_literal
          expression_stmt
            literal
          expression_stmt
            literal
          expression_stmt
            literal
    token
    expression_statement
      assignment_expression
        variable_declaration
          token
          type_expression
            expression_stmt
              literal
          expression_stmt
            literal
          scalar
            token
            token
        token
        token
    token
    expression_statement
      assignment_expression
        variable_declaration
          token
          expression_stmt
            literal
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
          scalar
            token
            token
        token
        anonymous_array_expression
          expression_stmt
            literal
          list_expression
            token
            expression_stmt
              literal
            token
            expression_stmt
              literal
            token
          expression_stmt
            literal
    token
    expression_statement
      assignment_expression
        variable_declaration
          token
          expression_stmt
            literal
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
          scalar
            token
            token
        token
        anonymous_hash_expression
          token
          token
    token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 257 characters
  Type Annotations:
    VarAnnotation: $public_field :: Int at 2:1
    VarAnnotation: $private_field :: private at 3:1
    VarAnnotation: $protected_field :: protected at 4:1
    VarAnnotation: $readonly_field :: ArrayRef[Int] at 5:1
    VarAnnotation: $class_field :: HashRef[Str] at 6:1
    FieldAnnotation: $public_field :: Int at 2:1
    FieldAnnotation: Str :: private at 3:1
    FieldAnnotation: Bool :: protected at 4:1
    FieldAnnotation: ArrayRef[Int] :: readonly at 5:1
    FieldAnnotation: HashRef[Str] :: static at 6:1
  Root: source_file
  Tree Structure:
  source_file
    expression_stmt
      literal
    expression_statement
      assignment_expression
        variable_declaration
          token
          type_expression
            expression_stmt
              literal
          scalar
            token
            token
        token
        token
    token
    expression_statement
      assignment_expression
        variable_declaration
          token
          type_expression
            expression_stmt
              literal
          expression_stmt
            literal
          scalar
            token
            token
        token
        interpolated_string_literal
          expression_stmt
            literal
          expression_stmt
            literal
          expression_stmt
            literal
    token
    expression_statement
      assignment_expression
        variable_declaration
          token
          type_expression
            expression_stmt
              literal
          expression_stmt
            literal
          scalar
            token
            token
        token
        token
    token
    expression_statement
      assignment_expression
        variable_declaration
          token
          expression_stmt
            literal
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
          scalar
            token
            token
        token
        anonymous_array_expression
          expression_stmt
            literal
          list_expression
            token
            expression_stmt
              literal
            token
            expression_stmt
              literal
            token
          expression_stmt
            literal
    token
    expression_statement
      assignment_expression
        variable_declaration
          token
          expression_stmt
            literal
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
          scalar
            token
            token
        token
        anonymous_hash_expression
          token
          token
    token
}
```

# Expected Type Errors

```
(none)
```
