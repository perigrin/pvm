---
category: typed-perl
subcategory: parameterized-types
tags:
    - method-signatures
    - return-types
    - parameters
    - parameterized-types
type_check: true
---

# Method Signatures

Parameterized types in method signatures

```perl
method HashRef[Int] process(ArrayRef[Str] $input) {
  return {};
}

method ArrayRef[Result[Str, Error]] transform(Map[Str, Int] $data) {
  return [];
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 151 characters
  Type Annotations:
    MethodReturnAnnotation: process :: HashRef[Int] at 1:8
    MethodParamAnnotation: $input :: ArrayRef[Str] at 1:29
    MethodReturnAnnotation: transform :: ArrayRef[Result[Str, Error]] at 5:8
    MethodParamAnnotation: $data :: Map[Str, Int] at 5:47
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      type_expr
      block_stmt
        token
        return_stmt
          literal
        token
        token
    method_decl
      type_expr
      block_stmt
        token
        return_stmt
          literal
        token
        token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 151 characters
  Type Annotations:
    MethodReturnAnnotation: process :: HashRef[Int] at 1:8
    MethodParamAnnotation: $input :: ArrayRef[Str] at 1:29
    MethodReturnAnnotation: transform :: ArrayRef[Result[Str, Error]] at 5:8
    MethodParamAnnotation: $data :: Map[Str, Int] at 5:47
  Root: source_file
  Tree Structure:
  source_file
    method_decl
      type_expr
      block_stmt
        token
        return_stmt
          literal
        token
        token
    method_decl
      type_expr
      block_stmt
        token
        return_stmt
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
  return {};
}

method transform($data) {
  return [];
}
```

## Typed Perl Output

```perl
method HashRef[Int] process(ArrayRef[Str] $input) {
  return {};
}

method ArrayRef[Result[Str, Error]] transform(Map[Str, Int] $data) {
  return [];
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

**Note**: Method signatures with parameterized return types and typed parameters are not yet supported by the tree-sitter grammar. This syntax would currently produce parse errors and cannot generate a meaningful AST.

```
(parse error - method signatures not supported)
```

## JSON AST

```json
{
  "error": "Method signatures with parameterized types not yet supported by grammar",
  "note": "This syntax requires grammar extensions for: method declarations with return types, parameterized types (Type[Param]), and typed method parameters"
}
```

# Expected Type Errors

```
(none)
```
