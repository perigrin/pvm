---
category: typed-perl
subcategory: methods-fields
tags:
    - complex-fields
    - parameterized-types
    - custom-types
    - field-declarations
type_check: true
---

# Complex Field Types

Field declarations with complex parameterized types and custom types

```perl
field ArrayRef[Int] $numbers = [];
field HashRef[Str] $config = {};
field CodeRef[Int, Str] $formatter;
field ArrayRef[MyType] $items;
field HashRef[ArrayRef[Str]] $grouped_data = {};
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 183 characters
  Type Annotations:
    VarAnnotation: $numbers :: ArrayRef[Int] at 1:1
    VarAnnotation: $config :: HashRef[Str] at 2:1
    VarAnnotation: $formatter :: CodeRef[Int, Str] at 3:1
    VarAnnotation: $items :: ArrayRef[MyType] at 4:1
    VarAnnotation: $grouped_data :: HashRef[ArrayRef[Str]] at 5:1
    FieldAnnotation: $numbers :: ArrayRef[Int] at 1:1
    FieldAnnotation: $config :: HashRef[Str] at 2:1
    FieldAnnotation: Str] :: CodeRef[Int, at 3:1
    FieldAnnotation: $items :: ArrayRef[MyType] at 4:1
    FieldAnnotation: $grouped_data :: HashRef[ArrayRef[Str]] at 5:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      assignment_expression
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
        anonymous_array_expression
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
        scalar
          token
          token
    token
    expression_statement
      assignment_expression
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
  Source length: 183 characters
  Type Annotations:
    VarAnnotation: $numbers :: ArrayRef[Int] at 1:1
    VarAnnotation: $config :: HashRef[Str] at 2:1
    VarAnnotation: $formatter :: CodeRef[Int, Str] at 3:1
    VarAnnotation: $items :: ArrayRef[MyType] at 4:1
    VarAnnotation: $grouped_data :: HashRef[ArrayRef[Str]] at 5:1
    FieldAnnotation: $numbers :: ArrayRef[Int] at 1:1
    FieldAnnotation: $config :: HashRef[Str] at 2:1
    FieldAnnotation: Str] :: CodeRef[Int, at 3:1
    FieldAnnotation: $items :: ArrayRef[MyType] at 4:1
    FieldAnnotation: $grouped_data :: HashRef[ArrayRef[Str]] at 5:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      assignment_expression
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
        anonymous_array_expression
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
        scalar
          token
          token
    token
    expression_statement
      assignment_expression
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
        anonymous_hash_expression
          token
          token
    token
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
field $numbers = [];
field $config = {};
field $formatter;
field $items;
field $grouped_data = {};
```

## Typed Perl Output

```perl
field ArrayRef[Int] $numbers = [];
field HashRef[Str] $config = {};
field CodeRef[Int, Str] $formatter;
field ArrayRef[MyType] $items;
field HashRef[ArrayRef[Str]] $grouped_data = {};
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
