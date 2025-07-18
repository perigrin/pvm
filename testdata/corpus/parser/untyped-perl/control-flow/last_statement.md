---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - last
    - loop_control
---

# Last Statement

Last statement to break out of loop

```perl
foreach my $item (@list) {
    process($item);
    last if stop_condition($item);
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
foreach my $item (@list) {
    process($item);
    last if stop_condition($item);
}
```

## Typed Perl Output

```perl
foreach my $item (@list) {
    process($item);
    last if stop_condition($item);
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected AST

## Text AST

```
AST {
  Path: /tmp/simple_last.pl
  Source length: 6 characters
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      loopex_expression
        expression_stmt
          literal
    token
}
```

## JSON AST

```json
{
  "path": "/tmp/simple_last.pl",
  "root": {
    "type": "source_file",
    "start": {
      "Line": 1,
      "Column": 1,
      "Offset": 0
    },
    "end": {
      "Line": 2,
      "Column": 1,
      "Offset": 6
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
          "Column": 5,
          "Offset": 4
        },
        "children": [
          {
            "type": "loopex_expression",
            "start": {
              "Line": 1,
              "Column": 1,
              "Offset": 0
            },
            "end": {
              "Line": 1,
              "Column": 5,
              "Offset": 4
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
                  "Column": 5,
                  "Offset": 4
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
                      "Column": 5,
                      "Offset": 4
                    },
                    "value": "last",
                    "kind": "string"
                  }
                ]
              }
            ]
          }
        ]
      },
      {
        "type": "token",
        "start": {
          "Line": 1,
          "Column": 5,
          "Offset": 4
        },
        "end": {
          "Line": 1,
          "Column": 6,
          "Offset": 5
        },
        "text": ";"
      }
    ]
  },
  "type_annotations": [],
  "errors": [],
  "source_length": 6
}
```
