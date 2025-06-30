---
category: typed-perl
subcategory: simple-annotations
tags:
    - typed-variables
    - built-in-types
    - variable-declarations
type_check: true
---

# Basic Typed Variables

Basic typed variable declarations with built-in types

```perl
my Int $count = 42;
my Str $name = "example";
my Bool $flag = 1;
my Num $pi = 3.14159;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 86 characters
  Type Annotations:
    VarAnnotation: $count :: Int at 1:1
    VarAnnotation: $name :: Str at 2:1
    VarAnnotation: $flag :: Bool at 3:1
    VarAnnotation: $pi :: Num at 4:1
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
  Source length: 86 characters
  Type Annotations:
    VarAnnotation: $count :: Int at 1:1
    VarAnnotation: $name :: Str at 2:1
    VarAnnotation: $flag :: Bool at 3:1
    VarAnnotation: $pi :: Num at 4:1
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
my $count = 42;
my $name = "example";
my $flag = 1;
my $pi = 3.14159;
```

## Typed Perl Output

```perl
my Int $count = 42;
my Str $name = "example";
my Bool $flag = 1;
my Num $pi = 3.14159;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
