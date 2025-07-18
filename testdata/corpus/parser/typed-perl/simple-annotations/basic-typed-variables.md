---
category: typed-perl
subcategory: simple-annotations
tags:
    - typed-variables
    - built-in-types
    - variable-declarations
type_check: true
---

# Basic Typed Variables

Basic typed variable declarations with built-in types

```perl
my Int $count = 42;
my Str $name = "example";
my Bool $flag = 1;
my Num $pi = 3.14159;
```

# Expected AST

## Before Type Inference

### Text Format

```
AST {
  Path:
  Source length: 86 characters
  Type Annotations:
    VarAnnotation: $count :: Int at 1:1
    VarAnnotation: $name :: Str at 2:1
    VarAnnotation: $flag :: Bool at 3:1
    VarAnnotation: $pi :: Num at 4:1
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
  "path": "/tmp/basic_typed_vars.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 4,
      "Column": 22,
      "Offset": 86
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
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 46
        },
        "end": {
          "Line": 3,
          "Column": 18,
          "Offset": 63
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 46
            },
            "end": {
              "Line": 3,
              "Column": 18,
              "Offset": 63
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 46
                },
                "end": {
                  "Line": 3,
                  "Column": 18,
                  "Offset": 63
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
          "Column": 18,
          "Offset": 63
        },
        "end": {
          "Line": 3,
          "Column": 19,
          "Offset": 64
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 4,
          "Column": 1,
          "Offset": 65
        },
        "end": {
          "Line": 4,
          "Column": 21,
          "Offset": 85
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 4,
              "Column": 1,
              "Offset": 65
            },
            "end": {
              "Line": 4,
              "Column": 21,
              "Offset": 85
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 4,
                  "Column": 1,
                  "Offset": 65
                },
                "end": {
                  "Line": 4,
                  "Column": 21,
                  "Offset": 85
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
          "Column": 21,
          "Offset": 85
        },
        "end": {
          "Line": 4,
          "Column": 22,
          "Offset": 86
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
    },
    {
      "annotated_item": "$flag",
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
      "annotated_item": "$pi",
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
  "source_length": 86
}
```

## After Type Inference

### Text Format

```
AST {
  Path:
  Source length: 86 characters
  Type Annotations:
    VarAnnotation: $count :: Int at 1:1
    VarAnnotation: $name :: Str at 2:1
    VarAnnotation: $flag :: Bool at 3:1
    VarAnnotation: $pi :: Num at 4:1
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
  "path": "/tmp/basic_typed_vars.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 4,
      "Column": 22,
      "Offset": 86
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
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 46
        },
        "end": {
          "Line": 3,
          "Column": 18,
          "Offset": 63
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 46
            },
            "end": {
              "Line": 3,
              "Column": 18,
              "Offset": 63
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 46
                },
                "end": {
                  "Line": 3,
                  "Column": 18,
                  "Offset": 63
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
          "Column": 18,
          "Offset": 63
        },
        "end": {
          "Line": 3,
          "Column": 19,
          "Offset": 64
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 4,
          "Column": 1,
          "Offset": 65
        },
        "end": {
          "Line": 4,
          "Column": 21,
          "Offset": 85
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 4,
              "Column": 1,
              "Offset": 65
            },
            "end": {
              "Line": 4,
              "Column": 21,
              "Offset": 85
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 4,
                  "Column": 1,
                  "Offset": 65
                },
                "end": {
                  "Line": 4,
                  "Column": 21,
                  "Offset": 85
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
          "Column": 21,
          "Offset": 85
        },
        "end": {
          "Line": 4,
          "Column": 22,
          "Offset": 86
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
    },
    {
      "annotated_item": "$flag",
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
      "annotated_item": "$pi",
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
  "source_length": 86
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $count = 42;
my $name = "example";
my $flag = 1;
my $pi = 3.14159;
```

## Typed Perl Output

```perl
my Int $count = 42;
my Str $name = "example";
my Bool $flag = 1;
my Num $pi = 3.14159;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
