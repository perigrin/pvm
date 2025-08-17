---
category: typed-perl
subcategory: simple-annotations
tags:
    - optional-types
    - maybe-types
    - undefined
    - typed-variables
type_check: true
---

# Optional Annotations

Type annotations including optional/maybe types and undefined values

```perl
my Int $with_type;
my Str $also_typed = undef;
my Optional[Int] $maybe_int;
my Maybe[Str] $maybe_str;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 101 characters
  Type Annotations:
    VarAnnotation: $with_type :: Int at 1:1
    VarAnnotation: $also_typed :: Str at 2:1
    VarAnnotation: $maybe_int :: Optional[Int] at 3:1
    VarAnnotation: $maybe_str :: Maybe[Str] at 4:1
  Root: source_file
  Tree Structure:
  source_file
    expression_stmt
      literal
    token
    expression_statement
      var_decl
        variable
        literal
    token
    expression_stmt
      literal
    token
    expression_stmt
      literal
    token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 101 characters
  Type Annotations:
    VarAnnotation: $with_type :: Int at 1:1
    VarAnnotation: $also_typed :: Str at 2:1
    VarAnnotation: $maybe_int :: Optional[Int] at 3:1
    VarAnnotation: $maybe_str :: Maybe[Str] at 4:1
  Root: source_file
  Tree Structure:
  source_file
    expression_stmt
      literal
    token
    expression_statement
      var_decl
        variable
        literal
    token
    expression_stmt
      literal
    token
    expression_stmt
      literal
    token
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $with_type;
my $also_typed = undef;
my $maybe_int;
my $maybe_str;
```

## Typed Perl Output

```perl
my Int $with_type;
my Str $also_typed = undef;
my Optional[Int] $maybe_int;
my Maybe[Str] $maybe_str;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
