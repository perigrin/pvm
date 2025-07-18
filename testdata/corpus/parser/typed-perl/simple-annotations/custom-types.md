---
category: typed-perl
subcategory: simple-annotations
tags:
    - custom-types
    - package-qualified
    - typed-variables
type_check: true
---

# Custom Types

Variable declarations with custom and package-qualified types

```perl
my MyType $custom;
my Package::CustomType $qualified;
my UserClass $user = UserClass->new();
```

# Expected AST

## Before Type Inference

### Text Format

```
AST {
  Path: /tmp/custom_types.pl
  Source length: 93 characters
  Type Annotations:
    VarAnnotation: $custom :: MyType at 1:1
    VarAnnotation: $qualified :: Package::CustomType at 2:1
    VarAnnotation: $user :: UserClass at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      variable_declaration
        token
        type_expression
          expression_stmt
            literal
        scalar
          token
          token
    token
    expression_statement
      variable_declaration
        token
        type_expression
          expression_stmt
            literal
        scalar
          token
          token
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
  "path": "/tmp/custom_types.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 4,
      "Column": 1,
      "Offset": 93
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
          "Column": 18,
          "Offset": 17
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 1,
              "Column": 1,
              "Offset": 0
            },
            "end": {
              "Line": 1,
              "Column": 18,
              "Offset": 17
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 1,
                  "Offset": 0
                },
                "end": {
                  "Line": 1,
                  "Column": 3,
                  "Offset": 2
                },
                "text": "my"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 1,
                  "Column": 4,
                  "Offset": 3
                },
                "end": {
                  "Line": 1,
                  "Column": 10,
                  "Offset": 9
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 1,
                      "Column": 4,
                      "Offset": 3
                    },
                    "end": {
                      "Line": 1,
                      "Column": 10,
                      "Offset": 9
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 1,
                          "Column": 4,
                          "Offset": 3
                        },
                        "end": {
                          "Line": 1,
                          "Column": 10,
                          "Offset": 9
                        },
                        "value": "MyType",
                        "kind": "string"
                      }
                    ]
                  }
                ]
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 1,
                  "Column": 11,
                  "Offset": 10
                },
                "end": {
                  "Line": 1,
                  "Column": 18,
                  "Offset": 17
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 11,
                      "Offset": 10
                    },
                    "end": {
                      "Line": 1,
                      "Column": 12,
                      "Offset": 11
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 12,
                      "Offset": 11
                    },
                    "end": {
                      "Line": 1,
                      "Column": 18,
                      "Offset": 17
                    },
                    "text": "custom"
                  }
                ]
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 1,
          "Column": 18,
          "Offset": 17
        },
        "end": {
          "Line": 1,
          "Column": 19,
          "Offset": 18
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 19
        },
        "end": {
          "Line": 2,
          "Column": 34,
          "Offset": 52
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 19
            },
            "end": {
              "Line": 2,
              "Column": 34,
              "Offset": 52
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 19
                },
                "end": {
                  "Line": 2,
                  "Column": 3,
                  "Offset": 21
                },
                "text": "my"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 2,
                  "Column": 4,
                  "Offset": 22
                },
                "end": {
                  "Line": 2,
                  "Column": 23,
                  "Offset": 41
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 2,
                      "Column": 4,
                      "Offset": 22
                    },
                    "end": {
                      "Line": 2,
                      "Column": 23,
                      "Offset": 41
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 2,
                          "Column": 4,
                          "Offset": 22
                        },
                        "end": {
                          "Line": 2,
                          "Column": 23,
                          "Offset": 41
                        },
                        "value": "Package::CustomType",
                        "kind": "string"
                      }
                    ]
                  }
                ]
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 2,
                  "Column": 24,
                  "Offset": 42
                },
                "end": {
                  "Line": 2,
                  "Column": 34,
                  "Offset": 52
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 24,
                      "Offset": 42
                    },
                    "end": {
                      "Line": 2,
                      "Column": 25,
                      "Offset": 43
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 25,
                      "Offset": 43
                    },
                    "end": {
                      "Line": 2,
                      "Column": 34,
                      "Offset": 52
                    },
                    "text": "qualified"
                  }
                ]
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 2,
          "Column": 34,
          "Offset": 52
        },
        "end": {
          "Line": 2,
          "Column": 35,
          "Offset": 53
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 54
        },
        "end": {
          "Line": 3,
          "Column": 38,
          "Offset": 91
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 54
            },
            "end": {
              "Line": 3,
              "Column": 38,
              "Offset": 91
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 54
                },
                "end": {
                  "Line": 3,
                  "Column": 38,
                  "Offset": 91
                },
                "name": "UserClass",
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
          "Column": 38,
          "Offset": 91
        },
        "end": {
          "Line": 3,
          "Column": 39,
          "Offset": 92
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "$custom",
      "type_expression": {
        "Kind": 0,
        "Name": "MyType",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "MyType"
      },
      "position": {
        "Line": 1,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$qualified",
      "type_expression": {
        "Kind": 0,
        "Name": "Package::CustomType",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Package::CustomType"
      },
      "position": {
        "Line": 2,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$user",
      "type_expression": {
        "Kind": 0,
        "Name": "UserClass",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "UserClass"
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
  "source_length": 93
}
```

