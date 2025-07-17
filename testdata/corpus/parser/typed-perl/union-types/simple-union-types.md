---
category: typed-perl
subcategory: union-types
tags:
    - simple-unions
    - union-types
    - variable-declarations
type_check: true
---

# Simple Union Types

Simple union type expressions with two types

```perl
my Int|Str $flexible = 42;
my Bool|Undef $maybe_flag;
my Num|Str $mixed_value = "text";
```

# Expected AST

## Before Type Inference

### Text Format

```
AST {
  Path:
  Source length: 87 characters
  Type Annotations:
    VarAnnotation: $flexible :: Int|Str at 1:1
    VarAnnotation: $maybe_flag :: Bool|Undef at 2:1
    VarAnnotation: $mixed_value :: Num|Str at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      var_decl
        variable
    token
    expression_statement
      variable_declaration
        token
        type_expression
          union_type
            type_expression
              expression_stmt
                literal
            expression_stmt
              literal
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
  "path": "/tmp/union_types_test.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 3,
      "Column": 34,
      "Offset": 87
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
          "Column": 26,
          "Offset": 25
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
              "Column": 26,
              "Offset": 25
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
                  "Column": 26,
                  "Offset": 25
                },
                "name": "Int|Str",
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
          "Column": 26,
          "Offset": 25
        },
        "end": {
          "Line": 1,
          "Column": 27,
          "Offset": 26
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 27
        },
        "end": {
          "Line": 2,
          "Column": 26,
          "Offset": 52
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 27
            },
            "end": {
              "Line": 2,
              "Column": 26,
              "Offset": 52
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 27
                },
                "end": {
                  "Line": 2,
                  "Column": 3,
                  "Offset": 29
                },
                "text": "my"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 2,
                  "Column": 4,
                  "Offset": 30
                },
                "end": {
                  "Line": 2,
                  "Column": 14,
                  "Offset": 40
                },
                "children": [
                  {
                    "type": "union_type",
                    "start": {
                      "Line": 2,
                      "Column": 4,
                      "Offset": 30
                    },
                    "end": {
                      "Line": 2,
                      "Column": 14,
                      "Offset": 40
                    },
                    "children": [
                      {
                        "type": "type_expression",
                        "start": {
                          "Line": 2,
                          "Column": 4,
                          "Offset": 30
                        },
                        "end": {
                          "Line": 2,
                          "Column": 8,
                          "Offset": 34
                        },
                        "children": [
                          {
                            "type": "expression_stmt",
                            "start": {
                              "Line": 2,
                              "Column": 4,
                              "Offset": 30
                            },
                            "end": {
                              "Line": 2,
                              "Column": 8,
                              "Offset": 34
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 2,
                                  "Column": 4,
                                  "Offset": 30
                                },
                                "end": {
                                  "Line": 2,
                                  "Column": 8,
                                  "Offset": 34
                                },
                                "value": "Bool",
                                "kind": "string"
                              }
                            ]
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 2,
                          "Column": 8,
                          "Offset": 34
                        },
                        "end": {
                          "Line": 2,
                          "Column": 9,
                          "Offset": 35
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 2,
                              "Column": 8,
                              "Offset": 34
                            },
                            "end": {
                              "Line": 2,
                              "Column": 9,
                              "Offset": 35
                            },
                            "value": "|",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "type_expression",
                        "start": {
                          "Line": 2,
                          "Column": 9,
                          "Offset": 35
                        },
                        "end": {
                          "Line": 2,
                          "Column": 14,
                          "Offset": 40
                        },
                        "children": [
                          {
                            "type": "expression_stmt",
                            "start": {
                              "Line": 2,
                              "Column": 9,
                              "Offset": 35
                            },
                            "end": {
                              "Line": 2,
                              "Column": 14,
                              "Offset": 40
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 2,
                                  "Column": 9,
                                  "Offset": 35
                                },
                                "end": {
                                  "Line": 2,
                                  "Column": 14,
                                  "Offset": 40
                                },
                                "value": "Undef",
                                "kind": "string"
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  }
                ]
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 2,
                  "Column": 15,
                  "Offset": 41
                },
                "end": {
                  "Line": 2,
                  "Column": 26,
                  "Offset": 52
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 15,
                      "Offset": 41
                    },
                    "end": {
                      "Line": 2,
                      "Column": 16,
                      "Offset": 42
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 16,
                      "Offset": 42
                    },
                    "end": {
                      "Line": 2,
                      "Column": 26,
                      "Offset": 52
                    },
                    "text": "maybe_flag"
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
          "Column": 26,
          "Offset": 52
        },
        "end": {
          "Line": 2,
          "Column": 27,
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
          "Column": 33,
          "Offset": 86
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
              "Column": 33,
              "Offset": 86
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
                  "Column": 33,
                  "Offset": 86
                },
                "name": "Num|Str",
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
          "Column": 33,
          "Offset": 86
        },
        "end": {
          "Line": 3,
          "Column": 34,
          "Offset": 87
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "$flexible",
      "type_expression": {
        "Kind": 0,
        "Name": "Int|Str",
        "Parameters": null,
        "IsUnion": true,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": [
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
          },
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
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Int|Str"
      },
      "position": {
        "Line": 1,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$maybe_flag",
      "type_expression": {
        "Kind": 0,
        "Name": "Bool|Undef",
        "Parameters": null,
        "IsUnion": true,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": [
          {
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
          {
            "Kind": 0,
            "Name": "Undef",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "Undef"
          }
        ],
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Bool|Undef"
      },
      "position": {
        "Line": 2,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$mixed_value",
      "type_expression": {
        "Kind": 0,
        "Name": "Num|Str",
        "Parameters": null,
        "IsUnion": true,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": [
          {
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
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Num|Str"
      },
      "position": {
        "Line": 3,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "source_length": 87
}
```

