---
category: untyped-perl
subcategory: variables
tags:
    - context
    - expressions
    - scoping
    - variables
---

# Variable Context Variations

Test variable declarations in different contexts (statement, expression)

```perl
if (my $scoped = get_value()) { print $scoped; }
for my $iterator (1..10) { print $iterator; }
while (my $line = <DATA>) { chomp $line; }
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
if (my $scoped = get_value()) { print $scoped; }
for my $iterator (1..10) { print $iterator; }
while (my $line = <DATA>) { chomp $line; }
```

## Typed Perl Output

```perl
if (my $scoped = get_value()) { print $scoped; }
for my $iterator (1..10) { print $iterator; }
while (my $line = <DATA>) { chomp $line; }
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
source_file
  conditional_statement
    expression_stmt
      literal
    token
    var_decl
      variable
    token
    block_stmt
      token
      expression_stmt
        literal
      token
      token
  for_statement
    expression_stmt
      literal
    token
    scalar
      token
      token
    token
    binary_expression
      token
      expression_stmt
        literal
      token
    token
    block_stmt
      token
      expression_stmt
        literal
      token
      token
  loop_statement
    expression_stmt
      literal
    token
    var_decl
      variable
    token
    block_stmt
      token
      expression_stmt
        literal
      token
      token
```

## JSON AST

