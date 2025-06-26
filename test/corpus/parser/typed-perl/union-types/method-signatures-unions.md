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
method process(Int|Str $input) returns Bool|Str {
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
  Source length: 119 characters
  Type Annotations:
    MethodReturnAnnotation: process :: Bool|Str at 1:40
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
  Source length: 119 characters
  Type Annotations:
    MethodReturnAnnotation: process :: Bool|Str at 1:40
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
{ if (ref $input) {
        return "Invalid";
    } return 1; }
```

## Typed Perl Output

```perl
method process(Int|Str $input) returns Bool|Str {
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
