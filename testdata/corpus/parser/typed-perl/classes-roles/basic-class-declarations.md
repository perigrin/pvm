---
category: typed-perl
subcategory: classes-roles
tags:
    - class-declaration
    - typed-fields
    - typed-methods
    - basic
type_check: true
---

# Basic Class Declarations

Basic class with typed fields and methods

```perl
class User {
    field Str $name;
    field Int $age;
    field Optional[Email] $email;

    method User new(Str $name, Int $age) {
        return bless {
            name => $name,
            age => $age
        }, __PACKAGE__;
    }

    method Str get_name() {
        return $name;
    }
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 294 characters
  Type Annotations:
    MethodReturnAnnotation: new :: User at 6:12
    MethodParamAnnotation: $name :: Str at 6:21
    MethodParamAnnotation: $age :: Int at 6:32
    MethodReturnAnnotation: get_name :: Str at 13:12
    VarAnnotation: User :: class at 1:1
    VarAnnotation: $name :: Str at 2:5
    VarAnnotation: $age :: Int at 3:5
    VarAnnotation: $email :: Optional[Email] at 4:5
  Root: source_file
  Tree Structure:
  source_file
    class_decl
      field_decl
        variable
      field_decl
        variable
      field_decl
        variable
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
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 294 characters
  Type Annotations:
    MethodReturnAnnotation: new :: User at 6:12
    MethodParamAnnotation: $name :: Str at 6:21
    MethodParamAnnotation: $age :: Int at 6:32
    MethodReturnAnnotation: get_name :: Str at 13:12
    VarAnnotation: User :: class at 1:1
    VarAnnotation: $name :: Str at 2:5
    VarAnnotation: $age :: Int at 3:5
    VarAnnotation: $email :: Optional[Email] at 4:5
  Root: source_file
  Tree Structure:
  source_file
    class_decl
      field_decl
        variable
      field_decl
        variable
      field_decl
        variable
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
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
class User {
    field $name;
    field $age;
    field $email;

    method new($name, $age) {
        return bless {
            name => $name,
            age => $age
        }, __PACKAGE__;
    }

    method get_name() {
        return $name;
    }
}
```

## Typed Perl Output

```perl
class User {
    field Str $name;
    field Int $age;
    field Optional[Email] $email;

    method User new(Str $name, Int $age) {
        return bless {
            name => $name,
            age => $age
        }, __PACKAGE__;
    }

    method Str get_name() {
        return $name;
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
