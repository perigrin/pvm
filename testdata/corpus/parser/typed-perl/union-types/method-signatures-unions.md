---
category: typed-perl
subcategory: union-types
tags:
    - method-signatures
    - parameter-types
    - return-types
    - union-types
type_check: true
---

# Method Signatures Unions

Union types in method parameter and return type signatures

```perl
method Bool|Str process(Int|Str $input) {
    if (ref $input) {
        return "Invalid";
    }
    return 1;
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 111 characters
  Type Annotations:
    MethodReturnAnnotation: process :: Bool|Str at 1:8
    MethodParamAnnotation: $input :: Int|Str at 1:1
    MethodParamAnnotation: $input :: Int|Str at 1:1
  Root: source_file
  Tree Structure:
  source_file
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
  Source length: 111 characters
  Type Annotations:
    MethodReturnAnnotation: process :: Bool|Str at 1:8
    MethodParamAnnotation: $input :: Int|Str at 1:1
    MethodParamAnnotation: $input :: Int|Str at 1:1
  Root: source_file
  Tree Structure:
  source_file
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
method process($input) {
    if (ref $input) {
        return "Invalid";
    }
    return 1;
}
```

## Typed Perl Output

```perl
method Bool|Str process(Int|Str $input) {
    if (ref $input) {
        return "Invalid";
    }
    return 1;
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
