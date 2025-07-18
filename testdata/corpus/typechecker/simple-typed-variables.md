---
category: typechecker
subcategory: variable-declarations
tags:
    - basic-types
    - symbol-binding
    - variable-declarations
type_check: true
---

# Simple Typed Variables

Basic typed variable declarations with proper symbol binding behavior.

```perl
my Int $count = 42;
my Str $name = "Alice";
my Bool $active = true;
```

# Expected Symbol Table

The binder should create symbols for the **variables**, not the **type names**.
Type names (Int, Str, Bool) are type references and should not appear as symbols.
The improved binder now properly extracts type annotations and flags typed variables.

```
=== SYMBOLS ===
scalar active :: Bool [lexical|typed] at 1:1
scalar count :: Int [lexical|typed] at 1:1
scalar name :: Str [lexical|typed] at 1:1
=== TYPE ERRORS ===
No type errors
```

# Expected Type Analysis

```
Variable $count: type Int (inferred from annotation)
Variable $name: type Str (inferred from annotation)
Variable $active: type Bool (inferred from annotation)
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $count = 42;
my $name = "Alice";
my $active = true;
```

## Typed Perl Output

```perl
use v5.36;
my Int $count = 42;
my Str $name = "Alice";
my Bool $active = true;
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
  expression_statement
    var_decl
      variable
  token
```

## JSON AST

```json
{
  "path": "simple-typed-variables.pl",
  "root": {
    "type": "source_file",
    "start": {"Line": 1, "Column": 1, "Offset": 0},
    "end": {"Line": 4, "Column": 1, "Offset": 68},
    "children": [
      {
        "type": "expression_statement",
        "start": {"Line": 1, "Column": 1, "Offset": 0},
        "end": {"Line": 1, "Column": 19, "Offset": 18},
        "children": [
          {
            "type": "var_decl",
            "start": {"Line": 1, "Column": 1, "Offset": 0},
            "end": {"Line": 1, "Column": 19, "Offset": 18},
            "children": [
              {
                "type": "variable",
                "start": {"Line": 1, "Column": 1, "Offset": 0},
                "end": {"Line": 1, "Column": 19, "Offset": 18},
                "name": "Int",
                "sigil": "$"
              }
            ],
            "decl_type": "my"
          }
        ]
      },
      {
        "type": "expression_statement",
        "start": {"Line": 2, "Column": 1, "Offset": 20},
        "end": {"Line": 2, "Column": 23, "Offset": 42},
        "children": [
          {
            "type": "var_decl",
            "start": {"Line": 2, "Column": 1, "Offset": 20},
            "end": {"Line": 2, "Column": 23, "Offset": 42},
            "children": [
              {
                "type": "variable",
                "start": {"Line": 2, "Column": 1, "Offset": 20},
                "end": {"Line": 2, "Column": 23, "Offset": 42},
                "name": "Str",
                "sigil": "$"
              }
            ],
            "decl_type": "my"
          }
        ]
      },
      {
        "type": "expression_statement",
        "start": {"Line": 3, "Column": 1, "Offset": 44},
        "end": {"Line": 3, "Column": 23, "Offset": 66},
        "children": [
          {
            "type": "var_decl",
            "start": {"Line": 3, "Column": 1, "Offset": 44},
            "end": {"Line": 3, "Column": 23, "Offset": 66},
            "children": [
              {
                "type": "variable",
                "start": {"Line": 3, "Column": 1, "Offset": 44},
                "end": {"Line": 3, "Column": 23, "Offset": 66},
                "name": "Bool",
                "sigil": "$"
              }
            ],
            "decl_type": "my"
          }
        ]
      }
    ]
  },
  "type_annotations": [
    {
      "annotated_item": "$count",
      "type_expression": {
        "Kind": 0,
        "Name": "Int",
        "OriginalString": "Int"
      },
      "position": {"Line": 1, "Column": 1, "Offset": 0},
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$name",
      "type_expression": {
        "Kind": 0,
        "Name": "Str",
        "OriginalString": "Str"
      },
      "position": {"Line": 2, "Column": 1, "Offset": 0},
      "kind": "VarAnnotation"
    },
    {
      "annotated_item": "$active",
      "type_expression": {
        "Kind": 0,
        "Name": "Bool",
        "OriginalString": "Bool"
      },
      "position": {"Line": 3, "Column": 1, "Offset": 0},
      "kind": "VarAnnotation"
    }
  ],
  "errors": [],
  "source_length": 68
}
```

# Test Notes

This test verifies the core symbol binding fix where built-in type names
(Int, Str, Bool) are correctly treated as type references rather than
being incorrectly added to the symbol table as variable symbols.
