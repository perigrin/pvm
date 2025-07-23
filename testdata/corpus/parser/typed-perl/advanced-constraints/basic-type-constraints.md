---
category: typed-perl
subcategory: advanced-constraints
type_check: true
tags:
    - type-constraints
    - where-clauses
---

# Basic Type Constraints

Test basic type constraint parsing with where clauses

```perl
class Container<T> where T: Serializable { }
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
class Container<T> where T:  { }
```

## Typed Perl Output

```perl
class Container<T> where T: Serializable { }
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Before Type Inference

### Text Format
```
AST {
  Path: /tmp/basic-type-constraints.pl
  Source length: 44 characters
  Type Annotations:
    VarAnnotation: Container :: class at 1:1
  Root: source_file
  Tree Structure:
  source_file
    class_decl
}
```

### JSON Format
```json
{
  "path": "/tmp/basic-type-constraints.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 1,
      "Column": 45,
      "Offset": 44
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
          "Column": 45,
          "Offset": 44
        },
        "name": "Container"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "Container",
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
  "source_length": 44
}
```

## After Type Inference

### Text Format
```
# Type inference not yet fully implemented
```

### JSON Format
```json
# Type inference not yet fully implemented
```

# Expected Type Errors

(none)
