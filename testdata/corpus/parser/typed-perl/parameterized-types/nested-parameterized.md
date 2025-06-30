---
category: typed-perl
subcategory: parameterized-types
tags:
    - nested-parameters
    - complex-nesting
    - ArrayRef
    - HashRef
    - parameterized-types
type_check: true
---

# Nested Parameterized

Nested parameterized types with multiple levels

```perl
my ArrayRef[ArrayRef[Int]] @matrix;
my HashRef[ArrayRef[Str]] %grouped_strings;
my ArrayRef[HashRef[Int]] @array_of_hashes;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 123 characters
  Type Annotations:
    VarAnnotation: @matrix :: ArrayRef[ArrayRef[Int]] at 1:1
    VarAnnotation: %grouped_strings :: HashRef[ArrayRef[Str]] at 2:1
    VarAnnotation: @array_of_hashes :: ArrayRef[HashRef[Int]] at 3:1
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
  Source length: 123 characters
  Type Annotations:
    VarAnnotation: @matrix :: ArrayRef[ArrayRef[Int]] at 1:1
    VarAnnotation: %grouped_strings :: HashRef[ArrayRef[Str]] at 2:1
    VarAnnotation: @array_of_hashes :: ArrayRef[HashRef[Int]] at 3:1
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
        array
          expression_stmt
            literal
          token
    token
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my @matrix;
my %grouped_strings;
my @array_of_hashes;
```

## Typed Perl Output

```perl
my ArrayRef[ArrayRef[Int]] @matrix;
my HashRef[ArrayRef[Str]] %grouped_strings;
my ArrayRef[HashRef[Int]] @array_of_hashes;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
