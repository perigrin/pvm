---
category: typed-perl
subcategory: classes-roles
tags:
    - class-declaration
    - typed-fields
    - typed-methods
    - basic
type_check: true
---

# Basic Class Declarations

Basic class with typed fields and methods

```perl
class User {
    field Str $name;
    field Int $age;
    field Optional[Email] $email;

    method User new(Str $name, Int $age) {
        return bless {
            name => $name,
            age => $age
        }, __PACKAGE__;
    }

    method Str get_name() {
        return $name;
    }
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 294 characters
  Type Annotations:
    MethodReturnAnnotation: new :: User at 6:12
    MethodParamAnnotation: $name :: Str at 6:21
    MethodParamAnnotation: $age :: Int at 6:32
    MethodReturnAnnotation: get_name :: Str at 13:12
    VarAnnotation: User :: class at 1:1
    VarAnnotation: $name :: Str at 2:5
    VarAnnotation: $age :: Int at 3:5
    VarAnnotation: $email :: Optional[Email] at 4:5
  Root: source_file
  Tree Structure:
  source_file
    class_decl
      field_decl
        variable
      field_decl
        variable
      field_decl
        variable
      method_decl
        type_expr
        block_stmt
          token
          expression_stmt
            literal
          token
          token
      method_decl
        type_expr
        block_stmt
          token
          expression_stmt
            literal
          token
          token
}
```

## After Type Inference

```
AST {
  Path:
  Source length: 294 characters
  Type Annotations:
    MethodReturnAnnotation: new :: User at 6:12
    MethodParamAnnotation: $name :: Str at 6:21
    MethodParamAnnotation: $age :: Int at 6:32
    MethodReturnAnnotation: get_name :: Str at 13:12
    VarAnnotation: User :: class at 1:1
    VarAnnotation: $name :: Str at 2:5
    VarAnnotation: $age :: Int at 3:5
    VarAnnotation: $email :: Optional[Email] at 4:5
  Root: source_file
  Tree Structure:
  source_file
    class_decl
      field_decl
        variable
      field_decl
        variable
      field_decl
        variable
      method_decl
        type_expr
        block_stmt
          token
          expression_stmt
            literal
          token
          token
      method_decl
        type_expr
        block_stmt
          token
          expression_stmt
            literal
          token
          token
}
```

## JSON AST

