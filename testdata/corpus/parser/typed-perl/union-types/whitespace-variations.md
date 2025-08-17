---
category: typed-perl
subcategory: union-types
tags:
    - whitespace
    - formatting
    - union-types
type_check: true
---

# Whitespace Variations

Union types with different whitespace formatting variations

```perl
my Int | Str $spaced;
my Int|Str|Bool $compact;
my  Num  |  Str  |  Bool  $extra_spaced;
my Int|
    Str|
    Bool $multiline;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 126 characters
  Type Annotations:
    VarAnnotation: $spaced :: Int | Str at 1:1
    VarAnnotation: $compact :: Int|Str|Bool at 2:1
    VarAnnotation: $extra_spaced :: Num  |  Str  |  Bool at 3:1
    VarAnnotation: $multiline :: Int|
    Str|
    Bool at 4:1
  Root: source_file
  Tree Structure:
  source_file
    expression_stmt
      literal
    token
    expression_stmt
      literal
    token
    expression_stmt
      literal
    token
    expression_stmt
      literal
    token
}
```

## Text AST

```
AST {
  Path: /tmp/whitespace-variations.pl
  Source length: 127 characters
  Type Annotations:
    VarAnnotation: $spaced :: Int | Str at 1:1
    VarAnnotation: $compact :: Int|Str|Bool at 2:1
    VarAnnotation: $extra_spaced :: Num  |  Str  |  Bool at 3:1
    VarAnnotation: $multiline :: Int|
    Str|
    Bool at 4:1
  Root: source_file
  Tree Structure:
  source_file
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
      variable_declaration
        token
        type_expression
          union_type
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
      variable_declaration
        token
        type_expression
          union_type
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
      variable_declaration
        token
        type_expression
          union_type
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
            expression_stmt
              literal
            type_expression
              expression_stmt
                literal
        scalar
          token
          token
    token
}
```

## JSON AST

```json
{
  "path": "/tmp/whitespace-variations.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 7,
      "Column": 1,
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
              "Column": 21,
              "Offset": 20
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
                  "Column": 13,
                  "Offset": 12
                },
                "children": [
                  {
                    "type": "union_type",
                    "start": {
                      "Line": 1,
                      "Column": 4,
                      "Offset": 3
                    },
                    "end": {
                      "Line": 1,
                      "Column": 13,
                      "Offset": 12
                    },
                    "children": [
                      {
                        "type": "type_expression",
                        "start": {
                          "Line": 1,
                          "Column": 4,
                          "Offset": 3
                        },
                        "end": {
                          "Line": 1,
                          "Column": 7,
                          "Offset": 6
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
                              "Column": 7,
                              "Offset": 6
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
                                  "Column": 7,
                                  "Offset": 6
                                },
                                "value": "Int",
                                "kind": "string"
                              }
                            ]
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 8,
                          "Offset": 7
                        },
                        "end": {
                          "Line": 1,
                          "Column": 9,
                          "Offset": 8
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 8,
                              "Offset": 7
                            },
                            "end": {
                              "Line": 1,
                              "Column": 9,
                              "Offset": 8
                            },
                            "value": "|",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "type_expression",
                        "start": {
                          "Line": 1,
                          "Column": 10,
                          "Offset": 9
                        },
                        "end": {
                          "Line": 1,
                          "Column": 13,
                          "Offset": 12
                        },
                        "children": [
                          {
                            "type": "expression_stmt",
                            "start": {
                              "Line": 1,
                              "Column": 10,
                              "Offset": 9
                            },
                            "end": {
                              "Line": 1,
                              "Column": 13,
                              "Offset": 12
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 1,
                                  "Column": 10,
                                  "Offset": 9
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 13,
                                  "Offset": 12
                                },
                                "value": "Str",
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
                  "Line": 1,
                  "Column": 14,
                  "Offset": 13
                },
                "end": {
                  "Line": 1,
                  "Column": 21,
                  "Offset": 20
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 14,
                      "Offset": 13
                    },
                    "end": {
                      "Line": 1,
                      "Column": 15,
                      "Offset": 14
                    },
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 15,
                      "Offset": 14
                    },
                    "end": {
                      "Line": 1,
                      "Column": 21,
                      "Offset": 20
                    },
                    "text": "spaced"
                  }
                ]
              }
            ]
          }
        ]
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "$spaced",
      "type_expression": {
        "Kind": 0,
        "Name": "Int | Str",
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
        "OriginalString": "Int | Str"
      },
      "position": {
        "Line": 1,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$compact",
      "type_expression": {
        "Kind": 0,
        "Name": "Int|Str|Bool",
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
          },
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
          }
        ],
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Int|Str|Bool"
      },
      "position": {
        "Line": 2,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$extra_spaced",
      "type_expression": {
        "Kind": 0,
        "Name": "Num  |  Str  |  Bool",
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
          },
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
          }
        ],
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Num  |  Str  |  Bool"
      },
      "position": {
        "Line": 3,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$multiline",
      "type_expression": {
        "Kind": 0,
        "Name": "Int|\n    Str|\n    Bool",
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
          },
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
          }
        ],
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Int|\n    Str|\n    Bool"
      },
      "position": {
        "Line": 4,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "source_length": 127
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 126 characters
  Type Annotations:
    VarAnnotation: $spaced :: Int | Str at 1:1
    VarAnnotation: $compact :: Int|Str|Bool at 2:1
    VarAnnotation: $extra_spaced :: Num  |  Str  |  Bool at 3:1
    VarAnnotation: $multiline :: Int|
    Str|
    Bool at 4:1
  Root: source_file
  Tree Structure:
  source_file
    expression_stmt
      literal
    token
    expression_stmt
      literal
    token
    expression_stmt
      literal
    token
    expression_stmt
      literal
    token
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $spaced;
my $compact;
my $extra_spaced;
my $multiline;
```

## Typed Perl Output

```perl
my Int | Str $spaced;
my Int|Str|Bool $compact;
my  Num  |  Str  |  Bool  $extra_spaced;
my Int|
    Str|
    Bool $multiline;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
