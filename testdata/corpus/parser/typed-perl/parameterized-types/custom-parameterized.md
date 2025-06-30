---
category: typed-perl
subcategory: parameterized-types
tags:
    - custom-types
    - generics
    - package-qualified
    - parameterized-types
type_check: true
---

# Custom Parameterized

Custom parameterized types and package-qualified generics

```perl
my Container[MyType] $custom_container;
my Package::Generic[Int] $qualified;
my Result[UserData, ErrorCode] $result;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 116 characters
  Type Annotations:
    VarAnnotation: $custom_container :: Container[MyType] at 1:1
    VarAnnotation: $qualified :: Package::Generic[Int] at 2:1
    VarAnnotation: $result :: Result[UserData, ErrorCode] at 3:1
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
  Source length: 116 characters
  Type Annotations:
    VarAnnotation: $custom_container :: Container[MyType] at 1:1
    VarAnnotation: $qualified :: Package::Generic[Int] at 2:1
    VarAnnotation: $result :: Result[UserData, ErrorCode] at 3:1
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
        scalar
          token
          token
    token
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $custom_container;
my $qualified;
my $result;
```

## Typed Perl Output

```perl
my Container[MyType] $custom_container;
my Package::Generic[Int] $qualified;
my Result[UserData, ErrorCode] $result;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
