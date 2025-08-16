---
category: typed-perl
subcategory: classes-roles
tags:
    - constructor
    - destructor
    - BUILD-method
    - DESTROY-method
    - lifecycle-management
type_check: true
---

# Constructor Destructor Methods

Class with constructor, destructor, and lifecycle methods

```perl
class Resource {
    field Str $name;
    field FileHandle $handle;
    field Bool $is_open = 0;

    method Void BUILD(Str $name, Optional[Str] $mode = 'r') {
        $self->{name} = $name;
        $self->{handle} = IO::File->new($name, $mode);
        $self->{is_open} = defined $self->{handle};
    }

    method Resource new(Str $name, Optional[Str] $mode = 'r') {
        my $self = bless {}, __PACKAGE__;
        $self->BUILD($name, $mode);
        return $self;
    }

    method Void DESTROY() {
        $self->close() if $is_open;
    }

    method Bool close() {
        return 0 unless $is_open;
        my $result = $handle->close();
        $is_open = 0;
        return $result;
    }

    method Optional[Str] read(Int $bytes) {
        return undef unless $is_open;
        my $data;
        my $read_bytes = $handle->read($data, $bytes);
        return defined $read_bytes ? $data : undef;
    }
}
```

# Expected AST

## Before Type Inference

```
AST {
  Path:
  Source length: 913 characters
  Type Annotations:
    MethodReturnAnnotation: BUILD :: Void at 6:12
    MethodParamAnnotation: $name :: Str at 6:23
    MethodParamAnnotation: $mode :: Optional[Str] at 6:34
    MethodReturnAnnotation: new :: Resource at 12:12
    MethodReturnAnnotation: DESTROY :: Void at 18:12
    MethodReturnAnnotation: close :: Bool at 22:12
    MethodReturnAnnotation: read :: Optional[Str] at 29:12
    MethodParamAnnotation: $bytes :: Int at 29:31
    VarAnnotation: Resource :: class at 1:1
    VarAnnotation: $name :: Str at 2:5
    VarAnnotation: $handle :: FileHandle at 3:5
    VarAnnotation: $is_open :: Bool at 4:5
  Root: source_file
  Tree Structure:
  source_file
    class_decl
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
          expression_stmt
            literal
          token
          expression_stmt
            literal
          token
          token
      method_decl
        type_expr
        block_stmt
          token
          var_decl
            variable
          token
          expression_stmt
            literal
          token
          return_stmt
            variable
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
      method_decl
        type_expr
        block_stmt
          token
          expression_stmt
            literal
          token
          var_decl
            variable
            literal
          token
          expression_stmt
            literal
          token
          return_stmt
            variable
          token
          token
      method_decl
        type_expr
        block_stmt
          token
          expression_stmt
            literal
          token
          var_decl
            variable
            variable
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
  Source length: 913 characters
  Type Annotations:
    MethodReturnAnnotation: BUILD :: Void at 6:12
    MethodParamAnnotation: $name :: Str at 6:23
    MethodParamAnnotation: $mode :: Optional[Str] at 6:34
    MethodReturnAnnotation: new :: Resource at 12:12
    MethodReturnAnnotation: DESTROY :: Void at 18:12
    MethodReturnAnnotation: close :: Bool at 22:12
    MethodReturnAnnotation: read :: Optional[Str] at 29:12
    MethodParamAnnotation: $bytes :: Int at 29:31
    VarAnnotation: Resource :: class at 1:1
    VarAnnotation: $name :: Str at 2:5
    VarAnnotation: $handle :: FileHandle at 3:5
    VarAnnotation: $is_open :: Bool at 4:5
  Root: source_file
  Tree Structure:
  source_file
    class_decl
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
          expression_stmt
            literal
          token
          expression_stmt
            literal
          token
          token
      method_decl
        type_expr
        block_stmt
          token
          var_decl
            variable
          token
          expression_stmt
            literal
          token
          return_stmt
            variable
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
      method_decl
        type_expr
        block_stmt
          token
          expression_stmt
            literal
          token
          var_decl
            variable
            literal
          token
          expression_stmt
            literal
          token
          return_stmt
            variable
          token
          token
      method_decl
        type_expr
        block_stmt
          token
          expression_stmt
            literal
          token
          var_decl
            variable
            variable
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
  "path": "/tmp/constructor-destructor.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 36,
      "Column": 1,
      "Offset": 914
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
          "Line": 35,
          "Column": 2,
          "Offset": 913
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
              "Column": 29,
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
                  "Column": 29,
                  "Offset": 0
                },
                "name": "handle",
                "sigil": "$"
              }
            ],
            "name": "handle"
          },
          {
            "type": "method_decl",
            "start": {
              "Line": 6,
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
                  "Line": 6,
                  "Column": 61,
                  "Offset": 158
                },
                "end": {
                  "Line": 10,
                  "Column": 6,
                  "Offset": 303
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 6,
                      "Column": 61,
                      "Offset": 158
                    },
                    "end": {
                      "Line": 6,
                      "Column": 62,
                      "Offset": 159
                    },
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 7,
                      "Column": 9,
                      "Offset": 168
                    },
                    "end": {
                      "Line": 7,
                      "Column": 30,
                      "Offset": 189
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 7,
                          "Column": 9,
                          "Offset": 168
                        },
                        "end": {
                          "Line": 7,
                          "Column": 30,
                          "Offset": 189
                        },
                        "value": "$self->{name} = $name",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 7,
                      "Column": 30,
                      "Offset": 189
                    },
                    "end": {
                      "Line": 7,
                      "Column": 31,
                      "Offset": 190
                    },
                    "text": ";"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 8,
                      "Column": 9,
                      "Offset": 199
                    },
                    "end": {
                      "Line": 8,
                      "Column": 54,
                      "Offset": 244
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 8,
                          "Column": 9,
                          "Offset": 199
                        },
                        "end": {
                          "Line": 8,
                          "Column": 54,
                          "Offset": 244
                        },
                        "value": "$self->{handle} = IO::File->new($name, $mode)",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 8,
                      "Column": 54,
                      "Offset": 244
                    },
                    "end": {
                      "Line": 8,
                      "Column": 55,
                      "Offset": 245
                    },
                    "text": ";"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 9,
                      "Column": 9,
                      "Offset": 254
                    },
                    "end": {
                      "Line": 9,
                      "Column": 51,
                      "Offset": 296
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 9,
                          "Column": 9,
                          "Offset": 254
                        },
                        "end": {
                          "Line": 9,
                          "Column": 51,
                          "Offset": 296
                        },
                        "value": "$self->{is_open} = defined $self->{handle}",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 9,
                      "Column": 51,
                      "Offset": 296
                    },
                    "end": {
                      "Line": 9,
                      "Column": 52,
                      "Offset": 297
                    },
                    "text": ";"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 10,
                      "Column": 5,
                      "Offset": 302
                    },
                    "end": {
                      "Line": 10,
                      "Column": 6,
                      "Offset": 303
                    },
                    "text": "}"
                  }
                ]
              }
            ],
            "name": "BUILD"
          },
          {
            "type": "method_decl",
            "start": {
              "Line": 12,
              "Column": 5,
              "Offset": 0
            },
            "end": {
              "Line": 16,
              "Column": 6,
              "Offset": 0
            },
            "children": [
              {
                "type": "block_stmt",
                "start": {
                  "Line": 12,
                  "Column": 63,
                  "Offset": 367
                },
                "end": {
                  "Line": 16,
                  "Column": 6,
                  "Offset": 474
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 12,
                      "Column": 63,
                      "Offset": 367
                    },
                    "end": {
                      "Line": 12,
                      "Column": 64,
                      "Offset": 368
                    },
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 13,
                      "Column": 9,
                      "Offset": 377
                    },
                    "end": {
                      "Line": 13,
                      "Column": 41,
                      "Offset": 409
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 13,
                          "Column": 9,
                          "Offset": 377
                        },
                        "end": {
                          "Line": 13,
                          "Column": 41,
                          "Offset": 409
                        },
                        "value": "my $self = bless {}, __PACKAGE__",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 13,
                      "Column": 41,
                      "Offset": 409
                    },
                    "end": {
                      "Line": 13,
                      "Column": 42,
                      "Offset": 410
                    },
                    "text": ";"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 14,
                      "Column": 9,
                      "Offset": 419
                    },
                    "end": {
                      "Line": 14,
                      "Column": 35,
                      "Offset": 445
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 14,
                          "Column": 9,
                          "Offset": 419
                        },
                        "end": {
                          "Line": 14,
                          "Column": 35,
                          "Offset": 445
                        },
                        "value": "$self->BUILD($name, $mode)",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 14,
                      "Column": 35,
                      "Offset": 445
                    },
                    "end": {
                      "Line": 14,
                      "Column": 36,
                      "Offset": 446
                    },
                    "text": ";"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 15,
                      "Column": 9,
                      "Offset": 455
                    },
                    "end": {
                      "Line": 15,
                      "Column": 21,
                      "Offset": 467
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 15,
                          "Column": 9,
                          "Offset": 455
                        },
                        "end": {
                          "Line": 15,
                          "Column": 21,
                          "Offset": 467
                        },
                        "value": "return $self",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 15,
                      "Column": 21,
                      "Offset": 467
                    },
                    "end": {
                      "Line": 15,
                      "Column": 22,
                      "Offset": 468
                    },
                    "text": ";"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 16,
                      "Column": 5,
                      "Offset": 473
                    },
                    "end": {
                      "Line": 16,
                      "Column": 6,
                      "Offset": 474
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
              "Line": 18,
              "Column": 5,
              "Offset": 0
            },
            "end": {
              "Line": 20,
              "Column": 6,
              "Offset": 0
            },
            "children": [
              {
                "type": "block_stmt",
                "start": {
                  "Line": 18,
                  "Column": 27,
                  "Offset": 502
                },
                "end": {
                  "Line": 20,
                  "Column": 6,
                  "Offset": 545
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 18,
                      "Column": 27,
                      "Offset": 502
                    },
                    "end": {
                      "Line": 18,
                      "Column": 28,
                      "Offset": 503
                    },
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 19,
                      "Column": 9,
                      "Offset": 512
                    },
                    "end": {
                      "Line": 19,
                      "Column": 35,
                      "Offset": 538
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 19,
                          "Column": 9,
                          "Offset": 512
                        },
                        "end": {
                          "Line": 19,
                          "Column": 35,
                          "Offset": 538
                        },
                        "value": "$self->close() if $is_open",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 19,
                      "Column": 35,
                      "Offset": 538
                    },
                    "end": {
                      "Line": 19,
                      "Column": 36,
                      "Offset": 539
                    },
                    "text": ";"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 20,
                      "Column": 5,
                      "Offset": 544
                    },
                    "end": {
                      "Line": 20,
                      "Column": 6,
                      "Offset": 545
                    },
                    "text": "}"
                  }
                ]
              }
            ],
            "name": "DESTROY"
          },
          {
            "type": "method_decl",
            "start": {
              "Line": 22,
              "Column": 5,
              "Offset": 0
            },
            "end": {
              "Line": 27,
              "Column": 6,
              "Offset": 0
            },
            "children": [
              {
                "type": "block_stmt",
                "start": {
                  "Line": 22,
                  "Column": 25,
                  "Offset": 571
                },
                "end": {
                  "Line": 27,
                  "Column": 6,
                  "Offset": 697
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 22,
                      "Column": 25,
                      "Offset": 571
                    },
                    "end": {
                      "Line": 22,
                      "Column": 26,
                      "Offset": 572
                    },
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 23,
                      "Column": 9,
                      "Offset": 581
                    },
                    "end": {
                      "Line": 23,
                      "Column": 33,
                      "Offset": 605
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 23,
                          "Column": 9,
                          "Offset": 581
                        },
                        "end": {
                          "Line": 23,
                          "Column": 33,
                          "Offset": 605
                        },
                        "value": "return 0 unless $is_open",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 23,
                      "Column": 33,
                      "Offset": 605
                    },
                    "end": {
                      "Line": 23,
                      "Column": 34,
                      "Offset": 606
                    },
                    "text": ";"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 24,
                      "Column": 9,
                      "Offset": 615
                    },
                    "end": {
                      "Line": 24,
                      "Column": 38,
                      "Offset": 644
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 24,
                          "Column": 9,
                          "Offset": 615
                        },
                        "end": {
                          "Line": 24,
                          "Column": 38,
                          "Offset": 644
                        },
                        "value": "my $result = $handle->close()",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 24,
                      "Column": 38,
                      "Offset": 644
                    },
                    "end": {
                      "Line": 24,
                      "Column": 39,
                      "Offset": 645
                    },
                    "text": ";"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 25,
                      "Column": 9,
                      "Offset": 654
                    },
                    "end": {
                      "Line": 25,
                      "Column": 21,
                      "Offset": 666
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 25,
                          "Column": 9,
                          "Offset": 654
                        },
                        "end": {
                          "Line": 25,
                          "Column": 21,
                          "Offset": 666
                        },
                        "value": "$is_open = 0",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 25,
                      "Column": 21,
                      "Offset": 666
                    },
                    "end": {
                      "Line": 25,
                      "Column": 22,
                      "Offset": 667
                    },
                    "text": ";"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 26,
                      "Column": 9,
                      "Offset": 676
                    },
                    "end": {
                      "Line": 26,
                      "Column": 23,
                      "Offset": 690
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 26,
                          "Column": 9,
                          "Offset": 676
                        },
                        "end": {
                          "Line": 26,
                          "Column": 23,
                          "Offset": 690
                        },
                        "value": "return $result",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 26,
                      "Column": 23,
                      "Offset": 690
                    },
                    "end": {
                      "Line": 26,
                      "Column": 24,
                      "Offset": 691
                    },
                    "text": ";"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 27,
                      "Column": 5,
                      "Offset": 696
                    },
                    "end": {
                      "Line": 27,
                      "Column": 6,
                      "Offset": 697
                    },
                    "text": "}"
                  }
                ]
              }
            ],
            "name": "close"
          },
          {
            "type": "method_decl",
            "start": {
              "Line": 29,
              "Column": 5,
              "Offset": 0
            },
            "end": {
              "Line": 34,
              "Column": 6,
              "Offset": 0
            },
            "children": [
              {
                "type": "block_stmt",
                "start": {
                  "Line": 29,
                  "Column": 43,
                  "Offset": 741
                },
                "end": {
                  "Line": 34,
                  "Column": 6,
                  "Offset": 911
                },
                "children": [
                  {
                    "type": "token",
                    "start": {
                      "Line": 29,
                      "Column": 43,
                      "Offset": 741
                    },
                    "end": {
                      "Line": 29,
                      "Column": 44,
                      "Offset": 742
                    },
                    "text": "{"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 30,
                      "Column": 9,
                      "Offset": 751
                    },
                    "end": {
                      "Line": 30,
                      "Column": 37,
                      "Offset": 779
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 30,
                          "Column": 9,
                          "Offset": 751
                        },
                        "end": {
                          "Line": 30,
                          "Column": 37,
                          "Offset": 779
                        },
                        "value": "return undef unless $is_open",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 30,
                      "Column": 37,
                      "Offset": 779
                    },
                    "end": {
                      "Line": 30,
                      "Column": 38,
                      "Offset": 780
                    },
                    "text": ";"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 31,
                      "Column": 9,
                      "Offset": 789
                    },
                    "end": {
                      "Line": 31,
                      "Column": 17,
                      "Offset": 797
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 31,
                          "Column": 9,
                          "Offset": 789
                        },
                        "end": {
                          "Line": 31,
                          "Column": 17,
                          "Offset": 797
                        },
                        "value": "my $data",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 31,
                      "Column": 17,
                      "Offset": 797
                    },
                    "end": {
                      "Line": 31,
                      "Column": 18,
                      "Offset": 798
                    },
                    "text": ";"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 32,
                      "Column": 9,
                      "Offset": 807
                    },
                    "end": {
                      "Line": 32,
                      "Column": 54,
                      "Offset": 852
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 32,
                          "Column": 9,
                          "Offset": 807
                        },
                        "end": {
                          "Line": 32,
                          "Column": 54,
                          "Offset": 852
                        },
                        "value": "my $read_bytes = $handle->read($data, $bytes)",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 32,
                      "Column": 54,
                      "Offset": 852
                    },
                    "end": {
                      "Line": 32,
                      "Column": 55,
                      "Offset": 853
                    },
                    "text": ";"
                  },
                  {
                    "type": "expression_stmt",
                    "start": {
                      "Line": 33,
                      "Column": 9,
                      "Offset": 862
                    },
                    "end": {
                      "Line": 33,
                      "Column": 57,
                      "Offset": 910
                    },
                    "children": [
                      {
                        "type": "literal",
                        "start": {
                          "Line": 33,
                          "Column": 9,
                          "Offset": 862
                        },
                        "end": {
                          "Line": 33,
                          "Column": 57,
                          "Offset": 910
                        },
                        "value": "return defined $read_bytes ? $data : undef",
                        "kind": "string"
                      }
                    ]
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 33,
                      "Column": 57,
                      "Offset": 910
                    },
                    "end": {
                      "Line": 33,
                      "Column": 58,
                      "Offset": 911
                    },
                    "text": ";"
                  },
                  {
                    "type": "token",
                    "start": {
                      "Line": 34,
                      "Column": 5,
                      "Offset": 912
                    },
                    "end": {
                      "Line": 34,
                      "Column": 6,
                      "Offset": 913
                    },
                    "text": "}"
                  }
                ]
              }
            ],
            "name": "read"
          }
        ],
        "name": "Resource"
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "BUILD",
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
        "Column": 23,
        "Offset": 0
      },
      "kind": "MethodParamAnnotation"
    },
    {
      "annotated_item": "$mode",
      "type_expression": {
        "Kind": 4,
        "Name": "Optional[Str]",
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
        "OriginalString": "Optional[Str]"
      },
      "position": {
        "Line": 6,
        "Column": 34,
        "Offset": 0
      },
      "kind": "MethodParamAnnotation"
    },
    {
      "annotated_item": "new",
      "type_expression": {
        "Kind": 0,
        "Name": "Resource",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "Resource"
      },
      "position": {
        "Line": 12,
        "Column": 12,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
    },
    {
      "annotated_item": "DESTROY",
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
        "Line": 18,
        "Column": 12,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
    },
    {
      "annotated_item": "close",
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
        "Line": 22,
        "Column": 12,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
    },
    {
      "annotated_item": "read",
      "type_expression": {
        "Kind": 4,
        "Name": "Optional[Str]",
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
        "OriginalString": "Optional[Str]"
      },
      "position": {
        "Line": 29,
        "Column": 12,
        "Offset": 0
      },
      "kind": "MethodReturnAnnotation"
    },
    {
      "annotated_item": "$bytes",
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
        "Line": 29,
        "Column": 31,
        "Offset": 0
      },
      "kind": "MethodParamAnnotation"
    },
    {
      "annotated_item": "Resource",
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
      "annotated_item": "$handle",
      "type_expression": {
        "Kind": 0,
        "Name": "FileHandle",
        "Parameters": null,
        "IsUnion": false,
        "IsIntersection": false,
        "IsNegation": false,
        "UnionTypes": null,
        "IntersectionTypes": null,
        "NegatedType": null,
        "Constraint": null,
        "OriginalString": "FileHandle"
      },
      "position": {
        "Line": 3,
        "Column": 5,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$is_open",
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
        "Line": 4,
        "Column": 5,
        "Offset": 0
      },
      "kind": "VarAnnotation"
    }
  ],
  "errors": [],
  "source_length": 914
}
```


# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
class Resource {
    field $name;
    field $handle;
    field $is_open = 0;

    method BUILD($name, $mode = 'r') {
        $self->{name} = $name;
        $self->{handle} = IO::File->new($name, $mode);
        $self->{is_open} = defined $self->{handle};
    }

    method new($name, $mode = 'r') {
        my $self = bless {}, __PACKAGE__;
        $self->BUILD($name, $mode);
        return $self;
    }

    method DESTROY() {
        $self->close() if $is_open;
    }

    method close() {
        return 0 unless $is_open;
        my $result = $handle->close();
        $is_open = 0;
        return $result;
    }

    method read($bytes) {
        return undef unless $is_open;
        my $data;
        my $read_bytes = $handle->read($data, $bytes);
        return defined $read_bytes ? $data : undef;
    }
}
```

## Typed Perl Output

```perl
class Resource {
    field Str $name;
    field FileHandle $handle;
    field Bool $is_open = 0;

    method Void BUILD(Str $name, Optional[Str] $mode = 'r') {
        $self->{name} = $name;
        $self->{handle} = IO::File->new($name, $mode);
        $self->{is_open} = defined $self->{handle};
    }

    method Resource new(Str $name, Optional[Str] $mode = 'r') {
        my $self = bless {}, __PACKAGE__;
        $self->BUILD($name, $mode);
        return $self;
    }

    method Void DESTROY() {
        $self->close() if $is_open;
    }

    method Bool close() {
        return 0 unless $is_open;
        my $result = $handle->close();
        $is_open = 0;
        return $result;
    }

    method Optional[Str] read(Int $bytes) {
        return undef unless $is_open;
        my $data;
        my $read_bytes = $handle->read($data, $bytes);
        return defined $read_bytes ? $data : undef;
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
