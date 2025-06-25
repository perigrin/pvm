---
category: typed-perl
subcategory: classes-roles
tags:
    - generic-role
    - type-parameters
    - type-constraints
    - parameterized-role-methods
type_check: true
---

# Generic Role Declarations

Generic roles with type parameters and constraints

```perl
role Processable<T> where T: Defined {
    method process(T $input) returns ProcessResult;
    method validate(T $input) returns Bool;
}

role EventHandler<T> where T: Event {
    field ArrayRef[CodeRef[T, Void]] $handlers = [];

    method add_handler(CodeRef[T, Void] $handler) returns Void {
        push @{$handlers}, $handler;
    }

    method handle_event(T $event) returns Void {
        for my $handler (@{$handlers}) {
            $handler->($event);
        }
    }

    method handler_count() returns Int {
        return scalar @{$handlers};
    }
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 562 characters
  Type Annotations:
    MethodReturnAnnotation: process :: ProcessResult at 2:38
    MethodReturnAnnotation: validate :: Bool at 3:39
    VarAnnotation: $handlers :: ArrayRef[CodeRef[T, Void]] at 7:5
    MethodReturnAnnotation: add_handler :: Void at 9:59
    MethodReturnAnnotation: handle_event :: Void at 13:43
    MethodReturnAnnotation: handler_count :: Int at 19:36
    MethodParamAnnotation: $input :: T at 2:1
    MethodParamAnnotation: $input :: T at 3:1
    FieldAnnotation: Void]] :: ArrayRef[CodeRef[T, at 7:1
    MethodParamAnnotation: $handler :: Void] at 9:1
    MethodParamAnnotation: $event :: T at 13:1
  Root: source_file
  Tree Structure:
  source_file
    role_decl
      method_decl
      method_decl
    role_decl
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
  Source length: 562 characters
  Type Annotations:
    MethodReturnAnnotation: process :: ProcessResult at 2:38
    MethodReturnAnnotation: validate :: Bool at 3:39
    VarAnnotation: $handlers :: ArrayRef[CodeRef[T, Void]] at 7:5
    MethodReturnAnnotation: add_handler :: Void at 9:59
    MethodReturnAnnotation: handle_event :: Void at 13:43
    MethodReturnAnnotation: handler_count :: Int at 19:36
    MethodParamAnnotation: $input :: T at 2:1
    MethodParamAnnotation: $input :: T at 3:1
    FieldAnnotation: Void]] :: ArrayRef[CodeRef[T, at 7:1
    MethodParamAnnotation: $handler :: Void] at 9:1
    MethodParamAnnotation: $event :: T at 13:1
  Root: source_file
  Tree Structure:
  source_file
    role_decl
      method_decl
      method_decl
    role_decl
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
role Processable<T> where T: Defined {
    method process(T $input) returns ProcessResult;
    method validate(T $input) returns Bool;
}

{ push @{$handlers}, $handler; }{ for my $handler (@{$handlers}) {
            $handler->($event);
        } }{ return scalar @{$handlers} }
```

## Typed Perl Output

```perl
role Processable<T> where T: Defined {
    method process(T $input) returns ProcessResult;
    method validate(T $input) returns Bool;
}

role EventHandler<T> where T: Event {
    field ArrayRef[CodeRef[T, Void]] $handlers = [];

    method add_handler(CodeRef[T, Void] $handler) returns Void {
        push @{$handlers}, $handler;
    }

    method handle_event(T $event) returns Void {
        for my $handler (@{$handlers}) {
            $handler->($event);
        }
    }

    method handler_count() returns Int {
        return scalar @{$handlers};
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
