---
category: untyped-perl
subcategory: subroutines
tags:
    - subroutines
    - functions
    - basic
---

# Simple Subroutine

Basic subroutine definition and call

```perl
sub greet {
    my $name = shift;
    return "Hello, $name!";
}

my $message = greet("World");
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
sub greet {
    my $name = shift;
    return "Hello, $name!";
}

my $message = greet("World");
```

## Typed Perl Output

```perl
sub greet {
    my $name = shift;
    return "Hello, $name!";
}

my $message = greet("World");
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
      expression_stmt
        literal
      token
      token
  expression_statement
    var_decl
      variable
  token
```

## JSON Format

```json
{
  "path": "/tmp/simple-subroutine.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 6,
      "Column": 30,
      "Offset": 94
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
          "Line": 4,
          "Column": 2,
          "Offset": 63
        },
        "children": [
          {
            "type": "block_stmt",
            "start": {
              "Line": 1,
              "Column": 11,
              "Offset": 0
            },
            "end": {
              "Line": 4,
              "Column": 2,
              "Offset": 0
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
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 2,
                  "Column": 5,
                  "Offset": 16
                },
                "end": {
                  "Line": 2,
                  "Column": 21,
                  "Offset": 32
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 2,
                      "Column": 5,
                      "Offset": 16
                    },
                    "end": {
                      "Line": 2,
                      "Column": 21,
                      "Offset": 32
                    },
                    "value": "my $name = shift",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 21,
                  "Offset": 32
                },
                "end": {
                  "Line": 2,
                  "Column": 22,
                  "Offset": 33
                },
                "text": ";"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 3,
                  "Column": 5,
                  "Offset": 38
                },
                "end": {
                  "Line": 3,
                  "Column": 27,
                  "Offset": 60
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 3,
                      "Column": 5,
                      "Offset": 38
                    },
                    "end": {
                      "Line": 3,
                      "Column": 27,
                      "Offset": 60
                    },
                    "value": "return \"Hello, $name!\"",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 27,
                  "Offset": 60
                },
                "end": {
                  "Line": 3,
                  "Column": 28,
                  "Offset": 61
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 4,
                  "Column": 1,
                  "Offset": 62
                },
                "end": {
                  "Line": 4,
                  "Column": 2,
                  "Offset": 63
                },
                "text": "}"
              }
            ]
          }
        ],
        "name": "greet"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 6,
          "Column": 1,
          "Offset": 65
        },
        "end": {
          "Line": 6,
          "Column": 29,
          "Offset": 93
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 6,
              "Column": 1,
              "Offset": 65
            },
            "end": {
              "Line": 6,
              "Column": 29,
              "Offset": 93
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 6,
                  "Column": 1,
                  "Offset": 65
                },
                "end": {
                  "Line": 6,
                  "Column": 29,
                  "Offset": 93
                },
                "name": "message",
                "sigil": "$"
              }
            ],
            "decl_type": "my"
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 6,
          "Column": 29,
          "Offset": 93
        },
        "end": {
          "Line": 6,
          "Column": 30,
          "Offset": 94
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 94
}
```
