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

    method Bool is_stale() {
        return 0 unless defined $cached_at;
        return time() - $cached_at->epoch > 3600;
    }

    method Void invalidate() {
        $cached_at = undef;
    }
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 399 characters
  Type Annotations:
    MethodReturnAnnotation: Str :: serialize at 2:12
    MethodReturnAnnotation: Self :: deserialize at 3:43
    VarAnnotation: $cached_at :: Optional[DateTime] at 7:5
    MethodReturnAnnotation: Str :: cache_key at 9:12
    MethodReturnAnnotation: Bool :: is_stale at 11:31
    MethodReturnAnnotation: Void :: invalidate at 16:33
    MethodParamAnnotation: $data :: Str at 3:1
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
  Source length: 399 characters
  Type Annotations:
    MethodReturnAnnotation: Str :: serialize at 2:12
    MethodReturnAnnotation: Self :: deserialize at 3:43
    VarAnnotation: $cached_at :: Optional[DateTime] at 7:5
    MethodReturnAnnotation: Str :: cache_key at 9:12
    MethodReturnAnnotation: Bool :: is_stale at 11:31
    MethodReturnAnnotation: Void :: invalidate at 16:33
    MethodParamAnnotation: $data :: Str at 3:1
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


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
role Serializable {
    method serialize() returns Str;
    method deserialize(Str $data) returns Self;
}

{ return 0 unless defined $cached_at; return time() - $cached_at->epoch > 3600; }{ $cached_at = undef; }
```

## Typed Perl Output

```perl
role Serializable {
    method serialize() returns Str;
    method deserialize(Str $data) returns Self;
}

role Cacheable {
    field Optional[DateTime] $cached_at;

    method cache_key() returns Str;

    method Bool is_stale() {
        return 0 unless defined $cached_at;
        return time() - $cached_at->epoch > 3600;
    }

    method Void invalidate() {
        $cached_at = undef;
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
