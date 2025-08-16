---
category: typed-perl
subcategory: classes-roles
tags:
    - inheritance
    - role-composition
    - class-declaration
type_check: true
---

# Class Inheritance

Class with inheritance and role composition

```perl
class Document : BaseDocument does Serializable, Cacheable {
    field Str $content;
    field DateTime $created;
    field Optional[UserRef] $author;

    method Str serialize() {
        return encode_json({
            content => $content,
            created => $created->iso8601,
            author => $author ? $author->id : undef
        });
    }

    method Self deserialize(Str $data) {
        my $decoded = decode_json($data);
        return __PACKAGE__->new(
            content => $decoded->{content},
            created => DateTime->from_epoch(epoch => $decoded->{created}),
            author => $decoded->{author} ? UserRef->new(id => $decoded->{author}) : undef
        );
    }
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 699 characters
  Type Annotations:
    MethodReturnAnnotation: serialize :: Str at 6:12
    MethodReturnAnnotation: deserialize :: Self at 14:12
    MethodParamAnnotation: $data :: Str at 14:29
    VarAnnotation: Document :: class at 1:1
    VarAnnotation: $content :: Str at 2:5
    VarAnnotation: $created :: DateTime at 3:5
    VarAnnotation: $author :: Optional[UserRef] at 4:5
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
          return_stmt
            literal
          token
          token
      method_decl
        type_expr
        block_stmt
          token
          var_decl
            variable
            literal
          token
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
  Source length: 699 characters
  Type Annotations:
    MethodReturnAnnotation: serialize :: Str at 6:12
    MethodReturnAnnotation: deserialize :: Self at 14:12
    MethodParamAnnotation: $data :: Str at 14:29
    VarAnnotation: Document :: class at 1:1
    VarAnnotation: $content :: Str at 2:5
    VarAnnotation: $created :: DateTime at 3:5
    VarAnnotation: $author :: Optional[UserRef] at 4:5
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
          return_stmt
            literal
          token
          token
      method_decl
        type_expr
        block_stmt
          token
          var_decl
            variable
            literal
          token
          return_stmt
            literal
          token
          token
}
```

## JSON AST

