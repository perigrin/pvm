---
category: typed-perl
subcategory: parameterized-types
tags:
    - field-declarations
    - class-fields
    - parameterized-types
type_check: true
---

# Field Declarations

Parameterized types in field declarations

```perl
field ArrayRef[MyClass] $objects;
field HashRef[ArrayRef[Str]] $nested_data;
field Optional[ArrayRef[Int]] $maybe_numbers;
```

# Expected AST

## Before Type Inference

### Text Format

```
AST {
  Path:
  Source length: 122 characters
  Type Annotations:
    VarAnnotation: $objects :: ArrayRef[MyClass] at 1:1
    VarAnnotation: $nested_data :: HashRef[ArrayRef[Str]] at 2:1
    VarAnnotation: $maybe_numbers :: Optional[ArrayRef[Int]] at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      variable_declaration
        token
        type_expression
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                expression_stmt
                  literal
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
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                parameterized_type
                  expression_stmt
                    literal
                  expression_stmt
                    literal
                  type_parameter_list
                    type_expression
                      expression_stmt
                        literal
                  expression_stmt
                    literal
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
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                parameterized_type
                  expression_stmt
                    literal
                  expression_stmt
                    literal
                  type_parameter_list
                    type_expression
                      expression_stmt
                        literal
                  expression_stmt
                    literal
            expression_stmt
              literal
        scalar
          token
          token
    token
}
```

### JSON Format

```json
{
  "path": "/tmp/field-declarations.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 3,
      "Column": 46,
      "Offset": 122
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
          "Column": 33,
          "Offset": 32
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
              "Column": 33,
              "Offset": 32
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
                  "Column": 24,
                  "Offset": 23
                },
                "children": [
                  {
                    "type": "parameterized_type",
                    "start": {
                      "Line": 1,
                      "Column": 7,
                      "Offset": 6
                    },
                    "end": {
                      "Line": 1,
                      "Column": 24,
                      "Offset": 23
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
                          "Column": 15,
                          "Offset": 14
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
                              "Column": 15,
                              "Offset": 14
                            },
                            "value": "ArrayRef",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 15,
                          "Offset": 14
                        },
                        "end": {
                          "Line": 1,
                          "Column": 16,
                          "Offset": 15
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 15,
                              "Offset": 14
                            },
                            "end": {
                              "Line": 1,
                              "Column": 16,
                              "Offset": 15
                            },
                            "value": "[",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "type_parameter_list",
                        "start": {
                          "Line": 1,
                          "Column": 16,
                          "Offset": 15
                        },
                        "end": {
                          "Line": 1,
                          "Column": 23,
                          "Offset": 22
                        },
                        "children": [
                          {
                            "type": "type_expression",
                            "start": {
                              "Line": 1,
                              "Column": 16,
                              "Offset": 15
                            },
                            "end": {
                              "Line": 1,
                              "Column": 23,
                              "Offset": 22
                            },
                            "children": [
                              {
                                "type": "expression_stmt",
                                "start": {
                                  "Line": 1,
                                  "Column": 16,
                                  "Offset": 15
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 23,
                                  "Offset": 22
                                },
                                "children": [
                                  {
                                    "type": "literal",
                                    "start": {
                                      "Line": 1,
                                      "Column": 16,
                                      "Offset": 15
                                    },
                                    "end": {
                                      "Line": 1,
                                      "Column": 23,
                                      "Offset": 22
                                    },
                                    "value": "MyClass",
                                    "kind": "string"
                                  }
                                ]
                              }
                            ]
                          }
                        ]
                      },
                      {
                        "type": "expression_stmt",
                        "start": {
                          "Line": 1,
                          "Column": 23,
                          "Offset": 22
                        },
                        "end": {
                          "Line": 1,
                          "Column": 24,
                          "Offset": 23
                        },
                        "children": [
                          {
                            "type": "literal",
                            "start": {
                              "Line": 1,
                              "Column": 23,
                              "Offset": 22
                            },
                            "end": {
                              "Line": 1,
                              "Column": 24,
                              "Offset": 23
                            },
                            "value": "]",
                            "kind": "string"
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
                  "Column": 25,
                  "Offset": 24
                },
                "end": {
                  "Line": 1,
                  "Column": 33,
                  "Offset": 32
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 25,
                      "Offset": 24
                    },
                    "end": {
                      "Line": 1,
                      "Column": 26,
                      "Offset": 25
                    },
                    "text": "$"
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
                      "Column": 33,
                      "Offset": 32
                    },
                    "text": "objects"
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
      "annotated_item": "$objects",
      "type_expression": {
        "Kind": 4,
        "Name": "ArrayRef[MyClass]",
        "Parameters": [
          {
            "Kind": 0,
            "Name": "MyClass",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "MyClass"
          }
        ],
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "ArrayRef[MyClass]"
      },
      "position": {
        "Line": 1,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "errors": [],
  "source_length": 122
}
```

## After Type Inference

### Text Format

```
AST {
  Path:
  Source length: 122 characters
  Type Annotations:
    VarAnnotation: $objects :: ArrayRef[MyClass] at 1:1
    VarAnnotation: $nested_data :: HashRef[ArrayRef[Str]] at 2:1
    VarAnnotation: $maybe_numbers :: Optional[ArrayRef[Int]] at 3:1
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      variable_declaration
        token
        type_expression
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                expression_stmt
                  literal
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
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                parameterized_type
                  expression_stmt
                    literal
                  expression_stmt
                    literal
                  type_parameter_list
                    type_expression
                      expression_stmt
                        literal
                  expression_stmt
                    literal
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
          parameterized_type
            expression_stmt
              literal
            expression_stmt
              literal
            type_parameter_list
              type_expression
                parameterized_type
                  expression_stmt
                    literal
                  expression_stmt
                    literal
                  type_parameter_list
                    type_expression
                      expression_stmt
                        literal
                  expression_stmt
                    literal
            expression_stmt
              literal
        scalar
          token
          token
    token
}
```

### JSON Format

```json
{
  "path": "/tmp/field-declarations.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 3,
      "Column": 46,
      "Offset": 122
    }
  },
  "type_annotations": [
    {
      "annotated_item": "$objects",
      "type_expression": {
        "Kind": 4,
        "Name": "ArrayRef[MyClass]",
        "Parameters": [
          {
            "Kind": 0,
            "Name": "MyClass",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "MyClass"
          }
        ],
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "ArrayRef[MyClass]"
      },
      "position": {
        "Line": 1,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$nested_data",
      "type_expression": {
        "Kind": 4,
        "Name": "HashRef[ArrayRef[Str]]",
        "Parameters": [
          {
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
          }
        ],
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "HashRef[ArrayRef[Str]]"
      },
      "position": {
        "Line": 2,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$maybe_numbers",
      "type_expression": {
        "Kind": 4,
        "Name": "Optional[ArrayRef[Int]]",
        "Parameters": [
          {
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
          }
        ],
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Optional[ArrayRef[Int]]"
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
  "source_length": 122
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
field $objects;
field $nested_data;
field $maybe_numbers;
```

## Typed Perl Output

```perl
field ArrayRef[MyClass] $objects;
field HashRef[ArrayRef[Str]] $nested_data;
field Optional[ArrayRef[Int]] $maybe_numbers;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
