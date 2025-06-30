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
method get_number() returns Int {
    return 42;
}

method get_array() returns ArrayRef[Str] {
    return ["a", "b", "c"];
}

method get_hash() returns HashRef[Int] {
    return { count => 5, total => 100 };
}

method get_nothing() returns Void {
    # Side effects only
    print "Done\n";
}

method get_optional() returns Optional[Str] {
    return undef;
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 359 characters
  Type Annotations:
    MethodReturnAnnotation: get_number :: Int at 1:29
    MethodReturnAnnotation: get_array :: ArrayRef[Str] at 5:28
    MethodReturnAnnotation: get_hash :: HashRef[Int] at 9:27
    MethodReturnAnnotation: get_nothing :: Void at 13:30
    MethodReturnAnnotation: get_optional :: Optional[Str] at 18:31
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
  Source length: 359 characters
  Type Annotations:
    MethodReturnAnnotation: get_number :: Int at 1:29
    MethodReturnAnnotation: get_array :: ArrayRef[Str] at 5:28
    MethodReturnAnnotation: get_hash :: HashRef[Int] at 9:27
    MethodReturnAnnotation: get_nothing :: Void at 13:30
    MethodReturnAnnotation: get_optional :: Optional[Str] at 18:31
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
{ return 42; }{ return ["a", "b", "c"]; }{ return { count => 5, total => 100 } }{ # Side effects only; print "Done\n"; }{ return undef; }
```

## Typed Perl Output

```perl
method get_number() returns Int {
    return 42;
}

method get_array() returns ArrayRef[Str] {
    return ["a", "b", "c"];
}

method get_hash() returns HashRef[Int] {
    return { count => 5, total => 100 };
}

method get_nothing() returns Void {
    # Side effects only
    print "Done\n";
}

method get_optional() returns Optional[Str] {
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
