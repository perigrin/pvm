---
category: typed-perl
subcategory: simple-annotations
tags:
    - complex-assignments
    - expressions
    - typed-variables
type_check: true
---

# Complex Assignments

Type annotations with complex assignment expressions

```perl
my Int $calculated = $base + $increment;
my Str $interpolated = "Value: $count";
my Bool $comparison = $a > $b;
my Num $result = $x * $y + $z;
```

# Expected AST

## Before Type Inference

### Text Format

```
AST {
  Path: /tmp/complex_assignments.pl
  Source length: 143 characters
  Type Annotations:
    VarAnnotation: $calculated :: Int at 1:1
    VarAnnotation: $interpolated :: Str at 2:1
    VarAnnotation: $comparison :: Bool at 3:1
    VarAnnotation: $result :: Num at 4:1
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
    expression_statement
      var_decl
        variable
    token
}
```

### JSON Format

```json
{
  "path": "/tmp/complex_assignments.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 5,
      "Column": 1,
      "Offset": 143
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
          "Column": 40,
          "Offset": 39
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
              "Column": 40,
              "Offset": 39
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
                  "Column": 40,
                  "Offset": 39
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
          "Column": 40,
          "Offset": 39
        },
        "end": {
          "Line": 1,
          "Column": 41,
          "Offset": 40
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 41
        },
        "end": {
          "Line": 2,
          "Column": 39,
          "Offset": 79
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 41
            },
            "end": {
              "Line": 2,
              "Column": 39,
              "Offset": 79
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 41
                },
                "end": {
                  "Line": 2,
                  "Column": 39,
                  "Offset": 79
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
          "Column": 39,
          "Offset": 79
        },
        "end": {
          "Line": 2,
          "Column": 40,
          "Offset": 80
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 81
        },
        "end": {
          "Line": 3,
          "Column": 30,
          "Offset": 110
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 81
            },
            "end": {
              "Line": 3,
              "Column": 30,
              "Offset": 110
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 81
                },
                "end": {
                  "Line": 3,
                  "Column": 30,
                  "Offset": 110
                },
                "name": "Bool",
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
          "Column": 30,
          "Offset": 110
        },
        "end": {
          "Line": 3,
          "Column": 31,
          "Offset": 111
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 4,
          "Column": 1,
          "Offset": 112
        },
        "end": {
          "Line": 4,
          "Column": 30,
          "Offset": 141
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 4,
              "Column": 1,
              "Offset": 112
            },
            "end": {
              "Line": 4,
              "Column": 30,
              "Offset": 141
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 4,
                  "Column": 1,
                  "Offset": 112
                },
                "end": {
                  "Line": 4,
                  "Column": 30,
                  "Offset": 141
                },
                "name": "Num",
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
          "Line": 4,
          "Column": 30,
          "Offset": 141
        },
        "end": {
          "Line": 4,
          "Column": 31,
          "Offset": 142
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "$calculated",
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
      "annotated_item": "$interpolated",
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
    },
    {
      "annotated_item": "$comparison",
      "type_expression": {
        "Kind": 0,
        "Name": "Bool",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Bool"
      },
      "position": {
        "Line": 3,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$result",
      "type_expression": {
        "Kind": 0,
        "Name": "Num",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Num"
      },
      "position": {
        "Line": 4,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "errors": [],
  "source_length": 143
}
```

## After Type Inference

### Text Format

```
AST {
  Path: /tmp/complex_assignments.pl
  Source length: 143 characters
  Type Annotations:
    VarAnnotation: $calculated :: Int at 1:1
    VarAnnotation: $interpolated :: Str at 2:1
    VarAnnotation: $comparison :: Bool at 3:1
    VarAnnotation: $result :: Num at 4:1
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
    expression_statement
      var_decl
        variable
    token
}
```

### JSON Format

```json
{
  "path": "/tmp/complex_assignments.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 5,
      "Column": 1,
      "Offset": 143
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
          "Column": 40,
          "Offset": 39
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
              "Column": 40,
              "Offset": 39
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
                  "Column": 40,
                  "Offset": 39
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
          "Column": 40,
          "Offset": 39
        },
        "end": {
          "Line": 1,
          "Column": 41,
          "Offset": 40
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 41
        },
        "end": {
          "Line": 2,
          "Column": 39,
          "Offset": 79
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 41
            },
            "end": {
              "Line": 2,
              "Column": 39,
              "Offset": 79
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 41
                },
                "end": {
                  "Line": 2,
                  "Column": 39,
                  "Offset": 79
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
          "Column": 39,
          "Offset": 79
        },
        "end": {
          "Line": 2,
          "Column": 40,
          "Offset": 80
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 81
        },
        "end": {
          "Line": 3,
          "Column": 30,
          "Offset": 110
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 81
            },
            "end": {
              "Line": 3,
              "Column": 30,
              "Offset": 110
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 81
                },
                "end": {
                  "Line": 3,
                  "Column": 30,
                  "Offset": 110
                },
                "name": "Bool",
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
          "Column": 30,
          "Offset": 110
        },
        "end": {
          "Line": 3,
          "Column": 31,
          "Offset": 111
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 4,
          "Column": 1,
          "Offset": 112
        },
        "end": {
          "Line": 4,
          "Column": 30,
          "Offset": 141
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 4,
              "Column": 1,
              "Offset": 112
            },
            "end": {
              "Line": 4,
              "Column": 30,
              "Offset": 141
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 4,
                  "Column": 1,
                  "Offset": 112
                },
                "end": {
                  "Line": 4,
                  "Column": 30,
                  "Offset": 141
                },
                "name": "Num",
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
          "Line": 4,
          "Column": 30,
          "Offset": 141
        },
        "end": {
          "Line": 4,
          "Column": 31,
          "Offset": 142
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "$calculated",
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
      "annotated_item": "$interpolated",
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
    },
    {
      "annotated_item": "$comparison",
      "type_expression": {
        "Kind": 0,
        "Name": "Bool",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Bool"
      },
      "position": {
        "Line": 3,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$result",
      "type_expression": {
        "Kind": 0,
        "Name": "Num",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Num"
      },
      "position": {
        "Line": 4,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "errors": [],
  "source_length": 143
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $calculated = $base + $increment;
my $interpolated = "Value: $count";
my $comparison = $a > $b;
my $result = $x * $y + $z;
```

## Typed Perl Output

```perl
my Int $calculated = $base + $increment;
my Str $interpolated = "Value: $count";
my Bool $comparison = $a > $b;
my Num $result = $x * $y + $z;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
