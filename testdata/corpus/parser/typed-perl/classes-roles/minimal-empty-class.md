---
category: typed-perl
subcategory: classes-roles
tags:
    - class-declaration
    - minimal
    - empty-class
type_check: true
---

# Minimal Empty Class

Simplest possible class declaration

```perl
class Foo {}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 12 characters
  Type Annotations:
    VarAnnotation: Foo :: class at 1:1
  Root: source_file
  Tree Structure:
  source_file
    class_decl
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 12 characters
  Type Annotations:
    VarAnnotation: Foo :: class at 1:1
  Root: source_file
  Tree Structure:
  source_file
    class_decl
}
```

## JSON AST

```json
{
  "path": "/tmp/minimal-class.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 2,
      "Column": 1,
      "Offset": 13
    },
    "children": [
      {
        "type": "class_decl",
        "start": {
          "Line": 1,
          "Column": 1,
          "Offset": 0
        },
        "end": {
          "Line": 1,
          "Column": 13,
          "Offset": 12
        },
        "name": "Foo"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "Foo",
      "type_expression": {
        "Kind": 0,
        "Name": "class",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "class"
      },
      "position": {
        "Line": 1,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "errors": [],
  "source_length": 13
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
class Foo {}
```

## Typed Perl Output

```perl
class Foo {}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
