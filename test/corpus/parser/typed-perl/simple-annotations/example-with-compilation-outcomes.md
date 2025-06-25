---
category: typed-perl
subcategory: simple-annotations
tags:
    - variables
    - compilation-outcomes
type_check: true
---

# Example With Compilation Outcomes

Basic variable declaration with type annotations and expected compilation outcomes.

```perl
my Int $count = 42;
my Str $name = "example";
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 46 characters
  Type Annotations:
    VarAnnotation: $count :: Int at 1:1
    VarAnnotation: $name :: Str at 2:1
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
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 46 characters
  Type Annotations:
    VarAnnotation: $count :: Int at 1:1
    VarAnnotation: $name :: Str at 2:1
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
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $count = 42;
my $name = "example";
```

## Typed Perl Output

```perl
use v5.36;
my Int $count = 42;
my Str $name = "example";
```

## Inferred Perl Output

```perl
use v5.36;
my Int $count = 42;
my Str $name = "example";
```

# Expected Type Errors

```
(none)
```