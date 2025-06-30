---
category: typed-perl
subcategory: simple-annotations
tags:
    - scoping
    - our
    - state
    - local
    - typed-variables
type_check: true
---

# Scoping Keywords

Type annotations with different scoping keywords (our, state, local)

```perl
our Int $global_counter = 0;
state Str $persistent_cache = "";
local Num $localized = 1.0;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 90 characters
  Type Annotations:
    VarAnnotation: $global_counter :: Int at 1:1
    VarAnnotation: $persistent_cache :: Str at 2:1
    VarAnnotation: $localized :: Num at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      var_decl
        variable
    token
    expression_statement
      var_decl
        variable
    token
    expression_statement
      localization_expression
        expression_stmt
          literal
        ambiguous_function_call_expression
          expression_stmt
            literal
          assignment_expression
            scalar
              token
              token
            token
            token
    token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 90 characters
  Type Annotations:
    VarAnnotation: $global_counter :: Int at 1:1
    VarAnnotation: $persistent_cache :: Str at 2:1
    VarAnnotation: $localized :: Num at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      var_decl
        variable
    token
    expression_statement
      var_decl
        variable
    token
    expression_statement
      localization_expression
        expression_stmt
          literal
        ambiguous_function_call_expression
          expression_stmt
            literal
          assignment_expression
            scalar
              token
              token
            token
            token
    token
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
our $global_counter = 0;
state $persistent_cache = "";
local $localized = 1.0;
```

## Typed Perl Output

```perl
our Int $global_counter = 0;
state Str $persistent_cache = "";
local Num $localized = 1.0;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
