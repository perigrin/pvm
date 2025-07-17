---
category: typed-perl
subcategory: simple-annotations
tags:
    - scoping
    - our
    - state
    - local
    - typed-variables
type_check: true
---

# Scoping Keywords

Type annotations with different scoping keywords (our, state, local)

```perl
our Int $global_counter = 0;
state Str $persistent_cache = "";
local Num $localized = 1.0;
```

# Expected AST

## Before Type Inference

### Text Format

```
AST {
  Path:
  Source length: 90 characters
  Type Annotations:
    VarAnnotation: $global_counter :: Int at 1:1
    VarAnnotation: $persistent_cache :: Str at 2:1
    VarAnnotation: $localized :: Num at 3:1
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
      assignment_expression
        variable_declaration
          expression_stmt
            literal
          type_expression
            expression_stmt
              literal
          scalar
            token
            token
        token
        token
    token
}
```

### JSON Format

```json
{
  "path": "/tmp/scoping-keywords.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 3,
      "Column": 28,
      "Offset": 90
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
          "Column": 28,
          "Offset": 27
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
              "Column": 28,
              "Offset": 27
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
                  "Column": 28,
                  "Offset": 27
                },
                "name": "Int",
                "sigil": "$"
              }
            ],
            "decl_type": "our"
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 1,
          "Column": 28,
          "Offset": 27
        },
        "end": {
          "Line": 1,
          "Column": 29,
          "Offset": 28
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 29
        },
        "end": {
          "Line": 2,
          "Column": 33,
          "Offset": 61
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 29
            },
            "end": {
              "Line": 2,
              "Column": 33,
              "Offset": 61
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 29
                },
                "end": {
                  "Line": 2,
                  "Column": 33,
                  "Offset": 61
                },
                "name": "Str",
                "sigil": "$"
              }
            ],
            "decl_type": "state"
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 2,
          "Column": 33,
          "Offset": 61
        },
        "end": {
          "Line": 2,
          "Column": 34,
          "Offset": 62
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 63
        },
        "end": {
          "Line": 3,
          "Column": 27,
          "Offset": 89
        },
        "children": [
          {
            "type": "assignment_expression",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 63
            },
            "end": {
              "Line": 3,
              "Column": 27,
              "Offset": 89
            },
            "children": [
              {
                "type": "variable_declaration",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 63
                },
                "end": {
                  "Line": 3,
                  "Column": 21,
                  "Offset": 83
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 3,
                      "Column": 1,
                      "Offset": 63
                    },
                    "end": {
                      "Line": 3,
                      "Column": 6,
                      "Offset": 68
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 3,
                          "Column": 1,
                          "Offset": 63
                        },
                        "end": {
                          "Line": 3,
                          "Column": 6,
                          "Offset": 68
                        },
                        "value": "local",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "type_expression",
                    "start": {
                      "Line": 3,
                      "Column": 7,
                      "Offset": 69
                    },
                    "end": {
                      "Line": 3,
                      "Column": 10,
                      "Offset": 72
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 3,
                          "Column": 7,
                          "Offset": 69
                        },
                        "end": {
                          "Line": 3,
                          "Column": 10,
                          "Offset": 72
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 7,
                              "Offset": 69
                            },
                            "end": {
                              "Line": 3,
                              "Column": 10,
                              "Offset": 72
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
                      "Line": 3,
                      "Column": 11,
                      "Offset": 73
                    },
                    "end": {
                      "Line": 3,
                      "Column": 21,
                      "Offset": 83
                    },
                    "children": [
                      {
                        "type": "token",
                        "start": {
                          "Line": 3,
                          "Column": 11,
                          "Offset": 73
                        },
                        "end": {
                          "Line": 3,
                          "Column": 12,
                          "Offset": 74
                        },
                        "text": "$"
                      },
                      {
                        "type": "token",
                        "start": {
                          "Line": 3,
                          "Column": 12,
                          "Offset": 74
                        },
                        "end": {
                          "Line": 3,
                          "Column": 21,
                          "Offset": 83
                        },
                        "text": "localized"
                      }
                    ]
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 22,
                  "Offset": 84
                },
                "end": {
                  "Line": 3,
                  "Column": 23,
                  "Offset": 85
                },
                "text": "="
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 24,
                  "Offset": 86
                },
                "end": {
                  "Line": 3,
                  "Column": 27,
                  "Offset": 89
                },
                "text": "1.0"
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 3,
          "Column": 27,
          "Offset": 89
        },
        "end": {
          "Line": 3,
          "Column": 28,
          "Offset": 90
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "$global_counter",
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
      "annotated_item": "$persistent_cache",
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
      "annotated_item": "$localized",
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
        "Line": 3,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "errors": [],
  "source_length": 90
}
```

