---
category: untyped-perl
subcategory: expressions
tags:
    - expressions
    - arithmetic
    - assignment
---

# Basic Expressions

Basic arithmetic and assignment expressions

```perl
my $a = 10;
my $b = 5;
my $sum = $a + $b;
my $product = $a * $b;
my $difference = $a - $b;
my $quotient = $a / $b;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $a = 10;
my $b = 5;
my $sum = $a + $b;
my $product = $a * $b;
my $difference = $a - $b;
my $quotient = $a / $b;
```

## Typed Perl Output

```perl
my $a = 10;
my $b = 5;
my $sum = $a + $b;
my $product = $a * $b;
my $difference = $a - $b;
my $quotient = $a / $b;
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
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
    var_decl
      variable
  token
  expression_statement
    var_decl
      variable
  token
  expression_statement
    var_decl
      variable
  token
  expression_statement
    var_decl
      variable
  token
```

## JSON AST

```json
{
  "path": "/tmp/basic-expressions.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 6,
      "Column": 24,
      "Offset": 114
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
          "Column": 11,
          "Offset": 10
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
              "Column": 11,
              "Offset": 10
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
                  "Column": 11,
                  "Offset": 10
                },
                "name": "a",
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
          "Line": 1,
          "Column": 11,
          "Offset": 10
        },
        "end": {
          "Line": 1,
          "Column": 12,
          "Offset": 11
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 12
        },
        "end": {
          "Line": 2,
          "Column": 10,
          "Offset": 21
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 12
            },
            "end": {
              "Line": 2,
              "Column": 10,
              "Offset": 21
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 12
                },
                "end": {
                  "Line": 2,
                  "Column": 10,
                  "Offset": 21
                },
                "name": "b",
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
          "Line": 2,
          "Column": 10,
          "Offset": 21
        },
        "end": {
          "Line": 2,
          "Column": 11,
          "Offset": 22
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 23
        },
        "end": {
          "Line": 3,
          "Column": 18,
          "Offset": 40
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 23
            },
            "end": {
              "Line": 3,
              "Column": 18,
              "Offset": 40
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 23
                },
                "end": {
                  "Line": 3,
                  "Column": 18,
                  "Offset": 40
                },
                "name": "sum",
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
          "Line": 3,
          "Column": 18,
          "Offset": 40
        },
        "end": {
          "Line": 3,
          "Column": 19,
          "Offset": 41
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 4,
          "Column": 1,
          "Offset": 42
        },
        "end": {
          "Line": 4,
          "Column": 22,
          "Offset": 63
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 4,
              "Column": 1,
              "Offset": 42
            },
            "end": {
              "Line": 4,
              "Column": 22,
              "Offset": 63
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 4,
                  "Column": 1,
                  "Offset": 42
                },
                "end": {
                  "Line": 4,
                  "Column": 22,
                  "Offset": 63
                },
                "name": "product",
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
          "Line": 4,
          "Column": 22,
          "Offset": 63
        },
        "end": {
          "Line": 4,
          "Column": 23,
          "Offset": 64
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 5,
          "Column": 1,
          "Offset": 65
        },
        "end": {
          "Line": 5,
          "Column": 25,
          "Offset": 89
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 5,
              "Column": 1,
              "Offset": 65
            },
            "end": {
              "Line": 5,
              "Column": 25,
              "Offset": 89
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 5,
                  "Column": 1,
                  "Offset": 65
                },
                "end": {
                  "Line": 5,
                  "Column": 25,
                  "Offset": 89
                },
                "name": "difference",
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
          "Line": 5,
          "Column": 25,
          "Offset": 89
        },
        "end": {
          "Line": 5,
          "Column": 26,
          "Offset": 90
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 6,
          "Column": 1,
          "Offset": 91
        },
        "end": {
          "Line": 6,
          "Column": 23,
          "Offset": 113
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 6,
              "Column": 1,
              "Offset": 91
            },
            "end": {
              "Line": 6,
              "Column": 23,
              "Offset": 113
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 6,
                  "Column": 1,
                  "Offset": 91
                },
                "end": {
                  "Line": 6,
                  "Column": 23,
                  "Offset": 113
                },
                "name": "quotient",
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
          "Column": 23,
          "Offset": 113
        },
        "end": {
          "Line": 6,
          "Column": 24,
          "Offset": 114
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 114
}
```
