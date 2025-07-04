---
category: typed-perl
subcategory: methods-fields
tags:
    - method-definitions
    - return-types
    - complex-return-types
    - parameterized-types
type_check: true
---

# Method Return Types

Methods with various return type annotations including complex types

```perl
method Int get_number() {
    return 42;
}

method ArrayRef[Str] get_array() {
    return ["a", "b", "c"];
}

method HashRef[Int] get_hash() {
    return { count => 5, total => 100 };
}

method Void get_nothing() {
    # Side effects only
    print "Done\n";
}

method Optional[Str] get_optional() {
    return undef;
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 319 characters
  Type Annotations:
    MethodReturnAnnotation: get_number :: Int at 1:8
    MethodReturnAnnotation: get_array :: ArrayRef[Str] at 5:8
    MethodReturnAnnotation: get_hash :: HashRef[Int] at 9:8
    MethodReturnAnnotation: get_nothing :: Void at 13:8
    MethodReturnAnnotation: get_optional :: Optional[Str] at 18:8
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        token
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        token
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        token
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        expression_stmt
          literal
        token
        token
    method_decl
      block_stmt
        token
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
  Source length: 319 characters
  Type Annotations:
    MethodReturnAnnotation: get_number :: Int at 1:8
    MethodReturnAnnotation: get_array :: ArrayRef[Str] at 5:8
    MethodReturnAnnotation: get_hash :: HashRef[Int] at 9:8
    MethodReturnAnnotation: get_nothing :: Void at 13:8
    MethodReturnAnnotation: get_optional :: Optional[Str] at 18:8
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        token
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        token
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        token
        token
    method_decl
      block_stmt
        token
        expression_stmt
          literal
        expression_stmt
          literal
        token
        token
    method_decl
      block_stmt
        token
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
method get_number() {
    return 42;
}

method get_array() {
    return ["a", "b", "c"];
}

method get_hash() {
    return { count => 5, total => 100 };
}

method get_nothing() {
    # Side effects only
    print "Done\n";
}

method get_optional() {
    return undef;
}
```

## Typed Perl Output

```perl
method Int get_number() {
    return 42;
}

method ArrayRef[Str] get_array() {
    return ["a", "b", "c"];
}

method HashRef[Int] get_hash() {
    return { count => 5, total => 100 };
}

method Void get_nothing() {
    # Side effects only
    print "Done\n";
}

method Optional[Str] get_optional() {
    return undef;
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