```json
{
  "path": "/tmp/basic-class.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 17,
      "Column": 1,
      "Offset": 295
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
          "Line": 16,
          "Column": 2,
          "Offset": 294
        },
        "children": [
          {
            "type": "field_decl",
            "start": {
              "Line": 2,
              "Column": 5,
              "Offset": 0
            },
            "end": {
              "Line": 2,
              "Column": 20,
              "Offset": 0
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 2,
                  "Column": 5,
                  "Offset": 0
                },
                "end": {
                  "Line": 2,
                  "Column": 20,
                  "Offset": 0
                },
                "name": "name",
                "sigil": "$"
              }
            ],
            "name": "name"
          },
          {
            "type": "field_decl",
            "start": {
              "Line": 3,
              "Column": 5,
              "Offset": 0
            },
            "end": {
              "Line": 3,
              "Column": 19,
              "Offset": 0
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 3,
                  "Column": 5,
                  "Offset": 0
                },
                "end": {
                  "Line": 3,
                  "Column": 19,
                  "Offset": 0
                },
                "name": "age",
                "sigil": "$"
              }
            ],
            "name": "age"
          },
          {
            "type": "field_decl",
            "start": {
              "Line": 4,
              "Column": 5,
              "Offset": 0
            },
            "end": {
              "Line": 4,
              "Column": 33,
              "Offset": 0
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 4,
                  "Column": 5,
                  "Offset": 0
                },
                "end": {
                  "Line": 4,
                  "Column": 33,
                  "Offset": 0
                },
                "name": "email",
                "sigil": "$"
              }
            ],
            "name": "email"
          },
          {
            "type": "method_decl",
            "start": {
              "Line": 6,
              "Column": 5,
              "Offset": 0
            },
            "end": {
              "Line": 11,
              "Column": 6,
              "Offset": 0
            },
            "children": [
              {
                "type": "block_stmt",
                "start": {
                  "Line": 6,
                  "Column": 42,
                  "Offset": 130
                },
                "end": {
                  "Line": 11,
                  "Column": 6,
                  "Offset": 235
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 6,
                      "Column": 42,
                      "Offset": 130
                    },
                    "end": {
                      "Line": 6,
                      "Column": 43,
                      "Offset": 131
                    },
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 7,
                      "Column": 9,
                      "Offset": 140
                    },
                    "end": {
                      "Line": 10,
                      "Column": 23,
                      "Offset": 228
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 7,
                          "Column": 9,
                          "Offset": 140
                        },
                        "end": {
                          "Line": 10,
                          "Column": 23,
                          "Offset": 228
                        },
                        "value": "return bless {\n            name => $name,\n            age => $age\n        }, __PACKAGE__",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 10,
                      "Column": 23,
                      "Offset": 228
                    },
                    "end": {
                      "Line": 10,
                      "Column": 24,
                      "Offset": 229
                    },
                    "text": ";"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 11,
                      "Column": 5,
                      "Offset": 234
                    },
                    "end": {
                      "Line": 11,
                      "Column": 6,
                      "Offset": 235
                    },
                    "text": "}"
                  }
                ]
              }
            ],
            "name": "new"
          },
          {
            "type": "method_decl",
            "start": {
              "Line": 13,
              "Column": 5,
              "Offset": 0
            },
            "end": {
              "Line": 15,
              "Column": 6,
              "Offset": 0
            },
            "children": [
              {
                "type": "block_stmt",
                "start": {
                  "Line": 13,
                  "Column": 27,
                  "Offset": 263
                },
                "end": {
                  "Line": 15,
                  "Column": 6,
                  "Offset": 292
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 13,
                      "Column": 27,
                      "Offset": 263
                    },
                    "end": {
                      "Line": 13,
                      "Column": 28,
                      "Offset": 264
                    },
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 14,
                      "Column": 9,
                      "Offset": 273
                    },
                    "end": {
                      "Line": 14,
                      "Column": 21,
                      "Offset": 285
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 14,
                          "Column": 9,
                          "Offset": 273
                        },
                        "end": {
                          "Line": 14,
                          "Column": 21,
                          "Offset": 285
                        },
                        "value": "return $name",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 14,
                      "Column": 21,
                      "Offset": 285
                    },
                    "end": {
                      "Line": 14,
                      "Column": 22,
                      "Offset": 286
                    },
                    "text": ";"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 15,
                      "Column": 5,
                      "Offset": 291
                    },
                    "end": {
                      "Line": 15,
                      "Column": 6,
                      "Offset": 292
                    },
                    "text": "}"
                  }
                ]
              }
            ],
            "name": "get_name"
          }
        ],
        "name": "User"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "new",
      "type_expression": {
        "Kind": 0,
        "Name": "User",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "User"
      },
      "position": {
        "Line": 6,
        "Column": 12,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
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
        "Line": 6,
        "Column": 21,
        "Offset": 0
      },
      "kind": "MethodParamAnnotation"
    },
    {
      "annotated_item": "$age",
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
        "Line": 6,
        "Column": 32,
        "Offset": 0
      },
      "kind": "MethodParamAnnotation"
    },
    {
      "annotated_item": "get_name",
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
        "Line": 13,
        "Column": 12,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
    },
    {
      "annotated_item": "User",
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
        "Column": 5,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$age",
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
        "Line": 3,
        "Column": 5,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$email",
      "type_expression": {
        "Kind": 4,
        "Name": "Optional[Email]",
        "Parameters": [
          {
            "Kind": 0,
            "Name": "Email",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "Email"
          }
        ],
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Optional[Email]"
      },
      "position": {
        "Line": 4,
        "Column": 5,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "errors": [],
  "source_length": 295
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
class User {
    field $name;
    field $age;
    field $email;

    method new($name, $age) {
        return bless {
            name => $name,
            age => $age
        }, __PACKAGE__;
    }

    method get_name() {
        return $name;
    }
}
```

## Typed Perl Output

```perl
class User {
    field Str $name;
    field Int $age;
    field Optional[Email] $email;

    method User new(Str $name, Int $age) {
        return bless {
            name => $name,
            age => $age
        }, __PACKAGE__;
    }

    method Str get_name() {
        return $name;
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