```json
{
  "path": "/tmp/class-inheritance.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 23,
      "Column": 1,
      "Offset": 700
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
          "Line": 22,
          "Column": 2,
          "Offset": 699
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
              "Column": 23,
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
                  "Column": 23,
                  "Offset": 0
                },
                "name": "content",
                "sigil": "$"
              }
            ],
            "name": "content"
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
              "Column": 28,
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
                  "Column": 28,
                  "Offset": 0
                },
                "name": "created",
                "sigil": "$"
              }
            ],
            "name": "created"
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
              "Column": 36,
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
                  "Column": 36,
                  "Offset": 0
                },
                "name": "author",
                "sigil": "$"
              }
            ],
            "name": "author"
          },
          {
            "type": "method_decl",
            "start": {
              "Line": 6,
              "Column": 5,
              "Offset": 0
            },
            "end": {
              "Line": 12,
              "Column": 6,
              "Offset": 0
            },
            "children": [
              {
                "type": "block_stmt",
                "start": {
                  "Line": 6,
                  "Column": 28,
                  "Offset": 179
                },
                "end": {
                  "Line": 12,
                  "Column": 6,
                  "Offset": 354
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 6,
                      "Column": 28,
                      "Offset": 179
                    },
                    "end": {
                      "Line": 6,
                      "Column": 29,
                      "Offset": 180
                    },
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 7,
                      "Column": 9,
                      "Offset": 189
                    },
                    "end": {
                      "Line": 11,
                      "Column": 11,
                      "Offset": 347
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 7,
                          "Column": 9,
                          "Offset": 189
                        },
                        "end": {
                          "Line": 11,
                          "Column": 11,
                          "Offset": 347
                        },
                        "value": "return encode_json({\n            content => $content,\n            created => $created->iso8601,\n            author => $author ? $author->id : undef\n        })",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 11,
                      "Column": 11,
                      "Offset": 347
                    },
                    "end": {
                      "Line": 11,
                      "Column": 12,
                      "Offset": 348
                    },
                    "text": ";"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 12,
                      "Column": 5,
                      "Offset": 353
                    },
                    "end": {
                      "Line": 12,
                      "Column": 6,
                      "Offset": 354
                    },
                    "text": "}"
                  }
                ]
              }
            ],
            "name": "serialize"
          },
          {
            "type": "method_decl",
            "start": {
              "Line": 14,
              "Column": 5,
              "Offset": 0
            },
            "end": {
              "Line": 21,
              "Column": 6,
              "Offset": 0
            },
            "children": [
              {
                "type": "block_stmt",
                "start": {
                  "Line": 14,
                  "Column": 40,
                  "Offset": 395
                },
                "end": {
                  "Line": 21,
                  "Column": 6,
                  "Offset": 697
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 14,
                      "Column": 40,
                      "Offset": 395
                    },
                    "end": {
                      "Line": 14,
                      "Column": 41,
                      "Offset": 396
                    },
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 15,
                      "Column": 9,
                      "Offset": 405
                    },
                    "end": {
                      "Line": 15,
                      "Column": 41,
                      "Offset": 437
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 15,
                          "Column": 9,
                          "Offset": 405
                        },
                        "end": {
                          "Line": 15,
                          "Column": 41,
                          "Offset": 437
                        },
                        "value": "my $decoded = decode_json($data)",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 15,
                      "Column": 41,
                      "Offset": 437
                    },
                    "end": {
                      "Line": 15,
                      "Column": 42,
                      "Offset": 438
                    },
                    "text": ";"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 16,
                      "Column": 9,
                      "Offset": 447
                    },
                    "end": {
                      "Line": 20,
                      "Column": 10,
                      "Offset": 690
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 16,
                          "Column": 9,
                          "Offset": 447
                        },
                        "end": {
                          "Line": 20,
                          "Column": 10,
                          "Offset": 690
                        },
                        "value": "return __PACKAGE__->new(\n            content => $decoded->{content},\n            created => DateTime->from_epoch(epoch => $decoded->{created}),\n            author => $decoded->{author} ? UserRef->new(id => $decoded->{author}) : undef\n        )",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 20,
                      "Column": 10,
                      "Offset": 690
                    },
                    "end": {
                      "Line": 20,
                      "Column": 11,
                      "Offset": 691
                    },
                    "text": ";"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 21,
                      "Column": 5,
                      "Offset": 696
                    },
                    "end": {
                      "Line": 21,
                      "Column": 6,
                      "Offset": 697
                    },
                    "text": "}"
                  }
                ]
              }
            ],
            "name": "deserialize"
          }
        ],
        "name": "Document"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "serialize",
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
        "Column": 12,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
    },
    {
      "annotated_item": "deserialize",
      "type_expression": {
        "Kind": 0,
        "Name": "Self",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Self"
      },
      "position": {
        "Line": 14,
        "Column": 12,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
    },
    {
      "annotated_item": "$data",
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
        "Line": 14,
        "Column": 29,
        "Offset": 0
      },
      "kind": "MethodParamAnnotation"
    },
    {
      "annotated_item": "Document",
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
      "annotated_item": "$content",
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
      "annotated_item": "$created",
      "type_expression": {
        "Kind": 0,
        "Name": "DateTime",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "DateTime"
      },
      "position": {
        "Line": 3,
        "Column": 5,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$author",
      "type_expression": {
        "Kind": 4,
        "Name": "Optional[UserRef]",
        "Parameters": [
          {
            "Kind": 0,
            "Name": "UserRef",
            "Parameters": null,
            "IsUnion": false,
            "IsIntersection": false,
            "IsNegation": false,
            "UnionTypes": null,
            "IntersectionTypes": null,
            "NegatedType": null,
            "Constraint": null,
            "OriginalString": "UserRef"
          }
        ],
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Optional[UserRef]"
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
  "source_length": 700
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
class Document : BaseDocument does Serializable, Cacheable {
    field $content;
    field $created;
    field $author;

    method serialize() {
        return encode_json({
            content => $content,
            created => $created->iso8601,
            author => $author ? $author->id : undef
        });
    }

    method deserialize($data) {
        my $decoded = decode_json($data);
        return __PACKAGE__->new(
            content => $decoded->{content},
            created => DateTime->from_epoch(epoch => $decoded->{created}),
            author => $decoded->{author} ? UserRef->new(id => $decoded->{author}) : undef
        );
    }
}
```

## Typed Perl Output

```perl
class Document : BaseDocument does Serializable, Cacheable {
    field Str $content;
    field DateTime $created;
    field Optional[UserRef] $author;

    method Str serialize() {
        return encode_json({
            content => $content,
            created => $created->iso8601,
            author => $author ? $author->id : undef
        });
    }

    method Self deserialize(Str $data) {
        my $decoded = decode_json($data);
        return __PACKAGE__->new(
            content => $decoded->{content},
            created => DateTime->from_epoch(epoch => $decoded->{created}),
            author => $decoded->{author} ? UserRef->new(id => $decoded->{author}) : undef
        );
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
