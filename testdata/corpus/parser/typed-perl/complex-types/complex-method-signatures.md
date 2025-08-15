---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - method-signatures
    - parameterized-return-types
    - complex-types
    - parameterized-types
---

# Complex Method Signatures

Complex method signatures with advanced parameter and return types

```perl
method HashRef[ArrayRef[Int]|Str] transform(ArrayRef[HashRef[Int|Str]] $input, CodeRef[Str, Bool] $validator) {
    return {};
}

method Result[Array[ProcessedData], ProcessingError] process(Map[Str, ArrayRef[Data|Error]] $complex_input, Optional[Handler[Request|Response]] $handler) {
    return success([]);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
method transform($input, $validator) {
    return {};
}

method process($complex_input, $handler) {
    return success([]);
}
```

## Typed Perl Output

```perl
method HashRef[ArrayRef[Int]|Str] transform(ArrayRef[HashRef[Int|Str]] $input, CodeRef[Str, Bool] $validator) {
    return {};
}

method Result[Array[ProcessedData], ProcessingError] process(Map[Str, ArrayRef[Data|Error]] $complex_input, Optional[Handler[Request|Response]] $handler) {
    return success([]);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Before Type Inference

```
AST {
  Path: /tmp/complex-method-signatures.pl
  Source length: 311 characters
  Type Annotations:
    MethodReturnAnnotation: transform :: HashRef[ArrayRef[Int]|Str] at 1:8
    MethodParamAnnotation: $input :: ArrayRef[HashRef[Int|Str]] at 1:45
    MethodParamAnnotation: $validator :: CodeRef[Str, Bool] at 1:80
    MethodReturnAnnotation: process :: Result[Array[ProcessedData], ProcessingError] at 5:8
    MethodParamAnnotation: $complex_input :: Map[Str, ArrayRef[Data|Error]] at 5:62
    MethodParamAnnotation: $handler :: Optional[Handler[Request|Response]] at 5:109
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
    method_decl
      type_expr
      block_stmt
        token
        return_stmt
          literal
        token
}
```

## After Type Inference

```
AST {
  Path: /tmp/complex-method-signatures.pl
  Source length: 311 characters
  Type Annotations:
    MethodReturnAnnotation: transform :: HashRef[ArrayRef[Int]|Str] at 1:8
    MethodParamAnnotation: $input :: ArrayRef[HashRef[Int|Str]] at 1:45
    MethodParamAnnotation: $validator :: CodeRef[Str, Bool] at 1:80
    MethodReturnAnnotation: process :: Result[Array[ProcessedData], ProcessingError] at 5:8
    MethodParamAnnotation: $complex_input :: Map[Str, ArrayRef[Data|Error]] at 5:62
    MethodParamAnnotation: $handler :: Optional[Handler[Request|Response]] at 5:109
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
    method_decl
      type_expr
      block_stmt
        token
        return_stmt
          literal
        token
}
```

## JSON AST

```json
{
  "path": "/tmp/complex-method-signatures.pl",
  "root": {
    "type": "source_file",
    "start": { "Line": 1, "Column": 1, "Offset": 0 },
    "end": { "Line": 7, "Column": 2, "Offset": 311 },
    "children": [
      {
        "type": "method_decl",
        "start": { "Line": 1, "Column": 1, "Offset": 0 },
        "end": { "Line": 3, "Column": 2, "Offset": 153 },
        "children": [
          {
            "type": "block_stmt",
            "start": { "Line": 1, "Column": 121, "Offset": 120 },
            "end": { "Line": 3, "Column": 2, "Offset": 153 },
            "children": [
              { "type": "token", "start": { "Line": 1, "Column": 121, "Offset": 120 }, "end": { "Line": 1, "Column": 122, "Offset": 121 }, "text": "{" },
              {
                "type": "expression_stmt",
                "start": { "Line": 2, "Column": 5, "Offset": 126 },
                "end": { "Line": 2, "Column": 17, "Offset": 138 },
                "children": [
                  { "type": "literal", "start": { "Line": 2, "Column": 5, "Offset": 126 }, "end": { "Line": 2, "Column": 17, "Offset": 138 }, "value": "return {};", "kind": "string" }
                ]
              },
              { "type": "token", "start": { "Line": 3, "Column": 1, "Offset": 152 }, "end": { "Line": 3, "Column": 2, "Offset": 153 }, "text": "}" }
            ]
          }
        ]
      },
      {
        "type": "method_decl",
        "start": { "Line": 5, "Column": 1, "Offset": 155 },
        "end": { "Line": 7, "Column": 2, "Offset": 311 },
        "children": [
          {
            "type": "block_stmt",
            "start": { "Line": 5, "Column": 159, "Offset": 313 },
            "end": { "Line": 7, "Column": 2, "Offset": 311 },
            "children": [
              { "type": "token", "start": { "Line": 5, "Column": 159, "Offset": 313 }, "end": { "Line": 5, "Column": 160, "Offset": 314 }, "text": "{" },
              {
                "type": "expression_stmt",
                "start": { "Line": 6, "Column": 5, "Offset": 319 },
                "end": { "Line": 6, "Column": 24, "Offset": 338 },
                "children": [
                  { "type": "literal", "start": { "Line": 6, "Column": 5, "Offset": 319 }, "end": { "Line": 6, "Column": 24, "Offset": 338 }, "value": "return success([]);", "kind": "string" }
                ]
              },
              { "type": "token", "start": { "Line": 7, "Column": 1, "Offset": 310 }, "end": { "Line": 7, "Column": 2, "Offset": 311 }, "text": "}" }
            ]
          }
        ]
      }
    ]
  }
}
```

# Expected Type Errors

(none)
