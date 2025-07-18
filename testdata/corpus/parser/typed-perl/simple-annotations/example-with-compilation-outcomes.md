---
category: typed-perl
subcategory: simple-annotations
tags:
    - variables
    - compilation-outcomes
type_check: true
---

# Example With Compilation Outcomes

Basic variable declaration with type annotations and expected compilation outcomes.

```perl
my Int $count = 42;
my Str $name = "example";
```

# Expected AST

## Before Type Inference

### Text Format

```
AST {
  Path: /tmp/example_with_compilation_outcomes.pl
  Source length: 46 characters
  Type Annotations:
    VarAnnotation: $count :: Int at 1:1
    VarAnnotation: $name :: Str at 2:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      var_decl
        variable
    token
    expression_statement
      var_decl
        variable
    token
}
```

### JSON Format

```json
{
  "path": "/tmp/example_with_compilation_outcomes.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 3,
      "Column": 1,
      "Offset": 46
    },
    "children": [
      {
        "type": "expression_statement",
        "start": {
          "Line": 1,
          "Column": 1,
          "Offset": 0
        },
        "end": {
          "Line": 1,
          "Column": 19,
          "Offset": 18
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 1,
              "Column": 1,
              "Offset": 0
            },
            "end": {
              "Line": 1,
              "Column": 19,
              "Offset": 18
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 1,
                  "Column": 1,
                  "Offset": 0
                },
                "end": {
                  "Line": 1,
                  "Column": 19,
                  "Offset": 18
                },
                "name": "Int",
                "sigil": "$"
              }
            ],
            "decl_type": "my"
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 1,
          "Column": 19,
          "Offset": 18
        },
        "end": {
          "Line": 1,
          "Column": 20,
          "Offset": 19
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 20
        },
        "end": {
          "Line": 2,
          "Column": 25,
          "Offset": 44
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 20
            },
            "end": {
              "Line": 2,
              "Column": 25,
              "Offset": 44
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 20
                },
                "end": {
                  "Line": 2,
                  "Column": 25,
                  "Offset": 44
                },
                "name": "Str",
                "sigil": "$"
              }
            ],
            "decl_type": "my"
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 2,
          "Column": 25,
          "Offset": 44
        },
        "end": {
          "Line": 2,
          "Column": 26,
          "Offset": 45
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "$count",
      "type_expression": {
        "Kind": 0,
        "Name": "Int",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Int"
      },
      "position": {
        "Line": 1,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$name",
      "type_expression": {
        "Kind": 0,
        "Name": "Str",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Str"
      },
      "position": {
        "Line": 2,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "errors": [],
  "source_length": 46
}
```

## After Type Inference

### Text Format

```
AST {
  Path: /tmp/example_with_compilation_outcomes.pl
  Source length: 46 characters
  Type Annotations:
    VarAnnotation: $count :: Int at 1:1
    VarAnnotation: $name :: Str at 2:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      var_decl
        variable
    token
    expression_statement
      var_decl
        variable
    token
}
```

### JSON Format

```json
{
  "path": "/tmp/example_with_compilation_outcomes.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 3,
      "Column": 1,
      "Offset": 46
    },
    "children": [
      {
        "type": "expression_statement",
        "start": {
          "Line": 1,
          "Column": 1,
          "Offset": 0
        },
        "end": {
          "Line": 1,
          "Column": 19,
          "Offset": 18
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 1,
              "Column": 1,
              "Offset": 0
            },
            "end": {
              "Line": 1,
              "Column": 19,
              "Offset": 18
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 1,
                  "Column": 1,
                  "Offset": 0
                },
                "end": {
                  "Line": 1,
                  "Column": 19,
                  "Offset": 18
                },
                "name": "Int",
                "sigil": "$"
              }
            ],
            "decl_type": "my"
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 1,
          "Column": 19,
          "Offset": 18
        },
        "end": {
          "Line": 1,
          "Column": 20,
          "Offset": 19
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 20
        },
        "end": {
          "Line": 2,
          "Column": 25,
          "Offset": 44
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 20
            },
            "end": {
              "Line": 2,
              "Column": 25,
              "Offset": 44
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 20
                },
                "end": {
                  "Line": 2,
                  "Column": 25,
                  "Offset": 44
                },
                "name": "Str",
                "sigil": "$"
              }
            ],
            "decl_type": "my"
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 2,
          "Column": 25,
          "Offset": 44
        },
        "end": {
          "Line": 2,
          "Column": 26,
          "Offset": 45
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "$count",
      "type_expression": {
        "Kind": 0,
        "Name": "Int",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Int"
      },
      "position": {
        "Line": 1,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$name",
      "type_expression": {
        "Kind": 0,
        "Name": "Str",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Str"
      },
      "position": {
        "Line": 2,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "errors": [],
  "source_length": 46
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $count = 42;
my $name = "example";
```

## Typed Perl Output

```perl
use v5.36;
my Int $count = 42;
my Str $name = "example";
```

## Inferred Perl Output

```perl
use v5.36;
my Int $count = 42;
my Str $name = "example";
```

# Expected Type Errors

```
(none)
```