## After Type Inference

### Text Format

```
AST {
  Path:
  Source length: 87 characters
  Type Annotations:
    VarAnnotation: $flexible :: Int|Str at 1:1
    VarAnnotation: $maybe_flag :: Bool|Undef at 2:1
    VarAnnotation: $mixed_value :: Num|Str at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      var_decl
        variable
    token
    expression_statement
      variable_declaration
        token
        type_expression
          union_type
            type_expression
              expression_stmt
                literal
            expression_stmt
              literal
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
  "path": "/tmp/union_types_test.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 3,
      "Column": 34,
      "Offset": 87
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
          "Column": 26,
          "Offset": 25
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
              "Column": 26,
              "Offset": 25
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
                  "Column": 26,
                  "Offset": 25
                },
                "name": "Int|Str",
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
          "Column": 26,
          "Offset": 25
        },
        "end": {
          "Line": 1,
          "Column": 27,
          "Offset": 26
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 27
        },
        "end": {
          "Line": 2,
          "Column": 26,
          "Offset": 52
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 27
            },
            "end": {
              "Line": 2,
              "Column": 26,
              "Offset": 52
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 27
                },
                "end": {
                  "Line": 2,
                  "Column": 3,
                  "Offset": 29
                },
                "text": "my"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 2,
                  "Column": 4,
                  "Offset": 30
                },
                "end": {
                  "Line": 2,
                  "Column": 14,
                  "Offset": 40
                },
                "children": [
                  {
                    "type": "union_type",
                    "start": {
                      "Line": 2,
                      "Column": 4,
                      "Offset": 30
                    },
                    "end": {
                      "Line": 2,
                      "Column": 14,
                      "Offset": 40
                    },
                    "children": [
                      {
                        "type": "type_expression",
                        "start": {
                          "Line": 2,
                          "Column": 4,
                          "Offset": 30
                        },
                        "end": {
                          "Line": 2,
                          "Column": 8,
                          "Offset": 34
                        },
                        "children": [
                          {
                            "type": "expression_stmt",
                            "start": {
                              "Line": 2,
                              "Column": 4,
                              "Offset": 30
                            },
                            "end": {
                              "Line": 2,
                              "Column": 8,
                              "Offset": 34
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 2,
                                  "Column": 4,
                                  "Offset": 30
                                },
                                "end": {
                                  "Line": 2,
                                  "Column": 8,
                                  "Offset": 34
                                },
                                "value": "Bool",
                                "kind": "string"
                              }
                            ]
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 2,
                          "Column": 8,
                          "Offset": 34
                        },
                        "end": {
                          "Line": 2,
                          "Column": 9,
                          "Offset": 35
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 2,
                              "Column": 8,
                              "Offset": 34
                            },
                            "end": {
                              "Line": 2,
                              "Column": 9,
                              "Offset": 35
                            },
                            "value": "|",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "type_expression",
                        "start": {
                          "Line": 2,
                          "Column": 9,
                          "Offset": 35
                        },
                        "end": {
                          "Line": 2,
                          "Column": 14,
                          "Offset": 40
                        },
                        "children": [
                          {
                            "type": "expression_stmt",
                            "start": {
                              "Line": 2,
                              "Column": 9,
                              "Offset": 35
                            },
                            "end": {
                              "Line": 2,
                              "Column": 14,
                              "Offset": 40
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 2,
                                  "Column": 9,
                                  "Offset": 35
                                },
                                "end": {
                                  "Line": 2,
                                  "Column": 14,
                                  "Offset": 40
                                },
                                "value": "Undef",
                                "kind": "string"
                              }
                            ]
                          }
                        ]
                      }
                    ]
                  }
                ]
              },
              {
                "type": "scalar",
                "start": {
                  "Line": 2,
                  "Column": 15,
                  "Offset": 41
                },
                "end": {
                  "Line": 2,
                  "Column": 26,
                  "Offset": 52
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 15,
                      "Offset": 41
                    },
                    "end": {
                      "Line": 2,
                      "Column": 16,
                      "Offset": 42
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 16,
                      "Offset": 42
                    },
                    "end": {
                      "Line": 2,
                      "Column": 26,
                      "Offset": 52
                    },
                    "text": "maybe_flag"
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
          "Column": 26,
          "Offset": 52
        },
        "end": {
          "Line": 2,
          "Column": 27,
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
          "Column": 33,
          "Offset": 86
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
              "Column": 33,
              "Offset": 86
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
                  "Column": 33,
                  "Offset": 86
                },
                "name": "Num|Str",
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
          "Column": 33,
          "Offset": 86
        },
        "end": {
          "Line": 3,
          "Column": 34,
          "Offset": 87
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "$flexible",
      "type_expression": {
        "Kind": 0,
        "Name": "Int|Str",
        "Parameters": null,
        "IsUnion": true,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": [
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
          },
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
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Int|Str"
      },
      "position": {
        "Line": 1,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$maybe_flag",
      "type_expression": {
        "Kind": 0,
        "Name": "Bool|Undef",
        "Parameters": null,
        "IsUnion": true,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": [
          {
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
          {
            "Kind": 0,
            "Name": "Undef",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "Undef"
          }
        ],
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Bool|Undef"
      },
      "position": {
        "Line": 2,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$mixed_value",
      "type_expression": {
        "Kind": 0,
        "Name": "Num|Str",
        "Parameters": null,
        "IsUnion": true,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": [
          {
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
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Num|Str"
      },
      "position": {
        "Line": 3,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "source_length": 87
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $flexible = 42;
my $maybe_flag;
my $mixed_value = "text";
```

## Typed Perl Output

```perl
my Int|Str $flexible = 42;
my Bool|Undef $maybe_flag;
my Num|Str $mixed_value = "text";
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
