---
category: typed-perl
subcategory: simple-annotations
tags:
    - arrays
    - hashes
    - parameterized-types
    - typed-variables
type_check: true
---

# Typed Arrays Hashes

Typed array and hash declarations with basic parameterized types

```perl
my ArrayRef[Int] @numbers = (1, 2, 3);
my HashRef[Str] %config = (key => 'value');
my ArrayRef[Str] @strings = ("a", "b", "c");
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 127 characters
  Type Annotations:
    VarAnnotation: @numbers :: ArrayRef[Int] at 1:1
    VarAnnotation: %config :: HashRef[Str] at 2:1
    VarAnnotation: @strings :: ArrayRef[Str] at 3:1
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
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 127 characters
  Type Annotations:
    VarAnnotation: @numbers :: ArrayRef[Int] at 1:1
    VarAnnotation: %config :: HashRef[Str] at 2:1
    VarAnnotation: @strings :: ArrayRef[Str] at 3:1
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
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
my ArrayRef[Int] @numbers = (1, 2, 3);
my HashRef[Str] %config = (key => 'value');
my ArrayRef[Str] @strings = ("a", "b", "c");
```

## Typed Perl Output

```perl
my ArrayRef[Int] @numbers = (1, 2, 3);
my HashRef[Str] %config = (key => 'value');
my ArrayRef[Str] @strings = ("a", "b", "c");
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
