---
category: typed-perl
subcategory: simple-annotations
tags:
    - arrays
    - hashes
    - parameterized-types
    - typed-variables
type_check: true
---

# Typed Arrays Hashes

Typed array and hash declarations with basic parameterized types

```perl
my ArrayRef[Int] @numbers = (1, 2, 3);
my HashRef[Str] %config = (key => 'value');
my ArrayRef[Str] @strings = ("a", "b", "c");
```

# Expected AST

## Before Type Inference

### Text Format

```
AST {
  Path:
  Source length: 127 characters
  Type Annotations:
    VarAnnotation: @numbers :: ArrayRef[Int] at 1:1
    VarAnnotation: %config :: HashRef[Str] at 2:1
    VarAnnotation: @strings :: ArrayRef[Str] at 3:1
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
    expression_statement
      var_decl
        variable
    token
}
```

### JSON Format

```json
{
  "path": "/tmp/typed-arrays-hashes.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 3,
      "Column": 45,
      "Offset": 127
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
          "Column": 38,
          "Offset": 37
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
              "Column": 38,
              "Offset": 37
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
                  "Column": 38,
                  "Offset": 37
                },
                "name": "ArrayRef[Int]",
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
          "Column": 38,
          "Offset": 37
        },
        "end": {
          "Line": 1,
          "Column": 39,
          "Offset": 38
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 39
        },
        "end": {
          "Line": 2,
          "Column": 43,
          "Offset": 81
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 39
            },
            "end": {
              "Line": 2,
              "Column": 43,
              "Offset": 81
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 39
                },
                "end": {
                  "Line": 2,
                  "Column": 43,
                  "Offset": 81
                },
                "name": "HashRef[Str]",
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
          "Column": 43,
          "Offset": 81
        },
        "end": {
          "Line": 2,
          "Column": 44,
          "Offset": 82
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 83
        },
        "end": {
          "Line": 3,
          "Column": 44,
          "Offset": 126
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 83
            },
            "end": {
              "Line": 3,
              "Column": 44,
              "Offset": 126
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 83
                },
                "end": {
                  "Line": 3,
                  "Column": 44,
                  "Offset": 126
                },
                "name": "ArrayRef[Str]",
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
          "Line": 3,
          "Column": 44,
          "Offset": 126
        },
        "end": {
          "Line": 3,
          "Column": 45,
          "Offset": 127
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "@numbers",
      "type_expression": {
        "Kind": 4,
        "Name": "ArrayRef[Int]",
        "Parameters": [
          {
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
          }
        ],
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "ArrayRef[Int]"
      },
      "position": {
        "Line": 1,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "%config",
      "type_expression": {
        "Kind": 4,
        "Name": "HashRef[Str]",
        "Parameters": [
          {
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
          }
        ],
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "HashRef[Str]"
      },
      "position": {
        "Line": 2,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "@strings",
      "type_expression": {
        "Kind": 4,
        "Name": "ArrayRef[Str]",
        "Parameters": [
          {
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
          }
        ],
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "ArrayRef[Str]"
      },
      "position": {
        "Line": 3,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "errors": [],
  "source_length": 127
}
```

## After Type Inference

### Text Format

```
AST {
  Path:
  Source length: 127 characters
  Type Annotations:
    VarAnnotation: @numbers :: ArrayRef[Int] at 1:1
    VarAnnotation: %config :: HashRef[Str] at 2:1
    VarAnnotation: @strings :: ArrayRef[Str] at 3:1
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
    expression_statement
      var_decl
        variable
    token
}
```

### JSON Format

```json
{
  "path": "/tmp/typed-arrays-hashes.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 3,
      "Column": 45,
      "Offset": 127
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
          "Column": 38,
          "Offset": 37
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
              "Column": 38,
              "Offset": 37
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
                  "Column": 38,
                  "Offset": 37
                },
                "name": "ArrayRef[Int]",
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
          "Column": 38,
          "Offset": 37
        },
        "end": {
          "Line": 1,
          "Column": 39,
          "Offset": 38
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 39
        },
        "end": {
          "Line": 2,
          "Column": 43,
          "Offset": 81
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 39
            },
            "end": {
              "Line": 2,
              "Column": 43,
              "Offset": 81
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 39
                },
                "end": {
                  "Line": 2,
                  "Column": 43,
                  "Offset": 81
                },
                "name": "HashRef[Str]",
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
          "Column": 43,
          "Offset": 81
        },
        "end": {
          "Line": 2,
          "Column": 44,
          "Offset": 82
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 83
        },
        "end": {
          "Line": 3,
          "Column": 44,
          "Offset": 126
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 83
            },
            "end": {
              "Line": 3,
              "Column": 44,
              "Offset": 126
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 83
                },
                "end": {
                  "Line": 3,
                  "Column": 44,
                  "Offset": 126
                },
                "name": "ArrayRef[Str]",
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
          "Line": 3,
          "Column": 44,
          "Offset": 126
        },
        "end": {
          "Line": 3,
          "Column": 45,
          "Offset": 127
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "@numbers",
      "type_expression": {
        "Kind": 4,
        "Name": "ArrayRef[Int]",
        "Parameters": [
          {
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
          }
        ],
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "ArrayRef[Int]"
      },
      "position": {
        "Line": 1,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "%config",
      "type_expression": {
        "Kind": 4,
        "Name": "HashRef[Str]",
        "Parameters": [
          {
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
          }
        ],
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "HashRef[Str]"
      },
      "position": {
        "Line": 2,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "@strings",
      "type_expression": {
        "Kind": 4,
        "Name": "ArrayRef[Str]",
        "Parameters": [
          {
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
          }
        ],
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "ArrayRef[Str]"
      },
      "position": {
        "Line": 3,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "errors": [],
  "source_length": 127
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my @numbers = (1, 2, 3);
my %config = (key => 'value');
my @strings = ("a", "b", "c");
```

## Typed Perl Output

```perl
my ArrayRef[Int] @numbers = (1, 2, 3);
my HashRef[Str] %config = (key => 'value');
my ArrayRef[Str] @strings = ("a", "b", "c");
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
