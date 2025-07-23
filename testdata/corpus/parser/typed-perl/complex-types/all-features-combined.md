---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - all-features
    - complex-combinations
    - intersection-types
    - negation-types
    - parameterized-types
    - union-types
    - type-assertions
---

# All Features Combined

Complex combination of all type features: unions, intersections, negations, parameterized types, and assertions

```perl
method Result[Map[Str, Array[ProcessedItem|FailureReason]], ProcessingError&Detailed] complex_processing(ArrayRef[HashRef[Int|Str]&!Undef] $validated_data, (Processor[Request]|Handler[Response])&Configured $handler, Optional[Logger[Info|Error]] $logger) {
    my $transformed = $validated_data as ArrayRef[Data&Processed];
    return success($transformed->map(sub { process($_) }));
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
method complex_processing($validated_data, $handler, $logger) {
    my $transformed = $validated_data;
    return success($transformed->map(sub { process($_) }));
}
```

## Typed Perl Output

```perl
method Result[Map[Str, Array[ProcessedItem|FailureReason]], ProcessingError&Detailed] complex_processing(ArrayRef[HashRef[Int|Str]&!Undef] $validated_data, (Processor[Request]|Handler[Response])&Configured $handler, Optional[Logger[Info|Error]] $logger) {
    my $transformed = $validated_data as ArrayRef[Data&Processed];
    return success($transformed->map(sub { process($_) }));
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
AST {
  Path: /tmp/all-features-combined.pl
  Source length: 384 characters
  Type Annotations:
    MethodReturnAnnotation: complex_processing :: Result[Map[Str, Array[ProcessedItem|FailureReason]], ProcessingError&Detailed] at 1:8
    MethodParamAnnotation: $validated_data :: ArrayRef[HashRef[Int|Str]&!Undef] at 1:106
    MethodParamAnnotation: $handler :: Configured at 1:196
    MethodParamAnnotation: $logger :: Optional[Logger[Info|Error]] at 1:217
    VarAnnotation: $validated_data :: ArrayRef[Data&Processed] at 2:23
  Root: source_file
  Tree Structure:
  source_file
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
}
```

## JSON AST

```json
{
  "path": "/tmp/all-features-combined.pl",
  "root": {
    "type": "source_file",
    "start": { "Line": 1, "Column": 1, "Offset": 0 },
    "end": { "Line": 4, "Column": 2, "Offset": 384 },
    "children": [
      {
        "type": "method_decl",
        "start": { "Line": 1, "Column": 1, "Offset": 0 },
        "end": { "Line": 4, "Column": 2, "Offset": 384 },
        "children": [
          {
            "type": "block_stmt",
            "start": { "Line": 1, "Column": 264, "Offset": 263 },
            "end": { "Line": 4, "Column": 2, "Offset": 384 },
            "children": [
              { "type": "token", "text": "{" },
              {
                "type": "expression_stmt",
                "start": { "Line": 2, "Column": 5, "Offset": 268 },
                "end": { "Line": 2, "Column": 64, "Offset": 327 },
                "children": [
                  {
                    "type": "literal",
                    "value": "my $transformed = $validated_data as ArrayRef[Data&Processed];",
                    "kind": "string",
                    "note": "Contains type assertion with intersection type"
                  }
                ]
              },
              {
                "type": "expression_stmt",
                "start": { "Line": 3, "Column": 5, "Offset": 332 },
                "end": { "Line": 3, "Column": 49, "Offset": 376 },
                "children": [
                  {
                    "type": "literal",
                    "value": "return success($transformed->map(sub { process($_) }));",
                    "kind": "string"
                  }
                ]
              },
              { "type": "token", "text": "}" }
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
