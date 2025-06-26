---
category: typed-perl
subcategory: simple-annotations
tags:
    - complex-assignments
    - expressions
    - typed-variables
type_check: true
---

# Complex Assignments

Type annotations with complex assignment expressions

```perl
my Int $calculated = $base + $increment;
my Str $interpolated = "Value: $count";
my Bool $comparison = $a > $b;
my Num $result = $x * $y + $z;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 142 characters
  Type Annotations:
    VarAnnotation: $calculated :: Int at 1:1
    VarAnnotation: $interpolated :: Str at 2:1
    VarAnnotation: $comparison :: Bool at 3:1
    VarAnnotation: $result :: Num at 4:1
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
  Source length: 142 characters
  Type Annotations:
    VarAnnotation: $calculated :: Int at 1:1
    VarAnnotation: $interpolated :: Str at 2:1
    VarAnnotation: $comparison :: Bool at 3:1
    VarAnnotation: $result :: Num at 4:1
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
my $calculated = $base + $increment;
my $interpolated = "Value: $count";
my $comparison = $a > $b;
my $result = $x * $y + $z;
```

## Typed Perl Output

```perl
my Int $calculated = $base + $increment;
my Str $interpolated = "Value: $count";
my Bool $comparison = $a > $b;
my Num $result = $x * $y + $z;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