```json
{
  "path": "/tmp/variable_context_variations_test.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 3,
      "Column": 43,
      "Offset": 137
    },
    "children": [
      {
        "type": "conditional_statement",
        "start": {
          "Line": 1,
          "Column": 1,
          "Offset": 0
        },
        "end": {
          "Line": 1,
          "Column": 49,
          "Offset": 48
        },
        "children": [
          {
            "type": "expression_stmt",
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
            "children": [
              {
                "type": "literal",
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
                "value": "if",
                "kind": "string"
              }
            ]
          },
          {
            "type": "token",
            "start": {
              "Line": 1,
              "Column": 4,
              "Offset": 3
            },
            "end": {
              "Line": 1,
              "Column": 5,
              "Offset": 4
            },
            "text": "("
          },
          {
            "type": "var_decl",
            "start": {
              "Line": 1,
              "Column": 5,
              "Offset": 4
            },
            "end": {
              "Line": 1,
              "Column": 29,
              "Offset": 28
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 1,
                  "Column": 5,
                  "Offset": 4
                },
                "end": {
                  "Line": 1,
                  "Column": 29,
                  "Offset": 28
                },
                "name": "scoped",
                "sigil": "$"
              }
            ],
            "decl_type": "my"
          },
          {
            "type": "token",
            "start": {
              "Line": 1,
              "Column": 29,
              "Offset": 28
            },
            "end": {
              "Line": 1,
              "Column": 30,
              "Offset": 29
            },
            "text": ")"
          },
          {
            "type": "block_stmt",
            "start": {
              "Line": 1,
              "Column": 31,
              "Offset": 30
            },
            "end": {
              "Line": 1,
              "Column": 49,
              "Offset": 48
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 31,
                  "Offset": 30
                },
                "end": {
                  "Line": 1,
                  "Column": 32,
                  "Offset": 31
                },
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 1,
                  "Column": 33,
                  "Offset": 32
                },
                "end": {
                  "Line": 1,
                  "Column": 46,
                  "Offset": 45
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 1,
                      "Column": 33,
                      "Offset": 32
                    },
                    "end": {
                      "Line": 1,
                      "Column": 46,
                      "Offset": 45
                    },
                    "value": "print $scoped",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 46,
                  "Offset": 45
                },
                "end": {
                  "Line": 1,
                  "Column": 47,
                  "Offset": 46
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 1,
                  "Column": 48,
                  "Offset": 47
                },
                "end": {
                  "Line": 1,
                  "Column": 49,
                  "Offset": 48
                },
                "text": "}"
              }
            ]
          }
        ]
      },
      {
        "type": "for_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 49
        },
        "end": {
          "Line": 2,
          "Column": 46,
          "Offset": 94
        },
        "children": [
          {
            "type": "expression_stmt",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 49
            },
            "end": {
              "Line": 2,
              "Column": 4,
              "Offset": 52
            },
            "children": [
              {
                "type": "literal",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 49
                },
                "end": {
                  "Line": 2,
                  "Column": 4,
                  "Offset": 52
                },
                "value": "for",
                "kind": "string"
              }
            ]
          },
          {
            "type": "token",
            "start": {
              "Line": 2,
              "Column": 5,
              "Offset": 53
            },
            "end": {
              "Line": 2,
              "Column": 7,
              "Offset": 55
            },
            "text": "my"
          },
          {
            "type": "scalar",
            "start": {
              "Line": 2,
              "Column": 8,
              "Offset": 56
            },
            "end": {
              "Line": 2,
              "Column": 17,
              "Offset": 65
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 8,
                  "Offset": 56
                },
                "end": {
                  "Line": 2,
                  "Column": 9,
                  "Offset": 57
                },
                "text": "$"
              },
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 9,
                  "Offset": 57
                },
                "end": {
                  "Line": 2,
                  "Column": 17,
                  "Offset": 65
                },
                "text": "iterator"
              }
            ]
          },
          {
            "type": "token",
            "start": {
              "Line": 2,
              "Column": 18,
              "Offset": 66
            },
            "end": {
              "Line": 2,
              "Column": 19,
              "Offset": 67
            },
            "text": "("
          },
          {
            "type": "binary_expression",
            "start": {
              "Line": 2,
              "Column": 19,
              "Offset": 67
            },
            "end": {
              "Line": 2,
              "Column": 24,
              "Offset": 72
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 19,
                  "Offset": 67
                },
                "end": {
                  "Line": 2,
                  "Column": 21,
                  "Offset": 69
                },
                "text": "1."
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 2,
                  "Column": 21,
                  "Offset": 69
                },
                "end": {
                  "Line": 2,
                  "Column": 22,
                  "Offset": 70
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 2,
                      "Column": 21,
                      "Offset": 69
                    },
                    "end": {
                      "Line": 2,
                      "Column": 22,
                      "Offset": 70
                    },
                    "value": ".",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 22,
                  "Offset": 70
                },
                "end": {
                  "Line": 2,
                  "Column": 24,
                  "Offset": 72
                },
                "text": "10"
              }
            ]
          },
          {
            "type": "token",
            "start": {
              "Line": 2,
              "Column": 24,
              "Offset": 72
            },
            "end": {
              "Line": 2,
              "Column": 25,
              "Offset": 73
            },
            "text": ")"
          },
          {
            "type": "block_stmt",
            "start": {
              "Line": 2,
              "Column": 26,
              "Offset": 74
            },
            "end": {
              "Line": 2,
              "Column": 46,
              "Offset": 94
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 26,
                  "Offset": 74
                },
                "end": {
                  "Line": 2,
                  "Column": 27,
                  "Offset": 75
                },
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 2,
                  "Column": 28,
                  "Offset": 76
                },
                "end": {
                  "Line": 2,
                  "Column": 43,
                  "Offset": 91
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 2,
                      "Column": 28,
                      "Offset": 76
                    },
                    "end": {
                      "Line": 2,
                      "Column": 43,
                      "Offset": 91
                    },
                    "value": "print $iterator",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 43,
                  "Offset": 91
                },
                "end": {
                  "Line": 2,
                  "Column": 44,
                  "Offset": 92
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 45,
                  "Offset": 93
                },
                "end": {
                  "Line": 2,
                  "Column": 46,
                  "Offset": 94
                },
                "text": "}"
              }
            ]
          }
        ]
      },
      {
        "type": "loop_statement",
        "start": {
          "Line": 3,
          "Column": 1,
          "Offset": 95
        },
        "end": {
          "Line": 3,
          "Column": 43,
          "Offset": 137
        },
        "children": [
          {
            "type": "expression_stmt",
            "start": {
              "Line": 3,
              "Column": 1,
              "Offset": 95
            },
            "end": {
              "Line": 3,
              "Column": 6,
              "Offset": 100
            },
            "children": [
              {
                "type": "literal",
                "start": {
                  "Line": 3,
                  "Column": 1,
                  "Offset": 95
                },
                "end": {
                  "Line": 3,
                  "Column": 6,
                  "Offset": 100
                },
                "value": "while",
                "kind": "string"
              }
            ]
          },
          {
            "type": "token",
            "start": {
              "Line": 3,
              "Column": 7,
              "Offset": 101
            },
            "end": {
              "Line": 3,
              "Column": 8,
              "Offset": 102
            },
            "text": "("
          },
          {
            "type": "var_decl",
            "start": {
              "Line": 3,
              "Column": 8,
              "Offset": 102
            },
            "end": {
              "Line": 3,
              "Column": 25,
              "Offset": 119
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 3,
                  "Column": 8,
                  "Offset": 102
                },
                "end": {
                  "Line": 3,
                  "Column": 25,
                  "Offset": 119
                },
                "name": "line",
                "sigil": "$"
              }
            ],
            "decl_type": "my"
          },
          {
            "type": "token",
            "start": {
              "Line": 3,
              "Column": 25,
              "Offset": 119
            },
            "end": {
              "Line": 3,
              "Column": 26,
              "Offset": 120
            },
            "text": ")"
          },
          {
            "type": "block_stmt",
            "start": {
              "Line": 3,
              "Column": 27,
              "Offset": 121
            },
            "end": {
              "Line": 3,
              "Column": 43,
              "Offset": 137
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 27,
                  "Offset": 121
                },
                "end": {
                  "Line": 3,
                  "Column": 28,
                  "Offset": 122
                },
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 3,
                  "Column": 29,
                  "Offset": 123
                },
                "end": {
                  "Line": 3,
                  "Column": 40,
                  "Offset": 134
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 3,
                      "Column": 29,
                      "Offset": 123
                    },
                    "end": {
                      "Line": 3,
                      "Column": 40,
                      "Offset": 134
                    },
                    "value": "chomp $line",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 40,
                  "Offset": 134
                },
                "end": {
                  "Line": 3,
                  "Column": 41,
                  "Offset": 135
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 42,
                  "Offset": 136
                },
                "end": {
                  "Line": 3,
                  "Column": 43,
                  "Offset": 137
                },
                "text": "}"
              }
            ]
          }
        ]
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 137
}
```
