---
category: typed-perl
subcategory: classes-roles
tags:
    - generic-class
    - type-parameters
    - type-constraints
    - parameterized-methods
type_check: true
---

# Generic Class Declarations

Generic class with type parameters and constraints

```perl
class Container<T> where T: Serializable {
    field ArrayRef[T] $items = [];

    method Void add(T $item) {
        push @{$items}, $item;
    }

    method ArrayRef[T] get_all() {
        return $items;
    }

    method Optional[T] find(CodeRef[T, Bool] $predicate) {
        for my $item (@{$items}) {
            return $item if $predicate->($item);
        }
        return undef;
    }
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 395 characters
  Type Annotations:
    MethodReturnAnnotation: add :: Void at 4:12
    MethodParamAnnotation: $item :: T at 4:21
    MethodReturnAnnotation: get_all :: ArrayRef[T] at 8:12
    MethodReturnAnnotation: find :: Optional[T] at 12:12
    MethodParamAnnotation: $predicate :: CodeRef[T, Bool] at 12:29
    VarAnnotation: Container :: class at 1:1
    VarAnnotation: $items :: ArrayRef[T] at 2:5
  Root: source_file
  Tree Structure:
  source_file
    class_decl
      method_decl
        type_expr
        block_stmt
          token
          return_stmt
            literal
          token
          token
      method_decl
        type_expr
        block_stmt
          token
          return_stmt
            literal
          token
          token
      method_decl
        type_expr
        block_stmt
          token
          return_stmt
            literal
          return_stmt
            literal
          token
          token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 395 characters
  Type Annotations:
    MethodReturnAnnotation: add :: Void at 4:12
    MethodParamAnnotation: $item :: T at 4:21
    MethodReturnAnnotation: get_all :: ArrayRef[T] at 8:12
    MethodReturnAnnotation: find :: Optional[T] at 12:12
    MethodParamAnnotation: $predicate :: CodeRef[T, Bool] at 12:29
    VarAnnotation: Container :: class at 1:1
    VarAnnotation: $items :: ArrayRef[T] at 2:5
  Root: source_file
  Tree Structure:
  source_file
    class_decl
      method_decl
        type_expr
        block_stmt
          token
          return_stmt
            literal
          token
          token
      method_decl
        type_expr
        block_stmt
          token
          return_stmt
            literal
          token
          token
      method_decl
        type_expr
        block_stmt
          token
          return_stmt
            literal
          return_stmt
            literal
          token
          token
}
```

## JSON AST

