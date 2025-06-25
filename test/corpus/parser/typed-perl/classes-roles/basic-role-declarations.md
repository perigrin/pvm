---
category: typed-perl
subcategory: classes-roles
tags:
    - role-declaration
    - required-methods
    - provided-methods
    - basic
type_check: true
---

# Basic Role Declarations

Basic role declarations with required and provided methods

```perl
role Serializable {
    method serialize() returns Str;
    method deserialize(Str $data) returns Self;
}

role Cacheable {
    field Optional[DateTime] $cached_at;

    method cache_key() returns Str;

    method is_stale() returns Bool {
        return 0 unless defined $cached_at;
        return time() - $cached_at->epoch > 3600;
    }

    method invalidate() returns Void {
        $cached_at = undef;
    }
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 415 characters
  Type Annotations:
    MethodReturnAnnotation: serialize :: Str at 2:32
    MethodReturnAnnotation: deserialize :: Self at 3:43
    VarAnnotation: $cached_at :: Optional[DateTime] at 7:5
    MethodReturnAnnotation: cache_key :: Str at 9:32
    MethodReturnAnnotation: is_stale :: Bool at 11:31
    MethodReturnAnnotation: invalidate :: Void at 16:33
    MethodParamAnnotation: $data :: Str at 3:1
    FieldAnnotation: $cached_at :: Optional[DateTime] at 7:1
  Root: source_file
  Tree Structure:
  source_file
    role_decl
      method_decl
      method_decl
    role_decl
      method_decl
      method_decl
        block_stmt
          token
          expression_stmt
            literal
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
  Source length: 415 characters
  Type Annotations:
    MethodReturnAnnotation: serialize :: Str at 2:32
    MethodReturnAnnotation: deserialize :: Self at 3:43
    VarAnnotation: $cached_at :: Optional[DateTime] at 7:5
    MethodReturnAnnotation: cache_key :: Str at 9:32
    MethodReturnAnnotation: is_stale :: Bool at 11:31
    MethodReturnAnnotation: invalidate :: Void at 16:33
    MethodParamAnnotation: $data :: Str at 3:1
    FieldAnnotation: $cached_at :: Optional[DateTime] at 7:1
  Root: source_file
  Tree Structure:
  source_file
    role_decl
      method_decl
      method_decl
    role_decl
      method_decl
      method_decl
        block_stmt
          token
          expression_stmt
            literal
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

# Expected Type Errors

```
(none)
```
