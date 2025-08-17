---
category: typed-perl
subcategory: methods-fields
tags:
    - field-declarations
    - field-initialization
    - typed-fields
type_check: true
---

# Basic Field Declarations

Basic typed field declarations with and without initializers

```perl
field Int $count = 0;
field Str $name;
field Bool $is_active = 1;
field Num $rate = 3.14;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 89 characters
  Type Annotations:
    VarAnnotation: $count :: Int at 1:1
    VarAnnotation: $name :: Str at 2:1
    VarAnnotation: $is_active :: Bool at 3:1
    VarAnnotation: $rate :: Num at 4:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      assignment_expression
        token
        token
    token
    expression_stmt
      literal
    token
    expression_statement
      assignment_expression
        token
        token
    token
    expression_statement
      assignment_expression
        token
        token
    token
}
```

## JSON AST

```json
{
  "path": "/tmp/test_code_1.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 4,
      "Column": 24,
      "Offset": 89
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
          "Column": 21,
          "Offset": 20
        },
        "children": [
          {
            "type": "assignment_expression",
            "start": {
              "Line": 1,
              "Column": 1,
              "Offset": 0
            },
            "end": {
              "Line": 1,
              "Column": 21,
              "Offset": 20
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
                  "Column": 17,
                  "Offset": 16
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
                      "Column": 6,
                      "Offset": 5
                    },
                    "text": "field"
                  },
                  {
                    "type": "type_expression",
                    "start": {
                      "Line": 1,
                      "Column": 7,
                      "Offset": 6
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
                          "Column": 7,
                          "Offset": 6
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
                              "Column": 7,
                              "Offset": 6
                            },
                            "end": {
                              "Line": 1,
                              "Column": 10,
                              "Offset": 9
                            },
                            "value": "Int",
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
                      "Column": 17,
                      "Offset": 16
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
                          "Column": 17,
                          "Offset": 16
                        },
                        "text": "count"
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
                "text": "="
              },
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 20,
                  "Offset": 19
                },
                "end": {
                  "Line": 1,
                  "Column": 21,
                  "Offset": 20
                },
                "text": "0"
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 1,
          "Column": 21,
          "Offset": 20
        },
        "end": {
          "Line": 1,
          "Column": 22,
          "Offset": 21
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 22
        },
        "end": {
          "Line": 2,
          "Column": 16,
          "Offset": 37
        },
        "children": [
          {
            "type": "variable_declaration",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 22
            },
            "end": {
              "Line": 2,
              "Column": 16,
              "Offset": 37
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 22
                },
                "end": {
                  "Line": 2,
                  "Column": 6,
                  "Offset": 27
                },
                "text": "field"
              },
              {
                "type": "type_expression",
                "start": {
                  "Line": 2,
                  "Column": 7,
                  "Offset": 28
                },
                "end": {
                  "Line": 2,
                  "Column": 10,
                  "Offset": 31
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 2,
                      "Column": 7,
                      "Offset": 28
                    },
                    "end": {
                      "Line": 2,
                      "Column": 10,
                      "Offset": 31
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 2,
                          "Column": 7,
                          "Offset": 28
                        },
                        "end": {
                          "Line": 2,
                          "Column": 10,
                          "Offset": 31
                        },
                        "value": "Str",
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
                  "Column": 11,
                  "Offset": 32
                },
                "end": {
                  "Line": 2,
                  "Column": 16,
                  "Offset": 37
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 11,
                      "Offset": 32
                    },
                    "end": {
                      "Line": 2,
                      "Column": 12,
                      "Offset": 33
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 2,
                      "Column": 12,
                      "Offset": 33
                    },
                    "end": {
                      "Line": 2,
                      "Column": 16,
                      "Offset": 37
                    },
                    "text": "name"
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
          "Column": 16,
          "Offset": 37
        },
        "end": {
          "Line": 2,
          "Column": 17,
          "Offset": 38
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 39
        },
        "end": {
          "Line": 3,
          "Column": 26,
          "Offset": 64
        },
        "children": [
          {
            "type": "assignment_expression",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 39
            },
            "end": {
              "Line": 3,
              "Column": 26,
              "Offset": 64
            },
            "children": [
              {
                "type": "variable_declaration",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 39
                },
                "end": {
                  "Line": 3,
                  "Column": 22,
                  "Offset": 60
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 3,
                      "Column": 1,
                      "Offset": 39
                    },
                    "end": {
                      "Line": 3,
                      "Column": 6,
                      "Offset": 44
                    },
                    "text": "field"
                  },
                  {
                    "type": "type_expression",
                    "start": {
                      "Line": 3,
                      "Column": 7,
                      "Offset": 45
                    },
                    "end": {
                      "Line": 3,
                      "Column": 11,
                      "Offset": 49
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 3,
                          "Column": 7,
                          "Offset": 45
                        },
                        "end": {
                          "Line": 3,
                          "Column": 11,
                          "Offset": 49
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 7,
                              "Offset": 45
                            },
                            "end": {
                              "Line": 3,
                              "Column": 11,
                              "Offset": 49
                            },
                            "value": "Bool",
                            "kind": "string"
                          }
                        ]
                      }
                    ]
                  },
                  {
                    "type": "scalar",
                    "start": {
                      "Line": 3,
                      "Column": 12,
                      "Offset": 50
                    },
                    "end": {
                      "Line": 3,
                      "Column": 22,
                      "Offset": 60
                    },
                    "children": [
                      {
                        "type": "token",
                        "start": {
                          "Line": 3,
                          "Column": 12,
                          "Offset": 50
                        },
                        "end": {
                          "Line": 3,
                          "Column": 13,
                          "Offset": 51
                        },
                        "text": "$"
                      },
                      {
                        "type": "token",
                        "start": {
                          "Line": 3,
                          "Column": 13,
                          "Offset": 51
                        },
                        "end": {
                          "Line": 3,
                          "Column": 22,
                          "Offset": 60
                        },
                        "text": "is_active"
                      }
                    ]
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 23,
                  "Offset": 61
                },
                "end": {
                  "Line": 3,
                  "Column": 24,
                  "Offset": 62
                },
                "text": "="
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 25,
                  "Offset": 63
                },
                "end": {
                  "Line": 3,
                  "Column": 26,
                  "Offset": 64
                },
                "text": "1"
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 3,
          "Column": 26,
          "Offset": 64
        },
        "end": {
          "Line": 3,
          "Column": 27,
          "Offset": 65
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 4,
          "Column": 1,
          "Offset": 66
        },
        "end": {
          "Line": 4,
          "Column": 23,
          "Offset": 88
        },
        "children": [
          {
            "type": "assignment_expression",
            "start": {
              "Line": 4,
              "Column": 1,
              "Offset": 66
            },
            "end": {
              "Line": 4,
              "Column": 23,
              "Offset": 88
            },
            "children": [
              {
                "type": "variable_declaration",
                "start": {
                  "Line": 4,
                  "Column": 1,
                  "Offset": 66
                },
                "end": {
                  "Line": 4,
                  "Column": 16,
                  "Offset": 81
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 4,
                      "Column": 1,
                      "Offset": 66
                    },
                    "end": {
                      "Line": 4,
                      "Column": 6,
                      "Offset": 71
                    },
                    "text": "field"
                  },
                  {
                    "type": "type_expression",
                    "start": {
                      "Line": 4,
                      "Column": 7,
                      "Offset": 72
                    },
                    "end": {
                      "Line": 4,
                      "Column": 10,
                      "Offset": 75
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 4,
                          "Column": 7,
                          "Offset": 72
                        },
                        "end": {
                          "Line": 4,
                          "Column": 10,
                          "Offset": 75
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 4,
                              "Column": 7,
                              "Offset": 72
                            },
                            "end": {
                              "Line": 4,
                              "Column": 10,
                              "Offset": 75
                            },
                            "value": "Num",
                            "kind": "string"
                          }
                        ]
                      }
                    ]
                  },
                  {
                    "type": "scalar",
                    "start": {
                      "Line": 4,
                      "Column": 11,
                      "Offset": 76
                    },
                    "end": {
                      "Line": 4,
                      "Column": 16,
                      "Offset": 81
                    },
                    "children": [
                      {
                        "type": "token",
                        "start": {
                          "Line": 4,
                          "Column": 11,
                          "Offset": 76
                        },
                        "end": {
                          "Line": 4,
                          "Column": 12,
                          "Offset": 77
                        },
                        "text": "$"
                      },
                      {
                        "type": "token",
                        "start": {
                          "Line": 4,
                          "Column": 12,
                          "Offset": 77
                        },
                        "end": {
                          "Line": 4,
                          "Column": 16,
                          "Offset": 81
                        },
                        "text": "rate"
                      }
                    ]
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 4,
                  "Column": 17,
                  "Offset": 82
                },
                "end": {
                  "Line": 4,
                  "Column": 18,
                  "Offset": 83
                },
                "text": "="
              },
              {
                "type": "token",
                "start": {
                  "Line": 4,
                  "Column": 19,
                  "Offset": 84
                },
                "end": {
                  "Line": 4,
                  "Column": 23,
                  "Offset": 88
                },
                "text": "3.14"
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 4,
          "Column": 23,
          "Offset": 88
        },
        "end": {
          "Line": 4,
          "Column": 24,
          "Offset": 89
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
      "annotated_item": "$is_active",
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
      "annotated_item": "$rate",
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
  "source_length": 89
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 89 characters
  Type Annotations:
    VarAnnotation: $count :: Int at 1:1
    VarAnnotation: $name :: Str at 2:1
    VarAnnotation: $is_active :: Bool at 3:1
    VarAnnotation: $rate :: Num at 4:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      assignment_expression
        token
        token
    token
    expression_stmt
      literal
    token
    expression_statement
      assignment_expression
        token
        token
    token
    expression_statement
      assignment_expression
        token
        token
    token
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
field $count = 0;
field $name;
field $is_active = 1;
field $rate = 3.14;
```

## Typed Perl Output

```perl
field Int $count = 0;
field Str $name;
field Bool $is_active = 1;
field Num $rate = 3.14;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
