---
category: untyped-perl
subcategory: expressions
tags:
    - strings
    - concatenation
    - interpolation
---

# String Operations

Basic string operations and interpolation

```perl
my $first = "Hello";
my $second = "World";
my $greeting = $first . ", " . $second . "!";
my $interpolated = "Message: $greeting";
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $first = "Hello";
my $second = "World";
my $greeting = $first . ", " . $second . "!";
my $interpolated = "Message: $greeting";
```

## Typed Perl Output

```perl
my $first = "Hello";
my $second = "World";
my $greeting = $first . ", " . $second . "!";
my $interpolated = "Message: $greeting";
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
  "path": "/tmp/string-operations.pl",
  "root": {
    "type": "source_file",
    "start": { "Line": 1, "Column": 1, "Offset": 0 },
    "end": { "Line": 4, "Column": 41, "Offset": 129 },
    "children": [
      {
        "type": "expression_statement",
        "start": { "Line": 1, "Column": 1, "Offset": 0 },
        "end": { "Line": 1, "Column": 20, "Offset": 19 },
        "children": [
          {
            "type": "var_decl",
            "start": { "Line": 1, "Column": 1, "Offset": 0 },
            "end": { "Line": 1, "Column": 20, "Offset": 19 },
            "children": [
              {
                "type": "variable",
                "start": { "Line": 1, "Column": 1, "Offset": 0 },
                "end": { "Line": 1, "Column": 20, "Offset": 19 },
                "name": "first",
                "sigil": "$"
              }
            ],
            "decl_type": "my"
          }
        ]
      },
      {
        "type": "token",
        "start": { "Line": 1, "Column": 20, "Offset": 19 },
        "end": { "Line": 1, "Column": 21, "Offset": 20 },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 129
}
```
