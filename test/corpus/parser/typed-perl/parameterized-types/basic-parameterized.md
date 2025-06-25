---
category: typed-perl
subcategory: parameterized-types
tags:
    - basic-parameters
    - ArrayRef
    - HashRef
    - CodeRef
    - parameterized-types
type_check: true
---

# Basic Parameterized

Basic parameterized types with single type parameters

```perl
my ArrayRef[Int] @numbers;
my HashRef[Str] %strings;
my CodeRef[Int, Str] $function;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 84 characters
  Type Annotations:
    VarAnnotation: @numbers :: ArrayRef[Int] at 1:1
    VarAnnotation: %strings :: HashRef[Str] at 2:1
    VarAnnotation: $function :: CodeRef[Int, Str] at 3:1
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
  Source length: 84 characters
  Type Annotations:
    VarAnnotation: @numbers :: ArrayRef[Int] at 1:1
    VarAnnotation: %strings :: HashRef[Str] at 2:1
    VarAnnotation: $function :: CodeRef[Int, Str] at 3:1
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
        scalar
          token
          token
    token
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
my @numbers;
my %strings;
my $function;
```

## Typed Perl Output

```perl
my ArrayRef[Int] @numbers;
my HashRef[Str] %strings;
my CodeRef[Int, Str] $function;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
