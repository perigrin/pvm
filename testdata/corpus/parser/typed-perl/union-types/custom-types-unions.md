---
category: typed-perl
subcategory: union-types
tags:
    - custom-types
    - package-qualified
    - union-types
type_check: true
---

# Custom Types Unions

Union types with custom and package-qualified type names

```perl
my MyClass|OtherClass $object;
my Package::Type1|Package::Type2 $qualified;
my UserType|SystemType|DefaultType $flexible;
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 121 characters
  Type Annotations:
    VarAnnotation: $object :: MyClass|OtherClass at 1:1
    VarAnnotation: $qualified :: Package::Type1|Package::Type2 at 2:1
    VarAnnotation: $flexible :: UserType|SystemType|DefaultType at 3:1
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

## Text AST

```
AST {
  Path: /tmp/custom-types-unions.pl
  Source length: 122 characters
  Type Annotations:
    VarAnnotation: $object :: MyClass|OtherClass at 1:1
    VarAnnotation: $qualified :: Package::Type1|Package::Type2 at 2:1
    VarAnnotation: $flexible :: UserType|SystemType|DefaultType at 3:1
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
  "path": "/tmp/custom-types-unions.pl",
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
          "Column": 30,
          "Offset": 29
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
              "Column": 30,
              "Offset": 29
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
                  "Column": 22,
                  "Offset": 21
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
                      "Column": 22,
                      "Offset": 21
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
                          "Column": 11,
                          "Offset": 10
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
                              "Column": 11,
                              "Offset": 10
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
                                  "Column": 11,
                                  "Offset": 10
                                },
                                "value": "MyClass",
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
                          "Column": 11,
                          "Offset": 10
                        },
                        "end": {
                          "Line": 1,
                          "Column": 12,
                          "Offset": 11
                        },
                        "children": [
                          {
                            "type": "literal",
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
                            "value": "|",
                            "kind": "string"
                          }
                        ]
                      },
                      {
                        "type": "type_expression",
                        "start": {
                          "Line": 1,
                          "Column": 12,
                          "Offset": 11
                        },
                        "end": {
                          "Line": 1,
                          "Column": 22,
                          "Offset": 21
                        },
                        "children": [
                          {
                            "type": "expression_stmt",
                            "start": {
                              "Line": 1,
                              "Column": 12,
                              "Offset": 11
                            },
                            "end": {
                              "Line": 1,
                              "Column": 22,
                              "Offset": 21
                            },
                            "children": [
                              {
                                "type": "literal",
                                "start": {
                                  "Line": 1,
                                  "Column": 12,
                                  "Offset": 11
                                },
                                "end": {
                                  "Line": 1,
                                  "Column": 22,
                                  "Offset": 21
                                },
                                "value": "OtherClass",
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
                  "Column": 23,
                  "Offset": 22
                },
                "end": {
                  "Line": 1,
                  "Column": 30,
                  "Offset": 29
                },
                "children": [
                  {
                    "type": "token",
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
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 1,
                      "Column": 24,
                      "Offset": 23
                    },
                    "end": {
                      "Line": 1,
                      "Column": 30,
                      "Offset": 29
                    },
                    "text": "object"
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
      "annotated_item": "$object",
      "type_expression": {
        "Kind": 0,
        "Name": "MyClass|OtherClass",
        "Parameters": null,
        "IsUnion": true,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": [
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
          },
          {
            "Kind": 0,
            "Name": "OtherClass",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "OtherClass"
          }
        ],
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "MyClass|OtherClass"
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
        "Name": "Package::Type1|Package::Type2",
        "Parameters": null,
        "IsUnion": true,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": [
          {
            "Kind": 0,
            "Name": "Package::Type1",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "Package::Type1"
          },
          {
            "Kind": 0,
            "Name": "Package::Type2",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "Package::Type2"
          }
        ],
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Package::Type1|Package::Type2"
      },
      "position": {
        "Line": 2,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$flexible",
      "type_expression": {
        "Kind": 0,
        "Name": "UserType|SystemType|DefaultType",
        "Parameters": null,
        "IsUnion": true,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": [
          {
            "Kind": 0,
            "Name": "UserType",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "UserType"
          },
          {
            "Kind": 0,
            "Name": "SystemType",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "SystemType"
          },
          {
            "Kind": 0,
            "Name": "DefaultType",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "DefaultType"
          }
        ],
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "UserType|SystemType|DefaultType"
      },
      "position": {
        "Line": 3,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "source_length": 122
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 121 characters
  Type Annotations:
    VarAnnotation: $object :: MyClass|OtherClass at 1:1
    VarAnnotation: $qualified :: Package::Type1|Package::Type2 at 2:1
    VarAnnotation: $flexible :: UserType|SystemType|DefaultType at 3:1
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


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $object;
my $qualified;
my $flexible;
```

## Typed Perl Output

```perl
my MyClass|OtherClass $object;
my Package::Type1|Package::Type2 $qualified;
my UserType|SystemType|DefaultType $flexible;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
