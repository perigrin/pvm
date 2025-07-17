---
category: untyped-perl
subcategory: variables
tags:
    - variables
    - declarations
    - scalars
---

# Simple Variable Declaration

Basic scalar variable declaration with assignment

```perl
my $name = "example";
my $count = 42;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $name = "example";
my $count = 42;
```

## Typed Perl Output

```perl
my $name = "example";
my $count = 42;
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
  expression_statement
    var_decl
      variable
  token
```

## JSON Format

```json
{
  "path": "/tmp/test_code.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 2,
      "Column": 16,
      "Offset": 37
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
          "Column": 21,
          "Offset": 20
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
              "Column": 21,
              "Offset": 20
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
                  "Column": 21,
                  "Offset": 20
                },
                "name": "name",
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
          "Column": 21,
          "Offset": 20
        },
        "end": {
          "Line": 1,
          "Column": 22,
          "Offset": 21
        },
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {
          "Line": 2,
          "Column": 1,
          "Offset": 22
        },
        "end": {
          "Line": 2,
          "Column": 15,
          "Offset": 36
        },
        "children": [
          {
            "type": "var_decl",
            "start": {
              "Line": 2,
              "Column": 1,
              "Offset": 22
            },
            "end": {
              "Line": 2,
              "Column": 15,
              "Offset": 36
            },
            "children": [
              {
                "type": "variable",
                "start": {
                  "Line": 2,
                  "Column": 1,
                  "Offset": 22
                },
                "end": {
                  "Line": 2,
                  "Column": 15,
                  "Offset": 36
                },
                "name": "count",
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
          "Column": 15,
          "Offset": 36
        },
        "end": {
          "Line": 2,
          "Column": 16,
          "Offset": 37
        },
        "text": ";"
      }
    ]
  },
  "source_length": 37
}
