---
category: typed-perl
subcategory: classes-roles
tags:
    - generic-class
    - type-parameters
    - type-constraints
    - parameterized-methods
type_check: true
---

# Generic Class Declarations

Generic class with type parameters and constraints

```perl
class Container<T> where T: Serializable {
    field ArrayRef[T] $items = [];

    method Void add(T $item) {
        push @{$items}, $item;
    }

    method ArrayRef[T] get_all() {
        return $items;
    }

    method Optional[T] find(CodeRef[T, Bool] $predicate) {
        for my $item (@{$items}) {
            return $item if $predicate->($item);
        }
        return undef;
    }
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 395 characters
  Type Annotations:
    MethodReturnAnnotation: add :: Void at 4:12
    MethodReturnAnnotation: get_all :: ArrayRef[T] at 8:12
    MethodReturnAnnotation: find :: Optional[T] at 12:12
    VarAnnotation: Container :: class at 1:1
    VarAnnotation: $items :: ArrayRef[T] at 2:5
    MethodReturnAnnotation: add :: Void at 4:12
    MethodReturnAnnotation: get_all :: ArrayRef[T] at 8:12
    MethodReturnAnnotation: find :: Optional[T] at 12:12
    MethodParamAnnotation: $item :: T at 4:1
    MethodParamAnnotation: $predicate :: Bool] at 12:1
  Root: source_file
  Tree Structure:
  source_file
    class_decl
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
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 395 characters
  Type Annotations:
    MethodReturnAnnotation: add :: Void at 4:12
    MethodReturnAnnotation: get_all :: ArrayRef[T] at 8:12
    MethodReturnAnnotation: find :: Optional[T] at 12:12
    VarAnnotation: Container :: class at 1:1
    VarAnnotation: $items :: ArrayRef[T] at 2:5
    MethodReturnAnnotation: add :: Void at 4:12
    MethodReturnAnnotation: get_all :: ArrayRef[T] at 8:12
    MethodReturnAnnotation: find :: Optional[T] at 12:12
    MethodParamAnnotation: $item :: T at 4:1
    MethodParamAnnotation: $predicate :: Bool] at 12:1
  Root: source_file
  Tree Structure:
  source_file
    class_decl
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
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
class Container<T> where T: Serializable {
    field ArrayRef[T] $items = [];

    method Void add(T $item) {
        push @{$items}, $item;
    }

    method ArrayRef[T] get_all() {
        return $items;
    }

    method Optional[T] find(CodeRef[T, Bool] $predicate) {
        for my $item (@{$items}) {
            return $item if $predicate->($item);
        }
        return undef;
    }
}
```

## Typed Perl Output

```perl
class Container<T> where T: Serializable {
    field ArrayRef[T] $items = [];

    method Void add(T $item) {
        push @{$items}, $item;
    }

    method ArrayRef[T] get_all() {
        return $items;
    }

    method Optional[T] find(CodeRef[T, Bool] $predicate) {
        for my $item (@{$items}) {
            return $item if $predicate->($item);
        }
        return undef;
    }
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
