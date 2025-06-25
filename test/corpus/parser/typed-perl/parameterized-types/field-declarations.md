---
category: typed-perl
subcategory: parameterized-types
tags:
    - field-declarations
    - class-fields
    - parameterized-types
type_check: true
---

# Field Declarations

Parameterized types in field declarations

```perl
field ArrayRef[MyClass] $objects;
field HashRef[ArrayRef[Str]] $nested_data;
field Optional[ArrayRef[Int]] $maybe_numbers;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 122 characters
  Type Annotations:
    VarAnnotation: $objects :: ArrayRef[MyClass] at 1:1
    VarAnnotation: $nested_data :: HashRef[ArrayRef[Str]] at 2:1
    VarAnnotation: $maybe_numbers :: Optional[ArrayRef[Int]] at 3:1
    FieldAnnotation: $objects :: ArrayRef[MyClass] at 1:1
    FieldAnnotation: $nested_data :: HashRef[ArrayRef[Str]] at 2:1
    FieldAnnotation: $maybe_numbers :: Optional[ArrayRef[Int]] at 3:1
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
  Source length: 122 characters
  Type Annotations:
    VarAnnotation: $objects :: ArrayRef[MyClass] at 1:1
    VarAnnotation: $nested_data :: HashRef[ArrayRef[Str]] at 2:1
    VarAnnotation: $maybe_numbers :: Optional[ArrayRef[Int]] at 3:1
    FieldAnnotation: $objects :: ArrayRef[MyClass] at 1:1
    FieldAnnotation: $nested_data :: HashRef[ArrayRef[Str]] at 2:1
    FieldAnnotation: $maybe_numbers :: Optional[ArrayRef[Int]] at 3:1
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
