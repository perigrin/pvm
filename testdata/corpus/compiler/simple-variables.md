---
category: compiler
subcategory: simple-variables
tags:
    - typed-variables
    - clean-perl-output
type_check: false
---

# Simple Typed Variable

Basic typed variable declaration compilation

```perl
my Int $count = 42;
print "Count: $count\n";
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $count = 42;
print "Count: $count\n";
```

# Simple String Variable

String typed variable compilation

```perl
my Str $name = "hello";
print "Name: $name\n";
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $name = "hello";
print "Name: $name\n";
```

## Text AST

```
source_file
  expression_statement
    var_decl
      variable
  token
```

## JSON AST

```json
{
  "path": "simple-variables.pl",
  "root": {
    "type": "source_file",
    "start": {"Line": 1, "Column": 1, "Offset": 0},
    "end": {"Line": 1, "Column": 20, "Offset": 19},
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
    }
  ],
  "errors": [],
  "source_length": 19
}
```