## After Type Inference

### Text Format

```
AST {
  Path: /tmp/custom_types.pl
  Source length: 93 characters
  Type Annotations:
    VarAnnotation: $custom :: MyType at 1:1
    VarAnnotation: $qualified :: Package::CustomType at 2:1
    VarAnnotation: $user :: UserClass at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      variable_declaration
        token
        type_expression
          expression_stmt
            literal
        scalar
          token
          token
    token
    expression_statement
      variable_declaration
        token
        type_expression
          expression_stmt
            literal
        scalar
          token
          token
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
  "path": "/tmp/custom_types.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 4,
      "Column": 1,
      "Offset": 93
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
          "Column": 18,
          "Offset": 17
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 1,
              "Column": 1,
              "Offset": 0
            },
            "end": {
              "Line": 1,
              "Column": 18,
              "Offset": 17
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 1,
                  "Offset": 0
                },
                "end": {
                  "Line": 1,
                  "Column": 3,
                  "Offset": 2
                },
                "text": "my"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 1,
                  "Column": 4,
                  "Offset": 3
                },
                "end": {
                  "Line": 1,
                  "Column": 10,
                  "Offset": 9
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 1,
                      "Column": 4,
                      "Offset": 3
                    },
                    "end": {
                      "Line": 1,
                      "Column": 10,
                      "Offset": 9
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 1,
                          "Column": 4,
                          "Offset": 3
                        },
                        "end": {
                          "Line": 1,
                          "Column": 10,
                          "Offset": 9
                        },
                        "value": "MyType",
                        "kind": "string"
                      }
                    ]
                  }
                ]
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 1,
                  "Column": 11,
                  "Offset": 10
                },
                "end": {
                  "Line": 1,
                  "Column": 18,
                  "Offset": 17
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 11,
                      "Offset": 10
                    },
                    "end": {
                      "Line": 1,
                      "Column": 12,
                      "Offset": 11
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 12,
                      "Offset": 11
                    },
                    "end": {
                      "Line": 1,
                      "Column": 18,
                      "Offset": 17
                    },
                    "text": "custom"
                  }
                ]
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 1,
          "Column": 18,
          "Offset": 17
        },
        "end": {
          "Line": 1,
          "Column": 19,
          "Offset": 18
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 19
        },
        "end": {
          "Line": 2,
          "Column": 34,
          "Offset": 52
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 19
            },
            "end": {
              "Line": 2,
              "Column": 34,
              "Offset": 52
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 19
                },
                "end": {
                  "Line": 2,
                  "Column": 3,
                  "Offset": 21
                },
                "text": "my"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 2,
                  "Column": 4,
                  "Offset": 22
                },
                "end": {
                  "Line": 2,
                  "Column": 23,
                  "Offset": 41
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 2,
                      "Column": 4,
                      "Offset": 22
                    },
                    "end": {
                      "Line": 2,
                      "Column": 23,
                      "Offset": 41
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 2,
                          "Column": 4,
                          "Offset": 22
                        },
                        "end": {
                          "Line": 2,
                          "Column": 23,
                          "Offset": 41
                        },
                        "value": "Package::CustomType",
                        "kind": "string"
                      }
                    ]
                  }
                ]
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 2,
                  "Column": 24,
                  "Offset": 42
                },
                "end": {
                  "Line": 2,
                  "Column": 34,
                  "Offset": 52
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 24,
                      "Offset": 42
                    },
                    "end": {
                      "Line": 2,
                      "Column": 25,
                      "Offset": 43
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 25,
                      "Offset": 43
                    },
                    "end": {
                      "Line": 2,
                      "Column": 34,
                      "Offset": 52
                    },
                    "text": "qualified"
                  }
                ]
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 2,
          "Column": 34,
          "Offset": 52
        },
        "end": {
          "Line": 2,
          "Column": 35,
          "Offset": 53
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 54
        },
        "end": {
          "Line": 3,
          "Column": 38,
          "Offset": 91
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 54
            },
            "end": {
              "Line": 3,
              "Column": 38,
              "Offset": 91
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 54
                },
                "end": {
                  "Line": 3,
                  "Column": 38,
                  "Offset": 91
                },
                "name": "UserClass",
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
          "Column": 38,
          "Offset": 91
        },
        "end": {
          "Line": 3,
          "Column": 39,
          "Offset": 92
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "$custom",
      "type_expression": {
        "Kind": 0,
        "Name": "MyType",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "MyType"
      },
      "position": {
        "Line": 1,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$qualified",
      "type_expression": {
        "Kind": 0,
        "Name": "Package::CustomType",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Package::CustomType"
      },
      "position": {
        "Line": 2,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$user",
      "type_expression": {
        "Kind": 0,
        "Name": "UserClass",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "UserClass"
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
  "source_length": 93
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $custom;
my $qualified;
my $user = UserClass->new();
```

## Typed Perl Output

```perl
my MyType $custom;
my Package::CustomType $qualified;
my UserClass $user = UserClass->new();
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
