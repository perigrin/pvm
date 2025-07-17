---
category: untyped-perl
subcategory: subroutines
tags:
    - basic
    - definitions
    - parameters
    - qualified
---

# Basic Subroutine Definitions

Test basic subroutine definitions and simple subroutines

```perl
sub simple_sub {
    return "result";
}

sub with_params {
    my ($first, $second) = @_;
    return $first + $second;
}

sub Package::qualified_sub {
    return 42;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
sub simple_sub {
    return "result";
}

sub with_params {
    my ($first, $second) = @_;
    return $first + $second;
}

sub Package::qualified_sub {
    return 42;
}
```

## Typed Perl Output

```perl
sub simple_sub {
    return "result";
}

sub with_params {
    my ($first, $second) = @_;
    return $first + $second;
}

sub Package::qualified_sub {
    return 42;
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text Format

```
source_file
  sub_decl
    block_stmt
      token
      expression_stmt
        literal
      token
      token
  sub_decl
    block_stmt
      token
      expression_stmt
        literal
      token
      expression_stmt
        literal
      token
      token
  sub_decl
    block_stmt
      token
      expression_stmt
        literal
      token
      token
```

## JSON Format

```json
{
  "path": "/tmp/basic_subroutine_definitions.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 12,
      "Column": 2,
      "Offset": 167
    },
    "children": [
      {
        "type": "sub_decl",
        "start": {
          "Line": 1,
          "Column": 1,
          "Offset": 0
        },
        "end": {
          "Line": 3,
          "Column": 2,
          "Offset": 39
        },
        "children": [
          {
            "type": "block_stmt",
            "start": {
              "Line": 1,
              "Column": 16,
              "Offset": 0
            },
            "end": {
              "Line": 3,
              "Column": 2,
              "Offset": 0
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 16,
                  "Offset": 15
                },
                "end": {
                  "Line": 1,
                  "Column": 17,
                  "Offset": 16
                },
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 2,
                  "Column": 5,
                  "Offset": 21
                },
                "end": {
                  "Line": 2,
                  "Column": 20,
                  "Offset": 36
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 2,
                      "Column": 5,
                      "Offset": 21
                    },
                    "end": {
                      "Line": 2,
                      "Column": 20,
                      "Offset": 36
                    },
                    "value": "return \"result\"",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 20,
                  "Offset": 36
                },
                "end": {
                  "Line": 2,
                  "Column": 21,
                  "Offset": 37
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 38
                },
                "end": {
                  "Line": 3,
                  "Column": 2,
                  "Offset": 39
                },
                "text": "}"
              }
            ]
          }
        ],
        "name": "simple_sub"
      },
      {
        "type": "sub_decl",
        "start": {
          "Line": 5,
          "Column": 1,
          "Offset": 41
        },
        "end": {
          "Line": 8,
          "Column": 2,
          "Offset": 120
        },
        "children": [
          {
            "type": "block_stmt",
            "start": {
              "Line": 5,
              "Column": 17,
              "Offset": 0
            },
            "end": {
              "Line": 8,
              "Column": 2,
              "Offset": 0
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 5,
                  "Column": 17,
                  "Offset": 57
                },
                "end": {
                  "Line": 5,
                  "Column": 18,
                  "Offset": 58
                },
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 6,
                  "Column": 5,
                  "Offset": 63
                },
                "end": {
                  "Line": 6,
                  "Column": 30,
                  "Offset": 88
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 6,
                      "Column": 5,
                      "Offset": 63
                    },
                    "end": {
                      "Line": 6,
                      "Column": 30,
                      "Offset": 88
                    },
                    "value": "my ($first, $second) = @_",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 6,
                  "Column": 30,
                  "Offset": 88
                },
                "end": {
                  "Line": 6,
                  "Column": 31,
                  "Offset": 89
                },
                "text": ";"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 7,
                  "Column": 5,
                  "Offset": 94
                },
                "end": {
                  "Line": 7,
                  "Column": 28,
                  "Offset": 117
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 7,
                      "Column": 5,
                      "Offset": 94
                    },
                    "end": {
                      "Line": 7,
                      "Column": 28,
                      "Offset": 117
                    },
                    "value": "return $first + $second",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 7,
                  "Column": 28,
                  "Offset": 117
                },
                "end": {
                  "Line": 7,
                  "Column": 29,
                  "Offset": 118
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 8,
                  "Column": 1,
                  "Offset": 119
                },
                "end": {
                  "Line": 8,
                  "Column": 2,
                  "Offset": 120
                },
                "text": "}"
              }
            ]
          }
        ],
        "name": "with_params"
      },
      {
        "type": "sub_decl",
        "start": {
          "Line": 10,
          "Column": 1,
          "Offset": 122
        },
        "end": {
          "Line": 12,
          "Column": 2,
          "Offset": 167
        },
        "children": [
          {
            "type": "block_stmt",
            "start": {
              "Line": 10,
              "Column": 28,
              "Offset": 0
            },
            "end": {
              "Line": 12,
              "Column": 2,
              "Offset": 0
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 10,
                  "Column": 28,
                  "Offset": 149
                },
                "end": {
                  "Line": 10,
                  "Column": 29,
                  "Offset": 150
                },
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 11,
                  "Column": 5,
                  "Offset": 155
                },
                "end": {
                  "Line": 11,
                  "Column": 14,
                  "Offset": 164
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 11,
                      "Column": 5,
                      "Offset": 155
                    },
                    "end": {
                      "Line": 11,
                      "Column": 14,
                      "Offset": 164
                    },
                    "value": "return 42",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 11,
                  "Column": 14,
                  "Offset": 164
                },
                "end": {
                  "Line": 11,
                  "Column": 15,
                  "Offset": 165
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 12,
                  "Column": 1,
                  "Offset": 166
                },
                "end": {
                  "Line": 12,
                  "Column": 2,
                  "Offset": 167
                },
                "text": "}"
              }
            ]
          }
        ],
        "name": "Package::qualified_sub"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 167
}
```