## After Type Inference

### Text Format

```
AST {
  Path:
  Source length: 90 characters
  Type Annotations:
    VarAnnotation: $global_counter :: Int at 1:1
    VarAnnotation: $persistent_cache :: Str at 2:1
    VarAnnotation: $localized :: Num at 3:1
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
      assignment_expression
        variable_declaration
          expression_stmt
            literal
          type_expression
            expression_stmt
              literal
          scalar
            token
            token
        token
        token
    token
}
```

### JSON Format

```json
{
  "path": "/tmp/scoping-keywords.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 3,
      "Column": 28,
      "Offset": 90
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
          "Column": 28,
          "Offset": 27
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
              "Column": 28,
              "Offset": 27
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
                  "Column": 28,
                  "Offset": 27
                },
                "name": "Int",
                "sigil": "$"
              }
            ],
            "decl_type": "our"
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 1,
          "Column": 28,
          "Offset": 27
        },
        "end": {
          "Line": 1,
          "Column": 29,
          "Offset": 28
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 29
        },
        "end": {
          "Line": 2,
          "Column": 33,
          "Offset": 61
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 29
            },
            "end": {
              "Line": 2,
              "Column": 33,
              "Offset": 61
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 29
                },
                "end": {
                  "Line": 2,
                  "Column": 33,
                  "Offset": 61
                },
                "name": "Str",
                "sigil": "$"
              }
            ],
            "decl_type": "state"
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 2,
          "Column": 33,
          "Offset": 61
        },
        "end": {
          "Line": 2,
          "Column": 34,
          "Offset": 62
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 63
        },
        "end": {
          "Line": 3,
          "Column": 27,
          "Offset": 89
        },
        "children": [
          {
            "type": "assignment_expression",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 63
            },
            "end": {
              "Line": 3,
              "Column": 27,
              "Offset": 89
            },
            "children": [
              {
                "type": "variable_declaration",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 63
                },
                "end": {
                  "Line": 3,
                  "Column": 21,
                  "Offset": 83
                },
                "children": [
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 3,
                      "Column": 1,
                      "Offset": 63
                    },
                    "end": {
                      "Line": 3,
                      "Column": 6,
                      "Offset": 68
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 3,
                          "Column": 1,
                          "Offset": 63
                        },
                        "end": {
                          "Line": 3,
                          "Column": 6,
                          "Offset": 68
                        },
                        "value": "local",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "type_expression",
                    "start": {
                      "Line": 3,
                      "Column": 7,
                      "Offset": 69
                    },
                    "end": {
                      "Line": 3,
                      "Column": 10,
                      "Offset": 72
                    },
                    "children": [
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 3,
                          "Column": 7,
                          "Offset": 69
                        },
                        "end": {
                          "Line": 3,
                          "Column": 10,
                          "Offset": 72
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 3,
                              "Column": 7,
                              "Offset": 69
                            },
                            "end": {
                              "Line": 3,
                              "Column": 10,
                              "Offset": 72
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
                      "Line": 3,
                      "Column": 11,
                      "Offset": 73
                    },
                    "end": {
                      "Line": 3,
                      "Column": 21,
                      "Offset": 83
                    },
                    "children": [
                      {
                        "type": "token",
                        "start": {
                          "Line": 3,
                          "Column": 11,
                          "Offset": 73
                        },
                        "end": {
                          "Line": 3,
                          "Column": 12,
                          "Offset": 74
                        },
                        "text": "$"
                      },
                      {
                        "type": "token",
                        "start": {
                          "Line": 3,
                          "Column": 12,
                          "Offset": 74
                        },
                        "end": {
                          "Line": 3,
                          "Column": 21,
                          "Offset": 83
                        },
                        "text": "localized"
                      }
                    ]
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 22,
                  "Offset": 84
                },
                "end": {
                  "Line": 3,
                  "Column": 23,
                  "Offset": 85
                },
                "text": "="
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 24,
                  "Offset": 86
                },
                "end": {
                  "Line": 3,
                  "Column": 27,
                  "Offset": 89
                },
                "text": "1.0"
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 3,
          "Column": 27,
          "Offset": 89
        },
        "end": {
          "Line": 3,
          "Column": 28,
          "Offset": 90
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "$global_counter",
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
      "annotated_item": "$persistent_cache",
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
      "annotated_item": "$localized",
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
        "Line": 3,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "errors": [],
  "source_length": 90
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
our $global_counter = 0;
state $persistent_cache = "";
local $localized = 1.0;
```

## Typed Perl Output

```perl
our Int $global_counter = 0;
state Str $persistent_cache = "";
local Num $localized = 1.0;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
