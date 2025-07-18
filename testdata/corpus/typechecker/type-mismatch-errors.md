---
category: typechecker
subcategory: type-errors
tags:
    - type-mismatch
    - error-detection
    - assignment-validation
type_check: true
---

# Type Mismatch Errors

Test cases for detecting type assignment mismatches.

```perl
my Int $number = 42;
my Str $text = "hello";

# These should cause type errors
$number = "not a number";
$text = 123;
```

# Expected Symbol Table

The binder properly extracts type annotations and flags typed variables.
Assignment validation is not yet implemented in the type checker.

```
=== SYMBOLS ===
scalar number :: Int [lexical|typed] at 1:1
scalar text :: Str [lexical|typed] at 1:1
=== TYPE ERRORS ===
No type errors
```

# Expected Type Analysis

```
Variable $number: declared as Int, assigned Str (ERROR)
Variable $text: declared as Str, assigned Int (ERROR)
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $number = 42;
my $text = "hello";

# These should cause type errors
$number = "not a number";
$text = 123;
```

## Typed Perl Output

```perl
use v5.36;
my Int $number = 42;
my Str $text = "hello";

# These should cause type errors
$number = "not a number";
$text = 123;
```

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
  expression_stmt
    literal
  expression_statement
    assignment_expression
      scalar
        token
        token
      token
      interpolated_string_literal
        expression_stmt
          literal
        expression_stmt
          literal
        expression_stmt
          literal
  token
  expression_statement
    assignment_expression
      scalar
        token
        token
      token
      token
  token
```

## JSON AST

```json
{
  "path": "type-mismatch-errors.pl",
  "root": {
    "type": "source_file",
    "start": {"Line": 1, "Column": 1, "Offset": 0},
    "end": {"Line": 7, "Column": 1, "Offset": 118},
    "children": [
      {
        "type": "expression_statement",
        "start": {"Line": 1, "Column": 1, "Offset": 0},
        "end": {"Line": 1, "Column": 20, "Offset": 19},
        "children": [
          {
            "type": "var_decl",
            "start": {"Line": 1, "Column": 1, "Offset": 0},
            "end": {"Line": 1, "Column": 20, "Offset": 19},
            "children": [
              {
                "type": "variable",
                "start": {"Line": 1, "Column": 1, "Offset": 0},
                "end": {"Line": 1, "Column": 20, "Offset": 19},
                "name": "Int",
                "sigil": "$"
              }
            ],
            "decl_type": "my"
          }
        ]
      },
      {
        "type": "token",
        "start": {"Line": 1, "Column": 20, "Offset": 19},
        "end": {"Line": 1, "Column": 21, "Offset": 20},
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {"Line": 2, "Column": 1, "Offset": 21},
        "end": {"Line": 2, "Column": 23, "Offset": 43},
        "children": [
          {
            "type": "var_decl",
            "start": {"Line": 2, "Column": 1, "Offset": 21},
            "end": {"Line": 2, "Column": 23, "Offset": 43},
            "children": [
              {
                "type": "variable",
                "start": {"Line": 2, "Column": 1, "Offset": 21},
                "end": {"Line": 2, "Column": 23, "Offset": 43},
                "name": "Str",
                "sigil": "$"
              }
            ],
            "decl_type": "my"
          }
        ]
      },
      {
        "type": "token",
        "start": {"Line": 2, "Column": 23, "Offset": 43},
        "end": {"Line": 2, "Column": 24, "Offset": 44},
        "text": ";"
      },
      {
        "type": "expression_statement",
        "start": {"Line": 5, "Column": 1, "Offset": 79},
        "end": {"Line": 5, "Column": 25, "Offset": 103},
        "children": [
          {
            "type": "assignment_expression",
            "start": {"Line": 5, "Column": 1, "Offset": 79},
            "end": {"Line": 5, "Column": 25, "Offset": 103},
            "children": [
              {
                "type": "scalar",
                "start": {"Line": 5, "Column": 1, "Offset": 79},
                "end": {"Line": 5, "Column": 8, "Offset": 86},
                "children": [
                  {
                    "type": "token",
                    "start": {"Line": 5, "Column": 1, "Offset": 79},
                    "end": {"Line": 5, "Column": 2, "Offset": 80},
                    "text": "$"
                  },
                  {
                    "type": "token",
                    "start": {"Line": 5, "Column": 2, "Offset": 80},
                    "end": {"Line": 5, "Column": 8, "Offset": 86},
                    "text": "number"
                  }
                ]
              },
              {
                "type": "token",
                "start": {"Line": 5, "Column": 9, "Offset": 87},
                "end": {"Line": 5, "Column": 10, "Offset": 88},
                "text": "="
              },
              {
                "type": "interpolated_string_literal",
                "start": {"Line": 5, "Column": 11, "Offset": 89},
                "end": {"Line": 5, "Column": 25, "Offset": 103}
              }
            ]
          }
        ]
      },
      {
        "type": "expression_statement",
        "start": {"Line": 6, "Column": 1, "Offset": 105},
        "end": {"Line": 6, "Column": 12, "Offset": 116},
        "children": [
          {
            "type": "assignment_expression",
            "start": {"Line": 6, "Column": 1, "Offset": 105},
            "end": {"Line": 6, "Column": 12, "Offset": 116},
            "children": [
              {
                "type": "scalar",
                "start": {"Line": 6, "Column": 1, "Offset": 105},
                "end": {"Line": 6, "Column": 6, "Offset": 110}
              },
              {
                "type": "token",
                "start": {"Line": 6, "Column": 9, "Offset": 113},
                "end": {"Line": 6, "Column": 12, "Offset": 116},
                "text": "123"
              }
            ]
          }
        ]
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "$number",
      "type_expression": {
        "Kind": 0,
        "Name": "Int",
        "OriginalString": "Int"
      },
      "position": {"Line": 1, "Column": 1, "Offset": 0},
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$text",
      "type_expression": {
        "Kind": 0,
        "Name": "Str",
        "OriginalString": "Str"
      },
      "position": {"Line": 2, "Column": 1, "Offset": 0},
      "kind": "VarAnnotation"
    }
  ],
  "errors": [],
  "source_length": 118
}
```

# Test Notes

This test verifies that the type checker properly detects assignment
mismatches between incompatible types and reports meaningful error messages.