```json
{
  "path": "/tmp/generic-class.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 19,
      "Column": 1,
      "Offset": 396
    },
    "children": [
      {
        "type": "class_decl",
        "start": {
          "Line": 1,
          "Column": 1,
          "Offset": 0
        },
        "end": {
          "Line": 18,
          "Column": 2,
          "Offset": 395
        },
        "children": [
          {
            "type": "method_decl",
            "start": {
              "Line": 4,
              "Column": 5,
              "Offset": 0
            },
            "end": {
              "Line": 6,
              "Column": 6,
              "Offset": 0
            },
            "children": [
              {
                "type": "block_stmt",
                "start": {
                  "Line": 4,
                  "Column": 30,
                  "Offset": 108
                },
                "end": {
                  "Line": 6,
                  "Column": 6,
                  "Offset": 146
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 4,
                      "Column": 30,
                      "Offset": 108
                    },
                    "end": {
                      "Line": 4,
                      "Column": 31,
                      "Offset": 109
                    },
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 5,
                      "Column": 9,
                      "Offset": 118
                    },
                    "end": {
                      "Line": 5,
                      "Column": 30,
                      "Offset": 139
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 5,
                          "Column": 9,
                          "Offset": 118
                        },
                        "end": {
                          "Line": 5,
                          "Column": 30,
                          "Offset": 139
                        },
                        "value": "push @{$items}, $item",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 5,
                      "Column": 30,
                      "Offset": 139
                    },
                    "end": {
                      "Line": 5,
                      "Column": 31,
                      "Offset": 140
                    },
                    "text": ";"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 6,
                      "Column": 5,
                      "Offset": 145
                    },
                    "end": {
                      "Line": 6,
                      "Column": 6,
                      "Offset": 146
                    },
                    "text": "}"
                  }
                ]
              }
            ],
            "name": "add"
          },
          {
            "type": "method_decl",
            "start": {
              "Line": 8,
              "Column": 5,
              "Offset": 0
            },
            "end": {
              "Line": 10,
              "Column": 6,
              "Offset": 0
            },
            "children": [
              {
                "type": "block_stmt",
                "start": {
                  "Line": 8,
                  "Column": 34,
                  "Offset": 181
                },
                "end": {
                  "Line": 10,
                  "Column": 6,
                  "Offset": 211
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 8,
                      "Column": 34,
                      "Offset": 181
                    },
                    "end": {
                      "Line": 8,
                      "Column": 35,
                      "Offset": 182
                    },
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 9,
                      "Column": 9,
                      "Offset": 191
                    },
                    "end": {
                      "Line": 9,
                      "Column": 22,
                      "Offset": 204
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 9,
                          "Column": 9,
                          "Offset": 191
                        },
                        "end": {
                          "Line": 9,
                          "Column": 22,
                          "Offset": 204
                        },
                        "value": "return $items",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 9,
                      "Column": 22,
                      "Offset": 204
                    },
                    "end": {
                      "Line": 9,
                      "Column": 23,
                      "Offset": 205
                    },
                    "text": ";"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 10,
                      "Column": 5,
                      "Offset": 210
                    },
                    "end": {
                      "Line": 10,
                      "Column": 6,
                      "Offset": 211
                    },
                    "text": "}"
                  }
                ]
              }
            ],
            "name": "get_all"
          },
          {
            "type": "method_decl",
            "start": {
              "Line": 12,
              "Column": 5,
              "Offset": 0
            },
            "end": {
              "Line": 17,
              "Column": 6,
              "Offset": 0
            },
            "children": [
              {
                "type": "block_stmt",
                "start": {
                  "Line": 12,
                  "Column": 58,
                  "Offset": 270
                },
                "end": {
                  "Line": 17,
                  "Column": 6,
                  "Offset": 393
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 12,
                      "Column": 58,
                      "Offset": 270
                    },
                    "end": {
                      "Line": 12,
                      "Column": 59,
                      "Offset": 271
                    },
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 13,
                      "Column": 9,
                      "Offset": 280
                    },
                    "end": {
                      "Line": 15,
                      "Column": 10,
                      "Offset": 365
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 13,
                          "Column": 9,
                          "Offset": 280
                        },
                        "end": {
                          "Line": 15,
                          "Column": 10,
                          "Offset": 365
                        },
                        "value": "for my $item (@{$items}) {\n            return $item if $predicate->($item);\n        }",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 16,
                      "Column": 9,
                      "Offset": 374
                    },
                    "end": {
                      "Line": 16,
                      "Column": 21,
                      "Offset": 386
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 16,
                          "Column": 9,
                          "Offset": 374
                        },
                        "end": {
                          "Line": 16,
                          "Column": 21,
                          "Offset": 386
                        },
                        "value": "return undef",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 16,
                      "Column": 21,
                      "Offset": 386
                    },
                    "end": {
                      "Line": 16,
                      "Column": 22,
                      "Offset": 387
                    },
                    "text": ";"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 17,
                      "Column": 5,
                      "Offset": 392
                    },
                    "end": {
                      "Line": 17,
                      "Column": 6,
                      "Offset": 393
                    },
                    "text": "}"
                  }
                ]
              }
            ],
            "name": "find"
          }
        ],
        "name": "Container"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "add",
      "type_expression": {
        "Kind": 0,
        "Name": "Void",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Void"
      },
      "position": {
        "Line": 4,
        "Column": 12,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
    },
    {
      "annotated_item": "$item",
      "type_expression": {
        "Kind": 0,
        "Name": "T",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "T"
      },
      "position": {
        "Line": 4,
        "Column": 21,
        "Offset": 0
      },
      "kind": "MethodParamAnnotation"
    },
    {
      "annotated_item": "get_all",
      "type_expression": {
        "Kind": 4,
        "Name": "ArrayRef[T]",
        "Parameters": [
          {
            "Kind": 0,
            "Name": "T",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "T"
          }
        ],
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "ArrayRef[T]"
      },
      "position": {
        "Line": 8,
        "Column": 12,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
    },
    {
      "annotated_item": "find",
      "type_expression": {
        "Kind": 4,
        "Name": "Optional[T]",
        "Parameters": [
          {
            "Kind": 0,
            "Name": "T",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "T"
          }
        ],
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Optional[T]"
      },
      "position": {
        "Line": 12,
        "Column": 12,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
    },
    {
      "annotated_item": "$predicate",
      "type_expression": {
        "Kind": 4,
        "Name": "CodeRef[T,Bool]",
        "Parameters": [
          {
            "Kind": 0,
            "Name": "T",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "T"
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
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "CodeRef[T, Bool]"
      },
      "position": {
        "Line": 12,
        "Column": 29,
        "Offset": 0
      },
      "kind": "MethodParamAnnotation"
    },
    {
      "annotated_item": "Container",
      "type_expression": {
        "Kind": 0,
        "Name": "class",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "class"
      },
      "position": {
        "Line": 1,
        "Column": 1,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$items",
      "type_expression": {
        "Kind": 4,
        "Name": "ArrayRef[T]",
        "Parameters": [
          {
            "Kind": 0,
            "Name": "T",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "T"
          }
        ],
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "ArrayRef[T]"
      },
      "position": {
        "Line": 2,
        "Column": 5,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "errors": [],
  "source_length": 396
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
class Container<T> where T:  {
    field $items = [];

    method add($item) {
        push @{$items}, $item;
    }

    method get_all() {
        return $items;
    }

    method find($predicate) {
        for my $item (@{$items}) {
            return $item if $predicate->($item);
        }
        return undef;
    }
}
```

## Typed Perl Output

```perl
class Container<T> where T: Serializable {
    field ArrayRef[T] $items = [];

    method Void add(T $item) {
        push @{$items}, $item;
    }

    method ArrayRef[T] get_all() {
        return $items;
    }

    method Optional[T] find(CodeRef[T, Bool] $predicate) {
        for my $item (@{$items}) {
            return $item if $predicate->($item);
        }
        return undef;
    }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none)
```
