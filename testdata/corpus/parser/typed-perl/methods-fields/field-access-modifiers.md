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
field readonly Int $readonly_field;
field static Str $class_field;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 221 characters
  Type Annotations:
    VarAnnotation: $public_field :: Int at 2:1
    VarAnnotation: $private_field :: private at 3:1
    VarAnnotation: $protected_field :: protected at 4:1
    VarAnnotation: $readonly_field :: readonly at 5:1
    VarAnnotation: $class_field :: static at 6:1
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
  Source length: 221 characters
  Type Annotations:
    VarAnnotation: $public_field :: Int at 2:1
    VarAnnotation: $private_field :: private at 3:1
    VarAnnotation: $protected_field :: protected at 4:1
    VarAnnotation: $readonly_field :: readonly at 5:1
    VarAnnotation: $class_field :: static at 6:1
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


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
# Different field visibility patterns
field $public_field = 1;
field $private_field = "secret";
field $protected_field = 0;
field readonly $readonly_field = [1, 2, 3];
field static $class_field = {};
```

## Typed Perl Output

```perl
# Different field visibility patterns
field Int $public_field = 1;
field private Str $private_field = "secret";
field protected Bool $protected_field = 0;
field readonly Int $readonly_field;
field static Str $class_field;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
