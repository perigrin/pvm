---
category: untyped-perl
subcategory: control-flow
tags:
    - control-flow
    - loops
    - foreach
---

# Simple Loop

Basic foreach loop over an array

```perl
my @items = ("apple", "banana", "cherry");
for my $item (@items) {
    print "Item: $item\n";
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my @items = ("apple", "banana", "cherry");
for my $item (@items) {
    print "Item: $item\n";
}
```

## Typed Perl Output

```perl
my @items = ("apple", "banana", "cherry");
for my $item (@items) {
    print "Item: $item\n";
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
  expression_statement
    var_decl
      variable
  token
  for_statement
    expression_stmt
      literal
    token
    scalar
      token
      token
    token
    array
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
```

## JSON Format

```json
{
  "path": "/tmp/simple-loop.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 4,
      "Column": 2,
      "Offset": 95
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
          "Column": 42,
          "Offset": 41
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
              "Column": 42,
              "Offset": 41
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
                  "Column": 42,
                  "Offset": 41
                },
                "name": "items",
                "sigil": "@"
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
          "Column": 42,
          "Offset": 41
        },
        "end": {
          "Line": 1,
          "Column": 43,
          "Offset": 42
        },
        "text": ";"
      },
      {
        "type": "for_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 43
        },
        "end": {
          "Line": 4,
          "Column": 2,
          "Offset": 95
        },
        "children": [
          {
            "type": "expression_stmt",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 43
            },
            "end": {
              "Line": 2,
              "Column": 4,
              "Offset": 46
            },
            "children": [
              {
                "type": "literal",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 43
                },
                "end": {
                  "Line": 2,
                  "Column": 4,
                  "Offset": 46
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
              "Offset": 47
            },
            "end": {
              "Line": 2,
              "Column": 7,
              "Offset": 49
            },
            "text": "my"
          },
          {
            "type": "scalar",
            "start": {
              "Line": 2,
              "Column": 8,
              "Offset": 50
            },
            "end": {
              "Line": 2,
              "Column": 13,
              "Offset": 55
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 8,
                  "Offset": 50
                },
                "end": {
                  "Line": 2,
                  "Column": 9,
                  "Offset": 51
                },
                "text": "$"
              },
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 9,
                  "Offset": 51
                },
                "end": {
                  "Line": 2,
                  "Column": 13,
                  "Offset": 55
                },
                "text": "item"
              }
            ]
          },
          {
            "type": "token",
            "start": {
              "Line": 2,
              "Column": 14,
              "Offset": 56
            },
            "end": {
              "Line": 2,
              "Column": 15,
              "Offset": 57
            },
            "text": "("
          },
          {
            "type": "array",
            "start": {
              "Line": 2,
              "Column": 15,
              "Offset": 57
            },
            "end": {
              "Line": 2,
              "Column": 21,
              "Offset": 63
            },
            "children": [
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 2,
                  "Column": 15,
                  "Offset": 57
                },
                "end": {
                  "Line": 2,
                  "Column": 16,
                  "Offset": 58
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 2,
                      "Column": 15,
                      "Offset": 57
                    },
                    "end": {
                      "Line": 2,
                      "Column": 16,
                      "Offset": 58
                    },
                    "value": "@",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 16,
                  "Offset": 58
                },
                "end": {
                  "Line": 2,
                  "Column": 21,
                  "Offset": 63
                },
                "text": "items"
              }
            ]
          },
          {
            "type": "token",
            "start": {
              "Line": 2,
              "Column": 21,
              "Offset": 63
            },
            "end": {
              "Line": 2,
              "Column": 22,
              "Offset": 64
            },
            "text": ")"
          },
          {
            "type": "block_stmt",
            "start": {
              "Line": 2,
              "Column": 23,
              "Offset": 65
            },
            "end": {
              "Line": 4,
              "Column": 2,
              "Offset": 95
            },
            "children": [
              {
                "type": "token",
                "start": {
                  "Line": 2,
                  "Column": 23,
                  "Offset": 65
                },
                "end": {
                  "Line": 2,
                  "Column": 24,
                  "Offset": 66
                },
                "text": "{"
              },
              {
                "type": "expression_stmt",
                "start": {
                  "Line": 3,
                  "Column": 5,
                  "Offset": 71
                },
                "end": {
                  "Line": 3,
                  "Column": 26,
                  "Offset": 92
                },
                "children": [
                  {
                    "type": "literal",
                    "start": {
                      "Line": 3,
                      "Column": 5,
                      "Offset": 71
                    },
                    "end": {
                      "Line": 3,
                      "Column": 26,
                      "Offset": 92
                    },
                    "value": "print \"Item: $item\\n\"",
                    "kind": "string"
                  }
                ]
              },
              {
                "type": "token",
                "start": {
                  "Line": 3,
                  "Column": 26,
                  "Offset": 92
                },
                "end": {
                  "Line": 3,
                  "Column": 27,
                  "Offset": 93
                },
                "text": ";"
              },
              {
                "type": "token",
                "start": {
                  "Line": 4,
                  "Column": 1,
                  "Offset": 94
                },
                "end": {
                  "Line": 4,
                  "Column": 2,
                  "Offset": 95
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
  "source_length": 95
}
```
